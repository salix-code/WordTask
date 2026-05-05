<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/userStore'
import { login as apiLogin } from '@/api/account'

const router = useRouter()
const userStore = useUserStore()

const accountName = ref('')
const isLoading = ref(false)
const errorMsg = ref('')

async function onLogin() {
  if (!accountName.value) {
    errorMsg.value = '请输入账号'
    return
  }
  isLoading.value = true
  errorMsg.value = ''
  try {
    const res = await apiLogin(accountName.value)
    userStore.login(res)
    router.push(res.isAdmin ? '/admin' : '/student')
  } catch (err: any) {
    errorMsg.value = (err as Error).message || '网络错误，请稍后再试'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="flex-1 flex items-center justify-center p-6">
    <div class="card w-full max-w-sm bg-base-100 shadow-xl border border-base-300">
      <div class="card-body gap-4">
        <h2 class="card-title">登录</h2>
        <p class="text-sm text-base-content/60">请输入您的账号以继续。</p>

        <div class="form-control">
          <input
            v-model="accountName"
            type="text"
            class="input input-bordered"
            placeholder=""
            @keyup.enter="onLogin"
          />
        </div>

        <div v-if="errorMsg" class="text-error text-sm">{{ errorMsg }}</div>

        <button
          class="btn btn-primary min-h-touch"
          :class="{ 'btn-disabled': isLoading }"
          @click="onLogin"
        >
          <span v-if="isLoading" class="loading loading-spinner"></span>
          登录
        </button>
        
      </div>
    </div>
  </div>
</template>
