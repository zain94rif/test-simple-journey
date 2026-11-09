<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'
import { toApiError } from '@/lib/errors'

const auth = useAuthStore()
const router = useRouter()
const toast = useToast()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')

const errors = ref<{
  name?: string
  email?: string
  password?: string
  confirmPassword?: string
}>({})
const formError = ref<string | null>(null)

const loading = computed(() => auth.loading)

function validate(): boolean {
  errors.value = {}
  if (name.value.trim().length < 2) errors.value.name = 'Nama minimal 2 karakter.'
  else if (name.value.trim().length > 100) errors.value.name = 'Nama maksimal 100 karakter.'

  const emailValue = email.value.trim()
  if (!emailValue || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailValue)) {
    errors.value.email = 'Masukkan email yang valid.'
  }

  const passwordValue = password.value
  if (passwordValue.length < 8) {
    errors.value.password = 'Password minimal 8 karakter dan mengandung huruf serta angka.'
  } else if (!/[A-Za-z]/.test(passwordValue) || !/[0-9]/.test(passwordValue)) {
    errors.value.password = 'Password minimal 8 karakter dan mengandung huruf serta angka.'
  }

  if (confirmPassword.value !== passwordValue) {
    errors.value.confirmPassword = 'Konfirmasi password tidak cocok.'
  }
  return Object.keys(errors.value).length === 0
}

async function submit() {
  formError.value = null
  if (!validate()) return
  try {
    await auth.register({
      name: name.value.trim(),
      email: email.value.trim(),
      password: password.value,
    })
    toast.success('Akun berhasil dibuat. Silakan masuk.')
    await router.replace({
      name: 'login',
      query: { registered: '1', email: email.value.trim() },
    })
  } catch (err) {
    const apiError = toApiError(err)
    if (apiError.code === 'DUPLICATE_EMAIL') {
      errors.value.email = 'Email sudah terdaftar.'
    } else if (apiError.details) {
      for (const [field, message] of Object.entries(apiError.details)) {
        if (field in errors.value) {
          ;(errors.value as Record<string, string>)[field] = message
        } else {
          formError.value = message
        }
      }
      if (!Object.keys(apiError.details).some((f) => f in errors.value)) {
        formError.value = apiError.message
      }
    } else {
      formError.value = apiError.message
    }
  }
}
</script>

<template>
  <div>
    <h2 class="mb-1 text-lg font-bold text-slate-900">Daftar Akun</h2>
    <p class="mb-4 text-sm text-slate-500">
      Akun baru selalu memiliki peran <strong>Requester</strong>.
    </p>

    <div v-if="formError" class="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">
      {{ formError }}
    </div>

    <form class="space-y-4" novalidate @submit.prevent="submit">
      <AppInput v-model="name" label="Nama lengkap" autocomplete="name" :error="errors.name" required />
      <AppInput v-model="email" label="Email" type="email" autocomplete="email" :error="errors.email" required />
      <AppInput v-model="password" label="Password" type="password" autocomplete="new-password" :error="errors.password" required />
      <AppInput
        v-model="confirmPassword"
        label="Konfirmasi password"
        type="password"
        autocomplete="new-password"
        :error="errors.confirmPassword"
        required
      />
      <AppButton type="submit" block :loading="loading">Daftar</AppButton>
    </form>

    <p class="mt-4 text-center text-sm text-slate-500">
      Sudah punya akun?
      <router-link :to="{ name: 'login' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Masuk
      </router-link>
    </p>
  </div>
</template>
