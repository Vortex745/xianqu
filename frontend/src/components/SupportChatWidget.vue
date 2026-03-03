<template>
  <div class="support-widget-root">
    <button class="support-fab" type="button" @click="handleOpenDialog" aria-label="打开AI助手" title="AI助手">
      <img :src="aiAvatars.normal64" alt="在线客服" class="fab-avatar" />
    </button>

    <transition name="chat-veil">
      <div v-if="openDialog" class="dialog-mask" @click.self="closeDialog">
        <section class="chat-dialog" role="dialog" aria-modal="true" :aria-label="assistantTitle">
          <header class="chat-head">
            <div class="head-left">
              <img :src="assistantHeaderAvatar" alt="闲趣AI客服头像" class="head-avatar" />
              <div class="head-text">
                <h3>{{ assistantTitle }}</h3>
                <p>{{ assistantStatusText }}</p>
                <div class="mode-switch" role="tablist" aria-label="对话模式切换">
                  <button
                    type="button"
                    class="mode-tab"
                    :class="{ active: activeMode === MODE_SUPPORT }"
                    :disabled="isSending"
                    @click="switchMode(MODE_SUPPORT)"
                  >
                    AI客服
                  </button>
                  <button
                    type="button"
                    class="mode-tab"
                    :class="{ active: activeMode === MODE_AGENT }"
                    :disabled="isSending"
                    @click="switchMode(MODE_AGENT)"
                  >
                    AI Agent
                  </button>
                </div>
              </div>
            </div>
            <div class="head-right">
              <button class="icon-btn" type="button" @click="closeDialog" aria-label="关闭对话" title="关闭对话">
                <Icon icon="solar:close-circle-bold-duotone" />
              </button>
            </div>
          </header>

          <div class="chat-scroll-wrapper">
            <div ref="listRef" class="chat-scroll" @scroll="handleScroll">
              <template v-for="(msg, index) in activeMessages" :key="msg.id">
                <div v-if="shouldShowDateSeparator(index)" class="time-divider">
                  {{ formatFullDate(msg.createdAt) }}
                </div>

                <article class="bubble-row" :class="msg.role === 'user' ? 'from-user' : 'from-assistant'">
                  <img class="bubble-avatar" :src="getMessageAvatar(msg)" :alt="msg.role === 'user' ? '用户头像' : '客服头像'" />

                  <div class="bubble-stack">
                    <div class="bubble-text" v-if="msg.text" :class="{ 'bubble-error': msg.state === 'offline' && msg.role === 'assistant' }">
                      {{ msg.text }}
                    </div>
                    <div class="bubble-meta">
                      <span v-if="showTimestampSetting" class="time-text">{{ formatTimeHHMM(msg.createdAt) }}</span>
                      <span
                        v-if="msg.role === 'user' && msg.status === 'failed'"
                        class="msg-status status-failed"
                      >
                        发送失败
                      </span>
                    </div>
                    <button
                      v-if="msg.role === 'user' && msg.status === 'failed'"
                      class="retry-btn"
                      type="button"
                      @click="retryMessage(msg)"
                      title="重新发送"
                    >
                      <Icon icon="solar:restart-bold" /> 重试
                    </button>
                  </div>
                </article>
              </template>

              <div v-if="isSending" class="typing-row">
                <img :src="aiAvatars.loading64" alt="AI输入中" class="typing-avatar" />
                <span>{{ typingLabel }}<span class="typing-dots"></span></span>
              </div>
            </div>

            <button v-if="hasNewMessage" class="new-message-fab" @click="scrollToBottomEvent">
              <Icon icon="solar:arrow-down-linear" /> 有新消息
            </button>
          </div>

          <footer class="chat-input-zone">
            <div class="input-actions-row">
              <div class="input-container">
                <textarea
                  v-model="draft"
                  class="chat-input"
                  rows="1"
                  maxlength="2000"
                  :disabled="isSending"
                  :placeholder="inputPlaceholder"
                  @keydown.enter.exact.prevent="sendMessage"
                ></textarea>
              </div>

              <button class="send-btn" type="button" @click="sendMessage" :disabled="!canSend">
                发送
              </button>
            </div>
          </footer>
        </section>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Icon } from '@iconify/vue'
