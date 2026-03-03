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
            <button class="stage-image-button" type="button" @click="openViewer(activeGalleryIndex)">
              <img
                  v-if="!stageImageFailed"
                  :src="activeGalleryImage"
                  :alt="product.name || '商品主图'"
                  class="main-img"
                  :class="{ 'is-sold': product.status !== 1 }"
                  @error="handleStageImageError"
              />
              <div v-else class="image-slot">
                <el-icon><Picture /></el-icon>
              </div>
            </button>

            <div v-if="product.status !== 1" class="status-stamp">
              <span>{{ product.status === 2 ? '已卖出' : '已下架' }}</span>
            </div>
          </div>

          <div class="gallery-footer">
            <div class="thumb-strip" v-if="galleryImages.length">
              <button
                  v-for="(image, index) in galleryImages"
                  :key="`${image}-${index}`"
                  class="thumb-btn"
                  :class="{ active: activeGalleryIndex === index }"
                  type="button"
                  @click="selectGalleryImage(index)"
              >
                <img :src="image" :alt="`${product.name || '商品'}图片 ${index + 1}`" />
              </button>
            </div>

            <div class="zoom-tip" @click="openViewer(activeGalleryIndex)" style="cursor: pointer;">
              <el-icon><ZoomIn /></el-icon>
              查看大图 · {{ activeGalleryIndex + 1 }}/{{ galleryImages.length }}
            </div>
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

    <transition name="viewer-fade">
      <div
          v-if="viewerVisible"
          class="viewer-mask"
          :style="viewerMaskStyle"
          @click.self="closeViewer"
      >
        <section class="viewer-shell" :style="viewerSheetStyle">
          <div class="viewer-glow"></div>
          <div class="viewer-topbar">
            <div class="viewer-meta">
              <div class="viewer-index">{{ viewerIndexLabel }}</div>
              <div class="viewer-caption">{{ product.name || '宝贝大图' }}</div>
            </div>
            <div class="viewer-tools">
              <button class="viewer-tool" type="button" @click="zoomViewer(-0.25)">
                <span class="viewer-tool-label">-</span>
              </button>
              <button class="viewer-tool" type="button" @click="zoomViewer(0.25)">
                <span class="viewer-tool-label">+</span>
              </button>
              <button class="viewer-tool close" type="button" @click="closeViewer">
                <el-icon><Close /></el-icon>
              </button>
            </div>
          </div>

          <div
              class="viewer-stage"
              @wheel.prevent="handleWheelZoom"
              @touchstart="handleViewerTouchStart"
              @touchmove="handleViewerTouchMove"
              @touchend="handleViewerTouchEnd"
              @touchcancel="handleViewerTouchEnd"
          >
            <button v-if="canSwitchImage" class="viewer-nav prev" type="button" @click="changeViewerImage(-1)">
              <el-icon><ArrowLeft /></el-icon>
            </button>

            <div class="viewer-image-wrap">
              <div v-if="viewerImageLoading" class="viewer-loading">
                <span class="viewer-loading-orb"></span>
                <span>大图加载中</span>
              </div>
              <div v-else-if="viewerImageError" class="viewer-error">
                <el-icon><Picture /></el-icon>
                <span>这张图暂时没刷出来</span>
              </div>
              <img
                  v-show="!viewerImageLoading && !viewerImageError"
                  :src="viewerImage"
                  :alt="`${product.name || '商品'}大图 ${viewerIndex + 1}`"
                  class="viewer-image"
                  :style="viewerImageStyle"
                  draggable="false"
                  @load="handleViewerImageLoad"
                  @error="handleViewerImageError"
              />
            </div>

            <button v-if="canSwitchImage" class="viewer-nav next" type="button" @click="changeViewerImage(1)">
              <el-icon><ArrowRight /></el-icon>
            </button>
          </div>

          <div class="viewer-hint">双指缩放 · 左右切图 · 下滑收起</div>

          <div class="viewer-filmstrip" v-if="galleryImages.length > 1">
            <button
                v-for="(image, index) in galleryImages"
                :key="`viewer-${image}-${index}`"
                class="viewer-thumb"
                :class="{ active: viewerIndex === index }"
                type="button"
                @click="setViewerIndex(index)"
            >
              <img :src="image" :alt="`${product.name || '商品'}缩略图 ${index + 1}`" />
            </button>
          </div>
        </section>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request, { resolveBackendAssetUrl } from '@/utils/request'
