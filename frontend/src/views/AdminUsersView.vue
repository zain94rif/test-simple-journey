<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from 'lucide-vue-next'
import AppSelect from '@/components/ui/AppSelect.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import PaginationBar from '@/components/ui/PaginationBar.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import UserTable from '@/components/users/UserTable.vue'
import { useUserStore } from '@/stores/users'
import { useToast } from '@/composables/useToast'
import { useConfirm } from '@/composables/useConfirm'
import { toApiError } from '@/lib/errors'
import { ROLE_OPTIONS } from '@/lib/constants'
import type { User } from '@/lib/types'

const store = useUserStore()
const toast = useToast()
const confirm = useConfirm()
const router = useRouter()

const errorState = ref<string | null>(null)
const roleFilterOptions = computed(() => ROLE_OPTIONS)

async function refresh() {
  errorState.value = null
  try {
    await store.fetchUsers()
  } catch (err) {
    errorState.value = toApiError(err).message
  }
}

onMounted(refresh)

async function onRoleChange(user: User, role: string) {
  const previousRole = user.role
  const ok = await confirm.confirm({
    title: 'Ubah peran pengguna',
    message: `Ubah peran ${user.name} dari ${previousRole} menjadi ${role}?`,
    confirmLabel: 'Ubah peran',
  })
  if (!ok) return
  try {
    const updated = await store.updateRole(user.id, { role })
    const index = store.items.findIndex((item) => item.id === user.id)
    if (index >= 0) store.items[index] = { ...updated }
    toast.success(`Peran ${updated.name} kini ${updated.role}.`)
  } catch (err) {
    const apiError = toApiError(err)
    if (apiError.code === 'LAST_ADMIN') {
      toast.error('Admin terakhir tidak dapat diturunkan.')
    } else {
      toast.error(apiError.message)
    }
    if (apiError.status === 403) {
      await router.replace({ name: 'forbidden' })
    }
    void refresh()
  }
}

function changePage(page: number) {
  store.setFilters({ page })
  void refresh()
}
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-4">
    <div>
      <h1 class="text-xl font-bold text-slate-900">Kelola Pengguna</h1>
      <p class="text-sm text-slate-500">{{ store.meta.total }} pengguna terdaftar</p>
    </div>

    <div class="rounded-xl border border-slate-200 bg-white p-3 sm:p-4">
      <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <div class="relative">
          <label class="mb-1 block text-sm font-medium text-slate-700">Cari nama / email</label>
          <Search class="absolute left-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-slate-400" aria-hidden="true" />
          <input
            type="search"
            :value="store.filters.search"
            placeholder="Cari pengguna…"
            class="block w-full rounded-lg border border-slate-300 bg-white py-2 pl-9 pr-3 text-sm shadow-sm"
            aria-label="Cari pengguna"
            @input="
              ($event.target as HTMLInputElement).value !== undefined &&
                store.setFilters({
                  search: ($event.target as HTMLInputElement).value,
                })
            "
            @change="refresh"
          />
          <p class="mt-1 text-xs text-slate-400">Tekan Enter untuk mencari.</p>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700">Peran</label>
          <AppSelect
            :model-value="store.filters.role"
            :options="roleFilterOptions"
            @update:model-value="
              (value: string) => {
                store.setFilters({ role: value })
                void refresh()
              }
            "
          >
            <option value="">Semua</option>
          </AppSelect>
        </div>
      </div>
    </div>

    <SkeletonBlock v-if="store.loading" height="240px" />
    <ErrorState v-else-if="errorState" :message="errorState" @retry="refresh" />
    <EmptyState
      v-else-if="store.items.length === 0"
      title="Belum ada pengguna"
      description="Belum ada data yang cocok dengan filter."
    />
    <template v-else>
      <UserTable
        :users="store.items"
        :updating-id="store.updatingId"
        @update:role="onRoleChange"
      />
      <PaginationBar
        :page="store.meta.page"
        :per-page="store.meta.per_page"
        :total="store.meta.total"
        :total-pages="store.meta.total_pages"
        @update:page="changePage"
        @update:per-page="
          (value: number) => {
            store.setFilters({ per_page: value, page: 1 })
            void refresh()
          }
        "
      />
    </template>
  </div>
</template>
