import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useReviewStore } from '@/store/reviewStore'
import { useSettingStore } from '@/store/settingStore'

describe('reviewStore 批次与日进度', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initDaily 会初始化 dailyTotal 并切出首个 session', () => {
    const review = useReviewStore()
    const setting = useSettingStore()
    setting.setBatchSize(5)
    review.initDaily()
    expect(review.dailyTotal).toBeGreaterThan(0)
    expect(review.sessionQueue.length).toBe(Math.min(5, review.dailyTotal))
    expect(review.sessionIndex).toBe(0)
  })

  it('submitCurrentWord 会推进 sessionIndex 与 dailyCompleted', () => {
    const review = useReviewStore()
    review.initDaily()
    const before = review.dailyCompleted
    review.submitCurrentWord('known')
    expect(review.dailyCompleted).toBe(before + 1)
    expect(review.sessionIndex).toBe(1)
    expect(review.flipped).toBe(false)
  })

  it('一个批次全部评分后 isSessionFinished 为 true', () => {
    const review = useReviewStore()
    const setting = useSettingStore()
    setting.setBatchSize(3)
    review.initDaily()
    for (let i = 0; i < 3; i++) review.submitCurrentWord('known')
    expect(review.isSessionFinished).toBe(true)
  })
})
