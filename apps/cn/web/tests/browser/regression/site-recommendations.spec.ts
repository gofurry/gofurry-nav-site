import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'

const similar = (page: Page) => page.locator('[data-site-similar]')
const items = (page: Page) => page.locator('[data-site-similar-id]')
async function open(page: Page, path = '/en/site/41?tab=overview') {
  const counted = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  const html = await openRuntime(page, path)
  await (await counted).finished()
  return html
}
async function mode(page: Page, value: 'sfw' | 'nsfw') {
  await page.evaluate(value => {
    localStorage.setItem('mode', value)
    window.dispatchEvent(new CustomEvent('mode-change', { detail: { mode: value, displayMode: value } }))
  }, value)
}
async function review(page: Page, name: string) {
  const directory = process.env.GOFURRY_SITE_DETAIL_REVIEW_DIR
  if (!directory) return
  await mkdir(directory, { recursive: true }); await settleRuntime(page)
  await page.screenshot({ path: join(directory, 'task-e-' + name + '.png'), fullPage: true })
}

test('recommendation SSR is parallel Site/language data with SFW-only markup', async ({ request, runtime }) => {
  runtime.state.recommendations = 'ready'
  const held = runtime.hold(url => url.pathname.endsWith('/recommendations'))
  try {
    const pending = request.get('/en/site/41')
    await expect.poll(() => held.received, { timeout: 90000 }).toBe(true)
    await expect.poll(() => runtime.count('/sites/41/detail')).toBe(1)
    await expect.poll(() => runtime.count('/sites/41/insights')).toBe(1)
    expect(runtime.count('/observations')).toBe(0); expect(runtime.count('/sites/41/view')).toBe(0)
    held.release(); const response = await pending
    expect(response.status()).toBe(200)
    const markup = (await response.text()).match(/<section id="site-similar-workspace"[\s\S]*?<\/section>/)?.[0]
    expect(markup).toBeDefined(); expect(markup).toContain('Similar fixture 42'); expect(markup).toContain('24,088')
    expect(markup).not.toContain('Similar fixture 43'); expect(markup).not.toContain('Similar fixture 47')
    expect(markup).not.toMatch(/first\.example|second\.example|similar\d\.example/)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights', '/api/v2/nav/sites/41/recommendations'])
    expect(Object.fromEntries(runtime.calls.find(call => call.url.pathname.endsWith('/recommendations'))!.url.searchParams)).toEqual({ lang: 'en', limit: '8' })
    runtime.assertQuiet()
  } finally { held.release() }
})

for (const theme of ['light', 'dark'] as const) test('Desktop Similar and local mode filtering ' + theme, async ({ page, context, runtime }) => {
  runtime.state.recommendations = 'ready'; runtime.state.historyCount = 100
  await page.setViewportSize({ width: 1440, height: 900 })
  await context.addInitScript(theme => { localStorage.setItem('theme', theme); localStorage.setItem('mode', 'nsfw') }, theme)
  const html = await open(page, '/en/site/41')
  expect(html.match(/<section id="site-similar-workspace"[\s\S]*?<\/section>/)?.[0]).not.toContain('Similar fixture 43')
  await expect(items(page)).toHaveCount(8)
  await expect(items(page).locator('img')).toHaveCount(8)
  await expect.poll(() => items(page).locator('img').evaluateAll(images => images.every(image => (image as HTMLImageElement).complete && (image as HTMLImageElement).naturalWidth > 0))).toBe(true)
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  expect(runtime.calls).toHaveLength(5)
  const target = await page.locator('[data-site-target-context]').boundingBox(), box = await similar(page).boundingBox()
  expect(box!.y).toBeGreaterThan(target!.y + target!.height)
  expect(box!.x).toBeCloseTo(target!.x, 0)
  await expect(page.locator('[data-site-aux-tab]')).toBeHidden()
  await expect(items(page).first()).toHaveAttribute('href', '/en/site/42')
  await expect(items(page).first()).toContainText('Similar fixture 42')
  await expect(items(page).first()).toContainText('A community to explore.')
  await expect(items(page).first()).toContainText('24,088')
  await expect(similar(page)).not.toContainText(/first\.example|second\.example|similar\d\.example/)
  await expect(page.locator('[data-site-similar-id="41"]')).toHaveCount(0)
  await expect(similar(page)).not.toContainText(/HTTP|Ping|CDN|Best match|score|similarity|NSFW/)
  await expect(items(page).first()).toHaveCSS('transition-duration', '0.5s, 0.5s')
  await expect(items(page).first()).toHaveCSS('transform', 'none'); await expect(items(page).first()).toHaveCSS('border-width', '0px')
  await mode(page, 'sfw'); await expect(items(page)).toHaveCount(6)
  await mode(page, 'nsfw'); await expect(items(page)).toHaveCount(8)
  expect(runtime.count('/recommendations')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, 'desktop-' + theme)
  runtime.assertQuiet()
})

