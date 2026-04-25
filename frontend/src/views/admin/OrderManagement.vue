<template>
  <div class="page-wrapper">
    <div class="content-card">
      <div class="card-header">
        <div class="left-panel">
          <div class="page-title-wrap">
            <span class="page-icon" style="background:#f5f3ff;color:#8b5cf6;"><el-icon><List /></el-icon></span>
            <div>
              <h2 class="page-title">订单</h2>
              <span class="subtitle">交易与状态</span>
            </div>
          </div>
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
import request, { resolveUrl } from '@/utils/request'
import { Search, Refresh } from '@/icons/tw-icons.js'

const loading = ref(false)
const orderList = ref([])
const keyword = ref('')
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const fixUrl = (url) => resolveUrl(url)

const fetchOrders = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const res = await request.get('/api/admin/orders', {
      headers: { Authorization: token }
    })
    orderList.value = (res.data || []).map(o => {
      if (o.user && o.user.avatar) o.user.avatar = resolveUrl(o.user.avatar)
      if (o.seller && o.seller.avatar) o.seller.avatar = resolveUrl(o.seller.avatar)
      return o
    })
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
@use 'sass:color';

$gray-50: #fafaf9;
$gray-100: #f5f5f4;
$gray-200: #e7e5e4;
$gray-400: #a8a29e;
$gray-500: #78716c;
$gray-600: #57534e;
$gray-800: #292524;
$gray-900: #1c1917;
$amber-300: #fcd34d;
$amber-400: #fbbf24;
$amber-500: #f59e0b;
$blue: #3b82f6;
$green: #10b981;
$orange: #f59e0b;
$danger: #ef4444;
$border-light: rgba(28, 25, 23, 0.06);

.page-wrapper { height: 100%; display: flex; flex-direction: column; padding-right: 12px; }

.content-card {
  background: #fff; border-radius: 18px; flex: 1; display: flex; flex-direction: column;
  overflow: hidden; border: 1px solid $border-light;
  box-shadow: 0 1px 3px rgba(28,25,23,0.04), 0 1px 2px rgba(28,25,23,0.02);
}

.card-header {
  padding: 22px 28px; display: flex; justify-content: space-between; align-items: center;
  border-bottom: 1px solid $border-light;

  .left-panel {
    .page-title-wrap {
      display: flex;
      align-items: center;
      gap: 12px;
    }
    .page-icon {
      width: 40px;
      height: 40px;
      border-radius: 12px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      flex-shrink: 0;
    }
    .page-title {
      margin: 0; font-family: 'Outfit', sans-serif;
      font-size: 22px; font-weight: 700; color: $gray-900;
      letter-spacing: -0.03em;
    }
    .subtitle { font-size: 13px; color: $gray-400; margin-top: 4px; display: block; font-weight: 500; }
  }

  .right-panel { display: flex; align-items: center; gap: 10px; }

  .search-capsule {
    display: flex; align-items: center;
    background: $gray-100; border-radius: 9999px;
    padding: 6px 6px 6px 16px; width: 300px;
    border: 1px solid transparent;
    transition: all 0.25s ease;

    &:focus-within {
      background: #fff; border-color: $amber-300;
      box-shadow: 0 0 0 3px rgba($amber-300, 0.15);
    }

    .search-icon { color: $gray-400; margin-right: 8px; font-size: 16px; }
    input {
      border: none; background: transparent; outline: none;
      flex: 1; font-size: 13px; color: $gray-900;
      font-family: 'Outfit', sans-serif;
      &::placeholder { color: $gray-400; }
    }
    button {
      background: $gray-800; color: #fff; border: none;
      padding: 7px 18px; border-radius: 9999px;
      font-weight: 600; cursor: pointer; font-size: 12px;
      font-family: 'Outfit', sans-serif;
      transition: all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1);
      &:hover { background: $gray-900; transform: translateY(-1px); }
      &:active { transform: scale(0.97); }
    }
  }
}

.table-container {
  flex: 1; padding: 0 20px; overflow-y: auto;

  &::-webkit-scrollbar { width: 6px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $gray-200; border-radius: 9999px; }
}

