export type InsightRange = '30d' | '90d' | 'all'
export type NavInsightMetricKey = 'ipv6' | 'tls13' | 'security_txt'
export type GameInsightMetricKey = 'free' | 'windows' | 'linux'
export type InsightMetricKey = NavInsightMetricKey | GameInsightMetricKey
export type InsightDomain = 'site' | 'game'
export type SiteInsightCapabilityKey = NavInsightMetricKey
export type SiteInsightCapabilityState =
  | 'supported'
  | 'unsupported'
  | 'stale'
  | 'not_probed'
  | 'unavailable'
  | 'unknown'
  | 'not_applicable'
export type GameInsightPriceState = 'free' | 'priced' | 'unknown' | 'unpriced'

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

export interface SiteInsightCapability {
  key: SiteInsightCapabilityKey
  as_of: string | null
  state: SiteInsightCapabilityState
  ecosystem: {
    value: number | null
    coverage: number | null
  }
}

export interface SiteInsights {
  site: InsightEntityRef
  capabilities: SiteInsightCapability[]
  recent_changes: InsightChange[]
}

export interface GameInsightState {
  free: boolean | null
  windows: boolean | null
  linux: boolean | null
  release: string | null
  as_of: string | null
}

export interface GameInsightPlayerSummary {
  current: number | null
  peak_30d: number | null
  as_of: string | null
}

export interface GameInsightPrice {
  region: 'CN'
  state: GameInsightPriceState
  currency: string | null
  initial_amount: number | null
  final_amount: number | null
  discount_percent: number | null
  as_of: string | null
}

export interface GameInsights {
  game: InsightEntityRef
  state: GameInsightState
  players: GameInsightPlayerSummary
  price: GameInsightPrice | null
  recent_changes: InsightChange[]
}

export interface GameInsightPlayerPoint {
  date: string
  min: number | null
  max: number
  avg: number | null
}

export interface GameInsightPlayerHistory {
  requested_range: InsightRange
  available_from: string | null
  available_through: string | null
  points: GameInsightPlayerPoint[]
}

export interface GameInsightPricePoint {
  date: string
  state: GameInsightPriceState
  currency: string | null
  initial_amount: number | null
  final_amount: number | null
  discount_percent: number | null
}

export interface GameInsightPriceHistory {
  requested_range: InsightRange
  available_from: string | null
  available_through: string | null
  points: GameInsightPricePoint[]
}
