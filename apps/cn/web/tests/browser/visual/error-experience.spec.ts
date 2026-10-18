import { test, expect } from '../fixtures/error-experience'

for (const viewport of ['desktop', 'mobile'] as const) {
  for (const theme of ['light', 'dark'] as const) {
    test(`Error 404: ${theme} / ${viewport}`, async ({ page, errorExperience }) => {
      await errorExperience.open404({ theme, ...(viewport === 'desktop'
        ? { width: 1440, height: 900 } : { width: 390, height: 844 }) })
      await expect.poll(() => page.locator('img').evaluateAll(images => images.every(image => image.complete && image.naturalWidth > 0))).toBe(true)
      await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
      await page.mouse.move(1, 1)
      await page.evaluate(async () => {
        const finite = document.getAnimations().filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
        await Promise.all(finite.map(animation => animation.finished.catch(() => {})))
        await document.fonts.ready
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
      })
      const geometry = await errorExperience.root.evaluate(el => {
        const rect = el.getBoundingClientRect()
        return {
          width: rect.width, height: rect.height, scrollWidth: el.scrollWidth, clientWidth: el.clientWidth,
          documentWidth: document.documentElement.scrollWidth, viewportWidth: innerWidth,
          reducedMotion: matchMedia('(prefers-reduced-motion: reduce)').matches,
        }
      })
      expect(geometry.width).toBeGreaterThan(0)
      expect(geometry.height).toBeGreaterThan(0)
      expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.clientWidth + 1)
      expect(geometry.documentWidth).toBeLessThanOrEqual(geometry.viewportWidth + 1)
      expect(geometry.reducedMotion).toBe(true)
      errorExperience.assertQuiet()
      await expect(errorExperience.root).toHaveScreenshot(`error-404-${theme}-${viewport}.png`)
    })
  }
}
