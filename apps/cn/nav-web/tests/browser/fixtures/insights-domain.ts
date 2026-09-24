import { domainFixtureResponse, domainMetricKeys, domainDimensionKeys } from '../../../scripts/fixtures/insights-domain.mjs'
import { runtimeTest } from './insights-runtime'
export const test = runtimeTest(() => ({ panelFailure: false, overviewFailure: false, breakdownFailure: false, trendMode: 'normal' }),
  url => url.pathname === '/api/v2/game/home' || /^\/api\/v2\/(nav|game)\/insights\/(overview|metrics\/(ipv6|tls13|http2|hsts|csp|security_txt|certificate_verified|free|windows|mac|linux)\/(trend|breakdown(?:\/[a-z_]+\/[\w-]+\/trend)?))$/.test(url.pathname),
  (url, media, _body, state) => domainFixtureResponse(url, media, state),
)
export { domainMetricKeys, domainDimensionKeys }
export { expect, openRuntime, keyboardFocus } from './insights-runtime'
