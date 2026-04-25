<template>
  <div class="admin-login">
    <div class="admin-login__shell">
      <section class="admin-login__intro" aria-hidden="true">
        <h1>管理工作台</h1>
        <p>商品管理 · 订单跟进 · 数据查看</p>
      </section>

      <section class="admin-login__panel">
        <div class="admin-login__brand">
          <div class="admin-login__brand-mark">
            <User />
          </div>
          <div class="admin-login__brand-title">
            <strong>XIANQU</strong>
            <span>后台管理</span>
          </div>
        </div>

        <form class="admin-login__form" @submit.prevent="handleLogin">
          <label class="admin-login__label">
            <span>账号</span>
            <div class="admin-login__field">
              <User />
              <input
                v-model.trim="form.username"
                type="text"
                autocomplete="username"
                placeholder="管理员账号"
              />
            </div>
          </label>

          <label class="admin-login__label">
            <span>密码</span>
            <div class="admin-login__field">
              <Lock />
              <input
                v-model="form.password"
                type="password"
                autocomplete="current-password"
                placeholder="登录密码"
              />
            </div>
          </label>

          <button class="admin-login__submit" type="submit" :disabled="loading">
            <span>{{ loading ? '验证中...' : '进入后台' }}</span>
            <Right v-if="!loading" />
          </button>
        </form>
      </section>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import request from '@/utils/request'
import { ElMessage } from '@/ui/feedback'
import { User, Lock, Right } from '@/icons/tw-icons.js'

const router = useRouter()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const handleLogin = async () => {
  if (!form.username) {
    ElMessage.warning('请输入账号')
    return
  }

  if (!form.password) {
    ElMessage.warning('请输入密码')
    return
  }

  loading.value = true
  try {
    const res = await request.post('/api/admin/login', form)
    localStorage.setItem('admin_token', res.token)
    localStorage.setItem('admin_user', JSON.stringify(res.admin))
    ElMessage.success('登录成功')
    router.replace('/admin/dashboard')
  } catch (error) {
    console.error(error)
    ElMessage.error('登录失败，请检查账号密码')
  } finally {
    loading.value = false
  }
}
</script>
