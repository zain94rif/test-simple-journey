<script setup lang="ts">
import AppBadge from '@/components/ui/AppBadge.vue'
import type { Status } from '@/lib/types'
import { statusLabel } from '@/lib/format'

const props = defineProps<{ status: Status; size?: 'sm' | 'md' }>()

const colorFor: Record<Status, 'blue' | 'amber' | 'emerald' | 'slate'> = {
  open: 'blue',
  in_progress: 'amber',
  resolved: 'emerald',
  closed: 'slate',
}

const sizeClass = props.size === 'md' ? 'px-2.5 py-1 text-sm' : ''
</script>

<template>
  <AppBadge :color="colorFor[props.status]" :class="sizeClass">
    {{ statusLabel(props.status) }}
  </AppBadge>
</template>
