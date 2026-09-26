import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'

async function openOverview(page: Page, path = '/en/site/41') {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  const html = await openRuntime(page, path)
  await (await view).finished()
  return html
}

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) {
  test(`Site Overview complete snapshot ${width} ${theme}`, async ({ page, context, runtime }) => {
    runtime.state.fullCapabilities = true
    runtime.state.manyChanges = true
    await page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    const html = await openOverview(page)
    for (const hook of ['data-site-overview-health', 'data-site-capability-snapshot', 'data-site-recent-changes']) expect(html).toContain(hook)
    await expect(page.locator('[data-site-overview-health]')).toHaveAttribute('data-site-status', 'healthy')
    await expect(page.locator('[data-site-status-distribution]')).toHaveText('2 / 2 targets healthy')
    await expect(page.locator('[data-site-overview-health] time')).toHaveText('2026-09-26 12:18:00 UTC')
    await expect(page.locator('[data-site-overview-attention]')).toHaveCount(0)
    await expect(page.locator('[data-site-capability]')).toHaveCount(7)
    for (const [key, state] of [['ipv6', 'supported'], ['http2', 'unsupported'], ['tls13', 'stale'],
      ['certificate_verified', 'not_probed'], ['hsts', 'unavailable'], ['csp', 'unknown'], ['security_txt', 'not_applicable']]) {
      await expect(page.locator(`[data-site-capability="${key}"]`)).toHaveAttribute('data-site-capability-state', state!)
    }
    await expect(page.locator('[data-site-capability="http2"] dd')).toHaveAttribute('data-tone', 'neutral')
    expect(await page.locator('[data-site-capability-group="network"] [data-site-capability]').evaluateAll(nodes => nodes.map(node => node.getAttribute('data-site-capability')))).toEqual(['ipv6', 'http2'])
    expect(await page.locator('[data-site-change]').evaluateAll(nodes => nodes.map(node => node.getAttribute('data-site-change-type')))).toEqual([
      'site.http2.enabled', 'site.tls13.enabled', 'site.hsts.added', 'site.csp.added',
    ])
    await expect(page.locator('[data-site-change] time').nth(1)).toHaveText('2026-09-23')
    await expect(page.locator('[data-site-overview]')).not.toContainText('%')
    await expect(page.locator('[data-site-overview] [data-site-protocol], [data-site-overview] [data-site-history-points], [data-site-insights]')).toHaveCount(0)
    const boxes = await page.locator('[data-site-overview-columns] > section').evaluateAll(nodes => nodes.map(node => {
      const rect = node.getBoundingClientRect()
      return { x: rect.x, y: rect.y, width: rect.width, bottom: rect.bottom }
    }))
    if (width === 1440) {
      expect(boxes[0]!.y).toBe(boxes[1]!.y)
      expect(boxes[0]!.width / boxes[1]!.width).toBeCloseTo(1.5, 1)
    } else {
      expect(boxes[1]!.y).toBeGreaterThan(boxes[0]!.bottom)
      expect(boxes[0]!.width).toBeCloseTo(boxes[1]!.width, 0)
    }
    await assertRuntimeSurface(page, '[data-site-overview]', theme)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([
      '/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights', '/api/v2/nav/sites/41/view',
    ])
    runtime.assertQuiet()
  })
}

for (const scenario of ['stale', 'missing', 'unknown', 'zero', 'mixed'] as const) {
  test('Site summary distinguishes ' + scenario, async ({ page, runtime }) => {
    if (scenario === 'missing') runtime.state.missingSummary = true
    else runtime.state.summaryScenario = scenario
    if (scenario === 'mixed') runtime.state.extraTargets = ['fallback.example']
    await openOverview(page)
    const health = page.locator('[data-site-overview-health]')
    const attention = page.locator('[data-site-overview-attention]')
    await expect(health).toHaveAttribute('data-site-summary-state', scenario === 'missing' ? 'missing' : scenario === 'stale' ? 'stale' : 'ready')
    if (scenario === 'stale') {
      await expect(health).toContainText('Healthy')
      await expect(health).toContainText('Summary may be stale')
    } else if (scenario === 'missing') {
      await expect(health).toContainText('No complete site health summary')
      await expect(health).not.toHaveAttribute('data-site-status')
      await expect(page.locator('[data-site-health="status"]')).toContainText('Healthy')
    } else if (scenario === 'unknown') {
      await expect(health).toHaveAttribute('data-site-status', 'unknown')
      await expect(page.locator('[data-site-status-distribution]')).toHaveText('2 Unknown')
      await expect(attention.locator('[data-site-attention-target]')).toHaveCount(0)
    } else if (scenario === 'zero') await expect(health).toContainText('No collected targets')
    else {
      await expect(page.locator('[data-site-status-distribution]')).toHaveText('1 Healthy · 2 Down')
      await expect(attention.locator('li')).toHaveText(['alt.example is not responding.', 'fallback.example · Down'])
      await expect(attention).not.toContainText('raw_failure_code')
    }
    await expect(attention).toHaveCount(scenario === 'zero' ? 0 : 1)
    expect(runtime.calls).toHaveLength(3)
    runtime.assertQuiet()
  })
}

