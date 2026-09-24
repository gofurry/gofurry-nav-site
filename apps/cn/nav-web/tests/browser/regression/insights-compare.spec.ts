import type { Page } from '@playwright/test'
import { test, expect, pathFor, compareGroups, openRuntime, revealImages, keyboardFocus } from '../fixtures/insights-compare'
const columns = (page: Page) => page.locator('[data-compare-entity-id]').evaluateAll(els => els.map(el => Number(el.getAttribute('data-compare-entity-id'))))
const selected = (page: Page) => page.locator('[data-selected-entity]').evaluateAll(els => els.map(el => Number(el.getAttribute('data-selected-entity'))))
async function layout(page: Page, matrix = true) {
  expect(await page.evaluate(() => Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth)).toBeLessThanOrEqual(1)
  if (!matrix) return
  const pane = page.locator('.insight-compare-matrix-scroll')
  await pane.scrollIntoViewIfNeeded()
  if (page.viewportSize()!.width === 390) expect(await pane.evaluate(el => el.scrollWidth > el.clientWidth)).toBe(true)
  await pane.evaluate(el => { el.scrollLeft = 180; el.scrollTop = 230 })
  const box = await pane.evaluate(el => {
    const rect = el.getBoundingClientRect(), corner = el.querySelector('thead th')!.getBoundingClientRect(), fact = el.querySelector('th[scope="row"]')!.getBoundingClientRect()
    return { left: corner.left - rect.left, top: corner.top - rect.top, fact: fact.left - rect.left,
      background: getComputedStyle(el.querySelector('thead th')!).backgroundColor }
  })
  for (const value of [box.left, box.top, box.fact]) expect(Math.abs(value - 1)).toBeLessThanOrEqual(1)
  expect(box.background).not.toMatch(/^rgba/)
  await keyboardFocus(pane)
  await pane.evaluate(el => { el.scrollLeft = 0; el.scrollTop = 0 })
}
for (const domain of ['site', 'game'] as const) for (const locale of ['zh', 'en']) {
  test('Compare SSR selection validation, identity and noindex ' + domain + ' ' + locale, async ({ request, runtime }) => {
    for (const ids of ['', '1', '2,1', '2,1,4,3', '2,1,2', '1,bad', '1,2,3,4,5']) {
      const before = runtime.calls.length
      const response = await request.get(pathFor(domain, locale, ids, 'US')), html = await response.text()
      expect(response.status()).toBe(200); expect(html).toMatch(/<h1>[^<]+<\/h1>/)
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
  for (const width of [1440, 1024, 390]) for (const ids of ['', '2,1', '2,1,4,3']) {
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
  await page.locator('[data-selected-entity="' + removed + '"] button').click()
  await expect.poll(() => columns(page)).toEqual(chosen)
  if (domain === 'game') {
    await page.locator('[data-workspace-option="HK"]').click()
    await expect(page).toHaveURL(url => url.searchParams.get('region') === 'HK')
    await expect(page.locator('[data-compare-status="ready"]')).toBeVisible()
    expect(runtime.state.searches).toHaveLength(1)
  }
  expect(runtime.count('/sites/directory')).toBe(directory)
  await page.locator('button:has(img[alt="EN"])').first().click()
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
  for (const width of [1440, 1024, 390]) test('Compare Dark matrix ' + domain + ' ' + width, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    await openRuntime(page, pathFor(domain, 'en', '2,1,4,3'))
    await page.getByRole('button', { name: 'Toggle theme icon', exact: true }).click()
    await expect(page.locator('html')).toHaveClass(/dark/)
    await layout(page); runtime.assertQuiet()
  })
}
