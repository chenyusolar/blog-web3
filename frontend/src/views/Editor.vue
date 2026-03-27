<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { 
  Sparkles, Send, Tag as TagIcon, LayoutGrid, FileImage, Save,
  Bold, Italic, List, ListOrdered, Quote, Heading1, Heading2, Code, Undo, Redo, 
  Strikethrough, Image as ImageIcon, Youtube as YoutubeIcon,
  Table as TableIcon, MinusSquare, Columns, Rows, Trash2, Paperclip, Link as LinkIcon
} from 'lucide-vue-next'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import { Image } from '@tiptap/extension-image'
import { Youtube } from '@tiptap/extension-youtube'
import { Table } from '@tiptap/extension-table'
import { TableRow } from '@tiptap/extension-table-row'
import { TableCell } from '@tiptap/extension-table-cell'
import { TableHeader } from '@tiptap/extension-table-header'
import { Link } from '@tiptap/extension-link'

interface Category {
  id: number
  name: string
}

const router = useRouter()
const route = useRoute()
const categories = ref<Category[]>([])
const loading = ref(false)
const isEditMode = ref(false)
const blogId = ref<string | null>(null)

const fileAttachmentInput = ref<HTMLInputElement | null>(null)

const editor = useEditor({
  content: '',
  extensions: [
    StarterKit,
    Image,
    Youtube.configure({
      width: 480,
      height: 320,
    }),
    Table.configure({
      resizable: true,
    }),
    TableRow,
    TableHeader,
    TableCell,
    Link.configure({
      openOnClick: false,
      HTMLAttributes: {
        rel: 'noopener noreferrer nofollow',
        class: 'text-indigo-600 font-bold underline cursor-pointer',
      },
    }),
  ],
  editorProps: {
    attributes: {
      class: 'prose prose-indigo prose-lg max-w-none focus:outline-none min-h-[450px] p-6 text-slate-700 leading-relaxed',
    },
  },
  onUpdate: ({ editor }) => {
    blogForm.value.content = editor.getHTML()
  },
})

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

const setLink = () => {
  if (!editor.value) return
  
  const previousUrl = editor.value.getAttributes('link').href
  const url = window.prompt('请输入链接 URL', previousUrl)

  // cancelled
  if (url === null) {
    return
  }

  // empty
  if (url === '') {
    editor.value.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }

  // update link
  editor.value.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
}

const addImage = () => {
  const url = window.prompt('请输入图片 URL')
  if (url && editor.value) {
    editor.value.chain().focus().setImage({ src: url }).run()
  }
}

const addYoutubeVideo = () => {
  const url = window.prompt('请输入 YouTube 视频 URL')
  if (url && editor.value) {
    editor.value.chain().focus().setYoutubeVideo({ src: url }).run()
  }
}

const addTable = () => {
  if (editor.value) {
    editor.value.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
  }
}

const triggerFileAttachment = () => {
  fileAttachmentInput.value?.click()
}

const handleFileAttachment = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !editor.value) return

  const formData = new FormData()
  formData.append('file', file)
  
  try {
    const res = await axios.post('http://localhost:8888/api/v1/upload', formData)
    const fileUrl = 'http://localhost:8888' + res.data.file_url
    const linkHtml = `<p><a href="${fileUrl}" class="flex items-center gap-2 text-indigo-600 font-bold hover:underline" target="_blank">🔗 附件: ${file.name}</a></p>`
    editor.value.chain().focus().insertContent(linkHtml).run()
  } catch (err) {
    alert('上传附件失败')
  }
}

