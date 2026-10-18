import { test, expect, quote } from '../fixtures/nav-revealed-content'

for (const theme of ['light', 'dark'] as const) {
  test(`Revealed Core / ${theme}`, async ({ revealed }) => {
    const scene = await revealed.open({ width: 1000, theme })
    await scene.reveal()
    await scene.alignContent()
    await expect(scene.dock).toBeHidden()
    await expect(scene.bar.locator('.nav-transition-bar__time')).toHaveText('2026年9月18日 12:40')
    await scene.quoteTrigger.focus()
    await expect(scene.author).toHaveText(quote.author)
    await scene.groups.first().getByRole('heading', { name: '社区', exact: true }).hover()
    await expect(scene.groupPopover).toHaveText(scene.data.home.groups[0]!.info)
    await scene.inside(scene.author)
    await scene.inside(scene.groupPopover)
    const clip = await scene.clip([scene.bar, scene.spotlight, scene.groups.first(), scene.author, scene.groupPopover])
    await expect(scene.page).toHaveScreenshot(`nav-revealed-core-${theme}-desktop.png`, { clip })
  })

  test(`Site Popover / ${theme}`, async ({ revealed }) => {
    const scene = await revealed.open({ width: 1000, theme })
    await scene.reveal()
    const card = await scene.hoverTopCard()
    const rows = scene.sitePopover.locator('.site-popover__domain')
    await expect(rows).toHaveCount(2)
    await expect(scene.sitePopover.locator('.site-popover__domain-metrics')).toHaveText(['0% / 32', '100% / -'])
    await rows.first().hover()
    await expect(rows.first()).not.toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
    const clip = await scene.clip([card, scene.sitePopover])
    await expect(scene.page).toHaveScreenshot(`nav-site-popover-${theme}-desktop.png`, { clip })
  })

  test(`ToolDock Search / ${theme}`, async ({ revealed }) => {
    const scene = await revealed.open({ theme })
    await scene.reveal()
    await scene.openSearch()
    await scene.searchInput.fill('wolf')
    await expect(scene.results.locator('strong')).toHaveText(['Wolf Community', 'Wolf Archive'])
    await scene.results.first().hover()
    await expect(scene.searchInput).toBeFocused()
    const clip = await scene.clip([scene.rail, scene.searchPanel])
    await expect(scene.page).toHaveScreenshot(`nav-tool-dock-search-${theme}-desktop.png`, { clip })
  })

  test(`Spotlight / ${theme}`, async ({ revealed }) => {
    const scene = await revealed.open({ width: 960, theme })
    await scene.reveal()
    await scene.alignContent()
    await expect(scene.dock).toBeHidden()
    await expect(scene.content.locator('.spotlight-panel:visible')).toHaveCount(2)
    for (const panel of [scene.featured, scene.popular]) {
      await expect(panel.locator('.spotlight-panel__pager span')).toHaveText('1/2')
    }
    const rows = scene.featured.locator('.spotlight-site')
    await expect(rows.first().locator('.spotlight-site__rank')).toHaveAttribute('aria-label', '已浏览')
    await expect(rows.nth(1).locator('.spotlight-site__rank')).toHaveText('2')
    await rows.nth(1).hover()
    const clip = await scene.clip([scene.spotlight])
    await expect(scene.page).toHaveScreenshot(`nav-spotlight-${theme}-desktop.png`, { clip })
  })
}
