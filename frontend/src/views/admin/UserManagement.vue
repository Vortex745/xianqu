<template>
  <div class="page-wrapper">
    <div class="content-card">
      <div class="card-header">
        <div class="left-panel">
          <div class="page-title-wrap">
            <span class="page-icon" style="background:#ecfdf5;color:#10b981;"><el-icon><User /></el-icon></span>
            <div>
              <h2 class="page-title">用户</h2>
              <span class="subtitle">账号与权限</span>
            </div>
          </div>
        </div>
        <div class="right-panel">
          <div class="search-capsule">
            <el-icon class="search-icon"><Search /></el-icon>
            <input
                v-model="keyword"
                placeholder="搜索昵称或账号..."
                @keyup.enter="fetchUsers"
                clearable
            />
            <button @click="fetchUsers">查询</button>
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
            :data="userList"
            v-loading="loading"
            style="width: 100%"
            :header-cell-style="{ background: '#fff', color: '#8c9bae', fontWeight: '600', borderBottom: '1px solid #f0f0f0' }"
            :cell-style="{ borderBottom: '1px solid #f7f7f7' }"
        >
          <el-table-column label="用户" min-width="240">
            <template #default="scope">
              <div class="user-profile-cell">
                <div class="avatar-box">
                  <el-avatar :size="48" :src="scope.row.avatar || defaultAvatar" />
                  <div class="status-indicator" :class="scope.row.status === 1 ? 'online' : 'offline'"></div>
                </div>
                <div class="info">
                  <div class="nick">{{ scope.row.nickname || scope.row.username }}</div>

                  <div class="id-row">
                    ID: {{ scope.row.id }}
                    <span class="divider">|</span>
                    @{{ scope.row.username }}
                    <el-tag v-if="scope.row.role === 'admin'" size="small" type="danger" effect="plain" class="admin-tag">管理员</el-tag>
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column prop="created_at" label="注册日期" width="200">
            <template #default="scope">
              <div class="date-cell">
                <el-icon><Calendar /></el-icon>
                <span>{{ formatDate(scope.row.created_at) }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="账号状态" width="150">
            <template #default="scope">
              <div class="status-badge" :class="scope.row.status === 1 ? 'active' : 'banned'">
                {{ scope.row.status === 1 ? '正常' : '已封禁' }}
              </div>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="120" align="right" fixed="right">
            <template #default="scope">
              <div v-if="scope.row.role !== 'admin'">
                <el-button
                    v-if="scope.row.status === 1"
                    type="danger" link
                    @click="handleStatusChange(scope.row, 0)"
                >封禁</el-button>
                <el-button
                    v-else
                    type="success" link
                    @click="handleStatusChange(scope.row, 1)"
                >解封</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-footer">
        <el-pagination
            background
            layout="prev, pager, next"
            :total="userList.length"
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
import { Search, Calendar, Refresh } from '@/icons/tw-icons.js'

const loading = ref(false)
const rawUserList = ref([])
const keyword = ref('')
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const userList = computed(() => {
  if (!keyword.value) return rawUserList.value
  const key = keyword.value.toLowerCase()
  return rawUserList.value.filter(u =>
    u.username.toLowerCase().includes(key) ||
    (u.nickname && u.nickname.toLowerCase().includes(key))
  )
})

const fetchUsers = async () => {
  loading.value = true
  try {
    const token = localStorage.getItem('admin_token')
    const res = await request.get('/api/admin/users', {
      headers: { Authorization: token }
    })
    rawUserList.value = (res.data || []).map(u => ({
      ...u,
      avatar: resolveUrl(u.avatar)
    }))
  } catch (e) {
    console.error(e)
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleRefresh = () => {
  keyword.value = ''
  fetchUsers()
}

const handleStatusChange = (row, targetStatus) => {
  const action = targetStatus === 0 ? '封禁' : '解封'
  // 确认弹窗
  ElMessageBox.confirm(`确定要${action}用户 ${row.nickname || row.username} 吗？`, '操作确认', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning', center: true, customClass: 'warm-theme-box'
  }).then(async () => {
    const token = localStorage.getItem('admin_token')
    await request.put(`/api/admin/users/${row.id}/status`, { status: targetStatus }, {
      headers: { Authorization: token }
    })
    ElMessage.success('操作成功')
    row.status = targetStatus // 本地直接更新状态，无需刷新列表
  }).catch(() => {})
}

const formatDate = (iso) => {
  if (!iso) return '-'
  return new Date(iso).toLocaleDateString()
}

onMounted(fetchUsers)
</script>

<style scoped lang="scss">
@use 'sass:color';

$gray-50: #fafaf9;
$gray-100: #f5f5f4;
$gray-200: #e7e5e4;
$gray-400: #a8a29e;
$gray-500: #78716c;
$gray-600: #57534e;
$gray-700: #44403c;
$gray-800: #292524;
$gray-900: #1c1917;
$amber-300: #fcd34d;
$amber-400: #fbbf24;
$amber-500: #f59e0b;
$success: #10b981;
$danger: #ef4444;
$border-light: rgba(28, 25, 23, 0.06);

.page-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding-right: 12px;
}

.content-card {
  background: #fff;
  border-radius: 18px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid $border-light;
  box-shadow: 0 1px 3px rgba(28,25,23,0.04), 0 1px 2px rgba(28,25,23,0.02);
}

.card-header {
  padding: 22px 28px;
  display: flex;
  justify-content: space-between;
  align-items: center;
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
      margin: 0;
      font-family: 'Outfit', sans-serif;
      font-size: 22px; font-weight: 700; color: $gray-900;
      letter-spacing: -0.03em;
    }
    .subtitle {
      font-size: 13px; color: $gray-400; margin-top: 4px; display: block;
      font-weight: 500;
    }
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
  flex: 1;
  padding: 0 20px;
  overflow-y: auto;

  &::-webkit-scrollbar { width: 6px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $gray-200; border-radius: 9999px; }
}

.user-profile-cell {
  display: flex; align-items: center; gap: 14px;
  .avatar-box {
    position: relative;
    .status-indicator {
      position: absolute; bottom: 2px; right: 2px; width: 10px; height: 10px;
      border-radius: 50%; border: 2px solid #fff;
      &.online { background: $success; box-shadow: 0 0 6px rgba($success, 0.4); }
      &.offline { background: $danger; }
    }
  }
  .info {
    .nick { font-weight: 600; font-size: 14px; color: $gray-900; }
    .id-row {
      font-size: 12px; color: $gray-400; margin-top: 3px;
      display: flex; align-items: center; gap: 4px;
      font-family: 'JetBrains Mono', monospace;
      .divider { color: $gray-200; }
    }
    .admin-tag { transform: scale(0.9); margin-left: 4px; }
  }
}

.status-badge {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 4px 12px; border-radius: 9999px;
  font-size: 12px; font-weight: 600;

  &::before {
    content: ''; width: 6px; height: 6px; border-radius: 50%;
  }

  &.active {
    background: rgba($success, 0.08); color: color.adjust($success, $lightness: -10%);
    &::before { background: $success; }
  }
  &.banned {
    background: rgba($danger, 0.08); color: color.adjust($danger, $lightness: -10%);
    &::before { background: $danger; }
  }
}

.date-cell {
  display: flex; align-items: center; gap: 8px;
  color: $gray-500; font-size: 13px;
}

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
  padding: 18px 28px;
  border-top: 1px solid $border-light;
  display: flex;
  justify-content: flex-end;
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
