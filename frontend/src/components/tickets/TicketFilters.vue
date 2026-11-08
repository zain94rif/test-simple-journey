<script setup lang="ts">
import { Search, X } from 'lucide-vue-next'
import { useDebounce } from '@/composables/useDebounce'
import { watch } from 'vue'
import { PRIORITY_OPTIONS, SORT_OPTIONS, STATUS_OPTIONS } from '@/lib/constants'
import AppSelect from '@/components/ui/AppSelect.vue'
import type { TicketFilters } from '@/lib/types'

const props = withDefaults(
  defineProps<{
    filters: TicketFilters
    showAssignee?: boolean
    isAgent?: boolean
    active: boolean
  }>(),
  { showAssignee: false, isAgent: false },
)

const emit = defineEmits<{
  (e: 'update:filter', patch: Partial<TicketFilters>): void
  (e: 'reset'): void
}>()

const search = useDebounce(props.filters.search, 350)
// Keep the debounced search in sync with external resets/URL restore.
watch(
  () => props.filters.search,
  (value) => {
    if (value !== search.value && value === '') search.value = ''
  },
)
watch(search, (value) => {
  if (value !== props.filters.search) emit('update:filter', { search: value })
})

const assigneeOptions: { value: string; label: string }[] = props.isAgent
  ? [
      { value: '', label: 'Semua' },
      { value: 'unassigned', label: 'Belum di-assign' },
      { value: 'me', label: 'Saya' },
    ]
  : [
      { value: '', label: 'Semua' },
      { value: 'unassigned', label: 'Belum di-assign' },
    ]
</script>

<template>
  <div class="rounded-xl border border-slate-200 bg-white p-3 sm:p-4">
    <div class="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-5">
      <div class="lg:col-span-2">
        <label class="mb-1 block text-sm font-medium text-slate-700">Cari judul</label>
        <div class="relative">
          <Search class="absolute left-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <input
            v-model="search"
            type="search"
            placeholder="Cari tiket…"
            aria-label="Cari tiket berdasarkan judul"
            class="block w-full rounded-lg border border-slate-300 bg-white py-2 pl-9 pr-8 text-sm shadow-sm"
          />
          <button
            v-if="search"
            type="button"
            class="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-slate-400 hover:text-slate-600"
            aria-label="Hapus pencarian"
            @click="search = ''"
          >
            <X class="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700">Status</label>
        <AppSelect
          :model-value="props.filters.status"
          :options="STATUS_OPTIONS"
          @update:model-value="(value: string) => emit('update:filter', { status: value })"
        >
          <option value="">Semua</option>
        </AppSelect>
      </div>
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700">Prioritas</label>
        <AppSelect
          :model-value="props.filters.priority"
          :options="PRIORITY_OPTIONS"
          @update:model-value="(value: string) => emit('update:filter', { priority: value })"
        >
          <option value="">Semua</option>
        </AppSelect>
      </div>
      <div v-if="props.showAssignee">
        <label class="mb-1 block text-sm font-medium text-slate-700">Assignee</label>
        <AppSelect
          :model-value="props.filters.assignee"
          :options="assigneeOptions"
          @update:model-value="(value: string) => emit('update:filter', { assignee: value })"
        />
      </div>
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700">Urutan</label>
        <AppSelect
          :model-value="props.filters.sort"
          :options="SORT_OPTIONS"
          @update:model-value="(value: string) => emit('update:filter', { sort: value, order: value.split(':')[1] })"
        />
      </div>
    </div>
    <div v-if="props.active" class="mt-3 flex justify-end">
      <button
        type="button"
        class="text-sm font-medium text-indigo-600 hover:text-indigo-700"
        @click="((search = ''), emit('reset'))"
      >
        Reset filter
      </button>
    </div>
  </div>
</template>
