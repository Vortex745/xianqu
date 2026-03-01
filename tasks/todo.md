# 闲趣项目部署方案 - TODO

## 1. 准备阶段 (Preparation)
- [x] 分析后端 (Go) 数据库模型并生成 SQL `tasks/migrate.sql`
- [x] 分析 AI 服务 (Python) 依赖与配置
- [x] 迁移数据库到 Neon PostgreSQL (完成迁移并验证连接)

## 2. 数据库迁移 (Neon Database)
- [x] 为 Neon 创建所有表 (User, Product, Order, Cart, Favorite, Message, Admin, AIModel)
- [x] 设置初始管理员数据 (Admin Init)
- [x] 获取 Neon 连接字符串并更新后端适配 (使用 PGBouncer 端口以应对 Serverless)

## 3. 后端部署 (Vercel Backend)
- [x] **探讨核心方案**: 
    - 方案 B: 部署到 Vercel (支持 Go, 通过 `api/index.go` 桥接)
- [x] 配置环境变量 (DB_URL, JWT_SECRET)
- [x] 修复上传逻辑：使用内存处理图片并返回 Base64 (避免 Vercel 无法写入本地文件的问题)
- [x] 部署后端至 `api.315279.xyz`

## 4. 前端部署 (Cloudflare/EdgeOne Pages)
- [x] 修改 `frontend/src/utils/request.js`：将 `VITE_API_URL` 指向 `https://api.315279.xyz`
- [x] 修复硬编码的 `127.0.0.1:8081` 和 `localhost:8081` (全局使用 `resolveUrl` 助手)
- [x] 修复图片上传绝对路径：`Publish.vue` 的 `el-upload` 使用 `uploadUrl`
- [x] 修复前端 WebSocket 连接域名冲突
- [x] 绑定主域名 `xianqu.xyz`

## 5. 验证与优化
- [x] 测试登录注册 (解决连接异常问题)
- [x] 验证图片上传：现在产品图片由于使用 Base64 可以在 Serverless 下正常存储
- [x] 验证 AI 模型管理：修复了后端探测 400 提示及前端请求 404
- [x] 验证 WebSocket 实时通信

## 6. 评审 (2026-03-01)
### 修复记录
1. **图片上传失败**：原因为后端试图写入 Vercel 只读文件系统。已改为 Base64 直传并将 Data URL 存库。
2. **连接异常 (Frontend)**：由于 `VITE_API_URL` 为空，前端请求 fallback 到自身域名导致 404。已设置为生产 API 域名。
3. **AI 模型管理**：前端验证拦截及错误的 API 前缀导致无法保存，已通过 `request.js` 统一修复。
4. **硬编码 URL**：多个组件硬编码了 `127.0.0.1` 导致线上无法显示头像/图片。已通过 `resolveUrl` 全局处理。

### 下一步建议
- 监控 Neon 数据库存储：Base64 方案会加速存储消耗（1天 ~10MB 左右如果有活跃上传）。
- 前后端同步：如果用户再次进行生产构建，需确保 `.env.production` 生效。
