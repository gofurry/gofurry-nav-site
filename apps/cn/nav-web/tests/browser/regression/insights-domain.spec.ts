import type { Page } from '@playwright/test'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import { assertRuntimeSurface, revealImages } from '../fixtures/insights-runtime'
import { test, expect, domainMetricKeys, domainDimensionKeys, openRuntime, keyboardFocus } from '../fixtures/insights-domain'

const routes = [
  ['/insights/sites?metric=ipv6&range=30d&dimension=country', 'site'],
  ['/en/insights/sites?metric=tls13&range=90d&dimension=group', 'site'],
  ['/insights/games?metric=free&range=30d&dimension=primary_tag', 'game'],
  ['/en/insights/games?metric=linux&range=90d&dimension=tag', 'game'],
] as const
async function layout(page: Page, width: number, domain: string) {
  const data = await page.evaluate(() => ({
    overflow: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth,
    chart: document.querySelector('.insights-trend')!.getBoundingClientRect().toJSON(),
    context: document.querySelector('.insight-trend-context')!.getBoundingClientRect().toJSON(),
    tracks: getComputedStyle(document.querySelector('.insight-trend-workspace')!).gridTemplateColumns.split(' ').map(Number.parseFloat),
  }))
  expect(data.overflow).toBeLessThanOrEqual(1)
  if (domain === 'site' ? width > 760 : width > 1100) {
    expect(data.tracks).toHaveLength(2)
    // Grid tracks measure the available columns, including the context's own
    // internal padding. Do not compare chart width to only the sidebar text box.
    const ratio = domain === 'site' ? 2 : 3
    expect(data.tracks[0]! / data.tracks[1]!).toBeGreaterThan(ratio - .15)
    expect(data.tracks[0]! / data.tracks[1]!).toBeLessThan(ratio + .15)
    expect(Math.abs(data.chart.width - data.tracks[0]!)).toBeLessThan(1)
    expect(Math.abs(data.context.width - data.tracks[1]!)).toBeLessThan(1)
    expect(data.context.left).toBeGreaterThanOrEqual(data.chart.right)
  } else {
    expect(data.tracks).toHaveLength(1)
    expect(data.context.top).toBeGreaterThanOrEqual(data.chart.bottom)
  }
}

