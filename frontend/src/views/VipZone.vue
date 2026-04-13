<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { ArrowRight, Lock, Crown, ChevronLeft, ChevronRight, AlertCircle, CheckCircle, Clock } from 'lucide-vue-next'
import axios from 'axios'

interface Blog {
  id: number
  title: string
  content: string
  image_url: string
  category?: { name: string }
  tags?: { id: number, name: string }[]
  author?: { username: string }
  created_at: string
  is_vip: boolean
}

const router = useRouter()
const auth = useAuthStore()
const blogs = ref<Blog[]>([])
const loading = ref(true)
const needsVip = ref(false)
const hasPending = ref(false)
const totalBlogs = ref(0)
const currentPage = ref(1)
const pageSize = 18

const fetchVipBlogs = async () => {
  try {
    const res = await axios.get('http://localhost:8888/api/v1/blogs/vip', {
      params: { page: currentPage.value, page_size: pageSize },
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    blogs.value = res.data.blogs
    totalBlogs.value = res.data.total
    needsVip.value = false
  } catch (err: any) {
    if (err.response?.status === 403) {
      needsVip.value = true
      checkVipStatus()
    }
  } finally {
    loading.value = false
  }
}

const checkVipStatus = async () => {
  if (!auth.isLoggedIn) return
  try {
    const res = await axios.get('http://localhost:8888/api/v1/vip/status', {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    hasPending.value = res.data.has_pending
  } catch (err) {
    console.error('Failed to check VIP status:', err)
  }
}

const applyVip = async () => {
  try {
    await axios.post('http://localhost:8888/api/v1/vip/apply', {}, {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    hasPending.value = true
    alert('会员申请已提交，请等待管理员审核')
  } catch (err: any) {
    alert(err.response?.data?.error || '申请失败')
  }
}

const totalPages = computed(() => Math.ceil(totalBlogs.value / pageSize))

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchVipBlogs()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(fetchVipBlogs)
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-12">
    <!-- Header -->
    <div class="text-center mb-10">
      <div class="inline-flex items-center justify-center p-4 bg-amber-50 rounded-2xl mb-4">
        <Crown class="w-8 h-8 text-amber-500" />
      </div>
      <h1 class="text-3xl font-bold text-slate-900 mb-2">会员专区</h1>
      <p class="text-slate-500">会员专享优质内容，解锁更多价值</p>
    </div>

    <!-- VIP Application Section -->
    <div v-if="needsVip && auth.isLoggedIn" class="max-w-md mx-auto mb-12">
      <div class="bg-white rounded-3xl border border-amber-200 shadow-lg overflow-hidden">
        <div class="bg-amber-50 p-6 text-center">
          <Lock class="w-10 h-10 text-amber-500 mx-auto mb-3" />
          <h3 class="text-lg font-bold text-amber-800 mb-1">需要会员权限</h3>
          <p class="text-amber-600 text-sm">升级会员解锁所有专享文章</p>
        </div>
        
        <div v-if="hasPending" class="p-6 text-center">
          <Clock class="w-8 h-8 text-blue-500 mx-auto mb-3" />
          <p class="font-bold text-slate-900 mb-1">申请审核中</p>
          <p class="text-slate-500 text-sm">您的会员申请已提交，请等待管理员审核</p>
        </div>
        
        <div v-else class="p-6">
          <div class="flex items-center justify-between mb-4 p-3 bg-slate-50 rounded-xl">
            <span class="text-sm font-medium text-slate-600">会员费用</span>
            <span class="text-lg font-black text-amber-600">1,000 BLOG</span>
          </div>
          <button 
            @click="applyVip"
            class="w-full bg-amber-500 text-white py-3 rounded-xl font-bold hover:bg-amber-600 transition flex items-center justify-center gap-2 shadow-lg"
          >
            <Crown class="w-4 h-4" /> 申请会员
          </button>
          <p class="text-xs text-slate-400 text-center mt-3">申请后代币将支付给管理员，审核通过后成为会员</p>
        </div>
      </div>
    </div>

    <div v-if="needsVip && !auth.isLoggedIn" class="flex flex-col items-center justify-center py-20">
      <Lock class="w-16 h-16 text-amber-500 mb-4" />
      <h3 class="text-xl font-bold text-slate-900 mb-2">需要登录</h3>
      <p class="text-slate-500 mb-6">请先登录后查看会员专享内容</p>
      <router-link to="/login" class="inline-flex items-center gap-2 px-6 py-3 bg-indigo-600 text-white font-bold rounded-2xl hover:bg-indigo-700 transition shadow-lg">
        立即登录
      </router-link>
    </div>

    <div v-if="loading" class="flex flex-col items-center justify-center py-20">
      <div class="animate-spin border-4 border-amber-500 border-t-transparent rounded-full w-12 h-12 mb-4"></div>
      <p class="text-slate-500 font-medium">加载会员内容...</p>
    </div>

    <div v-else-if="!blogs || blogs.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
      <Crown class="w-16 h-16 mb-4" />
      <p class="font-medium">暂无会员专享文章</p>
    </div>

    <div v-else>
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <article 
          v-for="blog in blogs" 
          :key="blog.id" 
          @click="router.push(`/blog/${blog.id}`)"
          class="bg-white rounded-3xl overflow-hidden shadow-sm hover:shadow-xl transition-all border border-amber-100 group flex flex-col cursor-pointer"
        >
          <div class="relative overflow-hidden h-36 flex-shrink-0">
            <img :src="blog.image_url" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
            <div class="absolute inset-0 bg-gradient-to-t from-black/30 to-transparent"></div>
            
            <div class="absolute top-3 left-3 flex gap-2">
              <span class="bg-amber-500 text-white px-2.5 py-1 rounded-full text-[10px] font-bold shadow-sm flex items-center gap-1">
                <Crown class="w-3 h-3" /> 会员专享
              </span>
            </div>
          </div>
          
          <div class="p-5 flex flex-col flex-grow">
            <h2 class="text-sm font-bold mb-3 leading-snug group-hover:text-amber-600 transition">
              {{ blog.title }}
            </h2>
            
            <div class="flex items-center justify-between mt-auto pt-3 border-t border-slate-50">
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-full bg-amber-50 border border-amber-100 flex items-center justify-center text-[9px] font-bold text-amber-600 uppercase">
                  {{ blog.author?.username?.[0] || '?' }}
                </div>
                <span class="text-[10px] font-bold text-slate-700">{{ blog.author?.username }}</span>
              </div>
              <span class="text-[10px] text-slate-400 font-medium">
                {{ new Date(blog.created_at).toLocaleDateString('zh-CN') }}
              </span>
            </div>
          </div>
        </article>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-3 mt-12">
        <button 
          @click="changePage(currentPage - 1)"
          :disabled="currentPage === 1"
          class="flex items-center gap-1 px-4 py-2 rounded-xl text-sm font-medium transition disabled:opacity-40 disabled:cursor-not-allowed bg-white border border-slate-200 hover:bg-slate-50 text-slate-600 shadow-sm"
        >
          <ChevronLeft class="w-4 h-4" /> 上一页
        </button>
        
        <span class="px-4 py-2 text-sm font-medium text-slate-500 bg-white rounded-xl border border-slate-100">
          第 {{ currentPage }} 页 / 共 {{ totalPages }} 页
        </span>

        <button 
          @click="changePage(currentPage + 1)"
          :disabled="currentPage === totalPages"
          class="flex items-center gap-1 px-4 py-2 rounded-xl text-sm font-medium transition disabled:opacity-40 disabled:cursor-not-allowed bg-white border border-slate-200 hover:bg-slate-50 text-slate-600 shadow-sm"
        >
          下一页 <ChevronRight class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>
</template>
