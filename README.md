<div align="center">

# 🎪 闲趣 · XIANQU

**校园二手交易平台**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org/)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.115+-009688?style=flat-square&logo=fastapi)](https://fastapi.tiangolo.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-336791?style=flat-square&logo=postgresql)](https://www.postgresql.org/)

</div>

---

## 项目简介

**闲趣**是一个面向校园场景的二手商品交易平台，涵盖商品发布与浏览、搜索筛选、购物车与订单流转、买卖双方即时私信、管理员后台，以及基于 FastAPI + LangChain 的 AI 客服、AI Agent 与智能推荐能力。

项目采用**前后端分离 + 独立 AI 微服务**的三层架构：

```
前端 (Vue 3 + Vite)  ←→  后端 (Go + Gin)  ←→  AI 服务 (FastAPI + LangChain)
                               ↕
                         PostgreSQL
```

---

## ✨ 功能亮点

| 模块 | 功能 |
|------|------|
| 🛒 商品交易 | 瀑布流浏览、关键词搜索、分类筛选、商品详情、收藏、购物车、下单、模拟支付 |
| 💬 即时沟通 | 基于 WebSocket 的买卖双方私信聊天，支持 Emoji |
| 📦 订单管理 | 订单状态流转（待付款→已付款→已确认→退款），买卖双方视角分离 |
| 🛡️ 管理后台 | 用户管理、商品审核、订单管理、数据分析看板、AI 模型管理 |
| 🤖 AI 客服 | 基于 LangChain 的多轮对话客服，回答平台相关问题 |
| 🧠 AI Agent | 识别用户意图，自动调用站内接口（搜索商品、查询订单等） |
| 📊 AI 推荐 | 结合用户行为与候选商品，返回个性化推荐列表 |
| 🚀 便捷联调 | 前端开发脚本自动检测并拉起后端，一条命令启动全栈环境 |
| 📦 单服务部署 | 后端通过 `go:embed` 托管前端静态资源，支持单二进制部署 |

---

## 🛠️ 技术栈

| 层级 | 技术选型 |
|------|---------|
| **前端框架** | Vue 3、Vite 7、Pinia、Vue Router 4 |
| **前端 UI** | Element Plus、Bootstrap 5、ECharts 5、GSAP 3、Iconify |
| **前端工具** | Axios、Sass、vue3-emoji-picker |
| **后端框架** | Go 1.24+、Gin、GORM |
| **后端组件** | JWT 鉴权、gorilla/websocket、go:embed |
| **数据库** | PostgreSQL 14+ |
| **AI 服务** | FastAPI、LangChain、langchain-openai |
| **AI 算法** | scikit-learn、scipy（协同过滤推荐） |
| **AI 模型** | DeepSeek（默认，可通过管理后台切换） |

---

## 📁 目录结构

```
xianqu-main/
├── frontend/                      # Vue 3 前端
│   ├── scripts/
│   │   └── dev.mjs                # 联调脚本，自动拉起后端
│   ├── src/
│   │   ├── views/                 # 页面视图
│   │   │   ├── admin/             # 管理后台页面
│   │   │   ├── Home.vue           # 首页（瀑布流）
│   │   │   ├── ProductDetail.vue  # 商品详情
│   │   │   ├── Cart.vue           # 购物车
│   │   │   ├── ChatRoom.vue       # 聊天室
│   │   │   └── ...
│   │   ├── components/            # 业务组件
│   │   ├── stores/                # Pinia 状态管理
│   │   ├── router/                # 前端路由
│   │   └── utils/                 # 请求封装、动画工具等
│   ├── package.json
│   └── vite.config.js
│
├── backend/                       # Go 后端
│   ├── main.go                    # 服务入口
│   ├── config/                    # 数据库连接与初始化
│   ├── core/
│   │   ├── app/                   # Gin 引擎与路由注册
│   │   ├── controllers/           # 控制器层
│   │   ├── services/              # 业务逻辑层
│   │   ├── models/                # GORM 数据模型
│   │   ├── middleware/            # JWT 等中间件
│   │   └── utils/                 # 工具函数
│   ├── pkg/
│   │   └── ws/                    # WebSocket Hub / Client
│   ├── uploads/                   # 上传文件目录
│   └── go.mod
│
├── ai_service/                    # FastAPI AI 微服务
│   ├── main.py                    # 入口
│   ├── app/
│   │   ├── api/                   # 路由层
│   │   ├── langchain_module/      # AI 客服（多轮对话）
│   │   ├── agent_module/          # AI Agent（意图识别 + 工具调用）
│   │   ├── recommend_module/      # 推荐算法
│   │   ├── core/                  # 配置与依赖注入
│   │   └── schemas/               # 请求 / 响应模型
│   ├── knowledge/                 # 知识库快照（供客服使用）
│   ├── requirements.txt
│   └── main.py
│
└── README.md
```

---

## 🚀 快速开始

### 环境依赖

| 依赖 | 建议版本 |
|------|---------|
| Node.js | `^20.19.0` 或 `>=22.12.0` |
| Go | `1.24+` |
| PostgreSQL | `14+` |
| Python | `3.10+` |

---

### 第一步：准备数据库

后端启动时会自动连接 PostgreSQL 并执行数据库迁移，需要提前配置以下**任意一个**环境变量：

**PowerShell：**
```powershell
$env:DB_URL="postgres://postgres:postgres@127.0.0.1:5432/xianqu?sslmode=disable&TimeZone=Asia/Shanghai"
```

**Bash / macOS / Linux：**
```bash
export DB_URL="postgres://postgres:postgres@127.0.0.1:5432/xianqu?sslmode=disable&TimeZone=Asia/Shanghai"
```

> 支持的变量名：`DB_URL`、`DATABASE_URL`、`POSTGRES_DSN`，三者任选其一。

---

### 第二步：启动前后端联调环境（推荐）

前端内置的 `dev.mjs` 脚本会自动检测 `8081` 端口，若后端未运行则自动在 `backend/` 目录执行 `go run .`。

```bash
cd frontend
npm install
npm run dev
```

启动成功后访问：

- **前端**：http://localhost:5173
- **后端**：http://localhost:8081

---

### 第三步（可选）：启动 AI 服务

```bash
cd ai_service

# 创建虚拟环境
python -m venv .venv
```

**Windows：**
```powershell
.venv\Scripts\activate
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8008
```

**macOS / Linux：**
```bash
source .venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8008
```

> AI 服务默认监听 `http://localhost:8008`。前端的 `/ai/*` 请求会由 Vite 代理转发，无需额外配置。

---

### 单独启动各服务

<details>
<summary>仅启动后端</summary>

```bash
cd backend
go mod tidy
go run .
```

</details>

<details>
<summary>仅启动前端（不自动拉起后端）</summary>

```bash
cd frontend
npm install
npm run dev:vite
```

</details>

---

## ⚙️ 环境变量

### 前端（`frontend/.env`）

| 变量 | 说明 | 默认行为 |
|------|------|---------|
| `VITE_AI_BASE_URL` | AI 服务基地址 | 留空时走 `/ai` 代理到 `localhost:8008` |

复制示例文件后按需修改：
```bash
cp frontend/.env.example frontend/.env
```

---

### 后端

后端直接读取系统环境变量，优先级如下（取到第一个非空值即止）：

| 变量 | 说明 |
|------|------|
| `DB_URL` | PostgreSQL 连接串 |
| `DATABASE_URL` | 连接串别名（兼容 Heroku / Vercel） |
| `POSTGRES_DSN` | 连接串别名 |

---

### AI 服务（`ai_service/.env`）

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DEEPSEEK_API_KEY` | 模型服务密钥 | **必填** |
| `DEEPSEEK_BASE_URL` | 模型服务地址 | `https://api.deepseek.com/v1` |
| `DEEPSEEK_MODEL` | 默认模型名 | `deepseek-chat` |
| `DEEPSEEK_TEMPERATURE` | 生成温度 | `0.4` |
| `DEEPSEEK_TIMEOUT` | LLM 请求超时（秒） | `30` |
| `BACKEND_API_BASE_URL` | 后端 API 基地址（供 Agent 调用） | `http://localhost:8081/api` |
| `BACKEND_TIMEOUT` | AI 调后端超时（秒） | `12` |
| `ALLOWED_ORIGINS` | CORS 允许来源 | `http://localhost:5173` |

```bash
cp ai_service/.env.example ai_service/.env
# 填写 DEEPSEEK_API_KEY 后启动
```

---

## 🗺️ 前端路由

### 商城页

| 路径 | 说明 | 需要登录 |
|------|------|:--------:|
| `/` | 首页（瀑布流） | ✗ |
| `/search` | 搜索页 | ✗ |
| `/product/:id` | 商品详情 | ✗ |
| `/publish` | 发布商品 | ✓ |
| `/product/manage/:id` | 编辑商品 | ✓ |
| `/cart` | 购物车 | ✓ |
| `/pay/mock` | 模拟支付 | ✓ |
| `/orders` | 我的订单 | ✓ |
| `/mysales` | 我卖出的 | ✓ |
| `/profile` | 个人主页 | ✓ |
| `/user/:id` | 用户主页 | ✓ |
| `/messages` | 会话列表 | ✓ |
| `/chat/:id` | 聊天室 | ✓ |

### 管理后台

| 路径 | 说明 |
|------|------|
| `/admin/login` | 管理员登录 |
| `/admin/dashboard` | 仪表盘 |
| `/admin/users` | 用户管理 |
| `/admin/products` | 商品管理 / 审核 |
| `/admin/orders` | 订单管理 |
| `/admin/analytics` | 行为分析 |
| `/admin/ai-models` | AI 模型管理 |
| `/admin/ai-dashboard` | AI 数据看板 |

---

## 📡 后端接口概览

所有接口统一以 `/api` 为前缀。

### 公开接口（无需鉴权）

```
POST   /api/register
POST   /api/login
POST   /api/auth/send-code
POST   /api/auth/verify-login
GET    /api/health
GET    /api/ws                    # WebSocket 握手
GET    /api/products              # 商品列表
GET    /api/products/recommend    # 推荐商品
GET    /api/products/:id          # 商品详情
GET    /api/categories            # 分类列表
```

### 用户登录后接口

```
# 用户
GET/PUT  /api/user/data
GET/PUT  /api/user/profile
PUT      /api/user/password
GET      /api/users/:id

# 商品
POST     /api/products
PUT      /api/products/:id

# 收藏
GET      /api/favorites/check
POST     /api/favorites/add
DELETE   /api/favorites/remove

# 购物车
GET      /api/cart
POST     /api/cart/add
DELETE   /api/cart/:id

# 订单
GET      /api/orders
POST     /api/orders/batch
POST     /api/orders/:id/pay
POST     /api/orders/:id/confirm_pay
POST     /api/orders/:id/confirm
POST     /api/orders/:id/refund

# 聊天
GET      /api/chat/contacts
GET      /api/chat/messages

# 上传
POST     /api/upload

# 行为上报
POST     /api/behavior
```

### AI 模型开放接口

```
GET      /api/ai-models/active
GET      /api/ai-models/secret/:id
POST     /api/ai-models/usage
```

### 管理员接口（需 Admin Token）

```
POST     /api/admin/login
GET      /api/admin/init
GET      /api/admin/info
GET      /api/admin/stats
GET/PUT  /api/admin/users
PUT      /api/admin/users/:id/status
GET/PUT  /api/admin/products
PUT      /api/admin/products/:id/audit
GET      /api/admin/orders
GET      /api/admin/analytics/*
GET/POST/PUT/DELETE  /api/admin/ai-models*
```

---

## 🤖 AI 服务接口

AI 服务同时挂载了**无前缀**和 `/ai` 前缀两套路由，适配不同代理场景。

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/health` | 健康检查 |
| `POST` | `/chat` | AI 客服（多轮对话） |
| `DELETE` | `/session/{session_id}` | 清理客服会话 |
| `POST` | `/agent/chat` | AI Agent（意图识别 + 工具调用） |
| `DELETE` | `/agent/session/{session_id}` | 清理 Agent 会话 |
| `POST` | `/recommend` | 推荐结果生成 |

> 以上路径均可加 `/ai` 前缀访问，例如 `/ai/chat`、`/ai/agent/chat`。

---

## 🔑 管理员初始化

首次部署后，访问以下接口自动创建默认管理员账号：

```
GET http://localhost:8081/api/admin/init
```

默认凭据：

| 字段 | 值 |
|------|----|
| 用户名 | `admin` |
| 密码 | `123456` |

后台登录地址：`http://localhost:5173/admin/login`

> ⚠️ 生产环境请在初始化后立即修改默认密码。

---

## 📦 构建与部署

### 方式一：分离部署

分别构建并部署前端、后端、AI 服务到各自的服务器或容器。

```bash
# 构建前端
cd frontend && npm run build
# 产物输出至 frontend/dist/

# 构建后端
cd backend && go build -o xianqu-server .
```

---

### 方式二：单服务部署（推荐生产）

后端通过 `go:embed` 读取 `backend/dist/` 中的前端静态资源，只需部署一个二进制文件。

```bash
# 1. 构建前端
cd frontend
npm install && npm run build

# 2. 将前端产物复制到后端目录
cp -r dist ../backend/dist

# 3. 构建后端（含嵌入资源）
cd ../backend
go build -o xianqu-server .

# 4. 启动
./xianqu-server
```

服务默认监听 `http://localhost:8081`（前后端共用同一端口）。

---

### AI 服务部署

AI 服务可独立部署，建议使用 `uvicorn` + `gunicorn` 或直接容器化：

```bash
cd ai_service
uvicorn main:app --host 0.0.0.0 --port 8008 --workers 2
```

---

## 🔧 开发参考

| 文件 | 说明 |
|------|------|
| `frontend/vite.config.js` | Vite 配置（含代理规则） |
| `frontend/scripts/dev.mjs` | 全栈联调启动脚本 |
| `backend/core/app/app.go` | Gin 路由注册总入口 |
| `backend/config/` | 数据库连接与自动迁移 |
| `ai_service/app/application.py` | FastAPI 应用初始化 |
| `ai_service/app/model_manager.py` | AI 模型管理（支持动态切换） |
| `ai_service/knowledge/` | AI 客服知识库文件 |

---

## ❓ 常见问题

**Q：后端启动报错"数据库地址缺失"？**
> 请先配置 `DB_URL`、`DATABASE_URL` 或 `POSTGRES_DSN` 环境变量之一，确保 PostgreSQL 服务已启动且数据库已创建。

**Q：页面正常打开，但 AI 功能不可用？**
> 前端把 `/ai/*` 代理到 `localhost:8008`，需要额外启动 `ai_service` 并在 `.env` 中填写有效的 `DEEPSEEK_API_KEY`。

**Q：上传的图片在哪里？**
> 后端在运行目录下自动创建 `uploads/` 文件夹，并通过 `/uploads/*` 路径对外提供访问。

**Q：如何切换 AI 模型？**
> 登录管理后台 → AI 模型管理，可添加、配置并切换不同的模型供应商和模型名称，无需重启服务。

**Q：WebSocket 连接失败？**
> 确认后端正常运行，且前端代理配置中 `/api/ws` 已正确转发，检查浏览器控制台是否有跨域或鉴权错误。

---

## 📄 License

本项目当前定位为**学习 / 课程设计 / 个人实验**用途。  
若有意用于生产或商业场景，请自行补充完善许可证、安全审计与部署规范。