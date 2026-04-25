import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
            'element-plus': fileURLToPath(new URL('./src/ui/feedback.js', import.meta.url)),
            '@element-plus/icons-vue': fileURLToPath(new URL('./src/icons/tw-icons.js', import.meta.url))
        }
    },
    // ★★★ 核心修复：添加 server 代理配置 ★★★
    // 这部分让前端 :5173 的请求能转发到后端 :8081
    server: {
        port: 5173, // 前端默认端口
        proxy: {
            '/api': {
                target: 'http://localhost:8081', // 后端接口地址
                changeOrigin: true,
                secure: false,
                ws: true
            },
            '/ai': {
                target: 'http://localhost:8008', // AI 客服服务
                changeOrigin: true,
                secure: false,
                rewrite: (path) => path.replace(/^\/ai/, '')
            },
            '/uploads': {
                target: 'http://localhost:8081', // 图片资源代理
                changeOrigin: true,
                secure: false
            }
        }
    },
    // ★★★ 你的 Build 配置 (保持不变) ★★★
    build: {
        outDir: 'dist', // 输出目录名
        assetsDir: 'assets', // 静态资源目录
        emptyOutDir: true, // 每次打包前清空目录
    }
})
