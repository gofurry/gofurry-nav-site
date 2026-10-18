import { runtimeTest } from './insights-runtime'
import { seoState, seoResponse } from './seo-recovery'
import { securityState, securityEvidence } from './site-security-data'
import { isSiteTrend, siteInsightScenario, siteRecentChanges, siteTrendResponse } from './site-insights-data'

export const longTarget = 'a-long-collected-target-name-for-layout-review.community.infrastructure.example'

// #109 owns only scenario data; Nitro, request instances, release gates and
// browser diagnostics remain in the shared runtime fixture.
export const test = runtimeTest(
  () => ({ ...seoState(), insightsEmpty: false, viewFailure: false, viewCount: 2, noTargetEvidence: false, missingSummary: false, longTarget: false, primaryStatus: 200, extraTargets: [] as string[],
    recommendations: 'empty' as 'empty' | 'ready' | 'adult' | 'unavailable' | 'failure',
    summaryScenario: 'healthy' as 'healthy' | 'mixed' | 'stale' | 'unknown' | 'zero', summaryChangesOnTarget: false, fullCapabilities: false, manyChanges: false,
    cdnScenario: 'none' as 'none' | 'reliable' | 'unreliable', protocolScenario: 'normal' as 'normal' | 'slow' | 'failure' | 'missing', observationRich: false, observationHealthReasons: true, webProbeScenario: 'ready' as 'ready' | 'not-found' | 'failure' | 'stale' | 'missing', historyCount: 1, historyFailure: false, historyNoRtt: false, noRedirects: false, redirectCount: 2, noCname: false, longEvidence: false, security: securityState(), intelligence: siteInsightScenario() }),
  url => /^\/api\/v2\/nav\/sites\/(41|42|999999999)\/(detail|insights|recommendations|view)$/.test(url.pathname)
    || url.pathname === '/api/v2/nav/home'
    || /^\/api\/v2\/nav\/sites\/(41|42)\/targets\/(target|alt)\.example\/observations$/.test(url.pathname) || isSiteTrend(url),
  (url, media, _body, state) => {
    if (url.pathname === '/api/v2/nav/home') return { data: {
      schema_version: 4, groups: [{ id: '12', name: 'Site entry fixture', priority: 1, sites: [{
        id: '41', name: 'Site fixture 41', info: 'Open the collected Target from Home.',
        domain: ['target.example', 'alt.example'], icon: '', nsfw: '0', welfare: '0', view_count: 1,
      }] }], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
      saying: { saying: 'Fixture', author: 'Fixture' }, ping: {}, hero: { desktop: null, mobile: null },
    } }
    if (url.pathname.endsWith('/recommendations')) {
      const id = Number(url.pathname.split('/').at(-2)), lang = url.searchParams.get('lang')
      if (url.searchParams.size !== 2 || !['zh', 'en'].includes(lang ?? '') || url.searchParams.get('limit') !== '8') throw new Error('Invalid recommendation identity ' + url)
      if (state.recommendations === 'failure') return { status: 503 }
      return { data: { schema_version: 1, generated_at: '2026-09-27T00:00:00Z', site_id: id,
        state: state.recommendations === 'unavailable' ? 'unavailable' : 'ready',
        items: ['empty', 'unavailable'].includes(state.recommendations) ? [] : Array.from({ length: 8 }, (_, index) => ({
          id: String(id + index + 1), name: `${lang === 'en' ? 'Similar fixture' : '相似站点示例'} ${id + index + 1}`,
          domain: index === 0 ? '{"domain":["first.example","second.example"]}' : `similar${index}.example`, info: lang === 'en' ? 'A community to explore.' : '探索新的社群。',
          icon: '', country: 'CN', welfare: '1', nsfw: state.recommendations === 'adult' || index === 1 || index === 5 ? '1' : '0',
          view_count: 24088 + index, create_time: '', update_time: '',
        })) } }
    }
    if (isSiteTrend(url)) return siteTrendResponse(url, state.intelligence.trend)
    if (url.pathname.endsWith('/view')) return state.viewFailure ? { status: 503 } : { data: { site_id: Number(url.pathname.split('/').at(-2)), view_count: state.viewCount } }
    if (url.pathname.endsWith('/insights') && state.insightsEmpty) return {
      data: { site: { id: 41, name: 'Site fixture 41' }, capabilities: [], recent_changes: [] },
    }
    if (url.pathname.endsWith('/observations')) {
      if (state.historyFailure) return { status: 503 }
      const target = decodeURIComponent(url.pathname.split('/').at(-2)!)
      return { data: { target, protocol: 'ping', items: Array.from({ length: state.historyCount }, (_, index) => ({
        target, protocol: 'ping', status: state.historyNoRtt ? 'failure' : 'success',
        observed_at: new Date(Date.UTC(2026, 8, 26, 12, 0) - index * 60000).toISOString(), duration_ms: 24,
        payload: state.historyNoRtt ? {} : { avg_rtt_ms: (target === 'target.example' ? 24 : 70) + index % 9, loss_rate: index === 1 ? 10 : 0 },
      })) } }
    }
    const scenarioUrl = new URL(url)
    if ((state.longTarget && url.searchParams.get('target') === longTarget) || state.extraTargets.includes(url.searchParams.get('target') ?? '')) scenarioUrl.searchParams.set('target', 'alt.example')
    const reply = seoResponse(scenarioUrl, media, state)
    if (url.pathname.endsWith('/insights') && !reply.status) {
      const data = reply.data as Record<string, unknown>
      return { data: { ...data,
        ...(state.fullCapabilities ? { capabilities: [
          ['ipv6', 'supported'], ['http2', 'unsupported'], ['tls13', 'stale'], ['certificate_verified', 'not_probed'],
          ['hsts', 'unavailable'], ['csp', 'unknown'], ['security_txt', 'not_applicable'],
        ].map(([key, capabilityState]) => ({ key, state: capabilityState, as_of: '2026-09-24', ecosystem: { value: .5, coverage: .8 } })) } : {}),
        ...(state.manyChanges ? { recent_changes: [
          ['site.ipv6.enabled', '2026-09-20', null], ['site.http2.enabled', '2026-09-24', '2026-09-24T12:34:00Z'],
          ['site.tls13.enabled', '2026-09-23', null], ['site.hsts.added', '2026-09-22', null], ['site.csp.added', '2026-09-21', null],
        ].map(([type, date, occurred_at]) => ({ type, date, occurred_at, entity: { id: 41, name: 'Site fixture 41' }, detail: null })) } : {}),
        ...(state.intelligence.rich ? { recent_changes: siteRecentChanges() } : {}),
        ...(state.intelligence.nullEcosystem ? { capabilities: [{ key: 'ipv6', state: 'unknown', as_of: null, ecosystem: { value: null, coverage: 0 } }] } : {}),
      } }
    }
    if (url.pathname.endsWith('/detail') && !reply.status) {
      const target = url.searchParams.get('target') || 'target.example'
      const security = state.security.enabled ? securityEvidence(target, state.security) : null
      const observed = '2026-08-30T12:00:00Z'
      const alternative = state.longTarget ? longTarget : 'alt.example'
      const changed = state.summaryChangesOnTarget && target !== 'target.example'
      const mixed = state.summaryScenario === 'mixed' || changed
      const unknown = state.summaryScenario === 'unknown'
      const hosts = state.summaryScenario === 'zero' ? [] : ['target.example', alternative, ...state.extraTargets]
      const targets = hosts.map(host => ({ target: host, status: unknown ? 'unknown' : mixed && host !== 'target.example' ? 'down' : 'healthy',
        reason_codes: mixed && host !== 'target.example' ? ['http_failed'] : [], reason_messages: mixed && host !== 'target.example' ? ['后端旧中文 HTTP 失败'] : [] }))
      const ip = target === 'target.example' ? '203.0.113.41' : '203.0.113.42'
      const envelope = (protocol: string, payload: unknown) => ({ target, protocol, status: 'success', observed_at: observed, duration_ms: 24, payload })
      const dns = { A: [{ type: 'A', value: ip, ttl: 0, asn: 'AS64496', isp: 'Fixture ISP', country: 'CN', reverse_ptr: 'ptr.example', dnssec: false, provider_type: 'origin', hijacked: false }],
        AAAA: [], CNAME: state.noCname ? [] : [{ type: 'CNAME', value: 'edge.example', ttl: 60, children: [{ type: 'A', value: ip, ttl: 0 }] }],
        MX: [{ type: 'MX', value: 'mail.example', ttl: 3600 }], NS: [{ type: 'NS', value: 'ns1.example', ttl: 3600 }],
        TXT: [{ type: 'TXT', value: state.longEvidence ? 'v=DKIM1; p=' + 'M'.repeat(350) : 'v=spf1 -all', ttl: 120 }], CAA: [{ type: 'CAA', value: '0 issue "ca.example"', ttl: 120 }],
        SOA: [{ type: 'SOA', value: 'ns1.example admin.example 1 2 3 4 5', ttl: 3600 }],
        cname_chain_depth: state.noCname ? 0 : 1, cname_terminal: state.noCname ? '' : 'edge.example', risk_flags: ['ptr_empty', 'low_ttl', 'private_ip', 'nxdomain_with_answer', 'collector_reported_signal'] }
      const lightEnvelope = (protocol: string, payload: Record<string, unknown>) => {
        const absent = { exists: false, status_code: 404 }
        const evidence = state.webProbeScenario === 'not-found' && protocol !== 'rdap'
          ? protocol === 'page_assets' ? { icon: absent, manifest: absent } : absent : payload
        return { ...envelope(protocol, state.webProbeScenario === 'failure' ? {} : evidence), status: state.webProbeScenario === 'failure' ? 'failure' : 'success' }
      }
      return { data: { ...reply.data as Record<string, unknown>, selected_target: target,
        site_summary: state.missingSummary ? { state: 'missing', status: 'unknown', targets: null } : {
          state: state.summaryScenario === 'stale' ? 'stale' : 'ready', status: mixed ? 'degraded' : unknown ? 'unknown' : 'healthy',
          target_count: hosts.length, targets, generated_at: changed ? '2026-09-26T13:00:00Z' : '2026-09-26T12:18:00Z',
          status_counts: unknown ? { unknown: hosts.length } : mixed ? { healthy: 1, down: hosts.length - 1 } : { healthy: hosts.length },
          reason_messages: mixed ? ['后端旧中文 HTTP 失败'] : [], reason_codes: mixed ? ['http_failed'] : [],
          target_relation_hints: [{ relation: 'shared_canonical', host: 'target.example', targets: ['target.example', alternative] }] },
        target_summary: state.noTargetEvidence ? { state: 'missing', target, status: 'unknown', observed_at: '0001-01-01T00:00:00Z', protocols: {} } : { state: 'ready', target, status: target === 'target.example' ? 'healthy' : 'warning', observed_at: observed,
          reason_messages: state.observationRich ? ['后端旧中文 DNS 观测过期'] : [], reason_codes: state.observationRich && state.observationHealthReasons ? ['dns_missing_or_stale'] : [],
          protocols: Object.fromEntries(['ping', 'http', 'dns'].map(protocol => [protocol, { protocol, status: state.protocolScenario === 'failure' ? 'failure' : state.protocolScenario === 'missing' ? 'unknown' : 'success', observed_at: observed, duration_ms: state.protocolScenario === 'slow' ? 4302 : 24, stale: state.observationRich && protocol === 'dns' }])),
          edge_provider_hints: state.cdnScenario === 'reliable' ? [
            { provider: 'aliyun', hint_type: 'waf', confidence: 'high', evidence: [] },
            { provider: 'tencent_cloud', hint_type: 'cdn', confidence: 'medium', evidence: [] },
            { provider: target === 'target.example' ? 'cloudflare' : 'fastly', hint_type: 'cdn', confidence: 'high', evidence: [] },
          ] : state.cdnScenario === 'unreliable' ? [
            { provider: 'cloudflare', hint_type: 'waf', confidence: 'high', evidence: [] },
            { provider: 'tencent_cloud', hint_type: 'cdn', confidence: 'low', evidence: [] },
          ] : [{ provider: 'cloudflare', hint_type: 'response_header', confidence: 'medium', evidence: [{ source: 'http', field: 'server', value: 'cloudflare' }] }] },
        latest_core: state.noTargetEvidence ? null : {
        target, protocols: { ...(state.observationRich ? { ping: envelope('ping', { avg_rtt_ms: target === 'target.example' ? 24 : 70, jitter_ms: 0, loss_rate: 0, resolved_ip: ip }), dns: envelope('dns', dns) } : {}), http: {
          target, status: 'success', observed_at: observed, duration_ms: 120,
          payload: { final_url: `https://${target}/`, status_code: target === 'target.example' ? state.primaryStatus : 201,
            response_time_ms: 120, http_protocol: 'HTTP/2', headers: {}, meta: {},
            ...(state.observationRich ? { dns_lookup_ms: 5, tcp_connect_ms: 7, tls_handshake_ms: 11, ttfb_ms: 90, transfer_ms: 13,
              remote_ip: ip, body_read_bytes: 0, content_type: 'text/html', title: `${target} page`, html_charset: 'UTF-8',
              headers: { Server: ['fixture-edge'], 'Content-Type': ['text/html'], 'Cache-Control': ['max-age=0'], 'X-Fixture': ['full-header-value' + (state.longEvidence ? 'x'.repeat(350) : '')], 'Strict-Transport-Security': ['max-age=31536000'] },
              meta: { description: 'Observed page description' + (state.longEvidence ? 'x'.repeat(350) : ''), keywords: 'furry, community' },
              redirect_chain: state.noRedirects ? [] : Array.from({ length: state.redirectCount }, (_, index) => index === 0 ? `http://${target}/` : index === 1 ? `https://${target}/` : `https://${target}/hop-${index}`) } : {}),
          },
          ...(security?.http ?? {}),
        } },
      }, light_probe_state: security?.light ?? (state.observationRich && state.webProbeScenario !== 'missing' ? { target, state: state.webProbeScenario === 'stale' ? 'stale' : 'ready', protocols: {
        robots: lightEnvelope('robots', { exists: true, status_code: 200, sitemap_count: 1, sitemaps: ['https://target.example/sitemap.xml'], user_agent_star_present: true, global_disallow_all: false }),
        llms_txt: lightEnvelope('llms_txt', { exists: true, path: '/llms.txt', status_code: 200, title: 'Community guide', heading_count: 2, link_count: 1, optional_section_present: false, body_read_bytes: 256, headings: ['About', 'Resources'], links: ['https://target.example/about'] }),
        page_assets: lightEnvelope('page_assets', { icon: { exists: true, source_url: 'https://target.example/favicon.ico', content_type: 'image/x-icon', status_code: 200 }, manifest: { exists: true, source_url: 'https://target.example/manifest.json', name: 'Fixture app', display: 'standalone', theme_color: '#123456', start_url: '/', scope: '/', icons_count: 2 } }),
        rdap: lightEnvelope('rdap', { registrable_domain: 'target.example', registrar: 'Fixture registrar', expires_at: '2027-09-26', statuses: ['active'], nameservers: ['ns1.example'], dnssec_delegation_signed: false }),
        security_txt: envelope('security_txt', { contact: ['SECURITY_ONLY_CONTACT'] }), port_check: envelope('port_check', { service: 'SECURITY_ONLY_PORT' }), waf_canary: envelope('waf_canary', { name: 'SECURITY_ONLY_WAF' }),
      } } : null) } }
    }
    return reply
  },
  { NUXT_PUBLIC_SITE_URL: 'https://go-furry.com', NUXT_PUBLIC_I18N_BASE_URL: 'https://go-furry.com' },
)
export { expect, openRuntime, settleRuntime, assertRuntimeSurface } from './insights-runtime'
