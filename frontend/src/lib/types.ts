export type Role = 'requester' | 'agent' | 'admin'
export type Status = 'open' | 'in_progress' | 'resolved' | 'closed'
export type Priority = 'low' | 'medium' | 'high' | 'urgent'
export type Category = 'hardware' | 'software' | 'network' | 'access' | 'other'

export interface User {
  id: number
  name: string
  email: string
  role: Role
  created_at: string
}

export interface TicketUser {
  id: number
  name: string
  email: string
}

export interface Ticket {
  id: number
  title: string
  description: string
  category: Category
  priority: Priority
  status: Status
  requester: TicketUser | null
  assignee: TicketUser | null
  created_at: string
  updated_at: string
}

export interface Comment {
  id: number
  ticket_id: number
  author: {
    id: number
    name: string
    role: Role
  }
  body: string
  is_internal: boolean
  created_at: string
}

export interface TicketDetail extends Ticket {
  comments: Comment[]
}

export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface ListResponse<T> {
  data: T[]
  meta?: PaginationMeta
}

export interface SingleResponse<T> {
  data: T
}

export interface LoginResponse {
  token: string
  token_type: string
  expires_at: string
  user: User
}

export interface TicketFilters {
  status: string
  priority: string
  assignee: string
  unassigned: boolean
  search: string
  sort: string
  order: string
  page: number
  per_page: number
}

export interface UserFilters {
  role: string
  search: string
  page: number
  per_page: number
}

export interface AdminRolePayload {
  role: string
}
