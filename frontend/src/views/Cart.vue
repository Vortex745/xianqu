<template>
  <div class="cart-page">
    <div class="bg-orb orb-1"></div>
    <div class="bg-orb orb-2"></div>

    <div class="container">
      <div class="page-header glass-panel">
        <div class="left" @click="$router.push('/')">
          <el-icon><ArrowLeft /></el-icon> <span>返回首页</span>
        </div>
        <div class="center">购物车 <span class="count">{{ cartList.length }}件</span></div>
        <div class="right" @click="isEditMode = !isEditMode">{{ isEditMode ? '完成' : '管理' }}</div>
      </div>

      <div class="cart-layout" :class="{ 'has-items': cartList.length > 0 }" v-loading="loading">
        <div class="cart-main">
          <div v-if="cartList.length > 0" class="cart-list-shell glass-panel">
            <div class="list-title-row">
              <h3>商品列表</h3>
              <p class="select-hint">已选 <strong>{{ selectedIds.length }}</strong> 件</p>
            </div>
            <div class="cart-list">
              <transition-group name="list-anim">
                <div
                    v-for="item in cartList"
                    :key="item.id"
                    class="cart-item-card"
                    :class="{ 'is-selected': selectedIds.includes(item.id) }"
                    @click="toggleSelect(item.id)"
                >
                  <div class="checkbox-area" @click.stop="toggleSelect(item.id)">
                    <div class="custom-checkbox" :class="{ checked: selectedIds.includes(item.id) }">
                      <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
                    </div>
                  </div>

                  <div class="image-wrapper" @click.stop="$router.push(`/product/${item.product?.id}`)">
                    <el-image :src="item.product?.image || defaultImg" fit="cover" class="thumb" loading="lazy" />
                    <div v-if="item.product?.status !== 1" class="invalid-mask">失效</div>
                  </div>

                  <div class="info-wrapper">
                    <div class="title-row">
                      <h3 class="product-title">{{ item.product?.name || item.product?.title || '商品信息待更新' }}</h3>
                      <el-button
                          v-if="isEditMode"
                          type="danger"
                          circle
                          size="small"
                          class="delete-btn"
                          @click.stop="deleteItem(item.id)"
                      >
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                    <div class="meta-row">
                      <div class="seller-tag">
                        <el-avatar :size="16" :src="item.product?.seller?.avatar || defaultAvatar" />
                        <span>{{ getSellerName(item.product) }}</span>
                      </div>
                    </div>
                    <div class="price-row">
                      <div class="price"><span class="symbol">¥</span><span class="num">{{ item.product?.price }}</span></div>
                      <span class="qty-tag">x{{ item.count || 1 }}</span>
                    </div>
                  </div>
                </div>
              </transition-group>
            </div>
          </div>

          <div v-else-if="!loading" class="empty-cart glass-panel">
            <div class="empty-deco">
              <div class="deco-blob blob-1"></div>
              <div class="deco-blob blob-2"></div>
            </div>
            <div class="empty-content">
              <div class="empty-icon-wrap">
                <svg class="cart-svg" viewBox="0 0 80 80" fill="none">
                  <circle cx="40" cy="40" r="38" stroke="#1a1f29" stroke-width="2" stroke-dasharray="6 4" opacity="0.15"/>
                  <path d="M24 28h4l6 24h20l5-16H32" stroke="#1a1f29" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
                  <circle cx="36" cy="58" r="3" fill="#1a1f29"/>
                  <circle cx="52" cy="58" r="3" fill="#1a1f29"/>
                  <path d="M38 38h8M42 34v8" stroke="#ffdf5d" stroke-width="2.5" stroke-linecap="round"/>
                </svg>
              </div>
              <div class="empty-text">
                <h3>购物车暂无商品</h3>
                <p>浏览商品页面，将心仪商品加入购物车</p>
              </div>
              <button class="go-home-btn" @click="$router.push('/')">
                <span>浏览商品</span>
                <svg viewBox="0 0 20 20" fill="none" class="btn-arrow">
                  <path d="M4 10h12M12 6l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
            </div>
          </div>
        </div>

        <aside v-if="cartList.length > 0" class="desktop-summary glass-panel">
          <div class="summary-header">订单摘要</div>
          <div class="summary-row clickable" @click="toggleSelectAll">
            <div class="custom-checkbox" :class="{ checked: isAllSelected }">
              <el-icon v-if="isAllSelected"><Check /></el-icon>
            </div>
            <span>全选</span>
            <span class="value">{{ selectedIds.length }}/{{ cartList.length }}</span>
          </div>
          <div class="summary-row">
            <span>已选商品</span>
            <span class="value">{{ selectedIds.length }} 件</span>
          </div>
          <div class="summary-row total">
            <span>合计</span>
            <span class="amount">¥{{ totalPrice }}</span>
          </div>
          <button
              v-if="!isEditMode"
              class="summary-btn checkout"
              :disabled="selectedIds.length === 0"
              @click="handleCheckout"
          >
            立即结算
          </button>
          <button
              v-else
              class="summary-btn danger"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
          >
            删除所选
          </button>
        </aside>
      </div>
    </div>

    <div class="checkout-bar-wrapper mobile-only" v-if="cartList.length > 0">
      <div class="checkout-bar glass-panel">
        <div class="select-all" @click="toggleSelectAll">
          <div class="custom-checkbox" :class="{ checked: isAllSelected }">
            <el-icon v-if="isAllSelected"><Check /></el-icon>
          </div>
          <span>全选</span>
        </div>

        <div class="right-section" v-if="!isEditMode">
          <div class="total-section">
            <span class="label">合计:</span>
            <span class="total-price"><span class="symbol">¥</span>{{ totalPrice }}</span>
          </div>
          <button
              class="btn-checkout"
              :disabled="selectedIds.length === 0"
              @click="handleCheckout"
          >
            立即结算 ({{ selectedIds.length }})
          </button>
        </div>

        <div class="right-section" v-else>
          <button
              class="btn-delete-all"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
          >
            删除 ({{ selectedIds.length }})
          </button>
        </div>
      </div>
    </div>

    <PaymentModal
        v-model="payVisible"
        :order="payOrderInfo"
        @success="onPaySuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import request, { resolveUrl } from '@/utils/request'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Delete, Check, ShoppingCart, Goods } from '@element-plus/icons-vue'
