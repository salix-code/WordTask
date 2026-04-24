/**
 * 统一的后端响应结构：{ code, data, msg }
 * 与 CLAUDE.md 中约定的 API 返回格式保持一致
 */
export interface ApiResponse<T = unknown> {
  code: number
  data: T
  msg: string
}

export type ReviewQuality = 'forgot' | 'vague' | 'known'
