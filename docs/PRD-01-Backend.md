# PRD-01 — Backend (Go)

| Item | Keterangan |
|---|---|
| Produk | Helpdesk Internal — REST API |
| Versi | 1.0 |
| Tanggal | 8 November 2026 |
| Status | Draft untuk implementasi |
| Dokumen induk | [PRD-00-Umum.md](./PRD-00-Umum.md) |
| Bahasa | Go 1.22+ |

---

## 1. Ringkasan dan tujuan

Membangun REST API helpdesk yang aman, konsisten, dan mudah diuji. Backend adalah **satu-satunya
otoritas** untuk aturan bisnis: autentikasi, otorisasi, kepemilikan tiket, transisi status, dan
visibilitas catatan internal. Frontend tidak boleh diandalkan untuk menegakkan aturan apa pun.

### Tujuan teknis
1. Struktur berlapis (handler → service → repository) sehingga aturan bisnis dapat diuji tanpa HTTP.
2. Semua error memiliki bentuk response dan kode HTTP yang konsisten.
3. Aturan transisi status terpusat di satu tempat dan diuji menyeluruh.
4. Migrasi dan seed dapat dijalankan berulang tanpa merusak data.
5. Satu perintah Docker menjalankan seluruh stack.

### Non-goals
- Bukan sistem multi-tenant.
- Tidak menyediakan refresh token, OAuth, atau SSO.
- Tidak menyimpan audit log perubahan status (bonus opsional).

## 2. Stack dan dependensi

| Kebutuhan | Pilihan | Alasan |
|---|---|---|
| Bahasa | Go 1.22+ | Sesuai ketentuan; routing stdlib sudah kuat |
| HTTP framework | Gin | Binding/validasi, middleware, ekosistem matang |
| Database | PostgreSQL 16 | Constraint kuat, enum check, index ekspresi |
| Driver | `jackc/pgx/v5` + `pgxpool` | Performa dan kontrol SQL |
| Migrasi | `golang-migrate/migrate` (embedded) | Versioned, dapat dijalankan ulang |
| JWT | `golang-jwt/jwt/v5` | Standar HS256 |
| Hash password | `golang.org/x/crypto/bcrypt` | Sederhana, cost dapat diatur |
| Validasi | `go-playground/validator/v10` via Gin binding | Konsisten dengan struct |
| Config | `caarlos0/env/v11` | Parsing env ke struct |
| Logging | `log/slog` (JSON) | Stdlib, terstruktur |
| Testing | `testing`, `httptest`, `testify/require` | Standar, mudah dibaca |

Semua dependensi open source dan boleh digunakan sesuai ketentuan test.

## 3. Arsitektur berlapis

```mermaid
flowchart TD
    R[Router + Middleware] --> H[Handler / HTTP]
    H --> S[Service / Use Case]
    S --> RP[Repository / SQL]
    RP --> DB[(PostgreSQL)]
    S --> AUTH[Auth: bcrypt + JWT]
    H -. DTO .-> S
```

### Aturan dependensi
- `handler` hanya menangani HTTP: parsing, validasi bentuk, mapping error, response.
- `service` memuat aturan bisnis dan otorisasi; **tidak** mengetahui `*gin.Context`.
- `repository` hanya SQL; tidak memuat aturan bisnis.
- `domain` berisi entity, enum, error sentinel, dan interface repository.
- Arah dependensi satu arah: `http → service → repository → domain`. Tidak ada import balik.

### Penanganan error
Gunakan error sentinel di `domain/errors.go`, lalu mapping terpusat:

| Error domain | HTTP | Kode |
|---|---|---|
| `ErrValidation` | 400 | `VALIDATION_ERROR` |
| `ErrUnauthenticated` | 401 | `UNAUTHENTICATED` |
| `ErrInvalidCredentials` | 401 | `INVALID_CREDENTIALS` |
| `ErrForbidden` | 403 | `FORBIDDEN` |
| `ErrInternalNoteForbidden` | 403 | `INTERNAL_NOTE_FORBIDDEN` |
| `ErrNotFound` | 404 | `NOT_FOUND` |
| `ErrDuplicateEmail` | 409 | `DUPLICATE_EMAIL` |
| `ErrInvalidTransition` | 409 | `INVALID_TRANSITION` |
| `ErrSelfRoleChange` | 409 | `SELF_ROLE_CHANGE` |
| `ErrLastAdmin` | 409 | `LAST_ADMIN` |
| lainnya | 500 | `INTERNAL_ERROR` |

Error 500 wajib di-log lengkap (dengan request ID), tetapi response hanya berisi pesan generik.

