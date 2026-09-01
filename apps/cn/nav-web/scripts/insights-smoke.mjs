#!/usr/bin/env node
import { launchPerfBrowser, normalizeBaseUrl, parseArgs, toAbsoluteUrl } from './perf/shared.mjs'

const args = parseArgs()
const baseUrl = normalizeBaseUrl(args['base-url'] || process.env.INSIGHTS_BASE_URL || 'http://localhost:3000')
const insightsRoutes = [
  '/insights',
  '/insights/sites?metric=ipv6&range=30d',
  '/insights/games?metric=free&range=30d',
  '/en/insights',
  '/en/insights/sites?metric=ipv6&range=90d',
  '/en/insights/games?metric=free&range=90d',
]
const removedWorkshopRoutes = [
  '/workshop',
  '/workshop/tools',
  '/workshop/developer',
  '/workshop/discussion',
  '/en/workshop',
  '/en/workshop/tools',
  '/en/workshop/developer',
  '/en/workshop/discussion',
]

for (const route of insightsRoutes) {
  const response = await fetch(toAbsoluteUrl(baseUrl, route), { redirect: 'manual' })
  const html = await response.text()
  assert(response.status === 200, `${route} expected HTTP 200, received ${response.status}`)
  assert(html.includes('insights-page'), `${route} did not SSR the Insights page shell`)
  assert(/<h1(?:\s|>)/.test(html), `${route} did not SSR a visible h1`)
  console.log(`[insights] SSR ${route} -> 200`)
}

for (const route of removedWorkshopRoutes) {
  const response = await fetch(toAbsoluteUrl(baseUrl, route), { redirect: 'manual' })
  assert(response.status === 404, `${route} expected ordinary HTTP 404, received ${response.status}`)
  assert(!response.headers.has('location'), `${route} must not redirect`)
  console.log(`[insights] removed ${route} -> 404`)
}

const browser = await launchPerfBrowser()
try {
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 }, locale: 'zh-CN' })
  const page = await context.newPage()
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/sites?metric=invalid&range=bad'), {
    waitUntil: 'domcontentloaded',
    timeout: 30000,
  })
  await page.waitForSelector('.insights-domain-page')
  await page.waitForURL(url => url.searchParams.get('metric') === 'ipv6' && url.searchParams.get('range') === '30d')
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-metric') === 'ipv6', 'invalid metric was not normalized before rendering')

  await page.locator('[data-metric-key="security_txt"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'security_txt')
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-metric') === 'security_txt', 'site metric selection did not update locally')

  await page.locator('[data-range="all"]').click()
  await page.waitForURL(url => url.searchParams.get('range') === 'all')
  assert(await page.locator('.insights-page').count() === 1, 'range interaction blanked the page shell')

  await page.locator('button:has(img[alt="EN"])').first().click()
  await page.waitForURL(url => url.pathname === '/en/insights/sites')
  const localizedUrl = new URL(page.url())
  assert(localizedUrl.searchParams.get('metric') === 'security_txt', 'locale switch dropped metric query state')
  assert(localizedUrl.searchParams.get('range') === 'all', 'locale switch dropped range query state')

  await page.setViewportSize({ width: 390, height: 844 })
  await page.reload({ waitUntil: 'domcontentloaded' })
  await page.waitForSelector('.insights-domain-page')
  const mobileState = await page.evaluate(() => ({
    overflow: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - document.documentElement.clientWidth,
    text: document.body.innerText,
  }))
  assert(mobileState.overflow <= 2, `mobile Insights page overflowed horizontally by ${mobileState.overflow}px`)
  for (const forbidden of ['undefined', 'NaN', 'null%']) {
    assert(!mobileState.text.includes(forbidden), `mobile Insights page exposed ${forbidden}`)
  }

  await page.setViewportSize({ width: 1440, height: 900 })
  await page.route('**/api/v2/nav/insights/metrics/**', async (route) => {
    const url = route.request().url()
    if (url.includes('/ipv6/')) {
      await route.abort('failed')
      return
    }
    const metric = url.includes('/security_txt/') ? 'security_txt' : 'tls13'
    const points = metric === 'security_txt'
      ? [{ date: '2026-08-31', value: 0, coverage: 1 }]
      : []
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          key: metric,
          requested_range: '30d',
          available_from: points[0]?.date ?? null,
          available_through: points[0]?.date ?? null,
          points,
        },
      }),
    })
  })
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/sites?metric=ipv6&range=30d'), { waitUntil: 'domcontentloaded' })
  await page.locator('[data-metric-key="tls13"]').click()
  await page.getByText('暂无可用历史数据', { exact: true }).waitFor()
  await page.locator('[data-metric-key="security_txt"]').click()
  await page.getByText('正在积累历史数据', { exact: true }).waitFor()
  await page.locator('[data-metric-key="ipv6"]').click()
  await page.getByText('洞察数据暂不可用', { exact: true }).first().waitFor()
  console.log('[insights] empty, one-point, zero-value, and request-error semantics passed')

  const mockOverview = (domain) => ({
    generated_at: '2026-09-01T12:00:00Z',
    entity_count: domain === 'site' ? 12 : 8,
    changes_7d: 1,
    metrics: [],
    recent_changes: [{
      type: domain === 'site' ? 'site.ipv6.enabled' : 'game.windows.added',
      date: '2026-09-01',
      occurred_at: domain === 'site' ? null : '2026-09-01T12:00:00Z',
      entity: { id: domain === 'site' ? 41 : 82, name: domain === 'site' ? 'Site fixture' : 'Game fixture' },
      detail: null,
    }],
  })
  await page.route('**/api/v2/nav/insights/overview', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ code: 1, data: mockOverview('site') }),
  }))
  await page.route('**/api/v2/game/insights/overview', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ code: 1, data: mockOverview('game') }),
  }))
  await page.goto(toAbsoluteUrl(baseUrl, '/'), { waitUntil: 'domcontentloaded' })
  await page.getByRole('link', { name: '洞察', exact: true }).first().click()
  await page.waitForURL(url => url.pathname === '/insights')
  const changeHrefs = await page.locator('[data-change-link]').evaluateAll(links => links.map(link => link.getAttribute('href')))
  assert(changeHrefs.includes('/site/41'), 'site change did not link to the existing Site detail route')
  assert(changeHrefs.includes('/games/82'), 'game change did not link to the existing Game detail route')
  await page.locator('[data-change-link]').first().click()
  await page.waitForURL(url => /^\/(?:en\/)?(?:site|games)\//.test(url.pathname))
  console.log('[insights] mixed Recent Changes feed and entity links passed')

  await page.goto(toAbsoluteUrl(baseUrl, '/insights/games?metric=free&range=30d'), { waitUntil: 'domcontentloaded' })
  await page.locator('[data-metric-key="linux"]').click()
  await page.locator('[data-range="90d"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'linux' && url.searchParams.get('range') === '90d')
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-metric') === 'linux', 'game metric interaction did not update locally')

  await context.close()
  console.log('[insights] URL state, locale preservation, interactions, and mobile overflow passed')
} finally {
  await browser.close()
}

console.log('[insights] smoke passed')

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
