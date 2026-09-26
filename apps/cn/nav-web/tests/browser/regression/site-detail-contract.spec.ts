import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'

const initialPaths = ['/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights', '/api/v2/nav/sites/41/view']

for (const prefix of ['', '/en']) for (const query of ['', '?domain=target.example']) {
  test(`Site/Target SSR ownership ${prefix || 'zh'} ${query || 'entity'}`, async ({ request, runtime }) => {
    const response = await request.get(prefix + '/site/41' + query)
    expect(response.status()).toBe(200)
    const html = await response.text()
    expect(html).toContain('data-site-detail')
    expect(html).toContain('data-site-capabilities-state="ready"')
    expect(html).toContain('data-site-capability-state="unknown"')
    expect(html).toContain('data-site-capability-state="unavailable"')
    expect(html).toContain('data-site-capability-state="unsupported"')
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(initialPaths.slice(0, 2))
    const detail = runtime.calls.find(call => call.url.pathname.endsWith('/detail'))!
    expect(detail.url.searchParams.get('lang')).toBe(prefix ? 'en' : 'zh')
    expect(detail.url.searchParams.get('target')).toBe(query ? 'target.example' : null)
    expect(runtime.calls.find(call => call.url.pathname.endsWith('/insights'))!.url.search).toBe('')
    runtime.assertQuiet()
  })
}

for (const query of ['', '?tab=observation&view=dns', '?tab=security&view=tls', '?tab=insights&metric=tls13&range=90d']) {
  test('hydrated target switch fetches only detail and preserves workspace ' + (query || 'overview'), async ({ page, runtime }) => {
    const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view')
      && response.request().method() === 'POST')
    await openRuntime(page, '/site/41' + query)
    expect((await view).status()).toBe(200)
    await (await view).finished()
    await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'target.example')
    await expect(page.getByText('HTTP 200', { exact: true }).first()).toBeVisible()
    const startsInInsights = query.includes('tab=insights')
    if (!startsInInsights) await page.locator('[data-site-primary-tab="insights"]').click()
    const insightsBefore = await page.locator('[data-site-insights]').textContent()
    if (!startsInInsights) {
      await page.goBack()
      await expect(page).toHaveURL('/site/41' + query)
    }
    expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(initialPaths)

    await page.locator('[data-site-target-trigger]').click()
    const option = page.locator('[data-site-target-option="alt.example"]')
    await expect(option).toBeVisible()
    const detail = page.waitForResponse(response => {
      const url = new URL(response.url())
      return url.pathname.endsWith('/sites/41/detail') && url.searchParams.get('target') === 'alt.example'
    })
    await option.click()
    expect((await detail).status()).toBe(200)
    await (await detail).finished()
    await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
    await expect(page.getByText('HTTP 201', { exact: true }).first()).toBeVisible()
    await expect(page).toHaveURL(url => url.pathname === '/site/41' && url.searchParams.get('domain') === 'alt.example')
    const expected = new URLSearchParams(query)
    expected.set('domain', 'alt.example')
    expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual(Object.fromEntries(expected))
    if (!startsInInsights) await page.locator('[data-site-primary-tab="insights"]').click()
    await expect(page.locator('[data-site-insights]')).toHaveText(insightsBefore!)
    await assertRuntimeSurface(page, '[data-site-detail]', 'light')
    expect(runtime.calls.slice(3).map(call => [call.url.pathname, call.url.searchParams.get('target')])).toEqual([
      ['/api/v2/nav/sites/41/detail', 'alt.example'],
    ])
    expect(runtime.count('/sites/41/insights')).toBe(1)
    expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.calls.every(call => call.completed)).toBe(true)
    expect(new URL((await page.locator('link[rel="canonical"]').getAttribute('href'))!).search).toBe('')
    runtime.assertQuiet()
  })
}

test('invalid UI query falls back without becoming a business target or issuing extra requests', async ({ request, runtime }) => {
  for (const [query, tab] of [['tab=banana', 'overview'], ['tab=observation&view=tls', 'observation'],
    ['tab=security&view=dns', 'security'], ['tab=insights&metric=banana&range=7d', 'insights'],
    ['tab&view&metric&range', 'overview'], ['tab=security&tab=insights&view=tls', 'security']]) {
    const start = runtime.calls.length
    const response = await request.get('/site/41?domain=alt.example&' + query)
    expect(response.status()).toBe(200)
    const html = await response.text()
    expect(html).toContain('data-site-target="alt.example"')
    expect(html).toContain(`data-site-tab="${tab}"`)
    if (tab === 'overview') expect(html).toContain('data-site-capabilities-state="ready"')
    if (tab === 'insights') expect(html).toContain('data-site-insights-state="ready"')
    expect(runtime.calls.slice(start).map(call => call.url.pathname).sort()).toEqual(initialPaths.slice(0, 2))
    const detail = runtime.calls.slice(start).find(call => call.url.pathname.endsWith('/detail'))!
    expect([...detail.url.searchParams.keys()].sort()).toEqual(['lang', 'payload_mode', 'target'])
  }
  runtime.assertQuiet()
})

