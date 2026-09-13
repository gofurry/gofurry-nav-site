import assert from 'node:assert/strict'
import { mkdtemp } from 'node:fs/promises'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { compareFixtureResponse, compareGroups } from './fixtures/insights-compare.mjs'

const pathFor = (domain, locale = 'zh', ids = '', region = 'CN') => `${locale === 'en' ? '/en' : ''}/insights/${domain === 'site' ? 'sites' : 'games'}/compare?${new URLSearchParams({ ...(ids ? { ids } : {}), ...(domain === 'game' ? { region } : {}) })}`
const columnIDs = page => page.locator('[data-compare-entity-id]').evaluateAll(els => els.map(el => Number(el.dataset.compareEntityId)))
const selectedIDs = page => page.locator('[data-selected-entity]').evaluateAll(els => els.map(el => Number(el.dataset.selectedEntity)))

export async function runCompareSmoke() {
  const state = { searches: [], compareFailure: false, directoryFailure: false, searchFailure: false, insufficient: false, reverse: false }
  const app = await startInsightsFixtureApp((url, media, body) => compareFixtureResponse(url, media, body, state))
  const { base, requests } = app
  let browser
  const apiRequests = () => requests.filter(url => url.pathname.startsWith('/api/'))
  try {
    for (const domain of ['site', 'game']) for (const locale of ['zh', 'en']) {
      for (const ids of ['', '1', '2,1', '2,1,4,3', '2,1,2', '1,bad', '1,2,3,4,5']) {
        requests.length = 0
        const response = await fetch(base + pathFor(domain, locale, ids, 'US')), html = await response.text()
        assert.equal(response.status, 200)
        assert.match(html, /<h1>[^<]+<\/h1>/)
        assert.match(html, /<meta[^>]*name="robots"[^>]*content="noindex, follow"/)
        const ready = ['2,1', '2,1,4,3', '2,1,2'].includes(ids)
        assert.equal(apiRequests().length, Number(ready), 'SSR fetched auxiliary data or duplicate Compare')
        if (ready) {
          assert.deepEqual([...html.matchAll(/data-compare-entity-id="(\d+)"/g)].map(match => Number(match[1])), [...new Set(ids.split(',').map(Number))])
          assert(html.includes('Cedar Archive') && html.includes('Fox Atlas'), 'identity missing from SSR')
          assert(html.includes(domain === 'site' ? '/nav/sites/2/icon/' + 'a'.repeat(32) + '.svg' : '/media/game-2.svg'), 'visual missing from SSR')
          const entityRoute = `${locale === 'en' ? '/en' : ''}/${domain === 'site' ? 'site' : 'games'}/2`
          assert(html.includes(`href="${entityRoute}"`), 'localized entity link changed')
        }
      }
      console.log(`[compare] ${domain} ${locale} SSR 0/1/2/4/invalid, noindex, request budget PASS`)
    }
    browser = await launchPerfBrowser()
    const context = await browser.newContext({ locale: 'zh-CN', colorScheme: 'light' }), page = await context.newPage()
    const errors = []
    page.on('pageerror', error => errors.push(error.message))
    let imageFailure = false
    await page.route('**/*', route => {
      const url = new URL(route.request().url())
      if (imageFailure && (url.pathname.startsWith('/media/') || url.pathname.startsWith('/nav/sites/') || url.pathname === '/defaultLogo.svg')) return route.abort()
      return ['127.0.0.1', 'localhost'].includes(url.hostname) || ['data:', 'blob:'].includes(url.protocol) ? route.continue() : route.abort()
    })
    const screenshots = await mkdtemp(join(tmpdir(), 'gofurry-compare-b5-'))
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const domain of ['site', 'game']) for (const locale of ['zh', 'en']) for (const ids of ['', '2,1', '2,1,4,3']) {
        requests.length = 0
        await page.goto(base + pathFor(domain, locale, ids, 'US'), { waitUntil: 'networkidle' })
        assert(await page.locator('main h1').isVisible())
        assert(!/insights\.[a-zA-Z]+\./.test(await page.locator('.insights-container').innerText()), 'unresolved locale key leaked into UI')
        assert.equal(apiRequests().length, Number(!!ids) + Number(domain === 'site'), 'hydration request budget changed')
        assert.deepEqual(await selectedIDs(page), ids ? ids.split(',').map(Number) : [])
        if (ids) {
          assert.deepEqual(await columnIDs(page), ids.split(',').map(Number))
          assert.deepEqual(await page.locator('[data-compare-group]').evaluateAll(els => els.map(el => el.dataset.compareGroup)), compareGroups[domain])
          assert.equal(await page.locator('[data-compare-fact]').count(), domain === 'site' ? 10 : 15, 'matrix lost facts')
          await checkMatrix(page, width)
          if (domain === 'game') {
            assert.equal((await page.locator('[data-current-player-available="true"]').first().textContent()).trim(), '0')
            assert.equal((await page.locator('[data-current-player-available="false"]').first().textContent()).trim(), '—')
            assert((await page.locator('[data-price-state="priced"]').first().innerText()).includes('$0.00'))
          }
        }
        await checkPage(page, width)
        await page.evaluate(() => window.scrollTo(0, 0))
        await page.screenshot({ path: join(screenshots, `${domain}-${locale}-${ids ? ids.split(',').length : 0}-${width}.png`), fullPage: true })
      }
      console.log(`[compare] ${width}px zh/en builders + two/four columns, local scroll/sticky PASS`)
    }
    await exercisePicker(page, base, state, requests)
    await exerciseFailures(page, base, state)
    imageFailure = true
    for (const domain of ['site', 'game']) {
      await page.goto(base + pathFor(domain, 'en', '2,1'), { waitUntil: 'networkidle' })
      await page.locator('.insight-compare-matrix-scroll').scrollIntoViewIfNeeded()
      await page.waitForTimeout(200)
      assert.equal(await page.locator('.insight-compare-matrix .insight-entity-media img').count(), 0)
      assert.equal(await page.locator('.insight-compare-matrix [role="img"][aria-label]').count(), 2)
    }
    imageFailure = false
    await page.getByRole('button', { name: 'Toggle theme icon', exact: true }).click()
    for (const width of [1440, 1024, 390]) {
      await page.setViewportSize({ width, height: 1000 })
      for (const domain of ['site', 'game']) {
        await page.goto(base + pathFor(domain, 'en', '2,1,4,3'), { waitUntil: 'networkidle' })
        assert(await page.locator('html').evaluate(el => el.classList.contains('dark')))
        await checkMatrix(page, width)
        await checkPage(page, width)
        await page.locator('.insight-compare-matrix-scroll').scrollIntoViewIfNeeded()
        await page.screenshot({ path: join(screenshots, `${domain}-dark-${width}.png`) })
      }
    }
    assert.deepEqual(errors, [], 'Compare browser exception')
    assert(apiRequests().every(url => /\/(?:insights\/compare|sites\/directory|search\/simple)$/.test(url.pathname)), 'per-entity frontend lookup added')
    console.log('[compare] failures, insufficient data, stale search, fallback, dark matrix PASS')
    console.log(`[compare] screenshots (outside Git): ${screenshots}`)
  } catch (error) { console.error(app.logs()); throw error }
  finally { await browser?.close(); await app.close() }
}

