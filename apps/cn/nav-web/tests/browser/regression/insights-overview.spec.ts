import { test, expect, sources, openRuntime, revealImages, keyboardFocus } from '../fixtures/insights-overview'
import type { Page } from '@playwright/test'
import { assertRuntimeSurface } from '../fixtures/insights-runtime'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'

function stats(html: string) {
  const dl = html.match(/<dl[^>]*data-overview-summary[^>]*>([\s\S]*?)<\/dl>/)?.[1] || ''
  return ['site', 'game'].map(domain => html.match(new RegExp(`<dd data-ecosystem-count="${domain}">(.*?)</dd>`))?.[1] ?? '')
    .concat([...dl.matchAll(/<dd>(.*?)<\/dd>/g)].map(match => match[1]))
}

test('Overview three SSR sources start independently', async ({ request, runtime }) => {
  const gates = sources.map(source => runtime.hold(url => url.pathname === source))
  const pending = request.get('/insights')
  try { await Promise.all(gates.map(gate => gate.wait())) } finally { gates.forEach(gate => gate.release()) }
  expect((await pending).status()).toBe(200)
  await Promise.all(gates.map(gate => gate.done()))
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  runtime.assertQuiet()
})

for (const candidateCase of ['adult', 'tagged', 'missing-tags', 'no-image', 'stale', 'invalid-time', 'invalid-count', 'zero', 'long-title']) {
  test('Overview safe data cover with ' + candidateCase, async ({ page, runtime }) => {
    runtime.state.candidateCase = candidateCase
    await page.setViewportSize({ width: 390, height: 900 })
    await openRuntime(page, '/insights')
    await expect(page.locator('[data-overview-hero]')).toHaveAttribute('data-hero-mode', 'data')
    await expect(page.locator('[data-overview-games] img, [data-overview-activity] [data-domain="game"] img')).toHaveCount(0)
    // No positive cover qualification exists, so neither artwork nor an old
    // Panel observation is substituted for the ecosystem's metric snapshot.
    await expect(page.locator('[data-overview-games] [data-ecosystem-count]')).toHaveText('213')
    await expect(page.locator('[data-overview-games] [data-ecosystem-metric]')).toHaveCount(3)
    await expect(page.locator('[data-overview-games]')).not.toContainText('game fixture')
    await expect(page.locator('[data-overview-ecosystems] a[href*="/games/91"], [data-overview-ecosystems] [data-pulse]')).toHaveCount(0)
    await layout(page, 390)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
    runtime.assertQuiet()
  })
}
for (const metricCase of ['zero', 'fallback', 'missing-date', 'empty', 'hero-only', 'sparse', 'duplicate']) {
  test('Overview metric evidence ' + metricCase, async ({ page, runtime }) => {
    runtime.state.metricCase = metricCase
    const html = await openRuntime(page, '/en/insights')
    if (metricCase === 'zero') {
      expect(stats(html)).toEqual(['0', '0', '0'])
      await expect(page.locator('[data-hero-domain="site"] progress')).toHaveAttribute('value', '0')
      await expect(page.locator('[data-hero-domain="site"] [data-hero-delta]')).toHaveText('0.0 percentage points')
      await expect(page.locator('[data-ecosystem-metric="free"] [data-ecosystem-value]')).toHaveText('0.0%')
      await expect(page.locator('[data-ecosystem-metric="tls13"] [data-ecosystem-value]')).toHaveText('0.0%')
    }
    if (metricCase === 'fallback') {
      await expect(page.locator('[data-hero-domain="site"]')).toHaveAttribute('data-hero-metric', 'tls13')
      await expect(page.locator('[data-hero-domain="site"] a')).toHaveAttribute('href', '/en/insights/sites?metric=tls13')
      expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['http2', 'hsts', 'csp'])
    }
    if (metricCase === 'missing-date') {
      await expect(page.locator('[data-overview-hero] time')).toHaveCount(0)
      await expect(page.locator('[data-hero-domain="site"]')).toContainText('Date unavailable')
      await expect(page.locator('[data-hero-domain="site"] [data-hero-delta]')).toHaveText('—')
      await expect(page.locator('[data-overview-ecosystems] time')).toHaveCount(0)
      for (const coverage of await page.locator('[data-overview-ecosystems] [data-ecosystem-coverage]').all()) await expect(coverage).toHaveText('—')
      await expect(page.locator('[data-ecosystem-metric="tls13"]')).toContainText('Date unavailable')
    }
    if (metricCase === 'empty') {
      await expect(page.locator('[data-overview-hero] progress')).toHaveCount(0)
      await expect(page.locator('[data-overview-hero] [data-hero-unavailable]')).toHaveCount(2)
      await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-empty]')).toHaveCount(2)
      await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-metric]')).toHaveCount(0)
      expect(stats(html)).toEqual(['238', '213', '47'])
    }
    if (metricCase === 'hero-only') {
      await expect(page.locator('[data-hero-domain="site"]')).toHaveAttribute('data-hero-metric', 'ipv6')
      await expect(page.locator('[data-overview-sites] [data-ecosystem-empty]')).toHaveCount(1)
      await expect(page.locator('[data-overview-sites] [data-ecosystem-metric]')).toHaveCount(0)
      await expect(page.locator('[data-overview-games] [data-ecosystem-metric]')).toHaveCount(1)
    }
    if (metricCase === 'sparse') {
      expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['mac', 'linux'])
      expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['csp', 'certificate_verified'])
      await expect(page.locator('[data-ecosystem-metric="mac"] [data-ecosystem-value]')).toHaveText('0.0%')
    }
    if (metricCase === 'duplicate') {
      expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['free', 'windows', 'mac'])
      expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['tls13', 'http2', 'hsts'])
    }
    await expect(page.locator('[data-overview-hero] a')).toHaveCount(2)
    await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-link]')).toHaveCount(2)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
    runtime.assertQuiet()
  })
}
for (const countCase of ['missing', 'invalid']) test('Overview unavailable ecosystem counts retain metrics and links ' + countCase, async ({ page, runtime }) => {
  runtime.state.countCase = countCase
  const html = await openRuntime(page, '/insights')
  expect(stats(html)).toEqual(['—', '—', '47'])
  await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-metric]')).toHaveCount(6)
  await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-link]')).toHaveCount(2)
  runtime.assertQuiet()
})
for (const count of [0, 1, 3]) test('Overview merged ordinary events stay bounded ' + count, async ({ page, runtime }) => {
  runtime.state.eventCount = count
  await openRuntime(page, '/insights')
  await expect(page.locator('[data-overview-activity] [data-change-link]')).toHaveCount(Math.min(count * 2, 5))
  await expect(page.locator('[data-overview-activity] > .overview-section-heading a')).toHaveAttribute('href', '/insights/changes')
  runtime.assertQuiet()
})
async function layout(page: Page, width: number) {
  const result = await page.evaluate(() => {
    const site = document.querySelector('[data-overview-sites]')!.getBoundingClientRect()
    const game = document.querySelector('[data-overview-games]')!.getBoundingClientRect()
    const heroGame = document.querySelector('[data-hero-domain="game"]')!.getBoundingClientRect()
    const heroSite = document.querySelector('[data-hero-domain="site"]')!.getBoundingClientRect()
    const box = (selector: string) => document.querySelector(selector)!.getBoundingClientRect().toJSON()
    return { overflow: Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - innerWidth, site: site.toJSON(), game: game.toJSON(), heroGame: heroGame.toJSON(), heroSite: heroSite.toJSON(), header: box('[data-overview-header]'), hero: box('[data-overview-hero]'), ecosystems: box('[data-overview-ecosystems]'), summary: box('[data-overview-summary]'), activity: box('[data-overview-activity]') }
  })
  expect(result.overflow).toBeLessThanOrEqual(0)
  expect(result.hero.top).toBeGreaterThanOrEqual(result.header.bottom)
  expect(result.ecosystems.top).toBeGreaterThanOrEqual(result.hero.bottom)
  expect(result.summary.top).toBeGreaterThanOrEqual(result.ecosystems.bottom)
  expect(result.activity.top).toBeGreaterThanOrEqual(result.summary.bottom)
  if (width >= 1200) {
    expect(result.heroGame.width / result.heroSite.width).toBeGreaterThan(1.9)
    expect(result.heroGame.width / result.heroSite.width).toBeLessThan(2.2)
    expect(Math.abs(result.heroGame.top - result.heroSite.top)).toBeLessThan(2)
  } else expect(result.heroSite.top).toBeGreaterThanOrEqual(result.heroGame.bottom)
  if (width >= 960) {
    expect(result.game.width / result.site.width).toBeCloseTo(1, 1)
    expect(Math.abs(result.site.top - result.game.top)).toBeLessThan(2)
    expect(result.site.left).toBeGreaterThan(result.game.left)
  } else expect(result.site.top).toBeGreaterThanOrEqual(result.game.bottom)
}
for (const prefix of ['', '/en']) for (const theme of ['light', 'dark'] as const) for (const width of [1440, 1024, 768, 390]) {
  test(`Overview SSR, hydration, media and layout ${prefix} ${theme} ${width}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    const gameArt: string[] = []
    page.on('request', request => { if (request.resourceType() === 'image' && /\/game(?:-\d+)?\.svg/.test(request.url())) gameArt.push(request.url()) })
    const html = await openRuntime(page, prefix + '/insights')
    expect(stats(html)).toEqual(['238', '213', '47'])
    for (const section of ['header', 'hero', 'ecosystems', 'activity', 'sites', 'games', 'explore']) expect(html).toContain('data-overview-' + section)
    await expect(page.locator('h1')).toHaveCount(1)
    await expect(page.locator('h1')).toHaveText(prefix ? 'A Closer Look at the Furry World.' : '看见生态的另一面。')
    expect(html).toContain('<meta name="description"')
    for (const path of ['/insights/sites', '/insights/games', '/insights/changes', '/insights/sites/certificates', '/insights/sites/compare', '/insights/games/players', '/insights/games/prices', '/insights/games/languages', '/insights/games/compare']) expect(html).toContain('href="' + prefix + path + '"')
    for (const path of ['/site/41', '/site/42', '/games/82', '/games/83']) expect(html).toContain('href="' + prefix + path + '"')
    for (const id of [91, 92, 93]) expect(html).not.toContain(`href="${prefix}/games/${id}"`)
    expect(html).toContain('/nav/sites/41/icon/' + 'a'.repeat(32) + '.svg')
    expect(html).toContain('/defaultLogo.svg'); expect(html).toContain('data-media-state="fallback"')
    expect(html).not.toContain('datetime="2026-09-01T10:00:00.000Z"')
    await expect(page.locator('[data-hero-domain="site"]')).toHaveAttribute('data-hero-metric', 'ipv6')
    await expect(page.locator('[data-hero-domain="site"]')).toContainText('63.0%')
    await expect(page.locator('[data-hero-domain="site"]')).toContainText('95.0%')
    await expect(page.locator('[data-hero-domain="site"] [data-hero-delta]')).toHaveText(prefix ? '-1.1 percentage points' : '-1.1 个百分点')
    await expect(page.locator('[data-hero-domain="game"] [data-hero-delta]')).toHaveText(prefix ? '+4.2 percentage points' : '+4.2 个百分点')
    await expect(page.locator('[data-hero-domain="site"] time')).toHaveAttribute('datetime', '2026-08-31')
    await expect(page.locator('[data-hero-domain="game"] time')).toHaveAttribute('datetime', '2026-08-29')
    await expect(page.locator('[data-hero-domain="site"] a')).toHaveAttribute('href', prefix + '/insights/sites?metric=ipv6')
    await expect(page.locator('[data-overview-hero] img, [data-overview-hero] canvas, [data-overview-games] img, [data-overview-activity] [data-domain="game"] img')).toHaveCount(0)
    await expect(page.locator('.insight-activity-item--hero')).toHaveCount(0)
    await expect(page.locator('[data-overview-activity] [data-change-link]')).toHaveCount(4)
    await expect(page.getByRole('heading', { level: 2, name: prefix ? 'Explore the ecosystems' : '探索生态', exact: true })).toBeVisible()
    await expect(page.locator('[data-overview-summary] dd')).toHaveCount(1)
    await expect(page.locator('[data-overview-games] [data-ecosystem-count="game"]')).toHaveText('213')
    await expect(page.locator('[data-overview-sites] [data-ecosystem-count="site"]')).toHaveText('238')
    await expect(page.locator('[data-ecosystem-count]')).toHaveCount(2)
    for (const [domain, path] of [['games', '/insights/games'], ['sites', '/insights/sites']]) {
      await expect(page.locator(`[data-overview-${domain}] [data-ecosystem-link]`)).toHaveAttribute('href', prefix + path)
    }
    expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['free', 'windows', 'mac'])
    expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['tls13', 'http2', 'hsts'])
    await expect(page.locator('[data-overview-sites] [data-ecosystem-metric="ipv6"]')).toHaveCount(0)
    await expect(page.locator('[data-ecosystem-metric="hsts"] [data-ecosystem-value]')).toHaveText('0.0%')
    await expect(page.locator('[data-ecosystem-metric="tls13"] [data-ecosystem-coverage]')).toHaveText('90.0%')
    await expect(page.locator('[data-ecosystem-metric="tls13"] time')).toHaveAttribute('datetime', '2026-08-30')
    await expect(page.locator('[data-ecosystem-metric="windows"] time')).toHaveAttribute('datetime', '2026-08-28')
    await expect(page.locator('[data-overview-activity] time').filter({ hasText: /^2026-09-01$/ })).toHaveAttribute('datetime', '2026-09-01')
    expect(await page.locator('progress').evaluateAll(elements => elements.every(el => el.value >= 0 && el.value <= 1 && el.max === 1 && el.getAttribute('aria-labelledby')))).toBe(true)
    await layout(page, width); await revealImages(page)
    await keyboardFocus(page.locator('.insights-primary-nav a').last())
    await keyboardFocus(page.locator('[data-hero-domain="game"] a'))
    await keyboardFocus(page.locator('[data-hero-domain="site"] a'))
    await keyboardFocus(page.locator('[data-overview-games] [data-ecosystem-link]'))
    await keyboardFocus(page.locator('[data-overview-sites] [data-ecosystem-link]'))
    await assertRuntimeSurface(page, '.insights-overview-page', theme)
    expect(await page.locator('a a').count()).toBe(0)
    expect(gameArt).toEqual([])
    if (process.env.GOFURRY_OVERVIEW_REVIEW_DIR) {
      await mkdir(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, { recursive: true })
      await page.mouse.click(0, 0); await page.evaluate(() => window.scrollTo(0, 0))
      await page.screenshot({ path: join(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, `${prefix ? 'en' : 'zh'}-${theme}-${width}.png`), fullPage: true })
      if (width === 1440 || width === 390) await page.locator('[data-overview-ecosystems]').screenshot({ path: join(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, `ecosystems-${prefix ? 'en' : 'zh'}-${theme}-${width}.png`) })
    }
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
    runtime.assertQuiet()
  })
}
for (const [source, values] of [['nav', ['—', '213', '—']], ['game', ['238', '—', '—']], ['panel', ['238', '213', '47']], ['all', ['—', '—', '—']]] as const) {
  test('Overview independent failure ' + source, async ({ page, runtime }) => {
    runtime.state.failure = source
    const browserSources: string[] = []
    page.on('request', request => { if (sources.includes(new URL(request.url()).pathname)) browserSources.push(request.url()) })
    const html = await openRuntime(page, '/insights')
    expect(stats(html)).toEqual(values)
    expect(html.includes('data-ecosystem-metric="free"')).toBe(source !== 'game' && source !== 'all')
    expect(html.includes('data-ecosystem-metric="tls13"')).toBe(source !== 'nav' && source !== 'all')
    expect(html.includes('data-change-link')).toBe(source !== 'all')
    await expect(page.locator('[data-hero-domain="game"] [data-hero-unavailable]')).toHaveCount(source === 'game' || source === 'all' ? 1 : 0)
    await expect(page.locator('[data-hero-domain="site"] [data-hero-unavailable]')).toHaveCount(source === 'nav' || source === 'all' ? 1 : 0)
    await expect(page.locator('[data-overview-hero] a')).toHaveCount(2)
    await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-link]')).toHaveCount(2)
    await expect(page.locator('[data-overview-games] [data-ecosystem-unavailable]')).toHaveCount(source === 'game' || source === 'all' ? 1 : 0)
    await expect(page.locator('[data-overview-sites] [data-ecosystem-unavailable]')).toHaveCount(source === 'nav' || source === 'all' ? 1 : 0)
    // The unchanged Overview services retain ofetch's one retry on 503. Only
    // Home explicitly disables retries; hydration must add no source request.
    const expected = [...sources]
    if (source === 'nav' || source === 'all') expected.push(sources[0]!)
    if (source === 'game' || source === 'all') expected.push(sources[1]!)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(expected.sort())
    expect(browserSources).toEqual([])
    await layout(page, 1440)
    if (process.env.GOFURRY_OVERVIEW_REVIEW_DIR) {
      await mkdir(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, { recursive: true })
      await page.screenshot({ path: join(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, `failure-${source}.png`), fullPage: true })
    }
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
  expect(html).toContain('data-ecosystem-metric="free"'); expect(html).toContain('data-ecosystem-metric="tls13"')
  expect(runtime.count('/game/home')).toBe(1)
  gate.release(); await gate.done(); runtime.assertQuiet()
})
test('Overview ordinary Site event and real media failures preserve identity and data Hero', async ({ page, runtime }) => {
  runtime.state.siteHero = true
  await page.setViewportSize({ width: 390, height: 1000 })
  await openRuntime(page, '/insights')
  await expect(page.locator('[data-change-link]').first()).toHaveAttribute('data-domain', 'site')
  await expect(page.locator('.insight-activity-item--hero')).toHaveCount(0)
  expect((await page.locator('[data-change-link]').first().locator('img').boundingBox())!.width).toBeLessThanOrEqual(72)
  await revealImages(page)
  runtime.failImages = true
  await openRuntime(page, '/insights'); await revealImages(page)
  await expect(page.locator('main .insight-entity-media img')).toHaveCount(0)
  expect(await page.locator('main [role="img"][aria-label]').count()).toBeGreaterThan(0)
  await expect(page.locator('[data-hero-domain="site"]')).toContainText('63.0%')
  await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  await layout(page, 390); runtime.assertQuiet()
})

for (const width of [1440, 390]) test(`Overview Dark ready shell ${width}`, async ({ page, context, runtime }) => {
  await page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
  await context.addInitScript(() => localStorage.setItem('theme', 'dark'))
  const html = await openRuntime(page, '/insights')
  expect(stats(html)).toEqual(['238', '213', '47'])
  for (const part of ['header', 'activity', 'sites', 'games', 'explore']) await expect(page.locator('[data-overview-' + part + ']')).toBeVisible()
  await expect(page.locator('.insights-primary-nav')).toBeVisible()
  await revealImages(page)
  await assertRuntimeSurface(page, '.insights-overview-page', 'dark')
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  runtime.assertQuiet()
})
