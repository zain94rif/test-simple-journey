<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { X } from 'lucide-vue-next'
import AppSidebar from './AppSidebar.vue'
import AppTopbar from './AppTopbar.vue'
import AppModal from '@/components/ui/AppModal.vue'
import AppButton from '@/components/ui/AppButton.vue'
import { useConfirmStore } from '@/composables/useConfirm'

const route = useRoute()
const drawerOpen = ref(false)
const confirmStore = useConfirmStore()

const activeName = ref<string | null>(route.name as string | null)
watch(
  () => route.name,
  (name) => {
    activeName.value = (name as string) ?? null
    drawerOpen.value = false
    if (route.meta.title) document.title = `${String(route.meta.title)} — Helpdesk`
  },
)

function hideOverflow(useHidden: boolean) {
  document.body.classList.toggle('overflow-hidden', useHidden)
}

watch(drawerOpen, (open) => hideOverflow(open))
onMounted(() => {
  if (!drawerOpen.value) hideOverflow(false)
})
</script>

<template>
  <div class="flex min-h-screen bg-slate-50">
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-white focus:px-3 focus:py-2 focus:shadow"
    >
      Lewati ke konten utama
    </a>

    <!-- Desktop sidebar -->
    <aside class="hidden w-60 shrink-0 border-r border-slate-200 bg-white lg:block">
      <div class="sticky top-0 flex h-14 items-center gap-2 border-b border-slate-200 px-4">
        <span class="h-7 w-7 rounded-lg bg-indigo-600"></span>
        <span class="text-sm font-bold text-slate-800">Helpdesk</span>
      </div>
      <AppSidebar :active-name="activeName" />
    </aside>

    <div class="flex min-w-0 flex-1 flex-col">
      <AppTopbar :show-menu-button="true" @toggle-menu="drawerOpen = true" />
      <main id="main-content" class="flex-1 px-4 py-6 sm:px-6 lg:px-8">
        <!-- Focus moves here after route change -->
        <router-view />
      </main>
    </div>

    <!-- Mobile drawer -->
    <Teleport to="body">
      <div v-if="drawerOpen" class="fixed inset-0 z-40 lg:hidden" role="dialog" aria-modal="true" aria-label="Menu navigasi">
        <div class="absolute inset-0 bg-slate-900/40" @click="drawerOpen = false"></div>
        <nav class="absolute inset-y-0 left-0 w-64 border-r border-slate-200 bg-white shadow-xl">
          <div class="flex h-14 items-center justify-between border-b border-slate-200 px-4">
            <span class="text-sm font-bold text-slate-800">Helpdesk</span>
            <button
              type="button"
              class="rounded-lg p-1 text-slate-400 hover:text-slate-600"
              aria-label="Tutup menu"
              @click="drawerOpen = false"
            >
              <X class="h-5 w-5" />
            </button>
          </div>
          <AppSidebar :active-name="activeName" @navigate="drawerOpen = false" />
        </nav>
      </div>
    </Teleport>

    <!-- Confirm dialog host -->
    <AppModal
      :open="Boolean(confirmStore.pending)"
      :title="confirmStore.pending?.title ?? ''"
      :closeable="false"
      @close="confirmStore.answer(false)"
    >
      <p class="text-sm text-slate-600">{{ confirmStore.pending?.message }}</p>
      <div class="mt-4 flex justify-end gap-2">
        <AppButton variant="secondary" size="sm" @click="confirmStore.answer(false)">
          {{ confirmStore.pending?.cancelLabel ?? 'Batal' }}
        </AppButton>
        <AppButton
          size="sm"
          :variant="confirmStore.pending?.danger ? 'danger' : 'primary'"
          @click="confirmStore.answer(true)"
        >
          {{ confirmStore.pending?.confirmLabel ?? 'Konfirmasi' }}
        </AppButton>
      </div>
    </AppModal>
  </div>
</template>
