<template>
  <div class="orders-page" v-loading="loading">

    <div class="nav-header">
      <div class="container nav-content">
        <BackHomePill />
        <div class="page-title">
          <IconifySymbol icon="lucide:clipboard-list" size="18" color="#202633" />
          <span>我的订单</span>
        </div>
        <div class="placeholder"></div>
      </div>
    </div>

    <div class="tabs-bar">
      <div class="container tabs-inner">
        <div
            v-for="tab in tabs"
            :key="tab.value"
            class="tab-item"
            :class="{ active: currentTab === tab.value }"
            @click="currentTab = tab.value"
        >
          <IconifySymbol :icon="tab.icon" size="15" :color="currentTab === tab.value ? '#202633' : '#7b8596'" />
          <span>{{ tab.label }}</span>
          <div class="active-line" v-if="currentTab === tab.value"></div>
        </div>
      </div>
    </div>

    <div class="container main-content">
      <div v-if="filteredOrders.length === 0 && !loading" class="empty-state">
        <div class="empty-art">
          <IconifySymbol icon="lucide:package-open" size="62" color="#c3cad6" />
        </div>
        <div class="empty-title">暂无相关订单</div>
        <div class="empty-sub">先去首页看看新上架的好物</div>
        <button class="go-home-btn" @click="$router.push('/')">去逛逛</button>
      </div>

      <div v-else class="order-list">
        <div v-for="order in filteredOrders" :key="order.id" class="order-card animate-up">

          <div class="card-header">
            <div class="shop-info" @click="contactSeller(order.seller?.id)">
              <el-avatar :size="24" :src="order.seller?.avatar || defaultAvatar" />
              <span class="shop-name">{{ order.seller?.nickname || order.seller?.username || '闲趣卖家' }}</span>
              <IconifySymbol icon="lucide:chevron-right" size="16" color="#9aa3b2" />
            </div>
            <div class="status-badge" :class="getStatusClass(order.status)">
              <IconifySymbol
                :icon="getStatusIcon(order.status)"
                size="14"
                :color="order.status === 1 ? '#ff6a32' : order.status === 4 ? '#31a45f' : order.status === 5 ? '#9aa1ad' : '#4f6fa8'"
              />
              {{ getStatusText(order.status) }}
            </div>
          </div>

          <div class="card-body" @click="$router.push(`/product/${order.product_id}`)">
            <div class="img-box">
              <el-image :src="fixUrl(order.product?.image)" fit="cover" class="prod-img" lazy>
                <template #error>
                  <div class="err-slot">
                    <IconifySymbol icon="lucide:image-off" size="24" color="#a7b0bf" />
                  </div>
                </template>
              </el-image>
            </div>

            <div class="info-box">
              <div class="prod-title">{{ order.product?.name || order.product?.title || '商品信息已失效' }}</div>
              <div class="prod-desc">{{ order.product?.description || '暂无描述...' }}</div>
              <div class="tags">
                <span class="tag"><IconifySymbol icon="lucide:truck" size="12" color="#b57b08" /> 包邮</span>
                <span class="tag"><IconifySymbol icon="lucide:shield-check" size="12" color="#b57b08" /> 担保交易</span>
              </div>
            </div>

            <div class="price-box">
              <div class="price"><span class="symbol">¥</span>{{ order.price }}</div>
              <div class="qty">x 1</div>
            </div>
          </div>

          <div class="card-footer">
            <div class="total-info">
              实付: <span class="amount">¥ {{ order.price }}</span>
            </div>

            <div class="actions">
              <!-- 加入倒计时显示 -->
              <div v-if="order.status === 1" class="countdown-text" style="color: #ff5000; font-size: 13px; font-weight: bold; margin-right: 12px; display: flex; align-items: center;">
                <IconifySymbol icon="lucide:timer" size="14" style="margin-right:4px;" />
                {{ getCountdownStr(order) }}
              </div>

              <button class="btn btn-contact" @click="contactSeller(order.seller?.id)">
                <IconifySymbol icon="lucide:message-circle" size="13" color="#646f7e" />
                联系卖家
              </button>

              <button
                  v-if="order.status === 1"
                  class="btn btn-primary"
                  @click="payNow(order)"
              >
                <IconifySymbol icon="lucide:wallet" size="13" color="#ffdf5d" />
                立即支付
              </button>

              <button
                  v-if="order.status === 2 || order.status === 3"
                  class="btn btn-confirm"
                  @click="confirmReceive(order.id)"
              >
                <IconifySymbol icon="lucide:badge-check" size="13" color="#c57c10" />
                确认收货
              </button>

              <button v-if="order.status === 4" class="btn btn-outline">评价</button>
              
              <button 
                  v-if="[1, 2, 3].includes(order.status)" 
                  class="btn btn-outline" 
                  style="color: #ff4d4f; border-color: #ff4d4f; background: #fff1f0;"
                  @click="handleRefund(order)"
              >
                  {{ order.status === 1 ? '取消订单' : '申请退款' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import request, { resolveUrl } from '@/utils/request'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import IconifySymbol from '@/components/IconifySymbol.vue'
import BackHomePill from '@/components/BackHomePill.vue'

const router = useRouter()
const loading = ref(false)
const orders = ref([])
// 0=全部, 1=待支付, 2=待发货/运输中, 4=已完成
const currentTab = ref(0)
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const systemTimeDiff = ref(0)
const now = ref(Date.now())
let timerId = null

const tabs = [
  { label: '全部', value: 0, icon: 'lucide:list' },
  { label: '待付款', value: 1, icon: 'lucide:clock-3' },
  { label: '待发货', value: 2, icon: 'lucide:truck' },
  { label: '已完成', value: 4, icon: 'lucide:circle-check' }
]

// 修复图片路径
const fixUrl = (url) => resolveUrl(url)

const fetchOrders = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/orders')
    orders.value = (res.data || []).map(order => {
      if (order.seller && order.seller.avatar) {
        order.seller.avatar = resolveUrl(order.seller.avatar)
      }
      return order
    })
    if (res.system_time) {
      systemTimeDiff.value = Date.now() - new Date(res.system_time).getTime()
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

// 前端过滤 (后端目前是返回全部，如果数据量大建议后端过滤)
const filteredOrders = computed(() => {
  if (currentTab.value === 0) return orders.value

  // 待发货 Tab 同时显示 status 2(待发货) 和 3(运输中)
  if (currentTab.value === 2) {
    return orders.value.filter(o => o.status === 2 || o.status === 3)
  }

  return orders.value.filter(o => o.status === currentTab.value)
})

const getStatusText = (status) => {
  const map = { 1: '待付款', 2: '待发货', 3: '运输中', 4: '交易成功', 5: '已取消' }
  return map[status] || '未知状态'
}

const getStatusClass = (status) => {
  if (status === 1) return 'wait'
  if (status === 4) return 'success'
  if (status === 5) return 'cancel'
  return 'normal'
}

const getStatusIcon = (status) => {
  if (status === 1) return 'lucide:clock-3'
  if (status === 2 || status === 3) return 'lucide:truck'
  if (status === 4) return 'lucide:badge-check'
  if (status === 5) return 'lucide:circle-x'
  return 'lucide:circle-help'
}

const getCountdownStr = (order) => {
  if (order.status !== 1 || !order.expired_at) return ''
  const expiredTime = new Date(order.expired_at).getTime()
  const currentServerTime = now.value - systemTimeDiff.value
  const remainingMs = expiredTime - currentServerTime
  if (remainingMs <= 0) {
    if (order.status === 1) order.status = 5
    return '已超时关闭'
  }
  const mm = String(Math.floor((remainingMs / 1000 / 60) % 60)).padStart(2, '0')
  const ss = String(Math.floor((remainingMs / 1000) % 60)).padStart(2, '0')
  return `剩余支付时间 ${mm}:${ss}`
}

const contactSeller = (sellerId) => {
  if (sellerId) router.push(`/chat/${sellerId}`)
}

const buildPayPath = (payUrl, extraQuery = {}) => {
  const [path = '/pay/mock', rawQuery = ''] = String(payUrl || '/pay/mock').split('?')
  const params = new URLSearchParams(rawQuery)
  Object.entries(extraQuery).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      params.set(key, String(value))
    }
  })
  const query = params.toString()
  return query ? `${path}?${query}` : path
}

