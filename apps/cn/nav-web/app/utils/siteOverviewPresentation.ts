import type { HealthStatus, SiteHealthSummary, TargetHealthSummaryItem } from '~/types/nav'
import type { SiteInsights, SiteInsightCapabilityState } from '~/types/insights'
import { siteCapabilityRegistry } from './siteCapabilityRegistry'
import { formatInsightChangeWhen, insightChangeI18nKey, insightChangeOrder } from './insightChanges'

type Tone = 'good' | 'neutral' | 'muted' | 'warning' | 'bad'
type Translate = (key: string, values?: Record<string, string | number>) => string
type CapabilityState = SiteInsightCapabilityState | 'missing'
const statuses: HealthStatus[] = ['healthy', 'warning', 'degraded', 'down', 'unknown']
const count = (value: unknown) => typeof value === 'number' && Number.isFinite(value) && value >= 0 ? Math.floor(value) : null
const clean = (values?: string[]) => [...new Set((values ?? []).map(value => value.trim()).filter(Boolean))]
const normalized = (value: string) => value.trim().replace(/\s+/g, ' ').toLowerCase()
const timestamp = (value?: string) => value && value !== '0001-01-01T00:00:00Z' && Number.isFinite(Date.parse(value))
  ? new Date(value).toISOString() : null
const healthTone = (status: HealthStatus | null): Tone => status === 'healthy' ? 'good' : status === 'down' ? 'bad'
  : status === 'warning' || status === 'degraded' ? 'warning' : 'muted'
const capabilityTone = (state: CapabilityState): Tone => state === 'supported' ? 'good'
  : state === 'stale' || state === 'unavailable' ? 'warning' : state === 'unsupported' ? 'neutral' : 'muted'

function needsAttention(target: TargetHealthSummaryItem) {
  if (['warning', 'degraded', 'down'].includes(target.status)) return true
  if (clean(target.reason_codes).some(code => /(^|_)stale($|_)/.test(code))) return true
  return target.status === 'unknown' && Boolean(timestamp(target.observed_at)
    || clean(target.reason_messages).length || clean(target.reason_codes).length)
}

