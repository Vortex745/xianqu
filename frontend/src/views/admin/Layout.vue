<template>
  <div class="admin-layout">
    <div class="sidebar-wrapper">
      <div class="sidebar-content">
        <div class="brand-header">
          <div class="logo-box">
            <div class="circle-shape"></div>
            <div class="square-shape"></div>
          </div>
          <div class="brand-text">
            <span class="main">XIANQU</span>
            <span class="sub">Admin Panel</span>
          </div>
        </div>

        <nav class="floating-menu">
          <button
            v-for="item in menuItems"
            :key="item.path"
            class="nav-item"
            :class="{ 'is-active': activeMenu === item.path }"
            @click="handleSelect(item.path)"
          >
            <div class="menu-icon" :style="{ color: item.color }">
              <el-icon><component :is="item.icon" /></el-icon>
            </div>
            <span class="menu-label">{{ item.label }}</span>
          </button>
        </nav>

        <div class="sidebar-footer">
          <div class="dot"></div>
          <span class="ver">v1.0.2</span>
        </div>
      </div>
    </div>

    <div class="main-viewport">
      <div class="top-header">
        <div class="breadcrumb-area">
          <span class="page-title">{{ currentRouteName }}</span>
        </div>

        <div class="user-area">
          <div class="user-profile">
            <el-avatar :size="36" :src="admin.avatar || defaultAvatar" class="avatar" />
            <div class="info">
              <span class="name">{{ admin.nickname || admin.username || 'Administrator' }}</span>
              <span class="role">Super Admin</span>
            </div>
          </div>
          <div class="divider"></div>
          <div class="icon-btn logout" @click="logout" title="退出登录">
            <el-icon><SwitchButton /></el-icon>
          </div>
        </div>
      </div>

      <div class="content-body">
        <router-view v-slot="{ Component }">
          <transition name="fade-scale" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Odometer, User, Goods, List, SwitchButton, Cpu, DataLine, Coin } from '@/icons/tw-icons.js'
import { ElMessageBox, ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

const handleSelect = (index) => {
  router.push(index)
}

const defaultAvatar = 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'

// ★★★ 核心修复：读取本地管理员信息 ★★★
const admin = ref(JSON.parse(localStorage.getItem('admin_user') || '{}'))

const menuItems = [
  { path: '/admin/dashboard', label: '仪表盘', icon: Odometer, color: '#3b82f6' },
  { path: '/admin/users',     label: '用户管理', icon: User,     color: '#10b981' },
  { path: '/admin/products',  label: '商品管理', icon: Goods,    color: '#f59e0b' },
  { path: '/admin/orders',    label: '订单管理', icon: List,     color: '#8b5cf6' },
  { path: '/admin/ai-models', label: 'AI模型',   icon: Cpu,      color: '#06b6d4' },
  { path: '/admin/analytics', label: '行为分析', icon: DataLine, color: '#ec4899' },
  { path: '/admin/ai-dashboard', label: 'AI看板', icon: Coin,   color: '#f97316' },
]



const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/admin/products')) return '/admin/products'
  if (path.startsWith('/admin/users')) return '/admin/users'
  if (path.startsWith('/admin/orders')) return '/admin/orders'
  if (path.startsWith('/admin/ai-models')) return '/admin/ai-models'
  if (path.startsWith('/admin/analytics')) return '/admin/analytics'
  if (path.startsWith('/admin/ai-dashboard')) return '/admin/ai-dashboard'
  return path
})

const currentRouteName = computed(() => {
  const map = {
    '/admin/dashboard': '仪表盘 Dashboard',
    '/admin/users': '用户管理 User',
    '/admin/products': '商品管理 Product',
    '/admin/orders': '订单中心 Order',
    '/admin/analytics': '行为分析 Analytics',
    '/admin/ai-models': 'AI模型管理 Model',
    '/admin/ai-dashboard': 'AI数据看板 Dashboard'
  }
  return map[activeMenu.value] || '后台管理'
})

const logout = () => {
  ElMessageBox.confirm('确定要退出管理后台吗？', '提示', { type: 'warning', center: true, customClass: 'warm-theme-box' })
      .then(() => {
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_user')
        router.push('/admin/login')
        ElMessage.success('已退出')
      })
}
</script>

<style scoped lang="scss">
/* 定义配色变量 */
$sidebar-bg: #1a1a1a;
$active-bg: linear-gradient(135deg, #ffdf5d 0%, #ffca28 100%);
$active-glow: 0 8px 20px rgba(255, 223, 93, 0.3);

/* 基础布局 */
.admin-layout {
  display: flex;
  height: 100vh;
  background: #f8fafc; /* Lighter, more neutral gray */
  overflow: hidden;
  font-family: 'Inter', sans-serif;
}

/* ==============================
   1. 悬浮侧边栏 (Floating Sidebar)
   ============================== */
.sidebar-wrapper {
  position: fixed;
  top: 16px; bottom: 16px; left: 16px;
  width: 80px; /* 默认收起宽度 */
  z-index: 2000;
  transition: width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);

  .sidebar-content {
    height: 100%; width: 100%;
    background: rgba(15, 15, 17, 0.85); /* Deeper, more premium dark */
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    border: 1px solid rgba(255, 255, 255, 0.06); /* Whisper-thin border */
    border-radius: 24px;
    box-shadow: 0 12px 40px -12px rgba(0, 0, 0, 0.4);
    display: flex; flex-direction: column; overflow: hidden;
  }

  /* 鼠标悬停展开 */
  &:hover {
    width: 240px;
    .brand-text { opacity: 1; transform: translateX(0); }
    .menu-label { opacity: 1; transform: translateX(0); }
    .sidebar-footer .ver { opacity: 1; }
  }
}

