<template>
  <div class="chat-room-page">
    <div class="nav-header">
      <div class="nav-content container">
        <div class="back-btn" @click="$router.go(-1)">
          <el-icon><ArrowLeft /></el-icon>
        </div>
        <div class="user-info">
          <span class="name">{{ targetUser.nickname || targetUser.username || ('用户 ' + targetId) }}</span>
          <span class="status-badge">在线</span>
        </div>
        <div class="placeholder"></div>
      </div>
    </div>

    <div class="chat-body" ref="msgBoxRef" @click="showEmoji = false">
      <div class="container">
        <div class="chat-tip">安全提示：交易过程中请勿轻信外部链接，谨防诈骗</div>

        <div
            v-for="(msg, index) in messages"
            :key="index"
            class="message-row animate-slide-in"
            :class="{ 'self': isMe(msg.sender_id) }"
        >
          <el-avatar
              :size="42"
              :src="resolveUrl(isMe(msg.sender_id) ? (user.avatar || defaultAvatar) : (targetUser.avatar || defaultAvatar))"
              class="avatar-img"
          />

          <div class="bubble-wrapper">
            <div v-if="msg.type === 1 || msg.type === 'text'" class="bubble text-bubble">
              {{ msg.content }}
            </div>

            <el-image
                v-else-if="msg.type === 2 || msg.type === 'image'"
                :src="resolveUrl(msg.content)"
                class="bubble image-bubble"
                :preview-src-list="[resolveUrl(msg.content)]"
                fit="cover"
                hide-on-click-modal
            >
              <template #error>
                <div class="img-error"><el-icon><Picture /></el-icon></div>
              </template>
            </el-image>
          </div>
        </div>
      </div>
    </div>

    <div class="chat-footer">
      <div class="container footer-inner">

        <transition name="fade-up">
          <div v-if="showEmoji" class="emoji-popover">
            <EmojiPicker :native="true" @select="onSelectEmoji" class="custom-picker" />
          </div>
        </transition>

        <div class="tool-bar">
          <div class="tool-btn" @click.stop="showEmoji = !showEmoji" title="表情">
            <span>😀</span>
          </div>
          <el-upload
              :action="uploadUrl"
              name="file"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleImageUpload"
              class="upload-wrapper"
          >
            <div class="tool-btn icon-btn" title="发送图片">
              <el-icon><Picture /></el-icon>
            </div>
          </el-upload>
        </div>

        <div class="input-row">
          <input
              v-model="inputText"
              type="text"
              placeholder="聊一聊宝贝细节..."
              @keyup.enter="sendMessage"
              @focus="showEmoji = false"
          />
          <button
              class="send-btn"
              @click="sendMessage"
              :disabled="!inputText.trim()"
              :class="{ 'active': inputText.trim() }"
          >
            <el-icon><Promotion /></el-icon>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { useRoute } from 'vue-router'
import request, { resolveUrl } from '@/utils/request'
import { ArrowLeft, Picture, Promotion } from '@element-plus/icons-vue'
import EmojiPicker from 'vue3-emoji-picker'
import 'vue3-emoji-picker/css'

const route = useRoute()
const targetId = Number(route.params.id || route.params.targetId)
const user = JSON.parse(localStorage.getItem('user') || '{}')
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const showEmoji = ref(false)
const targetUser = ref({}) // 对方信息
const messages = ref([])
const inputText = ref('')
const msgBoxRef = ref(null)
let socket = null

const uploadUrl = computed(() => {
  const baseUrl = import.meta.env.VITE_API_URL || '/'
  return baseUrl.endsWith('/') ? `${baseUrl}api/upload` : `${baseUrl}/api/upload`
})

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token')
  return { Authorization: token || '' }
})

const isMe = (senderId) => {
  return Number(senderId) === Number(user.id)
}

const scrollToBottom = () => {
  nextTick(() => {
    if (msgBoxRef.value) {
      msgBoxRef.value.scrollTo({
        top: msgBoxRef.value.scrollHeight,
        behavior: 'smooth'
      })
    }
  })
}

const onSelectEmoji = (emoji) => {
  inputText.value += emoji.i
}

// 获取对方信息
const fetchTargetInfo = async () => {
  try {
    const res = await request.get(`/api/users/${targetId}`)
    targetUser.value = res.data || {}
  } catch (e) {
    console.error('获取对方信息失败', e)
    targetUser.value = { nickname: `用户 ${targetId}` }
  }
}

// 获取历史消息
const fetchHistory = async () => {
  try {
    const res = await request.get('/api/chat/messages', { params: { target_id: targetId } })
    messages.value = res.data || []
    scrollToBottom()
  } catch (e) {
    console.error(e)
  }
}

// 初始化 WebSocket
const initWebSocket = () => {
  const token = localStorage.getItem('token')
  if (!token) return

  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const apiBase = import.meta.env.VITE_API_URL || ''
  const wsHost = apiBase ? apiBase.replace(/^https?:\/\//, '') : (import.meta.env.DEV ? 'localhost:8081' : window.location.host)
  const wsUrl = `${wsProtocol}://${wsHost}/api/ws?token=${encodeURIComponent(token)}`
  
  socket = new WebSocket(wsUrl)

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      const isRelevant =
          (Number(msg.sender_id) === targetId && Number(msg.receiver_id) === Number(user.id)) ||
          (Number(msg.sender_id) === Number(user.id) && Number(msg.receiver_id) === targetId)

      if (isRelevant) {
        messages.value.push(msg)
        scrollToBottom()
      }
    } catch (e) {}
  }

  socket.onopen = () => console.log('WS 已连接')
  socket.onclose = () => console.log('WS 断开')
}

