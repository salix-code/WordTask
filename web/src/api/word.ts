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

/** 获取今日待复习单词 */
export function fetchTodayWords(wordbook: string): Promise<TodayWordsResponse> {
  const userStore = useUserStore()
  return http
    .get('/words/today', {
      params: {
        userId: userStore.userId,
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
      userId: userStore.userId,
      wordbook,
      wordId: Number(wordId),
      quality,
    })
    .then(() => void 0)
}

