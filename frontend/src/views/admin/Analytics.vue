<template>
  <div class="analytics-page">
    <section class="analytics-hero panel">
      <div class="analytics-hero__top">
        <div class="title-wrap">
          <span class="title-wrap__eyebrow">用户行为洞察</span>
          <h2>行为分析</h2>
          <p>追踪近期浏览、收藏、加购与成交走势，快速判断哪里在放大转化，哪里正在流失注意力。</p>
        </div>

        <div class="analytics-hero__actions">
          <div class="range-note">
            <span class="range-note__label">数据窗口</span>
            <strong>{{ activeRangeLabel }}</strong>
            <small>更新于 {{ lastUpdatedAt || '--:--' }}</small>
          </div>

          <div class="range-switch">
            <button
              v-for="item in rangeOptions"
              :key="item.value"
              class="range-btn"
              :class="{ active: range === item.value }"
              @click="range = item.value"
            >
              {{ item.label }}
            </button>
          </div>
        </div>
      </div>

      <div class="insight-grid">
        <article class="insight-card">
          <span class="insight-card__label">高意向动作</span>
          <strong>{{ formatCount(intentActions) }}</strong>
          <p>收藏、加购与成交累计，能更快看出用户是否真的准备下单。</p>
        </article>

        <article class="insight-card">
          <span class="insight-card__label">峰值活跃时段</span>
          <strong>{{ peakActivityLabel }}</strong>
          <p>热力图里最活跃的时间窗，适合安排重点活动和内容曝光。</p>
        </article>

        <article class="insight-card">
          <span class="insight-card__label">当前热商品</span>
          <strong>{{ leadProductLabel }}</strong>
          <p>{{ leadProductMeta }}</p>
        </article>
      </div>
    </section>

    <template v-if="loading">
      <section class="loading-grid" aria-label="行为分析加载中">
        <div class="loading-kpis">
          <article v-for="card in 4" :key="card" class="loading-card">
            <span class="loading-card__icon"></span>
            <div class="loading-card__meta">
              <span class="loading-line short"></span>
              <span class="loading-line medium"></span>
            </div>
          </article>
        </div>

        <article class="loading-panel">
          <div class="loading-panel__head">
            <span class="loading-line short"></span>
            <span class="loading-pill"></span>
          </div>
          <div class="loading-chart loading-chart--trend"></div>
        </article>

        <div class="loading-split">
          <article class="loading-panel">
            <div class="loading-panel__head">
              <span class="loading-line short"></span>
            </div>
            <div class="loading-table">
              <span v-for="row in 5" :key="row" class="loading-table__row"></span>
            </div>
          </article>

          <article class="loading-panel">
            <div class="loading-panel__head">
              <span class="loading-line short"></span>
              <span class="loading-pill"></span>
            </div>
            <div class="loading-chart loading-chart--pie"></div>
          </article>
        </div>

        <article class="loading-panel">
          <div class="loading-panel__head">
            <span class="loading-line short"></span>
            <span class="loading-pill"></span>
          </div>
          <div class="loading-chart loading-chart--heat"></div>
        </article>
      </section>
    </template>

    <template v-else>
      <section class="kpi-grid">
        <article class="kpi-card">
          <div class="kpi-icon amber">
            <IconifySymbol icon="lucide:users-round" :size="20" color="#111" />
          </div>
          <div class="kpi-meta">
            <span>日活用户</span>
            <strong>{{ formatCount(overview.dau) }}</strong>
          </div>
        </article>

        <article class="kpi-card">
          <div class="kpi-icon gray">
            <IconifySymbol icon="lucide:package-search" :size="20" color="#111" />
          </div>
          <div class="kpi-meta">
            <span>在售商品</span>
            <strong>{{ formatCount(overview.on_sale_count) }}</strong>
          </div>
        </article>

        <article class="kpi-card">
          <div class="kpi-icon mint">
            <IconifySymbol icon="lucide:banknote" :size="20" color="#111" />
          </div>
          <div class="kpi-meta">
            <span>成交总额</span>
            <strong>¥{{ formatMoney(overview.gmv) }}</strong>
          </div>
        </article>

        <article class="kpi-card">
          <div class="kpi-icon clay">
            <IconifySymbol icon="lucide:target" :size="20" color="#111" />
          </div>
          <div class="kpi-meta">
            <span>浏览转化率</span>
            <strong>{{ formatPercent(overview.conversion_rate) }}</strong>
          </div>
        </article>
      </section>

      <section class="panel trend-panel">
        <div class="panel-head">
          <div class="panel-head__copy">
            <h3>行为趋势</h3>
            <p>浏览、收藏、加购与成交的变化轨迹</p>
          </div>
          <button class="export-btn" @click="exportChart(trendChart, '行为趋势')">导出图片</button>
        </div>

        <div class="chart-shell" :class="{ 'is-empty': !hasTrendData }">
          <div ref="trendEl" class="chart-box"></div>
          <div v-if="!hasTrendData" class="chart-empty">
            <strong>暂无趋势数据</strong>
            <p>切换时间窗口后重试，或等待用户行为数据同步进来。</p>
          </div>
        </div>
      </section>

      <section class="mid-grid">
        <article class="panel table-panel">
          <div class="panel-head">
            <div class="panel-head__copy">
              <h3>热门商品 Top10</h3>
              <p>优先关注浏览与成交都在上升的商品</p>
            </div>
          </div>

          <div v-if="hotProducts.length === 0" class="panel-empty">暂无热门商品数据</div>
          <div v-else class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>商品</th>
                  <th>分类</th>
                  <th>浏览</th>
                  <th>收藏</th>
                  <th>成交</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in hotProducts" :key="item.product_id">
                  <td>{{ item.name || '未命名' }}</td>
                  <td>{{ item.category || '其他' }}</td>
                  <td>{{ formatCount(item.view_count) }}</td>
                  <td>{{ formatCount(item.like_count) }}</td>
                  <td>{{ formatCount(item.buy_count) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>

        <article class="panel pie-panel">
          <div class="panel-head">
            <div class="panel-head__copy">
              <h3>分类成交占比</h3>
              <p>判断成交主要集中在哪些品类</p>
            </div>
            <button class="export-btn" @click="exportChart(pieChart, '分类占比')">导出图片</button>
          </div>

          <div class="chart-shell compact" :class="{ 'is-empty': !hasCategoryData }">
            <div ref="pieEl" class="chart-box compact"></div>
            <div v-if="!hasCategoryData" class="chart-empty">
              <strong>暂无分类成交数据</strong>
              <p>分类成交出现后，这里会自动给出占比结构。</p>
            </div>
          </div>
        </article>
      </section>

      <section class="panel heat-panel">
        <div class="panel-head">
          <div class="panel-head__copy">
            <h3>用户活跃热力图</h3>
            <p>看清一周内每天每小时的活跃分布</p>
          </div>
          <button class="export-btn" @click="exportChart(heatChart, '活跃热力图')">导出图片</button>
        </div>

        <div class="chart-shell heat" :class="{ 'is-empty': !hasActivityData }">
          <div ref="heatEl" class="chart-box heat"></div>
          <div v-if="!hasActivityData" class="chart-empty">
            <strong>暂无活跃热力数据</strong>
            <p>后续有用户行为产生时，这里会显示不同日期和时段的活跃强度。</p>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from '@/ui/feedback'
import request from '@/utils/request'
import IconifySymbol from '@/components/IconifySymbol.vue'

const loading = ref(true)
const lastUpdatedAt = ref('')
const range = ref(7)
const rangeOptions = [
  { label: '今日', value: 1 },
  { label: '近7天', value: 7 },
  { label: '近30天', value: 30 }
]

const overview = ref({})
const trend = ref({ dates: [], series: { view: [], like: [], cart: [], buy: [] } })
const hotProducts = ref([])
const categoryStats = ref([])
const activity = ref({ days: [], hours: [], data: [], max_count: 0 })

const trendEl = ref(null)
const pieEl = ref(null)
const heatEl = ref(null)

let trendChart = null
let pieChart = null
let heatChart = null

const formatMoney = (value) => Number(value || 0).toFixed(2)
const formatPercent = (value) => `${Number(value || 0).toFixed(2)}%`
const formatCount = (value) => new Intl.NumberFormat('zh-CN').format(Number(value || 0))

const activeRangeLabel = computed(() => rangeOptions.find((item) => item.value === range.value)?.label || `${range.value}天`)
const hasTrendData = computed(() => {
  const dates = trend.value?.dates || []
  const series = trend.value?.series || {}
  return dates.length > 0 && ['view', 'like', 'cart', 'buy'].some((key) => (series[key] || []).length > 0)
})
const hasCategoryData = computed(() => (categoryStats.value || []).length > 0)
const hasActivityData = computed(() => (activity.value?.data || []).length > 0)
const intentActions = computed(() => {
  const series = trend.value?.series || {}
  return ['like', 'cart', 'buy'].reduce(
    (sum, key) => sum + (series[key] || []).reduce((total, item) => total + Number(item || 0), 0),
    0
  )
})

const normalizeHeatPoint = (item) => {
  if (Array.isArray(item)) {
    return {
      x: Number(item[0] || 0),
      y: Number(item[1] || 0),
      value: Number(item[2] || 0)
    }
  }

  return {
    x: Number(item?.hour_index ?? item?.hour ?? item?.x ?? 0),
    y: Number(item?.day_index ?? item?.day ?? item?.y ?? 0),
    value: Number(item?.count ?? item?.value ?? 0)
  }
}

const peakActivityLabel = computed(() => {
  const points = (activity.value?.data || []).map(normalizeHeatPoint)
  if (!points.length) return '待生成'

  const peak = points.reduce((best, current) => (current.value > best.value ? current : best), points[0])
  const day = activity.value?.days?.[peak.y] || '近期'
  const rawHour = activity.value?.hours?.[peak.x]

  if (rawHour === undefined || rawHour === null || rawHour === '') {
    return `${day} 全天`
  }

  const hourLabel =
    typeof rawHour === 'number' || /^\d+$/.test(String(rawHour))
      ? `${String(rawHour).padStart(2, '0')}:00`
      : String(rawHour)

  return `${day} ${hourLabel}`
})

const leadProductLabel = computed(() => hotProducts.value?.[0]?.name || '等待成交数据')
const leadProductMeta = computed(() => {
  const leadProduct = hotProducts.value?.[0]
  if (!leadProduct) return '热门商品出现后，这里会自动补充成交与浏览焦点。'

  const signal = Number(leadProduct.buy_count || leadProduct.like_count || leadProduct.view_count || 0)
  return `${formatCount(signal)} 次关键行为，当前最值得优先跟进。`
})

const fetchAll = async () => {
  loading.value = true

  try {
    const token = localStorage.getItem('admin_token')
    const headers = { Authorization: token }
    const params = { range: range.value }

    const [overviewRes, trendRes, hotRes, categoryRes, activityRes] = await Promise.all([
      request.get('/api/admin/analytics/overview', { params, headers }),
      request.get('/api/admin/analytics/behavior-trend', { params, headers }),
      request.get('/api/admin/analytics/hot-products', { params, headers }),
      request.get('/api/admin/analytics/category-stats', { params, headers }),
      request.get('/api/admin/analytics/user-activity', { params, headers })
    ])

    overview.value = overviewRes || {}
    trend.value = trendRes || { dates: [], series: { view: [], like: [], cart: [], buy: [] } }
    hotProducts.value = hotRes?.list || hotRes || []
    categoryStats.value = categoryRes?.list || categoryRes || []
    activity.value = activityRes || { days: [], hours: [], data: [], max_count: 0 }
    lastUpdatedAt.value = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  } catch (error) {
    console.error('Analytics fetch error:', error)
    ElMessage.error('分析数据加载失败')
  } finally {
    loading.value = false
    await nextTick()
    renderTrend()
    renderPie()
    renderHeat()
  }
}

const ensureChart = (instance, elRef) => {
  const element = elRef.value
  if (!element) return null

  if (instance && !instance.isDisposed?.() && instance.getDom() === element) return instance
  if (instance && !instance.isDisposed?.()) instance.dispose()

  return echarts.init(element)
}

const renderTrend = () => {
  trendChart = ensureChart(trendChart, trendEl)
  if (!trendChart) return

  const dates = trend.value?.dates || []
  const series = trend.value?.series || {}

  if (!dates.length || !['view', 'like', 'cart', 'buy'].some((key) => (series[key] || []).length > 0)) {
    trendChart.clear()
    return
  }

  trendChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { top: 8, textStyle: { color: '#4d5563' } },
    grid: { left: 24, right: 20, top: 46, bottom: 24, containLabel: true },
    xAxis: { type: 'category', data: dates, axisLine: { lineStyle: { color: '#d6dbe3' } } },
    yAxis: { type: 'value', axisLine: { show: false }, splitLine: { lineStyle: { color: '#edf0f6' } } },
    series: [
      { name: '浏览', type: 'line', smooth: true, data: series.view || [], lineStyle: { color: '#111111' }, itemStyle: { color: '#111111' } },
      { name: '收藏', type: 'line', smooth: true, data: series.like || [], lineStyle: { color: '#f08c00' }, itemStyle: { color: '#f08c00' } },
      { name: '加购', type: 'line', smooth: true, data: series.cart || [], lineStyle: { color: '#e85d04' }, itemStyle: { color: '#e85d04' } },
      { name: '成交', type: 'line', smooth: true, data: series.buy || [], lineStyle: { color: '#2f9e44' }, itemStyle: { color: '#2f9e44' } }
    ]
  })
}

