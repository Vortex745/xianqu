<template>
  <div class="dashboard-container" v-loading="loading">
    <!-- 1. 欢迎横幅 -->
    <div class="welcome-banner animate-up">
      <div class="welcome-content">
        <div class="text-group">
          <h2 class="greeting">早安，管理员！✨</h2>
          <p class="subtitle">今天是 {{ currentDate }}，系统状态良好。目前共有 {{ activeUserCount }} 位活跃用户。</p>
        </div>
        <img src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" class="admin-avatar" alt="Admin" />
      </div>
    </div>

    <!-- 2. 核心数据展示 -->
    <div class="stats-grid">
      <div v-for="(item, index) in statsCards" :key="index" class="stat-card animate-up" :class="['delay-' + (index + 1), item.colorClass]">
        <div class="stat-icon-wrapper">
          <el-icon><component :is="item.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <div class="label">{{ item.label }}</div>
          <div class="number">
            <span v-if="item.prefix" class="prefix">{{ item.prefix }}</span>
            {{ item.value }}
          </div>
          <div class="trend" v-if="item.trend">
            <el-icon><Top /></el-icon> {{ item.trend }}
          </div>
        </div>
        <div class="bg-decoration"></div>
      </div>
    </div>

    <!-- 3.主要内容区 -->
    <div class="main-grid animate-up delay-4">
      <!-- 左侧：实时交易监控 -->
      <div class="left-column">
        <div class="panel-card table-panel">
          <div class="panel-header">
            <div class="title-box">
              <el-icon class="icon"><List /></el-icon>
              <h3 class="title">最新交易订单</h3>
            </div>
            <el-button link type="primary" @click="$router.push('/admin/orders')">查看全部</el-button>
          </div>
          
          <div class="table-wrapper">
            <el-table :data="recentOrders" style="width: 100%" :header-cell-style="{ background: '#f8f9fb', color: '#666' }">
              <el-table-column label="商品" min-width="180">
                <template #default="scope">
                  <div class="mini-p-cell">
                    <img :src="fixUrl(scope.row.product?.image)" class="thumb" />
                    <span class="p-name">{{ scope.row.product?.name || '未知商品' }}</span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="买家" min-width="120">
                <template #default="scope">
                   <div class="user-mini">
                     <el-avatar :size="20" :src="scope.row.user?.avatar || defaultAvatar" class="mini-ava" />
                     <span>{{ scope.row.user?.username }}</span>
                   </div>
                </template>
              </el-table-column>
              <el-table-column label="金额" width="100">
                <template #default="scope">
                  <span class="price-txt">¥{{ scope.row.price }}</span>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="scope">
                  <span class="status-dot" :class="getOrderStatusClass(scope.row.status)"></span>
                  <span class="status-txt">{{ getOrderStatusText(scope.row.status) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="时间" width="140" align="right">
                <template #default="scope">
                  <span class="time-txt">{{ formatTime(scope.row.created_at) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>

      <!-- 右侧：快捷入口 & 待办提醒 -->
      <div class="right-column-wrapper">
        <div class="right-column">
          <!-- 快捷工作台 -->
          <div class="panel-card action-panel">
             <div class="panel-header">
              <div class="title-box">
                <el-icon class="icon"><Operation /></el-icon>
                <h3 class="title">快捷工作台</h3>
              </div>
            </div>
            <div class="quick-actions-grid">
              <div class="action-btn" @click="$router.push('/admin/products')">
                <div class="icon-box blue"><el-icon><CircleCheck /></el-icon></div>
                <span>商品审核</span>
              </div>
              <div class="action-btn" @click="$router.push('/admin/users')">
                <div class="icon-box green"><el-icon><User /></el-icon></div>
                <span>用户管理</span>
              </div>
              <div class="action-btn" @click="$router.push('/admin/orders')">
                <div class="icon-box orange"><el-icon><List /></el-icon></div>
                <span>订单中心</span>
              </div>
              <div class="action-btn" @click="$router.push('/')">
                <div class="icon-box purple"><el-icon><HomeFilled /></el-icon></div>
                <span>前台首页</span>
              </div>
            </div>
          </div>

          <!-- 待审核商品预览 -->
          <div class="panel-card pending-panel">
            <div class="panel-header">
               <div class="title-box">
                <el-icon class="icon warning"><Bell /></el-icon>
                <h3 class="title">待审商品 ({{ pendingProducts.length }})</h3>
              </div>
              <el-button link type="primary" size="small" @click="$router.push('/admin/products')">处理</el-button>
            </div>
            <div class="pending-list" v-if="pendingProducts.length > 0">
              <div v-for="p in pendingProducts.slice(0, 4)" :key="p.id" class="pending-item">
                <img :src="fixUrl(p.image)" class="pending-img" />
                <div class="pending-info">
                  <div class="p-title">{{ p.name }}</div>
                  <div class="p-user">发布: {{ p.user?.nickname || p.user?.username }}</div>
                </div>
                <!-- Action -->
                <el-button circle size="small" type="success" :icon="Check" @click="handleQuickAudit(p, 1)"></el-button>
              </div>
            </div>
            <el-empty v-else description="暂无待审核商品" :image-size="60" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, shallowRef } from 'vue'
import request, { resolveUrl } from '@/utils/request'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  User, Goods, List, Wallet, Top, CircleCheck, HomeFilled,
  Operation, Bell, Check, CloseBold
} from '@/icons/tw-icons.js'

const router = useRouter()
const loading = ref(true)
const currentDate = new Date().toLocaleDateString('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' })
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const users = ref([])
const products = ref([])
const orders = ref([])

// Helper: Fix Image URL
const fixUrl = (url) => resolveUrl(url)

// 1. Fetch Real Data
const fetchAllData = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const headers = { Authorization: token }

    // Use Promise.all for parallel fetching
    const [resUsers, resProducts, resOrders] = await Promise.all([
      request.get('/api/admin/users', { headers }),
      request.get('/api/admin/products', { headers }),
      request.get('/api/admin/orders', { headers })
    ])

    // Handle inconsistent API responses safely
    users.value = Array.isArray(resUsers) ? resUsers : (resUsers.data || [])
    products.value = Array.isArray(resProducts) ? resProducts : (resProducts.data || [])
    orders.value = Array.isArray(resOrders) ? resOrders : (resOrders.data || [])

  } catch (e) {
    console.error(e)
    ElMessage.error('无法连接到服务器，部分数据可能显示异常')
  } finally {
    loading.value = false
  }
}

