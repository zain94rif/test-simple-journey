<script setup lang="ts">
import { computed, ref, useId } from 'vue'

const props = withDefaults(
  defineProps<{
    label: string
    modelValue: string
    type?: string
    placeholder?: string
    error?: string | null
    hint?: string | null
    required?: boolean
    disabled?: boolean
    autocomplete?: string
  }>(),
  {
    type: 'text',
    placeholder: undefined,
    error: null,
    hint: null,
    required: false,
    disabled: false,
    autocomplete: undefined,
  },
)

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const inputId = useId()
const inputEl = ref<HTMLInputElement | null>(null)

/** Programmatic focus support (used by forms to jump to the first error). */
function focus() {
  inputEl.value?.focus()
}

defineExpose({ focus })
const describedBy = computed(() =>
  props.error ? `${inputId}-error` : props.hint ? `${inputId}-hint` : undefined,
)
</script>

<template>
  <div class="w-full">
    <label :for="inputId" class="mb-1 block text-sm font-medium text-slate-700">
      {{ props.label }}
      <span v-if="props.required" class="text-red-600" aria-hidden="true">*</span>
      <span v-if="props.required" class="sr-only">(wajib diisi)</span>
    </label>
    <input
      :id="inputId"
      ref="inputEl"
      :type="props.type"
      :value="props.modelValue"
      :placeholder="props.placeholder"
      :required="props.required"
      :disabled="props.disabled"
      :autocomplete="props.autocomplete"
      :aria-invalid="props.error ? 'true' : undefined"
      :aria-describedby="describedBy"
      :class="[
        'block w-full rounded-lg border bg-white px-3 py-2 text-sm shadow-sm transition-colors',
        props.error
          ? 'border-red-400 focus:border-red-500'
          : 'border-slate-300 focus:border-indigo-500',
        props.disabled ? 'bg-slate-100 text-slate-400' : '',
      ]"
      @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="props.error" :id="`${inputId}-error`" class="mt-1 text-xs text-red-600">
      {{ props.error }}
    </p>
    <p v-else-if="props.hint" :id="`${inputId}-hint`" class="mt-1 text-xs text-slate-500">
      {{ props.hint }}
    </p>
  </div>
</template>