async function checkPage(page, width) {
  const overflow = await page.evaluate(() => Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth)
  assert(overflow <= 1, `${width}px page horizontal overflow`)
}
async function checkMatrix(page, width) {
  const pane = page.locator('.insight-compare-matrix-scroll')
  await pane.scrollIntoViewIfNeeded()
  if (width === 390) assert(await pane.evaluate(el => el.scrollWidth > el.clientWidth), 'matrix must scroll locally')
  await pane.evaluate(el => { el.scrollLeft = 180; el.scrollTop = 230 })
  const geometry = await pane.evaluate(el => {
    const box = el.getBoundingClientRect(), corner = el.querySelector('thead th').getBoundingClientRect(), fact = el.querySelector('th[scope="row"]').getBoundingClientRect()
    return { cornerLeft: corner.left - box.left, cornerTop: corner.top - box.top, factLeft: fact.left - box.left, background: getComputedStyle(el.querySelector('thead th')).backgroundColor }
  })
  assert(Math.abs(geometry.cornerLeft - 1) <= 1 && Math.abs(geometry.cornerTop - 1) <= 1 && Math.abs(geometry.factLeft - 1) <= 1, `sticky drift: ${JSON.stringify(geometry)}`)
  assert(!geometry.background.startsWith('rgba'), 'sticky background must be opaque')
  await page.keyboard.press('Tab')
  await pane.focus()
  assert(await pane.evaluate(el => getComputedStyle(el).outlineStyle !== 'none'))
  await pane.evaluate(el => { el.scrollLeft = 0; el.scrollTop = 0 })
}

