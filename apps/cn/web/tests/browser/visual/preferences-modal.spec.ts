import { test, expect } from './fixtures/preferences-visual'

for (const viewport of ['desktop', 'mobile'] as const) {
  for (const theme of ['light', 'dark'] as const) {
    test(`real Preferences home: ${theme} / ${viewport}`, async ({ mountPreferencesVisual }) => {
      const backdrop = await mountPreferencesVisual({ page: 'home', theme, viewport })
      await expect(backdrop).toHaveScreenshot(`preferences-home-${theme}-${viewport}.png`)
    })
  }
  for (const page of ['background', 'routing'] as const) {
    test(`real Preferences ${page}: light / ${viewport}`, async ({ mountPreferencesVisual }) => {
      const backdrop = await mountPreferencesVisual({ page, theme: 'light', viewport })
      await expect(backdrop).toHaveScreenshot(`preferences-${page}-light-${viewport}.png`)
    })
  }
}
