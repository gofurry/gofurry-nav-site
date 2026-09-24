import { runtimeTest, type Reply } from './insights-runtime'
import { mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'

export interface SEOState { failure: 'site' | 'game' | 'sitemap' | ''; siteInsightsFailure: boolean; gameInsightsFailure: boolean }
export const seoState = (): SEOState => ({ failure: '', siteInsightsFailure: false, gameInsightsFailure: false })
export const allowedSEO = (url: URL) => [
  '/api/v2/nav/home', '/api/v2/nav/sites/index', '/api/v2/nav/site-groups', '/api/v2/game/list',
  '/api/v2/game/info', '/api/v2/game/home', '/api/v2/game/reviews', '/api/v2/game/recommend/similar',
].includes(url.pathname) || /^\/api\/v2\/nav\/sites\/(41|42|999999999)\/(detail|insights|view)$/.test(url.pathname)
  || /^\/api\/v2\/game\/games\/(82|83|999999999)\/(insights(?:\/(players|prices))?|view|daily)$/.test(url.pathname)
  || url.pathname === '/api/v2/nav/site-groups/12/sites'

export function seoResponse(url: URL, media: string, state: SEOState): Reply {
  const path = url.pathname
  if (path === '/api/v2/nav/home') return { data: { schema_version: 4, groups: [],
    spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
    saying: { saying: 'Fixture', author: 'Fixture' }, ping: {}, hero: { desktop: null, mobile: null } } }
  if (path === '/api/v2/nav/sites/index') return state.failure === 'sitemap' ? { status: 503 }
    : { data: { items: [{ id: 41, domains: ['target.example', 'alt.example'] }] } }
  if (path === '/api/v2/nav/site-groups') return { data: [{ id: '12', name: 'Fixture community' }] }
  if (path === '/api/v2/game/list') return { data: [{ game_id: '82' }] }
  if (path === '/api/v2/nav/site-groups/12/sites') return { data: { group: { id: '12', name: 'Fixture community', info: 'Isolated group summary' },
    schema_version: 1, generated_at: '2026-08-30T12:00:00Z', state: 'ready',
    items: [], page: 1, page_size: 24, total: 0, has_more: false } }
  const siteID = path.match(/\/sites\/(\d+)\//)?.[1]
  if (siteID && path.endsWith('/detail')) {
    if (state.failure === 'site') return { status: 503 }
    if (siteID === '999999999' || (url.searchParams.has('target') && !['target.example', 'alt.example'].includes(url.searchParams.get('target')!))) return { status: 404 }
    return { data: { site: { id: Number(siteID), name: 'Site fixture ' + siteID, info: 'Independent Site detail.',
      icon: '', country: 'CN', nsfw: '0', welfare: '0', view_count: 1 },
    selected_target: url.searchParams.get('target') || 'target.example', latest_core: null,
    site_summary: { targets: [{ target: 'target.example' }, { target: 'alt.example' }] },
    target_summary: null, light_probe_state: null } }
  }
  if (siteID && path.endsWith('/insights')) return state.siteInsightsFailure || siteID === '42' ? { status: 503 } : { data: {
    site: { id: Number(siteID), name: 'Site fixture ' + siteID },
    capabilities: [
      { key: 'ipv6', as_of: '2026-08-30', state: 'unknown', ecosystem: { value: .5, coverage: .8 } },
      { key: 'tls13', as_of: '2026-08-30', state: 'unavailable', ecosystem: { value: .6, coverage: .8 } },
      { key: 'security_txt', as_of: '2026-08-30', state: 'unsupported', ecosystem: { value: .2, coverage: .8 } },
    ],
    recent_changes: [{ type: 'site.ipv6.enabled', date: '2026-08-30', occurred_at: null, entity: { id: Number(siteID), name: 'Site fixture ' + siteID }, detail: null }],
  } }
  if (path.endsWith('/view')) return { data: { site_id: Number(siteID), game_id: 82, view_count: 2 } }
  if (path === '/api/v2/game/home') return { data: mockGameHome(media) }
  if (path === '/api/v2/game/info') {
    if (state.failure === 'game') return { status: 503 }
    const id = url.searchParams.get('id')
    if (!['82', '83'].includes(id!)) return { status: 404 }
    return { data: { id, appid: Number(id), name: 'Game fixture ' + id, summary: 'Independent game detail.',
      about_the_game: '<p>SSR introduction.</p>', site: { view_count: 1, resources: [], groups: [], links: [] },
      platforms: { windows: true }, prices: [], news: [], tags: [], developers: [], publishers: [],
      media: { screenshots: [], movies: [], assets: [] }, requirements: {}, support_info: {}, extra: {} } }
  }
  if (/\/games\/\d+\/insights$/.test(path)) return state.gameInsightsFailure || path.includes('/83/') ? { status: 503 } : { data: {
    game: { id: 82, name: 'Game fixture 82' }, state: { free: false, windows: true, mac: true, linux: null, release: 'available', as_of: '2026-08-30' },
    players: { current: 0, peak_30d: 3, average_30d: 1.5, as_of: '2026-08-30', observed_days_30d: 28, sample_coverage_30d: .93 },
    price: null, regional_prices: { as_of: null, regions: [] }, recent_changes: [],
  } }
  if (/\/insights\/(players|prices)$/.test(path)) return { data: { region: 'CN', available_from: null, available_through: null, points: [] } }
  if (path.endsWith('/reviews')) return { data: { total: 0, remarks: [] } }
  if (path.endsWith('/daily') || path.endsWith('/recommend/similar')) return { data: [] }
  throw new Error('Unmapped SEO response ' + path)
}
export const test = runtimeTest(seoState, allowedSEO, (url, media, _body, state) => seoResponse(url, media, state), {
  NUXT_PUBLIC_SITE_URL: 'https://go-furry.com',
  NUXT_PUBLIC_I18N_BASE_URL: 'https://go-furry.com',
})
export { expect } from './insights-runtime'
