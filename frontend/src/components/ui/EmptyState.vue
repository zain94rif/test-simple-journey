<script setup lang="ts">
import { computed, useSlots } from 'vue'
import { Inbox } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    title: string
    description?: string
    /** "none": nothing exists yet; "filter": current filters exclude everything. */
    kind?: 'none' | 'filter'
  }>(),
  {
    description: undefined,
    kind: 'none',
  },
)

const slots = useSlots()
const hasAction = computed(() => Boolean(slots.action))
const icon = computed(() => (props.kind === 'filter' ? Inbox : Inbox))
</script>

<template>
  <div class="flex flex-col items-center justify-center rounded-xl border border-slate-200 bg-white px-6 py-16 text-center">
    <component :is="icon" class="mb-3 h-10 w-10 text-slate-300" />
    <h3 class="text-sm font-semibold text-slate-800">{{ props.title }}</h3>
    <p v-if="props.description" class="mt-1 max-w-sm text-sm text-slate-500">
      {{ props.description }}
    </p>
    <div v-if="hasAction" class="mt-4 flex gap-2">
      <slot name="action" />
    </div>
  </div>
</template>
