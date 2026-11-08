import { defineStore } from 'pinia'

export type ToastType = 'success' | 'error' | 'info'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
}

let nextId = 1

export const useToastStore = defineStore('toast', {
  state: () => ({
    toasts: [] as ToastItem[],
  }),
  actions: {
    push(type: ToastType, message: string) {
      const id = nextId++
      this.toasts.push({ id, type, message })
      window.setTimeout(() => this.dismiss(id), 4000)
      return id
    },
    success(message: string) {
      return this.push('success', message)
    },
    error(message: string) {
      return this.push('error', message)
    },
    info(message: string) {
      return this.push('info', message)
    },
    dismiss(id: number) {
      this.toasts = this.toasts.filter((t) => t.id !== id)
    },
  },
})