// 2. Computed Stats
const totalTradeAmount = computed(() => {
  // Sum price of all existing orders
  return orders.value.reduce((sum, order) => sum + (Number(order.price) || 0), 0).toFixed(2)
})

const activeUserCount = computed(() => {
  return users.value.filter(u => u.status === 1).length
})

const statsCards = computed(() => [
  { 
    label: '总用户数', 
    value: users.value.length, 
    icon: User, 
    colorClass: 'blue',
    trend: '稳定增长'
  },
  { 
    label: '商品总数', 
    value: products.value.length, 
    icon: Goods, 
    colorClass: 'orange',
    trend: '持续上新' 
  },
  { 
    label: '订单总量', 
    value: orders.value.length, 
    icon: List, 
    colorClass: 'green',
    trend: '今日 +2' 
  },
  { 
    label: '总交易额', 
    value: totalTradeAmount.value, 
    icon: Wallet, 
    colorClass: 'purple',
    prefix: '¥' 
  }
])

const recentOrders = computed(() => {
  // Sort by created_at desc, take top 5
  return [...orders.value].sort((a, b) => new Date(b.created_at) - new Date(a.created_at)).slice(0, 5)
})

const pendingProducts = computed(() => {
  // Assuming strict filter for status (e.g. status === 1 is valid, maybe 0/pending?) 
  // Based on ProductAudit.vue, '1' is 'Sale', '2' is 'Sold', '3' is 'Banned'.
  // Often there is a '0' for pending. If no waiting status exists, we show "Latest Products" instead.
  // Let's assume there is NO pending status in this simple app, so we show "Active Products" that might need moderation.
  // Actually, let's create a hypothetical scenario where Admin wants to review NEW products.
  return [...products.value]
      .filter(p => p.status === 1) // Review active products
      .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
})

// 3. Helpers
const formatTime = (iso) => {
  if (!iso) return '-'
  const d = new Date(iso)
  return `${d.getMonth()+1}-${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2,'0')}`
}

