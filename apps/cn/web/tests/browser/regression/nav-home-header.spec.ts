import { test, expect, customSeed, searchPopupURL } from '../fixtures/nav-home-header'
import { mkdir, writeFile } from 'node:fs/promises'
import type { Locator, Page } from '@playwright/test'

async function pointerSort(page: Page, modal: Locator, from: number, to: number) {
  const grip = modal.locator('.quick-modal-grip').nth(from)
  const start = (await grip.boundingBox())!, target = (await modal.locator('.quick-modal-item').nth(to).boundingBox())!
  await page.mouse.move(start.x + start.width / 2, start.y + start.height / 2)
  await page.mouse.down()
  await page.mouse.move(start.x + start.width / 2, target.y + target.height / 2 + Math.sign(to - from) * 6, { steps: 8 })
  await expect(modal.locator('.quick-modal-item-dragging')).toHaveCount(1)
  await expect.poll(async () => {
    const shift = await modal.locator('.quick-modal-item').nth(to).evaluate(el => {
      const transform = getComputedStyle(el).transform
      return transform === 'none' ? 0 : new DOMMatrixReadOnly(transform).m42
    })
    return shift * Math.sign(to - from)
  }).toBeLessThan(-1)
}

for (const reducedMotion of ['no-preference', 'reduce'] as const) {
  test(`145-C1.1 handle sorting previews, commits once and cancels safely ${reducedMotion}`, async ({ page, header }) => {
    await page.emulateMedia({ reducedMotion })
    await header.open(900)
    const manage = header.quick.getByRole('button', { name: '管理快捷站点', exact: true })
    await manage.click()
    const { modal } = header
    await expect(modal).toHaveAttribute('role', 'dialog')
    await expect(modal.locator('#custom-site-name')).toBeFocused()
    await expect(modal.locator('[draggable]')).toHaveCount(0)
    await expect(modal.locator('.quick-modal-close svg, .quick-modal-delete svg')).toHaveCount(3)
    await page.evaluate(() => {
      const save = Storage.prototype.setItem
      document.documentElement.dataset.customWrites = '0'
      Storage.prototype.setItem = function (key, value) {
        if (key === 'navCustomSites') document.documentElement.dataset.customWrites = String(Number(document.documentElement.dataset.customWrites) + 1)
        return save.call(this, key, value)
      }
      document.addEventListener('pointerdown', event => { document.documentElement.dataset.sortPointer = String(event.pointerId) })
    })
    const writes = page.locator('html')
    const saved = () => page.evaluate(() => JSON.parse(localStorage.getItem('navCustomSites')!))
    const recent = await page.evaluate(() => localStorage.getItem('recentSites'))
    await pointerSort(page, modal, 0, 1)
    expect(await saved()).toEqual(customSeed)
    await expect(writes).toHaveAttribute('data-custom-writes', '0')
    const moving = modal.locator('.quick-modal-item').nth(1)
    await expect(moving).toHaveCSS('transition-duration', reducedMotion === 'reduce' ? '0s' : '0.2s')
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await expect(moving).toHaveCSS('transition-duration', '0s')
    await page.mouse.up()
    await expect(writes).toHaveAttribute('data-custom-writes', '1')
    expect(await saved()).toEqual([...customSeed].reverse())
    await expect(modal.locator('.quick-modal-item-name')).toHaveText(['Wiki', 'Artwork'])
    const artwork = modal.locator('.quick-modal-grip').filter({ has: page.locator('svg') }).nth(1)
    await artwork.focus(); await artwork.press('ArrowUp')
    await expect(modal.locator('.quick-modal-grip').first()).toBeFocused()
    await expect(modal.getByRole('status')).toContainText('第 1 项')
    await expect(writes).toHaveAttribute('data-custom-writes', '2')
    expect(await saved()).toEqual(customSeed)
    for (const cancel of ['Escape', 'pointercancel', 'blur'] as const) {
      await pointerSort(page, modal, 0, 1)
      if (cancel === 'Escape') await page.keyboard.press('Escape')
      else if (cancel === 'pointercancel') await modal.locator('.quick-modal-grip').first().dispatchEvent('pointercancel', { pointerId: Number(await writes.getAttribute('data-sort-pointer')), isPrimary: true })
      else await page.evaluate(() => window.dispatchEvent(new Event('blur')))
      await page.mouse.up()
      await expect(modal.locator('.quick-modal-item-dragging')).toHaveCount(0)
      await expect(modal.locator('.quick-modal-item').first()).toHaveCSS('transform', 'none')
      expect(await saved()).toEqual(customSeed)
      await expect(writes).toHaveAttribute('data-custom-writes', '2')
    }
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await pointerSort(page, modal, 0, 1)
    await expect(moving).toHaveCSS('transition-duration', '0.2s')
    await page.keyboard.press('Escape'); await page.mouse.up()
    await modal.locator('.quick-modal-delete').last().focus()
    await page.keyboard.press('Tab'); await expect(modal.locator('.quick-modal-close')).toBeFocused()
    await page.keyboard.press('Shift+Tab'); await expect(modal.locator('.quick-modal-delete').last()).toBeFocused()
    await page.keyboard.press('Escape'); await expect(modal).toHaveCount(0); await expect(manage).toBeFocused()
    await manage.click(); await expect(modal.locator('.quick-modal-item-name')).toHaveText(['Artwork', 'Wiki'])
    await page.locator('.quick-modal-backdrop').click({ position: { x: 3, y: 400 } })
    await expect(modal).toHaveCount(0); await expect(manage).toBeFocused()
    expect(await page.evaluate(() => localStorage.getItem('recentSites'))).toBe(recent)
    header.assertQuiet()
  })
}

