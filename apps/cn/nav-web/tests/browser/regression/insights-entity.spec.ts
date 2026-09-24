import { test, expect, openRuntime, revealImages } from '../fixtures/insights-entity'

for (const prefix of ['', '/en']) test('Entity SSR keeps Site-level Insights separate from target observations ' + (prefix || 'zh'), async ({ request, runtime }) => {
  const entity = await request.get(prefix + '/site/41')
  expect(entity.status()).toBe(200); expect(await entity.text()).toContain('data-site-insights')
  const before = runtime.count('/sites/41/insights')
  const target = await request.get(prefix + '/site/41?domain=target.example')
  expect(target.status()).toBe(200); expect(await target.text()).not.toContain('data-site-insights')
  expect(runtime.count('/sites/41/insights')).toBe(before)
  const game = await request.get(prefix + '/games/82'), html = await game.text()
  expect(game.status()).toBe(200)
  expect(html).toContain('game-insights:82'); expect(html).toContain('peak_30d')
  expect(html).toContain('data-game-tab="insights"')
  runtime.assertQuiet()
})

for (const id of [41, 42, 82, 83]) test('Overview real change link preserves independent entity surface ' + id, async ({ page, runtime }) => {
  await openRuntime(page, '/insights'); await revealImages(page)
  const site = id < 80, path = (site ? '/site/' : '/games/') + id
  const view = page.waitForResponse(res => new URL(res.url()).pathname === '/api/v2/' + (site ? 'nav/sites/' : 'game/games/') + id + '/view'
    && res.request().method() === 'POST' && res.status() === 200)
  await page.locator('[data-change-link][href="' + path + '"]').click()
  await expect(page).toHaveURL(url => url.pathname === path)
  await expect(page.locator(site ? '.site-detail-page' : '.game-detail-page')).toBeVisible()
  await (await view).finished()
  if (site) {
    if (id === 42) await expect(page.locator('[data-site-insights-unavailable]')).toBeVisible()
    else {
      await expect(page.locator('[data-site-insights]')).toBeVisible()
      for (const [key, state] of [['ipv6', 'unknown'], ['tls13', 'unavailable'], ['security_txt', 'unsupported']]) {
        await expect(page.locator('[data-capability-key="' + key + '"]')).toHaveAttribute('data-capability-state', state!)
      }
      await expect(page.getByText('同日统计 · 2026-08-30', { exact: true })).toHaveCount(3)
      await expect(page.locator('[data-entity-timeline] time').first()).toHaveText('2026-08-30')
    }
  } else {
    await page.locator('[data-game-tab="insights"]').click()
    if (id === 83) await expect(page.locator('[data-game-summary-unavailable]')).toBeVisible()
    else await expect(page.locator('[data-current-players]')).toHaveText('0 人')
    await expect(page.locator('.game-detail-tabs')).toBeVisible()
  }
  runtime.assertQuiet()
})
