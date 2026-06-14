import type {
  ApiResult,
  AuthState,
  BatchDocumentsRequest,
  BatchResult,
  ChunkItem,
  ChunkPreviewResponse,
  ChunkPreviewVariantInput,
  DocumentItem,
  HealthInfo,
  Overview,
  PageResult,
  QueryFilters,
  QueryResponse,
  QuerySource,
  SyncStatusResponse,
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

async function readApiError(response: Response) {
  try {
    const result = (await response.json()) as ApiResult<unknown>
    return new Error(result.message || '请求失败')
  } catch {
    return new Error('请求失败')
  }
}

type QueryStreamHandlers = {
  onStatus?: (payload: { stage: string; message: string }) => void
  onSources?: (sources: QuerySource[]) => void
  onDelta?: (text: string) => void
  onDone?: (response: QueryResponse) => void
  onError?: (message: string) => void
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

export function syncStatus() {
  return request<SyncStatusResponse>('/api/v1/admin/sync/status')
}

export function runSync(
  source: 'nav_sites' | 'game_details' | 'game_news' | 'all',
) {
  return request<{ accepted: boolean; source: string }>('/api/v1/admin/sync/run', {
    method: 'POST',
    body: JSON.stringify({ source }),
  })
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

export function listDocuments(params: {
  page: number
  page_size: number
  status: string
  keyword: string
  source_type?: string[]
  category?: string
  language?: string
}) {
  const query = new URLSearchParams({
    page: String(params.page),
    page_size: String(params.page_size),
  })
  if (params.status) query.set('status', params.status)
  if (params.keyword) query.set('keyword', params.keyword)
  if (params.source_type?.length) query.set('source_type', params.source_type.join(','))
  if (params.category) query.set('category', params.category)
  if (params.language) query.set('language', params.language)
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

export function batchReindexDocuments(payload: BatchDocumentsRequest) {
  return request<BatchResult>('/api/v1/admin/documents/reindex', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function retryFailedDocuments(payload: BatchDocumentsRequest) {
  return request<BatchResult>('/api/v1/admin/documents/retry-failed', {
    method: 'POST',
    body: JSON.stringify(payload),
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

export function queryRag(question: string, topK: number, filters?: QueryFilters, includeDetails = false) {
  return request<QueryResponse>('/api/v1/chat/query', {
    method: 'POST',
    body: JSON.stringify({ question, top_k: topK, filters, include_details: includeDetails }),
  })
}

export async function queryRagStream(
  question: string,
  topK: number,
  filters: QueryFilters | undefined,
  handlers: QueryStreamHandlers = {},
  signal?: AbortSignal,
  includeDetails = false,
) {
  const response = await fetch('/api/v1/chat/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ question, top_k: topK, filters, include_details: includeDetails }),
    credentials: 'include',
    signal,
  })

  if (!response.ok) {
    throw await readApiError(response)
  }
  if (!response.body) {
    throw new Error('流式响应不可用')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let eventName = ''
  let dataLines: string[] = []
  let finalResponse: QueryResponse | null = null
  let answerText = ''
  let sourceList: QuerySource[] = []

  const dispatch = () => {
    const payload = dataLines.join('\n').trim()
    dataLines = []
    const currentEvent = eventName || 'message'
    eventName = ''
    if (!payload) return
    if (payload === '[DONE]') return
    try {
      if (currentEvent === 'status') {
        const parsed = JSON.parse(payload) as { stage: string; message: string }
        handlers.onStatus?.(parsed)
        return
      }
      if (currentEvent === 'sources') {
        const parsed = JSON.parse(payload) as { sources: QuerySource[] }
        sourceList = parsed.sources || []
        handlers.onSources?.(sourceList)
        return
      }
      if (currentEvent === 'delta') {
        const parsed = JSON.parse(payload) as { text: string }
        answerText += parsed.text || ''
        handlers.onDelta?.(parsed.text || '')
        return
      }
      if (currentEvent === 'error') {
        const parsed = JSON.parse(payload) as { message?: string }
        const message = parsed.message || '请求失败'
        handlers.onError?.(message)
        throw new Error(message)
      }
      if (currentEvent === 'done') {
        finalResponse = JSON.parse(payload) as QueryResponse
        handlers.onDone?.(finalResponse)
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : '流式数据解析失败'
      handlers.onError?.(message)
      throw new Error(message)
    }
  }

  const consumeLine = (line: string) => {
    if (line.startsWith('event:')) {
      eventName = line.slice('event:'.length).trim()
      return
    }
    if (line.startsWith('data:')) {
      dataLines.push(line.slice('data:'.length).trim())
      return
    }
    if (line === '') {
      dispatch()
    }
  }

  try {
    while (true) {
      const { value, done } = await reader.read()
      if (value) {
        buffer += decoder.decode(value, { stream: true })
        let newlineIndex = buffer.indexOf('\n')
        while (newlineIndex >= 0) {
          const line = buffer.slice(0, newlineIndex).replace(/\r$/, '')
          buffer = buffer.slice(newlineIndex + 1)
          consumeLine(line)
          newlineIndex = buffer.indexOf('\n')
        }
      }
      if (done) {
        break
      }
    }

    if (buffer.trim()) {
      consumeLine(buffer.replace(/\r$/, ''))
    }
    dispatch()
  } finally {
    reader.releaseLock()
  }

  if (!finalResponse) {
    finalResponse = {
      answer: answerText,
      sources: sourceList,
      usage: {
        top_k: topK,
        embedding_model: '',
      },
    }
  }
  return finalResponse
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