import PaymentModal from '../components/PaymentModal.vue'

const router = useRouter()
const loading = ref(false)
const cartList = ref([])
const selectedIds = ref([])
const isEditMode = ref(false)
const user = ref(JSON.parse(localStorage.getItem('user') || '{}'))
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const defaultImg = 'https://via.placeholder.com/150?text=No+Image'

const payVisible = ref(false)
const payOrderInfo = ref({})
const sellerCache = new Map()

const normalizeAssetUrl = (url) => resolveUrl(url)

// 获取卖家名字
const getSellerName = (p) => {
  if (!p) return '未知卖家'
  return p.seller?.nickname || p.seller?.username || p.user?.nickname || p.user?.username || '神秘卖家'
}

const ensureSellerInfo = async (product) => {
  if (!product) return
  // 与详情页数据结构对齐：优先使用 seller
  if (!product.seller && product.user) {
    product.seller = product.user
  }

  // 已有卖家昵称/用户名则认为数据完整
  if (product.seller?.nickname || product.seller?.username) {
    if (product.seller.avatar) {
      product.seller.avatar = normalizeAssetUrl(product.seller.avatar)
    }
    return
  }

  const sellerId = Number(product.user_id || product.seller?.id || 0)
  if (!sellerId) return

  if (sellerCache.has(sellerId)) {
    product.seller = sellerCache.get(sellerId)
    return
  }

  try {
    const res = await request.get(`/api/users/${sellerId}`)
    const seller = res?.data || {}
    if (seller.avatar) seller.avatar = normalizeAssetUrl(seller.avatar)
    sellerCache.set(sellerId, seller)
    product.seller = seller
  } catch (e) {
    // 静默兜底，界面继续使用默认文案
  }
}

