<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useReviewStore } from '@/store/reviewStore'
import { useSettingStore } from '@/store/settingStore'

const router = useRouter()
const review = useReviewStore()
const setting = useSettingStore()

onMounted(async () => {
  if (review.dailyTotal === 0) await review.initDaily()
})

function startReview() {
  // 若上一轮已完结，重新开始
  if (review.isDailyFinished) review.initDaily()
  router.push('/review')
}

function goSetupCycle() {
  router.push('/setup-cycle')
}
</script>

<template>
  <div class="flex-1 flex flex-col items-center justify-center p-6">
    <div class="max-w-md w-full card bg-base-100 shadow-xl border border-base-300">
      <div class="card-body items-center text-center gap-4">
        <div class="text-3xl font-bold">WordTask</div>
        <p class="text-base-content/70">科学记忆法 · 每组 {{ setting.batchSize }} 个单词</p>

        <!-- 没有进行中的周期 -->
        <template v-if="review.noCycle">
          <div class="alert alert-warning w-full">
            <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
            </svg>
            <span>暂无进行中的单词周期，请先录入单词表。</span>
          </div>
          <button class="btn btn-primary btn-wide min-h-touch" @click="goSetupCycle">
            录入单词表
          </button>
        </template>

        <!-- 正常状态 -->
        <template v-else>
          <div class="stats stats-horizontal shadow w-full">
            <div class="stat place-items-center">
              <div class="stat-title">今日总量</div>
              <div class="stat-value text-primary">{{ review.dailyTotal }}</div>
            </div>
            <div class="stat place-items-center">
              <div class="stat-title">已完成</div>
              <div class="stat-value">{{ review.dailyCompleted }}</div>
            </div>
          </div>

          <button class="btn btn-primary btn-wide min-h-touch" @click="startReview">
            开始背诵
          </button>

          <button class="btn btn-outline btn-wide btn-sm" @click="goSetupCycle">
            录入单词表
          </button>
        </template>
      </div>
    </div>
  </div>
</template>
