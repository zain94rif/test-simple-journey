# PRD-02 — Frontend (Vue 3)

| Item | Keterangan |
|---|---|
| Produk | Helpdesk Internal — Web App |
| Versi | 1.0 |
| Tanggal | 8 November 2026 |
| Status | Draft untuk implementasi |
| Dokumen induk | [PRD-00-Umum.md](./PRD-00-Umum.md) |
| Backend | [PRD-01-Backend.md](./PRD-01-Backend.md) |
| Stack | Vue 3 + Vite + Pinia + Vue Router |

---

## 1. Ringkasan dan tujuan

Membangun antarmuka web yang bersih, cepat, dan aman untuk tiga peran helpdesk. Frontend
mengonsumsi REST API pada PRD-01 dan **tidak menegakkan aturan bisnis sebagai pengganti backend**;
UI hanya menyembunyikan aksi yang tidak relevan dan menampilkan error server dengan jelas.

### Tujuan
1. Requester dapat membuat, memantau, dan menutup tiket tanpa pelatihan.
2. Agent dapat mengambil dan menyelesaikan tiket dengan sedikit klik.
3. Admin dapat mengelola peran pengguna dengan aman.
4. Semua halaman memiliki state loading, kosong, dan error yang jelas.
5. Nyaman dipakai di laptop dan ponsel.

### Non-goals
- Tidak ada mode offline/PWA.
- Tidak ada tema gelap pada versi wajib (bonus).
- Tidak ada upload lampiran (bonus backend).
- Tidak ada notifikasi real-time (bonus).

## 2. Stack dan dependensi

| Kebutuhan | Pilihan | Alasan |
|---|---|---|
| Framework | Vue 3 Composition API + `<script setup>` | Sesuai ketentuan |
| Build tool | Vite 5+ | Cepat, konfigurasi sederhana |
| Bahasa | TypeScript (disarankan) atau JavaScript | Tipe membuat kontrak API lebih aman |
| Routing | Vue Router 4 | Guard & lazy loading |
| State | Pinia | Standar Vue 3, mudah diuji |
| HTTP | Axios | Interceptor request/response |
| Styling | Tailwind CSS 3/4 | Konsisten, responsif, tanpa UI kit berat |
| Ikon | `lucide-vue-next` | Ringan dan konsisten |
| Notifikasi | Toast kustom (`useToast`) | Menghindari dependensi berlebih |
| Testing | Vitest + Vue Test Utils + MSW | Unit/komponen dengan API mock |
| Lint/format | ESLint + Prettier | Kualitas kode |

> Jika memakai JavaScript, kontrak API tetap didokumentasikan dengan JSDoc. TypeScript lebih
> diutamakan untuk nilai tambah, tetapi bukan syarat kelulusan.

## 3. Prinsip desain dan UX

1. **Satu tujuan per layar** — daftar untuk menemukan, detail untuk bertindak.
2. **Aksi destruktif diberi konfirmasi** — tutup, batalkan, buka kembali.
3. **Status terlihat sekilas** — badge warna + teks, tidak hanya warna.
4. **Error dapat ditindaklanjuti** — pesan jelas, tombol retry, tidak ada layar kosong misterius.
5. **Mobile-first** — tabel berubah menjadi kartu di layar kecil.
6. **Konsisten** — satu komponen badge, satu gaya tombol, satu pola form.
7. **Aman** — tidak pernah merender HTML dari input pengguna (`v-html` dilarang untuk konten user).

### 3.1 Design token

| Token | Nilai |
|---|---|
| Primary | Indigo 600 (`#4f46e5`), hover 700 |
| Background | Slate 50 |
| Surface | White, border Slate 200 |
| Text utama | Slate 900; sekunder Slate 500 |
| Status `open` | Biru (bg-blue-100/text-blue-700) |
| Status `in_progress` | Amber (bg-amber-100/text-amber-700) |
| Status `resolved` | Hijau (bg-emerald-100/text-emerald-700) |
| Status `closed` | Abu (bg-slate-100/text-slate-600) |
| Prioritas `low` | Abu; `medium` biru; `high` oranye; `urgent` merah |
| Radius | `rounded-lg` (8px), kartu `rounded-xl` (12px) |
| Font | Inter / system-ui |
| Spacing | Kelipatan 4px |

## 4. Information architecture dan peta rute

