<template>
  <div class="admin-login-bg">
    <div class="login-card animate-up">
      <div class="logo">
        <div class="icon-box"><div class="shape"></div></div>
        <h1>XIANQU</h1>
        <p>ADMIN CENTER</p>
      </div>

      <div class="welcome-text">欢迎回来 👋 <br><span>请输入您的管理员账号以继续</span></div>

      <div class="form-area">
        <div class="input-group">
          <el-icon><User /></el-icon>
          <input v-model="form.username" placeholder="admin" @keyup.enter="handleLogin" />
        </div>
        <div class="input-group">
          <el-icon><Lock /></el-icon>
          <input v-model="form.password" type="password" placeholder="••••••" @keyup.enter="handleLogin" />
        </div>

        <button class="btn-login" @click="handleLogin" :disabled="loading">
          {{ loading ? '登录中...' : '登 录 系 统 →' }}
        </button>
      </div>

      <div class="footer">© 2025 Xianqu Team. All rights reserved.</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

const handleLogin = async () => {
  if (!form.username || !form.password) return ElMessage.warning('请输入账号密码')

  loading.value = true
  try {
    // 请求后台登录接口
    const res = await request.post('/api/admin/login', form)

    ElMessage.success('登录成功')

    // ★★★ 核心修复：必须存为 admin_token ★★★
    // 这样 request.js 在请求 /api/admin/* 接口时，才会带上这个 Token
    localStorage.setItem('admin_token', res.token)
    localStorage.setItem('admin_user', JSON.stringify(res.admin))

    // 跳转到后台仪表盘
    router.push('/admin/users')

  } catch (e) {
    console.error(e)
    // 显示具体错误 (如: "管理员账号不存在")
    ElMessage.error(e.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
/* 深色科技风背景 */
.admin-login-bg {
  min-height: 100vh;
  background: #1a1a1a;
  background-image: radial-gradient(circle at 50% 30%, #333 0%, #1a1a1a 60%);
  display: flex; align-items: center; justify-content: center;
}

.login-card {
  background: #fff; width: 420px; padding: 50px 40px; border-radius: 24px; text-align: center;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);

  .logo {
    margin-bottom: 30px;
    .icon-box {
      width: 40px; height: 40px; border: 2px solid #333; border-radius: 8px; margin: 0 auto 10px; position: relative;
      .shape { position: absolute; width: 20px; height: 20px; background: #ffdf5d; border-radius: 50%; top: -5px; right: -5px; }
    }
    h1 { font-size: 24px; font-weight: 900; letter-spacing: 4px; margin: 0; color: #333; }
    p { font-size: 12px; color: #999; letter-spacing: 2px; margin-top: 4px; font-weight: bold; }
  }

  .welcome-text { font-size: 18px; font-weight: bold; color: #333; margin-bottom: 30px; span { font-size: 13px; color: #999; font-weight: normal; } }

  .form-area {
    .input-group {
      background: #f5f5f5; border-radius: 12px; height: 50px; display: flex; align-items: center; padding: 0 16px; margin-bottom: 16px; transition: 0.3s;
      &:focus-within { background: #fff; box-shadow: 0 0 0 2px #ffdf5d; }
      .el-icon { font-size: 18px; color: #999; margin-right: 12px; }
      input { border: none; background: transparent; outline: none; flex: 1; font-size: 15px; color: #333; }
    }

    .btn-login {
      width: 100%; height: 50px; background: #333; color: #ffdf5d; font-weight: bold; font-size: 16px; border: none; border-radius: 12px; cursor: pointer; margin-top: 10px; transition: 0.3s;
      &:hover { background: #000; transform: translateY(-2px); box-shadow: 0 8px 20px rgba(0,0,0,0.3); }
      &:disabled { opacity: 0.7; cursor: not-allowed; transform: none; }
    }
  }

  .footer { margin-top: 40px; font-size: 12px; color: #ccc; }
}

.animate-up { animation: fadeInUp 0.6s cubic-bezier(0.2, 0.8, 0.2, 1); }
@keyframes fadeInUp { from { opacity: 0; transform: translateY(40px); } to { opacity: 1; transform: translateY(0); } }
</style>