import { ElMessage } from 'element-plus'
import {
  ArrowLeft, Picture, ZoomIn, View, Star, StarFilled,
  ArrowRight, Location, MoreFilled, Close
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const loading = ref(true)
const product = ref({})
const isLiked = ref(false)
const buyLoading = ref(false)
const collectCount = ref(0)
const user = JSON.parse(localStorage.getItem('user') || 'null')
const activeGalleryIndex = ref(0)
const stageImageFailed = ref(false)
const viewerVisible = ref(false)
const viewerIndex = ref(0)
const viewerScale = ref(1)
const viewerOffsetX = ref(0)
const viewerOffsetY = ref(0)
const viewerSwipeOffsetX = ref(0)
const viewerDismissOffsetY = ref(0)
const viewerImageLoading = ref(false)
const viewerImageError = ref(false)

const defaultImg = 'https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png'
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const touchState = {
  mode: 'idle',
  startX: 0,
  startY: 0,
  lastX: 0,
  lastY: 0,
  startScale: 1,
  startOffsetX: 0,
  startOffsetY: 0,
  pinchDistance: 0
}

const isOwner = computed(() => {
  return user && product.value.user_id && String(user.id) === String(product.value.user_id)
})

const clamp = (value, min, max) => Math.min(Math.max(value, min), max)

const normalizeAsset = (value) => resolveBackendAssetUrl(value) || ''

const extractImageCandidates = (source) => {
  const list = []

  const pushValue = (value) => {
    if (!value) return
    if (Array.isArray(value)) {
      value.forEach(pushValue)
      return
    }
    if (typeof value === 'object') {
      pushValue(value.image || value.url || value.src)
      return
    }

    const text = String(value).trim()
    if (!text) return

    if ((text.startsWith('[') && text.endsWith(']')) || (text.startsWith('{') && text.endsWith('}'))) {
      try {
        pushValue(JSON.parse(text))
        return
      } catch (error) {}
    }

    if (!/^https?:\/\//i.test(text) && !text.startsWith('data:') && !text.startsWith('blob:') && /[\n,]/.test(text)) {
      text.split(/[\n,]+/).forEach(pushValue)
      return
    }

    list.push(text)
  }

  pushValue(source)
  return list
}

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

const galleryImages = computed(() => {
  const candidates = [
    ...extractImageCandidates(product.value.images),
    ...extractImageCandidates(product.value.image_list),
    ...extractImageCandidates(product.value.imageList),
    ...extractImageCandidates(product.value.gallery),
    ...extractImageCandidates(product.value.image)
  ]

  const normalized = candidates
    .map(normalizeAsset)
    .filter(Boolean)

  const uniqueImages = [...new Set(normalized)]
  return uniqueImages.length ? uniqueImages : [defaultImg]
})

const activeGalleryImage = computed(() => {
  return galleryImages.value[activeGalleryIndex.value] || galleryImages.value[0] || defaultImg
})

const viewerImage = computed(() => {
  return galleryImages.value[viewerIndex.value] || activeGalleryImage.value
})

const canSwitchImage = computed(() => galleryImages.value.length > 1)

const viewerIndexLabel = computed(() => `${viewerIndex.value + 1} / ${galleryImages.value.length}`)

const viewerMaskStyle = computed(() => ({
  background: `rgba(7, 9, 14, ${Math.max(0.28, 0.92 - viewerDismissOffsetY.value / 420)})`
}))

const viewerSheetStyle = computed(() => ({
  transform: `translate3d(0, ${viewerDismissOffsetY.value}px, 0)`
}))

const viewerImageStyle = computed(() => ({
  transform: `translate3d(${viewerOffsetX.value + viewerSwipeOffsetX.value}px, ${viewerOffsetY.value}px, 0) scale(${viewerScale.value})`
}))

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

const resetViewerMotion = () => {
  viewerScale.value = 1
  viewerOffsetX.value = 0
  viewerOffsetY.value = 0
  viewerSwipeOffsetX.value = 0
  viewerDismissOffsetY.value = 0
}

const primeViewerImage = () => {
  viewerImageLoading.value = true
  viewerImageError.value = false
}

const handleStageImageError = () => {
  stageImageFailed.value = true
}

const handleViewerImageLoad = () => {
  viewerImageLoading.value = false
  viewerImageError.value = false
}

const handleViewerImageError = () => {
  viewerImageLoading.value = false
  viewerImageError.value = true
}

const selectGalleryImage = (index) => {
  activeGalleryIndex.value = clamp(index, 0, galleryImages.value.length - 1)
  stageImageFailed.value = false
}

const setViewerIndex = (index) => {
  const nextIndex = clamp(index, 0, galleryImages.value.length - 1)
  viewerIndex.value = nextIndex
  activeGalleryIndex.value = nextIndex
  stageImageFailed.value = false
  resetViewerMotion()
  primeViewerImage()
}

const changeViewerImage = (delta) => {
  if (!canSwitchImage.value) return
  const total = galleryImages.value.length
  const nextIndex = (viewerIndex.value + delta + total) % total
  setViewerIndex(nextIndex)
}

const zoomViewer = (delta) => {
  const nextScale = clamp(Number((viewerScale.value + delta).toFixed(2)), 1, 4)
  viewerScale.value = nextScale
  if (nextScale === 1) {
    viewerOffsetX.value = 0
    viewerOffsetY.value = 0
  }
}

const openViewer = async (index = activeGalleryIndex.value) => {
  setViewerIndex(index)
  viewerVisible.value = true
  await nextTick()
}

const closeViewer = () => {
  viewerVisible.value = false
}

const handleViewerKeydown = (event) => {
  if (!viewerVisible.value) return
  if (event.key === 'Escape') {
    closeViewer()
    return
  }
  if (event.key === 'ArrowLeft') {
    changeViewerImage(-1)
    return
  }
  if (event.key === 'ArrowRight') {
    changeViewerImage(1)
  }
}

const getTouchDistance = (touchA, touchB) => {
  const deltaX = touchA.clientX - touchB.clientX
  const deltaY = touchA.clientY - touchB.clientY
  return Math.hypot(deltaX, deltaY)
}

const handleViewerTouchStart = (event) => {
  if (!viewerVisible.value || !event.touches.length) return

  if (event.touches.length === 2) {
    touchState.mode = 'pinch'
    touchState.startScale = viewerScale.value
    touchState.startOffsetX = viewerOffsetX.value
    touchState.startOffsetY = viewerOffsetY.value
    touchState.pinchDistance = getTouchDistance(event.touches[0], event.touches[1])
    viewerSwipeOffsetX.value = 0
    viewerDismissOffsetY.value = 0
    return
  }

  const touch = event.touches[0]
  touchState.mode = 'pending'
  touchState.startX = touch.clientX
  touchState.startY = touch.clientY
  touchState.lastX = touch.clientX
  touchState.lastY = touch.clientY
  touchState.startOffsetX = viewerOffsetX.value
  touchState.startOffsetY = viewerOffsetY.value
}

const handleViewerTouchMove = (event) => {
  if (!viewerVisible.value || !event.touches.length) return

  if (event.touches.length === 2) {
    if (touchState.mode !== 'pinch') {
      handleViewerTouchStart(event)
      return
    }
    const nextDistance = getTouchDistance(event.touches[0], event.touches[1])
    if (!touchState.pinchDistance) return
    const ratio = nextDistance / touchState.pinchDistance
    viewerScale.value = clamp(Number((touchState.startScale * ratio).toFixed(2)), 1, 4)
    return
  }

  const touch = event.touches[0]
  const deltaX = touch.clientX - touchState.startX
  const deltaY = touch.clientY - touchState.startY
  touchState.lastX = touch.clientX
  touchState.lastY = touch.clientY

  if (touchState.mode === 'pending') {
    if (Math.abs(deltaX) < 6 && Math.abs(deltaY) < 6) return
    if (viewerScale.value > 1.02) {
      touchState.mode = 'pan'
    } else if (Math.abs(deltaY) > Math.abs(deltaX) && deltaY > 0) {
      touchState.mode = 'dismiss'
    } else {
      touchState.mode = 'swipe'
    }
  }

  if (touchState.mode === 'pan') {
    event.preventDefault()
    viewerOffsetX.value = touchState.startOffsetX + deltaX
    viewerOffsetY.value = touchState.startOffsetY + deltaY
    return
  }

  if (touchState.mode === 'dismiss') {
    viewerDismissOffsetY.value = Math.max(0, deltaY)
    return
  }

  if (touchState.mode === 'swipe') {
    viewerSwipeOffsetX.value = deltaX
  }
}

const handleViewerTouchEnd = () => {
  if (!viewerVisible.value) return

  const deltaX = touchState.lastX - touchState.startX
  const deltaY = touchState.lastY - touchState.startY

  if (touchState.mode === 'dismiss') {
    if (deltaY > 120) {
      closeViewer()
    } else {
      viewerDismissOffsetY.value = 0
    }
  } else if (touchState.mode === 'swipe') {
    if (Math.abs(deltaX) > 70 && canSwitchImage.value) {
      changeViewerImage(deltaX < 0 ? 1 : -1)
    }
    viewerSwipeOffsetX.value = 0
  } else if (touchState.mode === 'pan' && viewerScale.value <= 1.02) {
    resetViewerMotion()
  } else if (touchState.mode === 'pinch' && viewerScale.value <= 1.02) {
    resetViewerMotion()
  }

  touchState.mode = 'idle'
}

const handleWheelZoom = (event) => {
  if (!viewerVisible.value) return
  zoomViewer(event.deltaY < 0 ? 0.2 : -0.2)
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
    const data = res.data || {}

    // 图片路径处理
    if (data.image) data.image = normalizeAsset(data.image)

    // 头像处理
    if (data.seller && data.seller.avatar) {
      data.seller.avatar = normalizeAsset(data.seller.avatar)
    }

    data.is_negotiable = isOptionEnabled(data.is_negotiable ?? data.isNegotiable)
    data.is_home_delivery = isOptionEnabled(data.is_home_delivery ?? data.isHomeDelivery)
    data.is_self_pickup = isOptionEnabled(data.is_self_pickup ?? data.isSelfPickup)

    product.value = data
    activeGalleryIndex.value = 0
    stageImageFailed.value = false
    viewerIndex.value = 0

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
  window.addEventListener('keydown', handleViewerKeydown)
})

