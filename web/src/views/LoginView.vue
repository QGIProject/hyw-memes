<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../services/api'

const emit = defineEmits(['login-success'])
const router = useRouter()

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const handleSubmit = async () => {
  if (!username.value || !password.value) {
    error.value = '请填写用户名和密码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const res = await authApi.login(username.value, password.value)
    localStorage.setItem('token', res.data.token)
    emit('login-success')
    router.push('/')
  } catch (err) {
    error.value = err.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex justify-center">
    <div class="card w-full max-w-md bg-base-100 shadow-xl">
      <div class="card-body">
        <h2 class="card-title text-2xl justify-center mb-4">🔐 用户登录</h2>
        
        <div v-if="error" class="alert alert-error mb-4">
          <span>{{ error }}</span>
        </div>

        <form @submit.prevent="handleSubmit">
          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">用户名</span>
            </label>
            <input 
              v-model="username"
              type="text" 
              placeholder="请输入用户名" 
              class="input input-bordered"
              :disabled="loading"
            />
          </div>

          <div class="form-control mb-6">
            <label class="label">
              <span class="label-text">密码</span>
            </label>
            <input 
              v-model="password"
              type="password" 
              placeholder="请输入密码" 
              class="input input-bordered"
              :disabled="loading"
            />
          </div>

          <button 
            type="submit" 
            class="btn btn-primary w-full"
            :disabled="loading"
          >
            <span v-if="loading" class="loading loading-spinner"></span>
            {{ loading ? '登录中...' : '登录' }}
          </button>
        </form>

        <div class="divider">OR</div>

        <p class="text-center">
          还没有账户？
          <router-link to="/register" class="link link-primary">立即注册</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
