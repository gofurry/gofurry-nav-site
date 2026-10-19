import { test, expect, quote, weatherURL } from '../fixtures/nav-revealed-content'
import { mkdir } from 'node:fs/promises'

test('145-C1 utility search keeps focus, cache and instant Reduced exit on the shared shell', async ({ revealed }) => {
  const scene = await revealed.open({ width: 900 })
  const { page, searchPanel, searchInput } = scene
  const trigger = scene.dock.getByRole('button', { name: '搜索站点', exact: true })
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await page.exposeFunction('reduceUtilityMotion', () => page.emulateMedia({ reducedMotion: 'reduce' }))
  await page.evaluate(() => {
    document.addEventListener('transitionrun', event => {
      if (!(event.target instanceof HTMLElement) || !event.target.matches('.gf-nav__search-panel') || document.documentElement.dataset.utilityReduced) return
      document.documentElement.dataset.utilityReduced = 'requested'
      const bridge = window as Window & { reduceUtilityMotion?: () => Promise<void> }
      void bridge.reduceUtilityMotion!().then(() => { document.documentElement.dataset.utilityReduced = 'done' })
    })
  })
  expect(scene.directoryCount()).toBe(0)
  await scene.openSearch()
  await expect(page.locator('html')).toHaveAttribute('data-utility-reduced', 'done')
  await expect(searchPanel).toHaveCSS('opacity', '1')
  await expect(searchPanel).toHaveCSS('transform', 'none')
  await searchInput.fill('wolf')
  await expect(scene.results.locator('strong')).toHaveText(['Wolf Community', 'Wolf Archive'])
  await searchInput.press('Tab'); await expect(scene.results.first()).toBeFocused()
  await scene.results.first().press('Shift+Tab'); await expect(searchInput).toBeFocused()
  const panelBox = (await searchPanel.boundingBox())!, rowBox = (await scene.rail.boundingBox())!
  expect(panelBox.y).toBeGreaterThanOrEqual(rowBox.y + rowBox.height)
  expect(panelBox.x).toBeGreaterThanOrEqual(0)
  expect(panelBox.x + panelBox.width).toBeLessThanOrEqual(900)
  if (process.env.GOFURRY_145_C1_REVIEW === '1') {
    await mkdir('../../../.cache/145-c1-review', { recursive: true })
    await page.screenshot({ path: '../../../.cache/145-c1-review/shared-search-900.png' })
  }
  await searchInput.press('Escape'); await expect(searchPanel).toHaveCount(0); await expect(trigger).toBeFocused()
  for (let i = 0; i < 3; i++) {
    await trigger.click(); await expect(searchInput).toBeFocused()
    await trigger.click(); await expect(searchPanel).toHaveCount(0)
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await trigger.click(); await expect(searchInput).toBeFocused()
  await page.locator('.search-input').click()
  await expect(searchPanel).toHaveCount(0)
  await expect(page.locator('.search-input')).toBeFocused()
  await trigger.click(); await expect(searchInput).toBeFocused()
  const preferencesTrigger = page.locator('.gf-nav__mode-button')
  await scene.openPreferences()
  await expect(searchPanel).toHaveCount(0)
  const preferences = page.locator('.gf-preferences-modal')
  await expect(preferences).toBeVisible()
  await preferences.getByRole('button', { name: '取消', exact: true }).click()
  await expect(preferences).toHaveCount(0)
  await expect(preferencesTrigger).toBeFocused()
  await trigger.click(); await expect(searchInput).toBeFocused()
  await page.locator('a[href="/about"]').first().click()
  await expect(searchPanel).toHaveCount(0)
  await expect(scene.dock).toBeVisible()
  expect(scene.directoryCount()).toBe(1)
  await expect(page.locator('.nav-tool-dock, .game-tool-dock')).toHaveCount(0)
  scene.assertQuiet()
})

for (const lang of ['zh', 'en'] as const) for (const dataset of ['full', 'partial', 'empty'] as const) {
  test(`145-B actual SSR content ${lang} ${dataset}`, async ({ revealed }) => {
    const scene = await revealed.open({ lang, dataset, width: lang === 'en' ? 390 : 1440, mode: 'nsfw' })
    const ssr = await scene.page.evaluate(html => {
      const dom = new DOMParser().parseFromString(html, 'text/html')
      const content = dom.querySelector('.nav-content')!
      return { main: dom.querySelectorAll('main').length, footer: dom.querySelectorAll('.gf-footer-shell').length,
        groups: [...content.querySelectorAll('.nav-group-title')].map(el => el.textContent?.trim()),
        panels: content.querySelectorAll('.spotlight-panel').length,
        emptyPanels: content.querySelectorAll('.spotlight-panel__empty').length,
        cards: [...content.querySelectorAll('.nav-site-card')].map(el => el.textContent),
        spotlight: [...content.querySelectorAll('.spotlight-site strong')].map(el => el.textContent),
        images: [...content.querySelectorAll('img')].map(el => [el.loading, el.decoding, el.getAttribute('fetchpriority')]),
        weather: dom.querySelector('iframe')?.getAttribute('loading') }
    }, scene.ssr)
    // Existing Spotlight keeps empty siblings when any panel has data.
    expect(ssr).toMatchObject({ main: 1, footer: 1, panels: dataset === 'empty' ? 0 : 4,
      emptyPanels: dataset === 'partial' ? 3 : 0, weather: 'lazy' })
    expect(ssr.groups).toEqual(dataset === 'full' ? ['社区', '工具'] : [])
    expect(ssr.cards).toHaveLength(dataset === 'full' ? 6 : 0)
    if (dataset === 'full') {
      expect(ssr.cards.join(' ')).toContain('Wolf Community · 发现、交流与创作')
      expect(ssr.spotlight).toContain('精选资源 1')
      expect(ssr.cards.join(' ')).not.toContain('Wolf After Dark')
      expect(ssr.spotlight).not.toContain('Night Showcase')
      // The saved client mode applies after SFW hydration, without changing SSR.
      await expect(scene.card('Wolf After Dark')).toHaveCount(1)
      await expect(scene.featured.getByText('Night Showcase', { exact: true })).toHaveCount(1)
    }
    for (const image of ssr.images) expect(image).toEqual(['lazy', 'async', 'low'])
    expect(scene.directoryCount()).toBe(0)
    expect(await scene.page.locator('.gf-footer-shell').evaluate(el => getComputedStyle(el).position)).toBe('relative')
    expect(await scene.bar.evaluate(el => el.getBoundingClientRect().bottom <= document.querySelector('.nav-content')!.getBoundingClientRect().top)).toBe(true)
    scene.assertQuiet()
  })
}

test('145-B native wheel and keyboard remain continuous across live Reduced changes', async ({ context, revealed }) => {
  await context.addInitScript(() => {
    const scroll = Element.prototype.scrollIntoView
    Element.prototype.scrollIntoView = function (options) {
      document.documentElement.dataset.productScrollCalls = String(Number(document.documentElement.dataset.productScrollCalls || 0) + 1)
      scroll.call(this, options)
    }
    window.addEventListener('wheel', event => { document.documentElement.dataset.wheelCancelled = String(event.defaultPrevented) }, { passive: true })
  })
  const scene = await revealed.open()
  const { page } = scene
  for (const motion of ['no-preference', 'reduce', 'no-preference'] as const) {
    await page.emulateMedia({ reducedMotion: motion })
    await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }))
    await page.mouse.move(10, 300)
    await page.mouse.wheel(0, 120)
    await expect.poll(() => page.evaluate(() => scrollY)).toBe(120)
    await expect(page.locator('html')).toHaveAttribute('data-wheel-cancelled', 'false')
    for (const target of [160, 200, 240]) {
      await page.mouse.wheel(0, 40)
      await expect.poll(() => page.evaluate(() => scrollY)).toBe(target)
    }
    expect(await scene.shell.evaluate(el => el.getBoundingClientRect().top)).toBe(660)
    await expect(scene.dock).toBeVisible()
    if (motion === 'no-preference' && process.env.GOFURRY_145_B_REVIEW === '1') {
      await mkdir('../../../.cache/145-b-review', { recursive: true })
      await page.screenshot({ path: '../../../.cache/145-b-review/desktop-hero-content-boundary.png' })
    }
    await page.keyboard.press('PageDown')
    await expect.poll(() => page.evaluate(() => scrollY)).toBeGreaterThan(500)
    await expect(scene.dock).toBeVisible()
    await expect(page.locator('html')).not.toHaveAttribute('data-product-scroll-calls')
  }
  expect(scene.directoryCount()).toBe(0)
  await page.keyboard.press('End')
  await expect.poll(() => page.evaluate(() => Math.abs(scrollY + innerHeight - document.documentElement.scrollHeight))).toBeLessThanOrEqual(1)
  await expect(scene.dock).toBeVisible()
  await scene.openSearch()
  await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }))
  await expect(scene.dock).toBeVisible()
  await expect(scene.searchPanel).toBeVisible()
  await scene.searchInput.press('Escape')
  await expect(scene.searchPanel).toHaveCount(0)
  expect(await page.evaluate(() => document.activeElement?.isConnected)).toBe(true)
  await page.locator('a[href="/about"]').first().click()
  await expect(scene.shell).toHaveCount(0)
  await page.mouse.wheel(0, 120)
  await expect(scene.dock).toBeVisible()
  scene.assertQuiet()
})