const fetchCart = async () => {
  if (!user.value.id) return router.push('/')
  loading.value = true
  try {
    const res = await request.get('/api/cart')

    // 图片路径修复逻辑
    const list = (res.data || []).map(item => {
      if (item.product?.image) {
        item.product.image = normalizeAssetUrl(item.product.image)
      }
      return item
    })

    await Promise.all(list.map(item => ensureSellerInfo(item.product)))
    cartList.value = list
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

// ★★★ 修复点3：计算总价时乘以数量 ★★★
const totalPrice = computed(() => {
  let sum = 0
  cartList.value.forEach(item => {
    if (selectedIds.value.includes(item.id) && item.product?.status === 1) {
      sum += Number(item.product.price) * (item.count || 1)
    }
  })
  return sum.toFixed(2)
})

const isAllSelected = computed(() => {
  // 只统计有效商品
  const validItems = cartList.value.filter(i => isEditMode.value || i.product?.status === 1)
  if (validItems.length === 0) return false
  return validItems.every(i => selectedIds.value.includes(i.id))
})

const toggleSelect = (id) => {
  const item = cartList.value.find(i => i.id === id)
  // 非管理模式下，失效商品不能选
  if (!isEditMode.value && item?.product?.status !== 1) return

  const index = selectedIds.value.indexOf(id)
  index > -1 ? selectedIds.value.splice(index, 1) : selectedIds.value.push(id)
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = cartList.value
        .filter(i => isEditMode.value || i.product?.status === 1)
        .map(i => i.id)
  }
}

const deleteItem = (id) => {
  ElMessageBox.confirm('确定要把这件宝贝移出购物车吗？', '提示', {
    confirmButtonText: '移除', cancelButtonText: '取消', type: 'warning', center: true, customClass: 'warm-theme-box'
  }).then(async () => {
    await request.delete(`/api/cart/${id}`)
    ElMessage.success('已移除')
    cartList.value = cartList.value.filter(i => i.id !== id)
    selectedIds.value = selectedIds.value.filter(sid => sid !== id)
  }).catch(() => {})
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedIds.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedIds.value.length} 件商品吗？`, '提示', { type: 'warning' })
    for (const id of selectedIds.value) {
      await request.delete(`/api/cart/${id}`)
    }
    ElMessage.success('已删除')
    selectedIds.value = []
    fetchCart()
  } catch (e) {}
}

// ★★★ 核心修复：批量下单逻辑 ★★★
const handleCheckout = async () => {
  if (selectedIds.value.length === 0) return

  loading.value = true
  try {
    // 1. 调用后端批量下单接口
    const res = await request.post('/api/orders/batch', {
      cart_ids: selectedIds.value,
      address: "默认收货地址"
    })

    // 2. 获取生成的订单数据
    const orders = res.data || []
    if (!orders || orders.length === 0) throw new Error('下单异常，未返回订单信息')

    // 提取 ID 数组，供 PaymentModal 批量支付使用
    const orderIds = orders.map(o => o.id)
    const totalAmount = orders.reduce((sum, o) => sum + o.price, 0)

    // 3. 打开支付弹窗
    payOrderInfo.value = {
      ids: orderIds,      // 传入数组，支持批量支付
      amount: totalAmount,
      isBatch: true,
      order_no: orders[0].order_no // 取第一个订单号用于展示二维码
    }
    payVisible.value = true

  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.error || '结算失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 支付成功回调
const onPaySuccess = () => {
  selectedIds.value = []
  fetchCart() // 重新拉取购物车(已结算的会被后端移除)
  // 延迟跳转，提升体验
  setTimeout(() => {
    router.push('/orders')
  }, 500)
}

onMounted(fetchCart)
</script>

<style scoped lang="scss">
@use "@/assets/tokens.scss" as *;

.cart-page {
  min-height: 100vh;
  background: 
    url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E"),
    linear-gradient(165deg, #f8fafb 0%, #eef2f7 35%, #e8ecf3 100%);
  background-blend-mode: overlay, normal;
  padding-bottom: 72px;
  position: relative;
  overflow-x: hidden;
}
.bg-orb {
  position: fixed;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.28;
  z-index: 0;
  pointer-events: none;
  animation: float-drift 20s ease-in-out infinite;
  &.orb-1 { width: 420px; height: 420px; background: linear-gradient(135deg, $primary, #ffc940); top: -200px; left: -140px; }
  &.orb-2 { width: 380px; height: 380px; background: linear-gradient(225deg, #a8d4ff, #7eb8f0); bottom: -160px; right: -140px; animation-delay: -10s; }
}
@keyframes float-drift {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(15px, -20px) scale(1.05); }
  66% { transform: translate(-10px, 15px) scale(0.95); }
}
.container {
  width: 100%;
  max-width: 1320px;
  margin: 0 auto;
  padding: 30px 24px 0;
  position: relative;
  z-index: 1;
  box-sizing: border-box; 
}

.glass-panel {
  background: rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border: 1px solid rgba(255, 255, 255, 0.68);
  box-shadow: 0 14px 34px rgba(10, 14, 24, 0.05);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 28px;
  height: 76px;
  margin-bottom: 22px;
  border-radius: 20px;
  .left {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    font-size: 16px;
    color: #5f6877;
    transition: 0.2s;
    &:hover { color: #1d2430; }
  }
  .center {
    font-size: 34px;
    font-weight: 900;
    color: #202633;
    letter-spacing: 0.4px;
    .count { font-size: 22px; font-weight: 700; color: #8c95a4; margin-left: 3px; }
  }
  .right {
    font-size: 15px;
    font-weight: 800;
    cursor: pointer;
    color: #2b3241;
    padding: 8px 15px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.75);
    border: 1px solid rgba(30, 38, 52, 0.08);
    &:hover { color: #171a21; background: #fff; }
  }
}

.cart-layout {
  display: block;
  &.has-items {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 326px;
    gap: 22px;
    align-items: start;
  }
}
.cart-main {
  min-height: 520px;
}

.cart-list-shell {
  border-radius: 22px;
  padding: 14px;
}

.list-title-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding: 10px 12px 16px;
  border-bottom: 1px dashed #e6ebf2;
  h3 {
    margin: 0;
    font-size: 19px;
    color: #1f2633;
    font-weight: 900;
    letter-spacing: -0.3px;
  }
  .select-hint {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: #7a8495;
    strong {
      color: #ff5a24;
      font-weight: 800;
      font-size: 15px;
    }
  }
}
.cart-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 10px;
}

.custom-checkbox {
  width: 23px;
  height: 23px;
  border-radius: 50%;
  border: 2px solid #c3cad8;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  transition: 0.2s;
  &.checked { background: #141923; border-color: #141923; }
}

.cart-item-card {
  display: grid;
  grid-template-columns: auto 132px minmax(0, 1fr);
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 15px;
  transition: all 0.28s cubic-bezier(0.22, 1, 0.36, 1);
  cursor: pointer;
  border: 1px solid #e8edf5;
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 12px 26px rgba(8, 13, 24, 0.08);
    border-color: #dde4ef;
  }
  &.is-selected { border-color: #eecf54; background: #fffdf2; }
  .checkbox-area { padding: 0 2px; }
  .image-wrapper {
    width: 132px;
    height: 132px;
    border-radius: 12px;
    overflow: hidden;
    background: #f3f5f9;
    position: relative;
    .thumb { width: 100%; height: 100%; object-fit: cover; }
    .invalid-mask {
      position: absolute;
      inset: 0;
      background: rgba(11, 14, 20, 0.64);
      color: #fff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 12px;
      font-weight: 700;
    }
  }
  .info-wrapper {
    min-height: 132px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    .title-row {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 14px;
      .product-title {
        margin: 0;
        font-size: 18px;
        font-weight: 800;
        color: #2a3140;
        line-height: 1.38;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
      }
      .delete-btn {
        color: #ff4d4f;
        background: #fff1f1;
        border: none;
        &:hover { background: #ff4d4f; color: #fff; }
      }
    }
    .meta-row {
      .seller-tag {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        background: #f2f5fa;
        padding: 4px 10px;
        border-radius: 999px;
        font-size: 12px;
        color: #677184;
        font-weight: 700;
      }
    }
    .price-row {
      display: flex;
      justify-content: space-between;
      align-items: flex-end;
      .price {
        color: #ff5a24;
        font-weight: 900;
        .symbol { font-size: 14px; margin-right: 2px; }
        .num { font-size: 30px; letter-spacing: 0.2px; }
      }
      .qty-tag {
        color: #677285;
        font-size: 12px;
        background: #f2f5fa;
        padding: 3px 8px;
        border-radius: 999px;
        font-weight: 700;
      }
    }
  }
}

.empty-cart {
  position: relative;
  padding: 100px 20px 120px;
  border-radius: 28px;
  min-height: 520px;
  overflow: hidden;
  max-width: 860px;
  margin: 0 auto;
  
  .empty-deco {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
    .deco-blob {
      position: absolute;
      border-radius: 50%;
      filter: blur(60px);
      opacity: 0.35;
    }
    .blob-1 {
      width: 260px;
      height: 260px;
      background: linear-gradient(135deg, #ffdf5d, #ffe98a);
      top: -80px;
      right: 15%;
    }
    .blob-2 {
      width: 200px;
      height: 200px;
      background: linear-gradient(225deg, #b8e0ff, #8fc9f5);
      bottom: -60px;
      left: 20%;
    }
  }
  
  .empty-content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    max-width: 400px;
    margin: 0 auto;
  }
  
  .empty-icon-wrap {
    width: 120px;
    height: 120px;
    border-radius: 32px;
    background:
      radial-gradient(circle at 30% 25%, rgba(255, 223, 93, 0.45), transparent 60%),
      linear-gradient(155deg, #fefefe, #f0f4f9);
    border: 1px solid rgba(227, 232, 240, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 
      inset 0 1px 0 rgba(255, 255, 255, 0.9),
      0 20px 40px rgba(13, 20, 34, 0.1),
      0 8px 16px rgba(13, 20, 34, 0.06);
    margin-bottom: 28px;
    transform: rotate(-3deg);
    
    .cart-svg {
      width: 64px;
      height: 64px;
    }
  }
  
  .empty-text {
    text-align: center;
    h3 {
      margin: 0 0 10px;
      font-size: 28px;
      color: #1f2836;
      font-weight: 900;
      letter-spacing: -0.5px;
    }
    p {
      color: #6b7a8f;
      margin: 0;
      font-size: 15px;
      font-weight: 600;
      line-height: 1.5;
    }
  }
  
  .go-home-btn {
    margin-top: 32px;
    height: 52px;
    padding: 0 28px;
    display: inline-flex;
    align-items: center;
    gap: 10px;
    font-weight: 800;
    font-size: 15px;
    border: none;
    border-radius: 16px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    background: $ink;
    color: $primary;
    box-shadow: 
      0 10px 24px rgba(11, 16, 24, 0.25),
      0 4px 8px rgba(11, 16, 24, 0.15);
    
    .btn-arrow {
      width: 18px;
      height: 18px;
      transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    }
    
    &:hover {
      background: #0f131b;
      transform: translateY(-3px);
      box-shadow: 
        0 14px 32px rgba(11, 16, 24, 0.28),
        0 6px 12px rgba(11, 16, 24, 0.18);
      .btn-arrow { transform: translateX(4px); }
    }
    &:active {
      transform: translateY(-1px);
    }
  }
}

.desktop-summary {
  border-radius: 22px;
  padding: 20px 20px 18px;
  position: sticky;
  top: 98px;
  background: 
    linear-gradient(145deg, rgba(255, 255, 255, 0.92), rgba(255, 255, 255, 0.85)),
    url("data:image/svg+xml,%3Csvg viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='3'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.03'/%3E%3C/svg%3E");
  border: 1px solid rgba(255, 255, 255, 0.7);
  
  .summary-header {
    font-size: 20px;
    color: #1f2836;
    font-weight: 900;
    margin-bottom: 16px;
    letter-spacing: -0.3px;
  }
  .summary-row {
    display: flex;
    align-items: center;
    gap: 10px;
    justify-content: space-between;
    border-bottom: 1px dashed #e7ebf2;
    padding: 14px 0;
    color: #5f697a;
    font-size: 14px;
    font-weight: 700;
    transition: all 0.2s ease;
    
    &.clickable {
      cursor: pointer;
      border-radius: 10px;
      margin: 0 -8px;
      padding: 14px 8px;
      &:hover {
        background: rgba(255, 223, 93, 0.12);
      }
    }
    .value { color: #2d3646; font-weight: 800; margin-left: auto; }
    &.total {
      border-bottom: none;
      margin-top: 6px;
      padding-top: 18px;
      border-top: 2px solid #eef2f7;
      .amount {
        color: #ff5a24;
        font-size: 32px;
        font-weight: 900;
        letter-spacing: -0.5px;
      }
    }
  }
  .summary-btn {
    margin-top: 16px;
    width: 100%;
    height: 50px;
    border-radius: 14px;
    border: none;
    font-size: 16px;
    font-weight: 900;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
    &.checkout {
      background: $ink;
      color: $primary;
      box-shadow: 0 8px 20px rgba(9, 12, 19, 0.18);
      &:hover { 
        transform: translateY(-2px); 
        box-shadow: 0 12px 28px rgba(9, 12, 19, 0.24); 
      }
    }
    &.danger {
      background: linear-gradient(135deg, #fff1f1, #ffe8e8);
      color: #e54143;
      border: 1px solid rgba(229, 65, 67, 0.15);
      &:hover { 
        background: linear-gradient(135deg, #ffe8e8, #ffd9d9);
        transform: translateY(-2px);
      }
    }
    &:disabled {
      opacity: 0.5;
      cursor: not-allowed;
      transform: none !important;
      box-shadow: none !important;
    }
  }
}

.checkout-bar-wrapper {
  position: fixed;
  bottom: 18px;
  left: 0;
  right: 0;
  z-index: 99;
  display: flex;
  justify-content: center;
  pointer-events: none;
}
.checkout-bar {
  pointer-events: auto;
  width: 100%;
  max-width: 980px;
  margin: 0 16px;
  height: 66px;
  border-radius: 99px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px 0 18px;
  .select-all {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 700;
    color: #3d4555;
  }
  .right-section {
    display: flex;
    align-items: center;
    gap: 14px;
    .total-section {
      display: flex;
      align-items: baseline;
      gap: 6px;
      .label { font-size: 14px; color: #646d7d; font-weight: 700; }
      .total-price {
        color: #ff5a24;
        font-weight: 900;
        font-size: 22px;
        .symbol { font-size: 15px; margin-right: 2px; }
      }
    }
    .btn-checkout {
      height: 46px;
      padding: 0 30px;
      background: $ink;
      color: $primary;
      border: none;
      border-radius: 999px;
      font-size: 15px;
      font-weight: 900;
      cursor: pointer;
      transition: 0.24s;
      &:hover { transform: translateY(-2px); background: #0f1219; }
      &:disabled { background: #d3d7df; color: #fff; cursor: not-allowed; transform: none; }
    }
    .btn-delete-all {
      height: 42px;
      padding: 0 20px;
      background: #ff4d4f;
      color: #fff;
      border-radius: 999px;
      border: none;
      font-size: 14px;
      font-weight: 800;
      cursor: pointer;
      &:disabled { opacity: 0.55; cursor: not-allowed; }
    }
  }
}

.mobile-only { display: none; }

.list-anim-enter-active, .list-anim-leave-active { transition: all 0.32s ease; }
.list-anim-enter-from { opacity: 0; transform: translateY(14px); }
.list-anim-leave-to { opacity: 0; transform: translateX(-60px); }

@media (max-width: 1160px) {
  .container { max-width: 1020px; padding: 20px 16px 0; }
  .page-header .center { font-size: 26px; .count { font-size: 18px; } }
  .cart-layout {
    grid-template-columns: 1fr;
  }
  .desktop-summary { display: none; }
  .mobile-only { display: flex; }
  .empty-cart {
    min-height: 420px;
    padding: 64px 16px;
    h3 { font-size: 24px; }
  }
}

@media (max-width: 720px) {
  .container { padding: 14px 12px 0; }
  .page-header {
    height: 58px;
    padding: 0 14px;
    border-radius: 14px;
    .center { font-size: 20px; .count { font-size: 15px; } }
  }
  .cart-item-card {
    grid-template-columns: auto 92px minmax(0, 1fr);
    padding: 12px;
    gap: 12px;
    .image-wrapper { width: 92px; height: 92px; border-radius: 10px; }
    .info-wrapper {
      min-height: 92px;
      .title-row .product-title { font-size: 14px; }
      .price-row .price .num { font-size: 20px; }
    }
  }
  .cart-list-shell {
    padding: 10px;
    border-radius: 16px;
  }
  .empty-cart {
    min-height: 320px;
    padding: 56px 16px;
    h3 { font-size: 22px; }
  }
  .mobile-only { display: flex; }
}
</style>
