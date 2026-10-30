import { runtimeTest } from './insights-runtime'
import { seoResponse, seoState } from './seo-recovery'

const allowed = (url: URL) => ['/api/v2/nav/home/hero', '/api/v2/nav/sites/directory',
  '/api/v2/nav/sites/index', '/api/v2/game/list', '/api/v2/game/info'].includes(url.pathname)

export const test = runtimeTest(() => ({}), allowed, (url, media) => {
  if (url.pathname === '/api/v2/nav/home/hero') return { data: { hero: { desktop: null, mobile: null } } }
  if (url.pathname === '/api/v2/nav/sites/directory') return { data: [{
    id: '41', name: 'Site fixture 41', info: 'Independent Site detail.', domain: 'target.example',
    icon: '', country: 'CN', nsfw: '0', welfare: '0', view_count: 1,
  }, { id: '42', name: 'Other profile sentinel', info: '', domain: 'alt.example' }] }
  return seoResponse(url, media, seoState())
}, { NUXT_PUBLIC_TRANSITION_MODE: 'true' })