const fetchBlogData = async (id: string) => {
  try {
    const res = await axios.get(`http://localhost:8888/api/v1/blogs/${id}`)
    const data = res.data.blog
    blogForm.value = {
      title: data.title,
      content: data.content,
      image_url: data.image_url,
      category_id: data.category_id,
      tag_names: data.tags?.map((t: any) => t.name) || []
    }
    // Update Tiptap content
    if (editor.value) {
      editor.value.commands.setContent(data.content)
    }
  } catch (err) {
    alert('加载文章失败')
    router.push('/dashboard')
  }
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
  
  loading.value = true
  try {
    if (isEditMode.value && blogId.value) {
      await axios.put(`http://localhost:8888/api/v1/blogs/${blogId.value}`, blogForm.value)
    } else {
      await axios.post('http://localhost:8888/api/v1/blogs', blogForm.value)
    }
    router.push('/dashboard')
  } catch (err) {
    alert('保存失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchCategories()
  const id = route.query.id as string
  if (id) {
    isEditMode.value = true
    blogId.value = id
    await fetchBlogData(id)
  }
})
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

      <div class="flex flex-col flex-grow">
        <label class="text-xs font-bold text-slate-400 uppercase tracking-widest px-2">文章正文</label>
        <section class="flex-grow bg-white rounded-3xl shadow-xl shadow-indigo-50/50 border border-indigo-100/50 flex flex-col overflow-hidden">
          <!-- Tiptap Toolbar -->
          <div v-if="editor" class="flex flex-wrap items-center gap-1 p-2 border-b border-indigo-50 bg-slate-50/50 sticky top-0 z-10">
            <button 
              @click="editor.chain().focus().toggleBold().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('bold') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="加粗"
            >
              <Bold class="w-4 h-4" />
            </button>
            <button 
              @click="editor.chain().focus().toggleItalic().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('italic') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="斜体"
            >
              <Italic class="w-4 h-4" />
            </button>
            <button 
              @click="editor.chain().focus().toggleStrike().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('strike') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="删除线"
            >
              <Strikethrough class="w-4 h-4" />
            </button>
            <div class="w-px h-6 bg-slate-200 mx-1"></div>
            <button 
              @click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
              :class="['p-2 rounded-lg transition', editor.isActive('heading', { level: 1 }) ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="一级标题"
            >
              <Heading1 class="w-4 h-4" />
            </button>
            <button 
              @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
              :class="['p-2 rounded-lg transition', editor.isActive('heading', { level: 2 }) ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="二级标题"
            >
              <Heading2 class="w-4 h-4" />
            </button>
            <div class="w-px h-6 bg-slate-200 mx-1"></div>
            <button 
              @click="editor.chain().focus().toggleBulletList().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('bulletList') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="无序列表"
            >
              <List class="w-4 h-4" />
            </button>
            <button 
              @click="editor.chain().focus().toggleOrderedList().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('orderedList') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="有序列表"
            >
              <ListOrdered class="w-4 h-4" />
            </button>
            <div class="w-px h-6 bg-slate-200 mx-1"></div>
            <button 
              @click="editor.chain().focus().toggleBlockquote().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('blockquote') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="引用"
            >
              <Quote class="w-4 h-4" />
            </button>
            <button 
              @click="editor.chain().focus().toggleCode().run()"
              :class="['p-2 rounded-lg transition', editor.isActive('code') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="行内代码"
            >
              <Code class="w-4 h-4" />
            </button>
            <button 
              @click="setLink"
              :class="['p-2 rounded-lg transition', editor.isActive('link') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="插入链接"
            >
              <LinkIcon class="w-4 h-4" />
            </button>
            <div class="w-px h-6 bg-slate-200 mx-1"></div>
            <button 
              @click="addImage"
              class="p-2 text-slate-500 hover:bg-white hover:text-indigo-600 rounded-lg transition"
              title="插入图片 URL"
            >
              <ImageIcon class="w-4 h-4" />
            </button>
            <button 
              @click="addYoutubeVideo"
              class="p-2 text-slate-500 hover:bg-white hover:text-indigo-600 rounded-lg transition"
              title="插入 Youtube 视频"
            >
              <YoutubeIcon class="w-4 h-4" />
            </button>
            <button 
              @click="triggerFileAttachment"
              class="p-2 text-slate-500 hover:bg-white hover:text-indigo-600 rounded-lg transition"
              title="上传附件"
            >
              <Paperclip class="w-4 h-4" />
            </button>
            <input 
              type="file" 
              ref="fileAttachmentInput" 
              class="hidden" 
              @change="handleFileAttachment"
            />
            
            <div class="w-px h-6 bg-slate-200 mx-1"></div>
            <button 
              @click="addTable"
              :class="['p-2 rounded-lg transition', editor.isActive('table') ? 'bg-indigo-600 text-white' : 'text-slate-500 hover:bg-white hover:text-indigo-600']"
              title="插入表格"
            >
              <TableIcon class="w-4 h-4" />
            </button>
            <template v-if="editor.isActive('table')">
              <button @click="editor.chain().focus().addColumnAfter().run()" class="p-2 text-slate-500 hover:text-indigo-600" title="增加列"><Columns class="w-4 h-4" /></button>
              <button @click="editor.chain().focus().addRowAfter().run()" class="p-2 text-slate-500 hover:text-indigo-600" title="增加行"><Rows class="w-4 h-4" /></button>
              <button @click="editor.chain().focus().deleteColumn().run()" class="p-2 text-slate-500 hover:text-rose-600" title="删除列"><MinusSquare class="w-4 h-4 rotate-90" /></button>
              <button @click="editor.chain().focus().deleteRow().run()" class="p-2 text-slate-500 hover:text-rose-600" title="删除行"><MinusSquare class="w-4 h-4" /></button>
              <button @click="editor.chain().focus().deleteTable().run()" class="p-2 text-rose-600" title="删除表格"><Trash2 class="w-4 h-4" /></button>
            </template>

            <div class="w-px h-6 bg-slate-200 mx-1"></div>
            <button 
              @click="editor.chain().focus().undo().run()"
              class="p-2 text-slate-500 hover:bg-white hover:text-indigo-600 rounded-lg transition"
              title="撤销"
            >
              <Undo class="w-4 h-4" />
            </button>
            <button 
              @click="editor.chain().focus().redo().run()"
              class="p-2 text-slate-500 hover:bg-white hover:text-indigo-600 rounded-lg transition"
              title="重做"
            >
              <Redo class="w-4 h-4" />
            </button>
          </div>
          <EditorContent :editor="editor" class="flex-grow overflow-y-auto" />
        </section>
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
          :disabled="loading"
          class="w-full bg-slate-900 text-white font-bold py-5 rounded-3xl text-sm flex items-center justify-center gap-3 hover:bg-slate-800 transition shadow-xl shadow-slate-200"
        >
          <template v-if="loading">
            <span class="animate-spin text-lg">馃攱</span>
            保存中...
          </template>
          <template v-else>
            <Save v-if="isEditMode" class="w-4 h-4" />
            <Send v-else class="w-4 h-4" /> 
            {{ isEditMode ? '提交更新' : '发布博文' }}
          </template>
        </button>
      </section>
    </aside>
  </div>
</template>