import { resolveBackendAssetUrl } from '@/utils/request'

const resolveAIBase = () => {
  const apiBase = String(import.meta.env.VITE_API_URL || '').trim()
  const fallbackBase = apiBase ? `${apiBase.replace(/\/$/, '')}/ai` : '/ai'
  const explicit = String(import.meta.env.VITE_AI_BASE_URL || '').trim()
  if (explicit) {
    const normalized = explicit.replace(/\/$/, '')
    // Force legacy direct host to use backend /ai proxy to avoid CORS preflight failures.
    if (/^https?:\/\/ai\.530745\.xyz(?:\/.*)?$/i.test(normalized)) {
      return fallbackBase
    }
    return normalized
  }

  return fallbackBase
}

const apiBase = resolveAIBase()
const isCrossOriginAIRequest = () => {
  if (typeof window === 'undefined') return false
  try {
    return new URL(apiBase, window.location.origin).origin !== window.location.origin
  } catch (error) {
    return false
  }
}

const isLikelyCorsOrPreflightFailure = (error) => {
  const message = String(error?.message || '').trim().toLowerCase()
  if (!message) return false
  if (message.includes('cors') || message.includes('preflight') || message.includes('disallowed cors origin')) {
    return true
  }
  return isCrossOriginAIRequest() && (error?.name === 'TypeError' || error instanceof TypeError) && message.includes('failed to fetch')
}

const route = useRoute()
const router = useRouter()
const MODE_SUPPORT = 'support'
const MODE_AGENT = 'agent'
const MODE_LIST = [MODE_SUPPORT, MODE_AGENT]

const MODE_META = {
  [MODE_SUPPORT]: {
    title: '闲趣AI客服',
    typingLabel: '客服正在输入',
    inputPlaceholder: '请输入您的问题...',
    endpoint: '/chat',
    greeting: '你好，这里是闲趣AI客服，请问有什么可以帮到您'
  },
  [MODE_AGENT]: {
    title: '闲趣AI Agent',
    typingLabel: 'Agent 正在处理',
    inputPlaceholder: '说出你想执行的操作，比如“帮我查运动分类商品”',
    endpoint: '/agent/chat',
    greeting: '你好，我是闲趣AI Agent，可以帮你查商品、改资料、管订单和购物车。'
  }
}

const aiAvatars = {
  normal64: '/avatars/ai/ai-assistant-normal-64.png',
  normal128: '/avatars/ai/ai-assistant-normal-128.png',
  loading64: '/avatars/ai/ai-assistant-loading-64.png',
  loading128: '/avatars/ai/ai-assistant-loading-128.png',
  offline64: '/avatars/ai/ai-assistant-offline-64.png',
  offline128: '/avatars/ai/ai-assistant-offline-128.png'
}

const defaultUserAvatar = '/avatars/user-default-64.png'
const openDialog = ref(false)
const isSending = ref(false)
const draft = ref('')
const listRef = ref(null)
const assistantState = ref('normal')
const currentUser = ref(null)
const showTimestampSetting = ref(true)
const isAtBottom = ref(true)
const hasNewMessage = ref(false)
const activeMode = ref(MODE_SUPPORT)
const sessionByMode = ref({
  [MODE_SUPPORT]: '',
  [MODE_AGENT]: ''
})

const buildInitialMessages = (mode) => ([
  {
    id: `${mode}-init-1`,
    role: 'assistant',
    text: MODE_META[mode]?.greeting || MODE_META[MODE_SUPPORT].greeting,
    state: 'normal',
    createdAt: Date.now() - 3000
  }
])

const messagesByMode = ref({
  [MODE_SUPPORT]: buildInitialMessages(MODE_SUPPORT),
  [MODE_AGENT]: buildInitialMessages(MODE_AGENT)
})

const ensureModeMessages = (mode) => {
  if (!Array.isArray(messagesByMode.value[mode])) {
    messagesByMode.value[mode] = buildInitialMessages(mode)
  }
  return messagesByMode.value[mode]
}

