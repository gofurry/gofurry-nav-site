import { changesFixtureResponse, changeCategories } from '../../../scripts/fixtures/insights-changes.mjs'
import { runtimeTest } from './insights-runtime'
export const test = runtimeTest(() => ({ failure: false, moreFailure: false, empty: false, delayRange: '' }),
  url => /^\/api\/v2\/(nav|game)\/insights\/changes$/.test(url.pathname),
  (url, media, _body, state) => changesFixtureResponse(url, media, state),
)
export const pathFor = (domain: string, locale = 'zh') => (locale === 'en' ? '/en' : '') + '/insights/changes?domain=' + domain + '&range=30d'
export { changeCategories }
export { expect, openRuntime, revealImages, keyboardFocus, settleRuntime } from './insights-runtime'
