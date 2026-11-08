<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import { PRIORITY_OPTIONS } from '@/lib/constants'
import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/users'
import { useTicketStore, type UpdateTicketPayload } from '@/stores/tickets'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'
import type { Ticket } from '@/lib/types'

const props = defineProps<{ ticket: Ticket }>()
const emit = defineEmits<{ (e: 'updated', ticket: Ticket): void }>()

const auth = useAuthStore()
const userStore = useUserStore()
const ticketStore = useTicketStore()
const toast = useToast()

const priority = ref(props.ticket.priority)
const assigneeChoice = ref<string>(props.ticket.assignee ? String(props.ticket.assignee.id) : 'none')
const saving = ref(false)
const error = ref<string | null>(null)

watch(
  () => props.ticket,
  (t) => {
    priority.value = t.priority
    assigneeChoice.value = t.assignee ? String(t.assignee.id) : 'none'
  },
)

// Admin sees the real staff list; agents only get "me"/"none".
const staffOptions = computed(() => {
  if (!auth.isAdmin) {
    return [
      { value: 'none', label: 'Tidak ada' },
      ...(auth.user ? [{ value: String(auth.user.id), label: 'Saya' }] : []),
    ]
  }
  const options: { value: string; label: string }[] = [{ value: 'none', label: 'Tidak ada' }]
  for (const user of userStore.items) {
    if (user.role === 'agent' || user.role === 'admin') {
      options.push({ value: String(user.id), label: `${user.name} (${user.role})` })
    }
  }
  return options
})

async function loadStaff() {
  if (!auth.isAdmin || userStore.items.length > 0) return
  userStore.setFilters({ role: '', per_page: 100, page: 1, search: '' })
  try {
    await userStore.fetchUsers()
  } catch (err) {
    toast.error(toApiError(err).message)
  }
}

loadStaff()

async function save() {
  error.value = null
  const patch: UpdateTicketPayload = { priority: priority.value }
  if (assigneeChoice.value === 'none') {
    patch.assignee_id = null
  } else {
    const id = Number(assigneeChoice.value)
    if (Number.isFinite(id)) patch.assignee_id = id
  }

  saving.value = true
  try {
    const updated = await ticketStore.updateTicket(props.ticket.id, patch)
    toast.success('Perubahan disimpan.')
    emit('updated', updated)
  } catch (err) {
    error.value = toApiError(err).message
    toast.error(error.value)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <h3 class="text-sm font-semibold text-slate-800">Penanganan</h3>
    <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
      {{ error }}
    </div>
    <AppSelect v-model="priority" label="Prioritas" :options="PRIORITY_OPTIONS" :disabled="saving" />
    <AppSelect v-model="assigneeChoice" label="Assignee" :options="staffOptions" :disabled="saving || !auth.isAdmin" />
    <p v-if="!auth.isAdmin" class="text-xs text-slate-500">
      Agent hanya dapat menugaskan tiket kepada dirinya sendiri.
    </p>
    <AppButton size="sm" :loading="saving" @click="save">Simpan perubahan</AppButton>
  </div>
</template>