test('145-B touch scroll keeps mobile content and the existing bottom-tab handoff', async ({ context, revealed }) => {
  const scene = await revealed.open({ width: 390 })
  const { page } = scene
  const client = await context.newCDPSession(page)
  await client.send('Emulation.setTouchEmulationEnabled', { enabled: true })
  await client.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x: 10, y: 300 }] })
  for (const y of [280, 260, 240, 220, 200]) {
    await client.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x: 10, y }] })
    await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => resolve())))
  }
  await expect.poll(() => page.evaluate(() => scrollY)).toBeGreaterThan(30)
  expect(await page.evaluate(() => scrollY)).toBeLessThan(200)
  await client.send('Input.dispatchTouchEvent', { type: 'touchCancel', touchPoints: [] })
  await expect(scene.dock).toBeHidden()
  await page.mouse.wheel(0, 300)
  await expect(page.locator('.mobile-bottom-tabs')).toHaveAttribute('aria-hidden', 'false')
  await page.mouse.wheel(0, -150)
  await expect(page.locator('.mobile-bottom-tabs')).toHaveAttribute('aria-hidden', 'false')
  if (process.env.GOFURRY_145_B_REVIEW === '1') {
    await mkdir('../../../.cache/145-b-review', { recursive: true })
    await page.screenshot({ path: '../../../.cache/145-b-review/mobile-natural-scroll.png' })
  }
  await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }))
  await expect(page.locator('.mobile-bottom-tabs')).toHaveAttribute('aria-hidden', 'true')
  await expect(page.locator('.gf-nav')).not.toHaveAttribute('inert')
  expect(scene.directoryCount()).toBe(0)
  scene.assertQuiet()
})

