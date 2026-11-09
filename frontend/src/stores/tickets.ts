import { defineStore } from 'pinia'
import { getList, getSingle, patchSingle, postSingle } from '@/lib/api'
import { useAuthStore } from './auth'
import type {
  Comment,
  PaginationMeta,
  Ticket,
  TicketDetail,
  TicketFilters,
} from '@/lib/types'

export interface CreateTicketPayload {
  title: string
  description: string
  category: string
  priority: string
}

export interface UpdateTicketPayload {
  status?: string
  priority?: string
  assignee_id?: number | null
}

// Module-level cancellation handle for the ticket list.
let listAborter: AbortController | undefined

/** How long a successful ticket-list response is considered fresh (ms). */
const LIST_FRESH_MS = 30_000

/**
 * Cache bookkeeping so re-entering the list page with the same filters does
 * not trigger a redundant request. Cleared by every mutation so fresh data
 * is always shown after create/update/comment.
 */
interface ListSnapshot {
  params: Record<string, string | number | boolean>
  requesterScope: boolean
  fetchedAt: number
}
let lastList: ListSnapshot | null = null

export const useTicketStore = defineStore('tickets', {
  state: () => ({
    items: [] as Ticket[],
    meta: { page: 1, per_page: 20, total: 0, total_pages: 0 } as PaginationMeta,
    filters: {
      status: '',
      priority: '',
      assignee: '',
      unassigned: false,
      search: '',
      sort: 'created_at:desc',
      order: 'desc',
      page: 1,
      per_page: 20,
    } as TicketFilters,
    listLoading: false,
    listError: null as string | null,
    detail: null as TicketDetail | null,
    detailLoading: false,
    detailError: null as string | null,
    mutationLoading: false,
  }),
  actions: {
    /** Build API params from the current filter state, honoring role scope. */
    buildParams(): Record<string, string | number | boolean> {
      const auth = useAuthStore()
      const f = this.filters
      const [sortField, sortOrder] = f.sort.split(':')
      const params: Record<string, string | number | boolean> = {
        sort: sortField,
        order: sortOrder ?? 'desc',
        page: f.page,
        per_page: f.per_page,
      }
      if (f.status) params.status = f.status
      if (f.priority) params.priority = f.priority
      if (f.search) params.search = f.search
      if (!auth.isRequester) {
        if (f.assignee === 'unassigned' || f.unassigned) params.unassigned = true
        if (f.assignee === 'me' && auth.user?.id !== undefined)
          params.assignee_id = auth.user.id
      }
      return params
    },
    async fetchTickets() {
      // Cancel an in-flight request so a fast filter change never races.
      if (listAborter) listAborter.abort()
      const controller = new AbortController()
      listAborter = controller
      this.listLoading = true
      this.listError = null
      try {
        const params = this.buildParams()
        const res = await getList<Ticket>('/tickets', {
          params,
          signal: controller.signal,
        })
        this.items = res.data
        if (res.meta) this.meta = res.meta
        lastList = {
          params,
          requesterScope: useAuthStore().isRequester,
          fetchedAt: Date.now(),
        }
      } catch (err) {
        if (controller.signal.aborted) return
        const message =
          (err as { message?: string })?.message ?? 'Tidak dapat memuat daftar tiket.'
        this.listError = message
        throw err
      } finally {
        if (listAborter === controller) {
          listAborter = undefined
          this.listLoading = false
        }
      }
    },    async fetchTicket(id: number | string) {
      this.detailLoading = true
      this.detailError = null
      try {
        this.detail = await getSingle<TicketDetail>(`/tickets/${id}`)
      } catch (err) {
        const message =
          (err as { message?: string })?.message ?? 'Tidak dapat memuat tiket.'
        this.detailError = message
        throw err
      } finally {
        this.detailLoading = false
      }
    },
    async createTicket(payload: CreateTicketPayload) {
      this.mutationLoading = true
      try {
        return await postSingle<Ticket>('/tickets', payload)
      } finally {
        this.mutationLoading = false
        lastList = null // a new ticket must appear on the next visit
      }
    },
    async updateTicket(id: number | string, patch: UpdateTicketPayload) {
      this.mutationLoading = true
      try {
        return await patchSingle<Ticket>(`/tickets/${id}`, patch)
      } finally {
        this.mutationLoading = false
        lastList = null
      }
    },
    async addComment(id: number | string, body: string, isInternal: boolean): Promise<Comment> {
      this.mutationLoading = true
      try {
        const comment = await postSingle<Comment>(`/tickets/${id}/comments`, {
          body,
          is_internal: isInternal,
        })
        // Recentre on server truth instead of patching local state.
        await this.fetchTicket(id)
        return comment
      } finally {
        this.mutationLoading = false
        lastList = null
      }
    },
    /**
     * True when the store already holds a fresh list for exactly these
     * params, so the view can skip the request on re-entry.
     */
    canUseCachedList(): boolean {
      if (!lastList || this.items.length === 0) return false
      if (Date.now() - lastList.fetchedAt > LIST_FRESH_MS) return false
      const auth = useAuthStore()
      if (auth.isRequester !== lastList.requesterScope) return false
      const next = this.buildParams()
      const prev = lastList.params
      const keys = new Set([...Object.keys(prev), ...Object.keys(next)])
      for (const key of keys) {
        if (prev[key] !== next[key]) return false
      }
      return true
    },
    /** Forget the cached list snapshot (used after logout or in tests). */
    invalidateListCache() {
      lastList = null
    },
    setFilters(partial: Partial<TicketFilters>) {
      const pageChanged = partial.page !== undefined
      this.filters = { ...this.filters, ...partial }
      if (!pageChanged) this.filters.page = 1
    },
    resetFilters() {
      this.filters = {
        status: '',
        priority: '',
        assignee: '',
        unassigned: false,
        search: '',
        sort: 'created_at:desc',
        order: 'desc',
        page: 1,
        per_page: 20,
      }
    },
    resetDetail() {
      this.detail = null
      this.detailError = null
    },
    hasActiveFilters(): boolean {
      const f = this.filters
      return Boolean(
        f.status || f.priority || f.assignee || f.search || f.unassigned ||
        f.sort !== 'created_at:desc' || f.page !== 1 || f.per_page !== 20,
      )
    },
  },
})
