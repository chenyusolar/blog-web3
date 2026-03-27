<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Clock, Tag, Send, MessageCircle, Share2, RotateCw } from 'lucide-vue-next'

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
}

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const blog = ref<Blog | null>(null)
const comments = ref<Comment[]>([])
const commentContent = ref('')

const fetchBlog = async () => {
  const res = await axios.get(`http://localhost:8888/api/v1/blogs/${route.params.id}`)
  blog.value = res.data.blog
  comments.value = res.data.comments
}

const addComment = async () => {
  if (!commentContent.value) return
  await axios.post('http://localhost:8888/api/v1/comments', {
    blog_id: parseInt(route.params.id as string),
    content: commentContent.value
  })
  commentContent.value = ''
  await fetchBlog()
}

const forwarding = ref(false)
const handleForward = async () => {
  if (!auth.isLoggedIn) {
    router.push('/login')
    return
  }
  forwarding.value = true
  try {
    await axios.post(`http://localhost:8888/api/v1/blogs/${route.params.id}/forward`)
    alert('转发成功！您已获得 BLOG 代币奖励。')
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
          <Clock class="w-3.5 h-3.5" />
          {{ new Date(blog.created_at).toLocaleDateString() }}
        </div>
      </div>
      <h1 class="text-4xl md:text-5xl font-extrabold leading-tight mb-8">{{ blog.title }}</h1>
      <div class="flex items-center gap-3 py-6 border-y border-slate-100">
        <div class="w-10 h-10 rounded-full bg-indigo-50 flex items-center justify-center font-bold text-indigo-600">
          {{ blog.author?.username[0] }}
        </div>
        <div>
          <div class="text-sm font-bold text-slate-800">{{ blog.author?.username }}</div>
          <div class="text-[10px] text-slate-400 uppercase tracking-widest font-bold italic">作者</div>
        </div>
        <button 
          @click="handleForward"
          :disabled="forwarding"
          class="ml-auto bg-slate-900 text-white px-6 py-2 rounded-xl text-sm font-bold flex items-center gap-2 hover:bg-slate-800 transition shadow-lg disabled:opacity-50"
        >
          <RotateCw v-if="forwarding" class="w-4 h-4 animate-spin" />
          <Share2 v-else class="w-4 h-4" />
          {{ forwarding ? '转发中...' : '转发并赚取 BLOG' }}
        </button>
      </div>
    </header>

    <!-- Image -->
    <img :src="blog.image_url" class="w-full h-96 object-cover rounded-3xl mb-12 shadow-lg shadow-indigo-100" />

    <!-- Content -->
    <div class="prose prose-indigo max-w-none text-slate-700 leading-loose text-lg mb-20 whitespace-pre-wrap">
      {{ blog.content }}
    </div>

    <!-- Tags -->
    <div class="flex flex-wrap gap-2 mb-20">
      <div v-for="tag in blog.tags" :key="tag.id" class="flex items-center gap-1.5 bg-slate-50 border border-slate-100 px-4 py-2 rounded-xl text-sm font-semibold text-slate-500">
        <Tag class="w-4 h-4 text-indigo-600" />
        #{{ tag.name }}
      </div>
    </div>

    <!-- Comments -->
    <section class="bg-slate-50/50 rounded-3xl p-8 border border-slate-100 shadow-sm">
      <h3 class="text-xl font-bold mb-8 flex items-center gap-2">
        <MessageCircle class="w-6 h-6 text-indigo-600" />
        评论区 ({{ comments.length }})
      </h3>
      
      <div v-if="auth.isLoggedIn" class="mb-12">
        <textarea 
          v-model="commentContent"
          placeholder="分享你的见解..." 
          class="w-full p-6 rounded-2xl border border-slate-200 focus:outline-none focus:ring-4 focus:ring-indigo-100 min-h-[120px] mb-4 text-sm font-medium"
        ></textarea>
        <button 
          @click="addComment"
          class="bg-indigo-600 text-white px-8 py-3 rounded-xl font-bold text-sm shadow-md hover:shadow-lg transition flex items-center gap-2"
        >
          <Send class="w-4 h-4" /> 发表评论
        </button>
      </div>
      <div v-else class="bg-white border rounded-2xl p-6 text-center text-slate-400 mb-12 italic text-sm">
        请登录后发表评论
      </div>

      <div class="space-y-6">
        <div v-for="c in comments" :key="c.id" class="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm transition hover:shadow-md">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-8 h-8 rounded-full bg-slate-100 flex items-center justify-center font-bold text-slate-500 text-xs uppercase">
              {{ c.user?.username[0] }}
            </div>
            <div class="text-xs font-bold text-slate-700">{{ c.user?.username }}</div>
            <div class="text-[10px] text-slate-400 font-medium ml-auto uppercase">{{ new Date(c.created_at).toLocaleDateString() }}</div>
          </div>
          <p class="text-slate-600 text-sm leading-relaxed">{{ c.content }}</p>
        </div>
      </div>
    </section>
  </div>
</template>
