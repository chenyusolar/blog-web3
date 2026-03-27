<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import axios from 'axios'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { 
  FolderOpen, ArrowRight, Search, Share2, CornerUpRight 
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
}

const blogs = ref<Blog[]>([])
const categories = ref<Category[]>([])
const searchQuery = ref('')
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const showShareModal = ref(false)
const selectedBlog = ref<{id: number, title: string} | null>(null)

const fetchBlogs = async () => {
  const categoryId = route.query.category_id
  const res = await axios.get(`http://localhost:8888/api/v1/blogs${categoryId ? '?category_id=' + categoryId : ''}`)
  blogs.value = res.data
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

onMounted(() => {
  fetchBlogs()
  fetchCategories()
})

watch(() => route.query.category_id, fetchBlogs)
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

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <article 
          v-for="blog in filteredBlogs" 
          :key="blog.id" 
          class="bg-white rounded-3xl overflow-hidden shadow-sm hover:shadow-xl transition-all border border-slate-100 group flex flex-col"
        >
          <div class="relative overflow-hidden h-52">
            <img :src="blog.image_url" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
            
            <!-- Badges -->
            <div class="absolute top-4 left-4 flex gap-2">
              <span class="bg-white/90 backdrop-blur px-3 py-1 rounded-full text-[10px] font-bold text-indigo-600 shadow-sm uppercase tracking-wider">
                {{ blog.category?.name || '未分类' }}
              </span>
            </div>

            <!-- Quick Action Share -->
            <button 
              @click.stop.prevent="handleShare(blog)"
              class="absolute top-4 right-4 w-9 h-9 bg-white/90 backdrop-blur rounded-full flex items-center justify-center text-slate-600 hover:bg-indigo-600 hover:text-white transition-all shadow-sm opacity-0 group-hover:opacity-100 translate-y-2 group-hover:translate-y-0"
              title="分享赚钱"
            >
              <Share2 class="w-4 h-4" />
            </button>
          </div>
          
          <div class="p-8 flex flex-col flex-grow">
            <div class="flex flex-wrap gap-2 mb-4">
              <span v-for="tag in blog.tags" :key="tag.id" class="text-[10px] uppercase tracking-wider font-bold bg-slate-100 text-slate-500 px-2 py-0.5 rounded">
                #{{ tag.name }}
              </span>
            </div>
            
            <h2 class="text-xl font-bold mb-4 line-clamp-2 h-14 group-hover:text-indigo-600 transition">
              {{ blog.title }}
            </h2>
            
            <p class="text-slate-500 text-sm mb-6 line-clamp-2 leading-relaxed h-10">
              {{ blog.content }}
            </p>

            <div class="flex items-center justify-between pt-6 border-t border-slate-50 mt-auto">
              <div class="flex items-center gap-2">
                <div class="w-7 h-7 rounded-full bg-indigo-50 border border-indigo-100 flex items-center justify-center text-[10px] font-bold text-indigo-600 uppercase">
                  {{ blog.author?.username?.[0] || '?' }}
                </div>
                <span class="text-xs font-bold text-slate-700">{{ blog.author?.username }}</span>
              </div>
              <router-link :to="`/blog/${blog.id}`" class="text-indigo-600 font-bold text-sm flex items-center gap-1 group/btn">
                阅读文章 <ArrowRight class="w-4 h-4 group-hover/btn:translate-x-1 transition-transform" />
              </router-link>
            </div>
          </div>
        </article>
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
