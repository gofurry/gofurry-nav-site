import type { GameV2ListItem, GameV2PanelRecord } from '@/types/game'
import type { InsightFeedItem, InsightMetric, InsightMetricKey, InsightOverview } from '@/types/insights'
import { selectGamePulse } from './insightGamePulse'

export const overviewHeroSiteKeys = ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'] as const
export const overviewHeroGameKeys = ['free', 'windows', 'mac', 'linux'] as const

export function overviewRatio(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 1 ? value : null
}

export function overviewFactDate(value: unknown): string | null {
  if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null
  const time = Date.parse(value)
  return Number.isFinite(time) && new Date(time).toISOString().slice(0, 10) === value ? value : null
}

export function selectOverviewMetric(overview: InsightOverview | null, keys: readonly InsightMetricKey[]) {
  for (const key of keys) {
    const metric = overview?.metrics?.find(item => item.key === key && overviewRatio(item.value) !== null)
    if (metric) return { ...metric, coverage: overviewRatio(metric.coverage), as_of: overviewFactDate(metric.as_of) }
  }
  return null
}

export function overviewEntityCount(value: unknown): number | null {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0 ? value : null
}

export function selectOverviewEcosystemMetrics(overview: InsightOverview | null, domain: 'game' | 'site') {
  const keys = domain === 'game' ? overviewHeroGameKeys : overviewHeroSiteKeys
  // Use the Hero's exact selector, including valid zero and fallback priority.
  // If its metric is the only reliable Site metric, leave this list empty.
  const heroKey = domain === 'site' ? selectOverviewMetric(overview, overviewHeroSiteKeys)?.key : null
  return keys.filter(key => key !== heroKey).flatMap(key => {
    const metric = selectOverviewMetric(overview, [key])
    return metric ? [metric] : []
  }).slice(0, 3)
}

export function overviewSample(metric: Pick<InsightMetric, 'known' | 'eligible'>): string | null {
  return Number.isSafeInteger(metric.known) && Number.isSafeInteger(metric.eligible)
    && metric.known >= 0 && metric.eligible >= metric.known ? `${metric.known} / ${metric.eligible}` : null
}

// Overview-only: the API supplies a ratio difference, never relative growth.
export function formatOverviewPercentagePoints(value: unknown, locale: string) {
  if (typeof value !== 'number' || !Number.isFinite(value) || Math.abs(value) > 1) return null
  const rounded = Math.round(value * 1000) / 10
  const number = new Intl.NumberFormat(locale === 'en' ? 'en-US' : 'zh-CN', { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(Object.is(rounded, -0) ? 0 : rounded)
  return `${rounded > 0 ? '+' : ''}${number}`
}

// The request owner serializes now into the SSR payload. A freshness window is
// optional evidence policy, not an assumption about the collector's cadence.
export function overviewPlayerObservation(game: GameV2ListItem, now: number, maxAgeMs: number | null = null) {
  const observation = game.online_count
  if (observation?.status !== 'success' || !Number.isFinite(observation.count) || observation.count < 0) return null
  const value = observation.collected_at
  if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(value) || !overviewFactDate(value.slice(0, 10))) return null
  const time = Date.parse(value)
  if (!Number.isFinite(time) || !Number.isFinite(now) || time > now) return null
  const freshness = maxAgeMs !== null && Number.isFinite(maxAgeMs) && maxAgeMs >= 0
    ? (now - time <= maxAgeMs ? 'fresh' : 'stale') : 'unverified'
  const collectedAt = new Date(time).toISOString()
  return { count: observation.count, collectedAt, label: collectedAt.slice(0, 19).replace('T', ' ') + ' UTC', freshness }
}

export function auditOverviewGameCover(game: GameV2ListItem, now: number, maxAgeMs: number | null = null) {
  // Neither current Overview nor Home DTO carries positive public-cover approval.
  // Populated tags, an absent adult tag, or an injected `sfw` field cannot grant it.
  return {
    gameId: game.id,
    safety: game.tags?.some(tag => tag.code === 'adult') ? 'adult' : 'unverified',
    hasImage: [game.header_url, game.capsule_url].some(value => typeof value === 'string' && /^https?:\/\/\S+$/.test(value)),
    observation: overviewPlayerObservation(game, now, maxAgeMs),
    coverEligible: false as const,
  }
}

export function selectOverviewPulse(panel: GameV2PanelRecord | null, now: number) {
  return selectGamePulse(panel ? {
    ...panel,
    top_online: (panel.top_online ?? []).filter(game => auditOverviewGameCover(game, now).observation !== null),
  } : null)
}

// Keep shared Domain media behavior unchanged; only Overview rejects game art.
export function overviewActivityWithoutGameArt(item: InsightFeedItem): InsightFeedItem {
  return item.domain === 'game' ? { ...item, entity: { id: item.entity.id, name: item.entity.name } } : item
}