test('145-C1.1 seven shortcuts keep handle-only drag and bounded scrolling', async ({ page, header }) => {
  await header.open()
  await header.quick.getByRole('button', { name: '管理快捷站点', exact: true }).click()
  const { modal } = header
  for (let i = 1; i <= 5; i++) {
    const url = `https://example.com/${i}`
    header.allowExampleIcon(url)
    await modal.locator('#custom-site-name').fill(`Example ${i}`)
    await modal.locator('#custom-site-url').fill(url)
    await modal.locator('button[type="submit"]').click()
  }
  await expect(modal.locator('.quick-modal-item')).toHaveCount(7)
  await expect(modal.locator('button[type="submit"]')).toBeDisabled()
  await page.setViewportSize({ width: 1440, height: 600 })
  const row = modal.locator('.quick-modal-item').first(), text = row.locator('.quick-modal-item-name')
  const box = (await text.boundingBox())!
  await page.mouse.move(box.x + 10, box.y + 5); await page.mouse.down()
  await page.mouse.move(box.x + 10, box.y + 75); await page.mouse.up()
  await expect(modal.locator('.quick-modal-item-dragging')).toHaveCount(0)
  await expect(modal.locator('.quick-modal-item-name').first()).toHaveText('Artwork')
  await pointerSort(page, modal, 0, 2)
  if (process.env.GOFURRY_145_C1_1_REVIEW) {
    await mkdir('../../../.cache/145-c1-1-review/final', { recursive: true })
    await modal.screenshot({ path: '../../../.cache/145-c1-1-review/final/quick-drag-1440.png' })
  }
  await page.mouse.up()
  await expect(modal.locator('.quick-modal-item-name').nth(2)).toHaveText('Artwork')
  const last = modal.locator('.quick-modal-grip').last()
  await last.focus(); await last.press('ArrowUp')
  await expect(modal.locator('.quick-modal-item-name').nth(5)).toHaveText('Example 5')
  await expect(modal.locator('.quick-modal-grip').nth(5)).toBeFocused()
  await modal.locator('.quick-modal-grip').first().focus()
  await modal.locator('.quick-modal-list').evaluate(el => { el.scrollTop = 0 })
  for (let index = 1; index < 7; index++) {
    await page.keyboard.press('ArrowDown')
    await expect(modal.locator('.quick-modal-grip').nth(index)).toBeFocused()
    expect(await modal.locator('.quick-modal-grip').nth(index).evaluate(el => {
      const row = el.closest('.quick-modal-item')!.getBoundingClientRect()
      const list = el.closest('.quick-modal-list')!.getBoundingClientRect()
      return row.top >= list.top - 1 && row.bottom <= list.bottom + 1
    })).toBe(true)
  }
  if (process.env.GOFURRY_145_C1_1_REVIEW) {
    await modal.screenshot({ path: '../../../.cache/145-c1-1-review/final/quick-keyboard-1440.png' })
  }
  const names = await modal.locator('.quick-modal-item-name').allTextContents()
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('navCustomSites')!).map((item: { name: string }) => item.name))).toEqual(names)
  await modal.locator('.quick-modal-close').click(); await expect(modal).toHaveCount(0)
  header.assertQuiet()
})