const payNow = async (order) => {
  try {
    loading.value = true
    // 请求后端获取支付链接
    const res = await request.post(`/api/orders/${order.id}/pay`)
    if (res.pay_url) {
      const payPath = buildPayPath(res.pay_url, {
        product_id: order.product_id,
        product_name: order.product?.name || order.product?.title || '',
        price: order.price,
        quantity: 1,
        return_to: '/orders',
        source: 'order_list',
        source_path: '/orders'
      })
      if (/^https?:\/\//i.test(payPath)) {
        window.location.href = payPath
        return
      }
      router.push(payPath)
    } else {
      ElMessage.error('获取支付链接失败')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.error || '支付发起失败')
  } finally {
    loading.value = false
  }
}

const confirmReceive = async (id) => {
  try {
    await request.put(`/api/orders/${id}/confirm`)
    ElMessage.success('收货成功！交易完成')
    // 刷新列表以更新状态
    fetchOrders()
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  }
}

const handleRefund = (order) => {
  if (order.status === 1) {
    // 待付款订单：直接支持取消操作
    ElMessageBox.confirm(
      '确定要取消这个订单吗？取消后商品将重新上架供他人购买。',
      '取消确认',
      {
        confirmButtonText: '确定取消',
        cancelButtonText: '再想想',
        type: 'warning'
      }
    ).then(async () => {
      try {
        loading.value = true
        await request.put(`/api/orders/${order.id}/refund`)
        ElMessage.success('订单已取消，商品已恢复上架')
        fetchOrders()
      } catch (e) {
        ElMessage.error(e.response?.data?.error || '取消失败')
      } finally {
        loading.value = false
      }
    }).catch(() => {})
  } else {
    // 已支付订单：引导联系卖家协商
    ElMessageBox.confirm(
      '当前自助退单/退款功能尚未开放，如有需求请直接联系卖家协商处理。',
      '功能未开放',
      {
        confirmButtonText: '联系卖家',
        cancelButtonText: '取消',
        type: 'warning',
        center: true
      }
    ).then(() => {
      contactSeller(order.seller?.id)
    }).catch(() => {})
  }
}

