<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { 
  Calendar, Tag, MessageCircle, Send, Share2, CornerUpRight 
} from 'lucide-vue-next'

interface Comment {
  id: number
  blog_id: number
  user_id: number
  content: string
  created_at: string
  user: { username: string }
}

interface Blog {
  id: number
  title: string
  content: string
  image_url: string
  created_at: string
  author: { username: string }
  category: { name: string }
  tags: { id: number, name: string }[]
  is_forward: boolean
  original_blog: Blog | null
}

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const blog = ref<Blog | null>(null)
const comments = ref<Comment[]>([])
const commentContent = ref('')

const fetchBlog = async () => {
  try {
    const res = await axios.get(`http://localhost:8888/api/v1/blogs/${route.params.id}`)
    blog.value = res.data.blog
    comments.value = res.data.comments
  } catch (err) {
    console.error('Failed to fetch blog:', err)
  }
}

const addComment = async () => {
  if (!commentContent.value || !auth.isLoggedIn) return
  try {
    await axios.post('http://localhost:8888/api/v1/comments', {
      blog_id: parseInt(route.params.id as string),
      content: commentContent.value
    }, {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    commentContent.value = ''
    await fetchBlog()
  } catch (err) {
    alert('评论失败')
  }
}

const forwarding = ref(false)
const handleForward = async () => {
  if (!auth.isLoggedIn) {
    router.push('/login')
    return
  }
  forwarding.value = true
  try {
    await axios.post(`http://localhost:8888/api/v1/blogs/${route.params.id}/forward`, {}, {
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    alert('转发成功！您已获得 BLOG 代币奖励。')
    router.push('/dashboard')
  } catch (err) {
    console.error(err)
    alert('转发失败')
  } finally {
    forwarding.value = false
  }
}

onMounted(fetchBlog)
</script>

<template>
  <div v-if="blog" class="max-w-4xl mx-auto px-4 py-16">
    <!-- Header -->
    <header class="mb-12">
      <div class="flex items-center gap-2 mb-6">
        <span class="bg-indigo-600 text-white px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider">
          {{ blog.category?.name || '未分类' }}
        </span>
        <div class="flex items-center gap-2 text-slate-400 text-xs font-medium ml-4">
          <Calendar class="w-3.5 h-3.5" />
          {{ new Date(blog.created_at).toLocaleDateString() }}
        </div>
      </div>
      <h1 class="text-4xl md:text-5xl font-extrabold leading-tight mb-8">{{ blog.title }}</h1>
      
      <!-- Author & Forward Action -->
      <div class="flex flex-col md:flex-row md:items-center gap-6 justify-between p-6 bg-slate-50 solid rounded-3xl border border-slate-100 mb-12 shadow-sm">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 rounded-full bg-indigo-50 border-2 border-white shadow-sm flex items-center justify-center font-bold text-indigo-600 text-xl">
            {{ blog.author?.username[0] }}
          </div>
          <div>
            <div class="font-bold text-slate-900 leading-tight">{{ blog.author?.username }}</div>
            <div v-if="blog.is_forward && blog.original_blog" class="text-xs text-indigo-600 font-bold flex items-center gap-1 mt-0.5">
              <CornerUpRight class="w-3 h-3" />
              转发自: {{ blog.original_blog.author?.username }}
            </div>
            <div v-else class="text-xs text-slate-400 font-medium mt-0.5 uppercase tracking-widest italic">主创作者</div>
          </div>
        </div>
        
        <div class="flex items-center gap-2">
          <button 
            @click="handleForward"
            :disabled="forwarding"
            class="flex items-center gap-2 px-6 py-3 bg-white border border-slate-200 text-slate-700 font-bold rounded-2xl hover:bg-slate-50 hover:border-indigo-100 hover:text-indigo-600 transition-all shadow-sm disabled:opacity-50"
          >
            <Share2 class="w-4 h-4" :class="{ 'animate-pulse': forwarding }" />
            <span>{{ forwarding ? '处理中...' : '一键转发' }}</span>
          </button>
        </div>
      </div>
    </header>

    <!-- Image -->
    <div class="relative w-full h-96 rounded-3xl overflow-hidden mb-12 shadow-2xl">
      <img :src="blog.image_url" class="w-full h-full object-cover" />
      <div class="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent"></div>
    </div>

    <!-- Content -->
    <div class="prose prose-indigo prose-lg max-w-none text-slate-700 mb-20" v-html="blog.content">
    </div>

    <!-- Tags -->
    <div class="flex flex-wrap gap-2 mb-20">
      <div v-for="tag in blog.tags" :key="tag.id" class="flex items-center gap-1.5 bg-slate-50 border border-slate-100 px-4 py-2 rounded-xl text-sm font-semibold text-slate-500">
        <Tag class="w-4 h-4 text-indigo-600" />
        #{{ tag.name }}
      </div>
    </div>

    <!-- Comments -->
    <section class="bg-slate-50/50 rounded-3xl p-8 border border-slate-100 shadow-sm mb-20">
      <h3 class="text-xl font-bold mb-8 flex items-center gap-2 font-display">
        <MessageCircle class="w-6 h-6 text-indigo-600" />
        交流探讨 ({{ comments.length }})
      </h3>
      
      <div v-if="auth.isLoggedIn" class="mb-12">
        <textarea 
          v-model="commentContent"
          placeholder="分享你的见解或提问..." 
          class="w-full p-6 rounded-2xl border border-slate-200 focus:outline-none focus:ring-4 focus:ring-indigo-100 min-h-[120px] mb-4 text-sm font-medium transition-all"
        ></textarea>
        <button 
          @click="addComment"
          class="bg-indigo-600 text-white px-8 py-3 rounded-xl font-bold text-sm shadow-indigo-200 shadow-lg hover:shadow-indigo-300 transition-all flex items-center gap-2 transform active:scale-95"
        >
          <Send class="w-4 h-4" /> 发表评论
        </button>
      </div>
      <div v-else class="bg-white border border-slate-100 rounded-2xl p-8 text-center text-slate-400 mb-12 italic text-sm">
        欢迎加入讨论，请先登录后发表评论
      </div>

      <div class="space-y-6">
        <div v-for="c in comments" :key="c.id" class="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm transition hover:shadow-md">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-8 h-8 rounded-full bg-slate-100 border border-slate-200 flex items-center justify-center font-bold text-slate-500 text-xs uppercase">
              {{ c.user?.username[0] }}
            </div>
            <div class="text-xs font-bold text-slate-700">{{ c.user?.username }}</div>
            <div class="text-[10px] text-slate-400 font-medium ml-auto uppercase tracking-wider">{{ new Date(c.created_at).toLocaleDateString() }}</div>
          </div>
          <p class="text-slate-600 text-sm leading-relaxed">{{ c.content }}</p>
        </div>
      </div>
    </section>
  </div>
</template>
