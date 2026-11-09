<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Plus } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import ErrorState from '@/components/ui/ErrorState.vue'
import PaginationBar from '@/components/ui/PaginationBar.vue'
import SkeletonBlock from '@/components/ui/SkeletonBlock.vue'
import TicketTable from '@/components/tickets/TicketTable.vue'
import TicketCardList from '@/components/tickets/TicketCardList.vue'
import TicketFilters from '@/components/tickets/TicketFilters.vue'
import { useTicketStore } from '@/stores/tickets'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'
import type { TicketFilters as Filters } from '@/lib/types'

const store = useTicketStore()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()

const FILTER_KEYS = ['status', 'priority', 'assignee', 'search', 'sort', 'order'] as const

let syncingFromUrl = false

function numberFromQuery(value: unknown, fallback: number): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
}

function stringFromQuery(value: unknown, fallback: string): string {
  return typeof value === 'string' ? value : fallback
}

/** Restore filters and page from the URL query (shareable links). */
function applyRouteQuery() {
  syncingFromUrl = true
  store.setFilters({
    status: stringFromQuery(route.query.status, ''),
    priority: stringFromQuery(route.query.priority, ''),
    assignee: stringFromQuery(route.query.assignee, ''),
    search: stringFromQuery(route.query.search, ''),
    sort: stringFromQuery(route.query.sort, 'created_at:desc'),
    order: stringFromQuery(route.query.order, 'desc'),
    page: numberFromQuery(route.query.page, 1),
    per_page: numberFromQuery(route.query.per_page, 20),
  })
  syncingFromUrl = false
}

function queryFromFilters(): Record<string, string> {
  const f = store.filters
  const query: Record<string, string> = {}
  for (const key of FILTER_KEYS) {
    if (f[key]) query[key] = String(f[key])
  }
  if (f.page !== 1) query.page = String(f.page)
  if (f.per_page !== 20) query.per_page = String(f.per_page)
  return query
}

async function refresh() {
  store.listError = null
  try {
    await store.fetchTickets()
  } catch (err) {
    // listError telah disimpan di store untuk ErrorState; tidak perlu toast.
    if (!store.listError) toast.error(toApiError(err))
  }
}

const activeFilters = computed(() => store.hasActiveFilters())

watch(
  () => ({ ...store.filters }),
  async () => {
    if (syncingFromUrl) return
    // Identical, still-fresh request (e.g. URL filter restore on re-entry):
    // keep the rendered cache and do not hit the API again.
    if (store.canUseCachedList()) {
      await router.replace({ query: queryFromFilters() })
      return
    }
    // Any filter change resets to page 1 unless the change itself was the page.
    await router.replace({ query: queryFromFilters() })
    await refresh()
  },
  { deep: true },
)

watch(
  () => route.query,
  () => {
    const currentQuery = queryFromFilters()
    const incoming = Object.fromEntries(
      Object.entries(route.query as Record<string, unknown>).map(([k, v]) => [
        k,
        typeof v === 'string' ? v : '',
      ]),
    )
    const same =
      Object.keys(incoming).length === Object.keys(currentQuery).length &&
      FILTER_KEYS.concat(['page', 'per_page']).every((key) => incoming[key] === currentQuery[key])
    if (same) return
    applyRouteQuery()
    void refresh()
  },
)

onMounted(() => {
  applyRouteQuery()
  // Re-entering the page with an identical, still-fresh request: serve from
  // the store cache instead of hitting the API again.
  if (!store.canUseCachedList()) void refresh()
})

function onFilterChange(patch: Partial<Filters>) {
  store.setFilters(patch)
}

function resetFilters() {
  store.resetFilters()
  void router.replace({ query: {} })
}

function openDetail(id: number) {
  void router.push({ name: 'ticket-detail', params: { id: String(id) } })
}

function changePage(page: number) {
  store.setFilters({ page })
}

function changePerPage(perPage: number) {
  store.setFilters({ per_page: perPage, page: 1 })
}
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-xl font-bold text-slate-900">Tiket</h1>
        <p class="text-sm text-slate-500">
          {{ store.meta.total }} tiket ·
          <span v-if="auth.isRequester">hanya tiket Anda</span>
          <span v-else>semua tiket</span>
        </p>
      </div>
      <AppButton @click="router.push({ name: 'ticket-create' })">
        <Plus class="h-4 w-4" /> Buat Tiket
      </AppButton>
    </div>

    <TicketFilters
      v-if="auth.isStaff"
      :filters="store.filters"
      :show-assignee="auth.isStaff"
      :is-agent="auth.isAgent"
      :active="activeFilters"
      @update:filter="onFilterChange"
      @reset="resetFilters"
    />
    <TicketFilters
      v-else
      :filters="store.filters"
      :show-assignee="false"
      :active="activeFilters"
      @update:filter="onFilterChange"
      @reset="resetFilters"
    />

    <template v-if="store.listLoading">
      <div class="space-y-2">
        <SkeletonBlock v-for="i in 5" :key="i" height="56px" />
      </div>
    </template>
    <template v-else-if="store.listError">
      <ErrorState :message="String(store.listError)" retry-label="Coba lagi" @retry="refresh" />
    </template>
    <template v-else-if="store.items.length === 0">
      <EmptyState
        :kind="activeFilters ? 'filter' : 'none'"
        :title="activeFilters ? 'Tidak ada tiket yang cocok' : 'Belum ada tiket'"
        :description="
          activeFilters
            ? 'Coba ubah atau hapus filter untuk memperluas hasil.'
            : 'Buat tiket pertama Anda dan tim akan segera menangani.'"
      >
        <template #action>
          <AppButton v-if="!activeFilters" size="sm" @click="router.push({ name: 'ticket-create' })">
            <Plus class="h-4 w-4" /> Buat tiket pertama
          </AppButton>
          <AppButton v-else size="sm" variant="secondary" @click="resetFilters">Hapus filter</AppButton>
        </template>
      </EmptyState>
    </template>
    <template v-else>
      <!-- Desktop: table; Mobile: cards -->
      <div class="hidden md:block">
        <TicketTable :tickets="store.items" @open="openDetail" />
      </div>
      <div class="md:hidden">
        <TicketCardList :tickets="store.items" />
      </div>
      <PaginationBar
        :page="store.meta.page"
        :per-page="store.meta.per_page"
        :total="store.meta.total"
        :total-pages="store.meta.total_pages"
        @update:page="changePage"
        @update:per-page="changePerPage"
      />
    </template>
  </div>
</template>
