<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import axios from 'axios'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { 
  FolderOpen, ArrowRight, Search, Share2, CornerUpRight, TrendingUp, Tag, ChevronLeft, ChevronRight, Crown
} from 'lucide-vue-next'
import ShareModal from '../components/ShareModal.vue'

interface Category {
  id: number
  name: string
}

interface Blog {
  id: number
  title: string
  content: string
  image_url: string
  category?: { name: string }
  tags?: { id: number, name: string }[]
  author?: { username: string }
  created_at: string
  share_count: number
  view_count: number
  is_vip: boolean
}

interface Tag {
  id: number
  name: string
  count: number
}

const blogs = ref<Blog[]>([])
const hotBlogs = ref<Blog[]>([])
const hotTags = ref<Tag[]>([])
const categories = ref<Category[]>([])
const searchQuery = ref('')
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const showShareModal = ref(false)
const selectedBlog = ref<{id: number, title: string} | null>(null)

const totalBlogs = ref(0)
const currentPage = ref(1)
const pageSize = 18

const fetchBlogs = async () => {
  const categoryId = route.query.category_id
  const res = await axios.get(`http://localhost:8888/api/v1/blogs`, {
    params: {
      category_id: categoryId,
      page: currentPage.value,
      page_size: pageSize
    }
  })
  blogs.value = res.data.blogs
  totalBlogs.value = res.data.total
}

const fetchHotBlogs = async () => {
  try {
    const res = await axios.get('http://localhost:8888/api/v1/blogs/hot')
    hotBlogs.value = res.data
  } catch (err) {
    console.error('Failed to fetch hot blogs:', err)
  }
}

const fetchHotTags = async () => {
  try {
    const res = await axios.get('http://localhost:8888/api/v1/tags/hot')
    hotTags.value = res.data
  } catch (err) {
    console.error('Failed to fetch hot tags:', err)
  }
}

const fetchCategories = async () => {
  const res = await axios.get('http://localhost:8888/api/v1/categories')
  categories.value = res.data
}

const filteredBlogs = computed(() => {
  if (!searchQuery.value) return blogs.value
  const q = searchQuery.value.toLowerCase()
  return blogs.value.filter(b => 
    b.title.toLowerCase().includes(q) || 
    b.author?.username?.toLowerCase().includes(q)
  )
})

const handleShare = (blog: Blog) => {
  if (!auth.isLoggedIn) {
    router.push('/login')
    return
  }
  selectedBlog.value = { id: blog.id, title: blog.title }
  showShareModal.value = true
}

const totalPages = computed(() => Math.ceil(totalBlogs.value / pageSize))

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchBlogs()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  fetchBlogs()
  fetchCategories()
  fetchHotBlogs()
  fetchHotTags()
})

