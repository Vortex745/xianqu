<template>
  <div class="profile-page">
    <div class="profile-header-bg"></div>

    <div class="container">
      <div class="profile-nav animate-up">
        <div class="back-btn glass-effect" @click="$router.push('/')">
          <el-icon><ArrowLeft /></el-icon> 返回首页
        </div>
      </div>

      <div class="main-layout">
        <div class="left-sidebar animate-up">
          <div class="user-card">
            <div class="avatar-section">
              <el-avatar :size="100" :src="userInfo.avatar || defaultAvatar" class="user-avatar" />
              <div class="edit-avatar-btn" @click="triggerFileUpload" v-if="isMe">
                <el-icon><Camera /></el-icon>
              </div>
              <input v-if="isMe" type="file" ref="fileInput" accept="image/png, image/jpeg, image/jpg" class="hidden-file-input" @change="handleFileChange" />
            </div>

            <h2 class="nickname">{{ userInfo.nickname || userInfo.username }}</h2>
            <div class="username-badge">ID: {{ displayId }}</div>
            <div class="info-list">
              <div class="info-item">
                <el-icon><Message /></el-icon>
                <span>{{ userInfo.email || '暂未绑定邮箱' }}</span>
              </div>
            </div>



            <div class="btn-group" v-if="isMe">
              <button class="btn-primary" @click="openEditProfileModal">编辑资料</button>
              <button class="btn-outline" @click="openPasswordModal">修改密码</button>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-item">
              <div class="num">{{ myProducts.length }}</div>
              <div class="label">发布</div>
            </div>
            <div class="divider"></div>
            <div class="stat-item" v-if="isMe">
              <div class="num">{{ myFavorites.length }}</div>
              <div class="label">收藏</div>
            </div>
          </div>
        </div>

        <div class="right-content animate-up delay-1">
          <div class="content-card">
            <el-tabs v-model="activeTab" class="custom-tabs">
              <el-tab-pane name="products">
                <template #label>
                  <span class="tab-rich-label">
                    <span class="tab-icon-badge"><el-icon><Box /></el-icon></span>
                    <span>{{ isMe ? '我发布的' : 'Ta发布的' }}</span>
                  </span>
                </template>
                <div class="grid-list" v-if="myProducts.length > 0">
                  <div v-for="item in myProducts" :key="item.id" class="mini-product-card" @click="handleProductClick(item.id)">
                    <div class="img-box">
                      <img :src="fixImageUrl(item.image)" loading="lazy" />
                      <div v-if="item.status !== 1" class="status-mask" :class="item.status === 2 ? 'sold' : 'off'">
                        <span>{{ item.status === 2 ? '已售出' : '已下架' }}</span>
                      </div>
                      <div class="manage-hint" v-if="isMe"><el-icon><Edit /></el-icon> 管理宝贝</div>
                    </div>
                    <div class="info">
                      <div class="title">{{ item.name || item.title }}</div>
                      <div class="price-row">
                        <span class="currency">¥</span><span class="amount">{{ item.price }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="empty-panel">
                  <div class="empty-visual"></div>
                  <p class="empty-title">{{ isMe ? '暂无发布任何宝贝' : 'Ta还没发布过宝贝' }}</p>
                  <p class="empty-sub">{{ isMe ? '快去发布你的第一个闲置吧！' : '去看看其他好物吧' }}</p>
                  <button v-if="isMe" class="empty-action" @click="$router.push('/publish')">
                    <el-icon><Plus /></el-icon>
                    去发布
                  </button>
                </div>
              </el-tab-pane>

              <el-tab-pane name="favorites" v-if="isMe">
                <template #label>
                  <span class="tab-rich-label">
                    <span class="tab-icon-badge"><el-icon><Star /></el-icon></span>
                    <span>收藏</span>
                  </span>
                </template>
                <div class="grid-list" v-if="myFavorites.length > 0">
                  <div v-for="item in myFavorites" :key="item.id" class="mini-product-card" @click="goToProduct(item.product?.id)">
                    <div class="img-box">
                      <img :src="fixImageUrl(item.product?.image)" loading="lazy" />
                      <div v-if="item.product?.status !== 1" class="status-mask" :class="item.product?.status === 2 ? 'sold' : 'off'">
                        <span>{{ item.product?.status === 2 ? '已售出' : '已下架' }}</span>
                      </div>
                    </div>
                    <div class="info">
                      <div class="title" :class="{'text-gray': item.product?.status !== 1}">
                        {{ item.product?.name || item.product?.title || '未知商品' }}
                      </div>
                      <div class="price-row">
                        <span class="currency">¥</span><span class="amount">{{ item.product?.price }}</span>
                      </div>
                      <div class="seller-mini">
                        <el-avatar :size="16" :src="defaultAvatar" class="mini-avatar" />
                        <span class="name">{{ getSellerName(item.product) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="empty-panel">
                  <div class="empty-visual"></div>
                  <p class="empty-title">暂无收藏宝贝</p>
                  <p class="empty-sub">遇到喜欢的商品就收藏起来吧！</p>
                  <button class="empty-action ghost" @click="$router.push('/')">
                    <el-icon><Goods /></el-icon>
                    去逛逛
                  </button>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="editProfileVisible"
      title="编辑资料"
      width="460px"
      align-center
      class="custom-dialog premium-modal"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <div class="dialog-body">
        <div class="dialog-avatar-wrapper" @click="triggerFileUpload">
          <div class="avatar-container">
             <el-avatar :size="100" :src="previewAvatar || fixImageUrl(userInfo.avatar) || defaultAvatar" class="dialog-avatar" />
             <div class="avatar-overlay">
               <el-icon><Camera /></el-icon>
             </div>
          </div>
          <p class="avatar-tip">点击更换头像</p>
        </div>

        <el-form :model="profileForm" label-position="top" class="edit-form">
          <el-form-item label="昵称">
            <el-input v-model="profileForm.nickname" placeholder="好的名字更容易被人记住" class="premium-input" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="profileForm.email" placeholder="请输入邮箱地址" class="premium-input" />
          </el-form-item>
          <el-form-item label="验证码" v-if="profileForm.email && profileForm.email !== userInfo.email">
            <div style="display: flex; gap: 10px; width: 100%;">
              <el-input v-model="profileForm.code" placeholder="输入6位验证码" class="premium-input" style="flex: 1;" :maxlength="6" />
              <button type="button" class="btn-primary" :disabled="countdown > 0 || !emailRegex.test(profileForm.email)" style="white-space: nowrap; padding: 0 16px; border-radius: 8px; font-weight: 600;" @click="sendCode">
                {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
              </button>
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer centered-footer">
          <button class="btn-ghost" @click="editProfileVisible = false">取消</button>
          <button class="btn-save" @click="submitProfileEdit" :disabled="isSubmitting">
            {{ isSubmitting ? '保存中...' : '保存修改' }}
          </button>
        </div>
      </template>
    </el-dialog>

    <el-dialog
      v-model="pwdVisible"
      title="修改密码"
      width="460px"
      max-width="92vw"
      align-center
      class="custom-dialog password-modal"
      destroy-on-close
      :close-on-click-modal="false"
    >
      <div class="pwd-tip">
        <el-icon><Lock /></el-icon>
        <span>修改成功后将自动退出并重新登录</span>
      </div>
      <el-form :model="pwdForm" label-width="0" class="pwd-form">
        <el-form-item label="当前密码">
          <el-input v-model="pwdForm.old_password" type="password" placeholder="请输入当前密码" show-password size="large" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.new_password" type="password" placeholder="请输入新密码" show-password size="large" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="pwd-footer">
          <button class="pwd-btn ghost" @click="pwdVisible = false">取消</button>
          <button class="pwd-btn solid" @click="submitPwd">确认修改</button>
        </div>
      </template>
    </el-dialog>

    <!-- 强制设置初始密码弹窗 -->
    <el-dialog
      v-model="forcePwdVisible"
      title="为保障账户安全，请立即设置登录密码"
      width="460px"
      max-width="92vw"
      align-center
      class="custom-dialog password-modal force-pwd-modal"
      :show-close="false"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <el-form :model="forcePwdForm" label-width="0" class="pwd-form mt-4">
        <el-form-item label="新密码">
          <el-input v-model="forcePwdForm.new_password" type="password" placeholder="请输入新密码" show-password size="large" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="forcePwdForm.confirm_password" type="password" placeholder="请再次输入新密码" show-password size="large" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="pwd-footer single-btn-footer">
          <button class="pwd-btn solid w-full" @click="submitForcePwd" :disabled="isSubmittingPwd">
            {{ isSubmittingPwd ? '提交中...' : '确认提交' }}
          </button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, watch } from 'vue'
import request from '@/utils/request'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Camera, Iphone, Edit, ArrowLeft, Goods, Star, Plus, Lock, Box, Message } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const userInfo = ref({})
const myProducts = ref([])
const myFavorites = ref([])
const activeTab = ref('products')
const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'
const isSubmitting = ref(false)

const editProfileVisible = ref(false)
const pwdVisible = ref(false)
const profileForm = reactive({ nickname: '', avatar: '', email: '', code: '' })
const pwdForm = reactive({ old_password: '', new_password: '' })

const isMe = computed(() => {
  const localUser = JSON.parse(localStorage.getItem('user') || '{}')
  // If no route param id, it's me (default /profile)
  if (!route.params.id) return true
  // If route param id matches local user id, it's me
  return String(route.params.id) === String(localUser.id)
})

// ID生成
const displayId = computed(() => {
  const dateStr = userInfo.value.created_at
  const date = dateStr ? new Date(dateStr) : new Date()
  const year = (date && !isNaN(date.getFullYear())) ? date.getFullYear() : new Date().getFullYear()
  const idStr = String(userInfo.value.id || '').padStart(4, '0')
  return `xq${year}${idStr}`
})

// 本地上传相关
const fileInput = ref(null)
const previewAvatar = ref('')
const selectedFile = ref(null)

// 辅助函数：修复图片路径
const fixImageUrl = (url) => {
  if (!url) return ''
  let fixedUrl = url
  if (!fixedUrl.startsWith('http')) {
    fixedUrl = 'http://127.0.0.1:8081' + fixedUrl
  }
  return fixedUrl.replace('localhost', '127.0.0.1')
}

// 获取卖家名字 (收藏列表中使用)
const getSellerName = (product) => {
  if (!product) return '未知'
  const s = product.seller || product.user || {}
  return s.nickname || s.username || '闲趣用户'
}

const triggerFileUpload = () => { fileInput.value.click() }

const handleFileChange = (event) => {
  const file = event.target.files[0]
  if (!file) return
  if (!['image/jpeg', 'image/png', 'image/jpg'].includes(file.type)) { return ElMessage.error('只能上传 JPG/PNG 格式的图片') }
  if (file.size > 5 * 1024 * 1024) { return ElMessage.error('图片大小不能超过 5MB') }

  selectedFile.value = file
  const reader = new FileReader()
  reader.onload = (e) => { previewAvatar.value = e.target.result }
  reader.readAsDataURL(file)

  if (!editProfileVisible.value) openEditProfileModal()
}

// 核心数据获取
const fetchUserData = async () => {
  try {
    const localUser = JSON.parse(localStorage.getItem('user') || '{}')
    const targetId = route.params.id || localUser.id
    
    if (!targetId) { router.push('/'); return }

    if (isMe.value) {
      userInfo.value = localUser
    } else {
      // Visitor mode: use query info or placeholder
      userInfo.value = {
        id: targetId,
        nickname: route.query.nickname || '神秘用户',
        username: route.query.nickname || `User_${targetId}`,
        avatar: route.query.avatar || '',
        email: route.query.email || '未公开',
      }
    }

    // 获取发布的商品
    const resProducts = await request.get('/api/products', { params: { user_id: targetId } })
    myProducts.value = (resProducts.list || []).filter(item => Number(item.user_id) === Number(targetId))
    
    // Refine visitor info if possible from products
    if (!isMe.value && myProducts.value.length > 0) {
       const first = myProducts.value[0]
       const seller = first.seller || first.user
       if (seller) {
         userInfo.value.nickname = seller.nickname || seller.username || userInfo.value.nickname
         userInfo.value.avatar = seller.avatar || userInfo.value.avatar
       }
    }

    // 获取收藏列表 (Only for self)
    if (isMe.value) {
        const resFav = await request.get('/api/user/data?type=favorites')
        myFavorites.value = (resFav.data || []).filter(item => item.product && item.product.id)
    } else {
        myFavorites.value = []
    }

  } catch (e) { console.error(e) }
}

const goToManage = (id) => { router.push(`/publish?id=${id}`) }
const goToProduct = (id) => { if (id) router.push(`/product/${id}`) }

const handleProductClick = (id) => {
  if (isMe.value) {
    goToManage(id)
  } else {
    // Visitor views product logic
    goToProduct(id)
  }
}

watch(() => route.fullPath, () => {
    fetchUserData()
})

const openEditProfileModal = () => {
  Object.assign(profileForm, { nickname: userInfo.value.nickname, email: userInfo.value.email, code: '' })
  if (!selectedFile.value) { previewAvatar.value = fixImageUrl(userInfo.value.avatar) }
  editProfileVisible.value = true
}

const submitProfileEdit = async () => {
  if (isSubmitting.value) return
  isSubmitting.value = true
  try {
    let newAvatarUrl = userInfo.value.avatar
    if (selectedFile.value) {
      const formData = new FormData()
      formData.append('file', selectedFile.value)
      const uploadRes = await request.post('/api/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })

      let uploadedUrl = uploadRes.url
      if (uploadedUrl && !uploadedUrl.startsWith('http')) { uploadedUrl = 'http://127.0.0.1:8081' + uploadedUrl }
      uploadedUrl = uploadedUrl.replace('localhost', '127.0.0.1')
      newAvatarUrl = uploadedUrl
    }

    const updateData = { ...profileForm, avatar: newAvatarUrl }
    
    if (profileForm.email && profileForm.email !== userInfo.value.email) {
      if (!profileForm.code) return ElMessage.warning('请输入验证码')
    }

    await request.put('/api/user/profile', updateData)

    ElMessage.success('保存修改成功')
    const newUser = { ...userInfo.value, ...updateData }
    localStorage.setItem('user', JSON.stringify(newUser))
    userInfo.value = newUser
    selectedFile.value = null
    editProfileVisible.value = false

  } catch (e) { ElMessage.error(e.response?.data?.error || '保存失败，请稍后重试') }
  finally { isSubmitting.value = false }
}

const openPasswordModal = () => { pwdForm.old_password = ''; pwdForm.new_password = ''; pwdVisible.value = true }
const submitPwd = async () => { if (!pwdForm.old_password || !pwdForm.new_password) return ElMessage.warning('旧密码和新密码不能为空'); try { await request.put('/api/user/password', pwdForm); ElMessage.success('修改成功，请重新登录'); localStorage.clear(); window.location.href = '/' } catch (e) { ElMessage.error(e.response?.data?.error || '修改失败，请重试') } }

// 强制设置密码逻辑
const forcePwdVisible = ref(false)
const forcePwdForm = reactive({ new_password: '', confirm_password: '' })
const isSubmittingPwd = ref(false)

const checkForcePassword = () => {
  if (localStorage.getItem('needSetPassword') === 'true' && isMe.value) {
    forcePwdVisible.value = true
  }
}

const submitForcePwd = async () => {
  if (!forcePwdForm.new_password || forcePwdForm.new_password.length < 6) return ElMessage.warning('密码长度至少需要6位')
  if (forcePwdForm.new_password !== forcePwdForm.confirm_password) return ElMessage.warning('两次输入的密码不一致')

  if (isSubmittingPwd.value) return
  isSubmittingPwd.value = true
  try {
    await request.put('/api/user/password', { old_password: '', new_password: forcePwdForm.new_password })
    ElMessage.success('登录密码设置成功')
    localStorage.removeItem('needSetPassword')
    forcePwdVisible.value = false
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '密码设置失败，请重试')
  } finally {
    isSubmittingPwd.value = false
  }
}

const emailRegex = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/
const countdown = ref(0)
let timer = null

const sendCode = async () => {
  if (!emailRegex.test(profileForm.email)) {
    return ElMessage.warning('邮箱格式不正确')
  }
  try {
    await request.post('/api/auth/send-code', { email: profileForm.email })
    ElMessage.success('验证码已发送至您的邮箱，请查收')
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) {
    ElMessage.error(e.response?.data?.error || '发送验证码失败')
  }
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

onMounted(() => {
  fetchUserData()
  checkForcePassword()
})
</script>

<style scoped lang="scss">
$primary: #ffdf5d;
$text-main: #333;
$text-light: #999;
$bg-page: #f6f7f9;

/* 全局布局 */
.profile-page {
  min-height: 100vh;
  background: #fbfbfc;
  padding: 0 0 60px;
  position: relative;
}
.profile-header-bg {
  position: absolute; top: 0; left: 0; width: 100%; height: 320px;
  background: linear-gradient(180deg, rgba(255, 223, 93, 0.45) 0%, rgba(255, 223, 93, 0.1) 60%, transparent 100%);
  z-index: 0;
}
.container { max-width: 1180px; margin: 0 auto; padding: 0 20px; box-sizing: border-box; }
.hidden-file-input { display: none; }

/* ★★★ 新增：返回按钮区域 ★★★ */
.profile-nav {
  margin-bottom: 18px;
  position: relative; z-index: 10;

  .back-btn {
    display: inline-flex; align-items: center; gap: 6px;
    padding: 10px 24px; border-radius: 99px;
    background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(10px);
    font-weight: bold; color: #333; cursor: pointer;
    box-shadow: 0 4px 16px rgba(0,0,0,0.06); transition: 0.3s;
    border: 1px solid #fff;
    margin-top: 20px;

    &:hover { transform: translateX(-4px); background: #fff; box-shadow: 0 6px 16px rgba(0,0,0,0.1); }
  }
}

.main-layout { display: grid; grid-template-columns: 340px 1fr; gap: 28px; align-items: start; position: relative; z-index: 5; }
.left-sidebar { position: sticky; top: 84px; }

/* 左侧：用户信息卡片 */
.user-card {
  background: #fff;
  border-radius: 26px;
  padding: 36px 30px;
  text-align: center;
  box-shadow: 0 14px 28px rgba(22, 30, 43, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.92);

  .avatar-section {
    position: relative; display: inline-block; margin-bottom: 20px; cursor: pointer;
    .user-avatar { border: 4px solid #fff; box-shadow: 0 4px 16px rgba(0,0,0,0.08); transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1); }
    .edit-avatar-btn {
      position: absolute; bottom: 0; right: 0; width: 40px; height: 40px; background: #fff; color: #333; border-radius: 50%; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: 0.2s; border: 1px solid #eee; box-shadow: 0 4px 10px rgba(0,0,0,0.1); font-size: 20px;
      &:hover { background: $primary; border-color: $primary; color: #000; transform: scale(1.1); }
    }
    &:hover .user-avatar { transform: scale(1.05) rotate(-2deg); }
  }

  .nickname { margin: 0 0 8px 0; font-size: 26px; font-weight: 800; color: $text-main; }
  .username-badge { color: #8a93a1; font-size: 13px; margin-bottom: 30px; background: #f5f7fb; padding: 4px 12px; border-radius: 99px; display: inline-block; font-weight: 700; }

  .info-list {
    text-align: left; padding: 0 10px; margin-bottom: 30px;
    .info-item { display: flex; align-items: center; gap: 14px; margin-bottom: 16px; color: #4e5868; font-size: 15px; font-weight: 600; .el-icon { font-size: 20px; color: #b3bac7; } }
  }

  .btn-group {
    display: flex; flex-direction: column; gap: 14px;
    button { width: 100%; height: 50px; border-radius: 999px; font-weight: 800; cursor: pointer; transition: 0.2s; font-size: 16px; display: flex; align-items: center; justify-content: center; letter-spacing: 0.5px; }
    .btn-primary { background: #1a1a1a; border: none; color: $primary; &:hover { background: #000; transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.15); } }
    .btn-outline { background: #fff; border: 2px solid #eee; color: #666; &:hover { border-color: #333; color: #333; background: #fafafa; } }
  }
}

/* 左侧：统计卡片 */
.stat-card {
  background: #fff; border-radius: 22px; padding: 24px; display: flex; justify-content: space-around; align-items: center; box-shadow: 0 10px 20px rgba(23, 31, 44, 0.06); margin-top: 20px; border: 1px solid rgba(255, 255, 255, 0.9);
  .stat-item { text-align: center; cursor: pointer; transition: 0.2s; &:hover { transform: translateY(-3px); } }
  .num { font-size: 24px; font-weight: 900; color: $text-main; font-family: 'Arial Black', sans-serif; }
  .label { font-size: 13px; color: $text-light; margin-top: 6px; font-weight: 600; }
  .divider { width: 1px; height: 36px; background: #eee; }
}

/* 右侧：内容区域 */
.content-card {
  background: linear-gradient(180deg, #ffffff 0%, #fcfdff 100%);
  border-radius: 24px;
  padding: 18px 28px 26px;
  min-height: 500px;
  box-shadow: 0 14px 30px rgba(23, 31, 44, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.86);
}

/* Tab 美化 */
:deep(.custom-tabs .el-tabs__item) { font-size: 16px; height: 54px; color: #9099a8; transition: 0.24s cubic-bezier(0.22, 1, 0.36, 1); &.is-active { font-weight: 800; color: #1f2632; font-size: 18px; } &:hover { color: #566070; } }
:deep(.custom-tabs .el-tabs__active-bar) { background-color: $primary; height: 4px; border-radius: 4px; bottom: 6px; width: 42px !important; margin-left: 18px; } /* 短横线风格 */
:deep(.custom-tabs .el-tabs__nav-wrap::after) { height: 1px; background-color: #eceff4; }
:deep(.custom-tabs .el-tabs__content) { padding-top: 10px; min-height: 430px; }
.custom-tabs :deep(.el-tabs__header) { gap: 12px; }
.custom-tabs :deep(.el-tabs__item) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 132px;
}
.custom-tabs :deep(.el-tabs__item.is-active .tab-rich-label) { color: #f5d64d; }
.custom-tabs :deep(.el-tabs__item.is-active .tab-icon-badge) {
  background: rgba(255, 223, 93, 0.18);
  color: #f5d64d;
}
.custom-tabs :deep(.el-tabs__item:not(.is-active):hover .tab-icon-badge) {
  background: rgba(255, 223, 93, 0.26);
  color: #414b5d;
}
.tab-rich-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 700;
  white-space: nowrap;
}
.tab-icon-badge {
  width: 20px;
  height: 20px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(247, 210, 57, 0.16);
  color: #505a6b;
  font-size: 12px;
}

.empty-panel {
  min-height: 430px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 10px;
  color: #7f8796;

  .empty-visual {
    width: 128px;
    height: 128px;
    border-radius: 26px;
    border: 1px solid #e6eaf2;
    background:
      radial-gradient(circle at 22% 24%, rgba(255, 223, 93, 0.36), rgba(255, 223, 93, 0) 62%),
      linear-gradient(145deg, #fafbfd, #eff2f8);
    box-shadow: inset 0 1px 0 rgba(255,255,255,0.7), 0 12px 24px rgba(14, 24, 45, 0.06);
  }
  .empty-title {
    margin: 2px 0 0;
    font-size: 19px;
    color: #2c3442;
    font-weight: 800;
  }
  .empty-sub {
    margin: 0 0 8px;
    font-size: 13px;
    color: #8a92a1;
    font-weight: 600;
  }
  .empty-action {
    min-width: 116px;
    height: 40px;
    border-radius: 999px;
    border: none;
    background: #1a1a1a;
    color: $primary;
    font-weight: 800;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: 0.2s ease;
    &:hover { transform: translateY(-2px); box-shadow: 0 8px 18px rgba(0,0,0,0.14); }
    &.ghost {
      background: #fff;
      color: #323844;
      border: 1px solid #e7e9ef;
      &:hover { background: #f7f9fd; box-shadow: none; }
    }
  }
}

/* 迷你商品列表 */
.grid-list {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 20px; padding-top: 24px;
}

.mini-product-card {
  border-radius: 16px; overflow: hidden; background: #fff; border: 1px solid #f9f9f9; cursor: pointer; transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1); position: relative;
  &:hover { transform: translateY(-6px); box-shadow: 0 12px 30px rgba(0,0,0,0.06); .manage-hint { opacity: 1; transform: translateY(0); } }

  .img-box {
    width: 100%; aspect-ratio: 1; background: #f9f9f9; position: relative; overflow: hidden;
    img { width: 100%; height: 100%; object-fit: cover; transition: transform 0.5s; }
    &:hover img { transform: scale(1.05); }

    .status-mask {
      position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; backdrop-filter: grayscale(100%) blur(2px);
      span { border: 3px solid #fff; color: #fff; padding: 6px 16px; font-size: 14px; font-weight: 900; transform: rotate(-10deg); border-radius: 8px; letter-spacing: 2px; }
      &.sold { background: rgba(0,0,0,0.6); }
      &.off { background: rgba(100,100,100,0.7); }
    }

    .manage-hint {
      position: absolute; bottom: 0; left: 0; right: 0; background: rgba(255, 223, 93, 0.96);
      color: #1a1a1a; font-size: 13px; padding: 10px; text-align: center; font-weight: bold;
      opacity: 0; transform: translateY(100%); transition: all 0.3s;
      display: flex; align-items: center; justify-content: center; gap: 6px;
    }
  }

  .info {
    padding: 14px;
    .title { font-size: 15px; color: $text-main; margin-bottom: 8px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; font-weight: 500; &.text-gray { color: #aaa; text-decoration: line-through; } }
    .price-row { display: flex; align-items: baseline; gap: 2px; .currency { font-size: 12px; font-weight: bold; color: #ff5000; } .amount { font-size: 18px; font-weight: 800; color: #ff5000; } }
    .seller-mini { display: flex; align-items: center; gap: 6px; margin-top: 8px; font-size: 12px; color: #bbb; .mini-avatar { border: 1px solid #f0f0f0; } }
  }
}

/* 弹窗高级样式优化 */
.premium-modal {
  border-radius: 28px !important;
  box-shadow: 0 24px 48px rgba(0,0,0,0.15) !important;
  overflow: hidden;

  :deep(.el-dialog__header) {
    padding: 18px 24px 12px !important;
    margin: 0;
    .el-dialog__title { font-size: 26px; font-weight: 900; color: #1a1a1a; letter-spacing: 1px; }
  }

  :deep(.el-dialog__body) { padding: 10px 28px 18px !important; overflow-x: hidden !important; overflow-y: auto !important; }
  :deep(.el-dialog__footer) { padding: 0 28px 26px !important; }
}

.dialog-body { width: 100%; max-width: 100%; overflow-x: hidden; }

.dialog-avatar-wrapper {
  display: flex; flex-direction: column; align-items: center; cursor: pointer; margin-bottom: 28px;
  .avatar-container {
    position: relative; width: 100px; height: 100px; border-radius: 50%; overflow: hidden;
    border: 4px solid #fff; box-shadow: 0 8px 24px rgba(255, 223, 93, 0.25); transition: 0.3s;
    .dialog-avatar { width: 100%; height: 100%; display: block; }
    .avatar-overlay {
      position: absolute; inset: 0; background: rgba(0,0,0,0.3);
      display: flex; align-items: center; justify-content: center;
      opacity: 0; transition: 0.3s; color: #fff; font-size: 24px;
    }
  }
  .avatar-tip { margin-top: 10px; font-size: 13px; color: #999; font-weight: 600; transition: 0.3s; }

  &:hover {
    .avatar-container { transform: scale(1.05); border-color: $primary; box-shadow: 0 12px 30px rgba(255, 223, 93, 0.4); }
    .avatar-overlay { opacity: 1; }
    .avatar-tip { color: $primary; }
  }
}

.edit-form {
  :deep(.el-form-item__label) { font-size: 14px; font-weight: 700; color: #444; margin-bottom: 8px !important; }
}

/* Premium Input Style */
:deep(.premium-input .el-input__wrapper) {
  background-color: #f7f8fa;
  border-radius: 12px;
  box-shadow: none !important;
  padding: 10px 16px;
  transition: all 0.2s ease;
  height: 48px;

  &.is-focus, &:hover {
    background-color: #fff;
    box-shadow: 0 0 0 2px $primary !important;
  }
  input { font-weight: 600; color: #333; }
}

.dialog-footer {
  display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
  button {
    height: 48px; border-radius: 14px; font-size: 16px; font-weight: 800;
    cursor: pointer; transition: all 0.2s cubic-bezier(0.25, 0.8, 0.25, 1); border: none;
  }
  .btn-ghost { background: #f5f5f5; color: #666; &:hover { background: #eee; color: #333; transform: scale(0.98); } }
  .btn-save {
    background: #1a1a1a; color: $primary;
    &:hover { background: #000; box-shadow: 0 8px 20px rgba(0,0,0,0.2); transform: translateY(-2px); }
    &:active { transform: scale(0.98); }
    &:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }
  }
}

.password-modal {
  :deep(.el-dialog__body) { padding: 16px 24px 10px !important; overflow-x: hidden !important; overflow-y: auto !important; }
  :deep(.el-dialog__footer) { padding: 0 24px 22px !important; }
}
.pwd-tip {
  margin-bottom: 14px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(255, 223, 93, 0.16);
  color: #4f5a6b;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 700;
}
.pwd-form {
  :deep(.el-form-item__label) {
    color: #5a6372;
    font-size: 13px;
    font-weight: 700;
    margin-bottom: 6px !important;
  }
}
.pwd-footer {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.pwd-btn {
  height: 46px;
  border-radius: 12px;
  border: none;
  font-size: 16px;
  font-weight: 800;
  cursor: pointer;
  transition: 0.2s;
}
.pwd-btn.ghost {
  background: #f2f4f7;
  color: #5a6372;
}
.pwd-btn.ghost:hover {
  background: #e8ebf0;
}
.pwd-btn.solid {
  background: #17191f;
  color: $primary;
}
.pwd-btn.solid:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 18px rgba(15, 19, 28, 0.2);
}

/* 动画 */
.animate-up { animation: fadeInUp 0.6s cubic-bezier(0.25, 0.8, 0.25, 1) backwards; }
.delay-1 { animation-delay: 0.15s; }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(30px); } to { opacity: 1; transform: translateY(0); } }

@media (max-width: 768px) {
  .main-layout { grid-template-columns: 1fr; }
  .left-sidebar { position: static; top: auto; }
  .profile-header-bg { height: 160px; margin-bottom: -80px; }
  .content-card { padding: 14px 14px 18px; min-height: auto; }
  .user-card { padding: 24px 20px; }
  .empty-panel { min-height: 320px; }
  .dialog-footer { grid-template-columns: 1fr; }
  .pwd-footer { grid-template-columns: 1fr; }
  .grid-list { grid-template-columns: repeat(2, 1fr); gap: 12px; }
}
</style>
