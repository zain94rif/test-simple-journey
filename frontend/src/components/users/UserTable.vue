<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/lib/types'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppSpinner from '@/components/ui/AppSpinner.vue'
import { ROLE_OPTIONS } from '@/lib/constants'
import { formatDateTime } from '@/lib/format'

const props = defineProps<{
  users: User[]
  loading?: boolean
  updatingId?: number
}>()

const emit = defineEmits<{
  (e: 'update:role', user: User, role: string): void
}>()

const auth = useAuthStore()
const roleOptions = computed(() => ROLE_OPTIONS)

function canEdit(user: User): boolean {
  return auth.user?.id !== user.id
}
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-slate-200 bg-white">
    <table class="min-w-full divide-y divide-slate-200 text-sm">
      <thead class="bg-slate-50 text-left text-xs font-medium uppercase tracking-wide text-slate-500">
        <tr>
          <th class="px-4 py-3">Nama</th>
          <th class="px-4 py-3">Email</th>
          <th class="px-4 py-3">Peran</th>
          <th class="hidden px-4 py-3 md:table-cell">Dibuat</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-100">
        <tr v-for="user in props.users" :key="user.id">
          <td class="px-4 py-3 font-medium text-slate-800">
            {{ user.name }}
            <span v-if="auth.user?.id === user.id" class="ml-1 text-xs text-slate-400">(Anda)</span>
          </td>
          <td class="px-4 py-3 text-slate-600">{{ user.email }}</td>
          <td class="px-4 py-3">
            <div class="flex items-center gap-2">
              <AppSelect
                :model-value="user.role"
                :options="roleOptions"
                :disabled="!canEdit(user) || props.updatingId === user.id"
                :title="canEdit(user) ? undefined : 'Anda tidak dapat mengubah peran sendiri.'"
                class="w-40"
                @update:model-value="(value: string) => emit('update:role', user, value)"
              />
              <AppSpinner v-if="props.updatingId === user.id" class="h-4 w-4 text-slate-500" />
            </div>
          </td>
          <td class="hidden px-4 py-3 text-slate-500 md:table-cell">
            {{ formatDateTime(user.created_at) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
