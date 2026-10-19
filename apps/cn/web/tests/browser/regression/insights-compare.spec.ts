import type { Page } from '@playwright/test'
import { mkdir, writeFile } from 'node:fs/promises'
import { test, expect, pathFor, compareGroups, openRuntime, revealImages, keyboardFocus } from '../fixtures/insights-compare'
const columns = (page: Page) => page.locator('[data-compare-entity-id]').evaluateAll(els => els.map(el => Number(el.getAttribute('data-compare-entity-id'))))
const selected = (page: Page) => page.locator('[data-selected-entity]').evaluateAll(els => els.map(el => Number(el.getAttribute('data-selected-entity'))))
const colorAlpha = (color: string) => Number(color.match(/\/\s*([\d.]+)\s*\)$/)?.[1] ?? color.match(/^rgba\([^,]+,[^,]+,[^,]+,\s*([\d.]+)\)$/)?.[1] ?? (color === 'transparent' ? 0 : 1))

async function observePickerMotion(page: Page) {
  await page.locator('.insight-compare-picker').evaluate(root => {
    const seen = new WeakSet<Element>()
    const entries: { kind: string; id: string | null; opacity: string; transform: string; duration: string; property: string }[] = []
    new MutationObserver(() => {
      for (const node of root.querySelectorAll('.compare-popup-enter-from, .compare-selected-enter-from')) {
        if (seen.has(node)) continue
        seen.add(node)
        const css = getComputedStyle(node)
        entries.push({ kind: node.matches('.compare-popup-enter-from') ? 'popup' : 'selected', id: node.getAttribute('data-selected-entity'),
          opacity: css.opacity, transform: css.transform, duration: css.transitionDuration, property: css.transitionProperty })
      }
      root.setAttribute('data-c2-entries', JSON.stringify(entries))
    }).observe(root, { childList: true, subtree: true, attributes: true, attributeFilter: ['class'] })
  })
}
async function pickerEntries(page: Page) {
  return JSON.parse((await page.locator('.insight-compare-picker').getAttribute('data-c2-entries')) || '[]') as {
    kind: string; id: string | null; opacity: string; transform: string; duration: string; property: string
  }[]
}
async function screenshotPicker(page: Page, name: string) {
  if (!process.env.GOFURRY_C2_REVIEW) return
  const picker = page.locator('.insight-compare-picker')
  await picker.evaluate(async root => {
    await document.fonts.ready
    await Promise.all([...root.querySelectorAll('img')].map(img => img.decode()))
    await Promise.all(root.getAnimations({ subtree: true }).map(animation => animation.finished.catch(() => {})))
  })
  await mkdir('.cache/139-c2-review', { recursive: true })
  await page.screenshot({ path: `.cache/139-c2-review/${name}.png` })
}

