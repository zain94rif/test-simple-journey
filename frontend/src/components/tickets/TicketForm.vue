<script setup lang="ts">
import AppButton from '@/components/ui/AppButton.vue'
import { CATEGORY_OPTIONS, PRIORITY_OPTIONS } from '@/lib/constants'
import AppInput from '@/components/ui/AppInput.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppTextarea from '@/components/ui/AppTextarea.vue'
import { ref } from 'vue'
import type { CreateTicketPayload } from '@/stores/tickets'

export interface FormErrors {
  title?: string
  description?: string
  category?: string
  priority?: string
}

defineProps<{
  errors?: FormErrors
  loading?: boolean
  serverError?: string | null
}>()

const emit = defineEmits<{ (e: 'submit', payload: CreateTicketPayload): void; (e: 'cancel'): void }>()

const title = ref('')
const description = ref('')
const category = ref('other')
const priority = ref('medium')

function validate(): FormErrors {
  const errors: FormErrors = {}
  if (title.value.trim().length < 3) {
    errors.title = 'Judul minimal 3 karakter.'
  } else if (title.value.length > 150) {
    errors.title = 'Judul maksimal 150 karakter.'
  }
  const descriptionLength = description.value.trim().length
  if (descriptionLength < 10) {
    errors.description = 'Deskripsi minimal 10 karakter.'
  } else if (description.value.length > 5000) {
    errors.description = 'Deskripsi maksimal 5000 karakter.'
  }
  if (!CATEGORY_OPTIONS.some((c) => c.value === category.value)) {
    errors.category = 'Pilih kategori.'
  }
  if (!PRIORITY_OPTIONS.some((p) => p.value === priority.value)) {
    errors.priority = 'Pilih prioritas.'
  }
  return errors
}

function onSubmit() {
  const errors = validate()
  if (Object.keys(errors).length > 0) return
  emit('submit', {
    title: title.value.trim(),
    description: description.value.trim(),
    category: category.value,
    priority: priority.value,
  })
}
</script>

<template>
  <form class="max-w-2xl space-y-4" novalidate @submit.prevent="onSubmit">
    <div
      v-if="serverError"
      class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
      role="alert"
    >
      {{ serverError }}
    </div>
    <AppInput
      v-model="title"
      label="Judul"
      placeholder="Ringkas masalah yang dialami"
      :error="errors?.title"
      :hint="errors?.title ? undefined : '3–150 karakter'"
      required
    />
    <AppTextarea
      v-model="description"
      label="Deskripsi"
      :rows="6"
      :error="errors?.description"
      :hint="errors?.description ? undefined : 'Jelaskan kronologi dan dampak kerja. Minimal 10 karakter.'"
      :counter="5000"
      :max-length="5000"
      required
    />
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <AppSelect
        v-model="category"
        label="Kategori"
        :options="CATEGORY_OPTIONS"
        :error="errors?.category"
      />
      <AppSelect
        v-model="priority"
        label="Prioritas"
        :options="PRIORITY_OPTIONS"
        :error="errors?.priority"
      />
    </div>
    <div class="flex gap-3 pt-2">
      <AppButton type="submit" :loading="loading">Buat Tiket</AppButton>
      <AppButton variant="secondary" @click="emit('cancel')">Batal</AppButton>
    </div>
  </form>
</template>
