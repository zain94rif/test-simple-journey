# Laporan Perkembangan — Helpdesk Technical Test

**Tanggal:** 9 November 2026  
**Status keseluruhan:** Backend selesai dan terverifikasi; frontend selesai, lolos build/type/lint/test, dan terverifikasi end-to-end dengan backend.  
**Acuan:** `docs/PRD-00-Umum.md`, `docs/PRD-01-Backend.md`, `docs/PRD-02-Frontend.md`

---

## 1. Ringkasan

Backend Helpdesk sudah diimplementasikan penuh sesuai kontrak PRD-00 dan PRD-01:

- REST API Go dengan Gin, PostgreSQL (pgx/v5), JWT HS256, dan bcrypt.
- Aturan peran, status, kepemilikan tiket, dan catatan internal ditegakkan di service layer.
- Migrasi embedded dan seeder idempotent.
- Seed data diperkaya agar semua peran, filter, komentar, dan daftar pengguna dapat diuji manual.
- Dokumentasi OpenAPI/Swagger tersedia dan dapat dicoba langsung dari browser.
- Unit test, `go vet`, `go test -race`, build Podman, serta uji end-to-end lokal sudah lulus.

Frontend Vue 3 + Vite + Pinia + Vue Router dengan Bun juga sudah dibangun sesuai PRD-02: seluruh rute, guard peran, filter URL-synced, halaman admin, dan halaman error tersedia. `tsc`, ESLint, build produksi, dan 33 unit test lulus, dan alur login → daftar tiket terverifikasi langsung terhadap backend berjalan. Detail ada di bagian **8a**.

---

## 2. Stack dan Struktur

| Komponen | Pilihan |
|---|---|
| Bahasa | Go 1.25 (modul), toolchain lokal Go 1.27.3 |
| HTTP framework | Gin v1.12 |
| Database | PostgreSQL (lokal 18.6; PRD mensyaratkan 16+) |
| Driver DB | `github.com/jackc/pgx/v5` + pgxpool |
| Autentikasi | JWT HS256, bcrypt cost 12 |
| Migrasi | Runner embedded sendiri (`schema_migrations` + advisory lock) |
| Logging | `log/slog` (JSON) |
| Dokumentasi | Swag v1.16.6 + gin-swagger |
| Container | Dockerfile multi-stage, terverifikasi dengan Podman 6.1.4 |

```text
backend/
├── cmd/server/main.go            # wiring, retry DB, graceful shutdown, anotasi Swagger
├── docs/                         # hasil generate Swagger (docs.go, swagger.json/yaml)
├── internal/
│   ├── auth/                     # bcrypt + JWT manager
│   ├── config/                   # env parsing & validasi
│   ├── domain/                   # entity, enum, error contract, matriks transisi
│   ├── http/
│   │   ├── handler/              # handler tipis + DTO response
│   │   ├── middleware/           # JWT, role, CORS, request ID, body limit, log, recovery
│   │   ├── response/             # envelope sukses/error + pagination meta
│   │   └── router.go             # routing + Swagger UI
│   ├── migrate/                  # runner migrasi idempotent
│   ├── repository/postgres/      # query SQL parameterized
│   ├── seed/                     # akun + data demo idempotent
│   └── service/                  # seluruh aturan bisnis & otorisasi
├── migrations/                   # SQL embedded (up/down)
├── Dockerfile
├── .env.example
└── go.mod / go.sum
```

---

## 3. Endpoint yang Tersedia

