<template>
  <img
    v-if="!hasError"
    :src="iconUrl"
    :alt="alt"
    class="iconify-symbol"
    :style="{
      width: normalizedSize,
      height: normalizedSize,
      opacity: disabled ? 0.45 : 1
    }"
    loading="lazy"
    decoding="async"
    @error="hasError = true"
  />
  <span
    v-else
    class="iconify-fallback"
    :style="{
      width: normalizedSize,
      height: normalizedSize,
      opacity: disabled ? 0.45 : 1
    }"
  ></span>
</template>

<script setup>
import { computed, ref, watch } from 'vue'

const props = defineProps({
  icon: {
    type: String,
    required: true
  },
  color: {
    type: String,
    default: '#303845'
  },
  size: {
    type: [String, Number],
    default: 18
  },
  alt: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const normalizedSize = computed(() => {
  if (typeof props.size === 'number') return `${props.size}px`
  return /^\d+$/.test(props.size) ? `${props.size}px` : props.size
})

const iconUrl = computed(() => {
  const iconName = String(props.icon || '').trim()
  const iconColor = encodeURIComponent(props.color || '#303845')
  return `https://api.iconify.design/${iconName}.svg?color=${iconColor}`
})

const hasError = ref(false)

watch(
  () => [props.icon, props.color],
  () => {
    hasError.value = false
  }
)
</script>

<style scoped>
.iconify-symbol {
  display: inline-block;
  vertical-align: middle;
  flex-shrink: 0;
  pointer-events: none;
}

.iconify-fallback {
  display: inline-block;
  border-radius: 999px;
  background: currentColor;
  opacity: 0.2;
}
</style>