| Path | Nama | Halaman | Guard | Layout |
|---|---|---|---|---|
| `/login` | `login` | Login | `guestOnly` | AuthLayout |
| `/register` | `register` | Register | `guestOnly` | AuthLayout |
| `/` | `home` | Redirect ke `/tickets` | — | — |
| `/tickets` | `tickets` | Daftar tiket | `requiresAuth` | AppLayout |
| `/tickets/new` | `ticket-create` | Buat tiket | `requiresAuth` | AppLayout |
| `/tickets/:id` | `ticket-detail` | Detail tiket | `requiresAuth` | AppLayout |
| `/admin/users` | `admin-users` | Kelola pengguna | `requiresAuth`, `roles: ['admin']` | AppLayout |
| `/forbidden` | `forbidden` | 403 | — | AppLayout / minimal |
| `/:pathMatch(.*)*` | `not-found` | 404 | — | Minimal |

Semua halaman AppLayout memakai navigasi yang sama:
- Desktop: sidebar kiri (Daftar Tiket, Buat Tiket, Kelola Pengguna [admin]) + topbar profil.
- Mobile: topbar dengan tombol menu → drawer; tombol utama tetap terlihat.
- Menampilkan nama, email, dan badge peran pengguna; menu logout.

### 4.1 Judul dokumen
Setiap route menetapkan `meta.title`; guard afterEach mengubah `document.title` menjadi
`<judul> — Helpdesk`.

## 5. Routing dan navigation guard

### 5.1 Meta route
```ts
{
  path: '/admin/users',
  name: 'admin-users',
  component: () => import('@/views/AdminUsersView.vue'),
  meta: { requiresAuth: true, roles: ['admin'], title: 'Kelola Pengguna' }
}
```

### 5.2 Perilaku guard

| Kondisi | Aksi |
|---|---|
| Belum login, membuka route `requiresAuth` | Redirect `/login?redirect=<path>` |
| Sudah login, membuka `guestOnly` (`/login`, `/register`) | Redirect `/tickets` |
| Login, peran tidak ada di `meta.roles` | Redirect `/forbidden` |
| Token ada tetapi `/api/me` gagal (401) | Hapus sesi, redirect `/login` |
| Route tidak dikenal | Halaman 404 |
| Sudah login, membuka `/` | Redirect `/tickets` |

### 5.3 Bootstrap sesi
```ts
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.initialized) await auth.bootstrap() // fetchMe bila token ada

  if (to.meta.guestOnly && auth.isAuthenticated) return { name: 'tickets' }
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.roles && !to.meta.roles.includes(auth.role)) {
    return { name: 'forbidden' }
  }
})
```
- `bootstrap()` hanya berjalan sekali per page load dan menangani race dengan `initialized`.
- Redirect setelah login **hanya** menerima path internal: diawali `/`, bukan `//`, bukan URL absolut.
  Selain itu jatuh ke `/tickets`.

### 5.4 Navigasi
- Setelah login sukses: `router.replace(redirect || '/tickets')`.
- Setelah register sukses: `/login` dengan query `registered=1` dan email terisi.
- Setelah create tiket: `/tickets/:id` + toast sukses.
- Tombol kembali pada detail: kembali ke daftar dengan query filter terakhir (`router.back()` bila
  memungkinkan, fallback `/tickets`).

## 6. State management

### 6.1 `useAuthStore`

```ts
state: {
  token: string | null          // diinisialisasi dari localStorage
  user: User | null
  initialized: boolean
  loading: boolean
  error: string | null
}
getters: {
  isAuthenticated: (s) => Boolean(s.token && s.user)
  role: (s) => s.user?.role ?? null
  isAdmin / isAgent / isRequester
  displayName
}
actions: {
  bootstrap()       // token ada -> GET /api/me; gagal -> logout()
  login(email, password)
  register(payload)
  fetchMe()
  logout()          // hapus token + user + redirect login
}
```
- Token disimpan di `localStorage` dengan key `helpdesk_token`.
- User **tidak** disimpan sebagai otoritas; selalu diambil dari `/api/me` saat boot.
- `logout()` tidak memanggil API (tidak ada endpoint logout); cukup bersihkan state.
- Perubahan peran oleh admin pada sesi sendiri dicegah backend; jika peran berubah di sesi lain,
  perubahan berlaku saat request berikutnya (backend memuat role dari DB).

### 6.2 `useTicketStore`

```ts
state: {
  items: TicketSummary[]
  meta: { page, per_page, total, total_pages }
  filters: { status, priority, assignee, unassigned, search, sort, order, page, per_page }
  listLoading, listError
  detail: TicketDetail | null
  detailLoading, detailError
  mutationLoading
}
actions: {
  fetchTickets(), fetchTicket(id), createTicket(payload),
  updateTicket(id, payload), addComment(id, payload),
  setFilters(partial), resetFilters(), resetDetail()
}
```
- Filter disimpan di store **dan** disinkronkan ke query string URL.
- `fetchTickets` membatalkan request lama bila filter berubah cepat (AbortController) untuk
  menghindari hasil balapan.
