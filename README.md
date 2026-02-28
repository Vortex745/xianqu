# 🛒 闲趣 XIANQU — 校园二手交易平台

> 买卖闲置、即时聊天、AI 智能客服，专为校园场景打造的全栈二手交易应用。

基于 **Vue 3 + Go (Gin) + SQLite + FastAPI (AI)** 构建，前后端分离，集成 AI 客服与 AI Agent，支持一键启动开发环境。

---

## ✨ 功能一览

### 买家端
| 功能 | 说明 |
|------|------|
| 🏠 首页浏览 | 商品瀑布流展示，支持分类筛选 |
| 🔍 搜索 | 关键词搜索商品 |
| 📦 商品详情 | 查看图片、价格、卖家信息、浏览量 |
| ❤️ 收藏（想要） | 一键收藏感兴趣的商品 |
| 🛒 购物车 | 加入购物车，批量下单 |
| 💳 下单 & 支付 | 创建订单，模拟支付（MockPay） |
| 📋 我的订单 | 查看订单状态、确认收货、申请退款 |
| 💬 即时聊天 | 基于 WebSocket 的站内实时消息 |

### 卖家端
| 功能 | 说明 |
|------|------|
| 📝 发布商品 | 上传图片、填写详情、选择分类 |
| ✏️ 商品管理 | 编辑已发布商品信息 |
| 📊 我卖出的 | 查看销售订单与状态 |

### AI 智能助手
| 功能 | 说明 |
|------|------|
| 🤖 AI 客服 | 基于 LangChain 的智能问答，解答平台使用问题 |
| 🧠 AI Agent | 意图识别 + 站内功能调用，帮用户查商品、管理收藏、查订单等 |

### 管理后台
| 功能 | 说明 |
|------|------|
| 📈 数据看板 | 用户数、商品数、订单数等统计 |
| 👥 用户管理 | 查看用户列表，启用/禁用账号 |
| ✅ 商品审核 | 审核上架商品 |
| 📦 订单管理 | 查看全部订单信息 |

---

## 🛠 技术栈

| 层 | 技术 |
|----|------|
| **前端** | Vue 3、Vite 7、Vue Router 4、Pinia 3、Element Plus、Bootstrap 5、Axios、Sass |
| **后端** | Go 1.24、Gin、GORM、JWT、gorilla/websocket |
| **AI 服务** | Python、FastAPI、LangChain、DeepSeek LLM |
| **数据库** | SQLite（纯 Go 驱动，免 CGO） |
| **实时通信** | WebSocket（自建 Hub + Client 模型） |
| **部署** | 支持后端 `embed` 内嵌前端产物，单二进制部署 |

---

## 📂 项目结构

```
xianqu/
├── frontend/                    # 前端工程
│   ├── src/
│   │   ├── views/              # 页面组件（15 个页面）
│   │   │   ├── Home.vue        #   首页
│   │   │   ├── Login.vue       #   登录/注册
│   │   │   ├── Search.vue      #   搜索
│   │   │   ├── ProductDetail.vue   # 商品详情
│   │   │   ├── Publish.vue     #   发布商品
│   │   │   ├── ProductManage.vue   # 商品管理（卖家）
│   │   │   ├── Cart.vue        #   购物车
│   │   │   ├── MockPay.vue     #   模拟支付
│   │   │   ├── UserOrders.vue  #   我的订单
│   │   │   ├── UserSales.vue   #   我卖出的
│   │   │   ├── UserProfile.vue #   个人中心
│   │   │   ├── MessageList.vue #   消息列表
│   │   │   ├── ChatRoom.vue    #   聊天室
│   │   │   └── admin/          #   后台管理（4 个页面）
│   │   ├── router/             # 前端路由
│   │   ├── stores/             # Pinia 状态管理
│   │   ├── components/         # 公共组件（含 AI 客服浮窗）
│   │   ├── ui/                 # UI 基础组件库
│   │   ├── icons/              # 图标系统
│   │   └── utils/              # 工具函数（请求封装、WebSocket）
│   ├── scripts/dev.mjs         # 一键开发脚本
│   └── vite.config.js          # Vite 配置 & 代理
│
├── backend/                     # 后端工程
│   ├── main.go                 # 服务入口（路由注册、embed 前端）
│   ├── config/
│   │   └── db.go               # 数据库初始化 & 自动迁移
│   ├── internal/
│   │   ├── controllers/        # API 控制器（7 个）
│   │   ├── middleware/         # 中间件（JWT 鉴权）
│   │   ├── models/             # 数据模型（7 个）
│   │   ├── services/           # 业务逻辑层
│   │   └── utils/              # 工具（加密、JWT 签发）
│   ├── pkg/ws/                 # WebSocket Hub & Client
│   └── uploads/                # 用户上传文件（运行时自动创建）
│
├── ai_service/                  # AI 智能服务
│   ├── main.py                 # FastAPI 启动入口
│   └── app/
│       ├── application.py      # 应用工厂
│       ├── api/routes.py       # HTTP 路由层
│       ├── schemas/chat.py     # 请求响应模型
│       ├── core/config.py      # 环境配置
│       ├── langchain_module/   # AI 客服模块（LangChain）
│       │   ├── service.py      #   客服链路入口
│       │   ├── history_store.py#   会话记忆存储
│       │   ├── llm_factory.py  #   统一 LLM 构建
│       │   └── text_utils.py   #   纯文本清洗
│       └── agent_module/       # AI Agent 模块
│           ├── service.py      #   Agent 主流程
│           ├── intent_router.py#   意图识别与动作规划
│           └── backend_tools.py#   站内功能白名单执行器
│
└── README.md
```

