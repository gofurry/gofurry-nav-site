import type {
  ApiResult,
  AuthState,
  ChunkItem,
  ChunkPreviewResponse,
  ChunkPreviewVariantInput,
  DocumentItem,
  HealthInfo,
  Overview,
  PageResult,
  QueryResponse,
} from './types'

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const response = await fetch(path, {
    ...init,
    headers,
    credentials: 'include',
  })
  const result = (await response.json()) as ApiResult<T>
  if (!response.ok || result.code !== 1) {
    throw new Error(result.message || '请求失败')
  }
  return result.data
}

export function authState() {
  return request<AuthState>('/api/v1/admin/auth/state')
}

export function login(password: string) {
  return request<AuthState>('/api/v1/admin/auth/login', {
    method: 'POST',
    body: JSON.stringify({ password }),
  })
}

export function logout() {
  return request<{ authenticated: boolean }>('/api/v1/admin/auth/logout', { method: 'POST' })
}

export function health() {
  return request<HealthInfo>('/api/v1/health')
}

export function overview() {
  return request<Overview>('/api/v1/admin/overview')
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

export function listChunks(documentId: number, page = 1, pageSize = 100) {
  return request<PageResult<ChunkItem>>(
    `/api/v1/admin/documents/${documentId}/chunks?page=${page}&page_size=${pageSize}`,
  )
}

export function deleteDocument(documentId: number) {
  return request<{ deleted: boolean }>(`/api/v1/admin/documents/${documentId}`, { method: 'DELETE' })
}

export function reindexDocument(documentId: number) {
  return request<{ document_id: number; status: string }>(`/api/v1/admin/documents/${documentId}/reindex`, {
    method: 'POST',
  })
}

export function updateChunk(chunkId: number, content: string) {
  return request<ChunkItem>(`/api/v1/admin/chunks/${chunkId}`, {
    method: 'PATCH',
    body: JSON.stringify({ content }),
  })
}

export function deleteChunk(chunkId: number) {
  return request<{ deleted: boolean }>(`/api/v1/admin/chunks/${chunkId}`, { method: 'DELETE' })
}

export function queryRag(question: string, topK: number) {
  return request<QueryResponse>('/api/v1/chat/query', {
    method: 'POST',
    body: JSON.stringify({ question, top_k: topK }),
  })
}

export function chunkPreview(payload: {
  document_id?: number
  text?: string
  variants?: ChunkPreviewVariantInput[]
}) {
  return request<ChunkPreviewResponse>('/api/v1/admin/debug/chunk-preview', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}
