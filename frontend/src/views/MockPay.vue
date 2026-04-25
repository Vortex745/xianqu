<template>
  <div class="mock-pay-container">
    <div class="pay-card" v-loading="pageLoading || paying">
      <!-- 头部状态栏 -->
      <div class="pay-header">
        <div class="brand">
          <img v-if="payType === 'alipay'" src="/images/支付宝.svg" alt="AliPay" style="height: 28px;" />
          <img v-else src="/images/wechatpay.svg" alt="WeChatPay" style="height: 28px;" />
          <span class="sandbox-tag">沙盒环境</span>
        </div>
        <div class="order-amount">
          <span class="currency">¥</span>
          <span class="amount">{{ amount }}</span>
        </div>
        <div class="order-info">
          <p>订单号：{{ order.order_no || order.id }}</p>
          <p>商品：{{ displayProductName }}</p>
          <p>数量：x {{ quantity }}</p>
        </div>
      </div>

      <!-- 支付方式切换 -->
      <div class="pay-methods">
        <div class="method-item" :class="{ active: payType === 'alipay' }" @click="payType = 'alipay'">
          <img src="/images/支付宝.svg" class="pay-icon" />
          <span>支付宝</span>
          <el-icon class="check-icon" v-if="payType === 'alipay'"><Select /></el-icon>
        </div>
        <div class="method-item" :class="{ active: payType === 'wechat' }" @click="payType = 'wechat'">
          <img src="/images/wechatpay.svg" class="pay-icon" />
          <span>微信支付</span>
          <el-icon class="check-icon" v-if="payType === 'wechat'"><Select /></el-icon>
        </div>
      </div>

      <!-- 扫码区域 -->
      <div class="qr-section">
        <div class="qr-box" :class="payType">
          <!-- 模拟二维码 -->
          <img src="https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=MockPayment" alt="QRCode" />
          <div class="logo-center">
            <img :src="payType === 'alipay' ? '/images/支付宝.svg' : '/images/wechatpay.svg'" class="qr-center-icon" />
          </div>
        </div>
        <p class="scan-tip">
          请使用 <span :class="payType === 'alipay' ? 'ali-text' : 'wx-text'">{{ payType === 'alipay' ? '支付宝' : '微信' }}</span> 扫一扫
        </p>
        <p class="scan-subtip">二维码2小时内有效</p>
      </div>

      <!-- 模拟操作区 -->
      <div class="action-area">
        <button class="btn-confirm" :disabled="paying" @click="handleConfirmPay">
          {{ paying ? '支付中...' : '确认支付' }}
        </button>
        <div class="cancel-link" @click="handleCancel">取消支付，返回订单页</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import { Select } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const orderId = computed(() => {
  const raw = route.query.order_id
  return raw ? String(raw) : ''
})
const quantity = computed(() => {
  const raw = Number(route.query.quantity || 1)
  return Number.isFinite(raw) && raw > 0 ? raw : 1
})
const fallbackProductName = computed(() => {
  const raw = route.query.product_name
  if (!raw) return ''
  try {
    return decodeURIComponent(String(raw))
  } catch (e) {
    return String(raw)
  }
})
const amount = ref('0.00')
const order = ref({})
const payType = ref('alipay') // alipay | wechat
const pageLoading = ref(false)
const paying = ref(false)

const formatAmount = (value) => {
  const num = Number(value)
  return Number.isFinite(num) ? num.toFixed(2) : '0.00'
}

const normalizeOrderList = (payload) => {
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload?.data)) return payload.data
  if (Array.isArray(payload?.list)) return payload.list
  return []
}

const appendQuery = (targetPath, query = {}) => {
  const [path, rawQuery = ''] = String(targetPath || '/orders').split('?')
  const params = new URLSearchParams(rawQuery)
  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      params.set(key, String(value))
    }
  })
  const str = params.toString()
  return str ? `${path}?${str}` : path
}

const resolveReturnPath = (state = '') => {
  const raw = route.query.return_to
  let target = '/orders'
  if (raw && typeof raw === 'string') {
    try {
      const decoded = decodeURIComponent(raw)
      if (decoded.startsWith('/')) target = decoded
    } catch (e) {}
  }
  return appendQuery(target, {
    pay_state: state,
    order_id: orderId.value
  })
}

