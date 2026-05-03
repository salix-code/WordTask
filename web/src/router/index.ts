import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/store/userStore'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/review',
    name: 'Review',
    component: () => import('@/views/Review.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
  },
  {
    path: '/setup-cycle',
    name: 'SetupCycle',
    component: () => import('@/views/SetupCycle.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, _from, next) => {
  // 在守卫内部获取 store 实例，确保 Pinia 已激活
  const userStore = useUserStore()

  const isLoggedIn = userStore.isLoggedIn
  const requiresAuth = to.meta.requiresAuth

  if (requiresAuth && !isLoggedIn) {
    // 需要认证但未登录，跳转到登录页
    next({ name: 'Login' })
  } else if (to.name === 'Login' && isLoggedIn) {
    // 已登录但要访问登录页，跳转到首页
    next({ name: 'Home' })
  } else {
    // 其他情况，正常放行
    next()
  }
})

export default router
