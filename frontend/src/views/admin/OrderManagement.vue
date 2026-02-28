<template>
  <div class="page-wrapper">
    <div class="content-card">
      <div class="card-header">
        <div class="left-panel">
          <h2 class="page-title">订单中心</h2>
          <span class="subtitle">监控全平台交易链路与状态</span>
        </div>
        <div class="right-panel">
          <div class="search-capsule">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
                v-model="keyword"
                placeholder="搜索订单号..."
                @keyup.enter="fetchOrders"
                clearable
            />
            <button @click="fetchOrders">搜索</button>
          </div>
          <el-tooltip content="刷新数据" placement="top">
            <div class="refresh-btn" @click="handleRefresh">
              <el-icon><Refresh /></el-icon>
            </div>
          </el-tooltip>
        </div>
      </div>

      <div class="table-container">
        <el-table
            :data="filteredList"
            v-loading="loading"
            style="width: 100%"
            :header-cell-style="{ background: '#fff', color: '#8c9bae', fontWeight: '600', borderBottom: '1px solid #f0f0f0' }"
            :cell-style="{ borderBottom: '1px solid #f7f7f7' }"
        >
          <el-table-column prop="order_no" label="订单号" width="200">
            <template #default="scope">
              <span class="mono-code">{{ scope.row.order_no }}</span>
            </template>
          </el-table-column>

          <el-table-column label="商品" min-width="220">
            <template #default="scope">
              <div class="p-cell">
                <img :src="fixUrl(scope.row.product?.image)" class="p-thumb" />
                <span class="p-name">{{ scope.row.product?.name || scope.row.product?.title || '商品已失效' }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="交易链路" min-width="280">
            <template #default="scope">
              <div class="flow-chart">
                <div class="node">
                  <el-avatar :size="32" :src="scope.row.user?.avatar || defaultAvatar" class="ava" />
                  <span class="role-badge buy">买</span>
                  <div class="tip-name">{{ scope.row.user?.username }}</div>
                </div>

                <div class="link-line">
                  <div class="line-track"></div>
                  <div class="status-bubble" :class="getStatusClass(scope.row.status)">
                    {{ getStatusText(scope.row.status) }}
                  </div>
                </div>

                <div class="node">
                  <el-avatar :size="32" :src="scope.row.seller?.avatar || defaultAvatar" class="ava" />
                  <span class="role-badge sell">卖</span>
                  <div class="tip-name">{{ scope.row.seller?.username || '未知' }}</div>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="金额" width="120" align="center">
            <template #default="scope">
              <span class="money-text">¥{{ scope.row.price }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="created_at" label="创建时间" width="180" align="right">
            <template #default="scope">
              <span class="date-text">{{ formatDate(scope.row.created_at) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-footer">
        <el-pagination
            background
            layout="prev, pager, next"
            :total="filteredList.length"
            :page-size="10"
            disabled
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import request from '@/utils/request'
import { Search, Refresh } from '@element-plus/icons-vue'

const loading = ref(false)
const orderList = ref([])
const keyword = ref('')
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const fixUrl = (url) => {
  if (!url) return ''
  if (!url.startsWith('http')) return 'http://127.0.0.1:8081' + url
  return url.replace('localhost', '127.0.0.1')
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const res = await request.get('/api/admin/orders', {
      headers: { Authorization: token }
    })
    orderList.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
    keyword.value = ''
    fetchOrders()
}

// 前端过滤搜索 (后端未做筛选时使用)
const filteredList = computed(() => {
  if (!keyword.value) return orderList.value
  const key = keyword.value.toLowerCase()
  return orderList.value.filter(o =>
      (o.order_no && o.order_no.toLowerCase().includes(key)) ||
      (o.product?.name && o.product.name.toLowerCase().includes(key))
  )
})

const formatDate = (iso) => iso ? new Date(iso).toLocaleString() : '-'

// 状态映射 (1:待支付, 2:待发货, 3:运输中, 4:已完成, 5:已取消)
const getStatusText = (s) => {
  const map = { 1: '待付款', 2: '待发货', 3: '运输中', 4: '已完成', 5: '已取消' }
  return map[s] || '未知状态'
}

const getStatusClass = (s) => {
  if (s === 1) return 'orange' // 待支付
  if (s === 4) return 'green'  // 已完成
  if (s === 5) return 'gray'   // 已取消
  return 'blue'                // 进行中
}

onMounted(fetchOrders)
</script>

<style scoped lang="scss">
.page-wrapper { height: 100%; display: flex; flex-direction: column; padding-right: 12px; }
.content-card { background: #fff; border-radius: 20px; flex: 1; display: flex; flex-direction: column; overflow: hidden; box-shadow: 0 4px 24px rgba(0,0,0,0.02); }

.card-header {
  padding: 24px 32px; display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #f7f7f7;
  .left-panel { .page-title { margin: 0; font-size: 22px; font-weight: 800; color: #1a1a1a; } .subtitle { font-size: 13px; color: #999; margin-top: 4px; display: block; } }
  .right-panel { display: flex; align-items: center; gap: 12px; }
  .search-capsule {
    display: flex; align-items: center; background: #f4f6f8; border-radius: 99px; padding: 4px 4px 4px 16px; width: 320px; transition: 0.3s;
    &:focus-within { background: #fff; box-shadow: 0 0 0 2px #ffdf5d; }
    .search-icon { color: #999; margin-right: 8px; }
    input { border: none; background: transparent; outline: none; flex: 1; font-size: 14px; color: #333; }
    button { background: #1a1a1a; color: #ffdf5d; border: none; padding: 8px 20px; border-radius: 99px; font-weight: bold; cursor: pointer; transition: 0.2s; &:hover { transform: scale(1.05); } }
  }
}

.table-container { flex: 1; padding: 0 24px; overflow-y: auto; }

.mono-code { font-family: 'Consolas', monospace; color: #666; font-size: 13px; background: #f9fafb; padding: 2px 6px; border-radius: 4px; border: 1px solid #eee; }

.p-cell {
  display: flex; align-items: center; gap: 10px;
  .p-thumb { width: 40px; height: 40px; border-radius: 6px; border: 1px solid #f0f0f0; object-fit: cover; background: #f9f9f9; }
  .p-name { font-weight: 600; font-size: 13px; color: #333; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 150px; }
}

/* 交易流可视化 */
.flow-chart {
  display: flex; align-items: center; justify-content: center; padding: 0 10px; width: 100%;
  .node {
    position: relative; width: 32px; height: 32px; display: flex; flex-direction: column; align-items: center;
    .ava { border: 2px solid #fff; box-shadow: 0 2px 8px rgba(0,0,0,0.1); z-index: 2; position: relative; }
    .role-badge {
      position: absolute; top: -2px; right: -4px; width: 14px; height: 14px; border-radius: 50%; z-index: 3;
      font-size: 9px; color: #fff; display: flex; align-items: center; justify-content: center; border: 1px solid #fff;
      &.buy { background: #3b82f6; }
      &.sell { background: #f59e0b; }
    }
    .tip-name { font-size: 10px; color: #999; position: absolute; bottom: -16px; white-space: nowrap; }
  }

  .link-line {
    flex: 1; margin: 0 8px; position: relative; height: 24px; display: flex; align-items: center; justify-content: center;
    .line-track { position: absolute; width: 100%; height: 2px; background: #e5e7eb; z-index: 0; border-radius: 2px; }
    .status-bubble {
      position: relative; z-index: 1; font-size: 11px; font-weight: bold; padding: 2px 10px; border-radius: 99px; background: #fff; border: 1px solid #e5e7eb; box-shadow: 0 2px 4px rgba(0,0,0,0.02);
      &.red { color: #ef4444; border-color: #fecaca; background: #fef2f2; }
      &.orange { color: #f59e0b; border-color: #fde68a; background: #fffbeb; }
      &.blue { color: #3b82f6; border-color: #bfdbfe; background: #eff6ff; }
      &.green { color: #10b981; border-color: #a7f3d0; background: #ecfdf5; }
      &.gray { color: #9ca3af; background: #f3f4f6; }
    }
  }
}

.money-text { font-weight: 800; color: #ff5000; font-family: 'DIN', sans-serif; font-size: 15px; }
.date-text { color: #9ca3af; font-size: 12px; }

.refresh-btn {
  width: 40px; height: 40px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: #f4f6f8; color: #666; cursor: pointer; transition: 0.3s;
  &:hover { background: #1a1a1a; color: #ffdf5d; transform: rotate(180deg); }
}

.pagination-footer { padding: 20px 32px; border-top: 1px solid #f7f7f7; display: flex; justify-content: flex-end; }
</style>