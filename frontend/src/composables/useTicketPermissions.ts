import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { Status, Ticket } from '@/lib/types'

export interface TicketPermissions {
  canClaim: boolean
  canClose: boolean
  canResolve: boolean
  canReopen: boolean
  canCancel: boolean
  canEditPriority: boolean
  canEditAssignee: boolean
  canCommentInternal: boolean
}

/**
 * UI-only helper: decides which action buttons to render. The backend always
 * re-validates; error responses must still be handled.
 */
export function useTicketPermissions(ticket: () => Ticket | null | undefined) {
  const auth = useAuthStore()

  const isOwner = computed(() => {
    const t = ticket()
    return Boolean(t && auth.user && t.requester && t.requester.id === auth.user.id)
  })

  const perms = computed<TicketPermissions>(() => {
    const t = ticket()
    if (!t) {
      return {
        canClaim: false,
        canClose: false,
        canResolve: false,
        canReopen: false,
        canCancel: false,
        canEditPriority: false,
        canEditAssignee: false,
        canCommentInternal: false,
      }
    }
    const role = auth.role
    const staff = role === 'agent' || role === 'admin'
    const status: Status = t.status

    const canClaim = status === 'open' && staff
    const canResolve = status === 'in_progress' && staff
    const canClose = (status === 'resolved' && (staff || isOwner.value)) ||
      (status === 'open' && (isOwner.value || role === 'admin'))
    const canReopen = status === 'resolved' && isOwner.value
    const canCancel = status === 'open' && isOwner.value
    const canEditPriority = staff
    const canEditAssignee = staff
    const canCommentInternal = staff

    return {
      canClaim,
      canClose,
      canResolve,
      canReopen,
      canCancel,
      canEditPriority,
      canEditAssignee,
      canCommentInternal,
    }
  })

  return perms
}