for (const [width, height] of [[1440, 900], [1440, 600], [1440, 480], [900, 700], [390, 844]] as const) {
  test(`145-C1 Hero group and hint clearance ${width}x${height}`, async ({ page, header }) => {
    await header.open(width)
    await page.setViewportSize({ width, height })
    await header.settle(header.search)
    const hint = page.locator('.nav-header__scroll-hint')
    const search = page.locator('.nav-header__search'), quick = page.locator('.nav-header__quick-access')
    if (width >= 768) {
      expect((await search.boundingBox())!.y).toBe(width >= 1024 ? height <= 520 ? 88 : 112 : 100)
      expect((await quick.boundingBox())!.y).toBe(width >= 1024 ? height <= 520 ? 268 : 292 : 272)
      const fits = await hint.evaluate(el => el.getBoundingClientRect().top >= Math.max(
        document.querySelector('.nav-header__search')!.getBoundingClientRect().bottom,
        document.querySelector('.nav-header__quick-access')!.getBoundingClientRect().bottom) + 24)
      await expect(hint).toHaveCSS('visibility', fits ? 'visible' : 'hidden')
      if (height === 900) expect(fits).toBe(true)
      if (height === 480) expect(fits).toBe(false)
      expect((await quick.boundingBox())!.y + (await quick.boundingBox())!.height).toBeLessThanOrEqual(height)
    } else {
      await expect(hint).toBeHidden()
      await expect(page.locator('.gf-nav__utilities')).toBeHidden()
      expect((await header.root.boundingBox())!.height).toBe(360)
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    if (process.env.GOFURRY_145_C1_REVIEW === '1') {
      await mkdir('../../../.cache/145-c1-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/145-c1-review/home-${width}x${height}-light.png` })
      await page.locator('.gf-nav').getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
      await header.settle(header.search)
      await page.screenshot({ path: `../../../.cache/145-c1-review/home-${width}x${height}-dark.png` })
    }
    header.assertQuiet()
  })
}

test('B2 Quick Modal preserves the Normal 180/220ms contract', async ({ page, header }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await header.open()
  await page.evaluate(() => {
    new MutationObserver(() => {
      const root = document.querySelector('.quick-modal-enter-active')
      if (!root) return
      const panel = root.querySelector('.quick-modal-panel')!
      document.documentElement.dataset.quickEntry = JSON.stringify({ backdrop: getComputedStyle(root).transitionDuration,
        panel: getComputedStyle(panel).transitionDuration, transform: getComputedStyle(panel).transform })
    }).observe(document.body, { subtree: true, childList: true })
  })
  await header.quick.getByRole('button', { name: '管理快捷站点', exact: true }).click()
  await expect(header.modal).toHaveCSS('opacity', '1')
  expect(await page.locator('.gf-nav__actions').evaluate(el => {
    const box = el.getBoundingClientRect()
    return !!document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2)?.closest('.quick-modal-backdrop')
  })).toBe(true)
  expect(await page.evaluate(() => JSON.parse(document.documentElement.dataset.quickEntry!)))
    .toEqual({ backdrop: '0.18s', panel: '0.22s, 0.18s', transform: 'matrix(1, 0, 0, 1, 0, 4)' })
  await header.modal.locator('.quick-modal-close').click()
  await expect(header.modal).toHaveCount(0)
  header.assertQuiet()
})

test('B2 Quick Modal live Reduced exit and rapid reopen leave no interactive ghost', async ({ page, header }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await header.open()
  await page.exposeFunction('reduceQuickMotion', async () => { await page.emulateMedia({ reducedMotion: 'reduce' }) })
  await page.evaluate(() => {
    document.addEventListener('transitionrun', event => {
      const root = event.target
      if (!(root instanceof HTMLElement) || !root.matches('.quick-modal-backdrop')) return
      const phase = root.classList.contains('quick-modal-leave-active') ? 'leave' : 'enter'
      if (document.documentElement.dataset.quickRestored) {
        if (phase === 'enter') document.documentElement.dataset.quickRestored = getComputedStyle(root).transitionDuration
        return
      }
      if (root.dataset[phase + 'Reduced']) return
      root.dataset[phase + 'Reduced'] = 'true'
      const bridge = window as Window & { reduceQuickMotion?: () => Promise<void> }
      void bridge.reduceQuickMotion!().then(() => {
        if (phase === 'leave') {
          root.querySelector<HTMLInputElement>('input')!.focus()
          document.documentElement.dataset.quickExit = JSON.stringify({ connected: root.isConnected,
            visibility: getComputedStyle(root).visibility, focused: root.contains(document.activeElement),
            hit: root.contains(document.elementFromPoint(3, 400)) })
        }
        document.documentElement.dataset.quickReduced = phase
      })
    })
  })
  const manage = header.quick.getByRole('button', { name: '管理快捷站点', exact: true })
  const root = page.locator('.quick-modal-backdrop')
  await manage.click()
  await expect(page.locator('html')).toHaveAttribute('data-quick-reduced', 'enter')
  await expect(root).toHaveCSS('opacity', '1'); await expect(header.modal).toHaveCSS('transform', 'none')
  await header.modal.locator('#custom-site-name').fill('discarded draft')
  await expect(root).not.toHaveClass(/quick-modal-enter-active/)
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await expect(root).toHaveCSS('opacity', '1')
  await header.modal.locator('.quick-modal-close').click()
  await expect(page.locator('html')).toHaveAttribute('data-quick-reduced', 'leave')
  const exit = await page.evaluate(() => JSON.parse(document.documentElement.dataset.quickExit!))
  expect(exit).toMatchObject({ focused: false, hit: false })
  if (exit.connected) expect(exit.visibility).toBe('hidden')
  if (process.env.GOFURRY_B2_REVIEW === '1') {
    await mkdir('.cache/139-b2-review', { recursive: true })
    await writeFile('.cache/139-b2-review/quick-leave-state.json', JSON.stringify(exit))
  }
  await manage.click(); await expect(root).toHaveCount(1)
  await expect(header.modal.locator('#custom-site-name')).toHaveValue('')
  await expect(header.modal).toHaveCSS('opacity', '1')
  await expect(header.modal).toHaveCSS('transform', 'none')
  if (process.env.GOFURRY_B2_REVIEW === '1') {
    await mkdir('.cache/139-b2-review', { recursive: true })
    await header.modal.screenshot({ path: '.cache/139-b2-review/quick-modal.png' })
  }
  await root.click({ position: { x: 3, y: 400 } }); await expect(root).toHaveCount(0)
  await manage.click(); await header.modal.locator('.quick-modal-close').click(); await expect(root).toHaveCount(0)
  await page.evaluate(() => { document.documentElement.dataset.quickRestored = 'waiting' })
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await manage.click()
  await expect(page.locator('html')).toHaveAttribute('data-quick-restored', '0.18s')
  await header.modal.locator('.quick-modal-close').click(); await expect(root).toHaveCount(0)
  // Real route unload after repeated close must not revive an old overlay.
  await page.mouse.move(20, 400); await page.mouse.wheel(0, 700)
  await page.getByRole('link', { name: '服务条款', exact: true }).click()
  await expect(page).toHaveURL(/\/terms$/)
  await page.emulateMedia({ reducedMotion: 'reduce' }); await expect(root).toHaveCount(0)
  header.assertQuiet()
})

test('Suggestions own one 220ms entry, instant exit and live Reduced loading without replay', async ({ page, header }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await header.open()
  let interruptedOpacity = 1
  await page.exposeFunction('reduceSuggestionsDuringEntry', async (opacity: number) => {
    interruptedOpacity = opacity
    await page.emulateMedia({ reducedMotion: 'reduce' })
  })
  await page.evaluate(() => {
    const entries: object[] = []
    new MutationObserver(() => {
      const list = document.querySelector<HTMLElement>('.search-suggestions-enter-active')
      if (!list || list.dataset.entryObserved) return
      list.dataset.entryObserved = 'true'
      const style = getComputedStyle(list)
      entries.push({ duration: style.transitionDuration, property: style.transitionProperty,
        animation: style.animationName, transform: style.transform, opacity: style.opacity })
      document.documentElement.dataset.suggestionEntries = JSON.stringify(entries)
    }).observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ['class'] })
    // Trigger the real media change at transitionrun; observe, never pause or
    // replace, the browser's animation (which would override CSS cancellation).
    let interrupted = false
    document.addEventListener('transitionrun', event => {
      const target = event.target
      if (target instanceof HTMLElement && target.matches('.search-suggestions-enter-active')) {
        target.dataset.entryInProgress = 'true'
        if (!interrupted) {
          interrupted = true
          const bridge = window as Window & { reduceSuggestionsDuringEntry?: (opacity: number) => Promise<void> }
          void bridge.reduceSuggestionsDuringEntry!(Number(getComputedStyle(target).opacity))
            .then(() => { target.dataset.reducedDuringEntry = 'true' })
        }
      }
    })
  })
  await expect(page.locator('.nav-header__scroll-dot')).toHaveCSS('animation-name', 'navHeaderBounce')
  await expect(page.locator('.nav-header__scroll-dot')).toHaveCSS('animation-duration', '1s')
  header.holdSuggestions()
  await header.input.fill('wolf'); await header.expectSuggestionRequest('wolf')
  await expect(header.suggestions).toHaveAttribute('data-reduced-during-entry', 'true')
  expect(interruptedOpacity).toBeLessThan(1)
  const entries = () => page.evaluate(() => JSON.parse(document.documentElement.dataset.suggestionEntries!))
  expect(await entries()).toEqual([{ duration: '0.22s, 0.22s', property: 'opacity, transform',
    animation: 'none', transform: 'matrix(1, 0, 0, 1, 0, -6)', opacity: '0' }])
  await expect(page.locator('.nav-header__scroll-dot')).toHaveCSS('animation-name', 'none')
  await expect(header.suggestions).toHaveCSS('animation-name', 'none')
  await expect(header.suggestions).toHaveCSS('opacity', '1')
  await expect(header.suggestions).toHaveCSS('transform', 'none')
  const spinner = header.suggestions.locator('.search-suggestion-spinner')
  await expect(spinner).toHaveCSS('animation-name', 'none')
  await expect(spinner).toBeVisible()
  await expect(header.suggestions.locator('.search-suggestion-loading')).toBeVisible()
  await expect(header.input).toBeFocused()
  if (process.env.GOFURRY_MOTION_REVIEW === '1') {
    await mkdir('.cache/139-a-review', { recursive: true })
    await header.suggestions.screenshot({ path: '.cache/139-a-review/nav-static-loading.png' })
  }
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await expect(spinner).toHaveCSS('animation-name', 'navSearchSpin')
  await expect(spinner).toHaveCSS('animation-duration', '1s')
  header.releaseSuggestions()
  await expect(header.items).toHaveCount(3)
  await expect(header.suggestions).toHaveAttribute('data-entry-observed', 'true')
  expect(await entries()).toHaveLength(1)
  if (process.env.GOFURRY_B1_REVIEW === '1') {
    await mkdir('.cache/139-b1-review', { recursive: true })
    await header.suggestions.screenshot({ path: '.cache/139-b1-review/search-ready.png' })
  }
  await header.input.press('Escape')
  expect(await header.suggestions.count()).toBe(0)
  await header.input.fill(''); await header.input.fill('wolf')
  await expect(header.suggestions).toHaveAttribute('data-entry-in-progress', 'true')
  expect(await entries()).toHaveLength(2)
  await page.mouse.click(20, 400)
  expect(await header.suggestions.count()).toBe(0)
  header.assertQuiet()
})

