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

        <el-menu
            :default-active="route.path"
            background-color="transparent"
            text-color="#8c9bae"
            active-text-color="#1a1a1a"
            router
            :collapse="false"
            class="floating-menu"
        >
          <el-menu-item index="/admin/dashboard">
            <div class="menu-icon"><el-icon><Odometer /></el-icon></div>
            <span class="menu-label">仪表盘</span>
          </el-menu-item>

          <el-menu-item index="/admin/users">
            <div class="menu-icon"><el-icon><User /></el-icon></div>
            <span class="menu-label">用户管理</span>
          </el-menu-item>

          <el-menu-item index="/admin/products">
            <div class="menu-icon"><el-icon><Goods /></el-icon></div>
            <span class="menu-label">商品审核</span>
          </el-menu-item>

          <el-menu-item index="/admin/orders">
            <div class="menu-icon"><el-icon><List /></el-icon></div>
            <span class="menu-label">订单管理</span>
          </el-menu-item>

          <el-menu-item index="/admin/ai-models">
            <div class="menu-icon"><el-icon><Cpu /></el-icon></div>
            <span class="menu-label">AI模型</span>
          </el-menu-item>

          <el-menu-item index="/admin/ai-dashboard">
            <div class="menu-icon"><el-icon><DataLine /></el-icon></div>
            <span class="menu-label">AI看板</span>
          </el-menu-item>
        </el-menu>

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
import { Odometer, User, Goods, List, SwitchButton, Cpu, DataLine } from '@/icons/tw-icons.js'
import { ElMessageBox, ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const defaultAvatar = 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'

// ★★★ 核心修复：读取本地管理员信息 ★★★
const admin = ref(JSON.parse(localStorage.getItem('admin_user') || '{}'))

const currentRouteName = computed(() => {
  const map = {
    '/admin/dashboard': '仪表盘 Dashboard',
    '/admin/users': '用户管理 User',
    '/admin/products': '商品审核 Product',
    '/admin/orders': '订单中心 Order',
    '/admin/ai-models': 'AI模型管理 Model',
    '/admin/ai-dashboard': 'AI数据看板 Dashboard'
  }
  return map[route.path] || '后台管理'
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
  background: #f2f4f7;
  overflow: hidden;
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
    background: rgba(26, 26, 26, 0.7); /* ui.glass Translucent Dark */
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 24px;
    box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.3);
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
    .main { color: #fff; font-weight: 900; font-size: 18px; letter-spacing: 1px; }
    .sub { color: #666; font-size: 12px; font-weight: 600; text-transform: uppercase; }
  }
}

/* 菜单样式 */
.floating-menu {
  border: none; padding: 10px 12px; flex: 1; background: transparent !important;
}

:deep(.el-menu-item) {
  height: 50px; line-height: 50px; margin-bottom: 8px; border-radius: 14px; padding: 0 !important; display: flex; align-items: center; transition: 0.2s;

  .menu-icon { width: 56px; height: 50px; display: flex; align-items: center; justify-content: center; font-size: 22px; flex-shrink: 0; transition: 0.2s; }
  .menu-label { font-size: 15px; font-weight: 500; opacity: 0; transform: translateX(10px); transition: all 0.3s ease; white-space: nowrap; }

  &:hover {
    background-color: rgba(255,255,255,0.1) !important;
    .menu-icon { color: #fff; transform: scale(1.1); }
  }

  &.is-active {
    background: $active-bg !important; box-shadow: $active-glow;
    .menu-icon { color: #1a1a1a; }
    .menu-label { color: #1a1a1a; font-weight: 800; }
  }
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
  height: 70px; background: #fff; border-radius: 20px;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 30px; box-shadow: 0 4px 20px rgba(0,0,0,0.02); margin-bottom: 20px;

  .page-title { font-size: 20px; font-weight: 800; color: #1a1a1a; }

  .user-area {
    display: flex; align-items: center; gap: 20px;
    .user-profile {
      display: flex; align-items: center; gap: 12px;
      .avatar { border: 2px solid #fff; box-shadow: 0 4px 10px rgba(0,0,0,0.1); }
      .info {
        display: flex; flex-direction: column; line-height: 1.3;
        .name { font-size: 14px; font-weight: 700; color: #333; }
        .role { font-size: 12px; color: #999; }
      }
    }
    .divider { width: 1px; height: 24px; background: #eee; }
    .logout {
      width: 40px; height: 40px; border-radius: 12px; background: #f9f9f9;
      display: flex; align-items: center; justify-content: center; color: #666;
      cursor: pointer; transition: 0.2s;
      &:hover { background: #fff1f0; color: #ff4d4f; }
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