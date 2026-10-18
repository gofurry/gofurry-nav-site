import { test as base, expect, type Page } from '@playwright/test'
import { readFile } from 'node:fs/promises'
import { setTimeout as delay } from 'node:timers/promises'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_SHARED_CDN_GROUP_PREFIXES } from '../../../app/utils/steamAssets'
import { STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

export const heroOrigins = { primary: 'https://primary.example', mirror: 'https://mirror.example' }
export const lifecycleKey = (variant: 'desktop' | 'mobile', hash = 'a') => `nav/hero/${variant}/${hash.repeat(32)}.avif`
const artwork = (color: string) => `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect width="100" height="100" fill="${color}"/></svg>`

interface LifecycleState {
  emptyMobile: boolean
  blockedHeroOrigins: Set<string>
  heroRequests: string[]
  homeCalls: URL[]
  heroCalls: URL[]
  expectedNetworkFailures: Set<string>
  failuresBeforeHydration: string[]
}

type FixtureApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
interface LifecycleWorker { app: FixtureApp; current: LifecycleState | null }
interface HeroRuntimeWindow extends Window {
  initialHeroNode?: Element | null
  heroAuxiliaryImages: HTMLImageElement[]
  failHeroAuxiliaryImages(): void
  seedHeroLocal?: Promise<void>
  pendingLocalReads: (() => void)[]
  localReadCount: number
  releaseAllLocalReads(): void
}
type HeroNuxtElement = Element & { __vue_app__: { config: { globalProperties: { $nuxt: {
  payload: { data: Record<string, unknown> }
  callHook(name: string, keys: string[]): Promise<void>
  $assetCDN: { probe(manual: boolean): Promise<unknown>; saveMode(mode: string): void }
} } } } }

interface HeroLifecycle extends LifecycleState {
  errors: string[]
  open(options?: { width?: number; seedLocalCache?: boolean; gateLocalRestore?: boolean }): Promise<{ ssrHTML: string }>
  currentHero(): Promise<string>
  waitForLocalReadGate(read: number): Promise<void>
  releaseNextLocalRead(): Promise<void>
  failAuxiliaryImages(): Promise<void>
  auxiliaryImageCount(): Promise<number>
  probeAndPinMirror(): Promise<void>
  refreshHome(): Promise<void>
  paintSample(): Promise<Buffer>
}

// Hero-only debt assertion, deliberately outside the generic error collector.
// Consume only the already-verified navigation evidence; later errors still fail.
export async function assertHeroHydration(page: Page, ssrHTML: string, errors: string[], homePath: '/' | '/en' = '/') {
  const mismatches = errors.filter(error => /hydration.*mismatch|mismatch.*hydration/i.test(error))
  if (mismatches.length) {
    expect(page.viewportSize()!.width).toBeLessThan(768)
    // English Home opts in explicitly; existing callers retain the '/' boundary.
    // See the Footer runtime follow-up in the #124 closure record.
    expect(new URL(page.url()).pathname).toBe(homePath)
    expect(mismatches).toHaveLength(1)
    expect(ssrHTML).not.toMatch(/class="[^"]*\bgf-footer-shell\b/)
    await expect(page.locator('.gf-footer-shell')).toHaveCount(1)
    await expectSSRHeroRetained(page)
    expect(errors.filter(error => !mismatches.includes(error))).toEqual([])
  } else expect(errors).toEqual([])
  errors.length = 0
}

export async function expectSSRHeroRetained(page: Page) {
  expect(await page.evaluate(() => {
    const initial = (window as unknown as HeroRuntimeWindow).initialHeroNode
    return Boolean(initial) && initial === document.querySelector('.nav-header__background:not([data-hero-pending])')
  })).toBe(true)
}

