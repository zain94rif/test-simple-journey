import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { toApiError } from './errors'
import { TOKEN_KEY } from './constants'

const PUBLIC_PATHS = ['/auth/login', '/auth/register']

type UnauthorizedHandler = (currentPath: string) => void

let onUnauthorized: UnauthorizedHandler | null = null
let isHandling401 = false

/** Registered once during app bootstrap; clears the session and redirects. */
export function setUnauthorizedHandler(handler: UnauthorizedHandler) {
  onUnauthorized = handler
}

export function resetUnauthorizedHandler() {
  onUnauthorized = null
  isHandling401 = false
}

export function readToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function writeToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export const http = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? '/api',
  timeout: 15_000,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const path = config.url ?? ''
  const isPublic = PUBLIC_PATHS.some((p) => path.startsWith(p))
  const token = readToken()
  if (token && !isPublic) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    const status = error.response?.status
    const config = error.config as InternalAxiosRequestConfig | undefined
    const path = config?.url ?? ''
    const isAuthCall = PUBLIC_PATHS.some((p) => path.startsWith(p))

    if (status === 401 && !isAuthCall) {
      if (!isHandling401 && onUnauthorized) {
        isHandling401 = true
        try {
          // The router path is not available here; the handler resolves it
          // from the router itself.
          onUnauthorized('')
        } finally {
          isHandling401 = false
        }
      }
    }
    return Promise.reject(toApiError(error))
  },
)