// 发送文本
const sendMessage = () => {
  if (!inputText.value.trim()) return

  const msgData = {
    receiver_id: targetId,
    content: inputText.value,
    type: 1 // 文本
  }

  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(msgData))
  }

  inputText.value = ''
  showEmoji.value = false
  scrollToBottom()
}

// 发送图片
const handleImageUpload = (res) => {
  const finalUrl = resolveUrl(res.url)
  if (finalUrl) {
    const msgData = {
      receiver_id: targetId,
      content: finalUrl,
      type: 2 // 图片
    }

    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(msgData))
      scrollToBottom()
    }
  }
}

onMounted(async () => {
  await fetchTargetInfo()
  await fetchHistory()
  initWebSocket()
})

onUnmounted(() => {
  if (socket) socket.close()
})
</script>

<style scoped lang="scss">
@use "@/assets/tokens.scss" as *;

.chat-room-page {
  height: 100vh;
  background: $bg-page;
  display: flex;
  flex-direction: column;
}

.container {
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  padding: 0 20px;
  box-sizing: border-box;
}

/* 顶部导航 */
.nav-header {
  height: 60px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(0,0,0,0.05);
  position: sticky;
  top: 0;
  z-index: 100;

  .nav-content {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .back-btn {
    width: 36px; height: 36px;
    display: flex; align-items: center; justify-content: center;
    border-radius: 50%;
    cursor: pointer;
    transition: 0.2s;
    font-size: 18px;
    color: #333;
    &:hover { background: #f0f0f0; }
  }

  .user-info {
    text-align: center;
    .name { font-size: 16px; font-weight: 800; color: #333; display: block; }
    .status-badge {
      font-size: 10px; color: #52c41a; background: #f6ffed;
      padding: 1px 6px; border-radius: 4px; border: 1px solid #b7eb8f;
    }
  }

  .placeholder { width: 36px; }
}

/* 消息区域 */
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px 0;

  .chat-tip {
    text-align: center; font-size: 12px; color: #999;
    background: rgba(0,0,0,0.03); padding: 4px 12px;
    border-radius: 99px; width: fit-content; margin: 0 auto 20px;
  }

  .message-row {
    display: flex; gap: 12px; margin-bottom: 24px;
    align-items: flex-start;

    .avatar-img {
      border: 2px solid #fff;
      box-shadow: 0 2px 6px rgba(0,0,0,0.05);
      flex-shrink: 0;
    }

    .bubble-wrapper {
      max-width: 70%;

      .bubble {
        padding: 12px 16px;
        border-radius: 16px;
        font-size: 15px;
        line-height: 1.6;
        position: relative;
        word-break: break-all;
      }

      .text-bubble {
        background: #fff;
        color: #333;
        border-top-left-radius: 4px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.03);
      }

      .image-bubble {
        border-radius: 12px;
        max-width: 200px;
        cursor: zoom-in;
        box-shadow: 0 4px 12px rgba(0,0,0,0.08);
        transition: transform 0.2s;
        &:hover { transform: scale(1.02); }
      }
    }

    &.self {
      flex-direction: row-reverse;
      .bubble-wrapper .text-bubble {
        background: $primary;
        color: #000;
        border-top-left-radius: 16px;
        border-top-right-radius: 4px;
        font-weight: 500;
      }
    }
  }
}

/* 底部输入栏 */
.chat-footer {
  background: #fff;
  padding: 10px 0 20px;
  border-top: 1px solid #f5f5f5;
  position: relative;
  z-index: 200;

  .footer-inner {
    display: flex; flex-direction: column; gap: 10px;
    position: relative;
  }

  .tool-bar {
    display: flex; gap: 16px; padding-left: 4px;
    .tool-btn {
      cursor: pointer; font-size: 20px; transition: 0.2s;
      &:hover { transform: scale(1.1); }
    }
    .icon-btn {
      font-size: 22px; color: #666; display: flex; align-items: center;
      &:hover { color: #333; }
    }
  }

  .input-row {
    display: flex; gap: 12px; align-items: center;

    input {
      flex: 1;
      height: 44px;
      border: 2px solid #f0f0f0;
      border-radius: 22px;
      padding: 0 18px;
      font-size: 15px;
      outline: none;
      transition: all 0.3s;
      background: #f9f9f9;

      &:focus {
        background: #fff;
        border-color: $primary;
        box-shadow: 0 0 0 3px rgba(255, 223, 93, 0.2);
      }
    }

    .send-btn {
      width: 44px; height: 44px;
      border-radius: 50%;
      border: none;
      background: #eee;
      color: #999;
      display: flex; align-items: center; justify-content: center;
      font-size: 18px;
      cursor: not-allowed;
      transition: 0.3s;

      &.active {
        background: #333;
        color: $primary;
        cursor: pointer;
        transform: rotate(0deg);
        &:hover { transform: scale(1.1) rotate(-10deg); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
      }
    }
  }

  /* 表情面板样式 */
  .emoji-popover {
    position: absolute;
    bottom: 100%;
    left: 20px;
    margin-bottom: 10px;
    z-index: 99;
    box-shadow: 0 8px 30px rgba(0,0,0,0.15);
    border-radius: 12px;
    overflow: hidden;
  }
}

/* 动画 */
.animate-slide-in {
  animation: messageSlide 0.3s ease-out;
}
@keyframes messageSlide {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.fade-up-enter-active, .fade-up-leave-active { transition: all 0.3s ease; }
.fade-up-enter-from, .fade-up-leave-to { opacity: 0; transform: translateY(10px); }

@media (max-width: 600px) {
  .container { padding: 0 12px; }
  .chat-body .message-row .bubble-wrapper { max-width: 80%; }
}
</style>