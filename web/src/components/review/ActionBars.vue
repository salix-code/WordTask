<script setup lang="ts">
import type { ReviewQuality } from '@/types/api'

const emit = defineEmits<{
  (e: 'rate', quality: ReviewQuality): void
}>()

function rate(q: ReviewQuality) {
  emit('rate', q)
}
</script>

<template>
  <div class="grid grid-cols-3 gap-3 w-full">
    <button
      class="btn btn-error text-white min-h-touch text-base font-semibold shadow-md"
      aria-label="忘记"
      @click="rate('forgot')"
    >
      忘记
    </button>
    <button
      class="btn btn-warning text-white min-h-touch text-base font-semibold shadow-md"
      aria-label="模糊"
      @click="rate('vague')"
    >
      模糊
    </button>
    <button
      class="btn btn-success text-white min-h-touch text-base font-semibold shadow-md"
      aria-label="认识"
      @click="rate('known')"
    >
      认识
    </button>
  </div>
</template>
