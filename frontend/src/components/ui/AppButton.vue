<script setup lang="ts">
import { computed } from 'vue'
import AppSpinner from './AppSpinner.vue'

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost'
    size?: 'sm' | 'md' | 'lg'
    type?: 'button' | 'submit'
    loading?: boolean
    disabled?: boolean
    block?: boolean
  }>(),
  {
    variant: 'primary',
    size: 'md',
    type: 'button',
    loading: false,
    disabled: false,
    block: false,
  },
)

const emit = defineEmits<{ (e: 'click', event: MouseEvent): void }>()

const variantClass = computed(() => {
  switch (props.variant) {
    case 'secondary':
      return 'bg-white border border-slate-300 text-slate-700 hover:bg-slate-100'
    case 'danger':
      return 'bg-red-600 text-white hover:bg-red-700'
    case 'ghost':
      return 'bg-transparent text-slate-600 hover:bg-slate-100'
    default:
      return 'bg-indigo-600 text-white hover:bg-indigo-700'
  }
})

const sizeClass = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'px-3 py-1.5 text-sm'
    case 'lg':
      return 'px-5 py-3 text-base'
    default:
      return 'px-4 py-2 text-sm'
  }
})

const isDisabled = computed(() => props.disabled || props.loading)

function onClick(event: MouseEvent) {
  if (isDisabled.value) return
  emit('click', event)
}
</script>

<template>
  <button
    :type="props.type"
    :disabled="isDisabled"
    :class="[
      'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors',
      sizeClass,
      variantClass,
      props.block ? 'w-full' : '',
      isDisabled ? 'cursor-not-allowed opacity-60' : '',
    ]"
    @click="onClick"
  >
    <AppSpinner v-if="props.loading" class="h-4 w-4" />
    <slot />
  </button>
</template>
