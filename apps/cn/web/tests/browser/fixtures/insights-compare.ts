import { compareFixtureResponse, compareGroups } from '../../../scripts/fixtures/insights-compare.mjs'
import { runtimeTest } from './insights-runtime'
export const test = runtimeTest(() => ({ searches: [] as string[], compareFailure: false, directoryFailure: false,
  searchFailure: false, insufficient: false, reverse: false, controlledTiming: true }),
  url => /^\/api\/v2\/(nav\/(insights\/compare|sites\/directory)|game\/(insights\/compare|search\/simple))$/.test(url.pathname),
  (url, media, body, state) => compareFixtureResponse(url, media, body, state),
)
export const pathFor = (domain: string, locale = 'zh', ids = '', region = 'CN') =>
  (locale === 'en' ? '/en' : '') + '/insights/' + (domain === 'site' ? 'sites' : 'games') + '/compare?' +
  new URLSearchParams({ ...(ids ? { ids } : {}), ...(domain === 'game' ? { region } : {}) })
export { compareGroups }
export { expect, openRuntime, revealImages, keyboardFocus } from './insights-runtime'
