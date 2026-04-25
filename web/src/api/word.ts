import http from './index'
import type { ReviewQuality } from '@/types/api'
import type { Word } from '@/types/word'

export interface TodayWordsResponse {
  dailyGoal: number
  completed: boolean
  words: Word[]
}

/** 获取今日待复习单词 */
export function fetchTodayWords(wordbook: string): Promise<TodayWordsResponse> {
  return http
    .get('/words/today', {
      params: {
        userId: '0000-0000-0000-0000',
        wordbook,
      },
    })
    .then((r) => {
      const raw = r.data as { dailyGoal: number; completed: boolean; words: Array<Omit<Word, 'id'> & { id: number }> }
      return {
        dailyGoal: raw.dailyGoal,
        completed: raw.completed,
        words: raw.words.map((w) => ({ ...w, id: String(w.id) })),
      }
    })
}

/** 通知服务器完成了一批，推进 offset */
export function advanceProgress(wordbook: string, count: number): Promise<void> {
  return http
    .post('/progress/advance', {
      userId: '0000-0000-0000-0000',
      wordbook,
      count,
    })
    .then(() => void 0)
}

/** 提交一次复习结果 */
export function submitReview(wordId: string, quality: ReviewQuality): Promise<void> {
  return http
    .post('/words/review', { wordId, quality, reviewedAt: Date.now() })
    .then(() => void 0)
}
