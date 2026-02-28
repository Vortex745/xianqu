<template>
  <el-dialog
      v-model="visible"
      title="闲趣收银台"
      width="460px"
      align-center
      class="payment-dialog"
      destroy-on-close
      :close-on-click-modal="false"
      :before-close="handleClose"
  >
    <div class="pay-content">
      <div v-if="step === 1" class="step-1">
        <div class="price-header">
          <span class="label">支付金额</span>
          <div class="amount">
            <span class="symbol">¥</span>{{ formatPrice(order.amount || order.price) }}
          </div>
        </div>

        <div class="methods">
          <div
              class="method-item"
              :class="{ active: paymentMethod === 'alipay' }"
              @click="paymentMethod = 'alipay'"
          >
            <div class="icon ali"><el-icon><Wallet /></el-icon></div>
            <span class="name">支付宝</span>
            <el-icon v-if="paymentMethod === 'alipay'" class="check"><Select /></el-icon>
          </div>

          <div
              class="method-item"
              :class="{ active: paymentMethod === 'wechat' }"
              @click="paymentMethod = 'wechat'"
          >
            <div class="icon wx"><el-icon><ChatDotRound /></el-icon></div>
            <span class="name">微信支付</span>
            <el-icon v-if="paymentMethod === 'wechat'" class="check"><Select /></el-icon>
          </div>
        </div>

        <div class="footer-btns">
          <button class="btn-cancel" @click="handleClose">稍后支付</button>
          <button class="btn-pay" @click="toScan">立即支付</button>
        </div>
      </div>

      <div v-else class="step-2">
        <div class="qr-title">
          <span v-if="paymentMethod === 'alipay'" style="color:#1677ff">请使用支付宝扫码</span>
          <span v-else style="color:#07c160">请使用微信扫码</span>
        </div>

        <div class="qr-box">
          <img :src="`https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=Pay:${order.order_no || 'Unknown'}`" class="qr-img" />

          <div v-if="paying" class="mask">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>支付处理中...</span>
          </div>
        </div>

        <div class="tips">模拟环境：点击下方按钮直接完成支付</div>

        <div class="actions">
          <el-button round @click="step = 1">返回选择</el-button>
          <el-button type="success" round :loading="paying" @click="confirmPay" class="btn-confirm-pay">
            模拟支付成功
          </el-button>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import { Select, Loading, Wallet, ChatDotRound } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: Boolean,
  order: { type: Object, default: () => ({}) }
})
const emit = defineEmits(['update:modelValue', 'success'])
const router = useRouter()

const visible = ref(false)
const step = ref(1)
const paymentMethod = ref('alipay')
const paying = ref(false)

// 监听弹窗打开，重置状态
watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    step.value = 1
    paying.value = false
  }
})

const formatPrice = (val) => {
  return Number(val || 0).toFixed(2)
}

// 进入扫码页
const toScan = () => {
  step.value = 2
}

// ★★★ 核心修复：关闭/取消时跳转到订单页 ★★★
// 因为订单已经生成了，不能让用户觉得“没买成”，引导去订单页继续支付
const handleClose = () => {
  visible.value = false
  emit('update:modelValue', false)

  // 如果是在支付步骤关闭，提示用户
  ElMessage.info('订单已生成，请在“我的订单”中完成支付')
  router.push('/orders')
}

// ★★★ 核心修复：确认支付（支持批量） ★★★
// 在 PaymentModal.vue 的 <script setup> 中

const confirmPay = async () => {
  paying.value = true
  try {
    // 1. 批量支付逻辑
    if (props.order.isBatch && props.order.ids && props.order.ids.length > 0) {
      const promises = props.order.ids.map(id => request.post(`/api/orders/${id}/pay`))
      await Promise.all(promises)
    }
    // 2. 单个支付逻辑 (修复点：确保这里读取的是 props.order.id)
    else if (props.order.id) {
      // ★★★ 核心修复：调用确认支付接口 ★★★
      await request.post(`/api/orders/${props.order.id}/confirm_pay`)
    } else {
      throw new Error('无效的订单信息') // 增加错误抛出，方便排查
    }

    ElMessage.success('支付成功！')

    visible.value = false
    emit('update:modelValue', false)
    emit('success') // 通知父组件刷新列表

  } catch (e) {
    console.error(e)
    // 显示后端返回的具体错误信息，方便调试
    ElMessage.error(e.response?.data?.error || '支付失败，请检查订单状态')
  } finally {
    paying.value = false
  }
}
</script>

<style scoped lang="scss">
$primary: #ffdf5d;

.payment-dialog { border-radius: 16px; overflow: hidden; }
.pay-content { padding: 10px 20px 30px; }

/* 头部金额 */
.price-header {
  text-align: center; margin-bottom: 30px;
  .label { color: #999; font-size: 14px; }
  .amount {
    font-size: 36px; font-weight: 900; color: #333; font-family: 'DIN Alternate', sans-serif;
    margin-top: 4px;
    .symbol { font-size: 20px; margin-right: 4px; }
  }
}

/* 支付方式列表 */
.methods {
  display: flex; flex-direction: column; gap: 12px; margin-bottom: 30px;
  .method-item {
    display: flex; align-items: center; padding: 16px; border: 1px solid #f0f0f0;
    border-radius: 12px; cursor: pointer; transition: 0.2s;
    &:hover { background: #f9f9f9; }
    &.active { border-color: $primary; background: #fffbf0; }

    .icon {
      width: 32px; height: 32px; border-radius: 8px;
      display: flex; align-items: center; justify-content: center;
      margin-right: 12px; color: #fff; font-size: 20px;
      &.ali { background: #1677ff; }
      &.wx { background: #07c160; }
    }

    .name { flex: 1; font-weight: bold; color: #333; font-size: 15px; }
    .check { color: #ff5000; font-weight: bold; font-size: 18px; }
  }
}

/* 底部按钮 */
.footer-btns {
  display: flex; gap: 12px;
  button {
    flex: 1; height: 46px; border-radius: 99px; font-weight: bold; cursor: pointer; font-size: 15px; transition: 0.2s; border: none;
  }
  .btn-cancel { background: #f5f5f5; color: #666; &:hover { background: #eee; } }
  .btn-pay { background: #333; color: $primary; &:hover { opacity: 0.9; } }
}

/* 扫码页 */
.step-2 {
  text-align: center;
  .qr-title { font-weight: bold; margin-bottom: 20px; font-size: 16px; }
  .qr-box {
    width: 200px; height: 200px; margin: 0 auto 20px; position: relative;
    border: 1px solid #eee; padding: 10px; border-radius: 12px; background: #fff;
    .qr-img { width: 100%; height: 100%; object-fit: contain; }
    .mask {
      position: absolute; inset: 0; background: rgba(255,255,255,0.95);
      display: flex; flex-direction: column; align-items: center; justify-content: center;
      gap: 10px; font-weight: bold; font-size: 14px; color: #333; border-radius: 12px;
      .is-loading { font-size: 28px; color: $primary; }
    }
  }
  .tips { font-size: 13px; color: #999; margin-bottom: 24px; }
  .actions {
    display: flex; justify-content: center; gap: 16px;
    .btn-confirm-pay { font-weight: bold; padding: 10px 24px; }
  }
}
</style>