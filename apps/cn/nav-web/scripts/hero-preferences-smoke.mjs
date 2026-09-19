import assert from 'node:assert/strict'
import { mkdir, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { launchPerfBrowser, reportsDir } from './perf/shared.mjs'

const origins = { primary: 'https://primary.example', mirror: 'https://mirror.example' }
const makeHero = (id, variant, hash = String(BigInt(id) % 10n)) => ({ id, name: `${variant} ${id}`, object_key: `nav/hero/${variant}/${hash.repeat(32)}.avif` })
const desktop = Array.from({ length: 48 }, (_, i) => makeHero(String(i + 10), 'desktop', (i + 10).toString(16).padStart(32, '0').slice(-1)))
// Unique keys across all pages, with no dependence on numeric ID conversion.
desktop.forEach((item, i) => { item.object_key = `nav/hero/desktop/${String(i + 10).padStart(32, '0')}.avif` })
const mobile = [makeHero('9007199254740993', 'mobile'), makeHero('9007199254740994', 'mobile')]
let heroGate = null, failedCatalogPage = 0
const resolveHero = url => {
  if (url.searchParams.get('hero_mode') === 'local') return { desktop: null, mobile: null }
  return { desktop: desktop.find(item => item.id === url.searchParams.get('hero_desktop_id')) ?? desktop[0], mobile: mobile.find(item => item.id === url.searchParams.get('hero_mobile_id')) ?? mobile[0] }
}
const app = await startInsightsFixtureApp(async url => {
  if (url.pathname === '/api/v2/nav/home') return { data: { schema_version: 4, groups: [], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] }, saying: null, ping: {}, hero: resolveHero(url) } }
  if (url.pathname === '/api/v2/nav/home/hero') { if (heroGate) await heroGate; return { data: { hero: resolveHero(url) } } }
  if (url.pathname === '/api/v2/nav/appearance/heroes') {
    const variant = url.searchParams.get('variant'), items = variant === 'desktop' ? desktop : mobile
    const page = Number(url.searchParams.get('page_num')), size = Number(url.searchParams.get('page_size'))
    if (page === failedCatalogPage) return { status: 503 }
    return { data: { schema_version: 1, variant, page_num: page, page_size: size, total: items.length, items: items.slice((page - 1) * size, page * size), selected: items.find(item => item.id === url.searchParams.get('selected_id')) ?? null } }
  }
  if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: [] } }
  return { data: [] }
}, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: origins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: origins.mirror })
const browser = await launchPerfBrowser()
const output = join(reportsDir, 'hero-preferences')
await mkdir(output, { recursive: true })
const errors = []
const cookieValue = async (context, name) => (await context.cookies()).find(cookie => cookie.name === name)?.value
const rendered = '.nav-header__background:not([data-hero-pending])'
const current = page => page.locator(rendered).evaluate(el => el.currentSrc || el.querySelector('img')?.currentSrc || '')
const heroCalls = () => app.requests.filter(url => url.pathname === '/api/v2/nav/home/hero')
const homeCalls = () => app.requests.filter(url => url.pathname === '/api/v2/nav/home')
const catalogCalls = () => app.requests.filter(url => url.pathname === '/api/v2/nav/appearance/heroes')
async function open(page) {
  const desktopButton = page.locator('.gf-nav__mode-button')
  if (await desktopButton.isVisible()) await desktopButton.click()
  else {
    await page.locator('.gf-nav__mobile-toggle').click()
    await page.locator('.gf-nav__mobile-action').click()
  }
  await page.locator('[data-hero-preferences]').waitFor()
}
async function cancel(page) { await page.locator('.gf-modal__header-actions .gf-button--ghost').click(); await page.locator('[data-hero-preferences]').waitFor({ state: 'detached' }) }
async function save(page) { await page.locator('.gf-modal__header-actions .gf-button--primary').click(); await page.locator('[data-hero-preferences]').waitFor({ state: 'detached' }) }
async function close(context) { await context.unrouteAll({ behavior: 'wait' }); await context.close() }
async function waitPosition(page, expected) {
  await page.waitForFunction(expected => {
    const picker = [...document.querySelectorAll('[data-hero-catalog]')].find(el => el.closest('[role="tabpanel"]').style.display !== 'none')
    return picker?.getAttribute('aria-busy') === 'false' && Number.parseInt(picker.querySelector('[data-hero-position]').textContent) === expected
  }, expected)
}
async function browseTo(page, target) {
  const picker = page.locator('[data-hero-catalog]:visible')
  await page.waitForFunction(() => [...document.querySelectorAll('[data-hero-catalog]')].some(el => el.closest('[role="tabpanel"]').style.display !== 'none' && el.getAttribute('aria-busy') === 'false' && Number.parseInt(el.querySelector('[data-hero-position]').textContent) > 0))
  let position = Number.parseInt(await picker.locator('[data-hero-position]').textContent())
  while (position !== target) {
    const direction = target > position ? 1 : -1
    await picker.getByRole('button', { name: direction > 0 ? '下一张背景' : '上一张背景', exact: true }).click()
    position += direction
    await waitPosition(page, position)
  }
}
async function setup({ width = 1440, mode = 'random', desktopId = null, mobileId = null, local = false } = {}) {
  app.requests.length = 0
  const context = await browser.newContext({ viewport: { width, height: 1000 }, locale: 'zh-CN', reducedMotion: 'reduce' })
  const cookies = [['gf_asset_cdn', 'primary'], ['gf_steam_asset_mode', 'global'], ['gf_asset_cdn_mode', 'primary']]
  if (mode) cookies.push(['gf_hero_mode', mode])
  if (desktopId) cookies.push(['gf_hero_desktop_id', desktopId])
  if (mobileId) cookies.push(['gf_hero_mobile_id', mobileId])
  await context.addCookies(cookies.map(([name, value]) => ({ name, value, url: app.base })))
  await context.addInitScript(({ base, local }) => {
    if (location.origin !== base) return
    localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt: Date.now(), primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
    new MutationObserver(() => { window.initialHeroNode ||= document.querySelector('.nav-header__background') }).observe(document, { subtree: true, childList: true })
    if (local) {
      localStorage.setItem('customNavHeaderBgFolderName', 'Saved folder')
      window.localSeed = new Promise((resolve, reject) => {
        const req = indexedDB.open('gofurry-custom-nav-header-bg', 1)
        req.onupgradeneeded = () => req.result.createObjectStore('directoryHandles')
        req.onerror = () => reject(req.error)
        req.onsuccess = () => {
          const db = req.result, tx = db.transaction('directoryHandles', 'readwrite')
          tx.objectStore('directoryHandles').put(new Blob(['<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40"><rect width="40" height="40" fill="#226633"/></svg>'], { type: 'image/svg+xml' }), 'nav-header-bg-cache')
          tx.oncomplete = () => { db.close(); resolve() }
        }
      })
    }
  }, { base: app.base, local })
  const images = [], control = { imageGate: null, fail: '' }
  await context.route('**/*', async route => {
    const url = new URL(route.request().url())
    if (url.origin === app.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
    if (Object.values(origins).includes(url.origin) && url.pathname.startsWith('/nav/hero/')) {
      images.push(url.href)
      if (control.imageGate) await control.imageGate
      if (control.fail && url.href.includes(control.fail)) return route.abort()
      // Local artwork with real viewport proportions for visual review only.
      const height = url.pathname.includes('/mobile/') ? 1200 : 506
      const colors = ['#6b8277', '#7c8c95', '#88826a']
      const color = colors[[...url.pathname].reduce((sum, char) => sum + char.charCodeAt(0), 0) % colors.length]
      return route.fulfill({ contentType: 'image/svg+xml', body: `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="${height}" viewBox="0 0 900 ${height}"><rect width="900" height="${height}" fill="#e5d5b6"/><circle cx="660" cy="${height * .28}" r="72" fill="#faf1d9"/><path d="M0 ${height}V${height * .72}L270 ${height * .32}L560 ${height}Z" fill="${color}"/><path d="M230 ${height}L620 ${height * .5}L900 ${height * .76}V${height}Z" fill="#475f59"/></svg>` })
    }
    return route.abort()
  })
  const page = await context.newPage()
  page.on('pageerror', error => errors.push(error.message))
  const response = await page.goto(app.base, { waitUntil: 'networkidle' })
  await page.waitForFunction(() => Boolean(document.querySelector('#__nuxt')?.__vue_app__))
  return { context, page, images, control, html: await response.text() }
}

try {
  // Record real computed input/toggle states alongside the existing Hero shots.
  // This evidence can be compared across style-only migrations without baking
  // a second palette into the test or introducing a screenshot framework.
  const foundation = {}
  for (const width of [1440, 390]) {
    const { page, context } = await setup({ width })
    for (const dark of [false, true]) {
      await page.evaluate(dark => document.documentElement.classList.toggle('dark', dark), dark)
      await open(page)
      const key = `${width}-${dark ? 'dark' : 'light'}`
      const capture = async state => {
        await page.waitForTimeout(300)
        foundation[`${key}-${state}`] = await page.evaluate(() => {
          const properties = ['display', 'width', 'height', 'padding', 'gap', 'border', 'borderRadius', 'backgroundColor', 'backgroundImage', 'boxShadow', 'color', 'fontSize', 'fontWeight', 'lineHeight', 'transform', 'outline', 'outlineOffset', 'flexShrink']
          const selectors = ['.gf-preferences-modal', '.gf-modal__header', '#mode-setting-input', '#quick-access-toggle', '#quick-access-toggle span', '.gf-modal__header-actions .gf-button--primary', '.preferences-sources button[aria-pressed="true"]']
          return Object.fromEntries(selectors.map(selector => {
            const style = getComputedStyle(document.querySelector(selector))
            return [selector, Object.fromEntries(properties.map(property => [property, style[property]]))]
          }))
        })
      }
      const toggle = page.locator('#quick-access-toggle')
      assert.equal(await toggle.getAttribute('aria-pressed'), 'true')
      await page.locator('#mode-setting-input').focus()
      await capture('input-focus-toggle-on')
      await page.screenshot({ path: join(output, `foundation-${key}-on.png`) })
      await toggle.focus()
      await page.keyboard.press('Space')
      assert.equal(await toggle.getAttribute('aria-pressed'), 'false')
      assert(await toggle.evaluate(el => el === document.activeElement), 'toggle lost keyboard focus')
      await capture('input-idle-toggle-off')
      await page.screenshot({ path: join(output, `foundation-${key}-off.png`) })
      assert.equal(await page.locator('.preferences-page').first().evaluate(el => el.scrollWidth > el.clientWidth + 1), false)
      await cancel(page)
      await open(page)
      assert.equal(await toggle.getAttribute('aria-pressed'), 'true', 'Cancel applied the toggle draft')
      await toggle.click()
      await save(page)
      assert.equal(await page.evaluate(() => localStorage.getItem('nav-header-show-quick-access')), '0')
      await open(page)
      assert.equal(await toggle.getAttribute('aria-pressed'), 'false', 'Save lost the toggle draft')
      await toggle.focus()
      await page.keyboard.press('Enter')
      assert.equal(await toggle.getAttribute('aria-pressed'), 'true')
      await save(page)
    }
    await close(context)
  }
  await writeFile(join(output, 'foundation-computed.json'), JSON.stringify(foundation, null, 2))
  console.log('[preferences] input idle/focus, toggle keyboard/ARIA, Save/Cancel, light/dark desktop/mobile PASS')

  for (const width of [1440, 390]) {
    const state = await setup({ width, mode: 'fixed', desktopId: '34', mobileId: mobile[1].id, local: true })
    const expected = width >= 768 ? desktop[24] : mobile[1]
    assert(state.html.includes(desktop[24].object_key) && state.html.includes(mobile[1].object_key), 'Fixed SSR did not resolve both ID cookies')
    assert(!state.html.includes(desktop[0].object_key), 'SSR rendered a random Hero before Fixed')
    assert.equal(await current(state.page), origins.primary + '/' + expected.object_key)
    assert.deepEqual(state.images, [origins.primary + '/' + expected.object_key], 'hydration or saved folder replaced Fixed Hero')
    assert(await state.page.evaluate(() => window.initialHeroNode === document.querySelector('.nav-header__background')), 'Fixed hydration replaced SSR element')
    assert.equal(homeCalls().length, 1)
    assert.equal(heroCalls().length, 0)
    await state.page.reload({ waitUntil: 'networkidle' })
    assert.equal(await current(state.page), origins.primary + '/' + expected.object_key)
    assert.equal(await cookieValue(state.context, 'gf_hero_mobile_id'), mobile[1].id)
    await close(state.context)
  }
  console.log('[hero preferences] Fixed SSR/hydration, independent bigint pins, reload and saved-folder isolation PASS')

  const previousKey = desktop[24].object_key
  desktop[24].object_key = 'nav/hero/desktop/' + 'f'.repeat(32) + '.avif'
  const replaced = await setup({ mode: 'fixed', desktopId: '34' })
  assert(replaced.html.includes(desktop[24].object_key) && !replaced.html.includes(previousKey), 'ID persisted an old object_key')
  await close(replaced.context)
  const invalid = await setup({ mode: 'fixed', desktopId: '999', mobileId: mobile[1].id })
  assert.equal(await current(invalid.page), origins.primary + '/' + desktop[0].object_key)
  assert(invalid.html.includes(mobile[1].object_key), 'invalid desktop pin lost mobile selection')
  await close(invalid.context)

  const local = await setup({ mode: 'local', local: true })
  assert(!local.html.includes('/nav/hero/'), 'Local SSR exposed cloud Hero URLs')
  await local.page.waitForFunction(selector => document.querySelector(selector)?.currentSrc?.startsWith('blob:'), rendered)
  assert.equal(local.images.length, 0, 'Local mode fetched cloud Hero')
  assert.equal(homeCalls()[0].searchParams.get('hero_mode'), 'local')
  assert.equal(heroCalls().length, 0)
  await open(local.page)
  await local.page.locator('[data-hero-preferences] .gf-modal__actions button').last().click()
  await cancel(local.page)
  assert((await current(local.page)).startsWith('blob:'))
  assert.equal(await cookieValue(local.context, 'gf_hero_mode'), 'local')
  assert.equal(await local.page.evaluate(() => localStorage.getItem('customNavHeaderBgFolderName')), 'Saved folder', 'Cancel deleted the local folder metadata')
  await open(local.page)
  await local.page.getByRole('button', { name: '云端随机', exact: true }).click()
  await save(local.page)
  await local.page.waitForFunction(selector => document.querySelector(selector)?.querySelector('img')?.currentSrc?.includes('primary.example'), rendered)
  assert.equal(heroCalls().length, 1)
  await local.page.reload({ waitUntil: 'networkidle' })
  assert.equal(await current(local.page), origins.primary + '/' + desktop[0].object_key, 'saved folder overrode explicit Random')
  await close(local.context)
  const missingLocal = await setup({ mode: 'local' })
  assert(!missingLocal.html.includes('/nav/hero/'))
  await missingLocal.page.waitForFunction(selector => document.querySelector(selector)?.querySelector('img')?.currentSrc?.includes('primary.example'), rendered)
  assert.equal(await cookieValue(missingLocal.context, 'gf_hero_mode'), 'random')
  assert.equal(heroCalls().length, 1, 'invalid Local retried more than once')
  await close(missingLocal.context)
  const legacy = await setup({ mode: null, local: true })
  await legacy.page.waitForFunction(selector => document.querySelector(selector)?.currentSrc?.startsWith('blob:'), rendered)
  assert.equal(await cookieValue(legacy.context, 'gf_hero_mode'), 'local')
  const reload = await legacy.page.reload({ waitUntil: 'networkidle' })
  assert(!(await reload.text()).includes('/nav/hero/'), 'legacy migration was not SSR-readable on the next visit')
  assert.equal(heroCalls().length, 0, 'legacy migration reselected cloud Hero')
  await close(legacy.context)
  console.log('[hero preferences] live metadata replacement, invalid pin, Local SSR, one-time legacy migration and missing-local fallback PASS')

  const ui = await setup()
  const before = await current(ui.page)
  await open(ui.page)
  assert.equal(await ui.page.getByRole('tab').count(), 3, 'Hero source created another preferences tab')
  assert.equal(catalogCalls().length, 0, 'opening Random preferences queried the entire Hero catalog')
  await ui.page.getByRole('button', { name: '固定云端背景', exact: true }).click()
  const picker = ui.page.locator('[data-hero-catalog]:visible')
  await picker.scrollIntoViewIfNeeded()
  await waitPosition(ui.page, 1)
  assert(await picker.getByRole('button', { name: '上一张背景', exact: true }).isDisabled())
  assert.equal(await picker.locator('select').count(), 0, 'Hero still uses a dropdown')
  assert.equal(await picker.getByRole('button').count(), 2, 'Picker has duplicate selection/random actions')
  assert.equal(catalogCalls().length, 1)
  assert.equal(new Set(ui.images).size, 1, 'opening Fixed downloaded all catalog entries')
  await browseTo(ui.page, 2)
  await ui.page.waitForFunction(() => document.querySelector('[data-hero-catalog="desktop"] img')?.complete)
  assert.equal(new Set(ui.images).size, 2, 'selecting a visible preview fetched hidden previews')
  await browseTo(ui.page, 12)
  assert.equal(catalogCalls().length, 1, 'in-page browsing requested metadata again')
  const beforeBoundary = new Set(ui.images).size
  await browseTo(ui.page, 13)
  assert.equal(catalogCalls().length, 2)
  assert.equal(catalogCalls().at(-1).searchParams.get('page_num'), '2')
  assert.equal(catalogCalls().at(-1).searchParams.get('selected_id'), '21')
  await ui.page.waitForFunction(() => document.querySelector('[data-hero-catalog="desktop"] img')?.complete)
  assert.equal(new Set(ui.images).size, beforeBoundary + 1, 'page boundary eagerly fetched unseen images')
  await browseTo(ui.page, 12)
  assert.equal(catalogCalls().length, 2, 'backward page boundary did not reuse metadata')
  await ui.page.getByRole('tab', { name: '手机', exact: true }).click()
  await browseTo(ui.page, 2)
  assert.equal(await ui.page.locator('[data-hero-catalog] img').count(), 1, 'inactive viewport retained a mounted preview')
  await ui.page.getByRole('tab', { name: '桌面', exact: true }).click()
  await waitPosition(ui.page, 12)
  assert.equal(catalogCalls().length, 3, 'viewport switch lost its browsing position/page cache')
  await cancel(ui.page)
  assert.equal(await current(ui.page), before)
  assert.equal(await cookieValue(ui.context, 'gf_hero_mode'), 'random')
  assert.equal(await cookieValue(ui.context, 'gf_hero_desktop_id'), undefined)
  assert.equal(heroCalls().length, 0, 'Cancel applied a Hero')

  await open(ui.page)
  await ui.page.getByRole('button', { name: '固定云端背景', exact: true }).click()
  await browseTo(ui.page, 3)
  await ui.page.getByRole('tab', { name: '手机', exact: true }).click()
  await browseTo(ui.page, 2)
  let releaseAPI, releaseImage
  // Simulate replacing the same ID's file after preview. The staged renderer
  // must actually wait for an uncached image, not reuse the decoded preview.
  desktop[2].object_key = 'nav/hero/desktop/' + 'e'.repeat(32) + '.avif'
  heroGate = new Promise(resolve => { releaseAPI = resolve })
  ui.control.imageGate = new Promise(resolve => { releaseImage = resolve })
  await save(ui.page)
  assert.equal(await current(ui.page), before, 'Save cleared Hero while the API was pending')
  assert.equal(await cookieValue(ui.context, 'gf_hero_mode'), 'fixed')
  assert.equal(await cookieValue(ui.context, 'gf_hero_desktop_id'), '12')
  assert.equal(await cookieValue(ui.context, 'gf_hero_mobile_id'), mobile[1].id)
  releaseAPI(); heroGate = null
  await delay(250)
  assert.equal(await current(ui.page), before, 'Save cleared Hero before the new display image loaded')
  releaseImage(); ui.control.imageGate = null
  await ui.page.waitForFunction(({ selector, key }) => document.querySelector(selector)?.querySelector('img')?.currentSrc?.endsWith(key), { selector: rendered, key: desktop[2].object_key })
  assert.equal(homeCalls().length, 1, 'Save refetched unrelated Home content')
  assert.equal(heroCalls().length, 1)
  assert.equal(await cookieValue(ui.context, 'gf_asset_cdn_mode'), 'primary')
  await ui.page.reload({ waitUntil: 'networkidle' })
  assert.equal(await current(ui.page), origins.primary + '/' + desktop[2].object_key)

  // Selecting only Mobile alongside a new route must not re-route Desktop's
  // unchanged, already successful resource during the frame handoff.
  const callsBeforeMobile = heroCalls().length
  await open(ui.page)
  await ui.page.getByRole('tab', { name: '手机', exact: true }).click()
  await browseTo(ui.page, 1)
  await ui.page.locator('.preferences-tabs:not(.preferences-tabs--compact)').getByRole('tab').last().click()
  await ui.page.locator('[data-resource-routing] .resource-route').first().getByRole('radio', { name: 'Cloudflare' }).check()
  await save(ui.page)
  await delay(500)
  assert.equal(heroCalls().length, callsBeforeMobile + 1)
  assert.equal(await current(ui.page), origins.primary + '/' + desktop[2].object_key, 'Mobile Save rerouted the unchanged Desktop resource')
  await ui.page.setViewportSize({ width: 390, height: 1000 })
  await ui.page.waitForFunction(({ selector, key }) => document.querySelector(selector)?.querySelector('img')?.currentSrc === 'https://mirror.example/' + key, { selector: rendered, key: mobile[0].object_key })
  await ui.page.setViewportSize({ width: 1440, height: 1000 })

  const retainedDesktop = await current(ui.page)
  await open(ui.page)
  await browseTo(ui.page, 5)
  desktop[4].object_key = 'nav/hero/desktop/' + 'd'.repeat(32) + '.avif'
  ui.control.fail = desktop[4].object_key
  await save(ui.page)
  await ui.page.waitForFunction(() => !document.querySelector('[data-hero-pending]'))
  // Wait for both actual renderer failures, not just the API or modal close.
  for (let attempts = 0; attempts < 100 && ui.images.filter(url => url.includes(desktop[4].object_key)).length < 2; attempts++) await delay(20)
  assert.deepEqual(ui.images.filter(url => url.includes(desktop[4].object_key)), [origins.mirror + '/' + desktop[4].object_key, origins.primary + '/' + desktop[4].object_key])
  await delay(100)
  assert.equal(await current(ui.page), retainedDesktop, 'exhausted staged fallback removed a successful old Hero')
  ui.control.fail = ''

  // Locate an out-of-page pin by stable metadata order, never by loading images.
  await ui.context.addCookies([{ name: 'gf_hero_desktop_id', value: '34', url: app.base }])
  await ui.page.reload({ waitUntil: 'networkidle' })
  const catalogBefore = catalogCalls().length
  await open(ui.page)
  await picker.scrollIntoViewIfNeeded()
  await waitPosition(ui.page, 25)
  assert.equal(catalogCalls().length, catalogBefore + 2, 'out-of-page selection scanned the catalog')
  assert.equal(catalogCalls().at(-1).searchParams.get('selected_id'), '34')
  assert.equal(await ui.page.locator('[data-hero-catalog] img').count(), 1)
  const locatedImages = ui.images.length
  await browseTo(ui.page, 24)
  assert.equal(catalogCalls().at(-1).searchParams.get('page_num'), '2', 'left arrow did not request the preceding page')
  await browseTo(ui.page, 25)
  assert(ui.images.length <= locatedImages + 2, 'browsing mounted hidden images')
  const viewportTabs = ui.page.locator('.preferences-tabs--compact').getByRole('tab')
  await viewportTabs.first().focus()
  await ui.page.keyboard.press('ArrowRight')
  await waitPosition(ui.page, 1)
  assert(await viewportTabs.last().evaluate(el => el === document.activeElement))
  await ui.page.keyboard.press('Home')
  await waitPosition(ui.page, 25)
  assert(await viewportTabs.first().evaluate(el => el === document.activeElement))
  await ui.page.locator('[data-hero-preferences] h3').click()
  await ui.page.mouse.move(2, 2)
  for (const width of [1440, 390, 320]) {
    await ui.page.setViewportSize({ width, height: 1000 })
    for (const dark of [false, true]) {
      await ui.page.evaluate(dark => document.documentElement.classList.toggle('dark', dark), dark)
      await picker.scrollIntoViewIfNeeded()
      const sourceStyles = await ui.page.evaluate(() => {
        const read = el => { const s = getComputedStyle(el); return ['padding', 'borderRadius', 'borderColor', 'backgroundColor', 'color', 'fontSize', 'lineHeight'].map(key => s[key]) }
        return [read(document.querySelector('[data-hero-preferences] .preferences-sources button[aria-pressed="true"]')), read(document.querySelector('[data-background-preferences] .preferences-sources button[aria-pressed="true"]'))]
      })
      assert.deepEqual(sourceStyles[0], sourceStyles[1], 'Hero source selector differs from page background controls')
      const widths = await ui.page.locator('[data-hero-preferences] .preferences-sources button').evaluateAll(elements => elements.map(el => el.getBoundingClientRect().width))
      assert(Math.max(...widths) - Math.min(...widths) < 1, 'source options are not equal-width columns')
      assert.equal(await ui.page.locator('.preferences-page').first().evaluate(el => el.scrollWidth > el.clientWidth + 1), false, 'Hero editor overflows horizontally')
      // Finish theme transitions so evidence shows the selected theme, not an
      // intermediate input/button color from the previous theme.
      await ui.page.screenshot({ path: join(output, `${width}-${dark ? 'dark' : 'light'}.png`), animations: 'disabled' })
    }
  }
  const source = ui.page.getByRole('button', { name: '固定云端背景', exact: true })
  await ui.page.keyboard.press('Tab')
  await source.focus()
  assert.notEqual(await source.evaluate(el => getComputedStyle(el).outlineStyle), 'none', 'keyboard focus is invisible')
  await cancel(ui.page)
  await close(ui.context)

  const boundary = await setup({ mode: 'fixed', desktopId: '57' })
  await open(boundary.page)
  await waitPosition(boundary.page, 48)
  const boundaryPicker = boundary.page.locator('[data-hero-catalog]:visible')
  assert(await boundaryPicker.getByRole('button', { name: '下一张背景', exact: true }).isDisabled(), 'last image can advance beyond the catalog')
  assert.equal(catalogCalls().length, 3, 'last-page pin did not use bounded metadata lookup')
  assert.deepEqual([...new Set(boundary.images)], [origins.primary + '/' + desktop[47].object_key], 'pin lookup loaded intermediate previews')
  await browseTo(boundary.page, 47)
  await browseTo(boundary.page, 48)
  assert.equal(await cookieValue(boundary.context, 'gf_hero_desktop_id'), '57')
  await cancel(boundary.page)
  await close(boundary.context)

  const retry = await setup({ mode: 'fixed', desktopId: '21' })
  await open(retry.page)
  await waitPosition(retry.page, 12)
  failedCatalogPage = 2
  await retry.page.getByRole('button', { name: '下一张背景', exact: true }).click()
  const retryPicker = retry.page.locator('[data-hero-catalog]:visible')
  await retryPicker.getByRole('alert').waitFor()
  await waitPosition(retry.page, 12)
  assert.equal(await cookieValue(retry.context, 'gf_hero_desktop_id'), '21', 'metadata failure changed persisted selection')
  failedCatalogPage = 0
  await retryPicker.getByRole('button', { name: '重试', exact: true }).click()
  await waitPosition(retry.page, 13)
  assert.equal(await cookieValue(retry.context, 'gf_hero_desktop_id'), '21', 'browsing implicitly saved the draft')
  await cancel(retry.page)
  await close(retry.context)
  assert.deepEqual(errors, [])
  console.log('[hero preferences] lazy metadata/preview paging, Cancel, Save without blank-first, ID cookies, selected_id, responsive themes/focus PASS')
} finally {
  for (const context of browser.contexts()) await context.unrouteAll({ behavior: 'ignoreErrors' })
  await browser.close()
  await app.close()
}
