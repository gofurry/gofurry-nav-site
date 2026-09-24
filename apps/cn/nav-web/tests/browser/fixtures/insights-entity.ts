import { runtimeTest } from './insights-runtime'
import { allowedSEO, seoState, seoResponse } from './seo-recovery'
import { mockOverview } from '../../../scripts/fixtures/insights-overview.mjs'
export const test = runtimeTest(seoState,
  url => allowedSEO(url) || ['/api/v2/nav/insights/overview', '/api/v2/game/insights/overview'].includes(url.pathname),
  (url, media, _body, state) => url.pathname.endsWith('/insights/overview')
    ? { data: mockOverview(url.pathname.includes('/nav/') ? 'site' : 'game', media) } : seoResponse(url, media, state),
)
export { expect, openRuntime, revealImages } from './insights-runtime'
