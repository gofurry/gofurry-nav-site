#!/usr/bin/env node
import { launchPerfBrowser, normalizeBaseUrl, parseArgs, toAbsoluteUrl } from './perf/shared.mjs'

const args = parseArgs()
const baseUrl = normalizeBaseUrl(args['base-url'] || process.env.INSIGHTS_BASE_URL || 'http://localhost:3000')
const entitySiteId = args['entity-site-id'] || process.env.INSIGHTS_SITE_ID || ''
const entityGameId = args['entity-game-id'] || process.env.INSIGHTS_GAME_ID || ''
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
const siteTargetRoutes = ['/site/1/target.example', '/en/site/1/target.example']
const localizedSiteRoutes = entitySiteId ? [`/site/${entitySiteId}`, `/en/site/${entitySiteId}`] : []
const localizedGameRoutes = entityGameId ? [`/games/${entityGameId}`, `/en/games/${entityGameId}`] : []

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

for (const route of siteTargetRoutes) {
  const response = await fetch(toAbsoluteUrl(baseUrl, route))
  const html = await response.text()
  assert(response.status === 200, `${route} expected the existing Site target route`)
  assert(!html.includes('data-site-insights'), `${route} conflated target observations with Site-level Insights`)
  console.log(`[insights] target route ${route} excludes Site-level Insights`)
}

for (const route of localizedSiteRoutes) {
  const response = await fetch(toAbsoluteUrl(baseUrl, route))
  const html = await response.text()
  assert(response.status === 200, `${route} expected HTTP 200`)
  assert(html.includes('data-site-insights'), `${route} did not SSR Site-level Insights`)
  console.log(`[insights] Site entity SSR ${route} -> 200`)
}

