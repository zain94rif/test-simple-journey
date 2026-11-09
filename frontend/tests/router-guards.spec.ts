import { beforeEach, describe, expect, it } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import routerApp from '@/router'
import { writeToken, clearToken } from '@/lib/http'
import { getSingle } from '@/lib/api'
import type { User } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

const getUser = vi.mocked(getSingle)

function makeUser(id: number, role: string): User {
  return { id, name: `U${id}`, email: `u${id}@x`, role, created_at: '' } as User
}

describe('router guards', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    getUser.mockReset()
  })

  it('unauthenticated user hitting /tickets is redirected to login with redirect query', async () => {
    await routerApp.push('/tickets')
    await routerApp.isReady()
    expect(routerApp.currentRoute.value.name).toBe('login')
    expect(routerApp.currentRoute.value.query.redirect).toBe('/tickets')
  })

  it('authenticated requester on /login is bounced back to /tickets (guestOnly)', async () => {
    writeToken('t-req')
    getUser.mockResolvedValueOnce(makeUser(10, 'requester'))
    await routerApp.push('/tickets')
    await routerApp.isReady()
    await routerApp.push('/login')
    expect(routerApp.currentRoute.value.name).toBe('tickets')
    clearToken()
  })

  it('requester is rejected from /admin/users with /forbidden', async () => {
    writeToken('t-req')
    getUser.mockResolvedValueOnce(makeUser(10, 'requester'))
    await routerApp.push({ name: 'admin-users' })
    await routerApp.isReady()
    expect(routerApp.currentRoute.value.name).toBe('forbidden')
    clearToken()
  })

  it('admin may open /admin/users', async () => {
    writeToken('t-admin')
    getUser.mockResolvedValueOnce(makeUser(1, 'admin'))
    await routerApp.push({ name: 'admin-users' })
    await routerApp.isReady()
    expect(routerApp.currentRoute.value.name).toBe('admin-users')
    clearToken()
  })

  it('unknown paths render the 404 route', async () => {
    await routerApp.push('/definitely/not/a/path')
    await routerApp.isReady()
    expect(routerApp.currentRoute.value.name).toBe('not-found')
  })
})
