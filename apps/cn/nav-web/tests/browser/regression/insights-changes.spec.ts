import type { Page } from '@playwright/test'
import { test, expect, pathFor, changeCategories, openRuntime, revealImages, keyboardFocus, settleRuntime } from '../fixtures/insights-changes'
const events = (page: Page) => page.locator('.insights-change-explorer-item')
async function ready(page: Page) { await expect(page.locator('.insights-change-explorer-feed')).toHaveAttribute('aria-busy', 'false') }
async function layout(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true)
  expect(await page.locator('.insights-change-explorer-filters, .insights-changes-day, .insights-change-explorer-item').evaluateAll(els => els.every(el => el.scrollWidth <= el.clientWidth + 1))).toBe(true)
  await expect(page.locator('.insights-change-explorer-filters')).toHaveCSS('box-shadow', 'none')
}
for (const domain of ['site', 'game']) for (const locale of ['zh', 'en']) for (const width of [1440, 1024, 390]) {
  test('Changes SSR/hydration, time precision and layout ' + domain + ' ' + locale + ' ' + width, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    const html = await openRuntime(page, pathFor(domain, locale))
    expect(html).toContain('<h1>' + (locale === 'en' ? 'Change Explorer' : '变化探索') + '</h1>')
    expect(html).toContain('insights-primary-nav'); expect(html).not.toContain('insights-domain-nav')
    expect(html).toContain('2026-09-09T13:32:00Z')
    expect(html).toContain('href="' + (locale === 'en' ? '/en' : '') + '/' + (domain === 'site' ? 'site' : 'games') + '/31"')
    expect(html).toContain(domain === 'site' ? '/nav/sites/31/icon/' : 'change-header.svg')
    expect(html).toContain(domain === 'site' ? '/defaultLogo.svg' : 'data-media-state="fallback"')
    await expect(events(page)).toHaveCount(4)
    await expect(events(page).nth(2).locator('strong')).toHaveText('#32')
    expect(await events(page).nth(2).locator('.insight-entity-media img, .insight-entity-media [role="img"]').evaluate(el => el.getAttribute('alt') ?? el.getAttribute('aria-label'))).toBe('#32')
    await expect(page.locator('[data-change-date]')).toHaveCount(2)
    await expect(events(page).nth(1).locator('time')).toHaveText('2026-09-09')
    await keyboardFocus(page.locator('[data-category-filter]').last())
    await layout(page)
    expect(runtime.calls).toHaveLength(1); runtime.assertQuiet()
  })
}
for (const domain of ['site', 'game'] as const) test('Changes filters, cursor failure/retry and locale ' + domain, async ({ page, runtime }) => {
  await openRuntime(page, pathFor(domain))
  for (const category of Object.keys(changeCategories[domain])) {
    const response = page.waitForResponse(res => new URL(res.url()).searchParams.get('category') === category)
    await page.locator('[data-category-filter="' + category + '"]').click()
    await response; await ready(page)
    expect(new URL(page.url()).searchParams.get('category')).toBe(category)
    await expect(events(page).first()).toHaveAttribute('data-event-type', changeCategories[domain][category as keyof typeof changeCategories[typeof domain]])
  }
  for (const range of ['7d', '90d', 'all', '30d']) {
    const response = page.waitForResponse(res => new URL(res.url()).searchParams.get('range') === range)
    await page.locator('[data-change-range="' + range + '"]').click(); await response; await ready(page)
    expect(new URL(page.url()).searchParams.get('range')).toBe(range)
  }
  const links = await events(page).evaluateAll(els => els.map(el => el.getAttribute('href')))
  runtime.state.moreFailure = true
  await page.locator('[data-load-more]').click()
  await expect(page.locator('.insights-change-explorer-feed__inline-error')).toBeVisible()
  expect(await events(page).evaluateAll(els => els.map(el => el.getAttribute('href')))).toEqual(links)
  expect(new URL(page.url()).searchParams.has('cursor')).toBe(false)
  runtime.state.moreFailure = false
  await page.locator('[data-load-more]').click(); await ready(page)
  await expect(events(page)).toHaveCount(6)
  await expect(page.locator('[data-change-date="2026-09-08"] li')).toHaveCount(3)
  await expect(page.locator('[data-load-more]')).toHaveCount(0)
  expect((await events(page).evaluateAll(els => els.map(el => el.getAttribute('href')))).slice(0, 4)).toEqual(links)
  const query = new URL(page.url()).search
  expect(new URL(page.url()).searchParams.has('cursor')).toBe(false)
  await page.locator('button:has(img[alt="EN"])').first().click()
  await expect(page).toHaveURL(url => url.pathname === '/en/insights/changes' && url.search === query)
  await expect(events(page).first()).toHaveAttribute('href', /^\/en\//)
  runtime.assertQuiet()
})
test('Changes normalize query and discard a held stale range response', async ({ page, runtime }) => {
  await openRuntime(page, '/insights/changes?domain=bad&range=bad&category=bad&cursor=leak&keep=yes')
  await expect(page).toHaveURL(url => url.searchParams.get('domain') === 'site' && url.searchParams.get('range') === '30d' && !url.searchParams.has('category') && !url.searchParams.has('cursor'))
  expect(new URL(page.url()).searchParams.get('keep')).toBe('yes')
  await page.locator('[data-domain-filter="game"]').click(); await ready(page)
  await expect(page).toHaveURL(url => url.searchParams.get('domain') === 'game')
  await expect(page.locator('[data-category-filter="capability"]')).toHaveCount(0)
  const gate = runtime.hold(url => url.searchParams.get('range') === '7d')
  await page.locator('[data-change-range="7d"]').click(); await gate.wait()
  const response = page.waitForResponse(res => new URL(res.url()).searchParams.get('range') === '90d')
  await page.locator('[data-change-range="90d"]').click(); await response; await ready(page)
  gate.release(); await gate.done(); await settleRuntime(page)
  await expect(events(page).first().locator('strong')).toContainText('90d')
  expect(new URL(page.url()).searchParams.get('range')).toBe('90d')
  await expect(events(page)).toHaveCount(4); runtime.assertQuiet()
})
for (const locale of ['zh', 'en']) test('Changes empty and real Retry ' + locale, async ({ page, runtime }) => {
  runtime.state.empty = true
  await openRuntime(page, pathFor('site', locale))
  await expect(events(page)).toHaveCount(0)
  await expect(page.locator('.insights-changes-state')).toContainText(locale === 'en' ? 'No public changes' : '暂无公开变化')
  runtime.state.empty = false; runtime.state.failure = true
  await openRuntime(page, pathFor('game', locale))
  await expect(page.locator('[data-retry-changes]')).toBeVisible()
  await expect(events(page)).toHaveCount(0)
  runtime.state.failure = false
  await page.locator('[data-retry-changes]').click(); await ready(page)
  await expect(events(page)).toHaveCount(4); runtime.assertQuiet()
})
for (const domain of ['site', 'game']) test('Changes failed media preserves accessible identity ' + domain, async ({ page, runtime }) => {
  runtime.failImages = true
  await openRuntime(page, pathFor(domain, 'en')); await revealImages(page)
  await expect(events(page).locator('[role="img"][aria-label]')).toHaveCount(4)
  await layout(page); runtime.assertQuiet()
})
for (const width of [1440, 1024, 390]) test('Changes Dark layout ' + width, async ({ page, runtime }) => {
  await page.setViewportSize({ width, height: 1000 })
  await openRuntime(page, pathFor('game', 'en'))
  await page.getByRole('button', { name: 'Toggle theme icon', exact: true }).first().click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  await layout(page); runtime.assertQuiet()
})
