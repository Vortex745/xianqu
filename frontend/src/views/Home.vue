<template>
  <div class="app-container">
    <nav class="navbar" :class="{ 'scrolled': isScrolled }">
      <div class="container navbar-content">
        <div class="brand" @click="goHome">
          <div class="logo-box">
            <div class="circle-shape"></div>
            <div class="square-shape"></div>
          </div>
          <span class="brand-text">XIANQU</span>
        </div>
        <div class="nav-search-bar" :class="{ 'visible': isScrolled }">
          <input v-model="keyword" placeholder="搜索宝贝..." @keyup.enter="handleSearch" />
          <button class="mini-search-btn" @click="handleSearch">搜索</button>
        </div>

        <div class="nav-actions">
          <div v-if="currentLocation" class="location-badge animate-fade">
            <el-icon><Location /></el-icon>
            <span class="loc-text">{{ currentLocation }}</span>
          </div>

          <div v-else-if="isLocating" class="location-badge animate-fade" style="opacity: 0.6">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span class="loc-text">定位中...</span>
          </div>

          <div v-if="user" class="user-area">
            <div class="icon-btn-wrapper" @click="goToMessages" title="消息中心">
              <el-badge :value="unreadCount" :max="99" :hidden="unreadCount <= 0" class="nav-badge">
                <div class="icon-circle"><el-icon><ChatDotRound /></el-icon></div>
              </el-badge>
            </div>

            <div class="icon-btn-wrapper" @click="$router.push('/cart')" title="购物车">
              <el-badge :value="cartCount" :max="99" :hidden="cartCount === 0" class="nav-badge cart-badge">
                <div class="icon-circle"><el-icon><ShoppingCart /></el-icon></div>
              </el-badge>
            </div>

            <el-dropdown
                class="user-dropdown-trigger"
                trigger="hover"
                placement="bottom"
                popper-class="custom-dock-popper"
                :show-timeout="100"
                :hide-timeout="140"
            >
              <div class="user-profile">
                <span class="nickname">{{ user.nickname || user.username || '闲趣用户' }}</span>
                <el-avatar :size="38" :src="resolveUrl(user.avatar || defaultAvatar)" class="avatar-img" />
              </div>

              <template #dropdown>
                <el-dropdown-menu class="custom-dock-menu">
                  <el-dropdown-item @click="$router.push('/profile')"><el-icon><User /></el-icon>个人中心</el-dropdown-item>
                  <el-dropdown-item @click="$router.push('/orders')"><el-icon><List /></el-icon>我的订单</el-dropdown-item>
                  <el-dropdown-item @click="$router.push('/mysales')"><el-icon><Money /></el-icon>我卖出的</el-dropdown-item>
                  <el-dropdown-item @click="handleSwitchAccount"><el-icon><Switch /></el-icon>切换账号</el-dropdown-item>
                  <el-dropdown-item divided @click="logout" class="logout-item"><el-icon><SwitchButton /></el-icon>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div v-else class="auth-btns">
            <button class="btn-login-pill" @click="openAuthModal('login')">登录 / 注册</button>
          </div>
        </div>
      </div>
    </nav>

    <section class="hero-section">
      <div class="container hero-inner" :style="heroStyle">
        <div class="title-wrapper animate-up">
          <h1 class="hero-title">让闲置<span class="highlight-text">游起来</span></h1>
          <svg class="doodle-underline" viewBox="0 0 200 20" preserveAspectRatio="none">
            <path d="M5,15 Q50,5 90,15 T190,5" fill="none" stroke="#333" stroke-width="3" stroke-linecap="round" />
          </svg>
        </div>
        <p class="hero-subtitle animate-up delay-1">社区和校园都能淘好物 · 让闲置流转更轻松</p>
        <div class="search-box-wrapper animate-up delay-2">
          <div class="search-box-large">
            <div class="input-wrapper">
              <el-icon class="search-icon"><Search /></el-icon>
              <input v-model="keyword" placeholder="搜索宝贝..." @keyup.enter="handleSearch" />
            </div>
            <button class="btn-search-large" @click="handleSearch">搜索</button>
          </div>
        </div>
      </div>
      <svg class="hero-wave" viewBox="0 0 1440 320"><path fill="#f6f7f9" fill-opacity="1" d="M0,224L48,213.3C96,203,192,181,288,181.3C384,181,480,203,576,224C672,245,768,267,864,261.3C960,256,1056,224,1152,202.7C1248,181,1344,171,1392,165.3L1440,160L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"></path></svg>
    </section>

    <main class="main-content">
      <div class="container">
        <div class="action-bar-wrapper">
          <div class="filter-container">
            <div class="static-filter-list">
              <div v-for="cat in categories" :key="cat.id" class="filter-chip" :class="{ active: currentCat === cat.id }" @click="filterCategory(cat.id)">
                {{ cat.name }}
              </div>
            </div>
          </div>

          <div class="location-switch-wrapper">
            <div class="goods-scope-toggle" role="tablist" aria-label="商品范围切换">
              <span class="scope-indicator" :class="{ near: onlyNearby }"></span>
              <button
                type="button"
                class="scope-btn"
                :class="{ active: !onlyNearby }"
                @click="handleGoodsScopeSwitch(false)"
              >
                全部商品
              </button>
              <button
                type="button"
                class="scope-btn"
                :class="{ active: onlyNearby, disabled: !currentLocation }"
                :disabled="!currentLocation"
                @click="handleGoodsScopeSwitch(true)"
              >
                周边好物
              </button>
            </div>
          </div>

          <button class="btn-publish-float" @click="goToPublish">
            <el-icon><Plus /></el-icon> 发布闲置
          </button>
        </div>

        <div class="section-title">
          {{ onlyNearby && currentLocation ? `周围好物 (${currentLocation})` : '猜你喜欢' }}
        </div>

        <div v-if="loading || productList.length > 0" class="product-grid-system" v-loading="loading" element-loading-text="加载中..." element-loading-background="rgba(255, 255, 255, 0.6)">
          <div v-for="item in productList" :key="item.id" class="grid-item animate-fade" @click="openDetail(item.id)">
            <div class="product-card" :class="{ 'is-sold': item.status !== 1 }">
              <div class="card-image">
                <img :src="item.image" loading="lazy" />
                <div v-if="item.status !== 1" class="sold-overlay">
                  <div class="sold-text">{{ item.status === 2 ? '已售出' : '已下架' }}</div>
                </div>
              </div>
              <div class="card-details">
                <div class="title">{{ item.name || item.title || '未知商品' }}</div>

                <div class="price-row">
                  <span class="currency">¥</span><span class="amount">{{ item.price }}</span>
                  <span class="wants" v-if="item.status === 1">{{ item.view_count || 0 }}人想要</span>
                </div>
                <div class="seller-row">
                  <img :src="resolveUrl(item.seller?.avatar || defaultAvatar)" />
                  <span class="name">{{ item.seller?.nickname || item.seller?.username || '闲趣用户' }}</span>
                  <span class="credit-tag" v-if="item.status === 1">信用极好</span>
                </div>
                <div class="location-tag" v-if="item.area">
                  <el-icon><Goods /></el-icon> {{ item.area }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-state">
          <el-empty :description="onlyNearby ? '附近暂无商品，看全站吧' : '暂无更多宝贝'" :image-size="160" />
        </div>
      </div>
    </main>

    <el-backtop :right="40" :bottom="80" class="custom-backtop"><div class="backtop-inner"><el-icon :size="20"><CaretTop /></el-icon><span class="back-text">TOP</span></div></el-backtop>

    <el-drawer v-model="userDrawer" title="个人中心" direction="rtl" size="300px" class="user-drawer" destroy-on-close>
      <div class="drawer-profile">
        <el-avatar :size="80" :src="resolveUrl(user?.avatar || defaultAvatar)" class="big-avatar" />
        <h3 class="username">{{ user?.nickname || user?.username || '闲趣用户' }}</h3>
        <p class="uid">ID: {{ user?.id }}</p>
      </div>
      <div class="drawer-menu">
        <div class="menu-item" @click="$router.push('/orders')"><el-icon><List /></el-icon> <span>我的订单</span> <el-icon class="arrow"><ArrowRight /></el-icon></div>
        <div class="menu-item" @click="$router.push('/mysales')"><el-icon><Money /></el-icon> <span>我卖出的</span> <el-icon class="arrow"><ArrowRight /></el-icon></div>
        <div class="menu-item" @click="$router.push('/profile')"><el-icon><User /></el-icon> <span>个人资料</span> <el-icon class="arrow"><ArrowRight /></el-icon></div>
        <div class="menu-item" @click="handleSwitchAccount"><el-icon><Switch /></el-icon> <span>切换账号</span> <el-icon class="arrow"><ArrowRight /></el-icon></div>
      </div>
      <div class="drawer-footer">
        <button class="btn-logout" @click="logout">退出登录</button>
      </div>
    </el-drawer>

    <AuthModal v-model="showAuthModal" @success="handleLoginSuccess" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import request, { resolveUrl } from '@/utils/request'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, CaretTop, Switch, ShoppingCart, User, List, SwitchButton, ChatDotRound, Money, ArrowRight, Location, Goods, Loading } from '@element-plus/icons-vue'
