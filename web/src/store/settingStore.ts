import { defineStore } from 'pinia'

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
  }),
  actions: {
    setBatchSize(n: number) {
      this.batchSize = Math.max(1, Math.floor(n))
    },
    setDailyGoal(n: number) {
      this.dailyGoal = Math.max(this.batchSize, Math.floor(n))
    },
  },
})
