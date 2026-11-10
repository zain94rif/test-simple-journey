<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import TicketForm from '@/components/tickets/TicketForm.vue'
import type { TicketFormErrors as FormErrors } from '@/lib/types'
import { useTicketStore } from '@/stores/tickets'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'

const router = useRouter()
const store = useTicketStore()
const toast = useToast()

const errors = ref<FormErrors>({})
const serverError = ref<string | null>(null)

async function onSubmit(payload: { title: string; description: string; category: string; priority: string }) {
  errors.value = {}
  serverError.value = null
  try {
    const ticket = await store.createTicket(payload)
    toast.success('Tiket berhasil dibuat.')
    await router.replace({ name: 'ticket-detail', params: { id: String(ticket.id) } })
  } catch (err) {
    const apiError = toApiError(err)
    if (apiError.details) {
      for (const [field, message] of Object.entries(apiError.details)) {
        if (field === 'title' || field === 'description' || field === 'category' || field === 'priority') {
          errors.value[field] = message
        } else {
          serverError.value = message
        }
      }
      if (!serverError.value && Object.keys(apiError.details).length === 0) {
        serverError.value = apiError.message
      }
    } else {
      serverError.value = apiError.message
      toast.error(apiError.message)
    }
  }
}

function onCancel() {
  void router.back()
}
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <h1 class="mb-1 text-xl font-bold text-slate-900">Buat Tiket</h1>
    <p class="mb-4 text-sm text-slate-500">Jelaskan masalah agar mudah ditindaklanjuti.</p>
    <div class="rounded-xl border border-slate-200 bg-white p-5">
      <TicketForm
        :errors="errors"
        :server-error="serverError"
        :loading="store.mutationLoading"
        @submit="onSubmit"
        @cancel="onCancel"
      />
    </div>
  </div>
</template>
