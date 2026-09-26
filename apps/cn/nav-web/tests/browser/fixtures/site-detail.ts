import { runtimeTest } from './insights-runtime'
import { seoState, seoResponse } from './seo-recovery'

// #109 owns only scenario data; Nitro, request instances, release gates and
// browser diagnostics remain in the shared runtime fixture.
export const test = runtimeTest(
  () => ({ ...seoState(), insightsEmpty: false, viewFailure: false }),
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
    const reply = seoResponse(url, media, state)
    if (url.pathname.endsWith('/detail') && !reply.status) {
      const target = url.searchParams.get('target') || 'target.example'
      return { data: { ...reply.data as Record<string, unknown>, latest_core: {
        target, protocols: { http: {
          status: 'success', observed_at: '2026-08-30T12:00:00Z', duration_ms: 120,
          payload: { final_url: `https://${target}/`, status_code: target === 'alt.example' ? 201 : 200,
            response_time_ms: 120, http_protocol: 'HTTP/2', headers: {}, meta: {} },
        } },
      } } }
    }
    return reply
  },
  { NUXT_PUBLIC_SITE_URL: 'https://go-furry.com', NUXT_PUBLIC_I18N_BASE_URL: 'https://go-furry.com' },
)
export { expect, openRuntime, settleRuntime, assertRuntimeSurface } from './insights-runtime'
