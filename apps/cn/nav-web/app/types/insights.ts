export type InsightRange = '30d' | '90d' | 'all'
export type InsightChangeRange = '7d' | InsightRange
export type NavInsightMetricKey = 'ipv6' | 'tls13' | 'http2' | 'hsts' | 'csp' | 'security_txt' | 'certificate_verified'
export type GameInsightMetricKey = 'free' | 'windows' | 'mac' | 'linux'
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
export type GameInsightRegion = 'CN' | 'US' | 'HK'
export type GamePlayerRankingMetric = 'latest_observed' | 'peak_30d' | 'average_30d'

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

export interface SiteCompareCapability {
  key: SiteInsightCapabilityKey
  state: SiteInsightCapabilityState
}

export interface SiteCompareCertificate {
  target: string
  not_after: string | null
  days_to_expiry: number | null
  expiry_status: CertificateExpiryStatus | null
  verified: boolean | null
  verification_issue: CertificateVerificationIssue | null
  issuer: string | null
  observed_at: string | null
}

export interface SiteCompareItem {
  site: InsightEntityRef
  capabilities: SiteCompareCapability[]
  certificate: SiteCompareCertificate | null
}

export interface SiteCompare {
  status: 'ready' | 'insufficient_data'
  as_of: string | null
  sites: SiteCompareItem[]
}

export type CertificateExpiryStatus = 'expired' | 'expires_within_7d' | 'expires_in_8_30d' | 'later'
export type CertificateVerificationIssue = 'hostname_mismatch' | 'unknown_authority' | 'expired' | 'not_yet_valid' | 'incompatible_usage' | 'other'

export interface CertificateInsightItem {
  site: InsightEntityRef
  target: string
  not_after: string | null
  days_to_expiry: number | null
  expiry_status: CertificateExpiryStatus | null
  verified: boolean | null
  verification_issue: CertificateVerificationIssue | null
  issuer: string | null
  observed_at: string | null
}

export interface CertificateInsightOverview {
  as_of: string | null
  reference_at: string | null
  freshness_seconds: number
  population: number
  eligible: number
  verification: { known: number; verified: number; failed: number; coverage: number | null }
  quality: { not_applicable: number; stale: number; not_probed: number; probe_failed: number; unknown: number }
  expiry: {
    known: number
    coverage: number | null
    expired: number
    expires_within_7d: number
    expires_in_8_30d: number
    later: number
  }
  expiry_attention: CertificateInsightItem[]
  verification_issues: CertificateInsightItem[]
}

export interface GameInsightState {
  free: boolean | null
  windows: boolean | null
  mac: boolean | null
  linux: boolean | null
  release: string | null
  as_of: string | null
}

export interface GameInsightPlayerSummary {
  current: number | null
  peak_30d: number | null
  average_30d: number | null
  as_of: string | null
  fact_through: string | null
  eligible_from_30d: string | null
  observed_days_30d: number
  successful_samples_30d: number
  sample_coverage_30d: number | null
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

export interface GameInsightObservedLow {
  amount: number
  currency: string
  first_seen: string
  observed_since: string
  initial_amount: number
  discount_percent: number
}

export interface GameInsightRegionalPrice {
  region: GameInsightRegion
  available: boolean
  state: GameInsightPriceState | null
  currency: string | null
  initial_amount: number | null
  final_amount: number | null
  discount_percent: number | null
  observed_low: GameInsightObservedLow | null
}

export interface GameInsightRegionalPrices {
  as_of: string | null
  regions: GameInsightRegionalPrice[]
}

export interface GameInsights {
  game: InsightEntityRef
  state: GameInsightState
  players: GameInsightPlayerSummary
  price: GameInsightPrice | null
  regional_prices: GameInsightRegionalPrices
  recent_changes: InsightChange[]
}

export interface GameComparePlayers {
  current_available: boolean
  current: number | null
  observed_at: string | null
  peak_30d: number | null
  average_30d: number | null
  eligible_from_30d: string | null
  observed_days_30d: number
  successful_samples_30d: number
  sample_coverage_30d: number | null
}

export interface GameCompareLanguages {
  evidence: 'fresh' | 'stale' | 'unobserved'
  supported: string[]
  explicit_full_audio: string[]
  unknown_names: string[]
}

export interface GameCompareItem {
  game: InsightEntityRef
  state: Omit<GameInsightState, 'as_of'>
  players: GameComparePlayers
  price: GameInsightRegionalPrice
  languages: GameCompareLanguages
}

export interface GameCompare {
  status: 'ready' | 'insufficient_data'
  region: GameInsightRegion
  state_as_of: string | null
  player_snapshot_scheduled_for: string | null
  player_fact_through: string | null
  games: GameCompareItem[]
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
  region: GameInsightRegion
  requested_range: InsightRange
  available_from: string | null
  available_through: string | null
  points: GameInsightPricePoint[]
}

export interface GamePlayerRankingItem {
  rank: number
  game: InsightEntityRef
  value: number
  observed_at: string | null
  eligible_from: string | null
  observed_days: number | null
  successful_samples: number | null
  sample_coverage: number | null
}

export interface GamePlayerRanking {
  metric: GamePlayerRankingMetric
  basis: 'scheduled_snapshot' | 'finalized_daily_facts'
  snapshot_scheduled_for: string | null
  latest_slot_scheduled_for: string | null
  observed_from: string | null
  observed_through: string | null
  window_from: string | null
  window_through: string | null
  population: number
  ranked: number
  entity_coverage: number | null
  items: GamePlayerRankingItem[]
}

export interface GamePriceOverview {
  region: GameInsightRegion
  as_of: string | null
  population: number
  priced: number
  free: number
  unpriced: number
  unknown: number
  unavailable: number
  known: number
  coverage: number | null
  discounted: number
  discounted_share: number | null
}

export interface GameDiscountItem {
  game: InsightEntityRef
  currency: string
  initial_amount: number
  final_amount: number
  discount_percent: number
  observed_low: GameInsightObservedLow | null
}
export interface GameDiscounts { region: GameInsightRegion; as_of: string | null; items: GameDiscountItem[] }

export interface GameLanguageItem {
  code: string
  steam_name: string
  supported_games: number
  share: number | null
  explicit_full_audio_games: number
  explicit_full_audio_share: number | null
}
export interface GameLanguageOverview {
  as_of: string | null
  freshness_seconds: number
  population: number
  fresh: number
  stale: number
  unobserved: number
  coverage: number | null
  fully_normalized_games: number
  unmapped_games: number
  unmapped_entries: number
  normalization_coverage: number | null
  items: GameLanguageItem[]
}
