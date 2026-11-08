import { defineStore } from 'pinia'

export interface ConfirmRequest {
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

interface Pending extends ConfirmRequest {
  resolver: (confirmed: boolean) => void
}

export const useConfirmStore = defineStore('confirm', {
  state: () => ({
    pending: null as Pending | null,
  }),
  actions: {
    open(request: ConfirmRequest): Promise<boolean> {
      return new Promise((resolve) => {
        // A new dialog replaces the previous unresolved one.
        this.pending?.resolver(false)
        this.pending = { ...request, resolver: resolve }
      })
    },
    answer(confirmed: boolean) {
      const current = this.pending
      this.pending = null
      current?.resolver(confirmed)
    },
  },
})

export function useConfirm() {
  const store = useConfirmStore()
  return {
    confirm: (request: ConfirmRequest) => store.open(request),
    state: store,
  }
}
