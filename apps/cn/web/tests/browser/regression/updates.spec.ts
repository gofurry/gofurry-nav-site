import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface, releaseItems, releaseSHA, indexPath, updatesNow } from '../fixtures/updates'

test.beforeEach(async ({ page }) => { await page.clock.setFixedTime(new Date(updatesNow)) })

for (const width of [1440, 390]) test(`Bounded SSR history loads additional pages only on explicit action ${width}`, async ({ page, runtime }) => {
  runtime.state.extraCount = 25
  await page.setViewportSize({ width, height: 900 })
  const html = await openRuntime(page, '/updates')
  expect(html).toContain('历史更新 16')
  expect(html).not.toContain('历史更新 17')
  await expect(page.locator('[data-updates-index] header')).toContainText('30 篇')
  await expect(page.locator('[data-updates-index] header')).not.toContainText('产品、导航与平台演进记录')
  await expect(page.locator('[data-update-history]')).toHaveCount(20)
  expect(runtime.calls).toHaveLength(1)
  const gate = runtime.hold(url => url.pathname === indexPath && url.searchParams.get('page') === '2')
  await page.locator('[data-updates-more]').click(); await gate.wait()
  await expect(page.locator('[data-updates-more]')).toBeDisabled()
  await expect(page.locator('[data-update-history]')).toHaveCount(20)
  expect(runtime.calls).toHaveLength(2)
  gate.release(); await gate.done()
  await expect(page.locator('[data-update-history]')).toHaveCount(29)
  await expect(page.locator('[data-updates-more]')).toHaveCount(0)
  expect(runtime.calls.map(call => call.url.search)).toEqual(['?lang=zh&page=1&page_size=21', '?lang=zh&page=2&page_size=21'])
  await assertRuntimeSurface(page, '[data-updates-index]', 'light')
  runtime.assertQuiet()
})

test('More-page failure preserves the visible archive and retries only that page', async ({ page, runtime }) => {
  runtime.state.extraCount = 25; runtime.state.failedPage = 2
  await openRuntime(page, '/updates')
  await page.locator('[data-updates-more]').click()
  await expect(page.locator('[data-updates-index] [role="alert"]')).toBeVisible()
  await expect(page.locator('[data-update-history]')).toHaveCount(20)
  await expect(page.locator('[data-update-latest]')).toBeVisible()
  runtime.state.failedPage = 0
  await page.locator('[data-updates-more]').click()
  await expect(page.locator('[data-update-history]')).toHaveCount(29)
  expect(runtime.calls.map(call => call.url.searchParams.get('page'))).toEqual(['1', '2', '2'])
  runtime.assertQuiet()
})

