<script setup lang="ts">
import { computed, useId } from 'vue'

export interface SelectOption {
  value: string
  label: string
}

const props = withDefaults(
  defineProps<{
    label?: string
    options: readonly SelectOption[]
    modelValue: string
    error?: string | null
    disabled?: boolean
    required?: boolean
  }>(),
  {
    label: undefined,
    error: null,
    disabled: false,
  },
)

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const selectId = useId()
const current = computed(() => (props.modelValue === undefined ? '' : String(props.modelValue)))
const describedBy = computed(() => (props.error ? `${selectId}-error` : undefined))
</script>

<template>
  <div class="w-full">
    <label v-if="props.label" :for="selectId" class="mb-1 block text-sm font-medium text-slate-700">
      {{ props.label }}
      <span v-if="props.required" class="text-red-600" aria-hidden="true">*</span>
      <span v-if="props.required" class="sr-only">(wajib diisi)</span>
    </label>
    <select
      :id="selectId"
      :value="current"
      :disabled="props.disabled"
      :required="props.required"
      :aria-invalid="props.error ? 'true' : undefined"
      :aria-describedby="describedBy"
      :class="[
        'block w-full rounded-lg border bg-white px-3 py-2 text-sm shadow-sm',
        props.error ? 'border-red-400' : 'border-slate-300 focus:border-indigo-500',
        props.disabled ? 'bg-slate-100 text-slate-400' : '',
      ]"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <slot></slot>
      <option v-for="option in props.options" :key="option.value" :value="option.value">
        {{ option.label }}
      </option>
    </select>
    <p v-if="props.error" :id="`${selectId}-error`" class="mt-1 text-xs text-red-600">
      {{ props.error }}
    </p>
  </div>
</template>
