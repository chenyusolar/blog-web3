<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { Wallet, Gift, Copy, CheckCircle2, TrendingUp, Loader2, CreditCard, Wallet as WalletIcon, DollarSign } from 'lucide-vue-next'
import axios from 'axios'

interface RewardLog {
  id: number
  type: string
  amount: number
  created_at: string
}

const auth = useAuthStore()
const wallet = ref({ address: '', balance: 0, referral_code: '' })
const rewards = ref<RewardLog[]>([])
const loading = ref(true)
const copied = ref(false)

const buyAmount = ref<number | null>(null)
const buyLoading = ref(false)
const selectedMethod = ref('')

const blogToReceive = computed(() => {
  return buyAmount.value ? buyAmount.value * 10 : 0
})

onMounted(async () => {
  try {
    const w = await auth.fetchWallet()
    wallet.value = w || { address: '', balance: 0, referral_code: '' }
    const r = await auth.fetchRewards()
    rewards.value = Array.isArray(r) ? r : []
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
})

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  copied.value = true
  setTimeout(() => copied.value = false, 2000)
}

const handleBuy = async (method: string) => {
  if (!buyAmount.value || buyAmount.value <= 0) return
  selectedMethod.value = method
  buyLoading.value = true
  try {
    await axios.post('http://localhost:8888/api/v1/user/buy', {
      amount_usd: buyAmount.value,
      payment_method: method
    }, {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    wallet.value.balance += blogToReceive.value
    alert(`成功购买 ${blogToReceive.value} BLOG 代币！`)
    buyAmount.value = null
  } catch (err) {
    alert('购买失败')
  } finally {
    buyLoading.value = false
    selectedMethod.value = ''
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto py-12 px-6">
    <div v-if="loading" class="flex flex-col items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-600 animate-spin" />
      <p class="text-slate-500 mt-4 font-medium">加载钱包数据...</p>
    </div>

    <template v-else>
      <div class="mb-10 flex items-end justify-between">
        <div>
          <h1 class="text-4xl font-black text-slate-900 tracking-tight">我的钱包</h1>
          <p class="text-slate-500 mt-2 font-medium">查看您的代币收益与推广详情</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-10">
        <!-- Balance Card -->
        <div class="md:col-span-2 bg-gradient-to-br from-indigo-600 to-violet-700 rounded-[32px] p-8 text-white shadow-2xl shadow-indigo-200 relative overflow-hidden group">
          <div class="absolute top-0 right-0 p-8 opacity-10 group-hover:scale-110 transition-transform duration-500">
            <Wallet class="w-32 h-32" />
          </div>
          <div class="relative z-10">
            <p class="text-indigo-100 font-bold mb-2 uppercase tracking-widest text-xs">BLOG Token Balance</p>
            <div class="flex items-baseline gap-3">
              <h2 class="text-6xl font-black tabular-nums">{{ wallet?.balance?.toFixed(2) || '0.00' }}</h2>
              <span class="text-2xl font-bold opacity-80 uppercase">Blog</span>
            </div>
            
            <div class="mt-8 pt-8 border-t border-white/10 flex items-center justify-between">
              <div>
                <p class="text-indigo-100 text-xs font-bold uppercase mb-1 opacity-60">Solana Wallet Address</p>
                <div class="flex items-center gap-2 group/addr cursor-pointer" @click="copyToClipboard(wallet?.address || '')">
                  <span class="font-mono text-sm opacity-90 truncate max-w-[200px] sm:max-w-none">{{ wallet?.address || '未生成' }}</span>
                  <Copy v-if="!copied" class="w-4 h-4 opacity-50 group-hover/addr:opacity-100 transition-opacity" />
                  <CheckCircle2 v-else class="w-4 h-4 text-emerald-400" />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Referral Card -->
        <div class="bg-white rounded-[32px] p-8 border border-slate-100 shadow-xl shadow-slate-200/50 flex flex-col justify-between">
          <div>
            <div class="w-12 h-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center mb-6">
              <Gift class="w-6 h-6" />
            </div>
            <h3 class="text-xl font-bold text-slate-900 mb-2">推广奖励</h3>
            <p class="text-slate-500 text-sm font-medium mb-6">分享您的邀请码，赚取三级返佣！</p>
          </div>
          
          <div class="space-y-4">
            <div class="p-4 bg-slate-50 rounded-2xl border border-slate-100">
              <p class="text-xs font-bold text-slate-400 uppercase mb-1">您的邀请码</p>
              <div class="flex items-center justify-between">
                <span class="text-xl font-black text-indigo-600 tracking-wider font-mono">{{ wallet?.referral_code || '---' }}</span>
                <button @click="copyToClipboard(wallet?.referral_code || '')" class="text-slate-400 hover:text-indigo-600 transition">
                  <Copy class="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Buy BLOG Section -->
      <div class="bg-white rounded-[40px] border border-slate-100 shadow-xl shadow-slate-100 overflow-hidden mb-10">
        <div class="p-8 border-b border-slate-50 flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
            <DollarSign class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-xl font-bold text-slate-900">购买 BLOG 代币</h3>
            <p class="text-slate-500 text-sm font-medium">1 USD = 10 BLOG</p>
          </div>
        </div>
        
        <div class="p-8">
          <div class="flex flex-col md:flex-row gap-8 items-start">
            <div class="flex-grow space-y-6 w-full">
              <div class="flex flex-col gap-2">
                <label class="text-xs font-bold text-slate-500 uppercase ml-1">支付金额 (USD)</label>
                <div class="relative">
                  <span class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 font-bold">$</span>
                  <input 
                    v-model.number="buyAmount"
                    type="number" 
                    min="1"
                    placeholder="输入金额..." 
                    class="w-full pl-8 pr-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl focus:ring-2 focus:ring-indigo-600/20 focus:bg-white transition-all font-medium text-slate-900 outline-none text-sm"
                  />
                </div>
              </div>

              <div v-if="blogToReceive > 0" class="p-4 bg-indigo-50 border border-indigo-100 rounded-2xl flex items-center justify-between">
                <span class="text-sm font-bold text-indigo-800">您将获得</span>
                <span class="text-2xl font-black text-indigo-600">{{ blogToReceive }} BLOG</span>
              </div>

              <div class="flex flex-col gap-3">
                <p class="text-xs font-bold text-slate-500 uppercase ml-1">选择支付方式</p>
                <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
                  <button 
                    @click="handleBuy('stripe')"
                    :disabled="buyLoading || !buyAmount"
                    class="flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 hover:border-indigo-500 hover:text-indigo-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <CreditCard class="w-4 h-4" /> Stripe
                  </button>
                  <button 
                    @click="handleBuy('paypal')"
                    :disabled="buyLoading || !buyAmount"
                    class="flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 hover:border-indigo-500 hover:text-indigo-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <CreditCard class="w-4 h-4" /> PayPal
                  </button>
                  <button 
                    @click="handleBuy('wise')"
                    :disabled="buyLoading || !buyAmount"
                    class="flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 hover:border-indigo-500 hover:text-indigo-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <CreditCard class="w-4 h-4" /> Wise
                  </button>
                  <button 
                    @click="handleBuy('metamask')"
                    :disabled="buyLoading || !buyAmount"
                    class="flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 hover:border-indigo-500 hover:text-indigo-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <WalletIcon class="w-4 h-4" /> MetaMask
                  </button>
                  <button 
                    @click="handleBuy('okx')"
                    :disabled="buyLoading || !buyAmount"
                    class="flex items-center justify-center gap-2 px-4 py-3 bg-white border border-slate-200 rounded-xl text-sm font-bold text-slate-700 hover:border-indigo-500 hover:text-indigo-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <WalletIcon class="w-4 h-4" /> OKX
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- History -->
      <div class="bg-white rounded-[40px] border border-slate-100 shadow-xl shadow-slate-100 overflow-hidden">
        <div class="p-8 border-b border-slate-50 flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
              <TrendingUp class="w-5 h-5" />
            </div>
            <h3 class="text-xl font-bold text-slate-900">奖励历史</h3>
          </div>
        </div>
        
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-slate-50/50">
                <th class="px-8 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">动作类型</th>
                <th class="px-8 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">金额 (BLOG)</th>
                <th class="px-8 py-4 text-xs font-bold text-slate-400 uppercase tracking-widest">时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              <tr v-for="log in rewards" :key="log.id" class="hover:bg-slate-50/50 transition">
                <td class="px-8 py-6">
                  <div class="flex items-center gap-2">
                    <span class="w-2 h-2 rounded-full" :class="log.type.startsWith('Referral') ? 'bg-amber-400' : 'bg-indigo-400'"></span>
                    <span class="font-bold text-slate-700">{{ log.type }}</span>
                  </div>
                </td>
                <td class="px-8 py-6 font-black text-slate-900">+{{ log.amount?.toFixed(2) }}</td>
                <td class="px-8 py-6 text-slate-400 text-sm">{{ new Date(log.created_at).toLocaleString() }}</td>
              </tr>
              <tr v-if="rewards.length === 0">
                <td colspan="3" class="px-8 py-20 text-center text-slate-400 font-medium">暂无奖励记录，开始创作或推广吧！</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
