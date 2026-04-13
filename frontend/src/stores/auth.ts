import { defineStore } from 'pinia'
import axios from 'axios'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    token: localStorage.getItem('token') || '',
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
  actions: {
    async login(username: string, password: string) {
      const res = await axios.post('http://localhost:8888/api/v1/login', { username, password })
      this.user = res.data.user
      this.token = res.data.token
      localStorage.setItem('user', JSON.stringify(res.data.user))
      localStorage.setItem('token', res.data.token)
    },
    async register(username: string, password: string, email: string, inviterCode?: string, importWallet?: string) {
      await axios.post('http://localhost:8888/api/v1/register', { username, password, email, inviter_code: inviterCode, import_wallet: importWallet })
    },
    async fetchWallet() {
      const res = await axios.get('http://localhost:8888/api/v1/user/wallet')
      return res.data
    },
    async fetchRewards() {
      const res = await axios.get('http://localhost:8888/api/v1/user/rewards')
      return res.data
    },
    async fetchProfile() {
      const res = await axios.get('http://localhost:8888/api/v1/user/profile')
      this.user = res.data
      localStorage.setItem('user', JSON.stringify(res.data))
    },
    async updateProfile(data: { username?: string, email?: string, avatar?: string }) {
      const res = await axios.put('http://localhost:8888/api/v1/user/profile', data)
      this.user = res.data
      localStorage.setItem('user', JSON.stringify(res.data))
    },
    async changePassword(oldPassword: string, newPassword: string) {
      await axios.put('http://localhost:8888/api/v1/user/password', { old_password: oldPassword, new_password: newPassword })
    },
    logout() {
      this.user = null
      this.token = ''
      localStorage.removeItem('user')
      localStorage.removeItem('token')
    }
  }
})

// Axios 拦截器注入 Token
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = token
  }
  return config
})