## 4. Struktur folder

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # wiring config, db, repo, service, router
├── internal/
│   ├── config/
│   │   └── config.go            # env parsing + validasi
│   ├── domain/
│   │   ├── user.go              # entity + enum role
│   │   ├── ticket.go            # entity + enum status/priority/category
│   │   ├── comment.go
│   │   ├── errors.go            # sentinel errors
│   │   └── repository.go        # interface repository
│   ├── auth/
│   │   ├── password.go          # bcrypt hash/verify
│   │   └── jwt.go               # sign/parse claims
│   ├── repository/
│   │   ├── postgres/
│   │   │   ├── user_repo.go
│   │   │   ├── ticket_repo.go
│   │   │   └── comment_repo.go
│   │   └── ...
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── ticket_service.go    # transisi + otorisasi
│   │   ├── comment_service.go
│   │   └── user_service.go
│   ├── http/
│   │   ├── router.go
│   │   ├── handler/
│   │   │   ├── auth_handler.go
│   │   │   ├── ticket_handler.go
│   │   │   ├── comment_handler.go
│   │   │   ├── user_handler.go
│   │   │   └── health_handler.go
│   │   ├── middleware/
│   │   │   ├── auth.go          # verifikasi JWT
│   │   │   ├── role.go          # requireRole
│   │   │   ├── request_id.go
│   │   │   ├── cors.go
│   │   │   └── recovery.go
│   │   └── response/
│   │       └── response.go      # envelope sukses + error
│   └── migrate/
│       └── migrate.go           # embed migrations + seed
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── seed/
│   └── seed.sql
├── tests/
│   ├── auth_test.go
│   ├── ticket_rbac_test.go
│   ├── transition_test.go
│   ├── comment_test.go
│   └── helpers_test.go
├── Dockerfile
├── go.mod
├── go.sum
└── .env.example
```

## 5. Domain model dan invarian

### 5.1 Entity User
```go
type Role string
const (
    RoleRequester Role = "requester"
    RoleAgent     Role = "agent"
    RoleAdmin     Role = "admin"
)

type User struct {
    ID           int64
    Name         string
    Email        string // disimpan lowercase
    PasswordHash string
    Role         Role
    CreatedAt    time.Time
}
```

### 5.2 Entity Ticket
```go
type Status string
const (
    StatusOpen       Status = "open"
    StatusInProgress Status = "in_progress"
    StatusResolved   Status = "resolved"
    StatusClosed     Status = "closed"
)

type Priority string
const (
    PriorityLow    Priority = "low"
    PriorityMedium Priority = "medium"
    PriorityHigh   Priority = "high"
    PriorityUrgent Priority = "urgent"
)

type Category string
const (
    CategoryHardware Category = "hardware"
    CategorySoftware Category = "software"
    CategoryNetwork  Category = "network"
    CategoryAccess   Category = "access"
    CategoryOther    Category = "other"
)

