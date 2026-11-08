<script setup lang="ts">
import { HelpCircle } from 'lucide-vue-next'
import { APP_NAME } from '@/lib/constants'
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-100 px-4 py-10">
    <div class="w-full max-w-md">
      <div class="mb-6 flex flex-col items-center text-center">
        <div class="mb-2 flex h-12 w-12 items-center justify-center rounded-xl bg-indigo-600 text-white">
          <HelpCircle class="h-6 w-6" />
        </div>
        <h1 class="text-xl font-bold text-slate-900">{{ APP_NAME }}</h1>
        <p class="mt-1 text-sm text-slate-500">Internal ticketing untuk tim</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <router-view />
      </div>
    </div>
  </div>
</template>
