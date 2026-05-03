<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { setupCycle, fetchCurrentCycle, updateCycle, clearAllCycles } from '@/api/cycle'
import { useReviewStore } from '@/store/reviewStore'
import type { WordStatus } from '@/types/cycle'

const router = useRouter()
const review = useReviewStore()

// 每行一个词，空格保留在词内
const textareaVal = ref('')
const results = ref<WordStatus[]>([])
const submitting = ref(false)
const successMsg = ref('')
const errorMsg = ref('')
// 是否处于编辑已有周期的模式
const isEditMode = ref(false)
// 清空确认弹窗
const showClearConfirm = ref(false)
const clearing = ref(false)
// 已有周期中标记为 known 的词（保留展示，防止用户误删）
const knownTerms = ref<Set<string>>(new Set())
const loading = ref(true)

const wordbook = 'KET'

// 从 textarea 解析出去重后的词列表（每行一词，忽略空行）
const tags = computed<string[]>(() => {
  const seen = new Set<string>()
  return textareaVal.value
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => {
      if (!line) return false
      const key = line.toLowerCase()
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
})

const canSubmit = computed(() => tags.value.length >= 1 && !submitting.value)

// 输入变动时清除旧校验结果
watch(textareaVal, () => {
  if (results.value.length > 0) {
    results.value = []
    successMsg.value = ''
    errorMsg.value = ''
  }
})

// 挂载时检查是否已有进行中的周期
onMounted(async () => {
  try {
    const res = await fetchCurrentCycle(wordbook)
    if (res.hasCycle && res.words.length > 0) {
      isEditMode.value = true
      // 预填所有词（包括 known 的）
      textareaVal.value = res.words.map((w) => w.term).join('\n')
      // 记录 known 词，供提示展示
      knownTerms.value = new Set(
        res.words.filter((w) => w.status === 'known').map((w) => w.term.toLowerCase()),
      )
    }
  } catch {
    // 拉取失败时静默降级为新建模式
  } finally {
    loading.value = false
  }
})

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

async function submit() {
  if (!canSubmit.value) return
  submitting.value = true
  successMsg.value = ''
  errorMsg.value = ''
  results.value = []
  try {
    if (isEditMode.value) {
      const res = await updateCycle({ wordbook, words: tags.value })
      results.value = res.results
      if (res.ok) {
        successMsg.value = '周期已更新！'
        review.reset()
        // 更新 knownTerms
        knownTerms.value = new Set(
          tags.value
            .filter((t) => knownTerms.value.has(t.toLowerCase()))
            .map((t) => t.toLowerCase()),
        )
      } else {
        const badCount = res.results.filter((r) => r.status !== 'valid').length
        errorMsg.value = `有 ${badCount} 个单词存在问题，请修改后重新提交。`
      }
    } else {
      const res = await setupCycle({ wordbook, words: tags.value })
      results.value = res.results
      if (res.ok) {
        successMsg.value = '周期创建成功！'
        review.reset()
        isEditMode.value = true
      } else {
        const badCount = res.results.filter((r) => r.status !== 'valid').length
        errorMsg.value = `有 ${badCount} 个单词存在问题，请修改后重新提交。`
      }
    }
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '提交失败，请稍后重试。'
  } finally {
    submitting.value = false
  }
}

async function clearCycles() {
  clearing.value = true
  try {
    await clearAllCycles(wordbook)
    // Reset all local state back to fresh-entry mode
    textareaVal.value = ''
    results.value = []
    successMsg.value = ''
    errorMsg.value = ''
    isEditMode.value = false
    knownTerms.value = new Set()
    review.reset()
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '清空失败，请稍后重试。'
  } finally {
    clearing.value = false
    showClearConfirm.value = false
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
        <h1 class="text-xl font-bold">{{ isEditMode ? '编辑当前周期' : '录入单词表' }}</h1>
        <button
          v-if="isEditMode"
          class="btn btn-error btn-sm btn-outline ml-auto"
          @click="showClearConfirm = true"
        >清空周期</button>
      </div>

      <!-- 清空确认弹窗 -->
      <div v-if="showClearConfirm" class="modal modal-open">
        <div class="modal-box">
          <h3 class="font-bold text-lg">确认清空所有周期？</h3>
          <p class="py-4 text-base-content/70">
            此操作将删除所有历史周期及进度，无法撤销。清空后可重新录入单词表。
          </p>
          <div class="modal-action">
            <button class="btn btn-ghost" :disabled="clearing" @click="showClearConfirm = false">取消</button>
            <button class="btn btn-error" :disabled="clearing" @click="clearCycles">
              <span v-if="clearing" class="loading loading-spinner loading-sm" />
              {{ clearing ? '清空中...' : '确认清空' }}
            </button>
          </div>
        </div>
        <div class="modal-backdrop" @click="showClearConfirm = false" />
      </div>

      <!-- 加载中 -->
      <div v-if="loading" class="flex justify-center py-8">
        <span class="loading loading-spinner loading-md" />
      </div>

      <template v-else>
        <!-- 说明 -->
        <p class="text-base-content/60 text-sm">
          每行一个单词，按 <kbd class="kbd kbd-sm">Enter</kbd> 换行；单词内部空格会保留（如 <em>go home</em>）。
          <template v-if="isEditMode">
            直接修改后提交即可同步到当前周期。
            <span v-if="knownTerms.size > 0" class="text-warning">⚠ 已掌握的词删除后进度不会回退。</span>
          </template>
          <template v-else>至少输入 1 个即可提交。</template>
        </p>

        <!-- 已掌握词提示 -->
        <div v-if="isEditMode && knownTerms.size > 0" class="flex flex-wrap gap-1 text-xs">
          <span class="text-base-content/50">已掌握：</span>
          <span
            v-for="t in [...knownTerms]"
            :key="t"
            class="badge badge-success badge-sm"
          >{{ t }}</span>
        </div>

        <!-- 文本输入区 -->
        <textarea
          v-model="textareaVal"
          placeholder="输入单词，每行一个…"
          rows="8"
          class="textarea textarea-bordered w-full text-sm font-mono leading-relaxed focus:textarea-primary resize-y"
        />

        <!-- 进度 + 标签预览 -->
        <div class="flex flex-col gap-2">
          <div class="flex justify-between text-sm text-base-content/50">
            <span>已录入 {{ tags.length }} 个</span>
            <span v-if="tags.length >= 1" class="text-success font-medium">可提交</span>
          </div>
          <!-- 校验结果标签（提交后显示） -->
          <div v-if="results.length > 0" class="flex flex-wrap gap-2">
            <span
              v-for="tag in tags"
              :key="tag"
              :class="[tagClass(tag), 'select-none']"
            >{{ tag }}</span>
          </div>
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
          <span><span class="badge badge-warning badge-sm mr-1" />其他周期已用过</span>
          <span><span class="badge badge-success badge-sm mr-1" />有效</span>
        </div>

        <!-- 提交按钮 -->
        <button
          class="btn btn-primary btn-wide self-center"
          :disabled="!canSubmit"
          @click="submit"
        >
          <span v-if="submitting" class="loading loading-spinner loading-sm" />
          {{ submitting ? (isEditMode ? '更新中...' : '校验中...') : (isEditMode ? '更新周期' : '提交') }}
        </button>
      </template>

    </div>
  </div>
</template>
