import http from './index'
import type { ReviewQuality } from '@/types/api'
import type { Word } from '@/types/word'

/** 获取今日待复习单词（M2 联调启用） */
export function fetchTodayWords(limit?: number): Promise<Word[]> {
  return http
    .get<Word[]>('/words/today', { params: limit ? { limit } : {} })
    .then((r) => r.data as unknown as Word[])
}

/** 提交一次复习结果 */
export function submitReview(wordId: string, quality: ReviewQuality): Promise<void> {
  return http
    .post('/words/review', { wordId, quality, reviewedAt: Date.now() })
    .then(() => void 0)
}
