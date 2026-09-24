import { test, expect, assertDetailAppearance } from '../fixtures/game-detail-contract'

for (const theme of ['light', 'dark'] as const) {
  for (const device of ['desktop', 'mobile'] as const) {
    const width = device === 'desktop' ? 1440 : 390
    test(`Detail overview ${theme} ${device}`, async ({ detail }) => {
      await detail.open({ theme, width }); detail.assertInitialReads()
      await detail.settle(); await assertDetailAppearance(detail)
      await detail.page.mouse.move(1, 1)
      const nav = await detail.page.locator('.gf-nav').boundingBox()
      const y = Math.ceil(nav!.y + nav!.height)
      detail.assertQuiet()
      await expect(detail.page).toHaveScreenshot(`game-detail-overview-${theme}-${device}.png`, { clip: { x: 0, y, width, height: detail.page.viewportSize()!.height - y } })
    })
    test(`Detail Insights ${theme} ${device}`, async ({ detail }) => {
      await detail.open({ theme, width }); await detail.tab('insights')
      await expect(detail.page.locator('[data-player-loaded-ranges]')).toHaveAttribute('data-player-loaded-ranges', '30d')
      await expect(detail.page.locator('[data-price-loaded-ranges]')).toHaveAttribute('data-price-loaded-ranges', 'CN:30d')
      await expect(detail.page.locator('[data-current-players]')).toHaveText('0 人')
      await expect(detail.page.locator('[data-player-history] canvas')).toBeVisible()
      await expect(detail.page.locator('.game-insights-price-list__low')).toHaveCSS('color', theme === 'dark' ? 'rgba(148, 163, 184, 0.66)' : 'rgba(120, 113, 108, 0.58)')
      await detail.page.mouse.move(1, 1); await detail.page.evaluate(() => (document.activeElement as HTMLElement)?.blur())
      const target = detail.page.locator(device === 'desktop' ? '[data-game-insights]' : '.game-insights-history-section')
      const clip = await detail.clip(target)
      await expect(detail.page.locator('.game-insights-price-list__discount')).toHaveCSS('color', theme === 'dark' ? 'rgb(134, 239, 172)' : 'rgb(21, 128, 61)')
      detail.assertQuiet()
      await expect(detail.page).toHaveScreenshot(`game-detail-insights-${theme}-${device}.png`, { clip })
    })
  }
  for (const [tab, target, name] of [
    ['gallery', '.game-detail-gallery', 'gallery'], ['comment', '.game-detail-comments', 'comments'],
    ['news', '.game-detail-news', 'news'], ['detail', '.game-detail-info', 'facts'],
  ]) test(`Detail ${name} ${theme} desktop`, async ({ detail }) => {
    await detail.open({ theme }); await detail.tab(tab!)
    if (tab === 'gallery') await expect(detail.page.locator('.game-detail-media-image')).toBeVisible()
    await detail.page.mouse.move(1, 1); await detail.page.evaluate(() => (document.activeElement as HTMLElement)?.blur())
    const clip = await detail.clip(detail.page.locator(target!))
    await assertDetailAppearance(detail); detail.assertQuiet()
    await expect(detail.page).toHaveScreenshot(`game-detail-${name}-${theme}-desktop.png`, { clip })
  })
  test(`Detail lightbox ${theme} mobile`, async ({ detail }) => {
    await detail.open({ theme, width: 390 }); await detail.tab('gallery')
    await detail.page.locator('.game-detail-media-image').click()
    const modal = detail.page.locator('.game-detail-lightbox')
    await expect(modal.locator('button')).toBeFocused(); await detail.settle(modal)
    await detail.page.mouse.move(1, 1)
    await expect(modal).toHaveCSS('background-color', theme === 'dark' ? 'rgba(0, 0, 0, 0.9)' : 'rgba(0, 0, 0, 0.86)')
    detail.assertQuiet()
    await expect(modal).toHaveScreenshot(`game-detail-lightbox-${theme}-mobile.png`)
  })
  test(`Detail adult lock ${theme} mobile`, async ({ detail }) => {
    await detail.open({ theme, width: 390, adult: true })
    const target = detail.page.locator('.blur-wrapper')
    await expect(target.locator('.blur-wrapper__content')).toHaveAttribute('inert', '')
    await expect(target.locator('.blur-wrapper__content')).toHaveCSS('filter', 'blur(24px)')
    await expect(target.locator('.blur-wrapper__notice')).toHaveCSS('backdrop-filter', 'blur(14px)')
    await detail.page.mouse.move(1, 1)
    const clip = await detail.clip(target, 20); detail.assertQuiet()
    await expect(detail.page).toHaveScreenshot(`game-detail-adult-locked-${theme}-mobile.png`, { clip })
  })
}
