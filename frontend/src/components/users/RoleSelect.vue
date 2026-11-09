<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/lib/types'
import AppSelect from '@/components/ui/AppSelect.vue'
import { ROLE_OPTIONS } from '@/lib/constants'

const props = defineProps<{
  user: User
  saving?: boolean
}>()

const emit = defineEmits<{ (e: 'change', role: string): void }>()

const auth = useAuthStore()

/** A user cannot change their own role (mirrors backend rule 409). */
const disabled = computed(
  () => props.saving === true || auth.user?.id === props.user.id,
)

const selfHint = computed(() =>
  auth.user?.id === props.user.id ? 'Anda tidak dapat mengubah peran sendiri.' : undefined,
)
</script>

<template>
  <div class="flex items-center gap-2">
    <AppSelect
      :model-value="props.user.role"
      :options="ROLE_OPTIONS"
      :disabled="disabled"
      :title="selfHint"
      class="w-40"
      aria-label="Peran pengguna"
      @update:model-value="(value: string) => emit('change', value)"
    />
  </div>
</template>
