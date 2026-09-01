export type InsightRange = '30d' | '90d' | 'all'
export type NavInsightMetricKey = 'ipv6' | 'tls13' | 'security_txt'
export type GameInsightMetricKey = 'free' | 'windows' | 'linux'
export type InsightMetricKey = NavInsightMetricKey | GameInsightMetricKey
export type InsightDomain = 'site' | 'game'

export interface InsightMetric {
  key: InsightMetricKey
  as_of: string
  value: number | null
  delta_30d: number | null
  coverage: number | null
  known: number
  eligible: number
  available_from: string | null
}

export interface InsightTrendPoint {
  date: string
  value: number | null
  coverage: number | null
}

export interface InsightMetricTrend {
  key: InsightMetricKey
  requested_range: InsightRange
  available_from: string | null
  available_through: string | null
  points: InsightTrendPoint[]
}

export interface InsightEntityRef {
  id: number
  name: string
}

export interface InsightChange {
  type: string
  date: string
  occurred_at: string | null
  entity: InsightEntityRef
  detail: unknown | null
}

export interface InsightOverview {
  generated_at: string
  entity_count: number
  changes_7d: number
  metrics: InsightMetric[]
  recent_changes: InsightChange[]
}

export interface InsightFeedItem extends InsightChange {
  domain: InsightDomain
}