- Setelah mutasi (create/update/comment), data detail di-refresh dari server; tidak mengandalkan
  asumsi lokal.

### 6.3 `useUserStore` (admin)
```ts
state: { items, meta, filters, loading, error, updatingId }
actions: { fetchUsers(), updateRole(id, role) }
```

### 6.4 `useToastStore`
- `toasts: { id, type: 'success'|'error'|'info', message }[]`
- `push()` otomatis hilang setelah 4 detik; `dismiss(id)`.
- Toast dirender dalam `aria-live="polite"`.

## 7. HTTP client terpusat

`src/lib/http.ts`:

```ts
const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 15_000,
  headers: { 'Content-Type': 'application/json' }
})
```

### 7.1 Request interceptor
- Baca token dari `localStorage` (bukan dari store, menghindari circular import) dan tambahkan
  `Authorization: Bearer <token>` bila ada.
- Jangan menambahkan header untuk endpoint publik (`/auth/login`, `/auth/register`).

### 7.2 Response interceptor
- Sukses: kembalikan `response.data`.
- `401`:
  - Untuk `/auth/login` atau `/auth/register`: teruskan error ke pemanggil (jangan redirect).
  - Selain itu: panggil callback `onUnauthorized()` yang didaftarkan saat bootstrap aplikasi.
    Callback menghapus sesi dan mengarahkan ke `/login?redirect=<current>`.
  - Gunakan flag `isHandling401` agar banyak request paralel tidak memicu redirect berulang.
- `403`, `404`, `409`, `400`, `500`: normalisasi menjadi `ApiError` dan lempar ke pemanggil.

### 7.3 Normalisasi error
```ts
type ApiError = {
  status: number
  code: string
  message: string
  details?: Record<string, string>
}
```
- Bila response tidak memiliki envelope error (mis. jaringan putus), gunakan pesan
  "Tidak dapat terhubung ke server. Periksa koneksi Anda." dengan `code = 'NETWORK_ERROR'`.
- Timeout menghasilkan pesan "Permintaan terlalu lama. Coba lagi.".
- Komponen **tidak** membaca `err.response` langsung; selalu memakai helper `toApiError`.

### 7.4 Aturan pemakaian
- Semua panggilan API lewat store, bukan langsung dari komponen (kecuali kebutuhan lokal).
- Tidak ada token yang di-log.
- Tidak ada `fetch` mentah di komponen.

## 8. Inventaris komponen

```
src/
├── components/
│   ├── ui/
│   │   ├── AppButton.vue          # variant, size, loading, disabled
│   │   ├── AppInput.vue           # label, error, hint, aria
│   │   ├── AppSelect.vue
│   │   ├── AppTextarea.vue
│   │   ├── AppBadge.vue
│   │   ├── AppModal.vue           # focus trap, ESC, backdrop
│   │   ├── AppSpinner.vue
│   │   ├── AppToast.vue
│   │   ├── EmptyState.vue
│   │   ├── ErrorState.vue
│   │   ├── SkeletonBlock.vue
│   │   └── PaginationBar.vue
│   ├── tickets/
│   │   ├── StatusBadge.vue
│   │   ├── PriorityBadge.vue
│   │   ├── TicketTable.vue        # desktop
│   │   ├── TicketCardList.vue     # mobile
│   │   ├── TicketFilters.vue
│   │   ├── TicketForm.vue
│   │   ├── TicketMeta.vue
│   │   ├── TicketActions.vue      # transisi sesuai peran
│   │   ├── TicketEditPanel.vue    # prioritas & assignee (agent/admin)
│   │   ├── CommentList.vue
│   │   ├── CommentItem.vue
│   │   └── CommentForm.vue
│   └── users/
│       ├── UserTable.vue
│       └── RoleSelect.vue
├── composables/
│   ├── useToast.ts
│   ├── useDebounce.ts
│   ├── useConfirm.ts
│   └── useTicketPermissions.ts   # UI-only helper
├── layouts/
│   ├── AppLayout.vue
│   ├── AuthLayout.vue
│   ├── AppSidebar.vue
│   └── AppTopbar.vue
├── lib/
│   ├── http.ts
│   ├── errors.ts
│   ├── format.ts                 # tanggal, inisial, label enum
│   └── constants.ts              # enum, opsi filter, warna badge
├── router/
│   └── index.ts
├── stores/
│   ├── auth.ts
│   ├── tickets.ts
│   ├── users.ts
│   └── toast.ts
├── views/
│   ├── LoginView.vue
│   ├── RegisterView.vue
│   ├── TicketListView.vue
│   ├── TicketCreateView.vue
│   ├── TicketDetailView.vue
│   ├── AdminUsersView.vue
│   ├── ForbiddenView.vue
│   └── NotFoundView.vue
├── App.vue
└── main.ts
```

