import assert from 'node:assert/strict'
import { mkdtemp } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'

// Real development CDN acceptance against the built production Nuxt application.
// API rows are isolated fixtures; all asset/probe responses come from real clouds.
// Failure scenarios abort selected CDN requests, never fulfill them with fake bytes.
const origins = { primary: process.env.NUXT_PUBLIC_ASSET_PRIMARY_BASE, mirror: process.env.NUXT_PUBLIC_ASSET_MIRROR_BASE }
for (const origin of Object.values(origins)) assert(/^https:\/\/assets-dev\.(?:go-furry|gofurry)\.com$/.test(origin || ''), 'Requires explicit development public origins (Nav Web ignored .env)')
const iconKey = 'nav/sites/9223372036854775000/icon/0bf88ea98db8746c4a3d2f917ae848d3.svg'
const desktopKey = 'nav/hero/desktop/6ff03891bcd930977a0d80a1f4ade3b6.avif'
const mobileKey = desktopKey.replace('/desktop/', '/mobile/')
const site = { id: '9223372036854775000', name: 'Managed asset acceptance', domain: 'example.com', info: 'Real development CDN icon', icon: iconKey, country: null, nsfw: '0', welfare: '0', view_count: 0, create_time: '', update_time: '' }
let mobileEnabled = true
const app = await startInsightsFixtureApp(async (url) => {
  if (url.pathname === '/api/v2/nav/home') return { data: { schema_version: 4, generated_at: new Date().toISOString(), cache_state: {}, groups: [{ id: '1', name: 'Acceptance', info: '', priority: 1, sites: [site] }], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] }, saying: null, ping: {}, hero: { desktop: { id: '1', object_key: desktopKey }, mobile: mobileEnabled ? { id: '2', object_key: mobileKey } : null } } }
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
  const background = '.gf-public-background__dots--left'
  const hero = '.nav-header__background--managed'
  async function closeContext(context) { await context.unrouteAll({ behavior: 'wait' }); await context.close() }
  async function loadedImage(page, locator) {
    await locator.scrollIntoViewIfNeeded()
    await page.waitForFunction(el => el instanceof HTMLImageElement && el.complete && el.naturalWidth > 0, await locator.elementHandle(), { timeout: 15000 })
  }
  async function assertAssetReads(state, provider, keys) {
    await state.page.waitForSelector('[data-pattern-status="default"]')
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
  assert((await normal.page.locator(background).evaluate(el => getComputedStyle(el).backgroundImage)).includes('radial-gradient'))
  // A tiny probe finishing does not imply that the original Hero request has
  // finished; route selection now deliberately leaves that request unchanged.
  if (!normal.assets.some(r => r.url().endsWith('/' + desktopKey) && r.ok())) {
    await normal.page.waitForResponse(response => response.url().endsWith('/' + desktopKey) && response.ok(), { timeout: 15000 })
  }
  assert(normal.assets.some(r => r.url().endsWith('/' + desktopKey) && r.ok()))
  assert.equal(normal.errors.length, 0, normal.errors.join('\n'))
  console.log('[assets real] Browser real icon/Hero, 4 full-body CORS probes, 12-hour cookie PASS', diagnostics)
  await closeContext(normal.context)

  const failPrimary = await newPage({ blocked: ['primary'] })
  await failPrimary.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.includes('assets-dev.gofurry.com'), hero, { timeout: 15000 })
  await failPrimary.page.mouse.wheel(0, 950)
  const mirrorIcon = failPrimary.page.locator('.nav-site-card__logo img').first()
  await loadedImage(failPrimary.page, mirrorIcon)
  assert((await mirrorIcon.getAttribute('src')).startsWith(origins.mirror))
  await assertAssetReads(failPrimary, 'mirror', [desktopKey, iconKey])
  await closeContext(failPrimary.context)
  console.log('[assets real] Primary failure -> real R2 CDN Hero / icon reads PASS')

  const failBoth = await newPage({ blocked: ['primary', 'mirror'] })
  await failBoth.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.startsWith('data:'), hero, { timeout: 15000 })
  await failBoth.page.waitForSelector('[data-pattern-status="default"]')
  await failBoth.page.mouse.wheel(0, 950)
  const defaultIcon = failBoth.page.locator('.nav-site-card__logo img').first()
  await loadedImage(failBoth.page, defaultIcon)
  assert.equal(await defaultIcon.getAttribute('src'), '/defaultLogo.svg')
  assert((await failBoth.page.locator(background).evaluate(el => getComputedStyle(el).backgroundImage)).includes('radial-gradient'))
  await closeContext(failBoth.context)
  console.log('[assets real] Both CDNs fail -> bundled icon / no Hero / CSS dots PASS')

  const mobile = await newPage({ mobile: true })
  await mobile.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.includes('/mobile/'), hero)
  if (!mobile.assets.some(r => r.url().endsWith('/' + mobileKey) && r.ok())) await mobile.page.waitForResponse(response => response.url().endsWith('/' + mobileKey) && response.ok(), { timeout: 15000 })
  assert(!mobile.assets.some(r => r.url().endsWith('/' + desktopKey)), 'mobile fetched desktop pool')
  await mobile.page.waitForSelector('[data-pattern-status="default"]')
  await mobile.page.evaluate(() => window.scrollTo(0, 450))
  await mobile.page.waitForFunction(() => document.querySelector('.mobile-bottom-tabs')?.getAttribute('aria-hidden') === 'false')
  await mobile.page.locator('.mobile-bottom-tabs button').click()
  await mobile.page.getByRole('tab', { name: '首屏设置', exact: true }).waitFor()
  await mobile.page.screenshot({ path: join(artifactDir, 'preferences-mobile-home.png') })
  await closeContext(mobile.context)
  mobileEnabled = false
  const emptyMobile = await newPage({ mobile: true })
  await emptyMobile.page.waitForFunction(selector => document.querySelector(selector + ' img')?.currentSrc.startsWith('data:'), hero)
  await emptyMobile.page.waitForSelector('[data-pattern-status="default"]')
  assert(!emptyMobile.assets.some(r => r.url().includes('/nav/hero/')), 'empty mobile pool borrowed desktop')
  await closeContext(emptyMobile.context)
  console.log('[assets real] Mobile real AVIF and empty pool isolation PASS')
  console.log('Real development CDN browser acceptance passed. Screenshot:', join(artifactDir, 'preferences-mobile-home.png'))
} catch (error) {
  if (currentPage && !currentPage.isClosed()) await currentPage.screenshot({ path: join(artifactDir, 'failure.png'), fullPage: false }).catch(() => {})
  console.error('Browser artifacts:', artifactDir, '\nNuxt:', app.logs())
  throw error
} finally {
  for (const context of browser.contexts()) await context.unrouteAll({ behavior: 'ignoreErrors' })
  await browser.close(); await app.close()
}