| Method | Path | Akses | Keterangan |
|---|---|---|---|
| GET | `/api/health` | Publik | Cek koneksi database |
| POST | `/api/auth/register` | Publik | Registrasi, peran selalu `requester` |
| POST | `/api/auth/login` | Publik | Login, mengembalikan JWT |
| GET | `/api/me` | Terautentikasi | Profil pengguna saat ini |
| GET | `/api/tickets` | Terautentikasi | List + filter + search + sort + pagination |
| POST | `/api/tickets` | Terautentikasi | Buat tiket (status awal `open`) |
| GET | `/api/tickets/{id}` | Pemilik / staff | Detail tiket + komentar |
| PATCH | `/api/tickets/{id}` | Pemilik / staff | Ubah status, prioritas, assignee |
| POST | `/api/tickets/{id}/comments` | Pemilik / staff | Komentar publik atau catatan internal |
| GET | `/api/users` | Admin | List + search + filter peran + pagination |
| PATCH | `/api/users/{id}/role` | Admin | Ubah peran pengguna |

Aturan bisnis yang sudah diterapkan:

- Requester hanya dapat melihat dan mengubah tiket miliknya; akses tiket orang lain menghasilkan `403`.
- Hanya 5 transisi status PRD yang diizinkan; selain itu `409 INVALID_TRANSITION`.
- Catatan internal tidak pernah dikirim ke requester, termasuk pada detail tiket.
- Agent hanya dapat mengambil/menugaskan tiket untuk dirinya sendiri; admin dapat menugaskan agent atau admin.
- Agent yang memindahkan tiket `open` ke `in_progress` otomatis menjadi assignee.
- Admin tidak dapat menurunkan perannya sendiri (`409 SELF_ROLE_CHANGE`) atau admin terakhir (`409 LAST_ADMIN`).
- Peran dibaca ulang dari database pada setiap request, sehingga perubahan peran langsung berlaku tanpa login ulang.
- Semua timestamp response dinormalisasi ke UTC.
- Error memakai envelope konsisten: `{"error":{"code","message","details?"}}`.
- Body request dibatasi 1 MB; setiap response menyertakan `X-Request-ID`.

---

## 4. Seed Data untuk Pengujian Manual

Seed dijalankan otomatis saat startup bila `AUTO_SEED=true`. Seluruh proses idempotent: akun dicek berdasarkan email, tiket berdasarkan requester + judul, dan komentar berdasarkan tiket + author + isi. Menjalankan server berulang kali tidak menggandakan data.

### 4.1 Akun

| Peran | Nama | Email | Password |
|---|---|---|---|
| Admin | Admin Helpdesk | `admin@helpdesk.local` | `Admin123!` |
| Agent | Agent Satu | `agent@helpdesk.local` | `Agent123!` |
| Agent | Agent Dua | `agent2@helpdesk.local` | `Agent123!` |
| Requester | Requester Satu | `requester@helpdesk.local` | `Requester123!` |
| Requester | Requester Dua | `requester2@helpdesk.local` | `Requester123!` |
| Requester | Requester Tiga | `requester3@helpdesk.local` | `Requester123!` |

Password akun tambahan otomatis mengikuti password akun dasar, tetapi dapat dioverride lewat `SEED_AGENT2_PASSWORD`, `SEED_REQUESTER2_PASSWORD`, dan `SEED_REQUESTER3_PASSWORD`.

### 4.2 Tiket demo

Total **12 tiket** dan **20 komentar**, mencakup seluruh status, prioritas, dan kategori:

| Status | Jumlah | Prioritas | Jumlah | Kategori | Jumlah |
|---|---|---|---|---|---|
| `open` | 5 | `urgent` | 2 | `hardware` | 3 |
| `in_progress` | 3 | `high` | 3 | `software` | 3 |
| `resolved` | 2 | `medium` | 3 | `network` | 3 |
| `closed` | 2 | `low` | 4 | `access` | 2 |
| | | | | `other` | 1 |

Distribusi per requester: masing-masing Requester Satu, Dua, dan Tiga memiliki 4 tiket. Sebagian tiket sudah di-assign ke Agent Satu/Agent Dua, sebagian belum di-assign, sehingga filter `assignee_id`, `unassigned`, dan pengujian klaim tiket dapat dilakukan.

