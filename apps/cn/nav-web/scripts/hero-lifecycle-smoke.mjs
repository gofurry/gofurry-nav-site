import assert from 'node:assert/strict'
import { mkdir, readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { launchPerfBrowser, reportsDir } from './perf/shared.mjs'

const origins = { primary: 'https://primary.example', mirror: 'https://mirror.example' }
const key = variant => 'nav/hero/' + variant + '/' + 'a'.repeat(32) + '.avif'
let homeRequests = 0, emptyMobile = false
const app = await startInsightsFixtureApp(url => {
  if (url.pathname === '/api/v2/nav/home') {
    homeRequests++
    const resourceKey = variant => homeRequests === 1 ? key(variant) : key(variant).replace(/a{32}/, 'b'.repeat(32))
    return { data: { schema_version: 4, groups: [], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
      saying: null, ping: {}, hero: { desktop: { id: '1', object_key: resourceKey('desktop') }, mobile: emptyMobile ? null : { id: '2', object_key: resourceKey('mobile') } } } }
  }
  if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: [] } }
  return { data: [] }
}, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: origins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: origins.mirror })
const browser = await launchPerfBrowser()
const output = join(reportsDir, 'hero-lifecycle')
await mkdir(output, { recursive: true })
const probeBody = await readFile(new URL('../tests/fixtures/cdn-probe.bin', import.meta.url))
const artwork = color => '<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect width="100" height="100" fill="' + color + '"/></svg>'
const failures = []
async function setup(width, blocked = [], local = false) {
  homeRequests = 0
  const context = await browser.newContext({ viewport: { width, height: 900 }, locale: 'zh-CN', reducedMotion: 'reduce' })
  await context.addCookies([{ name: 'gf_asset_cdn', value: 'primary', url: app.base }])
  await context.addInitScript(({ base, local }) => {
    if (location.origin !== base) return
    localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt: Date.now(), primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
    new MutationObserver(() => {
      window.initialHeroNode ||= document.querySelector('.nav-header__background--managed')
    }).observe(document, { childList: true, subtree: true })
    // Only independently constructed Image objects are faulted. CSS/DOM image
    // requests still use the browser's real renderer and the normal network.
    const NativeImage = window.Image
    const descriptor = Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, 'src')
    const probes = []
    window.heroAuxiliaryImages = probes
    window.failHeroAuxiliaryImages = () => probes.forEach(image => image.dispatchEvent(new Event('error')))
    window.Image = function (...args) {
      const image = new NativeImage(...args)
      Object.defineProperty(image, 'src', {
        get() { return descriptor.get.call(image) },
        set(url) {
          if (String(url).includes('/nav/hero/')) {
            // A failed auxiliary fetch must never veto the successful display.
            probes.push(image)
          } else descriptor.set.call(image, url)
        },
      })
      return image
    }
    window.Image.prototype = NativeImage.prototype
    if (local) {
      // Restore the actual persisted cached-local-image path, with no permission
      // dialog or filesystem handle mock. Gate IDB open so the managed SSR Hero
      // has painted first.
      const originalOpen = indexedDB.open.bind(indexedDB)
      window.seedHeroLocal = new Promise((resolve, reject) => {
        const request = originalOpen('gofurry-custom-nav-header-bg', 1)
        request.onupgradeneeded = () => request.result.createObjectStore('directoryHandles')
        request.onerror = () => reject(request.error)
        request.onsuccess = () => {
          const db = request.result
          const tx = db.transaction('directoryHandles', 'readwrite')
          tx.objectStore('directoryHandles').put(new Blob(['<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect width="100" height="100" fill="#227733"/></svg>'], { type: 'image/svg+xml' }), 'nav-header-bg-cache')
          tx.oncomplete = () => { db.close(); resolve() }
        }
      })
      localStorage.setItem('customNavHeaderBgFolderName', 'Fixture backgrounds')
      // App reads begin after this initialization script; delay that read to
      // expose mounted restoration separately from the display fetch.
      indexedDB.open = (...args) => {
        const request = originalOpen(...args)
        if (args[0] !== 'gofurry-custom-nav-header-bg') return request
        const add = request.addEventListener.bind(request)
        Object.defineProperty(request, 'onsuccess', { set(handler) {
          add('success', async event => { await window.seedHeroLocal; await new Promise(resolve => { window.releaseHeroLocal = resolve }); handler.call(request, event) }, { once: true })
        } })
        return request
      }
    }
  }, { base: app.base, local })
  const requests = []
  let releaseHydration
  const displayFailedBeforeHydration = new Promise(resolve => { releaseHydration = resolve })
  await context.route('**/*', async route => {
    const url = new URL(route.request().url())
    if (blocked.length && url.origin === app.base && url.pathname.startsWith('/_nuxt/') && url.pathname.endsWith('.js')) await displayFailedBeforeHydration
    if (url.origin === app.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
    if (!Object.values(origins).includes(url.origin)) return route.abort()
    if (url.pathname.endsWith('cdn.bin')) {
      await delay(url.origin === origins.primary ? 160 : 5)
      return route.fulfill({ contentType: 'application/octet-stream', headers: { 'Access-Control-Allow-Origin': '*' }, body: probeBody })
    }
    if (url.pathname.startsWith('/nav/hero/')) {
      requests.push(url.href)
      if (blocked.includes(url.origin)) return route.abort()
      return route.fulfill({ contentType: 'image/svg+xml', body: artwork(url.origin === origins.primary ? '#b34433' : '#3355bb') })
    }
    return route.abort()
  })
  const page = await context.newPage()
  const hydrationWarnings = []
  page.on('requestfailed', request => { if (request.url().includes('/nav/hero/')) releaseHydration() })
  page.on('pageerror', error => failures.push(error.message))
  page.on('console', message => { if (/hydration.*mismatch/i.test(message.text())) hydrationWarnings.push(message.text()) })
  const response = await page.goto(app.base, { waitUntil: 'networkidle' })
  await page.waitForFunction(() => Boolean(document.querySelector('#__nuxt')?.__vue_app__))
  assert(await page.evaluate(() => window.initialHeroNode === document.querySelector('.nav-header__background--managed')), 'hydration replaced the SSR Hero element')
  if (hydrationWarnings.length) {
    // Existing default.vue renders no footer on the server but renders one in
    // the initial narrow-screen client tree. Keep that unrelated layout issue
    // visible; the Hero-specific node/source/pixel checks above/below still run.
    assert(width < 768 && !(await response.text()).includes('class="gf-footer-shell') && await page.locator('.gf-footer-shell').count() === 1,
      'unexpected hydration mismatch: ' + hydrationWarnings.join('; '))
    console.log('[hero] existing mobile footer SSR/client mismatch observed; SSR Hero node retained (outside #121)')
  }
  return { page, context, requests }
}
async function displayed(page) {
  return page.locator('.nav-header__background--managed').evaluate(element => {
    const image = element.querySelector('img')
    return image ? image.currentSrc : getComputedStyle(element).backgroundImage
  })
}
async function paint(page) {
  await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))))
  return page.screenshot({ clip: { x: 4, y: 200, width: 16, height: 16 }, animations: 'disabled' })
}
async function close(context) { await context.unrouteAll({ behavior: 'wait' }); await context.close() }
try {
  for (const width of [1440, 390]) {
    const { page, context, requests } = await setup(width)
    const variant = width >= 768 ? 'desktop' : 'mobile'
    const initial = await displayed(page)
    assert(initial.includes(origins.primary + '/' + key(variant)))
    assert.deepEqual(requests, [origins.primary + '/' + key(variant)], 'only the viewport-specific display should fetch the Hero')
    const before = await paint(page)
    assert.equal(homeRequests, 1, 'SSR/hydration unexpectedly fetched a different random Hero')
    await page.screenshot({ path: join(output, 'hero-' + width + '-before.png') })
    await page.evaluate(() => window.failHeroAuxiliaryImages())
    await delay(250)
    assert.equal(await displayed(page), initial, 'successful displayed Hero was replaced by a failed hidden Image')
    assert.deepEqual(await paint(page), before, 'painted Hero pixels changed after auxiliary failure')
    assert.equal(await page.evaluate(() => window.heroAuxiliaryImages.length), 0, 'Hero still has an independent hidden Image failure authority')
    await page.evaluate(async () => {
      const cdn = document.querySelector('#__nuxt').__vue_app__.config.globalProperties.$nuxt.$assetCDN
      await cdn.probe(true)
      cdn.saveMode('mirror')
    })
    assert.equal(await displayed(page), initial, 'probe/Save replaced a successful displayed Hero')
    assert.equal(await page.evaluate(() => JSON.parse(localStorage.getItem('gf_asset_cdn_diagnostics')).selected), 'mirror')
    await page.evaluate(() => { window.dispatchEvent(new Event('focus')); window.dispatchEvent(new Event('resize')) })
    await delay(900)
    assert.equal(homeRequests, 1, 'hydration/mount/focus/prewarm refetched the random Hero pool')
    assert.equal(app.requests.filter(url => url.pathname === '/api/v2/nav/home/hero').length, 0, 'NavHeader requested a second Hero after receiving SSR keys')
    assert.equal(await displayed(page), initial)
    assert.deepEqual(requests, [origins.primary + '/' + key(variant)], 'existing Hero was fetched again')
    // Explicitly refreshing data is a new key, unlike mount/focus/probe work.
    await page.evaluate(() => { const nuxt = document.querySelector('#__nuxt').__vue_app__.config.globalProperties.$nuxt; return nuxt.callHook('app:data:refresh', Object.keys(nuxt.payload.data).filter(key => key.startsWith('nav-page:zh:'))) })
    await page.waitForFunction(() => document.querySelector('.nav-header__background--managed img').currentSrc.includes('mirror.example') && document.querySelector('.nav-header__background--managed img').currentSrc.includes('bbbbbbbb'))
    assert.equal(homeRequests, 2)
    await close(context)
    console.log('[hero] ' + width + 'px actual paint, late auxiliary failure, recommendation/Save, one automatic home request and explicit new key PASS')
  }
  // A genuine display request failure must still retry the mirror, then stop
  // with no Hero if both providers fail.
  for (const blocked of [[origins.primary], [origins.primary, origins.mirror]]) {
    const { page, context, requests } = await setup(1440, blocked)
    await page.waitForFunction(() => {
      const image = document.querySelector('.nav-header__background--managed img')
      return image?.complete && image.naturalWidth > 0
    })
    assert(requests.includes(origins.primary + '/' + key('desktop')))
    assert(requests.includes(origins.mirror + '/' + key('desktop')))
    assert.equal(requests.some(url => url.includes('/mobile/')), false)
    const source = await displayed(page)
    assert(blocked.length === 1 ? source.includes(origins.mirror) : source.startsWith('data:'), 'actual renderer failure lost fallback/terminal behavior')
    await close(context)
  }
  emptyMobile = true
  const empty = await setup(390)
  assert.equal(empty.requests.length, 0, 'empty mobile pool borrowed a desktop Hero')
  await close(empty.context)
  emptyMobile = false

  const local = await setup(1440, [], true)
  const managed = await displayed(local.page)
  assert(managed.includes(origins.primary))
  await local.page.waitForFunction(() => Boolean(window.releaseHeroLocal))
  // Two independent readonly opens (handle then cached blob) are gated.
  await local.page.evaluate(() => { window.releaseHeroLocal(); window.releaseHeroLocal = null })
  await local.page.waitForFunction(() => Boolean(window.releaseHeroLocal))
  await local.page.evaluate(() => window.releaseHeroLocal())
  await local.page.waitForFunction(() => {
    const element = document.querySelector('.nav-header__background:not(.nav-header__background--managed)')
    return element && (element.src?.startsWith('blob:') || getComputedStyle(element).backgroundImage.includes('blob:'))
  })
  assert.equal(homeRequests, 1, 'restoring a local background refetched home')
  const localURL = await local.page.locator('.nav-header__background:not(.nav-header__background--managed)').evaluate(el => el.src || getComputedStyle(el).backgroundImage)
  await local.page.evaluate(() => document.querySelector('#__nuxt').__vue_app__.config.globalProperties.$nuxt.$assetCDN.saveMode('mirror'))
  await delay(100)
  assert.equal(await local.page.locator('.nav-header__background:not(.nav-header__background--managed)').evaluate(el => el.src || getComputedStyle(el).backgroundImage), localURL)
  await close(local.context)
  assert.deepEqual(failures, [])
  console.log('[hero] real renderer fallback, empty pool isolation and saved local background restoration PASS')
} finally {
  for (const context of browser.contexts()) await context.unrouteAll({ behavior: 'ignoreErrors' })
  await browser.close()
  await app.close()
}