watch(galleryImages, (images) => {
  if (!images.length) return
  stageImageFailed.value = false
  if (activeGalleryIndex.value > images.length - 1) {
    activeGalleryIndex.value = 0
  }
  if (viewerIndex.value > images.length - 1) {
    viewerIndex.value = 0
  }
})

watch(viewerVisible, (visible) => {
  if (typeof document === 'undefined') return
  document.body.style.overflow = visible ? 'hidden' : ''
  if (!visible) {
    resetViewerMotion()
  }
})

watch(() => route.params.id, () => {
  fetchDetail()
})

onBeforeUnmount(() => {
  if (typeof document !== 'undefined') {
    document.body.style.overflow = ''
  }
  window.removeEventListener('keydown', handleViewerKeydown)
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

    .stage-image-button {
      width: 100%;
      height: 100%;
      border: none;
      background: transparent;
      padding: 0;
      cursor: zoom-in;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 28px;
    }

    .main-img {
      width: 100%; height: 100%; object-fit: contain;
      filter: drop-shadow(0 10px 30px rgba(0,0,0,0.1));
      transition: transform 0.3s cubic-bezier(0.2, 0.8, 0.2, 1);
      cursor: zoom-in;
      &:hover { transform: scale(1.05); }
      &.is-sold { filter: grayscale(100%); opacity: 0.8; }
    }

    .image-slot {
      width: min(76%, 420px);
      aspect-ratio: 1 / 1;
      border-radius: 28px;
      border: 1px solid rgba(0, 0, 0, 0.06);
      background: linear-gradient(145deg, rgba(255, 255, 255, 0.95), rgba(245, 247, 250, 0.98));
      box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9), 0 16px 40px rgba(22, 30, 43, 0.08);
      display: flex;
      align-items: center;
      justify-content: center;
      color: #b3bac7;
      font-size: 46px;
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
    padding: 14px 20px 18px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    align-items: center;
    justify-content: center;
    min-height: 92px;

    .thumb-strip {
      width: 100%;
      display: flex;
      gap: 10px;
      overflow-x: auto;
      padding-bottom: 2px;
      justify-content: center;

      &::-webkit-scrollbar {
        height: 4px;
      }

      &::-webkit-scrollbar-thumb {
        background: rgba(0, 0, 0, 0.12);
        border-radius: 999px;
      }
    }

    .thumb-btn {
      width: 58px;
      height: 58px;
      border-radius: 14px;
      overflow: hidden;
      border: 2px solid transparent;
      padding: 0;
      background: #fff;
      box-shadow: 0 6px 16px rgba(0,0,0,0.05);
      cursor: pointer;
      transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), border-color 0.22s cubic-bezier(0.22, 1, 0.36, 1);

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
      }

      &:hover {
        transform: translateY(-2px);
      }

      &.active {
        border-color: rgba(255, 223, 93, 0.95);
      }
    }

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

