<template>
  <div class="page-transition-wrapper" ref="wrapper">
    <slot />
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import gsap from 'gsap'
import { animConfig, shouldReduceMotion } from '@/utils/animations'

const wrapper = ref(null)

onMounted(() => {
  nextTick(() => {
    if (!wrapper.value || shouldReduceMotion()) return
    
    // Keep route changes crisp; this wrapper should feel immediate, not theatrical.
    gsap.from(wrapper.value, {
      opacity: 0,
      y: 12,
      scale: 0.992,
      duration: animConfig.duration.fast,
      ease: animConfig.ease.smooth,
      clearProps: 'transform'
    })
  })
})
</script>

<style scoped>
.page-transition-wrapper {
  min-height: 100vh;
}
</style>
