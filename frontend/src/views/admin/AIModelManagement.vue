<template>
  <div class="page-wrapper">
    <div class="content-card">
      <!-- ═══ 顶栏 ═══ -->
      <div class="card-header">
        <div class="left-panel">
          <h2 class="page-title">AI 模型管理</h2>
          <span class="subtitle">配置、启停各大模型，一键搞定</span>
        </div>
        <div class="right-panel">
          <div class="search-capsule">
            <el-icon class="search-icon"><Search /></el-icon>
            <input v-model="keyword" placeholder="按名称或提供商搜…" />
          </div>
          <el-tooltip content="刷新" placement="top">
            <div class="refresh-btn" @click="fetchList"><el-icon><Refresh /></el-icon></div>
          </el-tooltip>
          <el-button type="primary" class="add-btn" @click="openForm()">
            <el-icon><Plus /></el-icon> 新增模型
          </el-button>
        </div>
      </div>

      <!-- ═══ 表格 ═══ -->
      <div class="table-container">
        <el-table
          :data="filteredList"
          v-loading="loading"
          style="width:100%"
          :header-cell-style="{ background:'#fff', color:'#8c9bae', fontWeight:'600', borderBottom:'1px solid #f0f0f0' }"
          :cell-style="{ borderBottom:'1px solid #f7f7f7' }"
        >
          <el-table-column label="模型" min-width="260">
            <template #default="{ row }">
              <div class="model-cell">
                <div class="provider-badge" :class="'p-' + row.provider">
                  {{ providerLabel(row.provider) }}
                </div>
                <div class="model-info">
                  <span class="model-name">{{ row.model_name }}</span>
                  <span class="model-hint">{{ row.api_key_hint || '****' }}</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="单价" width="140">
            <template #default="{ row }">
              <span class="price-tag">¥{{ row.price_per_k }} <small>/千T</small></span>
            </template>
          </el-table-column>

          <el-table-column label="优先级" width="100" align="center">
            <template #default="{ row }">
              <span class="priority-num">{{ row.priority }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="120" align="center">
            <template #default="{ row }">
              <el-switch
                :model-value="row.status === 1"
                @change="(val) => toggleStatus(row, val)"
                active-color="#52c41a"
                inactive-color="#ddd"
              />
            </template>
          </el-table-column>

          <el-table-column label="操作" width="140" align="right" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openForm(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- ═══ 新增/编辑弹窗 ═══ -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑模型配置' : '新增模型配置'"
      width="560px"
      :close-on-click-modal="false"
      class="model-dialog"
    >
      <el-form :model="form" label-width="110px" :rules="formRules" ref="formRef">
        <el-form-item label="模型提供商" prop="provider">
          <el-select v-model="form.provider" @change="handleProviderChange" placeholder="选一个" style="width:100%">
            <el-option label="DeepSeek" value="deepseek" />
            <el-option label="阿里云百炼" value="aliyun" />
            <el-option label="MiniMax" value="minimax" />
            <el-option label="智谱AI（GLM）" value="zhipu" />
          </el-select>
        </el-form-item>

        <el-form-item label="模型名称" prop="model_name">
          <div style="display: flex; gap: 8px; width: 100%;">
            <el-input
              v-model="form.model_name"
              placeholder="如 deepseek-chat, qwen-max"
              style="flex: 1"
            />
            <el-select 
              v-if="detectedModels && detectedModels.length > 0" 
              v-model="form.model_name" 
              placeholder="选择已探测模型"
              style="width: 150px"
            >
              <el-option 
                v-for="model in detectedModels" 
                :key="model" 
                :label="model" 
                :value="model" 
              />
            </el-select>
          </div>
        </el-form-item>

        <el-form-item label="API 密钥" :prop="isEdit ? '' : 'api_key'">
          <div class="api-key-container">
            <el-input
              v-model="form.api_key"
              type="password"
              show-password
              :placeholder="isEdit ? '留空则不修改' : '填入密钥'"
              style="flex: 1"
            />
            <el-button
              type="success"
              plain
              class="detect-btn"
              :loading="detecting"
              @click="handleDetectModels"
              :disabled="!form.api_key"
            >
              探测模型
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="Base URL">
          <el-input v-model="form.base_url" placeholder="可选，自定义 API 地址" />
        </el-form-item>

        <el-form-item label="状态">
          <el-switch v-model="form.statusBool" active-text="启用" inactive-text="停用" active-color="#52c41a" />
        </el-form-item>

        <!-- 高级功能折叠面板 -->
        <el-collapse class="advanced-collapse" accordion>
          <el-collapse-item title="高级功能配置 (可选)" name="1">
            <el-form-item label="计费单价" prop="price_per_k" class="advanced-item">
              <el-input-number
                v-model="form.price_per_k"
                :min="0"
                :precision="4"
                :step="0.001"
                controls-position="right"
                style="width:100%"
              />
              <span class="unit-label">元 / 千Tokens</span>
            </el-form-item>

            <el-form-item label="优先级" class="advanced-item">
              <el-input-number v-model="form.priority" :min="0" :max="999" controls-position="right" style="width:200px" />
              <span class="unit-label">数值越大越优先</span>
            </el-form-item>
          </el-collapse-item>
        </el-collapse>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEdit ? '保存修改' : '确认创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'

const loading = ref(false)
const list = ref([])
const keyword = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const submitting = ref(false)
const formRef = ref(null)
const detecting = ref(false)
const detectedModels = ref([])


const handleDetectModels = async () => {
  if (!form.value.api_key) return
  if (!form.value.provider) {
    ElMessage.warning('请先选择模型提供商')
    return
  }
  detecting.value = true
  try {
    const res = await request.post('/api/admin/ai-models/detect', {
      provider: form.value.provider,
      api_key: form.value.api_key,
      base_url: form.value.base_url
    }, { headers: { Authorization: token() } })

    detectedModels.value = res.models || []
    if (detectedModels.value.length > 0) {
      ElMessage.success(`探测成功，找到 ${detectedModels.value.length} 个模型`)
      if (!form.value.model_name) {
        form.value.model_name = detectedModels.value[0]
      }
    } else {
      ElMessage.warning('未能探测到可用模型，请手动输入')
    }
  } catch (e) {
    console.warn('探测失败', e)
  } finally {
    detecting.value = false
  }
}

const PROVIDERS = { deepseek: 'DeepSeek', aliyun: '阿里云百炼', minimax: 'MiniMax', zhipu: '智谱AI' }
const providerLabel = (p) => PROVIDERS[p] || p

const defaultForm = () => ({
  provider: '',
  model_name: '',
  api_key: '',
  base_url: '',
  price_per_k: 0,
  priority: 0,
  statusBool: true
})

const handleProviderChange = (val) => {
  const defaultUrls = {
    deepseek: 'https://api.deepseek.com',
    aliyun: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    minimax: 'https://api.minimax.chat/v1',
    zhipu: 'https://open.bigmodel.cn/api/paas/v4'
  }
  if (defaultUrls[val]) {
    form.value.base_url = defaultUrls[val]
  }
}

const form = ref(defaultForm())

const formRules = {
  provider: [{ required: true, message: '请选择提供商', trigger: 'change' }],
  model_name: [{ required: true, message: '请填写模型名称', trigger: 'blur' }],
  api_key: [{ required: true, message: '请填写API密钥', trigger: 'blur' }],
  price_per_k: [{ required: true, message: '请设置单价', trigger: 'blur' }]
}

const filteredList = computed(() => {
  if (!keyword.value) return list.value
  const key = keyword.value.toLowerCase()
  return list.value.filter(m => {
    const modelName = m.model_name ? m.model_name.toLowerCase() : ''
    const providerName = (PROVIDERS[m.provider] || m.provider || '').toLowerCase()
    return modelName.includes(key) || providerName.includes(key)
  })
})

const token = () => localStorage.getItem('admin_token')

const fetchList = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/admin/ai-models', { headers: { Authorization: token() } })
    list.value = res.data || []
  } catch { ElMessage.error('获取列表失败') }
  finally { loading.value = false }
}