Data demo hanya dibuat bila `SEED_DEMO_DATA=true` (default). Set `SEED_DEMO_DATA=false` untuk hanya membuat enam akun tanpa tiket/komentar.

Untuk mengulang dari kondisi bersih:

```bash
psql -h localhost -U helpdesk -d helpdesk \
  -c "TRUNCATE comments, tickets, users RESTART IDENTITY CASCADE;"
cd backend && go run ./cmd/server
```

---

## 5. Dokumentasi Swagger

- UI: **http://localhost:8080/swagger/index.html**
- Spesifikasi JSON: **http://localhost:8080/swagger/doc.json**
- Spesifikasi YAML: `backend/docs/swagger.yaml`

Seluruh endpoint memiliki ringkasan, parameter, contoh request, skema response, dan kode error. Endpoint terproteksi memakai skema `BearerAuth`; pada UI klik **Authorize** lalu isi `Bearer <token>`.

Regenerasi dokumentasi setelah handler/annotation berubah:

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init \
  -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

File hasil generate (`backend/docs/`) ikut di-commit agar build tidak bergantung pada tool `swag` saat runtime.

---

## 6. Cara Menjalankan

### Lokal

Prasyarat: Go 1.25+, PostgreSQL 16+, database `helpdesk` dan role `helpdesk`.

```bash
cd backend
cp .env.example .env      # sesuaikan DATABASE_URL, JWT_SECRET, password seed
go run ./cmd/server
```

Migrasi dan seed berjalan otomatis. Cek kesehatan:

```bash
curl http://localhost:8080/api/health
# {"db":"up","status":"ok"}
```

### Podman

```bash
cd backend
podman build -t helpdesk-backend:test .
podman run --rm --network=host --env-file .env helpdesk-backend:test
```

Build dan run sudah diverifikasi; endpoint health mengembalikan status `ok`.

---

## 7. Hasil Verifikasi

| Pemeriksaan | Hasil |
|---|---|
| `go build ./...` | Lulus |
| `go vet ./...` | Lulus |
| `go test ./... -race -count=1` | Lulus, **35 test function** |
| Coverage `internal/service` | 72,5% |
| Coverage `internal/auth` | 81,5% |
| Seed idempotent | 6 user, 12 tiket, 20 komentar tetap sama setelah restart |
| Swagger UI & `doc.json` | HTTP 200, 9 path terdokumentasi |
| Podman build & run | Berhasil, health check `{"db":"up","status":"ok"}` |

Uji end-to-end manual yang sudah dilakukan:

- Register, login, `GET /me`, dan penolakan token kosong (`401`).
- Pembuatan tiket, klaim oleh agent, dan auto-assign.
- Catatan internal hanya terlihat oleh agent/admin; requester hanya melihat komentar publik.
- Requester tidak dapat membaca/mengubah tiket requester lain (`403`).
- Transisi status tidak valid (`open → resolved`) menghasilkan `409 INVALID_TRANSITION`.
- Duplicate email (termasuk beda kapitalisasi) menghasilkan `409 DUPLICATE_EMAIL`.
- Filter status/prioritas/assignee/unassigned, pencarian judul, sorting, dan pagination.
- Promosi peran oleh admin langsung berlaku pada JWT yang sama.
- Admin tidak dapat mengubah perannya sendiri (`409`).
- Payload JSON rusak, enum tidak valid, dan `per_page` di luar batas menghasilkan `400` dengan detail field.

---

## 8. Kendala dan Catatan Saat Ini

