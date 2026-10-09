import { test, expect, sources, openRuntime, revealImages, keyboardFocus } from '../fixtures/insights-overview'
import type { Page } from '@playwright/test'
import { assertRuntimeSurface } from '../fixtures/insights-runtime'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'

function stats(html: string) {
  return ['site', 'game'].map(domain => html.match(new RegExp(`<dd data-ecosystem-count="${domain}">(.*?)</dd>`))?.[1] ?? '')
    .concat(html.match(/<dd data-activity-total>(.*?)<\/dd>/)?.[1] ?? '')
}

for (const prefix of ['', '/en']) for (const visualCase of ['empty', 'populated']) test(`Overview ignores optional visual data ${prefix || 'zh'} ${visualCase}`, async ({ page, runtime }) => {
  runtime.state.visualCase = visualCase
  const mediaRequests: string[] = []
  page.on('request', request => { if (/steamstatic|akamaihd/.test(request.url())) mediaRequests.push(request.url()) })
  const html = await openRuntime(page, prefix + '/insights')
  expect(stats(html)).toEqual(['238', '213', '47'])
  await expect(page.locator('[data-overview-hero]')).toHaveAttribute('data-hero-mode', 'data')
  await expect(page.locator('[data-overview-hero] img, [data-overview-ecosystems] img, [data-overview-activity] img')).toHaveCount(0)
  expect((await page.locator('[data-overview-hero], [data-overview-ecosystems], [data-overview-activity]').allTextContents()).join(' ')).not.toContain('Visual candidate')
  await expect(page.locator('a[href$="/games/91"], a[href$="/games/92"], a[href$="/games/93"]')).toHaveCount(0)
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  expect(mediaRequests).toEqual([])
  runtime.assertQuiet()
})

for (const [prefix, title, description] of [
  ['', 'Furry 生态观测 - GoFurry', '查看 Furry 网站与游戏生态的公开指标、近期变化、统计覆盖和可靠历史数据，了解生态正在发生什么。'],
  ['/en', 'Furry Ecosystem - GoFurry', 'Explore public metrics and recent changes across the Furry website and game ecosystems, with transparent coverage and historical data availability.'],
] as const) test('Overview localized SEO survives SSR and hydration ' + (prefix || 'zh'), async ({ page, runtime }) => {
  const html = await openRuntime(page, prefix + '/insights')
  const head = html.match(/<head>([\s\S]*?)<\/head>/)?.[1] ?? ''
  expect(head).toContain(`<title>${title}</title>`)
  await expect(page).toHaveTitle(title)
  for (const [attribute, key, content] of [['name', 'description', description], ['property', 'og:description', description], ['property', 'og:title', title]] as const) {
    expect(head).toContain(`<meta ${attribute}="${key}" content="${content}">`)
    const selector = `meta[${attribute}="${key}"]`
    await expect(page.locator(selector)).toHaveCount(1)
    await expect(page.locator(selector)).toHaveAttribute('content', content)
  }
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  runtime.assertQuiet()
})

test('Overview two SSR sources start independently', async ({ request, runtime }) => {
  const gates = sources.map(source => runtime.hold(url => url.pathname === source))
  const pending = request.get('/insights')
  try { await Promise.all(gates.map(gate => gate.wait())) } finally { gates.forEach(gate => gate.release()) }
  expect((await pending).status()).toBe(200)
  await Promise.all(gates.map(gate => gate.done()))
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  runtime.assertQuiet()
})