test('Reduced mobile Home has SSR content and a static loading suggestion', async ({ page, header }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await header.open(390)
  await expect(page.locator('.nav-tool-dock')).toHaveCount(0)
  await expect(page.locator('.nav-content-shell')).toBeVisible()
  header.holdSuggestions()
  await header.input.fill('wolf'); await header.expectSuggestionRequest('wolf')
  await expect(header.suggestions.locator('.search-suggestion-spinner')).toHaveCSS('animation-name', 'none')
  await expect(header.suggestions.locator('.search-suggestion-loading')).toBeVisible()
  await expect(header.suggestions).toHaveCSS('opacity', '1')
  await expect(header.suggestions).toHaveCSS('transform', 'none')
  await expect(header.suggestions).toHaveCSS('transition-duration', '0s')
  header.releaseSuggestions(); await expect(header.items).toHaveCount(3)
  header.assertQuiet()
})

test('Search keeps typography, theme parity, debounce, keyboard selection and native scrolling', async ({ page, header }) => {
  await header.open()
  const { input, search, suggestions, items, quick } = header
  const category = search.locator('.search-category-row .search-chip-active')
  const platform = search.locator('.search-platform-row .search-chip-active')
  await expect(category).toHaveText('搜索')
  await expect(platform).toHaveText('必应')
  await expect(quick).toBeVisible()
  await expect(page.locator('.nav-tool-dock')).toHaveCount(0)
  for (const [chip, size, height] of [[category, '14px', '20px'], [platform, '12px', '16px']] as const) {
    await expect(chip).toHaveCSS('font-size', size)
    await expect(chip).toHaveCSS('line-height', height)
    await expect(chip).toHaveCSS('font-weight', '500')
    await expect(chip).toHaveCSS('border-radius', '12px')
    // The actual unlayered nav.less cascade overrides Tailwind transition-all.
    await expect(chip).toHaveCSS('transition-property', 'background, box-shadow, color')
    await expect(chip).toHaveCSS('transition-duration', '0.5s, 0.5s, 0.5s')
  }
  const appearance = () => search.locator('.search-category-row .search-chip-active').evaluate((chip) => {
    const elements = [chip, document.querySelector('.search-input')!, document.querySelector('.quick-site-tile')!]
    return elements.map(element => {
      const style = getComputedStyle(element)
      return Object.fromEntries(['background-color', 'color', 'border-radius', 'box-shadow', 'backdrop-filter',
        'font-size', 'line-height', 'font-weight'].map(property => [property, style.getPropertyValue(property)]))
    })
  })
  const light = await appearance()
  const theme = page.locator('.gf-nav').getByRole('button', { name: '切换明暗主题图标', exact: true })
  await theme.click()
  await expect(page.locator('html')).toHaveClass(/\bdark\b/)
  expect(await page.evaluate(() => localStorage.getItem('theme'))).toBe('dark')
  await header.settle(header.root)
  expect(await appearance()).toEqual(light)
  await theme.click()
  await expect(page.locator('html')).not.toHaveClass(/\bdark\b/)
  await header.settle(header.root)
  expect(await appearance()).toEqual(light)

  header.holdSuggestions()
  await input.focus()
  await input.fill('wolf')
  await header.expectSuggestionRequest('wolf')
  await expect(suggestions).toBeVisible()
  await expect(suggestions.locator('.search-suggestion-loading')).toBeVisible()
  await expect(suggestions.locator('.search-suggestion-spinner')).toBeVisible()
  header.releaseSuggestions()
  await expect(items).toHaveText(['wolf furry', 'wolf art', 'wolf game'])
  await expect(items.first()).toHaveClass(/\bsearch-suggestion-item-active\b/)
  await expect(suggestions.locator('.search-suggestion-header')).toHaveCSS('font-size', '12px')
  await expect(suggestions.locator('.search-suggestion-header')).toHaveCSS('line-height', '16px')
  await expect(items.first()).toHaveCSS('font-size', '14px')
  await expect(items.first()).toHaveCSS('line-height', '20px')
  await expect(items.first()).toHaveCSS('font-weight', '500')
  await input.press('ArrowDown')
  await expect(items.nth(1)).toHaveClass(/\bsearch-suggestion-item-active\b/)
  await expect(items.first()).not.toHaveClass(/\bsearch-suggestion-item-active\b/)
  await input.press('ArrowUp')
  await expect(items.first()).toHaveClass(/\bsearch-suggestion-item-active\b/)
  await input.press('Escape')
  await expect(suggestions).toHaveCount(0)
  await input.fill('noresult')
  await header.expectSuggestionRequest('noresult')
  await expect(suggestions.getByText('暂无搜索建议', { exact: true })).toBeVisible()
  await expect(items).toHaveCount(0)

  // Wheel outside the dropdown (which stops propagation) must reach the real
  // native wheel event while Search is focused. Only observe the event.
  await page.evaluate(() => {
    document.documentElement.dataset.headerWheelObserved = 'false'
    window.addEventListener('wheel', event => {
      document.documentElement.dataset.headerWheelObserved = 'true'
      document.documentElement.dataset.headerWheelCancelled = String(event.defaultPrevented)
    }, { once: true, passive: true })
  })
  await page.mouse.move(20, 400)
  await page.mouse.wheel(0, 180)
  await expect(page.locator('html')).toHaveAttribute('data-header-wheel-observed', 'true')
  await expect(page.locator('html')).toHaveAttribute('data-header-wheel-cancelled', 'false')
  await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))))
  await expect(page.locator('.nav-tool-dock')).toHaveCount(0)
  await input.press('Escape')
  await page.mouse.click(20, 400)
  await expect(input).not.toBeFocused()
  await page.mouse.wheel(0, 300)
  await expect.poll(() => page.evaluate(() => scrollY)).toBe(480)
  await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }))
  await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(0)
  await header.showSuggestions()
  header.allowSearchPopup()
  const popupPromise = page.waitForEvent('popup')
  await input.press('Enter')
  const popup = await popupPromise
  await popup.waitForLoadState('load')
  await expect(popup).toHaveURL(searchPopupURL)
  expect(new URL(popup.url()).searchParams.get('q')).toBe('wolf furry')
  header.assertQuiet()
})

