<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/lib/types'
import RoleSelect from './RoleSelect.vue'
import AppSpinner from '@/components/ui/AppSpinner.vue'
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
              <RoleSelect
                :user="user"
                :saving="props.updatingId === user.id"
                @change="(value: string) => emit('update:role', user, value)"
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