type Ticket struct {
    ID          int64
    Title       string
    Description string
    Category    Category
    Priority    Priority
    Status      Status
    RequesterID int64
    AssigneeID  *int64
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 5.3 Invarian
1. Tiket baru: `status = open`, `assignee_id = null`, `requester_id = aktor`.
2. `requester_id` tidak dapat diubah lewat API.
3. `updated_at` diperbarui setiap PATCH atau penambahan komentar (opsional untuk komentar).
4. Transisi status hanya boleh melalui fungsi `TicketService.ChangeStatus` (satu sumber kebenaran).
5. Assignee harus user dengan peran `agent` atau `admin`.
6. Field enum disimpan sebagai `VARCHAR + CHECK`, bukan tipe enum PostgreSQL, agar migrasi mudah.
7. Tidak ada hard delete.

## 6. Functional requirements

### 6.1 FR-BE-01 — Register
- `POST /api/auth/register`, publik.
- Input: `name`, `email`, `password`.
- Email di-trim dan di-lowercase. Unik secara case-insensitive.
- Password di-hash bcrypt dengan cost dari env (`BCRYPT_COST`, default 12).
- Peran selalu `requester`; jika klien mengirim `role`, nilai diabaikan.
- Response `201` berisi user tanpa `password_hash`.
- Email duplikat → `409 DUPLICATE_EMAIL`.
- Validasi gagal → `400 VALIDATION_ERROR` dengan `details` per field.

### 6.2 FR-BE-02 — Login
- `POST /api/auth/login`, publik.
- Input: `email`, `password`.
- Verifikasi bcrypt. Email tidak ditemukan dan password salah menghasilkan response **identik**
  (`401 INVALID_CREDENTIALS`, pesan "Email atau password salah") untuk mencegah enumerasi akun.
- Response `200`:
  ```json
  {
    "data": {
      "token": "<jwt>",
      "token_type": "Bearer",
      "expires_at": "2026-11-09T10:15:30Z",
      "user": { "id": 2, "name": "...", "email": "...", "role": "agent", "created_at": "..." }
    }
  }
  ```
- JWT claims: `sub` (string user id), `role`, `iat`, `exp`. Algoritma HS256. TTL dari `JWT_TTL`
  (default `24h`).
- Tidak ada refresh token pada versi ini.

### 6.3 FR-BE-03 — Middleware autentikasi dan peran
- Middleware `Auth` membaca `Authorization: Bearer <token>`.
- Token tidak ada / format salah / signature salah / kedaluwarsa → `401 UNAUTHENTICATED`.
- Setelah valid, middleware memuat `user_id` dan `role` ke context request.
- Middleware `RequireRole(roles...)` menolak peran yang tidak cocok → `403 FORBIDDEN`.
- Middleware memuat ulang user dari DB berdasarkan `sub` pada setiap request terautentikasi.
  Tujuannya agar perubahan peran oleh Admin langsung berlaku dan user yang dihapus/dinonaktifkan
  tidak lagi memiliki akses. Klaim `role` pada JWT hanya dipakai sebagai pemeriksaan awal;
  **otoritas final adalah nilai role terbaru di DB**. Biaya satu query ringan diterima.
- Middleware `Recovery` mengubah panic menjadi `500 INTERNAL_ERROR` tanpa membocorkan stack trace.

### 6.4 FR-BE-04 — Tiket (create & read)
- `POST /api/tickets` untuk semua peran login.
- Field yang diterima: `title`, `description`, `category`, `priority`.
- Field `status`, `requester_id`, `assignee_id`, `created_at` dari klien **diabaikan**.
- Response `201` berisi tiket lengkap dengan objek `requester` dan `assignee` (nullable).
- `GET /api/tickets/{id}`:
  - Requester hanya boleh jika `ticket.requester_id == actor.id`; selain itu `403 FORBIDDEN`.
  - Agent/Admin boleh semua tiket.
  - Response menyertakan `comments` yang sudah difilter:
    - Requester: hanya `is_internal = false`.
    - Agent/Admin: semua komentar.
  - Komentar diurutkan `created_at ASC`, lalu `id ASC`.
  - Tiket tidak ada → `404 NOT_FOUND`.

### 6.5 FR-BE-05 — Daftar tiket
- `GET /api/tickets` untuk semua peran login.
- **Scoping wajib**: bila aktor `requester`, repository menambahkan `requester_id = actor.id`
  dan mengabaikan parameter `requester_id`/`assignee_id`/`unassigned` dari klien.
- Filter: `status`, `priority`, `assignee_id`, `unassigned`, `search`.
- `search` mencocokkan `title ILIKE '%q%'`; trim input, maksimum 100 karakter.
- Sort: `created_at` (default), `updated_at`, `priority`, `status`.
  - `priority` memakai urutan `urgent=4, high=3, medium=2, low=1`.
  - `status` memakai urutan `open=1, in_progress=2, resolved=3, closed=4`.
  - `order=asc|desc`, default `desc`.
- Pagination: `page` default 1, `per_page` default 20, maksimum 100.
- Response menyertakan `meta` lengkap. Jika `page` melebihi total halaman, kembalikan `data: []`
  dengan meta yang benar (bukan error).
- Query invalid (enum salah, `page=0`, `per_page=1000`) → `400 VALIDATION_ERROR`.
- Wajib menggunakan parameterized query (tidak ada string interpolation nilai user).

### 6.6 FR-BE-06 — Komentar dan catatan internal
- `POST /api/tickets/{id}/comments`.
- Akses: pemilik tiket, Agent, Admin. Selain itu `403 FORBIDDEN`.
- Body: `body` (1–5000 karakter), `is_internal` (bool, default `false`).
- Requester mengirim `is_internal = true` → `403 INTERNAL_NOTE_FORBIDDEN`.
- Komentar publik dan internal sama-sama ditampilkan ke Agent/Admin; hanya publik yang terlihat
  oleh Requester.
- Response `201` berisi komentar dengan `author` (id, name, role).
- Komentar diizinkan pada semua status tiket.
- Komentar tidak dapat diedit/dihapus pada versi ini.

### 6.7 FR-BE-07 — Transisi status (aturan inti)
`PATCH /api/tickets/{id}` menerima kombinasi `status`, `priority`, `assignee_id` (minimal satu field).

#### Matriks otorisasi transisi

| Dari | Ke | Peran diizinkan | Syarat |
|---|---|---|---|
| `open` | `in_progress` | Agent, Admin | — |
| `in_progress` | `resolved` | Agent, Admin | — |
| `resolved` | `closed` | Requester, Agent, Admin | Requester harus pemilik |
| `resolved` | `open` | Requester | Harus pemilik |
| `open` | `closed` | Requester, Admin | Requester harus pemilik |

#### Algoritma validasi (wajib)
```
1. Muat tiket; tidak ada -> 404.
2. Tentukan apakah aktor boleh mengakses tiket (pemilik/agent/admin); jika tidak -> 403.
3. Jika status berubah:
   a. Cari aturan (from=status_lama, to=status_baru).
   b. Jika tidak ada aturan -> 409 INVALID_TRANSITION.
   c. Jika peran aktor tidak ada di aturan -> 409 INVALID_TRANSITION.
   d. Jika aturan mensyaratkan pemilik dan aktor bukan pemilik -> 403 FORBIDDEN.
4. Jika priority berubah: hanya Agent/Admin; selain itu 403.
5. Jika assignee_id berubah:
   - Admin: boleh null atau user berperan agent/admin yang ada.
   - Agent: hanya boleh mengisi dirinya sendiri, atau null jika assignee saat ini adalah dirinya.
   - Requester: 403.
6. Terapkan semua perubahan dalam satu transaksi DB.
7. Jika status menjadi in_progress dan assignee masih null dan aktor Agent:
   assignee_id = aktor.id.
8. Set updated_at = now(); kembalikan tiket terbaru.
```

#### Pesan error
- Transisi tidak ada / peran salah:
  `"Transisi dari 'open' ke 'resolved' tidak diizinkan untuk peran 'agent'."`
- Bukan pemilik:
  `"Anda tidak memiliki akses untuk mengubah tiket ini."`
- Assignee tidak valid:
  `"Assignee harus pengguna dengan peran agent atau admin."`

#### Contoh request
```json
{ "status": "in_progress" }
```
```json
{ "status": "resolved", "priority": "high", "assignee_id": 2 }
```
```json
{ "assignee_id": null }
```

#### Idempotensi
Mengirim `status` yang sama dengan status saat ini bukan transisi; diperlakukan sebagai no-op
yang valid (tidak error), selama aktor berhak mengubah tiket. Field lain tetap divalidasi.
Alasan: mencegah error palsu saat UI mengirim state yang sama.

### 6.8 FR-BE-08 — Manajemen pengguna (Admin)
- `GET /api/users`:
  - Hanya Admin (`403` untuk peran lain).
  - Query: `page`, `per_page` (default 20, max 100), `role`, `search` (nama/email).
  - Response: daftar user tanpa `password_hash` + meta.
- `PATCH /api/users/{id}/role`:
  - Hanya Admin.
  - Body `{ "role": "agent" }`, enum valid.
  - Admin tidak boleh mengubah perannya sendiri → `409 SELF_ROLE_CHANGE`.
  - Demosi Admin terakhir → `409 LAST_ADMIN`.
  - User tidak ada → `404 NOT_FOUND`.
  - Response berisi user terbaru.
- Perubahan peran berlaku pada request berikutnya karena klaim JWT lama masih membawa peran lama.
  **Keputusan**: middleware memuat peran terbaru dari DB untuk setiap request terautentikasi agar
  perubahan peran langsung berlaku. Ini mengorbankan satu query ringan demi keamanan.
  (Alternatif: token version/refresh; tidak dipakai agar sederhana.)

### 6.9 FR-BE-09 — Migrasi dan seed
- Migrasi versioned dengan `golang-migrate`, file SQL embedded di binary.
- Dijalankan otomatis saat startup bila `AUTO_MIGRATE=true` (default `true` di Docker).
- Harus aman dijalankan ulang (versioning + `IF NOT EXISTS`).
- Seed dijalankan bila `AUTO_SEED=true` (default `true` di Docker):
  - Satu `requester`, satu `agent`, satu `admin`.
  - Password dibaca dari env (`SEED_*_PASSWORD`), di-hash saat seed.
  - Idempotent: `INSERT ... ON CONFLICT (email) DO NOTHING`.
  - Seed tidak menimpa password/role user yang sudah ada.

### 6.10 FR-BE-10 — Health check
- `GET /api/health` publik.
- Response `200 { "status": "ok", "db": "up" }` bila DB dapat di-ping; `503` bila tidak.
- Dipakai sebagai Docker healthcheck.

## 7. API contract detail

Semua response mengikuti envelope pada PRD-00 §10.

### 7.1 POST `/api/auth/register`

Request:
```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "password": "Rahasia123"
}
```
Response `201`:
```json
{
  "data": {
    "id": 7,
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "role": "requester",
    "created_at": "2026-11-08T10:15:30Z"
  }
}
```
Error `400`:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Data yang dikirim tidak valid.",
    "details": {
      "email": "Format email tidak valid.",
      "password": "Password minimal 8 karakter."
    }
  }
}
```

### 7.2 POST `/api/auth/login`

Request:
```json
{ "email": "agent@helpdesk.local", "password": "Agent123!" }
```
Response `200`:
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_at": "2026-11-09T10:15:30Z",
    "user": {
      "id": 2,
      "name": "Agent Satu",
      "email": "agent@helpdesk.local",
      "role": "agent",
      "created_at": "2026-11-01T08:00:00Z"
    }
  }
}
```

