<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    color?: 'blue' | 'amber' | 'emerald' | 'slate' | 'red' | 'orange' | 'indigo'
  }>(),
  { color: 'slate' },
)

const colorClass = {
  blue: 'bg-blue-100 text-blue-700',
  amber: 'bg-amber-100 text-amber-700',
  emerald: 'bg-emerald-100 text-emerald-700',
  slate: 'bg-slate-100 text-slate-600',
  red: 'bg-red-100 text-red-700',
  orange: 'bg-orange-100 text-orange-700',
  indigo: 'bg-indigo-100 text-indigo-700',
}[props.color]
</script>

<template>
  <span
    :class="['inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium', colorClass]"
  >
    <slot />
  </span>
</template>
