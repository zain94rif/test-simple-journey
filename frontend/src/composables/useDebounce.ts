import { customRef } from 'vue'

/**
 * Returns a debounced ref: setting it schedules the value update after
 * `delayMs` of silence. Used for the search boxes.
 */
export function useDebounce<T>(initial: T, delayMs = 350) {
  let value = initial
  let timeout: ReturnType<typeof setTimeout> | undefined

  return customRef<T>((track, trigger) => ({
    get() {
      track()
      return value
    },
    set(newValue: T) {
      if (timeout) clearTimeout(timeout)
      timeout = setTimeout(() => {
        value = newValue
        trigger()
      }, delayMs)
    },
  }))
}
