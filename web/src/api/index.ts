import axios, { type AxiosInstance, AxiosError } from 'axios'
import type { ApiResponse } from '@/types/api'

const http: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10_000,
  headers: { 'Content-Type': 'application/json' },
})

// 请求拦截：附带 Token（M3 启用）
http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 响应拦截：统一剥离 { code, data, msg } 外层
http.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (body && typeof body.code === 'number') {
      if (body.code !== 200) {
        return Promise.reject(new Error(body.msg || `业务错误 ${body.code}`))
      }
      // 将 data 透出到调用方
      response.data = body.data
    }
    return response
  },
  (error: AxiosError) => {
    // 网络/HTTP 级别的错误
    const msg = (error.response?.data as ApiResponse)?.msg || error.message
    return Promise.reject(new Error(msg))
  },
)

export default http
