import type { Page } from '@playwright/test'
import { runtimeTest, openRuntime, settleRuntime, assertRuntimeSurface, expect } from './insights-runtime'
import { securityEvidence, securityState } from './site-security-data'
import { isSiteTrend } from './site-insights-data'

// The same instant owns the SSR evidence and browser clock. Real timers/RAF stay live.
export const siteVisualNow = '2026-09-27T12:00:00Z'
const target = 'target.example', alternate = 'alt.example', siteId = 41, views = 20457
// The existing loopback upstream serves these managed keys as fixed local SVGs.
const icon = (id: number) => `nav/sites/${id}/icon/${'8'.repeat(32)}.svg`
const names = ['兽人创作工坊', '绒毛社群', '灵感档案馆', '山林茶话会', '星河画室', '爪印游戏库', '月光放映室', '旅行者手册']

function scene() {
  const security = securityEvidence(target, { ...securityState(), days: undefined })
  const http = { ...security.http, observed_at: siteVisualNow, payload: { ...security.http.payload,
    cert_not_before: '2026-08-01T00:00:00Z', cert_not_after: '2026-11-11T12:00:00Z',
    dns_lookup_ms: 5, tcp_connect_ms: 7, tls_handshake_ms: 11, ttfb_ms: 90, transfer_ms: 13,
    remote_ip: '203.0.113.41', body_read_bytes: 18432, content_type: 'text/html; charset=utf-8', server: 'fixture-edge',
    redirect_chain: ['http://target.example/', 'https://target.example/', 'https://target.example/community',
      'https://target.example/zh', 'https://target.example/zh/home', 'https://target.example/welcome', 'https://target.example/index'],
  } }
  const targets = [{ target, status: 'healthy', reason_codes: [], reason_messages: [] },
    { target: alternate, status: 'down', reason_codes: ['http_failed'], reason_messages: [] }]
  return {
    detail: {
      site: { id: siteId, name: 'GoFurry 社群索引', info: '发现兽人社群的创作、游戏与交流空间。记录每一个站点的技术观测，让探索有迹可循。', icon: icon(siteId), country: 'CN', nsfw: '0', welfare: '1', view_count: views },
      selected_target: target,
      site_summary: { state: 'ready', status: 'degraded', generated_at: siteVisualNow, target_count: 2, targets,
        status_counts: { healthy: 1, down: 1 }, reason_codes: ['some_targets_degraded'], reason_messages: [] },
      target_summary: { state: 'ready', target, status: 'healthy', observed_at: siteVisualNow, reason_codes: [], reason_messages: [],
        protocols: Object.fromEntries(['ping', 'http', 'dns'].map(protocol => [protocol, { protocol, status: 'success', observed_at: siteVisualNow, duration_ms: protocol === 'http' ? 120 : 24, stale: false }])),
        edge_provider_hints: [{ provider: 'cloudflare', hint_type: 'cdn', confidence: 'high', evidence: [] }] },
      latest_core: { site_id: siteId, target, state: 'ready', protocols: {
        http,
        ping: { protocol: 'ping', target, status: 'success', observed_at: siteVisualNow, duration_ms: 24, payload: { avg_rtt_ms: 24, jitter_ms: 2, loss_rate: 0, resolved_ip: '203.0.113.41' } },
        dns: { protocol: 'dns', target, status: 'success', observed_at: siteVisualNow, duration_ms: 24, payload: { A: [{ type: 'A', value: '203.0.113.41', ttl: 300 }], AAAA: [], risk_flags: [] } },
      } },
      light_probe_state: null,
    },
    insights: {
      site: { id: siteId, name: 'GoFurry 社群索引' },
      capabilities: [
        ['ipv6', 'supported', .582, .921], ['http2', 'supported', .813, .953], ['tls13', 'supported', .791, .942],
        ['certificate_verified', 'supported', .931, .947], ['hsts', 'supported', .468, .924], ['csp', 'unsupported', .273, .912], ['security_txt', 'not_probed', .143, .864],
      ].map(([key, state, value, coverage]) => ({ key, state, as_of: '2026-09-27', ecosystem: { value, coverage } })),
      recent_changes: ['site.http2.enabled', 'site.primary_target.changed', 'site.tls_certificate.changed', 'site.ipv6.enabled', 'site.hsts.added', 'site.tls13.enabled']
        .map((type, index) => ({ type, date: `2026-09-${25 - index}`, occurred_at: index === 0 ? '2026-09-25T12:34:00Z' : null, entity: { id: siteId, name: 'GoFurry 社群索引' }, detail: null })),
    },
    recommendations: { schema_version: 1, generated_at: siteVisualNow, state: 'ready', site_id: siteId,
      items: names.map((name, index) => ({ id: String(42 + index), name, info: '分享创作与日常，遇见新的朋友。', domain: `similar-${index}.example`, icon: icon(42 + index), nsfw: '0', welfare: '1', country: 'CN', view_count: 24088 + index * 137, create_time: '', update_time: '' })) },
    history: { target, protocol: 'ping', items: Array.from({ length: 100 }, (_, index) => ({ target, protocol: 'ping', status: 'success',
      observed_at: new Date(Date.parse(siteVisualNow) - index * 60000).toISOString(), duration_ms: 24,
      payload: { avg_rtt_ms: [24, 29, 25, 32, 28, 26, 30, 23, 27][index % 9], loss_rate: 0 },
    })) },
  }
}

