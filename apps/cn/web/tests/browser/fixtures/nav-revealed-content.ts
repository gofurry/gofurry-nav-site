import { test as base, expect, type Locator, type Page } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { captureBrowserErrors } from './browser-errors'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import type { Group, NavHomeResponse, Site } from '../../../app/types/nav'

export const fixedNow = '2026-09-18T12:40:00+08:00'
export const quote = { author: 'GoFurry Fixture', content: '探索更多兽人世界', language: 'zh' as const }
export const weatherURL = 'https://i.tianqi.com/index.php?c=code&id=73&icon=1&num=3&color=d1d5dc'
const primary = 'https://revealed-primary.example'
const desktopKey = `nav/hero/desktop/${'a'.repeat(32)}.avif`
const mobileKey = `nav/hero/mobile/${'b'.repeat(32)}.avif`
const heroURLs = [primary + '/' + desktopKey, primary + '/' + mobileKey]
const heroSVG = '<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="1000"><rect width="1600" height="1000" fill="#697c83"/><path d="M0 1000V620L380 280 760 650 1160 400 1600 720V1000" fill="#465f64"/></svg>'
const iconSVG = '<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64"><rect width="64" height="64" rx="14" fill="#719489"/><path d="M16 48V16l16 16 16-16v32" fill="none" stroke="#fff5e9" stroke-width="6"/></svg>'

function site(id: number, name: string, domains = [`site-${id}.example`], nsfw = false): Site {
  return { id: String(id), name, domain: JSON.stringify({ domain: domains }),
    info: `${name} · 发现、交流与创作`, country: 'CN', nsfw: nsfw ? '1' : '0', welfare: '0',
    icon: `nav/sites/${id}/icon/${id.toString(16).padStart(32, '0')}.svg`,
    view_count: 1200 + id, create_time: '2026-09-12T08:30:00+08:00', update_time: '2026-09-18T09:00:00+08:00' }
}

function scenarioData() {
  const community = site(101, 'Wolf Community', ['community.example', 'mirror.example'])
  const archive = site(102, 'Wolf Archive')
  const adult = site(299, 'Wolf After Dark', ['after-dark.example'], true)
  const featured = Array.from({ length: 8 }, (_, i) => site(301 + i, `精选资源 ${i + 1}`))
  const adultFeatured = site(399, 'Night Showcase', ['night-showcase.example'], true)
  const directory = [community, archive, adult, site(205, 'Other Resource')]
  const groups: Group[] = [
    { id: '11', name: '社区', info: '交流创作、分享发现，与同好一起探索。', priority: 1,
      site_count: 12, has_more: true, detail_path: '/site-groups/11',
      sites: [community, archive, site(103, 'Art Library'), site(104, 'Creative Forum')] },
    { id: '12', name: '工具', info: '帮助创作与探索的实用工具。', priority: 2,
      site_count: 3, has_more: false, detail_path: '/site-groups/12',
      sites: [site(201, 'Palette Studio'), site(202, 'Writing Desk'), adult] },
  ]
  const home: NavHomeResponse = {
    schema_version: 4, generated_at: '2026-09-18T04:40:00Z', cache_state: {}, groups,
    spotlight: { page_size: 6, featured: [featured[0]!, adultFeatured, ...featured.slice(1)],
      popular: Array.from({ length: 7 }, (_, i) => site(401 + i, `热门资源 ${i + 1}`)),
      latest: Array.from({ length: 6 }, (_, i) => site(501 + i, `新收录 ${i + 1}`)),
      random: Array.from({ length: 6 }, (_, i) => site(601 + i, `随机发现 ${i + 1}`)) },
    saying: { ...quote },
    ping: { 'community.example': JSON.stringify({ status: 'up', delay: '32', loss: '0', time: '2026-09-18T04:39:00Z' }),
      'mirror.example': JSON.stringify({ status: 'down', delay: '-', loss: '100', time: '2026-09-18T04:39:00Z' }) },
    hero: { desktop: { id: '1', object_key: desktopKey }, mobile: { id: '2', object_key: mobileKey } },
  }
  return { home, directory, community, archive, adult, featured, adultFeatured }
}

