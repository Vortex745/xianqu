<template>
  <button class="back-home-pill" type="button" @click="goHome">
    <el-icon><ArrowLeft /></el-icon>
    <span>{{ label }}</span>
  </button>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'

const props = defineProps({
  label: {
    type: String,
    default: '返回首页'
  }
})

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<style scoped lang="scss">
.back-home-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 24px;
  border-radius: 99px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  font-weight: 700;
  color: #333;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
  transition: transform 0.24s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.24s cubic-bezier(0.22, 1, 0.36, 1), background-color 0.24s cubic-bezier(0.22, 1, 0.36, 1);
  border: 1px solid #fff;
  font-size: 15px;
  line-height: 1;

  &:hover {
    transform: translateX(-4px);
    background: #fff;
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.1);
  }

  &:active {
    transform: translateX(-2px) scale(0.98);
  }
}
</style>
