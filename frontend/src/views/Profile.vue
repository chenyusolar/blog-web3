<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { User, Mail, Camera, Save, Lock, ChevronRight, AlertCircle, CheckCircle2 } from 'lucide-vue-next'
import axios from 'axios'

const auth = useAuthStore()
const loading = ref(false)
const message = ref({ type: '', text: '' })

// Profile Form
const profileForm = ref({
  username: auth.user?.username || '',
  email: auth.user?.email || '',
  avatar: auth.user?.avatar || ''
})

// Password Form
const passwordForm = ref({
  old: '',
  new: '',
  confirm: ''
})

const handleUpdateProfile = async () => {
  loading.value = true
  message.value = { type: '', text: '' }
  try {
    await auth.updateProfile({
      username: profileForm.value.username,
      email: profileForm.value.email,
      avatar: profileForm.value.avatar
    })
    message.value = { type: 'success', text: '个人资料已更新' }
  } catch (err: any) {
    message.value = { type: 'error', text: err.response?.data?.error || '更新失败' }
  } finally {
    loading.value = false
  }
}

const handleChangePassword = async () => {
  if (passwordForm.value.new !== passwordForm.value.confirm) {
    message.value = { type: 'error', text: '两次输入的密码不一致' }
    return
  }
  
  loading.value = true
  try {
    await auth.changePassword(passwordForm.value.old, passwordForm.value.new)
    message.value = { type: 'success', text: '密码修改成功' }
    passwordForm.value = { old: '', new: '', confirm: '' }
  } catch (err: any) {
    message.value = { type: 'error', text: err.response?.data?.error || '修改失败' }
  } finally {
    loading.value = false
  }
}

const handleAvatarUpload = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    const res = await axios.post('http://localhost:8888/api/v1/upload', formData)
    profileForm.value.avatar = `http://localhost:8888${res.data.file_url}`
    message.value = { type: 'success', text: '头像上传成功，点击保存生效' }
  } catch (err) {
    message.value = { type: 'error', text: '头像上传失败' }
  }
}

