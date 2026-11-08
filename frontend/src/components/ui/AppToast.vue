<script setup lang="ts">
import { useToastStore } from '@/stores/toast'
import { CheckCircle2, Info, XCircle, X } from 'lucide-vue-next'

const store = useToastStore()

const icons = {
  success: CheckCircle2,
  error: XCircle,
  info: Info,
}

const colors = {
  success: 'text-emerald-600',
  error: 'text-red-600',
  info: 'text-indigo-600',
}
</script>

<template>
  <div
    class="pointer-events-none fixed inset-x-0 top-4 z-[60] flex flex-col items-center gap-2 px-4"
    aria-live="polite"
  >
    <transition-group name="toast">
      <div
        v-for="toast in store.toasts"
        :key="toast.id"
        class="pointer-events-auto flex w-full max-w-md items-start gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3 shadow-lg"
        role="status"
      >
        <component :is="icons[toast.type]" :class="['mt-0.5 h-5 w-5 shrink-0', colors[toast.type]]" />
        <p class="flex-1 text-sm text-slate-700">{{ toast.message }}</p>
        <button
          type="button"
          class="rounded p-0.5 text-slate-400 hover:text-slate-600"
          aria-label="Tutup notifikasi"
          @click="store.dismiss(toast.id)"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
    </transition-group>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.2s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
