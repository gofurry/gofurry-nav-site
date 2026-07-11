import type { ApiResult, PageResult } from './types'

type CsrfState = {
  token: string
  headerName: string
}

let csrfState: CsrfState | null = null

async function rawRequest<T>(path: string, init?: RequestInit): Promise<ApiResult<T>> {
  const response = await fetch(path, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  const result = (await response.json()) as ApiResult<T>
  if (!response.ok || result.code !== 1) {
    throw new Error(result.message || '请求失败')
  }
  return result
}

export async function ensureCsrf() {
  if (csrfState) {
    return csrfState
  }
  const result = await rawRequest<{ token: string; header_name: string }>('/csrf/token')
  csrfState = {
    token: result.data.token,
    headerName: result.data.header_name,
  }
  return csrfState
}

export function resetCsrf() {
  csrfState = null
}

export async function getJSON<T>(path: string) {
  const result = await rawRequest<T>(path, { method: 'GET' })
  return result.data
}

export async function sendJSON<T>(path: string, method: 'POST' | 'PUT' | 'DELETE', body?: unknown) {
  const headers: Record<string, string> = {}
  if (method === 'POST' || method === 'PUT' || method === 'DELETE') {
    const csrf = await ensureCsrf()
    headers[csrf.headerName] = csrf.token
  }
  const result = await rawRequest<T>(path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  return result.data
}

export async function listJSON<T>(path: string, pageNum: number, pageSize: number, keyword: string) {
  const params = new URLSearchParams({
    page_num: String(pageNum),
    page_size: String(pageSize),
  })
  if (keyword.trim()) {
    params.set('keyword', keyword.trim())
  }
  return getJSON<PageResult<T>>(`${path}?${params.toString()}`)
}

export async function listAllJSON<T>(path: string, keyword = '', pageSize = 200) {
  let pageNum = 1
  let total = 0
  const list: T[] = []

  while (pageNum === 1 || list.length < total) {
    const result = await listJSON<T>(path, pageNum, pageSize, keyword)
    total = result.total

    if (result.list.length === 0) {
      break
    }

    list.push(...result.list)
    pageNum += 1
  }

  return { total, list }
}
