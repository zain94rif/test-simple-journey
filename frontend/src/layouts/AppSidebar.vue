<script setup lang="ts">
import { computed } from 'vue'
import { ListPlus, List, Users } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'

defineProps<{
  activeName: string | null
}>()

const emit = defineEmits<{ (e: 'navigate'): void }>()

const auth = useAuthStore()

const links = computed(() => {
  const items = [
    { name: 'tickets', label: 'Daftar Tiket', icon: List },
    { name: 'ticket-create', label: 'Buat Tiket', icon: ListPlus },
  ]
  if (auth.isAdmin) {
    items.push({ name: 'admin-users', label: 'Kelola Pengguna', icon: Users })
  }
  return items
})
</script>

<template>
  <aside class="flex h-full w-full flex-col gap-1 p-3 lg:h-auto" aria-label="Navigasi utama">
    <router-link
      v-for="link in links"
      :key="link.name"
      :to="{ name: link.name }"
      :class="[
        'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
        activeName === link.name
          ? 'bg-indigo-50 text-indigo-700'
          : 'text-slate-600 hover:bg-slate-100',
      ]"
      @click="emit('navigate')"
    >
      <component :is="link.icon" class="h-4 w-4" />
      {{ link.label }}
    </router-link>
  </aside>
</template>
