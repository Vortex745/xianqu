<template>
  <div class="manage-page" v-loading="loading">
    <div class="top-nav">
      <div class="container">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item :to="{ path: '/profile' }">个人中心</el-breadcrumb-item>
          <el-breadcrumb-item>管理宝贝</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
    </div>

    <div class="container main-wrapper">
      <div class="left-column">
        <div class="image-preview-box">
          <el-image :src="form.image" fit="contain" class="main-img" />

          <div class="status-overlay" :class="getStatusClass(form.status)">
            <span class="status-text">{{ getStatusText(form.status) }}</span>
          </div>
        </div>

        <div class="control-panel">
          <div class="panel-title">商品状态控制</div>

          <div v-if="form.status === 2" class="locked-state">
            <el-icon class="lock-icon"><Lock /></el-icon>
            <p>宝贝已售出，交易快照已锁定</p>
            <span class="sub-tip">如需重新出售，请发布新商品</span>
          </div>

          <div v-else class="toggle-state">
            <p class="current-status">
              当前状态：
              <span :style="{ color: form.status === 1 ? '#67c23a' : '#f56c6c' }">
                {{ form.status === 1 ? '正在出售中' : '已放入仓库' }}
              </span>
            </p>
            <el-button
                v-if="form.status === 1"
                type="warning" round class="action-btn"
                @click="updateStatus(3)"
            >下架商品</el-button>

            <el-button
                v-if="form.status === 3"
                type="success" round class="action-btn"
                @click="updateStatus(1)"
            >重新上架</el-button>
          </div>
        </div>
      </div>

      <div class="right-column">
        <div class="form-header">
          <h2>编辑商品信息</h2>
          <span class="sub-text">完善信息有助于更快卖出哦</span>
        </div>

        <el-form :model="form" label-position="top" class="edit-form">
          <el-form-item label="商品标题">
            <el-input
                v-model="form.title"
                placeholder="请输入标题"
                size="large"
                :disabled="form.status === 2"
            />
          </el-form-item>

          <el-form-item label="商品描述">
            <el-input
                v-model="form.description"
                type="textarea"
                rows="8"
                placeholder="描述一下宝贝的转手原因、新旧程度..."
                resize="none"
                :disabled="form.status === 2"
            />
          </el-form-item>

          <el-form-item label="出售价格">
            <el-input
                v-model="form.price"
                size="large"
                class="price-input"
                :disabled="form.status === 2"
            >
              <template #prefix>¥</template>
            </el-input>
          </el-form-item>

          <div class="divider"></div>

          <div class="form-actions">
            <template v-if="form.status !== 2">
              <el-button type="primary" color="#ffdf5d" size="large" class="save-btn" @click="saveChanges">
                保存修改
              </el-button>
              <el-button size="large" round @click="$router.go(-1)">取消</el-button>
            </template>

            <template v-else>
              <el-button size="large" round disabled class="disabled-save">无法修改</el-button>
              <el-button size="large" round @click="$router.go(-1)">返回</el-button>
            </template>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const form = reactive({ id: 0, title: '', description: '', price: 0, image: '', status: 1 })

const fetchProduct = async () => {
  loading.value = true
  try {
    const res = await request.get(`/api/products/${route.params.id}`)
    const data = res.data
    const user = JSON.parse(localStorage.getItem('user') || '{}')

    // 权限校验
    if (data.user_id !== user.id) {
      ElMessage.error('您无权管理此商品')
      router.replace('/profile')
      return
    }
    Object.assign(form, data)
  } catch (e) { console.error(e) } finally { loading.value = false }
}

const saveChanges = async () => {
  try {
    await request.put(`/api/products/${form.id}`, {
      title: form.title, description: form.description, price: Number(form.price)
    })
    ElMessage.success('保存成功')
  } catch (e) { ElMessage.error('保存失败') }
}

const updateStatus = (newStatus) => {
  const actionText = newStatus === 1 ? '上架' : '下架'
  ElMessageBox.confirm(`确定要${actionText}该商品吗？`, '提示', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
  }).then(async () => {
    await request.put(`/api/products/${form.id}`, { status: newStatus })
    form.status = newStatus
    ElMessage.success(`商品已${actionText}`)
  })
}

const getStatusText = (s) => ({ 1: '正在出售', 2: '交易完成', 3: '已下架' }[s])
const getStatusClass = (s) => ({ 1: 'on-sale', 2: 'sold-out', 3: 'off-shelf' }[s])

onMounted(fetchProduct)
</script>

<style scoped lang="scss">
$primary: #ffdf5d;
.manage-page { min-height: 100vh; background: #f6f7f9; padding-bottom: 60px; }
.container { max-width: 1100px; margin: 0 auto; padding: 0 20px; }

.top-nav { height: 50px; display: flex; align-items: center; background: #fff; border-bottom: 1px solid #eee; margin-bottom: 30px; }

.main-wrapper {
  display: grid; grid-template-columns: 380px 1fr; gap: 30px;
  background: #fff; padding: 40px; border-radius: 20px; box-shadow: 0 4px 20px rgba(0,0,0,0.03); align-items: start;
}

/* 左侧 */
.left-column {
  .image-preview-box {
    width: 100%; aspect-ratio: 1; background: #f9f9f9; border-radius: 16px; position: relative; overflow: hidden; border: 1px solid #f0f0f0;
    .main-img { width: 100%; height: 100%; }

    .status-overlay {
      position: absolute; inset: 0; display: flex; align-items: center; justify-content: center;
      background: rgba(0,0,0,0); transition: 0.3s;
      .status-text { display: none; font-size: 24px; font-weight: 900; color: #fff; border: 3px solid #fff; padding: 8px 24px; border-radius: 8px; transform: rotate(-15deg); }

      &.sold-out { background: rgba(0,0,0,0.6); .status-text { display: block; } }
      &.off-shelf { background: rgba(0,0,0,0.7); .status-text { display: block; color: #ccc; border-color: #ccc; } }
    }
  }

  .control-panel {
    margin-top: 24px; padding: 24px; background: #fcfcfc; border-radius: 16px; border: 1px solid #f0f0f0; text-align: center;
    .panel-title { font-weight: bold; margin-bottom: 16px; color: #333; font-size: 15px; }

    .locked-state {
      color: #999;
      .lock-icon { font-size: 32px; margin-bottom: 8px; }
      p { font-weight: bold; margin-bottom: 4px; }
      .sub-tip { font-size: 12px; }
    }

    .toggle-state {
      .current-status { margin-bottom: 16px; font-size: 14px; font-weight: bold; }
      .action-btn { width: 100%; padding: 12px 0; font-weight: bold; }
    }
  }
}

/* 右侧表单 */
.right-column {
  padding-left: 20px;
  .form-header { margin-bottom: 30px; h2 { margin: 0 0 6px; font-size: 24px; color: #333; } .sub-text { color: #999; font-size: 14px; } }

  .edit-form {
    :deep(.el-form-item__label) { font-weight: bold; color: #333; }
    :deep(.el-input__wrapper), :deep(.el-textarea__inner) {
      box-shadow: 0 0 0 1px #eee inset; border-radius: 8px; padding: 10px 15px; transition: 0.2s;
      &.is-focus { box-shadow: 0 0 0 1px $primary inset; }
    }
    .price-input { width: 200px; }
  }

  .divider { height: 1px; background: #f0f0f0; margin: 30px 0; }

  .form-actions {
    display: flex; gap: 16px;
    .save-btn { color: #000; font-weight: bold; padding: 12px 40px; }
    .disabled-save { background: #f5f5f5; color: #999; border-color: #eee; }
  }
}
</style>