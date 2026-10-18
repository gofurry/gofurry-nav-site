import { test, expect, assertSiteGroupAppearance } from '../fixtures/site-groups'

for (const theme of ['light', 'dark'] as const) {
  for (const viewport of ['desktop', 'mobile'] as const) {
    test(`Site Groups / ${theme} / ${viewport}`, async ({ siteGroups }) => {
      const scene = await siteGroups.open({ width: viewport === 'desktop' ? 1440 : 390, height: 900, theme })
      await expect(scene.cards).toHaveCount(24)
      await assertSiteGroupAppearance(scene)
      await expect(scene.page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
      expect(await scene.panel.locator('.nav-site-grid').evaluate(element => getComputedStyle(element).gridTemplateColumns.split(' ').length))
        .toBe(viewport === 'desktop' ? 4 : 1)
      const clip = await scene.visualClip()
      await expect(scene.page).toHaveScreenshot(`site-group-${theme}-${viewport}.png`, { clip })
      scene.assertQuiet()
    })
  }
}