---

## 🚀 快速开始

### 环境要求

| 依赖 | 版本 |
|------|------|
| Node.js | `^20.19.0` 或 `>=22.12.0` |
| Go | `1.24+` |
| Python | `3.10+`（AI 服务） |
| 操作系统 | Windows / macOS / Linux |

### 方式一：一键启动（推荐）

```bash
# 1. 安装前端依赖
cd frontend
npm install

# 2. 一键启动前后端
npm run dev
```

脚本会自动完成以下事情：
1. 检测 `8081` 端口是否已有后端运行
2. 若无，自动在 `backend/` 目录执行 `go run .` 启动后端
3. 等待后端就绪后启动 Vite 开发服务器

启动后访问：
- 🌐 前端页面：`http://localhost:5173`
- 🔌 后端 API：`http://localhost:8081`

### 方式二：分别启动

**启动后端：**
```bash
cd backend
go mod tidy
go run .
# ✅ 服务运行在 http://localhost:8081
```

**启动前端：**
```bash
cd frontend
npm install
npm run dev:vite
# ✅ 页面运行在 http://localhost:5173
```

**启动 AI 服务（可选）：**
```bash
cd ai_service
python -m venv .venv

# Windows
.venv\Scripts\activate
# macOS / Linux
source .venv/bin/activate

pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8008
# ✅ AI 服务运行在 http://localhost:8008
```

> AI 服务需要配置 `ai_service/.env` 文件，设置 `DEEPSEEK_API_KEY` 等环境变量。可参考 `.env.example`。

---

## 🔧 构建部署

### 前端构建
```bash
cd frontend
npm run build
# 产物输出到 frontend/dist/
```

### 后端构建
```bash
cd backend
go build -o xianqu-server .
# 生成可执行文件 xianqu-server
```

### 单二进制部署（推荐）

项目后端通过 `go:embed` 内嵌前端静态资源，实现单文件部署：

```bash
# 1. 构建前端
cd frontend && npm run build

# 2. 复制产物到后端
cp -r dist/ ../backend/dist/

# 3. 构建后端（包含前端资源）
cd ../backend && go build -o xianqu-server .

# 4. 运行
./xianqu-server
# 访问 http://localhost:8081 即可使用完整应用
```

---

## 👤 管理后台

### 初始化管理员

首次使用需初始化管理员账号：

```
GET http://localhost:8081/api/admin/init
```

| 项 | 值 |
|----|----|
| 用户名 | `admin` |
| 密码 | `123456` |
| 后台入口 | `http://localhost:5173/admin/login` |

---

## 🗺 路由总览

### 前台页面

| 路径 | 页面 | 需要登录 |
|------|------|---------|
| `/` | 首页 | ❌ |
| `/search` | 搜索 | ❌ |
| `/product/:id` | 商品详情 | ❌ |
| `/publish` | 发布商品 | ✅ |
| `/product/manage/:id` | 管理商品 | ✅ |
| `/cart` | 购物车 | ✅ |
| `/pay/mock` | 模拟支付 | ✅ |
| `/orders` | 我的订单 | ✅ |
| `/mysales` | 我卖出的 | ✅ |
| `/profile` | 个人中心 | ✅ |
| `/user/:id` | 用户主页 | ✅ |
| `/messages` | 消息列表 | ✅ |
| `/chat/:id` | 聊天室 | ✅ |

### 后台管理

| 路径 | 页面 |
|------|------|
| `/admin/login` | 管理员登录 |
| `/admin/dashboard` | 数据看板 |
| `/admin/users` | 用户管理 |
| `/admin/products` | 商品审核 |
| `/admin/orders` | 订单管理 |

