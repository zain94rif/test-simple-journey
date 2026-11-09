import { beforeEach, describe, expect, it, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import { postSingle, getSingle } from '@/lib/api'
import { clearToken, writeToken, readToken } from '@/lib/http'
import type { User } from '@/lib/types'

vi.mock('@/lib/api', () => ({
  getSingle: vi.fn(),
  getList: vi.fn(),
  postSingle: vi.fn(),
  patchSingle: vi.fn(),
}))

const postMock = vi.mocked(postSingle)
const getMock = vi.mocked(getSingle)

const fakeUser: User = {
  id: 1,
  name: 'Admin Solusi',
  email: 'admin@helpdesk.local',
  role: 'admin',
  created_at: '2026-01-01T00:00:00Z',
}

describe('auth store', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    postMock.mockReset()
    getMock.mockReset()
  })

  it('bootstrap does nothing without a stored token and never calls /me', async () => {
    const auth = useAuthStore()
    await auth.bootstrap()
    expect(auth.initialized).toBe(true)
    expect(auth.user).toBeNull()
    expect(getMock).not.toHaveBeenCalled()
  })

  it('login posts credentials, stores the token, and persists it', async () => {
    postMock.mockResolvedValueOnce({ token: 'jwt-abc', user: fakeUser } as never)
    const auth = useAuthStore()
    await auth.login('admin@helpdesk.local', 'Admin123!')
    expect(postMock).toHaveBeenCalledWith('/auth/login', {
      email: 'admin@helpdesk.local',
      password: 'Admin123!',
    })
    expect(auth.token).toBe('jwt-abc')
    expect(auth.user?.role).toBe('admin')
    expect(auth.isAuthenticated).toBe(true)
    expect(readToken()).toBe('jwt-abc')
    expect(postMock.mock.calls[0][0]).toBe('/auth/login')
  })

  it('login failure clears loading state and propagates the error', async () => {
    postMock.mockRejectedValueOnce(new Error('INVALID_CREDENTIALS'))
    const auth = useAuthStore()
    await expect(auth.login('a@b.c', 'wrong')).rejects.toThrow('INVALID_CREDENTIALS')
    expect(auth.loading).toBe(false)
    expect(auth.error).toBeNull()
    expect(auth.token).toBeNull()
  })

  it('bootstrap with an invalid token clears the session (token never trusted)', async () => {
    writeToken('expired-token')
    getMock.mockRejectedValueOnce(new Error('UNAUTHORIZED'))
    const auth = useAuthStore()
    await auth.bootstrap()
    expect(auth.user).toBeNull()
    expect(auth.token).toBeNull()
    expect(readToken()).toBeNull()
    clearToken()
  })

  it('logout clears token and user', () => {
    writeToken('jwt-logout')
    const auth = useAuthStore()
    auth.token = 'jwt-logout'
    auth.user = fakeUser
    auth.logout()
    expect(auth.token).toBeNull()
    expect(auth.user).toBeNull()
    expect(readToken()).toBeNull()
    clearToken()
  })
})
