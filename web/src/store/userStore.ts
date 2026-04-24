import { defineStore } from 'pinia'

interface UserState {
  token: string
  userId: string
  account: string
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    token: localStorage.getItem('token') ?? '',
    userId: localStorage.getItem('userId') ?? '',
    account: localStorage.getItem('account') ?? '',
  }),
  getters: {
    isLogin: (s) => Boolean(s.token),
  },
  actions: {
    setAuth(payload: { token: string; userId: string; account: string }) {
      this.token = payload.token
      this.userId = payload.userId
      this.account = payload.account
      localStorage.setItem('token', payload.token)
      localStorage.setItem('userId', payload.userId)
      localStorage.setItem('account', payload.account)
    },
    logout() {
      this.token = ''
      this.userId = ''
      this.account = ''
      localStorage.removeItem('token')
      localStorage.removeItem('userId')
      localStorage.removeItem('account')
    },
  },
})
