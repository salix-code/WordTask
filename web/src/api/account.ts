import { http } from './index'

export interface LoginResponse {
  userId: number
  accountName: string
  isAdmin: boolean
}

export interface CreateAccountResponse {
  userId: number
  accountName: string
}

/**
 * 登录接口
 * @param accountName 账号名
 * @returns Promise with userId, accountName, and role flag
 */
export function login(accountName: string): Promise<LoginResponse> {
  return http.post('/account/login', { accountName }).then((r) => r.data)
}

export function createAccount(
  userId: number,
  accountName: string,
): Promise<CreateAccountResponse> {
  return http.post('/account/create', { accountName }, { params: { userId } }).then((r) => r.data)
}
