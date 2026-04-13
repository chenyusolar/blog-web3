<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { User, Lock, ArrowRight, Loader2, Sparkles } from 'lucide-vue-next'

const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const auth = useAuthStore()
const router = useRouter()

const handleLogin = async () => {
  if (!username.value || !password.value) {
    error.value = '请填写用户名和密码'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    await auth.login(username.value, password.value)
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.message || '登录失败，请检查用户名或密码'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[calc(100-4rem)] flex items-center justify-center p-6 bg-slate-50/50">
    <div class="w-full max-w-md">
      <!-- Logo/Brand Section -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-indigo-600 text-white shadow-xl shadow-indigo-200 mb-4">
          <Sparkles class="w-10 h-10" />
        </div>
        <h1 class="text-3xl font-bold text-slate-900">欢迎回来</h1>
        <p class="text-slate-500 mt-2 font-medium">登录您的 AIGen-Blog 账号</p>
      </div>

      <!-- Login Form Card -->
      <div class="bg-white p-8 rounded-[32px] shadow-xl shadow-slate-200/50 border border-slate-100">
        <form @submit.prevent="handleLogin" class="space-y-6">
          <div v-if="error" class="bg-red-50 text-red-500 p-4 rounded-xl text-sm font-semibold border border-red-100 flex items-center gap-2">
            <span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
            {{ error }}
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">用户名</label>
            <div class="relative group">
              <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none transition-colors group-focus-within:text-indigo-600 text-slate-400">
                <User class="w-5 h-5" />
              </div>
              <input 
                v-model="username"
                type="text" 
                placeholder="输入用户名"
                class="w-full pl-12 pr-4 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">密码</label>
            <div class="relative group">
              <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none transition-colors group-focus-within:text-indigo-600 text-slate-400">
                <Lock class="w-5 h-5" />
              </div>
              <input 
                v-model="password"
                type="password" 
                placeholder="输入密码"
                class="w-full pl-12 pr-4 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
              />
            </div>
          </div>

          <button 
            type="submit" 
            :disabled="loading"
            class="w-full bg-indigo-600 hover:bg-indigo-700 text-white py-4 rounded-2xl font-bold text-lg transition-all shadow-lg shadow-indigo-200 flex items-center justify-center gap-2 disabled:opacity-70 disabled:cursor-not-allowed group/btn"
          >
            <Loader2 v-if="loading" class="w-6 h-6 animate-spin" />
            <template v-else>
              立刻登录
              <ArrowRight class="w-6 h-6 group-hover/btn:translate-x-1 transition-transform" />
            </template>
          </button>
        </form>

        <div class="mt-8 pt-8 border-t border-slate-50 text-center">
          <p class="text-slate-500 font-medium">
            还没有账号？ 
            <router-link to="/register" class="text-indigo-600 font-bold hover:underline">免费注册</router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
