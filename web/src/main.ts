import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/styles/main.css'
import { useUserStore } from './store/userStore'

const app = createApp(App)

app.use(createPinia())

// 在挂载路由前检查登录状态是否超时
const userStore = useUserStore()
const TIMEOUT = 24 * 60 * 60 * 1000 // 24小时
if (userStore.loginTimestamp && Date.now() - userStore.loginTimestamp > TIMEOUT) {
  userStore.logout()
}

app.use(router)

app.mount('#app')
