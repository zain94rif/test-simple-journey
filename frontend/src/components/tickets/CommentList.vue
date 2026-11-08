<script setup lang="ts">
import type { Comment } from '@/lib/types'
import CommentItem from './CommentItem.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'

defineProps<{
  comments: Comment[]
  loading?: boolean
}>()
</script>

<template>
  <section aria-label="Komentar tiket">
    <h3 class="mb-3 text-sm font-semibold text-slate-800">
      Komentar <span class="font-normal text-slate-400">({{ comments.length }})</span>
    </h3>
    <div v-if="loading" class="space-y-3">
      <SkeletonBlock v-for="i in 3" :key="i" height="72px" />
    </div>
    <ul v-else-if="comments.length > 0" class="space-y-3">
      <CommentItem v-for="comment in comments" :key="comment.id" :comment="comment" />
    </ul>
    <p v-else class="rounded-xl border border-dashed border-slate-300 px-4 py-8 text-center text-sm text-slate-500">
      Belum ada komentar. Mulai diskusi.
    </p>
  </section>
</template>
