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
/* 全局样式重置 */
body {
  margin: 0;
  padding: 0;
  font-family: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', sans-serif;
}
</style>
