import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { useTicketPermissions } from '@/composables/useTicketPermissions'
import type { Ticket, User } from '@/lib/types'

// No network in unit tests: auth hydrates via getSingle only when a token exists.
vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

function makeTicket(overrides: Partial<Ticket> = {}): Ticket {
  return {
    id: 1,
    status: 'open',
    priority: 'medium',
    requester: { id: 100, name: 'Req', email: 'req@x', role: 'requester' } as unknown as User,
    ...overrides,
  } as unknown as Ticket
}

function loginAs(user: Partial<User> & Pick<User, 'id'>, role: string) {
  const auth = useAuthStore()
  auth.user = { id: user.id, name: 'U', email: 'u@x', role, created_at: '' } as User
  return auth
}

describe('useTicketPermissions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('returns all-false for a null ticket', () => {
    loginAs({ id: 1 }, 'admin')
    const perms = useTicketPermissions(() => null)
    expect(perms.value).toEqual({
      canClaim: false,
      canClose: false,
      canResolve: false,
      canReopen: false,
      canCancel: false,
      canEditPriority: false,
      canEditAssignee: false,
      canCommentInternal: false,
    })
  })

  it('agent can claim an open ticket but cannot reopen', () => {
    loginAs({ id: 2 }, 'agent')
    const perms = useTicketPermissions(() => makeTicket({ status: 'open' }))
    expect(perms.value.canClaim).toBe(true)
    expect(perms.value.canReopen).toBe(false)
    expect(perms.value.canClose).toBe(false)
    expect(perms.value.canResolve).toBe(false)
  })

  it('agent can resolve and close a resolved ticket (staff closing)', () => {
    loginAs({ id: 2 }, 'agent')
    const perms = useTicketPermissions(() => makeTicket({ status: 'in_progress' }))
    expect(perms.value.canResolve).toBe(true)
    const openToResolved = useTicketPermissions(() => makeTicket({ status: 'resolved' }))
    expect(openToResolved.value.canClose).toBe(true)
    expect(openToResolved.value.canReopen).toBe(false)
  })

  it('requester-owner on resolved ticket can close and reopen', () => {
    loginAs({ id: 100 }, 'requester')
    const perms = useTicketPermissions(() => makeTicket({ status: 'resolved' }))
    expect(perms.value.canClose).toBe(true)
    expect(perms.value.canReopen).toBe(true)
    expect(perms.value.canResolve).toBe(false)
    expect(perms.value.canEditPriority).toBe(false)
    expect(perms.value.canCommentInternal).toBe(false)
  })

  it('non-owner requester cannot act on open tickets owned by others', () => {
    loginAs({ id: 999 }, 'requester')
    const perms = useTicketPermissions(() => makeTicket({ status: 'open' }))
    expect(perms.value.canClaim).toBe(false)
    expect(perms.value.canClose).toBe(false)
    expect(perms.value.canCancel).toBe(false)
  })

  it('owner can cancel their own open ticket', () => {
    loginAs({ id: 100 }, 'requester')
    const perms = useTicketPermissions(() => makeTicket({ status: 'open' }))
    expect(perms.value.canCancel).toBe(true)
  })

  it('closed tickets permit no transitions other than nothing', () => {
    loginAs({ id: 100 }, 'requester')
    const perms = useTicketPermissions(() => makeTicket({ status: 'closed' }))
    for (const flag of [perms.value.canClaim, perms.value.canClose, perms.value.canResolve, perms.value.canReopen, perms.value.canCancel]) {
      expect(flag).toBe(false)
    }
  })

  it('staff can edit priority and assignee, requesters cannot', () => {
    loginAs({ id: 2 }, 'agent')
    let perms = useTicketPermissions(() => makeTicket())
    expect(perms.value.canEditPriority).toBe(true)
    expect(perms.value.canEditAssignee).toBe(true)

    loginAs({ id: 100 }, 'requester')
    perms = useTicketPermissions(() => makeTicket())
    expect(perms.value.canEditPriority).toBe(false)
    expect(perms.value.canEditAssignee).toBe(false)
  })
})
