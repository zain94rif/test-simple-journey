<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Lock } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import StatusBadge from '@/components/tickets/StatusBadge.vue'
import PriorityBadge from '@/components/tickets/PriorityBadge.vue'
import TicketMeta from '@/components/tickets/TicketMeta.vue'
import TicketActions from '@/components/tickets/TicketActions.vue'
import TicketEditPanel from '@/components/tickets/TicketEditPanel.vue'
import CommentList from '@/components/tickets/CommentList.vue'
import CommentForm from '@/components/tickets/CommentForm.vue'
import { useTicketStore } from '@/stores/tickets'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'

const store = useTicketStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()

const ticketId = computed(() => String(route.params.id))
const ticket = computed(() => store.detail)
const transitioning = computed(() => store.mutationLoading)

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  void router.replace({ name: 'tickets' })
}

async function load() {
  store.resetDetail()
  try {
    await store.fetchTicket(ticketId.value)
  } catch {
    // detailError berisi pesan; ErrorState merendernya.
  }
}

onMounted(load)

function onUpdated() {
  void store.fetchTicket(ticketId.value).catch(() => undefined)
}

function onFailed(err: unknown) {
  const apiError = toApiError(err)
  if (apiError.code === 'INVALID_TRANSITION' || apiError.status === 409) {
    toast.error('Status di server sudah berubah. Data tiket diperbarui.')
    void store.fetchTicket(ticketId.value).catch(() => undefined)
  }
}
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-4">
    <AppButton variant="ghost" size="sm" @click="goBack">
      <ArrowLeft class="h-4 w-4" /> Kembali
    </AppButton>

    <template v-if="store.detailLoading">
      <div class="space-y-3 rounded-xl border border-slate-200 bg-white p-5">
        <SkeletonBlock height="28px" width="60%" />
        <SkeletonBlock height="20px" width="40%" />
        <SkeletonBlock height="96px" />
        <SkeletonBlock height="24px" width="30%" />
      </div>
    </template>

    <ErrorState
      v-else-if="store.detailError"
      :message="store.detailError"
      :details="
        store.detailError?.includes('tidak ditemukan') ? 'Kode 404' :
        store.detailError?.includes('tidak memiliki') ? 'Kode 403' : 'Kode error: ' + String(ticketId)"
      @retry="load"
    />

    <template v-else-if="ticket">
      <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div class="space-y-4 lg:col-span-2">
          <div class="rounded-xl border border-slate-200 bg-white p-5">
            <div class="mb-2 flex flex-wrap items-center gap-2">
              <span class="font-mono text-xs text-slate-400">#{{ ticket.id }}</span>
              <StatusBadge :status="ticket.status" />
              <PriorityBadge :priority="ticket.priority" />
              <span v-if="ticket.status === 'closed'" class="inline-flex items-center gap-1 text-xs text-slate-400">
                <Lock class="h-3 w-3" /> tidak dapat diubah
              </span>
            </div>
            <h1 class="text-lg font-bold text-slate-900">{{ ticket.title }}</h1>
            <p class="mt-2 whitespace-pre-wrap text-sm text-slate-700">{{ ticket.description }}</p>
          </div>

          <div class="rounded-xl border border-slate-200 bg-white p-5">
            <TicketMeta :ticket="ticket" />
          </div>

          <div class="rounded-xl border border-slate-200 bg-white p-5">
            <CommentList :comments="ticket.comments" :loading="store.detailLoading" />
            <div v-if="ticket.status !== 'closed'" class="mt-5 border-t border-slate-100 pt-5">
              <CommentForm :ticket="ticket" />
            </div>
            <p v-else class="mt-5 border-t border-slate-100 pt-4 text-center text-sm text-slate-400">
              Tiket ditutup. Komentar baru tidak diperlukan.
            </p>
          </div>
        </div>

        <div class="space-y-4">
          <div class="rounded-xl border border-slate-200 bg-white p-5">
            <h3 class="mb-3 text-sm font-semibold text-slate-800">Aksi</h3>
            <TicketActions
              v-if="ticket.status !== 'closed'"
              :ticket="ticket"
              :loading="transitioning"
              @updated="onUpdated"
              @failed="onFailed"
            />
            <p v-else class="text-sm text-slate-400">Tiket sudah ditutup.</p>
          </div>
          <div v-if="auth.isStaff" class="rounded-xl border border-slate-200 bg-white p-5">
            <TicketEditPanel :ticket="ticket" @updated="onUpdated" />
          </div>
          <div class="rounded-xl border border-slate-200 bg-white p-5">
            <h3 class="mb-2 text-sm font-semibold text-slate-800">Ringkasan</h3>
            <ul class="space-y-1 text-sm text-slate-600">
              <li>Requester: {{ ticket.requester?.name ?? '—' }}</li>
              <li>Assignee: {{ ticket.assignee?.name ?? 'Belum ditangani' }}</li>
              <li>Status: {{ ticket.status }}</li>
            </ul>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