async function exercisePicker(page, base, state, requests) {
  await page.setViewportSize({ width: 1440, height: 1000 })
  for (const domain of ['site', 'game']) {
    await page.goto(base + pathFor(domain, 'zh'), { waitUntil: 'networkidle' })
    const input = page.getByRole('combobox')
    const directoryCount = requests.filter(url => url.pathname.endsWith('/sites/directory')).length
    state.searches.length = 0
    await input.fill('f')
    await input.fill('fo')
    await input.fill('fox')
    await page.locator('[role="option"]').first().waitFor()
    if (domain === 'game') assert.deepEqual(state.searches, ['fox'], 'Game search was not debounced')
    const firstID = Number(await page.locator('[role="option"]').first().getAttribute('data-picker-result'))
    assert.equal(firstID, domain === 'site' ? 3 : 2, 'search sort changed')
    await input.press('ArrowDown')
    assert(await input.getAttribute('aria-activedescendant'))
    await input.press('Enter')
    await page.waitForURL(url => url.searchParams.get('ids') === String(firstID))
    assert.equal(await page.locator('[data-compare-result]').count(), 0, 'one selection should stay builder')
    const chosen = [firstID]
    for (let count = 1; count < 4; count++) {
      await input.press('ArrowDown')
      const options = await page.locator('[role="option"]').evaluateAll(els => els.map(el => Number(el.dataset.pickerResult)))
      assert(options.every(id => !chosen.includes(id)), 'selected entities remained addable')
      chosen.push(options[0])
      await input.press('Enter')
      await page.waitForURL(url => url.searchParams.get('ids') === chosen.join(','))
      await page.locator('[data-compare-status="ready"]').waitFor()
      assert.deepEqual(await columnIDs(page), chosen)
    }
    await input.press('ArrowDown')
    assert.equal(await page.locator('[role="option"][aria-disabled="true"]').count(), await page.locator('[role="option"]').count())
    await input.press('Enter')
    assert.deepEqual(await selectedIDs(page), chosen, 'max-four limit failed')
    await input.press('Escape')
    assert.equal(await input.getAttribute('aria-expanded'), 'false')
    const removed = chosen.splice(1, 1)[0]
    await page.locator(`[data-selected-entity="${removed}"] button`).click()
    await page.waitForURL(url => url.searchParams.get('ids') === chosen.join(','))
    await page.locator('[data-compare-status="ready"]').waitFor()
    assert.deepEqual(await columnIDs(page), chosen)
    if (domain === 'game') {
      await page.locator('[data-workspace-option="HK"]').click()
      await page.waitForURL(url => url.searchParams.get('region') === 'HK')
      await page.locator('[data-compare-status="ready"]').waitFor()
      assert.deepEqual(await columnIDs(page), chosen)
    }
    assert.equal(requests.filter(url => url.pathname.endsWith('/sites/directory')).length, directoryCount, 'selection refetched Site directory')
    if (domain === 'game') assert.equal(state.searches.length, 1, 'selection refetched Game search')
    await page.locator('button:has(img[alt="EN"])').first().click()
    await page.waitForURL(url => url.pathname.startsWith('/en/'))
    await page.getByRole('combobox').waitFor()
    assert.equal(new URL(page.url()).searchParams.get('ids'), chosen.join(','))
    if (domain === 'game') assert.equal(new URL(page.url()).searchParams.get('region'), 'HK')
    assert.equal(await page.getByRole('combobox').inputValue(), '', 'locale kept search keyword')
    await page.goto(base + pathFor(domain, 'en', '1,bad', 'US'), { waitUntil: 'networkidle' })
    assert(await page.getByRole('alert').isVisible())
    await page.getByRole('button', { name: 'Reset selection', exact: true }).click()
    await page.waitForURL(url => !url.searchParams.has('ids'))
    if (domain === 'game') assert.equal(new URL(page.url()).searchParams.get('region'), 'US')
    console.log(`[compare] ${domain} keyboard/add/remove/max/order/region/locale/reset PASS`)
  }
}

async function exerciseFailures(page, base, state) {
  await page.goto(base + pathFor('game', 'en', '2,1'), { waitUntil: 'networkidle' })
  const input = page.getByRole('combobox')
  const slow = page.waitForRequest(request => request.url().endsWith('/game/search/simple') && request.postDataJSON()?.txt === 'slow')
  await input.fill('slow')
  await slow
  await input.fill('fox')
  await page.locator('[role="option"]').first().waitFor()
  await page.waitForTimeout(700)
  assert(!(await page.locator('[role="listbox"]').innerText()).includes('Stale result'), 'stale Game search leaked')
  state.searchFailure = true
  await input.fill('failure')
  await page.getByText('Search is unavailable. Your selection and comparison are preserved.').waitFor()
  assert.deepEqual(await columnIDs(page), [2,1])
  state.searchFailure = false
  state.directoryFailure = true
  await page.goto(base + pathFor('site', 'en', '2,1'), { waitUntil: 'networkidle' })
  assert.deepEqual(await columnIDs(page), [2,1])
  await page.getByRole('combobox').focus()
  assert(await page.getByText('Search is unavailable. Your selection and comparison are preserved.').isVisible())
  state.directoryFailure = false
  state.compareFailure = true
  for (const domain of ['site', 'game']) {
    await page.goto(base + pathFor(domain, 'en', '2,1'), { waitUntil: 'networkidle' })
    assert.equal(await page.locator('[data-compare-status="error"]').count(), 1)
    assert.deepEqual(await selectedIDs(page), [2,1])
    assert.equal(new URL(page.url()).searchParams.get('ids'), '2,1')
  }
  state.compareFailure = false
  state.insufficient = true
  for (const domain of ['site', 'game']) {
    await page.goto(base + pathFor(domain, 'en', '2,1'), { waitUntil: 'networkidle' })
    assert.equal(await page.locator('[data-compare-status="insufficient_data"]').count(), 1)
    assert.deepEqual(await selectedIDs(page), [2,1])
    assert.equal(await page.locator('[data-compare-result]').count(), 0)
  }
  state.insufficient = false
  state.reverse = true
  for (const domain of ['site', 'game']) {
    await page.goto(base + pathFor(domain, 'en', '2,1,4,3'), { waitUntil: 'networkidle' })
    assert.deepEqual(await columnIDs(page), [2,1,4,3], 'matrix followed response order instead of URL')
  }
  state.reverse = false
}