watch(() => route.query.category_id, () => {
  currentPage.value = 1
  fetchBlogs()
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-12 flex flex-col md:flex-row gap-12">
    <!-- Sidebar -->
    <aside class="w-full md:w-64 flex-shrink-0">
      <!-- Search Box -->
      <div class="mb-10 relative group">
        <input 
          v-model="searchQuery"
          type="text" 
          placeholder="搜索文章或作者..." 
          class="w-full pl-12 pr-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium focus:outline-none focus:ring-4 focus:ring-indigo-100 focus:bg-white transition-all shadow-sm"
        />
        <Search class="absolute left-4 top-3.5 w-4 h-4 text-slate-400 group-focus-within:text-indigo-600 transition-colors" />
      </div>

      <h3 class="font-bold text-lg mb-6 flex items-center gap-2">
        <FolderOpen class="w-5 h-5 text-indigo-600" />
        所有分类
      </h3>
      <div class="flex flex-col gap-2">
        <router-link 
          to="/" 
          class="px-4 py-2.5 rounded-xl transition text-sm font-medium"
          :class="!$route.query.category_id ? 'bg-indigo-600 text-white shadow-md' : 'hover:bg-indigo-50 text-slate-600'"
        >
          全部文章
        </router-link>
        <router-link 
          v-for="cat in categories" 
          :key="cat.id"
          :to="`/?category_id=${cat.id}`"
          class="px-4 py-2.5 rounded-xl transition text-sm font-medium"
          :class="String($route.query.category_id) === String(cat.id) ? 'bg-indigo-600 text-white shadow-md' : 'hover:bg-indigo-50 text-slate-600'"
        >
          {{ cat.name }}
        </router-link>
      </div>

      <!-- Hot Articles Section -->
      <div v-if="hotBlogs.length > 0" class="mt-8 pt-6 border-t border-slate-100">
        <h3 class="font-bold text-xs mb-3 flex items-center gap-1.5 text-slate-700 uppercase tracking-wider">
          <TrendingUp class="w-3.5 h-3.5 text-rose-500" />
          热门文章
        </h3>
        <div class="space-y-1">
          <router-link 
            v-for="(blog, index) in hotBlogs" 
            :key="blog.id"
            :to="`/blog/${blog.id}`"
            class="flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-slate-50 transition group"
          >
            <span class="flex-shrink-0 w-5 h-5 rounded-md bg-slate-100 text-slate-500 text-[10px] font-bold flex items-center justify-center group-hover:bg-indigo-600 group-hover:text-white transition">
              {{ index + 1 }}
            </span>
            <span class="text-xs font-medium text-slate-700 group-hover:text-indigo-600 transition truncate">
              {{ blog.title }}
            </span>
          </router-link>
        </div>
      </div>

      <!-- Hot Tags Section -->
      <div v-if="hotTags.length > 0" class="mt-6 pt-6 border-t border-slate-100">
        <h3 class="font-bold text-xs mb-3 flex items-center gap-1.5 text-slate-700 uppercase tracking-wider">
          <Tag class="w-3.5 h-3.5 text-indigo-500" />
          热门关键词
        </h3>
        <div class="flex flex-wrap gap-1.5">
          <span 
            v-for="tag in hotTags" 
            :key="tag.id"
            class="px-2 py-1 rounded-md bg-slate-50 text-slate-600 text-[10px] font-medium hover:bg-indigo-50 hover:text-indigo-600 transition cursor-pointer"
          >
            #{{ tag.name }}
          </span>
        </div>
      </div>
    </aside>

    <!-- Main -->
    <div class="flex-grow">
      <!-- Empty State -->
      <div v-if="filteredBlogs.length === 0" class="flex flex-col items-center justify-center py-20 text-slate-400">
        <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center mb-6">
          <Search class="w-10 h-10" />
        </div>
        <p class="font-medium">未找到匹配的文章</p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <article 
          v-for="blog in filteredBlogs" 
          :key="blog.id" 
          @click="router.push(`/blog/${blog.id}`)"
          class="bg-white rounded-3xl overflow-hidden shadow-sm hover:shadow-xl transition-all border border-slate-100 group flex flex-col cursor-pointer"
        >
          <div class="relative overflow-hidden h-36 flex-shrink-0">
            <img :src="blog.image_url" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
            
            <!-- Badges -->
            <div class="absolute top-3 left-3 flex gap-1.5">
              <span class="bg-white/90 backdrop-blur px-2.5 py-1 rounded-full text-[10px] font-bold text-indigo-600 shadow-sm uppercase tracking-wider">
                {{ blog.category?.name || '未分类' }}
              </span>
              <span v-if="blog.is_vip" class="bg-amber-500 text-white px-2 py-1 rounded-full text-[10px] font-bold shadow-sm flex items-center gap-0.5">
                <Crown class="w-2.5 h-2.5" /> 会员
              </span>
            </div>

            <!-- Quick Action Share -->
            <button 
              @click.stop.prevent="handleShare(blog)"
              class="absolute top-3 right-3 w-8 h-8 bg-white/90 backdrop-blur rounded-full flex items-center justify-center text-slate-600 hover:bg-indigo-600 hover:text-white transition-all shadow-sm opacity-0 group-hover:opacity-100 translate-y-2 group-hover:translate-y-0"
              title="分享赚钱"
            >
              <Share2 class="w-3.5 h-3.5" />
            </button>
          </div>
          
          <div class="p-5 flex flex-col flex-grow">
            <h2 class="text-sm font-bold mb-3 leading-snug group-hover:text-indigo-600 transition">
              {{ blog.title }}
            </h2>
            
            <div class="flex items-center justify-between mt-auto pt-3 border-t border-slate-50">
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-full bg-indigo-50 border border-indigo-100 flex items-center justify-center text-[9px] font-bold text-indigo-600 uppercase">
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

    <!-- Share Modal -->
    <ShareModal 
      v-if="selectedBlog"
      :show="showShareModal"
      :blog-id="selectedBlog.id"
      :blog-title="selectedBlog.title"
      @close="showShareModal = false"
    />
  </div>
</template>
