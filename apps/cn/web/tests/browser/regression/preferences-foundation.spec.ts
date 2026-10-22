import { test, expect } from '../fixtures/hero-preferences'
import { mkdir } from 'node:fs/promises'

test('Preferences interrupts only its own motion; close is inert, restores focus and can immediately reopen', async ({ page, preferences }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await preferences.open({ theme: 'light', seedSteamDiagnostics: true })
  await page.exposeFunction('interruptPreferencesMotion', async (action: string) => {
    if (action === 'reopen') await page.keyboard.press('Enter')
    else await page.emulateMedia({ reducedMotion: 'reduce' })
  })
  await page.evaluate(() => {
    const evidence: object[] = []
    document.documentElement.dataset.interruptPreferences = 'enter:reduce'
    new MutationObserver(() => {
      for (const root of document.querySelectorAll<HTMLElement>('.gf-modal-backdrop')) {
        const phase = root.classList.contains('preferences-modal-leave-active') ? 'leave'
          : root.classList.contains('preferences-modal-enter-active') ? 'enter' : null
        if (!phase || root.dataset[phase + 'Observed']) continue
        root.dataset[phase + 'Observed'] = 'true'
        const panel = root.querySelector<HTMLElement>('.gf-preferences-modal')!
        const style = getComputedStyle(panel)
        if (phase === 'leave') root.querySelector<HTMLInputElement>('#mode-setting-input')!.focus()
        evidence.push({ phase, backdrop: getComputedStyle(root).transitionDuration,
          panel: style.transitionDuration, transform: style.transform,
          inert: root.inert, hidden: root.getAttribute('aria-hidden'),
          pointer: getComputedStyle(root).pointerEvents, focusInside: root.contains(document.activeElement),
          returnedToOpener: document.activeElement?.matches('.gf-nav__mode-button'),
          quickAccess: localStorage.getItem('nav-header-show-quick-access') })
        document.documentElement.dataset.preferencesMotion = JSON.stringify(evidence)
      }
    }).observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ['class'] })
    document.addEventListener('transitionrun', event => {
      const target = event.target
      if (!(target instanceof HTMLElement) || !target.matches('.gf-modal-backdrop, .gf-preferences-modal')) return
      const root = target.closest<HTMLElement>('.gf-modal-backdrop')!
      const phase = root.classList.contains('preferences-modal-leave-active') ? 'leave' : 'enter'
      const [requestedPhase, action] = (document.documentElement.dataset.interruptPreferences ?? '').split(':')
      if (phase !== requestedPhase || root.dataset[phase + 'Interrupted']) return
      root.dataset[phase + 'Interrupted'] = 'true'
      const bridge = window as Window & { interruptPreferencesMotion?: (action: string) => Promise<void> }
      void bridge.interruptPreferencesMotion!(action!)
        .then(() => { document.documentElement.dataset.preferencesActionDone = phase + ':' + action })
    })
  })
  const opener = page.locator('.gf-nav__mode-button')
  const modal = page.locator('.gf-preferences-modal')
  const backdrop = page.locator('.gf-modal-backdrop')
  const input = page.locator('#mode-setting-input')
  const cancel = modal.getByRole('button', { name: '取消', exact: true })
  const evidence = () => page.evaluate(() => JSON.parse(document.documentElement.dataset.preferencesMotion!))
  await preferences.openPreferences()
  await expect(page.locator('html')).toHaveAttribute('data-preferences-action-done', 'enter:reduce')
  expect((await evidence())[0]).toMatchObject({ phase: 'enter', backdrop: '0.16s', panel: '0.22s, 0.22s',
    transform: 'matrix(1, 0, 0, 1, 0, 8)', inert: false })
  await expect(backdrop).not.toHaveClass(/preferences-modal-enter-active/)
  await expect(backdrop).toHaveCSS('opacity', '1')
  await expect(modal).toHaveCSS('opacity', '1')
  await expect(modal).toHaveCSS('transform', 'none')
  await expect(backdrop).toHaveAttribute('data-enter-observed', 'true')
  await input.fill('draft')

  const tabs = page.locator('.gf-preferences-modal > .preferences-tabs').getByRole('tab')
  await tabs.first().focus()
  for (const [key, index] of [['End', 1], ['ArrowLeft', 0], ['Home', 0]] as const) {
    await page.keyboard.press(key)
    await expect(tabs.nth(index)).toBeFocused()
    await expect(tabs.nth(index)).toHaveAttribute('aria-selected', 'true')
    await expect(page.locator('.preferences-page').nth(index)).not.toHaveAttribute('inert')
    await expect.poll(() => page.locator('.preferences-pages').evaluate(el => Math.round(el.scrollLeft / el.clientWidth))).toBe(index)
  }
  await page.keyboard.press('Tab')
  await expect(input).toBeFocused()
  // Restoring Normal never replays the interrupted entry or replaces the draft.
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await expect(backdrop).not.toHaveClass(/preferences-modal-enter-active/)
  await expect(input).toHaveValue('draft')
  await page.evaluate(() => { document.documentElement.dataset.interruptPreferences = 'leave:reopen' })
  await cancel.click()
  await expect(page.locator('html')).toHaveAttribute('data-preferences-action-done', 'leave:reopen')
  const firstLeave = (await evidence()).find((entry: { phase: string }) => entry.phase === 'leave')
  expect(firstLeave).toMatchObject({ backdrop: '0.16s', panel: '0.16s, 0.16s', inert: true,
    hidden: 'true', pointer: 'none', focusInside: false, returnedToOpener: true })
  await expect(opener).toBeFocused()
  // The observer attempted focus during leave; inert rejected it before the
  // real Enter on the restored opener interrupted that leave and reopened.
  await expect(modal).toHaveCount(1)
  await expect(backdrop).not.toHaveAttribute('inert')
  await expect(input).not.toHaveValue('draft')
  expect((await evidence()).filter((entry: { phase: string }) => entry.phase === 'enter')).toHaveLength(2)
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await expect(backdrop).not.toHaveClass(/preferences-modal-enter-active/)

  if (process.env.GOFURRY_B1_REVIEW === '1') {
    await mkdir('.cache/139-b1-review', { recursive: true })
    await modal.screenshot({ path: '.cache/139-b1-review/preferences-desktop-light.png' })
  }
  await page.locator('#quick-access-toggle').click()
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await page.evaluate(() => { document.documentElement.dataset.interruptPreferences = 'leave:reduce' })
  await modal.getByRole('button', { name: '保存', exact: true }).click()
  await expect(page.locator('html')).toHaveAttribute('data-preferences-action-done', 'leave:reduce')
  expect(await page.evaluate(() => localStorage.getItem('nav-header-show-quick-access'))).toBe('0')
  expect((await evidence()).filter((entry: { phase: string }) => entry.phase === 'leave').at(-1))
    .toMatchObject({ inert: true, focusInside: false, quickAccess: '0' })
  await expect(opener).toBeFocused()
  await expect(modal).toHaveCount(0)
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await page.evaluate(() => { document.documentElement.dataset.interruptPreferences = '' })
  await preferences.openPreferences()
  await expect(page.locator('#quick-access-toggle')).toHaveAttribute('aria-pressed', 'false')
  await expect(backdrop).not.toHaveClass(/preferences-modal-enter-active/)
  await cancel.click()
  await page.mouse.move(20, 400)
  await page.mouse.wheel(0, 700)
  await page.getByRole('link', { name: '服务条款', exact: true }).click()
  await expect(page).toHaveURL(/\/terms$/)
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await expect(modal).toHaveCount(0)
  expect(preferences.errors).toEqual([])
})

