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
