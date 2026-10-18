import { test, expect, openRuntime, settleRuntime } from '../fixtures/game-collections'
import { revealImages } from '../fixtures/insights-runtime'

for (const theme of ['light', 'dark'] as const) {
  for (const scene of ['index-desktop', 'timeline-desktop', 'timeline-mobile'] as const) {
    test(`Game Collections ${theme} ${scene}`, async ({ page, context, runtime }) => {
      await context.addInitScript(value => localStorage.setItem('theme', value), theme)
      await page.setViewportSize({ width: scene.endsWith('mobile') ? 390 : 1440, height: 900 })
      await openRuntime(page, scene.startsWith('index') ? '/games/collections' : '/games/collections/collection-3')
      await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
      await revealImages(page, '.game-collection-card')
      await page.evaluate(() => window.scrollTo(0, 0))
      await settleRuntime(page)
      const name = scene.startsWith('index') ? `game-collections-index-${theme}-desktop.png` : `game-collection-timeline-${theme}-${scene.endsWith('mobile') ? 'mobile' : 'desktop'}.png`
      await expect(page.locator('.game-collections-page')).toHaveScreenshot(name)
      runtime.assertQuiet()
    })
  }
}
