import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/',
    timeout: 15000
})

export const resolveUrl = (path) => {
    if (!path) return ''
    if (path.startsWith('http') || path.startsWith('https') || path.startsWith('data:') || path.startsWith('blob:')) {
        return path.replace('localhost', '127.0.0.1') // Cleanup for local consistency
    }
    const apiBase = import.meta.env.VITE_API_URL || ''
    const cleanBase = apiBase.endsWith('/') ? apiBase.slice(0, -1) : apiBase
    const cleanPath = path.startsWith('/') ? path : `/${path}`
    return `${cleanBase}${cleanPath}`
}

export const resolveBackendAssetUrl = (path) => {
    if (!path) return ''
    const value = String(path).trim()
    if (!value) return ''
    if (
        value.startsWith('http') ||
        value.startsWith('https') ||
        value.startsWith('data:') ||
        value.startsWith('blob:')
    ) {
        return resolveUrl(value)
    }
    if (
        value.startsWith('/avatars/') ||
        value.startsWith('/images/') ||
        value.startsWith('/assets/')
    ) {
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
