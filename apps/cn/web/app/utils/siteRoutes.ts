import { buildSiteDetailQuery, parseSiteDetailRouteState, selectSiteDetailTarget } from './siteDetailRouteState.ts'

export function siteEntityPath(siteId: string | number) {
  const id = encodeURIComponent(String(siteId))
  return `/site/${id}`
}

export function siteTargetPath(siteId: string | number, domain: string, query: Readonly<Record<string, unknown>> = {}) {
  const entityPath = siteEntityPath(siteId)
  const state = selectSiteDetailTarget(parseSiteDetailRouteState(query), domain)
  const search = Object.entries(buildSiteDetailQuery(state))
    .map(([key, value]) => `${key}=${encodeURIComponent(value)}`).join('&')
  return search ? `${entityPath}?${search}` : entityPath
}
