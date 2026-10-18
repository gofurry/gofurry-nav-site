import { test, expect, groupNames, assertGamesHomeAppearance, assertCardHover, assertStatsAppearance } from '../fixtures/games-home'
import { mkdir, writeFile } from 'node:fs/promises'

test('C1 Home Normal motion keeps card directions and track geometry', async ({ gamesHome, page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  const scene = await gamesHome.open({ dataset: 'news-populated' })
  const group = scene.group(0), shell = group.locator('.game-group-page-shell')
  const height = (await shell.boundingBox())!.height
  for (const direction of ['next', 'prev'] as const) {
    await group.locator('.game-group-pager button').nth(direction === 'next' ? 1 : 0).click()
    const entering = group.locator(`.game-page-${direction}-enter-active`)
    await expect(entering).toHaveCSS('transition-duration', '0.36s, 0.42s')
    await expect(entering).toHaveCSS('transition-property', 'opacity, transform')
    await expect(group.locator('.game-group-page-live')).toHaveCount(1)
    await expect(scene.cards(0)).toHaveCount(direction === 'next' ? 1 : 8)
    expect((await shell.boundingBox())!.height).toBe(height)
  }
  const stats = scene.stats.locator('.game-stats-page-track')
  await expect(stats).toHaveCSS('transition-duration', '0.36s')
  await scene.stats.locator('.stats-type-tab').last().click()
  await scene.stats.locator('.stats-page-tab').nth(1).click()
  await expect(stats).toHaveAttribute('style', /translate3d\(-100%/)
  await scene.settle(stats)
  expect(await stats.evaluate(el => new DOMMatrix(getComputedStyle(el).transform).m41)).toBeCloseTo(-(await stats.boundingBox())!.width, 1)
  const news = scene.news.locator('.news-track')
  await expect(news).toHaveCSS('transition-duration', '0.34s')
  await scene.news.locator('.news-nav-button').last().click()
  await expect(scene.news.locator('.news-pager__count')).toHaveText('2/3')
  await scene.settle(news)
  expect(await news.evaluate(el => new DOMMatrix(getComputedStyle(el).transform).m41)).toBeLessThan(0)
  scene.assertQuiet()
})

for (const config of [{ width: 1440, theme: 'light', locale: 'zh' }, { width: 390, theme: 'dark', locale: 'en' }] as const) {
  test(`C1 Home initial Reduced, live interruption and restoration (${config.width})`, async ({ gamesHome, page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' })
    const scene = await gamesHome.open({ ...config, dataset: 'news-populated' })
    const group = scene.group(0), next = group.locator('.game-group-pager button').last()
    const firstNames = await scene.cards(0).locator('.game-card__title').allTextContents()
    const shell = group.locator('.game-group-page-shell'), height = (await shell.boundingBox())!.height
    await next.focus(); await page.keyboard.press('Enter')
    await expect(group.locator('.game-group-page-live')).toHaveCount(1)
    await expect(scene.cards(0)).toHaveCount(1)
    await expect(next).toBeFocused()
    await expect(group.locator('.game-group-page-live')).toHaveCSS('opacity', '1')
    // Fast wrap still follows the latest real keyboard action.
    await page.keyboard.press('Enter'); await page.keyboard.press('Enter'); await page.keyboard.press('Enter')
    await expect(group.locator('.game-group-pager > span')).toHaveText('1/2')
    await expect(scene.cards(0).locator('.game-card__title')).toHaveText(firstNames)
    expect((await shell.boundingBox())!.height).toBe(height)

    await page.exposeFunction('c1ReduceHome', () => page.emulateMedia({ reducedMotion: 'reduce' }))
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await group.evaluate(root => {
      const onRun = (event: Event) => {
        const target = event.target as HTMLElement
        if (!target.matches('.game-page-next-enter-active')) return
        root.removeEventListener('transitionrun', onRun)
        const old = root.querySelector<HTMLElement>('.game-page-next-leave-active')!
        void (window as unknown as { c1ReduceHome(): Promise<void> }).c1ReduceHome().then(() => {
          old.querySelector<HTMLButtonElement>('button')!.focus()
          const box = old.getBoundingClientRect()
          root.setAttribute('data-c1-leave', JSON.stringify({ connected: old.isConnected,
            visibility: getComputedStyle(old).visibility, focused: old.contains(document.activeElement),
            hit: old.contains(document.elementFromPoint(box.x + 10, box.y + 10)) }))
        })
      }
      root.addEventListener('transitionrun', onRun)
    })
    await next.click()
    await expect(group).toHaveAttribute('data-c1-leave', /"focused":false,"hit":false/)
    const leave = JSON.parse((await group.getAttribute('data-c1-leave'))!)
    expect(!leave.connected || leave.visibility === 'hidden').toBe(true)
    await expect(scene.cards(0).filter({ visible: true })).toHaveCount(1)
    await expect(group.locator('.game-group-page-live')).toHaveCount(1)
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await group.locator('.game-group-pager button').first().click()
    await expect(group.locator('.game-page-prev-enter-active')).toHaveCSS('transition-duration', '0.36s, 0.42s')
    await expect(scene.cards(0)).toHaveCount(8)
    await scene.settle(group)
    if (process.env.GOFURRY_C1_REVIEW && config.width === 1440) {
      await mkdir('.cache/139-c1-review', { recursive: true })
      await writeFile('.cache/139-c1-review/home-leave-state.json', JSON.stringify(leave))
      await page.screenshot({ path: '.cache/139-c1-review/home-normal.png', clip: await scene.closureClip(group) })
    }

    await page.emulateMedia({ reducedMotion: 'reduce' })
    await scene.stats.locator('.stats-type-tab').last().click()
    const stats = scene.stats.locator('.game-stats-page-track')
    const news = scene.news.locator('.news-track')
    for (const owner of ['stats', 'news'] as const) {
      const track = owner === 'stats' ? stats : news
      const forward = owner === 'stats' ? scene.stats.locator('.stats-page-tab').nth(1) : scene.news.locator('.news-nav-button').last()
      const backward = owner === 'stats' ? scene.stats.locator('.stats-page-tab').first() : scene.news.locator('.news-nav-button').first()
      await expect(track).toHaveCSS('transition-duration', '0s')
      await forward.click()
      const target = await track.evaluate(el => el instanceof HTMLElement ? el.style.transform : '')
      expect(target).not.toMatch(/translate3d\(0(?:%|px),/)
      const atTarget = () => track.evaluate(el => {
        const node = el as HTMLElement
        const offset = Number.parseFloat(node.style.transform.match(/translate3d\(([-\d.]+)/)![1]!)
        const expected = node.style.transform.includes('%') ? offset / 100 * node.offsetWidth : offset
        return Math.abs(new DOMMatrix(getComputedStyle(node).transform).m41 - expected) < 1
      })
      expect(await atTarget()).toBe(true)
      await page.emulateMedia({ reducedMotion: 'no-preference' })
      await track.evaluate(el => {
        el.addEventListener('transitionrun', () => {
          void (window as unknown as { c1ReduceHome(): Promise<void> }).c1ReduceHome().then(() => el.setAttribute('data-c1-interrupted', 'true'))
        }, { once: true })
      })
      await backward.click()
      await expect(track).toHaveAttribute('data-c1-interrupted', 'true')
      await expect(track).toHaveCSS('transition-duration', '0s')
      expect(await atTarget()).toBe(true)
      await expect(track).toHaveAttribute('style', /translate3d\(0(?:%|px),/)
      await page.emulateMedia({ reducedMotion: 'no-preference' })
      await forward.click()
      await expect(track).toHaveCSS('transition-duration', owner === 'stats' ? '0.36s' : '0.34s')
      await scene.settle(track)
      expect(await atTarget()).toBe(true)
      await page.emulateMedia({ reducedMotion: 'reduce' })
    }
    await expect(scene.stats.locator('.stats-page-tab--active')).toHaveText('2')
    await expect(scene.stats.locator('.game-stats-page-slide[aria-hidden="false"]')).toHaveCount(1)
    await expect(scene.news.locator('.news-pager__count')).toHaveText('2/3')
    expect(await scene.news.locator('.news-progress-fill').evaluate(el => Number.parseFloat((el as HTMLElement).style.width))).toBeCloseTo(200 / 3, 4)
    if (process.env.GOFURRY_C1_REVIEW && config.width === 390) {
      await mkdir('.cache/139-c1-review', { recursive: true })
      await page.screenshot({ path: '.cache/139-c1-review/news-reduced-mobile.png', clip: await scene.closureClip(scene.news) })
    }
    scene.assertQuiet()
  })
}

for (const width of [390, 1024, 1440]) {
  test(`Collection shortcuts follow Sidebar only at ${width}`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ width, mode: 'nsfw', showcase: 'two-managed' })
    const shortcuts = scene.page.locator('.game-collection-shortcuts')
    if (width < 1280) {
      await expect(scene.sidebar).toBeHidden()
      await expect(shortcuts).toHaveCount(0)
    } else {
      await expect(scene.sidebar.locator('.game-collection-shortcuts')).toBeVisible()
      await expect(shortcuts.getByRole('link')).toHaveText(['森林故事 1', '森林故事 2', '森林故事 3', '森林故事 4', '森林故事 5', '全部分区'])
      expect(await shortcuts.evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length)).toBe(3)
    }
    await scene.page.evaluate(() => window.dispatchEvent(new CustomEvent('mode-change', { detail: { mode: 'sfw' } })))
    await scene.settle(scene.root)
    scene.assertQuiet()
  })
}
test('Sparse slots are compact, with no placeholder buttons', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ collections: 'sparse' })
  await expect(scene.sidebar.locator('.game-collection-shortcut')).toHaveText(['森林故事 1', '森林故事 3', '全部分区'])
  scene.assertQuiet()
})
test('Showcase and Collections recover independently in the same page', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ collections: 'slow', showcase: 'showcase-transient' })
  await expect(scene.showcase).toBeVisible()
  await expect(scene.page.locator('.game-collection-shortcuts')).toHaveCount(0)
  scene.releaseCollections()
  await expect(scene.sidebar.locator('.game-collection-shortcut')).toHaveCount(6)
  scene.assertQuiet()
})
for (const collections of ['empty' , 'failure', 'transient', 'slow'] as const) {
  test(`Optional Collections ${collections} recovers once only when unavailable`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ collections, dataset: 'news-populated' })
    await expect(scene.group(0)).toBeVisible()
    await expect(scene.stats).toBeVisible()
    await expect(scene.sidebar).toBeVisible()
    await expect(scene.news).toBeAttached()
    if (collections === 'slow') {
      expect(scene.rendered).not.toContain('game-collection-shortcuts')
      await expect(scene.page.locator('.game-collection-shortcuts')).toHaveCount(0)
      scene.releaseCollections()
    }
    await expect(scene.page.locator('.game-collection-shortcuts a')).toHaveCount(collections === 'empty' ? 1 : collections === 'failure' ? 0 : 6)
    await expect(scene.page.locator('[role="alert"]')).toHaveCount(0)
    scene.assertQuiet()
  })
}
for (const theme of ['light', 'dark'] as const) {
  test(`Sidebar shortcuts inherit Daily Game material (${theme})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ theme })
    const daily = scene.sidebar.getByRole('button', { name: '每日一游', exact: true })
    const shortcut = scene.sidebar.locator('.game-collection-shortcut').first()
    for (const property of ['min-height', 'background-color', 'border-color', 'border-radius', 'box-shadow', 'color']) {
      await expect(shortcut).toHaveCSS(property, await daily.evaluate((el, prop) => getComputedStyle(el).getPropertyValue(prop), property))
    }
    await expect(shortcut).toHaveCSS('min-height', '39.2px')
    await daily.hover(); await scene.settle(daily)
    const hover = await daily.evaluate(el => getComputedStyle(el).backgroundColor)
    await shortcut.hover(); await expect(shortcut).toHaveCSS('background-color', hover)
    await expect(shortcut).toHaveCSS('border-color', 'rgba(0, 0, 0, 0)')
    await scene.page.keyboard.press('Tab')
    await daily.focus()
    const focus = await daily.evaluate(el => {
      const css = getComputedStyle(el)
      return { outline: css.outline, offset: css.outlineOffset }
    })
    await shortcut.focus()
    await expect(shortcut).toHaveCSS('outline', focus.outline)
    await expect(shortcut).toHaveCSS('outline-offset', focus.offset)
    await scene.page.setViewportSize({ width: 1024, height: 900 })
    await expect(scene.page.locator('.game-collection-shortcuts')).toHaveCount(0)
    scene.assertQuiet()
  })
}

for (const theme of ['light', 'dark'] as const) {
  test(`Games Home SSR, local pagination, statistics and consumers (${theme})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ theme })
    const { page, stats, sidebar, dock } = scene
    await expect(page.locator('.gf-nav')).toHaveCSS('position', 'fixed')
    await expect(page.locator('.gf-nav__links [aria-current="page"]')).toHaveAttribute('href', '/games')
    expect(await scene.group(0).evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(88)
    await assertGamesHomeAppearance(scene)
    await assertCardHover(scene)
    await expect(sidebar).toBeVisible()
    await expect(dock).toBeVisible()
    await expect(dock.getByRole('button')).toHaveAttribute('aria-label', '去抽奖')
    await expect(dock.getByRole('link')).toHaveAttribute('href', 'https://github.com/gofurry/gofurry-nav-site/issues')
    const lists = [scene.data.panel.latest_games, scene.data.panel.updated_games, scene.data.panel.free_games, scene.data.panel.popular_games!]
    for (let index = 0; index < 4; index++) {
      const group = scene.group(index), name = groupNames[index]!
      await expect(scene.cards(index).locator('.game-card__title')).toHaveText(lists[index]!.slice(0, 8).map(item => item.name))
      await group.getByRole('button', { name: `${name} next page`, exact: true }).click()
      await expect(group.locator('.game-group-pager > span')).toHaveText('2/2')
      await expect(scene.cards(index)).toHaveCount(1)
      await expect(scene.cards(index).locator('.game-card__title')).toHaveText(lists[index]![8]!.name)
      await group.getByRole('button', { name: `${name} next page`, exact: true }).click()
      await expect(group.locator('.game-group-pager > span')).toHaveText('1/2')
      await expect(scene.cards(index)).toHaveCount(8)
      await group.getByRole('button', { name: `${name} previous page`, exact: true }).click()
      await expect(scene.cards(index)).toHaveCount(1)
      await group.getByRole('button', { name: `${name} previous page`, exact: true }).click()
      await expect(scene.cards(index)).toHaveCount(8)
    }
    await expect(scene.cards(0).first().locator('.game-card__release-badge')).toContainText('2026年9月1日')
    await expect(scene.cards(1).first().locator('.game-card__release-badge')).toContainText('2027年2月14日')
    const count = stats.locator('.stats-type-tab').filter({ hasText: '在线人数' })
    const price = stats.locator('.stats-type-tab').filter({ hasText: '价格' })
    await expect(stats.locator('.game-stats-row:not(.game-stats-row--placeholder)')).toHaveCount(16)
    await price.click()
    await expect(price).toHaveClass(/stats-type-tab--active/)
    await expect(count).toHaveClass(/stats-type-tab--idle/)
    await page.mouse.move(1, 1)
    await assertStatsAppearance(scene)
    await expect(stats.locator('.stats-page-tab')).toHaveCount(3)
    const active = stats.locator('.game-stats-page-slide[aria-hidden="false"]')
    await expect(active.locator('.game-stats-feature__title')).toHaveText([lists[0]![0]!.name, lists[0]![1]!.name])
    await expect(active.locator('.game-stats-feature__value').first()).toHaveText('¥48.00')
    await expect(active.locator('.game-stats-row__discount--deal').first()).toHaveText('-50%')
    await stats.locator('.stats-page-tab').nth(1).click()
    await expect(stats.locator('.stats-page-tab--active')).toHaveText('2')
    await count.click()
    await expect(count).toHaveClass(/stats-type-tab--active/)
    await expect(stats.locator('.stats-page-tab--active')).toHaveText('1')
    await expect(stats.locator('.game-stats-row:not(.game-stats-row--placeholder)')).toHaveCount(16)
    const card = scene.cards(0).first()
    await card.hover()
    await card.getByRole('button', { name: `${lists[0]![0]!.name} review` }).click()
    const dialog = page.locator('.review-dialog')
    await expect(dialog).toBeVisible()
    await expect(dialog).toContainText(lists[0]![0]!.name)
    await expect(dialog.locator('textarea')).toBeVisible()
    await dialog.locator('.review-dialog__close').click()
    await expect(dialog).toHaveCount(0)
    await expect(sidebar.locator('.sidebar-expand-button')).toBeVisible()
    await sidebar.locator('.sidebar-expand-button').click()
    await expect(sidebar.getByRole('link', { name: /Patreon/ })).toBeVisible()
    scene.assertQuiet()
  })

  test(`Games Home mobile layout and pagination (${theme})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ theme, width: 390 })
    await assertGamesHomeAppearance(scene)
    await expect(scene.sidebar).toBeHidden()
    await expect(scene.dock).toBeHidden()
    await expect(scene.cards(0)).toHaveCount(8)
    await scene.group(0).getByRole('button', { name: '最近发售 next page', exact: true }).click()
    await expect(scene.cards(0)).toHaveCount(1)
    await expect(scene.cards(0).locator('.game-card__title')).toHaveText('最近发售 9')
    await scene.group(0).getByRole('button', { name: '最近发售 previous page', exact: true }).click()
    await expect(scene.cards(0)).toHaveCount(8)
    await scene.settle(scene.group(0))
    const shell = await scene.group(0).locator('.game-group-page-shell').boundingBox()
    const last = await scene.cards(0).last().boundingBox()
    expect(last!.y + last!.height).toBeLessThanOrEqual(shell!.y + shell!.height)
    expect(await scene.page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    scene.assertQuiet()
  })
}
