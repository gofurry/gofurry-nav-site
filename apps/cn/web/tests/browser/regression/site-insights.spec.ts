import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'
const trend = (page: Page) => page.locator('[data-site-insight-trend]')
const metric = (page: Page, key: string) => page.locator(`[data-site-capability-matrix] [data-site-capability="${key}"]`)
const range = (page: Page, key: string) => page.locator(`[data-site-insight-range="${key}"]`)
const tab = (page: Page, key: string) => page.locator(`[data-site-primary-tab="${key}"]`)
async function openSite(page: Page, path = '/en/site/41?tab=insights') {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  const html = await openRuntime(page, path); await (await view).finished(); return html
}
async function ready(page: Page) { await expect(trend(page)).toHaveAttribute('data-site-insight-trend-state', 'ready'); await expect(trend(page).locator('canvas')).toBeVisible() }

for (const prefix of ['', '/en']) for (const domain of ['', '&domain=alt.example']) test('Site intelligence SSR without a trend dependency ' + prefix + domain, async ({ request, runtime }) => {
  runtime.state.fullCapabilities = true; runtime.state.intelligence.rich = true
  const response = await request.get(prefix + '/site/41?tab=insights&metric=tls13&range=90d' + domain), html = await response.text()
  expect(response.status()).toBe(200)
  expect(html).toContain('data-site-insights-workspace'); expect(html).toContain('data-site-insight-trend-state="loading"')
  expect(html.match(/data-site-capability="/g)).toHaveLength(7)
  expect(html).toContain('data-site-selected-metric="tls13"'); expect(html).toContain('data-site-trend-range="90d"')
  expect(html).toContain('data-site-insight-change')
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights', '/api/v2/nav/sites/41/recommendations'])
  runtime.assertQuiet()
})

test('Insights activation adds exactly one default trend; tab remount reuses cache', async ({ page, runtime }) => {
  await openSite(page, '/en/site/41?tab=overview'); expect(runtime.calls).toHaveLength(4); expect(runtime.count('/trend')).toBe(0)
  for (const workspace of ['security', 'observation', 'overview']) {
    await tab(page, workspace).click(); await expect(page.locator('[data-site-workspace]')).toHaveAttribute('data-site-workspace-tab', workspace)
    if (workspace === 'observation') await expect(page.locator('[data-site-performance-history-state]')).toHaveAttribute('data-site-performance-history-state', 'ready')
  }
  await settleRuntime(page); expect(runtime.calls).toHaveLength(5); expect(runtime.count('/observations')).toBe(1)
  await tab(page, 'insights').click(); await ready(page)
  expect(runtime.calls).toHaveLength(6)
  expect(runtime.calls.at(-1)!.url.pathname).toBe('/api/v2/nav/insights/metrics/ipv6/trend')
  expect(Object.fromEntries(runtime.calls.at(-1)!.url.searchParams)).toEqual({ range: '30d' })
  await tab(page, 'overview').click(); await expect(page.locator('[data-site-overview]')).toBeVisible()
  await page.goBack(); await ready(page)
  expect(runtime.calls).toHaveLength(6)
  await expect(page.locator('[data-site-insights-workspace] [role="tablist"]')).toHaveCount(0)
  runtime.assertQuiet()
})

test('metric/range are URL owned through keyboard, history, cache and reload', async ({ page, runtime }) => {
  const selected = async (metricKey: string, rangeKey: string) => {
    // An old ready surface can precede router.push; wait for the new identity first.
    await expect(page).toHaveURL(url => url.searchParams.get('metric') === metricKey && url.searchParams.get('range') === rangeKey)
    await expect(metric(page, metricKey)).toHaveAttribute('aria-pressed', 'true')
    await expect(range(page, rangeKey)).toHaveAttribute('aria-pressed', 'true')
    await expect(trend(page)).toHaveAttribute('data-site-trend-metric', metricKey)
    await expect(trend(page)).toHaveAttribute('data-site-trend-range', rangeKey)
    await ready(page)
  }
  const identities = () => runtime.calls.filter(call => call.url.pathname.endsWith('/trend')).map(call => `${call.url.pathname.split('/').at(-2)}|${call.url.searchParams.get('range')}`)
  await openSite(page, '/en/site/41?domain=alt.example&tab=insights&metric=tls13&range=90d'); await selected('tls13', '90d')
  await expect(page.locator('[data-site-insights-ecosystem]')).toHaveAttribute('href', '/en/insights/sites')
  await expect(page.locator('[data-site-insights-compare]')).toHaveAttribute('href', '/en/insights/sites/compare?ids=41')
  await metric(page, 'csp').focus(); await page.keyboard.press('Enter'); await selected('csp', '90d')
  expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ domain: 'alt.example', tab: 'insights', metric: 'csp', range: '90d' })
  await range(page, 'all').focus(); await page.keyboard.press(' '); await selected('csp', 'all')
  expect(runtime.count('/trend')).toBe(3)
  expect(identities()).toEqual(['tls13|90d', 'csp|90d', 'csp|all'])
  await page.goBack(); await selected('csp', '90d')
  await page.goBack(); await selected('tls13', '90d')
  await page.goForward(); await selected('csp', '90d')
  await page.goForward(); await selected('csp', 'all'); expect(runtime.count('/trend')).toBe(3)
  expect(runtime.count('/sites/41/detail')).toBe(1); expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/view'))
  expect((await page.reload({ waitUntil: 'domcontentloaded' }))?.status()).toBe(200); await (await view).finished(); await selected('csp', 'all')
  expect(runtime.count('/trend')).toBe(4); expect(runtime.calls).toHaveLength(12)
  expect(identities()).toEqual(['tls13|90d', 'csp|90d', 'csp|all', 'csp|all'])
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com/en/site/41')
  runtime.assertQuiet()
})

