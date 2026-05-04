export type ApiResult<T> = {
  code: number
  message: string
  data: T
}

export type PageResult<T> = {
  items: T[]
  total: number
}

export type DocumentItem = {
  id: number
  title: string
  source_type: string
  source_id?: string
  url?: string
  status: string
  error_message: string
  chunk_count: number
  created_at: string
  updated_at: string
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
  title: string
  url?: string
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
