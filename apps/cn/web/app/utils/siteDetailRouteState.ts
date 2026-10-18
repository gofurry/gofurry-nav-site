import { siteCapabilityKeys } from './siteCapabilityRegistry.ts'

export const siteDetailTabs = ['overview', 'observation', 'security', 'insights'] as const
export const siteObservationViews = ['overview', 'performance', 'http', 'dns', 'web'] as const
export const siteSecurityViews = ['overview', 'tls', 'web', 'exposure'] as const
export const siteInsightRanges = ['30d', '90d', 'all'] as const
export type SiteDetailTab = typeof siteDetailTabs[number]
export type SiteObservationView = typeof siteObservationViews[number]
export type SiteSecurityView = typeof siteSecurityViews[number]
export type SiteInsightMetric = typeof siteCapabilityKeys[number]
export type SiteInsightRange = typeof siteInsightRanges[number]
type SiteDetailQuery = Readonly<Record<string, unknown>>
type DomainState = { domain: string }
export type SiteDetailRouteState = DomainState & (
  | { tab: 'overview' }
  | { tab: 'observation'; view: typeof siteObservationViews[number] }
  | { tab: 'security'; view: typeof siteSecurityViews[number] }
  | { tab: 'insights'; metric: typeof siteCapabilityKeys[number]; range: typeof siteInsightRanges[number] }
)

// Vue Router has already decoded query strings. Repeated keys use the first
// value, including null/empty; never decode again or validate business ownership.
function queryValue(value: unknown): string {
  const first = Array.isArray(value) ? value[0] : value
  return typeof first === 'string' ? first.trim() : ''
}

function member<T extends string>(value: unknown, values: readonly T[], fallback: T): T {
  const candidate = queryValue(value)
  return values.find(item => item === candidate) ?? fallback
}

export function parseSiteDetailRouteState(query: SiteDetailQuery): SiteDetailRouteState {
  const domain = queryValue(query.domain)
  const requestedTab = queryValue(query.tab)
  const tab = member(query.tab, siteDetailTabs, 'observation')
  if (tab === 'observation') return { domain, tab, view: member(requestedTab && requestedTab !== 'observation' ? undefined : query.view, siteObservationViews, 'performance') }
  if (tab === 'security') return { domain, tab, view: member(query.view, siteSecurityViews, 'overview') }
  if (tab === 'insights') return {
    domain, tab,
    metric: member(query.metric, siteCapabilityKeys, 'ipv6'),
    range: member(query.range, siteInsightRanges, '30d'),
  }
  return { domain, tab }
}

// Builders always pass through the same normalizer. Defaults and foreign-tab
// state are omitted without rewriting the incoming URL during SSR/hydration.
export function buildSiteDetailQuery(state: SiteDetailRouteState): Record<string, string> {
  const normalized = parseSiteDetailRouteState(state)
  const query: Record<string, string> = {}
  if (normalized.domain) query.domain = normalized.domain
  if (normalized.tab === 'observation') {
    if (normalized.view !== 'performance') {
      query.tab = normalized.tab
      query.view = normalized.view
    }
  } else {
    query.tab = normalized.tab
    if (normalized.tab === 'security' && normalized.view !== 'overview') query.view = normalized.view
  }
  if (normalized.tab === 'insights') {
    if (normalized.metric !== 'ipv6') query.metric = normalized.metric
    if (normalized.range !== '30d') query.range = normalized.range
  }
  return query
}

export function selectSiteDetailTarget(state: SiteDetailRouteState, domain: string): SiteDetailRouteState {
  return parseSiteDetailRouteState({ ...state, domain })
}

export function selectSiteDetailTab(state: SiteDetailRouteState, tab: SiteDetailTab): SiteDetailRouteState {
  // Even a shared view name (e.g. web) belongs to its workspace, not the next tab.
  return parseSiteDetailRouteState(state.tab === tab ? state : { domain: state.domain, tab })
}

export function selectSiteObservationView(state: SiteDetailRouteState, view: SiteObservationView): SiteDetailRouteState {
  return parseSiteDetailRouteState({ domain: state.domain, tab: 'observation', view })
}

export function selectSiteSecurityView(state: SiteDetailRouteState, view: SiteSecurityView): SiteDetailRouteState {
  return parseSiteDetailRouteState({ domain: state.domain, tab: 'security', view })
}

export function selectSiteInsightMetric(state: SiteDetailRouteState, metric: SiteInsightMetric): SiteDetailRouteState {
  return parseSiteDetailRouteState({ ...selectSiteDetailTab(state, 'insights'), metric })
}

export function selectSiteInsightRange(state: SiteDetailRouteState, range: SiteInsightRange): SiteDetailRouteState {
  return parseSiteDetailRouteState({ ...selectSiteDetailTab(state, 'insights'), range })
}