for (const scenario of ['empty', 'unavailable', 'view-failure'] as const) {
  test('Overview optional slice boundary ' + scenario, async ({ page, runtime }) => {
    runtime.state.insightsEmpty = scenario === 'empty'
    runtime.state.siteInsightsFailure = scenario === 'unavailable'
    runtime.state.viewFailure = scenario === 'view-failure'
    const html = await openOverview(page)
    expect(html).toContain('data-site-overview-health')
    await expect(page.locator('[data-site-overview-health]')).toHaveAttribute('data-site-status', 'healthy')
    await expect(page.locator('[data-site-capability]')).toHaveCount(7)
    if (scenario !== 'view-failure') {
      await expect(page.locator('[data-site-capability-snapshot]')).toHaveAttribute('data-site-capabilities-state', scenario)
      await expect(page.locator(`[data-site-capability-state="${scenario === 'empty' ? 'missing' : 'unavailable'}"]`)).toHaveCount(7)
      await expect(page.locator('[data-site-recent-changes]')).toHaveAttribute('data-site-changes-state', scenario)
      await expect(page.locator('[data-site-recent-changes]')).toContainText(scenario === 'empty' ? 'No recent changes' : 'temporarily unavailable')
      await expect(page.locator('[data-site-change]')).toHaveCount(0)
    } else {
      await expect(page.locator('[data-site-capability="ipv6"]')).toHaveAttribute('data-site-capability-state', 'unknown')
      await expect(page.locator('[data-site-capability="http2"]')).toHaveAttribute('data-site-capability-state', 'missing')
      await expect(page.locator('[data-site-recent-changes]')).toHaveAttribute('data-site-changes-state', 'ready')
    }
    expect(runtime.count('/sites/41/detail')).toBe(1)
    expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.count('/sites/41/insights')).toBe(scenario === 'unavailable' ? 2 : 1)
    runtime.assertQuiet()
  })
}

test('Site snapshot survives pending Target switch, Overview remount and history without extra requests', async ({ page, runtime }) => {
  runtime.state.summaryChangesOnTarget = true
  runtime.state.fullCapabilities = true
  runtime.state.manyChanges = true
  await openOverview(page)
  const overview = page.locator('[data-site-overview]')
  const node = await overview.elementHandle()
  const before = await overview.textContent()
  const held = runtime.hold(url => url.pathname.endsWith('/sites/41/detail') && url.searchParams.get('target') === 'alt.example')
  try {
    await page.locator('[data-site-target-trigger]').click()
    await page.locator('[data-site-target-option="alt.example"]').click()
    await held.wait()
    await expect(page.locator('[data-site-target-pending]')).toBeVisible()
    await expect(overview).toHaveText(before!)
    expect(await node!.evaluate(element => element.isConnected)).toBe(true)
    held.release()
    await held.done()
    await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
    await expect(page.locator('[data-site-health="status"]')).toContainText('Warning')
    await expect(overview).toHaveText(before!)
    await page.locator('[data-site-overview-insights]').click()
    await expect(page.locator('[data-site-insights]')).toBeVisible()
    await expect(overview).toHaveCount(0)
    expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ domain: 'alt.example', tab: 'insights' })
    await page.goBack()
    await expect(overview).toHaveText(before!)
    await page.goForward()
    await expect(page.locator('[data-site-insights]')).toBeVisible()
    await page.locator('[data-site-primary-tab="overview"]').click()
    await expect(overview).toHaveText(before!)
    await settleRuntime(page)
    expect(runtime.count('/sites/41/detail')).toBe(2)
    expect(runtime.count('/sites/41/insights')).toBe(1)
    expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.calls).toHaveLength(4)
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com/en/site/41')
    // A new page session adopts the new Site summary; client Target switching did not.
    const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
    expect((await page.reload({ waitUntil: 'domcontentloaded' }))?.status()).toBe(200)
    await (await view).finished()
    await expect(page.locator('[data-site-overview-health]')).toHaveAttribute('data-site-status', 'degraded')
    await expect(page.locator('[data-site-overview-health] time')).toHaveText('2026-09-26 13:00:00 UTC')
    expect(runtime.calls).toHaveLength(7)
    runtime.assertQuiet()
  } finally { held.release() }
})

test.describe('cross-timezone SSR', () => {
  test.use({ timezoneId: 'America/Los_Angeles' })
  test('day precision stays a date and precise events hydrate in UTC', async ({ page, runtime }) => {
    runtime.state.manyChanges = true
    await openOverview(page)
    await expect(page.locator('[data-site-change] time').first()).toHaveText('Sep 24, 2026, 12:34 PM UTC')
    await expect(page.locator('[data-site-change] time').nth(1)).toHaveText('2026-09-23')
    runtime.assertQuiet()
  })
})
