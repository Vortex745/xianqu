<template>
  <div class="analytics-page">
    <header class="page-head">
      <div class="title-wrap">
        <h2>行为分析</h2>
        <p>看清用户动作</p>
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
    </header>

    <div v-if="loading" class="skeleton-board">
      <el-skeleton :rows="8" animated />
    </div>

    <template v-else>
      <section class="kpi-grid">
        <article class="kpi-card">
          <div class="kpi-icon amber">
            <IconifySymbol icon="lucide:users-round" :size="20" color="#111" />
          </div>
          <div class="kpi-meta">
            <span>日活用户</span>
            <strong>{{ overview.dau || 0 }}</strong>
          </div>
        </article>
        <article class="kpi-card">
          <div class="kpi-icon gray">
            <IconifySymbol icon="lucide:package-search" :size="20" color="#111" />
          </div>
          <div class="kpi-meta">
            <span>在售商品</span>
            <strong>{{ overview.on_sale_count || 0 }}</strong>
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
          <h3>行为趋势</h3>
          <button class="export-btn" @click="exportChart(trendChart, '行为趋势')">导出图片</button>
        </div>
        <div ref="trendEl" class="chart-box"></div>
      </section>

      <section class="mid-grid">
        <article class="panel table-panel">
          <div class="panel-head">
            <h3>热门商品 Top10</h3>
          </div>
          <div v-if="hotProducts.length === 0" class="panel-empty">暂无数据</div>
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
                  <td>{{ item.view_count || 0 }}</td>
                  <td>{{ item.like_count || 0 }}</td>
                  <td>{{ item.buy_count || 0 }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>

        <article class="panel pie-panel">
          <div class="panel-head">
            <h3>分类成交占比</h3>
            <button class="export-btn" @click="exportChart(pieChart, '分类占比')">导出图片</button>
          </div>
          <div ref="pieEl" class="chart-box compact"></div>
        </article>
      </section>

      <section class="panel heat-panel">
        <div class="panel-head">
          <h3>用户活跃热力图</h3>
          <button class="export-btn" @click="exportChart(heatChart, '活跃热力图')">导出图片</button>
        </div>
        <div ref="heatEl" class="chart-box heat"></div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'
import IconifySymbol from '@/components/IconifySymbol.vue'

const loading = ref(true)
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

const fetchAll = async () => {
  loading.value = true
  try {
    const params = { range: range.value }
    const [overviewRes, trendRes, hotRes, categoryRes, activityRes] = await Promise.all([
      request.get('/api/admin/analytics/overview', { params }),
      request.get('/api/admin/analytics/behavior-trend', { params }),
      request.get('/api/admin/analytics/hot-products', { params }),
      request.get('/api/admin/analytics/category-stats', { params }),
      request.get('/api/admin/analytics/user-activity', { params })
    ])

    overview.value = overviewRes || {}
    trend.value = trendRes || { dates: [], series: { view: [], like: [], cart: [], buy: [] } }
    hotProducts.value = hotRes?.list || []
    categoryStats.value = categoryRes?.list || []
    activity.value = activityRes || { days: [], hours: [], data: [], max_count: 0 }

    await nextTick()
    renderTrend()
    renderPie()
    renderHeat()
  } catch (error) {
    ElMessage.error('分析数据加载失败')
  } finally {
    loading.value = false
  }
}

const ensureChart = (instance, elRef) => {
  if (!elRef.value) return null
  if (instance) return instance
  return echarts.init(elRef.value)
}

const renderTrend = () => {
  trendChart = ensureChart(trendChart, trendEl)
  if (!trendChart) return
  const dates = trend.value?.dates || []
  const series = trend.value?.series || {}
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
  background:
    url("data:image/svg+xml,%3Csvg viewBox='0 0 180 180' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.88' numOctaves='3'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.05'/%3E%3C/svg%3E"),
    radial-gradient(circle at 12% 8%, rgba(255, 214, 102, 0.3), transparent 42%),
    radial-gradient(circle at 88% 4%, rgba(201, 230, 210, 0.3), transparent 38%),
    #f5f7fa;
}

.page-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.title-wrap h2 {
  margin: 0;
  font-size: 26px;
  color: #171c24;
  letter-spacing: -0.3px;
}

.title-wrap p {
  margin: 4px 0 0;
  color: #6c7483;
  font-size: 13px;
  font-weight: 700;
}

.range-switch {
  display: inline-flex;
  gap: 8px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid #e4e9f1;
  border-radius: 999px;
  padding: 4px;
}

.range-btn {
  border: none;
  background: transparent;
  color: #5b6472;
  border-radius: 999px;
  height: 34px;
  padding: 0 16px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.22s cubic-bezier(0.22, 1, 0.36, 1);
}

.range-btn.active {
  background: #111111;
  color: #ffd43b;
}

.skeleton-board,
.panel,
.kpi-card {
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.92);
  border-radius: 18px;
  backdrop-filter: blur(14px);
}

.skeleton-board {
  padding: 22px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.kpi-card {
  padding: 14px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.kpi-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.kpi-icon.amber { background: #ffe8a1; }
.kpi-icon.gray { background: #e9ecef; }
.kpi-icon.mint { background: #d3f9d8; }
.kpi-icon.clay { background: #ffd8a8; }

.kpi-meta span {
  display: block;
  font-size: 12px;
  color: #6a7281;
  font-weight: 700;
}

.kpi-meta strong {
  display: block;
  margin-top: 4px;
  font-size: 22px;
  color: #141a22;
}

.panel {
  margin-top: 12px;
  padding: 14px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.panel-head h3 {
  margin: 0;
  font-size: 16px;
  color: #1d2430;
}

.export-btn {
  border: 1px solid #d0d7e2;
  background: #fff;
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

.mid-grid {
  display: grid;
  grid-template-columns: 1.35fr 1fr;
  gap: 12px;
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

.panel-empty {
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8a93a2;
  font-weight: 700;
}

@media (max-width: 1160px) {
  .kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .mid-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .analytics-page {
    padding: 12px;
  }
  .page-head {
    flex-direction: column;
    align-items: flex-start;
  }
  .kpi-grid {
    grid-template-columns: 1fr;
  }
  .chart-box,
  .chart-box.heat {
    height: 280px;
  }
}
</style>

