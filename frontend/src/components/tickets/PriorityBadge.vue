<script setup lang="ts">
import AppBadge from '@/components/ui/AppBadge.vue'
import { Flame } from 'lucide-vue-next'
import type { Priority } from '@/lib/types'
import { priorityLabel } from '@/lib/format'

const props = defineProps<{ priority: Priority; size?: 'sm' | 'md' }>()

const colorFor: Record<Priority, 'slate' | 'blue' | 'orange' | 'red'> = {
  low: 'slate',
  medium: 'blue',
  high: 'orange',
  urgent: 'red',
}

const sizeClass = props.size === 'md' ? 'px-2.5 py-1 text-sm' : ''
</script>

<template>
  <AppBadge :color="colorFor[props.priority]" :class="sizeClass">
    <Flame v-if="props.priority === 'urgent'" class="h-3 w-3" aria-hidden="true" />
    {{ priorityLabel(props.priority) }}
  </AppBadge>
</template>