### 8.1 Komponen kunci

| Komponen | Tanggung jawab | Catatan |
|---|---|---|
| `AppButton` | Tombol konsisten | `loading` menampilkan spinner dan menonaktifkan klik ganda |
| `AppInput` | Label, error, hint | `aria-invalid`, `aria-describedby` otomatis |
| `StatusBadge` | Label + warna status | Selalu menampilkan teks, bukan warna saja |
| `PriorityBadge` | Label + ikon prioritas | `urgent` memakai ikon api |
| `TicketFilters` | Semua filter daftar | Debounce pencarian 350 ms; emit perubahan |
| `TicketActions` | Tombol transisi | Dihitung dari status + peran + kepemilikan |
| `CommentItem` | Satu komentar | Catatan internal bergaya amber + ikon kunci |
| `CommentForm` | Tambah komentar | Toggle internal hanya Agent/Admin |
| `EmptyState` | Tidak ada data | Membedakan "belum ada" vs. "filter tidak cocok" |
| `ErrorState` | Gagal memuat | Pesan + tombol "Coba lagi" |
| `PaginationBar` | Navigasi halaman | Menampilkan "1–20 dari 57" |

## 9. Spesifikasi halaman

### 9.1 Login (`/login`)

**Tujuan**: masuk dan mendapatkan sesi.

**Tampilan**
- Kartu terpusat (maks 420 px), logo/nama aplikasi, judul "Masuk ke Helpdesk".
- Field email, password (tombol tampil/sembunyikan), tombol "Masuk".
- Tautan "Belum punya akun? Daftar".
- Banner sukses bila query `registered=1`.

**Validasi klien**
| Field | Aturan | Pesan |
|---|---|---|
| email | wajib, format email | "Masukkan email yang valid." |
| password | wajib | "Password wajib diisi." |

**Perilaku**
- Submit menonaktifkan form dan menampilkan spinner.
- Error `401 INVALID_CREDENTIALS` ditampilkan sebagai banner: "Email atau password salah."
- Error jaringan ditampilkan sebagai banner dengan tombol coba lagi.
- Sukses: simpan token, ambil user dari response, redirect ke `redirect` yang aman atau `/tickets`.
- Tekan Enter men-submit.

**Acceptance criteria**
- [ ] Form tidak terkirim bila validasi gagal.
- [ ] Error server tampil tanpa menghapus input pengguna.
- [ ] Redirect `?redirect=/admin/users` dihormati untuk admin, ditolak untuk non-admin (guard).
- [ ] `?redirect=https://evil.example` diabaikan.

### 9.2 Register (`/register`)

**Tujuan**: membuat akun requester.

**Field**: nama, email, password, konfirmasi password.

**Validasi klien**
| Field | Aturan | Pesan |
|---|---|---|
| name | wajib, 2–100 | "Nama minimal 2 karakter." |
| email | wajib, format email, maks 255 | "Masukkan email yang valid." |
| password | wajib, min 8, mengandung huruf & angka | "Password minimal 8 karakter dan mengandung huruf serta angka." |
| confirmPassword | harus sama | "Konfirmasi password tidak cocok." |

**Perilaku**
- Info kecil: "Akun baru selalu memiliki peran Requester."
- Sukses `201`: redirect `/login?registered=1&email=<email>`, toast "Akun berhasil dibuat. Silakan masuk."
- `409 DUPLICATE_EMAIL`: error pada field email.
- `400`: map `details` ke field terkait; sisanya banner.

**Acceptance criteria**
- [ ] Tidak ada field peran di form.
- [ ] Password dan konfirmasi tidak pernah dikirim ke log/console.
- [ ] Email terisi otomatis di halaman login setelah register.

### 9.3 Daftar tiket (`/tickets`)

**Tujuan**: menemukan tiket dengan cepat.

**Tampilan**
- Header: judul "Tiket", jumlah total, tombol "+ Buat Tiket".
- Filter bar:
  - Input pencarian judul (ikon kaca pembesar, debounce 350 ms).
  - Select status: Semua, Open, In Progress, Resolved, Closed.
  - Select prioritas: Semua, Low, Medium, High, Urgent.
  - Select assignee (Agent/Admin): Semua, Belum di-assign, Saya, dan untuk Admin daftar agent.
  - Select urutan: Terbaru, Terlama, Prioritas tertinggi, Terakhir diperbarui.
  - Tombol "Reset filter" muncul bila ada filter aktif.
