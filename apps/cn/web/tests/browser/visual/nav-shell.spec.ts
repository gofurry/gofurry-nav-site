import { test, expect } from '../fixtures/nav-shell'
import type { Locator } from '@playwright/test'

async function assertIslandMaterial(nav: Locator) {
  await expect(nav).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  await expect(nav).toHaveCSS('border-width', '0px')
  await expect(nav).toHaveCSS('box-shadow', 'none')
  const shells = nav.locator('.gf-nav__shell:visible')
  await expect(shells).toHaveCount(2)
  for (const shell of await shells.all()) {
    await expect(shell).not.toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
    await expect(shell).toHaveCSS('backdrop-filter', /blur\(/)
    await expect(shell).toHaveCSS('border-width', '0px')
    await expect(shell).toHaveCSS('box-shadow', 'none')
    await expect(shell).toHaveCSS('background-image', 'none')
    await expect(shell.locator('path')).toHaveCSS('stroke', 'none')
  }
}

for (const theme of ['light', 'dark'] as const) {
  test(`Nav Shell: Standard Desktop / ${theme}`, async ({ page, navShell }) => {
    await navShell.open({ theme })
    const { nav, toggle } = navShell
    await expect(nav.getByText('GoFurry', { exact: true })).toBeVisible()
    await navShell.assertLinks('zh')
    await navShell.assertLanguage('CN')
    await expect(nav.getByRole('link', { name: 'GitHub', exact: true })).toBeVisible()
    await expect(nav.getByRole('button', { name: '切换明暗主题图标', exact: true })).toBeVisible()
    await expect(nav.getByRole('button', { name: '模式', exact: true })).toBeVisible()
    await expect(toggle).toBeHidden()
    await assertIslandMaterial(nav)
    await expect(nav.locator('.gf-nav__utility-row > a, .gf-nav__utility-row > button')).toHaveCount(3)
    const clip = await navShell.navClip()
    await expect(page).toHaveScreenshot(`nav-shell-standard-${theme}-desktop.png`, { clip })
  })

  test(`Nav Shell: Home Overlay Desktop / ${theme}`, async ({ page, navShell }) => {
    await navShell.open({ route: '/', theme })
    const { nav } = navShell
    await expect(nav).toHaveClass(/\bgf-nav--overlay\b/)
    await expect(nav.locator('nav').getByRole('link', { name: '站点导航', exact: true })).toHaveClass(/\bgf-nav__link--active\b/)
    await navShell.assertLanguage('CN')
    await assertIslandMaterial(nav)
    const clip = await navShell.navClip()
    await expect(page).toHaveScreenshot(`nav-shell-overlay-${theme}-desktop.png`, { clip })
  })

  test(`Nav Shell: Mobile Menu / ${theme}`, async ({ page, navShell }) => {
    await navShell.open({ width: 390, theme })
    const { menu, toggle } = navShell
    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-expanded', 'true')
    await expect(toggle).toHaveClass(/\bgf-nav__mobile-toggle--open\b/)
    await expect(menu).toBeVisible()
    await expect(menu).toHaveCSS('position', 'absolute')
    await navShell.assertLinks('zh', true)
    await expect(menu.getByRole('button', { name: /模式/ })).toBeVisible()
    await expect(menu.locator('.gf-nav__mobile-action-value')).toHaveText('--')
    await navShell.assertLanguage('CN', true)
    const clip = await navShell.navClip(true)
    await expect(page).toHaveScreenshot(`nav-shell-mobile-menu-${theme}.png`, { clip })
  })

  test(`Nav Shell: Home active BottomTab / ${theme}`, async ({ page, navShell }) => {
    await navShell.open({ route: '/', width: 390, theme })
    // Home hands off as content reaches the viewport midpoint, after the top shell exits.
    const hero = await page.locator('.nav-header').boundingBox()
    const header = await navShell.nav.boundingBox()
    await navShell.scrollTo(Math.ceil(Math.max(header!.y + header!.height + 8, hero!.y + hero!.height - page.viewportSize()!.height / 2)) + 16)
    await navShell.assertBottomTabs(true)
    const { bottomTabs } = navShell
    const home = bottomTabs.getByRole('link', { name: '站点导航', exact: true })
    await expect(home).toHaveClass(/\bmobile-bottom-tabs__item--active\b/)
    await expect(bottomTabs.locator('.mobile-bottom-tabs__item--active')).toHaveCount(1)
    await expect(home).not.toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
    await expect(home).not.toHaveCSS('border-top-color', 'rgba(0, 0, 0, 0)')
    await expect(bottomTabs).toHaveCSS('backdrop-filter', /blur\(/)
    const clip = await navShell.bottomClip()
    // A route-derived active item, with neutral mouse; no hover surrogate.
    await expect(home).not.toBeFocused()
    expect(await home.evaluate(element => element.matches(':hover'))).toBe(false)
    await expect(page).toHaveScreenshot(`nav-shell-bottom-tabs-${theme}-mobile.png`, { clip })
  })
}
