<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import Navbar from '@/components/common/Navbar.vue'
import WordCard from '@/components/review/WordCard.vue'
import ActionBars from '@/components/review/ActionBars.vue'
import SessionSummary from '@/components/review/SessionSummary.vue'
import { useReviewStore } from '@/store/reviewStore'
import { useSettingStore } from '@/store/settingStore'
import type { ReviewQuality } from '@/types/api'
import { submitReview, submitRevision } from '@/api/word'

const router = useRouter()
const review = useReviewStore()
const setting = useSettingStore()

onMounted(async () => {
  if (review.sessionQueue.length === 0) await review.initDaily()
})

const showSummary = computed(() => review.isSessionFinished)
const title = computed(() => (review.mode === 'revision' ? '复习中' : '背诵中'))
const isRevisionEmpty = computed(() => review.mode === 'revision' && review.dailyTotal === 0)

function handleRate(q: ReviewQuality) {
  review.submitCurrentWord(q)
  const wordId = review.history.at(-1)!.wordId
  if (review.mode === 'revision') {
    submitRevision(wordId, q, setting.wordbook).catch(console.warn)
    return
  }
  submitReview(wordId, q, setting.wordbook).catch(console.warn)
}

async function onContinue() {
  if (review.dailyRemaining === 0) return
  await review.completeSession()
}

function onGoHome() {
  router.push('/student')
}
</script>

<template>
  <Navbar :title="title" />
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

    <template v-else-if="isRevisionEmpty">
      <div class="flex-1 flex flex-col items-center justify-center gap-4 text-base-content/60">
        <div class="text-5xl">🧠</div>
        <div class="text-lg font-semibold">当前没有到期的复习单词</div>
        <button class="btn btn-outline" @click="onGoHome">返回首页</button>
      </div>
    </template>

    <template v-else-if="review.wordbookCompleted">
      <div class="flex-1 flex flex-col items-center justify-center gap-4 text-base-content/60">
        <div class="text-5xl">🎉</div>
        <div class="text-lg font-semibold">词库已全部学完！</div>
        <button class="btn btn-outline" @click="onGoHome">返回首页</button>
      </div>
    </template>

    <template v-else>
      <div class="flex-1 flex items-center justify-center text-base-content/60">
        加载中...
      </div>
    </template>
  </main>
</template>
