import http from './index'
import type { SetupCycleResult } from '@/types/cycle'

export interface SetupCyclePayload {
  userId: string
  wordbook: string
  words: string[]
}

/**
 * 提交周期单词表，由后端统一校验并创建新周期。
 * 返回 SetupCycleResult：
 *   ok=true  → 所有单词有效，周期已创建
 *   ok=false → 存在问题单词，results 中每项均有 status 标记
 */
export function setupCycle(payload: SetupCyclePayload): Promise<SetupCycleResult> {
  return http
    .post<SetupCycleResult>('/cycles/setup', payload)
    .then((r) => r.data)
}