test('Initial Reduced mobile Preferences closes immediately and returns to its actual menu action', async ({ page, preferences }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await preferences.open({ width: 390, theme: 'dark', seedSteamDiagnostics: true })
  await preferences.openPreferences()
  const modal = page.locator('.gf-preferences-modal')
  const backdrop = page.locator('.gf-modal-backdrop')
  await expect(backdrop).toHaveCSS('transition-duration', '0s')
  await expect(backdrop).toHaveCSS('opacity', '1')
  await expect(modal).toHaveCSS('transform', 'none')
  await expect(modal).toHaveCSS('transition-duration', '0s')
  await page.locator('#mode-setting-input').focus()
  await page.keyboard.press('Tab')
  await expect(page.locator('#quick-access-toggle')).toBeFocused()
  if (process.env.GOFURRY_B1_REVIEW === '1') {
    await mkdir('.cache/139-b1-review', { recursive: true })
    await modal.screenshot({ path: '.cache/139-b1-review/preferences-mobile-dark.png' })
  }
  await modal.getByRole('button', { name: '取消', exact: true }).click()
  expect(await modal.count()).toBe(0)
  await expect(page.locator('.gf-nav__mobile-action')).toBeFocused()
  await preferences.openPreferences()
  await preferences.cancelPreferences()
  await expect(page.locator('.gf-nav__mobile-action')).toBeFocused()
  expect(preferences.errors).toEqual([])
})

