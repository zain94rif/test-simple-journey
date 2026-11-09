<script setup lang="ts">
import { useRouter } from 'vue-router'
import { LockKeyhole } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'

const router = useRouter()
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-50 px-4">
    <div class="max-w-md rounded-xl border border-slate-200 bg-white p-8 text-center shadow-sm">
      <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
        <LockKeyhole class="h-6 w-6 text-red-600" />
      </div>
      <h1 class="text-lg font-bold text-slate-900">Akses ditolak</h1>
      <p class="mt-1 text-sm text-slate-500">
        Anda tidak memiliki izin untuk membuka halaman ini. Hubungi admin jika Anda merasa ini keliru.
      </p>
      <div class="mt-5 flex justify-center gap-2">
        <AppButton size="sm" @click="router.replace({ name: 'tickets' })">Kembali ke tiket</AppButton>
      </div>
    </div>
  </div>
</template>
