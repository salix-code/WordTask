<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useReviewStore } from '@/store/reviewStore'
import { useUserStore } from '@/store/userStore'
import { useSettingStore } from '@/store/settingStore'

const router = useRouter()
const review = useReviewStore()
const userStore = useUserStore()
const setting = useSettingStore()

onMounted(async () => {
  await setting.ensureWordbooksLoaded()
  if (review.dailyTotal === 0) await review.initDaily()
})

function startReview() {
  // 若上一轮已完结，重新开始
  if (review.isDailyFinished) review.initDaily()
  router.push('/review')
}

async function startRevision() {
  await review.initRevision()
  router.push('/review')
}

function logout() {
  userStore.logout()
  review.reset()
  router.push('/login')
}
</script>

<template>
  <div class="flex-1 bg-base-200/40 px-4 py-5 sm:px-6 sm:py-6">
    <div class="mx-auto w-full max-w-3xl">
      <div class="mb-4 flex justify-end">
        <button class="btn btn-ghost btn-sm min-h-touch" @click="logout">登出</button>
      </div>
      <div class="mx-auto w-full max-w-md card bg-base-100 shadow-2xl border border-base-300">
        <div class="card-body items-center text-center gap-4 sm:gap-5 py-7">
          <div class="text-4xl font-extrabold tracking-wide text-primary">{{ setting.wordbook }}</div>

          <!-- 没有进行中的周期 -->
          <template v-if="review.noCycle">
            <div class="alert alert-warning w-full shadow-sm">
              <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
              </svg>
              <span>暂无进行中的单词周期，请联系管理员录入。</span>
            </div>
            <button class="btn btn-outline btn-wide min-h-touch" @click="startRevision">
              复习
            </button>
          </template>

          <!-- 正常状态 -->
          <template v-else>
            <div class="stats stats-horizontal w-full bg-base-200/60 border border-base-300">
              <div class="stat place-items-center">
                <div class="stat-title text-base-content/60">今日总量</div>
                <div class="stat-value text-primary">{{ review.dailyTotal }}</div>
              </div>
              <div class="stat place-items-center">
                <div class="stat-title text-base-content/60">已完成</div>
                <div class="stat-value">{{ review.dailyCompleted }}</div>
              </div>
            </div>

            <button class="btn btn-primary btn-wide min-h-touch" @click="startReview">
              开始背诵
            </button>
            <button class="btn btn-secondary btn-wide min-h-touch" @click="startRevision">
              复习
            </button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
