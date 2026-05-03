<template>
  <div class="page-wrapper">
    <div class="content-card">
      <div class="card-header">
        <div class="left-panel">
          <div class="page-title-wrap">
            <span class="page-icon" style="background:#fffbeb;color:#f59e0b;"><el-icon><Goods /></el-icon></span>
            <div>
              <h2 class="page-title">商品</h2>
              <span class="subtitle">审核与状态</span>
            </div>
          </div>
        </div>
        <div class="right-panel">
          <div class="search-capsule">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
                v-model="keyword"
                placeholder="搜索商品名称..."
                @keyup.enter="fetchProducts"
                clearable
            />
            <button @click="fetchProducts">搜索</button>
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
            :data="productList"
            v-loading="loading"
            style="width: 100%"
            :header-cell-style="{ background: '#fff', color: '#8c9bae', fontWeight: '600', borderBottom: '1px solid #f0f0f0' }"
        >
          <el-table-column label="商品信息" min-width="320">
            <template #default="scope">
              <div class="product-cell">
                <div class="img-box">
                  <el-image
                      :src="fixUrl(scope.row.image)"
                      :preview-src-list="[fixUrl(scope.row.image)]"
                      fit="cover"
                      preview-teleported
                      class="p-img"
                  >
                    <template #error>
                      <div class="img-err"><el-icon><Picture /></el-icon></div>
                    </template>
                  </el-image>
                </div>
                <div class="p-text">
                  <div class="title" :title="scope.row.name">{{ scope.row.name || scope.row.title || '未知商品' }}</div>
                  <div class="desc">{{ scope.row.description || '暂无描述' }}</div>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="price" label="价格" width="120">
            <template #default="scope">
              <span class="price-tag">¥ {{ scope.row.price }}</span>
            </template>
          </el-table-column>

          <el-table-column label="发布人" width="180">
            <template #default="scope">
              <div class="seller-cell">
                <el-avatar :size="28" :src="scope.row.user?.avatar || defaultAvatar" />
                <span class="name">{{ scope.row.user?.nickname || scope.row.user?.username || '未知用户' }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="120">
            <template #default="scope">
              <span class="status-pill" :class="getStatusClass(scope.row.status)">
                {{ getStatusText(scope.row.status) }}
              </span>
            </template>
          </el-table-column>

          <el-table-column label="商品警告" width="210">
            <template #default="scope">
              <div class="warning-cell" :class="getWarningLevel(scope.row)">
                <IconExclamationCircleFill class="warning-icon" />
                <div class="warning-main">
                  <div class="warning-top">
                    <span>{{ getReportCount(scope.row) }} 次举报</span>
                    <span class="warning-label">{{ getWarningLabel(scope.row) }}</span>
                  </div>
                  <div class="warning-track">
                    <span class="warning-bar" :style="{ width: `${getWarningPercent(scope.row)}%` }"></span>
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="操作" fixed="right" width="140" align="right">
            <template #default="scope">
              <el-tooltip content="强制下架" placement="top" v-if="scope.row.status === 1">
                <div class="icon-btn danger" @click="handleAudit(scope.row, 3)">
                  <el-icon><CloseBold /></el-icon>
                </div>
              </el-tooltip>

              <el-tooltip content="重新上架" placement="top" v-if="scope.row.status === 3">
                <div class="icon-btn success" @click="handleAudit(scope.row, 1)">
                  <el-icon><Check /></el-icon>
                </div>
              </el-tooltip>

              <span v-if="scope.row.status === 2" class="disabled-text">已售出</span>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-footer">
        <el-pagination
            background
            layout="prev, pager, next"
            :total="productList.length"
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
import { ElMessage, ElMessageBox } from '@/ui/feedback'
import { IconExclamationCircleFill } from '@arco-design/web-vue/es/icon'
import { Search, CloseBold, Check, Picture, Refresh } from '@/icons/tw-icons.js'

const loading = ref(false)
const rawProductList = ref([])
const keyword = ref('')
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const productList = computed(() => {
  if (!keyword.value) return rawProductList.value
  const key = keyword.value.toLowerCase()
  return rawProductList.value.filter(p => 
    (p.name && p.name.toLowerCase().includes(key)) || 
    (p.description && p.description.toLowerCase().includes(key))
  )
})

// 修复图片 helper
const fixUrl = (url) => resolveUrl(url)

const fetchProducts = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const res = await request.get('/api/admin/products', {
      headers: { Authorization: token }
    })
    rawProductList.value = (res.data || []).map(p => {
       if (p.user && p.user.avatar) {
         p.user.avatar = resolveUrl(p.user.avatar)
       }
       return p
    })
  } catch (e) {
    console.error(e)
    ElMessage.error('获取商品列表失败')
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
  keyword.value = ''
  fetchProducts()
}

