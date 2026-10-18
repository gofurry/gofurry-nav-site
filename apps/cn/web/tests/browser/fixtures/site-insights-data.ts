export const siteInsightScenario = () => ({ trend: 'ready' as 'ready' | 'empty' | 'unavailable' | 'one' | 'gaps', rich: false, nullEcosystem: false })
export const isSiteTrend = (url: URL) => /^\/api\/v2\/nav\/insights\/metrics\/(ipv6|tls13|http2|hsts|csp|security_txt|certificate_verified)\/trend$/.test(url.pathname)
export function siteTrendResponse(url: URL, mode = 'ready') {
  if (mode === 'unavailable') return { status: 503 }
  return { data: { key: url.pathname.split('/').at(-2), requested_range: url.searchParams.get('range'), available_from: '2026-09-01', available_through: '2026-09-03',
    points: (mode === 'empty' ? [] : mode === 'one' ? [0] : mode === 'gaps' ? [null, null, null] : [.25, null, .5]).map((value, index) => ({ date: `2026-09-0${index + 1}`, value, coverage: .8 })) } }
}
export const siteRecentChanges = () => ['site.ipv6.enabled', 'site.primary_target.changed', 'site.tls_certificate.changed', 'site.future.changed', 'site.csp.added', 'site.http2.enabled']
  .map((type, index) => ({ type, date: `2026-09-2${index}`, occurred_at: index === 5 ? '2026-09-25T12:34:00Z' : null, entity: { id: 41, name: 'Site fixture 41' }, detail: null }))
