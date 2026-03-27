<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { Sparkles, Send, Tag as TagIcon, LayoutGrid, FileImage } from 'lucide-vue-next'

interface Category {
  id: number
  name: string
}

const router = useRouter()
const categories = ref<Category[]>([])
const loading = ref(false)

const blogForm = ref({
  title: '',
  content: '',
  image_url: '',
  category_id: null as number | null,
  tag_names: [] as string[]
})

const tagInput = ref('')
const prompt = ref('')

const fetchCategories = async () => {
  const res = await axios.get('http://localhost:8888/api/v1/categories')
  categories.value = res.data
}

const addTag = () => {
  if (tagInput.value && !blogForm.value.tag_names.includes(tagInput.value)) {
    blogForm.value.tag_names.push(tagInput.value)
    tagInput.value = ''
  }
}

const removeTag = (tag: string) => {
  blogForm.value.tag_names = blogForm.value.tag_names.filter(t => t !== tag)
}

const generateWithAI = async () => {
  if (!prompt.value) return
  loading.value = true
  try {
    const res = await axios.post('http://localhost:8888/api/v1/blogs/generate', { prompt: prompt.value })
    blogForm.value.title = res.data.title
    blogForm.value.content = res.data.content
    blogForm.value.image_url = res.data.image_url
  } finally {
    loading.value = false
  }
}

const publishBlog = async () => {
  if (!blogForm.value.title || !blogForm.value.content || !blogForm.value.category_id) {
    alert('请填写标题、正文并选择分类')
    return
  }
  await axios.post('http://localhost:8888/api/v1/blogs', blogForm.value)
  router.push('/')
}

onMounted(fetchCategories)
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-12 flex flex-col lg:flex-row gap-12 min-h-[calc(100vh-80px)]">
    <!-- Main Editor -->
    <div class="flex-grow bg-white rounded-3xl p-8 border border-slate-100 shadow-xl shadow-slate-100/50 flex flex-col gap-8">
      <div class="flex flex-col gap-2">
        <label class="text-xs font-bold text-slate-400 uppercase tracking-widest px-2">文章标题</label>
        <input 
          v-model="blogForm.title" 
          type="text" 
          placeholder="给你的博文起个响亮的标题..." 
          class="text-3xl font-extrabold p-4 rounded-2xl border border-transparent focus:border-indigo-100 focus:outline-none focus:ring-4 focus:ring-indigo-50/50 placeholder:text-slate-200 transition"
        />
      </div>

      <div class="flex flex-col gap-2 flex-grow">
        <label class="text-xs font-bold text-slate-400 uppercase tracking-widest px-2">文章正文</label>
        <textarea 
          v-model="blogForm.content" 
          placeholder="开始你的创作..." 
          class="flex-grow p-6 rounded-2xl border border-transparent focus:border-indigo-100 focus:outline-none focus:ring-4 focus:ring-indigo-50/50 min-h-[400px] text-lg text-slate-700 leading-relaxed transition"
        ></textarea>
      </div>
    </div>

    <!-- Sidebar Tools -->
    <aside class="w-full lg:w-96 flex flex-col gap-8">
      <!-- AI Generator -->
      <section class="bg-indigo-600 rounded-3xl p-8 text-white shadow-xl shadow-indigo-100">
        <h3 class="font-bold text-lg mb-4 flex items-center gap-2">
          <Sparkles class="w-5 h-5" />
          AI 智能助手
        </h3>
        <p class="text-xs opacity-80 mb-6 font-medium leading-relaxed">描述你想写的主题，Eino AI 将为你生成高质量内容、标题及配图。</p>
        <textarea 
          v-model="prompt"
          placeholder="例如: 关于 AI 在未来 10 年对艺术创作的影响..." 
          class="w-full bg-white/10 backdrop-blur-md p-4 rounded-2xl border border-white/20 text-sm placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-white/30 transition mb-4 min-h-[100px]"
        ></textarea>
        <button 
          @click="generateWithAI"
          :disabled="loading"
          class="w-full bg-white text-indigo-600 font-bold py-4 rounded-2xl text-sm flex items-center justify-center gap-2 hover:bg-indigo-50 transition shadow-lg shadow-indigo-800/20 disabled:opacity-50"
        >
          <span v-if="loading" class="animate-spin text-lg">馃攱</span>
          {{ loading ? '生成中...' : '一键生成' }}
        </button>
      </section>

      <!-- Settings -->
      <section class="bg-white rounded-3xl p-8 border border-slate-100 shadow-sm flex flex-col gap-8">
        <!-- Category -->
        <div class="flex flex-col gap-4">
          <h4 class="font-bold text-sm text-slate-800 flex items-center gap-2">
            <LayoutGrid class="w-4 h-4 text-indigo-600" />
            选择分类
          </h4>
          <select 
            v-model="blogForm.category_id"
            class="w-full p-4 rounded-2xl bg-slate-50 border border-slate-100 text-sm font-bold text-slate-600 focus:outline-none focus:ring-4 focus:ring-indigo-50/50 transition appearance-none"
          >
            <option :value="null" disabled>请选择分类</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
        </div>

        <!-- Tags -->
        <div class="flex flex-col gap-4">
          <h4 class="font-bold text-sm text-slate-800 flex items-center gap-2">
            <TagIcon class="w-4 h-4 text-indigo-600" />
            文章标签
          </h4>
          <div class="flex flex-wrap gap-2 mb-2">
            <span v-for="tag in blogForm.tag_names" :key="tag" class="bg-indigo-50 text-indigo-600 text-xs font-bold px-3 py-1.5 rounded-xl border border-indigo-100 flex items-center gap-2">
              {{ tag }}
              <button @click="removeTag(tag)" class="hover:text-red-500 font-black">脳</button>
            </span>
          </div>
          <div class="flex gap-2">
            <input 
              v-model="tagInput" 
              @keyup.enter="addTag"
              type="text" 
              placeholder="输入标签按回车..." 
              class="flex-grow p-4 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-medium focus:outline-none transition"
            />
          </div>
        </div>

        <!-- Image URL -->
        <div class="flex flex-col gap-4">
          <h4 class="font-bold text-sm text-slate-800 flex items-center gap-2">
            <FileImage class="w-4 h-4 text-indigo-600" />
            封面配图
          </h4>
          <input 
            v-model="blogForm.image_url"
            type="text" 
            placeholder="图片 URL (AI 自动生成或手动填写)" 
            class="w-full p-4 bg-slate-50 border border-slate-100 rounded-2xl text-xs font-medium focus:outline-none transition"
          />
          <div v-if="blogForm.image_url" class="rounded-2xl overflow-hidden border border-slate-100 aspect-video">
            <img :src="blogForm.image_url" class="w-full h-full object-cover" />
          </div>
        </div>

        <button 
          @click="publishBlog"
          class="w-full bg-slate-900 text-white font-bold py-5 rounded-3xl text-sm flex items-center justify-center gap-3 hover:bg-slate-800 transition shadow-xl shadow-slate-200"
        >
          <Send class="w-4 h-4" /> 发布博文
        </button>
      </section>
    </aside>
  </div>
</template>
