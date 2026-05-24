export interface ApiResult<T = unknown> {
  code: number
  data: T
  message?: string
  msg?: string
}

export class ApiError extends Error {
  status?: number
  code?: number

  constructor(message: string, status?: number, code?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export type ApiService = 'nav' | 'navV2' | 'game'
