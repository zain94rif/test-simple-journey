import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useTicketStore } from '@/stores/tickets'
import { useUserStore } from '@/stores/users'
import { useAuthStore } from '@/stores/auth'
import { getList } from '@/lib/api'
import type { Ticket, User } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

const getListMock = vi.mocked(getList)

const ticketList = (total: number) => ({
  data: Array.from({ length: total }, (_, i) => ({ id: i + 1, title: `T${i + 1}` })) as unknown as Ticket[],
  meta: { page: 1, per_page: 20, total, total_pages: Math.max(1, Math.ceil(total / 20)) },
})
const userList = (total: number) => ({
  data: Array.from({ length: total }, (_, i) => ({ id: i + 1, name: `U${i + 1}` })) as unknown as User[],
  meta: { page: 1, per_page: 20, total, total_pages: Math.max(1, Math.ceil(total / 20)) },
})

function loginAs(id: number, role: string) {
  const auth = useAuthStore()
  auth.user = { id, name: 'U', email: 'u@x', role, created_at: '' } as User
}

describe('ticket list: always fetch on (re-)mount (data berubah sering)', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    getListMock.mockReset()
    getListMock.mockResolvedValue(ticketList(12) as never)
  })

  it('fetches again on a second call even with identical filters', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    await store.fetchTickets()
    expect(getListMock).toHaveBeenCalledTimes(2)
  })

  it('listLoading resets between fetches (skeleton shows on re-entry)', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    await store.fetchTickets()
    const pending = store.fetchTickets()
    expect(store.listLoading).toBe(true)
    await pending
    expect(store.listLoading).toBe(false)
  })

  it('drops back to page 1 when a non-page filter changes', async () => {
    loginAs(1, 'admin')
    const store = useTicketStore()
    store.setFilters({ page: 3 })
    expect(store.filters.page).toBe(3)
    store.setFilters({ status: 'open' })
    expect(store.filters.page).toBe(1)
  })
})

describe('users cache (canUseCachedUsers — data jarang berubah)', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    getListMock.mockReset()
    getListMock.mockResolvedValue(userList(6) as never)
  })

  it('is false before any fetch and true right after a successful fetch', async () => {
    const store = useUserStore()
    expect(store.canUseCachedUsers()).toBe(false)
    await store.fetchUsers()
    expect(store.canUseCachedUsers()).toBe(true)
    expect(getListMock).toHaveBeenCalledTimes(1)
  })

  it('stays true while filters are unchanged (re-entry skip)', async () => {
    const store = useUserStore()
    await store.fetchUsers()
    store.setFilters({ page: 1 })
    expect(store.canUseCachedUsers()).toBe(true)
  })

  it('is false when filters change (search, role, page)', async () => {
    const store = useUserStore()
    await store.fetchUsers()
    store.setFilters({ page: 2 })
    expect(store.canUseCachedUsers()).toBe(false)
    store.setFilters({ page: 1, search: 'admin' })
    expect(store.canUseCachedUsers()).toBe(false)
  })

  it('is invalidated after updateRole and re-fetches next time', async () => {
    const { patchSingle } = await import('@/lib/api')
    vi.mocked(patchSingle).mockResolvedValue({ id: 2 } as never)
    const store = useUserStore()
    await store.fetchUsers()
    expect(store.canUseCachedUsers()).toBe(true)
    await store.updateRole(2, { role: 'agent' })
    expect(store.canUseCachedUsers()).toBe(false)
    await store.fetchUsers()
    expect(getListMock).toHaveBeenCalledTimes(2)
  })

  it('invalidateUsersCache clears the snapshot without touching items', async () => {
    const store = useUserStore()
    await store.fetchUsers()
    const items = store.items
    store.invalidateUsersCache()
    expect(store.canUseCachedUsers()).toBe(false)
    expect(store.items).toBe(items)
  })

  it('logout clears the users cache for the next account', async () => {
    loginAs(1, 'admin')
    const store = useUserStore()
    await store.fetchUsers()
    const auth = useAuthStore()
    auth.logout()
    expect(store.canUseCachedUsers()).toBe(false)
  })
})