### 7.3 GET `/api/me`
Header: `Authorization: Bearer <token>`.
Response `200`:
```json
{
  "data": {
    "id": 2,
    "name": "Agent Satu",
    "email": "agent@helpdesk.local",
    "role": "agent",
    "created_at": "2026-11-01T08:00:00Z"
  }
}
```

### 7.4 GET `/api/tickets`

Contoh: `GET /api/tickets?status=open&priority=high&search=laptop&sort=created_at&order=desc&page=1&per_page=10`

Response `200`:
```json
{
  "data": [
    {
      "id": 42,
      "title": "Laptop tidak menyala",
      "description": "Laptop mati total setelah update.",
      "category": "hardware",
      "priority": "high",
      "status": "open",
      "requester": { "id": 3, "name": "Budi Santoso", "email": "budi@example.com" },
      "assignee": null,
      "created_at": "2026-11-08T09:00:00Z",
      "updated_at": "2026-11-08T09:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 1, "total_pages": 1 }
}
```

### 7.5 POST `/api/tickets`

Request:
```json
{
  "title": "Laptop tidak menyala",
  "description": "Laptop mati total setelah update sistem.",
  "category": "hardware",
  "priority": "high"
}
```
Response `201`: objek tiket yang sama seperti di atas (tanpa `comments`).

