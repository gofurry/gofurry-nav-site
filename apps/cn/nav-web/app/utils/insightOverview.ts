import type { InsightEntityRef, InsightFeedItem, InsightOverview } from '@/types/insights'
import { insightChangeOrder } from './insightChanges'
import { formatInsightRatio } from './insightDimensions'

export const overviewSiteMetricKeys = ['tls13', 'ipv6', 'security_txt'] as const
export const overviewExploreGroups = {
  site: [
    { path: '/insights/sites', key: 'sites' },
    { path: '/insights/sites/certificates', key: 'certificates' },
    { path: '/insights/sites/compare', key: 'siteCompare' },
  ],
  game: [
    { path: '/insights/games', key: 'games' },
    { path: '/insights/games/players', key: 'players' },
    { path: '/insights/games/prices', key: 'prices' },
    { path: '/insights/games/languages', key: 'languages' },
    { path: '/insights/games/compare', key: 'gameCompare' },
  ],
} as const
export const overviewChangesPath = '/insights/changes'

export function overviewActivity(nav: InsightOverview | null, game: InsightOverview | null): InsightFeedItem[] {
  return [
    ...(nav?.recent_changes ?? []).map(item => ({ ...item, domain: 'site' as const })),
    ...(game?.recent_changes ?? []).map(item => ({ ...item, domain: 'game' as const })),
  ].sort((a, b) => insightChangeOrder(b) - insightChangeOrder(a)).slice(0, 5)
}

export function overviewSiteIdentities(nav: InsightOverview | null): InsightEntityRef[] {
  const entities = new Map<number, InsightEntityRef>()
  for (const item of nav?.recent_changes ?? []) {
    if (!entities.has(item.entity.id)) entities.set(item.entity.id, item.entity)
  }
  return [...entities.values()].slice(0, 5)
}

export function overviewGeneratedAt(nav: InsightOverview | null, game: InsightOverview | null) {
  const times = [nav?.generated_at, game?.generated_at]
    .map(value => value ? Date.parse(value) : NaN).filter(Number.isFinite)
  return times.length ? new Date(Math.min(...times)).toISOString() : null
}

export function formatOverviewSnapshot(value: string | null, locale: string) {
  if (!value) return '—'
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-US' : 'zh-CN', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', timeZone: 'UTC',
  }).format(new Date(value)) + ' UTC'
}

export function overviewSignal(value: number | null | undefined) {
  return typeof value === 'number' && Number.isFinite(value) ? Math.min(1, Math.max(0, value)) : null
}

export function formatOverviewDelta(value: number | null | undefined) {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${value > 0 ? '+' : ''}${formatInsightRatio(value)}`
}
