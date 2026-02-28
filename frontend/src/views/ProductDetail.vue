<template>
  <div class="product-detail-page">
    <div class="nav-header animate-down">
      <div class="nav-inner">
        <div class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon> 返回
        </div>
        <div class="brand-logo">XIANQU</div>
        <div class="action-btn">
          <el-icon><MoreFilled /></el-icon>
        </div>
      </div>
    </div>

    <div class="main-viewport" v-loading="loading">
      <div class="glass-card animate-up" v-if="product.id">

        <div class="gallery-section">
          <div class="image-stage">
            <el-image
                ref="imageRef"
                :src="product.image || defaultImg"
                fit="contain"
                class="main-img"
                :class="{ 'is-sold': product.status !== 1 }"
                :preview-src-list="[product.image || defaultImg]"
                hide-on-click-modal
            >
              <template #error>
                <div class="image-slot"><el-icon><Picture /></el-icon></div>
              </template>
            </el-image>

            <div v-if="product.status !== 1" class="status-stamp">
              <span>{{ product.status === 2 ? '已卖出' : '已下架' }}</span>
            </div>
          </div>

          <div class="gallery-footer">
            <div class="zoom-tip" @click="showLargeImage" style="cursor: pointer;"><el-icon><ZoomIn /></el-icon> 点击查看大图</div>
          </div>
        </div>

        <div class="info-section">
          <div class="info-scroll-area">
            <div class="seller-bar" @click="goToUserPage">
              <el-avatar :size="44" :src="product.seller?.avatar || defaultAvatar" class="seller-avatar" />
              <div class="seller-info">
                <div class="name">{{ product.seller?.nickname || product.seller?.username || '神秘卖家' }}</div>
                <div class="tags">
                  <span class="tag-credit">信用极好</span>
                  <span class="tag-time">发布于 {{ formatDate(product.created_at) }}</span>
                </div>
              </div>
              <div class="view-homepage">
                主页 <el-icon><ArrowRight /></el-icon>
              </div>
            </div>

            <div class="product-header">
              <div class="price-line">
                <span class="symbol">¥</span>
                <span class="num">{{ product.price }}</span>
                <span class="origin" v-if="product.original_price">¥{{ product.original_price }}</span>
              </div>
              <h1 class="title">{{ product.name }}</h1>
              <div class="meta-row" v-if="product.area || selectedTradeOptions.length">
                <div class="location-chip" v-if="product.area">
                  <el-icon><Location /></el-icon> {{ product.area }}
                </div>
                <div class="trade-tags-wrap" v-if="selectedTradeOptions.length">
                  <span class="trade-tag" v-for="opt in selectedTradeOptions" :key="opt">
                    {{ opt }}
                  </span>
                </div>
              </div>
            </div>

            <div class="divider"></div>

            <div class="desc-box">
              <p>{{ product.description || '卖家很懒，没有填写描述...' }}</p>
            </div>

            <div class="stat-row">
              <div class="stat-item"><el-icon><View /></el-icon> {{ product.view_count || 0 }} 浏览</div>
              <div class="stat-item"><el-icon><Star /></el-icon> {{ collectCount }} 想要</div>
            </div>
          </div>

          <div class="action-dock">
            <div class="icon-btn" :class="{ active: isLiked }" @click="toggleLike">
              <div class="icon-wrapper">
                <el-icon v-if="isLiked"><StarFilled /></el-icon>
                <el-icon v-else><Star /></el-icon>
              </div>
              <span class="text">{{ isLiked ? '已收藏' : '想要' }}</span>
            </div>

            <div class="main-btns">
              <template v-if="isOwner">
                <button class="btn btn-cart" disabled style="background: #f5f5f5; color: #999; cursor: not-allowed;">我是卖家</button>
                <button class="btn btn-buy" disabled style="background: #eee; color: #999; box-shadow: none; cursor: not-allowed;">不可购买</button>
              </template>
              <template v-else>
                <button class="btn btn-cart" @click="addToCart">加入购物车</button>
                <button class="btn btn-buy" :disabled="buyLoading" @click="handleBuy">
                  {{ buyLoading ? '跳转中...' : '立即购买' }}
                </button>
              </template>
            </div>
          </div>
        </div>

      </div>

      <div v-else-if="!loading" class="empty-state">
        <el-empty description="宝贝不见了..." />
        <button class="btn-back" @click="goBack">返回首页</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft, Picture, ZoomIn, View, Star, StarFilled,
  ArrowRight, Location, MoreFilled
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const product = ref({})
const isLiked = ref(false)
const buyLoading = ref(false)
const collectCount = ref(0)
const user = JSON.parse(localStorage.getItem('user') || 'null')
const imageRef = ref(null)

const defaultImg = 'https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png'
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const isOwner = computed(() => {
  return user && product.value.user_id && String(user.id) === String(product.value.user_id)
})