async function foundationAppearance(page: Page, width: number, domain: string) {
  const result = await page.evaluate(() => {
    const css = (selector: string) => getComputedStyle(document.querySelector(selector)!)
    const rail = document.querySelector('[data-metric-rail]')!, railBox = rail.getBoundingClientRect()
    const items = [...rail.querySelectorAll('button')]
    return {
      title: { weight: css('h1').fontWeight, size: parseFloat(css('h1').fontSize) },
      intro: { weight: css('.insights-domain-header__intro').fontWeight, size: parseFloat(css('.insights-domain-header__intro').fontSize) },
      count: { weight: css('[data-domain-count]').fontWeight, size: parseFloat(css('[data-domain-count]').fontSize) },
      snapshot: css('.insights-domain-header__snapshot').fontSize,
      headings: [...document.querySelectorAll('.insights-trend h2, [data-dimension-explorer] h2')].map(el => getComputedStyle(el).fontWeight),
      values: [...rail.querySelectorAll('strong')].map(el => getComputedStyle(el).fontWeight),
      labels: [...document.querySelectorAll('.insight-dimension-bars__label strong')].map(el => getComputedStyle(el).fontWeight),
      scrollable: rail.scrollWidth > rail.clientWidth,
      scrollbar: css('[data-metric-rail]').scrollbarWidth,
      partialItem: items.some(el => { const box = el.getBoundingClientRect(); return box.left < railBox.right && box.right > railBox.right + 1 }),
      controls: [...items, ...document.querySelectorAll('[data-dimension]')].map(el => ({ height: el.getBoundingClientRect().height, background: getComputedStyle(el).backgroundColor, shadow: getComputedStyle(el).boxShadow })),
      selectedLine: getComputedStyle(rail.querySelector('[aria-pressed="true"]')!, '::after').height,
      textFits: [...document.querySelectorAll('[data-domain-header] p, [data-metric-key] > *, .insight-trend-context dd, .insight-dimension-bars__label')].every(el => el.scrollWidth <= el.clientWidth + 1),
      surfaces: ['[data-domain-header]', '[data-metric-rail]', '.insight-trend-workspace', '.insights-trend', '[data-dimension-explorer]'].map(selector => {
        const style = css(selector)
        return { background: style.backgroundColor, shadow: style.boxShadow, left: style.borderLeftWidth, right: style.borderRightWidth }
      }),
    }
  })
  expect(result.title.weight).toBe('700'); expect(result.title.size).toBeGreaterThanOrEqual(30); expect(result.title.size).toBeLessThanOrEqual(46)
  expect(result.intro.weight).toBe('400'); expect(result.intro.size).toBeGreaterThanOrEqual(14); expect(result.intro.size).toBeLessThanOrEqual(15)
  expect(result.count.weight).toBe('600'); expect(result.count.size).toBeGreaterThanOrEqual(25); expect(result.count.size).toBeLessThanOrEqual(28)
  expect(result.snapshot).toBe('12px')
  for (const weight of [...result.headings, ...result.values, ...result.labels]) expect(weight).toBe('600')
  expect(result.scrollbar).toBe('thin'); expect(result.selectedLine).toBe('2px'); expect(result.textFits).toBe(true)
  await expect(page.locator('[data-dimension][aria-pressed="true"]')).toHaveCSS('font-weight', '600')
  await expect(page.locator('[data-dimension][aria-pressed="true"]')).toHaveCSS('border-bottom-width', '2px')
  if (width === 390) { expect(result.scrollable).toBe(true); expect(result.partialItem).toBe(true) }
  if (width === 1440) expect(result.scrollable).toBe(false)
  for (const control of result.controls) {
    expect(control.height).toBeGreaterThanOrEqual(44)
    expect(control.background).toBe('rgba(0, 0, 0, 0)'); expect(control.shadow).toBe('none')
  }
  for (const surface of result.surfaces) expect(surface).toEqual({ background: 'rgba(0, 0, 0, 0)', shadow: 'none', left: '0px', right: '0px' })
  await expect(page.locator('[data-domain-header] time')).toHaveAttribute('datetime', '2026-09-09T04:00:00Z')
  await expect(page.locator('.insight-trend-context')).toContainText('2026-09-08')
  await expect(page.locator('.insight-trend-context')).not.toContainText('2026-09-09')
  await expect(page.locator('.insight-trend-context__delta')).toContainText('+4.2%')
  if (domain === 'site') {
    await expect(page.locator('.insight-trend-context')).toContainText('2026-06-01')
    await expect(page.locator('[data-metric-key="certificate_verified"] strong')).toHaveText('0.0%')
    await expect(page.locator('[data-metric-key="security_txt"] strong')).toHaveText('—')
  }
}