const renderPie = () => {
  pieChart = ensureChart(pieChart, pieEl)
  if (!pieChart) return

  const pieData = (categoryStats.value || []).map((item) => ({
    name: item.category || '其他',
    value: Number(item.buy_count || 0)
  }))

  if (!pieData.length) {
    pieChart.clear()
    return
  }

  pieChart.setOption({
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { color: '#4d5563' } },
    series: [
      {
        name: '分类成交',
        type: 'pie',
        radius: ['36%', '64%'],
        center: ['50%', '46%'],
        avoidLabelOverlap: true,
        itemStyle: { borderColor: '#fff', borderWidth: 2 },
        data: pieData,
        color: ['#111111', '#f08c00', '#e67700', '#2f9e44', '#495057', '#fab005', '#fa9f42']
      }
    ]
  })
}

const renderHeat = () => {
  heatChart = ensureChart(heatChart, heatEl)
  if (!heatChart) return

  const data = activity.value?.data || []

  if (!data.length) {
    heatChart.clear()
    return
  }

  heatChart.setOption({
    tooltip: { position: 'top' },
    grid: { left: 58, right: 20, top: 16, bottom: 28 },
    xAxis: { type: 'category', data: activity.value?.hours || [], splitArea: { show: true } },
    yAxis: { type: 'category', data: activity.value?.days || [], splitArea: { show: true } },
    visualMap: {
      min: 0,
      max: Math.max(Number(activity.value?.max_count || 0), 5),
      calculable: true,
      orient: 'horizontal',
      left: 'center',
      bottom: -2,
      inRange: { color: ['#fff7e6', '#ffd8a8', '#ff922b', '#d9480f'] }
    },
    series: [
      {
        name: '活跃度',
        type: 'heatmap',
        data,
        emphasis: { itemStyle: { shadowBlur: 8, shadowColor: 'rgba(0, 0, 0, 0.24)' } }
      }
    ]
  })
}

