# PRD Helpdesk Internal — Index

Dokumen PRD untuk technical test Fullstack Shift Engineer, dipisah menjadi kontrak bersama,
Backend, dan Frontend.

| Dokumen | Isi | Audiens |
|---|---|---|
| [PRD-00-Umum.md](./PRD-00-Umum.md) | Latar belakang, peran, matriks akses, siklus status, kontrak API, model data, error, NFR, asumsi | Semua |
| [PRD-01-Backend.md](./PRD-01-Backend.md) | Arsitektur Go, struktur folder, API detail, validasi, skema DB, seed, env, keamanan, test, Docker | Backend |
| [PRD-02-Frontend.md](./PRD-02-Frontend.md) | Stack Vue, rute & guard, Pinia, HTTP client, komponen, spesifikasi 6 halaman, state, responsif, test, Docker | Frontend |

## Aturan inti (ringkas)

### Peran
- **Requester**: tiket miliknya saja; komentar publik; tutup/batalkan/buka kembali tiketnya.
- **Agent**: semua tiket; ambil tiket; ubah status/prioritas; catatan internal.
- **Admin**: semua hak Agent + kelola pengguna/peran.

### Transisi status yang diizinkan
| Dari | Ke | Oleh |
|---|---|---|
| `open` | `in_progress` | Agent, Admin |
| `in_progress` | `resolved` | Agent, Admin |
| `resolved` | `closed` | Requester, Agent, Admin |
| `resolved` | `open` | Requester |
| `open` | `closed` | Requester, Admin |

Selain tabel ini ditolak `409 INVALID_TRANSITION`.

### Keputusan teknis utama
- Backend: Go 1.22 + Gin + PostgreSQL 16 + pgx + bcrypt + JWT HS256, arsitektur handler → service → repository.
- Frontend: Vue 3 `<script setup>` + Vite + Pinia + Vue Router + Axios + Tailwind CSS.
- Error API memakai envelope `{ error: { code, message, details? } }`; sukses `{ data, meta? }`.
- Token disimpan di `localStorage`; 401 terpusat membersihkan sesi dan redirect ke login.
- Seed idempotent; migrasi versioned; konfigurasi via environment variable.
- Target test: 10 test backend dan 12 test frontend (melebihi minimum 5).

## Cara memakai dokumen ini
1. Baca PRD-00 terlebih dahulu sebagai kontrak.
2. Implementasi backend mengikuti PRD-01; frontend mengikuti PRD-02.
3. Jika ada perubahan kontrak, ubah PRD-00 dan kedua dokumen turunan agar tetap sinkron.
4. Asumsi yang belum dikonfirmasi rekrutmen tercantum di PRD-00 §12–13 dan wajib ditulis di README.

## Screenshot aplikasi

Tangkapan layar hasil uji manual (frontend di `http://localhost:5173`, backend Swagger di `http://localhost:8080/swagger/index.html`) tersedia di folder [`docs/photo/`](./photo/):

| File | Keterangan |
|---|---|
| `register.png` | Halaman registrasi akun baru; setelah berhasil user otomatis login sebagai `requester`. |
| `login.png` | Halaman login; memakai akun seed (mis. `admin@helpdesk.local` / `Admin123!`). |
| `make_ticket_page.png` | Form "Buat Tiket": field wajib ditandai `*`, pesan validasi per-field ("wajib diisi" vs batas minimum), ringkasan error, dan fokus otomatis ke field pertama yang salah. |
| `request_page.png` | Dashboard/requester view: daftar tiket milik sendiri dengan filter status & prioritas. |
| `ticket_view.png` | Detail tiket: kronologi komentar (publik/internal sesuai peran), aksi transisi status sesuai matriks akses. |
| `users_view.png` | Halaman admin: daftar pengguna dengan ubah peran (requester/agent/admin). |
| `swagger_BE.png` | Dokumentasi API Swagger backend yang di-serve di `/swagger/index.html`. |

## Menjalankan aplikasi dengan docker-compose (Podman)

Stack dev di root repo menggunakan `docker-compose.yml` dengan **hot-reload** — tidak perlu
build image ulang saat kode berubah (frontend via Vite HMR, backend via `air`).

```bash
# 1. Nyalakan stack (db + backend + frontend)
podman-compose -f docker-compose.yml up -d

# 2. Tunggu sampai semua kontainer (healthy) — cold start pertama ±1–2 menit
#    (backend perlu `go install air` + compile pertama; frontend perlu `bun install`)
watch podman ps                          # atau:
podman logs -f helpdesk-dev-backend      # tunggu log "server berjalan ... port 8080"
podman logs -f helpdesk-dev-frontend     # tunggu "Local: http://localhost:5173/"

# 3. Akses aplikasi
#   Frontend      : http://localhost:5173
#   Backend API   : http://localhost:8080/api
#   Swagger       : http://localhost:8080/swagger/index.html
#   DB (host)     : psql "postgres://helpdesk:helpdesk@localhost:5433/helpdesk"
#     (host port 5433 supaya tidak bentrok dengan Postgres native di 5432)

# 4. Hentikan / bersihkan
podman-compose -f docker-compose.yml down          # matikan (volume db tetap ada)
podman-compose -f docker-compose.yml down -v       # matikan + hapus volume db (data hilang)
```

**Akun seed** (idempotent, dibuat otomatis saat seeder berjalan):

| Peran | Email | Password |
|---|---|---|
| Admin | `admin@helpdesk.local` | `Admin123!` |
| Agent | `agent@helpdesk.local`, `agent2@helpdesk.local` | `Agent123!` |
| Requester | `requester1@helpdesk.local` … `requester3@helpdesk.local` | `Requester123!` |

**Catatan arsitektur compose:**
- `db`: `docker.io/library/postgres:16-alpine` + healthcheck; data di volume `pgdata`.
- `backend`: `docker.io/library/golang:1.26-alpine` menjalankan `air -c air.dev.toml`; kode di-bind-mount sehingga perubahan `*.go` langsung re-compile.
- `frontend`: `docker.io/oven/bun:1` menjalankan `bunx --bun vite --host`; kode di-bind-mount, `node_modules` di volume terpisah, polling watcher aktif (`CHOKIDAR_USEPOLLING`).
- Frontend mem-proxy `/api` ke `http://backend:8080` (via `VITE_PROXY_TARGET`), jadi cukup buka satu port 5173.