const handleAudit = (row, targetStatus) => {
  const actionText = targetStatus === 3 ? '强制下架' : '重新上架'
  ElMessageBox.confirm(`确定要${actionText}商品“${row.name || '此商品'}”吗？`, '提示', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: targetStatus === 3 ? 'warning' : 'info', center: true, customClass: 'warm-theme-box'
  }).then(async () => {
    const token = localStorage.getItem('admin_token')
    await request.put(`/api/admin/products/${row.id}/audit`, { status: targetStatus }, {
      headers: { Authorization: token }
    })
    ElMessage.success('操作成功')
    // 本地更新状态
    row.status = targetStatus
  }).catch(() => {})
}

const getStatusText = (s) => ({ 1: '在售', 2: '已售', 3: '违规下架' }[s] || '未知')
const getStatusClass = (s) => ({ 1: 'sale', 2: 'sold', 3: 'ban' }[s] || '')
const getReportCount = (row) => Number(row.report_count || 0)
const getWarningLevel = (row) => row.warning_level || (getReportCount(row) <= 5 ? 'green' : getReportCount(row) <= 10 ? 'yellow' : 'red')
const getWarningLabel = (row) => ({ green: '正常', yellow: '关注', red: '高危' }[getWarningLevel(row)] || '正常')
const getWarningPercent = (row) => {
  const count = getReportCount(row)
  if (count <= 0) return 0
  return Math.min(100, Math.max(8, Math.round((count / 15) * 100)))
}

onMounted(fetchProducts)
</script>

<style scoped lang="scss">
@use 'sass:color';

$gray-50: #f9fafb;
$gray-100: #f3f4f6;
$gray-200: #e5e7eb;
$gray-400: #9ca3af;
$gray-500: #6b7280;
$gray-600: #4b5563;
$gray-800: #1f2937;
$gray-900: #111827;
$amber-300: #fcd34d;
$amber-400: #fbbf24;
$amber-500: #f59e0b;
$success: #10b981;
$danger: #ef4444;
$warning: #f59e0b;
$border-light: rgba(17, 24, 39, 0.05);

.page-wrapper { height: 100%; display: flex; flex-direction: column; padding-right: 12px; }

.content-card {
  background: #ffffff; border-radius: 20px; flex: 1; display: flex; flex-direction: column;
  overflow: hidden; border: 1px solid $border-light;
  box-shadow: 0 4px 20px -4px rgba(17, 24, 39, 0.03);
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
      background: $gray-900; color: #fff; border: none;
      padding: 7px 18px; border-radius: 9999px;
      font-weight: 600; cursor: pointer; font-size: 12px;
      font-family: 'Outfit', sans-serif;
      transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
      &:hover { background: #000; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
      &:active { transform: scale(0.96); }
    }
  }
}

.table-container {
  flex: 1; padding: 0 20px; overflow-y: auto;

  &::-webkit-scrollbar { width: 6px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $gray-200; border-radius: 9999px; }
}

