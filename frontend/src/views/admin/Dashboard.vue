<template>
  <div class="page-wrapper dashboard-shell">
    <!-- Hero Section -->
    <section class="dashboard-hero" ref="heroRef">
      <div class="hero-content">
        <p class="dashboard-hero__eyebrow">运营概览</p>
        <h2 class="hero-headline">今日数据</h2>
        <p class="hero-sub">{{ currentDate }} · {{ activeUserCount }} 位活跃用户</p>
      </div>


    </section>

    <section v-if="loading" class="skeleton-board">
      <el-skeleton :rows="8" animated />
    </section>

    <template v-else>
      <!-- Stats Cards -->
      <section class="dashboard-stats" ref="statsRef">
        <article
          v-for="(item, index) in statsCards"
          :key="item.label"
          class="dashboard-card bento-card"
          :style="{ '--delay': index * 0.08 + 's' }"
        >
          <div class="dashboard-card__top">
            <div class="dashboard-card__icon" :style="{ background: item.bg, color: item.color }">
              <component :is="item.icon" />
            </div>
            <div class="dashboard-card__meta">
              <Top v-if="item.trend" />
              <span>{{ item.trend || '实时' }}</span>
            </div>
          </div>
          <div>
            <div class="dashboard-card__label">{{ item.label }}</div>
            <div class="dashboard-card__value">
              <span v-if="item.prefix">{{ item.prefix }}</span>{{ item.value }}
            </div>
          </div>
        </article>
      </section>

      <!-- Main Grid -->
      <section class="dashboard-grid">
        <!-- Recent Orders Panel -->
        <article class="dashboard-panel bento-card" ref="ordersRef">
          <div class="dashboard-panel__head">
            <h3 class="dashboard-panel__title">
              <span class="panel-icon" style="background:#f5f3ff;color:#8b5cf6;"><List /></span>
              <span>最近订单</span>
            </h3>
            <button class="dashboard-link-button" type="button" @click="$router.push('/admin/orders')">
              全部
            </button>
          </div>

          <table class="dashboard-orders">
            <thead>
              <tr>
                <th>商品</th>
                <th>买家</th>
                <th>金额</th>
                <th>状态</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="order in recentOrders" :key="order.id || order.order_no">
                <td>
                  <div class="dashboard-product">
                    <img
                      class="dashboard-product__thumb"
                      :src="fixUrl(order.product?.image)"
                      :alt="order.product?.name || '商品'"
                    />
                    <div>
                      <div class="dashboard-product__name">{{ order.product?.name || '已失效' }}</div>
                      <div class="dashboard-product__sub">{{ order.order_no || '-' }}</div>
                    </div>
                  </div>
                </td>
                <td>
                  <div class="dashboard-user">
                    <img
                      v-if="resolveAvatar(order.user?.avatar)"
                      class="dashboard-user__avatar"
                      :src="resolveAvatar(order.user?.avatar)"
                      :alt="order.user?.username || ''"
                    />
                    <span v-else class="dashboard-user__avatar fallback">{{ getInitial(order.user?.username) }}</span>
                    <div>
                      <div class="dashboard-user__name">{{ order.user?.username || '未知' }}</div>
                    </div>
                  </div>
                </td>
                <td class="dashboard-money">¥{{ order.price }}</td>
                <td>
                  <span class="dashboard-status" :class="getOrderStatusClass(order.status)">
                    {{ getOrderStatusText(order.status) }}
                  </span>
                </td>
                <td class="dashboard-time">{{ formatTime(order.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </article>

        <!-- Right Column -->
        <div class="dashboard-side">
          <!-- Quick Actions -->
          <article class="dashboard-panel bento-card" ref="actionsRef">
            <div class="dashboard-panel__head">
              <h3 class="dashboard-panel__title">
                <span class="panel-icon" style="background:#ecfeff;color:#06b6d4;"><Operation /></span>
                <span>快捷入口</span>
              </h3>
            </div>

            <div class="dashboard-actions">
              <button class="dashboard-action-button" type="button" @click="$router.push('/admin/products')">
                <span class="dashboard-action-button__icon" style="background:#fffbeb;color:#f59e0b;"><CircleCheck /></span>
                <span class="dashboard-action-button__text">
                  <strong>商品管理</strong>
                </span>
              </button>
              <button class="dashboard-action-button" type="button" @click="$router.push('/admin/users')">
                <span class="dashboard-action-button__icon" style="background:#eff6ff;color:#3b82f6;"><User /></span>
                <span class="dashboard-action-button__text">
                  <strong>用户管理</strong>
                </span>
              </button>
              <button class="dashboard-action-button" type="button" @click="$router.push('/admin/orders')">
                <span class="dashboard-action-button__icon" style="background:#f5f3ff;color:#8b5cf6;"><List /></span>
                <span class="dashboard-action-button__text">
                  <strong>订单中心</strong>
                </span>
              </button>
              <button class="dashboard-action-button" type="button" @click="$router.push('/')">
                <span class="dashboard-action-button__icon" style="background:#ecfdf5;color:#10b981;"><HomeFilled /></span>
                <span class="dashboard-action-button__text">
                  <strong>前台首页</strong>
                </span>
              </button>
            </div>
          </article>


        </div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { resolveUrl } from '@/utils/request'
import request from '@/utils/request'
import { ElMessage } from '@/ui/feedback'
import gsap from 'gsap'
import {
  User,
  Goods,
  List,
  Wallet,
  Top,
  CircleCheck,
  HomeFilled,
  Operation
} from '@/icons/tw-icons.js'

const loading = ref(true)
const heroRef = ref(null)
const statsRef = ref(null)
const ordersRef = ref(null)
const actionsRef = ref(null)

const currentDate = new Date().toLocaleDateString('zh-CN', {
  month: 'long',
  day: 'numeric',
  weekday: 'long'
})

const users = ref([])
const products = ref([])
const orders = ref([])

const fixUrl = (url) => resolveUrl(url)
const resolveAvatar = (url) => (url ? resolveUrl(url) : '')

const fetchAllData = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const headers = { Authorization: token }

    const [resUsers, resProducts, resOrders] = await Promise.all([
      request.get('/api/admin/users', { headers }),
      request.get('/api/admin/products', { headers }),
      request.get('/api/admin/orders', { headers })
    ])

    users.value = Array.isArray(resUsers) ? resUsers : (resUsers.data || [])
    products.value = Array.isArray(resProducts) ? resProducts : (resProducts.data || [])
    orders.value = Array.isArray(resOrders) ? resOrders : (resOrders.data || [])
  } catch (error) {
    console.error(error)
    ElMessage.error('数据加载失败')
  } finally {
    loading.value = false
    // Trigger GSAP animations after data loads
    setTimeout(initAnimations, 100)
  }
}

const initAnimations = () => {
  // Hero entrance
  if (heroRef.value) {
    gsap.from(heroRef.value.querySelectorAll('.hero-headline, .hero-sub'), {
      y: 40,
      opacity: 0,
      duration: 0.8,
      stagger: 0.1,
      ease: 'power3.out'
    })
  }

  // Stats cards stagger
  if (statsRef.value) {
    gsap.from(statsRef.value.querySelectorAll('.dashboard-card'), {
      y: 50,
      opacity: 0,
      duration: 0.7,
      stagger: 0.08,
      ease: 'power3.out',
      delay: 0.2
    })
  }

  // Panels reveal
  const panels = [ordersRef.value, actionsRef.value].filter(Boolean)
  panels.forEach((panel, i) => {
    gsap.from(panel, {
      y: 40,
      opacity: 0,
      duration: 0.7,
      ease: 'power2.out',
      delay: 0.4 + i * 0.1
    })
  })
}

const totalTradeAmount = computed(() => {
  return orders.value.reduce((sum, order) => sum + (Number(order.price) || 0), 0).toFixed(2)
})

const activeUserCount = computed(() => users.value.filter((user) => user.status === 1).length)

const statsCards = computed(() => [
  { label: '总用户', value: users.value.length, icon: User, trend: '账户', color: '#3b82f6', bg: '#eff6ff' },
  { label: '商品', value: products.value.length, icon: Goods, trend: '内容', color: '#f59e0b', bg: '#fffbeb' },
  { label: '订单', value: orders.value.length, icon: List, trend: '交易', color: '#10b981', bg: '#ecfdf5' },
  { label: '交易额', value: totalTradeAmount.value, icon: Wallet, prefix: '¥', trend: '累计', color: '#8b5cf6', bg: '#f5f3ff' }
])

const recentOrders = computed(() => {
  return [...orders.value]
    .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    .slice(0, 5)
})



const formatTime = (iso) => {
  if (!iso) return '-'
  const date = new Date(iso)
  return `${date.getMonth() + 1}-${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
}

const getOrderStatusText = (status) => ({ 1: '待付', 2: '待发', 3: '运送', 4: '完成', 5: '取消' }[status] || '未知')

const getOrderStatusClass = (status) => {
  if (status === 4) return 'success'
  if (status === 1) return 'warning'
  if (status === 5) return 'gray'
  return 'primary'
}

const getInitial = (name) => (name ? String(name).slice(0, 1).toUpperCase() : 'U')

onMounted(fetchAllData)
</script>

<style scoped lang="scss">
@use 'sass:color';

// ── Color Tokens ──
$ink: #111827; /* Richer dark gray instead of pure black/soft ink */
$ink-soft: #4b5563;
$ink-muted: #6b7280;
$surface: #f9fafb;
$surface-raised: #ffffff;
$border: rgba(17, 24, 39, 0.05); /* Whisper-thin border */
$border-strong: rgba(17, 24, 39, 0.1);
$accent: #111827; /* Sophisticated accent */
$accent-soft: rgba(17, 24, 39, 0.04);

// ── Layout ──
.dashboard-shell {
  padding: 0;
  height: 100%;
  overflow-x: hidden;
}

// ── Hero Section ──
.dashboard-hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding: 32px 32px 28px;
  margin: 0 0 24px;
  background:
    radial-gradient(ellipse 80% 60% at 20% 20%, rgba(255, 255, 255, 0.8), transparent),
    radial-gradient(ellipse 60% 80% at 80% 80%, rgba(240, 245, 250, 0.5), transparent),
    linear-gradient(160deg, #ffffff 0%, #f8fafc 100%);
  border-radius: 24px;
  border: 1px solid $border;
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: -50%;
    right: -20%;
    width: 500px;
    height: 500px;
    background: radial-gradient(circle, rgba(17, 24, 39, 0.03), transparent 70%);
    pointer-events: none;
  }
}

.hero-content {
  position: relative;
  z-index: 1;
}

.dashboard-hero__eyebrow {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: $accent;
  margin: 0 0 10px;
}

.hero-headline {
  font-family: 'Outfit', -apple-system, sans-serif;
  font-size: clamp(28px, 3.5vw, 40px);
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1.1;
  color: $ink;
  margin: 0;
}

.hero-sub {
  font-size: 15px;
  color: $ink-muted;
  margin: 10px 0 0;
  line-height: 1.5;
}



// ── Stats Cards ──
.dashboard-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  padding: 0 0 24px;
}

.bento-card {
  background: $surface-raised;
  border: 1px solid $border;
  border-radius: 20px;
  padding: 24px;
  box-shadow: 0 10px 30px -10px rgba(17, 24, 39, 0.04);
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1),
              box-shadow 0.3s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 20px 40px -12px rgba(17, 24, 39, 0.08);
  }
}

.dashboard-card {
  display: grid;
  gap: 14px;
}

.dashboard-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dashboard-card__icon {
  width: 44px;
  height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  font-size: 20px;
  transition: transform 0.2s ease;

  .bento-card:hover & {
    transform: scale(1.08);
  }
}

.dashboard-card__meta {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: $ink-muted;
  font-weight: 500;
}

.dashboard-card__label {
  font-size: 13px;
  color: $ink-muted;
  font-weight: 500;
  margin-bottom: 4px;
}

.dashboard-card__value {
  font-size: 28px;
  font-weight: 700;
  color: $ink;
  letter-spacing: -0.02em;
  font-family: 'JetBrains Mono', monospace;

  span {
    font-size: 20px;
    color: $ink-soft;
    margin-right: 2px;
  }
}

// ── Main Grid ──
.dashboard-grid {
  display: grid;
  grid-template-columns: 1.4fr 1fr;
  gap: 16px;
  align-items: start;
}

.dashboard-side {
  display: grid;
  gap: 16px;
}

// ── Panel ──
.dashboard-panel {
  overflow: hidden;
}

.dashboard-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.dashboard-panel__title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: $ink;
  margin: 0;

  .panel-icon {
    width: 32px;
    height: 32px;
    border-radius: 10px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
  }
}

.dashboard-link-button {
  font-size: 13px;
  font-weight: 600;
  color: $accent;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 8px;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);

  &:hover {
    background: $accent-soft;
    transform: translateY(-1px);
  }
  
  &:active {
    transform: scale(0.95);
  }
}

// ── Orders Table ──
.dashboard-orders {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;

  thead th {
    text-align: left;
    padding: 10px 8px;
    font-size: 11px;
    font-weight: 600;
    color: $ink-muted;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    border-bottom: 1px solid $border;
  }

  tbody td {
    padding: 12px 8px;
    border-bottom: 1px solid $border;
    vertical-align: middle;
  }

  tbody tr:last-child td {
    border-bottom: none;
  }

  tbody tr {
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);

    &:hover {
      background: $surface;
      transform: translateX(4px);
    }
  }
}

.dashboard-product {
  display: flex;
  align-items: center;
  gap: 10px;

  &__thumb {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    object-fit: cover;
    border: 1px solid $border;
  }

  &__name {
    font-weight: 600;
    color: $ink;
    font-size: 13px;
  }

  &__sub {
    font-size: 11px;
    color: $ink-muted;
    margin-top: 2px;
    font-family: 'JetBrains Mono', monospace;
  }
}

.dashboard-user {
  display: flex;
  align-items: center;
  gap: 8px;

  &__avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    object-fit: cover;
    border: 1px solid $border;

    &.fallback {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      background: linear-gradient(135deg, #e4e4e7, #d4d4d8);
      color: $ink-soft;
      font-size: 11px;
      font-weight: 700;
    }
  }

  &__name {
    font-size: 13px;
    color: $ink;
    font-weight: 500;
  }
}

.dashboard-money {
  font-family: 'JetBrains Mono', monospace;
  font-weight: 600;
  color: $ink;
  font-size: 13px;
}

.dashboard-time {
  font-size: 12px;
  color: $ink-muted;
  font-family: 'JetBrains Mono', monospace;
}

// ── Status Pills ──
.dashboard-status {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 600;

  &::before {
    content: '';
    width: 5px;
    height: 5px;
    border-radius: 50%;
  }

  &.success {
    background: rgba(21, 128, 61, 0.08);
    color: #15803d;
    &::before { background: #15803d; }
  }

  &.warning {
    background: rgba(180, 83, 9, 0.08);
    color: #b45309;
    &::before { background: #b45309; }
  }

  &.gray {
    background: rgba(113, 113, 122, 0.08);
    color: #71717a;
    &::before { background: #71717a; }
  }

  &.primary {
    background: rgba(59, 130, 246, 0.08);
    color: #3b82f6;
    &::before { background: #3b82f6; }
  }
}

// ── Quick Actions ──
.dashboard-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.dashboard-action-button {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 14px;
  background: $surface;
  border: 1px solid $border;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  text-align: left;

  &:hover {
    background: $surface-raised;
    border-color: $border-strong;
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(17, 24, 39, 0.05);
  }

  &:active {
    transform: scale(0.96);
  }

  &__icon {
    width: 36px;
    height: 36px;
    border-radius: 10px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    flex-shrink: 0;
  }

  &__text {
    font-size: 13px;
    color: $ink-soft;

    strong {
      display: block;
      color: $ink;
      font-weight: 600;
      font-size: 14px;
    }
  }
}



// ── Skeleton ──
.skeleton-board {
  padding: 24px;
}

// ── Responsive ──
@media (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }

  .dashboard-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .dashboard-hero {
    flex-direction: column;
    align-items: flex-start;
    gap: 20px;
    padding: 24px;
  }



  .dashboard-stats {
    grid-template-columns: 1fr;
  }

  .dashboard-actions {
    grid-template-columns: 1fr;
  }

  .dashboard-orders {
    font-size: 12px;

    thead th, tbody td {
      padding: 8px 6px;
    }
  }
}
</style>
