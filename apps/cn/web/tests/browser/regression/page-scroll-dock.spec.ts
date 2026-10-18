import { test, expect } from '../fixtures/page-scroll-dock'

test('PageScrollDock tracks document progress and steps back by a quarter', async ({ page, scrollDock }) => {
  await scrollDock.openTerms()
  const { dock } = scrollDock
  const maxScroll = await scrollDock.maxScroll()
  expect(maxScroll).toBeGreaterThan(320)
  await expect(dock).toHaveText('0%')
  await expect(dock).toHaveAttribute('aria-label', '当前位置 0%')
  await expect(dock).toHaveAttribute('title', '向上回退 25%')
  await expect(dock).not.toHaveClass(/\bpage-scroll-dock--visible\b/)
  await expect(dock).toHaveCSS('pointer-events', 'none')

  await scrollDock.scrollToProgress(50)
  await expect(dock).toHaveClass(/\bpage-scroll-dock--visible\b/)
  await expect(dock).toHaveText('50%')
  await expect(dock).toHaveCSS('opacity', '0.78')
  await expect(dock).toHaveCSS('pointer-events', 'auto')
  await dock.click()
  // One pixel of rounding in the real document, not a quarter of the viewport.
  await expect.poll(async () => Math.abs(await scrollDock.scrollProgress() - 25)).toBeLessThanOrEqual(100 / maxScroll)
  await expect(dock).toHaveText('25%')
  await expect(dock).toHaveAttribute('aria-label', '当前位置 25%')
  await expect(dock).toHaveClass(/\bpage-scroll-dock--visible\b/)

  await page.setViewportSize({ width: 390, height: 844 })
  await expect(dock).toHaveCount(0)
  scrollDock.assertQuiet()
})
