import assert from 'node:assert/strict'
import { mkdtemp } from 'node:fs/promises'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { changeCategories, changesFixtureResponse } from './fixtures/insights-changes.mjs'

const pathFor = (domain, locale = 'zh', category = '', range = '30d') => `${locale === 'en' ? '/en' : ''}/insights/changes?domain=${domain}&range=${range}${category ? '&category=' + category : ''}`
const events = page => page.locator('.insights-change-explorer-item')
const waitForFeed = page => page.locator('.insights-change-explorer-feed[aria-busy="false"]').waitFor()
const changeResponse = (page, test = () => true) => page.waitForResponse(response => {
  const url = new URL(response.url())
  return url.pathname.endsWith('/insights/changes') && test(url)
})

export async function runChangesSmoke() {
  const state = { failure: false, moreFailure: false, empty: false, delayRange: '' }
  const app = await startInsightsFixtureApp((url, mediaBase) => changesFixtureResponse(url, mediaBase, state))
  const { base, requests } = app
  const dataRequests = () => requests.filter(url => url.pathname.includes('/api/'))
  let browser
  try {
    // Each share URL has its own server-rendered snapshot; no directory or detail request.
    for (const domain of ['site', 'game']) for (const locale of ['zh', 'en']) {
      requests.length = 0
      const response = await fetch(base + pathFor(domain, locale))
      const html = await response.text()
      assert.equal(response.status, 200)
      assert(html.includes(`<h1>${locale === 'en' ? 'Change Explorer' : '变化探索'}</h1>`))
      assert(html.includes('aria-current="page"') && html.includes('insights-primary-nav'))
      assert(!html.includes('insights-domain-nav'), 'global Changes page gained a domain navigation')
      assert(html.includes('data-change-date="2026-09-09"') && html.includes('2026-09-09T13:32:00Z'))
      assert(html.includes(`${locale === 'en' ? '/en' : ''}/${domain === 'site' ? 'site' : 'games'}/31`))
      assert(html.includes(domain === 'site' ? '/nav/sites/31/icon/' : 'change-header.svg'))
      assert(html.includes(domain === 'site' ? '/defaultLogo.svg' : 'data-media-state="fallback"'))
      assert.equal(dataRequests().length, 1, 'SSR exceeded one change-list request')
      console.log(`[changes] SSR ${domain} ${locale}: identity, fallback, time precision, 1 request PASS`)
    }
    browser = await launchPerfBrowser()
    const context = await browser.newContext({ locale: 'zh-CN' })
    const page = await context.newPage()
    const errors = []
    page.on('pageerror', error => errors.push(error.message))
    await page.route('**/*', route => {
      const url = new URL(route.request().url())
      return ['127.0.0.1', 'localhost'].includes(url.hostname) || ['data:', 'blob:'].includes(url.protocol) ? route.continue() : route.abort()
    })
    const screenshots = await mkdtemp(join(tmpdir(), 'gofurry-changes-b6-'))
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const domain of ['site', 'game']) for (const locale of ['zh', 'en']) {
        requests.length = 0
        await page.goto(base + pathFor(domain, locale), { waitUntil: 'networkidle' })
        assert.equal(dataRequests().length, 1, 'hydration repeated the initial change request')
        assert(await page.locator('.insights-container h1').isVisible())
        assert.equal(await events(page).count(), 4, 'repeated-entity events were removed')
        assert.equal(await events(page).nth(2).locator('strong').innerText(), '#32', 'missing identity lost its ID fallback')
        assert.equal(await events(page).nth(2).locator('.insight-entity-media img, .insight-entity-media [role="img"]').evaluate(el => el.getAttribute('alt') ?? el.getAttribute('aria-label')), '#32', 'missing-name media lost its accessible fallback')
        assert.equal(await page.locator('[data-change-date]').count(), 2)
        assert.equal(await events(page).nth(1).locator('time').innerText(), '2026-09-09', 'day precision fabricated time')
        await assertLayout(page, width)
        await page.keyboard.press('Tab')
        await page.locator('[data-category-filter]').last().focus()
        assert(await page.locator('[data-category-filter]').last().evaluate(el => getComputedStyle(el).outlineStyle !== 'none'), 'filter focus invisible')
        await events(page).last().scrollIntoViewIfNeeded()
        await page.evaluate(() => window.scrollTo(0, 0))
        await page.screenshot({ path: join(screenshots, `${domain}-${locale}-${width}.png`), fullPage: true })
        console.log(`[changes] ${domain} ${locale} ${width}px: H1, filters, time, media and no overflow PASS`)
      }
    }

    await page.setViewportSize({ width: 1440, height: 1000 })
    for (const domain of ['site', 'game']) {
      await page.goto(base + pathFor(domain), { waitUntil: 'networkidle' })
      for (const category of Object.keys(changeCategories[domain])) {
        const response = changeResponse(page, url => url.searchParams.get('category') === category)
        await page.locator(`[data-category-filter="${category}"]`).click()
        await response
        await waitForFeed(page)
        assert.equal(new URL(page.url()).searchParams.get('category'), category)
        assert.equal(await events(page).first().getAttribute('data-event-type'), changeCategories[domain][category])
      }
      for (const range of ['7d', '90d', 'all', '30d']) {
        const response = changeResponse(page, url => url.searchParams.get('range') === range)
        await page.locator(`[data-change-range="${range}"]`).click()
        await response
        await waitForFeed(page)
        assert.equal(new URL(page.url()).searchParams.get('range'), range)
      }
      const firstLinks = await events(page).evaluateAll(els => els.map(el => el.getAttribute('href')))
      state.moreFailure = true
      await page.locator('[data-load-more]').click()
      await page.locator('.insights-change-explorer-feed__inline-error').waitFor()
      assert.deepEqual(await events(page).evaluateAll(els => els.map(el => el.getAttribute('href'))), firstLinks, 'load-more failure discarded events')
      assert(!new URL(page.url()).searchParams.has('cursor'))
      state.moreFailure = false
      const next = changeResponse(page, url => url.searchParams.get('cursor') === 'fixture-next-page')
      await page.locator('[data-load-more]').click()
      await next
      await waitForFeed(page)
      assert.equal(await events(page).count(), 6)
      assert.equal(await page.locator('[data-change-date="2026-09-08"] li').count(), 3, 'next page split same-date group')
      assert.equal(await page.locator('[data-load-more]').count(), 0)
      assert.deepEqual((await events(page).evaluateAll(els => els.map(el => el.getAttribute('href')))).slice(0,4), firstLinks, 'pagination reordered previous items')
      assert(!new URL(page.url()).searchParams.has('cursor'))
      const query = new URL(page.url()).search
      await page.locator('button:has(img[alt="EN"])').first().click()
      await page.waitForURL(url => url.pathname === '/en/insights/changes')
      await page.waitForLoadState('networkidle')
      assert.equal(new URL(page.url()).search, query, 'locale lost filters')
      assert((await events(page).first().getAttribute('href')).startsWith('/en/'))
      console.log(`[changes] ${domain}: categories/ranges, cursor retry/order, date merge and locale PASS`)
    }

    await page.goto(base + '/insights/changes?domain=bad&range=bad&category=bad&cursor=leak&keep=yes', { waitUntil: 'networkidle' })
    await page.waitForURL(url => url.searchParams.get('domain') === 'site' && url.searchParams.get('range') === '30d' && !url.searchParams.has('category') && !url.searchParams.has('cursor'))
    assert.equal(new URL(page.url()).searchParams.get('keep'), 'yes')
    const switchResponse = changeResponse(page, url => url.pathname.includes('/game/'))
    await page.locator('[data-domain-filter="game"]').click()
    await switchResponse
    await waitForFeed(page)
    assert.equal(new URL(page.url()).searchParams.get('domain'), 'game')
    assert.equal(await page.locator('[data-category-filter="capability"]').count(), 0)
    // An old range response must never replace a newer category result.
    state.delayRange = '7d'
    const slow = changeResponse(page, url => url.searchParams.get('range') === '7d')
    const slowStarted = page.waitForRequest(request => request.url().includes('/api/') && new URL(request.url()).searchParams.get('range') === '7d')
    await page.locator('[data-change-range="7d"]').click()
    await slowStarted
    const fast = changeResponse(page, url => url.searchParams.get('range') === '90d')
    await page.locator('[data-change-range="90d"]').click()
    await fast
    await waitForFeed(page)
    await slow
    assert((await events(page).first().locator('strong').innerText()).includes('90d'), 'stale range response replaced newer records')
    assert.equal(new URL(page.url()).searchParams.get('range'), '90d')
    assert.equal(await events(page).count(), 4)
    state.delayRange = ''
    console.log('[changes] invalid normalization, domain switch and stale-response protection PASS')

    for (const locale of ['zh', 'en']) {
      state.empty = true
      await page.goto(base + pathFor('site', locale), { waitUntil: 'networkidle' })
      assert.equal(await events(page).count(), 0)
      assert((await page.locator('.insights-changes-state').innerText()).includes(locale === 'en' ? 'No public changes' : '暂无公开变化'))
      state.empty = false
      state.failure = true
      await page.goto(base + pathFor('game', locale), { waitUntil: 'networkidle' })
      assert(await page.locator('[data-retry-changes]').isVisible())
      assert.equal(await events(page).count(), 0)
      state.failure = false
      await page.locator('[data-retry-changes]').click()
      await waitForFeed(page)
      assert.equal(await events(page).count(), 4)
    }
    await page.route('**/media/**', route => route.abort())
    await page.route('**/nav/sites/*/icon/*', route => route.abort())
    await page.route('**/defaultLogo.svg', route => route.abort())
    for (const domain of ['site', 'game']) {
      await page.goto(base + pathFor(domain, 'en'), { waitUntil: 'networkidle' })
      await events(page).last().scrollIntoViewIfNeeded()
      await page.waitForFunction(() => [...document.querySelectorAll('.insights-change-explorer-item .insight-entity-media')].every(el => el.dataset.mediaState === 'fallback'))
      assert.equal(await events(page).locator('[role="img"][aria-label]').count(), 4)
    }
    await page.unroute('**/media/**')
    await page.unroute('**/nav/sites/*/icon/*')
    await page.unroute('**/defaultLogo.svg')
    await page.goto(base + pathFor('game', 'en'), { waitUntil: 'networkidle' })
    await page.getByRole('button', { name: 'Toggle theme icon', exact: true }).first().click()
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      await events(page).first().scrollIntoViewIfNeeded()
      await assertLayout(page, width)
      await page.screenshot({ path: join(screenshots, `game-en-dark-${width}.png`) })
    }
    assert.deepEqual(errors, [], 'browser runtime errors')
    assert(dataRequests().every(url => /^\/api\/v2\/(nav|game)\/insights\/changes$/.test(url.pathname)), 'frontend requested another entity API')
    console.log(`[changes] empty/error/retry, image failures, light/dark and no entity requests PASS; screenshots: ${screenshots}`)
  } finally { await browser?.close(); await app.close() }
}

async function assertLayout(page, width) {
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), `${width}px page overflow`)
  assert(await page.locator('.insights-change-explorer-filters, .insights-changes-day, .insights-change-explorer-item').evaluateAll(els => els.every(el => el.scrollWidth <= el.clientWidth + 1)), `${width}px hidden content overflow`)
  assert(await page.locator('.insights-change-explorer-filters').evaluate(el => getComputedStyle(el).boxShadow === 'none'), 'filters restored a card surface')
}
