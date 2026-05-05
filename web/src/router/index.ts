import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/store/userStore'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'RootEntry',
    redirect: () => {
      const userStore = useUserStore()
      userStore.logout()
      return { name: 'Login' }
    },
  },
  {
    path: '/student',
    name: 'Student',
    component: () => import('@/views/Student.vue'),
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
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/Admin.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  {
    path: '/setup-cycle',
    name: 'SetupCycle',
    redirect: { name: 'Admin' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/student',
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
  const requiresAdmin = to.meta.requiresAdmin

  if (requiresAuth && !isLoggedIn) {
    // 需要认证但未登录，跳转到登录页
    next({ name: 'Login' })
  } else if (requiresAdmin && !userStore.isAdmin) {
    next({ name: 'Student' })
  } else if (to.name === 'Login' && isLoggedIn) {
    next({ name: userStore.isAdmin ? 'Admin' : 'Student' })
  } else {
    // 其他情况，正常放行
    next()
  }
})

export default router