.viewer-mask {
  position: fixed;
  inset: 0;
  z-index: 120;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px;
  background:
    radial-gradient(circle at top, rgba(255, 223, 93, 0.16), transparent 34%),
    rgba(6, 8, 12, 0.78);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  transition: background 0.24s cubic-bezier(0.22, 1, 0.36, 1);
}

.viewer-shell {
  width: min(1180px, 100%);
  height: min(90vh, 860px);
  border-radius: 34px;
  background:
    linear-gradient(180deg, rgba(13, 17, 24, 0.98), rgba(6, 8, 12, 0.98)),
    rgba(9, 12, 17, 0.98);
  border: 1px solid rgba(255,255,255,0.08);
  box-shadow: 0 32px 90px rgba(0,0,0,0.42);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1);
  position: relative;
}

.viewer-glow {
  position: absolute;
  inset: -12% auto auto -10%;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(255, 223, 93, 0.22), transparent 68%);
  pointer-events: none;
  filter: blur(8px);
}

.viewer-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 18px 22px 14px;
  color: rgba(255,255,255,0.88);
  position: relative;
  z-index: 1;
}

.viewer-meta {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.viewer-caption {
  min-width: 0;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: rgba(255, 255, 255, 0.78);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.viewer-index {
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 223, 93, 0.14);
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.4px;
  color: #ffdf5d;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.08);
}

.viewer-tools {
  display: flex;
  align-items: center;
  gap: 8px;
}

.viewer-tool {
  width: 40px;
  height: 40px;
  border-radius: 16px;
  border: 1px solid rgba(255,255,255,0.08);
  background: rgba(255,255,255,0.06);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), background 0.22s cubic-bezier(0.22, 1, 0.36, 1);

  &:hover {
    transform: translateY(-1px) scale(1.02);
    background: rgba(255,255,255,0.12);
  }

  &.close {
    background: rgba(255, 223, 93, 0.16);
    color: #ffdf5d;
    border-color: rgba(255, 223, 93, 0.22);
  }
}

