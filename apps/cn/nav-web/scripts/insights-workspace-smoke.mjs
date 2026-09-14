import assert from 'node:assert/strict'
import { mkdtemp } from 'node:fs/promises'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { workspaceFixtureResponse, workspaceMetricKeys, workspaceRegionKeys } from './fixtures/insights-workspace.mjs'

const routes = [
  ['/insights/games/players?metric=latest_observed', 'players', 1],
  ['/en/insights/games/players?metric=average_30d', 'players', 1],
  ['/insights/games/prices?region=CN', 'prices', 2],
  ['/en/insights/games/prices?region=US', 'prices', 2],
  ['/insights/games/languages', 'languages', 1],
  ['/en/insights/games/languages', 'languages', 1],
  ['/insights/sites/certificates', 'certificates', 1],
  ['/en/insights/sites/certificates', 'certificates', 1],
]

export async function runWorkspaceSmoke() {
  const state = { failure: '', empty: false }
  const app = await startInsightsFixtureApp((url, media) => workspaceFixtureResponse(url, media, state))
  const { base, requests } = app
  let browser
  const dataRequests = () => requests.filter(url => url.pathname.includes('/api/'))
  try {
    for (const [route, kind, count] of routes) {
      requests.length = 0
      const response = await fetch(base + route), html = await response.text()
      assert.equal(response.status, 200)
      assert.match(html, /<h1>[^<]+<\/h1>/, 'visible H1 missing from SSR')
      assert(html.includes('insights-domain-nav'), 'Domain navigation missing')
      assert.equal(dataRequests().length, count, 'SSR request budget changed')
      if (kind !== 'languages') {
        const prefix = route.startsWith('/en/') ? '/en' : ''
        assert(html.includes(`href="${prefix}/${kind === 'certificates' ? 'site/201' : 'games/101'}"`), 'localized identity link missing')
        assert(html.includes((kind === 'certificates' ? '/nav/sites/201/icon/' + 'a'.repeat(32) + '.svg' : '/media/game-0.svg')), 'identity asset missing from SSR')
        assert(html.includes(kind === 'certificates' ? '/defaultLogo.svg' : 'data-media-state="fallback"'), 'missing-asset fallback missing from SSR')
      }
      if (kind === 'players') assert(html.includes('189') && html.includes('213') && html.includes('88.7%'))
      if (kind === 'prices') assert(html.includes(route.startsWith('/en/') ? '$5.99' : '¥5.99'), 'regional minor-unit price changed')
      console.log(`[workspace] SSR ${route}, ${count} requests PASS`)
    }
    browser = await launchPerfBrowser()
    const context = await browser.newContext({ locale: 'zh-CN', colorScheme: 'light' })
    const page = await context.newPage(), errors = []
    page.on('pageerror', error => errors.push(error.message))
    let failImages = false
    await page.route('**/*', route => {
      const url = new URL(route.request().url())
      if (failImages && (url.pathname.startsWith('/media/') || url.pathname.startsWith('/nav/sites/') || url.pathname === '/defaultLogo.svg')) return route.abort()
      return ['127.0.0.1', 'localhost'].includes(url.hostname) || ['data:', 'blob:'].includes(url.protocol) ? route.continue() : route.abort()
    })
    const screenshots = await mkdtemp(join(tmpdir(), 'gofurry-workspace-b4-'))
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const [route, kind, count] of routes) {
        requests.length = 0
        await page.goto(base + route, { waitUntil: 'networkidle' })
        assert.equal(dataRequests().length, count, 'hydration duplicated a data request')
        assert(await page.locator('main h1').isVisible())
        assert.equal(await page.locator('.insights-domain-nav [aria-current="page"]').count(), 1)
        assert.equal(await page.locator('.insight-workspace-disclosure').getAttribute('open'), null)
        if (kind === 'players' || kind === 'prices') {
          const keys = kind === 'players' ? workspaceMetricKeys : workspaceRegionKeys
          assert.deepEqual(await page.locator('[data-workspace-option]').evaluateAll(els => els.map(el => el.dataset.workspaceOption)), keys)
          assert.equal(await page.locator('[data-workspace-option][aria-pressed="true"]').count(), 1)
          await page.keyboard.press('Tab')
          const last = page.locator('[data-workspace-option]').last()
          await last.focus()
          assert(await last.evaluate(el => {
            const box = el.getBoundingClientRect(), parent = el.parentElement.getBoundingClientRect()
            return getComputedStyle(el).outlineStyle !== 'none' && box.left >= parent.left - 1 && box.right <= parent.right + 1
          }), 'keyboard focus must remain visible within the selector')
        }
        if (kind === 'players') {
          assert.equal(await page.locator('[data-rank]').count(), 20)
          assert.equal(await page.locator('.insight-ranking-row--lead').count(), 3)
          assert((await page.locator('[data-rank="20"]').innerText()).includes('0'), 'genuine zero was lost')
        }
        if (kind === 'languages') {
          assert.equal(await page.locator('[data-language]').count(), 12)
          assert.equal(await page.locator('[data-workspace-raw]').getAttribute('open'), null)
          const signals = await page.locator('[data-language] progress').evaluateAll(els => els.map(el => ({ value: el.value, max: el.max })))
          assert(signals.every(signal => signal.max === 1 && signal.value >= 0 && signal.value <= 1))
          assert.equal(await page.locator('[data-language="th"] progress').count(), 0, 'missing share became a fake zero')
          await page.locator('[data-workspace-raw] summary').click()
          assert.equal(await page.locator('[data-workspace-raw] tbody tr').count(), 14)
          if (width === 390) assert(await page.locator('.insights-workspace-table-scroll').evaluate(el => el.scrollWidth > el.clientWidth), 'raw table must scroll locally')
          await assertLayout(page, width)
          await page.locator('[data-workspace-raw] summary').click()
        }
        if (kind === 'certificates') {
          assert.equal(await page.locator('[data-expiry]').count(), 4)
          assert.equal(await page.locator('[data-expiry-attention] [data-workspace-entity]').count(), 3)
          assert.equal(await page.locator('[data-verification-issues] [data-workspace-entity]').count(), 2)
        }
        await assertLayout(page, width)
        // Load identity thumbnails along the list before capturing the full-page evidence.
        for (const media of await page.locator('.insight-entity-media').all()) await media.scrollIntoViewIfNeeded()
        await page.waitForTimeout(100)
        assert.equal(await page.locator('.insight-entity-media img').evaluateAll(els => els.filter(el => el.complete && !el.naturalWidth).length), 0, 'broken thumbnail remained visible')
        await page.evaluate(() => window.scrollTo(0, 0))
        await page.screenshot({ path: join(screenshots, `${kind}-${route.startsWith('/en/') ? 'en' : 'zh'}-${width}.png`), fullPage: true })
        if (width === 390) {
          await page.locator('.insights-workspace-section').first().scrollIntoViewIfNeeded()
          await page.screenshot({ path: join(screenshots, `${kind}-${route.startsWith('/en/') ? 'en' : 'zh'}-mobile-detail.png`) })
        }
        console.log(`[workspace] ${kind} ${route.startsWith('/en/') ? 'en' : 'zh'} ${width}px layout, identity, raw data PASS`)
      }
    }

    // Existing query defaults and changes remain owned by each page.
    await page.setViewportSize({ width: 1440, height: 1000 })
    for (const [kind, queryKey, keys, endpoint] of [
      ['players', 'metric', workspaceMetricKeys, '/players/ranking'],
      ['prices', 'region', workspaceRegionKeys, '/prices/discounts'],
    ]) {
      await page.goto(`${base}/insights/games/${kind}?${queryKey}=invalid`, { waitUntil: 'networkidle' })
      assert.equal(await page.locator('[data-workspace-option][aria-pressed="true"]').getAttribute('data-workspace-option'), keys[0])
      for (const key of keys.slice(1)) {
        const response = page.waitForResponse(res => new URL(res.url()).pathname.endsWith(endpoint) && new URL(res.url()).searchParams.get(queryKey) === key)
        await page.locator(`[data-workspace-option="${key}"]`).click()
        await response
        await page.waitForURL(url => url.searchParams.get(queryKey) === key)
      }
      await page.locator('button:has(img[alt="EN"])').first().click()
      await page.waitForURL(url => url.pathname === `/en/insights/games/${kind}`)
      await page.waitForLoadState('networkidle')
      assert.equal(new URL(page.url()).searchParams.get(queryKey), keys.at(-1), 'locale switch dropped workspace query')
    }
    await page.setViewportSize({ width: 390, height: 1000 })
    for (const failure of ['/prices/overview', '/prices/discounts']) {
      state.failure = failure
      await page.goto(base + '/en/insights/games/prices?region=US', { waitUntil: 'networkidle' })
      if (failure.endsWith('overview')) {
        assert(await page.locator('[data-price-overview-error]').isVisible())
        assert.equal(await page.locator('.insight-discount-row').count(), 6)
      } else {
        assert(await page.locator('[data-discount-error]').isVisible())
        assert(await page.locator('.insights-workspace-summary').isVisible())
      }
    }
    for (const [kind, failure, route] of [
      ['players', '/players/ranking', '/insights/games/players'],
      ['languages', '/languages/overview', '/insights/games/languages'],
      ['certificates', '/certificates/overview', '/insights/sites/certificates'],
    ]) {
      state.failure = failure
      await page.goto(base + route, { waitUntil: 'networkidle' })
      assert(await page.locator('main .insights-empty-state').isVisible(), `${kind} failure state missing`)
      await assertLayout(page, 390)
    }
    state.failure = ''
    state.empty = true
    for (const [route] of routes.filter((_, index) => index % 2 === 0)) {
      await page.goto(base + route, { waitUntil: 'networkidle' })
      assert(await page.locator('main .insights-empty-state').first().isVisible(), 'empty workspace state missing')
      assert.equal(await page.locator('[data-workspace-entity], [data-language]').count(), 0)
    }
    state.empty = false
    failImages = true
    for (const [route, kind] of routes.filter(([, kind], index) => kind !== 'languages' && index % 2 === 0)) {
      await page.goto(base + route, { waitUntil: 'networkidle' })
      for (const media of await page.locator('.insight-entity-media').all()) await media.scrollIntoViewIfNeeded()
      await page.waitForTimeout(200)
      assert.equal(await page.locator('.insight-entity-media img').count(), 0, `${kind} failed image did not fall back`)
      assert(await page.locator('.insight-entity-media [role="img"][aria-label]').count() > 0, 'fallback lost accessible identity')
      await assertLayout(page, 390)
    }
    failImages = false
    await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const [route, kind] of routes.filter(([route]) => route.startsWith('/en/'))) {
        await page.goto(base + route, { waitUntil: 'networkidle' })
        assert(await page.locator('html').evaluate(el => el.classList.contains('dark')), 'dark theme preference did not persist')
        await assertLayout(page, width)
        await page.screenshot({ path: join(screenshots, `${kind}-en-dark-${width}.png`) })
      }
    }
    assert.deepEqual(errors, [], 'browser exceptions in workspace presentation')
    console.log('[workspace] query, locale, independent errors, empty states, failed images, dark theme PASS; no per-entity request')
    console.log(`[workspace] screenshots (outside Git): ${screenshots}`)
  } catch (error) {
    console.error(app.logs())
    throw error
  } finally {
    await browser?.close()
    await app.close()
  }
}

async function assertLayout(page, width) {
  const overflow = await page.evaluate(() => ({
    page: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth,
    rows: [...document.querySelectorAll('[data-workspace-entity], .insights-workspace-summary, .insights-workspace-header')].some(el => el.scrollWidth > el.clientWidth + 1),
  }))
  assert(overflow.page <= 1 && !overflow.rows, `${width}px workspace horizontal overflow: ${JSON.stringify(overflow)}`)
}
