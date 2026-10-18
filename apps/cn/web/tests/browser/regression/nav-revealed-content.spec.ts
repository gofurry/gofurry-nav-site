import { test, expect, quote, weatherURL } from '../fixtures/nav-revealed-content'
import { mkdir } from 'node:fs/promises'

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

test('Reduced Reveal and Spotlight settle the selected target through live preference changes', async ({ revealed }) => {
  const scene = await revealed.open({ theme: 'dark' })
  const { page, featured } = scene
  // Change after hydration: Reveal must read the current setting at the scroll.
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await scene.shell.evaluate(element => {
    window.addEventListener('wheel', event => element.setAttribute('data-observed-wheel-cancelled', String(event.defaultPrevented)), { once: true })
    const scroll = element.scrollIntoView.bind(element)
    element.scrollIntoView = options => {
      element.setAttribute('data-observed-scroll', typeof options === 'object' ? options.behavior || 'auto' : 'auto')
      scroll(options)
    }
  })
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await scene.reveal()
  await expect(scene.shell).toHaveAttribute('data-observed-scroll', 'instant')
  await expect(scene.shell).toHaveAttribute('data-observed-wheel-cancelled', 'true')
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

test('Normal motion baseline retains smooth Reveal and its final content position', async ({ revealed }) => {
  const scene = await revealed.open()
  await scene.page.emulateMedia({ reducedMotion: 'no-preference' })
  await scene.shell.evaluate(element => {
    window.addEventListener('wheel', event => element.setAttribute('data-observed-wheel-cancelled', String(event.defaultPrevented)), { once: true })
    const scroll = element.scrollIntoView.bind(element)
    element.scrollIntoView = options => {
      element.setAttribute('data-observed-scroll', typeof options === 'object' ? options.behavior || 'auto' : 'auto')
      scroll(options)
    }
  })
  await scene.reveal()
  await expect(scene.shell).toHaveAttribute('data-observed-scroll', 'smooth')
  await expect(scene.shell).toHaveAttribute('data-observed-wheel-cancelled', 'false')
  scene.assertQuiet()
})

test('Core cards, TransitionBar and real group/site popovers retain their runtime contract', async ({ revealed }) => {
  const scene = await revealed.open({ width: 1000 })
  const { page, bar, quoteTrigger, author, groupPopover, sitePopover } = scene
  await scene.reveal()
  await scene.alignContent()
  await expect(scene.dock).toBeHidden()
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

test('ToolDock caches its real directory, filters SFW search and opens a result without a view POST', async ({ revealed }) => {
  const scene = await revealed.open()
  await scene.reveal()
  await expect(scene.dock).toBeVisible()
  await scene.openSearch()
  await scene.searchInput.fill('wolf')
  await expect(scene.results.locator('strong')).toHaveText(['Wolf Community', 'Wolf Archive'])
  await expect(scene.searchPanel.getByText('Wolf After Dark', { exact: true })).toHaveCount(0)
  await scene.searchInput.fill('zz-no-result')
  await expect(scene.searchPanel.locator('.nav-tool-empty')).toHaveText('没有找到匹配站点')
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
