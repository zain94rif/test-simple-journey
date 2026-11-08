<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { Ticket } from '@/lib/types'
import StatusBadge from './StatusBadge.vue'
import PriorityBadge from './PriorityBadge.vue'
import { categoryLabel, relativeTime } from '@/lib/format'

defineProps<{ tickets: Ticket[] }>()

const router = useRouter()

function slug(id: number) {
  return String(id)
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <button
      v-for="ticket in tickets"
      :key="ticket.id"
      type="button"
      class="rounded-xl border border-slate-200 bg-white p-4 text-left shadow-sm transition-colors hover:border-indigo-300"
      @click="router.push({ name: 'ticket-detail', params: { id: slug(ticket.id) } })"
    >
      <div class="mb-2 flex items-start justify-between gap-3">
        <h3 class="font-medium text-slate-900">{{ ticket.title }}</h3>
        <span class="shrink-0 font-mono text-xs text-slate-400">#{{ ticket.id }}</span>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <StatusBadge :status="ticket.status" />
        <PriorityBadge :priority="ticket.priority" />
        <span class="text-xs text-slate-500">{{ categoryLabel(ticket.category) }}</span>
      </div>
      <p v-if="ticket.assignee" class="mt-2 text-xs text-slate-500">
        Ditangani oleh {{ ticket.assignee.name }} · {{ relativeTime(ticket.updated_at) }}
      </p>
      <p v-else class="mt-2 text-xs text-slate-400">Belum ditangani · {{ relativeTime(ticket.updated_at) }}</p>
    </button>
  </div>
</template>
