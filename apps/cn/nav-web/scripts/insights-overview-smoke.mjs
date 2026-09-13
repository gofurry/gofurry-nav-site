import assert from 'node:assert/strict'
import { mkdtemp } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { mockOverview, mockGameHome } from './fixtures/insights-overview.mjs'

const sourcePaths = {
  nav: '/api/v2/nav/insights/overview',
  game: '/api/v2/game/insights/overview',
  panel: '/api/v2/game/home',
}

// Invoked by insights:smoke -- --overview-fixtures. This deliberately tests only the
// Overview against local fixtures; it does not stand in for the full live-data smoke.
export async function runOverviewSmoke() {
  let failure = ''
  let siteHero = false
  let panelDelay = 0
  const requests = []
  const app = await startInsightsFixtureApp(async (url, mediaBase) => {
    const path = url.pathname
    requests.push(path)
    const source = Object.keys(sourcePaths).find(key => sourcePaths[key] === path)
    if (!source || failure === source || failure === 'all') return { status: 503 }
    if (source === 'panel' && panelDelay) await new Promise(resolve => setTimeout(resolve, panelDelay))
    const data = source === 'panel' ? mockGameHome(mediaBase) : mockOverview(source === 'nav' ? 'site' : 'game', mediaBase)
    if (siteHero && source === 'nav') data.recent_changes[0].occurred_at = '2026-09-02T12:00:00Z'
    return { data }
  })
  const { base, upstreamUrl } = app
  let browser
  try {
    for (const route of ['/insights', '/en/insights']) {
      requests.length = 0
      const response = await fetch(base + route)
      const html = await response.text()
      assert.equal(response.status, 200, `${route} HTTP status`)
      assert.match(html, /<h1>[^<]+<\/h1>/, 'visible localized H1 must SSR')
      assert.deepEqual(stats(html), ['238', '213', '47'], 'independent stats must SSR')
      for (const section of ['header', 'activity', 'sites', 'games', 'explore']) assert(html.includes(`data-overview-${section}`), `missing ${section}`)
      const prefix = route.startsWith('/en/') ? '/en' : ''
      for (const path of ['/site/41', '/site/42', '/games/82', '/games/83', '/games/91', '/games/92', '/games/93']) assert(html.includes(`href="${prefix}${path}"`), `missing localized ${path}`)
      assert(html.includes(`${upstreamUrl}/nav/sites/41/icon/${'a'.repeat(32)}.svg`) && html.includes('/defaultLogo.svg'), 'Site refs/default must SSR')
      assert(html.includes('data-media-state="fallback"'), 'missing Game header must SSR a neutral fallback')
      assert.match(html, /<time datetime="2026-09-01T10:00:00.000Z"/, 'earlier snapshot time must SSR')
      assert.deepEqual(requests.filter(path => Object.values(sourcePaths).includes(path)).sort(), Object.values(sourcePaths).sort(), 'Overview must request exactly three independent sources')
      assert(!requests.some(path => /\/(?:sites|games)\/\d+/.test(path)), 'Overview made a per-entity lookup')
      assert(!requests.some(path => path.endsWith('/panel/main')), 'Overview bypassed the prewarmed Home cache')
      console.log(`[overview] SSR ${route}: sections, stats, media, localized links and three requests PASS`)
    }
    for (const [source, expected] of [['nav', ['—', '213', '—']], ['game', ['238', '—', '—']], ['panel', ['238', '213', '47']], ['all', ['—', '—', '—']]]) {
      failure = source
      const response = await fetch(`${base}/insights`)
      const html = await response.text()
      assert.equal(response.status, 200, `${source} failure should stay local`)
      assert.deepEqual(stats(html), expected, `${source} failure fabricated a statistic`)
      assert.equal(html.includes('data-pulse="players"'), source !== 'panel' && source !== 'all', 'Panel failure independence')
      assert.equal(html.includes('data-metric="tls13"'), source !== 'nav' && source !== 'all', 'Site failure independence')
      assert.equal(html.includes('data-change-link'), source !== 'all', 'activity failure independence')
      console.log(`[overview] independent ${source} failure PASS`)
    }
    failure = ''
    panelDelay = 12000
    requests.length = 0
    const slowStarted = performance.now()
    const slowResponse = await fetch(base + '/insights')
    const slowHTML = await slowResponse.text()
    assert.equal(slowResponse.status, 200)
    assert(performance.now() - slowStarted < 11000, 'optional Game panel blocked SSR until the slow upstream completed')
    assert.deepEqual(stats(slowHTML), ['238', '213', '47'], 'slow Game media discarded independent overview facts')
    assert(!slowHTML.includes('data-pulse="players"'), 'timed-out panel fabricated a player observation')
    assert.deepEqual(requests.filter(path => path === sourcePaths.panel), [sourcePaths.panel], 'timed-out panel retried and extended SSR')
    panelDelay = 0
    console.log('[overview] slow cached panel is bounded; independent stats survive PASS')
    browser = await launchPerfBrowser()
    const context = await browser.newContext({ locale: 'zh-CN' })
    const page = await context.newPage()
    const pageErrors = []
    page.on('pageerror', error => pageErrors.push(error.message))
    // Keep fixture runs local even if the shell loads optional third-party resources.
    await page.route('**/*', route => {
      const url = new URL(route.request().url())
      return ['127.0.0.1', 'localhost'].includes(url.hostname) || ['data:', 'blob:'].includes(url.protocol) ? route.continue() : route.abort()
    })
    const screenshots = await mkdtemp(join(tmpdir(), 'gofurry-overview-b2-'))
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const prefix of ['', '/en']) {
        requests.length = 0
        await page.goto(`${base}${prefix}/insights`, { waitUntil: 'networkidle' })
        assert.deepEqual(requests.filter(path => Object.values(sourcePaths).includes(path)).sort(), Object.values(sourcePaths).sort(), 'hydration repeated Overview requests')
        assert(!requests.some(path => /\/(?:sites|games)\/\d+/.test(path)), 'browser added a per-entity request')
        assert(await page.locator('main h1').isVisible(), 'H1 is not visible')
        assert.equal(await page.locator('[data-overview-activity] [data-change-link]').count(), 4, 'fixture activity did not render')
        assert.equal(await page.locator('[data-pulse="players"] strong').textContent(), 'Active game fixture')
        assert((await page.locator('[data-pulse="players"]').textContent()).includes('2,800'), 'available Panel peak was dropped')
        assert((await page.locator('[data-pulse="discount"]').textContent()).includes('5.99'), 'minor-unit USD price was not preserved')
        const progress = await page.locator('progress').evaluateAll(elements => elements.map(el => ({ value: el.value, max: el.max, label: el.getAttribute('aria-labelledby') })))
        assert(progress.every(item => item.value >= 0 && item.value <= 1 && item.max === 1 && item.label), 'progress contract')
        await assertLayout(page, width)
        await revealImages(page)
        await page.screenshot({ path: join(screenshots, `${prefix ? 'en' : 'zh'}-${width}.png`), fullPage: true })
        await page.locator('.insights-primary-nav a').last().focus()
        assert(await page.locator('.insights-primary-nav a').last().evaluate(el => getComputedStyle(el).outlineStyle !== 'none'), 'navigation keyboard focus missing')
        console.log(`[overview] ${prefix || 'zh'} ${width}px layout, media and keyboard focus PASS`)
      }
    }
    siteHero = true
    await page.goto(`${base}/insights`, { waitUntil: 'networkidle' })
    assert.equal(await page.locator('.insight-activity-item--hero').getAttribute('data-domain'), 'site')
    assert((await page.locator('.insight-activity-item--hero img').boundingBox()).width <= 72, 'Site icon was stretched into a cover')
    await page.screenshot({ path: join(screenshots, 'site-hero-mobile.png'), fullPage: true })
    await page.route('**/media/**', route => route.abort())
    await page.route('**/nav/sites/*/icon/*', route => route.abort())
    await page.route('**/defaultLogo.svg', route => route.abort())
    await page.reload({ waitUntil: 'networkidle' })
    await revealImages(page)
    assert.equal(await page.locator('main .insight-entity-media img').count(), 0, 'failed Site/default/Game images did not reach fallback')
    assert(await page.locator('main [role="img"][aria-label]').count() > 0, 'fallback lost its accessible entity name')
    await assertLayout(page, 390)
    await page.screenshot({ path: join(screenshots, 'media-fallback-mobile.png'), fullPage: true })
    await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
    assert(await page.locator('html').evaluate(el => el.classList.contains('dark')), 'existing theme control did not enable dark mode')
    // The existing public background is fixed to the viewport; inspect dark mode
    // in viewports so a full-page capture does not extend beyond that surface.
    await page.screenshot({ path: join(screenshots, 'dark-fallback-mobile.png') })
    await page.locator('[data-overview-games]').scrollIntoViewIfNeeded()
    await page.screenshot({ path: join(screenshots, 'dark-games-mobile.png') })
    assert.deepEqual(pageErrors, [], 'Overview threw a browser exception')
    console.log('[overview] Site hero and image error fallback PASS')
    console.log(`[overview] screenshots (temporary, outside Git): ${screenshots}`)
  } catch (error) {
    console.error(app.logs())
    throw error
  } finally {
    await browser?.close()
    await app.close()
  }
}

