import type {
  NavHomeBackgrounds,
  NavHomePingResponse,
  NavHomeResponse,
  NavHomeSayingResponse,
  NavSearchSuggestionEngine,
  NavSearchSuggestionsResponse,
  NavSiteGroupPageResponse,
  NavSiteIndexResponse,
  NavUpdatesResponse,
  SayingModel,
  Site,
  SiteViewResponse,
} from '~/types/nav'
import type {
  InsightChangeExplorerPage,
  InsightChangeRange,
  InsightDimensionBreakdown,
  InsightDimensionTrend,
  InsightMetricTrend,
  InsightOverview,
  InsightRange,
  NavInsightMetricKey,
  SiteInsightChangeCategory,
  SiteInsightDimension,
  SiteInsights,
} from '~/types/insights'

export function getNavHome(lang: string): Promise<NavHomeResponse> {
  return useApi('navV2')('/nav/home', { query: { lang } })
}

export function getNavHomePing(): Promise<NavHomePingResponse> {
  return useApi('navV2')('/nav/home/ping')
}

export function getNavHomeSaying(lang: string): Promise<NavHomeSayingResponse> {
  return useApi('navV2')('/nav/home/saying', { query: { lang } })
}

export function getNavHomeBackgrounds(): Promise<NavHomeResponse['backgrounds']> {
  return useApi('navV2')<{ backgrounds: NavHomeBackgrounds }>('/nav/home/backgrounds').then((response) => response.backgrounds)
}

export function getNavSiteIndex(): Promise<NavSiteIndexResponse> {
  return useApi('navV2')('/nav/sites/index')
}

export function getNavSiteDirectory(lang: string): Promise<Site[]> {
  return useApi('navV2')('/nav/sites/directory', { query: { lang } })
}

export function getNavSiteGroupPage(siteGroupId: string, lang: string, page = 1, pageSize = 24): Promise<NavSiteGroupPageResponse> {
  return useApi('navV2')(`/nav/site-groups/${siteGroupId}/sites`, { query: { lang, page, page_size: pageSize } })
}

export function touchSiteView(siteId: string | number): Promise<SiteViewResponse> {
  return useApi('navV2')(`/nav/sites/${siteId}/view`, { method: 'POST' })
}

export function getSearchSuggestion(
  engine: NavSearchSuggestionEngine,
  keyword: string,
  signal?: AbortSignal
): Promise<NavSearchSuggestionsResponse> {
  return useApi('navV2')('/nav/search/suggestions', { query: { engine, q: keyword }, signal })
}

export function getNavUpdates(lang: 'zh' | 'en'): Promise<NavUpdatesResponse> {
  return useApi('navV2')('/nav/updates', { query: { lang } })
}

export function getNavInsightsOverview(): Promise<InsightOverview> {
  return useApi('navV2')('/nav/insights/overview')
}

export function getNavInsightsTrend(metric: NavInsightMetricKey, range: InsightRange): Promise<InsightMetricTrend> {
  return useApi('navV2')(`/nav/insights/metrics/${metric}/trend`, { query: { range } })
}

export function getNavInsightsBreakdown(metric: NavInsightMetricKey, dimension: SiteInsightDimension): Promise<InsightDimensionBreakdown> {
  return useApi('navV2')(`/nav/insights/metrics/${metric}/breakdown`, { query: { dimension } })
}

export function getNavInsightsSliceTrend(metric: NavInsightMetricKey, dimension: SiteInsightDimension, value: string, range: InsightRange): Promise<InsightDimensionTrend> {
  return useApi('navV2')(`/nav/insights/metrics/${metric}/breakdown/${dimension}/${encodeURIComponent(value)}/trend`, { query: { range } })
}

export function getNavInsightChanges(query: {
  range: InsightChangeRange
  category?: SiteInsightChangeCategory | ''
  type?: string
  cursor?: string
  limit?: number
}): Promise<InsightChangeExplorerPage> {
  return useApi('navV2')('/nav/insights/changes', { query })
}

export function getSiteInsights(siteId: string | number): Promise<SiteInsights> {
  return useApi('navV2')(`/nav/sites/${encodeURIComponent(String(siteId))}/insights`)
}
