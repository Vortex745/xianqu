<template>
  <el-dialog
      v-model="visible"
      width="800px"
      :show-close="false"
      class="auth-dialog"
      destroy-on-close
      align-center
      :close-on-click-modal="false"
      max-width="92vw"
  >
    <div class="auth-container">
      <div class="auth-left">
        <div class="brand-home">
          <div class="logo-box">
            <div class="circle-shape"></div>
            <div class="square-shape"></div>
          </div>
          <div class="brand-copy">
            <div class="brand-title">XIANQU</div>
            <div class="brand-sub">闲趣社区二手</div>
          </div>
        </div>
        <div class="slogan">
          <h3>发现闲置的价值</h3>
          <p>连接社区和校园的闲置交易</p>
        </div>
        <div class="circle c1"></div>
        <div class="circle c2"></div>
      </div>

      <div class="auth-right">
        <button type="button" class="close-btn" @click="close">
          <el-icon><Close /></el-icon>
        </button>

        <!-- 登录方式切换 Tab -->
        <div class="auth-tabs">
          <div 
            class="auth-tab" 
            :class="{ active: authMethod === 'email' }" 
            @click="switchAuthMethod('email')"
          >邮箱登录</div>
          <div 
            class="auth-tab" 
            :class="{ active: authMethod === 'password' }" 
            @click="switchAuthMethod('password')"
          >密码登录</div>
        </div>

        <h2 v-if="authMethod === 'password'">{{ isLogin ? '欢迎回来' : '创建账号' }}</h2>
        <p v-if="authMethod === 'password'" class="subtitle">{{ isLogin ? '登录后开始交易' : '注册只需 1 分钟' }}</p>

        <h2 v-if="authMethod === 'email'">快捷登录 / 注册</h2>
        <p v-if="authMethod === 'email'" class="subtitle">未注册邮箱验证后将自动创建账号</p>

        <el-form :model="form" class="auth-form" @submit.prevent>

          <!-- 邮箱登录表单 -->
          <template v-if="authMethod === 'email'">
            <div class="input-group">
              <el-input
                  v-model="form.email"
                  type="email"
                  placeholder="请输入邮箱地址"
                  class="custom-input"
              >
                 <template #prefix>
                   <el-icon><Monitor /></el-icon>
                 </template>
              </el-input>
            </div>
            
            <div class="input-group code-group">
              <el-input
                  v-model="form.code"
                  placeholder="6位验证码"
                  class="custom-input code-input"
                  maxlength="6"
              >
                <template #prefix>
                   <el-icon><Lock /></el-icon>
                 </template>
              </el-input>
              <button 
                type="button" 
                class="send-code-btn" 
                @click="sendCode" 
                :disabled="countdown > 0 || isSendingCode"
              >
                <span v-if="isSendingCode"><el-icon class="is-loading"><Loading /></el-icon> 发送中</span>
                <span v-else-if="countdown > 0">{{ countdown }}s 后重发</span>
                <span v-else>获取验证码</span>
              </button>
            </div>
          </template>

          <!-- 密码登录/注册表单 -->
          <template v-if="authMethod === 'password'">
            <div class="input-group">
              <el-input
                  v-model="form.email"
                  placeholder="请输入邮箱"
                  :prefix-icon="Message"
                  class="custom-input"
              />
            </div>

            <div class="input-group">
              <el-input
                  v-model="form.password"
                  type="password"
                  placeholder="密码 (至少4位)"
                  :prefix-icon="Lock"
                  show-password
                  class="custom-input"
              />
            </div>

            <div class="input-group animate-height" v-if="!isLogin">
              <el-input
                  v-model="form.confirmPassword"
                  type="password"
                  placeholder="请再次确认密码"
                  :prefix-icon="Lock"
                  show-password
                  class="custom-input"
              />
            </div>
          </template>

          <button
              type="button"
              class="submit-btn"
              @click="handleSubmit"
              :disabled="loading"
          >
            <span v-if="!loading">
              <template v-if="authMethod === 'email'">登 录 / 注 册</template>
              <template v-else>{{ isLogin ? '立 即 登 录' : '注 册 账 号' }}</template>
            </span>
            <el-icon v-else class="is-loading"><Loading /></el-icon>
          </button>

          <div class="toggle-area" v-if="authMethod === 'password'">
            <span v-if="isLogin">还没有账号？ <a @click="toggleMode">去注册</a></span>
            <span v-else>已有账号？ <a @click="toggleMode">去登录</a></span>
          </div>
        </el-form>
      </div>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, onUnmounted } from 'vue'
