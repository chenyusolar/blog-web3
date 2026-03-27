<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { User, Lock, Mail, ArrowRight, Loader2, Sparkles } from 'lucide-vue-next'

const username = ref('')
const email = ref('')
const password = ref('')
const inviterCode = ref('')
const importWallet = ref('')
const loading = ref(false)
const error = ref('')

const auth = useAuthStore()
const router = useRouter()

const handleRegister = async () => {
  if (!username.value || !password.value || !email.value) {
    error.value = '请填写所有必填项'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    await auth.register(username.value, password.value, email.value, inviterCode.value, importWallet.value)
    // 注册成功后自动登录或跳转到登录页
    router.push('/login')
  } catch (err: any) {
    error.value = err.response?.data?.error || '注册失败，用户名或邮箱可能已存在'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[calc(100vh-4rem)] flex items-center justify-center p-6 bg-slate-50/50">
    <div class="w-full max-w-md">
      <!-- Logo Section -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-indigo-600 text-white shadow-xl shadow-indigo-200 mb-4">
          <Sparkles class="w-10 h-10" />
        </div>
        <h1 class="text-3xl font-bold text-slate-900">加入 AIGen</h1>
        <p class="text-slate-500 mt-2 font-medium">开启您的 AI 驱动博客之旅</p>
      </div>

      <!-- Register Form Card -->
      <div class="bg-white p-8 rounded-[32px] shadow-xl shadow-slate-200/50 border border-slate-100">
        <form @submit.prevent="handleRegister" class="space-y-6">
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
                placeholder="选择一个响亮的用户名"
                class="w-full pl-12 pr-4 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">电子邮件</label>
            <div class="relative group">
              <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none transition-colors group-focus-within:text-indigo-600 text-slate-400">
                <Mail class="w-5 h-5" />
              </div>
              <input 
                v-model="email"
                type="email" 
                placeholder="your@email.com"
                class="w-full pl-12 pr-4 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">设置密码</label>
            <div class="relative group">
              <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none transition-colors group-focus-within:text-indigo-600 text-slate-400">
                <Lock class="w-5 h-5" />
              </div>
              <input 
                v-model="password"
                type="password" 
                placeholder="至少 8 位数字或字母"
                class="w-full pl-12 pr-4 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">邀请码 (可选)</label>
            <div class="relative group">
              <input 
                v-model="inviterCode"
                type="text" 
                placeholder="如果有的话，请填写"
                class="w-full px-6 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
              />
            </div>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-bold text-slate-700 ml-1">导入 Solana 地址 (可选)</label>
            <div class="relative group">
              <input 
                v-model="importWallet"
                type="text" 
                placeholder="留空则自动为您生成"
                class="w-full px-6 py-4 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 placeholder:text-slate-400 outline-none"
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
              创建账号
              <ArrowRight class="w-6 h-6 group-hover/btn:translate-x-1 transition-transform" />
            </template>
          </button>
        </form>

        <div class="mt-8 pt-8 border-t border-slate-50 text-center">
          <p class="text-slate-500 font-medium">
            已经有账号了？ 
            <router-link to="/login" class="text-indigo-600 font-bold hover:underline">立即登录</router-link>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