- Desktop: tabel dengan kolom: ID, Judul, Status, Prioritas, Kategori, Requester*, Assignee*, Diperbarui.
  Kolom bertanda `*` disembunyikan untuk Requester (requester = dirinya sendiri).
- Mobile (<768 px): daftar kartu berisi judul, badge status/prioritas, assignee, waktu.
- Footer: `PaginationBar` + pilihan 10/20/50 per halaman.

**Perilaku**
- Filter dan halaman disinkronkan ke query URL (`?status=open&page=2`). Refresh halaman
  mempertahankan filter.
- Perubahan filter mereset `page` ke 1.
- Klik baris/kartu → detail. Baris dapat difokuskan dan diaktifkan dengan Enter.
- Requester hanya menerima tiket miliknya dari backend; UI tidak menampilkan filter assignee.
- Setelah kembali dari detail, filter sebelumnya tetap aktif.

**State**
- Loading: skeleton tabel/kartu (5 baris).
- Empty (tanpa filter): ilustrasi + "Belum ada tiket" + CTA "Buat tiket pertama Anda".
- Empty (dengan filter): "Tidak ada tiket yang cocok" + tombol "Hapus filter".
- Error: `ErrorState` dengan pesan server + "Coba lagi".
- Pagination: tombol prev/next disabled pada batas; nomor halaman aktif.

**Acceptance criteria**
- [ ] Filter, pencarian, sort, dan pagination mengirim query yang benar ke API.
- [ ] URL dapat dibagikan dan menghasilkan tampilan yang sama.
- [ ] Requester tidak melihat tiket pengguna lain (diverifikasi lewat API).
- [ ] State loading, empty, dan error dapat dilihat dan diuji.
- [ ] Tabel berubah menjadi kartu di 375 px tanpa scroll horizontal.

### 9.4 Buat tiket (`/tickets/new`)

**Tujuan**: membuat tiket baru.

**Field**
| Field | Komponen | Default |
|---|---|---|
| Judul | input teks, counter 0/150 | kosong |
| Deskripsi | textarea, counter 0/5000 | kosong |
| Kategori | select (Hardware, Software, Network, Access, Other) | `other` |
| Prioritas | select (Low, Medium, High, Urgent) | `medium` |

**Validasi klien**
| Field | Aturan | Pesan |
|---|---|---|
| title | wajib, 3–150 | "Judul minimal 3 karakter." |
| description | wajib, 10–5000 | "Deskripsi minimal 10 karakter." |
| category | enum | "Pilih kategori." |
| priority | enum | "Pilih prioritas." |

**Perilaku**
- Tombol "Buat Tiket" dan "Batal".
- Submit sukses: toast "Tiket berhasil dibuat", redirect ke detail.
- Submit gagal: banner error + mapping `details` ke field.
- Mencegah double submit (tombol disabled + flag).
- Jika pengguna meninggalkan form yang sudah diisi, tampilkan konfirmasi (opsional, P1).

**Acceptance criteria**
- [ ] Tiket baru berstatus `open` dan muncul di daftar.
- [ ] Error validasi server tampil pada field yang tepat.
- [ ] Tidak ada field status/requester/assignee di form.

### 9.5 Detail tiket (`/tickets/:id`)

**Tujuan**: memahami tiket dan melakukan aksi yang diizinkan.

**Tata letak**
- Desktop ≥1024 px: dua kolom — konten utama (judul, deskripsi, komentar) dan sidebar aksi
  (status, prioritas, assignee, tombol transisi).
- Mobile: satu kolom; panel aksi di atas komentar, tombol sticky di bawah bila perlu.

**Konten**
- Header: `#ID`, judul, `StatusBadge`, `PriorityBadge`.
- Meta: Requester, Assignee (atau "Belum ditangani"), Kategori, Dibuat, Diperbarui.
- Deskripsi dengan `white-space: pre-wrap` (bukan `v-html`).
- Timeline komentar: penulis (nama + badge peran), waktu, isi, label "Catatan internal" bila ada.
- Form komentar di bawah; textarea + tombol kirim; toggle "Catatan internal" hanya Agent/Admin.

**Aksi berbasis peran dan status**

| Status | Requester (pemilik) | Agent / Admin |
|---|---|---|
| `open` | "Batalkan tiket" (open → closed) | "Ambil tiket" (open → in_progress), Admin juga "Tutup tiket" |
| `in_progress` | — | "Tandai selesai" (in_progress → resolved) |
| `resolved` | "Tutup tiket" (resolved → closed), "Buka kembali" (resolved → open) | "Tutup tiket" (resolved → closed) |
| `closed` | — (read-only) | — (read-only) |

