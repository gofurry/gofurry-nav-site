import { runtimeTest } from './insights-runtime'
import { seoState, seoResponse } from './seo-recovery'

export const longTarget = 'a-long-collected-target-name-for-layout-review.community.infrastructure.example'

// #109 owns only scenario data; Nitro, request instances, release gates and
// browser diagnostics remain in the shared runtime fixture.
export const test = runtimeTest(
  () => ({ ...seoState(), insightsEmpty: false, viewFailure: false, noTargetEvidence: false, missingSummary: false, longTarget: false, primaryStatus: 200, extraTargets: [] as string[] }),
  url => /^\/api\/v2\/nav\/sites\/(41|42|999999999)\/(detail|insights|view)$/.test(url.pathname)
    || /^\/api\/v2\/nav\/sites\/41\/targets\/(target|alt)\.example\/observations$/.test(url.pathname),
  (url, media, _body, state) => {
    if (url.pathname.endsWith('/view') && state.viewFailure) return { status: 503 }
    if (url.pathname.endsWith('/insights') && state.insightsEmpty) return {
      data: { site: { id: 41, name: 'Site fixture 41' }, capabilities: [], recent_changes: [] },
    }
    if (url.pathname.endsWith('/observations')) return { data: { items: [
      { status: 'success', observed_at: '2026-08-30T12:00:00Z', duration_ms: 24, payload: { avg_rtt_ms: 24, loss_rate: 0 } },
    ] } }
    const scenarioUrl = new URL(url)
    if ((state.longTarget && url.searchParams.get('target') === longTarget) || state.extraTargets.includes(url.searchParams.get('target') ?? '')) scenarioUrl.searchParams.set('target', 'alt.example')
    const reply = seoResponse(scenarioUrl, media, state)
    if (url.pathname.endsWith('/detail') && !reply.status) {
      const target = url.searchParams.get('target') || 'target.example'
      const observed = '2026-08-30T12:00:00Z'
      const alternative = state.longTarget ? longTarget : 'alt.example'
      return { data: { ...reply.data as Record<string, unknown>, selected_target: target,
        site_summary: state.missingSummary ? { state: 'missing', status: 'unknown', targets: null } : { state: 'ready', status: 'down', targets: ['target.example', alternative, ...state.extraTargets].map(target => ({ target })),
          target_relation_hints: [{ relation: 'shared_canonical', host: 'target.example', targets: ['target.example', alternative] }] },
        target_summary: state.noTargetEvidence ? { state: 'missing', target, status: 'unknown', observed_at: '0001-01-01T00:00:00Z', protocols: {} } : { state: 'ready', target, status: target === 'target.example' ? 'healthy' : 'warning', observed_at: observed,
          protocols: Object.fromEntries(['ping', 'http', 'dns'].map(protocol => [protocol, { protocol, status: 'success', observed_at: observed, duration_ms: 24, stale: false }])),
          edge_provider_hints: [{ provider: 'cloudflare', hint_type: 'response_header', confidence: 'medium', evidence: [{ source: 'http', field: 'server', value: 'cloudflare' }] }] },
        latest_core: state.noTargetEvidence ? null : {
        target, protocols: { http: {
          target, status: 'success', observed_at: observed, duration_ms: 120,
          payload: { final_url: `https://${target}/`, status_code: target === 'target.example' ? state.primaryStatus : 201,
            response_time_ms: 120, http_protocol: 'HTTP/2', headers: {}, meta: {} },
        } },
      } } }
    }
    return reply
  },
  { NUXT_PUBLIC_SITE_URL: 'https://go-furry.com', NUXT_PUBLIC_I18N_BASE_URL: 'https://go-furry.com' },
)
export { expect, openRuntime, settleRuntime, assertRuntimeSurface } from './insights-runtime'
