import { test, expect, css, assertSiteGroupAppearance } from '../fixtures/site-groups'

for (const theme of ['light', 'dark'] as const) {
  test(`Site Groups SSR, pagination and shared Nav consumer / ${theme}`, async ({ siteGroups }) => {
    const scene = await siteGroups.open({ width: 1000, theme, mode: 'nsfw' })
    const { page, cards, back, loadMore } = scene
    const dark = theme === 'dark'
    await assertSiteGroupAppearance(scene)
    await css(loadMore, { border: dark ? '1px solid rgba(226, 232, 240, 0.15)' : '1px solid rgba(126, 92, 58, 0.15)',
      'border-radius': '999px', 'background-color': dark ? 'rgba(226, 232, 240, 0.067)' : 'rgba(255, 250, 242, 0.42)',
      color: dark ? 'rgba(190, 208, 222, 0.72)' : 'rgba(124, 45, 18, 0.88)', 'font-size': '16px', 'line-height': '24px' })
    await cards.first().hover()
    await scene.settle(cards.first())
    await css(cards.first(), { 'background-color': dark ? 'rgb(30, 41, 59)' : 'rgb(255, 230, 191)',
      'box-shadow': dark ? 'rgba(203, 213, 225, 0.42) 0px 0px 0px 1px inset' : 'color(srgb 0.705882 0.376471 0.0941176 / 0.302824) 0px 0px 0px 1px inset' })
    const popover = page.locator('body > .site-popover')
    await expect(popover).toHaveClass(/site-popover--visible/)
    await expect(popover.locator('.site-popover__domain-name span')).toHaveText(['community.example', 'mirror.example'])
    await expect(popover.locator('.site-popover__domain-metrics')).toHaveText(['0% / 32', '100% / -'])
    await expect(popover.locator('h3')).toHaveText('Wolf Community')
    await expect(popover.locator('p')).toHaveText(scene.sites[0]!.info)
    const box = await popover.boundingBox()
    expect(box!.x).toBeGreaterThanOrEqual(0)
    expect(box!.y).toBeGreaterThanOrEqual(0)
    expect(box!.x + box!.width).toBeLessThanOrEqual(1000)
    expect(box!.y + box!.height).toBeLessThanOrEqual(900)
    await back.hover()
    await scene.settle(back)
    await css(back, { color: dark ? 'rgba(248, 250, 252, 0.96)' : 'rgba(71, 42, 20, 0.92)',
      'background-color': dark ? 'rgba(148, 163, 184, 0.12)' : 'rgba(255, 235, 205, 0.82)' })
    await scene.openSite()
    await scene.loadAll(async () => {
      await loadMore.hover()
      await scene.settle(loadMore)
      await css(loadMore, { 'background-color': dark ? 'rgba(148, 163, 184, 0.18)' : 'rgba(255, 235, 205, 0.82)',
        color: dark ? 'rgba(226, 232, 240, 0.9)' : 'rgb(124, 45, 18)' })
    })
  })

  for (const variant of ['missing', 'empty'] as const) {
    test(`Site Groups ${variant} state / ${theme}`, async ({ siteGroups }) => {
      const scene = await siteGroups.open({ variant, theme })
      await expect(scene.empty).toBeVisible()
      await expect(scene.cards).toHaveCount(0)
      await expect(scene.loadMore).toHaveCount(0)
      await assertSiteGroupAppearance(scene)
      await css(scene.empty, { color: theme === 'dark' ? 'rgba(203, 213, 225, 0.62)' : 'rgba(78, 60, 46, 0.76)', 'border-radius': '12px', 'font-size': '14px' })
      expect(scene.pageCalls()).toEqual([1])
      scene.assertQuiet()
    })
  }
}

for (const lang of ['zh', 'en'] as const) {
  test(`Site Groups Back Home keeps ${lang} locale`, async ({ siteGroups }) => {
    const scene = await siteGroups.open({ variant: 'empty', lang })
    await scene.goHome()
  })
}
