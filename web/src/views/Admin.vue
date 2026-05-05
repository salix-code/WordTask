<script setup lang="ts">
import { ref } from 'vue'
import SetupCycle from '@/views/SetupCycle.vue'
import { useUserStore } from '@/store/userStore'
import { createAccount } from '@/api/account'

type AdminTab = 'cycle' | 'user'
type UserActionTab = 'add' | 'search'

const activeTab = ref<AdminTab>('cycle')
const activeUserAction = ref<UserActionTab>('add')
const userStore = useUserStore()

const newAccountName = ref('')
const adding = ref(false)
const successMsg = ref('')
const errorMsg = ref('')

async function onAddUser() {
  const accountName = newAccountName.value.trim()
  if (!accountName) {
    errorMsg.value = '请输入账号名'
    successMsg.value = ''
    return
  }
  if (!userStore.userId) {
    errorMsg.value = '当前管理员信息无效，请重新登录'
    successMsg.value = ''
    return
  }

  adding.value = true
  successMsg.value = ''
  errorMsg.value = ''
  try {
    const res = await createAccount(userStore.userId, accountName)
    successMsg.value = `添加成功：${res.accountName}（ID: ${res.userId}）`
    newAccountName.value = ''
  } catch (err: unknown) {
    errorMsg.value = err instanceof Error ? err.message : '添加失败，请稍后重试'
  } finally {
    adding.value = false
  }
}
</script>

<template>
  <div class="flex-1 flex flex-col">
    <div class="max-w-6xl w-full mx-auto px-4 pt-4">
      <h1 class="text-xl font-bold mb-3">后台管理</h1>
      <div class="tabs tabs-boxed w-fit bg-base-200 border border-base-300 rounded-2xl p-1 shadow-sm">
        <button
          type="button"
          class="tab min-h-touch px-6 font-medium transition-colors duration-150"
          :class="activeTab === 'cycle'
            ? 'tab-active !bg-primary !text-primary-content shadow-sm'
            : 'text-base-content/70 hover:text-base-content'"
          @click="activeTab = 'cycle'"
        >
          周期管理
        </button>
        <button
          type="button"
          class="tab min-h-touch px-6 font-medium transition-colors duration-150"
          :class="activeTab === 'user'
            ? 'tab-active !bg-primary !text-primary-content shadow-sm'
            : 'text-base-content/70 hover:text-base-content'"
          @click="activeTab = 'user'"
        >
          用户管理
        </button>
      </div>
    </div>

    <SetupCycle v-if="activeTab === 'cycle'" />

    <div v-else class="flex-1 p-6">
      <div class="w-full max-w-6xl mx-auto">
        <div class="card bg-base-100 border border-base-300 shadow-sm">
          <div class="card-body gap-4 flex-row items-start">
            <div class="w-36 sm:w-44 shrink-0">
              <div class="bg-base-200 border border-base-300 rounded-2xl p-2 flex flex-col gap-2">
                <button
                  type="button"
                  class="btn btn-ghost min-h-touch justify-start px-4 normal-case rounded-xl transition-all duration-150"
                  :class="activeUserAction === 'add'
                    ? '!bg-primary !text-primary-content shadow-sm translate-x-2'
                    : 'text-base-content/70 hover:text-base-content hover:translate-x-1'"
                  @click="activeUserAction = 'add'"
                >
                  添加用户
                </button>
                <button
                  type="button"
                  class="btn btn-ghost min-h-touch justify-start px-4 normal-case rounded-xl transition-all duration-150"
                  :class="activeUserAction === 'search'
                    ? '!bg-primary !text-primary-content shadow-sm translate-x-2'
                    : 'text-base-content/70 hover:text-base-content hover:translate-x-1'"
                  @click="activeUserAction = 'search'"
                >
                  搜索用户
                </button>
              </div>
            </div>

            <div class="flex-1 min-h-[220px] border border-base-300 rounded-2xl p-4 bg-base-200 ml-1">
              <template v-if="activeUserAction === 'add'">
                <h3 class="text-lg font-semibold mb-2">添加用户</h3>
                <p class="text-sm text-base-content/70 mb-4">创建新的账号用于登录系统。</p>
                <div class="join join-vertical sm:join-horizontal w-full">
                  <input
                    v-model="newAccountName"
                    type="text"
                    class="input input-bordered join-item flex-1"
                    placeholder="输入新账号名"
                    @keyup.enter="onAddUser"
                  />
                  <button
                    class="btn btn-primary join-item min-h-touch"
                    :class="{ 'btn-disabled': adding }"
                    @click="onAddUser"
                  >
                    <span v-if="adding" class="loading loading-spinner loading-sm" />
                    添加用户
                  </button>
                </div>
                <div v-if="successMsg" class="alert alert-success py-2 mt-4">
                  <span class="text-sm">{{ successMsg }}</span>
                </div>
                <div v-if="errorMsg" class="alert alert-error py-2 mt-4">
                  <span class="text-sm">{{ errorMsg }}</span>
                </div>
              </template>

              <template v-else>
                <h3 class="text-lg font-semibold mb-2">搜索用户</h3>
                <p class="text-sm text-base-content/70 mb-4">按账号名查找用户（功能开发中）。</p>
                <div class="join join-vertical sm:join-horizontal w-full">
                  <input
                    type="text"
                    class="input input-bordered join-item flex-1"
                    placeholder="输入账号名进行搜索"
                    disabled
                  />
                  <button class="btn btn-outline join-item min-h-touch" disabled>搜索</button>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
