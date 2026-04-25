<template>
  <div class="order-page">
    <nav class="navbar">
      <div class="container navbar-inner">
        <BackHomePill />
        <div class="page-title">
          <IconifySymbol icon="lucide:shopping-bag" size="18" color="#1f2532" />
          <span>我卖出的</span>
        </div>
        <div class="placeholder"></div>
      </div>
    </nav>

    <div class="tabs-wrapper">
      <div class="container">
        <div class="custom-tabs">
          <div v-for="tab in tabs" :key="tab.value" class="tab-item" :class="{ active: currentStatus === tab.value }" @click="switchTab(tab.value)">
            <IconifySymbol :icon="tab.icon" size="15" :color="currentStatus === tab.value ? '#202633' : '#7b8596'" />
            <span>{{ tab.label }}</span>
            <div class="slider" v-if="currentStatus === tab.value"></div>
          </div>
        </div>
      </div>
    </div>

    <div class="container content-area">
      <div class="order-list" v-loading="loading">
        <div v-for="order in orderList" :key="order.id" class="order-card animate-up">
          <div class="card-header">
            <div class="shop-info">
              <el-avatar :size="24" :src="order.user?.avatar || defaultAvatar" />
              <span class="shop-name">买家: {{ order.user?.nickname || order.user?.username || '匿名用户' }}</span>
              <IconifySymbol icon="lucide:chevron-right" size="16" color="#9aa3b2" />
            </div>
            <div class="order-status" :class="getStatusClass(order.status)">
              <IconifySymbol
                :icon="getStatusIcon(order.status)"
                size="14"
                :color="order.status === 3 ? '#2da460' : order.status === 2 ? '#d28508' : order.status === 1 ? '#ff5e4d' : '#7b8596'"
              />
              {{ getStatusText(order.status) }}
            </div>
          </div>

          <div class="card-body" @click="$router.push(`/product/${order.product?.id}`)">
            <div class="img-wrapper">
              <img :src="order.product?.image" class="thumb" :class="{ 'is-sold': order.status >= 2 }" />
              <div v-if="order.status >= 2" class="sold-overlay">已卖出</div>
            </div>
            <div class="info">
              <div class="title">{{ order.product?.name || '未知商品' }}</div>
              <div class="tags">
                <span class="tag" :class="getStatusTagClass(order.status)">
                  <IconifySymbol icon="lucide:tag" size="12" :color="getStatusTagColor(order.status)" />
                  {{ getStatusText(order.status) }}
                </span>
              </div>
            </div>
            <div class="price-col">
              <div class="price"><small>¥</small>{{ order.price }}</div>
              <div class="qty">x1</div>
            </div>
          </div>

          <div class="card-footer">
            <div class="summary">订单号 {{ order.order_no || order.id }}</div>
            <div class="actions">
              <button class="btn-outline" @click="contactBuyer(order)">
                <IconifySymbol icon="lucide:message-circle" size="13" color="#646f7e" />
                联系买家
              </button>

              <button class="btn-primary" v-if="order.status === 1" @click="shipOrder(order)">
                <IconifySymbol icon="lucide:truck" size="13" color="#d28508" />
                去发货
              </button>
              <button class="btn-outline" v-if="order.status === 2">
                <IconifySymbol icon="lucide:clock-3" size="13" color="#646f7e" />
                等待收货
              </button>
            </div>
          </div>
        </div>
        <div v-if="!loading && orderList.length === 0" class="empty-state">
          <div class="empty-art">
            <IconifySymbol icon="lucide:inbox" size="58" color="#c2cad6" />
          </div>
          <div class="empty-title">您还没有卖出过宝贝</div>
          <div class="empty-sub">先发布几件，马上开始流转</div>
          <button class="go-publish-btn" @click="$router.push('/publish')">去发布</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request, { resolveUrl } from '@/utils/request'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import IconifySymbol from '@/components/IconifySymbol.vue'
import BackHomePill from '@/components/BackHomePill.vue'

const router = useRouter()
const loading = ref(false)
const orderList = ref([])
const currentStatus = ref(0)
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const tabs = [
  { label: '全部', value: 0, icon: 'lucide:list' },
  { label: '待发货', value: 2, icon: 'lucide:package-check' },
  { label: '已发货', value: 3, icon: 'lucide:truck' },
  { label: '已完成', value: 4, icon: 'lucide:badge-check' }
]

const allOrders = ref([])

