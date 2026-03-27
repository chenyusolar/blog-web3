<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Users, ShieldCheck, Settings, Save, Loader2 } from 'lucide-vue-next'

interface User {
  id: number
  username: string
  email: string
  wallet_address: string
  blog_balance: number
}

const stats = ref({ user_count: 0, blog_count: 0, comment_count: 0, total_issued: 0 })
const users = ref<User[]>([])
const config = ref<Record<string, string>>({})
const loading = ref(true)
const saving = ref(false)
const message = ref('')

onMounted(async () => {
  try {
    const [statsRes, usersRes, configRes] = await Promise.all([
      axios.get('http://localhost:8888/api/v1/admin/stats'),
      axios.get('http://localhost:8888/api/v1/admin/users'),
      axios.get('http://localhost:8888/api/v1/admin/config')
    ])
    stats.value = statsRes.data
    users.value = usersRes.data
    config.value = configRes.data
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
})

const updateConfig = async () => {
  saving.value = true
  try {
    await axios.put('http://localhost:8888/api/v1/admin/config', config.value)
    message.value = '配置更新成功！'
    setTimeout(() => message.value = '', 3000)
  } catch (err) {
    console.error(err)
    message.value = '配置更新失败'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="max-w-7xl mx-auto py-12 px-6">
    <div v-if="loading" class="flex flex-col items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-slate-900 animate-spin" />
      <p class="text-slate-500 mt-4 font-medium">进入加密管理区...</p>
    </div>

    <template v-else>
      <div class="mb-10 flex items-center gap-4">
        <div class="w-16 h-16 rounded-[24px] bg-slate-900 text-white flex items-center justify-center shadow-xl">
          <ShieldCheck class="w-8 h-8" />
        </div>
        <div>
          <h1 class="text-4xl font-black text-slate-900 tracking-tight">管理员面板</h1>
          <p class="text-slate-500 font-medium">系统概览与 Web3 参数设置</p>
        </div>
      </div>

      <!-- Stats Cards -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-12">
        <div class="bg-white p-8 rounded-[32px] border border-slate-100 shadow-xl shadow-slate-200/50">
          <p class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-2">总用户数</p>
          <h4 class="text-4xl font-black text-slate-900">{{ stats.user_count }}</h4>
        </div>
        <div class="bg-white p-8 rounded-[32px] border border-slate-100 shadow-xl shadow-slate-200/50">
          <p class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-2">文章总数</p>
          <h4 class="text-4xl font-black text-slate-900">{{ stats.blog_count }}</h4>
        </div>
        <div class="bg-white p-8 rounded-[32px] border border-slate-100 shadow-xl shadow-slate-200/50">
          <p class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-2">评论总数</p>
          <h4 class="text-4xl font-black text-slate-900">{{ stats.comment_count }}</h4>
        </div>
        <div class="bg-white p-8 rounded-[32px] border border-slate-100 shadow-xl shadow-slate-200/50 bg-indigo-50 border-indigo-100">
          <p class="text-xs font-bold text-indigo-400 uppercase tracking-widest mb-2">代币总发放</p>
          <h4 class="text-4xl font-black text-indigo-600">{{ stats.total_issued?.toFixed(2) }}</h4>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-12">
        <!-- Config Section -->
        <div class="lg:col-span-1 space-y-8">
          <div class="bg-white p-8 rounded-[40px] border border-slate-100 shadow-xl shadow-slate-100">
            <div class="flex items-center gap-3 mb-8">
              <Settings class="w-6 h-6 text-indigo-600" />
              <h3 class="text-xl font-bold text-slate-900">系统配置</h3>
            </div>

            <form @submit.prevent="updateConfig" class="space-y-6">
              <div v-for="(_, key) in config" :key="key" class="space-y-2">
                <label class="text-xs font-bold text-slate-500 uppercase ml-1">{{ key.replace(/_/g, ' ') }}</label>
                <input 
                  v-model="config[key]"
                  type="text"
                  class="w-full px-5 py-3 bg-slate-50 border-none rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 outline-none text-sm break-all"
                />
              </div>

              <button 
                type="submit" 
                :disabled="saving"
                class="w-full bg-slate-900 text-white py-4 rounded-2xl font-bold transition-all shadow-lg hover:bg-slate-800 flex items-center justify-center gap-2"
              >
                <Loader2 v-if="saving" class="w-5 h-5 animate-spin" />
                <Save v-else class="w-5 h-5" />
                保存基础配置
              </button>
              <p v-if="message" class="text-center text-sm font-bold text-emerald-600 truncate">{{ message }}</p>
            </form>
          </div>
        </div>

        <!-- User List Section -->
        <div class="lg:col-span-2 space-y-8">
          <div class="bg-white rounded-[40px] border border-slate-100 shadow-xl shadow-slate-100 overflow-hidden">
            <div class="p-8 border-b border-slate-50 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <Users class="w-6 h-6 text-indigo-600" />
                <h3 class="text-xl font-bold text-slate-900">用户管理</h3>
              </div>
            </div>
            
            <div class="overflow-x-auto">
              <table class="w-full text-left">
                <thead>
                  <tr class="bg-slate-50">
                    <th class="px-8 py-4 text-xs font-bold text-slate-400 uppercase">用户信息</th>
                    <th class="px-8 py-4 text-xs font-bold text-slate-400 uppercase">Solana 地址</th>
                    <th class="px-8 py-4 text-xs font-bold text-slate-400 uppercase">代币余额</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-50">
                  <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50/50 transition">
                    <td class="px-8 py-6">
                      <div>
                        <p class="font-bold text-slate-900">{{ user.username }}</p>
                        <p class="text-xs text-slate-400">{{ user.email }}</p>
                      </div>
                    </td>
                    <td class="px-8 py-6">
                      <span class="font-mono text-xs text-slate-500 block truncate max-w-[150px]" :title="user.wallet_address">
                        {{ user.wallet_address || '未生成' }}
                      </span>
                    </td>
                    <td class="px-8 py-6 text-right pr-12">
                      <span class="px-3 py-1 rounded-full bg-indigo-50 text-indigo-600 font-black text-sm">
                        {{ user.blog_balance?.toFixed(2) }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