test('Quick Access exposes all slots and persists real Manage add/delete actions', async ({ page, header }) => {
  await header.open()
  const { quick, modal } = header
  await expect(quick.getByText('最近浏览', { exact: true })).toBeVisible()
  await expect(quick.getByText('常用网站', { exact: true })).toBeVisible()
  const sections = quick.locator('.quick-access-section')
  await expect(sections).toHaveCount(2)
  await expect(sections.nth(0).locator('.quick-site-tile')).toHaveCount(8)
  await expect(sections.nth(1).locator('.quick-site-tile')).toHaveCount(8)
  await expect(sections.nth(0).locator('.quick-site-tile-empty')).toHaveCount(6)
  await expect(sections.nth(1).locator('.quick-site-tile-empty')).toHaveCount(5)
  await expect(quick.locator('[title="Community"] img')).toBeVisible()
  await expect(quick.locator('[title="Community"] img')).toHaveAttribute('src', 'https://a.favicon.im/community.example?larger=true&throw-error-on-404=true')
  await expect(quick.locator('[title="Artwork"] img')).toHaveAttribute('src', 'https://a.favicon.im/art.example?larger=true&throw-error-on-404=true')
  await expect(quick.locator('[title="Archive"] .quick-site-fallback')).toHaveText('A')
  await expect(quick.locator('[title="Archive"] img')).toHaveCount(0)
  await quick.getByRole('button', { name: '管理快捷站点', exact: true }).click()
  await expect(page.locator('.quick-modal-backdrop')).toBeVisible()
  await expect(modal).toBeVisible()
  await expect(modal.locator('.quick-modal-item')).toHaveCount(2)
  const submit = modal.locator('button[type="submit"]')
  const name = modal.locator('#custom-site-name')
  const url = modal.locator('#custom-site-url')
  const error = modal.locator('.quick-modal-error')
  await submit.click()
  await expect(error).toHaveText('请输入网站名称')
  await name.fill('Example Site')
  await submit.click()
  await expect(error).toHaveText('请输入网站链接')
  await url.fill('example.com')
  header.allowExampleIcon()
  await submit.click()
  await expect(modal.locator('.quick-modal-item')).toHaveCount(3)
  const saved = await page.evaluate(() => JSON.parse(localStorage.getItem('navCustomSites')!))
  expect(saved.slice(0, 2)).toEqual(customSeed)
  expect(saved[2]).toEqual({ id: expect.stringMatching(/^custom-/), name: 'Example Site', url: 'https://example.com' })
  await expect(quick.locator('[title="Example Site"] img')).toHaveAttribute('src', 'https://a.favicon.im/example.com?larger=true&throw-error-on-404=true')
  await expect(name).toHaveValue('')
  await expect(url).toHaveValue('')
  await expect(error).toHaveCount(0)
  await modal.locator('.quick-modal-item').filter({ hasText: 'Example Site' }).getByRole('button', { name: '删除网站', exact: true }).click()
  await expect(modal.locator('.quick-modal-item')).toHaveCount(2)
  await expect(modal.getByText('Example Site', { exact: true })).toHaveCount(0)
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('navCustomSites')!))).toEqual(customSeed)
  await pointerSort(page, modal, 0, 1)
  await page.mouse.up()
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('navCustomSites')!))).toEqual([...customSeed].reverse())
  header.assertQuiet()
})

