import { http } from './index'
import type { WordbookItem } from '@/types/api'

export interface WordbookListResponse {
  wordbooks: WordbookItem[]
}

export function fetchWordbooks(): Promise<WordbookListResponse> {
  return http
    .get<WordbookListResponse>('/wordbooks')
    .then((r) => r.data)
}
