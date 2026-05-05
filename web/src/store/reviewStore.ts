import { defineStore } from 'pinia'
import type { ReviewQuality } from '@/types/api'
import type { ReviewRecord, Word } from '@/types/word'
import { fetchReviewDueWords, fetchTodayWords } from '@/api/word'
import { useSettingStore } from './settingStore'

export type ReviewMode = 'learn' | 'revision'

interface ReviewState {
  /** 当前模式：学习新词 / 科学复习 */
  mode: ReviewMode
  /** 今日所有待复习单词（M1 从 mock；M2 从 /api/words/today） */
  dailyQueue: Word[]
  /** dailyQueue 中下一个待切入 session 的位置 */
  dailyIndex: number
  /** 该词库是否已全部学完 */
  wordbookCompleted: boolean
  /** 当前用户没有正在进行的周期 */
  noCycle: boolean
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
    mode: 'learn',
    dailyQueue: [],
    dailyIndex: 0,
    wordbookCompleted: false,
    noCycle: false,
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
    /** dailyQueue 中还未切入 session 的剩余数量 */
    dailyRemaining(state): number {
      return Math.max(0, state.dailyQueue.length - state.dailyIndex)
    },
    /** 当前批次进度字符串：2/5 */
    sessionProgressText(state): string {
      const total = state.sessionQueue.length
      const cur = Math.min(state.sessionIndex + 1, total)
      return total ? `${cur}/${total}` : '0/0'
    },
  },

  actions: {
    /**
     * 重置所有与当前词库/周期相关的状态。
     * 在录入新周期成功、切换词库等场景调用，触发下一次 initDaily 时重新拉取。
     */
    reset() {
      this.mode = 'learn'
      this.dailyQueue = []
      this.dailyIndex = 0
      this.wordbookCompleted = false
      this.noCycle = false
      this.dailyTotal = 0
      this.dailyCompleted = 0
      this.sessionQueue = []
      this.sessionIndex = 0
      this.flipped = false
      this.sessionResults = []
      this.history = []
    },

    /** 初始化今日队列：从服务器获取今日单词 */
    async initDaily() {
      this.mode = 'learn'
      const setting = useSettingStore()
      const response = await fetchTodayWords(setting.wordbook)
      if (response.noCycle) {
        this.noCycle = true
        this.wordbookCompleted = false
        this.dailyQueue = []
        this.dailyIndex = 0
        this.dailyTotal = 0
        this.dailyCompleted = 0
        return
      }
      if (response.completed) {
        this.noCycle = false
        this.wordbookCompleted = true
        this.dailyQueue = []
        this.dailyIndex = 0
        this.dailyTotal = 0
        this.dailyCompleted = 0
        return
      }
      this.noCycle = false
      this.wordbookCompleted = false
      this.dailyQueue = response.words
      this.dailyIndex = 0
      this.dailyTotal = response.words.length
      this.dailyCompleted = 0
      this.history = []
      this.startNextSession()
    },

    /** 初始化科学复习队列：从已完成周期拉取到期单词（SM-2） */
    async initRevision() {
      this.mode = 'revision'
      const setting = useSettingStore()
      const response = await fetchReviewDueWords(setting.wordbook, setting.batchSize)
      this.noCycle = false
      this.wordbookCompleted = false
      this.dailyQueue = response.words
      this.dailyIndex = 0
      this.dailyTotal = response.words.length
      this.dailyCompleted = 0
      this.history = []
      this.startNextSession()
    },

    /** 完成当前批次，开始下一批（进度已通过 submitReview per-word 上报） */
    async completeSession() {
      this.startNextSession()
    },

    /** 从 dailyQueue[dailyIndex ...] 中切出下一批（默认 5 个） */
    startNextSession() {
      const setting = useSettingStore()
      const start = this.dailyIndex
      const end = Math.min(start + setting.batchSize, this.dailyQueue.length)
      this.sessionQueue = this.dailyQueue.slice(start, end)
      this.dailyIndex = end
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
