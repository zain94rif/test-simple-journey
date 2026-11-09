<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Eye, EyeOff } from 'lucide-vue-next'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'
import { safeRedirectPath } from '@/lib/format'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()

const email = ref(typeof route.query.email === 'string' ? route.query.email : '')
const password = ref('')
const showPassword = ref(false)
const errors = ref<{ email?: string; password?: string }>({})
const formError = ref<string | null>(null)

const registeredBanner = computed(() => route.query.registered === '1')
const loading = computed(() => auth.loading)

function validate(): boolean {
  errors.value = {}
  const emailValue = email.value.trim()
  if (!emailValue) {
    errors.value.email = 'Email wajib diisi.'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailValue)) {
    errors.value.email = 'Masukkan email yang valid.'
  }
  if (!password.value) {
    errors.value.password = 'Password wajib diisi.'
  }
  return Object.keys(errors.value).length === 0
}

async function submit() {
  formError.value = null
  if (!validate()) return
  try {
    await auth.login(email.value.trim(), password.value)
    toast.success('Berhasil masuk.')
    password.value = ''
    await router.replace(safeRedirectPath(route.query.redirect))
  } catch (err) {
    const apiError = toApiError(err)
    formError.value = apiError.message
  }
}
</script>

<template>
  <div>
    <h2 class="mb-1 text-lg font-bold text-slate-900">Masuk ke Helpdesk</h2>
    <p class="mb-4 text-sm text-slate-500">Gunakan akun kerja Anda.</p>

    <div
      v-if="registeredBanner"
      class="mb-4 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700"
      role="status"
    >
      Akun berhasil dibuat. Silakan masuk dengan email terdaftar.
    </div>

    <div v-if="formError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
      {{ formError }}
    </div>

    <form class="space-y-4" novalidate @submit.prevent="submit">
      <AppInput
        v-model="email"
        label="Email"
        type="email"
        autocomplete="email"
        placeholder="nama@perusahaan.id"
        :error="errors.email"
        required
      />
      <div>
        <AppInput
          v-model="password"
          label="Password"
          :type="showPassword ? 'text' : 'password'"
          autocomplete="current-password"
          :error="errors.password"
          required
        />
        <button
          type="button"
          class="mt-1 inline-flex items-center gap-1 text-xs text-slate-500 hover:text-slate-700"
          @click="showPassword = !showPassword"
        >
          <component :is="showPassword ? EyeOff : Eye" class="h-3.5 w-3.5" aria-hidden="true" />
          {{ showPassword ? 'Sembunyikan' : 'Tampilkan' }} password
        </button>
      </div>
      <AppButton type="submit" block :loading="loading">Masuk</AppButton>
    </form>

    <p class="mt-4 text-center text-sm text-slate-500">
      Belum punya akun?
      <router-link :to="{ name: 'register' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Daftar
      </router-link>
    </p>
  </div>
</template>
