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

const VOWELS = new Set('aeiouAEIOU')
function charClass(ch: string) {
  return VOWELS.has(ch) ? 'text-primary' : ''
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
        <div class="flex flex-col items-center justify-center gap-4">
          <div class="text-6xl sm:text-8xl font-bold tracking-tight text-center">
            <span v-for="(ch, i) in word.text" :key="i" :class="charClass(ch)">{{ ch }}</span>
          </div>
          <div v-if="word.phonetic" class="text-2xl sm:text-3xl text-base-content/60 text-center">
            /{{ word.phonetic }}/
          </div>
        </div>
        <div class="absolute bottom-5 text-xs text-base-content/40">点击卡片查看释义</div>
      </div>

      <!-- 反面：释义 + 例句 -->
      <div
        class="flip-card-face flip-card-back absolute inset-0 rounded-3xl bg-base-100 shadow-xl border border-base-300 pt-12 px-6 pb-6 overflow-auto"
      >
        <div class="text-3xl font-semibold mb-4">
          <span v-for="(ch, i) in word.text" :key="i" :class="charClass(ch)">{{ ch }}</span>
        </div>
        <ul class="space-y-2 mb-5">
          <li
            v-for="(def, i) in word.definitions"
            :key="i"
            style="font-size: 1.5rem; line-height: 2rem;"
          >
            <span v-if="def.pos" class="badge badge-outline badge-lg mr-2 text-base">{{ def.pos }}</span>
            {{ def.meaning }}
          </li>
        </ul>
        <div v-if="word.examples && word.examples.length" class="divider text-sm">例句</div>
        <ul class="space-y-3">
          <li v-for="(ex, i) in word.examples" :key="i" class="text-lg">
            <div class="text-base-content/90">{{ ex.en }}</div>
            <div v-if="ex.zh" class="text-base-content/60 text-lg">{{ ex.zh }}</div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
