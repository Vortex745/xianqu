import axios from 'axios'
import { ElMessage } from 'element-plus'

const rawApiBase = String(import.meta.env.VITE_API_URL || '').trim()
const apiBase = rawApiBase.replace(/\/$/, '')
const LEGACY_LOCAL_ORIGIN_RE = /^https?:\/\/(?:localhost|127\.0\.0\.1)(?::\d+)?/i
const FRONTEND_STATIC_ASSET_PREFIXES = [
    '/avatars/ai/',
    '/avatars/user-default-',
    '/images/',
    '/favicon.ico',
    '/icon.ico'
]

const request = axios.create({
    baseURL: apiBase || '/',
    timeout: 15000
})

const isAbsoluteAssetUrl = (value = '') => {
    return /^https?:\/\//i.test(value) || value.startsWith('data:') || value.startsWith('blob:')
}

const rewriteLegacyLocalUrl = (value = '') => {
    if (!value || !apiBase || !LEGACY_LOCAL_ORIGIN_RE.test(value)) {
        return value
    }
    try {
        const legacyUrl = new URL(value)
        return `${apiBase}${legacyUrl.pathname}${legacyUrl.search}${legacyUrl.hash}`
    } catch (error) {
        return value.replace(LEGACY_LOCAL_ORIGIN_RE, apiBase)
    }
}

const isFrontendStaticAsset = (value = '') => {
    return FRONTEND_STATIC_ASSET_PREFIXES.some(prefix => value.startsWith(prefix))
}

export const resolveUrl = (path) => {
    const value = String(path || '').trim()
    if (!value) return ''
    if (isAbsoluteAssetUrl(value)) {
        return rewriteLegacyLocalUrl(value)
    }
    const cleanPath = value.startsWith('/') ? value : `/${value}`
    return apiBase ? `${apiBase}${cleanPath}` : cleanPath
}

export const resolveBackendAssetUrl = (path) => {
    if (!path) return ''
    const value = String(path).trim()
    if (!value) return ''
    if (isAbsoluteAssetUrl(value)) {
        return resolveUrl(value)
    }
    if (isFrontendStaticAsset(value)) {
        return value
    }
    return resolveUrl(value)
}

// 请求拦截器
request.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token')
        const adminToken = localStorage.getItem('admin_token')

        if (config.url.includes('/api/admin') && adminToken) {
            config.headers['Authorization'] = adminToken
        } else if (token) {
            config.headers['Authorization'] = token
        }
        return config
    },
    error => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
    response => response.data,
    error => {
        if (error.response) {
            const status = error.response.status
            const requestUrl = error.config.url

            // 1. 如果是其他接口报 401 -> 说明 Token 过期
            if (status === 401 && !requestUrl.includes('/login')) {
                if (!window.is401Shown) {
                    window.is401Shown = true
                    ElMessage.error('登录已过期，请重新登录')
                    localStorage.clear()
                    setTimeout(() => {
                        window.is401Shown = false
                        if (window.location.pathname.startsWith('/admin')) {
                            if (!window.location.pathname.includes('/login')) window.location.href = '/admin/login'
                        } else {
                            window.location.href = '/'
                        }
                    }, 1500)
                }
            } else {
                // 其他错误 (500等)
                ElMessage.error(error.response.data.error || '请求失败')
            }
        } else {
            ElMessage.error('网络连接异常')
        }
        return Promise.reject(error)
    }
)

export default request
