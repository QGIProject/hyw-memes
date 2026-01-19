<script setup>
import { ref, onMounted, computed } from 'vue'
import { authApi } from './services/api'

const isLoggedIn = ref(false)
const username = ref('')

const checkAuth = async () => {
  const token = localStorage.getItem('token')
  if (token) {
    try {
      const res = await authApi.getMe()
      isLoggedIn.value = true
      username.value = res.data.username
    } catch {
      localStorage.removeItem('token')
    }
  }
}

const logout = () => {
  localStorage.removeItem('token')
  isLoggedIn.value = false
  username.value = ''
  window.location.href = '/'
}

onMounted(checkAuth)
</script>

<template>
  <div class="min-h-screen bg-base-200">
    <!-- Navbar (Hidden on Admin) -->
    <div v-if="!$route.path.startsWith('/admin')" class="navbar bg-base-100 shadow-lg">
      <div class="flex-1">
        <router-link to="/" class="btn btn-ghost text-xl">
          🎭 梗图分享
        </router-link>
      </div>
      <div class="flex-none gap-2">
        <router-link to="/" class="btn btn-ghost">首页</router-link>
        
        <template v-if="isLoggedIn">
          <router-link to="/upload" class="btn btn-primary">上传</router-link>
          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button" class="btn btn-ghost">
              {{ username }}
            </div>
            <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-32">
              <li><a @click="logout">退出登录</a></li>
            </ul>
          </div>
        </template>
        
        <template v-else>
          <router-link to="/login" class="btn btn-ghost">登录</router-link>
          <router-link to="/register" class="btn btn-secondary">注册</router-link>
        </template>
        
        <router-link to="/admin" class="btn btn-ghost btn-sm">管理</router-link>
      </div>
    </div>

    <!-- Main Content (Standard container for non-admin, Full width for admin) -->
    <main :class="$route.path.startsWith('/admin') ? 'h-screen w-full' : 'container mx-auto px-4 py-8'">
      <router-view @login-success="checkAuth" />
    </main>

    <!-- Footer (Hidden on Admin) -->
    <footer v-if="!$route.path.startsWith('/admin')" class="footer footer-center p-4 bg-base-300 text-base-content">
      <p>梗图分享平台 © 2026</p>
    </footer>
  </div>
</template>