1. **Testcontainers sengaja ditunda.** Belum ada integration test yang menyentuh PostgreSQL sungguhan. Saat ini repository diuji tidak langsung melalui fake repository di service test. Setelah aplikasi selesai, disarankan menambah integration test dengan database sementara.
2. **Healthcheck Dockerfile tidak terbaca Podman.** Format OCI yang dipakai Podman mengabaikan instruksi `HEALTHCHECK` di Dockerfile. Saat `docker-compose.yml` dibuat, healthcheck harus didefinisikan ulang di level compose. Instruksi tersebut tetap dipertahankan agar kompatibel dengan Docker.
3. **Belum ada `docker-compose.yml`.** Compose penuh (backend + PostgreSQL + frontend) sengaja ditunda sampai frontend tersedia agar konfigurasi tidak ditulis dua kali.
4. **Versi PostgreSQL.** Pengembangan dan pengujian memakai PostgreSQL 18.6, sedangkan PRD menargetkan 16+. Query yang dipakai hanya fitur standar (`ILIKE`, `pg_trgm`, `TIMESTAMPTZ`), tetapi kombinasi dengan PostgreSQL 16 belum diuji langsung.
5. **Dokumentasi Swagger bersifat generated.** Setiap perubahan annotation atau DTO harus diikuti regenerasi `backend/docs`. Ada risiko dokumentasi usang bila langkah ini terlewat; perlu dijadikan bagian dari checklist rilis.
6. **Belum ada rate limiting / refresh token / audit log.** Ketiganya tidak diminta PRD. Saat ini token berumur 24 jam dan tidak dapat dicabut sebelum kedaluwarsa selain dengan mengganti `JWT_SECRET`.
7. **Seed memakai waktu relatif saat pertama dibuat.** Karena idempotent, timestamp tiket demo tidak bergeser pada restart berikutnya. Ini disengaja agar data manual test stabil, tetapi bukan pengganti data produksi.
8. **Seeder melakukan bcrypt hashing hanya untuk akun baru.** Seed pertama memerlukan sekitar 1,3 detik untuk enam akun (cost 12); startup berikutnya jauh lebih cepat.
9. **Frontend sudah dibangun dan terverifikasi end-to-end.** Lihat bagian **8a. Status Frontend**. Sebelum itu, seluruh kapabilitas hanya dapat diuji lewat curl/Postman/Swagger.
10. **Belum ada CI.** Build, vet, dan test masih dijalankan manual. Pipeline CI dapat ditambahkan setelah struktur frontend masuk ke repositori.
11. **Peringatan deprecation minor di frontend.** `pinia@2.3.1` dan `lucide-vue-next` masih diberi peringatan deprecated saat instalasi (API yang dipakai tetap stabil). Untuk masa depan, pertimbangkan upgrade ke `pinia@3` dan `lucide-vue-next` terbaru.
12. **Docker frontend belum diuji di Podman.** `Dockerfile` memakai `npm install` (bukan lockfile tertentu) sehingga build container tidak bergantung pada `bun.lock`, tetapi image belum dibangun di Podman untuk frontend.

---

## 8a. Status Frontend

Frontend mengikuti `docs/PRD-02-Frontend.md`, dibangun dengan **Vue 3 + `<script setup>` + TypeScript + Vite 6 + Pinia + Vue Router 4 + Axios + Tailwind 4**, semuanya dikelola lewat **Bun** di `frontend/`.

Hasil verifikasi:

| Butir | Status |
| --- | --- |
| `bun x tsc --noEmit` | 0 error |
| `bun x eslint .` | 0 error, 0 warning |
| `bun x vite build` | sukses, `dist/` ~198 kB JS (75 kB gzip) |
| `bun x vitest --run` | **51 test lulus di 9 file** (happy-dom + @vue/test-utils) |
| Dev server + proxy `/api` → `:8080` | `POST /api/auth/login`, `GET /api/me`, `GET /api/tickets`, `GET /api/tickets/:id` semua 200 |

Rute yang terdaftar di router dengan guard: `/login`, `/register` (guestOnly), `/tickets`, `/tickets/new`, `/tickets/:id` (requiresAuth), `/admin/users` (peran admin), `/forbidden`, serta fallback 404. Guard menjalankan `bootstrap()` sekali per load (token dari `localStorage["helpdesk_token"]`, sesi selalu divalidasi lewat `GET /me`), lalu menangani `guestOnly`, `requiresAuth` (dengan `?redirect=`), dan `roles` (admin; non-admin diarahkan ke `/forbidden`).