.mono-code {
  font-family: 'JetBrains Mono', monospace;
  color: $gray-600; font-size: 12px;
  background: $gray-50; padding: 3px 8px;
  border-radius: 6px; border: 1px solid $border-light;
}

.p-cell {
  display: flex; align-items: center; gap: 10px;
  .p-thumb {
    width: 38px; height: 38px; border-radius: 8px;
    border: 1px solid $border-light; object-fit: cover;
    background: $gray-50;
  }
  .p-name {
    font-weight: 600; font-size: 13px; color: $gray-900;
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 150px;
  }
}

.flow-chart {
  display: flex; align-items: center; justify-content: center; padding: 0 10px; width: 100%;

  .node {
    position: relative; width: 32px; height: 32px;
    display: flex; flex-direction: column; align-items: center;
    .ava {
      border: 2px solid #fff;
      box-shadow: 0 2px 8px rgba(28,25,23,0.08);
      z-index: 2; position: relative;
    }
    .role-badge {
      position: absolute; top: -2px; right: -4px;
      width: 14px; height: 14px; border-radius: 50%; z-index: 3;
      font-size: 9px; color: #fff;
      display: flex; align-items: center; justify-content: center;
      border: 1px solid #fff;
      &.buy { background: $blue; }
      &.sell { background: $amber-500; }
    }
    .tip-name {
      font-size: 10px; color: $gray-400;
      position: absolute; bottom: -16px; white-space: nowrap;
    }
  }

  .link-line {
    flex: 1; margin: 0 8px; position: relative; height: 24px;
    display: flex; align-items: center; justify-content: center;
    .line-track {
      position: absolute; width: 100%; height: 2px;
      background: $gray-200; z-index: 0; border-radius: 2px;
    }
    .status-bubble {
      position: relative; z-index: 1;
      font-size: 11px; font-weight: 600;
      padding: 3px 10px; border-radius: 9999px;
      background: #fff; border: 1px solid $gray-200;
      box-shadow: 0 1px 3px rgba(28,25,23,0.04);

      &.red {
        color: $danger; border-color: rgba($danger, 0.2);
        background: rgba($danger, 0.04);
      }
      &.orange {
        color: $orange; border-color: rgba($orange, 0.2);
        background: rgba($orange, 0.04);
      }
      &.blue {
        color: $blue; border-color: rgba($blue, 0.2);
        background: rgba($blue, 0.04);
      }
      &.green {
        color: $green; border-color: rgba($green, 0.2);
        background: rgba($green, 0.04);
      }
      &.gray {
        color: $gray-400; background: $gray-50;
        border-color: $gray-200;
      }
    }
  }
}

.money-text {
  font-weight: 700; color: $danger;
  font-family: 'JetBrains Mono', monospace; font-size: 14px;
}

.date-text { color: $gray-400; font-size: 12px; }

.refresh-btn {
  width: 38px; height: 38px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  background: $gray-100; color: $gray-500; cursor: pointer;
  transition: all 0.3s ease; border: 1px solid transparent;

  &:hover {
    background: #fff; border-color: rgba(28,25,23,0.1);
    color: $amber-500; transform: rotate(180deg);
  }
}

.pagination-footer {
  padding: 18px 28px; border-top: 1px solid $border-light;
  display: flex; justify-content: flex-end;
}

:deep(.el-table) {
  th.el-table__cell {
    background: $gray-50 !important; color: $gray-600 !important;
    font-weight: 600; font-size: 11px; text-transform: uppercase;
    letter-spacing: 0.04em; padding: 14px 20px !important;
    border-bottom: 1px solid $border-light !important;
  }
  td.el-table__cell {
    padding: 14px 20px !important;
    border-bottom: 1px solid $border-light !important;
    font-size: 13px; color: $gray-600;
  }
  tr:hover td.el-table__cell { background: $gray-50 !important; }
  tr { transition: background 0.15s ease; }
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column; align-items: flex-start; gap: 14px;
    .search-capsule { width: 100%; }
  }
}
</style>
