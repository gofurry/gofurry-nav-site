#!/usr/bin/env node
import { launchPerfBrowser, normalizeBaseUrl, parseArgs, toAbsoluteUrl } from './perf/shared.mjs'

const args = parseArgs()
const baseUrl = normalizeBaseUrl(args['base-url'] || process.env.INSIGHTS_BASE_URL || 'http://localhost:3000')
const entitySiteId = args['entity-site-id'] || process.env.INSIGHTS_SITE_ID || ''
const entitySiteDomain = args['entity-site-domain'] || process.env.INSIGHTS_SITE_DOMAIN || 'target.example'
const entityGameId = args['entity-game-id'] || process.env.INSIGHTS_GAME_ID || ''
const insightsRoutes = [
  '/insights',
  '/insights/sites?metric=ipv6&range=30d&dimension=country',
  '/insights/games?metric=free&range=30d&dimension=primary_tag',
  '/insights/games/players?metric=latest_observed',
  '/insights/games/prices?region=CN',
  '/insights/games/languages',
  '/insights/sites/compare',
  '/insights/games/compare?region=CN',
  '/insights/changes?domain=site&range=30d',
  '/en/insights',
  '/en/insights/sites?metric=ipv6&range=90d',
  '/en/insights/games?metric=free&range=90d',
  '/en/insights/games/players?metric=average_30d',
  '/en/insights/games/prices?region=US',
  '/en/insights/games/languages',
  '/en/insights/sites/compare',
  '/en/insights/games/compare?region=US',
  '/en/insights/changes?domain=game&range=7d',
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
const targetSiteId = entitySiteId || '1'
const siteTargetRoutes = [
  `/site/${targetSiteId}?domain=${encodeURIComponent(entitySiteDomain)}`,
  `/en/site/${targetSiteId}?domain=${encodeURIComponent(entitySiteDomain)}`,
]
const localizedSiteRoutes = entitySiteId ? [`/site/${entitySiteId}`, `/en/site/${entitySiteId}`] : []
const localizedGameRoutes = entityGameId ? [`/games/${entityGameId}`, `/en/games/${entityGameId}`] : []

for (const route of insightsRoutes) {
  const response = await fetch(toAbsoluteUrl(baseUrl, route), { redirect: 'manual' })
  const html = await response.text()
  assert(response.status === 200, `${route} expected HTTP 200, received ${response.status}`)
  assert(html.includes('insights-page'), `${route} did not SSR the Insights page shell`)
  assert(/<h1(?:\s|>)/.test(html), `${route} did not SSR a visible h1`)
  assert(!html.includes('insights-hero'), `${route} still SSR-rendered the retired large hero`)
  assert(html.includes('data-public-background'), `${route} did not SSR the app-level public background`)
  assert(html.includes('data-pattern-status="default"'), `${route} did not SSR the active public pattern`)
  assert(html.includes(route.startsWith('/en/') ? 'Ecosystem' : '生态观测'), `${route} did not expose the localized Ecosystem product name`)
  if (route.includes('/compare')) {
    assert(html.includes('name="robots" content="noindex, follow"'), `${route} did not SSR its noindex policy`)
  }
  console.log(`[insights] SSR ${route} -> 200`)
}

const patternResponse = await fetch(toAbsoluteUrl(baseUrl, '/web/background/gofurry-pattern.svg'))
assert(patternResponse.status === 200 && patternResponse.headers.get('content-type')?.includes('image/svg+xml'), 'default public pattern asset is not served as SVG')

const sitemap = await (await fetch(toAbsoluteUrl(baseUrl, '/sitemap.xml'))).text()
assert(!sitemap.includes('/insights/sites/compare') && !sitemap.includes('/insights/games/compare'), 'Compare routes leaked into sitemap')

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
  await page.route('**/api/v2/nav/sites/*/view', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ code: 1, data: { site_id: 41, view_count: 1 } }),
  }))
  await page.route('**/api/v2/game/games/*/view', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ code: 1, data: { game_id: 82, view_count: 1 } }),
  }))
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/sites?metric=invalid&range=bad&dimension=bad&slice=true'), {
    waitUntil: 'domcontentloaded',
    timeout: 30000,
  })
  await page.waitForSelector('.insights-domain-page')
  const patternState = await page.evaluate(() => {
    const pattern = document.querySelector('.gf-public-background__pattern')
    if (!pattern) return null
    const root = document.documentElement
    const wasDark = root.classList.contains('dark')
    root.classList.remove('dark')
    const light = getComputedStyle(pattern)
    const lightState = { color: light.backgroundColor, image: light.maskImage || light.webkitMaskImage, opacity: Number(light.opacity), repeat: light.maskRepeat || light.webkitMaskRepeat, size: light.maskSize || light.webkitMaskSize }
    root.classList.add('dark')
    const dark = getComputedStyle(pattern)
    const darkState = { color: dark.backgroundColor, image: dark.maskImage || dark.webkitMaskImage, opacity: Number(dark.opacity), repeat: dark.maskRepeat || dark.webkitMaskRepeat, size: dark.maskSize || dark.webkitMaskSize }
    root.classList.toggle('dark', wasDark)
    return { lightState, darkState }
  })
  assert(patternState?.lightState.image.includes('gofurry-pattern.svg') && patternState.darkState.image.includes('gofurry-pattern.svg'), 'public pattern is not active in both themes')
  assert(patternState?.lightState.repeat === 'repeat' && patternState.darkState.repeat === 'repeat', 'public pattern is not infinitely tiled')
  assert(patternState?.lightState.size === '160px 160px' && patternState.darkState.size === '160px 160px', 'public pattern does not use the default 160px tile size')
  assert(patternState?.lightState.color === 'rgb(168, 135, 115)' && patternState.darkState.color === 'rgb(215, 193, 175)', 'public pattern theme colors drifted')
  assert(patternState?.lightState.opacity === 0.065 && patternState.darkState.opacity === 0.045, 'public pattern theme opacity drifted')
  const desktopNavigation = await page.evaluate(() => {
    const shell = document.querySelector('.ecosystem-navigation')
    const primary = document.querySelector('.insights-nav')
    const context = document.querySelector('.site-intelligence-nav')
    if (!shell || !primary || !context) return null
    const primaryBox = primary.getBoundingClientRect()
    const contextBox = context.getBoundingClientRect()
    const primaryStyle = getComputedStyle(primary)
    const contextStyle = getComputedStyle(context)
    const primaryLink = primary.querySelector('a')
    const contextLink = context.querySelector('a')
    const primaryLinkStyle = primaryLink ? getComputedStyle(primaryLink) : null
    const contextLinkStyle = contextLink ? getComputedStyle(contextLink) : null
    return {
      direction: getComputedStyle(shell).flexDirection,
      primaryLeft: primaryBox.left,
      contextLeft: contextBox.left,
      centerDelta: Math.abs((primaryBox.top + primaryBox.height / 2) - (contextBox.top + contextBox.height / 2)),
      containerHeightDelta: Math.abs(primaryBox.height - contextBox.height),
      containerPaddingMatches: primaryStyle.padding === contextStyle.padding,
      containerBorderMatches: primaryStyle.borderWidth === contextStyle.borderWidth && primaryStyle.borderRadius === contextStyle.borderRadius,
      itemHeightDelta: primaryLink && contextLink ? Math.abs(primaryLink.getBoundingClientRect().height - contextLink.getBoundingClientRect().height) : Number.POSITIVE_INFINITY,
      itemTypographyMatches: primaryLinkStyle?.fontSize === contextLinkStyle?.fontSize && primaryLinkStyle?.fontWeight === contextLinkStyle?.fontWeight,
      itemPaddingMatches: primaryLinkStyle?.paddingInline === contextLinkStyle?.paddingInline,
    }
  })
  assert(desktopNavigation?.direction === 'row' && desktopNavigation.primaryLeft < desktopNavigation.contextLeft && desktopNavigation.centerDelta <= 2, 'desktop Ecosystem navigation was not left/right grouped')
  assert(desktopNavigation?.containerHeightDelta <= 1 && desktopNavigation.containerPaddingMatches && desktopNavigation.containerBorderMatches, 'primary and context navigation containers do not share one visual specification')
  assert(desktopNavigation?.itemHeightDelta <= 1 && desktopNavigation.itemTypographyMatches && desktopNavigation.itemPaddingMatches, 'primary and context navigation items do not share one visual specification')
  assert(await page.locator('.insights-hero').count() === 0, 'large Ecosystem hero remained visible')
  await page.waitForURL(url => url.searchParams.get('metric') === 'ipv6' && url.searchParams.get('range') === '30d' && url.searchParams.get('dimension') === 'country' && !url.searchParams.has('slice'))
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-metric') === 'ipv6', 'invalid metric was not normalized before rendering')
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-dimension') === 'country', 'invalid dimension was not normalized before rendering')

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
    navigationDirection: getComputedStyle(document.querySelector('.ecosystem-navigation')).flexDirection,
  }))
  assert(mobileState.overflow <= 2, `mobile Insights page overflowed horizontally by ${mobileState.overflow}px`)
  assert(mobileState.navigationDirection === 'column', 'small-screen Ecosystem navigation did not split into two centered rows')
  for (const forbidden of ['undefined', 'NaN', 'null%']) {
    assert(!mobileState.text.includes(forbidden), `mobile Insights page exposed ${forbidden}`)
  }
  await page.evaluate(() => window.scrollTo(0, 240))
  await page.waitForSelector('.mobile-bottom-tabs--visible')
  assert(await page.locator('.mobile-bottom-tabs__item').count() === 4, 'mobile bottom navigation did not expose four destinations')
  const activeMobileHref = await page.locator('.mobile-bottom-tabs__item--active').getAttribute('href')
  assert(activeMobileHref === '/en/insights', 'nested Ecosystem route did not activate the mobile Ecosystem destination')

  await page.setViewportSize({ width: 1440, height: 900 })
  await page.route('**/api/v2/nav/insights/metrics/**', async (route) => {
    const url = route.request().url()
    if (url.includes('/ipv6/')) {
      await route.abort('failed')
      return
    }
    const metric = url.includes('/security_txt/') ? 'security_txt' : 'tls13'
    if (url.includes('/breakdown/')) {
      const match = url.match(/\/breakdown\/([^/]+)\/([^/]+)\/trend/)
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ code: 1, data: {
          key: metric,
          dimension: match?.[1] || 'group',
          slice: { value: decodeURIComponent(match?.[2] || '12'), label: '分组 #12', label_en: 'Group #12' },
          slice_mode: 'overlapping', requested_range: '30d', as_of: '2026-09-01',
          available_from: '2026-08-31', available_through: '2026-09-01',
          points: [
            { date: '2026-08-31', population: 4, eligible: 4, known: 2, metric_value: 0.5, coverage: 0.5 },
            { date: '2026-09-01', population: 4, eligible: 0, known: 0, metric_value: null, coverage: null },
          ],
        } }),
      })
      return
    }
    if (url.includes('/breakdown')) {
      const dimension = new URL(url).searchParams.get('dimension') || 'country'
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ code: 1, data: {
          key: metric, dimension, slice_mode: dimension === 'group' ? 'overlapping' : 'partition', as_of: '2026-09-01',
          items: [{ value: dimension === 'group' ? '12' : 'CN', label: dimension === 'group' ? '分组 #12' : null, label_en: dimension === 'group' ? 'Group #12' : null, population: 4, eligible: 4, known: 2, metric_value: 0.5, coverage: 0.5 }],
        } }),
      })
      return
    }
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
  await page.waitForTimeout(500)
  await page.locator('[data-metric-key="tls13"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'tls13')
  await page.getByText('暂无可用历史数据', { exact: true }).waitFor()
  await page.locator('[data-metric-key="security_txt"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'security_txt')
  await page.getByText('正在积累历史数据', { exact: true }).waitFor()
  await page.locator('[data-metric-key="ipv6"]').click()
  await page.getByText('生态观测数据暂不可用', { exact: true }).first().waitFor()
  console.log('[insights] empty, one-point, zero-value, and request-error semantics passed')

  await page.locator('[data-metric-key="tls13"]').click()
  await page.locator('[data-dimension="group"]').click()
  await page.waitForURL(url => url.searchParams.get('dimension') === 'group' && !url.searchParams.has('slice'))
  await page.locator('[data-slice="12"]').click()
  await page.waitForURL(url => url.searchParams.get('slice') === '12')
  await page.waitForSelector('.insights-slice-trend')
  await page.locator('[data-metric-key="security_txt"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'security_txt' && url.searchParams.get('dimension') === 'group' && url.searchParams.get('slice') === '12')
  console.log('[insights] dimension URL state, slice selection, shared range, and isolated trend passed')

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
          state: { free: false, windows: true, mac: true, linux: null, release: 'available', as_of: '2026-08-30' },
          players: { current: 0, peak_30d: 3, average_30d: 1.5, as_of: '2026-08-30T12:00:00Z', fact_through: '2026-08-30', eligible_from_30d: '2026-08-01', observed_days_30d: 28, successful_samples_30d: 112, sample_coverage_30d: 0.93 },
          price: { region: 'CN', state: 'priced', currency: 'CNY', initial_amount: 5800, final_amount: 0, discount_percent: 100, as_of: '2026-08-30' },
          regional_prices: { as_of: '2026-08-30', regions: [
            { region: 'CN', available: true, state: 'priced', currency: 'CNY', initial_amount: 5800, final_amount: 0, discount_percent: 100, observed_low: { amount: 0, currency: 'CNY', first_seen: '2026-08-30', observed_since: '2026-08-28', initial_amount: 5800, discount_percent: 100 } },
            { region: 'US', available: true, state: 'unknown', currency: null, initial_amount: null, final_amount: null, discount_percent: null, observed_low: null },
            { region: 'HK', available: false, state: null, currency: null, initial_amount: null, final_amount: null, discount_percent: null, observed_low: null },
          ] },
          recent_changes: [
            { type: 'game.price.decreased', date: '2026-08-30', occurred_at: null, entity: { id, name: `Game fixture ${id}` }, detail: null },
            { type: 'game.windows.added', date: '2026-08-29', occurred_at: '2026-08-29T09:30:00Z', entity: { id, name: `Game fixture ${id}` }, detail: null },
          ],
        },
      }),
    })
  })
  const playerRequests = { '30d': 0, '90d': 0, all: 0 }
  const priceRequests = {}
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
    const requestUrl = new URL(route.request().url())
    const range = requestUrl.searchParams.get('range') || '30d'
    const region = requestUrl.searchParams.get('region') || 'CN'
    const key = `${region}:${range}`
    priceRequests[key] = (priceRequests[key] || 0) + 1
    if (range === '90d') return route.abort('failed')
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 1,
        data: {
          region, requested_range: range, available_from: '2026-08-28', available_through: '2026-08-30',
          points: [
            { date: '2026-08-28', state: 'free', currency: null, initial_amount: null, final_amount: null, discount_percent: null },
            { date: '2026-08-29', state: 'priced', currency: region === 'US' ? 'USD' : region === 'HK' ? 'HKD' : 'CNY', initial_amount: 5800, final_amount: 0, discount_percent: 100 },
            { date: '2026-08-30', state: 'unknown', currency: null, initial_amount: null, final_amount: null, discount_percent: null },
          ],
        },
      }),
    })
  })
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/sites?metric=ipv6&range=30d&dimension=country'), { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(500)
  await page.locator('.insights-nav__link[href="/insights"]').click()
  await page.waitForURL(url => url.pathname === '/insights')
  await page.waitForSelector('[data-change-link][href="/site/41"]')
  await page.waitForSelector('[data-change-link][href="/games/82"]')
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
  assert((await page.locator('[data-current-players]').textContent())?.trim() === '0 人', 'real player zero was not displayed as zero')
  assert(await page.locator('[data-price-region-summary]').count() === 3, 'Game overview did not expose all three regional prices')
  assert(await page.locator('[data-price-kind="priced"]').count() === 1, 'priced zero was confused with free')
  assert((await page.locator('[data-price-kind="priced"]').textContent())?.includes('¥0.00'), 'priced zero was not visibly priced')
  assert(await page.locator('[data-observed-low]').getByText('GoFurry 观测低价', { exact: false }).count() === 1, 'bounded observed-low product naming was not rendered')
  assert((await page.locator('.game-insights-platform-list').innerText()).includes('macOS') && (await page.locator('.game-insights-platform-list').innerText()).includes('支持'), 'Mac was not rendered as a peer platform state')
  assert(await page.locator('[data-entity-timeline][data-timeline-mode="compact"]').count() === 1, 'Game timeline did not default to compact mode')
  await page.getByRole('button', { name: '美国', exact: true }).click()
  await page.waitForSelector('[data-game-insights][data-price-region="US"][data-price-loaded-ranges*="US:30d"]')
  assert(await page.locator('[data-price-kind="unknown"]').count() === 1, 'explicit unknown regional price was collapsed')
  await page.getByRole('button', { name: '香港', exact: true }).click()
  await page.waitForSelector('[data-game-insights][data-price-region="HK"][data-price-loaded-ranges*="HK:30d"]')
  assert(await page.locator('[data-price-kind="missing"]').count() === 1, 'missing regional fact was collapsed into explicit unknown')
  await page.getByRole('button', { name: '中国', exact: true }).click()
  await page.waitForSelector('[data-game-insights][data-price-region="CN"]')
  assert(priceRequests['CN:30d'] === 1 && priceRequests['US:30d'] === 1 && priceRequests['HK:30d'] === 1, 'region+range price cache repeated or merged requests')
  const gameTimelineTimes = await page.locator('[data-entity-timeline] time').allTextContents()
  assert(gameTimelineTimes[0]?.trim() === '2026-08-30', 'day-precision Game event fabricated midnight')
  assert(gameTimelineTimes[1]?.includes(':'), 'exact Game event lost its time precision')
  await page.locator('[data-game-insights-range="90d"]').click()
  await page.waitForSelector('[data-game-insights][data-player-loaded-ranges*="90d"]')
  await page.locator('[data-price-history]').getByText('生态观测数据暂不可用', { exact: true }).waitFor()
  await page.locator('[data-game-insights-range="30d"]').click()
  await page.waitForTimeout(250)
  assert(playerRequests['30d'] === 1 && priceRequests['CN:30d'] === 1, 'returning to cached 30d repeated history requests')
  await page.locator('[data-game-insights-range="all"]').click()
  await page.waitForSelector('[data-game-insights][data-price-loaded-ranges*="all"]')
  await page.locator('[data-player-history]').getByText('生态观测数据暂不可用', { exact: true }).waitFor()
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
  await page.waitForTimeout(500)
  await page.locator('[data-metric-key="linux"]').click()
  await page.locator('[data-range="90d"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'linux' && url.searchParams.get('range') === '90d')
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-metric') === 'linux', 'game metric interaction did not update locally')
  await page.locator('[data-metric-key="mac"]').click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'mac')
  assert(await page.locator('.insights-domain-page').getAttribute('data-selected-metric') === 'mac', 'Mac did not enter the existing Metric/dimension product flow')

  await page.route('**/api/v2/game/insights/players/ranking**', (route) => {
    const metric = new URL(route.request().url()).searchParams.get('metric') || 'latest_observed'
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: {
      metric, basis: metric === 'latest_observed' ? 'scheduled_snapshot' : 'finalized_daily_facts',
      snapshot_scheduled_for: '2026-09-01T00:00:00Z', latest_slot_scheduled_for: '2026-09-02T00:00:00Z',
      observed_from: '2026-09-01T00:01:00Z', observed_through: '2026-09-01T00:02:00Z',
      window_from: '2026-08-03', window_through: '2026-09-01', population: 2, ranked: 1, entity_coverage: 0.5,
      items: [{ rank: 1, game: { id: 82, name: 'Zero fixture' }, value: 0, observed_at: metric === 'latest_observed' ? '2026-09-01T00:01:00Z' : null, eligible_from: '2026-08-03', observed_days: 28, successful_samples: 112, sample_coverage: null }],
    } }) })
  })
  await page.route('**/api/v2/game/insights/prices/overview**', (route) => {
    const region = new URL(route.request().url()).searchParams.get('region') || 'CN'
    if (region === 'HK') return route.abort('failed')
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: { region, as_of: '2026-09-01', population: 4, priced: 1, free: 1, unpriced: 1, unknown: 0, unavailable: 1, known: 3, coverage: 0.75, discounted: 1, discounted_share: 1 } }) })
  })
  await page.route('**/api/v2/game/insights/prices/discounts**', (route) => {
    const region = new URL(route.request().url()).searchParams.get('region') || 'CN'
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: { region, as_of: '2026-09-01', items: [{ game: { id: 82, name: 'Priced zero fixture' }, currency: region === 'US' ? 'USD' : 'CNY', initial_amount: 1000, final_amount: 0, discount_percent: 100, observed_low: { amount: 0, currency: region === 'US' ? 'USD' : 'CNY', first_seen: '2026-09-01', observed_since: '2026-08-30', initial_amount: 1000, discount_percent: 100 } }] } }) })
  })
  await page.route('**/api/v2/game/insights/languages/overview**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: {
    as_of: '2026-09-01', freshness_seconds: 259200, population: 3, fresh: 2, stale: 1, unobserved: 0, coverage: 2 / 3,
    fully_normalized_games: 1, unmapped_games: 1, unmapped_entries: 1, normalization_coverage: 0.5,
    items: [{ code: 'en', steam_name: 'English', supported_games: 2, share: 1, explicit_full_audio_games: 1, explicit_full_audio_share: 0.5 }],
  } }) }))

  await page.locator('.game-intelligence-nav a[href="/insights/games/players"]').click()
  await page.waitForSelector('[data-player-intelligence]')
  await page.getByRole('button', { name: '30 天观测均值', exact: true }).click()
  await page.waitForURL(url => url.searchParams.get('metric') === 'average_30d')
  await page.getByText('112 个成功样本', { exact: false }).waitFor()
  assert((await page.locator('.intelligence-table tbody tr td').nth(2).textContent())?.trim() === '0', 'Player ranking lost a real zero')
  await page.locator('.game-intelligence-nav a[href="/insights/games/prices"]').click()
  await page.waitForSelector('[data-regional-price-intelligence]')
  await page.getByRole('button', { name: '香港', exact: true }).click()
  await page.waitForURL(url => url.searchParams.get('region') === 'HK')
  await page.getByText('Priced zero fixture', { exact: true }).waitFor()
  assert(await page.locator('.intelligence-panel .intelligence-table').count() === 1, 'Price overview failure broke the independent discount list')
  await page.locator('.game-intelligence-nav a[href="/insights/games/languages"]').click()
  await page.waitForSelector('[data-language-intelligence]')
  await page.getByText('明确标注完整音频', { exact: true }).waitFor()
  await page.getByText('语言是重叠分布，各语言比例不能相加推导 100%。', { exact: true }).waitFor()
  console.log('[insights] P2.2 Player, regional Price, Mac, Language URL/zero/quality/failure semantics passed')

  const siteCapabilityKeys = ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified']
  await page.route('**/api/v2/nav/insights/compare**', (route) => {
    const ids = new URL(route.request().url()).searchParams.get('ids').split(',').map(Number)
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: {
      status: 'ready', as_of: '2026-09-01', sites: ids.map((id) => ({
        site: { id, name: `Site ${id}` },
        capabilities: siteCapabilityKeys.map((key, index) => ({ key, state: id === 42 && index === 0 ? 'unknown' : 'supported' })),
        certificate: id === 41 ? { target: 'site-41.example', not_after: '2026-09-08T00:00:00Z', days_to_expiry: 6, expiry_status: 'expires_within_7d', verified: true, verification_issue: null, issuer: 'Fixture CA', observed_at: '2026-09-01T23:00:00Z' } : null,
      })),
    } }) })
  })
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/sites/compare'), { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(500)
  await page.locator('.compare-builder input').fill('42,41,42')
  await page.locator('.compare-builder button[type="submit"]').click()
  await page.waitForURL(url => url.searchParams.get('ids') === '42,41')
  await page.waitForSelector('[data-site-compare][data-compare-count="2"] [data-compare-result]')
  assert((await page.locator('[data-compare-entity-id]').allTextContents()).map(value => value.trim()).join('|').includes('Site 42') && (await page.locator('[data-compare-entity-id]').first().getAttribute('data-compare-entity-id')) === '42', 'Site Compare lost first-appearance order or deduplication')
  assert(await page.locator('[data-capability-state="unknown"]').count() === 1, 'Site Compare collapsed unknown capability state')

  await page.route('**/api/v2/game/insights/compare**', (route) => {
    const url = new URL(route.request().url())
    const ids = url.searchParams.get('ids').split(',').map(Number)
    const region = url.searchParams.get('region') || 'CN'
    return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ code: 1, data: {
      status: 'ready', region, state_as_of: '2026-09-01', player_snapshot_scheduled_for: '2026-09-01T12:00:00Z', player_fact_through: '2026-09-01',
      games: ids.map((id) => ({ game: { id, name: `Game ${id}` }, state: { free: id === 83, windows: true, mac: null, linux: false },
        players: { current_available: id === 82, current: id === 82 ? 0 : null, observed_at: id === 82 ? '2026-09-01T12:01:00Z' : null, peak_30d: id === 82 ? 0 : null, average_30d: id === 82 ? 0 : null, eligible_from_30d: '2026-08-03', observed_days_30d: id === 82 ? 1 : 0, successful_samples_30d: id === 82 ? 1 : 0, sample_coverage_30d: null },
        price: id === 82 ? { region, available: true, state: 'priced', currency: region === 'US' ? 'USD' : 'CNY', initial_amount: 1000, final_amount: 0, discount_percent: 100, observed_low: { amount: 0, currency: region === 'US' ? 'USD' : 'CNY', first_seen: '2026-09-01', observed_since: '2026-09-01', initial_amount: 1000, discount_percent: 100 } } : { region, available: true, state: 'free', currency: null, initial_amount: null, final_amount: null, discount_percent: null, observed_low: null },
        languages: { evidence: id === 82 ? 'fresh' : 'stale', supported: ['en'], explicit_full_audio: id === 82 ? ['en'] : [], unknown_names: id === 83 ? ['Klingon'] : [] },
      })),
    } }) })
  })
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/games/compare?region=CN'), { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(500)
  await page.locator('.compare-builder input').fill('82,83')
  await page.locator('.compare-builder button[type="submit"]').click()
  await page.waitForURL(url => url.searchParams.get('ids') === '82,83' && url.searchParams.get('region') === 'CN')
  await page.waitForSelector('[data-game-compare][data-compare-count="2"] [data-compare-result]')
  assert((await page.locator('[data-current-player-available="true"]').textContent())?.trim() === '0', 'Game Compare changed real player zero')
  assert((await page.locator('[data-current-player-available="false"]').textContent())?.trim() === '—', 'Game Compare changed unavailable players into zero')
  assert(await page.locator('[data-price-state="priced"]').getByText('¥0.00', { exact: true }).count() === 1, 'Game Compare confused priced zero with Free')
  assert(await page.locator('[data-language-evidence="stale"]').count() === 1, 'Game Compare changed stale language evidence')
  await page.getByRole('button', { name: '美国', exact: true }).click()
  await page.waitForURL(url => url.searchParams.get('ids') === '82,83' && url.searchParams.get('region') === 'US')
  await page.waitForSelector('[data-game-compare][data-compare-region="US"]')
  console.log('[insights] Site/Game Compare URL, order, zero, stale, and regional semantics passed')

  const changePage = (domain, cursor) => ({
    items: [{
      domain, category: domain === 'site' ? 'capability' : 'price',
      type: domain === 'site' ? 'site.ipv6.enabled' : 'game.price.decreased',
      date: cursor ? '2026-08-31' : '2026-09-01', occurred_at: null,
      entity: { id: cursor ? 2 : 1, name: `${domain} fixture` }, detail: null,
    }],
    next_cursor: cursor ? null : 'opaque-next',
  })
  await page.route('**/api/v2/nav/insights/changes**', route => route.fulfill({
    status: 200, contentType: 'application/json',
    body: JSON.stringify({ code: 1, data: changePage('site', new URL(route.request().url()).searchParams.get('cursor')) }),
  }))
  await page.route('**/api/v2/game/insights/changes**', route => route.fulfill({
    status: 200, contentType: 'application/json',
    body: JSON.stringify({ code: 1, data: changePage('game', new URL(route.request().url()).searchParams.get('cursor')) }),
  }))
  await page.goto(toAbsoluteUrl(baseUrl, '/insights/changes?domain=bad&range=bad&category=bad&cursor=leak'), { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(500)
  await page.waitForURL(url => url.searchParams.get('domain') === 'site' && url.searchParams.get('range') === '30d' && !url.searchParams.has('category') && !url.searchParams.has('cursor'))
  await page.locator('[data-domain-filter="game"]').click()
  await page.waitForURL(url => url.searchParams.get('domain') === 'game')
  await page.locator('[data-category-filter="price"]').click()
  await page.waitForURL(url => url.searchParams.get('category') === 'price')
  await page.waitForSelector('[data-load-more]')
  assert(!new URL(page.url()).searchParams.has('cursor'), 'Explorer cursor leaked into URL')
  await page.locator('[data-load-more]').click()
  await page.waitForFunction(() => document.querySelectorAll('.insights-change-explorer-item').length === 2)
  assert(!new URL(page.url()).searchParams.has('cursor'), 'Load More placed cursor in URL')
  console.log('[insights] Change Explorer filters, reset, Load More, and cursor URL isolation passed')

  await context.close()
  console.log('[insights] URL state, locale preservation, interactions, and mobile overflow passed')
} finally {
  await browser.close()
}

console.log('[insights] smoke passed')

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
