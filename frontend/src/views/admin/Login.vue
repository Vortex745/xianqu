<template>
  <div class="login-container">
    <div class="decorative-circle c1"></div>
    <div class="decorative-circle c2"></div>

    <div class="login-card animate-up">
      <div class="brand-header">
        <div class="logo-shape">
          <div class="circle"></div>
          <div class="square"></div>
        </div>
        <div class="brand-title">
          <span class="main">XIANQU</span>
          <span class="sub">Admin Center</span>
        </div>
      </div>

      <div class="welcome-text">
        <h3>欢迎回来 👋</h3>
        <p>请输入您的管理员账号以继续</p>
      </div>

      <el-form :model="form" :rules="rules" ref="formRef" size="large" class="login-form" @submit.prevent>
        <el-form-item prop="username">
          <div class="input-wrapper">
            <el-icon class="icon"><User /></el-icon>
            <input
                v-model="form.username"
                type="text"
                placeholder="管理员账号"
                autocomplete="off"
            />
          </div>
        </el-form-item>

        <el-form-item prop="password">
          <div class="input-wrapper">
            <el-icon class="icon"><Lock /></el-icon>
            <input
                v-model="form.password"
                type="password"
                placeholder="登录密码"
                @keyup.enter="handleLogin"
            />
          </div>
        </el-form-item>

        <button
            class="submit-btn"
            :class="{ loading: loading }"
            @click.prevent="handleLogin"
            :disabled="loading"
        >
          <span v-if="!loading">登 录 系 统</span>
          <span v-else>登 录 中...</span>
          <el-icon v-if="!loading" class="arrow"><Right /></el-icon>
        </button>
      </el-form>

      <div class="footer-text">
        © 2025 Xianqu Team. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import { User, Lock, Right } from '@element-plus/icons-vue'

const router = useRouter()
const formRef = ref(null)
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = () => {
  if (!form.username || !form.password) return

  formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const res = await request.post('/api/admin/login', form)

        localStorage.setItem('admin_token', res.token)
        localStorage.setItem('admin_user', JSON.stringify(res.admin))

        ElMessage.success('登录成功，正在跳转...')

        // 延迟一下，让用户看清成功状态
        setTimeout(() => {
          window.location.href = '/admin/dashboard'
        }, 800)
      } catch (e) {
        console.error(e)
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped lang="scss">
$primary: #ffdf5d;
$dark-bg: #1a1c23;
$card-bg: #ffffff;

.login-container {
  height: 100vh;
  width: 100vw;
  background-color: $dark-bg;
  /* 高级深色渐变背景 */
  background-image: radial-gradient(circle at 10% 20%, #2b303b 0%, #1a1c23 90%);
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  overflow: hidden;
}

/* 背景装饰球 */
.decorative-circle {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  z-index: 1;

  &.c1 {
    width: 400px; height: 400px;
    background: #ffdf5d;
    top: -100px; left: -100px;
    animation: float 10s infinite ease-in-out;
  }
  &.c2 {
    width: 300px; height: 300px;
    background: #409eff;
    bottom: -50px; right: -50px;
    animation: float 12s infinite ease-in-out reverse;
  }
}

@keyframes float {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, 50px); }
}

.login-card {
  position: relative;
  z-index: 10;
  width: 420px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  text-align: center;
}

/* 1. 品牌头部 */
.brand-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 30px;

  .logo-shape {
    width: 48px; height: 48px; position: relative; margin-bottom: 16px;
    .circle {
      position: absolute; width: 32px; height: 32px; background: $primary; border-radius: 50%; top: 0; left: 0;
      box-shadow: 0 0 15px rgba(255, 223, 93, 0.6);
    }
    .square {
      position: absolute; width: 28px; height: 28px; border: 3px solid #333; border-radius: 6px; bottom: 0; right: 0;
      background: rgba(255,255,255,0.5); backdrop-filter: blur(4px);
    }
  }

  .brand-title {
    display: flex; flex-direction: column;
    .main { font-size: 24px; font-weight: 900; color: #333; letter-spacing: 2px; line-height: 1.2; }
    .sub { font-size: 12px; color: #999; letter-spacing: 4px; text-transform: uppercase; }
  }
}

/* 2. 欢迎语 */
.welcome-text {
  margin-bottom: 32px;
  h3 { font-size: 20px; color: #333; margin: 0 0 8px 0; }
  p { font-size: 14px; color: #999; margin: 0; }
}

/* 3. 表单 */
.login-form {
  .el-form-item { margin-bottom: 20px; }
}

/* 自定义输入框容器 */
.input-wrapper {
  box-sizing: border-box;
  width: 100%;
  height: 50px;
  background: #f5f7fa;
  border-radius: 99px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  transition: all 0.3s ease;
  border: 1px solid transparent;

  &:focus-within {
    background: #fff;
    border-color: $primary;
    box-shadow: 0 0 0 4px rgba(255, 223, 93, 0.2);
  }

  .icon { font-size: 18px; color: #a0a5a9; margin-right: 12px; }

  input {
    flex: 1;
    border: none;
    background: transparent;
    outline: none;
    font-size: 15px;
    color: #333;
    height: 100%;

    &::placeholder { color: #ccc; }
  }
}

/* 登录按钮 */
.submit-btn {
  width: 100%;
  height: 50px;
  border-radius: 99px;
  border: none;
  background: #333;
  color: $primary;
  font-size: 16px;
  font-weight: bold;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition: all 0.3s;
  margin-top: 10px;
  box-shadow: 0 8px 20px rgba(0,0,0,0.15);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 12px 24px rgba(0,0,0,0.2);
    background: #000;
  }

  &:active { transform: scale(0.98); }

  &.loading {
    opacity: 0.8;
    cursor: not-allowed;
  }

  .arrow { font-size: 18px; transition: transform 0.3s; }
  &:hover .arrow { transform: translateX(4px); }
}

.footer-text {
  margin-top: 40px;
  font-size: 12px;
  color: #ccc;
}

/* 进场动画 */
.animate-up { animation: fadeInUp 0.8s cubic-bezier(0.2, 0.8, 0.2, 1); }

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(40px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>