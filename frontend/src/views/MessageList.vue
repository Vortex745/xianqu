<template>
  <div class="message-list-page" v-loading="loading">
    <div class="nav-header">
      <div class="nav-content container">
        <div class="back-btn" @click="$router.push('/')">
          <el-icon><ArrowLeft /></el-icon> 返回首页
        </div>
        <div class="page-title">消息中心</div>
        <div class="placeholder"></div>
      </div>
    </div>

    <div class="container main-content">
      <div class="desktop-shell">
        <section class="chat-list-card glass-panel">
          <div class="list-head">
            <h3>近期会话</h3>
            <span class="contact-count">共 {{ contacts.length }} 位联系人</span>
          </div>

          <div v-if="contacts.length === 0 && !loading" class="empty-state">
            <div class="empty-deco">
              <div class="deco-ring ring-1"></div>
              <div class="deco-ring ring-2"></div>
            </div>
            <div class="empty-icon">
              <svg viewBox="0 0 64 64" fill="none" class="chat-svg">
                <path d="M8 14h36a4 4 0 014 4v20a4 4 0 01-4 4H24l-10 8v-8H8a4 4 0 01-4-4V18a4 4 0 014-4z" stroke="#1a1f29" stroke-width="2.5" stroke-linejoin="round"/>
                <circle cx="16" cy="28" r="2.5" fill="#1a1f29"/>
                <circle cx="26" cy="28" r="2.5" fill="#1a1f29"/>
                <circle cx="36" cy="28" r="2.5" fill="#1a1f29"/>
                <path d="M52 24v14a4 4 0 01-4 4h-6" stroke="#ffdf5d" stroke-width="2.5" stroke-linecap="round"/>
              </svg>
            </div>
            <div class="empty-text">
              <p class="main">暂无会话记录</p>
              <p class="sub">浏览感兴趣的商品，即可发起沟通</p>
            </div>
            <button class="go-home-btn" @click="$router.push('/')">
              <span>浏览商品</span>
              <svg viewBox="0 0 20 20" fill="none" class="btn-arrow">
                <path d="M4 10h12M12 6l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>

          <div v-else class="contact-scroll">
            <div
              v-for="item in contacts"
              :key="item.id"
              class="contact-item"
              @click="toChat(item.id)"
            >
              <div class="avatar-box">
                <el-avatar :size="50" :src="item.avatar || defaultAvatar" class="avatar-img" />
                <span v-if="Number(item.unread_count) > 0" class="unread-badge">
                  {{ Number(item.unread_count) > 99 ? '99+' : item.unread_count }}
                </span>
              </div>

              <div class="info-box">
                <div class="row-top">
                  <span class="nickname">{{ item.nickname || item.username || '用户' }}</span>
                  <span class="time">{{ formatTime(item.time) }}</span>
                </div>
                <p class="last-msg">{{ item.last_msg || '暂无消息记录' }}</p>
              </div>

              <div class="action-icon">
                <el-icon><ArrowRight /></el-icon>
              </div>
            </div>
          </div>
        </section>


      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { ArrowLeft, ArrowRight, ChatDotRound, Bell } from '@element-plus/icons-vue'

const router = useRouter()
const loading = ref(false)
const contacts = ref([])
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const unreadTotal = computed(() => {
  return contacts.value.reduce((sum, item) => sum + Number(item.unread_count || 0), 0)
})

const getCurrentUser = () => {
  try {
    const u = localStorage.getItem('user')
    return u ? JSON.parse(u) : null
  } catch (e) {
    return null
  }
}

const user = ref(getCurrentUser())

const fetchContacts = async () => {
  if (!user.value || !user.value.id) {
    router.push('/')
    return
  }

  loading.value = true
  try {
    const res = await request.get('/api/chat/contacts')
    contacts.value = res.data || []
  } catch (e) {
    console.error('获取消息列表失败:', e)
  } finally {
    loading.value = false
  }
}

const toChat = (targetId) => {
  router.push(`/chat/${targetId}`)
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()

  if (date.toDateString() === now.toDateString()) {
    return date.getHours().toString().padStart(2, '0') + ':' + date.getMinutes().toString().padStart(2, '0')
  } else if (date.getFullYear() === now.getFullYear()) {
    return (date.getMonth() + 1) + '-' + date.getDate()
  }
  return date.getFullYear() + '-' + (date.getMonth() + 1) + '-' + date.getDate()
}

onMounted(() => {
  fetchContacts()
})
</script>

<style scoped lang="scss">
@use "@/assets/tokens.scss" as *;

