<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { LogOut, Menu } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { initials, roleLabel } from '@/lib/format'
import AppBadge from '@/components/ui/AppBadge.vue'

const props = defineProps<{
  showMenuButton?: boolean
}>()

const emit = defineEmits<{ (e: 'toggle-menu'): void }>()

const auth = useAuthStore()
const router = useRouter()
const toast = useToast()
const menuOpen = ref(false)

async function logout() {
  auth.logout()
  menuOpen.value = false
  toast.info('Anda telah keluar.')
  await router.replace({ name: 'login' })
}
</script>

<template>
  <header class="sticky top-0 z-30 border-b border-slate-200 bg-white">
    <div class="flex items-center gap-3 px-4 py-3">
      <button
        v-if="props.showMenuButton"
        type="button"
        class="rounded-lg p-2 text-slate-600 hover:bg-slate-100 lg:hidden"
        aria-label="Buka menu"
        @click="emit('toggle-menu')"
      >
        <Menu class="h-5 w-5" />
      </button>
      <div class="flex-1"></div>
      <div class="relative">
        <button
          type="button"
          class="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm hover:bg-slate-100"
          aria-haspopup="menu"
          :aria-expanded="menuOpen"
          @click="menuOpen = !menuOpen"
        >
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-100 text-xs font-semibold text-indigo-700">
            {{ initials(auth.displayName || 'U') }}
          </span>
          <span class="hidden text-left sm:block">
            <span class="block font-medium text-slate-800">{{ auth.displayName }}</span>
            <span class="block text-xs text-slate-500">{{ auth.user?.email }}</span>
          </span>
        </button>
        <div
          v-if="menuOpen"
          class="absolute right-0 mt-2 w-56 rounded-xl border border-slate-200 bg-white p-2 shadow-lg"
          role="menu"
          @keydown.esc="menuOpen = false"
        >
          <div class="flex items-center justify-between px-2 py-1.5">
            <span class="text-xs text-slate-500">Peran Anda</span>
            <AppBadge color="indigo">{{ roleLabel(auth.role ?? '') }}</AppBadge>
          </div>
          <button
            type="button"
            class="mt-1 flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm text-slate-700 hover:bg-slate-100"
            role="menuitem"
            @click="logout"
          >
            <LogOut class="h-4 w-4" /> Keluar
          </button>
        </div>
      </div>
    </div>
  </header>
</template>
