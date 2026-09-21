import { test as base, expect, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Gate = { promise: Promise<void>, release(): void }
type Scenario = { requests: URL[], gate?: Gate }
type Worker = { app: App, current: Scenario | null }
type Clip = { x: number, y: number, width: number, height: number }
type Header = {
  root: Locator
  search: Locator
  input: Locator
  suggestions: Locator
  items: Locator
  quick: Locator
  modal: Locator
  open(width?: number): Promise<void>
  holdSuggestions(): void
  releaseSuggestions(): void
  expectSuggestionRequest(query: string): Promise<void>
  showSuggestions(): Promise<void>
  allowExampleIcon(): void
  allowSearchPopup(): void
  assertQuiet(): void
  settle(target: Locator): Promise<void>
  headerClip(): Promise<Clip>
  searchClip(): Promise<Clip>
  modalClip(): Promise<Clip>
}

const primary = 'https://header-primary.example'
const desktopKey = `nav/hero/desktop/${'a'.repeat(32)}.avif`
const mobileKey = `nav/hero/mobile/${'b'.repeat(32)}.avif`
const heroURLs = [primary + '/' + desktopKey, primary + '/' + mobileKey]
const weatherURL = 'https://i.tianqi.com/index.php?c=code&id=73&icon=1&num=3&color=d1d5dc'
export const searchPopupURL = 'https://www.bing.com/search?q=wolf%20furry'
export const recentSeed = [
  { id: 'recent-community', name: 'Community', url: 'https://community.example/' },
  { id: 'recent-archive', name: 'Archive', url: 'https://archive.example/' },
]
export const customSeed = [
  { id: 'custom-art', name: 'Artwork', url: 'https://art.example/' },
  { id: 'custom-wiki', name: 'Wiki', url: 'https://wiki.example/' },
]
const favicon = (url: string) => `https://favicon.im/${url}?larger=true`
const failedIcon = favicon(recentSeed[1]!.url)
const seededIcons = [...recentSeed, ...customSeed].map(site => favicon(site.url))
const exampleIcon = favicon('https://example.com')
const iconSVG = '<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"><rect width="32" height="32" rx="7" fill="#719489"/><path d="M8 24V8l8 8 8-8v16" fill="none" stroke="#fff5e9" stroke-width="3"/></svg>'
// Stable image bytes at exact managed URLs; the real picture/img, snapshot and
// rendering pipeline remain production-owned. No component or DOM substitution.
const heroSVG = '<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="1000" viewBox="0 0 1600 1000"><rect width="1600" height="1000" fill="#697c83"/><circle cx="1240" cy="200" r="110" fill="#c4b79b"/><path d="M0 1000V620L380 280 760 650 1160 400 1600 720V1000" fill="#465f64"/><path d="M0 1000V820L540 600 900 870 1390 620 1600 800V1000" fill="#314c50"/></svg>'

export const test = base.extend<{ header: Header }, { headerApp: Worker }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  headerApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp(async (url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Header API requires a test-scoped scenario')
      state.requests.push(url)
      if (url.pathname === '/api/v2/nav/home') return { data: {
        schema_version: 4, generated_at: '2026-09-18T12:40:00Z', cache_state: {}, groups: [],
        spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
        saying: { author: 'GoFurry', content: '探索更多兽人世界', language: 'zh' }, ping: {},
        hero: { desktop: { id: '1', object_key: desktopKey }, mobile: { id: '2', object_key: mobileKey } },
      } }
      if (url.pathname === '/api/v2/nav/search/suggestions' && url.searchParams.get('engine') === 'bing'
        && ['wolf', 'noresult'].includes(url.searchParams.get('q') ?? '')) {
        await state.gate?.promise
        const query = url.searchParams.get('q')!
        return { data: { schema_version: 1, generated_at: '2026-09-18T12:40:00Z',
          state: query === 'wolf' ? 'ready' : 'empty', engine: 'bing', query,
          suggestions: query === 'wolf' ? ['wolf furry', 'wolf art', 'wolf game'] : [], cache_state: 'hit' } }
      }
      return { status: 500 }
    }, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: 'https://header-mirror.example' })
    try { await use(worker) }
    finally { worker.current?.gate?.release(); await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ headerApp }, use) => { await use(headerApp.app.base) },
  header: async ({ page, context, headerApp }, use, testInfo) => {
    const state: Scenario = { requests: [] }
    headerApp.current = state
    const { app } = headerApp
    const upstreamStart = app.requests.length
    const expectedFailures = new Set<string>()
    const errors = captureBrowserErrors(page, expectedFailures)
    const popupErrors: string[][] = []
    const external: string[] = [], unexpectedFailed: string[] = [], injectedFailed: string[] = []
    const managed: string[] = [], icons: string[] = [], weather: string[] = [], popups: string[] = [], browserAPI: string[] = []
    let exampleAllowed = false, popupAllowed = false
    context.on('page', popup => { if (popup !== page) popupErrors.push(captureBrowserErrors(popup)) })
    context.on('requestfailed', request => {
      if (expectedFailures.has(request.url()) && request.failure()?.errorText === 'net::ERR_FAILED') injectedFailed.push(request.url())
      else unexpectedFailed.push(`${request.url()}: ${request.failure()?.errorText}`)
    })
    page.on('request', request => {
      const url = new URL(request.url())
      if (url.origin === app.base && url.pathname.startsWith('/api/')) browserAPI.push(url.pathname + url.search)
    })
    await context.route('**/*', async route => {
      const url = route.request().url()
      if (new URL(url).origin === app.base || /^(data|blob):/.test(url)) return route.continue()
      if (heroURLs.includes(url)) {
        managed.push(url)
        return route.fulfill({ contentType: 'image/svg+xml', body: heroSVG })
      }
      if (seededIcons.includes(url) || (exampleAllowed && url === exampleIcon)) {
        icons.push(url)
        if (url === failedIcon) {
          expectedFailures.add(url)
          return route.abort('failed')
        }
        return route.fulfill({ contentType: 'image/svg+xml', body: iconSVG })
      }
      if (url === weatherURL && route.request().isNavigationRequest()
        && route.request().frame().parentFrame() === page.mainFrame()) {
        weather.push(url)
        return route.fulfill({ contentType: 'text/html', body: '' })
      }
      if (popupAllowed && url === searchPopupURL && route.request().isNavigationRequest()) {
        popups.push(url)
        return route.fulfill({ contentType: 'text/html', body: '<title>Local Bing navigation boundary</title>' })
      }
      external.push(url)
      await route.abort('blockedbyclient')
    })
    const root = page.locator('.nav-header')
    const search = root.locator('.search-box-shell')
    const input = search.locator('.search-input')
    const suggestions = search.locator('.search-suggestion-list')
    const items = suggestions.locator('.search-suggestion-item')
    const quick = root.locator('.quick-access-grid')
    const modal = root.locator('.quick-modal-panel')
    const releaseSuggestions = () => { state.gate?.release(); state.gate = undefined }
    const assertQuiet = () => {
      expect(app.requests.slice(upstreamStart), 'No unaccounted fixture media/business upstream calls').toEqual(state.requests)
      const home = state.requests.filter(url => url.pathname === '/api/v2/nav/home')
      expect(home).toHaveLength(1)
      expect(home[0]!.searchParams.get('lang')).toBe('zh')
      expect(state.requests.filter(url => !['/api/v2/nav/home', '/api/v2/nav/search/suggestions'].includes(url.pathname))).toEqual([])
      expect(browserAPI.every(path => path.startsWith('/api/v2/nav/search/suggestions?')), 'Home is SSR-only; no hydration refetch or secondary Home API').toBe(true)
      for (const url of state.requests.filter(url => url.pathname.endsWith('/suggestions'))) {
        expect([...url.searchParams.keys()].sort()).toEqual(['engine', 'q'])
        expect(url.searchParams.get('engine')).toBe('bing')
        expect(['wolf', 'noresult']).toContain(url.searchParams.get('q'))
      }
      expect(managed).toEqual([heroURLs[page.viewportSize()!.width < 768 ? 1 : 0]])
      expect(weather).toEqual([weatherURL])
      expect(external).toEqual([])
      expect(unexpectedFailed).toEqual([])
      expect(errors).toEqual([])
      expect(popupErrors.flat()).toEqual([])
      expect(popups).toEqual(popupAllowed ? [searchPopupURL] : [])
    }
    const settle = async (target: Locator) => {
      // Preserve intentional input focus in suggestion/modal visual states.
      await page.mouse.move(1, 1)
      await expect.poll(() => target.locator('img').evaluateAll(images => images
        .filter(image => image.getClientRects().length > 0)
        .every(image => image.complete && image.naturalWidth > 0))).toBe(true)
      await target.evaluate(async element => {
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
        await Promise.all(element.getAnimations({ subtree: true })
          .filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
          .map(animation => animation.finished.catch(() => {})))
        await document.fonts.ready
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
      })
    }
    const expandedClip = async (targets: Locator[], margin: number) => {
      const boxes = await Promise.all(targets.map(target => target.boundingBox()))
      for (const box of boxes) expect(box).not.toBeNull()
      const viewport = page.viewportSize()!
      const x = Math.max(0, Math.floor(Math.min(...boxes.map(box => box!.x)) - margin))
      const y = Math.max(0, Math.floor(Math.min(...boxes.map(box => box!.y)) - margin))
      const right = Math.min(viewport.width, Math.ceil(Math.max(...boxes.map(box => box!.x + box!.width)) + margin))
      const bottom = Math.min(viewport.height, Math.ceil(Math.max(...boxes.map(box => box!.y + box!.height)) + margin))
      expect(right).toBeGreaterThan(x)
      expect(bottom).toBeGreaterThan(y)
      return { x, y, width: right - x, height: bottom - y }
    }
    const expectSuggestionRequest = async (query: string) => {
      await expect.poll(() => state.requests.filter(url => url.pathname.endsWith('/suggestions')
        && url.searchParams.get('engine') === 'bing' && url.searchParams.get('q') === query).length).toBeGreaterThan(0)
    }
    try {
      await use({
        root, search, input, suggestions, items, quick, modal, assertQuiet, settle,
        releaseSuggestions, expectSuggestionRequest,
        holdSuggestions() {
          expect(state.gate).toBeUndefined()
          let release!: () => void
          const promise = new Promise<void>(resolve => { release = resolve })
          state.gate = { promise, release }
        },
        allowExampleIcon() { exampleAllowed = true },
        allowSearchPopup() { popupAllowed = true },
        async open(width = 1440) {
          await page.setViewportSize({ width, height: width < 768 ? 844 : 900 })
          await context.addCookies(['gf_asset_cdn_mode=primary', 'gf_asset_cdn=primary', 'gf_hero_mode=random'].map(pair => {
            const [name, value] = pair.split('=')
            return { name: name!, value: value!, url: app.base }
          }))
          await context.addInitScript(({ origin, recent, custom, steamKey, sample }) => {
            if (location.origin !== origin) return
            localStorage.setItem('theme', 'light')
            localStorage.setItem('recentSites', JSON.stringify(recent))
            localStorage.setItem('navCustomSites', JSON.stringify(custom))
            localStorage.setItem('nav-header-show-quick-access', '1')
            const checkedAt = Date.now()
            localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
              primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
            localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
              china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
          }, { origin: app.base, recent: recentSeed, custom: customSeed, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
          const response = await page.goto('/', { waitUntil: 'load' })
          expect(response?.status()).toBe(200)
          const ssr = await response!.text()
          expect(ssr).toContain('nav-home-page')
          expect(ssr).toContain(desktopKey)
          expect(ssr).toContain(mobileKey)
          await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
          await expect(page.locator('html')).not.toHaveClass(/\bdark\b/)
          await expect(input).toBeVisible()
          await expect.poll(() => root.locator('.nav-header__background--managed img').evaluate((image: HTMLImageElement) =>
            image.complete && image.naturalWidth > 0 && image.currentSrc)).toBe(heroURLs[width < 768 ? 1 : 0])
          // Real Mobile reveal / Desktop prewarm may mount the weather consumer;
          // wait semantically, without owning unrelated Home animation timing.
          await expect(page.locator('.nav-content-shell')).toHaveCount(1)
          await expect.poll(() => weather.length).toBe(1)
          if (width >= 768) {
            await expect(quick).toBeVisible()
            await expect(quick.locator('[title="Archive"] .quick-site-fallback')).toHaveText('A')
            await expect.poll(() => injectedFailed).toEqual([failedIcon])
            await expect.poll(() => icons.length).toBe(4)
            await expect.poll(() => quick.locator('img').evaluateAll(images => images.length === 3
              && images.every(image => image.complete && image.naturalWidth > 0))).toBe(true)
          } else await expect(quick).toBeHidden()
          await settle(search)
          expect(state.requests).toHaveLength(1)
          expect(browserAPI).toEqual([])
          assertQuiet()
        },
        async showSuggestions() {
          await input.focus()
          await input.fill('wolf')
          await expectSuggestionRequest('wolf')
          await expect(items).toHaveText(['wolf furry', 'wolf art', 'wolf game'])
          await expect(items.first()).toHaveClass(/\bsearch-suggestion-item-active\b/)
          await expect(input).toBeFocused()
        },
        async headerClip() {
          await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
          await settle(root)
          const nav = await page.locator('.gf-nav').boundingBox()
          expect(nav).not.toBeNull()
          const viewport = page.viewportSize()!
          const y = Math.ceil(nav!.y + nav!.height)
          assertQuiet()
          return { x: 0, y, width: viewport.width, height: viewport.height - y }
        },
        async searchClip() {
          await settle(search)
          await expect(input).toBeFocused()
          await expect(items).toHaveCount(3)
          await expect(items.first()).toHaveClass(/\bsearch-suggestion-item-active\b/)
          assertQuiet()
          return expandedClip([search, suggestions], 16)
        },
        async modalClip() {
          await settle(modal)
          await expect(modal.locator('#custom-site-name')).toBeFocused()
          await expect(modal.locator('.quick-modal-error')).toHaveText('请输入网站名称')
          assertQuiet()
          return expandedClip([modal], 48)
        },
      })
      assertQuiet()
    } finally {
      releaseSuggestions()
      await testInfo.attach('header-network-evidence', { contentType: 'application/json', body: JSON.stringify({
        upstream: state.requests.map(String), browserAPI, managed, icons, injectedFailed, weather, popups,
        external, unexpectedFailed, errors, popupErrors,
      }, null, 2) })
      headerApp.current = null
      // Playwright owns the context and routes. Never await unrouteAll(wait)
      // while a route handler may still be completing during context teardown.
    }
  },
})

export { expect }
