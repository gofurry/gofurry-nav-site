import { test, expect } from '../fixtures/nav-shell'
import { mkdir, writeFile } from 'node:fs/promises'

for (const width of [900, 1440]) {
  test(`145-C1 continuous utility tier and brand top hysteresis at ${width}`, async ({ page, navShell }) => {
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await navShell.open({ width })
    const { nav } = navShell
    const brand = nav.locator('.gf-nav__brand'), actions = nav.locator('.gf-nav__actions')
    const tools = nav.locator('.gf-nav__utility-row')
    const centers = await nav.evaluate(root => {
      const center = (el: Element) => { const r = el.getBoundingClientRect(); return r.x + r.width / 2 }
      return { upper: [...root.querySelectorAll('.gf-nav__upper-tools > a, .gf-nav__upper-tools > button')].map(center),
        lower: [...root.querySelectorAll('.gf-nav__utility-row > a, .gf-nav__utility-row > button')].map(center) }
    })
    expect(centers.upper).toHaveLength(4); expect(centers.lower).toHaveLength(3)
    for (let i = 0; i < 3; i++) expect(Math.abs(centers.lower[i]! - (centers.upper[i]! + centers.upper[i + 1]!) / 2)).toBeLessThanOrEqual(1)
    await expect(tools.getByRole('link', { name: '去抽奖', exact: true })).toHaveAttribute('href', '/games/prize')
    await expect(tools.getByRole('link', { name: '反馈', exact: true })).toHaveAttribute('rel', 'noopener noreferrer')
    await expect(nav.locator('.gf-nav__shell-outline')).toHaveCount(0)
    for (const shell of await nav.locator('.gf-nav__shell:visible').all()) {
      await expect(shell).toHaveCSS('border-width', '0px')
      await expect(shell).toHaveCSS('box-shadow', 'none')
      await expect(shell).toHaveCSS('background-image', 'none')
      await expect(shell.locator('path')).toHaveCSS('stroke', 'none')
    }
    const top = (await actions.boundingBox())!.y
    await brand.focus()
    await navShell.scrollTo(150)
    await expect(brand).not.toHaveAttribute('inert')
    await nav.locator('.gf-nav__links a').first().focus()
    await expect(brand).toHaveAttribute('inert')
    await expect(brand).toHaveCSS('visibility', 'hidden')
    expect((await actions.boundingBox())!.y).toBe(top)
    await expect(tools).toBeVisible()
    if (process.env.GOFURRY_145_C1_REVIEW === '1') {
      await mkdir('../../../.cache/145-c1-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/145-c1-review/terms-${width}-scrolled.png` })
    }
    await navShell.scrollTo(40); await expect(brand).toHaveAttribute('inert')
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await expect(brand).toHaveCSS('transition-duration', '0s')
    await navShell.scrollTo(24); await expect(brand).not.toHaveAttribute('inert')
    await expect(brand).toHaveCSS('opacity', '1')
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await navShell.scrollTo(150); await expect(brand).toHaveAttribute('inert')
    await page.setViewportSize({ width: 390, height: 844 })
    await expect(nav).toHaveCSS('position', 'absolute')
    await expect(tools).toBeHidden()
    await navShell.assertBottomTabs(true)
    await navShell.scrollTo(0); await navShell.assertBottomTabs(false)
    navShell.assertQuiet()
  })
}

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
  await navShell.scrollTo(100); await navShell.assertBottomTabs(true)
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
  await navShell.scrollTo(100); await navShell.assertBottomTabs(true); await centered()
  await bottomTabs.getByRole('link').first().focus()
  await expect(bottomTabs.getByRole('link').first()).toBeFocused()
  if (process.env.GOFURRY_B2_REVIEW === '1') {
    await mkdir('.cache/139-b2-review', { recursive: true })
    await page.screenshot({ path: '.cache/139-b2-review/bottom-tabs.png', clip: await navShell.bottomClip() })
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  const read = () => page.locator('.gf-nav, .gf-nav__inner, .gf-nav__brand, .gf-nav__actions, .gf-nav__links').evaluateAll(nodes =>
    nodes.map(node => { const style = getComputedStyle(node); return { owner: node.classList[0],
      values: Object.fromEntries(['transition-property', 'background-color', 'border-bottom-color', 'box-shadow', 'color',
        'width', 'height', 'max-width', 'padding', 'margin', 'gap', 'left', 'top', 'opacity', 'transform'].map(key => [key, style.getPropertyValue(key)])) } }))
  await navShell.scrollTo(0)
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
  const utilities = nav.locator('.gf-nav__utility-row')
  await expect(utilities.getByRole('link', { name: 'Lottery', exact: true })).toHaveAttribute('href', '/en/games/prize')
  await expect(utilities.getByRole('button', { name: 'Search sites', exact: true })).toBeVisible()
  await expect(utilities.getByRole('link', { name: 'Feedback', exact: true })).toHaveAttribute('href', 'https://github.com/gofurry/gofurry-nav-site/issues')
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
  await navShell.scrollTo(100)
  // The measured top shell has left the viewport; the handoff is latched.
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
  await page.setViewportSize({ width: 768, height: 844 })
  await navShell.assertBottomTabs(false)
  await expect(bottomTabs).toHaveCSS('display', 'none')
  navShell.assertQuiet()
})


