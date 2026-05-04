import type { ApiResult, ChunkItem, DocumentItem, PageResult, QueryResponse } from './types'

const TOKEN_KEY = 'gofurry-rag-admin-token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token.trim())
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  const response = await fetch(path, {
    ...init,
    headers,
  })
  const result = (await response.json()) as ApiResult<T>
  if (!response.ok || result.code !== 1) {
    throw new Error(result.message || '请求失败')
  }
  return result.data
}

export function health() {
  return request<Record<string, unknown>>('/api/v1/health')
}

export function createTextDocument(payload: {
  title: string
  content: string
  source_type: string
  source_id: string
  url: string
  metadata: Record<string, unknown>
}) {
  return request<{ document_id: number; status: string }>('/api/v1/admin/documents/text', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function listDocuments(params: { page: number; page_size: number; status: string; keyword: string }) {
  const query = new URLSearchParams({
    page: String(params.page),
    page_size: String(params.page_size),
  })
  if (params.status) query.set('status', params.status)
  if (params.keyword) query.set('keyword', params.keyword)
  return request<PageResult<DocumentItem>>(`/api/v1/admin/documents?${query}`)
}

export function listChunks(documentId: number, page = 1, pageSize = 20) {
  return request<PageResult<ChunkItem>>(
    `/api/v1/admin/documents/${documentId}/chunks?page=${page}&page_size=${pageSize}`,
  )
}

export function deleteDocument(documentId: number) {
  return request<{ deleted: boolean }>(`/api/v1/admin/documents/${documentId}`, { method: 'DELETE' })
}

export function queryRag(question: string, topK: number) {
  return request<QueryResponse>('/api/v1/chat/query', {
    method: 'POST',
    body: JSON.stringify({ question, top_k: topK }),
  })
}
