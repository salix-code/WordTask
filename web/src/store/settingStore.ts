import { defineStore } from 'pinia'
import { fetchWordbooks } from '@/api/wordbook'
import type { WordbookItem } from '@/types/api'

/**
 * 用户偏好设置（批量大小等）
 * M1 阶段仅本地持久化；M3 可同步到后端
 */
export const useSettingStore = defineStore('setting', {
  state: () => ({
    /** 每组批次大小（"闯关模式"） */
    batchSize: 5,
    /** 每日目标 */
    dailyGoal: 20,
    /** 当前使用的词典 */
    wordbook: 'KET',
    /** 后端启用的词典列表 */
    wordbookOptions: [] as WordbookItem[],
    wordbookOptionsLoaded: false,
  }),
  actions: {
    setBatchSize(n: number) {
      this.batchSize = Math.max(1, Math.floor(n))
    },
    setDailyGoal(n: number) {
      this.dailyGoal = Math.max(this.batchSize, Math.floor(n))
    },
    setWordbook(code: string) {
      const candidate = this.wordbookOptions.find((w) => w.code === code)
      if (!candidate) return
      this.wordbook = candidate.code
    },
    async ensureWordbooksLoaded() {
      if (this.wordbookOptionsLoaded) {
        return
      }
      const res = await fetchWordbooks()
      this.wordbookOptions = res.wordbooks
      this.wordbookOptionsLoaded = true

      if (this.wordbookOptions.length === 0) {
        this.wordbook = ''
        return
      }
      if (!this.wordbookOptions.some((w) => w.code === this.wordbook)) {
        this.wordbook = this.wordbookOptions[0].code
      }
    },
  },
})