### 7.6 GET `/api/tickets/{id}`

Response `200` menambahkan `comments`:
```json
{
  "data": {
    "id": 42,
    "title": "Laptop tidak menyala",
    "description": "Laptop mati total setelah update sistem.",
    "category": "hardware",
    "priority": "high",
    "status": "in_progress",
    "requester": { "id": 3, "name": "Budi Santoso", "email": "budi@example.com" },
    "assignee": { "id": 2, "name": "Agent Satu", "email": "agent@helpdesk.local" },
    "created_at": "2026-11-08T09:00:00Z",
    "updated_at": "2026-11-08T09:30:00Z",
    "comments": [
      {
        "id": 10,
        "ticket_id": 42,
        "author": { "id": 2, "name": "Agent Satu", "role": "agent" },
        "body": "Sedang dicek oleh tim hardware.",
        "is_internal": false,
        "created_at": "2026-11-08T09:31:00Z"
      }
    ]
  }
}
```

### 7.7 PATCH `/api/tickets/{id}`

Request (salah satu atau kombinasi):
```json
{ "status": "in_progress", "priority": "urgent", "assignee_id": 2 }
```
Response `200`: tiket terbaru. Error:
```json
{
  "error": {
    "code": "INVALID_TRANSITION",
    "message": "Transisi dari 'open' ke 'resolved' tidak diizinkan untuk peran 'agent'.",
    "details": { "from": "open", "to": "resolved", "role": "agent" }
  }
}
```

### 7.8 POST `/api/tickets/{id}/comments`

Request:
```json
{ "body": "Mohon dicek juga adaptor charger-nya.", "is_internal": false }
```
Response `201`:
```json
{
  "data": {
    "id": 11,
    "ticket_id": 42,
    "author": { "id": 3, "name": "Budi Santoso", "role": "requester" },
    "body": "Mohon dicek juga adaptor charger-nya.",
    "is_internal": false,
    "created_at": "2026-11-08T09:40:00Z"
  }
}
```

### 7.9 GET `/api/users`

`GET /api/users?role=agent&search=budi&page=1&per_page=20`

Response `200`:
```json
{
  "data": [
    {
      "id": 2,
      "name": "Agent Satu",
      "email": "agent@helpdesk.local",
      "role": "agent",
      "created_at": "2026-11-01T08:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 1, "total_pages": 1 }
}
```

### 7.10 PATCH `/api/users/{id}/role`

Request: `{ "role": "admin" }`
Response `200`: user terbaru.

## 8. Aturan validasi

| Field | Aturan | Pesan contoh |
|---|---|---|
| `name` | wajib, 2–100 karakter, trim | "Nama minimal 2 karakter." |
| `email` | wajib, format email, maks 255, lowercase | "Format email tidak valid." |
| `password` | wajib, 8–72 byte, mengandung huruf dan angka | "Password minimal 8 karakter dan harus mengandung huruf dan angka." |
| `title` | wajib, 3–150 karakter, trim | "Judul minimal 3 karakter." |
| `description` | wajib, 10–5000 karakter | "Deskripsi minimal 10 karakter." |
| `category` | wajib, enum | "Kategori tidak valid." |
| `priority` | wajib saat create; enum saat patch | "Prioritas tidak valid." |
| `status` | enum saat patch | "Status tidak valid." |
| `assignee_id` | null atau integer > 0; user harus ada & berperan agent/admin | "Assignee tidak ditemukan." |
| `body` komentar | wajib, 1–5000 karakter setelah trim | "Komentar tidak boleh kosong." |
| `is_internal` | boolean, default false | — |
| `role` | enum | "Peran tidak valid." |
| `page` | integer ≥ 1 | "Parameter page tidak valid." |
| `per_page` | integer 1–100 | "per_page maksimal 100." |
| `sort` / `order` | enum | "Parameter sort tidak valid." |

Catatan: semua string di-trim sebelum validasi. Validasi dilakukan di handler (bentuk) dan service
(aturan bisnis) agar dapat diuji tanpa HTTP.

## 9. Skema database

### 9.1 DDL migrasi `000001_init.up.sql`