const exportChart = (chart, name) => {
  if (!chart) {
    ElMessage.warning('图表未就绪')
    return
  }

  const url = chart.getDataURL({ type: 'png', pixelRatio: 2, backgroundColor: '#ffffff' })
  const link = document.createElement('a')
  link.href = url
  link.download = `${name}-${Date.now()}.png`
  link.click()
}

const handleResize = () => {
  trendChart?.resize()
  pieChart?.resize()
  heatChart?.resize()
}

watch(range, fetchAll)

onMounted(() => {
  fetchAll()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  pieChart?.dispose()
  heatChart?.dispose()
})
</script>

<style scoped lang="scss">
.analytics-page {
  min-height: calc(100vh - 132px);
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background:
    radial-gradient(circle at 10% 8%, rgba(255, 214, 102, 0.18), transparent 30%),
    radial-gradient(circle at 92% 4%, rgba(191, 219, 254, 0.22), transparent 26%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.52), rgba(245, 247, 250, 0.96));
}

.analytics-hero {
  margin-top: 0;
  padding: 18px;
  background:
    linear-gradient(135deg, rgba(255, 248, 230, 0.92) 0%, rgba(255, 255, 255, 0.86) 50%, rgba(240, 249, 255, 0.9) 100%);
  border: 1px solid rgba(255, 255, 255, 0.95);
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.08);
}