- Tombol hanya muncul jika aksi valid; backend tetap memvalidasi.
- Semua aksi memakai modal konfirmasi dengan judul, konsekuensi, dan tombol konfirmasi.
- Setelah sukses: toast + refetch detail.
- Jika `409 INVALID_TRANSITION` (mis. data berubah di tab lain): tampilkan banner error dan refetch.
- Agent/Admin dapat mengubah prioritas dan assignee:
  - Select prioritas: low/medium/high/urgent.
  - Select assignee: Agent → "Saya" / "Tidak ada"; Admin → daftar agent/admin + "Tidak ada".
  - Tombol "Simpan perubahan" dengan loading; error ditampilkan.
- Komentar sukses: form dikosongkan, komentar baru muncul di timeline (refetch).
- Komentar internal tidak pernah dirender untuk Requester (backend sudah memfilter; UI tidak
  menyediakan toggle).

**State**
- Loading: skeleton header + body + komentar.
- `403`: "Anda tidak memiliki akses ke tiket ini" + tombol kembali.
- `404`: "Tiket tidak ditemukan" + tombol kembali.
- Error lain: `ErrorState` + retry.
- Aksi mutasi: tombol loading, form disabled, tidak ada double submit.
- Komentar kosong: "Belum ada komentar. Mulai diskusi."

**Acceptance criteria**
- [ ] Tombol transisi yang tampil persis sesuai tabel peran × status.
- [ ] Requester tidak melihat catatan internal atau toggle internal.
- [ ] Requester membuka tiket orang lain → halaman 403 yang ramah.
- [ ] Perubahan status langsung terlihat tanpa reload manual.
- [ ] Konfirmasi muncul sebelum aksi menutup/membatalkan/membuka kembali.

### 9.6 Kelola pengguna (`/admin/users`)

**Tujuan**: admin mengelola peran.

**Tampilan**
- Header "Kelola Pengguna" + total.
- Filter: pencarian nama/email (debounce), select peran (Semua/Requester/Agent/Admin).
- Tabel: Nama, Email, Peran (select inline), Dibuat, Aksi.
- Mobile: kartu dengan select peran.
- Pagination.

**Perilaku**
- Perubahan peran memakai modal konfirmasi: "Ubah peran X dari agent menjadi admin?"
- Baris milik sendiri: select disabled + tooltip "Anda tidak dapat mengubah peran sendiri."
- Sukses: toast + update baris.
- `409 LAST_ADMIN`: tampilkan error dan kembalikan select ke nilai sebelumnya.
- `403`: redirect ke `/forbidden` (atau tampilkan state akses ditolak).
- Loading/empty/error state sama seperti daftar tiket.

**Acceptance criteria**
- [ ] Hanya admin yang dapat membuka halaman (guard + backend).
- [ ] Role select menampilkan nilai terbaru setelah sukses.
- [ ] Peran sendiri tidak dapat diubah dari UI.
- [ ] Error server ditampilkan tanpa kehilangan data tabel.

### 9.7 Halaman 403 dan 404

- 403: ikon kunci, "Akses ditolak", penjelasan singkat, tombol "Kembali ke tiket".
- 404: ikon kompas, "Halaman tidak ditemukan", tombol "Ke beranda".
- Keduanya responsif dan tidak memerlukan login (kecuali 403 di dalam AppLayout).

## 10. Matriks visibilitas UI

| Elemen | Requester | Agent | Admin |
|---|---|---|---|
| Menu "Kelola Pengguna" | Tidak | Tidak | Ya |
| Tombol "Buat Tiket" | Ya | Ya | Ya |
| Filter assignee | Tidak | Ya (Saya/Belum) | Ya (semua agent) |
| Kolom Requester di tabel | Tidak | Ya | Ya |
| Panel ubah prioritas | Tidak | Ya | Ya |
| Panel ubah assignee | Tidak | Ya (diri sendiri) | Ya (semua) |
| Toggle catatan internal | Tidak | Ya | Ya |
| Tombol transisi | Sesuai status & kepemilikan | Sesuai status | Sesuai status |

Helper `useTicketPermissions(ticket)` mengembalikan `canChangeStatus(to)`, `canEditPriority`,
`canEditAssignee`, `canCommentInternal`. Helper ini **hanya untuk UI**; keputusan akhir di backend.

## 11. Pola loading, empty, dan error