/* Logo 区域 */
.brand-header {
  height: 80px; display: flex; align-items: center; padding-left: 24px; flex-shrink: 0;

  .logo-box {
    width: 32px; height: 32px; position: relative; flex-shrink: 0;
    .circle-shape { position: absolute; width: 20px; height: 20px; background: #ffdf5d; border-radius: 50%; top: 0; left: 0; }
    .square-shape { position: absolute; width: 18px; height: 18px; border: 2px solid #fff; bottom: 0; right: 0; border-radius: 4px; }
  }

  .brand-text {
    margin-left: 16px; display: flex; flex-direction: column;
    opacity: 0; transform: translateX(10px); transition: all 0.3s ease 0.1s; white-space: nowrap;
    .main { color: #fff; font-family: 'Outfit', sans-serif; font-weight: 800; font-size: 19px; letter-spacing: 0.02em; }
    .sub { color: #888; font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; }
  }
}

/* 菜单样式 */
.floating-menu {
  border: none; padding: 10px 12px; flex: 1; background: transparent;
  display: flex; flex-direction: column; gap: 4px;
}

.nav-item {
  width: 100%;
  height: 50px;
  border-radius: 14px;
  padding: 0;
  display: flex;
  align-items: center;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background 0.2s, transform 0.15s;
  text-align: left;

  .menu-icon {
    width: 56px;
    height: 50px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 22px;
    flex-shrink: 0;
    transition: transform 0.2s;
  }

  .menu-label {
    font-size: 15px;
    font-weight: 500;
    color: #8c9bae;
    opacity: 0;
    transform: translateX(10px);
    transition: opacity 0.3s ease, transform 0.3s ease;
    white-space: nowrap;
  }

  &:hover {
    background: rgba(255, 255, 255, 0.06);
    .menu-icon { transform: scale(1.08); }
  }

  &:active {
    transform: scale(0.97);
  }

  &.is-active {
    background: $active-bg;
    box-shadow: $active-glow;
    .menu-icon { color: #1a1a1a !important; }
    .menu-label { color: #1a1a1a; font-weight: 700; }
  }
}

/* sidebar hover 时展开 menu-label */
.sidebar-wrapper:hover .nav-item .menu-label {
  opacity: 1;
  transform: translateX(0);
}

/* 底部版本号 */
.sidebar-footer {
  height: 60px; display: flex; align-items: center; justify-content: center; gap: 10px; border-top: 1px solid rgba(255,255,255,0.05);
  .dot { width: 6px; height: 6px; background: #52c41a; border-radius: 50%; box-shadow: 0 0 6px #52c41a; }
  .ver { color: #666; font-size: 12px; opacity: 0; transition: 0.3s; }
}

/* ==============================
   2. 右侧主视图
   ============================== */
.main-viewport {
  flex: 1;
  margin-left: 104px; /* 留出侧边栏宽度 80 + 24 */
  display: flex; flex-direction: column;
  padding: 16px 20px 16px 0;
  transition: 0.3s;
}

.top-header {
  height: 70px; background: rgba(255, 255, 255, 0.7); border-radius: 20px;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 28px; box-shadow: 0 2px 12px rgba(10, 10, 12, 0.03), 0 8px 32px rgba(10, 10, 12, 0.02); margin-bottom: 24px;
  backdrop-filter: blur(24px) saturate(1.2);
  -webkit-backdrop-filter: blur(24px) saturate(1.2);
  border: 1px solid rgba(255,255,255,0.8);

  .page-title { font-family: 'Outfit', sans-serif; font-size: 20px; font-weight: 700; color: #111827; letter-spacing: -0.01em; }

  .user-area {
    display: flex; align-items: center; gap: 20px;
    .user-profile {
      display: flex; align-items: center; gap: 12px;
      .avatar { border: 2px solid #fff; box-shadow: 0 4px 10px rgba(0,0,0,0.1); }
      .info {
        display: flex; flex-direction: column; line-height: 1.3;
        .name { font-size: 14px; font-weight: 600; color: #111827; }
        .role { font-size: 12px; color: #6b7280; font-weight: 500; }
      }
    }
    .divider { width: 1px; height: 24px; background: rgba(10, 10, 12, 0.08); }
    .logout {
      width: 40px; height: 40px; border-radius: 12px; background: #f3f4f6;
      display: flex; align-items: center; justify-content: center; color: #4b5563;
      cursor: pointer; transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
      &:hover { background: #fee2e2; color: #ef4444; transform: translateY(-1px); }
      &:active { transform: scale(0.94); }
    }
  }
}

 .content-body {
  flex: 1; border-radius: 20px; overflow-y: auto; overflow-x: hidden; position: relative;
  display: flex; flex-direction: column;
}

/* 页面切换动画 */
.fade-scale-enter-active, .fade-scale-leave-active { transition: all 0.3s ease; }
.fade-scale-enter-from { opacity: 0; transform: scale(0.98) translateY(10px); }
.fade-scale-leave-to { opacity: 0; transform: scale(1.02); }
</style>
