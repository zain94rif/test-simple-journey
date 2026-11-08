import type { Category, Priority, Status } from './types'

export const APP_NAME = (import.meta.env.VITE_APP_NAME as string | undefined) ?? 'Helpdesk'

export const TOKEN_KEY = 'helpdesk_token'

export const STATUS_OPTIONS = [
  { value: 'open', label: 'Open' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'resolved', label: 'Resolved' },
  { value: 'closed', label: 'Closed' },
] as const

export const PRIORITY_OPTIONS = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
  { value: 'urgent', label: 'Urgent' },
] as const

export const CATEGORY_OPTIONS = [
  { value: 'hardware', label: 'Hardware' },
  { value: 'software', label: 'Software' },
  { value: 'network', label: 'Network' },
  { value: 'access', label: 'Access' },
  { value: 'other', label: 'Other' },
] as const

export const ROLE_OPTIONS = [
  { value: 'requester', label: 'Requester' },
  { value: 'agent', label: 'Agent' },
  { value: 'admin', label: 'Admin' },
] as const

export const SORT_OPTIONS = [
  { value: 'created_at:desc', label: 'Terbaru' },
  { value: 'created_at:asc', label: 'Terlama' },
  { value: 'priority:desc', label: 'Prioritas tertinggi' },
  { value: 'updated_at:desc', label: 'Terakhir diperbarui' },
] as const

export const PER_PAGE_OPTIONS = [10, 20, 50] as const

export const STATUS_LABELS: Record<Status, string> = {
  open: 'Open',
  in_progress: 'In Progress',
  resolved: 'Resolved',
  closed: 'Closed',
}

export const PRIORITY_LABELS: Record<Priority, string> = {
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  urgent: 'Urgent',
}

export const CATEGORY_LABELS: Record<Category, string> = {
  hardware: 'Hardware',
  software: 'Software',
  network: 'Network',
  access: 'Access',
  other: 'Other',
}

export const ROLE_LABELS: Record<string, string> = {
  requester: 'Requester',
  agent: 'Agent',
  admin: 'Admin',
}