| Konteks | Loading | Empty | Error |
|---|---|---|---|
| Daftar tiket | Skeleton 5 baris | EmptyState + CTA / hapus filter | ErrorState + retry |
| Detail tiket | Skeleton header + body | — | 403/404/ErrorState |
| Komentar | Skeleton 3 item | "Belum ada komentar" | Inline error + retry |
| Form submit | Spinner di tombol, input disabled | — | Banner + field errors |
| Admin users | Skeleton 5 baris | "Belum ada pengguna" | ErrorState + retry |
| Auth | Spinner tombol | — | Banner error |

Aturan:
- Tidak ada layout shift besar saat loading (skeleton meniru ukuran konten).
- Tombol retry memanggil ulang aksi yang sama.
- Error 500 menampilkan pesan generik + kode request bila tersedia; detail teknis hanya di console
  (tanpa token/data sensitif).

## 12. Responsivitas dan aksesibilitas

### 12.1 Breakpoint
| Rentang | Perilaku |
|---|---|
| <640 px | Satu kolom; tabel → kartu; filter dalam panel collapsible; tombol full-width |
| 640–1023 px | Dua kolom untuk filter; sidebar menjadi drawer |
| ≥1024 px | Sidebar tetap; detail dua kolom; tabel penuh |

- Tidak ada horizontal scroll pada 375 px.
- Target sentuh minimal 44×44 px.
- Modal tidak melebihi viewport; konten dapat di-scroll.

### 12.2 Aksesibilitas
- Semua input memiliki `<label for>`; error dihubungkan via `aria-describedby`.
- `aria-invalid="true"` pada field error.
- Fokus terlihat (`focus-visible:ring`).
- Modal: focus trap, `role="dialog"`, `aria-modal="true"`, ESC menutup, fokus kembali ke pemicu.
- Toast: `aria-live="polite"`.
- Badge status memakai teks, bukan hanya warna.
- Navigasi keyboard penuh; skip link ke konten utama.
- Setelah pindah route, fokus diarahkan ke heading halaman.
- Kontras teks minimal 4.5:1.

## 13. Pengujian frontend

### 13.1 Perangkat
- Vitest + Vue Test Utils + MSW (mock API).
- `@vue/test-utils` untuk mount komponen dengan Pinia dan router tiruan.
- E2E Playwright opsional untuk 3 alur utama.

### 13.2 Daftar test wajib (minimal 5; target 12)

| ID | Jenis | Skenario | Ekspektasi |
|---|---|---|---|
| FE-T01 | Unit (router) | Buka `/tickets` tanpa token | Redirect `/login?redirect=/tickets` |
| FE-T02 | Unit (router) | Agent membuka `/admin/users` | Redirect `/forbidden` |
| FE-T03 | Unit (router) | User login membuka `/login` | Redirect `/tickets` |
| FE-T04 | Unit (store) | `login()` sukses | Token tersimpan, user terisi, `isAuthenticated=true` |
| FE-T05 | Unit (http) | Response 401 pada endpoint terproteksi | Sesi dibersihkan, `onUnauthorized` dipanggil sekali |
| FE-T06 | Komponen | `TicketTable` tanpa data | `EmptyState` tampil, bukan tabel kosong |
| FE-T07 | Komponen | `StatusBadge`/`PriorityBadge` | Label dan kelas warna benar untuk semua nilai |
| FE-T08 | Komponen | Submit form create dengan judul kosong | Tidak memanggil API, pesan validasi tampil |
| FE-T09 | Unit (permissions) | Tabel peran × status | Tombol yang diizinkan saja yang muncul |
| FE-T10 | Komponen | `CommentForm` sebagai requester | Toggle internal tidak dirender |
| FE-T11 | Unit (filters) | Ubah filter + search | Query API dan URL diperbarui, `page` reset ke 1 |
| FE-T12 | Komponen | `RoleSelect` baris sendiri | Disabled |

### 13.3 Perintah
```bash
cd frontend
npm run test:unit -- --run
npm run lint
npm run build
```

Target coverage: **≥ 60%** pada `stores`, `router`, dan `lib`.

## 14. Build, environment, dan Docker

### 14.1 Environment
| Variabel | Default | Keterangan |
|---|---|---|
| `VITE_API_BASE_URL` | `/api` | Base URL API; di Docker memakai proxy Nginx |
| `VITE_APP_NAME` | `Helpdesk` | Nama di UI dan title |

File `.env.example` disediakan; `.env.local` tidak di-commit.

### 14.2 Build
- `npm ci` → `npm run build` → output `dist/`.
- Code splitting per route (`() => import(...)`).
- Asset diberi hash oleh Vite.
- Tidak ada `console.log` berisi data sensitif pada build produksi.

