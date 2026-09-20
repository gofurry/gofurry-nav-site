import { test, expect } from './fixtures/footer-shell'

for (const theme of ['light', 'dark'] as const) {
  test(`Footer and Shell: ${theme} / desktop`, async ({ page, mountFooterShell }) => {
    const clip = await mountFooterShell(theme)
    await expect(page).toHaveScreenshot(`footer-${theme}-desktop.png`, { clip })
  })
}
