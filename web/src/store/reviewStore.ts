import { defineStore } from 'pinia'
import type { ReviewQuality } from '@/types/api'
import type { ReviewRecord, Word } from '@/types/word'
import { mockWords } from '@/mock/words'
import { useSettingStore } from './settingStore'

interface ReviewState {
  /** 今日所有待复习单词（M1 从 mock；M2 从 /api/words/today） */
  dailyQueue: Word[]
  /** 今日总量（= dailyQueue.length） */
  dailyTotal: number
  /** 今日已完成数量 */
  dailyCompleted: number
  /** 当前一批（5 个）单词 */
  sessionQueue: Word[]
  /** 当前批次中的下标 */
  sessionIndex: number
  /** 卡片是否翻转 */
  flipped: boolean
  /** 本次 Session 的结算结果 */
  sessionResults: ReviewRecord[]
  /** 全部复习记录（供后续统计 / 上传） */
  history: ReviewRecord[]
}

export const useReviewStore = defineStore('review', {
  state: (): ReviewState => ({
    dailyQueue: [],
    dailyTotal: 0,
    dailyCompleted: 0,
    sessionQueue: [],
    sessionIndex: 0,
    flipped: false,
    sessionResults: [],
    history: [],
  }),

  getters: {
    currentWord(state): Word | undefined {
      return state.sessionQueue[state.sessionIndex]
    },
    /** 当前批次是否已完成 */
    isSessionFinished(state): boolean {
      return state.sessionQueue.length > 0 && state.sessionIndex >= state.sessionQueue.length
    },
    /** 今日整体是否完成 */
    isDailyFinished(state): boolean {
      return state.dailyCompleted >= state.dailyTotal && state.dailyTotal > 0
    },
    /** 当前批次进度字符串：2/5 */
    sessionProgressText(state): string {
      const total = state.sessionQueue.length
      const cur = Math.min(state.sessionIndex + 1, total)
      return total ? `${cur}/${total}` : '0/0'
    },
  },

  actions: {
    /** 初始化今日队列（M1：从 mock；M2：替换为 fetchTodayWords） */
    initDaily() {
      this.dailyQueue = [...mockWords]
      this.dailyTotal = this.dailyQueue.length
      this.dailyCompleted = 0
      this.history = []
      this.startNextSession()
    },

    /** 从 dailyQueue 中切出下一批（默认 5 个） */
    startNextSession() {
      const setting = useSettingStore()
      const batch = this.dailyQueue.splice(0, setting.batchSize)
      this.sessionQueue = batch
      this.sessionIndex = 0
      this.flipped = false
      this.sessionResults = []
    },

    flipCard() {
      this.flipped = !this.flipped
    },

    /**
     * 记录当前单词的评分并进入下一张
     * 注意：这里只做本地状态，API 上传在 views 层异步触发，避免阻塞交互
     */
    submitCurrentWord(quality: ReviewQuality) {
      const word = this.currentWord
      if (!word) return
      const record: ReviewRecord = {
        wordId: word.id,
        quality,
        reviewedAt: Date.now(),
      }
      this.sessionResults.push(record)
      this.history.push(record)
      this.dailyCompleted += 1
      this.sessionIndex += 1
      this.flipped = false
    },
  },
})
