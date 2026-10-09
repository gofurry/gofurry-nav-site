import { mockOverview, mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { runtimeTest } from './insights-runtime'
export const sources = ['/api/v2/nav/insights/overview', '/api/v2/game/insights/overview', '/api/v2/game/home']
export const test = runtimeTest(() => ({ failure: '', siteHero: false, metricCase: '', candidateCase: '', eventCount: -1 }),
  url => sources.includes(url.pathname),
  (url, media, _body, state) => {
    const source = url.pathname === sources[0] ? 'nav' : url.pathname === sources[1] ? 'game' : 'panel'
    if (state.failure === source || state.failure === 'all') return { status: 503 }
    const data = source === 'panel' ? mockGameHome(media) : mockOverview(source === 'nav' ? 'site' : 'game', media)
    if (state.siteHero && source === 'nav') data.recent_changes[0].occurred_at = '2026-09-02T12:00:00Z'
    if (source !== 'panel') {
      if (state.eventCount >= 0) data.recent_changes = Array.from({ length: state.eventCount }, (_, i) => ({
        ...data.recent_changes[i % 2], entity: { ...data.recent_changes[i % 2].entity, id: (source === 'nav' ? 41 : 81) + i },
      }))
      if (state.metricCase === 'empty') data.metrics = []
      if (state.metricCase === 'zero') { data.entity_count = 0; data.changes_7d = 0; data.metrics.forEach((metric: { value: number; delta_30d: number }) => { metric.value = 0; metric.delta_30d = 0 }) }
      if (source === 'nav' && state.metricCase === 'fallback') data.metrics.find((metric: { key: string }) => metric.key === 'ipv6').value = null
      if (state.metricCase === 'missing-date') data.metrics.forEach((metric: { as_of: string; coverage: null; delta_30d: null }) => { metric.as_of = ''; metric.coverage = null; metric.delta_30d = null })
    } else {
      const games = [...data.panel.top_online, ...data.panel.highest_discount, ...data.panel.latest_games]
      for (const game of games) {
        if (state.candidateCase === 'adult') game.tags = [{ code: 'adult' }]
        if (state.candidateCase === 'tagged') { game.tags = [{ code: 'adventure' }]; game.sfw = true }
        if (state.candidateCase === 'missing-tags') delete game.tags
        if (state.candidateCase === 'no-image') { game.header_url = ''; game.capsule_url = '' }
        if (state.candidateCase === 'stale') game.online_count.collected_at = '2000-01-01T00:00:00Z'
        if (state.candidateCase === 'invalid-time') game.online_count.collected_at = 'invalid'
        if (state.candidateCase === 'invalid-count') game.online_count.count = -1
        if (state.candidateCase === 'zero') game.online_count.count = 0
        if (state.candidateCase === 'long-title') game.name = '超长作品标题 LongTitleWithoutBreaks'.repeat(20)
      }
    }
    return { data }
  },
)
export { expect, openRuntime, revealImages, keyboardFocus } from './insights-runtime'