for (const locale of ['zh', 'en'] as const) for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  const prefix = locale === 'en' ? '/en' : ''
  test(`Release index SSR and hydration ${locale} ${width} ${theme}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.clock.setFixedTime(new Date(updatesNow))
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    const browserCalls: string[] = []
    page.on('request', request => { if (new URL(request.url()).pathname.startsWith(indexPath)) browserCalls.push(request.url()) })
    const html = await openRuntime(page, prefix + '/updates')
    expect(html).toContain('data-update-latest')
    expect(html).toContain(releaseItems(locale)[0]!.title)
    expect(html).toContain('data-updates-month="2025 / 12"')
    const root = page.locator('[data-updates-index]')
    await expect(root).toHaveAttribute('data-updates-state', 'ready')
    await expect(root.locator('[data-update-history]')).toHaveCount(4)
    await expect(root.locator('[data-updates-month]')).toHaveCount(3)
    const legacy = root.locator('[data-update-history="106"]')
    await expect(legacy.locator('[data-update-version], [data-update-commit], [data-update-summary]')).toHaveCount(0)
    await expect(root.locator('[data-update-markdown]')).toHaveCount(0)
    await expect(root.locator('button')).toHaveCount(0)
    const link = root.locator('[data-update-latest] [data-update-detail-link]')
    await expect(link).toHaveAttribute('href', prefix + '/updates/109')
    await link.focus()
    await expect(link).toBeFocused()
    expect(await link.evaluate(el => getComputedStyle(el).outlineStyle)).not.toBe('none')
    await assertRuntimeSurface(page, '[data-updates-index]', theme)
    expect(browserCalls).toEqual([])
    expect(runtime.calls.map(call => call.url.pathname + call.url.search)).toEqual([indexPath + '?lang=' + locale + '&page=1&page_size=21'])
    runtime.assertQuiet()
  })

  test(`Release article SSR and hydration ${locale} ${width} ${theme}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.clock.setFixedTime(new Date(updatesNow))
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    const browserCalls: string[] = []
    page.on('request', request => { if (new URL(request.url()).pathname.startsWith(indexPath)) browserCalls.push(request.url()) })
    const html = await openRuntime(page, prefix + '/updates/108')
    expect(html).toContain(releaseItems(locale)[1]!.title)
    expect(html).toMatch(/<h2>A clearer path through GoFurry|<h2>让每一次探索更清晰/)
    const body = page.locator('[data-update-markdown]')
    await expect(body.locator('h2')).toHaveCount(1)
    await expect(body.locator('h3')).toHaveCount(1)
    await expect(body.locator('ul li')).toHaveCount(2)
    await expect(body.locator('blockquote')).toHaveCount(1)
    await expect(body.locator('pre code')).toHaveAttribute('class', 'language-text')
    await expect(body.locator('p code')).toHaveText('Site')
    await expect(body.locator('a[href^="https:"]')).toHaveAttribute('rel', 'noopener noreferrer')
    await expect(body.locator('a[href^="https:"]')).toHaveAttribute('target', '_blank')
    const image = body.getByRole('img', { name: 'GoFurry' })
    await image.scrollIntoViewIfNeeded()
    await expect.poll(() => image.evaluate(el => (el as HTMLImageElement).complete && (el as HTMLImageElement).naturalWidth > 0)).toBe(true)
    await expect(image).toHaveAttribute('loading', 'lazy')
    await expect(image).toHaveAttribute('decoding', 'async')
    await expect(image).toHaveAttribute('referrerpolicy', 'no-referrer')
    await expect(page.locator('[data-update-commit]')).toHaveAttribute('href', 'https://github.com/gofurry/gofurry-nav-site/commit/' + releaseSHA)
    await expect(page.locator('[data-update-commit]')).toHaveText('53f429a')
    await expect(page.locator('[data-update-older]')).toHaveAttribute('href', prefix + '/updates/107')
    await expect(page.locator('[data-update-newer]')).toHaveAttribute('href', prefix + '/updates/109')
    await expect(page.locator('[data-update-older]')).toContainText(locale === 'zh' ? '上一篇' : 'Previous')
    await expect(page.locator('[data-update-newer]')).toContainText(locale === 'zh' ? '下一篇' : 'Next')
    await expect(page.locator('[data-update-detail] header')).not.toContainText(locale === 'zh' ? '发布记录' : 'RELEASE NOTE')
    await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com' + prefix + '/updates/108')
    await expect(page.locator('meta[property="og:url"]')).toHaveAttribute('content', 'https://go-furry.com' + prefix + '/updates/108')
    await expect(page.locator('meta[property="og:type"]')).toHaveAttribute('content', 'article')
    await expect(page.locator('link[rel="alternate"][hreflang="en-US"]')).toHaveAttribute('href', 'https://go-furry.com/en/updates/108')
    await expect(page.locator('meta[property="article:published_time"]')).toHaveAttribute('content', releaseItems(locale)[1]!.published_at)
    await expect(page).toHaveTitle(`v3.0.0-alpha.9 — ${releaseItems(locale)[1]!.title} | GoFurry`)
    await assertRuntimeSurface(page, '[data-update-detail]', theme)
    expect(browserCalls).toEqual([])
    expect(runtime.calls.map(call => call.url.pathname + call.url.search)).toEqual([indexPath + '/108?lang=' + locale])
    runtime.assertQuiet()
  })
}

