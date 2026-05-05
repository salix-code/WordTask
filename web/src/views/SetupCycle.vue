<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  setupCycle,
  fetchCycles,
  fetchCycleDetail,
  updateCycle,
  clearAllCycles,
} from '@/api/cycle'
import { useReviewStore } from '@/store/reviewStore'
import { useSettingStore } from '@/store/settingStore'
import type { WordStatus, CycleSummary } from '@/types/cycle'

const review = useReviewStore()
const setting = useSettingStore()
const selectedWordbook = computed({
  get: () => setting.wordbook,
  set: (value: string) => setting.setWordbook(value),
})
const wordbookOptions = computed(() => setting.wordbookOptions)

const cycles = ref<CycleSummary[]>([])
const selectedCycleId = ref<number | null>(null)
const isCreateMode = ref(false)

const textareaVal = ref('')
const results = ref<WordStatus[]>([])
const submitting = ref(false)
const successMsg = ref('')
const errorMsg = ref('')
const knownTerms = ref<Set<string>>(new Set())

const loadingList = ref(true)
const loadingDetail = ref(false)
const showClearConfirm = ref(false)
const clearing = ref(false)

const selectedCycle = computed(() => cycles.value.find((c) => c.cycleId === selectedCycleId.value) ?? null)
const isReadonlyCycle = computed(() => !isCreateMode.value && selectedCycle.value?.status !== 'ongoing')

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

const canSubmit = computed(() => Boolean(selectedWordbook.value) && tags.value.length >= 1 && !submitting.value && !isReadonlyCycle.value)

const modeTitle = computed(() => {
  if (isCreateMode.value) return '新增周期'
  if (!selectedCycle.value) return '请选择周期'
  if (selectedCycle.value.status === 'ongoing') return '编辑当前周期'
  if (selectedCycle.value.status === 'queued') return '查看后续周期'
  return '查看历史周期'
})

const submitLabel = computed(() => {
  if (submitting.value) return isCreateMode.value ? '创建中...' : '更新中...'
  return isCreateMode.value ? '创建周期' : '更新当前周期'
})

watch(textareaVal, () => {
  if (results.value.length > 0) {
    results.value = []
    successMsg.value = ''
    errorMsg.value = ''
  }
})

onMounted(async () => {
  await setting.ensureWordbooksLoaded()
  await loadCycles()
})

watch(
  () => selectedWordbook.value,
  async (next, prev) => {
    if (!next || next === prev) return
    await loadCycles()
  },
)

function resetFeedback() {
  results.value = []
  successMsg.value = ''
  errorMsg.value = ''
}

function syncKnownTerms(words: Array<{ term: string; status: string }>) {
  knownTerms.value = new Set(
    words
      .filter((w) => w.status === 'known')
      .map((w) => w.term.toLowerCase()),
  )
}

async function loadCycles(preferCycleId?: number) {
  loadingList.value = true
  resetFeedback()
  try {
    if (!selectedWordbook.value) {
      cycles.value = []
      startCreateMode()
      return
    }
    const res = await fetchCycles(selectedWordbook.value)
    cycles.value = res.cycles
    if (cycles.value.length === 0) {
      startCreateMode()
      return
    }

    const preferred = preferCycleId
      ? cycles.value.find((c) => c.cycleId === preferCycleId)
      : null
    const ongoing = cycles.value.find((c) => c.status === 'ongoing')
    const target = preferred ?? ongoing ?? cycles.value[0]
    await loadCycleDetail(target.cycleId)
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '加载周期列表失败，请稍后重试。'
  } finally {
    loadingList.value = false
  }
}

async function loadCycleDetail(cycleId: number) {
  loadingDetail.value = true
  resetFeedback()
  try {
    const detail = await fetchCycleDetail(cycleId, selectedWordbook.value)
    selectedCycleId.value = detail.cycleId
    isCreateMode.value = false
    textareaVal.value = detail.words.map((w) => w.term).join('\n')
    syncKnownTerms(detail.words)
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '加载周期详情失败，请稍后重试。'
  } finally {
    loadingDetail.value = false
  }
}

