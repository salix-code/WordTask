/**
 * 单词数据结构（前端使用）
 */
export interface Word {
  id: string
  /** 词头 */
  text: string
  /** 音标（IPA） */
  phonetic?: string
  /** 释义列表（词性 + 释义） */
  definitions: Array<{
    pos?: string
    meaning: string
  }>
  /** 例句 */
  examples?: Array<{
    en: string
    zh?: string
  }>
}

/** 单次复习的记录 */
export interface ReviewRecord {
  wordId: string
  quality: 'forgot' | 'vague' | 'known'
  /** 毫秒时间戳 */
  reviewedAt: number
}
