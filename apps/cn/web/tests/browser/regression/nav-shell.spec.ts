import { test, expect } from '../fixtures/nav-shell'
import { mkdir, writeFile } from 'node:fs/promises'

test('B2 Mobile Menu and Bottom Tabs motion contract', async ({ page, navShell }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await navShell.open({ width: 390 })
  const { menu, toggle, bottomTabs } = navShell
  await page.evaluate(() => {
    new MutationObserver(() => {
      const menu = document.querySelector('.mobile-menu-enter-active')
      if (menu) document.documentElement.dataset.menuEntry = JSON.stringify({
        duration: getComputedStyle(menu).transitionDuration, transform: getComputedStyle(menu).transform,
      })
    }).observe(document.body, { subtree: true, childList: true })
  })
  await toggle.click()
  await expect(menu).toHaveCSS('opacity', '1')
  expect(await page.evaluate(() => JSON.parse(document.documentElement.dataset.menuEntry!)))
    .toEqual({ duration: '0.2s, 0.2s', transform: 'matrix(1, 0, 0, 1, 0, -8)' })
  await navShell.assertLinks('zh', true)
  await toggle.click(); await expect(menu).toHaveCount(0)
  await navShell.scrollTo(73); await navShell.assertBottomTabs(true)
  await expect(bottomTabs).toHaveCSS('transition-duration', '0.22s, 0.22s')
  await expect(bottomTabs.locator('.mobile-bottom-tabs__item').first()).toHaveCSS('transition-duration', '0.16s, 0.16s')
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await expect(bottomTabs).toHaveCSS('transition-duration', '0s')
  await expect(bottomTabs.locator('.mobile-bottom-tabs__item').first()).toHaveCSS('transition-duration', '0s')
  const centered = async () => expect(await bottomTabs.evaluate(el => {
    const box = el.getBoundingClientRect(); return Math.abs(box.left + box.width / 2 - innerWidth / 2)
  })).toBeLessThan(1)
  await centered()
  await navShell.scrollTo(0); await navShell.assertBottomTabs(false); await centered()
  await navShell.scrollTo(73); await navShell.assertBottomTabs(true); await centered()
  await bottomTabs.getByRole('link').first().focus()
  await expect(bottomTabs.getByRole('link').first()).toBeFocused()
  if (process.env.GOFURRY_B2_REVIEW === '1') {
    await mkdir('.cache/139-b2-review', { recursive: true })
    await page.screenshot({ path: '.cache/139-b2-review/bottom-tabs.png', clip: await navShell.bottomClip() })
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  const read = () => page.locator('.gf-nav, .gf-nav__inner, .gf-nav__center, .gf-nav__actions, .gf-nav__links').evaluateAll(nodes =>
    nodes.map(node => { const style = getComputedStyle(node); return { owner: node.classList[0],
      values: Object.fromEntries(['transition-property', 'background-color', 'border-bottom-color', 'box-shadow', 'color',
        'width', 'height', 'max-width', 'padding', 'margin', 'gap', 'left', 'top', 'opacity', 'transform'].map(key => [key, style.getPropertyValue(key)])) } }))
  const mobileLight = await read()
  await navShell.nav.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await navShell.assertTheme('dark'); await navShell.settle(navShell.nav)
  const mobileDark = await read()
  await page.setViewportSize({ width: 1440, height: 900 }); await navShell.settle(navShell.nav)
  const desktopDark = await read()
  await navShell.language('EN').click(); await expect(page).toHaveURL(/\/en\/terms$/)
  await navShell.settle(navShell.nav)
  await mkdir('.cache/139-b2-review', { recursive: true })
  await writeFile('.cache/139-b2-review/nav-owner-final.json', JSON.stringify({ mobileLight, mobileDark, desktopDark, desktopEnglish: await read() }, null, 2))
  navShell.assertQuiet()
})

test('B2 Mobile Menu live Reduced exit releases hit testing and focus before rapid reopen', async ({ page, navShell }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await navShell.open({ width: 390, theme: 'dark' })
  await page.exposeFunction('reduceMenuMotion', async () => { await page.emulateMedia({ reducedMotion: 'reduce' }) })
  await page.evaluate(() => {
    document.addEventListener('transitionrun', event => {
      const root = event.target
      if (!(root instanceof HTMLElement) || !root.matches('.gf-nav__mobile-panel')) return
      const phase = root.classList.contains('mobile-menu-leave-active') ? 'leave' : 'enter'
      if (document.documentElement.dataset.menuRestored) {
        if (phase === 'enter') document.documentElement.dataset.menuRestored = getComputedStyle(root).transitionDuration
        return
      }
      if (root.dataset[phase + 'Reduced']) return
      root.dataset[phase + 'Reduced'] = 'true'
      const bridge = window as Window & { reduceMenuMotion?: () => Promise<void> }
      void bridge.reduceMenuMotion!().then(() => {
        if (phase === 'leave') {
          root.querySelector<HTMLElement>('a')!.focus()
          const box = root.getBoundingClientRect()
          document.documentElement.dataset.menuExit = JSON.stringify({ connected: root.isConnected,
            visibility: getComputedStyle(root).visibility, focused: root.contains(document.activeElement),
            hit: root.contains(document.elementFromPoint(box.x + 20, box.y + 20)) })
        }
        document.documentElement.dataset.menuReduced = phase
      })
    })
  })
  const { menu, toggle } = navShell
  await toggle.click()
  await expect(page.locator('html')).toHaveAttribute('data-menu-reduced', 'enter')
  await expect(menu).toHaveCSS('opacity', '1'); await expect(menu).toHaveCSS('transform', 'none')
  await navShell.assertLinks('zh', true)
  if (process.env.GOFURRY_B2_REVIEW === '1') {
    await mkdir('.cache/139-b2-review', { recursive: true })
    await menu.screenshot({ path: '.cache/139-b2-review/mobile-menu-dark.png' })
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await toggle.click()
  await expect(page.locator('html')).toHaveAttribute('data-menu-reduced', 'leave')
  await expect(toggle).toHaveAttribute('aria-expanded', 'false')
  const exit = await page.evaluate(() => JSON.parse(document.documentElement.dataset.menuExit!))
  expect(exit).toMatchObject({ focused: false, hit: false })
  if (exit.connected) expect(exit.visibility).toBe('hidden')
  if (process.env.GOFURRY_B2_REVIEW === '1') await writeFile('.cache/139-b2-review/menu-leave-state.json', JSON.stringify(exit))
  // The actual trigger must work even if Vue is still finishing its old leave.
  await toggle.click(); await expect(menu).toHaveCount(1); await expect(menu).toHaveCSS('opacity', '1')
  await toggle.click(); await expect(menu).toHaveCount(0)
  await toggle.click() // Initial Reduced entry; language navigation closes it.
  await navShell.language('EN', true).click()
  await expect(page).toHaveURL(/\/en\/terms$/); await expect(menu).toHaveCount(0)
  await page.evaluate(() => { document.documentElement.dataset.menuRestored = 'waiting' })
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await toggle.click()
  await expect(page.locator('html')).toHaveAttribute('data-menu-restored', '0.2s, 0.2s')
  await navShell.assertLinks('en', true)
  await toggle.click(); await expect(menu).toHaveCount(0)
  navShell.assertQuiet()
})

test('Desktop Nav Shell retains theme through real language navigation', async ({ page, navShell }) => {
  await navShell.open()
  const { nav, toggle } = navShell
  await expect(nav.getByText('GoFurry', { exact: true })).toBeVisible()
  await navShell.assertLinks('zh')
  await expect(nav.getByRole('link', { name: 'GitHub', exact: true })).toBeVisible()
  await expect(nav.getByRole('button', { name: '模式', exact: true })).toBeVisible()
  await expect(toggle).toBeHidden()
  await navShell.assertLanguage('CN')
  await nav.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await navShell.assertTheme('dark')
  await navShell.language('EN').click()
  await expect(page).toHaveURL(/\/en\/terms$/)
  await navShell.assertLinks('en')
  await navShell.assertLanguage('EN')
  await navShell.assertTheme('dark')
  await expect(nav).toHaveCount(1)
  navShell.assertQuiet()
})

test('Mobile Nav Shell keeps menu, locale and BottomTab state tied to real controls and scrolling', async ({ page, navShell }) => {
  await navShell.open({ width: 390 })
  const { nav, menu, toggle, bottomTabs } = navShell
  await expect(toggle).toBeVisible()
  await expect(toggle).toHaveAttribute('aria-expanded', 'false')
  await expect(menu).toHaveCount(0)
  await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(0)
  await navShell.assertBottomTabs(false)
  await toggle.click()
  await expect(toggle).toHaveAttribute('aria-expanded', 'true')
  await expect(menu).toBeVisible()
  await navShell.assertLinks('zh', true)
  await expect(menu.getByRole('button', { name: /模式/ })).toBeVisible()
  await navShell.assertLanguage('CN', true)
  await nav.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await navShell.assertTheme('dark')
  await expect(toggle).toHaveAttribute('aria-expanded', 'true')
  await expect(menu).toBeVisible()
  await navShell.language('EN', true).click()
  await expect(page).toHaveURL(/\/en\/terms$/)
  await expect(toggle).toHaveAttribute('aria-expanded', 'false')
  await expect(menu).toHaveCount(0)
  await navShell.assertTheme('dark')
  await toggle.click()
  await expect(menu).toBeVisible()
  await navShell.assertLinks('en', true)
  await expect(menu.getByRole('button', { name: /Mode/ })).toBeVisible()
  await navShell.assertLanguage('EN', true)
  await toggle.click()
  await expect(menu).toHaveCount(0)

  await navShell.scrollTo(0)
  await navShell.assertBottomTabs(false)
  await navShell.scrollTo(73)
  // Observable production state includes both scroll listeners and the 160ms sync path.
  await navShell.assertBottomTabs(true)
  const mode = bottomTabs.getByRole('button', { name: 'Mode', exact: true })
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await navShell.openBottomMode()
  const modal = page.locator('.gf-preferences-modal')
  await expect(modal).toBeVisible()
  await expect(mode).toHaveClass(/\bmobile-bottom-tabs__item--active\b/)
  await expect(mode).not.toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  await expect(mode).not.toHaveCSS('border-top-color', 'rgba(0, 0, 0, 0)')
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await expect(modal).toHaveCSS('transform', 'none')
  await modal.getByRole('button', { name: 'Cancel', exact: true }).click()
  await expect(modal).toHaveCount(0)
  await expect(mode).toBeFocused()
  await expect(mode).not.toHaveClass(/\bmobile-bottom-tabs__item--active\b/)
  await page.setViewportSize({ width: 640, height: 844 })
  await navShell.assertBottomTabs(false)
  await expect(bottomTabs).toHaveCSS('display', 'none')
  navShell.assertQuiet()
})
