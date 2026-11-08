<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import { PER_PAGE_OPTIONS } from '@/lib/constants'

const props = defineProps<{
  page: number
  perPage: number
  total: number
  totalPages: number
}>()

const emit = defineEmits<{
  (e: 'update:page', page: number): void
  (e: 'update:perPage', perPage: number): void
}>()

const rangeText = computed(() => {
  if (props.total === 0) return '0 data'
  const from = (props.page - 1) * props.perPage + 1
  const to = Math.min(props.page * props.perPage, props.total)
  return `${from}–${to} dari ${props.total}`
})

function pageNumbers(): number[] {
  const current = props.page
  const last = props.totalPages
  const pages = new Set<number>([1, last, current - 1, current, current + 1])
  const sorted = [...pages].filter((p) => p >= 1 && p <= last).sort((a, b) => a - b)
  return sorted
}

function gapAfter(index: number, pages: number[]): boolean {
  if (index + 1 >= pages.length) return false
  return pages[index + 1] - pages[index] > 1
}
</script>

<template>
  <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <p class="text-sm text-slate-500">{{ rangeText }}</p>
    <div class="flex items-center gap-3">
      <label class="hidden items-center gap-2 text-sm text-slate-500 sm:flex">
        <select
          :value="props.perPage"
          class="rounded-lg border border-slate-300 bg-white px-2 py-1 text-sm"
          aria-label="Per halaman"
          @change="emit('update:perPage', Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="option in PER_PAGE_OPTIONS" :key="option" :value="option">{{ option }}</option>
        </select>
        / halaman
      </label>
      <nav class="flex items-center gap-1" aria-label="Navigasi halaman">
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="props.page <= 1"
          aria-label="Halaman sebelumnya"
          @click="emit('update:page', props.page - 1)"
        >
          <ChevronLeft class="h-4 w-4" />
        </button>
        <template v-for="(p, i) in pageNumbers()" :key="p">
          <span v-if="gapAfter(i, pageNumbers())" class="px-1 text-slate-400" aria-hidden="true">…</span>
          <button
            type="button"
            class="min-w-[2rem] rounded-lg border px-2 py-1 text-sm"
            :class="
              p === props.page
                ? 'border-indigo-600 bg-indigo-600 text-white'
                : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-100'
            "
            :aria-current="p === props.page ? 'page' : undefined"
            @click="emit('update:page', p)"
          >
            {{ p }}
          </button>
        </template>
        <button
          type="button"
          class="rounded-lg border border-slate-200 bg-white p-1.5 text-slate-600 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
          :disabled="props.page >= props.totalPages"
          aria-label="Halaman berikutnya"
          @click="emit('update:page', props.page + 1)"
        >
          <ChevronRight class="h-4 w-4" />
        </button>
      </nav>
    </div>
  </div>
</template>
