import type { Page } from '@playwright/test'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'

async function openOverview(page: Page, path = '/en/site/41?tab=overview') {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  const html = await openRuntime(page, path)
  await (await view).finished()
  return html
}

async function reviewOverview(page: Page, name: string) {
  const directory = process.env.GOFURRY_OVERVIEW_REVIEW_DIR
  if (!directory) return
  await mkdir(directory, { recursive: true })
  await settleRuntime(page)
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.screenshot({ path: join(directory, name + '.png'), fullPage: true })
}

async function assertOverviewComposition(page: Page, width: number) {
  const overview = page.locator('[data-site-overview]')
  await expect(overview.locator(':scope > *')).toHaveCount(2)
  await expect(overview.locator('[data-site-overview-columns]')).toHaveCount(0)
  await expect(overview.locator('[data-site-overview-status-composite]')).toHaveCount(1)
  await expect(overview.locator('[data-site-overview-capability-composite]')).toHaveCount(1)
  const layout = await overview.evaluate(root => {
    const box = (node: Element) => {
      const rect = node.getBoundingClientRect(), style = getComputedStyle(node)
      return { x: rect.x, y: rect.y, width: rect.width, bottom: rect.bottom,
        border: style.borderTopWidth, radius: style.borderTopLeftRadius, shadow: style.boxShadow }
    }
    return {
      sections: Array.from(root.children).map(box),
      groups: Array.from(root.querySelectorAll('[data-site-capability-group]')).map(node => ({ ...box(node),
        fill: getComputedStyle(node).backgroundColor, bottomBorder: getComputedStyle(node).borderBottomWidth })),
      cards: Array.from(root.querySelectorAll('[data-site-change]')).map(box),
      rowBorders: Array.from(root.querySelectorAll('[data-site-capability]')).map(node => getComputedStyle(node).borderBottomWidth),
      statusColumns: getComputedStyle(root.querySelector('[data-site-overview-status-grid]')!).gridTemplateColumns.split(' ').length,
    }
  })
  for (let index = 1; index < layout.sections.length; index++) {
    expect(layout.sections[index]!.y).toBeGreaterThan(layout.sections[index - 1]!.bottom)
    expect(layout.sections[index]!.width).toBeCloseTo(layout.sections[0]!.width, 0)
    expect(layout.sections[index]!.x).toBeCloseTo(layout.sections[0]!.x, 0)
  }
  expect(layout.statusColumns).toBe(width >= 768 ? 3 : 2)
  expect(layout.groups).toHaveLength(3)
  expect(layout.groups.every(group => group.fill === 'rgba(0, 0, 0, 0)' && group.bottomBorder === '0px' && group.shadow === 'none')).toBe(true)
  expect(layout.rowBorders.every(border => border === '0px')).toBe(true)
  if (width >= 768) expect(new Set(layout.groups.map(group => group.y)).size).toBe(1)
  else expect(layout.groups[1]!.y).toBeGreaterThan(layout.groups[0]!.bottom)
  expect(layout.cards).toHaveLength(4)
  expect(layout.cards.every(card => card.border === '0px' && card.radius === '7px' && card.shadow === 'none')).toBe(true)
  const firstRow = layout.cards.filter(card => card.y === layout.cards[0]!.y)
  expect(firstRow).toHaveLength(width >= 768 ? 3 : 1)
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
    await expect(page.locator('[data-site-status-distribution]')).toHaveCount(0)
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
    await expect(page.locator('[data-site-overview-target-count]')).toHaveText('2 targets')
    await assertOverviewComposition(page, width)
    await assertRuntimeSurface(page, '[data-site-overview]', theme)
    await reviewOverview(page, `healthy-${width}-${theme}`)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([
      '/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights', '/api/v2/nav/sites/41/recommendations', '/api/v2/nav/sites/41/view',
    ])
    runtime.assertQuiet()
  })
}

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) {
  test(`Overview embeds Attention within its status composite ${width} ${theme}`, async ({ page, context, runtime }) => {
    runtime.state.summaryScenario = 'mixed'; runtime.state.extraTargets = ['third.example']
    runtime.state.fullCapabilities = true; runtime.state.intelligence.rich = true
    await page.setViewportSize({ width, height: 900 })
    await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await openOverview(page)
    const status = page.locator('[data-site-overview-status-composite]')
    const attention = status.locator('[data-site-overview-attention]')
    await expect(attention).toHaveCount(1)
    await expect(attention.locator('li')).toHaveText(['2 targets · HTTP is currently unreachable'])
    await expect(attention).not.toContainText('后端旧中文')
    const borders = await attention.evaluate(node => {
      const style = getComputedStyle(node)
      return [style.borderTopWidth, style.borderRightWidth, style.borderBottomWidth, style.borderLeftWidth, style.boxShadow]
    })
    expect(borders).toEqual(['0px', '0px', '0px', '0px', 'none'])
    await expect(status.locator('[data-site-overview-target-count]')).toHaveText('3 targets')
    await expect(page.locator('[data-site-change][data-site-change-category="certificate"]')).toHaveCount(1)
    await assertOverviewComposition(page, width)
    await assertRuntimeSurface(page, '[data-site-overview]', theme)
    await reviewOverview(page, `attention-${width}-${theme}`)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
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
      await expect(attention).toBeVisible()
    } else if (scenario === 'missing') {
      await expect(health).toContainText('No complete site health summary')
      await expect(health).not.toHaveAttribute('data-site-status')
      await expect(page.locator('[data-site-health="status"]')).toContainText('Healthy')
    } else if (scenario === 'unknown') {
      await expect(health).toHaveAttribute('data-site-status', 'unknown')
      await expect(page.locator('[data-site-status-distribution]')).toHaveCount(0)
      await expect(attention.locator('[data-site-attention-target]')).toHaveCount(0)
    } else if (scenario === 'zero') await expect(page.locator('[data-site-overview-target-count]')).toHaveText('0 targets')
    else {
      await expect(page.locator('[data-site-status-distribution]')).toHaveCount(0)
      await expect(attention.locator('li')).toHaveText(['2 targets · HTTP is currently unreachable'])
      await expect(attention).not.toContainText('后端旧中文')
    }
    await expect(attention).toHaveCount(scenario === 'zero' ? 0 : 1)
    expect(runtime.calls).toHaveLength(4)
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
      if (scenario === 'empty') await expect(page.locator('[data-site-capability-state="missing"]')).toHaveCount(7)
      else await expect(page.locator('[data-site-capability][data-site-capability-state]')).toHaveCount(0)
      await expect(page.locator('[data-site-recent-changes]')).toHaveAttribute('data-site-changes-state', scenario)
      await expect(page.locator('[data-site-recent-changes]')).toContainText(scenario === 'empty' ? 'No recent changes' : 'temporarily unavailable')
      await expect(page.locator('[data-site-change]')).toHaveCount(0)
      const empty = await page.locator('[data-site-recent-changes]').boundingBox()
      expect(empty!.height).toBeLessThan(100)
      await expect(page.locator('[data-site-overview-change-grid]')).toHaveCount(0)
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

test('Site snapshot survives pending Target switch, Overview remount and history without extra Site requests', async ({ page, runtime }) => {
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
    await expect(page.locator('[data-site-insight-trend]')).toHaveAttribute('data-site-insight-trend-state', 'ready')
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
    expect(runtime.calls).toHaveLength(6)
    expect(runtime.count('/trend')).toBe(1)
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com/en/site/41')
    // A new page session adopts the new Site summary; client Target switching did not.
    const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
    expect((await page.reload({ waitUntil: 'domcontentloaded' }))?.status()).toBe(200)
    await (await view).finished()
    await expect(page.locator('[data-site-overview-health]')).toHaveAttribute('data-site-status', 'degraded')
    await expect(page.locator('[data-site-overview-health] time')).toHaveText('2026-09-26 13:00:00 UTC')
    expect(runtime.calls).toHaveLength(10)
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
