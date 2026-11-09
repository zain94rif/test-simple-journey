import { defineStore } from 'pinia'
import { getSingle, postSingle } from '@/lib/api'
import { clearToken, readToken, writeToken } from '@/lib/http'
import type { LoginResponse, User } from '@/lib/types'
import { useUserStore } from './users'

interface RegisterPayload {
  name: string
  email: string
  password: string
}

export const useAuthStore = defineStore('auth', {
  state: () => {
    const token = readToken()
    return {
      token: token as string | null,
      user: null as User | null,
      initialized: false,
      loading: false,
      error: null as string | null,
    }
  },
  getters: {
    isAuthenticated: (state) => Boolean(state.token && state.user),
    role: (state) => state.user?.role ?? null,
    isAdmin(): boolean {
      return this.role === 'admin'
    },
    isAgent(): boolean {
      return this.role === 'agent'
    },
    isRequester(): boolean {
      return this.role === 'requester'
    },
    isStaff(): boolean {
      return this.role === 'agent' || this.role === 'admin'
    },
    displayName: (state) => state.user?.name ?? '',
  },
  actions: {
    /**
     * Runs once per page load. When a token exists, hydrate the user from the
     * server; the stored user is never trusted as the source of truth.
     */
    async bootstrap() {
      if (this.initialized) return
      this.initialized = true
      if (!this.token) return
      try {
        this.user = await getSingle<User>('/me')
      } catch {
        this.clearSession()
      }
    },
    async login(email: string, password: string) {
      this.loading = true
      this.error = null
      try {
        const res = await postSingle<LoginResponse>('/auth/login', { email, password })
        this.token = res.token
        writeToken(res.token)
        this.user = res.user
      } finally {
        this.loading = false
      }
    },
    async register(payload: RegisterPayload) {
      this.loading = true
      try {
        return await postSingle<User>('/auth/register', payload)
      } finally {
        this.loading = false
      }
    },
    async fetchMe() {
      this.user = await getSingle<User>('/me')
    },
    clearSession() {
      this.token = null
      this.user = null
      clearToken()
      // Never let the next account see the previous account's cached data.
      useUserStore().invalidateUsersCache()
    },
    logout() {
      this.clearSession()
    },
  },
})
