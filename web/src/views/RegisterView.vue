<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '../services/api'

const router = useRouter()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref(false)

const handleSubmit = async () => {
  if (!username.value || !password.value || !confirmPassword.value) {
    error.value = '请填写所有字段'
    return
  }

  if (password.value !== confirmPassword.value) {
    error.value = '两次密码输入不一致'
    return
  }

  if (password.value.length < 6) {
    error.value = '密码至少需要6个字符'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await authApi.register(username.value, password.value)
    success.value = true
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (err) {
    error.value = err.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex justify-center">
    <div class="card w-full max-w-md bg-base-100 shadow-xl">
      <div class="card-body">
        <h2 class="card-title text-2xl justify-center mb-4">📝 用户注册</h2>
        
        <div v-if="success" class="alert alert-success mb-4">
          <span>注册成功！正在跳转到登录页面...</span>
        </div>

        <div v-if="error" class="alert alert-error mb-4">
          <span>{{ error }}</span>
        </div>

        <form v-if="!success" @submit.prevent="handleSubmit">
          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">用户名</span>
            </label>
            <input 
              v-model="username"
              type="text" 
              placeholder="至少3个字符" 
              class="input input-bordered"
              :disabled="loading"
            />
          </div>

          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">密码</span>
            </label>
            <input 
              v-model="password"
              type="password" 
              placeholder="至少6个字符" 
              class="input input-bordered"
              :disabled="loading"
            />
          </div>

          <div class="form-control mb-6">
            <label class="label">
              <span class="label-text">确认密码</span>
            </label>
            <input 
              v-model="confirmPassword"
              type="password" 
              placeholder="再次输入密码" 
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
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>

        <div v-if="!success" class="divider">OR</div>

        <p v-if="!success" class="text-center">
          已有账户？
          <router-link to="/login" class="link link-primary">立即登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>
