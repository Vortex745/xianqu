import axios from 'axios'
import { ElMessage } from 'element-plus'

const service = axios.create({
    baseURL: '/',
    timeout: 5000
})

service.interceptors.request.use(
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

service.interceptors.response.use(
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

export default service