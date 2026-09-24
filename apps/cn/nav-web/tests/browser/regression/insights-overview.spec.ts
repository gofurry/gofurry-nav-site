import { test, expect, sources, openRuntime, revealImages, keyboardFocus } from '../fixtures/insights-overview'
import type { Page } from '@playwright/test'

function stats(html: string) {
  const dl = html.match(/<dl class="overview-stats">([\s\S]*?)<\/dl>/)?.[1] || ''
  return [...dl.matchAll(/<dd>(.*?)<\/dd>/g)].map(match => match[1])
}
async function layout(page: Page, width: number) {
  const result = await page.evaluate(() => {
    const site = document.querySelector('[data-overview-sites]')!.getBoundingClientRect()
    const game = document.querySelector('[data-overview-games]')!.getBoundingClientRect()
    return { overflow: document.documentElement.scrollWidth - innerWidth, site: site.toJSON(), game: game.toJSON() }
  })
  expect(result.overflow).toBeLessThanOrEqual(0)
  if (width >= 1200) { expect(result.game.width).toBeGreaterThan(result.site.width * 1.3); expect(Math.abs(result.site.top - result.game.top)).toBeLessThan(2) }
  if (width === 390) expect(result.game.top).toBeGreaterThan(result.site.top)
}
for (const prefix of ['', '/en']) for (const width of [1440, 1024, 390]) {
  test('Overview SSR, hydration, media and layout ' + prefix + ' ' + width, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    const html = await openRuntime(page, prefix + '/insights')
    expect(stats(html)).toEqual(['238', '213', '47'])
    for (const section of ['header', 'activity', 'sites', 'games', 'explore']) expect(html).toContain('data-overview-' + section)
    for (const path of ['/site/41', '/site/42', '/games/82', '/games/83', '/games/91', '/games/92', '/games/93']) expect(html).toContain('href="' + prefix + path + '"')
    expect(html).toContain('/nav/sites/41/icon/' + 'a'.repeat(32) + '.svg')
    expect(html).toContain('/defaultLogo.svg'); expect(html).toContain('data-media-state="fallback"')
    expect(html).toContain('datetime="2026-09-01T10:00:00.000Z"')
    await expect(page.locator('[data-overview-activity] [data-change-link]')).toHaveCount(4)
    await expect(page.locator('[data-pulse="players"] strong')).toHaveText('Active game fixture')
    await expect(page.locator('[data-pulse="players"]')).toContainText('2,800')
    await expect(page.locator('[data-pulse="discount"]')).toContainText('5.99')
    expect(await page.locator('progress').evaluateAll(elements => elements.every(el => el.value >= 0 && el.value <= 1 && el.max === 1 && el.getAttribute('aria-labelledby')))).toBe(true)
    await layout(page, width); await revealImages(page)
    await keyboardFocus(page.locator('.insights-primary-nav a').last())
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
    runtime.assertQuiet()
  })
}
for (const [source, values] of [['nav', ['—', '213', '—']], ['game', ['238', '—', '—']], ['panel', ['238', '213', '47']], ['all', ['—', '—', '—']]] as const) {
  test('Overview independent failure ' + source, async ({ request, runtime }) => {
    runtime.state.failure = source
    const response = await request.get('/insights'), html = await response.text()
    expect(response.status()).toBe(200); expect(stats(html)).toEqual(values)
    expect(html.includes('data-pulse="players"')).toBe(source !== 'panel' && source !== 'all')
    expect(html.includes('data-metric="tls13"')).toBe(source !== 'nav' && source !== 'all')
    expect(html.includes('data-change-link')).toBe(source !== 'all')
    runtime.assertQuiet()
  })
}
test('Overview optional panel timeout is bounded without retry or loss of independent facts', async ({ request, runtime }) => {
  const gate = runtime.hold(url => url.pathname === sources[2])
  const start = performance.now(), pending = request.get('/insights')
  await gate.wait()
  const response = await pending, html = await response.text()
  expect(performance.now() - start).toBeLessThan(11000)
  expect(response.status()).toBe(200); expect(gate.completed).toBe(false)
  expect(stats(html)).toEqual(['238', '213', '47']); expect(html).not.toContain('data-pulse="players"')
  expect(runtime.count('/game/home')).toBe(1)
  gate.release(); await gate.done(); runtime.assertQuiet()
})
test('Overview Site hero and real media failures preserve identity and Dark layout', async ({ page, runtime }) => {
  runtime.state.siteHero = true
  await page.setViewportSize({ width: 390, height: 1000 })
  await openRuntime(page, '/insights')
  await expect(page.locator('.insight-activity-item--hero')).toHaveAttribute('data-domain', 'site')
  expect((await page.locator('.insight-activity-item--hero img').boundingBox())!.width).toBeLessThanOrEqual(72)
  await revealImages(page)
  runtime.failImages = true
  await openRuntime(page, '/insights'); await revealImages(page)
  await expect(page.locator('main .insight-entity-media img')).toHaveCount(0)
  expect(await page.locator('main [role="img"][aria-label]').count()).toBeGreaterThan(0)
  await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  await layout(page, 390); runtime.assertQuiet()
})