const openForm = (row) => {
  if (row) {
    isEdit.value = true
    editId.value = row.ID || row.id
    form.value = {
      provider: row.provider,
      model_name: row.model_name,
      api_key: '',
      base_url: row.base_url || '',
      price_per_k: row.price_per_k,
      priority: row.priority,
      statusBool: row.status === 1
    }
    detectedModels.value = [] // 重置探测结果
  } else {
    isEdit.value = false
    editId.value = null
    form.value = defaultForm()
    detectedModels.value = []
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (formRef.value) {
    let valid = true
    try { valid = await formRef.value.validate() } catch { return }
    if (!valid) return
  }
  submitting.value = true
  const payload = {
    provider: form.value.provider,
    model_name: form.value.model_name,
    api_key: form.value.api_key || undefined,
    base_url: form.value.base_url,
    price_per_k: form.value.price_per_k,
    priority: form.value.priority,
    status: form.value.statusBool ? 1 : 0
  }
  try {
    if (isEdit.value) {
      await request.put(`/api/admin/ai-models/${editId.value}`, payload, { headers: { Authorization: token() } })
      ElMessage.success('更新成功')
    } else {
      await request.post('/api/admin/ai-models', payload, { headers: { Authorization: token() } })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (e) {
    console.warn('操作失败', e)
  } finally { submitting.value = false }
}

const toggleStatus = async (row, val) => {
  const status = val ? 1 : 0
  try {
    await request.put(`/api/admin/ai-models/${row.ID || row.id}/status`, { status }, { headers: { Authorization: token() } })
    row.status = status
  } catch { ElMessage.error('切换失败') }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(`确定删除模型 ${row.model_name}？删了就没了哦`, '危险操作', {
    confirmButtonText: '确认删除', cancelButtonText: '算了', type: 'warning', center: true
  }).then(async () => {
    await request.delete(`/api/admin/ai-models/${row.ID || row.id}`, { headers: { Authorization: token() } })
    ElMessage.success('已删除')
    fetchList()
  }).catch(() => {})
}

onMounted(fetchList)
</script>

<style scoped lang="scss">
.page-wrapper { height:100%; display:flex; flex-direction:column; padding-right:12px; }

.content-card {
  background:#fff; border-radius:20px; flex:1; display:flex; flex-direction:column; overflow:hidden;
  box-shadow:0 4px 24px rgba(0,0,0,0.02);
}

.card-header {
  padding:24px 32px; display:flex; justify-content:space-between; align-items:center;
  border-bottom:1px solid #f7f7f7;
  .left-panel {
    .page-title { margin:0; font-size:22px; font-weight:800; color:#1a1a1a; }
    .subtitle { font-size:13px; color:#999; margin-top:4px; display:block; }
  }
  .right-panel { display:flex; align-items:center; gap:12px; }
}

.search-capsule {
  display:flex; align-items:center; background:#f5f5f5; border-radius:12px; padding:0 14px; height:40px;
  .search-icon { color:#999; margin-right:8px; }
  input { border:none; background:transparent; outline:none; font-size:14px; width:160px; }
}

.refresh-btn {
  width:40px; height:40px; border-radius:12px; background:#f5f5f5; display:flex; align-items:center; justify-content:center;
  cursor:pointer; transition:0.2s; color:#666;
  &:hover { background:#e8e8e8; }
}

.add-btn { border-radius:12px; height:40px; font-weight:600; }

.table-container { flex:1; overflow:auto; padding:0 24px 24px; }

/* ─── 模型单元格 ─── */
.model-cell { display:flex; align-items:center; gap:12px; }
.provider-badge {
  padding:4px 12px; border-radius:8px; font-size:12px; font-weight:700; white-space:nowrap;
  &.p-deepseek { background:#e8f5e9; color:#2e7d32; }
  &.p-aliyun { background:#fff3e0; color:#e65100; }
  &.p-minimax { background:#e3f2fd; color:#1565c0; }
  &.p-zhipu { background:#f3e5f5; color:#7b1fa2; }
}
.model-info {
  display:flex; flex-direction:column; line-height:1.4;
  .model-name { font-weight:700; color:#1a1a1a; font-size:14px; }
  .model-hint { font-size:12px; color:#bbb; font-family:monospace; }
}

.price-tag {
  font-weight:700; color:#1a1a1a; font-size:15px;
  small { font-weight:400; color:#999; font-size:11px; }
}

.priority-num { font-weight:700; color:#666; font-size:14px; }
.desc-text { color:#999; font-size:13px; }

/* ─── 弹窗微调 ─── */
.api-key-container { display: flex; gap: 8px; width: 100%; }
.detect-btn { padding: 0 16px; border-radius: 8px; }
.model-suggest-item { font-size: 13px; font-weight: 500; }

.unit-label { margin-left:8px; color:#999; font-size:12px; }

.advanced-collapse {
  margin-top: 20px;
  border-top: 1px solid #f0f0f0;
  border-bottom: none;
}
:deep(.advanced-collapse .el-collapse-item__header) {
  font-weight: 600;
  color: #666;
  background-color: transparent;
  border-bottom: none;
}
:deep(.advanced-collapse .el-collapse-item__wrap) {
  border-bottom: none;
  background-color: transparent;
}
.advanced-item { margin-bottom: 12px; }

:deep(.model-dialog) {
  border-radius:20px !important;
  .el-dialog__header { padding:20px 24px 10px; font-weight:800; }
  .el-dialog__body { padding:10px 24px 20px; }
}
</style>
