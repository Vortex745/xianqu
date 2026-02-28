<template>
  <div class="page-wrapper">
    <div class="content-card">
      <div class="card-header">
        <div class="left-panel">
          <h2 class="page-title">商品审核</h2>
          <span class="subtitle">审核用户发布的内容，维护平台环境</span>
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
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, CloseBold, Check, Picture, Refresh } from '@element-plus/icons-vue'

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
const fixUrl = (url) => {
  if (!url) return ''
  if (!url.startsWith('http')) return 'http://127.0.0.1:8081' + url
  return url.replace('localhost', '127.0.0.1')
}

const fetchProducts = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const res = await request.get('/api/admin/products', {
      headers: { Authorization: token }
    })
    rawProductList.value = res.data || []
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

onMounted(fetchProducts)
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

.product-cell {
  display: flex; align-items: center; gap: 16px;
  .img-box {
    width: 60px; height: 60px; border-radius: 12px; overflow: hidden; border: 1px solid #f0f0f0; flex-shrink: 0; background: #f9f9f9; display: flex; align-items: center; justify-content: center;
    .p-img { width: 100%; height: 100%; transition: transform 0.3s; &:hover { transform: scale(1.1); } }
    .img-err { color: #ccc; font-size: 20px; }
  }
  .p-text {
    flex: 1; min-width: 0;
    .title { font-weight: 700; color: #333; font-size: 14px; margin-bottom: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .desc { color: #999; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 240px; }
  }
}

.price-tag { color: #ff4d4f; font-weight: 800; font-size: 16px; font-family: 'DIN', sans-serif; }

.seller-cell { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #666; font-weight: 500; }

.status-pill {
  padding: 4px 12px; border-radius: 99px; font-size: 12px; font-weight: 600;
  &.sale { background: #f0fdf4; color: #166534; border: 1px solid #dcfce7; }
  &.sold { background: #f3f4f6; color: #4b5563; border: 1px solid #e5e7eb; }
  &.ban { background: #fef2f2; color: #991b1b; border: 1px solid #fee2e2; }
}

.icon-btn {
  display: inline-flex; width: 32px; height: 32px; align-items: center; justify-content: center; border-radius: 50%; cursor: pointer; transition: 0.2s; margin-left: 8px;
  &.danger { background: #fff1f0; color: #ff4d4f; &:hover { background: #ff4d4f; color: #fff; } }
  &.success { background: #f6ffed; color: #52c41a; &:hover { background: #52c41a; color: #fff; } }
}

.refresh-btn {
  width: 40px; height: 40px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: #f4f6f8; color: #666; cursor: pointer; transition: 0.3s;
  &:hover { background: #1a1a1a; color: #ffdf5d; transform: rotate(180deg); }
}

.disabled-text { color: #ccc; font-size: 12px; padding-right: 8px; }

.pagination-footer { padding: 20px 32px; border-top: 1px solid #f7f7f7; display: flex; justify-content: flex-end; }
</style>