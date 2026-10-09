import { mockOverview } from '../../../scripts/fixtures/insights-overview.mjs'
import { runtimeTest } from './insights-runtime'
import type { InsightFeaturedVisual } from '../../../app/types/insights'
import { steamSharedAssetCandidates } from '../../../app/utils/steamAssets'
export const sources = ['/api/v2/nav/insights/overview', '/api/v2/game/insights/overview']
const overviewRuntime = runtimeTest(() => ({ failure: '', siteHero: false, metricCase: '', sampleCase: '', candidateCase: '', visualCase: '', siteVisualCase: '', brokenGame: 0, eventCount: -1, countCase: '' }),
  url => sources.includes(url.pathname),
  (url, media, _body, state) => {
    const source = url.pathname === sources[0] ? 'nav' : 'game'
    if (state.failure === source || state.failure === 'all') return { status: 503 }
    const data = mockOverview(source === 'nav' ? 'site' : 'game', media)
    if (source === 'game' && state.visualCase) {
      const ids = state.visualCase === 'empty' ? [] : state.visualCase === 'one' ? [91] : state.visualCase === 'two' ? [91, 92] : [91, 92, 93]
      const featured: InsightFeaturedVisual[] = ids.map((gameId, i) => ({
        game_id: gameId, name: ['林间旅人', '星光来信', '远方的河谷'][i]!, name_en: ['Forest Traveller', 'Letters from the Stars', ''][i]!,
        visual: { kind: 'game_header', asset: `https://shared.steamstatic.com/steam/apps/${gameId}/header.jpg` },
      }))
      if (state.visualCase === 'long') featured.forEach(item => { item.name = '超长作品标题'.repeat(12); item.name_en = 'LongUnbrokenGameTitle'.repeat(8) })
      if (state.visualCase === 'invalid') featured.forEach((item, i) => { if (i === 0) item.game_id = -1; if (i === 1) item.name = item.name_en = ''; if (i === 2) item.visual.asset = 'https://example.test/unverified.jpg' })
      if (state.visualCase === 'duplicate') featured.splice(1, 0, featured[0]!)
      Object.assign(data, { featured_visuals: featured })
    }
    if (state.siteHero && source === 'nav') data.recent_changes[0].occurred_at = '2026-09-02T12:00:00Z'
    if (state.eventCount >= 0) data.recent_changes = Array.from({ length: source === 'nav' ? Math.ceil(state.eventCount / 2) : Math.floor(state.eventCount / 2) }, (_, i) => ({
      ...data.recent_changes[i % 2], entity: { ...data.recent_changes[i % 2].entity, id: (source === 'nav' ? 41 : 81) + i },
    }))
    if (source === 'nav' && state.siteVisualCase) {
      const ids = state.siteVisualCase === 'empty' ? [] : [...new Set<number>(data.recent_changes.map((event: { entity: { id: number } }) => event.entity.id))]
      const siteVisuals = ids.map(id => ({ site_id: id, name: `Site ${id}`, visual: { kind: 'site_icon', asset: `nav/sites/${id}/icon/${'a'.repeat(32)}.png` } }))
      if (state.siteVisualCase === 'invalid') siteVisuals.forEach(item => { item.visual.asset = `nav/sites/999/icon/${'a'.repeat(32)}.png` })
      Object.assign(data, { site_visuals: siteVisuals })
    }
    if (state.metricCase === 'empty') data.metrics = []
    if (state.metricCase === 'zero') { data.entity_count = 0; data.changes_7d = 0; data.metrics.forEach((metric: { value: number; delta_30d: number }) => { metric.value = 0; metric.delta_30d = 0 }) }
    if (state.metricCase === 'fallback') data.metrics.find((metric: { key: string }) => metric.key === (source === 'nav' ? 'ipv6' : 'free')).value = null
    if (state.metricCase === 'missing-date') data.metrics.forEach((metric: { as_of: string; coverage: null; delta_30d: null }) => { metric.as_of = ''; metric.coverage = null; metric.delta_30d = null })
    if (state.metricCase === 'hero-only') data.metrics = data.metrics.filter((metric: { key: string }) => metric.key === (source === 'nav' ? 'ipv6' : 'free'))
    if (state.metricCase === 'sparse') {
      for (const metric of data.metrics) {
        if (['free', 'windows', 'tls13', 'http2', 'hsts', 'security_txt'].includes(metric.key)) metric.value = metric.key === 'tls13' ? 2 : metric.key === 'hsts' ? -1 : null
        if (metric.key === 'mac' || metric.key === 'csp') metric.value = 0
      }
    }
    if (state.metricCase === 'duplicate') data.metrics = [...data.metrics.slice().reverse(), ...data.metrics]
    if (state.metricCase === 'full-long') {
      data.metrics = data.metrics.filter((metric: { key: string }) => metric.key === (source === 'nav' ? 'certificate_verified' : 'windows'))
      Object.assign(data.metrics[0], { value: 1, delta_30d: null, known: Number.MAX_SAFE_INTEGER, eligible: Number.MAX_SAFE_INTEGER })
    }
    if (state.sampleCase) for (const metric of data.metrics) {
      Object.assign(metric, state.sampleCase === 'zero' ? { known: 0, eligible: 0 }
        : state.sampleCase === 'negative' ? { known: -1, eligible: 200 }
          : state.sampleCase === 'fraction' ? { known: 1.5, eligible: 200 }
            : { known: 201, eligible: 200 })
    }
    if (state.countCase) data.entity_count = state.countCase === 'missing' ? null : -1
    if (source === 'game' && state.candidateCase) {
      // Exercise the still-consumed Overview events. Uncontracted tags, SFW
      // flags and player observations must never qualify their artwork.
      for (const event of data.recent_changes) {
        const evidence = { tags: [] as { code: string }[] | undefined, sfw: false,
          header_url: `${media}/game.svg`, capsule_url: '',
          online_count: { status: 'success', count: 1240, collected_at: '2026-09-01T09:00:00Z' } }
        if (state.candidateCase === 'adult') evidence.tags = [{ code: 'adult' }]
        if (state.candidateCase === 'tagged') { evidence.tags = [{ code: 'adventure' }]; evidence.sfw = true }
        if (state.candidateCase === 'missing-tags') evidence.tags = undefined
        if (state.candidateCase === 'no-image') { evidence.header_url = ''; Object.assign(event.entity, { visual: undefined }) }
        if (state.candidateCase === 'stale') evidence.online_count.collected_at = '2000-01-01T00:00:00Z'
        if (state.candidateCase === 'invalid-time') evidence.online_count.collected_at = 'invalid'
        if (state.candidateCase === 'invalid-count') evidence.online_count.count = -1
        if (state.candidateCase === 'zero') evidence.online_count.count = 0
        if (state.candidateCase === 'long-title') event.entity.name = '超长作品标题 LongTitleWithoutBreaks'.repeat(20)
        Object.assign(event.entity, evidence)
      }
    }
    return { data }
  },
)

