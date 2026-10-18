import type { InsightEntityRef } from './insights'

export interface ComparePickerEntity extends InsightEntityRef { subtitle?: string }
export interface CompareMatrixCell { text: string; state?: string; attributes?: Record<string, string | boolean> }
export interface CompareMatrixGroup {
  key: string
  label: string
  rows: { key: string; label: string; cells: CompareMatrixCell[] }[]
}
