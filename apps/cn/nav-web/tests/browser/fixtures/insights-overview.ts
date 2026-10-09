import { mockOverview } from '../../../scripts/fixtures/insights-overview.mjs'
import { runtimeTest } from './insights-runtime'
export const sources = ['/api/v2/nav/insights/overview', '/api/v2/game/insights/overview']
export const test = runtimeTest(() => ({ failure: '', siteHero: false, metricCase: '', candidateCase: '', eventCount: -1, countCase: '' }),
  url => sources.includes(url.pathname),
  (url, media, _body, state) => {
    const source = url.pathname === sources[0] ? 'nav' : 'game'
    if (state.failure === source || state.failure === 'all') return { status: 503 }
    const data = mockOverview(source === 'nav' ? 'site' : 'game', media)
    if (state.siteHero && source === 'nav') data.recent_changes[0].occurred_at = '2026-09-02T12:00:00Z'
    if (state.eventCount >= 0) data.recent_changes = Array.from({ length: source === 'nav' ? Math.ceil(state.eventCount / 2) : Math.floor(state.eventCount / 2) }, (_, i) => ({
      ...data.recent_changes[i % 2], entity: { ...data.recent_changes[i % 2].entity, id: (source === 'nav' ? 41 : 81) + i },
    }))
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
export { expect, openRuntime, revealImages, keyboardFocus } from './insights-runtime'