test('Index links navigate to detail and older/newer without fetching the index again', async ({ page, runtime }) => {
  await openRuntime(page, '/updates')
  await page.locator('[data-update-history="108"] [data-update-detail-link]').click()
  await expect(page).toHaveURL(/\/updates\/108$/)
  await expect(page.locator('[data-update-detail] h1')).toHaveText(releaseItems('zh')[1]!.title)
  await settleRuntime(page)
  await page.locator('[data-update-newer]').click()
  await expect(page.locator('[data-update-detail] h1')).toHaveText(releaseItems('zh')[0]!.title)
  await expect(page.locator('[data-update-newer]')).toHaveCount(0)
  await settleRuntime(page)
  expect(runtime.calls.map(call => call.url.pathname)).toEqual([indexPath, indexPath + '/108', indexPath + '/109'])
  runtime.assertQuiet()
})

test('Legacy plain text detail omits optional metadata and has only the available neighbor', async ({ page, runtime }) => {
  await openRuntime(page, '/en/updates/105')
  await expect(page.locator('[data-update-detail] [data-update-summary], [data-update-detail] [data-update-commit], [data-update-detail] [data-update-version]')).toHaveCount(0)
  await expect(page.locator('[data-update-markdown] p')).toHaveText('A plain text legacy release.')
  await expect(page.locator('[data-update-older]')).toHaveCount(0)
  await expect(page.locator('meta[name="description"]')).toHaveAttribute('content', 'Read GoFurry feature updates, experience improvements and maintenance notes.')
  expect(runtime.calls).toHaveLength(1)
  runtime.assertQuiet()
})

test('Locale switch preserves article identity and requests only the new locale detail', async ({ page, runtime }) => {
  await page.setViewportSize({ width: 1440, height: 900 })
  await openRuntime(page, '/updates/108')
  await page.getByRole('button', { name: 'EN', exact: true }).click()
  await expect(page).toHaveURL(/\/en\/updates\/108$/)
  await expect(page.locator('[data-update-detail] h1')).toHaveText(releaseItems('en')[1]!.title)
  await expect(page.locator('[data-update-markdown] h2')).toHaveText('A clearer path through GoFurry')
  await settleRuntime(page)
  expect(runtime.calls.map(call => call.url.pathname + call.url.search)).toEqual([indexPath + '/108?lang=zh', indexPath + '/108?lang=en'])
  runtime.assertQuiet()
})

for (const state of ['empty', 'error'] as const) test('Index classifies ' + state + ' and supports local recovery', async ({ page, runtime }) => {
  runtime.state.indexState = state
  await openRuntime(page, '/updates')
  await expect(page.locator('[data-updates-index]')).toHaveAttribute('data-updates-state', state)
  await expect(page.locator('[data-update-latest]')).toHaveCount(0)
  expect(runtime.calls).toHaveLength(1)
  if (state === 'error') {
    runtime.state.indexState = 'ready'
    await page.getByRole('button', { name: '重试', exact: true }).click()
    await expect(page.locator('[data-update-latest]')).toBeVisible()
    expect(runtime.calls).toHaveLength(2)
  }
  runtime.assertQuiet()
})

for (const prefix of ['', '/en']) test('Hidden/missing detail is authoritative HTTP 404 ' + (prefix || 'zh'), async ({ request, runtime }) => {
  for (const id of ['999', '0', 'invalid']) {
    const response = await request.get(prefix + '/updates/' + id)
    expect(response.status()).toBe(404)
    expect(await response.text()).not.toContain('data-update-markdown')
  }
  expect(runtime.calls).toHaveLength(1)
  runtime.assertQuiet()
})

test('Detail service failure stays HTTP 503 instead of an empty article', async ({ request, runtime }) => {
  runtime.state.detailFailure = 503
  const response = await request.get('/updates/108')
  expect(response.status()).toBe(503)
  expect(await response.text()).not.toContain('data-update-markdown')
  expect(runtime.calls.every(call => call.url.pathname === indexPath + '/108')).toBe(true)
  runtime.assertQuiet()
})