test('Quick Access Space/Enter, focus, Cancel and Save retain their draft semantics', async ({ page, preferences }) => {
  await preferences.open()
  await preferences.openPreferences()
  const toggle = page.locator('#quick-access-toggle')
  const input = page.locator('#mode-setting-input')
  await input.focus()
  await expect(input).toBeFocused()
  await expect(toggle).toHaveAttribute('aria-pressed', 'true')
  await toggle.focus()
  await page.keyboard.press('Space')
  await expect(toggle).toHaveAttribute('aria-pressed', 'false')
  await expect(toggle).toBeFocused()
  await preferences.cancelPreferences()
  await preferences.openPreferences()
  await expect(toggle).toHaveAttribute('aria-pressed', 'true')
  await toggle.click()
  await preferences.savePreferences()
  expect(await page.evaluate(() => localStorage.getItem('nav-header-show-quick-access'))).toBe('0')
  await preferences.openPreferences()
  await expect(toggle).toHaveAttribute('aria-pressed', 'false')
  await toggle.focus()
  await page.keyboard.press('Enter')
  await expect(toggle).toHaveAttribute('aria-pressed', 'true')
  await expect(toggle).toBeFocused()
  await preferences.savePreferences()
  expect(await page.evaluate(() => localStorage.getItem('nav-header-show-quick-access'))).toBe('1')
  expect(preferences.errors).toEqual([])
})

for (const width of [1440, 390, 320]) for (const dark of [false, true]) {
  test(`${width}px ${dark ? 'dark' : 'light'} Preferences retain Hero source appearance with no overflow and visible focus`, async ({ page, preferences }) => {
    await preferences.open({ width, mode: 'fixed', desktopId: '34' })
    await page.evaluate(dark => document.documentElement.classList.toggle('dark', dark), dark)
    await preferences.openPreferences()
    await preferences.waitCatalogPosition(25)
    await page.locator('[data-hero-catalog]:visible').scrollIntoViewIfNeeded()
    await page.mouse.move(2, 2)
    // Finish CSS transitions without writing success screenshots or style JSON.
    await page.evaluate(async () => {
      const animations = document.querySelector('.gf-preferences-modal')!.getAnimations({ subtree: true })
      await Promise.all(animations.filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity).map(animation => animation.finished.catch(() => {})))
    })
    await expect(page.locator('.gf-preferences-modal > .preferences-tabs').getByRole('tab')).toHaveCount(2)
    const widths = await page.locator('[data-hero-preferences] .preferences-sources button').evaluateAll(elements => elements.map(el => el.getBoundingClientRect().width))
    expect(widths).toHaveLength(3)
    expect(Math.max(...widths) - Math.min(...widths)).toBeLessThan(1)
    expect(await page.locator('.preferences-page').first().evaluate(el => el.scrollWidth <= el.clientWidth + 1)).toBe(true)
    expect(await page.locator('.gf-preferences-modal').evaluate(el => el.getBoundingClientRect().right <= innerWidth && el.getBoundingClientRect().left >= 0)).toBe(true)
    await page.locator('#mode-setting-input').focus()
    await expect(page.locator('#mode-setting-input')).toBeFocused()
    await page.keyboard.press('Tab')
    const source = page.getByRole('button', { name: '固定云端背景', exact: true })
    await source.focus()
    await expect(source).toBeFocused()
    expect(await source.evaluate(el => getComputedStyle(el).outlineStyle)).not.toBe('none')
    await page.getByRole('button', { name: '云端随机', exact: true }).click()
    await expect(page.getByRole('button', { name: '云端随机', exact: true })).toHaveAttribute('aria-pressed', 'true')
    await preferences.cancelPreferences()
    expect(preferences.errors).toEqual([])
  })
}
