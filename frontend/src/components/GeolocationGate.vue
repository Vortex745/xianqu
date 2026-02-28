<template>
  <div v-if="showWarning" class="geo-gate-overlay">
    <div class="geo-gate-card animate-in">
      <div class="card-header">
        <div class="warning-icon-circ">
          <el-icon :size="40"><Warning /></el-icon>
        </div>
        <h2 class="card-title">环境兼容性提示</h2>
      </div>
      
      <div class="card-body">
        <p class="main-msg">
          检测到当前处于 <strong class="highlight">非安全连接 (HTTP)</strong> 部署环境。
        </p>
        <p class="detail-msg">
          由于现代浏览器的隐私与安全策略，<strong class="highlight">地理定位</strong> 与 <strong class="highlight">地图精准定位</strong> 功能在 HTTP 协议下无法正常调出。
        </p>
        <div class="solution-box">
          <div class="box-icon"><el-icon><Monitor /></el-icon></div>
          <div class="box-text">推荐通过 <strong>HTTPS</strong> 访问以获得完整的自动定位与“附近商品”体验。</div>
        </div>
      </div>
      
      <div class="card-footer">
        <button class="confirm-btn-primary" @click="handleConfirm">
          已了解，进入网站
          <el-icon class="btn-icon"><ArrowRight /></el-icon>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Warning, Monitor, ArrowRight } from '@element-plus/icons-vue'

const showWarning = ref(false)

const isSecureOrLocal = () => {
  // localhost, 127.0.0.1 或其它本地回环地址被浏览器视为安全环境
  const isLocal = window.location.hostname === 'localhost' || 
                  window.location.hostname === '127.0.0.1' || 
                  window.location.hostname === '0.0.0.0' ||
                  window.location.hostname === ''
  
  // 生产环境下如果没有 HTTPS，window.isSecureContext 为 false
  // 如果是本地开发环境，即使是 http 也有 isSecureContext = true (部分浏览器)
  // 这里结合判断：如果不是本地环境，且不是安全上下文，则触发警告
  return isLocal || window.isSecureContext
}

onMounted(() => {
  // 使用 sessionStorage 确保每次会话（新开标签页/重新打开浏览器）只提示一次
  const alreadyConfirmed = sessionStorage.getItem('geo_gate_confirmed')
  
  if (!alreadyConfirmed) {
    if (!isSecureOrLocal()) {
      showWarning.value = true
      // 禁止 body 滚动，防止不点击确定直接划走
      document.body.style.overflow = 'hidden'
      document.body.style.height = '100dvh'
    }
  }
})

const handleConfirm = () => {
  showWarning.value = false
  sessionStorage.setItem('geo_gate_confirmed', 'true')
  // 恢复滚动
  document.body.style.overflow = ''
  document.body.style.height = ''
}
</script>

<style scoped lang="scss">
.geo-gate-overlay {
  position: fixed;
  inset: 0;
  z-index: 999999; // 极高层级，覆盖所有 UI
  background: rgba(10, 12, 18, 0.88);
  backdrop-filter: blur(20px) saturate(1.8);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.geo-gate-card {
  width: 100%;
  max-width: 480px;
  background: #ffffff;
  border-radius: 40px;
  padding: 48px;
  box-shadow: 
    0 40px 100px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.1);
  text-align: center;
  position: relative;
}

.warning-icon-circ {
  width: 88px;
  height: 88px;
  background: #fff8e6;
  color: #f59e0b;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 24px;
  box-shadow: inset 0 2px 10px rgba(245, 158, 11, 0.1);
  
  .el-icon {
    filter: drop-shadow(0 4px 8px rgba(245, 158, 11, 0.2));
  }
}

.card-title {
  margin: 0 0 16px;
  font-size: 28px;
  font-weight: 900;
  color: #000;
  letter-spacing: -0.02em;
}

.card-body {
  color: #4b5563;
  line-height: 1.6;
  
  .main-msg {
    font-size: 17px;
    margin-bottom: 12px;
    color: #1f2937;
  }
  
  .detail-msg {
    font-size: 14px;
    margin-bottom: 24px;
    color: #6b7280;
  }
  
  .highlight {
    color: #000;
    font-weight: 800;
    background: linear-gradient(0deg, #ffdf5d 40%, transparent 40%);
    padding: 0 2px;
  }
}

.solution-box {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  background: #f8fafc;
  padding: 20px;
  border-radius: 20px;
  border: 1px solid #e2e8f0;
  text-align: left;
  
  .box-icon {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    background: #fff;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #64748b;
  }
  
  .box-text {
    font-size: 13px;
    color: #475569;
    line-height: 1.5;
    
    strong {
      color: #1e293b;
      font-weight: 700;
    }
  }
}

.card-footer {
  margin-top: 40px;
}

.confirm-btn-primary {
  width: 100%;
  height: 64px;
  background: #000;
  color: #ffdf5d;
  border: none;
  border-radius: 20px;
  font-size: 18px;
  font-weight: 900;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
  
  .btn-icon {
    font-size: 20px;
    transition: transform 0.3s;
  }
  
  &:hover {
    background: #111;
    transform: translateY(-3px) scale(1.02);
    box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
    
    .btn-icon {
      transform: translateX(4px);
    }
  }
  
  &:active {
    transform: translateY(0) scale(1);
  }
}

.animate-in {
  animation: card-appear 0.6s cubic-bezier(0.16, 1, 0.3, 1) both;
}

@keyframes card-appear {
  0% {
    opacity: 0;
    transform: translateY(60px) scale(0.9);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
