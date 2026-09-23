import { test, expect, assertSearchAppearance, assertFilterAppearance, searchClip } from '../fixtures/games-search'

for (const theme of ['light', 'dark'] as const) {
  for (const device of ['desktop', 'mobile'] as const) {
    test(`Search results ${theme} ${device}`, async ({ search }) => {
      await search.page.setViewportSize(device === 'desktop' ? { width: 1440, height: 900 } : { width: 390, height: 844 })
      await search.open({ theme, query: { pageSize: device === 'desktop' ? '4' : '2' } })
      await assertSearchAppearance(search, theme)
      await search.page.mouse.move(1, 1)
      const clip = await searchClip(search, '.search-toolbar, .search-result-shell', 8)
      await expect(search.page).toHaveScreenshot(`games-search-results-${theme}-${device}.png`, { clip })
    })

    test(`Search Filter ${theme} ${device}`, async ({ search }) => {
      await search.page.setViewportSize(device === 'desktop' ? { width: 1440, height: 900 } : { width: 390, height: 844 })
      await search.open({ theme, fixedNow: '2026-09-18T12:40:00Z' })
      await search.openFilter(); await search.leaf('zh Leaf 1-1').click()
      if (device === 'desktop') {
        await search.filter.locator('.dp__input').first().click()
        await expect(search.filter.locator('.dp__menu')).toBeVisible()
      } else {
        await search.filter.evaluate(element => { (document.activeElement as HTMLElement | null)?.blur(); element.scrollTop = 0 })
        // Scroll the real dialog's own viewport to the title after selecting a lower tag.
        await search.filter.locator('.game-search-filter-title').scrollIntoViewIfNeeded()
      }
      await search.page.mouse.move(1, 1)
      await assertFilterAppearance(search, theme)
      const clip = await searchClip(search, '.game-search-filter-panel', 24)
      await expect(search.page).toHaveScreenshot(`games-search-filter-${theme}-${device}.png`, { clip })
      await search.page.keyboard.press('Escape')
      if (device === 'desktop') await search.page.keyboard.press('Escape')
    })
  }

  test(`Search Jump ${theme} desktop`, async ({ search }) => {
    await search.open({ theme })
    await search.page.locator('.game-search-page-button').filter({ hasText: /^\.\.\.$/ }).first().click()
    const dialog = search.page.locator('.game-search-jump-dialog')
    await expect(dialog.locator('input')).toBeFocused(); await dialog.locator('input').fill('4')
    await expect(dialog).toHaveCSS('background-color', theme === 'dark' ? 'rgba(15, 23, 42, 0.94)' : 'rgba(255, 250, 242, 0.94)')
    await expect(dialog).toHaveCSS('border-radius', '10.4px')
    await search.page.mouse.move(1, 1)
    const clip = await searchClip(search, '.game-search-jump-dialog', 32)
    await expect(search.page).toHaveScreenshot(`games-search-jump-${theme}-desktop.png`, { clip })
    await search.page.keyboard.press('Escape')
  })

  for (const home of [true, false]) {
    test(`Shared simple search ${home ? 'Home' : 'Search'} ${theme}`, async ({ search }) => {
      await search.open({ home, theme }); await search.input.fill('wolf')
      const panel = search.page.locator('.search-results-panel')
      await expect(panel.locator('.search-result-card')).toHaveCount(4)
      await expect(search.input).toBeFocused()
      await expect(panel).toHaveCSS('border-radius', '13.12px')
      await expect(panel).toHaveCSS('backdrop-filter', 'blur(8px)')
      await expect(panel.locator('.search-result-title').first()).toHaveCSS('font-size', '13.12px')
      await expect(panel.locator('.search-result-title').first()).toHaveCSS('line-height', '15.088px')
      await search.page.mouse.move(1, 1)
      const clip = await searchClip(search, '.search-shell, .search-results-panel')
      await expect(search.page).toHaveScreenshot(`games-${home ? 'home' : 'search'}-simple-search-${theme}-desktop.png`, { clip })
    })
  }
}
