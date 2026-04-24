<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import Navbar from '@/components/common/Navbar.vue'
import WordCard from '@/components/review/WordCard.vue'
import ActionBars from '@/components/review/ActionBars.vue'
import SessionSummary from '@/components/review/SessionSummary.vue'
import { useReviewStore } from '@/store/reviewStore'
import type { ReviewQuality } from '@/types/api'
// M2 接入：import { submitReview } from '@/api/word'

const router = useRouter()
const review = useReviewStore()

onMounted(() => {
  if (review.sessionQueue.length === 0) review.initDaily()
})

const showSummary = computed(() => review.isSessionFinished)

function handleRate(q: ReviewQuality) {
  review.submitCurrentWord(q)
  // M2：异步上报，不阻塞 UI
  // submitReview(review.history.at(-1)!.wordId, q).catch(console.warn)
}

function onContinue() {
  if (review.dailyQueue.length === 0) return
  review.startNextSession()
}

function onGoHome() {
  router.push('/')
}
</script>

<template>
  <Navbar title="背诵中" />
  <main class="flex-1 max-w-xl w-full mx-auto px-4 py-4 flex flex-col gap-4">
    <template v-if="!showSummary && review.currentWord">
      <WordCard
        :word="review.currentWord"
        :flipped="review.flipped"
        @flip="review.flipCard"
      />
      <ActionBars @rate="handleRate" />
    </template>

    <template v-else-if="showSummary">
      <SessionSummary @continue="onContinue" @home="onGoHome" />
    </template>

    <template v-else>
      <div class="flex-1 flex items-center justify-center text-base-content/60">
        加载中...
      </div>
    </template>
  </main>
</template>
