import { test, expect } from '../fixtures/nav-shell'

for (const route of ['/', '/en'] as const) for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`Localized Home shell and real reveal ${route} ${width} ${theme}`, async ({ page, navShell }) => {
    await navShell.open({ route, width, theme })
    await expect(page.locator('.nav-home-page')).toBeVisible()
    await expect(page.locator('.nav-header')).toBeVisible()
    await expect(page.locator('.search-box-shell')).toBeVisible()
    if (width === 1440) {
      await navShell.assertLinks(route === '/en' ? 'en' : 'zh')
      await navShell.assertLanguage(route === '/en' ? 'EN' : 'CN')
      await page.mouse.move(10, 400)
      await page.mouse.wheel(0, 1000)
    }
    await expect(page.locator('.nav-content')).toBeVisible()
    await expect(page.locator('.gf-footer-shell')).toHaveCount(1)
    await navShell.settle(page.locator('.nav-content'))
    await expect.poll(() => page.evaluate(() => Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - document.documentElement.clientWidth)).toBeLessThanOrEqual(1)
    navShell.assertQuiet()
  })
}
