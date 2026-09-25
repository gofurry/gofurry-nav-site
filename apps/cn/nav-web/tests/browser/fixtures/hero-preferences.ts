import { test as base, expect, type Page, type Response } from '@playwright/test'
import { readFile } from 'node:fs/promises'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_SHARED_CDN_GROUP_PREFIXES } from '../../../app/utils/steamAssets'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'
import { assertHeroHydration, heroOrigins } from './hero-lifecycle'

interface HeroItem { id: string; name: string; object_key: string }
interface Gate { promise: Promise<void>; release(): void }
const newGate = (): Gate => {
  let release!: () => void
  const promise = new Promise<void>(resolve => { release = resolve })
  return { promise, release }
}
const mobileItem = (id: string): HeroItem => ({ id, name: `mobile ${id}`, object_key: `nav/hero/mobile/${String(BigInt(id) % 10n).repeat(32)}.avif` })
export const displayedHero = '.nav-header__background:not([data-hero-pending])'
interface PreferenceState {
  desktopCatalog: HeroItem[]
  mobileCatalog: HeroItem[]
  failedCatalogPage: number | null
  failedImageKey: string | null
  requestedImages: string[]
  expectedNetworkFailures: Set<string>
  requests: URL[]
  apiGate: Gate | null
  imageGate: Gate | null
}
type FixtureApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
interface PreferenceWorker { app: FixtureApp; current: PreferenceState | null }
interface PreferenceOpenOptions {
  width?: number
  height?: number
  theme?: 'light' | 'dark'
  fixedNow?: number
  seedSteamDiagnostics?: boolean
  mode?: 'random' | 'fixed' | 'local' | null
  desktopId?: string
  mobileId?: string
  seedLocalCache?: boolean
}
interface HeroPreferences extends PreferenceState {
  errors: string[]
  open(options?: PreferenceOpenOptions): Promise<{ ssrHTML: string }>
  reload(): Promise<{ ssrHTML: string }>
  holdHeroAPI(): void
  releaseHeroAPI(): void
  holdHeroImages(): void
  releaseHeroImages(): void
  replaceDesktopObjectKey(id: string, key: string): void
  homeCalls(): URL[]
  heroCalls(): URL[]
  catalogCalls(): URL[]
  currentHero(): Promise<string>
  openPreferences(): Promise<void>
  savePreferences(): Promise<void>
  cancelPreferences(): Promise<void>
  waitCatalogPosition(position: number): Promise<void>
  browseCatalogTo(position: number): Promise<void>
  cookieValue(name: string): Promise<string | undefined>
}

