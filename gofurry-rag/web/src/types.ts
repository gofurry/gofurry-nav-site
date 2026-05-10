export type ApiResult<T> = {
  code: number
  message: string
  data: T
}

export type PageResult<T> = {
  items: T[]
  total: number
}

export type AuthState = {
  initialized: boolean
  authenticated: boolean
  session_version?: number
}

export type Overview = {
  document_total: number
  chunk_total: number
  embedded_chunk_total: number
  pending_documents: number
  processing_documents: number
  ready_documents: number
  failed_documents: number
  queue_documents: number
  recent_failure_message?: string
  recent_failure_at?: string
  recent_failed_document_id?: number
  last_document_update_at?: string
  worker_state?: string
  worker_active_workers?: number
  worker_current_document_id?: number
  worker_last_document_id?: number
  worker_total_processed?: number
  worker_total_failed?: number
  worker_last_duration_ms?: number
  worker_average_duration_ms?: number
  worker_recent_error?: string
  worker_recent_error_at?: string
  worker_last_success_at?: string
  worker_last_started_at?: string
  worker_last_completed_at?: string
}

export type HealthInfo = {
  status: string
  app_name?: string
  embedding_model?: string
  database?: {
    type?: string
    name?: string
    host?: string
    port?: string
    connected?: boolean
    error?: string
  }
  ollama?: {
    base_url?: string
    model?: string
    embed_dim?: number
    healthy?: boolean
    error?: string
  }
}

export type DocumentItem = {
  id: number
  title: string
  source_type: string
  source_id?: string
  url?: string
  status: string
  error_message: string
  metadata?: Record<string, unknown>
  chunk_count: number
  retry_count: number
  last_error_at?: string
  processed_at?: string
  reindex_requested_at?: string
  last_indexed_at?: string
  created_at: string
  updated_at: string
}

export type DocumentFilters = {
  source_type?: string[]
  category?: string[]
  language?: string[]
  status?: string[]
}

export type BatchDocumentsRequest = {
  scope: 'all' | 'filters' | 'document_ids'
  document_ids?: number[]
  filters?: DocumentFilters
}

export type BatchResult = {
  accepted_count: number
  skipped_count: number
  status: string
}

export type ChunkItem = {
  id: number
  document_id: number
  chunk_index: number
  content: string
  token_count: number
  content_hash: string
  has_embedding: boolean
  embedding_dim: number
  created_at: string
}

export type QuerySource = {
  document_id: number
  chunk_id: number
  source_type: string
  source_id?: string
  title: string
  url?: string
  chunk_index: number
  token_count: number
  score: number
  content: string
}

export type QueryResponse = {
  answer: string
  sources: QuerySource[]
  usage: {
    top_k: number
    embedding_model: string
  }
}

export type ChunkPreviewVariantInput = {
  chunk_size: number
  chunk_overlap: number
}

export type ChunkPreviewChunk = {
  index: number
  char_count: number
  content: string
}

export type ChunkPreviewVariant = {
  chunk_size: number
  chunk_overlap: number
  chunk_count: number
  min_chars: number
  max_chars: number
  avg_chars: number
  chunks: ChunkPreviewChunk[]
}

export type ChunkPreviewResponse = {
  source: string
  title: string
  variants: ChunkPreviewVariant[]
}

export type QueryFilters = {
  source_type?: string[]
  document_ids?: number[]
  category?: string[]
  language?: string[]
}