test('all seven backend states remain facts while absent facts remain missing', async ({ page, runtime }) => {
  runtime.state.fullCapabilities = true
  await openSite(page); await ready(page)
  for (const [key, state] of [['ipv6', 'supported'], ['http2', 'unsupported'], ['tls13', 'stale'], ['certificate_verified', 'not_probed'], ['hsts', 'unavailable'], ['csp', 'unknown'], ['security_txt', 'not_applicable']]) {
    await expect(metric(page, key!)).toHaveAttribute('data-site-capability-state', state!)
  }
  expect(runtime.calls).toHaveLength(5); runtime.assertQuiet()
})

test('ecosystem null differs from zero and the selected detail uses the same fact', async ({ page, runtime }) => {
  runtime.state.intelligence.nullEcosystem = true
  await openSite(page); await ready(page)
  await expect(metric(page, 'ipv6').locator('[data-site-capability-adoption]')).toHaveText('—')
  await expect(metric(page, 'ipv6').locator('[data-site-capability-coverage]')).toHaveText('0.0%')
  await expect(page.locator('[data-site-selected-adoption]')).toHaveText('—'); await expect(page.locator('[data-site-selected-coverage]')).toHaveText('0.0%')
  await expect(metric(page, 'ipv6').locator('time')).toHaveText('—'); await expect(metric(page, 'http2')).toHaveAttribute('data-site-capability-state', 'missing')
  runtime.assertQuiet()
})

for (const mode of ['ready', 'one', 'gaps', 'empty', 'unavailable'] as const) test('active trend never renders an unclassified blank surface: ' + mode, async ({ page, runtime }) => {
  runtime.state.intelligence.trend = mode
  await openSite(page)
  const state = mode === 'empty' || mode === 'unavailable' ? mode : 'ready'
  await expect(trend(page)).toHaveAttribute('data-site-insight-trend-state', state)
  if (state === 'ready') {
    await ready(page)
    if (mode === 'gaps') await expect(page.locator('[data-site-trend-no-values]')).toContainText('no observed adoption values')
  } else await expect(page.locator('[data-site-trend-message]')).toContainText(mode === 'empty' ? 'No ecosystem trend data' : 'could not be loaded')
  await expect(page.locator('[data-site-insights-workspace]')).toHaveAttribute('data-site-insights-state', 'ready')
  await expect(metric(page, 'ipv6')).toHaveAttribute('data-site-capability-state', 'unknown')
  await expect(page.locator('[data-site-insight-change]')).toHaveCount(1)
  const before = runtime.calls.length
  await tab(page, 'overview').click(); await tab(page, 'insights').click()
  await expect(trend(page)).toHaveAttribute('data-site-insight-trend-state', state)
  await settleRuntime(page); expect(runtime.calls).toHaveLength(before); runtime.assertQuiet()
})

