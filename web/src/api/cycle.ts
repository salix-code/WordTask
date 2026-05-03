import { http } from './index'
import type { SetupCycleResult, CurrentCycleResult, UpdateCycleResult } from '@/types/cycle'
import { useUserStore } from '@/store/userStore'

export interface SetupCyclePayload {
  userId: string | null
  wordbook: string
  words: string[]
}

/**
 * 提交周期单词表，由后端统一校验并创建新周期。
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

/** 获取当前进行中周期的单词列表（无周期时 hasCycle=false） */
export function fetchCurrentCycle(wordbook: string): Promise<CurrentCycleResult> {
  const userStore = useUserStore()
  return http
    .get<CurrentCycleResult>('/cycles/current', {
      params: {
        userId: userStore.userId !== null ? String(userStore.userId) : null,
        wordbook,
      },
    })
    .then((r) => r.data)
}

/** 清空该 wordbook 的所有周期及进度，允许用户重新录入 */
export function clearAllCycles(wordbook: string): Promise<{ ok: boolean }> {
  const userStore = useUserStore()
  return http
    .delete<{ ok: boolean }>('/cycles/all', {
      params: {
        userId: userStore.userId !== null ? String(userStore.userId) : null,
        wordbook,
      },
    })
    .then((r) => r.data)
}

export function updateCycle(payload: Omit<SetupCyclePayload, 'userId'>): Promise<UpdateCycleResult> {
  const userStore = useUserStore()
  const fullPayload: SetupCyclePayload = {
    ...payload,
    userId: userStore.userId !== null ? String(userStore.userId) : null,
  }
  return http
    .put<UpdateCycleResult>('/cycles/current', fullPayload)
    .then((r) => r.data)
}
