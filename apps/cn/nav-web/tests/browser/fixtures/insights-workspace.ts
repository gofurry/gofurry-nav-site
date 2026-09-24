import { workspaceFixtureResponse, workspaceMetricKeys, workspaceRegionKeys } from '../../../scripts/fixtures/insights-workspace.mjs'
import { runtimeTest } from './insights-runtime'
export const test = runtimeTest(() => ({ failure: '', empty: false }),
  url => /^\/api\/v2\/(game\/insights\/(players\/ranking|prices\/(overview|discounts)|languages\/overview)|nav\/insights\/certificates\/overview)$/.test(url.pathname),
  (url, media, _body, state) => workspaceFixtureResponse(url, media, state),
)
export { workspaceMetricKeys, workspaceRegionKeys }
export { expect, openRuntime, revealImages, keyboardFocus } from './insights-runtime'
