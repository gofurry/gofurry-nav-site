import { test, expect, assertGamesHomeAppearance } from '../fixtures/games-home'

for (const theme of ['light', 'dark'] as const) {
  for (const device of ['desktop', 'mobile'] as const) {
    test(`Games Home ${theme} ${device}`, async ({ gamesHome }) => {
      const scene = await gamesHome.open({ theme, width: device === 'desktop' ? 1440 : 390 })
      if (device === 'desktop') {
        await expect(scene.sidebar).toBeVisible()
        await expect(scene.dock).toBeVisible()
      } else {
        await expect(scene.sidebar).toBeHidden()
        await expect(scene.dock).toBeHidden()
      }
      await assertGamesHomeAppearance(scene)
      await scene.visualReady()
      await expect(scene.page).toHaveScreenshot(`games-home-${theme}-${device}.png`)
      scene.assertQuiet()
    })
  }
}
