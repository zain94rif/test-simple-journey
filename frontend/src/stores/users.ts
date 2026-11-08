import { defineStore } from 'pinia'
import { getList, patchSingle } from '@/lib/api'
import type { AdminRolePayload, User, UserFilters } from '@/lib/types'
import type { PaginationMeta } from '@/lib/types'

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
        const res = await getList<User>('/users', { params: this.buildParams() })
        this.items = res.data
        if (res.meta) this.meta = res.meta
      } finally {
        this.loading = false
      }
    },
    async updateRole(id: number, payload: AdminRolePayload) {
      this.updatingId = id
      try {
        return await patchSingle<User>(`/users/${id}/role`, payload)
      } finally {
        this.updatingId = 0
      }
    },
    setFilters(partial: Partial<UserFilters>) {
      const pageChanged = partial.page !== undefined
      this.filters = { ...this.filters, ...partial }
      if (!pageChanged) this.filters.page = 1
    },
  },
})
