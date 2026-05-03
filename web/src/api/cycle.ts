import { http } from './index'
import type { SetupCycleResult } from '@/types/cycle'
import { useUserStore } from '@/store/userStore'

export interface SetupCyclePayload {
  userId: string | null
  wordbook: string
  words: string[]
}

/**
 * 提交周期单词表，由后端统一校验并创建新周期。
 * 返回 SetupCycleResult：
 *   ok=true  → 所有单词有效，周期已创建
 *   ok=false → 存在问题单词，results 中每项均有 status 标记
 */
export function setupCycle(payload: Omit<SetupCyclePayload, 'userId'>): Promise<SetupCycleResult> {
  const userStore = useUserStore()
  const fullPayload: SetupCyclePayload = {
    ...payload,
    userId: userStore.userId !== null ? String(userStore.userId) : null,
  }
  return http
    .post<SetupCycleResult>('/cycles/setup', fullPayload)
    .then((r) => r.data)
}
