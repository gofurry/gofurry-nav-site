import assert from 'node:assert/strict'
import { readFile, mkdtemp } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { launchPerfBrowser, rootDir } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'

// Real development CDN acceptance against the built production Nuxt application.
// API rows are isolated fixtures; all asset/probe responses come from real clouds.
// Failure scenarios abort selected CDN requests, never fulfill them with fake bytes.
const origins = { primary: process.env.NUXT_PUBLIC_ASSET_PRIMARY_BASE, mirror: process.env.NUXT_PUBLIC_ASSET_MIRROR_BASE }
for (const origin of Object.values(origins)) assert(/^https:\/\/assets-dev\.(?:go-furry|gofurry)\.com$/.test(origin || ''), 'Requires explicit development public origins (Nav Web ignored .env)')
const iconKey = 'nav/sites/9223372036854775000/icon/0bf88ea98db8746c4a3d2f917ae848d3.svg'
const desktopKey = 'nav/hero/desktop/6ff03891bcd930977a0d80a1f4ade3b6.avif'
const mobileKey = desktopKey.replace('/desktop/', '/mobile/')
const patternKey = 'nav/patterns/e3f8e3ed70a58cfdf4625b74f1e19563.svg'
const preference = { version: 1, source: 'server', pattern_id: '1', overrides: {} }
const site = { id: '9223372036854775000', name: 'Managed asset acceptance', domain: 'example.com', info: 'Real development CDN icon', icon: iconKey, country: null, nsfw: '0', welfare: '0', view_count: 0, create_time: '', update_time: '' }
let patternsEnabled = true, mobileEnabled = true
let patternSize = 120
const app = await startInsightsFixtureApp(async (url) => {
  if (url.pathname === '/api/v2/nav/home') return { data: { schema_version: 4, generated_at: new Date().toISOString(), cache_state: {}, groups: [{ id: '1', name: 'Acceptance', info: '', priority: 1, sites: [site] }], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] }, saying: null, ping: {}, hero: { desktop: { id: '1', object_key: desktopKey }, mobile: mobileEnabled ? { id: '2', object_key: mobileKey } : null } } }
  if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: patternsEnabled ? Array.from({ length: 7 }, (_, i) => ({ id: String(i + 1), name: `云端图案 ${i + 1}`, name_en: `Cloud pattern ${i + 1}`, object_key: patternKey, light_color: '#123456', dark_color: '#abcdef', light_opacity: .12, dark_opacity: .08, default_size_px: patternSize })) : [] } }
  return { data: [] }
}, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: origins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: origins.mirror })
const browser = await launchPerfBrowser()
// The dev buckets allow the developer's real Nuxt origin. Route only that local
// origin to this isolated build, leaving all CDN traffic untouched. This avoids
// stopping an already-running workstation frontend or widening cloud CORS rules.
const browserBase = 'http://localhost:3000'
const artifactDir = await mkdtemp(join(tmpdir(), 'gofurry-assets-browser-'))
let currentPage
try {
  for (const provider of ['primary', 'mirror']) {
    const response = await fetch(app.base + '/', { headers: provider === 'mirror' ? { Cookie: 'gf_asset_cdn=mirror' } : {} })
    const html = await response.text()
    assert.equal(response.status, 200)
    assert(html.includes(origins[provider] + '/' + desktopKey) && html.includes(origins[provider] + '/' + mobileKey), `SSR must select ${provider} without a probe`)
    assert(!html.includes('?probe='), 'SSR initiated a browser probe')
    assert(html.includes('data-pattern-status="default"'), 'SSR must use bundled page background')
    assert.match(html, /<h1[^>]*>/)
  }
  console.log('[assets real] SSR Primary default / Mirror cookie, independent pools, bundled background and H1 PASS')

  async function newPage({ blocked = [], mobile = false, cookie = true } = {}) {
    const context = await browser.newContext({ viewport: mobile ? { width: 390, height: 844 } : { width: 1440, height: 1000 }, locale: 'zh-CN' })
    await context.route(browserBase + '/**', async route => {
      const response = await route.fetch({ url: app.base + route.request().url().slice(browserBase.length) })
      await route.fulfill({ response })
    })
    if (cookie) await context.addCookies([{ name: 'gf_asset_cdn', value: 'primary', url: browserBase }])
    await context.addInitScript(({ pref, base }) => { if (window.top === window && location.origin === base && !localStorage.getItem('gf_background_preference')) localStorage.setItem('gf_background_preference', JSON.stringify(pref)) }, { pref: preference, base: browserBase })
    const page = await context.newPage(); currentPage = page
    await page.route('**/*', route => {
      const url = new URL(route.request().url())
      if (url.origin === browserBase) return route.fallback()
      if (Object.values(origins).includes(url.origin) || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      // Unrelated remote fonts/links are outside this asset acceptance scope.
      return route.abort()
    })
    const probes = [], assets = [], errors = [], mutations = [], failures = []
    page.on('requestfailed', request => { if (request.url().includes('assets-dev.')) failures.push({ url: request.url(), error: request.failure()?.errorText }) })
    page.on('pageerror', error => errors.push(error.message))
    page.on('console', message => { if (message.type() === 'error' && message.text().includes('CORS')) failures.push({ cors: message.text() }) })
    page.on('request', request => { if (!['GET', 'HEAD', 'OPTIONS'].includes(request.method())) mutations.push(request.url()) })
    page.on('response', response => {
      if (response.url().includes('/system/probes/cdn.bin')) probes.push(response)
      else if (Object.values(origins).some(origin => response.url().startsWith(origin + '/nav/'))) assets.push(response)
    })
    if (blocked.length) await page.route(/https:\/\/assets-dev\..+\/nav\//, route => blocked.some(provider => route.request().url().startsWith(origins[provider])) ? route.abort('failed') : route.continue())
    await page.goto(browserBase + '/', { waitUntil: 'domcontentloaded' })
    await page.waitForSelector('.nav-header__background--managed')
    return { page, context, probes, assets, errors, mutations, failures }
  }
  async function settleTab(page, index) {
    await page.waitForFunction(index => {
      const pages = document.querySelector('.preferences-pages')
      return pages && Math.abs(pages.scrollLeft - pages.clientWidth * index) < 2 && document.querySelectorAll('.preferences-tabs [role="tab"]')[index]?.getAttribute('aria-selected') === 'true'
    }, index)
  }
  const background = '.gf-public-background__pattern'
  const hero = '.nav-header__background--managed'
  async function closeContext(context) { await context.unrouteAll({ behavior: 'wait' }); await context.close() }
  async function loadedImage(page, locator) {
    await locator.scrollIntoViewIfNeeded()
    await page.waitForFunction(el => el instanceof HTMLImageElement && el.complete && el.naturalWidth > 0, await locator.elementHandle(), { timeout: 15000 })
  }
  async function assertAssetReads(state, provider, keys) {
    await state.page.waitForFunction(() => document.querySelector('[data-pattern-status="server"]'), null, { timeout: 15000 })
    for (const key of keys) {
      const response = state.assets.find(response => response.url() === origins[provider] + '/' + key)
      assert(response?.ok(), `${provider}/${key} was not fetched successfully from the real CDN`)
      assert((await response.body()).length > 0)
    }
  }

  const normal = await newPage({ cookie: false })
  await normal.page.waitForFunction(() => localStorage.getItem('gf_asset_cdn_diagnostics'), null, { timeout: 20000 })
  const diagnostics = await normal.page.evaluate(() => JSON.parse(localStorage.getItem('gf_asset_cdn_diagnostics')))
  assert(diagnostics.primaryMs !== null && diagnostics.mirrorMs !== null, 'Both real CDN CORS probes must complete, including the full-body hash: ' + JSON.stringify({ diagnostics, failures: normal.failures }))
  assert.equal(normal.probes.length, 4, 'two parallel rounds, no extra probes')
  for (const response of normal.probes) assert.equal((await response.body()).length, 8192)
  const cookie = (await normal.context.cookies()).find(c => c.name === 'gf_asset_cdn')
  assert(cookie && cookie.expires - Date.now() / 1000 > 43000 && cookie.expires - Date.now() / 1000 <= 43200, 'CDN preference needs a 12-hour cookie')
  await normal.page.mouse.wheel(0, 950)
  const icon = normal.page.locator('.nav-site-card__logo img').first()
  await loadedImage(normal.page, icon)
  await normal.page.waitForFunction(selector => getComputedStyle(document.querySelector(selector)).maskImage.includes('/nav/patterns/'), background)
  // A tiny probe finishing does not imply that the original Hero request has
  // finished; route selection now deliberately leaves that request unchanged.
  if (!normal.assets.some(r => r.url().endsWith('/' + desktopKey) && r.ok())) {
    await normal.page.waitForResponse(response => response.url().endsWith('/' + desktopKey) && response.ok(), { timeout: 15000 })
  }
  assert(normal.assets.some(r => r.url().endsWith('/' + desktopKey) && r.ok()))
  assert(normal.assets.some(r => r.url().endsWith('/' + patternKey) && r.ok()))
  assert.equal(normal.errors.length, 0, normal.errors.join('\n'))
  console.log('[assets real] Browser real icon/Hero/SVG, 4 full-body CORS probes, 12-hour cookie PASS', diagnostics)
  patternSize = 144
  await normal.page.evaluate(() => window.dispatchEvent(new Event('focus')))
  await normal.page.waitForFunction(selector => Number.parseFloat(getComputedStyle(document.querySelector(selector)).maskSize) === 144, background)
  assert.deepEqual(await normal.page.evaluate(() => JSON.parse(localStorage.getItem('gf_background_preference')).overrides), {}, 'server defaults were frozen into local overrides')

  // Use the actual preference editor: local files never leave this browser.
  await normal.page.locator('.gf-nav__mode-button').click()
  await normal.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(normal.page, 1)
  const editor = normal.page.locator('[data-background-preferences]')
  await editor.locator('[data-pattern-id="1"]').waitFor()
  const storedBefore = await normal.page.evaluate(() => localStorage.getItem('gf_background_preference'))
  assert.equal(await editor.locator('select,input[type=color],input[type=number]').count(), 0)
  await editor.getByRole('button', { name: '下一组图案' }).click()
  await normal.page.waitForFunction(() => document.querySelector('.background-preferences__strip').scrollLeft > 100)
  await editor.getByRole('button', { name: '云端图案 5', exact: true }).click()
  await editor.getByRole('button', { name: '选择图案颜色', exact: true }).click()
  await editor.getByRole('button', { name: '#6b7c62', exact: true }).click()
  await editor.getByRole('textbox', { name: '平铺尺寸', exact: true }).fill('180')
  await editor.getByRole('slider', { name: '透明度', exact: true }).fill('0.25')
  await normal.page.getByRole('tab', { name: '首屏设置', exact: true }).click()
  await normal.page.getByLabel('显示模式', { exact: true }).waitFor({ state: 'visible' })
  await normal.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(normal.page, 1)
  assert.equal(await editor.getByRole('textbox', { name: '平铺尺寸', exact: true }).inputValue(), '180')
  assert.equal(await editor.getByRole('textbox', { name: '图案颜色', exact: true }).inputValue(), '#6b7c62')
  assert.equal(await editor.locator('[data-pattern-id="5"]').getAttribute('aria-pressed'), 'true')
  await normal.page.locator('.gf-modal__header-actions .gf-button--ghost').click()
  assert.equal(await normal.page.evaluate(() => localStorage.getItem('gf_background_preference')), storedBefore, 'Cancel persisted draft changes')
  await normal.page.locator('.gf-nav__mode-button').click()
  assert.equal(await normal.page.getByRole('tab', { name: '首屏设置', exact: true }).getAttribute('aria-selected'), 'true')
  await normal.page.getByRole('tab', { name: '首屏设置', exact: true }).focus()
  await normal.page.keyboard.press('ArrowRight')
  await settleTab(normal.page, 1)
  await editor.getByRole('textbox', { name: '平铺尺寸', exact: true }).waitFor({ state: 'visible' })
  assert.equal(await editor.getByRole('textbox', { name: '平铺尺寸', exact: true }).inputValue(), String(patternSize))
  const boxes = await editor.locator('.background-preferences__controls > .background-preferences__field').evaluateAll(els => els.map(el => ({top:el.getBoundingClientRect().top,bottom:el.getBoundingClientRect().bottom})))
  assert(boxes.every(box => Math.abs(box.top - boxes[0].top) < 1 && Math.abs(box.bottom - boxes[0].bottom) < 1), 'Appearance controls are misaligned')
  await editor.getByRole('textbox', { name: '平铺尺寸', exact: true }).fill('0')
  await normal.page.locator('.gf-modal__header-actions .gf-button--primary').click()
  await editor.getByRole('alert').waitFor()
  await editor.getByRole('button', { name: '使用图案默认外观' }).click()
  assert.equal(await editor.getByRole('alert').count(), 0)
  await editor.evaluate(el => el.closest('.preferences-page').scrollTo(0, 0))
  console.log('[preferences] underline tabs, keyboard navigation, carousel selection, custom color, aligned controls, draft retention/cancel and size validation PASS')
  await normal.page.screenshot({ path: join(artifactDir, 'server-pattern-preview.png'), fullPage: false })
  await editor.getByRole('button', { name: '本地图片', exact: true }).click()
  await editor.locator('input[type=file]').setInputFiles({ name: 'local.svg', mimeType: 'image/svg+xml', buffer: Buffer.from('<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><circle cx="12" cy="12" r="5"/></svg>') })
  await editor.getByText('local.svg', { exact: true }).waitFor()
  assert.equal(await editor.getByRole('textbox', { name: '图案颜色', exact: true }).count(), 1)
  await editor.locator('input[type=range]').fill('0')
  await normal.page.locator('.gf-modal__header-actions .gf-button--primary').click()
  await normal.page.waitForSelector('[data-pattern-status="local"]')
  assert.equal(await normal.page.locator(background).evaluate(el => getComputedStyle(el).opacity), '0')
  await normal.page.reload({ waitUntil: 'domcontentloaded' })
  await normal.page.waitForSelector('[data-pattern-status="local"]')
  await normal.page.locator('.gf-nav__mode-button').click()
  await normal.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(normal.page, 1)
  const reopened = normal.page.locator('[data-background-preferences]')
  await reopened.locator('input[type=file]').setInputFiles({ name: 'local.png', mimeType: 'image/png', buffer: await readFile(join(rootDir, 'public/logo-mini.png')) })
  await reopened.getByText('local.png', { exact: true }).waitFor()
  assert.equal(await reopened.getByRole('textbox', { name: '图案颜色', exact: true }).count(), 0, 'raster images must hide SVG color controls')
  await normal.page.locator('.gf-modal__header-actions .gf-button--primary').click()
  await normal.page.waitForFunction(selector => getComputedStyle(document.querySelector(selector)).maskImage === 'none' && getComputedStyle(document.querySelector(selector)).backgroundImage.includes('blob:'), background)
  await normal.page.locator('.gf-nav__mode-button').click()
  await normal.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(normal.page, 1)
  await normal.page.getByRole('button', { name: '清除本地背景', exact: true }).click()
  await normal.page.locator('.gf-modal__header-actions .gf-button--primary').click()
  await normal.page.waitForSelector('[data-pattern-status="default"]')
  assert.deepEqual(await normal.page.evaluate(() => JSON.parse(localStorage.getItem('gf_background_preference'))), { version: 1, source: 'default', overrides: {} })
  assert.deepEqual(normal.mutations, [], 'Local background UI performed a network mutation')
  await normal.page.locator('.gf-nav__mode-button').click()
  await normal.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(normal.page, 1)
  await normal.page.screenshot({ path: join(artifactDir, 'background-preferences.png'), fullPage: false })
  await closeContext(normal.context)
  console.log('[assets real] Actual Vue local SVG/raster editor, zero opacity, reload persistence, clearing and no upload PASS')

  const failPrimary = await newPage({ blocked: ['primary'] })
  await failPrimary.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.includes('assets-dev.gofurry.com'), hero, { timeout: 15000 })
  await failPrimary.page.mouse.wheel(0, 950)
  const mirrorIcon = failPrimary.page.locator('.nav-site-card__logo img').first()
  await loadedImage(failPrimary.page, mirrorIcon)
  assert((await mirrorIcon.getAttribute('src')).startsWith(origins.mirror))
  await assertAssetReads(failPrimary, 'mirror', [desktopKey, patternKey, iconKey])
  await closeContext(failPrimary.context)
  console.log('[assets real] Primary failure -> real R2 CDN Hero / pattern / icon reads PASS')

  const failBoth = await newPage({ blocked: ['primary', 'mirror'] })
  await failBoth.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.startsWith('data:'), hero, { timeout: 15000 })
  await failBoth.page.waitForSelector('[data-pattern-status="default"]')
  await failBoth.page.mouse.wheel(0, 950)
  const defaultIcon = failBoth.page.locator('.nav-site-card__logo img').first()
  await loadedImage(failBoth.page, defaultIcon)
  assert.equal(await defaultIcon.getAttribute('src'), '/defaultLogo.svg')
  assert((await failBoth.page.locator(background).evaluate(el => getComputedStyle(el).maskImage)).includes('/web/background/gofurry-pattern.svg'))
  await closeContext(failBoth.context)
  console.log('[assets real] Both CDNs fail -> bundled icon / no Hero / bundled pattern PASS')

  const mobile = await newPage({ mobile: true })
  await mobile.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.includes('/mobile/'), hero)
  if (!mobile.assets.some(r => r.url().endsWith('/' + mobileKey) && r.ok())) await mobile.page.waitForResponse(response => response.url().endsWith('/' + mobileKey) && response.ok(), { timeout: 15000 })
  assert(!mobile.assets.some(r => r.url().endsWith('/' + desktopKey)), 'mobile fetched desktop pool')
  await mobile.page.waitForSelector('[data-pattern-status="server"]')
  await mobile.page.evaluate(() => window.scrollTo(0, 450))
  await mobile.page.waitForFunction(() => document.querySelector('.mobile-bottom-tabs')?.getAttribute('aria-hidden') === 'false')
  await mobile.page.locator('.mobile-bottom-tabs button').click()
  await mobile.page.getByRole('tab', { name: '首屏设置', exact: true }).waitFor()
  await mobile.page.screenshot({ path: join(artifactDir, 'preferences-mobile-home.png') })
  await mobile.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(mobile.page, 1)
  await mobile.page.locator('[data-pattern-id="1"]').waitFor()
  await mobile.page.screenshot({ path: join(artifactDir, 'preferences-mobile-background.png') })
  const overflow = await mobile.page.locator('.preferences-page').nth(1).evaluate(el => el.scrollWidth > el.clientWidth + 1)
  assert.equal(overflow, false, 'Mobile panel overflows horizontally')
  await mobile.page.locator('.gf-modal__header-actions .gf-button--ghost').click()
  await mobile.page.evaluate(() => localStorage.setItem('theme', 'dark'))
  await mobile.page.reload({ waitUntil: 'domcontentloaded' })
  await mobile.page.waitForSelector('[data-pattern-status="server"]')
  await mobile.page.evaluate(() => window.scrollTo(0, 450))
  await mobile.page.waitForFunction(() => document.querySelector('.mobile-bottom-tabs')?.getAttribute('aria-hidden') === 'false')
  await mobile.page.locator('.mobile-bottom-tabs button').click()
  await mobile.page.getByRole('tab', { name: '页面背景', exact: true }).click()
  await settleTab(mobile.page, 1)
  await mobile.page.locator('[data-pattern-id="1"]').waitFor()
  await mobile.page.screenshot({ path: join(artifactDir, 'preferences-mobile-dark.png') })
  assert.equal(await mobile.page.locator('html').evaluate(el => el.classList.contains('dark')), true)
  await mobile.page.locator('.preferences-pages').evaluate(el => el.scrollTo({ left: 0, behavior: 'instant' }))
  await mobile.page.waitForFunction(() => document.querySelector('[role="tab"][aria-selected="true"]')?.textContent.includes('首屏设置'))
  await closeContext(mobile.context)
  mobileEnabled = false; patternsEnabled = false
  const emptyMobile = await newPage({ mobile: true })
  await emptyMobile.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.startsWith('data:'), hero)
  await emptyMobile.page.waitForSelector('[data-pattern-status="default"]')
  assert(!emptyMobile.assets.some(r => r.url().includes('/nav/hero/')), 'empty mobile pool borrowed desktop')
  // Explicit CORS GET must work for each provider, even if fallback hid an
  // infrastructure error during the visual checks above.
  for (const origin of Object.values(origins)) {
    const cors = await emptyMobile.page.evaluate(async url => {
      try { const response = await fetch(url, { mode: 'cors', credentials: 'omit' }); return { status: response.status, length: (await response.arrayBuffer()).byteLength } }
      catch (error) { return { error: String(error) } }
    }, origin + '/' + patternKey)
    assert(cors.status === 200 && cors.length > 0, `${origin}: SVG must support real cross-origin CSS masks: ${JSON.stringify(cors)}`)
  }
  await closeContext(emptyMobile.context)
  console.log('[assets real] Mobile real AVIF, empty pool isolation and disabled catalog fallback PASS')
  console.log('Real development CDN browser acceptance passed. Screenshot:', join(artifactDir, 'background-preferences.png'))
} catch (error) {
  if (currentPage && !currentPage.isClosed()) console.error('Background state:', await currentPage.evaluate(() => ({ preference: localStorage.getItem('gf_background_preference'), style: document.querySelector('.gf-public-background__pattern')?.getAttribute('style'), maskSize: document.querySelector('.gf-public-background__pattern') ? getComputedStyle(document.querySelector('.gf-public-background__pattern')).maskSize : null })))
  if (currentPage && !currentPage.isClosed()) await currentPage.screenshot({ path: join(artifactDir, 'failure.png'), fullPage: false }).catch(() => {})
  console.error('Browser artifacts:', artifactDir, '\nNuxt:', app.logs())
  throw error
} finally {
  for (const context of browser.contexts()) await context.unrouteAll({ behavior: 'ignoreErrors' })
  await browser.close(); await app.close()
}