test('C2 Motion Normal popup enters once through Loading and only new selections fade', async ({ page, runtime }) => {
  const warnings: string[] = []
  page.on('console', message => { if (message.type() === 'warning') warnings.push(message.text()) })
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  const html = await openRuntime(page, pathFor('game', 'en', '2,1', 'US'))
  expect([...html.matchAll(/data-selected-entity="(\d+)"/g)].map(match => Number(match[1]))).toEqual([2, 1])
  const picker = page.locator('.insight-compare-picker'), input = page.getByRole('combobox')
  const popup = page.locator('.insight-compare-picker__popup')
  await expect(picker.locator('.insight-compare-selected')).toHaveJSProperty('tagName', 'UL')
  await expect(picker.locator('[data-selected-entity="2"]')).toHaveCSS('opacity', '1')
  await expect(picker.locator('[data-selected-entity="2"]')).toHaveCSS('transition-duration', '0s')
  await observePickerMotion(page)
  const gate = runtime.hold((url, body) => url.pathname.endsWith('/search/simple') && (body as { txt?: string })?.txt === 'fox')
  await input.fill('fox'); await gate.wait()
  expect(await pickerEntries(page)).toEqual([{ kind: 'popup', id: null, opacity: '0', transform: 'matrix(1, 0, 0, 1, 0, -4)',
    duration: '0.18s, 0.18s', property: 'opacity, transform' }])
  const original = await popup.elementHandle()
  await expect(popup.getByRole('listbox')).toHaveAttribute('aria-busy', 'true')
  gate.release(); await gate.done()
  await expect(page.getByRole('option').first()).toBeVisible()
  expect(await original!.evaluate(el => el === document.querySelector('.insight-compare-picker__popup'))).toBe(true)
  expect((await pickerEntries(page)).filter(entry => entry.kind === 'popup')).toHaveLength(1)
  await screenshotPicker(page, 'compare-normal-1440-light')
  await input.press('ArrowDown'); await input.press('Enter')
  await expect(page).toHaveURL(url => url.searchParams.get('ids') === '2,1,4')
  expect(await popup.count()).toBe(0)
  expect(await original!.evaluate(el => el.isConnected)).toBe(false)
  await expect(input).toHaveAttribute('aria-expanded', 'false')
  await expect(input).toBeFocused()
  expect((await pickerEntries(page)).filter(entry => entry.kind === 'selected')).toEqual([
    { kind: 'selected', id: '4', opacity: '0', transform: 'none', duration: '0.16s', property: 'opacity' },
  ])
  const added = await picker.locator('[data-selected-entity="4"]').elementHandle()
  await expect.poll(() => columns(page)).toEqual([2, 1, 4])
  await picker.locator('[data-selected-entity="4"] button').click()
  await expect(page).toHaveURL(url => url.searchParams.get('ids') === '2,1')
  expect(await picker.locator('[data-selected-entity="4"]').count()).toBe(0)
  expect(await added!.evaluate(el => el.isConnected)).toBe(false)
  await expect.poll(() => columns(page)).toEqual([2, 1])
  expect(runtime.state.searches).toEqual(['fox'])
  expect(await picker.locator('.insight-compare-selected li').evaluateAll(nodes => nodes.every(node =>
    getComputedStyle(node).transform === 'none' && !(node as HTMLElement).style.transform && !node.className.includes('-move')))).toBe(true)
  expect(await page.locator('.insight-compare-matrix').evaluate(root => root.getAnimations({ subtree: true }).length)).toBe(0)
  await input.focus()
  const closing = await popup.elementHandle()
  await input.press('Escape')
  expect(await closing!.evaluate(el => el.isConnected)).toBe(false)
  expect(await popup.count()).toBe(0)
  expect(warnings).toEqual([])
  if (process.env.GOFURRY_C2_REVIEW) await writeFile('.cache/139-c2-review/normal-entries.json', JSON.stringify(await pickerEntries(page), null, 2))
  runtime.assertQuiet()
})