function startCreateMode() {
  isCreateMode.value = true
  selectedCycleId.value = null
  textareaVal.value = ''
  knownTerms.value = new Set()
  resetFeedback()
}

async function submit() {
  if (!canSubmit.value) return
  submitting.value = true
  resetFeedback()
  try {
    if (isCreateMode.value) {
      const res = await setupCycle({ wordbook: selectedWordbook.value, words: tags.value })
      results.value = res.results
      if (res.ok) {
        successMsg.value = '周期创建成功！'
        review.reset()
        await loadCycles()
      } else {
        const badCount = res.results.filter((r) => r.status !== 'valid').length
        errorMsg.value = `有 ${badCount} 个单词存在问题，请修改后重新提交。`
      }
      return
    }

    const res = await updateCycle({ wordbook: selectedWordbook.value, words: tags.value })
    results.value = res.results
    if (res.ok) {
      successMsg.value = '当前周期已更新！'
      review.reset()
      if (selectedCycleId.value !== null) {
        await loadCycleDetail(selectedCycleId.value)
      }
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

async function clearCycles() {
  clearing.value = true
  try {
    await clearAllCycles(selectedWordbook.value)
    review.reset()
    await loadCycles()
  } catch (e: unknown) {
    errorMsg.value = e instanceof Error ? e.message : '清空失败，请稍后重试。'
  } finally {
    clearing.value = false
    showClearConfirm.value = false
  }
}

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

function formatCreatedAt(timeText: string): string {
  const date = new Date(timeText)
  if (Number.isNaN(date.getTime())) return timeText
  return date.toLocaleString()
}

</script>

<template>
  <div class="flex-1 flex flex-col items-center p-6">
    <div class="max-w-5xl w-full flex flex-col gap-5">
      <div class="flex items-center gap-3">
        <h1 class="text-xl font-bold">管理周期</h1>
        <select
          v-model="selectedWordbook"
          class="select select-bordered select-sm"
          :disabled="wordbookOptions.length === 0"
        >
          <option
            v-for="wb in wordbookOptions"
            :key="wb.code"
            :value="wb.code"
          >
            {{ wb.shortName }}（{{ wb.fullName }}）
          </option>
        </select>
        <button class="btn btn-primary btn-sm ml-auto" @click="startCreateMode">+ 新增周期</button>
        <button class="btn btn-error btn-sm btn-outline" @click="showClearConfirm = true">删除所有周期</button>
      </div>

      <div v-if="showClearConfirm" class="modal modal-open">
        <div class="modal-box">
          <h3 class="font-bold text-lg">确认删除所有周期？</h3>
          <p class="py-4 text-base-content/70">
            此操作会删除该词库下全部周期与相关进度，且不可撤销。
          </p>
          <div class="modal-action">
            <button class="btn btn-ghost" :disabled="clearing" @click="showClearConfirm = false">取消</button>
            <button class="btn btn-error" :disabled="clearing" @click="clearCycles">
              <span v-if="clearing" class="loading loading-spinner loading-sm" />
              {{ clearing ? '删除中...' : '确认删除' }}
            </button>
          </div>
        </div>
        <div class="modal-backdrop" @click="showClearConfirm = false" />
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-[360px_1fr] gap-5">
        <section class="card bg-base-100 border border-base-300 shadow-sm">
          <div class="card-body gap-3">
            <h2 class="card-title text-base">周期列表</h2>
            <div v-if="loadingList" class="flex justify-center py-6">
              <span class="loading loading-spinner loading-md" />
            </div>
            <template v-else>
              <div v-if="cycles.length === 0" class="text-sm text-base-content/60">
                暂无周期，请点击“新增周期”开始录入。
              </div>
              <div v-else class="flex flex-col gap-2">
                <button
                  v-for="cycle in cycles"
                  :key="cycle.cycleId"
                  class="btn justify-start h-auto py-3"
                  :class="cycle.cycleId === selectedCycleId && !isCreateMode ? 'btn-primary' : 'btn-ghost border border-base-300'"
                  @click="loadCycleDetail(cycle.cycleId)"
                >
                  <div class="flex flex-col items-start text-left w-full">
                    <div class="flex items-center gap-2">
                      <span class="font-medium">周期 #{{ cycle.cycleId }}</span>
                      <span
                        class="badge badge-sm"
                        :class="cycle.status === 'ongoing' ? 'badge-info' : (cycle.status === 'queued' ? 'badge-warning' : 'badge-neutral')"
                      >
                        {{ cycle.status === 'ongoing' ? '进行中' : (cycle.status === 'queued' ? '待生效' : '已完成') }}
                      </span>
                    </div>
                    <span class="text-xs opacity-70">
                      {{ cycle.knownWords }}/{{ cycle.totalWords }} 已掌握 · {{ formatCreatedAt(cycle.createdAt) }}
                    </span>
                  </div>
                </button>
              </div>
            </template>
          </div>
        </section>

        <section class="card bg-base-100 border border-base-300 shadow-sm">
          <div class="card-body gap-4">
            <h2 class="card-title text-base">{{ modeTitle }}</h2>

            <div v-if="loadingDetail" class="flex justify-center py-10">
              <span class="loading loading-spinner loading-md" />
            </div>

            <template v-else>
              <p class="text-base-content/60 text-sm">
                每行一个单词，按 <kbd class="kbd kbd-sm">Enter</kbd> 换行。
                <template v-if="isReadonlyCycle">仅当前进行中周期可编辑；后续/历史周期仅支持查看。</template>
                <template v-else-if="!isCreateMode">
                  当前进行中周期可编辑并提交更新。
                  <span v-if="knownTerms.size > 0" class="text-warning">⚠ 已掌握词删除后进度不会回退。</span>
                </template>
                <template v-else>新增周期至少输入 1 个单词。</template>
              </p>

              <div v-if="knownTerms.size > 0" class="flex flex-wrap gap-1 text-xs">
                <span class="text-base-content/50">已掌握：</span>
                <span
                  v-for="t in [...knownTerms]"
                  :key="t"
                  class="badge badge-success badge-sm"
                >{{ t }}</span>
              </div>

              <textarea
                v-model="textareaVal"
                placeholder="输入单词，每行一个…"
                rows="10"
                :disabled="isReadonlyCycle"
                class="textarea textarea-bordered w-full text-sm font-mono leading-relaxed focus:textarea-primary resize-y disabled:opacity-80"
              />

              <div class="flex flex-col gap-2">
                <div class="flex justify-between text-sm text-base-content/50">
                  <span>已录入 {{ tags.length }} 个</span>
                  <span v-if="canSubmit" class="text-success font-medium">可提交</span>
                </div>
                <div v-if="results.length > 0" class="flex flex-wrap gap-2">
                  <span
                    v-for="tag in tags"
                    :key="tag"
                    :class="[tagClass(tag), 'select-none']"
                  >{{ tag }}</span>
                </div>
              </div>

              <div v-if="errorMsg" class="alert alert-error text-sm">{{ errorMsg }}</div>
              <div v-if="successMsg" class="alert alert-success text-sm">
                {{ successMsg }}
              </div>

              <div v-if="results.length > 0 && !successMsg" class="flex gap-3 text-xs text-base-content/60">
                <span><span class="badge badge-error badge-sm mr-1" />不在词典中</span>
                <span><span class="badge badge-warning badge-sm mr-1" />其他周期已用过</span>
                <span><span class="badge badge-success badge-sm mr-1" />有效</span>
              </div>

              <button
                class="btn btn-primary self-center"
                :disabled="!canSubmit"
                @click="submit"
              >
                <span v-if="submitting" class="loading loading-spinner loading-sm" />
                {{ submitLabel }}
              </button>
            </template>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>
