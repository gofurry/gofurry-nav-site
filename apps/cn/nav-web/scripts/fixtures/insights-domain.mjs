import { mockOverview, mockGameHome } from './insights-overview.mjs'

export const domainMetricKeys = { site: ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'], game: ['free', 'windows', 'mac', 'linux'] }
export const domainDimensionKeys = { site: ['country', 'group', 'nsfw', 'public_interest'], game: ['primary_tag', 'tag'] }

export async function domainFixtureResponse(url, mediaBase, state) {
  const path = url.pathname
  if (path === '/api/v2/game/home') return state.panelFailure ? { status: 503 } : { data: mockGameHome(mediaBase) }
  const domain = path.startsWith('/api/v2/nav/') ? 'site' : 'game'
  if (path.endsWith('/insights/overview')) {
    if (state.overviewFailure) return { status: 503 }
    return { data: { ...mockOverview(domain, mediaBase), generated_at: '2026-09-09T04:00:00Z', metrics: domainMetricKeys[domain].map((key, i) => ({
      key, value: [0.468, 0.824, 0.912, 0.713, 0.216, null, 0][i], delta_30d: i === 5 ? null : i === 6 ? 0 : 0.042,
      coverage: 0.962, known: domain === 'site' ? 229 : 205, eligible: domain === 'site' ? 238 : 213, as_of: '2026-09-08', available_from: '2026-06-01',
    })) } }
  }
  const match = path.match(/\/insights\/metrics\/([^/]+)\/(trend|breakdown)(?:\/([^/]+)\/([^/]+)\/trend)?$/)
  if (!match) return { status: 503 }
  const [, key, kind, sliceDimension, slice] = match
  const dimension = sliceDimension || url.searchParams.get('dimension') || domainDimensionKeys[domain][0]
  const range = url.searchParams.get('range') || '30d'
  if (kind === 'breakdown' && !slice) {
    if (state.breakdownFailure) return { status: 503 }
    const values = dimension === 'country' ? ['CN', 'US', 'JP', 'DE', 'GB', 'FR', 'CA', 'AU', 'BR', 'NZ']
      : dimension === 'nsfw' ? ['sfw', 'nsfw', 'unknown']
        : dimension === 'public_interest' ? ['public_interest', 'standard', 'unknown']
          : Array.from({ length: 10 }, (_, i) => String(i + 31))
    return { data: { key, dimension, slice_mode: ['tag', 'group'].includes(dimension) ? 'overlapping' : 'partition', as_of: '2026-09-08', items: values.map((value, i) => {
      const population = [46, 39, 32, 26, 22, 18, 15, 12, 8, 5][i]
      const known = Math.floor(population * 0.9)
      return {
      value, label: ['视觉小说', '冒险', '独立', '角色扮演'][i] || `分组 ${i + 1}`, label_en: ['Visual novel', 'Adventure', 'Indie', 'Role-playing'][i] || `Group ${i + 1}`,
      population, eligible: population, known, metric_value: [0.824, null, 0, 0.63, 0.7, 0.35, 0.95, 0.4, 0.55, 0.5][i], coverage: known / population,
    } }) } }
  }
  if (state.trendMode === 'fail') return { status: 503 }
  if (state.trendMode === 'delayed') await new Promise(resolve => setTimeout(resolve, 350))
  const count = state.trendMode === 'empty' ? 0 : state.trendMode === 'one' ? 1 : range === '30d' ? 30 : 90
  const points = Array.from({ length: count }, (_, i) => {
    const date = new Date(Date.UTC(2026, 8, 8 - count + i + 1)).toISOString().slice(0, 10)
    const value = i === 10 ? null : i === 0 ? 0 : Math.min(0.9, 0.45 + i * 0.004)
    return slice ? { date, metric_value: value, population: 46, eligible: 44, known: 40, coverage: 0.91 } : { date, value, coverage: 0.91 }
  })
  return { data: { key, dimension, requested_range: range, available_from: points[0]?.date ?? null, available_through: points.at(-1)?.date ?? null, points,
    ...(slice ? { slice: { value: slice, label: '选中分组', label_en: 'Selected group' }, slice_mode: 'overlapping' } : {}),
  } }
}