test('empty Site Insights produces seven missing facts while ecosystem trend stays usable', async ({ page, runtime }) => {
  runtime.state.insightsEmpty = true
  await openSite(page); await ready(page)
  await expect(page.locator('[data-site-insights-workspace]')).toHaveAttribute('data-site-insights-state', 'empty')
  await expect(page.locator('[data-site-capability-matrix] [data-site-capability-state="missing"]')).toHaveCount(7)
  await expect(page.locator('[data-site-insight-changes]')).toHaveAttribute('data-site-changes-state', 'empty')
  await expect(page.locator('[data-site-insights-retry]')).toHaveCount(0)
  runtime.assertQuiet()
})

for (const workspace of ['overview', 'insights']) test('Site retry recovers P3 and P6 with one Site request from ' + workspace, async ({ page, runtime }) => {
  runtime.state.siteInsightsFailure = true
  await openSite(page); await ready(page)
  await expect(page.locator('[data-site-insights-workspace]')).toHaveAttribute('data-site-insights-state', 'unavailable')
  await expect(page.locator('[data-site-capability-matrix] [data-site-capability]')).toHaveCount(7)
  await expect(page.locator('[data-site-capability-matrix] [data-site-capability-state]')).toHaveCount(0)
  await expect(page.locator('[data-site-insight-changes]')).toHaveAttribute('data-site-changes-state', 'unavailable')
  if (workspace === 'overview') await tab(page, 'overview').click()
  runtime.state.siteInsightsFailure = false; runtime.state.fullCapabilities = true
  const before = runtime.calls.length, held = runtime.hold(url => url.pathname.endsWith('/sites/41/insights'))
  try {
    await page.locator('[data-site-insights-retry]').click(); await held.wait()
    await expect(page.locator('[data-site-insights-retry]')).toBeDisabled()
    held.release(); await held.done()
    await tab(page, 'overview').click(); await expect(page.locator('[data-site-capability-snapshot]')).toHaveAttribute('data-site-capabilities-state', 'ready')
    await expect(page.locator('[data-site-capability="ipv6"]')).toHaveAttribute('data-site-capability-state', 'supported')
    await tab(page, 'insights').click(); await ready(page)
    await expect(metric(page, 'ipv6')).toHaveAttribute('data-site-capability-state', 'supported')
    await expect(page.locator('[data-site-insights-retry]')).toHaveCount(0)
    expect(runtime.calls.slice(before).map(call => call.url.pathname)).toEqual(['/api/v2/nav/sites/41/insights'])
    runtime.assertQuiet()
  } finally { held.release() }
})

test('trend retry is isolated from Site facts and never refreshes Detail or View', async ({ page, runtime }) => {
  runtime.state.intelligence.trend = 'unavailable'
  await openSite(page); await expect(trend(page)).toHaveAttribute('data-site-insight-trend-state', 'unavailable')
  const facts = await page.locator('[data-site-capability-matrix]').textContent(), changes = await page.locator('[data-site-insight-changes]').textContent()
  const before = runtime.calls.length
  runtime.state.intelligence.trend = 'ready'
  await page.locator('[data-site-insight-trend-retry]').click(); await ready(page)
  expect(runtime.calls.slice(before).map(call => call.url.pathname)).toEqual(['/api/v2/nav/insights/metrics/ipv6/trend'])
  await expect(page.locator('[data-site-capability-matrix]')).toHaveText(facts!); await expect(page.locator('[data-site-insight-changes]')).toHaveText(changes!)
  runtime.assertQuiet()
})

test('late metric/range response only populates its cache and every active identity has a loading state', async ({ page, runtime }) => {
  await openSite(page); await ready(page)
  const held = runtime.hold(url => url.pathname.endsWith('/metrics/tls13/trend') && url.searchParams.get('range') === '30d')
  try {
    await metric(page, 'tls13').click(); await held.wait()
    await expect(trend(page)).toHaveAttribute('data-site-insight-trend-state', 'loading')
    await expect(page.locator('[data-site-trend-message]')).toContainText('Loading ecosystem adoption trend')
    await range(page, '90d').click(); await ready(page)
    held.release(); await held.done(); await settleRuntime(page)
    await expect(trend(page)).toHaveAttribute('data-site-trend-range', '90d')
    await expect(range(page, '90d')).toHaveAttribute('aria-pressed', 'true')
    await range(page, '30d').click(); await ready(page)
    expect(runtime.count('/trend')).toBe(3); expect(runtime.calls).toHaveLength(7)
    runtime.assertQuiet()
  } finally { held.release() }
})