for (const width of [1440, 1024, 900, 768, 767, 640, 430, 390, 375]) {
  test(`145-A floating islands and responsive handoff at ${width}`, async ({ page, navShell }) => {
    await navShell.open({ width })
    const { nav, toggle, menu } = navShell
    await expect(nav).toHaveCSS('position', width >= 768 ? 'fixed' : 'absolute')
    await expect(page.locator('.gf-app-shell > main')).toHaveCSS('padding-top', '0px')
    expect(await page.locator('.gf-static-page').evaluate(el => el.getBoundingClientRect().top)).toBe(0)
    expect(await page.locator('.legal-panel').evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(width >= 768 ? 88 : 76)
    await expect(nav.locator('.gf-nav__island')).toHaveCount(2)
    const outlines = await nav.locator('.gf-nav__shell:visible svg').evaluateAll(nodes => nodes.map((node, index) => {
      const svg = node as SVGSVGElement, path = svg.querySelector('clipPath path')! as SVGPathElement
      const { width: w, height: fullHeight } = svg.viewBox.baseVal
      const h = svg.closest('.gf-nav__island')!.getBoundingClientRect().height
      // Sample the actual full silhouette, rather than assert a particular path string.
      const edge = (y: number) => {
        let low = index === 0 ? w - 60 : 0, high = index === 0 ? w : 60
        for (let n = 0; n < 20; n++) {
          const mid = (low + high) / 2
          const inside = path.isPointInFill(new DOMPoint(mid, y))
          if (inside === (index === 0)) low = mid; else high = mid
        }
        return (low + high) / 2
      }
      const top = edge(h * 0.1), bottom = edge(h * 0.9)
      const shell = svg.parentElement!, island = shell.parentElement!
      const box = island.getBoundingClientRect(), surface = shell.getBoundingClientRect()
      const voidX = index === 0 ? box.right - 3 : box.left + 3
      const voidY = box.bottom - 4
      return { drop: (edge(h * 0.02) - edge(h * 0.98)) * 56 / h,
        slope: (top - bottom) / (h * 0.8),
        straightError: Math.max(...[0.25, 0.5, 0.75].map(t => Math.abs(edge(h * t) - (top + (bottom - top) * (t - 0.1) / 0.8)))),
        fullHeight: path.getBBox().height / fullHeight,
        sized: Math.abs(w - box.width) < 0.1 && Math.abs(fullHeight - surface.height) < 0.1 && surface.width === box.width,
        clipped: getComputedStyle(shell).clipPath.startsWith('url('), blur: getComputedStyle(shell).backdropFilter,
        controlsUnclipped: getComputedStyle(island).clipPath === 'none' && getComputedStyle(island).overflow === 'visible',
        cutPassesHits: !island.closest('.gf-nav')!.contains(document.elementFromPoint(voidX, voidY)) }
    }))
    expect(outlines).toHaveLength(2)
    expect(outlines[0]!.drop, 'Brand cut runs upper right to lower left').toBeGreaterThan(0)
    expect(outlines[1]!.drop, 'Actions cut runs upper left to lower right').toBeLessThan(0)
    for (const shape of outlines) {
      expect(Math.abs(shape.drop)).toBeGreaterThanOrEqual(14)
      expect(Math.abs(shape.drop)).toBeLessThanOrEqual(22)
      expect(shape.straightError, 'Middle 80% remains a diagonal straight line').toBeLessThan(0.15)
      expect(shape.fullHeight).toBeGreaterThan(0.97)
      expect(shape).toMatchObject({ sized: true, clipped: true, blur: 'blur(10px)', controlsUnclipped: true, cutPassesHits: true })
    }
    expect(Math.abs(outlines[0]!.slope + outlines[1]!.slope)).toBeLessThan(0.01)
    const geometry = await nav.evaluate(root => {
      const [brand, actions] = [...root.querySelectorAll('.gf-nav__island')].map(el => el.getBoundingClientRect())
      const hit = document.elementFromPoint((brand!.right + actions!.left) / 2, brand!.top + 16)
      return { gap: actions!.left - brand!.right, through: !root.contains(hit),
        overflow: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth,
        hitSizes: [...root.querySelectorAll<HTMLElement>('.gf-nav__actions a, .gf-nav__actions button')]
          .filter(el => el.getClientRects().length).every(el => el.offsetWidth >= 40 && el.offsetHeight >= 44),
        controlsHit: [...root.querySelectorAll<HTMLElement>('.gf-nav__brand, .gf-nav__actions a, .gf-nav__actions button')]
          .filter(el => el.getClientRects().length).every(el => {
            const r = el.getBoundingClientRect()
            return el.contains(document.elementFromPoint(r.x + r.width / 2, r.y + r.height / 2))
          }) }
    })
    expect(geometry.gap).toBeGreaterThanOrEqual(12)
    expect(geometry).toMatchObject({ through: true, overflow: 0, hitSizes: true, controlsHit: true })
    if (width >= 768) {
      await navShell.assertLinks('zh')
      await expect(toggle).toBeHidden()
      for (const link of await nav.locator('.gf-nav__links a').all()) {
        await expect(link.locator('svg')).toHaveCount(1)
        await expect(link).toHaveText('')
        await expect(link).toHaveAttribute('title', /.+/)
      }
      const before = await nav.boundingBox()
      await navShell.scrollTo(300)
      expect((await nav.boundingBox())!.y).toBe(before!.y)
      await navShell.assertBottomTabs(false)
    } else {
      await toggle.click()
      await navShell.assertLinks('zh', true)
      await expect(menu.getByRole('link', { name: 'GitHub', exact: true })).toHaveAttribute('rel', 'noopener noreferrer')
      await menu.getByRole('link').first().focus()
      await page.keyboard.press('Escape')
      await expect(menu).toHaveCount(0)
      await expect(toggle).toBeFocused()
      await navShell.scrollTo(110)
      await navShell.assertBottomTabs(true)
      await expect(nav).toHaveAttribute('inert')
      await navShell.scrollTo(40)
      await navShell.assertBottomTabs(true)
      await navShell.scrollTo(20)
      await navShell.assertBottomTabs(false)
      await expect(nav).not.toHaveAttribute('inert')
      await navShell.scrollTo(0)
      await toggle.click()
      await menu.getByRole('link').first().focus()
      await page.setViewportSize({ width: 768, height: 600 })
      await expect(menu).toHaveCount(0)
      await expect(nav.locator('.gf-nav__brand')).toBeFocused()
      await page.setViewportSize({ width, height: 844 })
      await expect(toggle).toHaveAttribute('aria-expanded', 'false')
    }
    await navShell.scrollTo(0)
    if (process.env.GOFURRY_145_A3_REVIEW === '1' && [390, 900, 1440].includes(width)) {
      await mkdir('../../../.cache/145-a3-review', { recursive: true })
      await navShell.settle(nav)
      await page.screenshot({ path: `../../../.cache/145-a3-review/shell-${width}-light.png` })
    }
    if (process.env.GOFURRY_145_REVIEW === '1') {
      await mkdir('../../../.cache/145-a2-review', { recursive: true })
      await navShell.settle(nav)
      await page.screenshot({ path: `../../../.cache/145-a2-review/shell-${width}.png` })
    }
    navShell.assertQuiet()
  })
}

for (const width of [375, 390, 430, 640, 767]) {
  test(`145-A Home handoff follows the Hero and persists until top at ${width}`, async ({ page, navShell }) => {
    await navShell.open({ width, route: '/' })
    expect(await page.locator('.nav-header').evaluate(el => el.getBoundingClientRect().top)).toBe(0)
    await navShell.assertBottomTabs(false)
    await navShell.scrollTo(30)
    await navShell.assertBottomTabs(false)
    const start = await page.evaluate(() => {
      const hero = document.querySelector('.nav-header')!.getBoundingClientRect()
      const nav = document.querySelector('.gf-nav')!.getBoundingClientRect()
      return Math.ceil(Math.max(nav.bottom + scrollY + 8, hero.bottom + scrollY - innerHeight / 2))
    })
    await navShell.scrollTo(start + 2)
    await navShell.assertBottomTabs(true)
    if (width === 390 && process.env.GOFURRY_145_REVIEW === '1') {
      await mkdir('../../../.cache/145-a2-review', { recursive: true })
      await page.screenshot({ path: '../../../.cache/145-a2-review/mobile-body-handoff.png' })
    }
    await navShell.scrollTo(40)
    await navShell.assertBottomTabs(true)
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await navShell.scrollTo(0)
    await navShell.assertBottomTabs(false)
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await expect(navShell.bottomTabs).toHaveCSS('transition-duration', '0s')
    if (process.env.GOFURRY_145_REVIEW === '1') {
      await mkdir('../../../.cache/145-a2-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/145-a2-review/home-${width}.png` })
    }
    navShell.assertQuiet()
  })
}

test('145-A locale keeps query/hash and history restores mobile handoff without moving the page', async ({ page, navShell }) => {
  await navShell.open({ width: 640 })
  await page.goto('/terms?source=shell#review')
  await navShell.language('EN').click()
  await expect(page).toHaveURL(/\/en\/terms\?source=shell#review$/)
  await navShell.toggle.click()
  await navShell.assertLinks('en', true)
  await navShell.toggle.click()
  await navShell.scrollTo(300)
  await navShell.assertBottomTabs(true)
  await page.goBack()
  await expect(page).toHaveURL(/\/terms\?source=shell#review$/)
  await navShell.assertBottomTabs(false)
  await page.goForward()
  await expect(page).toHaveURL(/\/en\/terms\?source=shell#review$/)
  await expect.poll(() => page.evaluate(() => scrollY)).toBe(300)
  await navShell.assertBottomTabs(true)
  navShell.assertQuiet()
})


for (const theme of ['light', 'dark'] as const) {
  test(`145-A desktop Hero keeps Search clear of fixed islands (${theme})`, async ({ page, navShell }) => {
    await navShell.open({ theme, route: '/' })
    const island = await navShell.nav.locator('.gf-nav__actions').boundingBox()
    const search = await page.locator('.nav-header__search').boundingBox()
    expect(search!.y).toBeGreaterThanOrEqual(island!.y + island!.height + 4)
    await navShell.assertLinks('zh')
    if (process.env.GOFURRY_145_REVIEW === '1') {
      await mkdir('../../../.cache/145-a2-review', { recursive: true })
      await navShell.settle(navShell.nav)
      await page.screenshot({ path: `../../../.cache/145-a2-review/home-desktop-${theme}.png` })
    }
    await page.mouse.move(10, 400)
    await page.mouse.wheel(0, 1000)
    await expect(page.locator('.nav-tool-dock')).toHaveCount(0)
    await expect(page.locator('.gf-nav__utilities')).toBeVisible()
    await expect(navShell.nav).toHaveCSS('position', 'fixed')
    expect((await navShell.nav.boundingBox())!.y).toBe(0)
    expect(await page.locator('.gf-nav__utilities').evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(72)
    navShell.assertQuiet()
  })
}

test('145-A short mobile menu stays reachable with a visible keyboard target', async ({ page, navShell }) => {
  await navShell.open({ width: 390, theme: 'dark' })
  await page.setViewportSize({ width: 390, height: 320 })
  await navShell.toggle.click()
  expect(await navShell.menu.evaluate(el => el.getBoundingClientRect().bottom)).toBeLessThanOrEqual(320)
  await navShell.menu.getByRole('button', { name: /模式/ }).focus()
  await expect(navShell.menu.getByRole('button', { name: /模式/ })).toBeFocused()
  if (process.env.GOFURRY_145_REVIEW === '1') {
    await mkdir('../../../.cache/145-a2-review', { recursive: true })
    await navShell.settle(navShell.menu)
    await page.screenshot({ path: '../../../.cache/145-a2-review/mobile-menu-short.png' })
  }
  await navShell.toggle.click()
  await expect(navShell.menu).toHaveCount(0)
  navShell.assertQuiet()
})


test('145-A rotating an open Bottom Tabs Preferences returns focus to the desktop shell', async ({ page, navShell }) => {
  await navShell.open({ width: 390 })
  await navShell.scrollTo(150)
  await navShell.assertBottomTabs(true)
  await navShell.openBottomMode()
  await page.locator('#mode-setting-input').focus()
  await page.setViewportSize({ width: 900, height: 600 })
  await expect(page.locator('.gf-preferences-modal')).toHaveCount(0)
  await navShell.assertBottomTabs(false)
  await expect(navShell.nav.locator('.gf-nav__brand')).toBeFocused()
  navShell.assertQuiet()
})

for (const width of [390, 900, 1440]) {
  test(`145-A3 full diagonal shells over dark Home at ${width}`, async ({ page, navShell }) => {
    await navShell.open({ width, theme: 'dark', route: '/' })
    await navShell.settle(navShell.nav)
    const brand = navShell.nav.locator('.gf-nav__brand')
    await brand.focus()
    await page.keyboard.press('Tab')
    const focused = navShell.nav.locator(':focus')
    await expect(focused).toHaveCount(1)
    await expect(focused).toHaveCSS('outline-width', '2px')
    await expect(focused).toHaveCSS('outline-style', 'solid')
    await expect(navShell.nav.locator('.gf-nav__shell')).toHaveCount(3)
    await expect(navShell.nav.locator('.gf-nav__shell:visible')).toHaveCount(2)
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBe(0)
    if (process.env.GOFURRY_145_A3_REVIEW === '1' || process.env.GOFURRY_145_C1_REVIEW === '1') {
      const folder = process.env.GOFURRY_145_C1_REVIEW === '1' ? '145-c1-review' : '145-a3-review'
      await mkdir(`../../../.cache/${folder}`, { recursive: true })
      await page.screenshot({ path: `../../../.cache/${folder}/pattern-home-${width}-dark-focus.png` })
    }
    navShell.assertQuiet()
  })
}

for (const route of ['/', '/en'] as const) {
  test(`145-A2 inline picker has natural Tab order and responsive cleanup (${route})`, async ({ page, navShell }) => {
    await navShell.open({ width: 390, theme: 'dark', route })
    const trigger = page.locator('.search-platform-trigger')
    const picker = page.locator('.search-platform-picker')
    const options = picker.locator('button')
    await trigger.click(); await expect(options).toHaveCount(8)
    await options.first().focus()
    for (let index = 1; index < 8; index++) { await page.keyboard.press('Tab'); await expect(options.nth(index)).toBeFocused() }
    await page.keyboard.press('Tab'); await expect(options).toHaveCount(1)
    await trigger.click(); await expect(options).toHaveCount(8)
    await page.setViewportSize({ width: 768, height: 600 })
    await navShell.expectWeather()
    await expect(trigger).toBeHidden()
    await expect(options).toHaveCount(1)
    await expect(page.locator('.search-platform-row')).toBeVisible()
    await expect(page.locator('.search-category-row [aria-pressed="true"]')).toBeFocused()
    const google = route === '/' ? '谷歌' : 'Google'
    await page.locator('.search-platform-row .search-chip').filter({ hasText: google }).click()
    await page.setViewportSize({ width: 320, height: 844 })
    await expect(trigger).toHaveText(google)
    const track = page.locator('.search-category-track')
    const row = page.locator('.search-category-items')
    expect(await track.evaluate(el => el.scrollWidth >= el.clientWidth)).toBe(true)
    if (route === '/en') expect(await track.evaluate(el => el.scrollWidth > el.clientWidth)).toBe(true)
    await row.locator('button').last().click()
    expect(await page.evaluate(() => scrollY)).toBe(0)
    expect(await row.locator('button').last().evaluate(el => {
      const item = el.getBoundingClientRect(), frame = el.closest('.search-category-track')!.getBoundingClientRect()
      return item.left >= frame.left - 1 && item.right <= frame.right + 1
    })).toBe(true)
    await trigger.click(); await expect(options).toHaveCount(2)
    await page.keyboard.press('Escape'); await expect(trigger).toBeFocused()
    await navShell.scrollTo(180); await navShell.assertBottomTabs(true)
    await navShell.scrollTo(0); await navShell.assertBottomTabs(false)
    navShell.assertQuiet()
  })
}