test('B2 Site Popover keeps directional motion, geometry and live Reduced state', async ({ revealed }) => {
  const scene = await revealed.open({ width: 1000 })
  const { page, sitePopover } = scene
  await scene.reveal()
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await page.exposeFunction('reducePopoverMotion', async () => {
    await page.emulateMedia({ reducedMotion: 'reduce' })
    const hidden = await sitePopover.evaluate(root => {
      const style = getComputedStyle(root)
      return { opacity: style.opacity, transform: style.transform, pointer: style.pointerEvents }
    })
    // Re-enter through the actual target before the existing 320ms cleanup;
    // assertions must not rely on a later polling interval finding that node.
    await scene.card(scene.data.community.name).hover()
    return hidden
  })
  await page.evaluate(() => {
    document.addEventListener('transitionend', event => {
      const root = event.target
      if (!(root instanceof HTMLElement) || !root.matches('.site-popover--hidden-top, .site-popover--hidden-bottom')) return
      const style = getComputedStyle(root)
      const placement = root.classList.contains('site-popover--hidden-top') ? 'top' : 'bottom'
      document.documentElement.dataset['popoverExit' + placement] = JSON.stringify({
        opacity: style.opacity, transform: style.transform, filter: style.filter, pointer: style.pointerEvents,
      })
    })
    document.addEventListener('transitionrun', event => {
      const root = event.target
      if (document.documentElement.dataset.reducePopover !== 'true' || !(root instanceof HTMLElement)
        || !root.matches('.site-popover--hidden-top, .site-popover--hidden-bottom')) return
      document.documentElement.dataset.reducePopover = 'requested'
      const bridge = window as Window & { reducePopoverMotion?: () => Promise<object> }
      void bridge.reducePopoverMotion!().then(hidden => {
        document.documentElement.dataset.popoverReducedState = JSON.stringify(hidden)
        document.documentElement.dataset.reducePopover = 'done'
      })
    })
  })
  const card = await scene.hoverTopCard()
  await expect(sitePopover).toHaveCSS('transition-property', 'opacity, transform')
  await expect(sitePopover).toHaveCSS('transition-duration', '0.22s, 0.22s')
  await expect(sitePopover).toHaveCSS('will-change', 'auto')
  await expect(sitePopover).toHaveCSS('filter', 'none')
  await expect(sitePopover).toHaveCSS('backdrop-filter', 'blur(12px)')
  await expect(sitePopover).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  await expect(sitePopover.locator('.site-popover__domain-metrics')).toHaveText(['0% / 32', '100% / -'])
  if (process.env.GOFURRY_B2_REVIEW === '1') {
    await mkdir('.cache/139-b2-review', { recursive: true })
    await page.screenshot({ path: '.cache/139-b2-review/site-popover-top.png', clip: await scene.clip([card, sitePopover]) })
  }
  await page.mouse.move(1, 1)
  await expect.poll(() => page.evaluate(() => JSON.parse(document.documentElement.dataset.popoverExittop || 'null')))
    .toEqual({ opacity: '0', transform: 'matrix(1, 0, 0, 1, 0, -6)', filter: 'none', pointer: 'none' })
  await expect(sitePopover).toHaveCount(0)
  await card.evaluate(el => window.scrollBy({ top: el.getBoundingClientRect().top - 200, behavior: 'instant' }))
  await card.hover(); await expect(sitePopover).toHaveClass(/site-popover--visible/)
  await scene.settle(sitePopover); await scene.inside(sitePopover)
  expect((await sitePopover.boundingBox())!.y).toBeGreaterThanOrEqual((await card.boundingBox())!.y + (await card.boundingBox())!.height)
  await sitePopover.locator('.site-popover__domain').first().hover()
  await expect(sitePopover).toHaveCSS('pointer-events', 'auto')
  await page.mouse.move(1, 1)
  await expect.poll(() => page.evaluate(() => JSON.parse(document.documentElement.dataset.popoverExitbottom || 'null')))
    .toEqual({ opacity: '0', transform: 'matrix(1, 0, 0, 1, 0, 6)', filter: 'none', pointer: 'none' })
  await expect(sitePopover).toHaveCount(0)

  await card.hover(); await expect(sitePopover).toHaveClass(/site-popover--visible/)
  await page.evaluate(() => { document.documentElement.dataset.reducePopover = 'true' })
  await page.mouse.move(1, 1)
  await expect(page.locator('html')).toHaveAttribute('data-reduce-popover', 'done')
  expect(await page.evaluate(() => JSON.parse(document.documentElement.dataset.popoverReducedState!)))
    .toEqual({ opacity: '0', transform: 'none', pointer: 'none' })
  await expect(sitePopover).toHaveCount(1); await expect(sitePopover).toHaveCSS('opacity', '1')
  await expect(sitePopover).toHaveCSS('transition-duration', '0s')
  // Position stays tied to the target immediately; left/top never transition.
  await page.evaluate(() => window.scrollBy({ top: 40, behavior: 'instant' }))
  await scene.inside(sitePopover)
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await expect(sitePopover).toHaveCSS('transition-duration', '0.22s, 0.22s')
  await page.locator('a[href="/about"]').first().click()
  await expect(scene.shell).toHaveCount(0); await expect(sitePopover).toHaveCount(0)
  await page.emulateMedia({ reducedMotion: 'reduce' })
  scene.assertQuiet()
})

