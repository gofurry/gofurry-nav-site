import type { ApiResult, PageResult } from './types'

type CsrfState = { token: string; headerName: string }
let csrfState: CsrfState | null = null

export class ApiError extends Error {
  readonly status: number
  readonly code?: number

  constructor(message: string, status: number, code?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    ...init,
  })
  let result: ApiResult<T>
  try { result = await response.json() as ApiResult<T> } catch { throw new ApiError('服务返回了无法解析的响应', response.status) }
  if (!response.ok || result.code !== 1) {
    if (response.status === 401) window.dispatchEvent(new Event('gofurry:unauthorized'))
    throw new ApiError(result.message || '请求失败', response.status, result.code)
  }
  return result.data
}

async function ensureCsrf() {
  if (csrfState) return csrfState
  const data = await request<{ token: string; header_name: string }>('/csrf/token')
  csrfState = { token: data.token, headerName: data.header_name }
  return csrfState
}

export function resetCsrf() { csrfState = null }
export function getJSON<T>(path: string) { return request<T>(path, { method: 'GET' }) }

export async function sendJSON<T>(path: string, method: 'POST' | 'PUT' | 'DELETE', body?: unknown) {
  const csrf = await ensureCsrf()
  return request<T>(path, { method, headers: { [csrf.headerName]: csrf.token }, body: body === undefined ? undefined : JSON.stringify(body) })
}

export function pageURL(path: string, page: number, pageSize: number, keyword = '') {
  const params = new URLSearchParams({ page_num: String(page), page_size: String(pageSize) })
  if (keyword.trim()) params.set('keyword', keyword.trim())
  return `${path}${path.includes('?') ? '&' : '?'}${params}`
}

export function listJSON<T>(path: string, page: number, pageSize: number, keyword = '') {
  return getJSON<PageResult<T>>(pageURL(path, page, pageSize, keyword))
}

export function errorMessage(error: unknown) { return error instanceof Error ? error.message : '操作失败，请稍后重试' }