test('Suggestion requests debounce for 600ms, abort stale work and use only q', async ({ page, header }) => {
  await header.open()
  await page.clock.install()
  await page.clock.pauseAt(new Date())
  header.holdSuggestions()
  await header.input.fill('wo')
  await page.clock.runFor(400)
  await header.input.fill('wolf')
  await page.clock.runFor(599)
  expect(header.requestQueries()).toEqual([])
  await page.clock.runFor(1)
  await header.expectSuggestionRequest('wolf')
  await expect(header.suggestions.locator('.search-suggestion-loading')).toBeVisible()
  await header.input.fill('noresult')
  await header.expectAborted('wolf')
  await page.clock.runFor(599)
  expect(header.requestQueries()).toEqual(['wolf'])
  await page.clock.runFor(1)
  await header.expectSuggestionRequest('noresult')
  await expect(header.suggestions.getByText('暂无搜索建议', { exact: true })).toBeVisible()
  header.releaseSuggestions()
  await page.clock.runFor(1000)
  await expect(header.items).toHaveCount(0)
  expect(header.requestQueries()).toEqual(['wolf', 'noresult'])
  header.assertQuiet()
})

for (const [platform, destination] of [
  ['谷歌', 'https://www.google.com/search?q=wolf%20furry'],
  ['小红书', 'https://www.xiaohongshu.com/search_result?keyword=wolf%20furry'],
] as const) {
  test(`${platform} destination uses the same suggestion API without changing Enter navigation`, async ({ page, header }) => {
    await header.open()
    await header.search.locator('.search-platform-row').getByText(platform, { exact: true }).click()
    await header.showSuggestions()
    header.allowSearchPopup(destination)
    const popupPromise = page.waitForEvent('popup')
    await header.input.press('Enter')
    const popup = await popupPromise
    await popup.waitForLoadState('load')
    await expect(popup).toHaveURL(destination)
    expect(header.requestQueries()).toEqual(['wolf'])
    header.assertQuiet()
  })
}

for (const failure of ['unavailable', 503, 429, 'network'] as const) {
  test(`Suggestion ${failure} quietly closes the dropdown and preserves search`, async ({ page, header }) => {
    await header.open()
    header.suggestion('furry', failure === 'unavailable' ? { state: 'unavailable' } : { failure })
    await page.clock.install()
    await page.clock.pauseAt(new Date())
    await header.input.fill('furry')
    const finished = page.waitForEvent(failure === 'network' ? 'requestfailed' : 'requestfinished',
      request => request.url().includes('/nav/search/suggestions?q=furry'))
    await page.clock.runFor(600)
    await finished
    await expect(header.suggestions).toHaveCount(0)
    await expect(page.getByRole('alert')).toHaveCount(0)
    await page.clock.runFor(5000)
    expect(header.requestQueries()).toEqual(['furry'])
    const destination = 'https://www.bing.com/search?q=furry'
    header.allowSearchPopup(destination)
    const popupPromise = page.waitForEvent('popup')
    await header.input.press('Enter')
    const popup = await popupPromise
    await popup.waitForLoadState('load')
    await expect(popup).toHaveURL(destination)
    header.assertQuiet()
  })
}

