import type { Page } from '@playwright/test'
import { assertRuntimeSurface } from '../fixtures/insights-runtime'
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
  }))
  expect(data.overflow).toBeLessThanOrEqual(1)
  if (width === 1440) expect(data.chart.width / data.context.width).toBeGreaterThan(domain === 'site' ? 1.8 : 4.5)
  if (width === 390) expect(data.context.top).toBeGreaterThan(data.chart.top)
}
for (const [route, domain] of routes) for (const width of [1440, 1024, 390]) {
  test('Domain SSR/hydration, full raw data, focus ' + route + ' ' + width, async ({ page, request, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
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
    const last = page.locator('[data-metric-key]').last()
    await keyboardFocus(last)
    expect(await last.evaluate(el => {
      const box = el.getBoundingClientRect(), rail = el.closest('[data-metric-rail]')!.getBoundingClientRect()
      return box.left >= rail.left - 1 && box.right <= rail.right + 1
    })).toBe(true)
    expect(runtime.calls.slice(offset)).toHaveLength(domain === 'site' ? 3 : 4)
    expect(runtime.count('/game/home')).toBe(domain === 'game' ? 2 : 0)
    runtime.assertQuiet()
  })
}
for (const domain of ['site', 'game'] as const) {
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