test('C2 Motion initial Reduced and live interruptions restore without replay at 390 Dark', async ({ page, runtime }) => {
  await page.setViewportSize({ width: 390, height: 1000 })
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openRuntime(page, pathFor('site', 'en', '2,1'))
  await page.getByRole('button', { name: 'Toggle theme icon', exact: true }).click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  const picker = page.locator('.insight-compare-picker'), input = page.getByRole('combobox')
  const popup = page.locator('.insight-compare-picker__popup')
  await observePickerMotion(page)
  await input.focus()
  await expect(page.getByRole('option').first()).toBeVisible()
  await input.press('ArrowDown')
  await expect(popup).toHaveCSS('opacity', '1')
  await expect(popup).toHaveCSS('transform', 'none')
  expect((await pickerEntries(page))[0]).toMatchObject({ kind: 'popup', opacity: '1', transform: 'none', duration: '0s' })
  await input.press('Enter')
  await expect(page).toHaveURL(url => url.searchParams.get('ids') === '2,1,4')
  expect(await popup.count()).toBe(0)
  expect((await pickerEntries(page)).find(entry => entry.kind === 'selected')).toMatchObject({ id: '4', opacity: '1', duration: '0s' })
  await expect(input).toBeFocused()
  await picker.locator('[data-selected-entity="4"] button').click()
  await expect(page).toHaveURL(url => url.searchParams.get('ids') === '2,1')

  await page.exposeFunction('c2Reduce', () => page.emulateMedia({ reducedMotion: 'reduce' }))
  for (const kind of ['popup', 'selected'] as const) {
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    await picker.evaluate((root, kind) => {
      const onRun = (event: Event) => {
        const target = event.target as HTMLElement
        if (!target.matches(`.compare-${kind}-enter-active`)) return
        root.removeEventListener('transitionrun', onRun)
        root.setAttribute(`data-c2-${kind}-duration`, getComputedStyle(target).transitionDuration)
        void (window as unknown as { c2Reduce(): Promise<void> }).c2Reduce().then(() => root.setAttribute(`data-c2-${kind}-reduced`, 'true'))
      }
      root.addEventListener('transitionrun', onRun)
    }, kind)
    if (kind === 'popup') await input.focus()
    else { await input.press('ArrowDown'); await input.press('Enter') }
    await expect(picker).toHaveAttribute(`data-c2-${kind}-reduced`, 'true')
    await expect(picker).toHaveAttribute(`data-c2-${kind}-duration`, kind === 'popup' ? '0.18s, 0.18s' : '0.16s')
    const target = kind === 'popup' ? popup : picker.locator('[data-selected-entity="4"]')
    await expect(target).toHaveCSS('opacity', '1')
    await expect(target).toHaveCSS('transform', 'none')
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    expect(await target.evaluate(el => el.getAnimations().length)).toBe(0)
    if (kind === 'selected') {
      await expect.poll(() => columns(page)).toEqual([2, 1, 4])
      await picker.locator('[data-selected-entity="4"] button').click()
      await expect(page).toHaveURL(url => url.searchParams.get('ids') === '2,1')
      expect(await target.count()).toBe(0)
    }
  }
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await input.focus()
  await screenshotPicker(page, 'compare-reduced-390-dark')
  await input.press('Escape')
  expect(await popup.count()).toBe(0)
  expect(runtime.count('/sites/directory')).toBe(1)
  expect(runtime.state.searches).toEqual([])
  await expect.poll(() => columns(page)).toEqual([2, 1])
  runtime.assertQuiet()
})

test('C2 Motion rapid add remove and route departure leave no stale interactive DOM', async ({ page, runtime }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await openRuntime(page, pathFor('game', 'en'))
  const picker = page.locator('.insight-compare-picker'), input = page.getByRole('combobox')
  const popup = page.locator('.insight-compare-picker__popup')
  await input.fill('fox')
  await expect(page.getByRole('option').first()).toBeVisible()
  // Remove from the real button during the new item's opacity transition.
  await picker.evaluate(root => root.addEventListener('transitionrun', function removeDuringEnter(event) {
    const target = event.target as HTMLElement
    if (!target.matches('.compare-selected-enter-active')) return
    root.removeEventListener('transitionrun', removeDuringEnter)
    root.setAttribute('data-c2-added', target.getAttribute('data-selected-entity')!)
    target.querySelector('button')!.click()
  }))
  await input.press('ArrowDown'); await input.press('Enter')
  await expect(picker).toHaveAttribute('data-c2-added', '2')
  await expect(page).toHaveURL(url => !url.searchParams.has('ids'))
  expect(await selected(page)).toEqual([])
  expect(await picker.locator('.compare-selected-leave-active').count()).toBe(0)
  for (let index = 0; index < 2; index++) {
    await input.press('ArrowDown'); await input.press('Enter')
    await expect(page).toHaveURL(url => url.searchParams.get('ids') === '2')
    expect(await selected(page)).toEqual([2])
    expect(await popup.count()).toBe(0)
    await expect(input).toBeFocused()
    await picker.locator('[data-selected-entity="2"] button').click()
    await expect(page).toHaveURL(url => !url.searchParams.has('ids'))
    expect(await selected(page)).toEqual([])
  }
  // Escape also interrupts entry synchronously, including rapid reopen.
  for (let index = 0; index < 2; index++) {
    await input.press('ArrowDown')
    const old = await popup.elementHandle()
    await input.press('Escape')
    expect(await old!.evaluate(el => el.isConnected)).toBe(false)
    expect(await popup.count()).toBe(0)
  }
  const oldPicker = await picker.elementHandle()
  await picker.evaluate(root => root.addEventListener('transitionrun', event => {
    if ((event.target as Element).matches('.compare-popup-enter-active')) {
      document.querySelector<HTMLAnchorElement>('a[href="/en/about"]')!.click()
    }
  }, { once: true }))
  await input.press('ArrowDown')
  await expect(page).toHaveURL(/\/en\/about$/)
  expect(await oldPicker!.evaluate(el => el.isConnected)).toBe(false)
  expect(await popup.count()).toBe(0)
  await page.locator('.about-avatar').evaluate(img => (img as HTMLImageElement).decode())
  await page.goBack()
  await expect(page.getByRole('combobox')).toBeVisible()
  expect(await selected(page)).toEqual([])
  expect(await popup.count()).toBe(0)
  expect(runtime.state.searches).toEqual(['fox'])
  runtime.assertQuiet()
})