export const test = base.extend<{ hero: HeroLifecycle }, { lifecycleApp: LifecycleWorker }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  lifecycleApp: [async ({}, use) => {
    const worker = { current: null } as LifecycleWorker
    worker.app = await startInsightsFixtureApp((url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Hero lifecycle API called without an active test scenario')
      if (url.pathname === '/api/v2/nav/home') {
        state.homeCalls.push(url)
        const hash = state.homeCalls.length === 1 ? 'a' : 'b'
        return { data: { schema_version: 4, groups: [], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
          saying: null, ping: {}, hero: { desktop: { id: '1', object_key: lifecycleKey('desktop', hash) }, mobile: state.emptyMobile ? null : { id: '2', object_key: lifecycleKey('mobile', hash) } } } }
      }
      if (url.pathname === '/api/v2/nav/home/hero') state.heroCalls.push(url)
      if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: [] } }
      return { data: [] }
    }, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: heroOrigins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: heroOrigins.mirror })
    try { await use(worker) }
    finally { await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ lifecycleApp }, use) => { await use(lifecycleApp.app.base) },
  hero: async ({ lifecycleApp, page, context }, use, testInfo) => {
    expect(lifecycleApp.current, 'previous lifecycle scenario was not released').toBeNull()
    const state: LifecycleState = { emptyMobile: false, blockedHeroOrigins: new Set(), heroRequests: [], homeCalls: [], heroCalls: [], expectedNetworkFailures: new Set(), failuresBeforeHydration: [] }
    lifecycleApp.current = state
    const errors = captureBrowserErrors(page, state.expectedNetworkFailures)
    const probeBody = await readFile(new URL('../../fixtures/cdn-probe.bin', import.meta.url))
    let releaseHydration!: () => void
    const hydrationGate = new Promise<void>(resolve => { releaseHydration = resolve })
    let hydrationReleased = false
    page.on('requestfailed', request => {
      if (state.expectedNetworkFailures.has(request.url()) && !hydrationReleased) {
        state.failuresBeforeHydration.push(request.url())
        hydrationReleased = true
        releaseHydration()
      }
    })
    const scenario: HeroLifecycle = Object.assign(state, {
      errors,
      async open({ width = 1440, seedLocalCache = false, gateLocalRestore = false } = {}) {
        await page.setViewportSize({ width, height: 900 })
        await context.addCookies([{ name: 'gf_asset_cdn', value: 'primary', url: lifecycleApp.app.base }])
        await context.addInitScript(({ baseURL, seedLocalCache, gateLocalRestore }) => {
          if (location.origin !== baseURL) return
          const evidence = window as unknown as HeroRuntimeWindow
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt: Date.now(), primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          new MutationObserver(() => { evidence.initialHeroNode ||= document.querySelector('.nav-header__background') })
            .observe(document, { childList: true, subtree: true })
          // Trap auxiliary Image authority only. DOM picture/img keep real loading.
          const NativeImage = window.Image
          const descriptor = Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, 'src')!
          evidence.heroAuxiliaryImages = []
          evidence.failHeroAuxiliaryImages = () => evidence.heroAuxiliaryImages.forEach(image => image.dispatchEvent(new Event('error')))
          window.Image = function (...args: ConstructorParameters<typeof Image>) {
            const image = new NativeImage(...args)
            Object.defineProperty(image, 'src', {
              get() { return descriptor.get!.call(image) },
              set(url: string) {
                if (String(url).includes('/nav/hero/')) evidence.heroAuxiliaryImages.push(image)
                else descriptor.set!.call(image, url)
              },
            })
            return image
          } as typeof Image
          window.Image.prototype = NativeImage.prototype
          evidence.pendingLocalReads = []
          evidence.localReadCount = 0
          let gateReads = gateLocalRestore
          evidence.releaseAllLocalReads = () => { gateReads = false; evidence.pendingLocalReads.splice(0).forEach(release => release()) }
          if (!seedLocalCache) return
          const originalOpen = indexedDB.open.bind(indexedDB)
          evidence.seedHeroLocal = new Promise<void>((resolve, reject) => {
            const request = originalOpen('gofurry-custom-nav-header-bg', 1)
            request.onupgradeneeded = () => request.result.createObjectStore('directoryHandles')
            request.onerror = () => reject(request.error)
            request.onsuccess = () => {
              const db = request.result, tx = db.transaction('directoryHandles', 'readwrite')
              tx.objectStore('directoryHandles').put(new Blob(['<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect width="100" height="100" fill="#227733"/></svg>'], { type: 'image/svg+xml' }), 'nav-header-bg-cache')
              tx.oncomplete = () => { db.close(); resolve() }
              tx.onerror = () => reject(tx.error)
            }
          })
          localStorage.setItem('customNavHeaderBgFolderName', 'Fixture backgrounds')
          indexedDB.open = (...args: Parameters<IDBFactory['open']>) => {
            const request = originalOpen(...args)
            if (args[0] !== 'gofurry-custom-nav-header-bg') return request
            Object.defineProperty(request, 'onsuccess', { set(handler: IDBOpenDBRequest['onsuccess']) {
              request.addEventListener('success', async event => {
                await evidence.seedHeroLocal
                evidence.localReadCount++
                if (gateReads) await new Promise<void>(resolve => { evidence.pendingLocalReads.push(resolve) })
                handler?.call(request, event)
              }, { once: true })
            } })
            return request
          }
        }, { baseURL: lifecycleApp.app.base, seedLocalCache, gateLocalRestore })
        await context.route('**/*', async route => {
          const url = new URL(route.request().url())
          if (state.blockedHeroOrigins.size && url.origin === lifecycleApp.app.base && url.pathname.startsWith('/_nuxt/') && url.pathname.endsWith('.js')) await hydrationGate
          if (url.origin === lifecycleApp.app.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
          if (Object.values(heroOrigins).includes(url.origin)) {
            if (url.pathname === '/system/probes/cdn.bin') {
              await delay(url.origin === heroOrigins.primary ? 160 : 5)
              return route.fulfill({ contentType: 'application/octet-stream', headers: { 'Access-Control-Allow-Origin': '*' }, body: probeBody })
            }
            if (url.pathname.startsWith('/nav/hero/')) {
              state.heroRequests.push(url.href)
              if (state.blockedHeroOrigins.has(url.origin)) {
                if (!hydrationReleased) {
                  expect(await page.evaluate(() => Boolean((document.querySelector('#__nuxt') as HeroNuxtElement)?.__vue_app__)), 'renderer must fail before hydration').toBe(false)
                }
                state.expectedNetworkFailures.add(url.href)
                return route.abort('failed')
              }
              return route.fulfill({ contentType: 'image/svg+xml', body: artwork(url.origin === heroOrigins.primary ? '#b34433' : '#3355bb') })
            }
          }
          if (Object.values(STEAM_SHARED_CDN_GROUP_PREFIXES).flat().includes(url.origin) && STEAM_PROBE_PATHS.includes(url.pathname)) {
            return route.fulfill({ contentType: 'image/svg+xml', body: artwork('#3355bb') })
          }
          return route.abort()
        })
        const response = await page.goto('/', { waitUntil: 'networkidle' })
        expect(response?.status()).toBe(200)
        const ssrHTML = await response!.text()
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as HeroNuxtElement)?.__vue_app__))
        await assertHeroHydration(page, ssrHTML, errors)
        return { ssrHTML }
      },
      currentHero: () => page.locator('.nav-header__background:not([data-hero-pending])').evaluate(el => (el as HTMLImageElement).currentSrc || el.querySelector('img')?.currentSrc || ''),
      waitForLocalReadGate: (read: number) => page.waitForFunction(read => {
        const evidence = window as unknown as HeroRuntimeWindow
        return evidence.localReadCount === read && evidence.pendingLocalReads.length === 1
      }, read).then(() => {}),
      releaseNextLocalRead: () => page.evaluate(() => {
        const release = (window as unknown as HeroRuntimeWindow).pendingLocalReads.shift()
        if (!release) throw new Error('No pending Hero local read')
        release()
      }),
      failAuxiliaryImages: () => page.evaluate(() => (window as unknown as HeroRuntimeWindow).failHeroAuxiliaryImages()),
      auxiliaryImageCount: () => page.evaluate(() => (window as unknown as HeroRuntimeWindow).heroAuxiliaryImages.length),
      probeAndPinMirror: () => page.evaluate(async () => {
        const cdn = (document.querySelector('#__nuxt') as HeroNuxtElement).__vue_app__.config.globalProperties.$nuxt.$assetCDN
        await cdn.probe(true)
        cdn.saveMode('mirror')
      }),
      refreshHome: () => page.evaluate(() => {
        const nuxt = (document.querySelector('#__nuxt') as HeroNuxtElement).__vue_app__.config.globalProperties.$nuxt
        return nuxt.callHook('app:data:refresh', Object.keys(nuxt.payload.data).filter(key => key.startsWith('nav-page:zh:')))
      }),
      async paintSample() {
        await page.evaluate(() => new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))))
        return page.screenshot({ clip: { x: 4, y: 200, width: 16, height: 16 }, animations: 'disabled' })
      },
    })
    try { await use(scenario); expect(errors).toEqual([]) }
    finally {
      releaseHydration()
      try {
        if (!page.isClosed()) await page.evaluate(() => (window as unknown as HeroRuntimeWindow).releaseAllLocalReads?.())
        await context.unrouteAll({ behavior: 'wait' })
        if (testInfo.status !== testInfo.expectedStatus) await testInfo.attach('hero-lifecycle.log', { body: lifecycleApp.app.logs(), contentType: 'text/plain' })
      } finally { lifecycleApp.current = null }
    }
  },
})

export { expect }
