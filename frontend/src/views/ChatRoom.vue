<template>
  <div class="chat-room-page" @click="showEmoji = false">
    <header class="nav-header">
      <div class="nav-content container">
        <button class="back-btn" type="button" @click="$router.go(-1)">
          <el-icon><ArrowLeft /></el-icon>
        </button>

        <div class="user-info">
          <span class="name">{{ targetUser.nickname || targetUser.username || ('用户 ' + targetId) }}</span>
          <span class="status-badge">{{ wsConnected ? '在线' : '重连中' }}</span>
        </div>

        <div class="placeholder"></div>
      </div>
    </header>

    <main class="chat-main container">
      <section class="conversation-card">
        <div class="chat-tip">安全提示：交易中别点陌生链接</div>

        <div class="messages-wrap" ref="msgBoxRef">
          <div v-if="messages.length === 0" class="empty-module">
            <div class="empty-title">还没开聊</div>
            <div class="empty-desc">发一句问问成色和价格</div>
          </div>

          <article
            v-for="(msg, index) in messages"
            :key="msg.id || `${msg.sender_id}-${msg.receiver_id}-${msg.created_at || index}`"
            class="message-row animate-slide-in"
            :class="{ self: isMe(msg.sender_id) }"
          >
            <el-avatar
              :size="40"
              :src="resolveUrl(isMe(msg.sender_id) ? (user.avatar || defaultAvatar) : (targetUser.avatar || defaultAvatar))"
              class="avatar-img"
            />

            <div class="bubble-wrapper">
              <div v-if="isTextMessage(msg)" class="bubble text-bubble">
                {{ msg.content }}
              </div>

              <el-image
                v-else-if="isImageMessage(msg)"
                :src="resolveUrl(msg.content)"
                class="bubble image-bubble"
                :preview-src-list="[resolveUrl(msg.content)]"
                fit="cover"
                hide-on-click-modal
              >
                <template #error>
                  <div class="img-error">
                    <el-icon><Picture /></el-icon>
                    <span>图片加载失败</span>
                  </div>
                </template>
              </el-image>

              <div v-else class="bubble text-bubble">
                {{ msg.content || '[暂不支持的消息]' }}
              </div>

              <time class="time-label">{{ formatMsgTime(msg.created_at) }}</time>
            </div>
          </article>
        </div>
      </section>
    </main>

    <footer class="chat-footer">
      <div class="container composer-shell" @click.stop>
        <transition name="fade-up">
          <div v-if="showEmoji" class="emoji-popover">
            <EmojiPicker :native="true" @select="onSelectEmoji" class="custom-picker" />
          </div>
        </transition>

        <section class="composer-card">
          <div class="tool-bar">
            <button class="tool-btn" type="button" title="表情" @click.stop="showEmoji = !showEmoji">
              <el-icon><ChatDotRound /></el-icon>
            </button>

            <el-upload
              :action="uploadUrl"
              name="file"
              :headers="uploadHeaders"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleImageUpload"
              :on-error="handleUploadError"
              class="upload-wrapper"
            >
              <button class="tool-btn" type="button" title="发送图片">
                <el-icon><Picture /></el-icon>
              </button>
            </el-upload>

            <button class="tool-btn reconnect-btn" type="button" title="重连" @click="manualReconnect">
              <el-icon><Refresh /></el-icon>
            </button>
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
              type="button"
              @click="sendMessage"
              :disabled="!canSend"
              :class="{ active: canSend }"
            >
              <el-icon><Promotion /></el-icon>
            </button>
          </div>
        </section>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import request, { resolveUrl } from '@/utils/request'
import { ElMessage } from 'element-plus'
import { ArrowLeft, ChatDotRound, Picture, Promotion, Refresh } from '@element-plus/icons-vue'
import EmojiPicker from 'vue3-emoji-picker'
import 'vue3-emoji-picker/css'

const route = useRoute()
const targetId = Number(route.params.id || route.params.targetId)
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const IMAGE_CONTENT_RE = /^(data:image\/|https?:\/\/|\/uploads\/|uploads\/)/i

const getUser = () => {
  try {
    return JSON.parse(localStorage.getItem('user') || '{}')
  } catch (error) {
    return {}
  }
}

const user = getUser()
const showEmoji = ref(false)
const targetUser = ref({})
const messages = ref([])
const inputText = ref('')
const msgBoxRef = ref(null)
const wsConnected = ref(false)
let socket = null
let reconnectTimer = null
let reconnectCount = 0

const canSend = computed(() => !!inputText.value.trim() && wsConnected.value)

const uploadUrl = computed(() => {
  const baseUrl = String(import.meta.env.VITE_API_URL || '/').trim()
  return baseUrl.endsWith('/') ? `${baseUrl}api/upload` : `${baseUrl}/api/upload`
})

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token')
  return { Authorization: token || '' }
})