const displayProductName = computed(() => {
  return order.value.product?.name || order.value.product?.title || fallbackProductName.value || '未命名商品'
})

const backToReturnPath = (state = '') => {
  router.replace(resolveReturnPath(state))
}

// 初始化：获取订单详情
const fetchOrder = async () => {
  if (!orderId.value) {
    ElMessage.error('订单不存在')
    backToReturnPath('invalid_order')
    return
  }
  amount.value = formatAmount(route.query.amount || route.query.price || 0)
  pageLoading.value = true
  try {
    // 复用获取订单列表接口，但逻辑上应该通常用 getDetail，这里简化处理
    const res = await request.get('/api/orders')
    const list = normalizeOrderList(res)
    const target = list.find(o => String(o.id) === String(orderId.value))

    if (target) {
      order.value = target
      amount.value = formatAmount(target.price ?? target.amount)
      if (target.status !== 1) {
        if ([2, 3, 4].includes(target.status)) {
          ElMessage.info('该订单已支付，请勿重复支付')
          backToReturnPath('already_paid')
          return
        }
        ElMessage.warning('该订单状态异常，无法支付')
        backToReturnPath('invalid_status')
      }
    } else {
      ElMessage.error('找不到该订单')
      backToReturnPath('not_found')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.error || '网络错误')
  } finally {
    pageLoading.value = false
  }
}

// 模拟支付逻辑
const handleConfirmPay = async () => {
  if (!orderId.value || paying.value) return
  paying.value = true
  try {
    await request.post(`/api/orders/${orderId.value}/confirm_pay`)
    ElMessage.success('支付成功！')
    backToReturnPath('success')
  } catch (e) {
    const msg = e.response?.data?.error || '支付失败'
    ElMessage.error(msg)
    if (msg.includes('已支付')) {
      backToReturnPath('already_paid')
    }
  } finally {
    paying.value = false
  }
}

const handleCancel = () => {
  ElMessage.info('订单已保留，可在订单页继续支付')
  backToReturnPath('cancel')
}

onMounted(() => {
  fetchOrder()
})
</script>

<style scoped lang="scss">
$primary: #ffdf5d;
$dark: #1a1a1a;
$bg-page: #edf1f5;