.viewer-tool-label {
  font-size: 20px;
  font-weight: 700;
  line-height: 1;
}

.viewer-stage {
  flex: 1;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 18px;
  padding: 0 22px 18px;
  touch-action: none;
  position: relative;
  z-index: 1;
}

.viewer-image-wrap {
  width: 100%;
  height: 100%;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 30px;
  border: 1px solid rgba(255,255,255,0.08);
  background:
    linear-gradient(180deg, rgba(255,255,255,0.04), rgba(255,255,255,0.01)),
    radial-gradient(circle at top, rgba(255, 223, 93, 0.08), transparent 28%),
    rgba(255,255,255,0.02);
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.04);
  position: relative;
}

.viewer-loading,
.viewer-error {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: rgba(255,255,255,0.72);
  font-size: 14px;
  font-weight: 700;
}

.viewer-loading-orb {
  width: 54px;
  height: 54px;
  border-radius: 50%;
  border: 2px solid rgba(255,255,255,0.12);
  border-top-color: #ffdf5d;
  border-right-color: rgba(255, 223, 93, 0.48);
  box-shadow: 0 0 24px rgba(255, 223, 93, 0.16);
  animation: viewerSpin 0.9s linear infinite;
}

.viewer-error {
  .el-icon {
    font-size: 38px;
    color: rgba(255,255,255,0.46);
  }
}

.viewer-image {
  max-width: calc(100% - 48px);
  max-height: calc(100% - 48px);
  object-fit: contain;
  user-select: none;
  transform-origin: center center;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1);
}

.viewer-nav {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  border: 1px solid rgba(255,255,255,0.08);
  background: rgba(255,255,255,0.08);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), background 0.22s cubic-bezier(0.22, 1, 0.36, 1);

  &:hover {
    transform: scale(1.06);
    background: rgba(255,255,255,0.16);
  }
}

.viewer-hint {
  text-align: center;
  color: rgba(255,255,255,0.56);
  font-size: 12px;
  letter-spacing: 0.08em;
  padding: 0 18px 14px;
  position: relative;
  z-index: 1;
}

