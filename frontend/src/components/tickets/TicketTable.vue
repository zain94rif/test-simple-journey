<script setup lang="ts">
import { computed } from 'vue'
import type { Ticket } from '@/lib/types'
import StatusBadge from './StatusBadge.vue'
import PriorityBadge from './PriorityBadge.vue'
import { categoryLabel, relativeTime } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'

defineProps<{ tickets: Ticket[] }>()

const emit = defineEmits<{ (e: 'open', id: number): void }>()

const auth = useAuthStore()
const showRequesterAndAssignee = computed(() => auth.isStaff)
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-slate-200 bg-white">
    <table class="min-w-full divide-y divide-slate-200 text-sm">
      <thead class="bg-slate-50 text-left text-xs font-medium uppercase tracking-wide text-slate-500">
        <tr>
          <th class="px-4 py-3">ID</th>
          <th class="px-4 py-3">Judul</th>
          <th class="px-4 py-3">Status</th>
          <th class="px-4 py-3">Prioritas</th>
          <th class="px-4 py-3">Kategori</th>
          <th v-if="showRequesterAndAssignee" class="px-4 py-3">Requester</th>
          <th v-if="showRequesterAndAssignee" class="px-4 py-3">Assignee</th>
          <th class="px-4 py-3">Diperbarui</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        <tr
          v-for="ticket in tickets"
          :key="ticket.id"
          tabindex="0"
          class="cursor-pointer transition-colors hover:bg-indigo-50/50"
          :data-testid="`ticket-row-${ticket.id}`"
          @click="emit('open', ticket.id)"
          @keydown.enter="emit('open', ticket.id)"
        >
          <td class="px-4 py-3 font-mono text-xs text-slate-400">#{{ ticket.id }}</td>
          <td class="max-w-xs truncate px-4 py-3 font-medium text-slate-800">
            {{ ticket.title }}
          </td>
          <td class="px-4 py-3"><StatusBadge :status="ticket.status" /></td>
          <td class="px-4 py-3"><PriorityBadge :priority="ticket.priority" /></td>
          <td class="px-4 py-3 text-slate-600">{{ categoryLabel(ticket.category) }}</td>
          <td v-if="showRequesterAndAssignee" class="px-4 py-3 text-slate-600">
            {{ ticket.requester?.name ?? '—' }}
          </td>
          <td v-if="showRequesterAndAssignee" class="px-4 py-3 text-slate-600">
            {{ ticket.assignee?.name ?? 'Belum ditangani' }}
          </td>
          <td class="px-4 py-3 text-slate-500">{{ relativeTime(ticket.updated_at) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
