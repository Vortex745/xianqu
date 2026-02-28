<template>
  <div class="page-wrapper">
    <div class="content-card">
      <div class="card-header">
        <div class="left-panel">
          <h2 class="page-title">用户管理</h2>
          <span class="subtitle">管理平台注册用户及其权限状态</span>
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
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Calendar, Refresh } from '@element-plus/icons-vue'

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
    rawUserList.value = res.data || []
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
/* 通用布局容器 */
.page-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding-right: 12px;
}

.content-card {
  background: #fff;
  border-radius: 20px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 4px 24px rgba(0,0,0,0.02);
}

/* 头部样式 */
.card-header {
  padding: 24px 32px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f7f7f7;

  .left-panel {
    .page-title { margin: 0; font-size: 22px; font-weight: 800; color: #1a1a1a; }
    .subtitle { font-size: 13px; color: #999; margin-top: 4px; display: block; }
  }
  .right-panel { display: flex; align-items: center; gap: 12px; }
  .search-capsule {
    display: flex;
    align-items: center;
    background: #f4f6f8;
    border-radius: 99px;
    padding: 4px 4px 4px 16px;
    width: 320px;
    transition: 0.3s;

    &:focus-within { background: #fff; box-shadow: 0 0 0 2px #ffdf5d; }

    .search-icon { color: #999; margin-right: 8px; }
    input { border: none; background: transparent; outline: none; flex: 1; font-size: 14px; color: #333; }
    button {
      background: #1a1a1a; color: #ffdf5d; border: none; padding: 8px 20px;
      border-radius: 99px; font-weight: bold; cursor: pointer; transition: 0.2s;
      &:hover { transform: scale(1.05); }
    }
  }
}

/* 表格区域 */
.table-container {
  flex: 1;
  padding: 0 24px;
  overflow-y: auto;
}

.user-profile-cell {
  display: flex; align-items: center; gap: 16px;
  .avatar-box {
    position: relative;
    .status-indicator {
      position: absolute; bottom: 2px; right: 2px; width: 10px; height: 10px;
      border-radius: 50%; border: 2px solid #fff;
      &.online { background: #4cd964; box-shadow: 0 0 4px #4cd964; }
      &.offline { background: #ff3b30; }
    }
  }
  .info {
    .nick { font-weight: 700; font-size: 15px; color: #333; }
    .id-row { font-size: 12px; color: #999; margin-top: 4px; display: flex; align-items: center; gap: 4px; .divider { color: #eee; } }
    .admin-tag { transform: scale(0.9); margin-left: 4px; }
  }
}

.status-badge {
  display: inline-block; padding: 4px 12px; border-radius: 6px; font-size: 12px; font-weight: 600;
  &.active { background: #f0fdf4; color: #166534; border: 1px solid #dcfce7; }
  &.banned { background: #fef2f2; color: #991b1b; border: 1px solid #fee2e2; }
}

.date-cell {
  display: flex; align-items: center; gap: 8px; color: #666; font-size: 13px;
}

.refresh-btn {
  width: 40px; height: 40px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  background: #f4f6f8; color: #666; cursor: pointer; transition: 0.3s;
  &:hover { background: #1a1a1a; color: #ffdf5d; transform: rotate(180deg); }
}

.pagination-footer {
  padding: 20px 32px;
  border-top: 1px solid #f7f7f7;
  display: flex;
  justify-content: flex-end;
}
</style>