```sql
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'requester'
                  CHECK (role IN ('requester', 'agent', 'admin')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uq
    ON users (lower(email));

CREATE TABLE IF NOT EXISTS tickets (
    id           BIGSERIAL PRIMARY KEY,
    title        VARCHAR(150) NOT NULL,
    description  TEXT         NOT NULL,
    category     VARCHAR(20)  NOT NULL
                 CHECK (category IN ('hardware', 'software', 'network', 'access', 'other')),
    priority     VARCHAR(20)  NOT NULL DEFAULT 'medium'
                 CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    status       VARCHAR(20)  NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    requester_id BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assignee_id  BIGINT       REFERENCES users(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tickets_requester_idx ON tickets (requester_id);
CREATE INDEX IF NOT EXISTS tickets_assignee_idx  ON tickets (assignee_id);
CREATE INDEX IF NOT EXISTS tickets_status_idx    ON tickets (status);
CREATE INDEX IF NOT EXISTS tickets_priority_idx  ON tickets (priority);
CREATE INDEX IF NOT EXISTS tickets_created_idx   ON tickets (created_at DESC);
CREATE INDEX IF NOT EXISTS tickets_updated_idx   ON tickets (updated_at DESC);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS tickets_title_trgm_idx
    ON tickets USING gin (title gin_trgm_ops);

CREATE TABLE IF NOT EXISTS comments (
    id          BIGSERIAL PRIMARY KEY,
    ticket_id   BIGINT      NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id   BIGINT      NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    body        TEXT        NOT NULL,
    is_internal BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS comments_ticket_idx ON comments (ticket_id, created_at);
```

`000001_init.down.sql` menghapus tabel dalam urutan dependensi:
```sql
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS users;
```

### 9.2 Query wajib

| Operasi | Pendekatan |
|---|---|
| Scoping requester | `WHERE requester_id = $1` selalu ditambahkan di repository |
| Filter dinamis | Bangun `WHERE` + `args` secara terkontrol; hanya placeholder `$n` |
| Sort prioritas | `ORDER BY CASE priority WHEN 'urgent' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END` |
| Sort status | `ORDER BY CASE status WHEN 'open' THEN 1 WHEN 'in_progress' THEN 2 WHEN 'resolved' THEN 3 ELSE 4 END` |
| Search | `title ILIKE '%' || $n || '%'` |
| Count total | query terpisah dengan filter sama, dalam transaksi read-only atau CTE |
| Ambil detail + komentar | dua query (tiket, komentar) atau `json_agg`; dua query lebih mudah diuji |

### 9.3 Seed

```sql
-- Password hash diisi oleh aplikasi (bcrypt), bukan plaintext.
-- Index unik pada lower(email) memungkinkan inferensi konflik ekspresi.
INSERT INTO users (name, email, password_hash, role)
VALUES
  ($1, $2, $3, 'admin'),
  ($4, $5, $6, 'agent'),
  ($7, $8, $9, 'requester')
ON CONFLICT (lower(email)) DO NOTHING;
```
Perilaku yang diwajibkan: seed idempotent dan tidak menimpa password/role user yang sudah ada.
Jika versi PostgreSQL target tidak mendukung inferensi konflik ekspresi, gunakan
`ON CONFLICT DO NOTHING` tanpa target atau cek keberadaan email di dalam transaksi.

## 10. Konfigurasi environment

| Variabel | Wajib | Default | Keterangan |
|---|---|---|---|
| `APP_ENV` | tidak | `development` | `development`/`production` |
| `APP_PORT` | tidak | `8080` | Port HTTP |
| `DATABASE_URL` | ya | — | `postgres://user:pass@host:5432/db?sslmode=disable` |
| `DB_MAX_CONNS` | tidak | `10` | Ukuran pool |
| `JWT_SECRET` | ya | — | Minimal 32 karakter; gagal start bila kosong di production |
| `JWT_TTL` | tidak | `24h` | Durasi token |
| `BCRYPT_COST` | tidak | `12` | 10–14 |
| `CORS_ORIGINS` | tidak | `http://localhost:5173` | Pisahkan koma |
| `AUTO_MIGRATE` | tidak | `true` | Jalankan migrasi saat start |
| `AUTO_SEED` | tidak | `true` | Jalankan seed saat start |
| `SEED_ADMIN_NAME` | tidak | `Admin Helpdesk` | — |
| `SEED_ADMIN_EMAIL` | tidak | `admin@helpdesk.local` | — |
| `SEED_ADMIN_PASSWORD` | ya bila seed aktif | — | Tanpa default di source |
| `SEED_AGENT_*` | ya bila seed aktif | — | Sama seperti admin |
| `SEED_REQUESTER_*` | ya bila seed aktif | — | Sama seperti admin |
| `LOG_LEVEL` | tidak | `info` | `debug`/`info`/`warn`/`error` |
| `MAX_BODY_BYTES` | tidak | `1048576` | Batas body request |

Aturan:
- Semua konfigurasi dibaca sekali di startup ke struct `Config`; tidak ada `os.Getenv` tersebar.
- Tidak ada kredensial nyata di source code. `.env` masuk `.gitignore`; `.env.example` hanya
  berisi placeholder.
