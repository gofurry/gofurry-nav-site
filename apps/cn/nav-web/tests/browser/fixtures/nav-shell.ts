import { test as base, expect, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'
import { assertHeroHydration } from './hero-lifecycle'

type ShellApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Theme = 'light' | 'dark'
type Clip = { x: number, y: number, width: number, height: number }
type NavShell = {
  nav: Locator
  menu: Locator
  toggle: Locator
  bottomTabs: Locator
  open(options?: { width?: number, theme?: Theme, route?: '/terms' | '/' }): Promise<void>
  assertTheme(theme: Theme): Promise<void>
  assertLinks(locale: 'zh' | 'en', mobile?: boolean): Promise<void>
  language(locale: 'CN' | 'EN', mobile?: boolean): Locator
  assertLanguage(locale: 'CN' | 'EN', mobile?: boolean): Promise<void>
  assertBottomTabs(visible: boolean): Promise<void>
  openBottomMode(): Promise<void>
  scrollTo(top: number): Promise<void>
  settle(target: Locator): Promise<void>
  navClip(includeMenu?: boolean): Promise<Clip>
  bottomClip(): Promise<Clip>
  assertQuiet(): void
}

const weatherURL = 'https://i.tianqi.com/index.php?c=code&id=73&icon=1&num=3&color=d1d5dc'

export const test = base.extend<{ navShell: NavShell }, { shellApp: ShellApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  shellApp: [async ({}, use) => {
    const app = await startInsightsFixtureApp((url: URL) => {
      if (url.pathname === '/api/v2/nav/home') return { data: {
        schema_version: 4, groups: [],
        spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
        saying: null, ping: {}, hero: { desktop: null, mobile: null },
      } }
      // Approved existing Home behavior: null saying triggers this focused fetch.
      if (url.pathname === '/api/v2/nav/home/saying') return { data: { schema_version: 1, saying: null } }
      if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: [] } }
      return { status: 500 }
    })
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ shellApp }, use) => { await use(shellApp.base) },
  navShell: async ({ page, context, shellApp }, use, testInfo) => {
    const errors = captureBrowserErrors(page)
    const external: string[] = [], failed: string[] = []
    const browserAPI: string[] = []
    const isolatedWeather: string[] = [], knownHydration: string[] = []
    let home = false
    let modeCatalogExpected = false
    const upstreamStart = shellApp.requests.length
    page.on('requestfailed', request => failed.push(`${request.url()}: ${request.failure()?.errorText}`))
    page.on('request', request => {
      const url = new URL(request.url())
      if (url.origin === shellApp.base && url.pathname.startsWith('/api/')) browserAPI.push(url.pathname)
    })
    await context.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === shellApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      // Approved test-only integration boundary, not a real weather request or
      // a replacement of any Nav/Hero component. All other external URLs fail.
      if (home && url.href === weatherURL && route.request().isNavigationRequest()
        && route.request().frame().parentFrame() === page.mainFrame()) {
        isolatedWeather.push(url.href)
        return route.fulfill({ status: 200, contentType: 'text/html', body: '' })
      }
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const nav = page.locator('.gf-nav')
    const menu = nav.locator('.gf-nav__mobile-panel')
    const toggle = nav.locator('.gf-nav__mobile-toggle')
    const bottomTabs = page.locator('.mobile-bottom-tabs')
    const language = (locale: 'CN' | 'EN', mobile = false) => nav.locator(mobile ? '.gf-nav__mobile-language' : '.gf-nav__icon-button')
      .filter({ has: page.getByRole('img', { name: locale, exact: true }) })
    const assertQuiet = () => {
      const upstream = shellApp.requests.slice(upstreamStart)
      const expected = home ? ['/api/v2/nav/home', '/api/v2/nav/home/saying'] : []
      if (modeCatalogExpected) expected.push('/api/v2/nav/appearance/patterns')
      expect(upstream.map(url => url.pathname), 'Only explicitly reached Home/Mode integrations may call upstream').toEqual(expected)
      expect(browserAPI, 'Home data must arrive in the SSR payload, never refetch during hydration')
        .toEqual(expected.filter(path => path !== '/api/v2/nav/home'))
      for (const url of upstream) {
        if (url.pathname === '/api/v2/nav/appearance/patterns') expect(url.search).toBe('')
        else expect(url.searchParams.get('lang')).toBe('zh')
      }
      expect(isolatedWeather, 'Only the exact Home weather frame is isolated').toEqual(home ? [weatherURL] : [])
      expect(external, 'No external origins may be requested').toEqual([])
      expect(failed, 'No resource failures are expected').toEqual([])
      // Preserve raw evidence and consume only the one already-verified initial
      // mobile Footer mismatch. A later identical message still fails.
      expect(errors.slice(knownHydration.length), 'No errors beyond the verified initial Footer debt are allowed').toEqual([])
    }
    const assertTheme = async (theme: Theme) => {
      await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
      await expect.poll(() => page.evaluate(() => localStorage.getItem('theme'))).toBe(theme)
    }
    // Only target/subtree motion is relevant. In particular, do not await all
    // document animations when a Nav capture shares the page with Home content.
    const settle = async (target: Locator) => {
      await expect.poll(() => target.locator('img').evaluateAll(images => images.every(image => image.complete && image.naturalWidth > 0))).toBe(true)
      await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
      await page.mouse.move(1, 1)
      await target.evaluate(async element => {
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
        const finite = element.getAnimations({ subtree: true }).filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
        await Promise.all(finite.map(animation => animation.finished.catch(() => {})))
        await document.fonts.ready
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
      })
    }
    try {
      await use({
        nav, menu, toggle, bottomTabs, language, assertTheme, assertQuiet, settle,
        async open({ width = 1440, theme = 'light', route = '/terms' } = {}) {
          home = route === '/'
          await page.setViewportSize({ width, height: width < 768 ? 844 : 900 })
          await context.addInitScript(({ origin, theme, steamKey, sample, home }) => {
            if (location.origin !== origin) return
            // Only seed an empty context: a real locale navigation must retain
            // the theme subsequently saved by the production Theme button.
            if (localStorage.getItem('theme') === null) localStorage.setItem('theme', theme)
            const checkedAt = Date.now()
            localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({
              selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20,
              primaryState: 'success', mirrorState: 'success',
            }))
            localStorage.setItem(steamKey, JSON.stringify({
              version: 1, selected: 'china', checkedAt, sample,
              china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
            }))
            if (home) {
              // Evidence required by the existing narrow Footer-debt guard.
              // Observe the SSR node; never alter Vue state or the rendered DOM.
              const observer = new MutationObserver(() => {
                const node = document.querySelector('.nav-header__background:not([data-hero-pending])')
                if (node) {
                  (window as Window & { initialHeroNode?: Element }).initialHeroNode = node
                  observer.disconnect()
                }
              })
              observer.observe(document, { childList: true, subtree: true })
            }
          }, { origin: shellApp.base, theme, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0], home })
          const response = await page.goto(route, { waitUntil: 'load' })
          expect(response?.status()).toBe(200)
          const ssr = await response!.text()
          expect(ssr).toContain('gf-nav')
          expect(ssr).toContain(home ? 'nav-home-page' : '服务条款')
          await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
          await assertTheme(theme)
          await expect(nav).toHaveCount(1)
          await expect(nav).toBeVisible()
          if (home) {
            await expect(nav).toHaveClass(/\bgf-nav--overlay\b/)
            // Real Mobile reveal / Desktop prewarm mounts content. Neither is
            // faked, and the Nav contract does not wait on its unrelated motion.
            await expect(page.locator('.nav-content-shell')).toHaveCount(1)
            await expect.poll(() => shellApp.requests.slice(upstreamStart).length).toBe(2)
            await expect.poll(() => isolatedWeather.length).toBe(1)
            const initialErrors = [...errors]
            await assertHeroHydration(page, ssr, [...initialErrors])
            knownHydration.push(...initialErrors)
          } else await expect(nav).not.toHaveClass(/\bgf-nav--overlay\b/)
          await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
          assertQuiet()
        },
        async assertLinks(locale, mobile = false) {
          const scope = mobile ? menu : nav.locator('nav')
          const names = locale === 'zh' ? ['站点导航', '热门兽游', '生态观测', '深度兽研'] : ['Navigation', 'Games', 'Ecosystem', 'DeepFurry']
          for (const name of names) await expect(scope.getByRole('link', { name, exact: true })).toBeVisible()
        },
        async assertLanguage(locale, mobile = false) {
          const active = mobile ? /\bgf-nav__mobile-language--active\b/ : /\bgf-nav__icon-button--active\b/
          await expect(language(locale, mobile)).toBeVisible()
          await expect(language(locale, mobile)).toHaveClass(active)
          await expect(language(locale === 'CN' ? 'EN' : 'CN', mobile)).toBeVisible()
          await expect(language(locale === 'CN' ? 'EN' : 'CN', mobile)).not.toHaveClass(active)
        },
        async assertBottomTabs(visible) {
          await expect(bottomTabs).toHaveCount(1)
          await expect.poll(() => bottomTabs.evaluate(el => el.classList.contains('mobile-bottom-tabs--visible'))).toBe(visible)
          await expect(bottomTabs).toHaveAttribute('aria-hidden', String(!visible))
          await expect(bottomTabs).toHaveCSS('pointer-events', visible ? 'auto' : 'none')
          if (visible) await expect(bottomTabs).toHaveCSS('opacity', '0.88')
          const items = bottomTabs.locator('.mobile-bottom-tabs__item')
          await expect(items).toHaveCount(4)
          for (const item of await items.all()) await expect(item).toHaveAttribute('tabindex', visible ? '0' : '-1')
        },
        async openBottomMode() {
          expect(home).toBe(false)
          expect(modeCatalogExpected).toBe(false)
          assertQuiet()
          modeCatalogExpected = true
          await bottomTabs.getByRole('button', { name: /^(模式|Mode)$/ }).click()
          await expect(page.locator('.gf-preferences-modal')).toBeVisible()
          // A real Modal mount eagerly initializes its Background Editor.
          // Permit exactly that one catalog only after the real Mode action.
          await expect.poll(() => shellApp.requests.slice(upstreamStart).length).toBe(1)
          assertQuiet()
        },
        async scrollTo(top) {
          await page.evaluate(value => window.scrollTo({ top: value, behavior: 'instant' }), top)
          await expect.poll(() => page.evaluate(() => window.scrollY)).toBe(top)
        },
        async navClip(includeMenu = false) {
          await settle(nav)
          const navBox = await nav.boundingBox()
          expect(navBox).not.toBeNull()
          const last = includeMenu ? await menu.boundingBox() : navBox
          expect(last).not.toBeNull()
          const viewport = page.viewportSize()!
          const y = Math.max(0, Math.floor(navBox!.y))
          const bottom = Math.min(viewport.height, Math.ceil(last!.y + last!.height + 20))
          expect(navBox!.y).toBeGreaterThanOrEqual(-1)
          expect(last!.y + last!.height).toBeLessThanOrEqual(viewport.height)
          const clip = { x: 0, y, width: viewport.width, height: bottom - y }
          expect(clip.height).toBeGreaterThan(navBox!.height)
          if (home) {
            for (const selector of ['.nav-header__search', '.nav-header__quick-access']) {
              const box = await page.locator(selector).boundingBox()
              expect(box).not.toBeNull()
              expect(box!.y, 'SearchBox/QuickAccess must stay outside the Nav golden').toBeGreaterThanOrEqual(bottom)
            }
          }
          assertQuiet()
          return clip
        },
        async bottomClip() {
          await settle(bottomTabs)
          const box = await bottomTabs.boundingBox()
          expect(box).not.toBeNull()
          const viewport = page.viewportSize()!
          const x = Math.max(0, Math.floor(box!.x - 12)), y = Math.max(0, Math.floor(box!.y - 12))
          const right = Math.min(viewport.width, Math.ceil(box!.x + box!.width + 12))
          const bottom = Math.min(viewport.height, Math.ceil(box!.y + box!.height + 12))
          const clip = { x, y, width: right - x, height: bottom - y }
          expect(clip.width).toBeGreaterThan(box!.width)
          expect(clip.height).toBeGreaterThan(box!.height)
          assertQuiet()
          return clip
        },
      })
      assertQuiet()
    } finally {
      await testInfo.attach('nav-shell-requests.json', {
        body: JSON.stringify({ upstream: shellApp.requests.slice(upstreamStart).map(url => url.href), browserAPI, modeCatalogExpected, isolatedWeather, external, failed, errors, knownHydration }),
        contentType: 'application/json',
      })
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('nav-shell-app.log', { body: shellApp.logs(), contentType: 'text/plain' })
      }
      // No gates or background fixture tasks; Playwright disposes context routes.
    }
  },
})

export { expect }
