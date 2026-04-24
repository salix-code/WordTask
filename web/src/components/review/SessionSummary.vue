<script setup lang="ts">
import { computed } from 'vue'
import { useReviewStore } from '@/store/reviewStore'

const review = useReviewStore()

const emit = defineEmits<{
  (e: 'continue'): void
  (e: 'home'): void
}>()

const stats = computed(() => {
  const total = review.sessionResults.length
  const known = review.sessionResults.filter((r) => r.quality === 'known').length
  const vague = review.sessionResults.filter((r) => r.quality === 'vague').length
  const forgot = review.sessionResults.filter((r) => r.quality === 'forgot').length
  return { total, known, vague, forgot }
})

const hasMore = computed(() => review.dailyQueue.length > 0)
</script>

<template>
  <div class="w-full flex flex-col items-center gap-5">
    <div class="text-2xl font-semibold">这一组完成啦 🎉</div>

    <div class="stats stats-horizontal shadow w-full">
      <div class="stat place-items-center">
        <div class="stat-title text-xs">认识</div>
        <div class="stat-value text-success text-2xl">{{ stats.known }}</div>
      </div>
      <div class="stat place-items-center">
        <div class="stat-title text-xs">模糊</div>
        <div class="stat-value text-warning text-2xl">{{ stats.vague }}</div>
      </div>
      <div class="stat place-items-center">
        <div class="stat-title text-xs">忘记</div>
        <div class="stat-value text-error text-2xl">{{ stats.forgot }}</div>
      </div>
    </div>

    <div class="text-sm text-base-content/60">
      今日已完成 {{ review.dailyCompleted }} / {{ review.dailyTotal }}
    </div>

    <div class="grid grid-cols-2 gap-3 w-full">
      <button
        class="btn btn-outline min-h-touch"
        @click="emit('home')"
      >
        休息一下
      </button>
      <button
        class="btn btn-primary min-h-touch"
        :disabled="!hasMore"
        @click="emit('continue')"
      >
        {{ hasMore ? '再背一组' : '今日已完成' }}
      </button>
    </div>
  </div>
</template>
