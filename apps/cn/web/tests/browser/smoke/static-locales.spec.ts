import { test, expect, openRuntime, settleRuntime } from '../fixtures/static-runtime'
import { assertNavigationClearance } from '../fixtures/navigation-clearance'

test('145-C1.1 Static entry clearance', async ({ page, runtime }) => {
  await openRuntime(page, '/en/terms')
  for (const width of [390, 768, 900, 1024, 1440]) {
    await page.setViewportSize({ width, height: 900 })
    await assertNavigationClearance(page, '.gf-static-page h1', `static-${width}`)
  }
  runtime.assertQuiet()
})

// The Chinese About/Terms combinations already belong to Static Visual.
const routes = [
  ['/en/about', 'about', [1440, 390]],
  ['/en/terms', 'legal', [390]],
  ['/privacy', 'legal', [390]],
  ['/en/privacy', 'legal', [390]],
] as const

for (const [route, kind, widths] of routes) for (const width of widths) for (const theme of ['light', 'dark']) {
  test(`Static locale runtime ${route} ${width} ${theme}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    const html = await openRuntime(page, route)
    expect(html).toContain(kind + '-page')
    await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
    await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
    const root = page.locator('.gf-static-page')
    await expect(root).toBeVisible()
    await expect(root).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
    const panels = root.locator('.' + kind + '-panel')
    await expect(panels).toHaveCount(kind === 'about' ? 3 : 1)
    if (kind === 'about') {
      for (const member of [panels.nth(0), panels.nth(1)]) await expect(member.locator('.about-avatar')).toHaveCount(1)
      if (width === 1440) {
        const faolan = (await panels.nth(0).boundingBox())!
        const yukongze = (await panels.nth(1).boundingBox())!
        const contact = (await panels.nth(2).boundingBox())!
        expect(yukongze.x).toBeCloseTo(faolan.x, 0)
        expect(yukongze.y).toBeGreaterThanOrEqual(faolan.y + faolan.height)
        expect(contact.x).toBeGreaterThan(faolan.x + faolan.width)
        expect(contact.y).toBeCloseTo(faolan.y, 0)
      }
    }
    for (const panel of await panels.all()) await expect(panel).toBeVisible()
    expect(await root.locator(kind === 'about' ? '.about-link' : '.gf-static-section-list').count()).toBeGreaterThan(0)
    expect((await root.innerText()).trim().length).toBeGreaterThan(20)
    await expect.poll(() => root.locator('img').evaluateAll(images => images.every(img => img.complete && img.naturalWidth > 0))).toBe(true)
    await settleRuntime(page)
    expect(await page.evaluate(() => Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - document.documentElement.clientWidth)).toBeLessThanOrEqual(1)
    expect(runtime.calls).toHaveLength(0)
    runtime.assertQuiet()
  })
}
