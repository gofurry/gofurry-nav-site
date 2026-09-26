import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'
const tab = (page: Page, view: string) => page.locator(`[data-site-observation-tab="${view}"]`)
const history = (page: Page) => page.locator('[data-site-performance-history-state]')
async function openObservation(page: Page, view = 'overview') {
  const counted = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  const html = await openRuntime(page, '/en/site/41?tab=observation' + (view === 'overview' ? '' : '&view=' + view))
  await (await counted).finished()
  return html
}
async function selectTarget(page: Page, target: string) {
  await page.locator('[data-site-target-trigger]').click()
  await page.locator(`[data-site-target-option="${target}"]`).click()
  await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', target)
}
async function activeView(page: Page, view: string) {
  await expect(tab(page, view)).toHaveAttribute('aria-selected', 'true')
  await expect(tab(page, view)).toHaveAttribute('tabindex', '0')
  await expect(page.locator('[data-site-observation-tab][tabindex="0"]')).toHaveCount(1)
  await expect(page.locator('[data-site-observation-view]')).toHaveAttribute('data-site-observation-view', view)
}

for (const view of ['overview', 'performance', 'http', 'dns', 'web']) test('Observation SSR evidence without history: ' + view, async ({ request, runtime }) => {
  runtime.state.observationRich = true
  const response = await request.get('/en/site/41?tab=observation&view=' + view)
  expect(response.status()).toBe(200)
  const html = await response.text()
  expect(html).toContain(`data-site-observation-view="${view}"`)
  if (view === 'performance') expect(html).toContain('data-site-performance-history-state="loading"')
  if (view === 'web') expect(html).toContain('data-site-web-metadata')
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights'])
  runtime.assertQuiet()
})

test('secondary navigation owns history/reload/keyboard and omits the default view', async ({ page, runtime }) => {
  runtime.state.observationRich = true
  await openObservation(page, 'http')
  await tab(page, 'dns').click(); await activeView(page, 'dns')
  await page.goBack(); await activeView(page, 'http')
  await page.goForward(); await activeView(page, 'dns')
  await tab(page, 'http').click(); await tab(page, 'http').focus()
  for (const [key, view] of [['ArrowRight', 'dns'], ['End', 'web'], ['Home', 'overview'], ['ArrowLeft', 'web']]) {
    await page.keyboard.press(key!); await activeView(page, view!); await expect(tab(page, view!)).toBeFocused()
    if (view === 'overview') expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ tab: 'observation' })
  }
  expect(runtime.calls).toHaveLength(3)
  const counted = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  expect((await page.reload({ waitUntil: 'domcontentloaded' }))?.status()).toBe(200)
  await (await counted).finished(); await activeView(page, 'web')
  expect(runtime.calls).toHaveLength(6)
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com/en/site/41')
  runtime.assertQuiet()
})

test('default Performance sample auto-loads exactly once, slices locally and survives view and Target cache reuse', async ({ page, runtime }) => {
  runtime.state.historyCount = 100; runtime.state.observationRich = true
  await openObservation(page)
  expect(runtime.count('/observations')).toBe(0)
  const held = runtime.hold(url => url.pathname.endsWith('/observations'))
  try {
    await tab(page, 'performance').click(); await held.wait()
    await expect(page.locator('[data-site-performance-sample="20"]')).toHaveAttribute('aria-pressed', 'true')
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'loading')
    await expect(page.locator('[data-site-performance-chart]')).toHaveCount(0)
    held.release(); await held.done()
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'ready')
    await expect(history(page)).toHaveAttribute('data-site-performance-points', '20')
    await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
    for (const sample of [60, 100]) {
      await page.locator(`[data-site-performance-sample="${sample}"]`).click()
      await expect(history(page)).toHaveAttribute('data-site-performance-points', String(sample))
    }
    await tab(page, 'http').click(); await expect(page.locator('[data-site-performance-chart]')).toHaveCount(0)
    await tab(page, 'performance').click(); await expect(history(page)).toHaveAttribute('data-site-performance-points', '100')
    await page.locator('[data-site-primary-tab="overview"]').click()
    const siteSnapshot = await page.locator('[data-site-overview]').textContent()
    await page.goBack(); await activeView(page, 'performance')
    expect(runtime.count('/observations')).toBe(1)
    expect(Object.fromEntries(runtime.calls.find(call => call.url.pathname.endsWith('/observations'))!.url.searchParams)).toEqual({ protocol: 'ping', limit: '100', payload_mode: 'preview' })
    await selectTarget(page, 'alt.example')
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'ready')
    await expect(history(page)).toHaveAttribute('data-site-performance-points', '20')
    await selectTarget(page, 'target.example')
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'ready')
    await page.locator('[data-site-primary-tab="overview"]').click()
    await expect(page.locator('[data-site-overview]')).toHaveText(siteSnapshot!)
    await settleRuntime(page)
    expect(runtime.count('/sites/41/detail')).toBe(3)
    expect(runtime.count('/observations')).toBe(2)
    expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.calls).toHaveLength(7)
    runtime.assertQuiet()
  } finally { held.release() }
})

