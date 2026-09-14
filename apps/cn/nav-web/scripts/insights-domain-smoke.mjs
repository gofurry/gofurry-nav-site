import assert from 'node:assert/strict'
import { mkdtemp } from 'node:fs/promises'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { domainFixtureResponse, domainMetricKeys, domainDimensionKeys } from './fixtures/insights-domain.mjs'

export async function runDomainSmoke() {
  const state = { panelFailure: false, overviewFailure: false, breakdownFailure: false, trendMode: 'normal' }
  const app = await startInsightsFixtureApp((url, mediaBase) => domainFixtureResponse(url, mediaBase, state))
  const { base, requests } = app
  let browser
  const routes = [
    ['/insights/sites?metric=ipv6&range=30d&dimension=country', 'site'],
    ['/en/insights/sites?metric=tls13&range=90d&dimension=group', 'site'],
    ['/insights/games?metric=free&range=30d&dimension=primary_tag', 'game'],
    ['/en/insights/games?metric=linux&range=90d&dimension=tag', 'game'],
  ]
  const dataRequests = () => requests.filter(url => url.pathname.includes('/api/'))
  try {
    for (const [route, domain] of [...routes,
      ['/insights/sites?metric=ipv6&range=30d&dimension=country&slice=CN', 'site'],
      ['/en/insights/games?metric=linux&range=90d&dimension=tag&slice=31', 'game'],
    ]) {
      requests.length = 0
      const response = await fetch(base + route)
      const html = await response.text()
      assert.equal(response.status, 200)
      assert.match(html, /<h1>[^<]+<\/h1>/, 'Domain H1 must be visible in SSR')
      assert(html.includes('insights-domain-nav'), 'B1 Domain navigation missing')
      assert(html.includes(`data-domain-count>${domain === 'site' ? 238 : 213}`), 'header count missing')
      for (const key of domainMetricKeys[domain]) assert(html.includes(`data-metric-key="${key}"`), `missing ${key}`)
      const selectedSlice = new URL(route, base).searchParams.has('slice')
      assert.equal(dataRequests().length, (domain === 'site' ? 3 : 4) + Number(selectedSlice), 'SSR initial request count changed')
      assert.equal(dataRequests().some(url => url.pathname.endsWith('/game/home')), domain === 'game', 'cached Game Home requested outside Game Domain')
      assert(!dataRequests().some(url => url.pathname.endsWith('/panel/main')), 'Domain bypassed the prewarmed Home cache')
      console.log(`[domain] SSR ${route}, ${dataRequests().length} requests PASS`)
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
    const screenshots = await mkdtemp(join(tmpdir(), 'gofurry-domain-b3-'))
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const [route, domain] of routes) {
        requests.length = 0
        await page.goto(base + route, { waitUntil: 'networkidle' })
        await page.locator('.insights-trend canvas').waitFor()
        assert.equal(dataRequests().length, domain === 'site' ? 3 : 4, 'hydration duplicated initial data requests')
        assert(await page.locator('main h1').isVisible())
        assert.equal(await page.locator('[data-metric-key]').count(), domainMetricKeys[domain].length)
        assert.equal(await page.locator('[data-dimension]').count(), domainDimensionKeys[domain].length)
        assert.equal(await page.locator('[data-slice]').count(), 8, 'default bars should show the first eight items')
        assert.equal(await page.locator('[data-dimension-raw]').getAttribute('open'), null, 'raw table should start collapsed')
        const first = page.locator('[data-slice]').first()
        const progress = await first.locator('progress').evaluate(el => ({ value: el.value, max: el.max }))
        assert.equal(progress.max, domain === 'site' ? 1 : 46, 'bar denominator changed its domain semantics')
        assert.equal(progress.value, domain === 'site' ? 0.824 : 46, 'bar value used the wrong signal')
        await page.locator('[data-dimension-raw] summary').click()
        assert.equal(await page.locator('[data-dimension-raw] tbody tr').count(), 10, 'raw data was truncated to the bar limit')
        await assertLayout(page, width, domain)
        if (width === 390) assert(await page.locator('.insight-dimension-raw__scroll').evaluate(el => el.scrollWidth > el.clientWidth), 'raw table did not scroll locally')
        await page.locator('[data-dimension-raw] summary').click()
        await page.locator('[data-domain-activity]').scrollIntoViewIfNeeded()
        await page.waitForTimeout(150)
        await page.evaluate(() => window.scrollTo(0, 0))
        const locale = route.startsWith('/en/') ? 'en' : 'zh'
        await page.screenshot({ path: join(screenshots, `${domain}-${locale}-${width}.png`), fullPage: true })
        await page.keyboard.press('Tab')
        await page.locator('[data-metric-key]').last().focus()
        assert(await page.locator('[data-metric-key]').last().evaluate(el => {
          const box = el.getBoundingClientRect(), rail = el.closest('[data-metric-rail]').getBoundingClientRect()
          return box.left >= rail.left - 1 && box.right <= rail.right + 1 && getComputedStyle(el).outlineStyle !== 'none'
        }), 'metric keyboard focus was hidden by horizontal scrolling')
        console.log(`[domain] ${domain} ${locale} ${width}px layout, chart, raw data, focus PASS`)
      }
    }

    await page.setViewportSize({ width: 1440, height: 1000 })
    for (const domain of ['site', 'game']) {
      const path = `/insights/${domain === 'site' ? 'sites' : 'games'}`
      const metric = domain === 'site' ? 'tls13' : 'linux'
      const dimension = domain === 'site' ? 'group' : 'tag'
      await page.goto(`${base}${path}?metric=invalid&range=bad&dimension=bad&slice=bad&keep=yes`, { waitUntil: 'networkidle' })
      await page.waitForURL(url => url.searchParams.get('metric') === domainMetricKeys[domain][0] && url.searchParams.get('dimension') === domainDimensionKeys[domain][0] && !url.searchParams.has('slice'))
      assert.equal(new URL(page.url()).searchParams.get('keep'), 'yes', 'normalization dropped unrelated query state')
      await page.locator(`[data-metric-key="${metric}"]`).click()
      await page.waitForURL(url => url.searchParams.get('metric') === metric)
      await page.locator('[data-range="90d"]').click()
      await page.waitForURL(url => url.searchParams.get('range') === '90d')
      await page.locator(`[data-dimension="${dimension}"]`).click()
      await page.locator('[data-overlapping]').waitFor()
      await page.locator('[data-slice="31"]').click()
      await page.waitForURL(url => url.searchParams.get('slice') === '31')
      await page.locator('.insights-slice-trend canvas').waitFor()
      await page.locator('[data-slice="31"]').click()
      await page.waitForURL(url => !url.searchParams.has('slice'))
      assert.equal(await page.locator('.insights-slice-trend').count(), 0, 'same slice did not toggle off')
      await page.locator('[data-slice="31"]').click()
      const waitForAllHistory = () => Promise.all([
        page.waitForResponse(response => new URL(response.url()).pathname.endsWith(`/metrics/${metric}/trend`) && new URL(response.url()).searchParams.get('range') === 'all'),
        page.waitForResponse(response => new URL(response.url()).pathname.endsWith(`/breakdown/${dimension}/31/trend`) && new URL(response.url()).searchParams.get('range') === 'all'),
      ])
      const allHistory = waitForAllHistory()
      await page.locator('[data-range="all"]').click()
      await page.waitForURL(url => url.searchParams.get('range') === 'all')
      await allHistory
      // A SPA URL changes before the new locale's async page has mounted.
      const localizedHistory = waitForAllHistory()
      await page.locator('button:has(img[alt="EN"])').first().click()
      await page.waitForURL(url => url.pathname === `/en${path}`)
      await localizedHistory
      await page.waitForLoadState('networkidle')
      await page.locator('.insights-slice-trend canvas').waitFor()
      const localized = new URL(page.url())
      for (const [key, value] of Object.entries({ metric, range: 'all', dimension, slice: '31', keep: 'yes' })) assert.equal(localized.searchParams.get(key), value, `locale switch lost ${key}`)
      requests.length = 0
      const response = await fetch(localized.href)
      const html = await response.text()
      assert.equal(response.status, 200)
      assert(!dataRequests().some(url => url.pathname.endsWith('/trend')), 'all-range history was requested during SSR')
      assert(html.includes('aria-busy="true"'), 'deferred chart did not expose its loading state')
      requests.length = 0
      const deferredRequests = []
      const observeHydration = request => {
        const url = new URL(request.url())
        if (url.pathname.endsWith('/trend') && url.searchParams.get('range') === 'all') deferredRequests.push(url)
      }
      page.on('request', observeHydration)
      await page.reload({ waitUntil: 'networkidle' })
      await page.locator('.insights-slice-trend canvas').waitFor()
      page.off('request', observeHydration)
      assert.equal(deferredRequests.length, 2, `deferred main/slice histories did not each load once after hydration: ${deferredRequests.map(url => url.pathname).join(', ')}`)
      console.log(`[domain] ${domain} query normalization, metric/range/dimension/slice, locale and deferred all PASS`)
    }

    state.panelFailure = true
    await page.goto(`${base}/insights/games?metric=free&range=30d&dimension=primary_tag`, { waitUntil: 'networkidle' })
    assert(await page.locator('.insights-game-pulse__unavailable').isVisible())
    await page.locator('.insights-trend canvas').waitFor()
    assert.equal(await page.locator('[data-slice]').count(), 8)
    assert(await page.locator('[data-domain-activity] [data-change-link]').count() > 0)
    state.panelFailure = false
    for (const [mode, text] of [['empty', '暂无可用历史数据'], ['one', '正在积累历史数据'], ['fail', '生态观测数据暂不可用']]) {
      state.trendMode = mode
      await page.goto(`${base}/insights/sites?metric=ipv6&range=30d&dimension=country`, { waitUntil: 'networkidle' })
      assert(await page.locator('.insights-trend').getByText(text, { exact: true }).isVisible(), `${mode} chart state lost`)
    }
    state.trendMode = 'delayed'
    const responseReady = page.waitForResponse(response => response.url().includes('/tls13/trend'))
    await page.locator('[data-metric-key="tls13"]').click()
    await page.locator('.insights-trend [aria-busy="true"]').waitFor()
    await responseReady
    await page.locator('.insights-trend canvas').waitFor()
    state.trendMode = 'normal'
    await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
    await page.locator('.insights-trend').scrollIntoViewIfNeeded()
    await page.waitForTimeout(200)
    await page.screenshot({ path: join(screenshots, 'site-dark-chart.png') })
    assert.deepEqual(errors, [], 'browser exception in Domain presentation')
    assert(!requests.some(url => /\/(?:sites|games)\/\d+|\/game\/info/.test(url.pathname)), 'Domain added a per-entity lookup')
    console.log('[domain] independent Panel failure, empty/one/error/loading charts PASS; no per-entity request')
    console.log(`[domain] screenshots (outside Git): ${screenshots}`)
  } catch (error) {
    console.error(app.logs())
    throw error
  } finally {
    await browser?.close()
    await app.close()
  }
}

async function assertLayout(page, width, domain) {
  const layout = await page.evaluate(() => {
    const chart = document.querySelector('.insights-trend').getBoundingClientRect()
    const context = document.querySelector('.insight-trend-context').getBoundingClientRect()
    return { overflow: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - window.innerWidth, chart: { width: chart.width, top: chart.top }, context: { width: context.width, top: context.top } }
  })
  assert(layout.overflow <= 1, `${width}px page horizontal overflow`)
  if (width === 1440) assert(layout.chart.width / layout.context.width > (domain === 'site' ? 1.8 : 4.5), 'Domain trend/context proportions regressed')
  if (width === 390) assert(layout.context.top > layout.chart.top, 'mobile trend context did not stack')
}