async function layout(page: Page, matrix = true) {
  const english = new URL(page.url()).pathname.startsWith('/en/'), site = new URL(page.url()).pathname.includes('/sites/')
  const title = site ? (english ? 'Site Capability Compare' : '网站能力对比') : (english ? 'Game Fact Compare' : '游戏事实对比')
  await expect(page.getByRole('heading', { level: 1, name: title, exact: true })).toHaveCount(1)
  await expect(page.locator('h1')).toHaveClass('sr-only')
  await expect(page.locator('h1')).toHaveCSS('clip-path', 'inset(50%)')
  await expect(page.locator('.insights-workspace-header p')).toHaveCount(0)
  const header = await page.locator('.insights-workspace-header').boundingBox()
  expect(header!.height).toBe(0)
  expect(await page.locator('.insight-compare-picker').evaluate(el => {
    const nav = document.querySelector('.ecosystem-navigation')!, gap = el.getBoundingClientRect().top - nav.getBoundingClientRect().bottom
    return gap >= 0 && gap <= parseFloat(getComputedStyle(nav).marginBottom) + 1
  })).toBe(true)

  expect(await page.evaluate(() => Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth)).toBeLessThanOrEqual(1)
  const material = await page.locator('.insight-compare-picker').evaluate(el => {
    const style = getComputedStyle(el)
    return { background: style.backgroundColor, border: style.borderTopWidth, radius: parseFloat(style.borderRadius), shadow: style.boxShadow,
      width: el.getBoundingClientRect().width, overflow: style.overflow, dark: document.documentElement.classList.contains('dark') }
  })
  expect(colorAlpha(material.background)).toBeGreaterThanOrEqual(material.dark ? .1 : .3)
  expect(colorAlpha(material.background)).toBeLessThanOrEqual(material.dark ? .14 : .4)
  expect(material.border).toBe('1px'); expect(material.shadow).toBe('none')
  expect(material.radius).toBeGreaterThanOrEqual(12); expect(material.radius).toBeLessThanOrEqual(16)
  expect(material.width).toBeLessThanOrEqual(960)
  expect(material.overflow).toBe('visible')
  await expect(page.locator('.insight-compare-result')).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  await expect(page.locator('.insight-compare-result')).toHaveCSS('box-shadow', 'none')
  await expect(page.locator('.insight-compare-selected li').first()).toHaveCount(matrix ? 1 : 0)
  if (matrix) await expect(page.locator('.insight-compare-selected li').first()).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  if (!matrix) return
  const pane = page.locator('.insight-compare-matrix-scroll')
  await expect(pane).toHaveCSS('background-color', material.background)
  await expect(pane).toHaveCSS('border-top-width', '1px')
  await expect(pane).toHaveCSS('box-shadow', 'none')
  await expect(pane).toHaveCSS('border-radius', material.radius + 'px')
  await expect(page.locator('.insight-compare-matrix')).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  const padding = await page.locator('.insight-compare-matrix td').first().evaluate(el => parseFloat(getComputedStyle(el).paddingTop))
  expect(padding).toBeGreaterThanOrEqual(10); expect(padding).toBeLessThanOrEqual(14)
  await pane.scrollIntoViewIfNeeded()
  if (page.viewportSize()!.width === 390) expect(await pane.evaluate(el => el.scrollWidth > el.clientWidth)).toBe(true)
  await pane.evaluate(el => { el.scrollLeft = 180; el.scrollTop = 230 })
  const box = await pane.evaluate(el => {
    const rect = el.getBoundingClientRect(), corner = el.querySelector('thead th')!.getBoundingClientRect(), fact = el.querySelector('th[scope="row"]')!.getBoundingClientRect()
    const cornerCell = el.querySelector('thead th')!, topCell = el.querySelector('thead th:nth-child(2)')!, factCell = el.querySelector('th[scope="row"]')!
    return { left: corner.left - rect.left, top: corner.top - rect.top, fact: fact.left - rect.left,
      x: el.scrollLeft, y: el.scrollTop, canX: el.scrollWidth > el.clientWidth, canY: el.scrollHeight > el.clientHeight,
      cornerZ: Number(getComputedStyle(cornerCell).zIndex), topZ: Number(getComputedStyle(topCell).zIndex), factZ: Number(getComputedStyle(factCell).zIndex),
      cornerOwnsPoint: cornerCell.contains(document.elementFromPoint(rect.left + 16, rect.top + 16)),
      factBackground: getComputedStyle(factCell).backgroundColor, topBackground: getComputedStyle(topCell).backgroundColor,
      background: getComputedStyle(el.querySelector('thead th')!).backgroundColor }
  })
  for (const value of [box.left, box.top, box.fact]) expect(Math.abs(value - 1)).toBeLessThanOrEqual(1)
  expect(box.background).not.toMatch(/^rgba/)
  if (box.canX) expect(box.x).toBeGreaterThan(0)
  if (box.canY) expect(box.y).toBeGreaterThan(0)
  if (page.viewportSize()!.width === 390) { expect(box.x).toBeGreaterThan(0); expect(box.y).toBeGreaterThan(0) }
  expect(box.cornerOwnsPoint).toBe(true)
  expect(box.cornerZ).toBeGreaterThan(box.topZ); expect(box.topZ).toBeGreaterThan(box.factZ)
  for (const color of [box.background, box.factBackground, box.topBackground]) expect(colorAlpha(color)).toBe(1)
  await keyboardFocus(pane)
  await pane.evaluate(el => { el.scrollLeft = 0; el.scrollTop = 0 })
}
for (const domain of ['site', 'game'] as const) for (const locale of ['zh', 'en']) {
  test('Compare SSR selection validation, identity and noindex ' + domain + ' ' + locale, async ({ request, runtime }) => {
    for (const ids of ['', '1', '2,1', '2,1,4,3', '2,1,2', '1,bad', '1,2,3,4,5']) {
      const before = runtime.calls.length
      const response = await request.get(pathFor(domain, locale, ids, 'US')), html = await response.text()
      expect(response.status()).toBe(200)
      const title = domain === 'site' ? (locale === 'en' ? 'Site Capability Compare' : '网站能力对比') : (locale === 'en' ? 'Game Fact Compare' : '游戏事实对比')
      expect([...html.matchAll(/<h1\b[^>]*>([^<]+)<\/h1>/g)].map(match => match[1])).toEqual([title])
      expect(html).toMatch(/<meta[^>]*name="robots"[^>]*content="noindex, follow"/)
      const ready = ['2,1', '2,1,4,3', '2,1,2'].includes(ids)
      expect(runtime.calls.length - before).toBe(Number(ready))
      if (ready) {
        expect([...html.matchAll(/data-compare-entity-id="(\d+)"/g)].map(match => Number(match[1]))).toEqual([...new Set(ids.split(',').map(Number))])
        expect(html).toContain('Cedar Archive'); expect(html).toContain('Fox Atlas')
        expect(html).toContain(domain === 'site' ? '/nav/sites/2/icon/' + 'a'.repeat(32) + '.svg' : '/media/game-2.svg')
        expect(html).toContain('href="' + (locale === 'en' ? '/en' : '') + '/' + (domain === 'site' ? 'site' : 'games') + '/2"')
      }
    }
    runtime.assertQuiet()
  })
  for (const width of [1440, 1024, 768, 390]) for (const ids of ['', '2,1', '2,1,4,3']) {
    test('Compare hydration and matrix ' + domain + ' ' + locale + ' ' + width + ' ' + (ids || 'builder'), async ({ page, runtime }) => {
      await page.setViewportSize({ width, height: 1000 })
      await openRuntime(page, pathFor(domain, locale, ids, 'US'))
      await expect(page.locator('.insights-container')).not.toContainText(/insights\.[a-zA-Z]+\./)
      await expect.poll(() => selected(page)).toEqual(ids ? ids.split(',').map(Number) : [])
      if (domain === 'site') await expect.poll(() => runtime.count('/sites/directory')).toBe(1)
      if (ids) {
        await expect.poll(() => columns(page)).toEqual(ids.split(',').map(Number))
        expect(await page.locator('[data-compare-group]').evaluateAll(els => els.map(el => el.getAttribute('data-compare-group')))).toEqual(compareGroups[domain])
        await expect(page.locator('[data-compare-fact]')).toHaveCount(domain === 'site' ? 10 : 15)
        if (domain === 'site') expect(await page.locator('[data-capability-state="unknown"]').count()).toBeGreaterThan(0)
        if (domain === 'game') {
          await expect(page.locator('[data-current-player-available="true"]').first()).toHaveText('0')
          await expect(page.locator('[data-current-player-available="false"]').first()).toHaveText('—')
          await expect(page.locator('[data-price-state="priced"]').first()).toContainText('$0.00')
          expect(await page.locator('[data-language-evidence="stale"]').count()).toBeGreaterThan(0)
        }
      }
      await layout(page, Boolean(ids))
      expect(runtime.calls).toHaveLength(Number(Boolean(ids)) + Number(domain === 'site'))
      runtime.assertQuiet()
    })
  }
}
for (const domain of ['site', 'game']) test('Compare keyboard picker, max four, cache, locale and reset ' + domain, async ({ page, runtime }) => {
  await openRuntime(page, pathFor(domain))
  const input = page.getByRole('combobox')
  await input.fill('f'); await input.fill('fo'); await input.fill('fox')
  await expect(page.getByRole('option').first()).toBeVisible()
  // Measure the populated popup; after all matching Sites are selected, the
  // legitimate empty message is shorter than the picker and need not overlap it.
  expect(await page.locator('.insight-compare-picker__popup').evaluate(el => {
    const rect = el.getBoundingClientRect(), picker = el.closest('.insight-compare-picker')!.getBoundingClientRect()
    const y = Math.min(rect.bottom - 8, innerHeight - 8)
    return rect.bottom > picker.bottom && el.contains(document.elementFromPoint(rect.left + 12, y))
  })).toBe(true)
  const first = Number(await page.getByRole('option').first().getAttribute('data-picker-result'))
  expect(first).toBe(domain === 'site' ? 3 : 2)
  if (domain === 'game') expect(runtime.state.searches).toEqual(['fox'])
  const directory = runtime.count('/sites/directory')
  await input.press('ArrowDown'); await expect(input).toHaveAttribute('aria-activedescendant', /.+/)
  await input.press('Enter'); await expect(page).toHaveURL(url => url.searchParams.get('ids') === String(first))
  await expect(page.locator('[data-compare-result]')).toHaveCount(0)
  const chosen = [first]
  for (let count = 1; count < 4; count++) {
    await input.press('ArrowDown')
    const options = await page.getByRole('option').evaluateAll(els => els.map(el => Number(el.getAttribute('data-picker-result'))))
    expect(options.every(id => !chosen.includes(id))).toBe(true)
    chosen.push(options[0]!)
    await input.press('Enter')
    await expect(page).toHaveURL(url => url.searchParams.get('ids') === chosen.join(','))
    await expect(page.locator('[data-compare-status="ready"]')).toBeVisible()
    await expect.poll(() => columns(page)).toEqual(chosen)
  }
  await input.press('ArrowDown')
  await expect(page.locator('[role="option"][aria-disabled="true"]')).toHaveCount(await page.getByRole('option').count())
  await input.press('Enter'); expect(await selected(page)).toEqual(chosen)
  await input.press('Escape'); await expect(input).toHaveAttribute('aria-expanded', 'false')
  const removed = chosen.splice(1, 1)[0]
  await keyboardFocus(page.locator('[data-selected-entity="' + removed + '"] button'))
  await page.locator('[data-selected-entity="' + removed + '"] button').click()
  await expect.poll(() => columns(page)).toEqual(chosen)
  if (domain === 'game') {
    await page.locator('[data-workspace-option="HK"]').click()
    await expect(page).toHaveURL(url => url.searchParams.get('region') === 'HK')
    await expect(page.locator('[data-compare-status="ready"]')).toBeVisible()
    expect(runtime.state.searches).toHaveLength(1)
  }
  expect(runtime.count('/sites/directory')).toBe(directory)
  await page.getByRole('button', { name: 'Switch to English', exact: true }).click()
  await expect(page).toHaveURL(url => url.pathname.startsWith('/en/') && url.searchParams.get('ids') === chosen.join(','))
  await expect(input).toHaveValue('')
  if (domain === 'game') expect(new URL(page.url()).searchParams.get('region')).toBe('HK')
  await openRuntime(page, pathFor(domain, 'en', '1,bad', 'US'))
  await expect(page.getByRole('alert')).toBeVisible()
  await page.getByRole('button', { name: 'Reset selection', exact: true }).click()
  await expect(page).toHaveURL(url => !url.searchParams.has('ids'))
  if (domain === 'game') expect(new URL(page.url()).searchParams.get('region')).toBe('US')
  runtime.assertQuiet()
})
test('Compare held stale search cannot replace new results; search failure preserves matrix', async ({ page, runtime }) => {
  await openRuntime(page, pathFor('game', 'en', '2,1'))
  const input = page.getByRole('combobox')
  const gate = runtime.hold((url, body) => url.pathname.endsWith('/search/simple') && (body as { txt?: string })?.txt === 'slow')
  await input.fill('slow'); await gate.wait()
  gate.expectAbort()
  const fast = page.waitForResponse(res => res.request().postDataJSON()?.txt === 'fox')
  await input.fill('fox'); await fast
  await expect(page.getByRole('option').first()).toBeVisible()
  gate.release(); await gate.done()
  await expect(page.getByRole('listbox')).not.toContainText('Stale result')
  runtime.state.searchFailure = true
  await input.fill('failure')
  await expect(page.getByText('Search is unavailable. Your selection and comparison are preserved.')).toBeVisible()
  expect(await columns(page)).toEqual([2, 1]); runtime.assertQuiet()
})
for (const domain of ['site', 'game']) for (const state of ['compareFailure', 'insufficient', 'reverse'] as const) {
  test('Compare preserves selected state ' + domain + ' ' + state, async ({ page, runtime }) => {
    runtime.state[state] = true
    const ids = state === 'reverse' ? '2,1,4,3' : '2,1'
    await openRuntime(page, pathFor(domain, 'en', ids))
    if (state === 'reverse') await expect.poll(() => columns(page)).toEqual([2, 1, 4, 3])
    else {
      await expect(page.locator('[data-compare-status="' + (state === 'compareFailure' ? 'error' : 'insufficient_data') + '"]')).toBeVisible()
      expect(await selected(page)).toEqual([2, 1])
      expect(new URL(page.url()).searchParams.get('ids')).toBe(ids)
      if (state === 'insufficient') await expect(page.locator('[data-compare-result]')).toHaveCount(0)
    }
    runtime.assertQuiet()
  })
}
test('Compare directory failure leaves Site matrix intact', async ({ page, runtime }) => {
  runtime.state.directoryFailure = true
  await openRuntime(page, pathFor('site', 'en', '2,1'))
  await page.getByRole('combobox').focus()
  await expect(page.getByText('Search is unavailable. Your selection and comparison are preserved.')).toBeVisible()
  expect(await columns(page)).toEqual([2, 1]); runtime.assertQuiet()
})
for (const domain of ['site', 'game']) {
  test('Compare real media fallback ' + domain, async ({ page, runtime }) => {
    runtime.failImages = true
    await openRuntime(page, pathFor(domain, 'en', '2,1'))
    await revealImages(page)
    await expect(page.locator('.insight-compare-matrix .insight-entity-media img')).toHaveCount(0)
    await expect(page.locator('.insight-compare-matrix [role="img"][aria-label]')).toHaveCount(2)
    runtime.assertQuiet()
  })
  for (const locale of ['zh', 'en']) for (const width of [1440, 1024, 768, 390]) test('Compare Dark matrix ' + domain + ' ' + width + ' ' + locale, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    await openRuntime(page, pathFor(domain, locale, '2,1,4,3'))
    await page.getByRole('button', { name: locale === 'en' ? 'Toggle theme icon' : '切换明暗主题图标', exact: true }).click()
    await expect(page.locator('html')).toHaveClass(/dark/)
    await layout(page); runtime.assertQuiet()
  })
}