---

## 📡 API 接口概览

所有接口以 `/api` 为前缀，需要登录的接口通过 `Authorization: Bearer <token>` 传递 JWT。

### 公开接口
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/register` | 用户注册 |
| POST | `/api/login` | 用户登录 |
| GET | `/api/products` | 商品列表 |
| GET | `/api/products/:id` | 商品详情 |
| GET | `/api/categories` | 分类列表 |
| GET | `/api/ws` | WebSocket 连接 |

### 用户接口（需登录）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/data` | 获取个人信息 |
| PUT | `/api/user/profile` | 更新个人资料 |
| PUT | `/api/user/password` | 修改密码 |
| POST | `/api/upload` | 上传文件 |
| GET | `/api/favorites/check` | 检查是否已收藏 |
| POST | `/api/favorites/add` | 添加收藏 |
| POST | `/api/favorites/remove` | 取消收藏 |
| POST | `/api/cart/add` | 加入购物车 |
| GET | `/api/cart` | 购物车列表 |
| DELETE | `/api/cart/:id` | 删除购物车项 |
| POST | `/api/products` | 发布商品 |
| PUT | `/api/products/:id` | 编辑商品 |
| POST | `/api/orders` | 创建订单 |
| POST | `/api/orders/batch` | 批量下单 |
| GET | `/api/orders` | 订单列表 |
| POST | `/api/orders/:id/pay` | 订单支付 |
| POST | `/api/orders/:id/confirm_pay` | 确认支付 |
| PUT | `/api/orders/:id/confirm` | 确认收货 |
| PUT | `/api/orders/:id/refund` | 申请退款 |
| GET | `/api/chat/contacts` | 聊天联系人 |
| GET | `/api/chat/messages` | 聊天记录 |

### AI 服务接口
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/ai/chat` | AI 客服对话 |
| DELETE | `/ai/session/{session_id}` | 清理客服会话 |
| POST | `/ai/agent/chat` | AI Agent 对话 |
| DELETE | `/ai/agent/session/{session_id}` | 清理 Agent 会话 |

### 管理后台接口
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/login` | 管理员登录 |
| GET | `/api/admin/init` | 初始化管理员 |
| GET | `/api/admin/info` | 管理员信息 |
| GET | `/api/admin/stats` | 统计数据 |
| GET | `/api/admin/users` | 用户列表 |
| PUT | `/api/admin/users/:id/status` | 修改用户状态 |
| GET | `/api/admin/products` | 商品列表 |
| PUT | `/api/admin/products/:id/audit` | 审核商品 |
| GET | `/api/admin/orders` | 订单列表 |

---

## 📝 数据模型

| 模型 | 文件 | 说明 |
|------|------|------|
| User | `models/user.go` | 用户信息（用户名、密码哈希、头像等） |
| Product | `models/product.go` | 商品（标题、价格、分类、图片、状态） |
| Order | `models/order.go` | 订单（买卖双方、金额、状态流转） |
| Cart | `models/cart.go` | 购物车项 |
| Favorite | `models/favorite.go` | 用户收藏 |
| Message | `models/message.go` | 聊天消息 |
| Admin | `models/admin.go` | 管理员账号 |

数据库文件默认保存在 `backend/xianqu.db`，应用启动时会自动建表和迁移。

---

## ⚙️ 配置说明

### 主服务配置

| 配置项 | 位置 | 默认值 |
|--------|------|--------|
| 后端端口 | `backend/main.go` | `8081` |
| 前端端口 | `frontend/vite.config.js` | `5173` |
| API 代理 | `frontend/vite.config.js` | `/api` → `localhost:8081` |
| 上传代理 | `frontend/vite.config.js` | `/uploads` → `localhost:8081` |
| AI 代理 | `frontend/vite.config.js` | `/ai` → `localhost:8008` |
| 数据库路径 | `backend/config/db.go` | `backend/xianqu.db` |
| 上传目录 | `backend/main.go` | `backend/uploads/` |
| 文件大小限制 | `backend/main.go` | `100 MB` |

### AI 服务配置（`ai_service/.env`）

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 | 必填 |
| `DEEPSEEK_BASE_URL` | API 地址 | `https://api.deepseek.com/v1` |
| `DEEPSEEK_MODEL` | 模型名称 | `deepseek-chat` |
| `DEEPSEEK_TEMPERATURE` | 生成温度 | `0.4` |
| `BACKEND_API_BASE_URL` | 主后端地址 | `http://localhost:8081/api` |

---

## 📜 License

本项目仅供学习交流使用。