import AuthModal from '../components/AuthModal.vue'

const router = useRouter()
const route = useRoute()
const keyword = ref('')
const loading = ref(false)
const productList = ref([])
const categories = ref([ { id: 0, name: '全部' }, { id: '数码', name: '数码' }, { id: '书籍', name: '书籍' }, { id: '生活', name: '生活' }, { id: '服饰', name: '服饰' }, { id: '运动', name: '运动' }, { id: '美妆', name: '美妆' }, { id: '乐器', name: '乐器' }, { id: '手游', name: '手游' }, { id: '兼职', name: '兼职' } ])
const currentCat = ref(0)
const isScrolled = ref(0)
const scrollY = ref(0)
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const cartCount = ref(0)
const unreadCount = ref(0)

const normalizeToken = (raw = '') => {
  const value = String(raw || '').trim()
  if (!value) return ''
  return value.toLowerCase().startsWith('bearer ') ? value.slice(7).trim() : value
}

const isTokenExpired = (token = '') => {
  try {
    const payload = token.split('.')[1]
    if (!payload) return true
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    if (!decoded?.exp) return false
    return Date.now() >= decoded.exp * 1000
  } catch (e) {
    return true
  }
}

const rawToken = normalizeToken(localStorage.getItem('token'))
const hasValidToken = !!rawToken && !isTokenExpired(rawToken)
const user = ref(hasValidToken ? JSON.parse(localStorage.getItem('user') || 'null') : null)
const hasToken = ref(hasValidToken)
if (!hasValidToken && localStorage.getItem('token')) {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
}

