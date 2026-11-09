import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useTicketStore } from '@/stores/tickets'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/users'
import { getList } from '@/lib/api'
import type { Ticket, User } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

const getListMock = vi.mocked(getList)

function listViewResponse(total: number) {
  return {
    data: Array.from({ length: total }, (_, i) => ({ id: i + 1, title: `T${i + 1}` })) as unknown as Ticket[],
    meta: { page: 1, per_page: 20, total, total_pages: Math.max(1, Math.ceil(total / 20)) },
  }
}

function loginAs(id: number, role: string) {
  const auth = useAuthStore()
  auth.user = { id, name: 'U', email: 'u@x', role, created_at: '' } as User
}

describe('ticket list cache (canUseCachedList)', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    getListMock.mockReset()
    getListMock.mockResolvedValue(listViewResponse(12) as never)
  })

  it('is false before any fetch and true right after a successful fetch', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    expect(store.canUseCachedList()).toBe(false)
    await store.fetchTickets()
    expect(store.canUseCachedList()).toBe(true)
    expect(getListMock).toHaveBeenCalledTimes(1)
  })

  it('stays true while params are unchanged (re-entry skip)', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    store.setFilters({ status: '' })
    expect(store.canUseCachedList()).toBe(true)
  })

  it('is false when filters change (page, status, etc.)', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    store.setFilters({ page: 2 })
    expect(store.canUseCachedList()).toBe(false)
    store.setFilters({ page: 1, status: 'open' })
    expect(store.canUseCachedList()).toBe(false) // status berbeda dari snapshot
  })

  it('is false when the account role scope changes (admin → requester)', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    loginAs(10, 'requester')
    expect(store.canUseCachedList()).toBe(false)
  })

  it('is false after a mutation invalidates the snapshot', async () => {
    loginAs(1, 'requester')
    const { postSingle } = await import('@/lib/api')
    vi.mocked(postSingle).mockResolvedValue({ id: 99 } as never)
    const store = useTicketStore()
    await store.fetchTickets()
    expect(store.canUseCachedList()).toBe(true)
    await store.createTicket({ title: 'New', description: 'd', category: 'hardware', priority: 'low' as never })
    expect(store.canUseCachedList()).toBe(false)
  })

  it('invalidateListCache clears the snapshot without touching items', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    const items = store.items
    store.invalidateListCache()
    expect(store.canUseCachedList()).toBe(false)
    expect(store.items).toBe(items)
  })

  it('logout clears the cache for the next account', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    const auth = useAuthStore()
    auth.logout()
    expect(store.canUseCachedList()).toBe(false)
    // user store untouched by this design decision: stays fetch-on-mount
    expect(useUserStore().fetchUsers).toBeTypeOf('function')
  })
})
