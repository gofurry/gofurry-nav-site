export type InsightsNavigationDomain = 'site' | 'game'

export const insightsPrimaryItems = [
  { label: 'insights.nav.overview', path: '/insights' },
  { label: 'insights.nav.sites', path: '/insights/sites' },
  { label: 'insights.nav.games', path: '/insights/games' },
  { label: 'insights.nav.changes', path: '/insights/changes' },
] as const

export const insightsDomainItems = {
  site: [
    { path: '/insights/sites', label: 'insights.siteIntelligence.ecosystem' },
    { path: '/insights/sites/certificates', label: 'insights.siteIntelligence.certificates' },
    { path: '/insights/sites/compare', label: 'insights.siteIntelligence.compare' },
  ],
  game: [
    { path: '/insights/games', label: 'insights.gameIntelligence.ecosystem' },
    { path: '/insights/games/players', label: 'insights.gameIntelligence.players' },
    { path: '/insights/games/prices', label: 'insights.gameIntelligence.prices' },
    { path: '/insights/games/languages', label: 'insights.gameIntelligence.languages' },
    { path: '/insights/games/compare', label: 'insights.gameIntelligence.compare' },
  ],
} as const

export function isInsightsPrimaryActive(routePath: string, path: string) {
  const normalized = normalizePath(routePath)
  return normalized === path || (path !== '/insights' && normalized.startsWith(`${path}/`))
}

export function isInsightsDomainActive(routePath: string, path: string) {
  return normalizePath(routePath) === path
}

export function revealInsightsNavigationLink(event: FocusEvent) {
  if (event.currentTarget instanceof HTMLElement) {
    event.currentTarget.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'instant' })
  }
}

function normalizePath(path: string) {
  return path.replace(/^\/en(?=\/|$)/, '') || '/'
}