for (const width of [1440, 390]) {
test(`IME owns confirmation and navigation keys; only final Chinese text starts one debounce (${width})`, async ({ page, context, header }) => {
  await header.open(width)
  header.suggestion('兽人', { items: ['兽人 游戏'] })
  await page.clock.install()
  await page.clock.pauseAt(new Date())
  await header.input.focus()
  await header.input.dispatchEvent('compositionstart')
  for (const text of ['shou', '兽']) {
    await header.input.fill(text)
    await page.clock.runFor(1200)
    expect(header.requestQueries()).toEqual([])
  }
  for (const key of ['Enter', 'ArrowDown', 'ArrowUp', 'Escape']) {
    // Both the composition lifecycle ref and the native flag guard before preventDefault.
    for (const isComposing of [false, true]) {
      expect(await header.input.evaluate((input, event) => input.dispatchEvent(new KeyboardEvent('keydown', {
        ...event, bubbles: true, cancelable: true,
      })), { key, isComposing })).toBe(true)
    }
  }
  expect(context.pages()).toHaveLength(1)
  await header.input.fill('兽人')
  await header.input.dispatchEvent('compositionend', { data: '兽人' })
  await page.clock.runFor(300)
  await header.input.dispatchEvent('input', { inputType: 'insertText', data: '兽人' })
  await page.clock.runFor(299)
  expect(header.requestQueries()).toEqual([])
  await page.clock.runFor(1)
  await header.expectSuggestionRequest('兽人')
  await expect(header.items).toHaveText(['兽人 游戏'])
  await page.clock.runFor(1000)
  expect(header.requestQueries()).toEqual(['兽人'])
  // Native composing can still be true after compositionend in some browsers.
  expect(await header.input.evaluate(input => input.dispatchEvent(new KeyboardEvent('keydown', {
    key: 'Enter', isComposing: true, bubbles: true, cancelable: true,
  })))).toBe(true)
  expect(context.pages()).toHaveLength(1)
  await header.input.press('Escape')
  await expect(header.suggestions).toHaveCount(0)
  header.assertQuiet()
})
}

test('Provider HTML remains literal text with safe keyword segments; other categories never fetch', async ({ page, header }) => {
  await header.open()
  const malicious = '<img src=x onerror="document.documentElement.dataset.suggestionXss=1">wolf<script>alert(1)</script>'
  header.suggestion('wolf', { items: [malicious] })
  await page.clock.install()
  await page.clock.pauseAt(new Date())
  await header.input.fill('wolf')
  await page.clock.runFor(600)
  await header.expectSuggestionRequest('wolf')
  await expect(header.items).toHaveText([malicious])
  await expect(header.items.locator('.search-highlight')).toHaveText('wolf')
  await expect(header.items.locator('img, script')).toHaveCount(0)
  await expect(page.locator('html')).not.toHaveAttribute('data-suggestion-xss')
  // Dispatch a real click without depending on animation frames while the clock is paused.
  await header.search.locator('.search-category-row').getByText('兽人', { exact: true }).dispatchEvent('click')
  await header.input.fill('other category')
  await page.clock.runFor(1200)
  await expect(header.suggestions).toHaveCount(0)
  expect(header.requestQueries()).toEqual(['wolf'])
  header.assertQuiet()
})

