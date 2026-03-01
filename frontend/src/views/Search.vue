<template>
  <div class="search-page">
    <div class="fixed-header">
      <div class="header-inner">
        <div class="left-icon" @click="$router.push('/')">
          <el-icon><ArrowLeft /></el-icon>
        </div>

        <div class="search-group">
          <div class="input-wrapper">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
                v-model="searchParams.keyword"
                placeholder="搜一下，发现新奇好物..."
                class="custom-input"
                @keyup.enter="doSearch"
            />
            <el-icon v-if="searchParams.keyword" class="clear-icon" @click="searchParams.keyword=''"><CircleCloseFilled /></el-icon>
          </div>
          <button class="search-btn" @click="doSearch">搜索</button>
        </div>

        <div
            class="filter-toggle-btn"
            :class="{ active: isFilterOpen }"
            @click="isFilterOpen = !isFilterOpen"
        >
          <span class="text">筛选</span>
          <el-icon class="icon" :class="{ 'is-rotated': isFilterOpen }"><Filter /></el-icon>
        </div>
      </div>
    </div>

    <div class="filter-section-wrapper" :style="{ marginTop: '70px' }">
      <el-collapse-transition>
        <div v-show="isFilterOpen" class="filter-panel">
          <div class="panel-content">
            <div class="filter-row">
              <div class="label">排序</div>
              <div class="options">
                <div class="chip" :class="{ active: sortType === 'default' }" @click="changeSort('default')">综合</div>
                <div class="chip" :class="{ active: sortType === 'time' }" @click="changeSort('time')">最新</div>
                <div class="chip price-chip" :class="{ active: sortType.includes('price') }" @click="togglePriceSort">
                  价格
                  <div class="sort-icons">
                    <el-icon class="up" :class="{ on: sortType==='price_asc' }"><CaretTop /></el-icon>
                    <el-icon class="down" :class="{ on: sortType==='price_desc' }"><CaretBottom /></el-icon>
                  </div>
                </div>
              </div>
            </div>

            <div class="filter-row">
              <div class="label">价格</div>
              <div class="price-inputs">
                <input v-model.number="searchParams.minPrice" type="number" placeholder="最低价" />
                <span class="dash"></span>
                <input v-model.number="searchParams.maxPrice" type="number" placeholder="最高价" />
              </div>
            </div>

            <div class="filter-row">
              <div class="label">分类</div>
              <div class="options wrap">
                <div class="chip" :class="{ active: searchParams.categoryId === 0 }" @click="changeCategory(0)">全部分类</div>
                <div v-for="cat in categories" :key="cat.id" class="chip" :class="{ active: searchParams.category === cat.name }" @click="changeCategory(cat.name)">{{ cat.name }}</div>
              </div>
            </div>

            <div class="panel-footer">
              <span class="reset-btn" @click="resetFilter">清除条件</span>
              <button class="confirm-btn" @click="doSearch">确认筛选</button>
            </div>
          </div>
        </div>
      </el-collapse-transition>
    </div>

    <div class="result-container" v-loading="loading">
      <div v-if="productList.length > 0" class="product-grid">
        <div v-for="item in productList" :key="item.id" class="grid-item animate-up" @click="toDetail(item.id)">
          <div class="product-card">
            <div class="img-box">
              <img :src="item.image" loading="lazy" />
              <div v-if="item.status !== 1" class="mask-sold">
                {{ item.status === 2 ? '已售出' : '已下架' }}
              </div>
            </div>
            <div class="info-box">
              <div class="title" v-html="highlightKeyword(item.name || item.title)"></div>

              <div class="tags">
                <span class="tag" v-if="item.is_negotiable">可小刀</span>
                <span class="tag" v-if="item.is_home_delivery">送货上门</span>
                <span class="tag" v-if="item.is_self_pickup">自提</span>
              </div>

              <div class="footer">
                <div class="price">¥ <span class="num">{{ item.price }}</span></div>
                <div class="user">
                  <img :src="item.seller?.avatar || item.user?.avatar || defaultAvatar" />
                  <span>{{ item.seller?.nickname || item.seller?.username || item.user?.nickname || item.user?.username || '闲趣用户' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="!loading" class="empty-state">
        <img src="https://cube.elemecdn.com/e/fd/0fc7d20532fdaf769a25683617711png.png" />
        <p>这里空空如也，换个词试试？</p>
        <span class="empty-tip">没有找到相关宝贝，您可以尝试减少关键词或更换分类</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '@/utils/request'
import { Search, ArrowLeft, Filter, CircleCloseFilled, CaretTop, CaretBottom } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const isFilterOpen = ref(false)
const productList = ref([])
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const sortType = ref('default')

// 搜索参数状态
const searchParams = reactive({
  keyword: route.query.q || '',
  minPrice: null,
  maxPrice: null,
  category: '', // 存分类名称字符串
  categoryId: 0 // 仅用于高亮"全部分类"
})

const categories = ref([
  { id: 1, name: '数码' }, { id: 2, name: '书籍' }, { id: 3, name: '生活' },
  { id: 4, name: '服饰' }, { id: 5, name: '运动' }, { id: 6, name: '美妆' },
  { id: 7, name: '乐器' }, { id: 8, name: '手游' }, { id: 9, name: '兼职' }
])

// 执行搜索
const doSearch = async () => {
  loading.value = true
  try {
    // 构造后端需要的参数
    const params = {
      search: searchParams.keyword ? searchParams.keyword.trim() : '', // 后端接收 search
      category: searchParams.category, // 后端接收 category 字符串
      //sort: sortType.value, // 如果后端支持排序
      //min_price: searchParams.minPrice,
      //max_price: searchParams.maxPrice
    }

    const res = await request.get('/api/products', { params })

    // ★★★ 图片路径修复 ★★★
    productList.value = (res.list || []).map(item => {
      let img = item.image
      if (img && !img.startsWith('http')) {
        img = 'http://127.0.0.1:8081' + img
      }
      if (img) {
        img = img.replace('localhost', '127.0.0.1')
      }
      return { ...item, image: img }
    })

    isFilterOpen.value = false

    // 更新 URL，方便分享和刷新
    router.replace({
      query: {
        q: searchParams.keyword.trim(),
        cat: searchParams.category
      }
    })
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const changeCategory = (name) => {
  if (name === 0) {
    searchParams.category = ''
    searchParams.categoryId = 0
  } else {
    searchParams.category = name
    // 只是为了高亮效果，不传给后端
    searchParams.categoryId = 1
  }
  doSearch()
}

const changeSort = (type) => { sortType.value = type; doSearch() }
const togglePriceSort = () => { sortType.value = sortType.value === 'price_asc' ? 'price_desc' : 'price_asc'; doSearch() }

const resetFilter = () => {
  sortType.value = 'default'
  searchParams.minPrice = null
  searchParams.maxPrice = null
  searchParams.category = ''
  searchParams.categoryId = 0
}

const toDetail = (id) => router.push(`/product/${id}`)

// 高亮关键词
const highlightKeyword = (text) => {
  if (!searchParams.keyword || !text) return text
  const reg = new RegExp(searchParams.keyword, 'gi')
  return text.replace(reg, (match) => `<span style="color:#ff8200;font-weight:bold">${match}</span>`)
}

// 监听路由变化（例如从别的页面跳过来搜索）
watch(() => route.query.q, (newVal) => {
  if (newVal !== searchParams.keyword) {
    searchParams.keyword = newVal || ''
    doSearch()
  }
})

onMounted(doSearch)
</script>

<style scoped lang="scss">
@use 'sass:color';
/* 变量定义 */
$primary: #ffdf5d;
$text-main: #333;
$bg-page: #f4f5f7;
$max-width: 1200px;

.search-page { min-height: 100vh; background: $bg-page; }

/* 1. 顶部栏 (高度 70px) */
.fixed-header {
  position: fixed; top: 0; left: 0; right: 0; height: 70px; background: #fff; z-index: 1000; box-shadow: 0 2px 10px rgba(0,0,0,0.04);
  .header-inner {
    max-width: $max-width; margin: 0 auto; height: 100%; display: flex; align-items: center; padding: 0 24px; gap: 16px;
  }

  .left-icon {
    font-size: 24px; cursor: pointer; color: #666; transition: .2s;
    &:hover { color: #000; background: #f5f5f5; border-radius: 50%; }
  }

  /* 搜索组合 */
  .search-group {
    flex: 1; display: flex; align-items: center; gap: 8px;

    .input-wrapper {
      flex: 1; height: 44px; background: #f2f3f5; border-radius: 12px; display: flex; align-items: center; padding: 0 16px; transition: 0.3s; border: 2px solid transparent;
      &:focus-within { background: #fff; border-color: $primary; box-shadow: 0 4px 12px rgba(255, 223, 93, 0.2); }
      .search-icon { color: #999; font-size: 18px; margin-right: 8px; }
      .custom-input { flex: 1; border: none; background: transparent; outline: none; font-size: 15px; color: #333; &::placeholder { color: #ccc; } }
      .clear-icon { cursor: pointer; color: #ccc; font-size: 18px; &:hover { color: #999; } }
    }

    .search-btn {
      height: 44px; padding: 0 28px; background: $primary; border: none; border-radius: 12px; font-size: 15px; font-weight: bold; cursor: pointer; transition: .2s;
      &:hover { opacity: 0.9; transform: translateY(-1px); }
      &:active { transform: scale(0.98); }
    }
  }

  /* 筛选开关 */
  .filter-toggle-btn {
    display: flex; align-items: center; gap: 4px; cursor: pointer; padding: 8px 12px; border-radius: 8px; color: #666; transition: .2s; user-select: none;
    &:hover { background: #f5f5f5; color: #333; }
    &.active { background: rgba(255, 223, 93, 0.2); color: #000; font-weight: bold; }
    .text { font-size: 14px; }
    .icon { font-size: 16px; transition: transform 0.3s; }
  }
}

/* 2. 筛选面板 */
.filter-section-wrapper { background: #fff; position: relative; z-index: 900; }
.filter-panel {
  border-top: 1px solid #f5f5f5; border-bottom: 1px solid #eee; background: #fff; overflow: hidden;
  .panel-content { max-width: $max-width; margin: 0 auto; padding: 24px; }

  .filter-row {
    display: flex; align-items: flex-start; margin-bottom: 20px;
    .label { width: 50px; font-size: 14px; color: #999; padding-top: 6px; font-weight: bold; }
    .options { flex: 1; display: flex; flex-wrap: wrap; gap: 10px; }

    .chip {
      padding: 6px 18px; background: #f7f8fa; border-radius: 6px; font-size: 13px; color: #555; cursor: pointer; transition: .2s; border: 1px solid transparent;
      &:hover { background: #eee; }
      &.active { background: $primary; color: #000; border-color: color.adjust($primary, $lightness: -5%); font-weight: bold; }
      &.price-chip { display: flex; align-items: center; gap: 4px; .sort-icons { display: flex; flex-direction: column; transform: scale(0.7); .el-icon { height: 8px; color: #ccc; &.on { color: #000; } } } }
    }

    .price-inputs {
      display: flex; align-items: center; gap: 8px;
      input { width: 100px; height: 32px; background: #f7f8fa; border: 1px solid transparent; border-radius: 6px; text-align: center; font-size: 13px; outline: none; transition: .2s; &:focus { background: #fff; border-color: $primary; } }
      .dash { width: 10px; height: 1px; background: #ccc; }
    }
  }

  .panel-footer {
    display: flex; justify-content: flex-end; align-items: center; gap: 16px; padding-top: 16px; border-top: 1px dashed #eee;
    .reset-btn { font-size: 14px; color: #666; cursor: pointer; &:hover { text-decoration: underline; } }
    .confirm-btn { padding: 8px 24px; background: #333; color: $primary; border: none; border-radius: 99px; font-size: 13px; font-weight: bold; cursor: pointer; transition: .2s; &:hover { background: #000; } }
  }
}

/* 3. 结果网格 */
.result-container {
  max-width: $max-width;
  margin: 0 auto;
  padding: 24px;
  min-height: calc(100vh - 100px);
  display: flex;
  flex-direction: column;
}
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 24px;
  flex: 1;
}

.product-card {
  background: #fff; border-radius: 16px; overflow: hidden; cursor: pointer; transition: all 0.3s; border: 1px solid transparent;
  &:hover { transform: translateY(-6px); box-shadow: 0 12px 24px rgba(0,0,0,0.06); border-color: rgba(255, 223, 93, 0.4); }

  .img-box {
    width: 100%; aspect-ratio: 1; background: #f9f9f9; position: relative;
    img { width: 100%; height: 100%; object-fit: cover; display: block; }
    .mask-sold { position: absolute; inset: 0; background: rgba(0,0,0,0.6); color: #fff; display: flex; align-items: center; justify-content: center; font-weight: bold; letter-spacing: 2px; }
  }

  .info-box {
    padding: 16px;
    .title { font-size: 15px; height: 44px; line-height: 1.5; overflow: hidden; margin-bottom: 8px; color: #333; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; font-weight: 500; }
    .tags { display: flex; gap: 6px; margin-bottom: 12px; .tag { font-size: 10px; padding: 2px 6px; border: 1px solid #eee; color: #999; border-radius: 4px; background: #fff; } }

    .footer {
      display: flex; justify-content: space-between; align-items: center;
      .price { color: #ff5000; font-size: 13px; font-weight: 900; .num { font-size: 20px; } }
      .user { display: flex; align-items: center; gap: 6px; img { width: 20px; height: 20px; border-radius: 50%; } span { font-size: 12px; color: #999; max-width: 80px; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; } }
    }
  }
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0 120px;
  text-align: center;
  animation: fadeIn 0.8s ease;

  img {
    width: 200px;
    margin-bottom: 24px;
    opacity: 0.9;
    filter: drop-shadow(0 10px 20px rgba(0,0,0,0.05));
  }

  p {
    color: #666;
    font-size: 16px;
    font-weight: 600;
    margin: 0;
  }

  .empty-tip {
    font-size: 13px;
    color: #999;
    margin-top: 8px;
  }
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-up { animation: fadeInUp 0.5s ease backwards; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(15px); } to { opacity: 1; transform: translateY(0); } }

/* 4. 移动端适配 (Mobile Responsive) */
@media (max-width: 768px) {
  .fixed-header {
    height: 60px;
    .header-inner {
      padding: 0 12px;
      gap: 10px;
    }
    .search-group {
      .input-wrapper {
        height: 38px;
        padding: 0 10px;
        border-radius: 8px;
        .search-icon { font-size: 16px; margin-right: 4px; }
        .custom-input { font-size: 14px; }
      }
      .search-btn {
        height: 38px;
        padding: 0 16px;
        border-radius: 8px;
        font-size: 14px;
      }
    }
    .filter-toggle-btn {
      padding: 6px 8px;
      .text { display: none; } /* 移动端隐藏"筛选"文字，仅保留图标 */
      .icon { font-size: 20px; }
    }
  }

  .filter-section-wrapper {
    margin-top: 60px !important;
  }

  .filter-panel {
    .panel-content { padding: 16px; }
    .filter-row {
      margin-bottom: 16px;
      .label { width: 40px; font-size: 13px; }
      .chip { padding: 4px 12px; font-size: 12px; }
      .price-inputs input { width: 80px; height: 30px; }
    }
  }

  .result-container {
    padding: 12px;
  }

  .product-grid {
    grid-template-columns: repeat(2, 1fr); /* 移动端双列布局 */
    gap: 12px;
  }

  .product-card {
    .info-box {
      padding: 10px;
      .title { font-size: 13px; height: 36px; -webkit-line-clamp: 2; margin-bottom: 6px; }
      .footer {
        flex-direction: column;
        align-items: flex-start;
        gap: 8px;
        .price { .num { font-size: 16px; } }
        .user { width: 100%; justify-content: flex-start; }
      }
    }
  }
}

@media (max-width: 480px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }
  .product-card .info-box .title { font-size: 12px; height: 32px; }
}
</style>
