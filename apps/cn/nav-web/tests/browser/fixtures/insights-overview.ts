import { mockOverview, mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { runtimeTest } from './insights-runtime'
export const sources = ['/api/v2/nav/insights/overview', '/api/v2/game/insights/overview', '/api/v2/game/home']
export const test = runtimeTest(() => ({ failure: '', siteHero: false }),
  url => sources.includes(url.pathname),
  (url, media, _body, state) => {
    const source = url.pathname === sources[0] ? 'nav' : url.pathname === sources[1] ? 'game' : 'panel'
    if (state.failure === source || state.failure === 'all') return { status: 503 }
    const data = source === 'panel' ? mockGameHome(media) : mockOverview(source === 'nav' ? 'site' : 'game', media)
    if (state.siteHero && source === 'nav') data.recent_changes[0].occurred_at = '2026-09-02T12:00:00Z'
    return { data }
  },
)
export { expect, openRuntime, revealImages, keyboardFocus } from './insights-runtime'
