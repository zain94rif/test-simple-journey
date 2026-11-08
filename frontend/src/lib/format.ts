import { CATEGORY_LABELS, PRIORITY_LABELS, STATUS_LABELS } from './constants'
import type { Category, Priority, Role, Status } from './types'

export function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: 'Asia/Jakarta',
  }).format(date)
}

export function relativeTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const diffMs = Date.now() - date.getTime()
  const minutes = Math.round(diffMs / 60_000)
  if (minutes < 1) return 'baru saja'
  if (minutes < 60) return `${minutes} menit lalu`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours} jam lalu`
  const days = Math.round(hours / 24)
  if (days < 7) return `${days} hari lalu`
  return formatDateTime(value)
}

export function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

export function statusLabel(status: Status): string {
  return STATUS_LABELS[status] ?? status
}

export function priorityLabel(priority: Priority): string {
  return PRIORITY_LABELS[priority] ?? priority
}

export function categoryLabel(category: Category): string {
  return CATEGORY_LABELS[category] ?? category
}

export function roleLabel(role: Role | string): string {
  return role.charAt(0).toUpperCase() + role.slice(1)
}

/** Only internal paths are accepted: must start with "/" and not "//". */
export function safeRedirectPath(raw: unknown): string {
  if (typeof raw !== 'string') return '/tickets'
  if (!raw.startsWith('/') || raw.startsWith('//')) return '/tickets'
  return raw
}
