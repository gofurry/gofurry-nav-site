import { test, expect, assertSearchAppearance } from '../fixtures/games-search'

for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`English Search ready shell ${width} ${theme}`, async ({ search }) => {
    await search.page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
    await search.open({ locale: 'en', theme })
    await assertSearchAppearance(search, theme)
    await expect(search.page.locator('.games-search-page')).toBeVisible()
    await expect(search.page.locator('.search-result-page-slide:not([aria-hidden="true"]) .search-result-grid')).toBeVisible()
    await expect(search.page.locator('.gf-pagination')).toBeVisible()
    await expect(search.page.locator('.game-search-pagination-total')).toContainText('29')
    await search.page.evaluate(async () => {
      await document.fonts.ready
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)))
    })
    expect(await search.page.evaluate(() => Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - document.documentElement.clientWidth)).toBeLessThanOrEqual(1)
    expect(search.calls('advanced')).toHaveLength(1)
    expect(search.calls('tags')).toHaveLength(1)
    search.assertQuiet()
  })
}

for (const width of [390, 1440]) for (const locale of ['zh', 'en'] as const) {
  test(`Explicit category and leaf-only tag contract (${width} ${locale})`, async ({ search }) => {
    await search.page.setViewportSize({ width, height: 900 })
    await search.open({ locale, legacyTags: true }); await search.openFilter()
    await expect(search.filter.locator('.game-search-filter-group-title')).toHaveText('Catalog Category')
    const chip = search.filter.locator('.game-search-filter-chip').filter({ hasText: /^Leaf Tag 4$/ })
    await expect(search.filter.locator('.game-search-filter-group-title').locator('..').locator('.game-search-filter-chip')).toHaveCount(1)
    await chip.click(); await expect(chip).toHaveAttribute('aria-pressed', 'true')
    await search.apply(); await search.waitResults('Result')
    await expect.poll(() => search.calls('advanced').length).toBe(2)
    expect(search.calls('advanced')[1]!.body).toMatchObject({ tag_list: [812345] })
    expect(new URL(search.page.url()).searchParams.get('tagList')).toBe('812345')
    await search.openFilter(); await expect(chip).toHaveAttribute('aria-pressed', 'true')
    await search.cancel(); search.assertQuiet()
  })
}

for (const locale of ['zh', 'en'] as const) {
  test(`Initial Search keeps its CSR/SSR shell and noindex contract (${locale})`, async ({ search }) => {
    await search.open({ locale })
    expect(search.rendered).not.toContain('A deterministic search lifecycle result.')
    expect(search.rendered.includes('games-search-page')).toBe(locale === 'en')
    await expect(search.page.locator('meta[name="robots"]')).toHaveAttribute('content', /noindex/)
    expect(search.calls('advanced')).toHaveLength(1); expect(search.calls('tags')).toHaveLength(1)
    await expect(search.page.locator('.game-search-pagination-total')).toContainText('29')
    await assertSearchAppearance(search, 'light')
    search.assertQuiet()
  })
}

test('Result paging and responsive grid retain visible cards without horizontal clipping', async ({ search }) => {
  await search.open()
  const page = search.page
  await page.locator('.game-search-page-button').filter({ hasText: /^2$/ }).click()
  await search.waitResults('Result 5')
  await expect(page.locator('[aria-current="page"]')).toHaveText('2')
  expect(search.calls('advanced')).toHaveLength(2)
  for (const width of [1440, 960, 700, 390, 320]) {
    await page.setViewportSize({ width, height: 900 })
    await expect.poll(() => page.evaluate(() => {
      const viewport = document.querySelector('.search-result-page-viewport')!.getBoundingClientRect()
      const cards = [...document.querySelectorAll('.search-result-page-slide:not([aria-hidden="true"]) .search-page-card')]
      return cards.length === 4 && document.documentElement.scrollWidth <= innerWidth + 1 && cards.every(card => {
        const rect = card.getBoundingClientRect()
        return rect.left >= viewport.left - 1 && rect.right <= viewport.right + 1 && rect.bottom <= viewport.bottom + 1
      })
    })).toBe(true)
  }
  search.assertQuiet()
})

for (const home of [true, false]) {
  test(`Shared simple search routes its selected result (${home ? 'Home' : 'Search'})`, async ({ search }) => {
    await search.open({ home })
    await search.input.fill('wolf')
    const panel = search.page.locator('.search-results-panel')
    await expect(panel.locator('.search-result-card')).toHaveCount(4)
    expect(search.calls('simple')).toHaveLength(1)
    for (const width of home ? [1440, 1280] : [1440, 1000, 390]) {
      await search.page.setViewportSize({ width, height: 900 })
      await expect.poll(() => panel.evaluate(element => {
        const shellWidth = element.closest('.search-shell')!.getBoundingClientRect().width
        return getComputedStyle(element).gridTemplateColumns.split(' ').length === (shellWidth >= 880 ? 4 : shellWidth >= 620 ? 3 : 2)
      })).toBe(true)
    }
    if (home) {
      await search.page.setViewportSize({ width: 390, height: 844 })
      await expect(search.input).toBeHidden()
      await search.page.setViewportSize({ width: 1440, height: 900 })
      await expect(search.input).toBeVisible()
      // Hiding the input can blur it; resume the real search interaction after resizing.
      await search.input.focus()
      await expect(panel.locator('.search-result-card')).toHaveCount(4)
    }
    search.allowDetail()
    const viewResponse = search.page.waitForResponse(response => {
      const url = new URL(response.url())
      return url.origin === new URL(search.page.url()).origin
        && url.pathname === '/api/v2/game/games/7100/view'
        && response.request().method() === 'POST' && response.status() === 200
    })
    await panel.locator('.search-result-card').first().click()
    await expect(search.page).toHaveURL(/\/games\/7100$/)
    await expect(search.page.locator('.game-detail-page')).toBeVisible()
    await (await viewResponse).finished()
    search.assertQuiet()
  })
}

test('Search review and Steam actions do not trigger card navigation; the card retains English routing', async ({ search }) => {
  await search.open({ locale: 'en' })
  const page = search.page, card = page.locator('.search-result-page-slide:not([aria-hidden="true"]) .search-page-card').first()
  await card.hover(); await card.locator('.search-review-button').click()
  await expect(page.locator('.review-dialog-backdrop')).toBeVisible()
  expect(new URL(page.url()).pathname).toBe('/en/games/search')
  await page.locator('.review-dialog__close').click()
  await expect(page.locator('.review-dialog-backdrop')).toHaveCount(0)
  const url = search.allowSteam(), popupEvent = page.waitForEvent('popup')
  await card.getByRole('link', { name: 'Steam' }).click()
  const popup = await popupEvent
  await expect(popup).toHaveURL(url); await popup.close()
  expect(new URL(page.url()).pathname).toBe('/en/games/search')
  search.allowDetail()
  const viewResponse = page.waitForResponse(response => {
    const url = new URL(response.url())
    return url.origin === new URL(page.url()).origin
      && url.pathname === '/api/v2/game/games/7100/view'
      && response.request().method() === 'POST' && response.status() === 200
  })
  await card.locator('.search-page-title').click()
  await expect(page).toHaveURL(/\/en\/games\/7100$/)
  await expect(page.locator('.game-detail-page')).toBeVisible()
  await (await viewResponse).finished()
  search.assertQuiet()
})