const isOptionEnabled = (value) => {
  return value === true || value === 1 || value === '1' || value === 'true'
}

const selectedTradeOptions = computed(() => {
  const options = [
    { label: '可小刀', value: product.value.is_negotiable ?? product.value.isNegotiable },
    { label: '送货上门', value: product.value.is_home_delivery ?? product.value.isHomeDelivery },
    { label: '自提', value: product.value.is_self_pickup ?? product.value.isSelfPickup }
  ]

  return options
    .filter((item) => isOptionEnabled(item.value))
    .map((item) => item.label)
})

const goBack = () => {
  if (window.history.length > 1) {
    router.go(-1)
  } else {
    router.push('/')
  }
}

const formatDate = (str) => {
  if (!str) return ''
  const d = new Date(str)
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

const showLargeImage = () => {
  if (imageRef.value && imageRef.value.$el) {
    const imgDiv = imageRef.value.$el.querySelector('.el-image__inner')
    if (imgDiv) {
      imgDiv.click()
    }
  }
}

const goToUserPage = () => {
  if (product.value.user_id) {
    router.push({
      path: `/user/${product.value.user_id}`,
      query: {
        nickname: product.value.seller?.nickname || product.value.seller?.username,
        avatar: product.value.seller?.avatar
      }
    })
  }
}

const fetchDetail = async () => {
  loading.value = true
  try {
    const res = await request.get(`/api/products/${route.params.id}`)
    let data = res.data

    // 图片路径处理
    if (data.image && !data.image.startsWith('http')) {
      data.image = 'http://127.0.0.1:8081' + data.image
    }
    if (data.image) data.image = data.image.replace('localhost', '127.0.0.1')

    // 头像处理
    if (data.seller && data.seller.avatar && !data.seller.avatar.startsWith('http')) {
      data.seller.avatar = 'http://127.0.0.1:8081' + data.seller.avatar
    }

    data.is_negotiable = isOptionEnabled(data.is_negotiable ?? data.isNegotiable)
    data.is_home_delivery = isOptionEnabled(data.is_home_delivery ?? data.isHomeDelivery)
    data.is_self_pickup = isOptionEnabled(data.is_self_pickup ?? data.isSelfPickup)

    product.value = data

    // 读取真实收藏数
    collectCount.value = res.collect_count || 0

    if (user) {
      checkFavorite()
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('获取宝贝详情失败')
  } finally {
    loading.value = false
  }
}

const checkFavorite = async () => {
  try {
    const res = await request.get('/api/favorites/check', {
      params: { user_id: user.id, product_id: product.value.id }
    })
    isLiked.value = res.is_favorite
  } catch (e) {}
}

const toggleLike = async () => {
  if (!user) return ElMessage.warning('请先登录')
  if (isOwner.value) return ElMessage.warning('不能收藏自己的商品')
  try {
    if (isLiked.value) {
      await request.post('/api/favorites/remove', { user_id: user.id, product_id: product.value.id })
      isLiked.value = false
      collectCount.value = Math.max(0, collectCount.value - 1)
    } else {
      await request.post('/api/favorites/add', { user_id: user.id, product_id: product.value.id })
      isLiked.value = true
      collectCount.value += 1
      ElMessage.success('已加入心愿单')
    }
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const addToCart = async () => {
  if (!user) return ElMessage.warning('请先登录')
  if (isOwner.value) return ElMessage.warning('不能购买自己的商品')
  if (product.value.status !== 1) return ElMessage.warning('商品不可购买')
  try {
    await request.post('/api/cart/add', { user_id: user.id, product_id: product.value.id, count: 1 })
    ElMessage.success('已加入购物车')
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '添加失败')
  }
}

const jumpToLogin = () => {
  router.push({
    path: '/',
    query: {
      auth: 'login',
      redirect: route.fullPath
    }
  })
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

const handleBuy = async () => {
  if (buyLoading.value) return
  if (!user) {
    ElMessage.warning('请先登录后购买')
    jumpToLogin()
    return
  }
  if (isOwner.value) return ElMessage.warning('不能购买自己的商品')
  if (product.value.status !== 1) return ElMessage.warning('商品不可购买')
  if (Number(product.value.count || 0) < 1) return ElMessage.warning('库存不足，暂时无法购买')
  try {
    buyLoading.value = true
    const createRes = await request.post('/api/orders', { product_id: product.value.id })
    const createdOrder = createRes?.data
    if (!createdOrder?.id) throw new Error('订单创建异常')

    const payRes = await request.post(`/api/orders/${createdOrder.id}/pay`)
    const payPath = buildPayPath(payRes?.pay_url, {
      amount: createdOrder.price ?? product.value.price,
      product_id: product.value.id,
      product_name: product.value.name || '',
      price: product.value.price,
      quantity: 1,
      return_to: '/orders',
      source: 'product_detail',
      source_path: route.fullPath
    })
    if (/^https?:\/\//i.test(payPath)) {
      window.location.href = payPath
      return
    }
    router.push(payPath)
  } catch (e) {
    console.error(e)
    if (e.response && e.response.status === 401) {
      ElMessage.warning('登录已失效，请重新登录')
      jumpToLogin()
    } else {
      const msg = e.response?.data?.error || '下单失败，请稍后重试'
      ElMessage.error(msg)
      if (msg.includes('售出') || msg.includes('下架') || msg.includes('状态')) {
        fetchDetail()
      }
    }
  } finally {
    buyLoading.value = false
  }
}

onMounted(() => {
  window.scrollTo(0, 0)
  fetchDetail()
})
</script>

<style scoped lang="scss">
/* 核心配色 */
$primary: #ffdf5d;
$dark: #1a1a1a;
$bg-gradient: radial-gradient(circle at 10% 20%, rgba(255, 223, 93, 0.15) 0%, #f6f7f9 80%);

.product-detail-page {
  height: 100vh;
  width: 100vw;
  background: $bg-gradient;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  font-family: "PingFang SC", sans-serif;
}

/* 顶部导航 */
.nav-header {
  height: 60px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  padding: 0 20px;
  z-index: 10;
  .nav-inner {
    width: 100%; max-width: 1200px; margin: 0 auto;
    display: flex; justify-content: space-between; align-items: center;
  }
  .back-btn {
    display: flex; align-items: center; gap: 4px;
    font-size: 14px; font-weight: bold; color: #333;
    cursor: pointer; padding: 8px 12px; border-radius: 99px;
    background: rgba(255,255,255,0.6); backdrop-filter: blur(4px);
    transition: 0.2s;
    &:hover { background: #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
  }
  .brand-logo { font-weight: 900; font-size: 16px; letter-spacing: 2px; color: $dark; }
  .action-btn {
    width: 32px; height: 32px; border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; color: #666;
    &:hover { background: rgba(0,0,0,0.05); }
  }
}

/* 主视口 */
.main-viewport {
  flex: 1;
  display: flex; justify-content: center; align-items: center;
  padding: 14px 28px 26px;
  overflow: hidden;
}

/* 毛玻璃卡片 */
.glass-card {
  width: 100%; max-width: 1140px; height: 100%; max-height: 760px;
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(20px);
  border-radius: 28px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.08), inset 0 0 0 1px rgba(255,255,255,0.6);
  display: flex;
  overflow: hidden;
}

/* 左侧画廊 */
.gallery-section {
  flex: 1.1;
  background: #fcfcfc;
  display: flex; flex-direction: column;
  position: relative;
  border-right: 1px solid rgba(0,0,0,0.03);

  .image-stage {
    flex: 1; position: relative;
    display: flex; align-items: center; justify-content: center;
    padding: 30px;

    .main-img {
      width: 100%; height: 100%; object-fit: contain;
      filter: drop-shadow(0 10px 30px rgba(0,0,0,0.1));
      transition: transform 0.3s cubic-bezier(0.2, 0.8, 0.2, 1);
      cursor: zoom-in;
      &:hover { transform: scale(1.05); }
      &.is-sold { filter: grayscale(100%); opacity: 0.8; }
    }

    .status-stamp {
      position: absolute; inset: 0; pointer-events: none;
      display: flex; align-items: center; justify-content: center;
      span {
        font-size: 24px; font-weight: 900; color: #fff;
        background: $dark; padding: 8px 24px;
        transform: rotate(-10deg); border-radius: 12px;
        box-shadow: 0 10px 30px rgba(0,0,0,0.2);
      }
    }
  }

  .gallery-footer {
    height: 60px; display: flex; align-items: center; justify-content: center;
    .zoom-tip { font-size: 12px; color: #999; display: flex; align-items: center; gap: 4px; }
  }
}

/* 右侧信息 */
.info-section {
  flex: 0.9;
  display: flex; flex-direction: column;
  background: rgba(255,255,255,0.4);

  .info-scroll-area {
    flex: 1;
    overflow-y: auto;
    padding: 24px 32px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    &::-webkit-scrollbar { width: 4px; }
    &::-webkit-scrollbar-thumb { background: #eee; border-radius: 2px; }
  }
}

/* 卖家栏 */
.seller-bar {
  display: flex; align-items: center; gap: 12px;
  margin-bottom: 0;
  padding: 12px;
  background: #fff; border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.02);
  cursor: pointer; transition: 0.2s;
  &:hover { transform: translateY(-2px); box-shadow: 0 8px 16px rgba(0,0,0,0.05); }

  .seller-avatar { border: 2px solid #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.05); }
  .seller-info {
    flex: 1;
    .name { font-weight: bold; font-size: 15px; color: $dark; margin-bottom: 4px; }
    .tags { display: flex; gap: 8px; font-size: 11px; }
    .tag-credit { color: #00b578; background: #e6fdfb; padding: 1px 6px; border-radius: 4px; }
    .tag-time { color: #999; }
  }
  .view-homepage { font-size: 12px; font-weight: bold; color: #666; display: flex; align-items: center; }
}

/* 商品头 */
.product-header {
  margin-bottom: 0;
  padding: 4px 2px 0;
  .price-line {
    display: flex; align-items: baseline; gap: 2px;
    color: #ff5000; line-height: 1; margin-bottom: 10px;
    .symbol { font-size: 20px; font-weight: bold; }
    .num { font-size: 40px; font-weight: 900; letter-spacing: -1px; }
    .origin { font-size: 14px; color: #999; text-decoration: line-through; margin-left: 8px; font-weight: normal; }
  }
  .title { font-size: 24px; font-weight: 800; color: #222; line-height: 1.3; margin: 0; }
}

.meta-row {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;

  .location-chip {
    display: inline-flex; align-items: center; gap: 4px;
    background: #f0f0f0; color: #666; font-size: 12px; font-weight: 600;
    padding: 4px 10px; border-radius: 6px;
  }
}

.divider { height: 1px; background: rgba(0,0,0,0.06); margin: 0; }

.trade-tags-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.trade-tag {
  font-size: 12px;
  font-weight: 800;
  color: #111;
  background: rgba(255, 223, 93, 0.8);
  border: 1px solid rgba(0, 0, 0, 0.06);
  padding: 5px 10px;
  border-radius: 999px;
}

.desc-box {
  min-height: 104px;
  margin: 0;
  border: 1px solid rgba(0, 0, 0, 0.05);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.78);
  padding: 14px 16px;
  p { font-size: 15px; line-height: 1.7; color: #444; white-space: pre-wrap; }
}

.stat-row {
  display: flex;
  gap: 10px;
  color: #999;
  font-size: 13px;
  .stat-item {
    display: flex;
    align-items: center;
    gap: 4px;
    background: #f4f5f7;
    border-radius: 999px;
    padding: 5px 10px;
  }
}

/* 底部操作坞 */
.action-dock {
  padding: 14px 24px;
  background: #fff;
  border-top: 1px solid rgba(0,0,0,0.03);
  display: flex; align-items: center; gap: 20px;

  .icon-btn {
    display: flex; flex-direction: column; align-items: center; gap: 4px;
    cursor: pointer; color: #999; transition: 0.2s;
    .icon-wrapper { font-size: 22px; height: 24px; display: flex; align-items: center; }
    .text { font-size: 11px; font-weight: bold; }
    &:hover { color: $dark; }
    &.active { color: #ff5000; .icon-wrapper { animation: pop 0.3s; } }
  }

  .main-btns {
    flex: 1; display: flex; gap: 12px;
    .btn {
      flex: 1; height: 44px; border: none; border-radius: 22px;
      font-size: 15px; font-weight: 800; cursor: pointer; transition: 0.2s;
      &:active { transform: scale(0.98); }
    }
    .btn-cart { background: #fff4d6; color: #bfa12f; &:hover { background: #ffe08a; } }
    .btn-buy {
      background: $dark; color: $primary;
      box-shadow: 0 8px 20px rgba(26, 26, 26, 0.2);
      &:hover { transform: translateY(-2px); box-shadow: 0 12px 24px rgba(26, 26, 26, 0.3); }
      &:disabled {
        cursor: not-allowed;
        opacity: 0.75;
        transform: none;
        box-shadow: none;
      }
    }
  }
}

.empty-state { text-align: center; margin-top: 100px; .btn-back { background: $dark; color: #fff; border: none; padding: 10px 24px; border-radius: 99px; cursor: pointer; } }

/* 动画定义 */
.animate-up { animation: fadeInUp 0.6s cubic-bezier(0.2, 0.8, 0.2, 1); }
.animate-down { animation: fadeInDown 0.6s cubic-bezier(0.2, 0.8, 0.2, 1); }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(40px); } to { opacity: 1; transform: translateY(0); } }
@keyframes fadeInDown { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }
@keyframes pop { 50% { transform: scale(1.3); } }

/* 响应式 */
@media (max-width: 900px) {
  .product-detail-page { height: auto; overflow-y: auto; }
  .main-viewport { padding: 10px; height: auto; display: block; }
  .glass-card { flex-direction: column; max-height: none; border-radius: 20px; }
  .gallery-section { height: 400px; border-right: none; border-bottom: 1px solid #eee; }
  .info-scroll-area { max-height: none; }
  .action-dock { position: sticky; bottom: 0; box-shadow: 0 -4px 20px rgba(0,0,0,0.05); }
}
</style>