export const siteURL = (value: Site) => `https://${JSON.parse(value.domain).domain[0]}`
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type RequestEvidence = { method: string, path: string }
type State = {
  data: ReturnType<typeof scenarioData>
  upstream: URL[]
  upstreamEnd?: number
  directoryAllowed: boolean
  expectedViews: Set<string>
  expectedPopups: Set<string>
  browserAPI: RequestEvidence[]
  assets: string[]
  weather: string[]
  popups: string[]
  external: string[]
  failures: string[]
  pages: Array<{ page: Page, errors: string[] }>
}
type Worker = { app: App, current: State | null }
type OpenOptions = { width?: number, theme?: 'light' | 'dark', mode?: 'sfw' | 'nsfw' }
type Scene = {
  page: Page
  data: State['data']
  ssr: string
  shell: Locator
  content: Locator
  bar: Locator
  quoteTrigger: Locator
  author: Locator
  groups: Locator
  groupPopover: Locator
  sitePopover: Locator
  dock: Locator
  rail: Locator
  searchPanel: Locator
  searchInput: Locator
  results: Locator
  panels: Locator
  spotlight: Locator
  featured: Locator
  popular: Locator
  card(name: string): Locator
  reveal(): Promise<void>
  alignContent(): Promise<void>
  openSearch(): Promise<void>
  directoryCount(): number
  hoverTopCard(): Promise<Locator>
  clickSite(target: Locator, value: Site, touchView: boolean): Promise<void>
  settle(...targets: Locator[]): Promise<void>
  inside(target: Locator): Promise<void>
  clip(targets: Locator[], margin?: number): Promise<{ x: number, y: number, width: number, height: number }>
  assertQuiet(): void
}

