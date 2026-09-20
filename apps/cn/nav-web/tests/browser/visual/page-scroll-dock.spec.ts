import { test, expect } from '../fixtures/page-scroll-dock'

for (const hovered of [false, true]) {
  test(`PageScrollDock: Light 50% ${hovered ? 'hover' : 'normal'}`, async ({ page, scrollDock }) => {
    await scrollDock.openTerms()
    const { dock } = scrollDock
    await scrollDock.scrollToProgress(50)
    await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
    await page.mouse.move(1, 1)
    await expect(dock).toHaveClass(/\bpage-scroll-dock--visible\b/)
    await expect(dock).toHaveText('50%')
    await expect(dock).toHaveAttribute('aria-label', '当前位置 50%')
    await expect(dock).toHaveAttribute('title', '向上回退 25%')
    await expect(dock).toHaveCSS('opacity', '0.78')
    await expect(dock).toHaveCSS('pointer-events', 'auto')
    await expect(dock).toHaveCSS('width', '46px')
    await expect(dock).toHaveCSS('height', '46px')
    await expect(dock).not.toHaveCSS('box-shadow', 'none')
    if (hovered) {
      await dock.hover()
      await expect(dock).toHaveText('50%')
      await expect(dock).toHaveCSS('opacity', '0.93')
      await expect(dock).not.toHaveCSS('filter', 'none')
      await expect(dock).not.toHaveCSS('box-shadow', 'none')
    }
    await page.evaluate(async () => {
      await document.fonts.ready
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
    })
    const box = await dock.boundingBox()
    expect(box).not.toBeNull()
    const viewport = page.viewportSize()!
    // Page clip uses viewport coordinates. Keep the real shadow/canvas around
    // the button, including its hover state, rather than cropping at its border.
    const x = Math.max(0, Math.floor(box!.x - 20)), y = Math.max(0, Math.floor(box!.y - 20))
    const right = Math.min(viewport.width, Math.ceil(box!.x + box!.width + 20))
    const bottom = Math.min(viewport.height, Math.ceil(box!.y + box!.height + 20))
    const clip = { x, y, width: right - x, height: bottom - y }
    expect(clip.width).toBeGreaterThan(box!.width)
    expect(clip.height).toBeGreaterThan(box!.height)
    scrollDock.assertQuiet()
    await expect(page).toHaveScreenshot(`page-scroll-dock-light-50${hovered ? '-hover' : ''}.png`, { clip })
  })
}
