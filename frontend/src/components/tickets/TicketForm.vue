<script setup lang="ts">
import AppButton from '@/components/ui/AppButton.vue'
import { CATEGORY_OPTIONS, PRIORITY_OPTIONS } from '@/lib/constants'
import AppInput from '@/components/ui/AppInput.vue'
import AppSelect from '@/components/ui/AppSelect.vue'
import AppTextarea from '@/components/ui/AppTextarea.vue'
import { computed, nextTick, ref } from 'vue'
import type { TicketFormErrors as FormErrors } from '@/lib/types'
import type { CreateTicketPayload } from '@/stores/tickets'

export type { FormErrors } from '@/lib/types'

const props = defineProps<{
  errors?: FormErrors
  loading?: boolean
  serverError?: string | null
}>()

const emit = defineEmits<{ (e: 'submit', payload: CreateTicketPayload): void; (e: 'cancel'): void }>()

const title = ref('')
const description = ref('')
const category = ref('other')
const priority = ref('medium')

// Element refs so the first invalid field receives focus on failed submit.
const titleEl = ref<InstanceType<typeof AppInput> | null>(null)
const descriptionEl = ref<InstanceType<typeof AppTextarea> | null>(null)

function validate(): FormErrors {
  const errors: FormErrors = {}
  if (!title.value.trim()) {
    errors.title = 'Judul wajib diisi.'
  } else if (title.value.trim().length < 3) {
    errors.title = 'Judul minimal 3 karakter.'
  } else if (title.value.length > 150) {
    errors.title = 'Judul maksimal 150 karakter.'
  }
  if (!description.value.trim()) {
    errors.description = 'Deskripsi wajib diisi.'
  } else if (description.value.trim().length < 10) {
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

// Server-side (view-mapped) error display added to per-field local errors.
const fieldErrors = ref<FormErrors>({})

const hasAnyError = computed(() =>
  Boolean(
    Object.values(fieldErrors.value).some(Boolean) ||
      Object.values(props.errors ?? {}).some(Boolean),
  ),
)

function errorFor(field: keyof FormErrors): string | undefined {
  return fieldErrors.value[field] ?? props.errors?.[field]
}

async function onSubmit() {
  fieldErrors.value = validate()
  if (Object.keys(fieldErrors.value).length > 0) {
    // Bring the first invalid field into view so users see why it failed.
    await nextTick()
    if (fieldErrors.value.title) titleEl.value?.focus?.()
    else if (fieldErrors.value.description) descriptionEl.value?.focus?.()
    return
  }
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
    <p class="text-xs text-slate-500">
      Kolom bertanda
      <span class="font-semibold text-red-600" aria-hidden="true">*</span>
      wajib diisi.
    </p>
    <p
      v-if="hasAnyError"
      class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
      role="alert"
    >
      Ada input yang perlu dipertimbangkan ulang; lihat penanda merah di bawah.
    </p>
    <AppInput
      ref="titleEl"
      v-model="title"
      label="Judul"
      placeholder="Ringkas masalah yang dialami"
      :error="errorFor('title')"
      :hint="errorFor('title') ? undefined : '3–150 karakter'"
      required
    />
    <AppTextarea
      ref="descriptionEl"
      v-model="description"
      label="Deskripsi"
      :rows="6"
      :error="errorFor('description')"
      :hint="errorFor('description') ? undefined : 'Jelaskan kronologi dan dampak kerja. Minimal 10 karakter.'"
      :counter="5000"
      :max-length="5000"
      required
    />
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <AppSelect
        v-model="category"
        label="Kategori"
        :options="CATEGORY_OPTIONS"
        :error="errorFor('category')"
        required
      />
      <AppSelect
        v-model="priority"
        label="Prioritas"
        :options="PRIORITY_OPTIONS"
        :error="errorFor('priority')"
        required
      />
    </div>
    <div class="flex gap-3 pt-2">
      <AppButton type="submit" :loading="loading" :disabled="loading">Buat Tiket</AppButton>
      <AppButton variant="secondary" :disabled="loading" @click="emit('cancel')">Batal</AppButton>
    </div>
  </form>
</template>
