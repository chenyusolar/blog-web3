import { createRouter, createWebHistory, type RouteLocationNormalized, type NavigationGuardNext } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/', component: () => import('../views/Home.vue') },
  { path: '/login', component: () => import('../views/Login.vue') },
  { path: '/register', component: () => import('../views/Register.vue') },
  { path: '/blog/:id', component: () => import('../views/BlogDetail.vue') },
  { 
    path: '/editor', 
    component: () => import('../views/Editor.vue'),
    beforeEnter: (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
      const auth = useAuthStore()
      auth.isLoggedIn ? next() : next('/login')
    }
  },
  { 
    path: '/profile', 
    component: () => import('../views/Profile.vue'),
    beforeEnter: (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
      const auth = useAuthStore()
      auth.isLoggedIn ? next() : next('/login')
    }
  },
  { 
    path: '/dashboard', 
    component: () => import('../views/Dashboard.vue'),
    beforeEnter: (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
      const auth = useAuthStore()
      auth.isLoggedIn ? next() : next('/login')
    }
  },
  { 
    path: '/wallet', 
    component: () => import('../views/Wallet.vue'),
    beforeEnter: (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
      const auth = useAuthStore()
      auth.isLoggedIn ? next() : next('/login')
    }
  },
  { 
    path: '/leaderboard', 
    component: () => import('../views/Leaderboard.vue')
  },
  { 
    path: '/admin', 
    component: () => import('../views/AdminPanel.vue'),
    beforeEnter: (_to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
      const auth = useAuthStore()
      auth.isLoggedIn && auth.user?.role === 'admin' ? next() : next('/')
    }
  }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
