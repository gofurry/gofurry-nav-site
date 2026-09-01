export type InsightRange = '30d' | '90d' | 'all'
export type InsightChangeRange = '7d' | InsightRange
export type NavInsightMetricKey = 'ipv6' | 'tls13' | 'security_txt'
export type GameInsightMetricKey = 'free' | 'windows' | 'linux'
export type InsightMetricKey = NavInsightMetricKey | GameInsightMetricKey
export type InsightDomain = 'site' | 'game'
export type SiteInsightDimension = 'country' | 'group' | 'nsfw' | 'public_interest'
export type GameInsightDimension = 'primary_tag' | 'tag'
export type InsightDimension = SiteInsightDimension | GameInsightDimension
export type InsightSliceMode = 'partition' | 'overlapping'
export type SiteInsightChangeCategory = 'capability' | 'target' | 'certificate'
export type GameInsightChangeCategory = 'pricing_model' | 'platform' | 'release' | 'price' | 'discount'
export type InsightChangeCategory = SiteInsightChangeCategory | GameInsightChangeCategory
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

export interface InsightDimensionSlice {
  value: string
  label: string | null
  label_en: string | null
  population: number
  eligible: number
  known: number
  metric_value: number | null
  coverage: number | null
}

export interface InsightDimensionBreakdown {
  key: InsightMetricKey
  dimension: InsightDimension
  slice_mode: InsightSliceMode
  as_of: string | null
  items: InsightDimensionSlice[]
}

export interface InsightDimensionSliceRef {
  value: string
  label: string | null
  label_en: string | null
}

export interface InsightDimensionTrendPoint {
  date: string
  population: number
  eligible: number
  known: number
  metric_value: number | null
  coverage: number | null
}

export interface InsightDimensionTrend {
  key: InsightMetricKey
  dimension: InsightDimension
  slice: InsightDimensionSliceRef
  slice_mode: InsightSliceMode
  requested_range: InsightRange
  as_of: string | null
  available_from: string | null
  available_through: string | null
  points: InsightDimensionTrendPoint[]
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

export interface InsightExplorerChange extends InsightChange {
  domain: InsightDomain
  category: InsightChangeCategory
}

export interface InsightChangeExplorerPage {
  items: InsightExplorerChange[]
  next_cursor: string | null
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