import { Message, Lock, Close, Loading, Monitor } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'

const props = defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const authMethod = ref('email') // 'email' or 'password'
const isLogin = ref(true) // only applies to 'password' method
const loading = ref(false)
const form = reactive({ 
  email: '', 
  password: '', 
  confirmPassword: '', 
  code: ''
})

// 验证码相关
const isSendingCode = ref(false)
const countdown = ref(0)
let timer = null

const close = () => { visible.value = false }

const switchAuthMethod = (method) => {
  authMethod.value = method
  resetForm()
}

const toggleMode = () => {
  isLogin.value = !isLogin.value
  resetForm()
}

const resetForm = () => {
  Object.assign(form, { email: '', password: '', confirmPassword: '', code: '' })
}

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const sendCode = async () => {
  if (!form.email || !emailRegex.test(form.email)) {
    ElMessage.warning('请输入有效的邮箱地址')
    return
  }

  isSendingCode.value = true
  try {
    const res = await request.post('/api/auth/send-code', { email: form.email })
    ElMessage.success(res.message || '验证码已发送，请查收')
    startCountdown(60)
  } catch (err) {
    if (err.response?.data?.countdown) {
      startCountdown(err.response.data.countdown)
    }
  } finally {
    isSendingCode.value = false
  }
}

const startCountdown = (seconds) => {
  countdown.value = seconds
  if (timer) clearInterval(timer)
  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
    }
  }, 1000)
}

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const handleSubmit = async () => {
  // --- 邮箱验证码登录/注册逻辑 ---
  if (authMethod.value === 'email') {
    if (!form.email || !emailRegex.test(form.email)) {
      ElMessage.warning('请输入有效的邮箱地址')
      return
    }
    if (!form.code || form.code.length !== 6) {
      ElMessage.warning('请输入6位验证码')
      return
    }
    
    loading.value = true
    try {
      const res = await request.post('/api/auth/verify-login', {
        email: form.email,
        code: form.code
      })
      
      ElMessage.success(res.message || '登录成功！')
      localStorage.setItem('token', res.token)
      localStorage.setItem('user', JSON.stringify(res.user))
      
      // 设置标志位，告知个人主页需弹出密码设置
      localStorage.setItem('needSetPassword', 'true')
      
      emit('success', res.user)
      close()
    } catch (err) {
      console.error('Email auth failed:', err)
    } finally {
      loading.value = false
    }
    return
  }

  // --- 密码登录/注册逻辑 ---
  if (!form.email || !emailRegex.test(form.email)) {
    ElMessage.warning('请输入有效的邮箱地址')
    return
  }
  if (!form.password) {
    ElMessage.warning('请输入密码')
    return
  }

  if (!isLogin.value) {
    if (form.password.length < 4) {
      ElMessage.warning('密码至少4位')
      return
    }

    if (form.password !== form.confirmPassword) {
      ElMessage.error('两次输入的密码不一致')
      return
    }
  }

  loading.value = true
  const url = isLogin.value ? '/api/login' : '/api/register'

  try {
    const submitData = {
      email: form.email,
      password: form.password
    }

    const res = await request.post(url, submitData)

    if (isLogin.value) {
      ElMessage.success('登录成功，欢迎回来！')
      localStorage.setItem('token', res.token)
      localStorage.setItem('user', JSON.stringify(res.user))
      emit('success', res.user)
      close()
    } else {
      ElMessage.success('注册成功，请登录')
      toggleMode()
    }
  } catch (err) {
    console.error('操作失败:', err)
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss">
.auth-dialog {
  border-radius: 24px !important; overflow: hidden; padding: 0 !important;
  .el-dialog__header { display: none; } .el-dialog__body { padding: 0 !important; }
}
</style>

<style scoped lang="scss">
.auth-container { display: flex; min-height: 540px; }

.auth-left {
  width: 320px; background: linear-gradient(135deg, #ffda44 0%, #ffc107 100%);
  position: relative; display: flex; flex-direction: column; justify-content: center; align-items: center; color: #333; overflow: hidden;
  .brand-home {
    display: flex;
    align-items: center;
    gap: 12px;
    z-index: 2;
    margin-top: -24px;
    .logo-box {
      width: 38px;
      height: 38px;
      position: relative;
      .circle-shape { position: absolute; width: 24px; height: 24px; background: #111; border-radius: 50%; top: 0; left: 0; }
      .square-shape { position: absolute; width: 20px; height: 20px; border: 2px solid #fff; bottom: 0; right: 0; border-radius: 4px; background: rgba(255,255,255,0.18); }
    }
    .brand-copy {
      display: flex;
      flex-direction: column;
      line-height: 1.1;
    }
    .brand-title {
      font-size: 26px;
      font-weight: 900;
      letter-spacing: 2px;
      color: #101216;
    }
    .brand-sub {
      margin-top: 4px;
      font-size: 12px;
      font-weight: 700;
      color: rgba(16, 18, 22, 0.72);
      letter-spacing: 1px;
    }
  }
  .slogan { text-align: center; margin-top: 20px; z-index: 2; h3 { margin: 0; font-size: 20px; } p { margin: 8px 0 0; opacity: 0.8; font-size: 14px; } }
  .circle { position: absolute; border-radius: 50%; background: rgba(255, 255, 255, 0.2); }
  .c1 { width: 200px; height: 200px; top: -50px; left: -50px; }
  .c2 { width: 150px; height: 150px; bottom: -30px; right: -30px; }
}

.auth-right {
  flex: 1; padding: 40px 50px; position: relative; display: flex; flex-direction: column; justify-content: center; background: #fff;
  .close-btn { position: absolute; top: 20px; right: 20px; background: transparent; border: none; font-size: 20px; cursor: pointer; color: #999; &:hover { color: #333; } }
  
  .auth-tabs {
    display: flex;
    gap: 16px;
    margin-bottom: 24px;
    border-bottom: 2px solid #f0f0f0;
    
    .auth-tab {
      padding: 8px 0;
      font-size: 16px;
      color: #666;
      cursor: pointer;
      position: relative;
      font-weight: 500;
      transition: color 0.3s;
      
      &.active {
        color: #333;
        font-weight: bold;
        
        &::after {
          content: '';
          position: absolute;
          bottom: -2px;
          left: 0;
          width: 100%;
          height: 3px;
          background: #ffda44;
          border-radius: 3px;
        }
      }
      
      &:hover:not(.active) {
        color: #000;
      }
    }
  }
  
  h2 { margin: 0 0 8px; font-size: 28px; color: #333; }
  .subtitle { margin: 0 0 24px; color: #999; font-size: 14px; }
  
  .input-group { 
    margin-bottom: 16px; 
    
    &.code-group {
      display: flex;
      gap: 12px;
      
      .code-input {
        flex: 1;
      }
      
      .send-code-btn {
        width: 120px;
        height: 36px;
        border-radius: 12px;
        background: #f0f2f5;
        color: #333;
        border: none;
        font-size: 13px;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 6px;
        
        &:hover:not(:disabled) {
          background: #e4e6eb;
        }
        
        &:disabled {
          background: #f5f5f5;
          color: #999;
          cursor: not-allowed;
        }
      }
    }
  }

  :deep(.custom-input .el-input__wrapper) {
    border-radius: 12px; padding: 8px 15px; box-shadow: 0 0 0 1px #eee; background: #f9f9f9; transition: all 0.3s;
    &.is-focus { box-shadow: 0 0 0 2px #ffda44 !important; background: #fff; }
  }

  .submit-btn {
    width: 100%; height: 48px; background: #333; color: #ffda44; border: none; border-radius: 12px; font-size: 16px; font-weight: bold; cursor: pointer; margin-top: 10px; transition: all 0.3s; display: flex; align-items: center; justify-content: center;
    &:hover { background: #000; transform: translateY(-2px); box-shadow: 0 8px 20px rgba(0,0,0,0.15); }
    &:disabled { opacity: 0.7; cursor: not-allowed; transform: none; }
  }

  .toggle-area {
    margin-top: 24px; text-align: center; font-size: 14px; color: #666;
    a { color: #ff9500; font-weight: bold; cursor: pointer; margin-left: 5px; &:hover { text-decoration: underline; } }
  }
}

.animate-height { animation: slideDown 0.3s ease; }
@keyframes slideDown { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }

@media (max-width: 860px) {
  .auth-container { flex-direction: column; min-height: auto; }
  .auth-left {
    width: 100%;
    min-height: 180px;
    justify-content: center;
    .brand-home { margin-top: 0; }
    .slogan { margin-top: 12px; }
    .c2 { right: -24px; bottom: -46px; }
  }
  .auth-right {
    padding: 28px 22px;
  }
}
</style>