for (const state of ['empty', 'unavailable', 'no-rtt']) test('Performance classifies ' + state + ' instead of a blank chart', async ({ page, runtime }) => {
  runtime.state.historyCount = state === 'empty' ? 0 : 3
  runtime.state.historyFailure = state === 'unavailable'; runtime.state.historyNoRtt = state === 'no-rtt'
  await openObservation(page, 'performance')
  await expect(history(page)).toHaveAttribute('data-site-performance-history-state', state === 'no-rtt' ? 'ready' : state)
  await expect(page.locator('[data-site-performance-chart]')).toHaveCount(0)
  await expect(page.locator('[data-site-performance]')).toContainText(state === 'empty' ? 'No Ping observations yet.' : state === 'unavailable' ? 'temporarily unavailable' : 'no RTT measurements')
  await tab(page, 'http').click(); await tab(page, 'performance').click()
  expect(runtime.count('/observations')).toBe(state === 'unavailable' ? 2 : 1) // Nitro GET retry on injected 503 only.
  if (state === 'unavailable') {
    runtime.state.historyFailure = false
    await page.locator('[data-site-history-retry]').click()
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'ready')
    expect(runtime.count('/observations')).toBe(3)
  }
  expect(runtime.count('/sites/41/detail')).toBe(1)
  expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  runtime.assertQuiet()
})

test('late A history cannot replace B and completed A is reusable from cache', async ({ page, runtime }) => {
  runtime.state.historyCount = 80
  const held = runtime.hold(url => url.pathname.endsWith('/targets/target.example/observations'))
  try {
    await openObservation(page, 'performance'); await held.wait()
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'loading')
    runtime.state.historyCount = 3
    await selectTarget(page, 'alt.example')
    await expect(history(page)).toHaveAttribute('data-site-performance-points', '3')
    held.release(); await held.done(); await settleRuntime(page)
    await expect(history(page)).toHaveAttribute('data-site-performance-points', '3')
    await page.locator('[data-site-history-table] summary').click()
    await expect(page.locator('[data-site-history-table] tbody tr').first()).toContainText('70 ms')
    await selectTarget(page, 'target.example')
    await expect(history(page)).toHaveAttribute('data-site-performance-points', '20')
    expect(runtime.count('/observations')).toBe(2)
    expect(runtime.count('/sites/41/detail')).toBe(3)
    expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.calls.every(call => call.completed)).toBe(true)
    runtime.assertQuiet()
  } finally { held.release() }
})

for (const slice of ['insights', 'view']) test('Observation is independent of optional ' + slice + ' failure', async ({ page, runtime }) => {
  runtime.state.siteInsightsFailure = slice === 'insights'; runtime.state.viewFailure = slice === 'view'
  runtime.state.observationRich = true
  await openObservation(page, 'performance')
  await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'ready')
  await expect(page.locator('[data-site-performance-kpi="response"]')).toContainText('120 ms')
  expect(runtime.count('/sites/41/detail')).toBe(1); expect(runtime.count('/observations')).toBe(1)
  runtime.assertQuiet()
})

