<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'
import AppButton from './AppButton.vue'

defineProps<{
  message: string
  details?: string
}>()

const emit = defineEmits<{ (e: 'retry'): void }>()
</script>

<template>
  <div class="flex flex-col items-center justify-center rounded-xl border border-red-200 bg-red-50 px-6 py-12 text-center">
    <AlertTriangle class="mb-3 h-8 w-8 text-red-500" />
    <h3 class="text-sm font-semibold text-red-800">Terjadi kesalahan</h3>
    <p class="mt-1 max-w-md text-sm text-red-700">{{ message }}</p>
    <p v-if="details" class="mt-0.5 text-xs text-red-500">{{ details }}</p>
    <AppButton variant="secondary" size="sm" class="mt-4" @click="emit('retry')">
      Coba lagi
    </AppButton>
  </div>
</template>
