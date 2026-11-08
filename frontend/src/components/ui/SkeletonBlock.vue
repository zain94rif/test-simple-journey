<script setup lang="ts">
withDefaults(
  defineProps<{
    class?: string
    width?: string
    height?: string
  }>(),
  { class: '', width: undefined, height: undefined },
)
</script>

<template>
  <div
    :class="`animate-pulse rounded-lg bg-slate-200 ${props.class ?? ''}`"
    :style="{ width: props.width, height: props.height }"
  ></div>
</template>
