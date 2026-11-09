<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Compass } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'

const router = useRouter()
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-50 px-4">
    <div class="max-w-md rounded-xl border border-slate-200 bg-white p-8 text-center shadow-sm">
      <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100">
        <Compass class="h-6 w-6 text-indigo-600" />
      </div>
      <h1 class="text-lg font-bold text-slate-900">Halaman tidak ditemukan</h1>
      <p class="mt-1 text-sm text-slate-500">Alamat yang Anda buka tidak dikenal atau sudah dipindahkan.</p>
      <div class="mt-5 flex justify-center gap-2">
        <AppButton size="sm" @click="router.replace({ name: 'tickets' })">Ke beranda</AppButton>
      </div>
    </div>
  </div>
</template>
