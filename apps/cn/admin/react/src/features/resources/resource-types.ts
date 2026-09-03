import type { z } from 'zod'

export type ResourceRecord = Record<string, unknown> & { id?: string | number }
export type ResourceFieldType = 'text' | 'textarea' | 'number' | 'boolean' | 'datetime' | 'select' | 'remote-select' | 'string-array'
export type ResourceOption = { label: string; value: string }

export type ResourceField = {
  key: string
  label: string
  type: ResourceFieldType
  help?: string
  placeholder?: string
  options?: ResourceOption[]
  optionEndpoint?: string
  section?: string
}

export type ResourceColumn<T extends ResourceRecord> = {
  key: string
  label: string
  hidden?: boolean
  format?: (value: unknown, row: T) => React.ReactNode
}

export type ResourceDefinition<T extends ResourceRecord = ResourceRecord> = {
  key: string
  section: 'nav' | 'game'
  title: string
  description: string
  listEndpoint: string
  detailEndpoint: string
  columns: ResourceColumn<T>[]
  fields: ResourceField[]
  defaults: T
  schema: z.ZodType<T>
}
