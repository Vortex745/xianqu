import request from '@/utils/request'

const HEARTBEAT_INTERVAL_MS = 15000

let heartbeatTimer = null

const hasToken = () => !!localStorage.getItem('token')

export const sendPresenceHeartbeat = async () => {
  if (!hasToken()) return false

  try {
    await request.post('/api/presence/heartbeat')
    return true
  } catch (error) {
    return false
  }
}

export const startPresenceHeartbeat = () => {
  if (heartbeatTimer || !hasToken()) return

  sendPresenceHeartbeat()
  heartbeatTimer = window.setInterval(() => {
    sendPresenceHeartbeat()
  }, HEARTBEAT_INTERVAL_MS)
}

export const stopPresenceHeartbeat = () => {
  if (!heartbeatTimer) return
  window.clearInterval(heartbeatTimer)
  heartbeatTimer = null
}

export const markPresenceOffline = () => {
  if (!hasToken()) return

  const token = localStorage.getItem('token')
  const apiBase = String(import.meta.env.VITE_API_URL || '').trim().replace(/\/$/, '')
  const endpoint = apiBase ? `${apiBase}/api/presence/offline` : '/api/presence/offline'

  if (navigator.sendBeacon) {
    navigator.sendBeacon(`${endpoint}?token=${encodeURIComponent(token)}`, new Blob([], { type: 'application/json' }))
    return
  }

  request.post('/api/presence/offline').catch(() => {})
}
