import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/layouts/AppLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: { name: 'tickets' },
  },
  {
    path: '/login',
    component: AuthLayout,
    children: [
      {
        path: '',
        name: 'login',
        component: () => import('@/views/LoginView.vue'),
        meta: { guestOnly: true, title: 'Masuk' },
      },
    ],
  },
  {
    path: '/register',
    component: AuthLayout,
    children: [
      {
        path: '',
        name: 'register',
        component: () => import('@/views/RegisterView.vue'),
        meta: { guestOnly: true, title: 'Daftar' },
      },
    ],
  },
  {
    path: '/tickets',
    component: AppLayout,
    children: [
      {
        path: '',
        name: 'tickets',
        component: () => import('@/views/TicketListView.vue'),
        meta: { requiresAuth: true, title: 'Tiket' },
      },
    ],
  },
  {
    path: '/tickets/new',
    component: AppLayout,
    children: [
      {
        path: '',
        name: 'ticket-create',
        component: () => import('@/views/TicketCreateView.vue'),
        meta: { requiresAuth: true, title: 'Buat Tiket' },
      },
    ],
  },
  {
    path: '/tickets/:id',
    component: AppLayout,
    children: [
      {
        path: '',
        name: 'ticket-detail',
        component: () => import('@/views/TicketDetailView.vue'),
        meta: { requiresAuth: true, title: 'Detail Tiket' },
      },
    ],
  },
  {
    path: '/admin/users',
    component: AppLayout,
    children: [
      {
        path: '',
        name: 'admin-users',
        component: () => import('@/views/AdminUsersView.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: 'Kelola Pengguna' },
      },
    ],
  },
  {
    path: '/forbidden',
    name: 'forbidden',
    component: () => import('@/views/ForbiddenView.vue'),
    meta: { title: 'Akses Ditolak' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { title: 'Tidak Ditemukan' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.initialized) await auth.bootstrap()

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'tickets' }
  }
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  const allowed = to.meta.roles as string[] | undefined
  if (allowed && !allowed.includes(auth.role ?? '')) {
    return { name: 'forbidden' }
  }
})

router.afterEach((to) => {
  const title = (to.meta.title as string | undefined) ?? 'Helpdesk'
  document.title = `${title} — Helpdesk`
})

export default router