export const test = base.extend<{ preferences: HeroPreferences }, { preferenceApp: PreferenceWorker }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  preferenceApp: [async ({}, use) => {
    const worker = { current: null } as PreferenceWorker
    worker.app = await startInsightsFixtureApp(async (url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Hero Preferences API called without an active test scenario')
      state.requests.push(url)
      const resolveHero = () => url.searchParams.get('hero_mode') === 'local' ? { desktop: null, mobile: null } : {
        desktop: state.desktopCatalog.find(item => item.id === url.searchParams.get('hero_desktop_id')) ?? state.desktopCatalog[0],
        mobile: state.mobileCatalog.find(item => item.id === url.searchParams.get('hero_mobile_id')) ?? state.mobileCatalog[0],
      }
      if (url.pathname === '/api/v2/nav/home') return { data: { schema_version: 4, groups: [], spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] }, saying: null, ping: {}, hero: resolveHero() } }
      if (url.pathname === '/api/v2/nav/home/hero') {
        if (state.apiGate) await state.apiGate.promise
        return { data: { hero: resolveHero() } }
      }
      if (url.pathname === '/api/v2/nav/appearance/heroes') {
        const variant = url.searchParams.get('variant'), items = variant === 'desktop' ? state.desktopCatalog : state.mobileCatalog
        const page = Number(url.searchParams.get('page_num')), size = Number(url.searchParams.get('page_size'))
        if (page === state.failedCatalogPage) return { status: 503 }
        return { data: { schema_version: 1, variant, page_num: page, page_size: size, total: items.length, items: items.slice((page - 1) * size, page * size), selected: items.find(item => item.id === url.searchParams.get('selected_id')) ?? null } }
      }
      if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: [] } }
      return { data: [] }
    }, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: heroOrigins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: heroOrigins.mirror })
    try { await use(worker) }
    finally { await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ preferenceApp }, use) => { await use(preferenceApp.app.base) },
  preferences: async ({ preferenceApp, page, context }, use, testInfo) => {
    expect(preferenceApp.current, 'previous Preferences scenario was not released').toBeNull()
    const state: PreferenceState = {
      desktopCatalog: Array.from({ length: 48 }, (_, i) => ({ id: String(i + 10), name: `desktop ${i + 10}`, object_key: `nav/hero/desktop/${String(i + 10).padStart(32, '0')}.avif` })),
      mobileCatalog: [mobileItem('9007199254740993'), mobileItem('9007199254740994')],
      failedCatalogPage: null, failedImageKey: null, requestedImages: [], expectedNetworkFailures: new Set(), requests: [], apiGate: null, imageGate: null,
    }
    preferenceApp.current = state
    const errors = captureBrowserErrors(page, state.expectedNetworkFailures)
    const probeBody = await readFile(new URL('../../fixtures/cdn-probe.bin', import.meta.url))
    const navigate = async (navigation: Promise<Response | null>) => {
      const response = await navigation
      expect(response?.status()).toBe(200)
      const ssrHTML = await response!.text()
      await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
      await assertHeroHydration(page, ssrHTML, errors)
      return { ssrHTML }
    }
    const waitCatalogPosition = async (position: number) => {
      await page.waitForFunction(expected => {
        const picker = [...document.querySelectorAll('[data-hero-catalog]')].find(el => (el.closest('[role="tabpanel"]') as HTMLElement).style.display !== 'none')
        return picker?.getAttribute('aria-busy') === 'false' && Number.parseInt(picker.querySelector('[data-hero-position]')!.textContent!) === expected
      }, position)
    }
    const finishPreferences = async (button: string) => {
      await page.locator(`.gf-modal__header-actions .gf-button--${button}`).click()
      await expect(page.locator('[data-hero-preferences]')).toHaveCount(0)
    }
    const scenario: HeroPreferences = Object.assign(state, {
      errors,
      async open({ width = 1440, height = 1000, theme, fixedNow, seedSteamDiagnostics = false, mode = 'random', desktopId, mobileId, seedLocalCache = false }: PreferenceOpenOptions = {}) {
        await page.setViewportSize({ width, height })
        const cookies = [['gf_asset_cdn', 'primary'], ['gf_steam_asset_mode', 'global'], ['gf_asset_cdn_mode', 'primary']]
        if (mode) cookies.push(['gf_hero_mode', mode])
        if (desktopId) cookies.push(['gf_hero_desktop_id', desktopId])
        if (mobileId) cookies.push(['gf_hero_mobile_id', mobileId])
        await context.addCookies(cookies.map(([name, value]) => ({ name: name!, value: value!, url: preferenceApp.app.base })))
        await context.addInitScript(({ baseURL, seedLocalCache, theme, fixedNow, seedSteamDiagnostics, steamDiagnosticsKey, steamProbePath }) => {
          if (location.origin !== baseURL) return
          // Optional Visual seeds; ordinary functional scenarios keep real time,
          // their existing viewport, and the production theme initialization.
          if (theme) localStorage.setItem('theme', theme)
          if (fixedNow !== undefined) Date.now = () => fixedNow
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt: Date.now(), primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          if (seedSteamDiagnostics) localStorage.setItem(steamDiagnosticsKey, JSON.stringify({
            version: 1, china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
            selected: 'china', checkedAt: Date.now(), sample: steamProbePath,
          }))
          const evidence = window as Window & { initialHeroNode?: Element | null }
          new MutationObserver(() => { evidence.initialHeroNode ||= document.querySelector('.nav-header__background') }).observe(document, { subtree: true, childList: true })
          // Seed a real cached Blob once, never a fake FileSystemDirectoryHandle.
          // Reload must exercise persisted storage, not create replacement data.
          if (!seedLocalCache || sessionStorage.getItem('hero-fixture-cache-seeded')) return
          localStorage.setItem('customNavHeaderBgFolderName', 'Saved folder')
          const originalOpen = indexedDB.open.bind(indexedDB)
          const seed = new Promise<void>((resolve, reject) => {
            const request = originalOpen('gofurry-custom-nav-header-bg', 1)
            request.onupgradeneeded = () => request.result.createObjectStore('directoryHandles')
            request.onerror = () => reject(request.error)
            request.onsuccess = () => {
              const db = request.result, tx = db.transaction('directoryHandles', 'readwrite')
              tx.objectStore('directoryHandles').put(new Blob(['<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40"><rect width="40" height="40" fill="#226633"/></svg>'], { type: 'image/svg+xml' }), 'nav-header-bg-cache')
              tx.oncomplete = () => { db.close(); sessionStorage.setItem('hero-fixture-cache-seeded', '1'); resolve() }
              tx.onerror = () => reject(tx.error)
            }
          })
          indexedDB.open = (...args: Parameters<IDBFactory['open']>) => {
            const request = originalOpen(...args)
            if (args[0] === 'gofurry-custom-nav-header-bg') Object.defineProperty(request, 'onsuccess', { set(handler: IDBOpenDBRequest['onsuccess']) {
              request.addEventListener('success', async event => { await seed; handler?.call(request, event) }, { once: true })
            } })
            return request
          }
        }, { baseURL: preferenceApp.app.base, seedLocalCache, theme, fixedNow, seedSteamDiagnostics,
          steamDiagnosticsKey: STEAM_DIAGNOSTICS_KEY, steamProbePath: STEAM_PROBE_PATHS[0] })
        await context.route('**/*', async route => {
          const url = new URL(route.request().url())
          if (url.origin === preferenceApp.app.base) {
            if (url.pathname === '/api/v2/nav/appearance/heroes' && Number(url.searchParams.get('page_num')) === state.failedCatalogPage) state.expectedNetworkFailures.add(url.href)
            return route.continue()
          }
          if (['data:', 'blob:'].includes(url.protocol)) return route.continue()
          if (Object.values(heroOrigins).includes(url.origin)) {
            if (url.pathname === '/system/probes/cdn.bin') return route.fulfill({ contentType: 'application/octet-stream', headers: { 'Access-Control-Allow-Origin': '*' }, body: probeBody })
            if (url.pathname.startsWith('/nav/hero/')) {
              state.requestedImages.push(url.href)
              if (state.imageGate) await state.imageGate.promise
              if (url.pathname === '/' + state.failedImageKey) {
                state.expectedNetworkFailures.add(url.href)
                return route.abort('failed')
              }
              const height = url.pathname.includes('/mobile/') ? 1200 : 506
              const colors = ['#6b8277', '#7c8c95', '#88826a']
              const color = colors[[...url.pathname].reduce((sum, char) => sum + char.charCodeAt(0), 0) % colors.length]
              return route.fulfill({ contentType: 'image/svg+xml', body: `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="${height}" viewBox="0 0 900 ${height}"><rect width="900" height="${height}" fill="#e5d5b6"/><circle cx="660" cy="${height * .28}" r="72" fill="#faf1d9"/><path d="M0 ${height}V${height * .72}L270 ${height * .32}L560 ${height}Z" fill="${color}"/><path d="M230 ${height}L620 ${height * .5}L900 ${height * .76}V${height}Z" fill="#475f59"/></svg>` })
            }
          }
          if (Object.values(STEAM_SHARED_CDN_GROUP_PREFIXES).flat().includes(url.origin) && STEAM_PROBE_PATHS.includes(url.pathname)) return route.fulfill({ contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="45"/>' })
          return route.abort()
        })
        return navigate(page.goto('/', { waitUntil: 'networkidle' }))
      },
      async reload() {
        // Navigation/network bookkeeping can remain busy after a reload. The
        // contract is hydrated SSR plus a loaded Hero, not whole-page network idle.
        const result = await navigate(page.reload({ waitUntil: 'load' }))
        await expectHeroLoaded(page)
        return result
      },
      holdHeroAPI() { expect(state.apiGate).toBeNull(); state.apiGate = newGate() },
      releaseHeroAPI() { state.apiGate?.release(); state.apiGate = null },
      holdHeroImages() { expect(state.imageGate).toBeNull(); state.imageGate = newGate() },
      releaseHeroImages() { state.imageGate?.release(); state.imageGate = null },
      replaceDesktopObjectKey(id: string, key: string) {
        const item = state.desktopCatalog.find(item => item.id === id)
        if (!item) throw new Error(`Unknown fixture Hero ${id}`)
        item.object_key = key
      },
      homeCalls: () => state.requests.filter(url => url.pathname === '/api/v2/nav/home'),
      heroCalls: () => state.requests.filter(url => url.pathname === '/api/v2/nav/home/hero'),
      catalogCalls: () => state.requests.filter(url => url.pathname === '/api/v2/nav/appearance/heroes'),
      currentHero: () => page.locator(displayedHero).evaluate(el => (el as HTMLImageElement).currentSrc || el.querySelector('img')?.currentSrc || ''),
      async openPreferences() {
        const desktopButton = page.locator('.gf-nav__mode-button')
        if (await desktopButton.isVisible()) await desktopButton.click()
        else {
          await page.locator('.gf-nav__mobile-toggle').click()
          await page.locator('.gf-nav__mobile-action').click()
        }
        await expect(page.locator('[data-hero-preferences]')).toBeVisible()
      },
      savePreferences: () => finishPreferences('primary'),
      cancelPreferences: () => finishPreferences('ghost'),
      waitCatalogPosition,
      async browseCatalogTo(target: number) {
        const picker = page.locator('[data-hero-catalog]:visible')
        await expect(picker).toHaveAttribute('aria-busy', 'false')
        let position = Number.parseInt((await picker.locator('[data-hero-position]').textContent())!)
        expect(position).toBeGreaterThan(0)
        while (position !== target) {
          const direction = target > position ? 1 : -1
          await picker.getByRole('button', { name: direction > 0 ? '下一张背景' : '上一张背景', exact: true }).click()
          position += direction
          await waitCatalogPosition(position)
        }
      },
      cookieValue: async (name: string) => (await context.cookies()).find(cookie => cookie.name === name)?.value,
    })
    try { await use(scenario); expect(errors).toEqual([]) }
    finally {
      scenario.releaseHeroAPI()
      scenario.releaseHeroImages()
      try {
        await context.unrouteAll({ behavior: 'wait' })
        if (testInfo.status !== testInfo.expectedStatus) {
          await testInfo.attach('hero-preferences.log', { body: preferenceApp.app.logs(), contentType: 'text/plain' })
          await testInfo.attach('hero-requests.json', { body: JSON.stringify({ images: state.requestedImages, api: state.requests.map(url => url.href) }, null, 2), contentType: 'application/json' })
        }
      } finally { preferenceApp.current = null }
    }
  },
})

export { expect, heroOrigins }

export async function expectHeroLoaded(page: Page) {
  await expect.poll(() => page.locator(displayedHero).evaluate(el => {
    const image = el instanceof HTMLImageElement ? el : el.querySelector('img')
    return Boolean(image?.complete && image.naturalWidth > 0)
  })).toBe(true)
}
