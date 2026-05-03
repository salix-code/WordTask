import { http } from './index'

export interface LoginResponse {
  userId: number
  accountName: string
}

/**
 * 登录接口
 * @param accountName 账号名
 * @returns Promise with userId and accountName
 */
export function login(accountName: string): Promise<LoginResponse> {
  return http.post('/account/login', { accountName }).then((r) => r.data)
}
