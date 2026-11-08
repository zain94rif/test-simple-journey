<script setup lang="ts">
import { computed } from 'vue'
import type { User } from '@/lib/types'
import AppSelect from '@/components/ui/AppSelect.vue'
import { ROLE_LABELS, ROLE_OPTIONS } from '@/lib/constants'

const props = defineProps<{
  user: User
  disabled?: boolean
}>()

const emit = defineEmits<{ (e: 'change', role: string): void }>()

const groups = computed(() => ({
  label: ROLE_LABELS[props.user.role] ?? props.user.role,
  options: ROLE_OPTIONS,
}))
</script>

<template>
  <div class="flex items-center gap-2">
    <AppSelect
      :model-value="props.user.role"
      :options="groups.options"
      :disabled="props.disabled"
      aria-label="Peran pengguna"
      @update:model-value="(value: string) => emit('change', value)"
    />
  </div>
</template>