### 14.3 Docker
- Multi-stage: `node:20-alpine` (build) → `nginx:1.27-alpine` (serve).
- `nginx.conf`:
  - `try_files $uri $uri/ /index.html;` untuk SPA fallback.
  - `location /api/ { proxy_pass http://backend:8080; }` + header `X-Forwarded-*`.
  - Gzip untuk JS/CSS/JSON.
  - Cache immutable untuk `/assets/`.
- Port kontainer 80, dipetakan ke host 8080 oleh compose.
- Healthcheck: `GET /` mengembalikan 200.

## 15. Acceptance criteria frontend

- [ ] Enam halaman wajib tersedia dan dapat diakses sesuai peran.
- [ ] Guard mencegah akses tanpa login dan akses admin oleh non-admin.
- [ ] Token ditambahkan otomatis pada setiap request terproteksi.
- [ ] 401 membersihkan sesi dan mengarahkan ke login (sekali, tanpa loop).
- [ ] Form login, register, dan create tiket memiliki validasi klien dan menampilkan error server.
- [ ] Daftar tiket mendukung filter, pencarian, pagination, sort, dan sinkronisasi URL.
- [ ] Detail tiket menampilkan komentar dan tombol transisi sesuai peran/status.
- [ ] Catatan internal tidak terlihat oleh requester.
- [ ] Admin dapat melihat pengguna dan mengubah peran; peran sendiri tidak dapat diubah.
- [ ] Semua halaman memiliki state loading, empty, dan error.
- [ ] Layout berfungsi di 375 px dan 1440 px tanpa scroll horizontal.
- [ ] Tidak ada `v-html` untuk konten pengguna.
- [ ] `npm run build` dan `npm run test:unit` lulus.
- [ ] Frontend berjalan melalui Nginx di `docker compose up`.

## 16. Bonus (opsional, prioritas menurun)

| Bonus | Dampak | Estimasi |
|---|---|---|
| Polling detail tiket 15 detik | Mendekati real-time | S |
| SSE notifikasi | Badge tiket baru | M |
| Halaman statistik admin | Ringkasan jumlah per status | S |
| Dark mode | Preferensi pengguna | S |
| E2E Playwright | Bukti alur utama | M |
| Upload lampiran | Drag & drop + progress | M |

## 17. Rencana kerja

| Prioritas | Tugas | Output |
|---|---|---|
| P0 | Setup Vite, Tailwind, router, Pinia, HTTP client | Aplikasi kosong berjalan |
| P0 | Auth store + guard + halaman login/register | Alur auth lengkap |
| P0 | Layout responsif + komponen UI dasar | Fondasi tampilan |
| P0 | Daftar tiket + filter + pagination | Halaman daftar |
| P0 | Form buat tiket | Halaman create |
| P0 | Detail tiket + komentar + aksi transisi | Halaman detail |
| P0 | Admin users + role select | Halaman admin |
| P0 | State loading/empty/error + 403/404 | UX lengkap |
| P1 | Unit test (minimal 5) + lint | Kualitas |
| P1 | Dockerfile + nginx | Satu perintah jalan |
| P2 | Bonus polling/statistik/E2E | Nilai tambah |

## 18. Risiko dan mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Daftar assignee untuk Agent | Agent tidak boleh akses `/api/users` | Sediakan opsi "Saya"/"Belum di-assign"; hanya Admin memuat daftar agent |
| Filter cepat memicu race | Hasil tidak konsisten | Debounce + AbortController |
| Redirect loop saat 401 | Aplikasi tidak dapat dipakai | Flag `isHandling401`, kecualikan endpoint auth |
| Peran berubah di sesi lain | UI menampilkan aksi lama | Backend menolak dengan 403/409; UI refetch dan tampilkan error |
| Token kedaluwarsa saat mengisi form | Kehilangan input | Simpan draft di memori; setelah login kembali, arahkan ulang bila memungkinkan |
| Tabel lebar di mobile | Tidak nyaman | Ganti ke kartu di <768 px |
| Error server tidak terpetakan | Pesan membingungkan | Helper `toApiError` + fallback generik |
| Scope melebar ke bonus | Wajib tidak selesai | Disiplin P0 → P1 → P2, dokumentasikan yang tertunda |

## 19. Catatan untuk README frontend

- Cara install: `npm ci`, `npm run dev` (butuh backend di `http://localhost:8080`).
- Konfigurasi `VITE_API_BASE_URL`.
- Perintah test, lint, build.
- Daftar akun seed (dari README utama).
- Keputusan: token di `localStorage`, guard di router, error dinormalisasi di satu tempat.
- Yang sengaja ditinggalkan: refresh token, SSE, upload, dark mode.