onMounted(() => {
  fetchOrders()
  timerId = setInterval(() => { now.value = Date.now() }, 1000)
})

onUnmounted(() => {
  if (timerId) clearInterval(timerId)
})
</script>

<style scoped lang="scss">
$primary: #ffdf5d;
$bg-page: #f6f7f9;

.orders-page { min-height: 100vh; background: $bg-page; padding-bottom: 40px; }

/* 顶部导航 */
.nav-header {
  height: 60px; background: #fff; position: sticky; top: 0; z-index: 100; box-shadow: 0 2px 8px rgba(0,0,0,0.02);
  .nav-content { height: 100%; display: flex; align-items: center; justify-content: space-between; }
  .page-title { font-weight: 800; font-size: 18px; color: #333; display: flex; align-items: center; gap: 8px; }
  .placeholder { width: 80px; }
}

/* Tabs */
.tabs-bar {
  background: #fff; margin-bottom: 20px; border-top: 1px solid #f9f9f9;
  .tabs-inner { display: flex; gap: 30px; }
  .tab-item {
    padding: 14px 0; font-size: 15px; color: #666; cursor: pointer; position: relative; transition: 0.2s;
    display: inline-flex; align-items: center; gap: 6px;
    &:hover { color: #333; }
    &.active { font-weight: 800; color: #333; font-size: 16px; }
    .active-line { position: absolute; bottom: 0; left: 50%; transform: translateX(-50%); width: 24px; height: 3px; background: $primary; border-radius: 2px; }
  }
}

.container { max-width: 800px; margin: 0 auto; padding: 0 16px; }

/* 订单卡片 */
.order-card {
  background: #fff; border-radius: 16px; padding: 20px; margin-bottom: 16px; box-shadow: 0 4px 12px rgba(0,0,0,0.02); transition: 0.2s;
  &:hover { box-shadow: 0 8px 20px rgba(0,0,0,0.05); transform: translateY(-2px); }

  .card-header {
    display: flex; justify-content: space-between; align-items: center; padding-bottom: 14px; border-bottom: 1px dashed #eee; margin-bottom: 14px;
    .shop-info { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; font-weight: bold; color: #333; &:hover { opacity: 0.8; } }
    .status-badge { font-size: 13px; font-weight: bold; display: inline-flex; align-items: center; gap: 4px; &.wait { color: #ff5000; } &.success { color: #52c41a; } &.cancel { color: #999; } &.normal { color: #1677ff; } }
  }

  .card-body {
    display: flex; gap: 16px; cursor: pointer;
    .img-box { width: 90px; height: 90px; border-radius: 12px; overflow: hidden; background: #f9f9f9; .prod-img { width: 100%; height: 100%; transition: transform 0.3s; &:hover { transform: scale(1.05); } } .err-slot { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; background: #f2f4f8; } }
    .info-box { flex: 1; .prod-title { font-size: 15px; font-weight: bold; color: #333; margin-bottom: 6px; line-height: 1.4; max-height: 42px; overflow: hidden; } .prod-desc { font-size: 13px; color: #999; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-bottom: 8px; } .tags .tag { font-size: 10px; background: #fffbe6; color: #d48806; padding: 2px 6px; border-radius: 4px; margin-right: 6px; display: inline-flex; align-items: center; gap: 3px; } }
    .price-box { text-align: right; .price { font-weight: 800; font-size: 16px; color: #333; } .qty { font-size: 12px; color: #999; margin-top: 4px; } }
  }

  .card-footer {
    display: flex; justify-content: space-between; align-items: center; margin-top: 20px; padding-top: 10px;
    .total-info { font-size: 13px; color: #666; .amount { font-size: 18px; font-weight: 800; color: #ff5000; margin-left: 4px; } }
    .actions {
      display: flex; gap: 10px;
      .btn { padding: 8px 20px; border-radius: 99px; font-size: 13px; cursor: pointer; border: 1px solid #ddd; background: #fff; font-weight: 600; transition: 0.2s; display: inline-flex; align-items: center; gap: 5px; }
      .btn-contact { color: #666; &:hover { border-color: #333; color: #333; } }
      .btn-primary { border: none; background: #1a1a1a; color: $primary; &:hover { background: #000; transform: translateY(-1px); box-shadow: 0 4px 10px rgba(0,0,0,0.15); } }
      .btn-confirm { border-color: $primary; color: #d48806; background: #fffbe6; &:hover { background: #fff1b8; } }
      .btn-outline { &:hover { background: #f5f5f5; } }
    }
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 74px;
  .empty-art {
    width: 112px;
    height: 112px;
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid #e6eaf2;
    background: linear-gradient(152deg, #fafbfd, #eef2f7);
    box-shadow: inset 0 1px 0 rgba(255,255,255,0.7);
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
  .go-home-btn { margin-top: 20px; padding: 10px 32px; background: #333; color: $primary; border: none; border-radius: 99px; font-weight: bold; cursor: pointer; transition: 0.2s; &:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); } }
}

.animate-up { animation: fadeInUp 0.4s ease; } @keyframes fadeInUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

/* ========== 手机端适配 ========== */
@media (max-width: 480px) {
  .nav-header {
    height: 50px;
    padding: 0 10px;
    .page-title { font-size: 16px; }
  }

  .tabs-bar {
    .tabs-inner { gap: 15px; justify-content: center; }
    .tab-item {
      font-size: 13px;
      &.active { font-size: 14px; }
    }
  }

  .container { padding: 0 12px; }

  .order-card {
    padding: 15px;
    .card-header {
      .shop-info { font-size: 12px; gap: 6px; }
      .status-badge { font-size: 11px; }
    }

    .card-body {
      gap: 10px;
      .img-box { width: 70px; height: 70px; }
      .info-box {
        .prod-title { font-size: 13px; margin-bottom: 4px; }
        .prod-desc { font-size: 11px; }
        .tags .tag { font-size: 9px; padding: 1px 4px; }
      }
      .price-box {
        .price { font-size: 14px; }
        .qty { font-size: 10px; }
      }
    }

    .card-footer {
      flex-direction: column;
      align-items: flex-end;
      gap: 12px;
      margin-top: 15px;
      .total-info { font-size: 11px; .amount { font-size: 16px; } }
      .actions {
        width: 100%;
        justify-content: flex-end;
        .btn { padding: 6px 14px; font-size: 12px; }
      }
    }
  }
}
</style>