test('Reduced Spotlight settles the selected target through live preference changes', async ({ revealed }) => {
  const scene = await revealed.open({ theme: 'dark' })
  const { page, featured } = scene
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await scene.reveal()
  const next = featured.getByRole('button', { name: '精选站点 下一页', exact: true })
  const prev = featured.getByRole('button', { name: '精选站点 上一页', exact: true })
  const track = featured.locator('.spotlight-panel__track')
  const settled = async (names: string[]) => {
    await expect(track.locator('.spotlight-panel__slide')).toHaveCount(1)
    await expect(track).not.toHaveClass(/--(?:next|prev|instant)/)
    await expect(track.locator('.spotlight-site strong')).toHaveText(names)
    await expect(track.locator('.spotlight-panel__slide')).toHaveAttribute('aria-hidden', 'false')
  }
  const first = scene.data.featured.slice(0, 6).map(site => site.name)
  const last = scene.data.featured.slice(6).map(site => site.name)
  await next.focus(); await next.press('Enter'); await settled(last)
  await expect(next).toBeFocused()
  await prev.press('Enter'); await settled(first)
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await prev.press('Enter') // Normal wrap to last.
  await expect(track).toHaveCSS('animation-name', 'spotlight-panel-prev')
  await next.press('Enter') // Interrupt with the latest selected page.
  await expect(track).toHaveCSS('animation-name', 'spotlight-panel-next')
  await page.emulateMedia({ reducedMotion: 'reduce' }); await settled(first)
  await next.press('Enter'); await settled(last)
  if (process.env.GOFURRY_MOTION_REVIEW === '1') {
    await mkdir('.cache/139-a-review', { recursive: true })
    await page.screenshot({ path: '.cache/139-a-review/nav-reveal-spotlight-dark.png', clip: await scene.clip([scene.bar, scene.spotlight]) })
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await next.press('Enter')
  await expect(track).toHaveCSS('animation-duration', '0.52s')
  await settled(first)
  // Real SPA navigation disposes the local media subscription and paging work.
  await next.press('Enter')
  await expect(track).toHaveClass(/spotlight-panel__track--next/)
  await page.locator('a[href="/about"]').first().click()
  await expect(scene.shell).toHaveCount(0)
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  scene.assertQuiet()
})

test('Core cards, TransitionBar and real group/site popovers retain their runtime contract', async ({ revealed }) => {
  const scene = await revealed.open({ width: 1000 })
  const { page, bar, quoteTrigger, author, groupPopover, sitePopover } = scene
  await scene.reveal()
  await scene.alignContent()
  await expect(scene.dock).toBeVisible()
  const card = scene.card(scene.data.community.name)
  await expect(card.locator('h3')).toHaveCSS('font-size', '16px')
  await expect(card.locator('h3')).toHaveCSS('line-height', '24px')
  await expect(card.locator('h3')).toHaveCSS('font-weight', '500')
  await expect(card.locator('p')).toHaveCSS('font-size', '12px')
  // Current unlayered nav.less line-height: 1.35 wins over text-xs (16px).
  await expect(card.locator('p')).toHaveCSS('line-height', '16.2px')
  await expect(bar.locator('.nav-transition-bar__time')).toHaveText('2026年9月18日 12:40')
  await expect(quoteTrigger).toHaveText(`"${quote.content}"`)
  await expect(bar.locator('iframe')).toHaveAttribute('src', weatherURL)
  await quoteTrigger.focus()
  await expect(author).toHaveText(quote.author)
  await scene.inside(author)
  await quoteTrigger.blur()
  await expect(author).toHaveCount(0)

  await scene.groups.first().getByRole('heading', { name: '社区', exact: true }).hover()
  await expect(groupPopover).toHaveText(scene.data.home.groups[0]!.info)
  await scene.inside(groupPopover)
  await page.mouse.move(1, 1)
  await expect(groupPopover).toHaveCount(0)

  await scene.hoverTopCard()
  await expect(sitePopover.locator('h3')).toHaveText(scene.data.community.name)
  await expect(sitePopover.locator('p')).toHaveText(scene.data.community.info)
  await expect(sitePopover.locator('.site-popover__domain-name')).toHaveText(['community.example', 'mirror.example'])
  await expect(sitePopover.locator('.site-popover__domain-metrics')).toHaveText(['0% / 32', '100% / -'])
  const indicators = await sitePopover.locator('.site-popover__domain-name img').evaluateAll(images => images.map(image => (image as HTMLImageElement).src))
  // Vite inlines these real status SVGs in the production build.
  expect(decodeURIComponent(indicators[0]!)).toContain("fill='#78B159'")
  expect(decodeURIComponent(indicators[1]!)).toContain("fill='#DD2E44'")
  scene.assertQuiet()
})

test('Navigation utilities cache the real directory, filters SFW search and opens a result without a view POST', async ({ revealed }) => {
  const scene = await revealed.open()
  await scene.reveal()
  await expect(scene.dock).toBeVisible()
  await scene.openSearch()
  await scene.searchInput.fill('wolf')
  await expect(scene.results.locator('strong')).toHaveText(['Wolf Community', 'Wolf Archive'])
  await expect(scene.searchPanel.getByText('Wolf After Dark', { exact: true })).toHaveCount(0)
  await scene.searchInput.fill('zz-no-result')
  await expect(scene.searchPanel.locator('.gf-nav__search-status')).toHaveText('没有找到匹配站点')
  await scene.searchInput.press('Escape')
  await expect(scene.searchPanel).toHaveCount(0)
  await scene.openSearch()
  await expect(scene.searchInput).toHaveValue('')
  expect(scene.directoryCount()).toBe(1)
  await scene.searchInput.fill('wolf')
  await expect(scene.results).toHaveCount(2)
  await scene.clickSite(scene.results.first(), scene.data.community, false)
  await expect(scene.searchPanel).toHaveCount(0)
  scene.assertQuiet()
})

test('Spotlight pages animate, wrap and mark real site visits with the exact view request', async ({ revealed }) => {
  const scene = await revealed.open()
  await scene.page.emulateMedia({ reducedMotion: 'no-preference' })
  await scene.reveal()
  await expect(scene.content.locator('.spotlight-panel:visible')).toHaveCount(4)
  const pager = scene.featured.locator('.spotlight-panel__pager span')
  await expect(pager).toHaveText('1/2')
  await expect(scene.popular.locator('.spotlight-panel__pager span')).toHaveText('1/2')
  await expect(scene.panels.nth(2).locator('.spotlight-panel__pager')).toHaveCount(0)
  await expect(scene.panels.nth(3).locator('.spotlight-panel__pager')).toHaveCount(0)
  const first = scene.featured.locator('.spotlight-site').first()
  await expect(first.locator('.spotlight-site__rank')).toHaveAttribute('aria-label', '已浏览')
  await expect(scene.content.getByText(scene.data.adultFeatured.name, { exact: true })).toHaveCount(0)
  const track = scene.featured.locator('.spotlight-panel__track')
  const move = async (direction: '下一页' | '上一页', page: string) => {
    await scene.featured.getByRole('button', { name: `精选站点 ${direction}`, exact: true }).click()
    await expect(track).toHaveClass(direction === '下一页' ? /spotlight-panel__track--next/ : /spotlight-panel__track--prev/)
    await expect(track).toHaveCSS('animation-name', direction === '下一页' ? 'spotlight-panel-next' : 'spotlight-panel-prev')
    await expect(track).toHaveCSS('animation-duration', '0.52s')
    await expect(pager).toHaveText(page)
    await expect(track).not.toHaveClass(/spotlight-panel__track--(?:next|prev|instant)/)
    await expect(track.locator('.spotlight-panel__slide')).toHaveCount(1)
    await expect(track.locator('.spotlight-panel__slide')).toHaveAttribute('aria-hidden', 'false')
    await scene.settle(scene.featured)
  }
  await move('下一页', '2/2')
  await move('上一页', '1/2')
  await move('上一页', '2/2') // Stable first -> last wrap, through the real pager.
  await move('下一页', '1/2')
  const unvisited = scene.featured.locator('.spotlight-site').filter({ hasText: scene.data.featured[1]!.name })
  await expect(unvisited.locator('.spotlight-site__rank')).not.toHaveAttribute('aria-label', '已浏览')
  await scene.clickSite(unvisited, scene.data.featured[1]!, true)
  await expect(unvisited.locator('.spotlight-site__rank')).toHaveAttribute('aria-label', '已浏览')
  scene.assertQuiet()
})

test('Real resizing selects the Spotlight panels and a fresh saved NSFW mode includes adult sites', async ({ revealed }) => {
  const sfw = await revealed.open()
  await sfw.reveal()
  for (const [width, count] of [[390, 1], [700, 2], [1100, 3], [1440, 4]] as const) {
    await sfw.page.setViewportSize({ width, height: 900 })
    await expect(sfw.content.locator('.spotlight-panel:visible')).toHaveCount(count)
  }
  await expect(sfw.card(sfw.data.adult.name)).toHaveCount(0)
  sfw.assertQuiet()
  const nsfw = await revealed.open({ mode: 'nsfw' })
  await nsfw.reveal()
  await expect(nsfw.card(nsfw.data.adult.name)).toBeVisible()
  await expect(nsfw.featured.getByText(nsfw.data.adultFeatured.name, { exact: true })).toBeVisible()
  await nsfw.openSearch()
  await nsfw.searchInput.fill('wolf')
  await expect(nsfw.results.locator('strong')).toHaveText(['Wolf Community', 'Wolf Archive', 'Wolf After Dark'])
  nsfw.assertQuiet()
})
