import type { Page } from '@playwright/test'
import { settleRuntime } from '../fixtures/insights-runtime'
import { test, expect, workspaceMetricKeys, workspaceRegionKeys, openRuntime, revealImages, keyboardFocus } from '../fixtures/insights-workspace'
const routes = [
  ['/insights/games/players?metric=latest_observed', 'players', 1],
  ['/en/insights/games/players?metric=average_30d', 'players', 1],
  ['/insights/games/prices?region=CN', 'prices', 2],
  ['/en/insights/games/prices?region=US', 'prices', 2],
  ['/insights/games/languages', 'languages', 1], ['/en/insights/games/languages', 'languages', 1],
  ['/insights/sites/certificates', 'certificates', 1], ['/en/insights/sites/certificates', 'certificates', 1],
] as const
async function layout(page: Page) {
  expect(await page.evaluate(() => ({
    page: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) <= innerWidth + 1,
    rows: [...document.querySelectorAll('[data-workspace-entity], .insights-workspace-summary, .insights-workspace-header')].every(el => el.scrollWidth <= el.clientWidth + 1),
  }))).toEqual({ page: true, rows: true })
}
for (const [route, kind, count] of routes) for (const width of [1440, 1024, 390]) {
  test('Workspace SSR/hydration, identity and layout ' + route + ' ' + width, async ({ page, request, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    const ssr = await request.get(route)
    expect(ssr.status()).toBe(200)
    expect(await ssr.text()).toMatch(/<h1>[^<]+<\/h1>/)
    expect(runtime.calls).toHaveLength(count)
    const offset = runtime.calls.length
    const html = await openRuntime(page, route)
    expect(html).toContain('insights-domain-nav')
    if (kind !== 'languages') {
      expect(html).toContain('href="' + (route.startsWith('/en/') ? '/en' : '') + (kind === 'certificates' ? '/site/201' : '/games/101') + '"')
      expect(html).toContain(kind === 'certificates' ? '/nav/sites/201/icon/' : '/media/game-0.svg')
      expect(html).toContain(kind === 'certificates' ? '/defaultLogo.svg' : 'data-media-state="fallback"')
    }
    if (kind === 'prices') expect(html).toContain(route.startsWith('/en/') ? '$5.99' : '¥5.99')
    if (kind === 'players') { expect(html).toContain('189'); expect(html).toContain('213'); expect(html).toContain('88.7%') }
    await expect(page.locator('.insights-domain-nav [aria-current="page"]')).toHaveCount(1)
    await expect(page.locator('.insight-workspace-disclosure')).not.toHaveAttribute('open')
    if (kind === 'players' || kind === 'prices') {
      expect(await page.locator('[data-workspace-option]').evaluateAll(els => els.map(el => (el as HTMLElement).dataset.workspaceOption)))
        .toEqual(kind === 'players' ? workspaceMetricKeys : workspaceRegionKeys)
      await expect(page.locator('[data-workspace-option][aria-pressed="true"]')).toHaveCount(1)
      await keyboardFocus(page.locator('[data-workspace-option]').last())
      expect(await page.locator('[data-workspace-option]').last().evaluate(el => {
        const box = el.getBoundingClientRect(), parent = el.parentElement!.getBoundingClientRect()
        return box.left >= parent.left - 1 && box.right <= parent.right + 1
      })).toBe(true)
    }
    if (kind === 'players') {
      await expect(page.locator('[data-rank]')).toHaveCount(20)
      await expect(page.locator('.insight-ranking-row--lead')).toHaveCount(3)
      await expect(page.locator('[data-rank="20"]')).toContainText('0')
    }
    if (kind === 'languages') {
      await expect(page.locator('[data-language]')).toHaveCount(12)
      await expect(page.locator('[data-workspace-raw]')).not.toHaveAttribute('open')
      expect(await page.locator('[data-language] progress').evaluateAll(els => els.every(el => el.max === 1 && el.value >= 0 && el.value <= 1))).toBe(true)
      await expect(page.locator('[data-language="th"] progress')).toHaveCount(0)
      await page.locator('[data-workspace-raw] summary').click()
      await expect(page.locator('[data-workspace-raw] tbody tr')).toHaveCount(14)
      if (width === 390) expect(await page.locator('.insights-workspace-table-scroll').evaluate(el => el.scrollWidth > el.clientWidth)).toBe(true)
    }
    if (kind === 'certificates') {
      await expect(page.locator('[data-expiry]')).toHaveCount(4)
      await expect(page.locator('[data-expiry-attention] [data-workspace-entity]')).toHaveCount(3)
      await expect(page.locator('[data-verification-issues] [data-workspace-entity]')).toHaveCount(2)
    }
    await revealImages(page); await layout(page)
    expect(runtime.calls.slice(offset)).toHaveLength(count); runtime.assertQuiet()
  })
}
for (const [kind, queryKey, keys, endpoint] of [
  ['players', 'metric', workspaceMetricKeys, '/players/ranking'],
  ['prices', 'region', workspaceRegionKeys, '/prices/discounts'],
] as const) {
  test('Workspace options and locale preserve query ' + kind, async ({ page, runtime }) => {
    await openRuntime(page, '/insights/games/' + kind + '?' + queryKey + '=invalid')
    await expect(page.locator('[data-workspace-option][aria-pressed="true"]')).toHaveAttribute('data-workspace-option', keys[0]!)
    for (const key of keys.slice(1)) {
      const response = page.waitForResponse(res => new URL(res.url()).pathname.endsWith(endpoint) && new URL(res.url()).searchParams.get(queryKey) === key)
      await page.locator('[data-workspace-option="' + key + '"]').click(); await response
      await expect(page).toHaveURL(url => url.searchParams.get(queryKey) === key)
    }
    // Locale may reuse the same locale-independent async-data cache.
    await page.locator('button:has(img[alt="EN"])').first().click()
    await expect(page).toHaveURL(url => url.pathname === '/en/insights/games/' + kind && url.searchParams.get(queryKey) === keys.at(-1))
    if (kind === 'players') await expect(page.locator('.insights-container')).toContainText('180')
    runtime.assertQuiet()
  })
}
for (const failure of ['/prices/overview', '/prices/discounts']) {
  test('Workspace independent Price failure ' + failure, async ({ page, runtime }) => {
    runtime.state.failure = failure; await page.setViewportSize({ width: 390, height: 1000 })
    await openRuntime(page, '/en/insights/games/prices?region=US')
    if (failure.endsWith('overview')) {
      await expect(page.locator('[data-price-overview-error]')).toBeVisible()
      await expect(page.locator('.insight-discount-row')).toHaveCount(6)
    } else {
      await expect(page.locator('[data-discount-error]')).toBeVisible()
      await expect(page.locator('.insights-workspace-summary')).toBeVisible()
    }
    await layout(page); runtime.assertQuiet()
  })
}
for (const [kind, failure, route] of [
  ['players', '/players/ranking', '/insights/games/players'],
  ['languages', '/languages/overview', '/insights/games/languages'],
  ['certificates', '/certificates/overview', '/insights/sites/certificates'],
]) test('Workspace failure ' + kind, async ({ page, runtime }) => {
  runtime.state.failure = failure!; await page.setViewportSize({ width: 390, height: 1000 })
  await openRuntime(page, route!)
  await expect(page.locator('main .insights-empty-state')).toBeVisible()
  await layout(page); runtime.assertQuiet()
})
for (const [route, kind] of routes.filter((_, index) => index % 2 === 0)) {
  test('Workspace empty ' + kind, async ({ page, runtime }) => {
    runtime.state.empty = true
    await openRuntime(page, route)
    await expect(page.locator('main .insights-empty-state').first()).toBeVisible()
    await expect(page.locator('[data-workspace-entity], [data-language]')).toHaveCount(0)
    runtime.assertQuiet()
  })
  if (kind !== 'languages') test('Workspace real image fallback ' + kind, async ({ page, runtime }) => {
    runtime.failImages = true; await page.setViewportSize({ width: 390, height: 1000 })
    await openRuntime(page, route); await revealImages(page)
    await expect(page.locator('.insight-entity-media img')).toHaveCount(0)
    expect(await page.locator('.insight-entity-media [role="img"][aria-label]').count()).toBeGreaterThan(0)
    await layout(page); runtime.assertQuiet()
  })
}
for (const [route, kind] of routes.filter(([route]) => route.startsWith('/en/'))) for (const width of [1440, 1024, 390]) {
  test('Workspace Dark persistence and layout ' + kind + ' ' + width, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 1000 }); await openRuntime(page, route)
    await page.getByRole('button', { name: 'Toggle theme icon', exact: true }).click()
    await expect(page.locator('html')).toHaveClass(/dark/)
    await revealImages(page); await settleRuntime(page)
    await page.reload({ waitUntil: 'load' })
    await expect(page.locator('html')).toHaveClass(/dark/)
    await layout(page); runtime.assertQuiet()
  })
}
