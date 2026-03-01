# 闲趣项目部署方案 - TODO

## 1. 准备阶段 (Preparation)
- [x] 分析后端 (Go) 数据库模型并生成 SQL `tasks/migrate.sql`
- [x] 分析 AI 服务 (Python) 依赖与配置
- [x] 检查 Supabase 控制台，创建 `xianqu` 数据库 (如果可能)

## 2. 数据库迁移 (Supabase Database)
- [x] 为 Supabase 创建所有表 (User, Product, Order, Cart, Favorite, Message, Admin)
- [ ] 设置初始管理员数据 (Admin Init)
- [ ] 获取 Supabase 连接字符串并更新后端配置

## 3. 后端部署 (Backend Hosting)
- [ ] **探讨核心方案**: 
    - 方案 A: 迁移到 Supabase Edge Functions (需要重写为 TS/Deno, 风险高)
    - 方案 B: 部署到 Vercel (支持 Go 和 Python, 且有 MCP 插件)
    - 方案 C: 部署到 EdgeOne (如果是全栈支持)
- [ ] 配置环境变量 (DB_URL, JWT_SECRET)
- [ ] 编写部署脚本

## 4. 前端部署 (EdgeOne Pages)
- [ ] 修改 `frontend/src/utils/request.js` 等，将 API 地址指向新的后端域名
- [x] 构建前端项目 `npm run build`

- [ ] 部署到 EdgeOne Pages
- [ ] 绑定域名 `315279.xyz`

## 5. 验证与优化
- [ ] 测试登录注册
- [ ] 测试 AI 客服对话
- [ ] 检查图片上传功能 (Supabase Storage?)
- [ ] 验证 SSL 和域名访问