.mock-pay-container {
  position: relative;
  min-height: 100dvh;
  height: 100dvh;
  background:
    radial-gradient(circle at 0% 0%, rgba(255, 223, 93, 0.22) 0%, rgba(255, 223, 93, 0) 34%),
    radial-gradient(circle at 100% 100%, rgba(0, 0, 0, 0.07) 0%, rgba(0, 0, 0, 0) 40%),
    linear-gradient(160deg, $bg-page 0%, #f7f8fa 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(8px, 2vh, 20px) clamp(10px, 3vw, 20px);
  font-family: "PingFang SC", sans-serif;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    pointer-events: none;
    background-image: radial-gradient(rgba(18, 18, 18, 0.06) 0.7px, transparent 0.7px);
    background-size: 4px 4px;
    opacity: 0.15;
  }

  > * {
    position: relative;
    z-index: 1;
  }

  .pay-icon { width: 24px; height: 24px; object-fit: contain; }
  .qr-center-icon { width: 100%; height: 100%; object-fit: contain; padding: 2px; background: #fff; border-radius: 4px; }
}

.pay-card {
  width: 100%;
  max-width: 420px;
  max-height: 100%;
  background: #fff;
  border-radius: clamp(16px, 2.6vh, 24px);
  box-shadow: 0 10px 40px rgba(0,0,0,0.08);
  overflow: hidden;
  padding-bottom: clamp(14px, 2.6vh, 24px);
  animation: fadeInUp 0.5s ease;
}

.pay-header {
  text-align: center;
  padding: clamp(20px, 4vh, 34px) 20px clamp(14px, 2.8vh, 24px);
  background: linear-gradient(180deg, rgba(255, 223, 93, 0.1) 0%, #fff 100%);

  .brand {
    display: flex; align-items: center; justify-content: center; gap: 8px; margin-bottom: clamp(8px, 1.6vh, 14px);
    img { height: 28px; object-fit: contain; }
    .sandbox-tag { background: $dark; color: $primary; font-size: 10px; padding: 2px 8px; border-radius: 4px; font-weight: bold; }
  }

  .order-amount {
    color: $dark; margin-bottom: clamp(8px, 1.6vh, 12px);
    .currency { font-size: clamp(18px, 3.2vh, 24px); font-weight: bold; vertical-align: top; margin-right: 4px; }
    .amount { font-size: clamp(34px, 6.2vh, 48px); font-weight: 900; letter-spacing: -1px; line-height: 1; }
  }

  .order-info {
    font-size: clamp(12px, 1.9vh, 14px); color: #666;
    p { margin: clamp(1px, 0.5vh, 4px) 0; line-height: 1.35; }
  }
}

.pay-methods {
  padding: clamp(8px, 1.8vh, 12px) clamp(18px, 5vw, 28px);
  display: flex; gap: clamp(10px, 2vw, 16px);

  .method-item {
    flex: 1;
    border: 2px solid #f5f5f5;
    background: #f9f9f9;
    border-radius: 12px;
    padding: clamp(10px, 1.9vh, 16px);
    display: flex; align-items: center; justify-content: center; gap: 8px;
    cursor: pointer;
    transition: all 0.2s;
    position: relative;
    opacity: 0.8;

    span { font-size: 15px; font-weight: bold; color: #333; }

    .check-icon {
      position: absolute; top: -8px; right: -8px;
      background: $dark; color: $primary; border-radius: 50%; padding: 2px;
      font-size: 12px; display: none;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }

    &.active {
      border-color: $dark; background: #fff; opacity: 1; transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0,0,0,0.05);
      .check-icon { display: block; }
    }
  }
}

.qr-section {
  text-align: center;
  padding: clamp(10px, 2.2vh, 18px) 0 clamp(14px, 2.6vh, 24px);

  .qr-box {
    width: clamp(132px, 24vh, 180px);
    height: clamp(132px, 24vh, 180px);
    margin: 0 auto clamp(10px, 1.8vh, 20px);
    position: relative;
    padding: clamp(8px, 1.2vh, 10px);
    border-radius: 16px;
    background: #fff;
    box-shadow: 0 8px 24px rgba(0,0,0,0.06);
    border: 1px solid #f0f0f0;

    img { width: 100%; height: 100%; display: block; border-radius: 8px; }

    .logo-center {
      position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);
      width: clamp(30px, 5vh, 40px);
      height: clamp(30px, 5vh, 40px);
      background: #fff; border-radius: 8px; padding: 4px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    }
  }

  .scan-tip { font-size: clamp(13px, 2.1vh, 15px); font-weight: bold; color: #333; margin-bottom: clamp(3px, 0.8vh, 6px); }
  .scan-subtip { font-size: clamp(11px, 1.6vh, 12px); color: #999; margin: 0; }
}

.action-area {
  padding: 0 clamp(18px, 5vw, 28px);
  .btn-confirm {
    width: 100%; height: clamp(42px, 6.2vh, 50px); border: none; border-radius: 99px; font-size: clamp(14px, 2.2vh, 16px); font-weight: 800; cursor: pointer; transition: 0.2s;
    background: $dark; color: $primary;
    box-shadow: 0 8px 20px rgba(0,0,0,0.15);
    &:hover { transform: translateY(-2px); box-shadow: 0 12px 24px rgba(0,0,0,0.25); background: #000; }
    &:active { transform: scale(0.98); }
    &:disabled { cursor: not-allowed; opacity: 0.75; transform: none; box-shadow: none; }
  }
  .cancel-link { text-align: center; margin-top: clamp(10px, 1.8vh, 16px); font-size: clamp(12px, 1.8vh, 13px); color: #999; cursor: pointer; font-weight: 500; &:hover { color: #333; } }
}

@media (max-height: 680px) {
  .pay-methods {
    .method-item {
      span { font-size: 14px; }
    }
  }
}

@keyframes fadeInUp { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
</style>