test('Target and workspace changes never re-request recommendations', async ({ page, runtime }) => {
  runtime.state.recommendations = 'ready'
  await open(page)
  const before = runtime.calls.length
  await page.locator('[data-site-target-trigger]').click(); await page.locator('[data-site-target-option="alt.example"]').click()
  await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
  expect(runtime.calls.slice(before).map(call => call.url.pathname)).toEqual(['/api/v2/nav/sites/41/detail'])
  await page.locator('[data-site-primary-tab="security"]').click(); await page.locator('[data-site-security-tab="tls"]').click()
  await page.locator('[data-site-primary-tab="insights"]').click()
  await expect(page.locator('[data-site-insight-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await page.locator('[data-site-capability="tls13"]').click(); await page.locator('[data-site-insight-range="90d"]').click()
  await expect(page.locator('[data-site-insight-trend]')).toHaveAttribute('data-site-trend-range', '90d')
  await settleRuntime(page)
  expect(runtime.count('/recommendations')).toBe(1); expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  runtime.assertQuiet()
})

test('Mobile Similar is a local accessible tab and Desktop restores the underlying route', async ({ page, runtime }) => {
  runtime.state.recommendations = 'ready'
  await page.setViewportSize({ width: 390, height: 844 }); await open(page, '/en/site/41')
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  const aux = page.locator('[data-site-aux-tab="similar"]'), route = page.url(), history = await page.evaluate(() => window.history.length)
  const seo = await page.locator('title, link[rel="canonical"], link[rel="alternate"], meta[name="description"]').evaluateAll(nodes => nodes.map(node => node.outerHTML))
  await page.locator('[data-site-primary-tab="observation"]').focus(); await page.keyboard.press('End')
  await expect(aux).toBeFocused(); await expect(aux).toHaveAttribute('aria-selected', 'true')
  const tabBounds = await page.locator('[data-site-primary-tabs]').boundingBox(), activeBounds = await aux.boundingBox()
  expect(activeBounds!.x + activeBounds!.width).toBeLessThanOrEqual(tabBounds!.x + tabBounds!.width + 1)
  await expect(similar(page)).toHaveAttribute('role', 'tabpanel'); await expect(similar(page)).toHaveAttribute('aria-labelledby', 'site-tab-similar')
  await expect(page.locator('[data-site-workspace]')).toBeHidden()
  expect(page.url()).toBe(route); expect(await page.evaluate(() => window.history.length)).toBe(history)
  expect(await page.locator('title, link[rel="canonical"], link[rel="alternate"], meta[name="description"]').evaluateAll(nodes => nodes.map(node => node.outerHTML))).toEqual(seo)
  expect(new Set(await items(page).evaluateAll(nodes => nodes.map(node => node.getBoundingClientRect().x))).size).toBe(1)
  await expect(page.locator('[data-site-primary-tabs] [role="tab"][tabindex="0"]')).toHaveCount(1)
  expect(await page.locator('[data-site-primary-tabs] [role="tab"]').evaluateAll(nodes => nodes.every(node => node.scrollWidth <= node.clientWidth))).toBe(true)
  await assertRuntimeSurface(page, '[data-site-detail]', 'light'); await review(page, 'mobile-light')
  await page.keyboard.press('ArrowLeft'); await expect(page.locator('[data-site-primary-tab="insights"]')).toHaveAttribute('aria-selected', 'true')
  await page.keyboard.press('Home'); await expect(page.locator('[data-site-overview]')).toBeVisible()
  await page.keyboard.press('ArrowLeft'); await expect(aux).toHaveAttribute('aria-selected', 'true')
  await page.keyboard.press('ArrowRight'); await expect(page.locator('[data-site-primary-tab="overview"]')).toHaveAttribute('aria-selected', 'true')
  await page.locator('[data-site-primary-tab="observation"]').click(); await aux.click()
  const beforeResize = page.url()
  await page.setViewportSize({ width: 1440, height: 900 })
  await expect(aux).toBeHidden(); await expect(page.locator('[data-site-workspace]')).toBeVisible()
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  expect(page.url()).toBe(beforeResize)
  await page.locator('[data-site-primary-tab="observation"]').focus(); await page.keyboard.press('End')
  await expect(page.locator('[data-site-primary-tab="insights"]')).toBeFocused()
  expect(runtime.count('/recommendations')).toBe(1); expect(runtime.count('/observations')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  await settleRuntime(page); runtime.assertQuiet()
})

test('raw adult candidates keep the mobile tab stable without exposing filtering reasons', async ({ page, runtime }) => {
  runtime.state.recommendations = 'adult'
  await page.setViewportSize({ width: 390, height: 844 }); const html = await open(page)
  expect(html).toContain('data-site-aux-tab="similar"')
  await page.locator('[data-site-aux-tab]').click()
  await expect(page.locator('[data-site-similar-empty]')).toHaveText('No similar sites are available to display.')
  await expect(items(page)).toHaveCount(0); await expect(similar(page)).not.toContainText('NSFW')
  await mode(page, 'nsfw'); await expect(items(page)).toHaveCount(8)
  await mode(page, 'sfw'); await expect(items(page)).toHaveCount(0)
  await expect(page.locator('[data-site-aux-tab]')).toHaveAttribute('aria-selected', 'true')
  await page.setViewportSize({ width: 1440, height: 900 }); await expect(similar(page)).toHaveCount(0)
  expect(runtime.count('/recommendations')).toBe(1); runtime.assertQuiet()
})

for (const state of ['empty', 'unavailable', 'failure'] as const) test('optional recommendation slice ' + state + ' leaves the Site page healthy', async ({ page, runtime }) => {
  runtime.state.recommendations = state
  await open(page)
  await expect(page.locator('[data-site-overview]')).toBeVisible()
  await expect(similar(page)).toHaveCount(0); await expect(page.locator('[data-site-aux-tab]')).toHaveCount(0)
  expect(runtime.count('/recommendations')).toBe(1); expect(runtime.calls).toHaveLength(4)
  runtime.assertQuiet()
})

test('recommendation locale identity changes and a Similar link counts the destination exactly once', async ({ page, runtime }) => {
  runtime.state.recommendations = 'ready'
  await page.setViewportSize({ width: 1440, height: 900 }); await open(page)
  await page.locator('button').filter({ has: page.locator('img[alt="CN"]') }).first().click()
  await expect(page).toHaveURL('/site/41?tab=overview')
  await expect(items(page).first()).toContainText('相似站点示例')
  expect(runtime.calls.filter(call => call.url.pathname.endsWith('/recommendations')).map(call => Object.fromEntries(call.url.searchParams))).toEqual([{ lang: 'en', limit: '8' }, { lang: 'zh', limit: '8' }])
  const destination = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/42/view') && response.request().method() === 'POST')
  await expect(items(page).first()).toHaveAttribute('href', '/site/42')
  await items(page).first().click(); await (await destination).finished()
  await expect(page).toHaveURL('/site/42')
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await settleRuntime(page)
  expect(runtime.count('/sites/42/view')).toBe(1)
  expect(runtime.count('/sites/42/recommendations')).toBe(1)
  expect(runtime.count('/sites/42/detail')).toBe(1)
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com/site/42')
  runtime.assertQuiet()
})
