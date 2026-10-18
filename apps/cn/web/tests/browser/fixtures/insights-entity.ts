import { runtimeTest } from './insights-runtime'
import { allowedSEO, seoState, seoResponse } from './seo-recovery'
import { mockOverview } from '../../../scripts/fixtures/insights-overview.mjs'
const isEntryHistory = (url: URL) => /^\/api\/v2\/nav\/sites\/(41|42)\/targets\/target\.example\/observations$/.test(url.pathname)
export const test = runtimeTest(seoState,
  url => allowedSEO(url) || isEntryHistory(url) || ['/api/v2/nav/insights/overview', '/api/v2/game/insights/overview'].includes(url.pathname),
  (url, media, _body, state) => isEntryHistory(url) ? { data: { target: 'target.example', protocol: 'ping', items: [] } } : url.pathname.endsWith('/insights/overview')
    ? { data: mockOverview(url.pathname.includes('/nav/') ? 'site' : 'game', media) } : seoResponse(url, media, state),
)
export { expect, openRuntime, revealImages } from './insights-runtime'