.analytics-hero__top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
}

.title-wrap {
  max-width: 760px;
}

.title-wrap__eyebrow {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: rgba(17, 17, 17, 0.06);
  color: #5a6472;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.title-wrap h2 {
  margin: 10px 0 0;
  font-size: 30px;
  color: #171c24;
  letter-spacing: -0.04em;
}

.title-wrap p {
  margin: 10px 0 0;
  color: #5f6b79;
  font-size: 14px;
  line-height: 1.7;
  font-weight: 600;
}

.analytics-hero__actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
}

.range-note {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  padding: 10px 14px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.74);
  border: 1px solid rgba(17, 17, 17, 0.06);
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.05);
}

.range-note__label {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #7b8794;
}

.range-note strong {
  font-size: 15px;
  color: #171c24;
}

.range-note small {
  color: #748091;
  font-size: 12px;
}

.range-switch {
  display: inline-flex;
  gap: 6px;
  background: rgba(255, 255, 255, 0.84);
  border: 1px solid rgba(17, 17, 17, 0.06);
  border-radius: 999px;
  padding: 4px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.range-btn {
  border: none;
  background: transparent;
  color: #5b6472;
  border-radius: 999px;
  min-height: 36px;
  padding: 0 18px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.22s cubic-bezier(0.22, 1, 0.36, 1);
}

.range-btn.active {
  background: #111111;
  color: #ffd43b;
  box-shadow: 0 10px 20px rgba(17, 17, 17, 0.16);
}

.insight-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.insight-card,
.loading-card,
.loading-panel,
.panel,
.kpi-card {
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.92);
  border-radius: 18px;
  backdrop-filter: blur(14px);
}