test('Target pending and completion cannot reset Site matrix/detail/trend/changes or fetch anything but Detail', async ({ page, runtime }) => {
  runtime.state.fullCapabilities = true; runtime.state.intelligence.rich = true
  await openSite(page, '/en/site/41?tab=insights&metric=tls13&range=90d&domain=target.example'); await ready(page)
  const workspace = page.locator('[data-site-insights-workspace]'), before = await workspace.textContent(), chart = await trend(page).locator('canvas').elementHandle()
  const offset = runtime.calls.length, held = runtime.hold(url => url.pathname.endsWith('/detail') && url.searchParams.get('target') === 'alt.example')
  try {
    await page.locator('[data-site-target-trigger]').click(); await page.locator('[data-site-target-option="alt.example"]').click(); await held.wait()
    await expect(workspace).toHaveText(before!); await ready(page)
    expect(await chart!.evaluate(node => node.isConnected)).toBe(true)
    held.release(); await held.done(); await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
    await expect(workspace).toHaveText(before!); expect(await chart!.evaluate(node => node.isConnected)).toBe(true)
    expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ domain: 'alt.example', tab: 'insights', metric: 'tls13', range: '90d' })
    expect(runtime.calls.slice(offset).map(call => [call.url.pathname, call.url.searchParams.get('target')])).toEqual([['/api/v2/nav/sites/41/detail', 'alt.example']])
    expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1); expect(runtime.count('/trend')).toBe(1)
    runtime.assertQuiet()
  } finally { held.release() }
})

test('Recent Site Changes keeps every returned event, shared categories and date precision', async ({ page, runtime }) => {
  runtime.state.intelligence.rich = true
  await openSite(page); await ready(page)
  await expect(page.locator('[data-site-insight-change]')).toHaveCount(6)
  await expect(page.locator('[data-site-insight-change] time').first()).toHaveText('Sep 25, 2026, 12:34 PM UTC')
  await expect(page.locator('[data-site-insight-change] time').nth(1)).toHaveText('2026-09-24')
  expect(await page.locator('[data-site-insight-change]').evaluateAll(nodes => nodes.map(node => node.getAttribute('data-site-change-category')))).toEqual(['capability', 'capability', 'unknown', 'certificate', 'target', 'capability'])
  await expect(page.locator('[data-site-insight-changes]')).not.toContainText('Complete History')
  runtime.assertQuiet()
})

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) test(`Site intelligence responsive ${width} ${theme}`, async ({ page, context, runtime }) => {
  runtime.state.fullCapabilities = true; runtime.state.intelligence.rich = true
  await page.setViewportSize({ width, height: 900 }); await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
  await openSite(page, '/site/41?tab=insights&metric=certificate_verified&range=all'); await ready(page)
  await expect(page.locator('[data-site-capability-matrix] [data-site-capability]')).toHaveCount(7)
  await assertRuntimeSurface(page, '[data-site-insights-workspace]', theme)
  const geometry = await metric(page, 'certificate_verified').evaluate(node => ({ width: node.clientWidth, scroll: node.scrollWidth, columns: getComputedStyle(node).gridTemplateColumns.split(' ').length }))
  expect(geometry.scroll).toBeLessThanOrEqual(geometry.width + 1); expect(geometry.columns).toBe(width === 1440 ? 5 : 2)
  const chartWidth = await page.locator('[data-site-insight-chart]').evaluate(node => node.clientWidth)
  expect(chartWidth).toBeGreaterThan(200)
  await metric(page, 'hsts').focus(); await page.keyboard.press('Enter'); await ready(page)
  await expect(metric(page, 'hsts')).toHaveAttribute('aria-pressed', 'true')
  await expect(metric(page, 'hsts')).toBeFocused()
  expect(runtime.count('/trend')).toBe(2); expect(runtime.count('/sites/41/insights')).toBe(1)
  runtime.assertQuiet()
})
