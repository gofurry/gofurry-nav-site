import { test, expect, customSeed, searchPopupURL } from '../fixtures/nav-home-header'
import { mkdir } from 'node:fs/promises'

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

test('Reduced mobile Home retains automatic reveal and a static loading suggestion', async ({ page, header }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await header.open(390)
  await expect(page.locator('.nav-tool-dock')).toHaveCount(1)
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

test('Search keeps typography, theme parity, debounce, keyboard selection and reveal lock', async ({ page, header }) => {
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
  // window listener while Search owns its reveal lock. Only observe the event.
  await page.evaluate(() => {
    document.documentElement.dataset.headerWheelObserved = 'false'
    window.addEventListener('wheel', () => { document.documentElement.dataset.headerWheelObserved = 'true' }, { once: true, passive: true })
  })
  await page.mouse.move(20, 400)
  await page.mouse.wheel(0, 180)
  await expect(page.locator('html')).toHaveAttribute('data-header-wheel-observed', 'true')
  await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))))
  await expect(page.locator('.nav-tool-dock')).toHaveCount(0)
  await input.press('Escape')
  await page.mouse.click(20, 400)
  await expect(input).not.toBeFocused()
  await page.mouse.wheel(0, 300)
  await expect(page.locator('.nav-tool-dock')).toHaveCount(1)
  // Let the real reveal's scrollIntoView finish before returning to Search;
  // otherwise its pending smooth scroll can race a one-shot scrollTo(0).
  await expect.poll(() => page.locator('.nav-content-shell').evaluate(element => {
    const target = Math.min(element.getBoundingClientRect().top + scrollY,
      document.documentElement.scrollHeight - innerHeight)
    return Math.abs(scrollY - target)
  })).toBeLessThanOrEqual(1)
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
  await expect(name).toHaveValue('')
  await expect(url).toHaveValue('')
  await expect(error).toHaveCount(0)
  await modal.locator('.quick-modal-item').filter({ hasText: 'Example Site' }).getByRole('button', { name: '删除网站', exact: true }).click()
  await expect(modal.locator('.quick-modal-item')).toHaveCount(2)
  await expect(modal.getByText('Example Site', { exact: true })).toHaveCount(0)
  expect(await page.evaluate(() => JSON.parse(localStorage.getItem('navCustomSites')!))).toEqual(customSeed)
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

test('IME owns confirmation and navigation keys; only final Chinese text starts one debounce', async ({ page, context, header }) => {
  await header.open()
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