.viewer-filmstrip {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding: 0 22px 22px;
  justify-content: center;
  position: relative;
  z-index: 1;

  &::-webkit-scrollbar {
    height: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: rgba(255,255,255,0.18);
    border-radius: 999px;
  }
}

.viewer-thumb {
  width: 62px;
  height: 62px;
  padding: 0;
  border-radius: 18px;
  overflow: hidden;
  border: 2px solid transparent;
  background: rgba(255,255,255,0.08);
  cursor: pointer;
  transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), border-color 0.22s cubic-bezier(0.22, 1, 0.36, 1);

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  &:hover {
    transform: translateY(-2px);
  }

  &.active {
    border-color: rgba(255, 223, 93, 0.9);
    box-shadow: 0 10px 24px rgba(255, 223, 93, 0.12);
  }
}

.viewer-fade-enter-active,
.viewer-fade-leave-active {
  transition: opacity 0.22s cubic-bezier(0.22, 1, 0.36, 1);
}

.viewer-fade-enter-from,
.viewer-fade-leave-to {
  opacity: 0;
}

/* 动画定义 */
.animate-up { animation: fadeInUp 0.6s cubic-bezier(0.2, 0.8, 0.2, 1); }
.animate-down { animation: fadeInDown 0.6s cubic-bezier(0.2, 0.8, 0.2, 1); }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(40px); } to { opacity: 1; transform: translateY(0); } }
@keyframes fadeInDown { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }
@keyframes pop { 50% { transform: scale(1.3); } }
@keyframes viewerSpin { to { transform: rotate(360deg); } }

/* 响应式 */
@media (max-width: 900px) {
  .product-detail-page { height: auto; overflow-y: auto; }
  .main-viewport { padding: 10px; height: auto; display: block; }
  .glass-card { flex-direction: column; max-height: none; border-radius: 20px; }
  .gallery-section { height: 400px; border-right: none; border-bottom: 1px solid #eee; }
  .info-scroll-area { max-height: none; padding: 16px 20px; }
  .action-dock { position: sticky; bottom: 0; box-shadow: 0 -4px 20px rgba(0,0,0,0.05); }

  .viewer-mask {
    padding: 0;
  }

  .viewer-shell {
    width: 100vw;
    height: 100vh;
    border-radius: 0;
  }

  .viewer-caption {
    max-width: 52vw;
  }

  .viewer-stage {
    gap: 8px;
    padding: 0 10px 10px;
    grid-template-columns: auto minmax(0, 1fr) auto;
  }

  .viewer-nav {
    width: 40px;
    height: 40px;
  }
}

@media (max-width: 480px) {
  .nav-header { 
    height: 50px; 
    padding: 0 12px;
    .brand-logo { font-size: 14px; }
    .back-btn { font-size: 13px; padding: 6px 10px; }
  }
  .gallery-section { height: 320px; }
  .product-header {
    .price-line {
      .num { font-size: 32px; }
      .symbol { font-size: 16px; }
    }
    .title { font-size: 20px; }
  }
  .desc-box { padding: 10px 12px; p { font-size: 14px; } }
  .action-dock {
    padding: 10px 16px;
    gap: 12px;
    .icon-btn .text { font-size: 10px; }
    .main-btns {
      gap: 8px;
      .btn { height: 40px; font-size: 14px; }
    }
  }

  .gallery-footer {
    padding: 12px 12px 16px;

    .thumb-strip {
      justify-content: flex-start;
    }

    .thumb-btn {
      width: 50px;
      height: 50px;
      border-radius: 12px;
    }
  }

  .viewer-topbar {
    padding: 12px;
  }

  .viewer-tools {
    gap: 6px;
  }

  .viewer-tool {
    width: 36px;
    height: 36px;
    border-radius: 12px;
  }

  .viewer-stage {
    grid-template-columns: minmax(0, 1fr);
    padding: 0 8px 8px;
  }

  .viewer-meta {
    max-width: calc(100vw - 150px);
  }

  .viewer-caption {
    font-size: 13px;
  }

  .viewer-nav {
    display: none;
  }

  .viewer-filmstrip {
    justify-content: flex-start;
    padding: 0 12px 14px;
  }

  .viewer-thumb {
    width: 54px;
    height: 54px;
    border-radius: 14px;
  }
}
</style>
