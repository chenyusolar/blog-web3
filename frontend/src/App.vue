<script setup lang="ts">
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'
import { Sparkles, User, LogOut, PenLine, Home, FileText, Wallet, ShieldCheck } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()

const logout = () => {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex flex-col">
    <!-- Navbar -->
    <nav class="bg-white border-b sticky top-0 z-20 shadow-sm">
      <div class="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
        <router-link to="/" class="flex items-center gap-2 font-bold text-xl text-indigo-600">
          <Sparkles class="w-6 h-6" />
          <span>AIGen-Blog</span>
        </router-link>

        <div class="flex items-center gap-6">
          <router-link to="/" class="text-slate-600 hover:text-indigo-600 flex items-center gap-1 font-medium transition">
            <Home class="w-4 h-4" /> 首页
          </router-link>
          
          <template v-if="auth.isLoggedIn">
            <router-link to="/editor" class="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-bold flex items-center gap-2 hover:bg-indigo-700 transition shadow-sm">
              <PenLine class="w-4 h-4" /> 发布
            </router-link>
            
            <div class="group relative h-full flex items-center gap-2 cursor-pointer border-l pl-6 ml-2">
              <div class="w-8 h-8 rounded-full bg-slate-200 flex items-center justify-center text-slate-500 overflow-hidden">
                <img v-if="auth.user?.avatar" :src="auth.user.avatar" class="w-full h-full object-cover" />
                <User v-else class="w-5 h-5" />
              </div>
              <span class="text-sm font-semibold text-slate-700">{{ auth.user?.username }}</span>
              
              <!-- Dropdown Menu -->
              <div class="absolute right-0 top-[calc(100%-10px)] pt-[10px] w-48 opacity-0 group-hover:opacity-100 transition-all pointer-events-none group-hover:pointer-events-auto transform origin-top translate-y-2 group-hover:translate-y-0">
                <div class="bg-white border rounded-xl shadow-xl py-2">
                  <router-link to="/profile" class="w-full text-left px-4 py-2 text-sm text-slate-600 hover:bg-slate-50 flex items-center gap-2">
                  <User class="w-4 h-4" /> 个人设置
                </router-link>
                <router-link to="/dashboard" class="w-full text-left px-4 py-2 text-sm text-slate-600 hover:bg-slate-50 flex items-center gap-2">
                  <FileText class="w-4 h-4" /> 文章管理
                </router-link>
                <router-link to="/wallet" class="w-full text-left px-4 py-2 text-sm text-slate-600 hover:bg-slate-50 flex items-center gap-2">
                  <Wallet class="w-4 h-4" /> 我的钱包
                </router-link>
                <router-link v-if="auth.user?.role === 'admin'" to="/admin" class="w-full text-left px-4 py-2 text-sm text-indigo-600 hover:bg-indigo-50 flex items-center gap-2">
                  <ShieldCheck class="w-4 h-4" /> 管理面板
                </router-link>
                <div class="h-px bg-slate-100 my-1 mx-2"></div>
                <button @click="logout" class="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2">
                  <LogOut class="w-4 h-4" /> 退出登录
                </button>
              </div>
            </div>
          </div>
        </template>

          <template v-else>
            <router-link to="/login" class="text-slate-600 hover:text-indigo-600 text-sm font-bold">登录</router-link>
            <router-link to="/register" class="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-bold hover:bg-indigo-700 transition">注册</router-link>
          </template>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="flex-grow">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t py-12">
      <div class="max-w-6xl mx-auto px-4 text-center text-slate-400 text-sm">
        &copy; 2026 AIGen-Blog. Powered by CloudWeGo Hertz & Eino AI.
      </div>
    </footer>
  </div>
</template>

<style>
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