const isMe = (senderId) => Number(senderId) === Number(user.id)

const formatMsgTime = (value) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}

const normalizeMessage = (raw) => {
  const message = raw && typeof raw === 'object' ? raw : {}
  const content = String(message.content ?? '').trim()
  const rawType = message.type ?? message.Type
  let typeNumber = Number(rawType)
  if (Number.isNaN(typeNumber)) {
    const typeText = String(rawType || '').toLowerCase()
    if (typeText === 'image' || typeText === 'img') {
      typeNumber = 2
    } else if (typeText === 'text' || typeText === 'emoji') {
      typeNumber = 1
    }
  }

  if (typeNumber !== 1 && typeNumber !== 2) {
    typeNumber = IMAGE_CONTENT_RE.test(content) ? 2 : 1
  }

  return {
    ...message,
    sender_id: Number(message.sender_id ?? message.SenderID ?? 0),
    receiver_id: Number(message.receiver_id ?? message.ReceiverID ?? 0),
    content,
    type: typeNumber,
    created_at: message.created_at || message.CreatedAt || new Date().toISOString()
  }
}

const isTextMessage = (msg) => normalizeMessage(msg).type === 1
const isImageMessage = (msg) => normalizeMessage(msg).type === 2

const scrollToBottom = () => {
  nextTick(() => {
    if (!msgBoxRef.value) return
    msgBoxRef.value.scrollTo({
      top: msgBoxRef.value.scrollHeight,
      behavior: 'smooth'
    })
  })
}

const onSelectEmoji = (emoji) => {
  const value = emoji?.i || emoji?.emoji || emoji?.native || (typeof emoji === 'string' ? emoji : '')
  if (!value) return
  inputText.value += value
}

const fetchTargetInfo = async () => {
  try {
    const res = await request.get(`/api/users/${targetId}`)
    targetUser.value = res.data || {}
  } catch (error) {
    console.error('获取对方信息失败', error)
    targetUser.value = { nickname: `用户 ${targetId}` }
  }
}

const fetchHistory = async () => {
  try {
    const res = await request.get('/api/chat/messages', { params: { target_id: targetId } })
    const rows = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : [])
    messages.value = rows.map(normalizeMessage).filter((msg) => msg.content)
    scrollToBottom()
  } catch (error) {
    console.error('获取历史消息失败', error)
  }
}

const buildWsUrl = () => {
  const token = localStorage.getItem('token')
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const rawApiBase = String(import.meta.env.VITE_API_URL || '').trim()
  let host = ''

  if (rawApiBase) {
    try {
      host = new URL(rawApiBase).host
    } catch (error) {
      host = rawApiBase.replace(/^https?:\/\//i, '').replace(/\/.*$/, '')
    }
  } else {
    host = import.meta.env.DEV ? 'localhost:8081' : window.location.host
  }

  return `${wsProtocol}://${host}/api/ws?token=${encodeURIComponent(token || '')}`
}

const clearReconnectTimer = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

const scheduleReconnect = () => {
  if (reconnectTimer || !localStorage.getItem('token')) return
  const delay = Math.min(8000, 1200 * Math.max(1, reconnectCount))
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    reconnectCount += 1
    initWebSocket()
  }, delay)
}

const initWebSocket = () => {
  const token = localStorage.getItem('token')
  if (!token) return

  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }

  socket = new WebSocket(buildWsUrl())

  socket.onopen = () => {
    wsConnected.value = true
    reconnectCount = 0
    clearReconnectTimer()
  }

  socket.onmessage = (event) => {
    try {
      const normalized = normalizeMessage(JSON.parse(event.data))
      const isRelevant =
        (normalized.sender_id === targetId && normalized.receiver_id === Number(user.id)) ||
        (normalized.sender_id === Number(user.id) && normalized.receiver_id === targetId)

      if (!isRelevant || !normalized.content) return
      messages.value.push(normalized)
      scrollToBottom()
    } catch (error) {
      console.error('消息解析失败', error)
    }
  }

  socket.onerror = () => {
    wsConnected.value = false
  }

  socket.onclose = () => {
    wsConnected.value = false
    scheduleReconnect()
  }
}

const sendWsMessage = (payload) => {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    ElMessage.warning('连接中断，正在重连')
    initWebSocket()
    return false
  }
  socket.send(JSON.stringify(payload))
  return true
}

const sendMessage = () => {
  const content = inputText.value.trim()
  if (!content) return

  const ok = sendWsMessage({
    receiver_id: targetId,
    content,
    type: 1
  })

  if (!ok) return
  inputText.value = ''
  showEmoji.value = false
}