/** Site-only projection. It has no Current Target input and performs no I/O. */
export function presentSiteOverview(
  summary: SiteHealthSummary | null, insights: SiteInsights | null, insightsUnavailable: boolean,
  t: Translate, locale: string,
) {
  const summaryState = summary?.state === 'ready' || summary?.state === 'stale' ? summary.state : 'missing'
  const status = summaryState === 'missing' ? null : statuses.find(status => status === summary?.status) ?? 'unknown'
  const statusLabel = status === null ? t('siteOverview.summaryMissing') : t(`siteDetail.states.${status}`)
  const targetCount = count(summary?.target_count)
  const statusItems = statuses.flatMap(status => {
    const value = count(summary?.status_counts?.[status]) ?? 0
    return value ? [{ status, count: value, label: t(`siteDetail.states.${status}`), tone: healthTone(status) }] : []
  })
  const allHealthy = targetCount !== null && targetCount > 0 && statusItems.length === 1
    && statusItems[0]?.status === 'healthy' && statusItems[0].count === targetCount
  const summaryText = summaryState === 'missing' ? '' : targetCount === 0 ? t('siteOverview.noTargets')
    : allHealthy ? t('siteOverview.healthyTargets', { healthy: targetCount, total: targetCount })
      : statusItems.length ? statusItems.map(item => `${item.count} ${item.label}`).join(' · ')
        : targetCount === null ? t('siteOverview.countUnavailable') : t('siteOverview.targetCount', { count: targetCount })
  const generatedAt = timestamp(summary?.generated_at)
  const health = { summaryState, status, statusLabel, tone: healthTone(status), targetCount, statusItems, summaryText,
    freshnessLabel: summaryState === 'stale' ? t('siteOverview.summaryStale') : '', generatedAt,
    generatedLabel: generatedAt ? generatedAt.slice(0, 19).replace('T', ' ') + ' UTC' : '—' }

  const attention: { key: string; message: string; target?: string; tone: Tone }[] = []
  const seen = new Set<string>()
  const add = (message: string, target?: string) => {
    const key = normalized(message)
    if (!key || seen.has(key)) return
    seen.add(key)
    attention.push({ key, message, target, tone: 'warning' })
  }
  const messages = clean(summary?.reason_messages)
  const codes = clean(summary?.reason_codes)
  const codeLabel = (code: string) => code === 'summary_stale' ? t('siteOverview.staleAttention')
    : code === 'summary_missing' ? t('siteOverview.missingAttention') : t('siteOverview.reasonCode', { code })
  const attentionTargets = (summary?.targets ?? []).filter(needsAttention)
  const hasHumanMessages = messages.length > 0 || attentionTargets.some(target => clean(target.reason_messages).length > 0)
  for (const message of messages.length ? messages : hasHumanMessages ? [] : codes.map(codeLabel)) add(message)
  for (const target of attentionTargets) {
    const targetMessages = clean(target.reason_messages)
    // Human messages that name this target or repeat its reason already explain it.
    const covered = messages.some(message => normalized(message).includes(normalized(target.target))
      || targetMessages.some(reason => normalized(reason) === normalized(message)))
    if (covered) continue
    const reasons = targetMessages.length ? targetMessages : hasHumanMessages ? [] : clean(target.reason_codes).map(codeLabel)
    if (reasons.length) {
      for (const message of reasons) add(normalized(message).includes(normalized(target.target)) ? message : `${target.target} · ${message}`, target.target)
    } else add(`${target.target} · ${t(`siteDetail.states.${target.status}`)}`, target.target)
  }
  if (!attention.length) {
    if (summaryState === 'missing') add(t('siteOverview.missingAttention'))
    else if (summaryState === 'stale') add(t('siteOverview.staleAttention'))
    else if (status !== 'healthy') add(t('siteOverview.statusAttention', { status: statusLabel }))
  }

  const unavailable = insightsUnavailable || insights === null
  const records = new Map((insights?.capabilities ?? []).map(item => [item.key, item]))
  const capabilities = siteCapabilityRegistry.map(item => {
    const state: CapabilityState = unavailable ? 'unavailable' : records.get(item.key)?.state ?? 'missing'
    return { key: item.key, category: item.category, label: t(item.labelKey), state, tone: capabilityTone(state),
      stateLabel: t(state === 'missing' ? 'siteOverview.capabilityMissing' : `insights.entity.states.${state}`) }
  })
  const capabilityState = unavailable ? 'unavailable' : capabilities.every(item => item.state === 'missing') ? 'empty' : 'ready'
  const capabilityGroups = [...new Set(siteCapabilityRegistry.map(item => item.category))].map(category => ({
    key: category, label: t(`siteOverview.categories.${category}`), items: capabilities.filter(item => item.category === category),
  }))
  const recentChanges = unavailable ? [] : [...insights.recent_changes]
    .sort((left, right) => insightChangeOrder(right) - insightChangeOrder(left)).slice(0, 4)
    .map((item, index) => ({ key: `${item.type}:${item.occurred_at || item.date}:${index}`, type: item.type,
      label: t(insightChangeI18nKey(item.type)), dateTime: item.occurred_at || item.date,
      // An explicit zone makes SSR and hydration agree even across time zones.
      when: formatInsightChangeWhen(item, locale, 'UTC'), precise: item.occurred_at !== null }))
  const changesState = unavailable ? 'unavailable' : recentChanges.length ? 'ready' : 'empty'
  return { health, attention, capabilityState, capabilities, capabilityGroups, changesState, recentChanges }
}

export type SiteOverviewPresentation = ReturnType<typeof presentSiteOverview>