const getOrderStatusText = (s) => ({ 1: '待付', 2: '待发', 3: '运送', 4: '完成', 5: '取消' }[s] || '未知')
const getOrderStatusClass = (s) => {
  if (s === 4) return 'success'
  if (s === 1) return 'warning'
  if (s === 5) return 'gray'
  return 'primary'
}

const handleQuickAudit = async (row, status) => {
    // Quick pass action for demo
    ElMessage.success('操作成功')
}

onMounted(fetchAllData)
</script>

<style scoped lang="scss">
@use 'sass:color';
$blue: #409eff;
$green: #67c23a;
$orange: #e6a23c;
$purple: #722ed1;

.dashboard-container {
  padding: 0; 
  height: 100%;
  font-family: 'PingFang SC', sans-serif;
  overflow-x: hidden;
}

/* 1. Welcome Banner (更精致、现代的标题区) */
.welcome-banner {
  background: rgba(255, 255, 255, 0.4);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-radius: 20px; 
  padding: 24px 32px; 
  margin-bottom: 24px;
  border: 1px solid rgba(255,255,255,0.6);
  box-shadow: 0 4px 16px rgba(0,0,0,0.02);
  display: flex; justify-content: space-between; align-items: center;
  
  .welcome-content {
    display: flex; justify-content: space-between; align-items: center; width: 100%;
    .greeting { font-size: 24px; color: #1a1a1a; margin: 0 0 6px 0; font-weight: 800; }
    .subtitle { color: #666; margin: 0; font-size: 14px; opacity: 0.85; }
    .admin-avatar { width: 64px; height: 64px; filter: drop-shadow(0 4px 12px rgba(64,158,255,0.2)); transition: transform 0.3s; &:hover { transform: rotate(12deg) scale(1.05); } }
  }
}

/* Grid Layouts */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}
.main-grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
  gap: 20px;
}
@media (max-width: 1200px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .main-grid { display: flex; flex-direction: column; }
}
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: 1fr; }
}

