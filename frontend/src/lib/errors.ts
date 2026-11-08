export interface ApiError {
  status: number
  code: string
  message: string
  details?: Record<string, string>
}

/**
 * Normalizes any thrown value into an ApiError. Components never read
 * err.response directly; they always use this helper.
 */
export function toApiError(err: unknown): ApiError {
  if (isApiError(err)) return err

  if (isAxiosLike(err)) {
    const data = err.response?.data as { error?: { code?: string; message?: string; details?: Record<string, string> } } | undefined
    if (data?.error) {
      return {
        status: err.response?.status ?? 500,
        code: data.error.code ?? 'INTERNAL_ERROR',
        message: data.error.message ?? 'Terjadi kesalahan pada server.',
        details: data.error.details,
      }
    }
    if (err.code === 'ECONNABORTED') {
      return { status: 0, code: 'TIMEOUT', message: 'Permintaan terlalu lama. Coba lagi.' }
    }
    return { status: 0, code: 'NETWORK_ERROR', message: 'Tidak dapat terhubung ke server. Periksa koneksi Anda.' }
  }

  return { status: 0, code: 'UNKNOWN_ERROR', message: 'Terjadi kesalahan yang tidak diketahui.' }
}

export function createApiError(partial: Partial<ApiError> & { message: string }): ApiError {
  return {
    status: partial.status ?? 0,
    code: partial.code ?? 'UNKNOWN_ERROR',
    message: partial.message,
    details: partial.details,
  }
}

interface ApiErrorExtras {
  status?: number
  code?: string
  details?: Record<string, string>
}

export function isApiError(err: unknown): err is ApiError & ApiErrorExtras {
  return (
    typeof err === 'object' &&
    err !== null &&
    'code' in err &&
    'message' in err &&
    typeof (err as ApiError).message === 'string' &&
    typeof (err as ApiError).code === 'string'
  )
}

interface AxiosLike {
  response?: {
    status?: number
    data?: unknown
  }
  code?: string
  config?: unknown
}

export function isAxiosLike(err: unknown): err is AxiosLike {
  return typeof err === 'object' && err !== null && ('response' in err || 'code' in err)
}
