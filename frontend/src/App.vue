<script setup>
import { computed } from 'vue'
import { RouterView } from 'vue-router'
import { useRoute } from 'vue-router'
import SupportChatWidget from '@/components/SupportChatWidget.vue'

import GeolocationGate from '@/components/GeolocationGate.vue'

const route = useRoute()
const showSupportWidget = computed(() => route.path === '/')
</script>

<template>
  <GeolocationGate />
  <RouterView />
  <SupportChatWidget v-if="showSupportWidget" />
</template>

<style>
body {
  margin: 0;
  padding: 0;
  font-family: "PingFang SC", sans-serif;
  background: #f6f7f9;
}
</style>