const beforeUpload = (file) => {
  if (file.size > 8 * 1024 * 1024) {
    ElMessage.warning('图片不能超过 8MB')
    return false
  }
  return true
}

const handleImageUpload = (res) => {
  const payload = res?.url ? res : (res?.data || {})
  const finalUrl = resolveUrl(payload.url)
  if (!finalUrl) {
    ElMessage.error('图片上传响应异常')
    return
  }

  const ok = sendWsMessage({
    receiver_id: targetId,
    content: finalUrl,
    type: 2
  })

  if (!ok) return
  showEmoji.value = false
}

const handleUploadError = () => {
  ElMessage.error('图片上传失败，请重试')
}

const manualReconnect = () => {
  clearReconnectTimer()
  if (socket && socket.readyState === WebSocket.OPEN) {
    ElMessage.success('连接正常')
    return
  }
  reconnectCount = 1
  initWebSocket()
}

onMounted(async () => {
  await fetchTargetInfo()
  await fetchHistory()
  initWebSocket()
})

onUnmounted(() => {
  clearReconnectTimer()
  if (socket) socket.close()
})
</script>

<style scoped lang="scss">
@use "@/assets/tokens.scss" as *;

.chat-room-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background:
    url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.82' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E"),
    radial-gradient(circle at 82% -10%, rgba(255, 223, 93, 0.33), transparent 42%),
    linear-gradient(165deg, #f7f9fc 0%, #ebeff6 54%, #e3e8f1 100%);
  background-blend-mode: overlay, normal, normal;
}

.container {
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  padding: 0 18px;
  box-sizing: border-box;
}

.nav-header {
  height: 74px;
  position: sticky;
  top: 0;
  z-index: 120;
  background: rgba(255, 255, 255, 0.84);
  border-bottom: 1px solid rgba(16, 23, 36, 0.08);
  backdrop-filter: blur(16px);

  .nav-content {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .back-btn {
    width: 38px;
    height: 38px;
    border: 0;
    border-radius: 12px;
    background: rgba(18, 24, 36, 0.04);
    color: #1f2530;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    cursor: pointer;
    transition: all 0.22s cubic-bezier(0.22, 1, 0.36, 1);

    &:hover {
      transform: translateY(-1px);
      background: rgba(18, 24, 36, 0.08);
    }
  }

  .user-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;

    .name {
      font-size: 20px;
      line-height: 1;
      font-weight: 800;
      color: #1d2430;
    }

    .status-badge {
      font-size: 11px;
      font-weight: 700;
      color: #3d8f27;
      background: #effdde;
      border: 1px solid #b7eb8f;
      border-radius: 999px;
      padding: 2px 8px;
    }
  }

  .placeholder {
    width: 38px;
    height: 38px;
  }
}

.chat-main {
  flex: 1;
  display: flex;
  align-items: stretch;
  padding: 16px 18px 8px;
}