for (const view of ['overview', 'http', 'dns', 'web']) test('non-Performance Target switch only refetches Detail: ' + view, async ({ page, runtime }) => {
  runtime.state.observationRich = true
  await openObservation(page, view)
  await selectTarget(page, 'alt.example')
  await activeView(page, view)
  await expect(page.locator('[data-site-observation]')).toHaveAttribute('data-site-observation-target', 'alt.example')
  await expect(page.locator('[data-site-health="http"]')).toContainText('HTTP 201')
  if (view === 'http' || view === 'overview') await expect(page.locator('[data-site-observation]')).toContainText('https://alt.example/')
  if (view === 'dns') await expect(page.locator('[data-site-dns]')).toContainText('203.0.113.42')
  if (view === 'web') await expect(page.locator('[data-site-web-metadata]')).toContainText('alt.example page')
  expect(runtime.calls).toHaveLength(4)
  expect(runtime.count('/sites/41/detail')).toBe(2); expect(runtime.count('/observations')).toBe(0)
  expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  runtime.assertQuiet()
})

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) {
  test(`HTTP DNS Web evidence and disclosure ${width} ${theme}`, async ({ page, context, runtime }) => {
    runtime.state.observationRich = true
    runtime.state.longEvidence = true
    await page.setViewportSize({ width, height: 900 })
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    await openObservation(page)
    await expect(page.locator('[data-site-observation-protocol]')).toHaveCount(3)
    await expect(page.locator('[data-site-observation-protocol="dns"]')).toContainText('Success')
    await expect(page.locator('[data-site-observation-protocol="dns"] [data-site-protocol-freshness]')).toContainText('Stale')
    await expect(page.locator('[data-site-observation-risks]')).toContainText('Observed DNS evidence needs review.')
    await expect(page.locator('[data-site-observation-risks]')).not.toContainText('raw_code_hidden')
    await tab(page, 'http').click()
    await expect(page.locator('[data-site-http] [data-site-evidence="bytes"]')).toContainText('0 B')
    await expect(page.locator('[data-site-http-redirects] li')).toHaveCount(2)
    const full = page.locator('[data-site-http-all-headers]')
    await expect(full).not.toHaveAttribute('open')
    await full.locator('summary').focus(); await page.keyboard.press('Enter')
    await expect(full).toHaveAttribute('open', '')
    await expect(full).toContainText('full-header-value')
    await expect(full).toContainText('strict-transport-security') // Raw header evidence, no policy interpretation.
    await assertRuntimeSurface(page, '[data-site-http]', theme)
    await tab(page, 'dns').click()
    await expect(page.locator('[data-site-dns-chain]')).toContainText('edge.example')
    await expect(page.locator('[data-site-dns-group="AAAA"]')).toHaveCount(0)
    for (const group of ['A', 'CNAME', 'MX', 'NS', 'TXT', 'CAA', 'SOA']) await expect(page.locator(`[data-site-dns-group="${group}"]`)).toBeVisible()
    const details = page.locator('[data-site-dns-group="A"] details').first()
    await details.locator('summary').click(); await expect(details).toContainText('AS64496')
    await expect(details).toContainText('Fixture ISP')
    await assertRuntimeSurface(page, '[data-site-dns]', theme)
    await tab(page, 'web').click()
    await expect(page.locator('[data-site-web-metadata]')).toContainText('UTF-8')
    await expect(page.locator('[data-site-web-probe]')).toHaveCount(4)
    for (const value of ['Community guide', 'Fixture app', 'Fixture registrar', 'Global Disallow', 'DNSSEC delegation']) await expect(page.locator('[data-site-web]')).toContainText(value)
    await expect(page.locator('[data-site-web]')).not.toContainText('SECURITY_ONLY')
    await expect(page.locator('[data-site-web]')).not.toContainText('security.txt')
    await assertRuntimeSurface(page, '[data-site-web]', theme)
    expect(runtime.calls).toHaveLength(3)
    runtime.assertQuiet()
  })
  test(`Performance chart and waterfall ${width} ${theme}`, async ({ page, context, runtime }) => {
    runtime.state.observationRich = true; runtime.state.historyCount = 100
    await page.setViewportSize({ width, height: 900 })
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    await openObservation(page, 'performance')
    await expect(history(page)).toHaveAttribute('data-site-performance-history-state', 'ready')
    const chart = page.locator('[data-site-performance-chart]')
    await expect(chart).toHaveAttribute('data-site-chart-ready', 'true')
    await expect(chart.locator('canvas')).toBeVisible()
    expect((await chart.boundingBox())!.height).toBeLessThan(300)
    await expect(page.locator('[data-site-timing]')).toHaveCount(6)
    await expect(page.locator('[data-site-timing="total"]')).toContainText('120 ms')
    await expect(page.locator('[data-site-performance-kpi="jitter"]')).toContainText('0 ms')
    await page.locator('[data-site-history-table] summary').click()
    await expect(page.locator('[data-site-history-table] tbody tr')).toHaveCount(20)
    await expect(page.locator('[data-site-history-table] tbody tr').nth(1)).toContainText('10%')
    await assertRuntimeSurface(page, '[data-site-performance]', theme)
    expect(runtime.calls).toHaveLength(4)
    runtime.assertQuiet()
  })
}

test('no redirect and no CNAME do not invent chains or history requests', async ({ page, runtime }) => {
  runtime.state.observationRich = true; runtime.state.noRedirects = true; runtime.state.noCname = true
  await openObservation(page, 'http')
  await expect(page.locator('[data-site-http-redirects]')).toHaveCount(0)
  await tab(page, 'dns').click()
  await expect(page.locator('[data-site-dns-chain], [data-site-dns-group="CNAME"]')).toHaveCount(0)
  expect(runtime.calls).toHaveLength(3)
  runtime.assertQuiet()
})