for (const width of [375, 390, 430, 640, 767, 768, 1440]) {
  test(`145-A2 centered categories and inline platform slots at ${width}`, async ({ page, context, header }) => {
    await header.open(width)
    const trigger = page.locator('.search-platform-trigger')
    const picker = page.locator('.search-platform-picker')
    const options = picker.locator('button')
    const categories = header.search.locator('.search-category-row button')
    const inputBox = (await header.input.boundingBox())!
    const categoryBox = (await header.search.locator('.search-category-items').boundingBox())!
    expect(Math.abs(categoryBox.x + categoryBox.width / 2 - inputBox.x - inputBox.width / 2)).toBeLessThanOrEqual(1)
    expect(new Set(await categories.evaluateAll(buttons => buttons.map(button => button.getBoundingClientRect().top))).size).toBe(1)
    if (width >= 768) {
      await expect(picker).toBeHidden()
      await expect(header.search.locator('.search-platform-row .search-chip')).toHaveCount(8)
      header.assertQuiet()
      return
    }
    await expect(header.search.locator('.search-platform-row')).toBeHidden()
    await expect(options).toHaveCount(1)
    await expect(trigger).toHaveText('必应')
    await expect(trigger).toHaveAttribute('aria-expanded', 'false')
    const anchor = (await trigger.boundingBox())!
    expect(Math.abs(anchor.x + anchor.width - inputBox.x - inputBox.width)).toBeLessThanOrEqual(1)
    expect(anchor.height).toBeGreaterThanOrEqual(44)
    // The collapsed empty slots must pass pointer input to the Hero, not an invisible grid.
    expect(await picker.evaluate(el => {
      const box = el.getBoundingClientRect()
      return !document.elementFromPoint(box.left + 2, box.top + 20)?.closest('.search-platform-picker')
    })).toBe(true)
    const before = await page.evaluate(() => ({ height: document.documentElement.scrollHeight, y: scrollY, overflow: document.body.style.overflow }))
    await header.input.fill('wolf & 兽人')
    await trigger.click()
    await expect(options).toHaveText(['谷歌', 'DuckDuckGo', 'B站', '必应', '小红书', '知乎', '微博', '推特'])
    await expect(trigger).toHaveAttribute('aria-expanded', 'true')
    const boxes = await options.evaluateAll(nodes => nodes.map(el => { const r = el.getBoundingClientRect(); return { x: r.x, y: r.y, bottom: r.bottom } }))
    expect(new Set(boxes.map(box => box.x)).size).toBe(4)
    expect(new Set(boxes.map(box => box.y)).size).toBe(2)
    expect(boxes[3]!.x).toBe(anchor.x)
    expect(boxes[7]!.bottom).toBeLessThan((await header.root.boundingBox())!.height)
    expect(await page.evaluate(() => ({ height: document.documentElement.scrollHeight, y: scrollY, overflow: document.body.style.overflow }))).toEqual(before)
    await expect(page.locator('dialog, .search-platform-sheet')).toHaveCount(0)
    await options.filter({ hasText: /^小红书$/ }).click()
    await expect(options).toHaveText(['谷歌', 'DuckDuckGo', 'B站', '小红书', '必应', '知乎', '微博', '推特'])
    await expect(trigger).toBeFocused()
    await expect(trigger).toHaveAttribute('aria-expanded', 'true')
    await expect(header.input).toHaveValue('wolf & 兽人')
    expect(context.pages()).toHaveLength(1)
    await trigger.click(); await expect(options).toHaveCount(1)
    expect((await trigger.boundingBox())!.x).toBe(anchor.x)
    await trigger.click()
    await expect(options).toHaveText(['谷歌', 'DuckDuckGo', 'B站', '小红书', '必应', '知乎', '微博', '推特'])
    if (process.env.GOFURRY_145_REVIEW === '1' && [375, 390, 430, 640].includes(width)) {
      await mkdir('../../../.cache/145-a2-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/145-a2-review/picker-${width}.png` })
    }
    await page.keyboard.press('Escape'); await expect(options).toHaveCount(1); await expect(trigger).toBeFocused()
    const destination = 'https://www.xiaohongshu.com/search_result?keyword=wolf%20%26%20%E5%85%BD%E4%BA%BA'
    header.allowSearchPopup(destination)
    const popupPromise = page.waitForEvent('popup')
    await header.search.locator('.search-icon').click()
    const popup = await popupPromise; await popup.waitForLoadState('load'); await expect(popup).toHaveURL(destination)
    expect(header.requestQueries()).toEqual([])
    for (const [category, count, selected] of [['兽人', 5, 'WikiFur'], ['游戏', 3, 'itch.io'], ['艺术', 5, 'Pixiv'], ['小说', 2, '兽人控小说站']] as const) {
      await categories.filter({ hasText: new RegExp(`^${category}$`) }).click()
      await expect(trigger).toHaveAttribute('aria-expanded', 'false')
      await expect(trigger).toHaveText(selected)
      await trigger.click(); await expect(options).toHaveCount(count)
      expect(new Set(await options.allTextContents()).size).toBe(count)
      expect(Math.abs((await trigger.boundingBox())!.x + (await trigger.boundingBox())!.width - inputBox.x - inputBox.width)).toBeLessThanOrEqual(1)
      await expect(trigger).toHaveCSS('grid-column-start', String(Math.min(4, count)))
    }
    await categories.first().click(); await expect(trigger).toHaveText('必应'); await expect(options).toHaveCount(1)
    expect(await page.evaluate(() => scrollY)).toBe(0)
    header.assertQuiet()
  })
}

test('145-A2 inline picker cancels suggestions and preserves live motion, focus and route cleanup', async ({ page, header }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await header.open(390)
  header.holdSuggestions()
  await header.input.fill('wolf'); await header.expectSuggestionRequest('wolf')
  const trigger = page.locator('.search-platform-trigger')
  const picker = page.locator('.search-platform-picker')
  const options = picker.locator('button')
  await page.exposeFunction('reducePickerMotion', () => page.emulateMedia({ reducedMotion: 'reduce' }))
  await page.evaluate(() => {
    document.addEventListener('transitionrun', event => {
      if (!(event.target instanceof Element) || !event.target.matches('.search-platform-option-enter-active') || document.documentElement.dataset.pickerEntry) return
      document.documentElement.dataset.pickerEntry = getComputedStyle(event.target).transitionDuration
      void (window as Window & { reducePickerMotion(): Promise<void> }).reducePickerMotion()
    })
  })
  await trigger.click(); await header.expectAborted('wolf')
  await expect(header.suggestions).toHaveCount(0); header.releaseSuggestions()
  await expect(options).toHaveCount(8)
  await expect(page.locator('html')).toHaveAttribute('data-picker-entry', '0.18s, 0.18s')
  await expect(options.first()).toHaveCSS('opacity', '1'); await expect(options.first()).toHaveCSS('transform', 'none')
  await expect(picker.locator('.search-platform-option-enter-active')).toHaveCount(0)
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await options.first().focus(); await page.keyboard.press('Escape')
  await expect(options).toHaveCount(1); await expect(trigger).toBeFocused()
  for (let index = 0; index < 3; index++) {
    await trigger.click(); await trigger.click(); await expect(options).toHaveCount(1)
  }
  await trigger.click()
  await page.mouse.click(5, 320)
  await expect(options).toHaveCount(1); await expect(header.suggestions).toHaveCount(0)
  expect(header.requestQueries()).toEqual(['wolf'])
  await trigger.click()
  // Input focus closes the grid, keeps the keyword and starts only the existing debounce.
  await header.input.focus(); await expect(options).toHaveCount(1)
  await header.input.press('Escape')
  await trigger.click()
  await page.evaluate(() => history.pushState({}, '', '/terms'))
  await page.evaluate(() => dispatchEvent(new PopStateEvent('popstate')))
  await expect(page.locator('.legal-panel')).toBeVisible()
  await expect(picker).toHaveCount(0)
  await expect(page.locator('body')).not.toHaveCSS('overflow', 'hidden')
  header.assertQuiet()
})