for (const route of localizedGameRoutes) {
  const response = await fetch(toAbsoluteUrl(baseUrl, route))
  const html = await response.text()
  assert(response.status === 200, `${route} expected HTTP 200`)
  assert(html.includes(`game-insights:${entityGameId}`) && html.includes('peak_30d'), `${route} did not SSR the Game Insights summary payload`)
  assert(html.includes('data-game-tab="insights"'), `${route} did not SSR the Insights tab after Introduction`)
  console.log(`[insights] Game summary SSR ${route} -> 200`)
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
    changes_7d: 2,
    metrics: [],
    recent_changes: domain === 'site'
      ? [
          { type: 'site.ipv6.enabled', date: '2026-09-01', occurred_at: null, entity: { id: 41, name: 'Site fixture' }, detail: null },
          { type: 'site.tls13.disabled', date: '2026-08-31', occurred_at: null, entity: { id: 42, name: 'Site failure fixture' }, detail: null },
        ]
      : [
          { type: 'game.windows.added', date: '2026-09-01', occurred_at: '2026-09-01T12:00:00Z', entity: { id: 82, name: 'Game fixture' }, detail: null },
          { type: 'game.linux.added', date: '2026-08-31', occurred_at: '2026-08-31T12:00:00Z', entity: { id: 83, name: 'Game failure fixture' }, detail: null },
        ],
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
  await page.route('**/api/v2/nav/sites/*/detail**', (route) => {
    const id = Number(new URL(route.request().url()).pathname.match(/sites\/(\d+)\/detail/)?.[1] || 0)
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          site: {
            id,
            name: `Site fixture ${id}`,
            info: 'Site detail remains available independently from Insights.',
            icon: '',
            country: 'CN',
            nsfw: '0',
            welfare: '0',
            view_count: 1,
          },
          selected_target: `site-${id}.example`,
          latest_core: null,
          site_summary: { targets: [{ target: `site-${id}.example` }, { target: `alt-${id}.example` }] },
          target_summary: null,
          light_probe_state: null,
        },
      }),
    })
  })
  await page.route('**/api/v2/nav/sites/*/insights', (route) => {
    const id = Number(new URL(route.request().url()).pathname.match(/sites\/(\d+)\/insights/)?.[1] || 0)
    if (id === 42) return route.abort('failed')
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          site: { id, name: `Site fixture ${id}` },
          capabilities: [
            { key: 'ipv6', as_of: '2026-08-30', state: 'unknown', ecosystem: { value: 0.5, coverage: 0.8 } },
            { key: 'tls13', as_of: '2026-08-30', state: 'unavailable', ecosystem: { value: 0.6, coverage: 0.8 } },
            { key: 'security_txt', as_of: '2026-08-30', state: 'unsupported', ecosystem: { value: 0.2, coverage: 0.8 } },
          ],
          recent_changes: [{
            type: 'site.ipv6.enabled', date: '2026-08-30', occurred_at: null,
            entity: { id, name: `Site fixture ${id}` }, detail: null,
          }],
        },
      }),
    })
  })
  await page.route('**/api/v2/game/info**', (route) => {
    const id = Number(new URL(route.request().url()).searchParams.get('id') || 0)
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          id: String(id), appid: id, name: `Game fixture ${id}`, summary: 'Game detail remains usable.',
          site: { view_count: 1, resources: [], groups: [], links: [] },
          platforms: { windows: true, linux: true }, prices: [], news: [], tags: [], developers: [], publishers: [],
          media: { screenshots: [], movies: [], assets: [] }, requirements: {}, support_info: {}, extra: {},
        },
      }),
    })
  })
  await page.route('**/api/v2/game/reviews**', route => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: { total: 0, data: [] } }),
  }))
  await page.route('**/api/v2/game/recommend/similar**', route => route.fulfill({
    status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: [] }),
  }))
  await page.route('**/api/v2/game/games/*/insights', (route) => {
    const id = Number(new URL(route.request().url()).pathname.match(/games\/(\d+)\/insights/)?.[1] || 0)
    if (id === 83) return route.abort('failed')
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          game: { id, name: `Game fixture ${id}` },
          state: { free: false, windows: true, linux: null, release: 'available', as_of: '2026-08-30' },
          players: { current: 0, peak_30d: 3, as_of: '2026-08-30T12:00:00Z' },
          price: { region: 'CN', state: 'priced', currency: 'CNY', initial_amount: 5800, final_amount: 0, discount_percent: 100, as_of: '2026-08-30' },
          recent_changes: [
            { type: 'game.price.decreased', date: '2026-08-30', occurred_at: null, entity: { id, name: `Game fixture ${id}` }, detail: null },
            { type: 'game.windows.added', date: '2026-08-29', occurred_at: '2026-08-29T09:30:00Z', entity: { id, name: `Game fixture ${id}` }, detail: null },
          ],
        },
      }),
    })
  })
  const playerRequests = { '30d': 0, '90d': 0, all: 0 }
  const priceRequests = { '30d': 0, '90d': 0, all: 0 }
  await page.route('**/api/v2/game/games/*/insights/players**', (route) => {
    const range = new URL(route.request().url()).searchParams.get('range') || '30d'
    playerRequests[range] += 1
    if (range === 'all') return route.abort('failed')
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          requested_range: range, available_from: '2026-08-29', available_through: '2026-08-30',
          points: [{ date: '2026-08-29', min: 0, max: 0, avg: 0 }, { date: '2026-08-30', min: 1, max: 3, avg: 2 }],
        },
      }),
    })
  })
  await page.route('**/api/v2/game/games/*/insights/prices**', (route) => {
    const range = new URL(route.request().url()).searchParams.get('range') || '30d'
    priceRequests[range] += 1
    if (range === '90d') return route.abort('failed')
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          requested_range: range, available_from: '2026-08-28', available_through: '2026-08-30',
          points: [
            { date: '2026-08-28', state: 'free', currency: null, initial_amount: null, final_amount: null, discount_percent: null },
            { date: '2026-08-29', state: 'priced', currency: 'CNY', initial_amount: 5800, final_amount: 0, discount_percent: 100 },
            { date: '2026-08-30', state: 'unknown', currency: null, initial_amount: null, final_amount: null, discount_percent: null },
          ],
        },
      }),
    })
  })
  await page.goto(toAbsoluteUrl(baseUrl, '/'), { waitUntil: 'domcontentloaded' })
  await page.getByRole('link', { name: '洞察', exact: true }).first().click()
  await page.waitForURL(url => url.pathname === '/insights')
  const changeHrefs = await page.locator('[data-change-link]').evaluateAll(links => links.map(link => link.getAttribute('href')))
  assert(changeHrefs.includes('/site/41'), 'site change did not link to the existing Site detail route')
  assert(changeHrefs.includes('/games/82'), 'game change did not link to the existing Game detail route')
  await page.locator('[data-change-link][href="/site/41"]').click()
  await page.waitForSelector('[data-site-insights]')
  assert(await page.locator('[data-capability-key="ipv6"]').getAttribute('data-capability-state') === 'unknown', 'unknown Site capability changed meaning')
  assert(await page.locator('[data-capability-key="tls13"]').getAttribute('data-capability-state') === 'unavailable', 'unavailable Site capability changed meaning')
  assert(await page.locator('[data-capability-key="security_txt"]').getAttribute('data-capability-state') === 'unsupported', 'unsupported Site capability was not explicit')
  assert(await page.getByText('同日统计 · 2026-08-30', { exact: true }).count() === 3, 'Site capability and ecosystem context did not expose the same fact date')
  assert((await page.locator('[data-entity-timeline] time').first().textContent())?.trim() === '2026-08-30', 'day-precision Site event fabricated a time')
  await page.goBack()
  await page.locator('[data-change-link][href="/site/42"]').click()
  await page.waitForSelector('[data-site-insights-unavailable]')
  assert(await page.locator('.site-detail-page').count() === 1, 'Site Insights failure broke the base Site detail')
  console.log('[insights] Site entity states, same-day context, timeline, and failure isolation passed')

  await page.goBack()
  await page.locator('[data-change-link][href="/games/82"]').click()
  await page.waitForSelector('[data-game-tab="insights"]')
  await page.locator('[data-game-tab="insights"]').click()
  await page.waitForSelector('[data-game-insights][data-player-loaded-ranges*="30d"][data-price-loaded-ranges*="30d"]')
  assert((await page.locator('[data-current-players]').textContent())?.trim() === '0', 'real player zero was not displayed as zero')
  assert(await page.locator('[data-price-kind="priced"]').count() === 1, 'priced zero was confused with free')
  assert((await page.locator('[data-price-kind="priced"]').textContent())?.includes('¥0.00'), 'priced zero was not visibly priced')
  const gameTimelineTimes = await page.locator('[data-entity-timeline] time').allTextContents()
  assert(gameTimelineTimes[0]?.trim() === '2026-08-30', 'day-precision Game event fabricated midnight')
  assert(gameTimelineTimes[1]?.includes(':'), 'exact Game event lost its time precision')
  await page.locator('[data-game-insights-range="90d"]').click()
  await page.waitForSelector('[data-game-insights][data-player-loaded-ranges*="90d"]')
  await page.locator('[data-price-history]').getByText('洞察数据暂不可用', { exact: true }).waitFor()
  await page.locator('[data-game-insights-range="30d"]').click()
  await page.waitForTimeout(250)
  assert(playerRequests['30d'] === 1 && priceRequests['30d'] === 1, 'returning to cached 30d repeated history requests')
  await page.locator('[data-game-insights-range="all"]').click()
  await page.waitForSelector('[data-game-insights][data-price-loaded-ranges*="all"]')
  await page.locator('[data-player-history]').getByText('洞察数据暂不可用', { exact: true }).waitFor()
  assert(await page.locator('[data-price-history] .game-insights-chart--visible').count() === 1, 'player failure prevented price history from rendering')
  console.log('[insights] Game zero, priced-zero, lazy range cache, and partial failures passed')

  await page.goBack()
  await page.locator('[data-change-link][href="/games/83"]').click()
  await page.waitForSelector('[data-game-tab="insights"]')
  await page.locator('[data-game-tab="insights"]').click()
  await page.waitForSelector('[data-game-summary-unavailable]')
  assert(await page.locator('.game-detail-tabs').count() === 1, 'Game summary failure broke the base Game detail')
  console.log('[insights] mixed feed links and Game summary failure isolation passed')

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