Aturan perizinan backend yang dimirror di UI:

- Filter daftar tiket tersinkron ke query URL (`?status=&priority=&assignee=&search=&page=`), pencarian memakai debounce, dan `AbortController` membatalkan request in-flight agar filter cepat tidak saling menimpa.
- `TicketActions` hanya menampilkan transisi yang sah menurut peran dan status; `TicketEditPanel` (prioritas/assignee) hanya untuk agent/admin.
- `RoleSelect` dinonaktifkan bila baris yang dipilih adalah pengguna yang sedang login (aturan "tidak boleh mengubah peran sendiri" di backend menghasilkan 409).
- `CommentForm` menyembunyikan opsi komentar internal untuk requester.
- Interceptor axios menangani 401 global: token dihapus dan pengguna diarahkan ke login dengan menyimpan `?redirect=`. Envelope error `{"error":{...}}` dinormalisasi di `src/lib/errors.ts`.

Unit test (`frontend/tests/`) mencakup subset FE-T01..FE-T12 dari PRD §13.2: matriks perizinan tiket (`useTicketPermissions`, 8 kasus peran/status), login/bootstrap/logout di store auth, guard router (redirect login untuk tamu dengan `?redirect=`, guestOnly mem-bounce pengguna terautentikasi, requester ditolak di `/admin/users` ke `/forbidden`, admin lolos, rute tak dikenal ke 404), label `StatusBadge`/`PriorityBadge`, `RoleSelect` (dinonaktifkan untuk diri sendiri dengan petunjuk, meng-emis event peran untuk pengguna lain), `CommentForm` (validasi komentar kosong, requester tidak memiliki opsi is-internal, staff dapat mengirim `is_internal=true`, textarea dikosongkan setelah kirim), `SkeletonBlock` (regresi crash yang ditemukan di runtime), dan strategi caching daftar: daftar tiket selalu di-fetch ulang setiap masuk halaman karena datanya dinamis (agent lain dapat mengubah status sewaktu-waktu), sementara daftar pengguna — yang hampir tidak berubah sendiri — dicache di Pinia dengan TTL 30 detik dan invalidasi otomatis setelah `updateRole` dan logout. MSW tidak digunakan; test memakai `vi.mock('@/lib/api')` ringan tanpa server mock tambahan.

File tambahan: `frontend/README.md` (cara menjalankan, menguji, struktur, akun seed), `frontend/Dockerfile` + `nginx.conf` (SPA fallback, proxy `/api/` ke `backend:8080`, gzip, cache aset). `docker-compose.yml` di root (dev, podman-compose): Postgres 16 + backend `golang:1.22-alpine` (selaras PRD) dengan air hot-reload + frontend `oven/bun:1.4.2-alpine` dengan Vite HMR — source di-mount sebagai volume sehingga edit kode tidak memerlukan build ulang image; Dockerfile produksi FE (bun 1.4.2) & BE (Go 1.22) selaras dengan toolchain compose tersebut.

---

## 9. Langkah Berikutnya

1. ~~Membangun frontend Vue 3 + Vite + Pinia + Vue Router dengan Bun sesuai `docs/PRD-02-Frontend.md`.~~ **Selesai** (lihat bagian 8a).
2. Menambahkan `docker-compose.yml` (PostgreSQL + backend + frontend) dengan healthcheck di level compose.
3. Menambahkan integration test PostgreSQL (menggantikan penundaan testcontainers).
4. Menyusun README root yang menggabungkan cara menjalankan backend, frontend, dan compose.
5. Menambahkan CI untuk `go build`, `go vet`, `go test -race`, serta build & test frontend.