- Untuk demo lokal, `docker-compose.yml` boleh menyediakan nilai default **dev-only** untuk seed
  dan DB, dengan peringatan di README bahwa nilai tersebut wajib diganti.

## 11. Keamanan

| Area | Kontrol |
|---|---|
| Password | bcrypt cost ≥ 12; tidak pernah di-log; maksimum 72 byte |
| JWT | HS256; secret ≥ 32 byte dari env; `exp` wajib; tolak alg selain HS256 |
| Otorisasi | Dicek di service, bukan hanya route; kepemilikan tiket diverifikasi setiap akses |
| SQL injection | Parameterized query; tidak ada concatenation nilai user |
| Mass assignment | DTO eksplisit; `role`, `requester_id`, `status` awal diabaikan dari input |
| Enumerasi akun | Pesan login seragam; respons register duplikat tetap jelas (kebutuhan UX) |
| Kebocoran data | `password_hash` tidak pernah di-serialize; catatan internal difilter di service |
| Body limit | 1 MB; tolak payload lebih besar dengan `413` |
| Rate limit | Bonus; bila diterapkan: 10 login/menit per IP |
| Header | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy` |
| CORS | Hanya origin dari `CORS_ORIGINS`; kredensial tidak diperlukan karena Bearer |
| Log | Jangan log token, password, atau body komentar penuh |

## 12. Strategi pengujian

### 12.1 Prinsip
- Test harus menguji **aturan bisnis**, bukan hanya status code.
- Layer service diuji dengan fake repository (cepat, tanpa DB).
- Integration test memakai PostgreSQL nyata (testcontainers atau service `db` pada compose)
  dan `httptest` untuk endpoint kritis.
- Setiap test berdiri sendiri: buat data unik, bersihkan setelahnya.

### 12.2 Daftar test wajib (minimal 5; target 10)

| ID | Jenis | Skenario | Ekspektasi |
|---|---|---|---|
| BE-T01 | Integration | Register dengan password valid | `201`, hash tersimpan (bukan plaintext), role `requester` |
| BE-T02 | Integration | Register email duplikat (beda kapitalisasi) | `409 DUPLICATE_EMAIL` |
| BE-T03 | Integration | Login benar vs. salah | `200` + JWT valid vs. `401 INVALID_CREDENTIALS` dengan pesan identik |
| BE-T04 | Integration | Akses tanpa token / token rusak / token kedaluwarsa | `401 UNAUTHENTICATED` |
| BE-T05 | Integration | Requester A membuka tiket Requester B | `403 FORBIDDEN`; list A tidak memuat tiket B |
| BE-T06 | Service (table-driven) | Semua kombinasi `(from, to, role)` pada matriks | Hanya 5 aturan yang lolos; sisanya `ErrInvalidTransition` |
| BE-T07 | Integration | Agent mengubah `open → resolved` | `409 INVALID_TRANSITION` + pesan jelas |
| BE-T08 | Integration | Requester melihat detail tiketnya yang memiliki catatan internal | Komentar internal tidak ada di response |
| BE-T09 | Integration | Requester mengirim `is_internal=true` | `403 INTERNAL_NOTE_FORBIDDEN` |
| BE-T10 | Integration | Agent mengambil tiket `open` | Status `in_progress`, assignee = agent |
| BE-T11 | Integration | Agent mencoba assign ke agent lain | `403 FORBIDDEN` |
| BE-T12 | Integration | List: filter status+priority, search, sort priority, pagination | Data dan `meta` sesuai |
| BE-T13 | Integration | Admin mengubah peran; admin mengubah perannya sendiri | `200` vs. `409 SELF_ROLE_CHANGE` |
| BE-T14 | Unit | Validasi create ticket mengabaikan `requester_id`/`status` dari klien | Nilai dari aktor/default dipakai |

### 12.3 Perintah
```bash
cd backend
go test ./... -race -count=1
go test ./... -cover
```
Target coverage: **≥ 70%** pada package `service` dan `http/handler`.

### 12.4 CI (bonus)
Workflow: `go vet`, `golangci-lint run`, `go test -race ./...`, build Docker. Tidak wajib.

## 13. Docker

### 13.1 Dockerfile backend (multi-stage)
- Stage build: `golang:1.22-alpine`, `CGO_ENABLED=0`, `go build -trimpath -ldflags="-s -w"`.
- Stage runtime: `gcr.io/distroless/static-debian12:nonroot` atau `alpine` dengan user non-root.
- Expose `8080`.
- `HEALTHCHECK` memanggil `/api/health` (atau healthcheck didefinisikan di compose).

### 13.2 docker-compose.yml (root repo)

| Service | Image/Build | Port | Depends on | Healthcheck |
|---|---|---|---|---|
| `db` | `postgres:16-alpine` | internal 5432 (opsional host 5433) | — | `pg_isready -U $POSTGRES_USER` |
| `backend` | build `./backend` | internal 8080 | `db` healthy | `GET /api/health` |
| `frontend` | build `./frontend` | host `8080:80` | `backend` started | `GET /` |

Ketentuan:
- Volume bernama `pgdata` untuk persistensi.
- Env backend diambil dari `.env` dengan fallback dev-only.
- `restart: unless-stopped`.
- Satu perintah: `docker compose up --build`.
- Backend menunggu DB siap (retry ping dengan backoff, bukan hanya `depends_on`).

## 14. Observability

- Log JSON: `ts`, `level`, `msg`, `request_id`, `method`, `path`, `status`, `latency_ms`, `user_id`.
- Request ID dibuat middleware bila tidak ada, dikembalikan di header.
- Health endpoint terpisah dari log akses agar tidak berisik (opsional).
- Metrik Prometheus tidak termasuk scope.

## 15. Acceptance criteria backend

- [ ] `POST /api/auth/register` membuat user `requester`; password ter-hash bcrypt.
- [ ] `POST /api/auth/login` mengembalikan JWT yang dapat dipakai di endpoint terproteksi.
- [ ] Token hilang/rusak/kedaluwarsa → `401`; peran salah → `403`.
- [ ] Requester hanya melihat tiket miliknya pada list dan detail.
- [ ] Catatan internal tidak pernah muncul pada response untuk requester.
- [ ] Kelima transisi valid berhasil; semua transisi lain → `409 INVALID_TRANSITION`.
- [ ] Requester tidak dapat mengubah prioritas/assignee.
- [ ] Agent hanya dapat assign ke dirinya sendiri; Admin ke agent/admin mana pun.
- [ ] List mendukung `status`, `priority`, `assignee_id`, `unassigned`, `search`, `sort`, `order`,
      `page`, `per_page` dengan hasil dan meta benar.
- [ ] Admin dapat melihat user dan mengubah peran; self-change dan last-admin ditolak.
- [ ] Migrasi + seed idempotent dan berjalan otomatis.
- [ ] Semua test wajib lulus dengan `-race`.
- [ ] `docker compose up --build` menyalakan DB, backend, frontend; healthcheck hijau.
- [ ] OpenAPI tersedia dan sesuai implementasi.
- [ ] Tidak ada kredensial nyata di source.

## 16. Bonus (opsional, prioritas menurun)

| Bonus | Dampak | Estimasi |
|---|---|---|
| Lampiran tiket | Validasi MIME + ukuran (maks 5 MB), simpan di volume | M |
| SSE notifikasi | Event `ticket.updated`, `comment.created` | M |
| SLA urgent | Kolom `sla_due_at`, job penanda `is_overdue` | S |
| Statistik admin | Endpoint agregat per status/prioritas | S |
| CI | GitHub Actions lint + test | S |
| Audit log | Tabel `ticket_events` | M |

## 17. Rencana kerja

| Prioritas | Tugas | Output |
|---|---|---|
| P0 | Setup repo, config, koneksi DB, migrasi, seed | Backend start + health |
| P0 | Auth (register/login/me) + JWT middleware | Test BE-T01–T04 |
| P0 | Domain tiket, repository, create/detail | Endpoint dasar |
| P0 | Transisi status + otorisasi | Test BE-T05–T11 |
| P0 | List filter/pagination/sort | Test BE-T12 |
| P0 | Komentar + catatan internal | Test BE-T08–T09 |
| P0 | Admin users | Test BE-T13 |
| P0 | Dockerfile + compose + healthcheck | Satu perintah jalan |
| P1 | OpenAPI + README backend | Dokumentasi |
| P1 | Perbaikan pesan error & validasi | UX error konsisten |
| P2 | Bonus SLA/statistik/CI | Nilai tambah |

## 18. Risiko dan mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Aturan transisi tersebar | Bug & test rapuh | Satu fungsi `canTransition(from,to,role,isOwner)` + table-driven test |
| Query dinamis rawan injeksi | Keamanan | Placeholder `$n`, whitelist kolom sort |
| Perubahan peran tidak langsung berlaku | Keamanan | Muat ulang role dari DB di middleware |
| Seed menimpa data | Kehilangan data | `ON CONFLICT DO NOTHING`, hanya saat `AUTO_SEED=true` |
| Test DB tidak tersedia di CI | Test gagal | testcontainers atau compose profile `test` |
| Waktu habis | Fitur kurang | Prioritaskan P0; dokumentasikan yang belum selesai |

## 19. Catatan arsitektur (ringkas untuk README)

- **Kenapa berlapis**: aturan transisi dan otorisasi berada di service sehingga dapat diuji tanpa
  HTTP/DB; handler tipis dan repository hanya SQL.
- **Kenapa Gin**: binding + middleware mengurangi boilerplate, tetapi aturan bisnis tetap framework
  agnostic.
- **Yang sengaja ditinggalkan**: refresh token, audit log, lampiran, notifikasi real-time.
- **Perbaikan berikutnya**: refresh token + rotasi, audit log, rate limit login, soft delete,
  observabilitas metrik.