.product-cell {
  display: flex; align-items: center; gap: 14px;
  .img-box {
    width: 56px; height: 56px; border-radius: 12px; overflow: hidden;
    border: 1px solid $border-light; flex-shrink: 0;
    background: $gray-50; display: flex; align-items: center; justify-content: center;
    .p-img {
      width: 100%; height: 100%; transition: transform 0.3s ease;
      &:hover { transform: scale(1.08); }
    }
    .img-err { color: $gray-400; font-size: 20px; }
  }
  .p-text {
    flex: 1; min-width: 0;
    .title {
      font-weight: 600; color: $gray-900; font-size: 14px;
      margin-bottom: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
    }
    .desc {
      color: $gray-400; font-size: 12px;
      overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 240px;
    }
  }
}

.price-tag {
  color: $danger; font-weight: 700; font-size: 15px;
  font-family: 'JetBrains Mono', monospace;
}

.seller-cell {
  display: flex; align-items: center; gap: 8px;
  font-size: 13px; color: $gray-500; font-weight: 500;
}

.status-pill {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 12px; border-radius: 9999px;
  font-size: 12px; font-weight: 600;

  &::before {
    content: ''; width: 6px; height: 6px; border-radius: 50%;
  }

  &.sale {
    background: rgba($success, 0.08); color: color.adjust($success, $lightness: -10%);
    &::before { background: $success; }
  }
  &.sold {
    background: $gray-100; color: $gray-600;
    &::before { background: $gray-400; }
  }
  &.ban {
    background: rgba($danger, 0.08); color: color.adjust($danger, $lightness: -10%);
    &::before { background: $danger; }
  }
}

.warning-cell {
  display: flex;
  align-items: center;
  gap: 10px;
  color: $success;

  .warning-icon {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    color: rgba($success, 0.42);
  }

  .warning-main {
    flex: 1;
    min-width: 0;
  }

  .warning-top {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 7px;
    color: $gray-600;
    font-size: 12px;
    font-weight: 700;
  }

  .warning-label {
    color: currentColor;
    flex-shrink: 0;
  }

  .warning-track {
    height: 7px;
    overflow: hidden;
    border-radius: 999px;
    background: $gray-100;
  }

  .warning-bar {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: currentColor;
    transition: width 0.24s ease;
  }

  &.yellow {
    color: $warning;

    .warning-icon {
      color: rgba($warning, 0.74);
    }
  }

  &.red {
    color: $danger;

    .warning-icon {
      color: $danger;
      filter: drop-shadow(0 0 6px rgba($danger, 0.32));
    }
  }
}

.icon-btn {
  display: inline-flex; width: 32px; height: 32px;
  align-items: center; justify-content: center;
  border-radius: 10px; cursor: pointer;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  margin-left: 8px;

  &.danger {
    background: rgba($danger, 0.08); color: $danger;
    &:hover { background: $danger; color: #fff; transform: translateY(-1px); box-shadow: 0 4px 12px rgba($danger, 0.2); }
    &:active { transform: scale(0.94); }
  }
  &.success {
    background: rgba($success, 0.08); color: $success;
    &:hover { background: $success; color: #fff; transform: translateY(-1px); box-shadow: 0 4px 12px rgba($success, 0.2); }
    &:active { transform: scale(0.94); }
  }
}

.refresh-btn {
  width: 38px; height: 38px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  background: $gray-100; color: $gray-500; cursor: pointer;
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1); border: 1px solid transparent;

  &:hover {
    background: #fff; border-color: rgba(17, 24, 39, 0.1);
    color: $gray-900; transform: rotate(90deg);
    box-shadow: 0 4px 12px rgba(17, 24, 39, 0.05);
  }
  
  &:active {
    transform: rotate(90deg) scale(0.94);
  }
}

.disabled-text { color: $gray-400; font-size: 12px; padding-right: 8px; }

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
    font-size: 13px; color: $gray-900;
  }
  tr:hover td.el-table__cell { background: $gray-50 !important; }
  tr { transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1); }
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column; align-items: flex-start; gap: 14px;
    .search-capsule { width: 100%; }
  }
}
</style>
