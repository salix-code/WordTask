<script setup lang="ts">
import type { Word } from '@/types/word'
import { speak } from '@/utils/speech'

interface Props {
  word: Word
  flipped: boolean
}
const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'flip'): void
}>()

function onSpeak(e: MouseEvent) {
  e.stopPropagation()
  speak(props.word.text)
}

function onClick() {
  emit('flip')
}
</script>

<template>
  <div class="flip-card w-full" @click="onClick">
    <div
      class="flip-card-inner relative w-full aspect-[3/4] sm:aspect-[4/5] cursor-pointer select-none"
      :class="{ 'is-flipped': flipped }"
    >
      <!-- 正面：单词 + 音标 -->
      <div
        class="flip-card-face absolute inset-0 rounded-3xl bg-base-100 shadow-xl border border-base-300 flex flex-col items-center justify-center p-6"
      >
        <button
          class="btn btn-circle btn-ghost absolute top-4 right-4 min-h-touch min-w-touch"
          aria-label="朗读单词"
          @click="onSpeak"
        >
          <span class="text-xl">🔊</span>
        </button>
        <div class="text-4xl sm:text-5xl font-bold tracking-tight">{{ word.text }}</div>
        <div v-if="word.phonetic" class="mt-3 text-base text-base-content/60">
          {{ word.phonetic }}
        </div>
        <div class="mt-auto text-xs text-base-content/40">点击卡片查看释义</div>
      </div>

      <!-- 反面：释义 + 例句 -->
      <div
        class="flip-card-face flip-card-back absolute inset-0 rounded-3xl bg-base-100 shadow-xl border border-base-300 p-6 overflow-auto"
      >
        <div class="text-2xl font-semibold mb-3">{{ word.text }}</div>
        <ul class="space-y-1 mb-4">
          <li
            v-for="(def, i) in word.definitions"
            :key="i"
            style="font-size: 1.25rem; line-height: 1.75rem;"
          >
            <span v-if="def.pos" class="badge badge-outline badge-sm mr-2">{{ def.pos }}</span>
            {{ def.meaning }}
          </li>
        </ul>
        <div v-if="word.examples && word.examples.length" class="divider text-xs">例句</div>
        <ul class="space-y-2">
          <li v-for="(ex, i) in word.examples" :key="i" class="text-sm">
            <div class="text-base-content/90">{{ ex.en }}</div>
            <div v-if="ex.zh" class="text-base-content/60">{{ ex.zh }}</div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
