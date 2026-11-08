<script setup lang="ts">
import { LogIn, Lock, CheckCircle2, RotateCcw, XCircle } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useToast } from '@/composables/useToast'
import { useTicketStore } from '@/stores/tickets'
import { useTicketPermissions } from '@/composables/useTicketPermissions'
import { toApiError } from '@/lib/errors'
import type { Ticket } from '@/lib/types'

const props = defineProps<{
  ticket: Ticket
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'updated', ticket: Ticket): void
  (e: 'failed', err: unknown): void
}>()

const perms = useTicketPermissions(() => props.ticket)
const confirm = useConfirm()
const toast = useToast()
const store = useTicketStore()

async function transition(patch: { status: string }, title: string, message: string, danger = false) {
  const ok = await confirm.confirm({ title, message, confirmLabel: title, danger })
  if (!ok) return
  try {
    const updated = await store.updateTicket(props.ticket.id, patch)
    toast.success('Status tiket diperbarui.')
    emit('updated', updated)
  } catch (err) {
    const apiError = toApiError(err)
    toast.error(apiError.message)
    emit('failed', err)
  }
}

function claim() {
  transition({ status: 'in_progress' }, 'Ambil tiket', 'Tiket akan dipindahkan ke In Progress dan menjadi tanggung jawab Anda. Lanjutkan?')
}
function resolve() {
  transition({ status: 'resolved' }, 'Tandai selesai', 'Tiket akan ditandai Resolved. Lanjutkan?')
}
function close() {
  transition({ status: 'closed' }, 'Tutup tiket', 'Tiket akan ditutup secara permanen dan tidak dapat dibuka kembali. Lanjutkan?', true)
}
function reopen() {
  transition({ status: 'open' }, 'Buka kembali', 'Tiket akan dibuka kembali ke status Open. Lanjutkan?')
}
function cancel() {
  transition({ status: 'closed' }, 'Batalkan tiket', 'Tiket akan dibatalkan (ditutup). Lanjutkan?', true)
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-2">
    <AppButton v-if="perms.canClaim" size="sm" :loading="loading" @click="claim">
      <LogIn class="h-4 w-4" /> Ambil tiket
    </AppButton>
    <AppButton v-if="perms.canResolve" size="sm" :loading="loading" @click="resolve">
      <CheckCircle2 class="h-4 w-4" /> Tandai selesai
    </AppButton>
    <AppButton
      v-if="perms.canClose && ticket.status !== 'open'"
      size="sm"
      variant="secondary"
      :loading="loading"
      @click="close"
    >
      <Lock class="h-4 w-4" /> Tutup tiket
    </AppButton>
    <AppButton v-if="perms.canReopen" size="sm" variant="secondary" :loading="loading" @click="reopen">
      <RotateCcw class="h-4 w-4" /> Buka kembali
    </AppButton>
    <AppButton
      v-if="ticket.status === 'open' && perms.canCancel"
      size="sm"
      variant="secondary"
      :loading="loading"
      @click="cancel"
    >
      <XCircle class="h-4 w-4" /> Batalkan tiket
    </AppButton>
    <AppButton v-else-if="perms.canClose && ticket.status === 'open'" size="sm" variant="secondary" :loading="loading" @click="close">
      <Lock class="h-4 w-4" /> Tutup tiket
    </AppButton>
  </div>
</template>
