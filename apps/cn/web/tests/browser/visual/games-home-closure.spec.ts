import { test, expect, gamesHomeClosureTime, assertNewsAppearance, assertLatestReviewsAppearance } from '../fixtures/games-home'

test.use({ gamesHomeFixedNow: gamesHomeClosureTime, timezoneId: 'UTC' })

for (const theme of ['light', 'dark'] as const) {
  for (const device of ['desktop', 'mobile'] as const) {
    test(`News ${theme} ${device}`, async ({ gamesHome }) => {
      const scene = await gamesHome.open({ theme, width: device === 'desktop' ? 1440 : 390, dataset: 'news-populated' })
      expect((scene.rendered.match(/class="news-card"/g) || [])).toHaveLength(3)
      await expect(scene.news.locator('.news-card')).toHaveCount(3)
      await assertNewsAppearance(scene)
      await expect(scene.news.locator('.news-pager__count')).toHaveText('1/3')
      const clip = await scene.closureClip(scene.news)
      await expect(scene.page).toHaveScreenshot(`games-home-news-${theme}-${device}.png`, { clip })
      scene.assertQuiet()
    })
  }
  test(`Latest Reviews ${theme} desktop`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ theme, dataset: 'reviews-populated' })
    await expect(scene.reviews.locator('.latest-review-item')).toHaveCount(3)
    for (const [index, time] of ['5 min ago', '2 hours ago', '3 days ago'].entries()) {
      expect(scene.rendered).toContain(time)
      await expect(scene.reviews.locator('.latest-review-item__meta').nth(index)).toContainText(time)
    }
    await assertLatestReviewsAppearance(scene)
    const clip = await scene.closureClip(scene.reviews)
    await expect(scene.page).toHaveScreenshot(`games-home-latest-reviews-${theme}-desktop.png`, { clip })
    scene.assertQuiet()
  })
}