.message-list-page {
  min-height: 100vh;
  background: 
    url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E"),
    linear-gradient(165deg, #f8fafb 0%, #eef2f7 35%, #e8ecf3 100%);
  background-blend-mode: overlay, normal;
  padding-bottom: 44px;
  position: relative;
  
  &::before {
    content: '';
    position: fixed;
    top: -200px;
    right: -150px;
    width: 400px;
    height: 400px;
    border-radius: 50%;
    background: linear-gradient(135deg, rgba(255, 223, 93, 0.25), rgba(255, 201, 64, 0.1));
    filter: blur(80px);
    pointer-events: none;
    z-index: 0;
  }
  
  &::after {
    content: '';
    position: fixed;
    bottom: -180px;
    left: -120px;
    width: 360px;
    height: 360px;
    border-radius: 50%;
    background: linear-gradient(225deg, rgba(168, 212, 255, 0.3), rgba(126, 184, 240, 0.1));
    filter: blur(80px);
    pointer-events: none;
    z-index: 0;
  }
}

.container {
  width: 100%;
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 24px;
  box-sizing: border-box;
  position: relative;
  z-index: 1;
}

.glass-panel {
  background: rgba(255, 255, 255, 0.84);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 14px 30px rgba(11, 18, 32, 0.06);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
}

.nav-header {
  height: 76px;
  position: sticky;
  top: 0;
  z-index: 80;
  background: rgba(255, 255, 255, 0.84);
  border-bottom: 1px solid rgba(18, 24, 36, 0.06);

  .nav-content {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .back-btn {
    cursor: pointer;
    font-size: 16px;
    font-weight: 700;
    color: #364050;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 9px 14px;
    border-radius: 999px;
    border: 1px solid rgba(39, 47, 62, 0.1);
    background: #fff;
    transition: 0.2s;
    &:hover { color: #121926; transform: translateY(-1px); }
  }

  .page-title {
    font-size: 34px;
    font-weight: 900;
    color: #242c3b;
    letter-spacing: 0.3px;
  }

  .placeholder { width: 112px; }
}

.main-content { margin-top: 24px; }

.desktop-shell {
  max-width: 720px;
  margin: 0 auto;
}

.chat-list-card {
  border-radius: 20px;
  min-height: 520px;
  overflow: hidden;
}

.list-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 20px 22px 16px;
  border-bottom: 1px dashed #e7ebf2;
  h3 {
    margin: 0;
    color: #1f2836;
    font-size: 20px;
    font-weight: 900;
    letter-spacing: -0.3px;
  }
  .contact-count {
    font-size: 13px;
    font-weight: 700;
    color: #7a8495;
    background: #f2f5fa;
    padding: 4px 10px;
    border-radius: 20px;
  }
}

.contact-scroll {
  max-height: calc(100vh - 220px);
  overflow: auto;
  padding: 8px 10px 10px;
}

.contact-item {
  display: flex;
  align-items: center;
  padding: 14px;
  border: 1px solid transparent;
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.24s cubic-bezier(0.22, 1, 0.36, 1);
  position: relative;
  margin-bottom: 2px;

  &:hover {
    background: #fffef6;
    border-color: #ecd777;
    transform: translateX(3px);
    .action-icon { opacity: 1; transform: translateX(0); }
  }

  .avatar-box {
    margin-right: 12px;
    position: relative;
    .avatar-img {
      border: 2px solid #fff;
      box-shadow: 0 6px 12px rgba(19, 27, 40, 0.08);
    }
    .unread-badge {
      position: absolute;
      top: -4px;
      right: -4px;
      min-width: 18px;
      height: 18px;
      padding: 0 4px;
      border-radius: 999px;
      background: #ff3b30;
      color: #fff;
      font-size: 10px;
      font-weight: 800;
      line-height: 18px;
      text-align: center;
      box-shadow: 0 0 0 2px #fff;
    }
  }

  .info-box {
    flex: 1;
    overflow: hidden;
    .row-top {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 4px;
      .nickname {
        font-size: 16px;
        font-weight: 800;
        color: #253043;
      }
      .time {
        font-size: 12px;
        color: #8b94a5;
        font-weight: 700;
      }
    }
    .last-msg {
      margin: 0;
      font-size: 13px;
      color: #657186;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      line-height: 1.35;
    }
  }

  .action-icon {
    margin-left: 10px;
    color: #9aa4b6;
    font-size: 16px;
    opacity: 0;
    transform: translateX(-4px);
    transition: 0.2s;
  }
}

.empty-state {
  min-height: 420px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 32px;
  position: relative;
  overflow: hidden;
  
  .empty-deco {
    position: absolute;
    inset: 0;
    pointer-events: none;
    .deco-ring {
      position: absolute;
      border-radius: 50%;
      border: 1px dashed rgba(26, 31, 41, 0.08);
    }
    .ring-1 {
      width: 300px;
      height: 300px;
      top: -80px;
      right: -60px;
    }
    .ring-2 {
      width: 200px;
      height: 200px;
      bottom: -40px;
      left: -30px;
    }
  }
  
  .empty-icon {
    width: 88px;
    height: 88px;
    border-radius: 24px;
    background: 
      radial-gradient(circle at 30% 25%, rgba(255, 223, 93, 0.4), transparent 55%),
      linear-gradient(155deg, #fefefe, #f3f6fb);
    border: 1px solid rgba(227, 232, 241, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 24px;
    box-shadow: 
      inset 0 1px 0 rgba(255, 255, 255, 0.9),
      0 16px 32px rgba(13, 20, 34, 0.1);
    transform: rotate(-4deg);
    position: relative;
    z-index: 1;
    
    .chat-svg {
      width: 48px;
      height: 48px;
    }
  }
  
  .empty-text {
    position: relative;
    z-index: 1;
    .main {
      margin: 0 0 6px;
      color: #1f2836;
      font-size: 22px;
      font-weight: 900;
      letter-spacing: -0.3px;
    }
    .sub {
      margin: 0;
      color: #6b7a8f;
      font-size: 14px;
      font-weight: 600;
      line-height: 1.5;
    }
  }
  
  .go-home-btn {
    margin-top: 28px;
    height: 48px;
    padding: 0 24px;
    border-radius: 14px;
    border: none;
    background: $ink;
    color: $primary;
    font-size: 14px;
    font-weight: 800;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    box-shadow: 0 10px 24px rgba(11, 16, 24, 0.22);
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    position: relative;
    z-index: 1;
    
    .btn-arrow {
      width: 16px;
      height: 16px;
      transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    }
    
    &:hover { 
      transform: translateY(-3px); 
      box-shadow: 0 14px 32px rgba(11, 16, 24, 0.26);
      .btn-arrow { transform: translateX(4px); }
    }
  }
}

.desktop-panel {
  border-radius: 22px;
  min-height: 520px;
  padding: 26px;
  position: sticky;
  top: 100px;
  z-index: 1;
  
  .panel-header {
    h3 {
      margin: 0;
      font-size: 28px;
      color: #1f2836;
      font-weight: 900;
      letter-spacing: -0.5px;
    }
    .sub {
      margin: 10px 0 0;
      font-size: 14px;
      color: #6a7588;
      font-weight: 600;
    }
  }
}

.stats-grid {
  margin-top: 28px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.stat-item {
  border-radius: 18px;
  padding: 18px;
  background: linear-gradient(145deg, #f8fafc, #f2f5fa);
  border: 1px solid #e6ebf3;
  display: flex;
  align-items: center;
  gap: 14px;
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 20px rgba(13, 20, 34, 0.08);
  }
  
  .stat-icon {
    width: 44px;
    height: 44px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    
    svg {
      width: 22px;
      height: 22px;
    }
    
    &.chat-icon {
      background: linear-gradient(135deg, #fff8dc, #ffefab);
      color: #8b7722;
    }
    
    &.unread-icon {
      background: linear-gradient(135deg, #e8f4ff, #cce5ff);
      color: #3b82f6;
    }
  }
  
  .stat-content {
    .value {
      font-size: 28px;
      color: #1f2736;
      font-weight: 900;
      line-height: 1;
      letter-spacing: -0.5px;
    }
    .label {
      margin-top: 4px;
      color: #778297;
      font-size: 12px;
      font-weight: 700;
    }
  }
}

.tips-card {
  margin-top: 20px;
  border-radius: 16px;
  padding: 16px 18px;
  background: linear-gradient(145deg, #fffef6, #fffce8);
  border: 1px solid #eedb8a;
  color: #3e495d;
  font-size: 13px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 12px;
  
  .tip-icon {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    background: rgba(255, 223, 93, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    color: #8b7722;
    flex-shrink: 0;
    
    svg {
      width: 18px;
      height: 18px;
    }
  }
}

@media (max-width: 1080px) {
  .container {
    padding: 0 14px;
  }
  .nav-header {
    height: 64px;
    .page-title { font-size: 24px; }
    .placeholder { width: 88px; }
  }
  .desktop-shell {
    grid-template-columns: 1fr;
    gap: 14px;
  }
  .desktop-panel {
    min-height: auto;
    position: relative;
    top: 0;
  }
  .chat-list-card {
    min-height: 480px;
  }
}

@media (max-width: 640px) {
  .nav-header {
    .back-btn {
      font-size: 14px;
      padding: 7px 10px;
    }
    .page-title {
      font-size: 20px;
    }
    .placeholder {
      width: 60px;
    }
  }
  .contact-item {
    padding: 12px;
    .info-box .row-top .nickname { font-size: 15px; }
  }
}
</style>
