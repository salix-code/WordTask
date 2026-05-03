import { defineStore } from 'pinia'

// 1. 从 localStorage 读取初始状态
const storedState = JSON.parse(localStorage.getItem('user_state') || '{}')

interface UserState {
  userId: number | null
  accountName: string | null
  loginTimestamp: number | null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    userId: storedState.userId || null,
    accountName: storedState.accountName || null,
    loginTimestamp: storedState.loginTimestamp || null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.userId,
  },
  actions: {
    // 2. 登录后保存状态到 state 和 localStorage
    login(payload: { userId: number; accountName: string }) {
      this.userId = payload.userId
      this.accountName = payload.accountName
      this.loginTimestamp = Date.now()

      localStorage.setItem(
        'user_state',
        JSON.stringify({
          userId: this.userId,
          accountName: this.accountName,
          loginTimestamp: this.loginTimestamp,
        }),
      )
    },
    // 3. 登出时清除 state 和 localStorage
    logout() {
      this.userId = null
      this.accountName = null
      this.loginTimestamp = null
      localStorage.removeItem('user_state')
    },
  },
})
