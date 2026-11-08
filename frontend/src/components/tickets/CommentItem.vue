<script setup lang="ts">
import AppBadge from '@/components/ui/AppBadge.vue'
import { Lock } from 'lucide-vue-next'
import type { Comment } from '@/lib/types'
import { formatDateTime, roleLabel } from '@/lib/format'

defineProps<{ comment: Comment }>()
</script>

<template>
  <li
    :class="[
      'rounded-xl border p-4',
      comment.is_internal ? 'border-amber-300 bg-amber-50' : 'border-slate-200 bg-white',
    ]"
  >
    <div class="mb-2 flex flex-wrap items-center gap-2">
      <span class="text-sm font-semibold text-slate-800">{{ comment.author?.name ?? 'Anonim' }}</span>
      <AppBadge color="indigo">{{ roleLabel(comment.author?.role ?? '') }}</AppBadge>
      <span
        v-if="comment.is_internal"
        class="inline-flex items-center gap-1 rounded-full bg-amber-200/60 px-2 py-0.5 text-xs font-medium text-amber-800"
      >
        <Lock class="h-3 w-3" aria-hidden="true" /> Catatan internal
      </span>
      <span class="ml-auto text-xs text-slate-400">{{ formatDateTime(comment.created_at) }}</span>
    </div>
    <p class="whitespace-pre-wrap text-sm text-slate-700">{{ comment.body }}</p>
  </li>
</template>