for (const scenario of ['missing-site', 'invalid-target', 'detail-failure'] as const) {
  test('detail is authoritative: ' + scenario, async ({ request, runtime }) => {
    if (scenario === 'detail-failure') runtime.state.failure = 'site'
    const path = scenario === 'missing-site' ? '/site/999999999'
      : scenario === 'invalid-target' ? '/site/41?domain=not-owned.example&tab=banana' : '/site/41'
    const response = await request.get(path)
    expect(response.status()).toBe(scenario === 'detail-failure' ? 503 : 404)
    expect(await response.text()).not.toContain('data-site-detail')
    expect(runtime.count('/view')).toBe(0)
    runtime.assertQuiet()
  })
}

for (const scenario of ['empty', 'unavailable', 'view-failure'] as const) {
  test('optional Site slices remain distinct: ' + scenario, async ({ page, runtime }) => {
    runtime.state.insightsEmpty = scenario === 'empty'
    runtime.state.siteInsightsFailure = scenario === 'unavailable'
    runtime.state.viewFailure = scenario === 'view-failure'
    const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
    const html = await openRuntime(page, '/site/41?domain=alt.example&tab=insights')
    const state = scenario === 'view-failure' ? 'ready' : scenario
    expect(html).toContain(`data-site-insights-state="${state}"`)
    expect((await view).status()).toBe(scenario === 'view-failure' ? 503 : 200)
    await (await view).finished()
    await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
    await expect(page.locator('[data-site-insights]')).toHaveAttribute('data-site-insights-state', state)
    await expect(page.locator('[data-site-insights-unavailable]')).toHaveCount(scenario === 'unavailable' ? 1 : 0)
    await expect(page.locator('[data-entity-timeline]')).toHaveCount(scenario === 'unavailable' ? 0 : 1)
    if (scenario === 'empty') await expect(page.locator('[data-entity-timeline] li')).toHaveCount(0)
    await settleRuntime(page)
    expect(runtime.count('/sites/41/detail')).toBe(1)
    expect(runtime.count('/sites/41/view')).toBe(1)
    // Existing GET transport retries once on 503; hydration never adds a fetch.
    expect(runtime.count('/sites/41/insights')).toBe(scenario === 'unavailable' ? 2 : 1)
    runtime.assertQuiet()
  })
}

test('a failed hydrated Target switch reaches the authoritative page error', async ({ page, runtime }) => {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  await openRuntime(page, '/site/41')
  await (await view).finished()
  runtime.state.failure = 'site'
  await page.locator('[data-site-target-trigger]').click()
  await page.locator('[data-site-target-option="alt.example"]').click()
  await expect(page).toHaveTitle('503 - GoFurry')
  await expect(page.locator('[data-site-detail]')).toHaveCount(0)
  await settleRuntime(page)
  // Client and Nitro proxy each retain one GET retry on injected 503:
  // one successful SSR call + two browser attempts × two proxy attempts.
  expect(runtime.count('/sites/41/detail')).toBe(5)
  expect(runtime.count('/sites/41/insights')).toBe(1)
  expect(runtime.count('/sites/41/view')).toBe(1)
  runtime.assertQuiet()
})

test('existing Ping history loads only on sample interaction and is not refetched by Target selection', async ({ page, runtime }) => {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  await openRuntime(page, '/site/41?tab=observation')
  await (await view).finished()
  await expect(page.locator('[data-site-history-points]')).toHaveAttribute('data-site-history-points', '0')
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(initialPaths)
  const history = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/observations'))
  await page.locator('[data-site-history-sample="twenty"]').click()
  expect((await history).status()).toBe(200)
  await (await history).finished()
  await expect(page.locator('[data-site-history-points]')).toHaveAttribute('data-site-history-points', '1')
  await page.locator('[data-site-history-sample="sixty"]').click()
  await page.locator('[data-site-target-trigger]').click()
  await page.locator('[data-site-target-option="alt.example"]').click()
  await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
  await expect(page.locator('[data-site-history-points]')).toHaveAttribute('data-site-history-points', '0')
  await settleRuntime(page)
  expect(runtime.calls.slice(3).map(call => call.url.pathname)).toEqual([
    '/api/v2/nav/sites/41/targets/target.example/observations', '/api/v2/nav/sites/41/detail',
  ])
  expect(runtime.count('/sites/41/insights')).toBe(1)
  expect(runtime.count('/sites/41/view')).toBe(1)
  runtime.assertQuiet()
})