const activeMessages = computed(() => ensureModeMessages(activeMode.value))
const assistantTitle = computed(() => MODE_META[activeMode.value]?.title || MODE_META[MODE_SUPPORT].title)
const typingLabel = computed(() => MODE_META[activeMode.value]?.typingLabel || MODE_META[MODE_SUPPORT].typingLabel)
const inputPlaceholder = computed(() => MODE_META[activeMode.value]?.inputPlaceholder || MODE_META[MODE_SUPPORT].inputPlaceholder)

const assistantHeaderAvatar = computed(() => {
  if (assistantState.value === 'loading') return aiAvatars.loading128
  if (assistantState.value === 'offline') return aiAvatars.offline128
  return aiAvatars.normal128
})

const assistantStatusText = computed(() => {
  if (assistantState.value === 'loading') return '处理中'
  if (assistantState.value === 'offline') return '离线'
  return '在线中'
})

const canSend = computed(() => !!draft.value.trim() && !isSending.value)

const normalizeToken = (raw = '') => {
  const value = String(raw || '').trim()
  if (!value) return ''
  return value.toLowerCase().startsWith('bearer ') ? value.slice(7).trim() : value
}

const isTokenExpired = (token = '') => {
  try {
    const payload = token.split('.')[1]
    if (!payload) return true
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    if (!decoded?.exp) return false
    return Date.now() >= decoded.exp * 1000
  } catch (error) {
    return true
  }
}

const parseStoredUser = () => {
  try {
    return JSON.parse(localStorage.getItem('user') || 'null')
  } catch (error) {
    return null
  }
}

const hasValidLogin = () => {
  const token = normalizeToken(localStorage.getItem('token'))
  return !!token && !isTokenExpired(token)
}

const getSessionStorageKey = (mode) => {
  const user = parseStoredUser()
  return user && user.id ? `xianqu_ai_session_id_${user.id}_${mode}` : `xianqu_ai_session_id_guest_${mode}`
}

const getMessageStorageKey = (mode) => {
  const user = parseStoredUser()
  return user && user.id ? `xianqu_ai_messages_${user.id}_${mode}` : `xianqu_ai_messages_guest_${mode}`
}

const SEVENTY_TWO_HOURS = 72 * 60 * 60 * 1000

const syncUserFromStorage = () => {
  currentUser.value = parseStoredUser()
}

const goToLogin = () => {
  ElMessage.warning('请先登录后咨询客服')
  router.push({
    path: '/',
    query: {
      ...route.query,
      auth: 'login',
      redirect: route.fullPath || '/'
    }
  })
}

const ensureLoggedIn = () => {
  syncUserFromStorage()
  if (hasValidLogin()) return true
  openDialog.value = false
  goToLogin()
  return false
}

const resolveUserAvatar = () => {
  const source = currentUser.value || {}
  return resolveBackendAssetUrl(source.avatar_url || source.avatar) || defaultUserAvatar
}

const getMessageAvatar = (msg) => {
  if (msg.role === 'user') return resolveUserAvatar()
  if (msg.state === 'offline') return aiAvatars.offline64
  if (msg.state === 'loading') return aiAvatars.loading64
  return aiAvatars.normal64
}

