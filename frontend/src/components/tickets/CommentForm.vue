<script setup lang="ts">
import { computed, ref } from 'vue'
import { Lock } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'
import AppTextarea from '@/components/ui/AppTextarea.vue'
import { useAuthStore } from '@/stores/auth'
import { useTicketStore } from '@/stores/tickets'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'
import type { Ticket } from '@/lib/types'

const props = defineProps<{ ticket: Ticket }>()
const emit = defineEmits<{ (e: 'commented'): void }>()

const auth = useAuthStore()
const store = useTicketStore()
const toast = useToast()

const body = ref('')
const isInternal = ref(false)
const error = ref<string | null>(null)
const sending = computed(() => store.mutationLoading)

const canToggleInternal = auth.isStaff

async function submit() {
  error.value = null
  if (body.value.trim().length === 0) {
    error.value = 'Komentar tidak boleh kosong.'
    return
  }
  try {
    await store.addComment(props.ticket.id, body.value.trim(), isInternal.value && canToggleInternal)
    body.value = ''
    isInternal.value = false
    toast.success('Komentar terkirim.')
    emit('commented')
  } catch (err) {
    const apiError = toApiError(err)
    if (apiError.details?.body) {
      error.value = apiError.details.body
    } else {
      error.value = apiError.message
      toast.error(apiError.message)
    }
  }
}
</script>

<template>
  <form class="space-y-2" novalidate @submit.prevent="submit">
    <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
      {{ error }}
    </div>
    <AppTextarea
      v-model="body"
      label="Tulis komentar"
      :rows="3"
      :error="error"
      :disabled="sending"
    />
    <div class="flex flex-wrap items-center justify-between gap-3">
      <label
        v-if="canToggleInternal"
        class="inline-flex cursor-pointer items-center gap-2 text-sm text-slate-600"
      >
        <input v-model="isInternal" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-indigo-600" />
        <Lock class="h-4 w-4 text-amber-600" aria-hidden="true" />
        Catatan internal
      </label>
      <AppButton type="submit" :loading="sending">Kirim</AppButton>
    </div>
  </form>
</template>
