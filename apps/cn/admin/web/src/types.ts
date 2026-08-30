export type ApiResult<T> = {
  code: number
  message: string
  data: T
}

export type PageResult<T> = {
  total: number
  list: T[]
}

export type AuthState = {
  initialized: boolean
  authenticated: boolean
  identity?: AuthIdentity
}

export type AuthIdentity = {
  account_id: number
  username: string
  display_name: string
  role: 'owner' | 'developer' | 'operator'
  status: 'active' | 'disabled'
  session_version: number
  capabilities: string[]
}

export type OptionItem = {
  id: string
  label: string
  extra?: string
}

export type SelectOption = {
  label: string
  value: string
}

export type KeyValue = {
  key: string
  value: string
}

export type FieldType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'bool'
  | 'string-array'
  | 'kv-array'
  | 'select'
  | 'remote-select'
  | 'remote-multi'
  | 'datetime'

export type ResourceField = {
  key: string
  label: string
  type: FieldType
  placeholder?: string
  options?: SelectOption[]
  optionEndpoint?: string
}

export type BulkReplaceConfig = {
  endpoint: string
  selectionEndpoint?: string
  title?: string
  description?: string
  ownerField: ResourceField
  targetField: ResourceField
}

export type ResourceConfig = {
  key: string
  title: string
  section: 'nav' | 'game'
  listEndpoint: string
  detailEndpoint: string
  columns: Array<{ key: string; label: string }>
  fields: ResourceField[]
  defaults: Record<string, unknown>
  bulkReplace?: BulkReplaceConfig
  additionalBulkReplace?: BulkReplaceConfig
}
