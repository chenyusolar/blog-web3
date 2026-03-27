<script setup lang="ts">
import { ref, computed } from 'vue'
import { 
  X, Twitter, Send, Facebook, Copy, Check, ExternalLink 
} from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
import axios from 'axios'

const props = defineProps<{
  show: boolean
  blogId: number
  blogTitle: string
}>()

const emit = defineEmits(['close'])

const auth = useAuthStore()
const copied = ref(false)

const shareUrl = computed(() => {
  const baseUrl = window.location.origin
  const refCode = auth.user?.referral_code || ''
  return `${baseUrl}/blog/${props.blogId}${refCode ? '?ref=' + refCode : ''}`
})

const networks = [
  { name: 'Twitter', icon: Twitter, color: 'bg-black text-white', shareUrl: (url: string, title: string) => `https://twitter.com/intent/tweet?url=${encodeURIComponent(url)}&text=${encodeURIComponent(title)}` },
  { name: 'Telegram', icon: Send, color: 'bg-[#229ED9] text-white', shareUrl: (url: string, title: string) => `https://t.me/share/url?url=${encodeURIComponent(url)}&text=${encodeURIComponent(title)}` },
  { name: 'Facebook', icon: Facebook, color: 'bg-[#1877F2] text-white', shareUrl: (url: string, _title: string) => `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}` }
]

const handleShareAction = async (networkName: string) => {
  if (auth.isLoggedIn) {
    try {
      await axios.post(`http://localhost:8888/api/v1/blogs/${props.blogId}/share`)
    } catch (err) {
      console.error('Failed to notify backend of share:', err)
    }
  }
  
  const network = networks.find(n => n.name === networkName)
  if (network) {
    window.open(network.shareUrl(shareUrl.value, props.blogTitle), '_blank')
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
    
    // Also record share when copying link
    if (auth.isLoggedIn) {
      await axios.post(`http://localhost:8888/api/v1/blogs/${props.blogId}/share`)
    }
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm" @click="$emit('close')"></div>
    
    <!-- Modal -->
    <div class="relative w-full max-w-sm bg-white rounded-[32px] shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
      <div class="p-8">
        <div class="flex items-center justify-between mb-8">
          <h3 class="text-xl font-bold text-slate-900">分享这篇文章</h3>
          <button @click="$emit('close')" class="p-2 hover:bg-slate-100 rounded-full transition">
            <X class="w-5 h-5 text-slate-400" />
          </button>
        </div>

        <p class="text-slate-500 text-sm mb-8 font-medium">
          分享到社交平台，吸引更多读者并获取 <span class="text-indigo-600 font-bold">BLOG</span> 代币奖励！
        </p>

        <!-- Social Networks -->
        <div class="grid grid-cols-3 gap-4 mb-10">
          <button 
            v-for="network in networks" 
            :key="network.name"
            @click="handleShareAction(network.name)"
            class="flex flex-col items-center gap-3 group"
          >
            <div :class="[network.color, 'w-14 h-14 rounded-2xl flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform duration-300']">
              <component :is="network.icon" class="w-6 h-6" />
            </div>
            <span class="text-xs font-bold text-slate-600">{{ network.name }}</span>
          </button>
        </div>

        <!-- Copy Link Section -->
        <div class="relative group">
          <div class="flex items-center gap-3 p-4 bg-slate-50 border border-slate-100 rounded-2xl">
            <div class="flex-grow overflow-hidden">
              <p class="text-[10px] text-slate-400 font-bold uppercase tracking-widest mb-1">专属推广链接</p>
              <p class="text-xs font-medium text-slate-600 truncate">{{ shareUrl }}</p>
            </div>
            <button 
              @click="copyToClipboard"
              class="flex-shrink-0 w-10 h-10 bg-white border border-slate-200 rounded-xl flex items-center justify-center text-slate-600 hover:text-indigo-600 hover:border-indigo-100 transition shadow-sm"
              :class="{ 'text-emerald-500 border-emerald-100': copied }"
            >
              <Check v-if="copied" class="w-5 h-5" />
              <Copy v-else class="w-5 h-5" />
            </button>
          </div>
          
          <div v-if="copied" class="absolute -top-10 left-1/2 -translate-x-1/2 bg-slate-900 text-white text-[10px] font-bold px-3 py-1.5 rounded-lg shadow-xl animate-in fade-in slide-in-from-bottom-2 duration-300">
            链接已复制到剪贴板
          </div>
        </div>
      </div>
      
      <!-- Footer Info -->
      <div class="bg-indigo-600 p-4 text-center">
        <p class="text-[10px] text-indigo-100 font-bold uppercase tracking-widest flex items-center justify-center gap-2">
          <ExternalLink class="w-3 h-3" />
          通过社交媒体曝光提升你的 Web3 影响力
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-in {
  animation: animate-in 0.2s ease-out;
}

@keyframes animate-in {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
