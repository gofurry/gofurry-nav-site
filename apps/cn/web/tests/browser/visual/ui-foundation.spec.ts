import { test, expect } from './fixtures/ui-foundation'

for (const viewport of ['desktop', 'mobile'] as const) {
  for (const theme of ['light', 'dark'] as const) {
    test(`shared primitive foundation: ${theme} / ${viewport}`, async ({ mountFoundation }) => {
      const foundation = await mountFoundation({ theme, viewport })
      await expect(foundation).toHaveScreenshot(`foundation-${theme}-${viewport}.png`)
    })
  }
}
