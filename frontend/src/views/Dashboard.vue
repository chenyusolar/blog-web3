<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  FileText, Trash2, Eye, Plus, Search, 
  Calendar, ExternalLink, AlertCircle, Sparkles
} from 'lucide-vue-next'
import axios from 'axios'

const router = useRouter()
const blogs = ref<any[]>([])
const loading = ref(true)
const searchQuery = ref('')
const message = ref({ type: '', text: '' })

const fetchMyBlogs = async () => {
  loading.value = true
  try {
    const res = await axios.get('http://localhost:8888/api/v1/user/blogs')
    blogs.value = res.data
  } catch (err) {
    message.value = { type: 'error', text: '加载文章失败' }
  } finally {
    loading.value = false
  }
}

const deleteBlog = async (id: number) => {
  if (!confirm('确定要删除这篇文章吗？此操作不可撤销。')) return
  
  try {
    await axios.delete(`http://localhost:8888/api/v1/blogs/${id}`)
    blogs.value = blogs.value.filter(b => b.id !== id)
    message.value = { type: 'success', text: '文章已删除' }
  } catch (err) {
    message.value = { type: 'error', text: '删除失败' }
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

onMounted(fetchMyBlogs)

// Filtered blogs based on search query
const filteredBlogs = () => {
  if (!searchQuery.value) return blogs.value
  return blogs.value.filter(b => 
    b.title.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    b.category?.name.toLowerCase().includes(searchQuery.value.toLowerCase())
  )
}
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-12">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-10">
      <div>
        <h1 class="text-3xl font-bold text-slate-900 flex items-center gap-3">
          <Sparkles class="w-8 h-8 text-indigo-600" />
          文章管理
        </h1>
        <p class="text-slate-500 mt-1">管理您发布的所有内容，查看数据并持续创作</p>
      </div>
      <router-link to="/editor" class="bg-indigo-600 text-white px-6 py-3 rounded-2xl font-bold flex items-center gap-2 hover:bg-indigo-700 transition shadow-lg shadow-indigo-100 transform hover:-translate-y-1">
        <Plus class="w-5 h-5" />
        创作新文章
      </router-link>
    </div>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-10">
      <div class="bg-white p-6 rounded-3xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-4 bg-indigo-50 text-indigo-600 rounded-2xl">
          <FileText class="w-6 h-6" />
        </div>
        <div>
          <div class="text-2xl font-bold text-slate-900">{{ blogs.length }}</div>
          <div class="text-sm text-slate-500 font-medium">文章总数</div>
        </div>
      </div>
      <div class="bg-white p-6 rounded-3xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-4 bg-emerald-50 text-emerald-600 rounded-2xl">
          <Eye class="w-6 h-6" />
        </div>
        <div>
          <div class="text-2xl font-bold text-slate-900">
            {{ blogs.reduce((acc, b) => acc + (b.view_count || 0), 0) }}
          </div>
          <div class="text-sm text-slate-500 font-medium">总浏览量</div>
        </div>
      </div>
      <div class="bg-white p-6 rounded-3xl border border-slate-100 shadow-sm flex items-center gap-4">
        <div class="p-4 bg-amber-50 text-amber-600 rounded-2xl">
          <Calendar class="w-6 h-6" />
        </div>
        <div>
          <div class="text-2xl font-bold text-slate-900">
            {{ blogs.length > 0 ? formatDate(blogs[0].created_at) : '暂无数据' }}
          </div>
          <div class="text-sm text-slate-500 font-medium">最近发布</div>
        </div>
      </div>
    </div>

    <!-- Search & Filters -->
    <div class="mb-6 flex items-center gap-4">
      <div class="relative flex-grow">
        <Search class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
        <input 
          v-model="searchQuery"
          type="text" 
          placeholder="搜索文章标题、分类..." 
          class="w-full pl-12 pr-4 py-3 bg-white border border-slate-200 rounded-2xl focus:ring-2 focus:ring-indigo-500 transition outline-none text-slate-700 font-medium shadow-sm"
        />
      </div>
    </div>

    <!-- Message Alert -->
    <div v-if="message.text" :class="[
      'mb-6 p-4 rounded-2xl flex items-center justify-between',
      message.type === 'success' ? 'bg-emerald-50 text-emerald-700 border border-emerald-100' : 'bg-rose-50 text-rose-700 border border-rose-100'
    ]">
      <div class="flex items-center gap-3">
        <AlertCircle class="w-5 h-5" />
        <span class="text-sm font-medium">{{ message.text }}</span>
      </div>
      <button @click="message.text = ''" class="text-lg">&times;</button>
    </div>

    <!-- Articles Table -->
    <div class="bg-white rounded-3xl border border-slate-100 shadow-sm overflow-hidden">
      <div v-if="loading" class="p-20 text-center">
        <div class="animate-spin border-4 border-indigo-600 border-t-transparent rounded-full w-12 h-12 mx-auto mb-4"></div>
        <p class="text-slate-500 font-medium">正在获取您的创作...</p>
      </div>

      <div v-else-if="blogs.length === 0" class="p-20 text-center">
        <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center mx-auto mb-6">
          <FileText class="w-10 h-10 text-slate-300" />
        </div>
        <h3 class="text-xl font-bold text-slate-900 mb-2">开启您的第一篇创作</h3>
        <p class="text-slate-500 mb-8">在这里记录您的想法，与世界分享知识</p>
        <router-link to="/editor" class="text-indigo-600 font-bold hover:underline">立即开始写作 &rarr;</router-link>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left">
          <thead>
            <tr class="bg-slate-50/50 border-b border-slate-100">
              <th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">文章详情</th>
              <th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">分类 / 标签</th>
              <th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">发布日期</th>
              <th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider">浏览量</th>
              <th class="px-6 py-4 text-xs font-bold text-slate-400 uppercase tracking-wider text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="blog in filteredBlogs()" :key="blog.id" class="hover:bg-slate-50/50 transition group">
              <td class="px-6 py-5">
                <div class="flex items-center gap-4">
                  <div class="w-16 h-12 rounded-lg bg-slate-100 flex-shrink-0 overflow-hidden">
                    <img v-if="blog.image_url" :src="blog.image_url" class="w-full h-full object-cover" />
                    <FileText v-else class="w-full h-full p-3 text-slate-300" />
                  </div>
                  <div>
                    <h4 class="font-bold text-slate-900 group-hover:text-indigo-600 transition line-clamp-1">{{ blog.title }}</h4>
                    <p class="text-xs text-slate-400 mt-1">ID: #{{ blog.id }}</p>
                  </div>
                </div>
              </td>
              <td class="px-6 py-5">
                <span v-if="blog.category" class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-bold bg-indigo-50 text-indigo-600">
                  {{ blog.category.name }}
                </span>
                <div class="flex flex-wrap gap-1 mt-2">
                  <span v-for="tag in blog.tags" :key="tag.id" class="text-[10px] text-slate-400 font-medium">
                    #{{ tag.name }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-5">
                <div class="text-sm text-slate-600 font-medium">{{ formatDate(blog.created_at) }}</div>
              </td>
              <td class="px-6 py-5">
                <div class="flex items-center gap-1.5 text-sm text-slate-600 font-bold">
                  <Eye class="w-4 h-4 text-slate-400" />
                  {{ blog.view_count || 0 }}
                </div>
              </td>
              <td class="px-6 py-5">
                <div class="flex items-center justify-end gap-2">
                  <button 
                    @click="router.push(`/blog/${blog.id}`)"
                    class="p-2 text-slate-400 hover:text-indigo-600 transition hover:bg-white hover:shadow-sm rounded-lg"
                    title="查看"
                  >
                    <ExternalLink class="w-5 h-5" />
                  </button>
                  <button 
                    @click="deleteBlog(blog.id)"
                    class="p-2 text-slate-400 hover:text-rose-600 transition hover:bg-white hover:shadow-sm rounded-lg"
                    title="删除"
                  >
                    <Trash2 class="w-5 h-5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
