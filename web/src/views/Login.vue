<script setup lang="ts">
// M3 阶段实现 - 预留骨架
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const account = ref('')
const password = ref('')

function onLogin() {
  // TODO(M3): 调用 /api/login
  router.push('/')
}
</script>

<template>
  <div class="flex-1 flex items-center justify-center p-6">
    <div class="card w-full max-w-sm bg-base-100 shadow-xl border border-base-300">
      <div class="card-body gap-3">
        <h2 class="card-title">登录</h2>
        <input v-model="account" class="input input-bordered" placeholder="账号" />
        <input v-model="password" type="password" class="input input-bordered" placeholder="密码" />
        <button class="btn btn-primary min-h-touch" @click="onLogin">登录</button>
        <p class="text-xs text-base-content/50">M3 里程碑完善</p>
      </div>
    </div>
  </div>
</template>
