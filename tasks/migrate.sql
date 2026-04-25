-- 闲趣项目 PostgreSQL 全量建表脚本

-- 管理员表
CREATE TABLE IF NOT EXISTS admins (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    username VARCHAR(32) UNIQUE NOT NULL,
    password VARCHAR(128) NOT NULL,
    nickname VARCHAR(32),
    avatar VARCHAR(255),
    phone VARCHAR(20),
    email VARCHAR(100),
    role_id INTEGER DEFAULT 1,
    status INTEGER DEFAULT 1,
    last_login_time TIMESTAMPTZ,
    last_login_ip VARCHAR(50)
);

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    nickname TEXT,
    avatar TEXT,
    phone TEXT,
    email TEXT,
    role TEXT DEFAULT 'user',
    status INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

-- 商品表
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL,
    image TEXT,
    area TEXT,
    category TEXT,
    status INTEGER DEFAULT 1,
    view_count INTEGER DEFAULT 0,
    count INTEGER DEFAULT 1,
    is_free_shipping BOOLEAN DEFAULT FALSE,
    is_negotiable BOOLEAN DEFAULT FALSE,
    is_home_delivery BOOLEAN DEFAULT FALSE,
    is_self_pickup BOOLEAN DEFAULT FALSE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    order_no TEXT UNIQUE,
    user_id BIGINT REFERENCES users(id),
    seller_id BIGINT REFERENCES users(id),
    product_id BIGINT REFERENCES products(id),
    price DECIMAL,
    status INTEGER DEFAULT 1
);

-- 订单日志表
CREATE TABLE IF NOT EXISTS order_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id),
    action TEXT,
    operator TEXT,
    detail TEXT,
    created_at TIMESTAMPTZ
);

-- 购物车表
CREATE TABLE IF NOT EXISTS carts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    count INTEGER DEFAULT 1,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

-- 收藏表
CREATE TABLE IF NOT EXISTS favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    created_at TIMESTAMPTZ
);

-- 用户行为表
CREATE TABLE IF NOT EXISTS user_behaviors (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    action VARCHAR(16) NOT NULL,
    weight DECIMAL(6, 2) NOT NULL,
    source VARCHAR(64),
    created_at TIMESTAMPTZ
);

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    sender_id BIGINT REFERENCES users(id),
    receiver_id BIGINT REFERENCES users(id),
    content TEXT,
    type INTEGER DEFAULT 1,
    is_read BOOLEAN DEFAULT FALSE
);

-- 邮箱验证码表
CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    code TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    failed_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ
);

-- AI 模型配置表
CREATE TABLE IF NOT EXISTS ai_models (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    provider VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    api_key VARCHAR(512) NOT NULL,
    base_url VARCHAR(255),
    price_per_k DECIMAL(10, 4) NOT NULL DEFAULT 0,
    status INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 0,
    description VARCHAR(255)
);

-- AI 用量记录表
CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    model_id BIGINT NOT NULL REFERENCES ai_models(id),
    app_type VARCHAR(20) NOT NULL,
    prompt_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    cost DECIMAL(10, 6) DEFAULT 0,
    session_id VARCHAR(64),
    user_id BIGINT DEFAULT 0,
    provider VARCHAR(50),
    model_name VARCHAR(100)
);

-- AI 每日汇总表
CREATE TABLE IF NOT EXISTS ai_usage_daily_stats (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL,
    model_id BIGINT NOT NULL REFERENCES ai_models(id),
    app_type VARCHAR(20) NOT NULL,
    total_tokens BIGINT DEFAULT 0,
    total_cost DECIMAL(12, 4) DEFAULT 0,
    call_count BIGINT DEFAULT 0,
    provider VARCHAR(50),
    model_name VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_admins_deleted_at ON admins(deleted_at);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);
CREATE INDEX IF NOT EXISTS idx_orders_expired_at ON orders(expired_at);
CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at);
CREATE INDEX IF NOT EXISTS idx_order_logs_order_id ON order_logs(order_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_email_created ON verification_codes(email, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_behaviors_created_at ON user_behaviors(created_at);
CREATE INDEX IF NOT EXISTS idx_user_behaviors_action_created ON user_behaviors(action, created_at);
CREATE INDEX IF NOT EXISTS idx_user_behaviors_user_created ON user_behaviors(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_user_behaviors_product_action ON user_behaviors(product_id, action);
CREATE INDEX IF NOT EXISTS idx_ai_models_deleted_at ON ai_models(deleted_at);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_deleted_at ON ai_usage_logs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_model_id ON ai_usage_logs(model_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_app_type ON ai_usage_logs(app_type);
CREATE INDEX IF NOT EXISTS idx_ai_usage_daily_stats_date ON ai_usage_daily_stats(date);
CREATE INDEX IF NOT EXISTS idx_ai_usage_daily_stats_model_id ON ai_usage_daily_stats(model_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_daily_stats_app_type ON ai_usage_daily_stats(app_type);
