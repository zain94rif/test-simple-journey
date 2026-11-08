<script setup lang="ts">
import { computed, useId } from 'vue'

const props = withDefaults(
  defineProps<{
    label: string
    modelValue: string
    rows?: number
    placeholder?: string
    error?: string | null
    counter?: number
    maxLength?: number
    disabled?: boolean
    required?: boolean
  }>(),
  {
    rows: 4,
    placeholder: undefined,
    error: null,
    counter: undefined,
    maxLength: undefined,
    disabled: false,
    required: false,
  },
)

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const fieldId = useId()
const charCount = computed(() => props.modelValue.length)
const describedBy = computed(() => (props.error ? `${fieldId}-error` : undefined))
</script>

<template>
  <div class="w-full">
    <div class="mb-1 flex items-center justify-between">
      <label :for="fieldId" class="block text-sm font-medium text-slate-700">
        {{ props.label }}
      </label>
      <span v-if="props.counter" class="text-xs text-slate-500">{{ charCount }}/{{ props.counter }}</span>
    </div>
    <textarea
      :id="fieldId"
      :value="props.modelValue"
      :rows="props.rows"
      :placeholder="props.placeholder"
      :maxlength="props.maxLength"
      :disabled="props.disabled"
      :required="props.required"
      :aria-invalid="props.error ? 'true' : undefined"
      :aria-describedby="describedBy"
      :class="[
        'block w-full rounded-lg border bg-white px-3 py-2 text-sm shadow-sm',
        props.error ? 'border-red-400' : 'border-slate-300 focus:border-indigo-500',
        props.disabled ? 'bg-slate-100 text-slate-400' : '',
      ]"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    ></textarea>
    <p v-if="props.error" :id="`${fieldId}-error`" class="mt-1 text-xs text-red-600">
      {{ props.error }}
    </p>
  </div>
</template>