const currentLocation = ref(localStorage.getItem('userLocation') || '')
const isLocating = ref(false)

// Cookie 工具函数：用于持久化存储状态
const setCookie = (name, value, days = 365) => {
  const d = new Date();
  d.setTime(d.getTime() + (days * 24 * 60 * 60 * 1000));
  document.cookie = `${name}=${value};expires=${d.toUTCString()};path=/`;
}
const getCookie = (name) => {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
  return match ? match[2] : null;
}

// ★★★ 默认开启“只看附近” (持久化：优先从 Cookie 读取) ★★★
const savedNearbyState = getCookie('user_pref_only_nearby')
// 如果 Cookie 不存在（首次访问或已清除），默认为 true；否则使用 Cookie 存储的值
const onlyNearby = ref(savedNearbyState === null ? true : savedNearbyState === 'true')

const userDrawer = ref(false)
const showAuthModal = ref(false)
let globalSocket = null

const handleScroll = () => { scrollY.value = window.scrollY; isScrolled.value = window.scrollY > 100 }
const heroStyle = computed(() => ({ opacity: Math.max(0, 1 - scrollY.value / 250), transform: `translateY(-${scrollY.value * 0.3}px)` }))

const handleSearch = () => { if (keyword.value.trim()) router.push({ path: '/search', query: { q: keyword.value } }) }
const refreshPage = () => { keyword.value = ''; currentCat.value = 0; if (router.currentRoute.value.path !== '/' && router.currentRoute.value.path !== '/home') { router.push('/') } else { fetchProducts() } }