.conversation-card {
  width: 100%;
  border-radius: 28px;
  padding: 16px 14px;
  border: 1px solid rgba(255, 255, 255, 0.78);
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 24px 54px rgba(11, 19, 33, 0.08);
  backdrop-filter: blur(12px);
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.chat-tip {
  align-self: center;
  font-size: 12px;
  font-weight: 600;
  color: #7d8795;
  padding: 6px 14px;
  border-radius: 999px;
  background: rgba(20, 28, 40, 0.05);
  border: 1px solid rgba(20, 28, 40, 0.06);
}

.messages-wrap {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  margin-top: 14px;
  padding: 6px 6px 8px;
}

.empty-module {
  border-radius: 18px;
  border: 1px dashed rgba(23, 33, 48, 0.16);
  background: rgba(248, 251, 255, 0.75);
  padding: 24px 18px;
  text-align: center;
  margin-top: 12px;

  .empty-title {
    font-size: 15px;
    font-weight: 800;
    color: #242f3f;
  }

  .empty-desc {
    margin-top: 6px;
    font-size: 13px;
    color: #6f7b8e;
  }
}

.message-row {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  margin-bottom: 18px;

  .avatar-img {
    flex-shrink: 0;
    border: 2px solid rgba(255, 255, 255, 0.9);
    box-shadow: 0 8px 16px rgba(21, 27, 39, 0.12);
  }

  .bubble-wrapper {
    max-width: min(74%, 520px);
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .bubble {
    border-radius: 16px;
    padding: 11px 14px;
    font-size: 15px;
    line-height: 1.56;
    word-break: break-word;
  }

  .text-bubble {
    color: #1e2734;
    background: #ffffff;
    border: 1px solid rgba(24, 34, 50, 0.08);
    box-shadow: 0 12px 22px rgba(25, 34, 47, 0.08);
  }

  .image-bubble {
    width: clamp(160px, 26vw, 280px);
    border-radius: 16px;
    overflow: hidden;
    border: 1px solid rgba(22, 32, 47, 0.1);
    box-shadow: 0 14px 24px rgba(27, 36, 49, 0.18);
    cursor: zoom-in;
    transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1);

    &:hover {
      transform: translateY(-1px);
    }
  }

  .img-error {
    width: 100%;
    min-height: 110px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    color: #6e7b8d;
    background: #f0f3f8;
    font-size: 12px;
  }

  .time-label {
    font-size: 11px;
    color: #8b95a4;
    margin-left: 6px;
  }

  &.self {
    flex-direction: row-reverse;

    .bubble-wrapper {
      align-items: flex-end;
    }

    .text-bubble {
      color: #151514;
      background: linear-gradient(130deg, #ffe991 0%, #ffdf5d 100%);
      border: 1px solid rgba(42, 34, 8, 0.08);
      box-shadow: 0 12px 24px rgba(150, 114, 12, 0.2);
    }

    .time-label {
      margin-left: 0;
      margin-right: 6px;
    }
  }
}

.chat-footer {
  position: sticky;
  bottom: 0;
  z-index: 140;
  padding: 8px 0 16px;
}

.composer-shell {
  position: relative;
}

.composer-card {
  border-radius: 22px;
  border: 1px solid rgba(255, 255, 255, 0.8);
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 18px 38px rgba(20, 28, 42, 0.12);
  backdrop-filter: blur(12px);
  padding: 10px 12px 12px;
}

.tool-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 2px 2px 8px;
}

.tool-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  border: 1px solid rgba(21, 30, 44, 0.12);
  background: rgba(250, 252, 255, 0.9);
  color: #415069;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.22s cubic-bezier(0.22, 1, 0.36, 1);

  &:hover {
    transform: translateY(-1px);
    border-color: rgba(255, 212, 74, 0.76);
    color: #2a3342;
    background: #fff4d0;
  }
}

.reconnect-btn {
  margin-left: auto;
}

.input-row {
  display: flex;
  gap: 10px;
  align-items: center;

  input {
    flex: 1;
    height: 46px;
    border-radius: 14px;
    border: 1px solid rgba(22, 32, 46, 0.12);
    background: rgba(248, 250, 254, 0.9);
    color: #212b38;
    font-size: 15px;
    padding: 0 14px;
    outline: none;
    transition: all 0.24s cubic-bezier(0.22, 1, 0.36, 1);

    &:focus {
      border-color: rgba(255, 211, 73, 0.94);
      background: #fff;
      box-shadow: 0 0 0 4px rgba(255, 223, 93, 0.22);
    }
  }
}

.send-btn {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  border: 0;
  background: #e8ebf0;
  color: #94a0af;
  font-size: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: not-allowed;
  transition: all 0.22s cubic-bezier(0.22, 1, 0.36, 1);

  &.active {
    cursor: pointer;
    color: #181818;
    background: linear-gradient(130deg, #ffe991 0%, #ffdf5d 100%);
    box-shadow: 0 10px 20px rgba(145, 106, 5, 0.2);

    &:hover {
      transform: translateY(-1px) scale(1.02);
    }
  }
}

.emoji-popover {
  position: absolute;
  left: 16px;
  bottom: calc(100% + 10px);
  z-index: 150;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 24px 44px rgba(17, 24, 35, 0.2);
}

.animate-slide-in {
  animation: messageSlide 0.26s cubic-bezier(0.22, 1, 0.36, 1);
}

@keyframes messageSlide {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.fade-up-enter-active,
.fade-up-leave-active {
  transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}

.fade-up-enter-from,
.fade-up-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

@media (max-width: 680px) {
  .container {
    padding: 0 12px;
  }

  .nav-header {
    height: 66px;

    .user-info {
      .name {
        font-size: 18px;
      }
    }
  }

  .chat-main {
    padding: 12px 0 8px;
  }

  .conversation-card {
    border-radius: 18px;
    padding: 12px 10px;
  }

  .message-row {
    gap: 8px;
    margin-bottom: 14px;

    .bubble-wrapper {
      max-width: 82%;
    }
  }

  .composer-card {
    border-radius: 16px;
    padding: 8px 8px 10px;
  }

  .tool-btn {
    width: 32px;
    height: 32px;
    border-radius: 9px;
    font-size: 17px;
  }

  .input-row {
    gap: 8px;

    input {
      height: 42px;
      border-radius: 12px;
      font-size: 14px;
      padding: 0 12px;
    }
  }

  .send-btn {
    width: 42px;
    height: 42px;
    border-radius: 12px;
  }

  .emoji-popover {
    left: 8px;
    right: 8px;
    width: auto;
  }
}
</style>
