<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { setupCycle } from '@/api/cycle'
import { useReviewStore } from '@/store/reviewStore'
import type { WordStatus } from '@/types/cycle'

const router = useRouter()
const review = useReviewStore()

const tags = ref<string[]>([])
const inputVal = ref('')
const results = ref<WordStatus[]>([])
const submitting = ref(false)
const successMsg = ref('')
const errorMsg = ref('')

// 当前 wordbook 固定为 KET（与现有 word.ts 保持一致，后续可扩展为 props/store）
const wordbook = 'KET'

const canSubmit = computed(() => tags.value.length >= 1 && !submitting.value)

// 根据已有 results 获取某个 tag 的状态
function tagStatus(term: string): WordStatus['status'] | null {
  if (results.value.length === 0) return null
  const r = results.value.find((x) => x.term.toLowerCase() === term.toLowerCase())
  return r?.status ?? null
}

function tagClass(term: string): string {
  const s = tagStatus(term)
  if (s === 'not_in_dict') return 'badge badge-error'
  if (s === 'already_used') return 'badge badge-warning'
  if (s === 'valid') return 'badge badge-success'
  return 'badge badge-neutral'
}

function addTag() {
  const word = inputVal.value.trim()
  if (!word) return
  // 避免重复
  if (tags.value.some((t) => t.toLowerCase() === word.toLowerCase())) {
    inputVal.value = ''
    return
  }
  tags.value.push(word)
  inputVal.value = ''
  // 若有结果，清除（因为列表已变动）
  if (results.value.length > 0) {
    results.value = []
    successMsg.value = ''
    errorMsg.value = ''
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault()
    addTag()
  } else if (e.key === 'Backspace' && inputVal.value === '' && tags.value.length > 0) {
    tags.value.pop()
  }
}

function removeTag(index: number) {
  tags.value.splice(index, 1)
  if (results.value.length > 0) {
    results.value = []
    successMsg.value = ''
    errorMsg.value = ''
  }
}

async function submit() {
  if (!canSubmit.value) return
  submitting.value = true
  successMsg.value = ''
  errorMsg.value = ''
  results.value = []
  try {
    const res = await setupCycle({ wordbook, words: tags.value })
    results.value = res.results
    if (res.ok) {
      successMsg.value = '周期创建成功！'
      // 让下次进入首页 / 复习页时重新拉取今日单词
      review.reset()
    } else {
      const badCount = res.results.filter((r) => r.status !== 'valid').length
      errorMsg.value = `有 ${badCount} 个单词存在问题，请修改后重新提交。`
    }
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '提交失败，请稍后重试。'
  } finally {
    submitting.value = false
  }
}

function goHome() {
  router.push('/')
}
</script>

<template>
  <div class="flex-1 flex flex-col items-center p-6">
    <div class="max-w-lg w-full flex flex-col gap-6">

      <!-- 标题 -->
      <div class="flex items-center gap-3">
        <button class="btn btn-ghost btn-sm" @click="goHome">← 返回</button>
        <h1 class="text-xl font-bold">录入单词表</h1>
      </div>

      <!-- 说明 -->
      <p class="text-base-content/60 text-sm">
        输入单词后按 <kbd class="kbd kbd-sm">Enter</kbd> 或 <kbd class="kbd kbd-sm">Space</kbd> 生成标签，至少输入 1 个即可提交。
      </p>

      <!-- 标签区 + 输入框 -->
      <div
        class="border border-base-300 rounded-xl p-3 min-h-32 flex flex-wrap gap-2 cursor-text focus-within:border-primary transition-colors"
        @click="($refs.inputRef as HTMLInputElement)?.focus()"
      >
        <span
          v-for="(tag, i) in tags"
          :key="i"
          :class="[tagClass(tag), 'gap-1 select-none']"
        >
          {{ tag }}
          <button
            class="ml-1 opacity-60 hover:opacity-100"
            @click.stop="removeTag(i)"
          >×</button>
        </span>

        <input
          ref="inputRef"
          v-model="inputVal"
          type="text"
          placeholder="输入单词..."
          class="outline-none bg-transparent min-w-24 flex-1 text-sm"
          @keydown="onKeydown"
        />
      </div>

      <!-- 进度 -->
      <div class="flex justify-between text-sm text-base-content/50">
        <span>已录入 {{ tags.length }} 个</span>
        <span v-if="tags.length >= 1" class="text-success font-medium">可提交</span>
      </div>

      <!-- 错误/成功提示 -->
      <div v-if="errorMsg" class="alert alert-error text-sm">{{ errorMsg }}</div>
      <div v-if="successMsg" class="alert alert-success text-sm">
        {{ successMsg }}
        <button class="btn btn-sm btn-ghost ml-auto" @click="goHome">返回主页</button>
      </div>

      <!-- 图例（仅校验后显示） -->
      <div v-if="results.length > 0 && !successMsg" class="flex gap-3 text-xs text-base-content/60">
        <span><span class="badge badge-error badge-sm mr-1" />不在词典中</span>
        <span><span class="badge badge-warning badge-sm mr-1" />本周期已用过</span>
        <span><span class="badge badge-success badge-sm mr-1" />有效</span>
      </div>

      <!-- 提交按钮 -->
      <button
        class="btn btn-primary btn-wide self-center"
        :disabled="!canSubmit"
        @click="submit"
      >
        <span v-if="submitting" class="loading loading-spinner loading-sm" />
        {{ submitting ? '校验中...' : '提交' }}
      </button>

    </div>
  </div>
</template>