// ★★★ 核心修改：带缓存 (Cookie/Storage) 的定位逻辑 ★★★
const getLocation = () => {
  return new Promise((resolve) => {
    // 1. 检查缓存：有效期 24 小时
    const cachedLoc = localStorage.getItem('userLocation')
    const cachedTime = localStorage.getItem('userLocationTime')
    const now = Date.now()
    const ONE_DAY = 24 * 60 * 60 * 1000

    if (cachedLoc && cachedTime && (now - Number(cachedTime) < ONE_DAY)) {
      currentLocation.value = cachedLoc
      resolve(cachedLoc)
      return
    }

    // 2. 无缓存或过期，调用 API
    isLocating.value = true

    const finish = (location = '') => {
      isLocating.value = false
      resolve(location)
    }

    const checkAndLocate = () => {
      if (typeof AMap !== 'undefined') {
        executeLocate()
      } else {
        setTimeout(checkAndLocate, 500)
      }
    }

    const executeLocate = () => {
      AMap.plugin(['AMap.Geolocation', 'AMap.Geocoder'], function() {
        const geolocation = new AMap.Geolocation({
          enableHighAccuracy: true,
          timeout: 10000,
          extensions: 'base'
        })

        geolocation.getCurrentPosition(function(status, result) {
          if (status !== 'complete') {
            finish('')
            return
          }

          const position = result.position
          const geocoder = new AMap.Geocoder({ radius: 500, extensions: 'all' })

          geocoder.getAddress([position.lng, position.lat], function(geoStatus, regeoResult) {
            if (geoStatus !== 'complete' || !regeoResult?.regeocode) {
              finish('')
              return
            }

            const regeo = regeoResult.regeocode
            const pois = regeo.pois || []
            let bestName = ''

            for (const poi of pois) {
              const type = poi.type || ''
              const name = poi.name || ''
              if (type.includes('学校') || type.includes('高等院校') || type.includes('中学') || type.includes('大学')) {
                bestName = name
                break
              }
              if (type.includes('住宅') || type.includes('居住') || type.includes('小区') || type.includes('公寓')) {
                if (!bestName) bestName = name
              }
            }

            if (!bestName) {
              bestName = regeo.addressComponent.neighborhood || regeo.addressComponent.building || regeo.addressComponent.township || regeo.formattedAddress
            }

            if (bestName) {
              currentLocation.value = bestName
              localStorage.setItem('userLocation', bestName)
              localStorage.setItem('userLocationTime', Date.now())
              finish(bestName)
              return
            }

            finish('')
          })
        })
      })
    }

    checkAndLocate()
  })
}

// ★★★ 修改：支持区域过滤参数 ★★★
const fetchProducts = async () => {
  loading.value = true
  try {
    const params = { page: 1, page_size: 20 }
    if (currentCat.value !== 0) params.category = currentCat.value
    else params.is_random = true

    // 如果开启了只看附近，且有定位，则传 area
    if (onlyNearby.value && currentLocation.value) {
      params.area = currentLocation.value
    }

    const res = await request.get('/api/products', { params })

    productList.value = (res.list || []).map(item => {
      return { ...item, image: resolveUrl(item.image) }
    })
    return true
  } catch (e) {
    productList.value = []
    return false
  } finally { loading.value = false }
}

const handleGoodsScopeSwitch = (nextNearby) => {
  if (nextNearby && !currentLocation.value) {
    ElMessage.info(isLocating.value ? '定位中，飞鸽传书...' : '定位到了才能看身边好物')
    return
  }
  if (onlyNearby.value === nextNearby) return
  onlyNearby.value = nextNearby
  setCookie('user_pref_only_nearby', onlyNearby.value)
  fetchProducts()
}

const fetchCartCount = async () => {
  if (!user.value || !hasToken.value) return false
  try {
    const res = await request.get('/api/cart')
    cartCount.value = (res.data || []).length
    return true
  } catch (e) {
    cartCount.value = 0
    return false
  }
}
const goToMessages = () => { unreadCount.value = 0; router.push('/messages') }
const goToPublish = () => { if (user.value && hasToken.value) { router.push('/publish') } else { ElMessage.warning('站住，先上车'); showAuthModal.value = true } }

const initGlobalWebSocket = () => {
  if (!hasToken.value || globalSocket) return
  const token = normalizeToken(localStorage.getItem('token'))
  if (!token) return
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const apiBase = import.meta.env.VITE_API_URL || ''
  const wsHost = apiBase ? apiBase.replace(/^https?:\/\//, '') : (import.meta.env.DEV ? 'localhost:8081' : window.location.host)
  const wsUrl = `${wsProtocol}://${wsHost}/api/ws?token=${encodeURIComponent(token)}`
  globalSocket = new WebSocket(wsUrl)
  globalSocket.onopen = () => {
    // reset badge on reconnect
    unreadCount.value = Math.max(0, unreadCount.value)
  }
  globalSocket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (Number(msg.receiver_id) === Number(user.value.id)) {
        unreadCount.value++
        ElMessage.info({ message: `有人撩你`, grouping: true })
      }
    } catch (e) {}
  }
  globalSocket.onerror = () => {
    if (globalSocket) {
      globalSocket.close()
      globalSocket = null
    }
  }
  globalSocket.onclose = () => {
    globalSocket = null
  }
}

