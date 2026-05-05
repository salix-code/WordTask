import { defineStore } from 'pinia'

const USER_STATE_KEY = 'user_state'

// 1. 读取初始状态：管理员优先从 sessionStorage（关闭网页即失效），普通用户从 localStorage
const storedState = JSON.parse(
  sessionStorage.getItem(USER_STATE_KEY) || localStorage.getItem(USER_STATE_KEY) || '{}',
)

// 历史数据兼容：若管理员状态残留在 localStorage，清理掉以确保“关闭网页即退出”
if (storedState.isAdmin) {
  localStorage.removeItem(USER_STATE_KEY)
}

interface UserState {
  userId: number | null
  accountName: string | null
  isAdmin: boolean
  loginTimestamp: number | null
}

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    userId: storedState.userId || null,
    accountName: storedState.accountName || null,
    isAdmin: Boolean(storedState.isAdmin),
    loginTimestamp: storedState.loginTimestamp || null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.userId,
  },
  actions: {
    // 2. 登录后保存状态：管理员写 sessionStorage，普通用户写 localStorage
    login(payload: { userId: number; accountName: string; isAdmin: boolean }) {
      this.userId = payload.userId
      this.accountName = payload.accountName
      this.isAdmin = payload.isAdmin
      this.loginTimestamp = Date.now()

      const nextState = JSON.stringify({
        userId: this.userId,
        accountName: this.accountName,
        isAdmin: this.isAdmin,
        loginTimestamp: this.loginTimestamp,
      })

      if (this.isAdmin) {
        sessionStorage.setItem(USER_STATE_KEY, nextState)
        localStorage.removeItem(USER_STATE_KEY)
        return
      }

      localStorage.setItem(USER_STATE_KEY, nextState)
      sessionStorage.removeItem(USER_STATE_KEY)
    },
    // 3. 登出时清除 state 与所有存储
    logout() {
      this.userId = null
      this.accountName = null
      this.isAdmin = false
      this.loginTimestamp = null
      localStorage.removeItem(USER_STATE_KEY)
      sessionStorage.removeItem(USER_STATE_KEY)
    },
  },
})
