import { test, expect } from './fixtures/static-pages'

for (const viewport of ['desktop', 'mobile'] as const) {
  for (const theme of ['light', 'dark'] as const) {
    test(`Static About: ${theme} / ${viewport}`, async ({ page, mountStaticPage }) => {
      await mountStaticPage({ page: 'about', theme, viewport })
      await expect(page).toHaveScreenshot(`static-about-${theme}-${viewport}.png`)
    })
  }
}

for (const theme of ['light', 'dark'] as const) {
  test(`Static Terms: ${theme} / mobile`, async ({ page, mountStaticPage }) => {
    await mountStaticPage({ page: 'terms', theme, viewport: 'mobile' })
    await expect(page).toHaveScreenshot(`static-terms-${theme}-mobile.png`)
  })
}