const handleLoginSuccess = async (u) => {
  user.value = u;
  hasToken.value = true;
  await fetchCartCount();
  initGlobalWebSocket();
  await getLocation();
  await fetchProducts();

  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''
  if (redirect && redirect.startsWith('/')) {
    router.replace(redirect)
  }
}

const handleSwitchAccount = () => {
  ElMessageBox.confirm('确定要切换其他账号吗？','切换账号',{ confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning', center: true, customClass: 'warm-theme-box' })
      .then(() => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        user.value = null; hasToken.value = false; userDrawer.value = false; if(globalSocket) { globalSocket.close(); globalSocket = null; } showAuthModal.value = true;
      }).catch(() => {})
}
const logout = () => {
  ElMessageBox.confirm('确定要退出登录吗？','退出登录',{ confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning', center: true, customClass: 'warm-theme-box' })
      .then(() => { localStorage.removeItem('token'); localStorage.removeItem('user'); user.value = null; hasToken.value = false; userDrawer.value = false; cartCount.value = 0; unreadCount.value = 0; if(globalSocket) { globalSocket.close(); globalSocket = null; } ElMessage.success('退出登录成功') }).catch(() => {})
}
const openDetail = (id) => { router.push(`/product/${id}`) }
const filterCategory = (id) => { currentCat.value = id; fetchProducts() }
const openAuthModal = (mode) => { showAuthModal.value = true }

const openAuthModalByRoute = () => {
  if (route.query.auth !== 'login') return
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : ''

  if (user.value && hasToken.value) {
    if (redirect && redirect.startsWith('/')) {
      router.replace(redirect)
    } else {
      router.replace('/')
    }
    return
  }
  showAuthModal.value = true
}

onMounted(async () => {
  window.addEventListener('scroll', handleScroll)
  const located = await getLocation()
  if (!located && onlyNearby.value) {
    onlyNearby.value = false
    setCookie('user_pref_only_nearby', 'false')
  }
  const productOk = await fetchProducts()
  if (hasToken.value && productOk) {
    const cartOk = await fetchCartCount()
    if (cartOk) initGlobalWebSocket()
  }
  openAuthModalByRoute()
})
onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  if (globalSocket) {
    globalSocket.close()
    globalSocket = null
  }
})

watch(
  () => route.query.auth,
  () => {
    openAuthModalByRoute()
  }
)
</script>

<style lang="scss">
/* 下拉菜单 - 亚克力圆角风格 */
.custom-dock-menu {
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 14px;
  box-shadow: 0 10px 24px rgba(18, 28, 45, 0.14), 0 2px 8px rgba(0,0,0,0.03);
  padding: 7px;
  min-width: 170px;

  .el-dropdown-menu__item {
    border-radius: 10px;
    padding: 9px 10px;
    margin: 0;
    font-weight: 700;
    color: #4d5565;
    transition: all 0.2s cubic-bezier(0.25, 0.8, 0.25, 1);
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 34px;
    font-size: 13px;
    text-align: left;

    .el-icon { font-size: 14px; color: #636c7b; }

    &:hover {
      background-color: #fff7d1 !important;
      color: #1f2632 !important;
      transform: translateX(2px);
      .el-icon { color: #1f2632; }
    }

    &.logout-item {
      color: #e54545;
      border-top: 1px dashed #e7ebf0;
      margin-top: 4px;
      padding-top: 10px;
      &:hover {
        background-color: #fff1f1 !important;
        color: #e54545 !important;
      }
    }
  }
}

.warm-theme-box { border-radius: 20px !important; .el-button { border-radius: 99px; } }
</style>

<style scoped lang="scss">
$primary: #ffdf5d; $bg-color: #f6f7f9;
.app-container { background: $bg-color; min-height: 100vh; font-family: "PingFang SC", sans-serif; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; width: 100%; box-sizing: border-box; }
.navbar { height: 64px; position: sticky; top: 0; z-index: 999; background: $primary; transition: all 0.3s; &.scrolled { box-shadow: 0 4px 12px rgba(0,0,0,0.05); } .navbar-content { height: 100%; display: flex; align-items: center; justify-content: space-between; } }
.brand {
  display: flex; align-items: center; gap: 12px; cursor: pointer; color: #000;
  .logo-box {
    width: 32px; height: 32px; position: relative;
    .circle-shape { position: absolute; width: 22px; height: 22px; background: #000; border-radius: 50%; top: 0; left: 0; z-index: 1; }
    .square-shape {
      position: absolute; width: 20px; height: 20px; border: 2px solid #fff; bottom: 0; right: 0; border-radius: 6px;
      background: rgba(255, 255, 255, 0.4); backdrop-filter: blur(4px); z-index: 2; box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    }
  }
  .brand-text { font-size: 24px; font-weight: 900; color: #000; letter-spacing: 2px; font-family: 'Arial Black', sans-serif; }
}
.nav-search-bar { opacity: 0; transform: translateY(10px); pointer-events: none; transition: all 0.3s; background: #fff; border-radius: 99px; padding: 4px 4px 4px 16px; display: flex; align-items: center; width: 320px; &.visible { opacity: 1; transform: translateY(0); pointer-events: auto; } input { border: none; outline: none; font-size: 14px; flex: 1; } .mini-search-btn { background: $primary; border: none; padding: 6px 16px; border-radius: 99px; font-weight: bold; cursor: pointer; } }
.nav-actions { display: flex; align-items: center; gap: 20px; font-weight: 600; .user-area { display: flex; align-items: center; gap: 20px; } .icon-btn-wrapper { cursor: pointer; transition: all 0.3s; display: flex; align-items: center; justify-content: center; height: 100%; .nav-badge { display: flex; align-items: center; } .icon-circle { width: 40px; height: 40px; background: #000; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: $primary; box-shadow: 0 4px 10px rgba(0,0,0,0.1); font-size: 20px; } &:hover { transform: scale(1.1) rotate(-5deg); } } .user-profile { display: flex; align-items: center; gap: 10px; cursor: pointer; outline: none; padding: 4px 4px 4px 16px; border-radius: 99px; transition: all 0.3s ease; border: 1px solid transparent; .nickname { font-size: 14px; font-weight: 700; color: #333; } .avatar-img { border: 2px solid #fff; transition: transform 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275); } &:focus { outline: none; } &:hover { background: rgba(255, 255, 255, 0.4); box-shadow: 0 4px 12px rgba(0,0,0,0.08); transform: translateY(-1px); .avatar-img { transform: scale(1.1); } } } .btn-login-pill { background: #000; color: $primary; border: none; padding: 8px 24px; border-radius: 99px; font-weight: bold; cursor: pointer; transition: 0.2s; &:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); } } }
.user-dropdown-trigger {
  position: relative;
  display: inline-flex;
  align-items: center;
}
.user-dropdown-trigger :deep(.el-dropdown__menu-wrap) {
  right: 4px;
  top: calc(100% + 8px);
}
.hero-section { position: relative; padding: 60px 0 100px; text-align: center; background: linear-gradient(180deg, $primary 0%, rgba(255, 223, 93, 0.2) 60%, $bg-color 100%); .hero-inner { position: relative; z-index: 2; } .title-wrapper { position: relative; display: inline-block; } .hero-title { font-size: 48px; font-weight: 900; margin-bottom: 8px; color: #2c2c2c; letter-spacing: 2px; position: relative; z-index: 2; text-shadow: 2px 2px 0px rgba(255, 255, 255, 0.5); .highlight-text { color: #000; position: relative; } } .doodle-underline { position: absolute; bottom: -10px; left: 0; width: 100%; height: 20px; z-index: 1; opacity: 0.6; path { stroke: #000; stroke-width: 3; stroke-linecap: round; fill: none; } } .hero-subtitle { font-size: 16px; color: #555; margin-bottom: 30px; margin-top: 10px; font-weight: 500; letter-spacing: 1px; } .search-box-wrapper { display: flex; justify-content: center; } .search-box-large { width: 420px; max-width: 600px; height: 60px; background: #fff; border-radius: 30px; display: flex; align-items: center; padding: 6px; border: 2px solid transparent; box-shadow: 0 8px 24px rgba(255, 218, 68, 0.25); transition: all 0.4s cubic-bezier(0.25, 0.8, 0.25, 1); .input-wrapper { flex: 1; display: flex; align-items: center; padding-left: 20px; cursor: text; .search-icon { font-size: 20px; color: #999; margin-right: 10px; } input { border: none; outline: none; font-size: 16px; width: 100%; height: 100%; background: transparent; } } .btn-search-large { background: $primary; color: #000; border: 2px solid transparent; height: 100%; padding: 0 32px; border-radius: 24px; font-size: 18px; font-weight: 900; cursor: pointer; transition: 0.2s; white-space: nowrap; &:hover { background: #f0cf4b; } } &:hover, &:focus-within { width: 600px; transform: translateY(-2px); box-shadow: 0 16px 32px rgba(255, 218, 68, 0.35); border-color: $primary; } } .hero-wave { position: absolute; bottom: 0; left: 0; width: 100%; z-index: 1; } }
.main-content { padding-bottom: 60px; position: relative; z-index: 3; margin-top: -60px; }
.action-bar-wrapper { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; .filter-container { flex: 1; display: flex; gap: 12px; } .static-filter-list { display: flex; gap: 12px; flex-wrap: wrap; } .filter-chip { padding: 8px 20px; background: #fff; border-radius: 99px; font-size: 14px; color: #666; cursor: pointer; font-weight: 600; box-shadow: 0 2px 8px rgba(0,0,0,0.03); transition: 0.2s; user-select: none; &:hover { transform: translateY(-2px); color: #333; } &.active { background: #000; color: $primary; } }

  /* ★★★ 优化：确保切换开关不被挤压 ★★★ */
  .location-switch-wrapper { display: flex; align-items: center; gap: 8px; margin-right: 20px; white-space: nowrap; }
  .btn-publish-float { background: #000; color: $primary; border: none; padding: 10px 24px; border-radius: 99px; font-weight: 700; display: flex; align-items: center; gap: 6px; cursor: pointer; box-shadow: 0 4px 16px rgba(0,0,0,0.2); transition: 0.2s; white-space: nowrap; &:hover { transform: translateY(-2px) scale(1.02); } } }
.section-title { font-size: 18px; font-weight: 900; margin-bottom: 16px; color: #333; border-left: 4px solid $primary; padding-left: 10px; }
.product-grid-system { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 20px; min-height: 60vh; transition: all 0.3s; }
.product-card { background: #fff; border-radius: 16px; overflow: hidden; border: none; transition: all 0.3s ease; display: flex; flex-direction: column; cursor: pointer; box-shadow: 0 2px 8px rgba(0,0,0,0.02); &:hover { transform: translateY(-6px); box-shadow: 0 12px 24px rgba(0,0,0,0.08); .card-overlay { opacity: 1; } } &.is-sold { filter: grayscale(100%); opacity: 0.8; &:hover { transform: none; box-shadow: 0 2px 8px rgba(0,0,0,0.02); cursor: default; } } .card-image { position: relative; padding-top: 100%; background: #f0f0f0; img { position: absolute; top: 0; left: 0; width: 100%; height: 100%; object-fit: cover; } .card-overlay { position: absolute; inset: 0; background: rgba(0,0,0,0.2); display: flex; align-items: center; justify-content: center; opacity: 0; transition: 0.3s; .view-text { background: #fff; padding: 6px 16px; border-radius: 99px; font-size: 12px; font-weight: bold; } } .sold-overlay { position: absolute; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; z-index: 2; .sold-text { border: 3px solid #fff; color: #fff; padding: 4px 12px; font-size: 16px; font-weight: 900; transform: rotate(-15deg); border-radius: 8px; letter-spacing: 2px; } } } .card-details { padding: 14px; .title { font-size: 15px; margin-bottom: 12px; height: 40px; overflow: hidden; line-height: 1.4; color: #333; } .price-row { display: flex; align-items: baseline; gap: 2px; margin-bottom: 12px; .currency { font-size: 12px; color: #ff5000; font-weight: bold; } .amount { font-size: 20px; color: #ff5000; font-weight: 800; } .wants { font-size: 12px; color: #999; margin-left: auto; } .sold-label { font-size: 12px; color: #999; margin-left: auto; font-weight: bold; } } .seller-row { display: flex; align-items: center; gap: 6px; img { width: 20px; height: 20px; border-radius: 50%; } .name { font-size: 12px; color: #666; } .credit-tag { font-size: 10px; background: #e6fdfb; color: #00b578; padding: 1px 4px; border-radius: 4px; margin-left: auto; } } .location-tag { margin-top: 8px; font-size: 11px; color: #999; display: flex; align-items: center; gap: 3px; } } }
.empty-state { min-height: 50vh; display: flex; align-items: center; justify-content: center; opacity: 0.8; }
.user-drawer { .drawer-profile { text-align: center; padding: 40px 0; background: #fffcf0; border-bottom: 1px solid #eee; .big-avatar { border: 4px solid #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.05); margin-bottom: 10px; } .username { font-size: 18px; margin: 0; font-weight: bold; color: #333; } .uid { font-size: 12px; color: #999; margin-top: 4px; } } .drawer-menu { padding: 10px 0; .menu-item { padding: 16px 24px; display: flex; align-items: center; gap: 12px; font-size: 15px; color: #333; cursor: pointer; transition: 0.2s; &:hover { background: #f9f9f9; } .arrow { margin-left: auto; color: #ddd; font-size: 12px; } } } .drawer-footer { position: absolute; bottom: 20px; left: 0; right: 0; padding: 0 20px; .btn-logout { width: 100%; height: 44px; border: 1px solid #fee; background: #fff5f5; color: #ff4d4f; border-radius: 12px; font-size: 15px; cursor: pointer; font-weight: bold; &:hover { background: #ff4d4f; color: #fff; } } } }
.custom-backtop { width: 44px; height: 44px; border-radius: 12px; background: #fff; color: #000; border: 2px solid #000; box-shadow: 4px 4px 0 #000; transition: all 0.2s; overflow: hidden; &:hover { transform: translate(-2px, -2px); box-shadow: 6px 6px 0 #000; color: $primary; } .backtop-inner { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; } .back-text { font-size: 10px; font-weight: 900; line-height: 1; margin-top: -2px; } }

.location-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(0,0,0,0.1);
  padding: 4px 12px;
  border-radius: 99px;
  font-size: 13px;
  font-weight: bold;
  color: #333;
  cursor: default;
  transition: 0.2s;

  &:hover {
    background: #fff;
    box-shadow: 0 4px 10px rgba(0,0,0,0.05);
  }

  .is-loading {
    animation: rotate 1s linear infinite;
  }
}

:deep(.cart-badge.el-badge .el-badge__content) {
  background: #ff3b30;
  color: #fff;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  font-weight: 800;
  box-shadow: 0 0 0 2px #f4d854;
  right: -8px;
  top: -7px;
}

.location-switch-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-right: 18px;
  min-width: 220px;
  flex: 0 0 auto;
}

.goods-scope-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
  padding: 4px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.76);
  border: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72), 0 3px 12px rgba(0, 0, 0, 0.06);
  width: 206px;
}

.scope-indicator {
  position: absolute;
  top: 4px;
  left: 4px;
  width: 98px;
  height: 32px;
  border-radius: 999px;
  background: #fff;
  box-shadow: 0 6px 14px rgba(11, 16, 26, 0.1);
  transition: transform 0.28s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.scope-indicator.near {
  transform: translateX(100%);
}

.scope-btn {
  height: 32px;
  width: 98px;
  padding: 0;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: #717a89;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
  transition: color 0.2s ease;
  position: relative;
  z-index: 1;
  outline: none;
  box-shadow: none;
}

.scope-btn.active {
  color: #111824;
  font-weight: 800;
}

.scope-btn:hover:not(.active):not(.disabled) {
  color: #3b4658;
}

.scope-btn.disabled,
.scope-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.scope-btn:focus,
.scope-btn:focus-visible {
  outline: none;
  box-shadow: none;
}

@media (max-width: 980px) {
  .location-switch-wrapper {
    min-width: 180px;
    margin-right: 8px;
  }
  .goods-scope-toggle { width: 180px; }
  .scope-indicator { width: 85px; }
  .scope-btn {
    width: 85px;
    font-size: 12px;
  }
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
