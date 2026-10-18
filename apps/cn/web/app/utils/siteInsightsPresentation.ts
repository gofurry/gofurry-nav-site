import type { SiteInsights, SiteInsightCapabilityKey, InsightTrendPoint } from '~/types/insights'
import { siteCapabilityRegistry } from './siteCapabilityRegistry'
import { formatInsightRatio } from './insightDimensions'
import { formatInsightChangeWhen, insightChangeI18nKey, insightChangeOrder, siteInsightChangeCategory } from './insightChanges'

export type SiteInsightsState = 'ready' | 'empty' | 'unavailable'
type Translate = (key: string) => string
export function siteInsightsState(data: SiteInsights | null): SiteInsightsState {
  return !data ? 'unavailable' : data.capabilities.length || data.recent_changes.length ? 'ready' : 'empty'
}

/** Site facts and ecosystem context only. There is deliberately no Target input. */
export function presentSiteInsights(data: SiteInsights | null, state: SiteInsightsState, metric: SiteInsightCapabilityKey, t: Translate, locale: string) {
  const unavailable = state === 'unavailable'
  const facts = new Map((data?.capabilities ?? []).map(item => [item.key, item]))
  const categories = [...new Set(siteCapabilityRegistry.map(item => item.category))]
  const rows = categories.flatMap(category => siteCapabilityRegistry.filter(item => item.category === category).map(item => {
    const fact = unavailable ? undefined : facts.get(item.key)
    // null means the slice failed, not a fabricated backend fact state.
    const factState = unavailable ? null : fact?.state ?? 'missing'
    return { key: item.key, category, categoryLabel: t('siteOverview.categories.' + category), label: t(item.labelKey),
      state: factState, stateLabel: factState === null ? '—' : t(factState === 'missing' ? 'siteOverview.capabilityMissing' : 'insights.entity.states.' + factState),
      tone: factState === 'supported' ? 'good' : factState === 'stale' || factState === 'unavailable' ? 'warning' : factState === 'unsupported' ? 'neutral' : 'muted',
      adoption: formatInsightRatio(fact?.ecosystem.value ?? null), coverage: formatInsightRatio(fact?.ecosystem.coverage ?? null),
      date: fact?.as_of || '—', shortDate: fact?.as_of?.slice(5) || '—', dateTime: fact?.as_of || undefined }
  }))
  const changes = unavailable ? [] : [...(data?.recent_changes ?? [])].sort((a, b) => insightChangeOrder(b) - insightChangeOrder(a)).map((item, index) => {
    const category = siteInsightChangeCategory(item.type)
    return { key: `${item.type}:${item.occurred_at || item.date}:${index}`, category: category ?? 'unknown',
      categoryLabel: t('siteIntelligence.changeCategories.' + (category ?? 'unknown')), label: t(insightChangeI18nKey(item.type)),
      dateTime: item.occurred_at || item.date, when: formatInsightChangeWhen(item, locale, 'UTC') + (item.occurred_at ? ' UTC' : ''), precise: Boolean(item.occurred_at) }
  })
  return { state, rows, groups: categories.map(category => ({ key: category, label: t('siteOverview.categories.' + category), rows: rows.filter(row => row.category === category) })),
    selected: rows.find(item => item.key === metric)!, changes,
    changesState: unavailable ? 'unavailable' as const : changes.length ? 'ready' as const : 'empty' as const }
}

export const siteInsightAdoptionValues = (points: InsightTrendPoint[]) => points.map(point =>
  point.value === null || !Number.isFinite(point.value) ? null : Number((point.value * 100).toFixed(4)))
export type SiteInsightsPresentation = ReturnType<typeof presentSiteInsights>