function stats(html) {
  const dl = html.match(/<dl class="overview-stats">([\s\S]*?)<\/dl>/)?.[1] || ''
  return [...dl.matchAll(/<dd>(.*?)<\/dd>/g)].map(match => match[1])
}
async function assertLayout(page, width) {
  const layout = await page.evaluate(() => {
    const site = document.querySelector('[data-overview-sites]').getBoundingClientRect()
    const game = document.querySelector('[data-overview-games]').getBoundingClientRect()
    return { overflow: document.documentElement.scrollWidth > window.innerWidth, site: { top: site.top, width: site.width }, game: { top: game.top, width: game.width } }
  })
  assert(!layout.overflow, `${width}px page has horizontal overflow`)
  if (width >= 1200) assert(layout.game.width > layout.site.width * 1.3 && Math.abs(layout.site.top - layout.game.top) < 2, 'desktop lost asymmetric split')
  if (width === 390) assert(layout.game.top > layout.site.top, 'mobile did not stack domains')
}
async function revealImages(page) {
  await page.evaluate(async () => {
    for (let top = 0; top < document.body.scrollHeight; top += 650) {
      window.scrollTo(0, top)
      await new Promise(resolve => setTimeout(resolve, 60))
    }
    window.scrollTo(0, 0)
  })
  await page.waitForTimeout(200)
}
