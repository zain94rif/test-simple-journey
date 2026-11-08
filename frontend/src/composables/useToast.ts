import { useToastStore } from '@/stores/toast'
import { toApiError } from '@/lib/errors'

/** Thin wrapper so components talk to a composable API, not the raw store. */
export function useToast() {
  const store = useToastStore()
  return {
    success: (message: string) => store.success(message),
    error: (err: unknown) => store.error(toApiError(err).message),
    errorText: (message: string) => store.error(message),
    info: (message: string) => store.info(message),
  }
}