// Overview-only media interception: stable layout substitutes, never product
// artwork or evidence of production qualification. Shared fixtures stay intact.
export const test = overviewRuntime.extend({
  page: async ({ page, context, runtime }, use) => {
    const approved = new Set([91, 92, 93].flatMap(id => steamSharedAssetCandidates(`https://shared.steamstatic.com/steam/apps/${id}/header.jpg`)))
    for (const url of approved) runtime.assets.add(url)
    await context.route(url => approved.has(url.href), route => {
      if (runtime.failImages) return route.fallback()
      if (runtime.state.brokenGame && route.request().url().includes(`/apps/${runtime.state.brokenGame}/`)) return route.fulfill({ contentType: 'image/svg+xml', body: 'invalid image fixture' })
      return route.fulfill({ contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="460" height="215" viewBox="0 0 460 215"><rect width="460" height="215" fill="#394b44"/><circle cx="350" cy="62" r="28" fill="#c8b398"/><path d="M0 215V155L110 58l120 157M190 215l125-123 145 110v13" fill="#778978"/><text x="28" y="180" fill="#fff" font-family="sans-serif" font-size="18">GAME · FIXTURE</text></svg>' })
    })
    await use(page)
  },
})
export { expect, openRuntime, revealImages, keyboardFocus } from './insights-runtime'
