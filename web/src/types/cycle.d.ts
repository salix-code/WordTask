/**
 * 单词校验状态码：
 *   valid        — 有效
 *   not_in_dict  — 不在词典中
 *   already_used — 已在其他周期中使用过
 */
export type WordStatusCode = 'valid' | 'not_in_dict' | 'already_used'

/** 单个单词的校验结果 */
export interface WordStatus {
  term: string
  status: WordStatusCode
}

/** POST /api/cycles/setup 的响应体 */
export interface SetupCycleResult {
  ok: boolean
  results: WordStatus[]
}

/** GET /api/cycles/current 的响应体 */
export interface CurrentCycleWord {
  term: string
  status: 'new' | 'known'
}

export interface CurrentCycleResult {
  hasCycle: boolean
  words: CurrentCycleWord[]
}

/** PUT /api/cycles/current 的响应体（与 SetupCycleResult 结构相同） */
export type UpdateCycleResult = SetupCycleResult

/** 周期摘要（列表项） */
export interface CycleSummary {
  cycleId: number
  status: 'ongoing' | 'queued' | 'completed'
  createdAt: string
  totalWords: number
  knownWords: number
}

/** GET /api/cycles 的响应体 */
export interface CycleListResult {
  cycles: CycleSummary[]
}

/** GET /api/cycles/:id 的响应体 */
export interface CycleDetailResult {
  cycleId: number
  status: 'ongoing' | 'queued' | 'completed'
  createdAt: string
  words: CurrentCycleWord[]
}
