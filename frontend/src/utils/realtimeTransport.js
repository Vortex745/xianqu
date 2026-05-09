const HTTP_ONLY_HOSTS = new Set([
  '315279.xyz',
])

const isVercelHosted = (host) => {
  const hostname = String(host || '').split(':')[0].toLowerCase()
  return hostname.endsWith('.vercel.app') || HTTP_ONLY_HOSTS.has(hostname)
}

export const isHttpOnlyRealtimeHost = () => {
  if (typeof window === 'undefined') return false
  const explicitWsUrl = String(import.meta.env.VITE_WS_URL || '').trim()
  if (explicitWsUrl) return false

  const apiBase = String(import.meta.env.VITE_API_URL || '').trim()
  if (apiBase) {
    try {
      return isVercelHosted(new URL(apiBase, window.location.origin).host)
    } catch (error) {
      return isVercelHosted(apiBase)
    }
  }

  return !import.meta.env.DEV && isVercelHosted(window.location.host)
}

export const supportsWebSocketTransport = () => !isHttpOnlyRealtimeHost()

export const hasExplicitWebSocketUrl = () => String(import.meta.env.VITE_WS_URL || '').trim() !== ''

export const buildWebSocketUrl = (token) => {
  const explicitWsUrl = String(import.meta.env.VITE_WS_URL || '').trim()
  if (explicitWsUrl) {
    const url = new URL(explicitWsUrl, window.location.origin)
    url.searchParams.set('token', token || '')
    return url.toString()
  }

  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const rawApiBase = String(import.meta.env.VITE_API_URL || '').trim()
  let host = ''

  if (rawApiBase) {
    try {
      host = new URL(rawApiBase, window.location.origin).host
    } catch (error) {
      host = rawApiBase.replace(/^https?:\/\//i, '').replace(/\/.*$/, '')
    }
  } else {
    host = import.meta.env.DEV ? 'localhost:8081' : window.location.host
  }

  return `${wsProtocol}://${host}/api/ws?token=${encodeURIComponent(token || '')}`
}
