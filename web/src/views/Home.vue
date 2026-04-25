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
</script>

<template>
  <div class="flex-1 flex flex-col items-center justify-center p-6">
    <div class="max-w-md w-full card bg-base-100 shadow-xl border border-base-300">
      <div class="card-body items-center text-center gap-4">
        <div class="text-3xl font-bold">WordTask</div>
        <p class="text-base-content/70">科学记忆法 · 每组 {{ setting.batchSize }} 个单词</p>

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
      </div>
    </div>
  </div>
</template>
