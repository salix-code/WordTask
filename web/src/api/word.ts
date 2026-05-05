import { http } from './index'
import type { ReviewQuality } from '@/types/api'
import type { Word } from '@/types/word'
import { useUserStore } from '@/store/userStore'

export interface TodayWordsResponse {
  dailyGoal: number
  noCycle: boolean
  completed: boolean
  words: Word[]
}

export interface ReviewDueResponse {
  words: Word[]
}

/** 获取今日待复习单词 */
export function fetchTodayWords(wordbook: string): Promise<TodayWordsResponse> {
  const userStore = useUserStore()
  return http
    .get('/words/today', {
      params: {
        userId: userStore.userId !== null ? String(userStore.userId) : null,
        wordbook,
      },
    })
    .then((r) => r.data)
}

/** 提交单词复习结果；quality='known' 时后端将单词标记为已掌握 */
export function submitReview(wordId: string, quality: ReviewQuality, wordbook: string): Promise<void> {
  const userStore = useUserStore()
  return http
    .post('/words/review', {
      userId: userStore.userId !== null ? String(userStore.userId) : null,
      wordbook,
      wordId: Number(wordId),
      quality,
    })
    .then(() => void 0)
}

/** 获取已完成周期中到期的复习单词（SM-2） */
export function fetchReviewDueWords(wordbook: string, limit: number): Promise<ReviewDueResponse> {
  const userStore = useUserStore()
  return http
    .get('/words/review/due', {
      params: {
        userId: userStore.userId !== null ? String(userStore.userId) : null,
        wordbook,
        limit,
      },
    })
    .then((r) => r.data)
}

/** 提交复习模式评分（SM-2 调度） */
export function submitRevision(wordId: string, quality: ReviewQuality, wordbook: string): Promise<void> {
  const userStore = useUserStore()
  return http
    .post('/words/review/revision', {
      userId: userStore.userId !== null ? String(userStore.userId) : null,
      wordbook,
      wordId: Number(wordId),
      quality,
    })
    .then(() => void 0)
}
