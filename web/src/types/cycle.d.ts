/**
 * 单词校验状态码：
 *   valid        — 有效
 *   not_in_dict  — 不在词典中
 *   already_used — 已在当前进行中的周期中使用过
 */
export type WordStatusCode = 'valid' | 'not_in_dict' | 'already_used'

/** 单个单词的校验结果 */
export interface WordStatus {
  term: string
  status: WordStatusCode
}

/** POST /api/cycles/setup 的响应体（嵌套在通用 ApiResponse.data 中） */
export interface SetupCycleResult {
  ok: boolean
  results: WordStatus[]
}