for (const [route, domain] of routes) for (const theme of ['light', 'dark'] as const) for (const width of [1440, 1024, 768, 390]) {
  test('Domain SSR/hydration, full raw data, focus ' + route + ' ' + theme + ' ' + width, async ({ page, context, request, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    const ssr = await request.get(route), ssrHTML = await ssr.text()
    expect(ssr.status()).toBe(200)
    expect(ssrHTML).toMatch(/<h1>[^<]+<\/h1>/)
    for (const key of domainMetricKeys[domain]) expect(ssrHTML).toContain('data-metric-key="' + key + '"')
    expect(runtime.calls).toHaveLength(domain === 'site' ? 3 : 4)
    const offset = runtime.calls.length
    const html = await openRuntime(page, route)
    expect(html).toContain('data-domain-count>' + (domain === 'site' ? 238 : 213))
    expect(html).toContain('insights-domain-nav')
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('.insight-domain-activity .insight-activity-item')).toHaveCount(2)
    await expect(page.locator('.insight-domain-activity .insight-entity-media')).toHaveCount(2)
    await expect(page.locator('.overview-event, [data-activity-icon]')).toHaveCount(0)
    const prefix = route.startsWith('/en/') ? '/en' : ''
    expect(await page.locator('.insight-domain-continue a').evaluateAll(elements => elements.map(el => el.getAttribute('href'))))
      .toEqual((domain === 'game' ? ['/insights/games/players', '/insights/games/prices', '/insights/games/languages', '/insights/games/compare'] : ['/insights/sites/certificates', '/insights/sites/compare']).map(path => prefix + path))
    await expect(page.locator('[data-metric-key]')).toHaveCount(domainMetricKeys[domain].length)
    await expect(page.locator('[data-dimension]')).toHaveCount(domainDimensionKeys[domain].length)
    await expect(page.locator('[data-slice]')).toHaveCount(8)
    await expect(page.locator('[data-dimension-raw]')).not.toHaveAttribute('open')
    expect(await page.locator('[data-slice] progress').first().evaluate(el => ({ value: el.value, max: el.max })))
      .toEqual(domain === 'site' ? { value: .824, max: 1 } : { value: 46, max: 46 })
    await page.locator('[data-dimension-raw] summary').click()
    await expect(page.locator('[data-dimension-raw] tbody tr')).toHaveCount(10)
    if (width === 390) expect(await page.locator('.insight-dimension-raw__scroll').evaluate(el => el.scrollWidth > el.clientWidth)).toBe(true)
    await layout(page, width, domain)
    await foundationAppearance(page, width, domain)
    await assertRuntimeSurface(page, '.insights-domain-page', theme)
    const last = page.locator('[data-metric-key]').last()
    await keyboardFocus(last)
    expect(await last.evaluate(el => {
      const box = el.getBoundingClientRect(), rail = el.closest('[data-metric-rail]')!.getBoundingClientRect()
      const css = getComputedStyle(el)
      return box.left >= rail.left - 1 && box.right <= rail.right + 1
        && el.matches(':focus-visible') && css.outlineStyle === 'solid' && parseFloat(css.outlineWidth) >= 2
    })).toBe(true)
    await expect(page.locator('[data-metric-key][aria-pressed="true"]')).toHaveAttribute('data-metric-key', new URL(route, 'http://fixture.test').searchParams.get('metric')!)
    // A few review images from this existing fixture, never accepted goldens.
    const reviewDir = process.env.GOFURRY_DOMAIN_REVIEW_DIR
    const representative = (domain === 'game' && !prefix && theme === 'light' && [1440, 390].includes(width))
      || (domain === 'site' && !prefix && theme === 'light' && width === 1440)
      || (domain === 'site' && prefix && theme === 'dark' && width === 768)
    const alternate = (domain === 'game' && prefix && theme === 'dark' && width === 1440)
      || (domain === 'site' && !prefix && theme === 'dark' && width === 390)
    if (reviewDir && (representative || alternate)) {
      await mkdir(reviewDir, { recursive: true })
      await page.locator('[data-dimension-raw] summary').click()
      await page.locator('[data-domain-header] h1').click()
      await page.locator('[data-metric-rail]').evaluate(el => { el.scrollLeft = 0 })
      await revealImages(page)
      const name = `${domain}-${prefix ? 'en' : 'zh'}-${theme}-${width}`
      if (representative) await page.screenshot({ path: join(reviewDir, `full-${name}.png`), fullPage: true })
      await page.locator('[data-metric-rail]').screenshot({ path: join(reviewDir, `rail-${name}.png`) })
      await page.locator('.insight-trend-workspace').screenshot({ path: join(reviewDir, `trend-${name}.png`) })
      await page.locator('[data-dimension-explorer]').screenshot({ path: join(reviewDir, `dimensions-${name}.png`) })
    }
    expect(runtime.calls.slice(offset)).toHaveLength(domain === 'site' ? 3 : 4)
    expect(runtime.count('/game/home')).toBe(domain === 'game' ? 2 : 0)
    runtime.assertQuiet()
  })
}
test('Domain Site artwork failure retains shared identity and entity links', async ({ page, runtime }) => {
  runtime.failImages = true
  await openRuntime(page, '/insights/sites')
  await revealImages(page, '.insight-domain-activity .insight-entity-media')
  await expect(page.locator('.insight-domain-activity img')).toHaveCount(0)
  await expect(page.locator('.insight-domain-activity [role="img"][aria-label]')).toHaveCount(2)
  await expect(page.locator('.insight-domain-activity [data-change-link]').first()).toHaveAttribute('href', '/site/41')
  await expect(page.locator('.insight-domain-activity [data-change-link]').first()).toContainText('Site fixture')
  await keyboardFocus(page.locator('.insight-domain-activity [data-change-link]').first())
  runtime.assertQuiet()
})
for (const domain of ['site', 'game'] as const) {
  test('Domain missing overview and breakdown stay independent ' + domain, async ({ page, runtime }) => {
    const path = '/insights/' + (domain === 'site' ? 'sites' : 'games')
    runtime.state.overviewFailure = true
    await openRuntime(page, path)
    await expect(page.locator('[data-domain-count]')).toHaveText('—')
    await expect(page.locator('[data-domain-header] time')).toHaveCount(0)
    await expect(page.locator('.insight-trend-context__value')).toHaveText('—')
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-slice]')).toHaveCount(8)
    runtime.state.overviewFailure = false; runtime.state.breakdownFailure = true
    await openRuntime(page, path)
    await expect(page.locator('[data-domain-count]')).toHaveText(domain === 'site' ? '238' : '213')
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-dimension-explorer] .insights-empty-state')).toHaveText('生态观测数据暂不可用')
    await expect(page.locator('[data-slice], [data-dimension-raw]')).toHaveCount(0)
    runtime.assertQuiet()
  })
  test('Domain selected slice SSR and deferred all-history ' + domain, async ({ page, request, runtime }) => {
    const path = '/insights/' + (domain === 'site' ? 'sites' : 'games')
    const metric = domain === 'site' ? 'tls13' : 'linux', dimension = domain === 'site' ? 'group' : 'tag'
    const response = await request.get(path + '?metric=' + metric + '&dimension=' + dimension + '&slice=31&range=30d')
    expect(response.status()).toBe(200)
    expect(runtime.calls).toHaveLength(domain === 'site' ? 4 : 5)
    const offset = runtime.calls.length
    const html = await openRuntime(page, path + '?metric=' + metric + '&dimension=' + dimension + '&slice=31&range=all')
    expect(html).toContain('aria-busy="true"')
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    const calls = runtime.calls.slice(offset)
    expect(calls.filter(call => call.url.pathname.endsWith('/trend'))).toHaveLength(2)
    // A request-only SSR pass cannot execute the client-deferred histories.
    const before = runtime.calls.length
    await request.get(path + '?metric=' + metric + '&dimension=' + dimension + '&slice=31&range=all')
    expect(runtime.calls.slice(before).some(call => call.url.pathname.endsWith('/trend'))).toBe(false)
    runtime.assertQuiet()
  })
  test('Domain query normalization, slice toggling and locale preserve state ' + domain, async ({ page, runtime }) => {
    const path = '/insights/' + (domain === 'site' ? 'sites' : 'games')
    const metric = domain === 'site' ? 'tls13' : 'linux', dimension = domain === 'site' ? 'group' : 'tag'
    await openRuntime(page, path + '?metric=invalid&range=bad&dimension=bad&slice=bad&keep=yes')
    await expect(page).toHaveURL(url => url.searchParams.get('metric') === domainMetricKeys[domain][0]
      && url.searchParams.get('dimension') === domainDimensionKeys[domain][0] && !url.searchParams.has('slice'))
    await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-metric', domainMetricKeys[domain][0]!)
    await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-dimension', domainDimensionKeys[domain][0]!)
    if (domain === 'game') {
      await page.locator('[data-metric-key="mac"]').click()
      await expect(page).toHaveURL(url => url.searchParams.get('metric') === 'mac')
      await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-metric', 'mac')
      await expect(page.locator('.insights-trend canvas')).toBeVisible()
    } else {
      await page.locator('[data-metric-key="security_txt"]').click()
      await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-metric', 'security_txt')
    }
    await page.locator('[data-metric-key="' + metric + '"]').click()
    await page.locator('[data-range="90d"]').click()
    await page.locator('[data-dimension="' + dimension + '"]').click()
    await expect(page.locator('[data-overlapping]')).toBeVisible()
    await page.locator('[data-slice="31"]').click()
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    await expect(page.locator('[data-slice="31"] strong')).toHaveCSS('text-decoration-line', 'underline')
    await expect(page.locator('[data-slice="31"] strong')).toHaveCSS('text-decoration-thickness', '2px')
    await page.locator('[data-slice="31"]').click()
    await expect(page.locator('.insights-slice-trend')).toHaveCount(0)
    await page.locator('[data-slice="31"]').click()
    const main = page.waitForResponse(res => res.url().includes('/metrics/' + metric + '/trend?range=all'))
    const slice = page.waitForResponse(res => res.url().includes('/breakdown/' + dimension + '/31/trend?range=all'))
    await page.locator('[data-range="all"]').click()
    await Promise.all([main, slice])
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    await page.locator('button:has(img[alt="EN"])').first().click()
    await expect(page).toHaveURL(url => url.pathname === '/en' + path)
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    const query = new URL(page.url()).searchParams
    for (const [key, value] of Object.entries({ metric, range: 'all', dimension, slice: '31', keep: 'yes' })) expect(query.get(key)).toBe(value)
    runtime.assertQuiet()
  })
}
for (const [mode, text] of [['empty', '暂无可用历史数据'], ['one', '正在积累历史数据'], ['fail', '生态观测数据暂不可用']]) {
  test('Domain independent history state ' + mode, async ({ page, runtime }) => {
    runtime.state.trendMode = mode!
    await openRuntime(page, routes[0][0])
    await expect(page.locator('.insights-trend').getByText(text!, { exact: true })).toBeVisible()
    runtime.assertQuiet()
  })
}
test('Domain panel failure and gated history leave independent content and Dark chart usable', async ({ page, runtime }) => {
  runtime.state.panelFailure = true
  await openRuntime(page, routes[2][0])
  await expect(page.locator('.insights-game-pulse__unavailable')).toBeVisible()
  await expect(page.locator('.insights-trend canvas')).toBeVisible()
  await expect(page.locator('[data-slice]')).toHaveCount(8)
  expect(await page.locator('[data-domain-activity] [data-change-link]').count()).toBeGreaterThan(0)
  runtime.state.panelFailure = false
  await openRuntime(page, routes[0][0])
  const gate = runtime.hold(url => url.pathname.endsWith('/tls13/trend'))
  await page.locator('[data-metric-key="tls13"]').click(); await gate.wait()
  await expect(page.locator('.insights-trend [aria-busy="true"]')).toBeVisible()
  gate.release(); await gate.done()
  await expect(page.locator('.insights-trend canvas')).toBeVisible()
  await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  runtime.assertQuiet()
})

for (const [route, domain] of [routes[0], routes[2]]) for (const width of [1440, 390]) {
  test(`Domain Dark ready shell ${domain} ${width}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
    await context.addInitScript(() => localStorage.setItem('theme', 'dark'))
    await openRuntime(page, route)
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-metric-rail]')).toBeVisible()
    await expect(page.locator('[data-dimension-explorer]')).toBeVisible()
    await expect(page.locator('.insights-data-info')).toBeVisible()
    await assertRuntimeSurface(page, '.insights-domain-page', 'dark')
    await layout(page, width, domain)
    expect(runtime.calls).toHaveLength(domain === 'site' ? 3 : 4)
    runtime.assertQuiet()
  })
}
