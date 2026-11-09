import { defineStore } from 'pinia'
import { getList, patchSingle } from '@/lib/api'
import type { AdminRolePayload, User, UserFilters } from '@/lib/types'
import type { PaginationMeta } from '@/lib/types'

/** How long a successful users response is considered fresh (ms). */
const USERS_FRESH_MS = 30_000

/**
 * Cache snapshot so re-entering the admin page with the same filters serves
 * the store without a request. User data rarely changes on its own; the
 * snapshot is invalidated by updateRole (the only mutation).
 */
interface UsersSnapshot {
  params: Record<string, string | number | boolean>
  fetchedAt: number
}
let lastUsers: UsersSnapshot | null = null

export const useUserStore = defineStore('users', {
  state: () => ({
    items: [] as User[],
    meta: { page: 1, per_page: 20, total: 0, total_pages: 0 } as PaginationMeta,
    filters: {
      role: '',
      search: '',
      page: 1,
      per_page: 20,
    } as UserFilters,
    loading: false,
    error: null as string | null,
    updatingId: 0,
  }),
  actions: {
    buildParams(): Record<string, string | number | boolean> {
      const f = this.filters
      const params: Record<string, string | number | boolean> = {
        page: f.page,
        per_page: f.per_page,
      }
      if (f.role) params.role = f.role
      if (f.search) params.search = f.search
      return params
    },
    async fetchUsers() {
      this.loading = true
      this.error = null
      try {
        const params = this.buildParams()
        const res = await getList<User>('/users', { params })
        this.items = res.data
        if (res.meta) this.meta = res.meta
        lastUsers = { params, fetchedAt: Date.now() }
      } finally {
        this.loading = false
      }
    },
    /** True when the store holds a fresh list for exactly these filters. */
    canUseCachedUsers(): boolean {
      if (!lastUsers || this.items.length === 0) return false
      if (Date.now() - lastUsers.fetchedAt > USERS_FRESH_MS) return false
      const next = this.buildParams()
      const prev = lastUsers.params
      const keys = new Set([...Object.keys(prev), ...Object.keys(next)])
      for (const key of keys) {
        if (prev[key] !== next[key]) return false
      }
      return true
    },
    /** Forget the cached snapshot (used after logout or in tests). */
    invalidateUsersCache() {
      lastUsers = null
    },
    async updateRole(id: number, payload: AdminRolePayload) {
      this.updatingId = id
      try {
        return await patchSingle<User>(`/users/${id}/role`, payload)
      } finally {
        this.updatingId = 0
        lastUsers = null // the role changed: show server truth next time
      }
    },
    setFilters(partial: Partial<UserFilters>) {
      const pageChanged = partial.page !== undefined
      this.filters = { ...this.filters, ...partial }
      if (!pageChanged) this.filters.page = 1
    },
  },
})