const base = runtimeTest(scene,
  url => /^\/api\/v2\/nav\/sites\/41\/(detail|insights|recommendations|view)$/.test(url.pathname)
    || url.pathname === '/api/v2/nav/sites/41/targets/target.example/observations' || isSiteTrend(url),
  (url, _media, _body, state) => {
    if (url.pathname.endsWith('/detail')) return { data: state.detail }
    if (url.pathname.endsWith('/insights')) return { data: state.insights }
    if (url.pathname.endsWith('/recommendations')) return { data: state.recommendations }
    if (url.pathname.endsWith('/view')) return { data: { site_id: siteId, view_count: views } }
    if (url.pathname.endsWith('/observations')) return { data: state.history }
    return { data: { key: url.pathname.split('/').at(-2), requested_range: url.searchParams.get('range'), available_from: '2026-09-01', available_through: '2026-09-27',
      points: Array.from({ length: 27 }, (_, index) => ({ date: `2026-09-${String(index + 1).padStart(2, '0')}`, value: index === 12 ? null : .72 + index * .003, coverage: .942 })) } }
  }, { NUXT_PUBLIC_SITE_URL: 'https://go-furry.com', NUXT_PUBLIC_I18N_BASE_URL: 'https://go-furry.com' }, { fixedTime: siteVisualNow })

export type SiteVisualScene = 'performance' | 'overview' | 'security-tls' | 'insights' | 'http' | 'similar'
type Options = { scene: SiteVisualScene; theme: 'light' | 'dark'; width: 1440 | 390 }

async function settleVisual(page: Page) {
  const root = page.locator('[data-site-detail]')
  // Loading is evidence-based; revealing images also covers native lazy managed images.
  for (const image of await root.locator('img:visible').all()) {
    await image.scrollIntoViewIfNeeded()
    await expect.poll(() => image.evaluate(node => (node as HTMLImageElement).complete && (node as HTMLImageElement).naturalWidth > 0)).toBe(true)
  }
  await page.evaluate(() => { (document.activeElement as HTMLElement | null)?.blur(); window.scrollTo({ top: 0, behavior: 'instant' }) })
  await page.mouse.move(0, 0)
  await settleRuntime(page)
  await page.evaluate(async () => {
    await document.fonts.ready
    await Promise.all(document.getAnimations().filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity).map(animation => animation.finished.catch(() => {})))
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
  })
}

export const test = base.extend<{ siteVisual: { open(options: Options): Promise<void>; capture(name: string): Promise<void> } }>({
  siteVisual: async ({ page, context, runtime }, use) => {
    await use({
      async open(options) {
        await page.setViewportSize({ width: options.width, height: 900 })
        await page.clock.setFixedTime(new Date(siteVisualNow))
        await context.addInitScript(theme => { localStorage.setItem('theme', theme); localStorage.setItem('mode', 'sfw') }, options.theme)
        const paths: Record<SiteVisualScene, string> = { performance: '', overview: '?tab=overview', 'security-tls': '?tab=security&view=tls', insights: '?tab=insights&metric=tls13&range=90d', http: '?tab=observation&view=http', similar: '' }
        const counted = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
        await openRuntime(page, '/site/41' + paths[options.scene]); await (await counted).finished()
        if (options.scene === 'performance' || options.scene === 'similar') await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
        if (options.scene === 'insights') await expect(page.locator('[data-site-insight-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
        if (options.scene === 'similar') {
          const url = page.url()
          await page.locator('[data-site-aux-tab="similar"]').click()
          await expect(page.locator('[data-site-similar]')).toHaveAttribute('role', 'tabpanel')
          await expect(page.locator('[data-site-workspace]')).toBeHidden()
          expect(page.url()).toBe(url)
          await expect(page.locator('[data-site-similar]')).not.toContainText(/similar-\d\.example/)
        } else if (options.scene === 'security-tls') {
          expect(runtime.state.detail.latest_core.protocols.http.payload.cert_days_left).toBeUndefined()
          await expect(page.locator('[data-site-certificate-expiry]')).toContainText('45')
        } else if (options.scene === 'http') await expect(page.locator('[data-site-redirect-node]')).toHaveCount(7)
        if (options.width === 1440) await expect(page.locator('[data-site-detail-aside] [data-site-similar]')).toBeVisible()
        // Unrelated fixed navigation tools have their own Visual owners.
        await page.addStyleTag({ content: '.page-scroll-dock, .mobile-bottom-tabs-root { display: none !important; }' })
        await settleVisual(page)
        await assertRuntimeSurface(page, '[data-site-detail]', options.theme)
        runtime.assertQuiet()
      },
      async capture(name) {
        await settleVisual(page)
        runtime.assertQuiet()
        await expect(page.locator('[data-site-detail]')).toHaveScreenshot(name)
        runtime.assertQuiet()
      },
    })
  },
})
export { expect }