const fetchOrders = async () => {
  if (!user.value.id) return router.push('/')
  loading.value = true
  try {
    const params = {
      user_id: user.value.id,
      role: 'seller'
    }
    const res = await request.get('/api/orders', { params })

    // 图片路径修复逻辑
    allOrders.value = (res.data || []).map(order => {
      if (order.product && order.product.image) {
        order.product.image = resolveUrl(order.product.image)
      }
      // 买家头像修复
      if (order.user && order.user.avatar) {
        order.user.avatar = resolveUrl(order.user.avatar)
      }
      return order
    })
    applyFilter()
  } catch (e) { console.error(e) } finally { loading.value = false }
}

const applyFilter = () => {
  if (currentStatus.value === 0) {
    orderList.value = allOrders.value
  } else {
    orderList.value = allOrders.value.filter(o => o.status === currentStatus.value)
  }
}

const switchTab = (val) => { currentStatus.value = val; applyFilter() }

const contactBuyer = (order) => {
  // order.user_id 是买家ID
  if (order.user_id) {
    router.push(`/chat/${order.user_id}`)
  }
}

// 模拟发货操作
const shipOrder = (order) => {
  ElMessageBox.confirm('确认已发货？', '发货确认').then(() => {
    // 这里需要后端支持发货接口，暂时仅前端模拟
    ElMessage.success('发货成功')
    // 实际应调用: await request.put(`/api/orders/${order.id}/ship`)
  })
}

const getStatusText = (status) => {
  switch (status) {
    case 1: return '待付款';
    case 2: return '待发货';
    case 3: return '已发货';
    case 4: return '已完成';
    case 5: return '已取消';
    default: return '未知'
  }
}
const getStatusClass = (status) => {
  switch (status) {
    case 4: return 'success';
    case 2: case 3: return 'warning';
    case 1: return 'danger';
    case 5: return 'muted';
    default: return ''
  }
}

const getStatusIcon = (status) => {
  switch (status) {
    case 1: return 'lucide:clock-3'
    case 2: return 'lucide:package-check'
    case 3: return 'lucide:truck'
    case 4: return 'lucide:badge-check'
    case 5: return 'lucide:circle-x'
    default: return 'lucide:circle-help'
  }
}

const getStatusTagClass = (status) => {
  switch (status) {
    case 4: return 'tag-success';
    case 2: case 3: return 'tag-warning';
    case 1: return 'tag-danger';
    case 5: return 'tag-muted';
    default: return ''
  }
}

const getStatusTagColor = (status) => {
  switch (status) {
    case 4: return '#2da460';
    case 2: case 3: return '#d28508';
    case 1: return '#ff5e4d';
    case 5: return '#999';
    default: return '#7b8596'
  }
}

onMounted(fetchOrders)
</script>

<style scoped lang="scss">
/* 复用 UserOrders.vue 的样式，完全一致 */
$primary: #ffdf5d;
$bg: #f6f7f9;

