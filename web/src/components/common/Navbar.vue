<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useReviewStore } from '@/store/reviewStore'

interface Props {
  title?: string
}
const props = withDefaults(defineProps<Props>(), { title: 'WordTask' })

const router = useRouter()
const review = useReviewStore()

const dailyPercent = computed(() => {
  if (!review.dailyTotal) return 0
  return Math.min(100, Math.round((review.dailyCompleted / review.dailyTotal) * 100))
})

function goHome() {
  router.push('/')
}
</script>

<template>
  <header class="sticky top-0 z-20 bg-base-100/90 backdrop-blur border-b border-base-300">
    <div class="max-w-xl mx-auto flex items-center gap-3 px-4 h-14">
      <button
        class="btn btn-ghost btn-sm min-h-touch min-w-touch"
        aria-label="返回首页"
        @click="goHome"
      >
        <span class="text-xl">←</span>
      </button>
      <h1 class="font-semibold text-base flex-1 truncate">{{ props.title }}</h1>
      <div class="text-xs text-base-content/60 tabular-nums">
        <span class="mr-2">批次 {{ review.sessionProgressText }}</span>
        <span>今日 {{ review.dailyCompleted }}/{{ review.dailyTotal }}</span>
      </div>
    </div>
    <div class="h-1 w-full bg-base-300">
      <div
        class="h-1 bg-primary transition-[width] duration-300"
        :style="{ width: dailyPercent + '%' }"
      />
    </div>
  </header>
</template>