.insight-card {
  padding: 16px 18px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.04);
}

.insight-card__label {
  display: block;
  font-size: 12px;
  font-weight: 800;
  color: #7a8697;
  letter-spacing: 0.04em;
}

.insight-card strong {
  display: block;
  margin-top: 8px;
  font-size: 20px;
  line-height: 1.2;
  color: #141a22;
}

.insight-card p {
  margin: 8px 0 0;
  color: #647180;
  font-size: 13px;
  line-height: 1.6;
}

.loading-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading-kpis,
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.loading-split,
.mid-grid {
  display: grid;
  grid-template-columns: 1.35fr 1fr;
  gap: 12px;
}

.loading-card,
.loading-panel {
  position: relative;
  overflow: hidden;
}

.loading-card::after,
.loading-panel::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(110deg, transparent 24%, rgba(255, 255, 255, 0.62) 50%, transparent 76%);
  transform: translateX(-120%);
  animation: loading-shimmer 1.6s ease-in-out infinite;
}

.loading-card {
  min-height: 88px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.loading-card__icon,
.loading-line,
.loading-pill,
.loading-chart,
.loading-table__row {
  background: rgba(148, 163, 184, 0.18);
}

.loading-card__icon {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  flex-shrink: 0;
}

.loading-card__meta {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.loading-panel {
  min-height: 300px;
  padding: 16px;
}

.loading-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.loading-line {
  display: block;
  height: 12px;
  border-radius: 999px;
}

.loading-line.short {
  width: 38%;
}

.loading-line.medium {
  width: 64%;
  height: 18px;
}

.loading-pill {
  width: 84px;
  height: 34px;
  border-radius: 999px;
}

.loading-chart {
  width: 100%;
  border-radius: 16px;
}

.loading-chart--trend {
  height: 290px;
}

.loading-chart--pie {
  height: 260px;
}

.loading-chart--heat {
  height: 300px;
}

.loading-table {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading-table__row {
  display: block;
  height: 38px;
  border-radius: 12px;
}

@keyframes loading-shimmer {
  100% {
    transform: translateX(120%);
  }
}

.kpi-card {
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
}

.kpi-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.kpi-icon.amber {
  background: #ffe8a1;
}

.kpi-icon.gray {
  background: #e9ecef;
}

.kpi-icon.mint {
  background: #d3f9d8;
}

.kpi-icon.clay {
  background: #ffd8a8;
}

.kpi-meta span {
  display: block;
  font-size: 12px;
  color: #6a7281;
  font-weight: 700;
}

.kpi-meta strong {
  display: block;
  margin-top: 4px;
  font-size: 24px;
  color: #141a22;
}

.panel {
  margin-top: 0;
  padding: 16px 18px 18px;
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.06);
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 12px;
}

.panel-head__copy {
  display: flex;
  flex-direction: column;
}

.panel-head__copy h3 {
  margin: 0;
  font-size: 16px;
  color: #1d2430;
}

.panel-head__copy p {
  margin: 4px 0 0;
  color: #6e7a8b;
  font-size: 13px;
  font-weight: 600;
}

.export-btn {
  border: 1px solid rgba(17, 17, 17, 0.08);
  background: rgba(255, 255, 255, 0.88);
  color: #374151;
  border-radius: 999px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 800;
  transition: all 0.2s cubic-bezier(0.22, 1, 0.36, 1);
}

.export-btn:hover {
  transform: translateY(-1px);
  border-color: #111111;
  box-shadow: 0 10px 18px rgba(17, 17, 17, 0.08);
}

.chart-shell {
  position: relative;
  border-radius: 18px;
  overflow: hidden;
}

.chart-shell.is-empty {
  border: 1px dashed #d9dee7;
  background: linear-gradient(180deg, rgba(248, 250, 252, 0.95), rgba(255, 255, 255, 0.9));
}

.chart-empty {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 24px;
  background: rgba(255, 255, 255, 0.82);
}

.chart-empty strong {
  color: #273140;
  font-size: 15px;
}

.chart-empty p {
  margin: 8px 0 0;
  max-width: 320px;
  color: #738091;
  font-size: 13px;
  line-height: 1.6;
}

.chart-box {
  width: 100%;
  height: 320px;
}

.chart-box.compact {
  height: 290px;
}

.chart-box.heat {
  height: 340px;
}

.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  text-align: left;
  padding: 10px 8px;
  border-bottom: 1px dashed #e6eaf2;
  color: #293241;
  font-size: 13px;
}

th {
  color: #6b7280;
  font-size: 12px;
  font-weight: 800;
}

tbody tr:last-child td {
  border-bottom: none;
}

.panel-empty {
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8a93a2;
  font-weight: 700;
  border: 1px dashed #d9dee7;
  border-radius: 16px;
  background: rgba(248, 250, 252, 0.88);
}

@media (max-width: 1160px) {
  .analytics-hero__top {
    flex-direction: column;
    align-items: flex-start;
  }

  .analytics-hero__actions {
    width: 100%;
    align-items: stretch;
  }

  .range-note {
    align-items: flex-start;
  }

  .insight-grid {
    grid-template-columns: 1fr;
  }

  .loading-kpis,
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .loading-split,
  .mid-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .analytics-page {
    padding: 12px;
  }

  .analytics-hero {
    padding: 14px;
  }

  .title-wrap h2 {
    font-size: 26px;
  }

  .analytics-hero__actions,
  .range-note,
  .range-switch {
    width: 100%;
  }

  .range-switch {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .range-btn {
    width: 100%;
  }

  .loading-kpis,
  .kpi-grid {
    grid-template-columns: 1fr;
  }

  .chart-box,
  .chart-box.heat {
    height: 280px;
  }
}
</style>