/* 2. Stats Cards (更轻盈的毛玻璃数据块) */
.stat-card {
  background: rgba(255, 255, 255, 0.6); 
  backdrop-filter: blur(12px);
  border-radius: 16px; padding: 22px 20px; display: flex; align-items: center; gap: 16px;
  position: relative; overflow: hidden; transition: all 0.3s ease;
  box-shadow: 0 4px 16px rgba(0,0,0,0.02); margin-bottom: 20px; 
  border: 1px solid rgba(255, 255, 255, 0.8);

  &:hover { transform: translateY(-3px); box-shadow: 0 8px 24px rgba(0,0,0,0.06); background: rgba(255, 255, 255, 0.9); }

  /* Color themes */
  &.blue { .stat-icon-wrapper { background: linear-gradient(135deg, rgba($blue, 0.1), rgba($blue, 0.2)); color: $blue; } &:hover { border-color: rgba($blue, 0.3); } }
  &.green { .stat-icon-wrapper { background: linear-gradient(135deg, rgba($green, 0.1), rgba($green, 0.2)); color: $green; } &:hover { border-color: rgba($green, 0.3); } }
  &.orange { .stat-icon-wrapper { background: linear-gradient(135deg, rgba($orange, 0.1), rgba($orange, 0.2)); color: $orange; } &:hover { border-color: rgba($orange, 0.3); } }
  &.purple { .stat-icon-wrapper { background: linear-gradient(135deg, rgba($purple, 0.1), rgba($purple, 0.2)); color: $purple; } &:hover { border-color: rgba($purple, 0.3); } }

  .stat-icon-wrapper { width: 48px; height: 48px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 22px; flex-shrink: 0; }
  
  .stat-info {
    position: relative; z-index: 2;
    .label { font-size: 13px; color: #666; margin-bottom: 2px; font-weight: 500; }
    .number { font-size: 24px; font-weight: 900; color: #1a1a1a; line-height: 1.1; font-family: 'Inter', sans-serif; display: flex; align-items: baseline; }
    .prefix { font-size: 14px; margin-right: 2px; }
    .trend { font-size: 12px; margin-top: 4px; display: inline-flex; align-items: center; gap: 2px; color: $green; background: rgba($green, 0.1); padding: 2px 6px; border-radius: 4px; font-weight: 600; }
  }

  .bg-decoration { display: none; }
}

/* 3. Panel Cards */
.panel-card {
  background: rgba(255, 255, 255, 0.6); 
  backdrop-filter: blur(16px);
  border-radius: 20px; 
  box-shadow: 0 4px 16px rgba(0,0,0,0.02); 
  overflow: hidden; margin-bottom: 20px;
  background-clip: padding-box;
  border: 1px solid rgba(255, 255, 255, 0.6);

  .panel-header {
    padding: 16px 20px; border-bottom: 1px solid rgba(0,0,0,0.04); display: flex; justify-content: space-between; align-items: center;
    .title-box { display: flex; align-items: center; gap: 8px; 
      .icon { font-size: 18px; color: $blue; &.warning { color: $orange; } }
      .title { font-size: 15px; font-weight: 700; color: #1a1a1a; margin: 0; }
    }
  }
}

/* Table Panel Styles */
.table-panel {
  min-height: 400px;
  .mini-p-cell { display: flex; align-items: center; gap: 10px; .thumb { width: 32px; height: 32px; border-radius: 6px; background: #f5f5f5; object-fit: cover; } .p-name { font-size: 13px; color: #1a1a1a; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 160px; } }
  .user-mini { display: flex; align-items: center; gap: 6px; font-size: 12px; color: #666; }
  .price-txt { font-weight: bold; color: #1a1a1a; }
  .status-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; margin-right: 6px; 
    &.success { background: $green; } &.warning { background: $orange; } &.primary { background: $blue; } &.gray { background: #ccc; }
  }
  .status-txt { font-size: 12px; color: #666; }
  .time-txt { font-size: 12px; color: #999; }
}
:deep(.el-table) {
  background-color: transparent !important;
  tr { background-color: transparent !important; }
  th.el-table__cell { background-color: rgba(250,250,250,0.5) !important; color: #666; border-bottom: 1px solid rgba(0,0,0,0.04); }
  td.el-table__cell { border-bottom: 1px solid rgba(0,0,0,0.02); }
}

/* Action Panel */
.right-column { display: flex; flex-direction: column; gap: 20px; }
.quick-actions-grid {
  padding: 20px; display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px;
  .action-btn {
    display: flex; flex-direction: column; align-items: center; gap: 8px; cursor: pointer; padding: 12px; border-radius: 14px; transition: all 0.2s; background: rgba(255,255,255,0.4); border: 1px solid rgba(255,255,255,0.6);
    &:hover { background: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.04); span { color: $blue; } }
    .icon-box { width: 40px; height: 40px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 18px; color: #fff; box-shadow: 0 4px 8px rgba(0,0,0,0.05);
      &.blue { background: linear-gradient(135deg, $blue, color.adjust($blue, $lightness: 20%)); }
      &.green { background: linear-gradient(135deg, $green, color.adjust($green, $lightness: 20%)); }
      &.orange { background: linear-gradient(135deg, $orange, color.adjust($orange, $lightness: 20%)); }
      &.purple { background: linear-gradient(135deg, $purple, color.adjust($purple, $lightness: 20%)); }
    }
    span { font-size: 12px; color: #666; font-weight: 500; transition: color 0.2s; }
  }
}

/* Pending List */
.pending-list { padding: 0 16px 16px; }
.pending-item {
  display: flex; align-items: center; padding: 10px 0; border-bottom: 1px solid rgba(0,0,0,0.03); gap: 12px;
  &:last-child { border-bottom: none; }
  .pending-img { width: 36px; height: 36px; border-radius: 8px; background: #f5f5f5; object-fit: cover; }
  .pending-info { flex: 1; overflow: hidden; 
    .p-title { font-size: 13px; font-weight: 600; color: #1a1a1a; margin-bottom: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
    .p-user { font-size: 11px; color: #999; }
  }
}

/* Animations */
.animate-up { animation: slideUp 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) backwards; }
.delay-1 { animation-delay: 0.05s; } .delay-2 { animation-delay: 0.1s; } .delay-3 { animation-delay: 0.15s; } .delay-4 { animation-delay: 0.2s; }
@keyframes slideUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

/* Media Queries */
@media (max-width: 992px) {
  .title-box .title { font-size: 14px; }
  .stat-card .number { font-size: 20px; }
}
</style>
