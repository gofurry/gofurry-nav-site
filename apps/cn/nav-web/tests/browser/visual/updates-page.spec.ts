import { test, expect } from '../fixtures/updates'

// Chromium's isolated SVG image documents copy native settings, not the page's
// CDP media override. Let the real divider SVG's reduced-motion CSS run too.
// Scoped here: other Visual owners and the shared runner remain unchanged.
test.use({ launchOptions: { args: ['--force-prefers-reduced-motion'] } })

for (const viewport of ['desktop', 'mobile'] as const) {
  for (const theme of ['light', 'dark'] as const) {
    test(`Updates: ${theme} / ${viewport}`, async ({ page, updates }) => {
      await updates.open({ theme, ...(viewport === 'desktop'
        ? { width: 1440, height: 900 } : { width: 390, height: 844 }) })
      const background = page.locator('.gf-app-shell > [data-public-background]')
      await expect(background).toHaveAttribute('data-pattern-status', 'default')
      await expect(background).toBeVisible()
      // Isolate only unrelated fixed tools; preserve the real Updates canvas and CSS.
      await page.addStyleTag({ content: '.page-scroll-dock, .mobile-bottom-tabs-root { display: none !important; }' })
      await expect.poll(() => updates.root.evaluate(el => Array.from(el.querySelectorAll('img'))
        .every(image => image.complete && image.naturalWidth > 0))).toBe(true)
      await updates.root.evaluate(el => el.scrollIntoView({ block: 'start', behavior: 'instant' }))
      await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
      await page.mouse.move(1, 1)
      await page.evaluate(async () => {
        const finite = document.getAnimations().filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
        await Promise.all(finite.map(animation => animation.finished.catch(() => {})))
        await document.fonts.ready
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
      })

      const geometry = await updates.root.evaluate(el => {
        const rect = el.getBoundingClientRect()
        return {
          top: rect.top, left: rect.left, right: rect.right, width: rect.width, height: rect.height,
          scrollWidth: el.scrollWidth, clientWidth: el.clientWidth,
          documentWidth: document.documentElement.scrollWidth, viewportWidth: innerWidth,
          reducedMotion: matchMedia('(prefers-reduced-motion: reduce)').matches,
          infiniteAnimations: el.getAnimations({ subtree: true }).filter(animation => animation.effect?.getComputedTiming().iterations === Infinity).length,
        }
      })
      expect(Math.abs(geometry.top), 'Capture starts at the Updates root').toBeLessThanOrEqual(1)
      expect(geometry.width).toBeGreaterThan(0)
      expect(geometry.height).toBeGreaterThan(0)
      expect(geometry.left).toBeGreaterThanOrEqual(-1)
      expect(geometry.right).toBeLessThanOrEqual(geometry.viewportWidth + 1)
      expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.clientWidth + 1)
      expect(geometry.documentWidth).toBeLessThanOrEqual(geometry.viewportWidth + 1)
      expect(geometry.reducedMotion).toBe(true)
      expect(geometry.infiniteAnimations, 'Production reduced motion must stop the latest marker pulse').toBe(0)
      await updates.assertInitial()
      await expect(background).toHaveAttribute('data-pattern-status', 'default')
      await expect(page).toHaveScreenshot(`updates-${theme}-${viewport}.png`)
    })
  }
}