export const test = base.extend<{ revealed: { open(options?: OpenOptions): Promise<Scene> } }, { revealedApp: Worker }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires destructured fixture arguments.
  revealedApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp((url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Revealed Content API requires an active scenario')
      state.upstream.push(url)
      if (url.pathname === '/api/v2/nav/home' && url.search === '?lang=zh') return { data: state.data.home }
      if (state.directoryAllowed && url.pathname === '/api/v2/nav/sites/directory' && url.search === '?lang=zh') return { data: state.data.directory }
      for (const id of state.expectedViews) {
        if (url.pathname === `/api/v2/nav/sites/${id}/view` && !url.search) return { data: { site_id: Number(id), view_count: 4321 } }
      }
      return { status: 500 }
    }, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: 'https://revealed-mirror.example' })
    try { await use(worker) }
    finally { worker.current = null; await worker.app.close() }
  }, { scope: 'worker' }],
  timezoneId: 'Asia/Shanghai',
  baseURL: async ({ revealedApp }, use) => { await use(revealedApp.app.base) },
  revealed: async ({ context, revealedApp }, use, testInfo) => {
    const { app } = revealedApp
    const scenes: Array<{ state: State, scene: Scene }> = []
    const evidence: State[] = []
    let active: State | null = null
    context.on('page', page => { active?.pages.push({ page, errors: captureBrowserErrors(page) }) })
    context.on('requestfailed', request => active?.failures.push(`${request.url()}: ${request.failure()?.errorText}`))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin === app.base && url.pathname.startsWith('/api/')) active?.browserAPI.push({ method: request.method(), path: url.pathname + url.search })
    })
    await context.route('**/*', async route => {
      const url = route.request().url()
      if (new URL(url).origin === app.base || /^(data|blob):/.test(url)) return route.continue()
      const state = active
      if (!state) return route.abort('blockedbyclient')
      const knownSites = [...state.data.home.groups.flatMap(group => group.sites),
        ...Object.values(state.data.home.spotlight).filter(Array.isArray).flat(), ...state.data.directory] as Site[]
      if (heroURLs.includes(url) || knownSites.some(value => url === primary + '/' + value.icon)) {
        state.assets.push(url)
        return route.fulfill({ contentType: 'image/svg+xml', body: heroURLs.includes(url) ? heroSVG : iconSVG })
      }
      if (url === weatherURL && route.request().isNavigationRequest() && route.request().frame().parentFrame()) {
        state.weather.push(url)
        return route.fulfill({ contentType: 'text/html', body: '' })
      }
      if (state.expectedPopups.has(url) && route.request().isNavigationRequest() && route.request().method() === 'GET') {
        state.popups.push(url)
        return route.fulfill({ contentType: 'text/html', body: '<title>Seeded site navigation boundary</title>' })
      }
      state.external.push(url)
      await route.abort('blockedbyclient')
    })
    try {
      await use({ async open({ width = 1440, theme = 'light', mode = 'sfw' } = {}) {
        // A second open starts a fresh document/storage/data scenario, never a
        // mode-change event or internal Vue mutation. Close old timers first.
        const previous = scenes.at(-1)
        if (previous) {
          previous.scene.assertQuiet()
          for (const { page } of previous.state.pages) await page.close()
          previous.scene.assertQuiet()
          previous.state.upstreamEnd = app.requests.length
        }
        const state: State = { data: scenarioData(), upstream: [], directoryAllowed: false,
          expectedViews: new Set(), expectedPopups: new Set(), browserAPI: [], assets: [],
          weather: [], popups: [], external: [], failures: [], pages: [] }
        active = state
        evidence.push(state)
        revealedApp.current = state
        const upstreamStart = app.requests.length
        await context.clearCookies()
        await context.addCookies(['gf_asset_cdn_mode=primary', 'gf_asset_cdn=primary', 'gf_hero_mode=random'].map(pair => {
          const [name, value] = pair.split('=')
          return { name: name!, value: value!, url: app.base }
        }))
        const page = await context.newPage()
        await page.setViewportSize({ width, height: 900 })
        await page.clock.setFixedTime(new Date(fixedNow))
        await page.addInitScript(({ origin, theme, mode, recent, steamKey, sample }) => {
          if (location.origin !== origin) return
          localStorage.clear()
          localStorage.setItem('theme', theme)
          localStorage.setItem('mode', mode)
          localStorage.setItem('recentSites', JSON.stringify([recent]))
          // Header/QuickAccess is P5.2-owned. Use its real saved preference,
          // avoiding an unrelated favicon boundary in this content contract.
          localStorage.setItem('nav-header-show-quick-access', '0')
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
            primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, mode, recent: { id: state.data.featured[0]!.id,
          name: state.data.featured[0]!.name, url: siteURL(state.data.featured[0]!) },
          steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
        const shell = page.locator('.nav-content-shell')
        const content = page.locator('.nav-content')
        const bar = page.locator('.nav-transition-bar')
        const quoteTrigger = page.locator('.nav-transition-bar__quote')
        const author = page.locator('.nav-transition-bar__author')
        const groups = content.locator('.nav-group-section')
        const groupPopover = page.locator('.group-popover')
        const sitePopover = page.locator('.site-popover')
        const dock = page.locator('.nav-tool-dock')
        const rail = dock.locator('.nav-tool-rail')
        const searchPanel = dock.getByRole('region', { name: '搜索站点', exact: true })
        const searchInput = searchPanel.getByPlaceholder('搜索站点名称或简介...')
        const results = searchPanel.locator('.nav-tool-results button')
        const panels = content.locator('.spotlight-panel')
        const spotlight = content.locator(':scope > section')
        const featured = panels.filter({ has: page.getByRole('heading', { name: '精选站点', exact: true }) })
        const popular = panels.filter({ has: page.getByRole('heading', { name: '热门站点', exact: true }) })
        const card = (name: string) => content.locator('.nav-site-card').filter({ has: page.getByRole('heading', { name, exact: true }) })
        const settle = async (...targets: Locator[]) => {
          for (const target of targets) {
            await expect.poll(() => target.locator('img').evaluateAll(images => images.filter(image => {
              const rect = image.getBoundingClientRect()
              return rect.width > 0 && rect.height > 0 && rect.top < innerHeight && rect.bottom > 0 && rect.left < innerWidth && rect.right > 0
            }).every(image => image.complete && image.naturalWidth > 0))).toBe(true)
            await target.evaluate(async element => {
              await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
              await Promise.all(element.getAnimations({ subtree: true })
                .filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
                .map(animation => animation.finished.catch(() => {})))
              await document.fonts.ready
              await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
            })
          }
        }
        const inside = async (target: Locator) => {
          const box = await target.boundingBox()
          expect(box).not.toBeNull()
          const viewport = page.viewportSize()!
          expect(box!.x).toBeGreaterThanOrEqual(0)
          expect(box!.y).toBeGreaterThanOrEqual(0)
          expect(box!.x + box!.width).toBeLessThanOrEqual(viewport.width)
          expect(box!.y + box!.height).toBeLessThanOrEqual(viewport.height)
        }
        const assertQuiet = () => {
          expect(app.requests.slice(upstreamStart, state.upstreamEnd)).toEqual(state.upstream)
          expect(state.upstream.map(url => url.pathname + url.search)).toEqual([
            '/api/v2/nav/home?lang=zh',
            ...(state.directoryAllowed ? ['/api/v2/nav/sites/directory?lang=zh'] : []),
            ...[...state.expectedViews].map(id => `/api/v2/nav/sites/${id}/view`),
          ])
          expect(state.browserAPI).toEqual([
            ...(state.directoryAllowed ? [{ method: 'GET', path: '/api/v2/nav/sites/directory?lang=zh' }] : []),
            ...[...state.expectedViews].map(id => ({ method: 'POST', path: `/api/v2/nav/sites/${id}/view` })),
          ])
          expect(state.weather).toEqual([weatherURL])
          expect(state.popups).toEqual([...state.expectedPopups])
          expect(state.external).toEqual([])
          expect(state.failures).toEqual([])
          expect(state.pages.flatMap(value => value.errors)).toEqual([])
        }
        const response = await page.goto('/', { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const ssr = await response!.text()
        expect(ssr).toContain(desktopKey)
        expect(ssr).toContain(mobileKey)
        expect(ssr).toContain('Wolf Community')
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect.poll(() => page.locator('html').evaluate(element => element.classList.contains('dark'))).toBe(theme === 'dark')
        await expect.poll(() => page.locator('.nav-header__background--managed img').evaluate((image: HTMLImageElement) =>
          image.complete && image.naturalWidth > 0 && image.currentSrc)).toBe(heroURLs[0])
        await expect(dock).toHaveCount(0)
        const scene: Scene = {
          page, data: state.data, ssr, shell, content, bar, quoteTrigger, author, groups, groupPopover, sitePopover,
          dock, rail, searchPanel, searchInput, results, panels, spotlight, featured, popular, card, settle, inside, assertQuiet,
          directoryCount: () => state.upstream.filter(url => url.pathname.endsWith('/directory')).length,
          async reveal() {
            await page.mouse.move(20, 300)
            await page.mouse.wheel(0, 300)
            await expect(dock).toHaveCount(1)
            await expect(content).toBeVisible()
            await expect.poll(() => shell.evaluate(element => {
              const target = Math.min(element.getBoundingClientRect().top + scrollY, document.documentElement.scrollHeight - innerHeight)
              return Math.abs(scrollY - target)
            })).toBeLessThanOrEqual(1)
            await expect.poll(() => state.weather.length).toBe(1)
            await settle(content, bar)
            assertQuiet()
          },
          async alignContent() {
            await shell.evaluate(element => element.scrollIntoView({ behavior: 'instant', block: 'start' }))
            await expect.poll(() => shell.evaluate(element => Math.abs(element.getBoundingClientRect().top))).toBeLessThanOrEqual(1)
          },
          async openSearch() {
            state.directoryAllowed = true
            await dock.getByRole('button', { name: '搜索站点', exact: true }).click()
            await expect(searchPanel).toBeVisible()
            await expect(searchInput).toBeFocused()
            await expect.poll(() => scene.directoryCount()).toBe(1)
          },
          async hoverTopCard() {
            const target = card(state.data.community.name)
            await target.evaluate(element => window.scrollBy({ top: element.getBoundingClientRect().top - 720, behavior: 'instant' }))
            await target.hover()
            await expect(sitePopover).toHaveClass(/\bsite-popover--visible\b/)
            await settle(target, sitePopover)
            await inside(sitePopover)
            const cardBox = await target.boundingBox(), box = await sitePopover.boundingBox()
            expect(box!.y + box!.height).toBeLessThanOrEqual(cardBox!.y)
            return target
          },
          async clickSite(target, value, touchView) {
            const url = new URL(siteURL(value)).href
            state.expectedPopups.add(url)
            if (touchView) state.expectedViews.add(value.id)
            const popupPromise = page.waitForEvent('popup')
            const viewResponse = touchView ? page.waitForResponse(response => new URL(response.url()).pathname === `/api/v2/nav/sites/${value.id}/view`) : undefined
            await target.click()
            const popup = await popupPromise
            await popup.waitForLoadState('load')
            await expect(popup).toHaveURL(url)
            if (viewResponse) {
              const response = await viewResponse
              expect(response.request().method()).toBe('POST')
              expect((await response.json()).data.view_count).toBe(4321)
            }
            expect(await page.evaluate(() => JSON.parse(localStorage.getItem('recentSites')!)[0])).toEqual({ id: value.id, name: value.name, url: siteURL(value) })
            assertQuiet()
          },
          async clip(targets, margin = 12) {
            await settle(...targets)
            const boxes = await Promise.all(targets.map(target => target.boundingBox()))
            for (const box of boxes) expect(box).not.toBeNull()
            const viewport = page.viewportSize()!
            const x = Math.max(0, Math.floor(Math.min(...boxes.map(box => box!.x)) - margin))
            const y = Math.max(0, Math.floor(Math.min(...boxes.map(box => box!.y)) - margin))
            const right = Math.min(viewport.width, Math.ceil(Math.max(...boxes.map(box => box!.x + box!.width)) + margin))
            const bottom = Math.min(viewport.height, Math.ceil(Math.max(...boxes.map(box => box!.y + box!.height)) + margin))
            expect(right).toBeGreaterThan(x)
            expect(bottom).toBeGreaterThan(y)
            expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
            assertQuiet()
            return { x, y, width: right - x, height: bottom - y }
          },
        }
        scenes.push({ state, scene })
        return scene
      } })
      for (const { scene } of scenes) scene.assertQuiet()
    } finally {
      await testInfo.attach('revealed-content-network-evidence', { contentType: 'application/json', body: JSON.stringify(evidence.map(state => ({
        upstream: state.upstream.map(String), browserAPI: state.browserAPI, assets: state.assets, weather: state.weather,
        popups: state.popups, external: state.external, failures: state.failures, errors: state.pages.flatMap(value => value.errors),
      })), null, 2) })
      // No held gates; Playwright disposes this test's context and registered routes.
      revealedApp.current = null
      active = null
    }
  },
})

export { expect }