for (const candidateCase of ['adult', 'tagged', 'missing-tags', 'no-image', 'stale', 'invalid-time', 'invalid-count', 'zero', 'long-title']) {
  test('Overview rejects unqualified event evidence with ' + candidateCase, async ({ page, runtime }) => {
    runtime.state.candidateCase = candidateCase
    await page.setViewportSize({ width: 390, height: 900 })
    await openRuntime(page, '/insights')
    await expect(page.locator('[data-overview-hero]')).toHaveAttribute('data-hero-mode', 'data')
    await expect(page.locator('[data-overview-games] img, [data-overview-activity] [data-domain="game"] img')).toHaveCount(0)
    // Even injected, uncontracted evidence cannot qualify Overview event art
    // or replace either ecosystem's authoritative metric snapshot.
    await expect(page.locator('[data-overview-games] [data-ecosystem-count]')).toHaveText('213')
    await expect(page.locator('[data-overview-games] [data-ecosystem-metric]')).toHaveCount(3)
    await expect(page.locator('[data-overview-games]')).not.toContainText('game fixture')
    await expect(page.locator('[data-overview-ecosystems] a[href*="/games/91"], [data-overview-ecosystems] [data-pulse]')).toHaveCount(0)
    await expect(page.locator('[data-overview-activity] [data-domain="game"]')).toHaveCount(2)
    await expect(page.locator('[data-overview-activity] [data-domain="game"]').first()).toHaveAttribute('href', '/games/82')
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
      await expect(page.locator('[data-hero-domain="game"]')).toHaveAttribute('data-hero-metric', 'free')
      await expect(page.locator('[data-hero-domain="game"] progress')).toHaveAttribute('value', '0')
      await expect(page.locator('[data-ecosystem-metric="windows"] [data-ecosystem-value]')).toHaveText('0.0%')
      await expect(page.locator('[data-ecosystem-metric="tls13"] [data-ecosystem-value]')).toHaveText('0.0%')
    }
    if (metricCase === 'fallback') {
      await expect(page.locator('[data-hero-domain="site"]')).toHaveAttribute('data-hero-metric', 'tls13')
      await expect(page.locator('[data-hero-domain="site"] a')).toHaveAttribute('href', '/en/insights/sites?metric=tls13')
      expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['http2', 'hsts', 'csp'])
      await expect(page.locator('[data-hero-domain="game"]')).toHaveAttribute('data-hero-metric', 'windows')
      await expect(page.locator('[data-hero-domain="game"] a')).toHaveAttribute('href', '/en/insights/games?metric=windows')
      expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['mac', 'linux'])
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
      await expect(page.locator('[data-overview-games] [data-ecosystem-empty]')).toHaveCount(1)
      await expect(page.locator('[data-overview-games] [data-ecosystem-metric]')).toHaveCount(0)
      expect(stats(html)).toEqual(['238', '213', '47'])
    }
    if (metricCase === 'sparse') {
      expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['linux'])
      expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['csp', 'certificate_verified'])
      await expect(page.locator('[data-hero-domain="game"]')).toHaveAttribute('data-hero-metric', 'mac')
      await expect(page.locator('[data-hero-domain="game"] progress')).toHaveAttribute('value', '0')
    }
    if (metricCase === 'duplicate') {
      expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['windows', 'mac', 'linux'])
      expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['tls13', 'http2', 'hsts'])
    }
    for (const [domain, section] of [['game', 'games'], ['site', 'sites']]) {
      const heroKey = await page.locator(`[data-hero-domain="${domain}"]`).getAttribute('data-hero-metric')
      if (heroKey) await expect(page.locator(`[data-overview-${section}] [data-ecosystem-metric="${heroKey}"]`)).toHaveCount(0)
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
for (const count of [0, 1, 2, 3, 4, 5, 8]) test('Overview merged ordinary events stay bounded ' + count, async ({ page, runtime }) => {
  runtime.state.eventCount = count
  await openRuntime(page, '/insights')
  await expect(page.locator('[data-overview-activity] [data-change-link]')).toHaveCount(Math.min(count, 5))
  await expect(page.locator('[data-activity-all]')).toHaveAttribute('href', '/insights/changes')
  await expect(page.locator('[data-activity-empty]')).toHaveCount(count === 0 ? 1 : 0)
  await expect(page.locator('[data-activity-total]')).toHaveText('47')
  await expect(page.locator('[data-overview-activity]')).toHaveAttribute('data-activity-state', 'complete')
  runtime.assertQuiet()
})
for (const prefix of ['', '/en']) for (const source of ['nav', 'game']) test(`Overview partial empty feed is distinct ${prefix} ${source}`, async ({ page, runtime }) => {
  runtime.state.failure = source
  runtime.state.eventCount = 0
  await openRuntime(page, prefix + '/insights')
  await expect(page.locator('[data-activity-partial]')).toContainText(prefix ? (source === 'nav' ? 'Website observations are unavailable' : 'Game observations are unavailable') : (source === 'nav' ? '网站观测暂不可用' : '游戏观测暂不可用'))
  await expect(page.locator('[data-activity-empty]')).toHaveText(prefix ? 'The available source has no recent public changes.' : '可用来源近期暂无可展示变化。')
  await expect(page.locator('[data-activity-unavailable]')).toHaveCount(0)
  await expect(page.locator('[data-activity-total]')).toHaveText('—')
  await expect(page.locator('[data-activity-all]')).toHaveAttribute('href', prefix + '/insights/changes')
  await expect(page.locator('[data-directory-link]')).toHaveCount(6)
  runtime.assertQuiet()
})
test.describe('Overview event precision across browser timezones', () => {
  test.use({ timezoneId: 'America/Los_Angeles' })
  for (const prefix of ['', '/en']) test('Overview day and exact events preserve SSR text ' + prefix, async ({ page, runtime }) => {
    const html = await openRuntime(page, prefix + '/insights')
    const exact = page.locator(`[data-change-link][href="${prefix}/games/82"] time`)
    const day = page.locator(`[data-change-link][href="${prefix}/site/41"] time`)
    await expect(exact).toHaveAttribute('datetime', '2026-09-01T12:00:00Z')
    await expect(exact).toContainText(prefix ? '12:00 PM UTC' : '12:00 UTC')
    await expect(day).toHaveAttribute('datetime', '2026-09-01')
    await expect(day).toHaveText('2026-09-01')
    expect(html).toContain(await exact.innerText())
    expect(html).not.toContain('datetime="2026-09-01T00:00:00')
    runtime.assertQuiet()
  })
})
async function layout(page: Page, width: number) {
  const result = await page.evaluate(() => {
    const site = document.querySelector('[data-overview-sites]')!.getBoundingClientRect()
    const game = document.querySelector('[data-overview-games]')!.getBoundingClientRect()
    const heroGame = document.querySelector('[data-hero-domain="game"]')!.getBoundingClientRect()
    const heroSite = document.querySelector('[data-hero-domain="site"]')!.getBoundingClientRect()
    const box = (selector: string) => document.querySelector(selector)!.getBoundingClientRect().toJSON()
    return { overflow: Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - innerWidth, site: site.toJSON(), game: game.toJSON(), heroGame: heroGame.toJSON(), heroSite: heroSite.toJSON(), header: box('[data-overview-header]'), hero: box('[data-overview-hero]'), ecosystems: box('[data-overview-ecosystems]'), activity: box('[data-overview-activity]'), activityHeading: box('[data-activity-heading]'), activityContent: box('[data-activity-content]'), explore: box('[data-overview-explore]'), gameDirectory: box('[data-directory-domain="game"]'), siteDirectory: box('[data-directory-domain="site"]'), events: [...document.querySelectorAll('[data-change-link]')].map(el => el.getBoundingClientRect().toJSON()) }
  })
  expect(result.overflow).toBeLessThanOrEqual(0)
  expect(result.hero.top - result.header.bottom).toBeCloseTo(width >= 1200 ? 28 : 24, 0)
  expect(result.ecosystems.top - result.hero.bottom).toBeCloseTo(56, 0)
  expect(result.activity.top).toBeGreaterThanOrEqual(result.ecosystems.bottom)
  expect(result.explore.top).toBeGreaterThanOrEqual(result.activity.bottom)
  if (width >= 1200) {
    expect(result.heroGame.width / result.heroSite.width).toBeGreaterThan(1.9)
    expect(result.heroGame.width / result.heroSite.width).toBeLessThan(2.2)
    expect(Math.abs(result.heroGame.top - result.heroSite.top)).toBeLessThan(2)
    expect(result.activityContent.width / result.activityHeading.width).toBeGreaterThan(2.8)
    expect(result.activityContent.width / result.activityHeading.width).toBeLessThan(3.5)
    expect(Math.abs(result.activityHeading.top - result.activityContent.top)).toBeLessThan(2)
  } else {
    expect(result.heroSite.top).toBeGreaterThanOrEqual(result.heroGame.bottom)
    expect(result.activityContent.top).toBeGreaterThanOrEqual(result.activityHeading.bottom)
  }
  if (width >= 960) {
    expect(result.game.width / result.site.width).toBeCloseTo(1, 1)
    expect(Math.abs(result.site.top - result.game.top)).toBeLessThan(2)
    expect(result.site.left).toBeGreaterThan(result.game.left)
    expect(result.gameDirectory.width / result.siteDirectory.width).toBeCloseTo(1, 1)
    expect(Math.abs(result.gameDirectory.top - result.siteDirectory.top)).toBeLessThan(2)
    expect(result.siteDirectory.left).toBeGreaterThan(result.gameDirectory.left)
  } else {
    expect(result.site.top).toBeGreaterThanOrEqual(result.game.bottom)
    expect(result.siteDirectory.top).toBeGreaterThanOrEqual(result.gameDirectory.bottom)
  }
  for (let index = 1; index < result.events.length; index++) expect(result.events[index]!.top).toBeGreaterThanOrEqual(result.events[index - 1]!.bottom)
}
async function heroAppearance(page: Page, width: number) {
  const result = await page.locator('[data-overview-hero]').evaluate(hero => {
    const rect = (el: Element) => el.getBoundingClientRect().toJSON()
    const game = hero.querySelector('[data-hero-domain="game"]')!, site = hero.querySelector('[data-hero-domain="site"]')!
    const primary = game.querySelector('[data-hero-primary]')!, evidence = game.querySelector('[data-hero-evidence]')!
    const gameValue = game.querySelector('[data-hero-value]')!, siteValue = site.querySelector('[data-hero-value]')!
    return {
      primary: rect(primary), evidence: rect(evidence), facts: [...evidence.children].map(rect),
      sitePrimary: rect(site.querySelector('[data-hero-primary]')!), siteEvidence: rect(site.querySelector('[data-hero-evidence]')!),
      gameSize: parseFloat(getComputedStyle(gameValue).fontSize), siteSize: parseFloat(getComputedStyle(siteValue).fontSize),
      distinctSurface: getComputedStyle(game).backgroundColor !== getComputedStyle(site).backgroundColor,
      shadows: [game, site].map(el => getComputedStyle(el).boxShadow),
      textFits: [gameValue, siteValue].every(el => {
        const range = document.createRange(); range.selectNodeContents(el)
        const text = range.getBoundingClientRect(), box = el.getBoundingClientRect()
        const panel = el.closest('[data-hero-domain]')!.getBoundingClientRect()
        // Font ascenders/descenders can exceed the CSS line box without clipping.
        // Require the complete text width and its vertical extent inside the panel.
        return text.width <= box.width + 1 && text.left >= panel.left && text.right <= panel.right
          && text.top >= panel.top && text.bottom <= panel.bottom
      }),
      evidenceFits: [...hero.querySelectorAll('[data-hero-evidence], [data-hero-fact], dd')].every(el => el.scrollWidth <= el.clientWidth + 1),
      linksBelow: [game, site].every(el => el.querySelector('a')!.getBoundingClientRect().top >= el.querySelector('[data-hero-evidence]')!.getBoundingClientRect().bottom),
    }
  })
  expect(result).toMatchObject({ distinctSurface: true, shadows: ['none', 'none'], textFits: true, evidenceFits: true, linksBelow: true })
  expect(result.gameSize).toBeGreaterThan(result.siteSize)
  if (width >= 1200) {
    expect(result.primary.width / result.evidence.width).toBeCloseTo(65 / 35, 1)
    expect(Math.abs(result.primary.top - result.evidence.top)).toBeLessThan(2)
    expect(result.evidence.left).toBeGreaterThanOrEqual(result.primary.right)
    expect(result.gameSize).toBeGreaterThanOrEqual(86)
    expect(result.siteSize).toBeGreaterThanOrEqual(58)
    expect(result.facts[1]!.top).toBeGreaterThanOrEqual(result.facts[0]!.bottom)
  } else {
    expect(result.evidence.top).toBeGreaterThanOrEqual(result.primary.bottom)
    expect(Math.abs(result.facts[0]!.top - result.facts[1]!.top)).toBeLessThan(2)
    expect(result.facts[1]!.left).toBeGreaterThanOrEqual(result.facts[0]!.right)
    if (width === 390) {
      expect(result.gameSize).toBeGreaterThanOrEqual(56); expect(result.gameSize).toBeLessThanOrEqual(68)
      expect(result.siteSize).toBeGreaterThanOrEqual(48); expect(result.siteSize).toBeLessThanOrEqual(56)
    }
  }
  expect(result.siteEvidence.top).toBeGreaterThanOrEqual(result.sitePrimary.bottom)
  for (const panel of await page.locator('[data-hero-domain]').all()) {
    await expect(panel.locator('h2')).toHaveCount(1)
    await expect(panel.locator('progress')).toHaveAttribute('aria-labelledby', await panel.locator('h2').getAttribute('id') as string)
    expect(await panel.locator('[data-hero-fact]').evaluateAll(elements => elements.map(el => el.getAttribute('data-hero-fact')))).toEqual(['delta', 'coverage', 'sample', 'date'])
    await expect(panel.locator('dt')).toHaveCount(4); await expect(panel.locator('dd')).toHaveCount(4)
  }
}

for (const prefix of ['', '/en']) for (const width of [1440, 390]) test(`Overview full ratio and long evidence remain readable ${prefix} ${width}`, async ({ page, runtime }) => {
  runtime.state.metricCase = 'full-long'
  await page.setViewportSize({ width, height: 1000 })
  const html = await openRuntime(page, prefix + '/insights')
  for (const [domain, path, metric] of [['game', 'games', 'windows'], ['site', 'sites', 'certificate_verified']]) {
    const panel = page.locator(`[data-hero-domain="${domain}"]`)
    await expect(panel.locator('[data-hero-value]')).toHaveText('100.0%')
    await expect(panel.locator('progress')).toHaveAttribute('value', '1')
    await expect(panel.locator('[data-hero-delta]')).toHaveText('—')
    await expect(panel.locator('[data-hero-fact="sample"] dd')).toHaveText('9007199254740991 / 9007199254740991')
    expect(html).toContain('9007199254740991 / 9007199254740991')
    await expect(panel.locator('a')).toHaveAttribute('href', `${prefix}/insights/${path}?metric=${metric}`)
  }
  await heroAppearance(page, width); await layout(page, width)
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  runtime.assertQuiet()
})

for (const sampleCase of ['zero', 'negative', 'fraction', 'over-eligible']) test('Overview evidence only displays valid samples ' + sampleCase, async ({ page, runtime }) => {
  runtime.state.sampleCase = sampleCase
  await openRuntime(page, '/en/insights')
  const samples = page.locator('[data-overview-hero] [data-hero-fact="sample"] dd')
  await expect(samples).toHaveCount(sampleCase === 'zero' ? 2 : 0)
  if (sampleCase === 'zero') for (const sample of await samples.all()) await expect(sample).toHaveText('0 / 0')
  await expect(page.locator('[data-overview-hero] progress')).toHaveCount(2)
  runtime.assertQuiet()
})

for (const failure of ['nav', 'game', 'all']) for (const width of [1440, 390]) test(`Overview missing Hero shrinks naturally ${failure} ${width}`, async ({ page, runtime }) => {
  await page.setViewportSize({ width, height: 1000 })
  await openRuntime(page, '/en/insights')
  const panels = page.locator('[data-hero-domain]')
  const ready = await panels.evaluateAll(elements => elements.map(el => el.getBoundingClientRect().height))
  runtime.state.failure = failure
  await openRuntime(page, '/en/insights')
  for (const [index, domain, source, path] of [[0, 'game', 'game', 'games'], [1, 'site', 'nav', 'sites']] as const) {
    if (failure !== 'all' && failure !== source) continue
    const panel = page.locator(`[data-hero-domain="${domain}"]`)
    expect((await panel.boundingBox())!.height).toBeLessThan(ready[index]! * .85)
    await expect(panel.locator('[data-hero-unavailable]')).toBeVisible()
    await expect(panel.locator('progress, [data-hero-value], [data-hero-evidence]')).toHaveCount(0)
    await expect(panel.locator('a')).toHaveAttribute('href', '/en/insights/' + path)
    await keyboardFocus(panel.locator('a'))
  }
  await layout(page, width); runtime.assertQuiet()
})

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
    // Remote artwork may remain in the DTO; it must not enter the homepage UI.
    expect(html.match(/<main[\s\S]*?<\/main>/)?.[0]).not.toContain('<img')
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
    await expect(page.locator('[data-overview-summary]')).toHaveCount(0)
    await expect(page.locator('[data-activity-heading] [data-activity-total]')).toHaveText('47')
    await expect(page.locator('[data-activity-total]')).toHaveCount(1)
    await expect(page.locator('[data-activity-icon] svg')).toHaveCount(4)
    expect(await page.locator('[data-change-link]').evaluateAll(elements => elements.map(el => el.getAttribute('href')))).toEqual(['/games/82', '/site/41', '/games/83', '/site/42'].map(path => prefix + path))
    await expect(page.locator('[data-overview-explore] h2')).toHaveText(prefix ? 'Explore deeper' : '深入探索')
    await expect(page.locator('[data-directory-domain="game"] [data-directory-link]')).toHaveCount(4)
    await expect(page.locator('[data-directory-domain="site"] [data-directory-link]')).toHaveCount(2)
    expect(await page.locator('[data-directory-link]').evaluateAll(elements => elements.map(el => el.getAttribute('href')))).toEqual(['/insights/games/players', '/insights/games/prices', '/insights/games/languages', '/insights/games/compare', '/insights/sites/certificates', '/insights/sites/compare'].map(path => prefix + path))
    for (const path of ['/insights/games', '/insights/sites', '/insights/changes']) await expect(page.locator(`[data-overview-explore] a[href="${prefix}${path}"]`)).toHaveCount(0)
    await expect(page.locator('[data-activity-all]')).toHaveAttribute('href', prefix + '/insights/changes')
    await expect(page.locator('[data-overview-games] [data-ecosystem-count="game"]')).toHaveText('213')
    await expect(page.locator('[data-overview-sites] [data-ecosystem-count="site"]')).toHaveText('238')
    await expect(page.locator('[data-ecosystem-count]')).toHaveCount(2)
    for (const [domain, path] of [['games', '/insights/games'], ['sites', '/insights/sites']]) {
      await expect(page.locator(`[data-overview-${domain}] [data-ecosystem-link]`)).toHaveAttribute('href', prefix + path)
    }
    expect(await page.locator('[data-overview-games] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['windows', 'mac', 'linux'])
    expect(await page.locator('[data-overview-sites] [data-ecosystem-metric]').evaluateAll(elements => elements.map(el => el.getAttribute('data-ecosystem-metric')))).toEqual(['tls13', 'http2', 'hsts'])
    await expect(page.locator('[data-overview-sites] [data-ecosystem-metric="ipv6"]')).toHaveCount(0)
    await expect(page.locator('[data-overview-games] [data-ecosystem-metric="free"]')).toHaveCount(0)
    await expect(page.locator('[data-ecosystem-metric="hsts"] [data-ecosystem-value]')).toHaveText('0.0%')
    await expect(page.locator('[data-ecosystem-metric="tls13"] [data-ecosystem-coverage]')).toHaveText('90.0%')
    await expect(page.locator('[data-ecosystem-metric="tls13"] time')).toHaveAttribute('datetime', '2026-08-30')
    await expect(page.locator('[data-ecosystem-metric="windows"] time')).toHaveAttribute('datetime', '2026-08-28')
    await expect(page.locator('[data-overview-activity] time').filter({ hasText: /^2026-09-01$/ })).toHaveAttribute('datetime', '2026-09-01')
    expect(await page.locator('progress').evaluateAll(elements => elements.every(el => el.value >= 0 && el.value <= 1 && el.max === 1 && el.getAttribute('aria-labelledby')))).toBe(true)
    await layout(page, width); await heroAppearance(page, width); await revealImages(page)
    if (process.env.GOFURRY_OVERVIEW_REVIEW_DIR && (width === 1440 || width === 390)) {
      await mkdir(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, { recursive: true })
      const name = `${prefix ? 'en' : 'zh'}-${theme}-${width}`
      await page.screenshot({ path: join(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, `cover-${name}.png`) })
      await page.locator('[data-overview-hero]').screenshot({ path: join(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, `hero-${name}.png`) })
    }
    await keyboardFocus(page.locator('.insights-primary-nav a').last())
    await keyboardFocus(page.locator('[data-hero-domain="game"] a'))
    await keyboardFocus(page.locator('[data-hero-domain="site"] a'))
    for (const link of await page.locator('[data-overview-hero] a').all()) {
      await link.scrollIntoViewIfNeeded()
      const before = await link.boundingBox()
      await link.hover()
      await expect(link).toHaveCSS('text-decoration-line', 'underline')
      await expect(link).toHaveCSS('transform', 'none')
      expect(await link.boundingBox()).toEqual(before)
      await keyboardFocus(link)
      expect(await link.evaluate(el => {
        const css = getComputedStyle(el)
        return el.matches(':focus-visible') && css.outlineStyle === 'solid' && css.boxShadow !== 'none'
      })).toBe(true)
    }
    await keyboardFocus(page.locator('[data-overview-games] [data-ecosystem-link]'))
    await keyboardFocus(page.locator('[data-overview-sites] [data-ecosystem-link]'))
    await keyboardFocus(page.locator('[data-change-link]').first())
    await keyboardFocus(page.locator('[data-activity-all]'))
    for (const link of await page.locator('[data-directory-link]').all()) await keyboardFocus(link)
    await assertRuntimeSurface(page, '.insights-overview-page', theme)
    expect(await page.locator('a a').count()).toBe(0)
    expect(gameArt).toEqual([])
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
    runtime.assertQuiet()
  })
}
for (const [source, values] of [['nav', ['—', '213', '—']], ['game', ['238', '—', '—']], ['all', ['—', '—', '—']]] as const) {
  test('Overview independent failure ' + source, async ({ page, runtime }) => {
    runtime.state.failure = source
    const browserSources: string[] = []
    page.on('request', request => { if (sources.includes(new URL(request.url()).pathname)) browserSources.push(request.url()) })
    const html = await openRuntime(page, '/insights')
    expect(stats(html)).toEqual(values)
    expect(html.includes('data-ecosystem-metric="windows"')).toBe(source !== 'game' && source !== 'all')
    expect(html.includes('data-ecosystem-metric="tls13"')).toBe(source !== 'nav' && source !== 'all')
    expect(html.includes('data-change-link')).toBe(source !== 'all')
    await expect(page.locator('[data-overview-activity]')).toHaveAttribute('data-activity-state', source === 'all' ? 'unavailable' : source === 'nav' ? 'game-only' : 'site-only')
    await expect(page.locator('[data-activity-total]')).toHaveText('—')
    await expect(page.locator('[data-activity-partial]')).toHaveCount(source === 'all' ? 0 : 1)
    await expect(page.locator('[data-activity-unavailable]')).toHaveCount(source === 'all' ? 1 : 0)
    await expect(page.locator('[data-activity-all]')).toHaveAttribute('href', '/insights/changes')
    await expect(page.locator('[data-hero-domain="game"] [data-hero-unavailable]')).toHaveCount(source === 'game' || source === 'all' ? 1 : 0)
    await expect(page.locator('[data-hero-domain="site"] [data-hero-unavailable]')).toHaveCount(source === 'nav' || source === 'all' ? 1 : 0)
    await expect(page.locator('[data-overview-hero] a')).toHaveCount(2)
    await expect(page.locator('[data-overview-ecosystems] [data-ecosystem-link]')).toHaveCount(2)
    await expect(page.locator('[data-overview-games] [data-ecosystem-unavailable]')).toHaveCount(source === 'game' || source === 'all' ? 1 : 0)
    await expect(page.locator('[data-overview-sites] [data-ecosystem-unavailable]')).toHaveCount(source === 'nav' || source === 'all' ? 1 : 0)
    // The unchanged Overview services retain ofetch's one retry on 503;
    // hydration must add no source request.
    const expected = [...sources]
    if (source === 'nav' || source === 'all') expected.push(sources[0]!)
    if (source === 'game' || source === 'all') expected.push(sources[1]!)
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(expected.sort())
    expect(browserSources).toEqual([])
    await layout(page, 1440)
    if (process.env.GOFURRY_OVERVIEW_REVIEW_DIR) {
      await mkdir(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, { recursive: true })
      await page.locator('[data-overview-hero]').screenshot({ path: join(process.env.GOFURRY_OVERVIEW_REVIEW_DIR, `hero-failure-${source}.png`) })
    }
    runtime.assertQuiet()
  })
}
for (const [prefix, locale] of [['', 'zh'], ['/en', 'en']]) test('Overview has no Game Home dependency or unused payload state ' + locale, async ({ page, runtime }) => {
  // An absent dependency cannot delay SSR. Any attempt to read Home is rejected
  // by the fixture allowlist and fails assertQuiet, including after hydration.
  const gate = runtime.hold(url => url.pathname === '/api/v2/game/home')
  const html = await openRuntime(page, prefix + '/insights')
  expect(gate.received).toBe(false)
  expect(stats(html)).toEqual(['238', '213', '47']); expect(html).not.toContain('data-pulse="players"')
  expect(html).toContain('data-ecosystem-metric="windows"'); expect(html).toContain('data-ecosystem-metric="tls13"')
  const snapshots = await page.evaluate(() => {
    const root = document.querySelector('#__nuxt') as Element & { __vue_app__: { $nuxt: { payload: { data: Record<string, object> } } } }
    return Object.entries(root.__vue_app__.$nuxt.payload.data)
      .filter(([key]) => key.startsWith('insights:overview:'))
      .map(([key, value]) => [key, Object.keys(value).sort()])
  })
  expect(snapshots).toEqual([[`insights:overview:${locale}`, ['game', 'nav']]])
  expect(runtime.count('/game/home')).toBe(0)
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual([...sources].sort())
  runtime.assertQuiet()
})
test('Overview neutral Site events ignore failing artwork and preserve identity and data Hero', async ({ page, runtime }) => {
  runtime.state.siteHero = true
  await page.setViewportSize({ width: 390, height: 1000 })
  const art: string[] = []
  page.on('request', request => { if (request.resourceType() === 'image' && (/\/media\/game|\/nav\/sites\/|\/defaultLogo\.svg/.test(request.url()))) art.push(request.url()) })
  await openRuntime(page, '/insights')
  await expect(page.locator('[data-change-link]').first()).toHaveAttribute('data-domain', 'site')
  await expect(page.locator('.insight-activity-item--hero')).toHaveCount(0)
  await expect(page.locator('[data-change-link]').first()).toHaveAttribute('href', '/site/41')
  await expect(page.locator('[data-activity-icon] svg')).toHaveCount(4)
  await expect(page.locator('[data-overview-activity] img')).toHaveCount(0)
  runtime.failImages = true
  await openRuntime(page, '/insights'); await revealImages(page)
  await expect(page.locator('main .insight-entity-media img')).toHaveCount(0)
  await expect(page.locator('[data-change-link]').first()).toContainText('Site fixture')
  await expect(page.locator('[data-change-link]').first().locator('time')).toHaveAttribute('datetime', '2026-09-02T12:00:00Z')
  expect(art).toEqual([])
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
