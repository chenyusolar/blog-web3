<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Trophy, Medal, Coins, Users, Crown } from 'lucide-vue-next'
import axios from 'axios'

interface LeaderboardUser {
  id: number
  username: string
  avatar: string
  blog_balance: number
  wallet_address: string
}

const users = ref<LeaderboardUser[]>([])
const loading = ref(true)

const fetchLeaderboard = async () => {
  try {
    const res = await axios.get('http://localhost:8888/api/v1/leaderboard')
    users.value = res.data
  } catch (err) {
    console.error('Failed to fetch leaderboard:', err)
  } finally {
    loading.value = false
  }
}

const getRankStyle = (index: number) => {
  if (index === 0) return 'bg-amber-50 border-amber-200'
  if (index === 1) return 'bg-slate-50 border-slate-200'
  if (index === 2) return 'bg-orange-50 border-orange-200'
  return 'bg-white border-slate-100'
}

const getRankIcon = (index: number) => {
  if (index === 0) return Crown
  if (index === 1) return Medal
  if (index === 2) return Medal
  return null
}

const getRankIconColor = (index: number) => {
  if (index === 0) return 'text-amber-500'
  if (index === 1) return 'text-slate-400'
  if (index === 2) return 'text-orange-400'
  return ''
}

onMounted(fetchLeaderboard)
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-12">
    <div class="text-center mb-10">
      <div class="inline-flex items-center justify-center p-4 bg-indigo-50 rounded-2xl mb-4">
        <Trophy class="w-8 h-8 text-indigo-600" />
      </div>
      <h1 class="text-3xl font-bold text-slate-900 mb-2">代币排行榜</h1>
      <p class="text-slate-500">查看持有 BLOG 代币最多的用户</p>
    </div>

    <div v-if="loading" class="flex flex-col items-center justify-center py-20">
      <div class="animate-spin border-4 border-indigo-600 border-t-transparent rounded-full w-12 h-12 mb-4"></div>
      <p class="text-slate-500 font-medium">加载排行榜...</p>
    </div>

    <div v-else-if="users.length === 0" class="bg-white rounded-3xl border border-slate-100 shadow-sm p-20 text-center">
      <Users class="w-12 h-12 text-slate-300 mx-auto mb-4" />
      <p class="text-slate-500">暂无数据</p>
    </div>

    <div v-else class="space-y-3">
      <div 
        v-for="(user, index) in users" 
        :key="user.id"
        :class="[
          'relative flex items-center gap-4 p-4 rounded-2xl border transition hover:shadow-md',
          getRankStyle(index)
        ]"
      >
        <div class="flex items-center justify-center w-12 h-12 rounded-full bg-white border border-slate-100">
          <component 
            v-if="getRankIcon(index)" 
            :is="getRankIcon(index)" 
            :class="['w-6 h-6', getRankIconColor(index)]"
          />
          <span v-else class="text-lg font-bold text-slate-400">{{ index + 1 }}</span>
        </div>

        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-indigo-100 flex items-center justify-center overflow-hidden">
              <img v-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover" />
              <span v-else class="text-indigo-600 font-bold">{{ user.username?.charAt(0).toUpperCase() }}</span>
            </div>
            <div>
              <p class="font-bold text-slate-900 truncate">{{ user.username }}</p>
              <p class="text-xs text-slate-400 font-mono truncate">{{ user.wallet_address?.slice(0, 8) }}...{{ user.wallet_address?.slice(-4) }}</p>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <Coins class="w-5 h-5 text-indigo-500" />
          <span class="text-xl font-black text-slate-900">{{ user.blog_balance?.toFixed(2) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