onMounted(async () => {
  try {
    await auth.fetchProfile()
    profileForm.value = {
      username: auth.user?.username || '',
      email: auth.user?.email || '',
      avatar: auth.user?.avatar || ''
    }
  } catch (err) {
    console.error('Failed to fetch profile', err)
  }
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-12">
    <div class="flex items-center gap-4 mb-8">
      <div class="p-3 bg-indigo-100 rounded-2xl text-indigo-600">
        <User class="w-8 h-8" />
      </div>
      <div>
        <h1 class="text-3xl font-bold text-slate-900">个人设置</h1>
        <p class="text-slate-500">管理您的个人信息、账号安全和偏好</p>
      </div>
    </div>

    <!-- Message Alert -->
    <div v-if="message.text" :class="[
      'mb-8 p-4 rounded-xl flex items-center gap-3 animate-in fade-in slide-in-from-top-4 duration-300',
      message.type === 'success' ? 'bg-emerald-50 text-emerald-700 border border-emerald-100' : 'bg-rose-50 text-rose-700 border border-rose-100'
    ]">
      <CheckCircle2 v-if="message.type === 'success'" class="w-5 h-5 flex-shrink-0" />
      <AlertCircle v-else class="w-5 h-5 flex-shrink-0" />
      <span class="text-sm font-medium">{{ message.text }}</span>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      <!-- Sidebar Navigation -->
      <div class="space-y-2">
        <button class="w-full text-left px-4 py-3 rounded-xl bg-white shadow-sm border border-slate-200 text-indigo-600 font-bold flex items-center justify-between group transition hover:border-indigo-200">
          <div class="flex items-center gap-3">
            <User class="w-5 h-5" />
            <span>基本信息</span>
          </div>
          <ChevronRight class="w-4 h-4 opacity-50 group-hover:translate-x-1 transition" />
        </button>
        <button class="w-full text-left px-4 py-3 rounded-xl bg-transparent text-slate-600 font-medium flex items-center justify-between group transition hover:bg-white hover:shadow-sm">
          <div class="flex items-center gap-3">
            <Lock class="w-5 h-5" />
            <span>账号安全</span>
          </div>
          <ChevronRight class="w-4 h-4 opacity-0 group-hover:opacity-50 group-hover:translate-x-1 transition" />
        </button>
      </div>

      <!-- Main Content Area -->
      <div class="md:col-span-2 space-y-8">
        <!-- Profile Section -->
        <section class="bg-white rounded-3xl p-8 shadow-sm border border-slate-100">
          <div class="mb-8 flex flex-col items-center">
            <div class="relative group">
              <div class="w-32 h-32 rounded-full ring-4 ring-slate-50 overflow-hidden bg-slate-100 flex items-center justify-center text-slate-400">
                <img v-if="profileForm.avatar" :src="profileForm.avatar" class="w-full h-full object-cover transition duration-500 group-hover:scale-110" />
                <User v-else class="w-12 h-12" />
              </div>
              <label class="absolute bottom-0 right-0 p-2 bg-indigo-600 text-white rounded-full cursor-pointer shadow-lg hover:bg-indigo-700 transition transform hover:scale-110">
                <Camera class="w-5 h-5" />
                <input type="file" class="hidden" accept="image/*" @change="handleAvatarUpload" />
              </label>
            </div>
            <p class="mt-4 text-sm text-slate-500 font-medium">点击右下角按钮上传新头像</p>
          </div>

          <form @submit.prevent="handleUpdateProfile" class="space-y-6">
            <div class="grid grid-cols-1 gap-6">
              <div class="space-y-2">
                <label class="text-sm font-bold text-slate-700">用户名</label>
                <div class="relative">
                  <User class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                  <input 
                    v-model="profileForm.username"
                    type="text" 
                    placeholder="您的显示名称" 
                    class="w-full pl-10 pr-4 py-3 bg-slate-50 border-none rounded-xl focus:ring-2 focus:ring-indigo-500 transition outline-none text-slate-700 font-medium"
                  />
                </div>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-bold text-slate-700">电子邮箱</label>
                <div class="relative">
                  <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                  <input 
                    v-model="profileForm.email"
                    type="email" 
                    placeholder="someone@example.com" 
                    class="w-full pl-10 pr-4 py-3 bg-slate-50 border-none rounded-xl focus:ring-2 focus:ring-indigo-500 transition outline-none text-slate-700 font-medium"
                  />
                </div>
              </div>
            </div>

            <button 
              type="submit"
              :disabled="loading"
              class="w-full py-3 bg-indigo-600 text-white rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-indigo-700 transition disabled:opacity-50 shadow-lg shadow-indigo-100"
            >
              <Save v-if="!loading" class="w-5 h-5" />
              <span v-else class="animate-spin border-2 border-white border-t-transparent rounded-full w-5 h-5 mr-2"></span>
              {{ loading ? '更新中...' : '保存更改' }}
            </button>
          </form>
        </section>

        <!-- Security Section -->
        <section class="bg-white rounded-3xl p-8 shadow-sm border border-slate-100">
          <div class="flex items-center gap-3 mb-6">
            <div class="p-2 bg-rose-100 rounded-lg text-rose-600">
              <Lock class="w-5 h-5" />
            </div>
            <h2 class="text-xl font-bold text-slate-900">修改密码</h2>
          </div>

          <form @submit.prevent="handleChangePassword" class="space-y-6">
            <div class="space-y-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-slate-600">当前密码</label>
                <input 
                  v-model="passwordForm.old"
                  type="password" 
                  class="w-full px-4 py-3 bg-slate-50 border-none rounded-xl focus:ring-2 focus:ring-indigo-500 transition outline-none"
                  required
                />
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="space-y-2">
                  <label class="text-sm font-medium text-slate-600">新密码</label>
                  <input 
                    v-model="passwordForm.new"
                    type="password" 
                    class="w-full px-4 py-3 bg-slate-50 border-none rounded-xl focus:ring-2 focus:ring-indigo-500 transition outline-none"
                    required
                  />
                </div>
                <div class="space-y-2">
                  <label class="text-sm font-medium text-slate-600">确认新密码</label>
                  <input 
                    v-model="passwordForm.confirm"
                    type="password" 
                    class="w-full px-4 py-3 bg-slate-50 border-none rounded-xl focus:ring-2 focus:ring-indigo-500 transition outline-none"
                    required
                  />
                </div>
              </div>
            </div>

            <button 
              type="submit"
              :disabled="loading"
              class="w-full py-3 bg-slate-900 text-white rounded-xl font-bold hover:bg-slate-800 transition disabled:opacity-50 shadow-lg shadow-slate-100"
            >
              {{ loading ? '处理中...' : '确认修改' }}
            </button>
          </form>
        </section>
      </div>
    </div>
  </div>
</template>
