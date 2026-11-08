<script setup lang="ts">
import { User, UserPlus } from 'lucide-vue-next'
import type { Ticket } from '@/lib/types'
import { categoryLabel, formatDateTime } from '@/lib/format'

defineProps<{ ticket: Ticket }>()
</script>

<template>
  <dl class="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
    <div class="flex items-start gap-2">
      <User class="mt-0.5 h-4 w-4 shrink-0 text-slate-400" aria-hidden="true" />
      <div>
        <dt class="text-xs font-medium uppercase tracking-wide text-slate-400">Requester</dt>
        <dd class="text-slate-700">{{ ticket.requester?.name ?? '—' }}</dd>
      </div>
    </div>
    <div class="flex items-start gap-2">
      <UserPlus class="mt-0.5 h-4 w-4 shrink-0 text-slate-400" aria-hidden="true" />
      <div>
        <dt class="text-xs font-medium uppercase tracking-wide text-slate-400">Assignee</dt>
        <dd class="text-slate-700">{{ ticket.assignee?.name ?? 'Belum ditangani' }}</dd>
      </div>
    </div>
    <div>
      <dt class="text-xs font-medium uppercase tracking-wide text-slate-400">Kategori</dt>
      <dd class="text-slate-700">{{ categoryLabel(ticket.category) }}</dd>
    </div>
    <div>
      <dt class="text-xs font-medium uppercase tracking-wide text-slate-400">Dibuat</dt>
      <dd class="text-slate-700">{{ formatDateTime(ticket.created_at) }}</dd>
    </div>
    <div>
      <dt class="text-xs font-medium uppercase tracking-wide text-slate-400">Diperbarui</dt>
      <dd class="text-slate-700">{{ formatDateTime(ticket.updated_at) }}</dd>
    </div>
  </dl>
</template>
