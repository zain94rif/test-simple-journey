# Helpdesk Frontend

Vue 3 + Vite SPA untuk aplikasi Helpdesk (Uji Teknis). Mengimplementasikan PRD-02.

## Teknologi

- **Vue 3** (Composition API, `<script setup>`) + **TypeScript**
- **Vite 6** — dev server & bundler
- **Pinia** — state management (`auth`, `tickets`, `users`, `toast`)
- **Vue Router 4** — routing dengan role guards
- **Axios** — HTTP client dengan interceptors (Bearer token, 401 handling, error normalization)
- **Tailwind CSS 4** (`@tailwindcss/vite`)
- **lucide-vue-next** — ikon
- **Vitest** + **@vue/test-utils** + **happy-dom** — unit test

## Perintah (Bun)

```bash
bun install        # install dependencies
bun run dev        # dev server di http://localhost:5173 (proxy /api → http://localhost:8080)
bun run build      # build produksi ke dist/
bun run preview    # serve hasil build
bun run test:unit  # vitest --run
bun run lint       # eslint .
bun run format     # prettier --write
```

Periksa tipe secara langsung: `bun x tsc --noEmit`.

## Konfigurasi

Lihat `.env.example`. `VITE_API_BASE_URL` default `/api` (proxy Vite → `http://localhost:8080`).

## Struktur Kunci

```
src/
├── lib/          http.ts (axios+token), api.ts, errors.ts, types.ts, constants.ts, format.ts
├── stores/       auth, tickets (list/abort/detail/mutasi), users, toast
├── composables/  useTicketPermissions, useToast, useDebounce, useConfirm
├── components/   ui/, tickets/, users/
├── layouts/      AppLayout (sidebar, topbar, drawer, modal konfirmasi), AuthLayout
├── views/        Login, Register, TicketList, TicketCreate, TicketDetail, AdminUsers, Forbidden, NotFound
└── router/       guards: bootstrap → guestOnly → requiresAuth → roles
```

## Rute

| Rute | Akses |
| --- | --- |
| `/login`, `/register` | khusus tamu |
| `/tickets`, `/tickets/new`, `/tickets/:id` | wajib login |
| `/admin/users` | hanya admin |
| `/forbidden`, 404 `/:pathMatch(.*)*` | publik |

Filter daftar tiket tersinkron ke query URL (`?status=&priority=&assignee=&search=&page=`).
Token disimpan di `localStorage["helpdesk_token"]`; sesi selalu di-hydrate dari `GET /me`.

## Testing

42 unit test (Vitest, happy-dom) meliputi: matriks perizinan tiket (`useTicketPermissions`),
login/bootstrap di store auth, guard router (guestOnly, requiresAuth, roles, 404),
label `StatusBadge`/`PriorityBadge`, `RoleSelect` dinonaktifkan untuk peran sendiri,
`CommentForm` (requester tidak dapat komentar internal), dan perilaku cache daftar
(tiket selalu fetch saat masuk halaman karena datanya dinamis; daftar pengguna
di-cache di Pinia dengan invalidasi setelah edit peran dan logout).

## Akun Seed

| Peran | Email |
| --- | --- |
| Admin | admin@helpdesk.local |
| Agent | agent@helpdesk.local, agent2@helpdesk.local |
| Requester | requester1@…3@helpdesk.local |

Sandi: `Admin123!` / `Agent123!` / `Requester123!` — seed aktif saat backend
dijalankan dengan `SEED_DEMO_DATA=true` (lihat `../backend`).

## Docker

`Dockerfile` (build node:20-alpine → serve nginx:1.27-alpine) dan `nginx.conf`
(SPA fallback + proxy `/api/` ke `backend:8080`).

```bash
podman build -t helpdesk-frontend .
podman run -p 8081:80 --network <jaringan-yang-sama-dengan-backend> helpdesk-frontend
```
