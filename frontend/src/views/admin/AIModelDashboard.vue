<template>
  <div class="page-wrapper">
    <!-- ═══ 顶部汇总卡片 ═══ -->
    <div class="summary-row">
      <div class="sum-card" v-for="item in summaryCards" :key="item.key" :class="item.cls">
        <div class="sum-icon"><el-icon><component :is="item.icon" /></el-icon></div>
        <div class="sum-body">
          <span class="sum-label">{{ item.label }}</span>
          <span class="sum-value">{{ item.value }}</span>
        </div>
      </div>
    </div>

    <!-- ═══ 筛选器 ═══ -->
    <div class="filter-bar">
      <div class="filter-left">
        <h2 class="section-title">用量趋势</h2>
      </div>
      <div class="filter-right">
        <div class="pill-group">
          <button
            v-for="d in [7, 30]" :key="d"
            :class="['pill', { active: days === d }]"
            @click="days = d; fetchData()"
          >近{{ d }}天</button>
        </div>
        <el-select v-model="filterProvider" placeholder="全部提供商" clearable size="small" style="width:140px" @change="fetchData">
          <el-option label="DeepSeek" value="deepseek" />
          <el-option label="阿里云百炼" value="aliyun" />
          <el-option label="MiniMax" value="minimax" />
          <el-option label="智谱AI" value="zhipu" />
        </el-select>
        <el-select v-model="filterApp" placeholder="全部应用" clearable size="small" style="width:140px" @change="fetchData">
          <el-option label="AI客服" value="customer_service" />
          <el-option label="Agent任务" value="agent" />
        </el-select>
      </div>
    </div>

    <!-- ═══ 图表 ═══ -->
    <div class="chart-row">
      <div class="chart-card">
        <h3 class="chart-title">Tokens 消耗量</h3>
        <div class="chart-body">
          <canvas ref="tokenChartRef"></canvas>
        </div>
      </div>
      <div class="chart-card">
        <h3 class="chart-title">成本趋势 (元)</h3>
        <div class="chart-body">
          <canvas ref="costChartRef"></canvas>
        </div>
      </div>
    </div>

    <!-- ═══ 明细表格 ═══ -->
    <div class="detail-card">
      <h3 class="detail-title">每日明细</h3>
      <el-table :data="tableData" style="width:100%" :header-cell-style="{ background:'#fafafa', color:'#8c9bae', fontWeight:600 }">
        <el-table-column prop="date" label="日期" width="130" />
        <el-table-column label="提供商" width="120">
          <template #default="{ row }">
            <span class="provider-dot" :class="'p-' + row.provider">{{ providerLabel(row.provider) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="model_name" label="模型" width="160" />
        <el-table-column label="应用" width="120">
          <template #default="{ row }">
            {{ row.app_type === 'customer_service' ? 'AI客服' : 'Agent任务' }}
          </template>
        </el-table-column>
        <el-table-column label="Tokens" width="140" align="right">
          <template #default="{ row }">
            <span class="num-cell">{{ formatNum(row.total_tokens) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="成本(元)" width="120" align="right">
          <template #default="{ row }">
            <span class="cost-cell">¥{{ row.total_cost?.toFixed(4) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="调用次数" width="100" align="right">
          <template #default="{ row }">{{ row.call_count }}</template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import { Coin, DataLine, Stopwatch } from '@element-plus/icons-vue'

const days = ref(7)
const filterProvider = ref('')
const filterApp = ref('')
const loading = ref(false)
const dashData = ref([])
const summaryData = ref({ total_tokens: 0, total_cost: 0, total_calls: 0 })

const tokenChartRef = ref(null)
const costChartRef = ref(null)
let tokenChart = null
let costChart = null

const PROVIDERS = { deepseek: 'DeepSeek', aliyun: '阿里云百炼', minimax: 'MiniMax', zhipu: '智谱AI' }
const providerLabel = (p) => PROVIDERS[p] || p
const COLORS = { deepseek: '#2e7d32', aliyun: '#e65100', minimax: '#1565c0', zhipu: '#7b1fa2' }
const formatNum = (n) => n ? n.toLocaleString() : '0'

const token = () => localStorage.getItem('admin_token')

const summaryCards = computed(() => [
  { key: 'tokens', label: '总 Tokens', value: formatNum(summaryData.value.total_tokens), icon: 'DataLine', cls: 'sc-blue' },
  { key: 'cost', label: '总成本', value: '¥' + (summaryData.value.total_cost || 0).toFixed(2), icon: 'Coin', cls: 'sc-orange' },
  { key: 'calls', label: '调用次数', value: formatNum(summaryData.value.total_calls), icon: 'Stopwatch', cls: 'sc-green' }
])

const tableData = computed(() => {
  return (dashData.value || []).map(item => ({
    ...item,
    date: item.date ? item.date.substring(0, 10) : ''
  }))
})

const fetchData = async () => {
  loading.value = true
  try {
    const params = { days: days.value }
    if (filterProvider.value) params.provider = filterProvider.value
    if (filterApp.value) params.app_type = filterApp.value
    const res = await request.get('/api/admin/ai-models/dashboard', {
      params,
      headers: { Authorization: token() }
    })
    dashData.value = res.items || []
    summaryData.value = res.summary || { total_tokens: 0, total_cost: 0, total_calls: 0 }
    await nextTick()
    renderCharts()
  } catch (e) {
    console.error(e)
    ElMessage.error('获取看板数据失败')
  } finally { loading.value = false }
}

/* ─── Canvas 图表（纯手绘，零依赖） ─── */
const renderCharts = () => {
  renderLineChart(tokenChartRef.value, dashData.value, 'total_tokens', 'Tokens')
  renderLineChart(costChartRef.value, dashData.value, 'total_cost', '元')
}

function renderLineChart(canvas, rawData, field, unit) {
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  const dpr = window.devicePixelRatio || 1
  const rect = canvas.parentElement.getBoundingClientRect()
  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr
  ctx.scale(dpr, dpr)
  const W = rect.width, H = rect.height
  ctx.clearRect(0, 0, W, H)

  // 按日期聚合
  const dateMap = {}
  rawData.forEach(item => {
    const d = (item.date || '').substring(0, 10)
    if (!d) return
    if (!dateMap[d]) dateMap[d] = {}
    const prov = item.provider || 'unknown'
    dateMap[d][prov] = (dateMap[d][prov] || 0) + (item[field] || 0)
  })

  const dates = Object.keys(dateMap).sort()
  if (dates.length === 0) {
    ctx.fillStyle = '#ccc'
    ctx.font = '14px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText('暂无数据', W / 2, H / 2)
    return
  }

  // 找到出现过的所有 provider
  const provSet = new Set()
  Object.values(dateMap).forEach(obj => Object.keys(obj).forEach(p => provSet.add(p)))
  const providers = [...provSet]

  // 计算最大值
  let maxVal = 0
  dates.forEach(d => {
    providers.forEach(p => {
      if (dateMap[d][p] > maxVal) maxVal = dateMap[d][p]
    })
  })
  if (maxVal === 0) maxVal = 1

  // 画布边距
  const padL = 60, padR = 20, padT = 20, padB = 40
  const cw = W - padL - padR
  const ch = H - padT - padB

  // 网格线
  ctx.strokeStyle = '#f0f0f0'
  ctx.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const y = padT + ch * (1 - i / 4)
    ctx.beginPath(); ctx.moveTo(padL, y); ctx.lineTo(W - padR, y); ctx.stroke()
    ctx.fillStyle = '#bbb'
    ctx.font = '11px sans-serif'
    ctx.textAlign = 'right'
    const label = field === 'total_cost' ? (maxVal * i / 4).toFixed(2) : Math.round(maxVal * i / 4)
    ctx.fillText(label, padL - 8, y + 4)
  }

  // X 轴标签
  ctx.textAlign = 'center'
  ctx.fillStyle = '#999'
  ctx.font = '11px sans-serif'
  const step = Math.max(1, Math.floor(dates.length / 8))
  dates.forEach((d, i) => {
    if (i % step === 0 || i === dates.length - 1) {
      const x = padL + (i / (dates.length - 1 || 1)) * cw
      ctx.fillText(d.substring(5), x, H - padB + 18)
    }
  })

  // 画折线
  providers.forEach(prov => {
    const color = COLORS[prov] || '#888'
    ctx.strokeStyle = color
    ctx.lineWidth = 2.5
    ctx.lineJoin = 'round'
    ctx.beginPath()
    dates.forEach((d, i) => {
      const val = dateMap[d][prov] || 0
      const x = padL + (i / (dates.length - 1 || 1)) * cw
      const y = padT + ch * (1 - val / maxVal)
      if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y)
    })
    ctx.stroke()

    // 数据点
    dates.forEach((d, i) => {
      const val = dateMap[d][prov] || 0
      const x = padL + (i / (dates.length - 1 || 1)) * cw
      const y = padT + ch * (1 - val / maxVal)
      ctx.fillStyle = '#fff'
      ctx.beginPath(); ctx.arc(x, y, 4, 0, Math.PI * 2); ctx.fill()
      ctx.strokeStyle = color; ctx.lineWidth = 2
      ctx.beginPath(); ctx.arc(x, y, 4, 0, Math.PI * 2); ctx.stroke()
    })
  })

  // 图例
  const legendY = padT - 4
  let legendX = padL
  providers.forEach(prov => {
    const label = PROVIDERS[prov] || prov
    ctx.fillStyle = COLORS[prov] || '#888'
    ctx.fillRect(legendX, legendY - 8, 12, 12)
    ctx.fillStyle = '#666'
    ctx.font = '12px sans-serif'
    ctx.textAlign = 'left'
    ctx.fillText(label, legendX + 16, legendY + 2)
    legendX += ctx.measureText(label).width + 36
  })
}

onMounted(fetchData)
</script>

<style scoped lang="scss">
.page-wrapper {
  height:100%; overflow-y:auto; padding-right:12px;
  display:flex; flex-direction:column; gap:20px;
}

/* ─── 汇总卡片 ─── */
.summary-row { display:flex; gap:16px; }
.sum-card {
  flex:1; background:#fff; border-radius:16px; padding:20px 24px; display:flex; align-items:center; gap:16px;
  box-shadow:0 2px 12px rgba(0,0,0,0.03);
  .sum-icon {
    width:48px; height:48px; border-radius:14px; display:flex; align-items:center; justify-content:center; font-size:22px;
  }
  .sum-body { display:flex; flex-direction:column; }
  .sum-label { font-size:13px; color:#999; }
  .sum-value { font-size:22px; font-weight:800; color:#1a1a1a; margin-top:2px; }
  &.sc-blue .sum-icon { background:#e3f2fd; color:#1565c0; }
  &.sc-orange .sum-icon { background:#fff3e0; color:#e65100; }
  &.sc-green .sum-icon { background:#e8f5e9; color:#2e7d32; }
}

/* ─── 筛选栏 ─── */
.filter-bar {
  display:flex; justify-content:space-between; align-items:center;
  .section-title { margin:0; font-size:18px; font-weight:800; color:#1a1a1a; }
  .filter-right { display:flex; align-items:center; gap:12px; }
}

.pill-group {
  display:flex; background:#f5f5f5; border-radius:10px; overflow:hidden;
  .pill {
    padding:6px 16px; border:none; background:transparent; cursor:pointer; font-size:13px; font-weight:600; color:#999;
    transition:0.2s;
    &.active { background:#1a1a1a; color:#fff; border-radius:10px; }
  }
}

/* ─── 图表 ─── */
.chart-row { display:flex; gap:16px; }
.chart-card {
  flex:1; background:#fff; border-radius:16px; padding:20px 24px; box-shadow:0 2px 12px rgba(0,0,0,0.03);
  .chart-title { margin:0 0 12px; font-size:15px; font-weight:700; color:#333; }
  .chart-body { height:240px; position:relative; }
  canvas { width:100%; height:100%; display:block; }
}

/* ─── 明细表格 ─── */
.detail-card {
  background:#fff; border-radius:16px; padding:20px 24px; box-shadow:0 2px 12px rgba(0,0,0,0.03);
  .detail-title { margin:0 0 16px; font-size:15px; font-weight:700; color:#333; }
}

.provider-dot {
  font-size:12px; font-weight:700; padding:2px 8px; border-radius:6px;
  &.p-deepseek { background:#e8f5e9; color:#2e7d32; }
  &.p-aliyun { background:#fff3e0; color:#e65100; }
  &.p-minimax { background:#e3f2fd; color:#1565c0; }
  &.p-zhipu { background:#f3e5f5; color:#7b1fa2; }
}

.num-cell { font-weight:700; color:#1a1a1a; }
.cost-cell { font-weight:700; color:#e65100; }
</style>