const forcePlainText = (rawText = '') => {
  const text = String(rawText || '')
  return text
    .replace(/```[\s\S]*?```/g, (chunk) => chunk.replace(/```/g, ''))
    .replace(/<[^>]+>/g, '')
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, (_, alt, url) => `${alt || '图片'}（详见：${url}）`)
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, title, url) => `${title}（详见：${url}）`)
    .replace(/^#{1,6}\s*/gm, '')
    .replace(/^\s*[-*+]\s+/gm, '')
    .replace(/\*\*|__|`|~~/g, '')
    .replace(/\r\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

const normalizeMessage = (item, index = 0) => ({
  id: item.id || `${Date.now()}-${index}`,
  role: item.role === 'user' ? 'user' : 'assistant',
  text: forcePlainText(item.text || ''),
  state: item.state || 'normal',
  status: item.status || (item.role === 'user' ? 'replied' : undefined),
  createdAt: Number(item.createdAt) || (Date.now() + index)
})

const pushMessage = (role, text, options = {}) => {
  const mode = options.mode || activeMode.value
  const targetMessages = ensureModeMessages(mode)
  const message = normalizeMessage(
    {
      id: options.id || `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      role,
      text: text || '我刚才没收到内容，请重试。',
      state: options.state || 'normal',
      status: options.status
    },
    targetMessages.length
  )
  targetMessages.push(message)
  return message.id
}

const updateMessageStatus = (messageId, status, mode = activeMode.value) => {
  const target = ensureModeMessages(mode).find((item) => item.id === messageId)
  if (target && target.role === 'user') {
    target.status = status
  }
}

const formatTimeHHMM = (timestamp) => {
  const date = new Date(Number(timestamp) || Date.now())
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false })
}

const formatFullDate = (timestamp) => {
  const date = new Date(Number(timestamp) || Date.now())
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}年${month}月${day}日`
}

const shouldShowDateSeparator = (index) => {
  if (index === 0) return true
  const current = new Date(Number(activeMessages.value[index]?.createdAt || 0))
  const previous = new Date(Number(activeMessages.value[index - 1]?.createdAt || 0))
  return current.toDateString() !== previous.toDateString()
}

const userStatusLabel = (status) => {
  if (status === 'pending') return '发送中'
  if (status === 'failed') return '发送失败'
  return '已回复'
}

const handleScroll = () => {
  if (!listRef.value) return
  const { scrollTop, scrollHeight, clientHeight } = listRef.value
  isAtBottom.value = Math.abs(scrollHeight - scrollTop - clientHeight) < 20
  if (isAtBottom.value) {
    hasNewMessage.value = false
  }
}

const scrollBottom = async () => {
  await nextTick()
  if (!listRef.value) return
  listRef.value.scrollTop = listRef.value.scrollHeight
  isAtBottom.value = true
  hasNewMessage.value = false
}

const scrollToBottomEvent = async () => {
  hasNewMessage.value = false
  await scrollBottom()
}

const loadMessagesForMode = (mode) => {
  try {
    if (!hasValidLogin()) {
      messagesByMode.value[mode] = buildInitialMessages(mode)
      return
    }

    const key = getMessageStorageKey(mode)
    const raw = localStorage.getItem(key)
    if (!raw) {
      messagesByMode.value[mode] = buildInitialMessages(mode)
      return
    }

    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed) || parsed.length === 0) {
      messagesByMode.value[mode] = buildInitialMessages(mode)
      return
    }

    const now = Date.now()
    const validMessages = parsed.filter(item => {
      return (now - (Number(item.createdAt) || 0)) <= SEVENTY_TWO_HOURS
    })

    if (validMessages.length === 0) {
      messagesByMode.value[mode] = buildInitialMessages(mode)
    } else {
      messagesByMode.value[mode] = validMessages.map((item, index) => normalizeMessage(item, index))
    }
  } catch (error) {
    messagesByMode.value[mode] = buildInitialMessages(mode)
  }
}

const persistMessagesForMode = (mode) => {
  if (!hasValidLogin()) {
    return
  }

  const currentMessages = ensureModeMessages(mode)
  const now = Date.now()
  const validMessages = currentMessages.filter(item => {
    return (now - (Number(item.createdAt) || 0)) <= SEVENTY_TWO_HOURS
  })

  if (validMessages.length !== currentMessages.length) {
    messagesByMode.value[mode] = validMessages.length ? validMessages : buildInitialMessages(mode)
  }

  const key = getMessageStorageKey(mode)
  localStorage.setItem(key, JSON.stringify(messagesByMode.value[mode]))
}

const resetAllModeState = () => {
  MODE_LIST.forEach((mode) => {
    messagesByMode.value[mode] = buildInitialMessages(mode)
    sessionByMode.value[mode] = ''
  })
}

const checkAndEnforcePolicy = () => {
  syncUserFromStorage()
  if (!hasValidLogin()) {
    const keysToRemove = []
    for (let i = 0; i < localStorage.length; i++) {
        const k = localStorage.key(i)
        if (k && (k.startsWith('xianqu_ai_messages_') || k.startsWith('xianqu_ai_session_id_'))) {
            keysToRemove.push(k)
        }
    }
    keysToRemove.forEach(k => localStorage.removeItem(k))
    resetAllModeState()
    return
  }

  MODE_LIST.forEach((mode) => {
    const key = getMessageStorageKey(mode)
    const raw = localStorage.getItem(key)
    if (!raw) return

    try {
      const parsed = JSON.parse(raw)
      if (!Array.isArray(parsed)) return
      const now = Date.now()
      const validMessages = parsed.filter(item => (now - (Number(item.createdAt) || 0)) <= SEVENTY_TWO_HOURS)
      localStorage.setItem(key, JSON.stringify(validMessages))
      if (mode === activeMode.value) {
        const current = ensureModeMessages(mode)
        const inMemoryValid = current.filter(item => (now - (Number(item.createdAt) || 0)) <= SEVENTY_TWO_HOURS)
        if (inMemoryValid.length !== current.length) {
          messagesByMode.value[mode] = inMemoryValid.length ? inMemoryValid : buildInitialMessages(mode)
        }
      }
    } catch (e) {}
  })
}

const handleOpenDialog = async () => {
  if (!ensureLoggedIn()) return
  openDialog.value = true
  assistantState.value = 'normal'
  loadMessagesForMode(activeMode.value)
  hasNewMessage.value = false
  await scrollBottom()
  setTimeout(handleScroll, 100)
}

const switchMode = async (mode) => {
  if (!MODE_LIST.includes(mode) || mode === activeMode.value || isSending.value) return
  activeMode.value = mode
  assistantState.value = 'normal'
  hasNewMessage.value = false

  if (!hasValidLogin()) {
    resetAllModeState()
    return
  }

  sessionByMode.value[mode] = localStorage.getItem(getSessionStorageKey(mode)) || sessionByMode.value[mode] || ''
  loadMessagesForMode(mode)
  await scrollBottom()
  setTimeout(handleScroll, 80)
}

const closeDialog = () => {
  openDialog.value = false
}

const parseSendErrorMessage = (error) => {
  if (isLikelyCorsOrPreflightFailure(error)) {
    return '跨域或预检失败，请稍后重试'
  }
  const message = String(error?.message || '').trim()
  const lower = message.toLowerCase()
  if (lower.includes('failed to fetch') || lower.includes('networkerror') || lower.includes('连接')) {
    return '网络连接异常，请检查网络后重试'
  }
  if (lower.includes('request failed: 503') || lower.includes('missing deepseek_api_key')) {
    return 'AI 助手暂时不在线，请稍后再试'
  }
  if (lower.includes('request failed: 502') || lower.includes('upstream connection failed')) {
    return 'AI 助手暂时无法响应，请稍后再试'
  }
  if (lower.includes('request failed: 500')) {
    return '抱歉，AI 助手遇到了一点问题，请稍后再试'
  }
  if (lower.includes('request failed: 404')) {
    return 'AI 服务地址未配置正确，请联系管理员'
  }
  if (lower.includes('request failed: 401') || lower.includes('request failed: 403')) {
    return '登录状态已过期，请重新登录后再试'
  }
  if (lower.includes('llm invoke failed')) {
    return 'AI 助手繁忙中，请稍后再试'
  }
  return '消息发送失败，请稍后重试'
}

const notifyAgentResult = (rawText = '') => {
  const text = forcePlainText(rawText)
  if (!text || activeMode.value !== MODE_AGENT) return
  if (text.includes('加入购物车')) {
    if (text.includes('未找到') || text.includes('没找到') || text.includes('失败')) {
      ElMessage.warning(text)
      return
    }
    if (text.includes('成功') || text.includes('已')) {
      ElMessage.success(text)
    }
  }
}

const sendMessage = async () => {
  if (!ensureLoggedIn() || !canSend.value) return

  const mode = activeMode.value
  const text = draft.value.trim()
  const pendingId = pushMessage('user', text, { status: 'pending', mode })
  draft.value = ''
  isSending.value = true
  assistantState.value = 'loading'
  await scrollBottom()

  try {
    const payload = {
      session_id: sessionByMode.value[mode] || undefined,
      message: text
    }

    if (mode === MODE_AGENT) {
      const token = normalizeToken(localStorage.getItem('token'))
      if (token) payload.auth_token = token
    }

    const response = await fetch(`${apiBase}${MODE_META[mode].endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })

    if (!response.ok) {
      let detail = ''
      try {
        detail = (await response.text()).trim()
      } catch (e) {}
      throw new Error(`request failed: ${response.status}${detail ? ` ${detail}` : ''}`)
    }

    const data = await response.json()
    if (data?.session_id) {
      sessionByMode.value[mode] = data.session_id
      localStorage.setItem(getSessionStorageKey(mode), data.session_id)
    }
    updateMessageStatus(pendingId, 'replied', mode)
    const answerText = data?.answer || '我这边没拿到回复。'
    pushMessage('assistant', answerText, { state: 'normal', mode })
    notifyAgentResult(answerText)
    assistantState.value = 'normal'
  } catch (error) {
    console.error('chat request failed:', error)
    const errorHint = parseSendErrorMessage(error)
    updateMessageStatus(pendingId, 'failed', mode)
    pushMessage(
      'assistant',
      errorHint,
      { state: 'offline', mode }
    )
    assistantState.value = 'offline'
  } finally {
    isSending.value = false
    await scrollBottom()
  }
}

const retryMessage = async (failedMsg) => {
  if (isSending.value || !failedMsg?.text) return
  const mode = activeMode.value
  const msgs = ensureModeMessages(mode)
  // Find and remove the failed user message and its companion error reply
  const failedIndex = msgs.findIndex(m => m.id === failedMsg.id)
  if (failedIndex >= 0) {
    // Remove the error reply right after it (if it exists)
    if (failedIndex + 1 < msgs.length && msgs[failedIndex + 1].role === 'assistant' && msgs[failedIndex + 1].state === 'offline') {
      msgs.splice(failedIndex, 2)
    } else {
      msgs.splice(failedIndex, 1)
    }
  }
  // Re-send with the original text
  draft.value = failedMsg.text
  await nextTick()
  await sendMessage()
}

const onStorageChange = () => {
  syncUserFromStorage()
  checkAndEnforcePolicy()
  loadMessagesForMode(activeMode.value)
}

let policyTimer = null

watch(
  () => activeMessages.value.length,
  async (newLen, oldLen) => {
    persistMessagesForMode(activeMode.value)
    if (newLen > (oldLen || 0)) {
      if (!isAtBottom.value && openDialog.value) {
        hasNewMessage.value = true
      } else {
        await scrollBottom()
      }
    }
  }
)

watch(
  activeMessages,
  () => {
    persistMessagesForMode(activeMode.value)
  },
  { deep: true }
)

watch(activeMode, async () => {
  hasNewMessage.value = false
  await scrollBottom()
  setTimeout(handleScroll, 80)
})

onMounted(() => {
  syncUserFromStorage()
  checkAndEnforcePolicy()
  MODE_LIST.forEach((mode) => {
    sessionByMode.value[mode] = localStorage.getItem(getSessionStorageKey(mode)) || ''
    loadMessagesForMode(mode)
  })
  window.addEventListener('storage', onStorageChange)
  policyTimer = setInterval(checkAndEnforcePolicy, 60000)
})

onBeforeUnmount(() => {
  window.removeEventListener('storage', onStorageChange)
  if (policyTimer) clearInterval(policyTimer)
})
</script>

<style scoped>
.support-widget-root {
  position: relative;
  z-index: 70;
  --brand-primary: var(--ui-primary, #ffdf5d);
  --brand-secondary: #edb79c;
  --brand-dark: var(--ui-dark, #14171f);
  --brand-panel: var(--ui-panel, rgba(255, 255, 255, 0.9));
  --brand-border: var(--ui-border, rgba(13, 18, 27, 0.08));
  --brand-text: var(--ui-text, #24262b);
  --brand-muted: var(--ui-muted, #79808d);
}

.support-fab {
  position: fixed;
  right: 36px;
  bottom: 18px;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.65);
  background:
    radial-gradient(circle at 28% 26%, rgba(255, 255, 255, 0.62), rgba(255, 255, 255, 0) 58%),
    linear-gradient(140deg, var(--brand-primary), #f1d66f 58%, var(--brand-secondary));
  box-shadow: 0 14px 34px rgba(7, 9, 14, 0.2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 260ms cubic-bezier(0.21, 0.87, 0.29, 1), box-shadow 260ms cubic-bezier(0.21, 0.87, 0.29, 1);
}

.support-fab:hover {
  transform: translateY(-2px);
  box-shadow: 0 18px 40px rgba(7, 9, 14, 0.24);
}

.fab-avatar {
  width: 26px;
  height: 26px;
}

.dialog-mask {
  position: fixed;
  inset: 0;
  display: flex;
  justify-content: flex-end;
  align-items: flex-end;
  padding: 16px;
  background: rgba(8, 12, 20, 0.52);
  backdrop-filter: blur(8px);
}

.chat-dialog {
  width: min(430px, calc(100vw - 12px));
  height: min(660px, calc(100vh - 12px));
  display: flex;
  flex-direction: column;
  border-radius: 24px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.6);
  background: rgba(255, 255, 255, 0.55);
  backdrop-filter: blur(28px);
  -webkit-backdrop-filter: blur(28px);
}

.chat-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.35);
  background: rgba(255, 255, 255, 0.15);
}

.head-right {
  display: flex;
  align-items: center;
}

.head-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.head-avatar {
  width: 44px;
  height: 44px;
}

.head-text h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 800;
  color: var(--brand-text);
}

.head-text p {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--brand-muted);
  font-weight: 600;
}

.mode-switch {
  margin-top: 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px;
  border-radius: 999px;
  background: rgba(20, 23, 31, 0.08);
}

.mode-tab {
  border: none;
  background: transparent;
  color: var(--brand-muted);
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: background-color 180ms cubic-bezier(0.25, 0.84, 0.28, 1), color 180ms cubic-bezier(0.25, 0.84, 0.28, 1);
}

.mode-tab.active {
  background: rgba(255, 255, 255, 0.92);
  color: var(--brand-text);
}

.mode-tab:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.icon-btn {
  width: 32px;
  height: 32px;
  border: 1px solid var(--brand-border);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.7);
  color: var(--brand-text);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  cursor: pointer;
  transition: transform 220ms cubic-bezier(0.25, 0.84, 0.28, 1), background-color 220ms cubic-bezier(0.25, 0.84, 0.28, 1);
}

.icon-btn:hover {
  transform: scale(1.05);
  background: #ffffff;
}

.chat-scroll-wrapper {
  position: relative;
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.chat-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 14px 12px;
}

.time-divider {
  margin: 8px 0;
  text-align: center;
  font-size: 11px;
  color: var(--brand-muted);
}

.bubble-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 10px;
}

.from-user {
  flex-direction: row-reverse;
}

.bubble-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--brand-border);
  flex-shrink: 0;
}

.bubble-stack {
  max-width: 82%;
  display: flex;
  flex-direction: row;
  align-items: flex-end;
  gap: 6px;
}

.from-user .bubble-stack {
  flex-direction: row-reverse;
}

.bubble-text {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.45;
  white-space: pre-wrap;
  word-break: break-word;
  flex: 0 1 auto;
}

.from-assistant .bubble-text {
  background: #F5F5F5 !important;
  color: var(--brand-text);
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.from-user .bubble-text {
  background: #E3F2FD !important;
  color: #1a1a1a;
  border: 1px solid rgba(227, 242, 253, 0.8);
}

.bubble-meta {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  margin-bottom: 2px;
  font-size: 11px;
  color: var(--brand-muted);
  flex: 0 0 auto;
}

.from-user .bubble-meta {
  align-items: flex-end;
}

.msg-status {
  font-weight: 700;
}

.status-pending {
  color: #a57e22;
}

.status-failed {
  color: #d14343;
}

.bubble-error {
  background: #FFF5F5 !important;
  border: 1px solid rgba(209, 67, 67, 0.15) !important;
  color: #8b3a3a;
}

.retry-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 10px;
  border: 1px solid rgba(209, 67, 67, 0.25);
  border-radius: 12px;
  background: rgba(255, 245, 245, 0.8);
  color: #c0392b;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  margin-bottom: 2px;
  transition: background 180ms, border-color 180ms;
  flex-shrink: 0;
}

.retry-btn:hover {
  background: rgba(255, 235, 235, 1);
  border-color: rgba(209, 67, 67, 0.45);
}

.retry-btn .iconify {
  font-size: 12px;
}

.status-replied,
.status-sent {
  color: #49966d;
}

.typing-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--brand-muted);
  margin-left: 2px;
}

.typing-dots::after {
  content: '';
  animation: typingDots 1.5s infinite steps(4, end);
}

@keyframes typingDots {
  0%, 20% { content: ''; }
  40% { content: '.'; }
  60% { content: '..'; }
  80%, 100% { content: '...'; }
}

.typing-avatar {
  width: 24px;
  height: 24px;
}

.chat-input-zone {
  border-top: 1px solid rgba(255, 255, 255, 0.4);
  padding: 10px 12px 12px;
  background: rgba(255, 255, 255, 0.25);
}

.input-actions-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.input-container {
  flex: 1;
  display: flex;
  position: relative;
}

.chat-input {
  width: 100%;
  resize: none;
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 20px;
  padding: 8px 14px;
  color: var(--brand-text);
  background: rgba(255, 255, 255, 0.65);
  box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.04);
  font-size: 14px;
  line-height: normal;
  min-height: 38px;
  font-family: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  outline: none;
  transition: border-color 180ms cubic-bezier(0.21, 0.87, 0.29, 1), box-shadow 180ms cubic-bezier(0.21, 0.87, 0.29, 1), opacity 180ms cubic-bezier(0.21, 0.87, 0.29, 1);
  overflow: hidden;
}

.chat-input:focus {
  border-color: rgba(255, 223, 93, 0.95);
  box-shadow: 0 0 0 3px rgba(255, 223, 93, 0.22);
}

.chat-input:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.ghost-btn,
.send-btn {
  height: 36px;
  border-radius: 18px;
  border: 1px solid transparent;
  font-size: 13px;
  font-weight: 800;
  cursor: pointer;
  transition: transform 200ms cubic-bezier(0.25, 0.84, 0.28, 1), opacity 200ms cubic-bezier(0.25, 0.84, 0.28, 1);
  padding: 0 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.ghost-btn.icon-only {
  width: 36px;
  padding: 0;
  border-radius: 50%;
}

.ghost-btn {
  color: var(--brand-text);
  background: #ffffff;
  border-color: var(--brand-border);
}

.send-btn {
  color: var(--brand-dark);
  background: linear-gradient(130deg, var(--brand-primary), #f3d86f 55%, var(--brand-secondary));
  padding: 0 18px;
}

.send-btn:hover:enabled,
.ghost-btn:hover:enabled {
  transform: translateY(-1px);
}

.send-btn:disabled,
.ghost-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.chat-veil-enter-active,
.chat-veil-leave-active {
  transition: opacity 300ms cubic-bezier(0.24, 0.82, 0.29, 1);
}

.chat-veil-enter-active .chat-dialog,
.chat-veil-leave-active .chat-dialog {
  transition: transform 400ms cubic-bezier(0.16, 1, 0.3, 1), opacity 400ms cubic-bezier(0.16, 1, 0.3, 1);
}

.chat-veil-enter-from,
.chat-veil-leave-to {
  opacity: 0;
}

.chat-veil-enter-from .chat-dialog,
.chat-veil-leave-to .chat-dialog {
  transform: translateX(40px) scale(0.95);
  opacity: 0;
}

@media (max-width: 767px) {
  .chat-veil-enter-from .chat-dialog,
  .chat-veil-leave-to .chat-dialog {
    transform: translateY(60px) scale(0.98);
  }
}

@media (max-width: 767px) {
  .support-fab {
    right: 18px;
    bottom: 18px;
  }

  .dialog-mask {
    padding: 0;
  }

  .chat-dialog {
    width: 100vw;
    height: 100vh;
    border-radius: 0;
  }
  
  .chat-head {
    border-radius: 0;
  }
}

.new-message-fab {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--brand-panel);
  color: var(--brand-text);
  padding: 8px 16px;
  border-radius: 20px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  cursor: pointer;
  font-size: 13px;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  border: 1px solid var(--brand-border);
}

.new-message-fab:hover {
  background: var(--brand-primary);
  color: var(--brand-dark);
}

.time-text {
  font-size: 11px;
}
</style>