.order-page { min-height: 100vh; background: $bg; padding-top: 110px; }
.container { max-width: 800px; margin: 0 auto; padding: 0 20px; }
.navbar {
  height: 60px; background: #fff; position: fixed; top: 0; left: 0; right: 0; z-index: 100; border-bottom: 1px solid #f0f0f0;
  .navbar-inner { height: 100%; display: flex; align-items: center; justify-content: space-between; }
  .page-title { display: inline-flex; align-items: center; gap: 8px; font-weight: 800; font-size: 18px; color: #202633; }
  .placeholder { width: 92px; }
}
.tabs-wrapper { position: fixed; top: 60px; left: 0; right: 0; z-index: 99; background: #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.03); }
.custom-tabs {
  display: flex; gap: 30px; padding: 0 10px;
  .tab-item {
    position: relative; padding: 14px 0; cursor: pointer; font-size: 15px; color: #666; transition: 0.3s; display: inline-flex; align-items: center; gap: 6px;
    &:hover { color: #333; }
    &.active { font-weight: bold; color: #000; font-size: 16px; }
    .slider { position: absolute; bottom: 0; left: 50%; transform: translateX(-50%); width: 20px; height: 3px; background: $primary; border-radius: 3px; }
  }
}
.content-area { padding-top: 20px; padding-bottom: 40px; }
.order-card {
  background: #fff; border-radius: 16px; padding: 16px; margin-bottom: 16px; box-shadow: 0 2px 8px rgba(0,0,0,0.02); transition: 0.2s;
  &:hover { transform: translateY(-2px); box-shadow: 0 8px 16px rgba(0,0,0,0.05); }
}
.card-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; padding-bottom: 12px; border-bottom: 1px solid #f9f9f9;
  .shop-info { display: flex; align-items: center; gap: 8px; font-size: 14px; font-weight: bold; }
  .order-status { font-size: 13px; color: #666; font-weight: 500; display: inline-flex; align-items: center; gap: 4px;
    &.success { color: #00b578; }
    &.warning { color: #ff9500; }
    &.danger { color: #ff4d4f; }
  }
}
.card-body {
  display: flex; gap: 12px; cursor: pointer;
  .img-wrapper {
    position: relative; width: 80px; height: 80px; border-radius: 8px; overflow: hidden;
    .thumb { width: 100%; height: 100%; object-fit: cover; background: #f9f9f9; transition: 0.3s; &.is-sold { filter: grayscale(100%); opacity: 0.8; } }
    .sold-overlay {
      position: absolute; inset: 0; display: flex; align-items: center; justify-content: center;
      background: rgba(0,0,0,0.5); color: #fff; font-size: 12px; font-weight: bold;
    }
  }
  .info { flex: 1; .title { font-size: 14px; font-weight: bold; margin-bottom: 8px; line-height: 1.4; height: 40px; overflow: hidden; } .tags { display: flex; gap: 4px; .tag { font-size: 10px; background: #fef0f0; color: #ff4d4f; padding: 2px 6px; border-radius: 4px; display: inline-flex; align-items: center; gap: 4px; } } }
  .price-col { text-align: right; .price { font-size: 16px; font-weight: bold; } .qty { font-size: 12px; color: #999; margin-top: 4px; } }
}
.card-footer {
  margin-top: 16px; display: flex; justify-content: space-between; align-items: center;
  .summary { font-size: 12px; color: #999; }
  .actions {
    display: flex; gap: 10px;
    button { padding: 6px 16px; border-radius: 99px; font-size: 13px; cursor: pointer; font-weight: 600; transition: 0.2s; background: #fff; border: 1px solid #ddd; display: inline-flex; align-items: center; gap: 5px; }
    .btn-primary { background: #fff; border: 1px solid $primary; color: #d48806; font-weight: bold; &:hover { background: #fffcf0; } }
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 84px 0 20px;

  .empty-art {
    width: 112px;
    height: 112px;
    border-radius: 24px;
    border: 1px solid #e6eaf2;
    background: linear-gradient(152deg, #fafbfd, #eef2f7);
    box-shadow: inset 0 1px 0 rgba(255,255,255,0.7);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .empty-title {
    margin-top: 14px;
    font-size: 22px;
    font-weight: 800;
    color: #293245;
  }

  .empty-sub {
    margin-top: 6px;
    font-size: 13px;
    color: #8892a2;
    font-weight: 600;
  }

  .go-publish-btn {
    margin-top: 20px;
    height: 40px;
    padding: 0 28px;
    border: none;
    border-radius: 999px;
    background: #111;
    color: #ffdf5d;
    font-size: 14px;
    font-weight: 800;
    cursor: pointer;
    transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), box-shadow 0.2s cubic-bezier(0.22, 1, 0.36, 1);
    &:hover { transform: translateY(-2px); box-shadow: 0 10px 18px rgba(0, 0, 0, 0.16); }
  }
}
.animate-up { animation: fadeInUp 0.5s ease backwards; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

/* ========== 手机端适配 ========== */
@media (max-width: 480px) {
  .nav-header {
    height: 50px;
    padding: 0 10px;
    .nav-inner h1 { font-size: 16px; }
  }

  .tabs-bar {
    .tabs-inner { gap: 15px; justify-content: center; }
    .tab-item {
      font-size: 13px;
      &.active { font-size: 14px; }
    }
  }

  .content-area { padding-top: 10px; }

  .order-card {
    padding: 12px;
    .card-header {
      margin-bottom: 10px;
      .shop-info { font-size: 12px; }
      .order-status { font-size: 11px; }
    }

    .card-body {
      gap: 10px;
      .img-wrapper { width: 70px; height: 70px; }
      .info {
        .title { font-size: 13px; height: auto; max-height: 36px; -webkit-line-clamp: 2; display: -webkit-box; -webkit-box-orient: vertical; margin-bottom: 4px; }
        .tags .tag { font-size: 9px; }
      }
      .price-col {
        .price { font-size: 14px; }
        .qty { font-size: 10px; }
      }
    }

    .card-footer {
      flex-direction: column;
      align-items: flex-end;
      gap: 10px;
      margin-top: 12px;
      .summary { font-size: 11px; }
      .actions {
        width: 100%;
        justify-content: flex-end;
        button { padding: 5px 12px; font-size: 12px; }
      }
    }
  }
}
</style>
