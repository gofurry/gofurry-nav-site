import { test as base, expect, type Locator, type Page } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import type { NavSiteGroupPageResponse, Site } from '../../../app/types/nav'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

const primary = 'https://site-groups-assets.example'
const groupPath = '/api/v2/nav/site-groups/42/sites'
const pingPath = '/api/v2/nav/home/ping'
const icon = '<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48"><rect width="48" height="48" rx="9" fill="#82654f"/><path d="M10 36V12l14 14 14-14v24" fill="none" stroke="#f4e9dd" stroke-width="4"/></svg>'
const group = { id: '42', name: '社区与平台', info: '兽迷社区、创作平台与交流站点', priority: 1, site_count: 30, detail_path: '/site-groups/42' }
const makeSites = (): Site[] => Array.from({ length: 30 }, (_, index) => ({
  id: String(4201 + index), name: index === 0 ? 'Wolf Community' : index === 29 ? 'Wolf After Dark' : `兽迷站点 ${String(index + 1).padStart(2, '0')}`,
  domain: JSON.stringify({ domain: index === 0 ? ['community.example', 'mirror.example'] : [`site-${index + 1}.example`] }),
  info: index === 0 ? '分享创作、发现同好，连接各地的兽迷社区。' : `第 ${index + 1} 个社区与创作平台的固定简介。`,
  country: 'CN', nsfw: index === 29 ? '1' : '0', welfare: '0', icon: `nav/sites/${4201 + index}/icon/${(4201 + index).toString(16).padStart(32, '0')}.svg`, view_count: 120 + index,
  create_time: '2026-09-18T04:40:00Z', update_time: '2026-09-18T04:40:00Z',
}))
type Theme = 'light' | 'dark'
type Variant = 'ready' | 'missing' | 'empty'
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type State = {
  sites: Site[], variant: Variant, lang: 'zh' | 'en', upstream: URL[],
  pageTwo: Promise<void>, releasePageTwo(): void, allowView: boolean, allowHome: boolean,
}
type Worker = { app: App, current: State | null }
type Scene = {
  page: Page, root: Locator, content: Locator, panel: Locator, cards: Locator, back: Locator, loadMore: Locator, empty: Locator,
  sites: Site[], ssr: string, theme: Theme, assertQuiet(): void, pageCalls(): number[],
  loadAll(beforeRelease?: () => Promise<void>): Promise<void>, openSite(): Promise<void>, goHome(): Promise<void>,
  settle(...targets: Locator[]): Promise<void>, visualClip(): Promise<{ x: number, y: number, width: number, height: number }>,
}
type Options = { width?: number, height?: number, theme?: Theme, variant?: Variant, lang?: 'zh' | 'en', mode?: 'sfw' | 'nsfw' }

export const test = base.extend<{ siteGroups: { open(options?: Options): Promise<Scene> } }, { siteGroupsApp: Worker }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires destructured fixture arguments.
  siteGroupsApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp(async (url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Site Groups API requires an active scenario')
      state.upstream.push(url)
      if (url.pathname === groupPath) {
        const number = Number(url.searchParams.get('page'))
        if (url.searchParams.get('lang') !== state.lang || url.searchParams.get('page_size') !== '24'
          || url.searchParams.size !== 3 || ![1, 2].includes(number)) return { status: 500 }
        if (number === 2) await state.pageTwo
        const ready = state.variant === 'ready'
        const data: NavSiteGroupPageResponse = {
          schema_version: 1, generated_at: '2026-09-18T04:40:00Z', state: state.variant === 'missing' ? 'missing' : 'ready',
          group: { ...group, site_count: ready ? 30 : 0 }, page: number, page_size: 24,
          total: ready ? 30 : 0, has_more: ready && number === 1,
          items: ready ? state.sites.slice((number - 1) * 24, number * 24) : [],
          ...(state.variant === 'missing' ? { reason_messages: ['这个分组的缓存暂时不可用。'] } : {}),
        }
        return { data }
      }
      if (url.pathname === pingPath && !url.search) return { data: {
        schema_version: 1, generated_at: '2026-09-18T04:40:00Z', state: 'ready',
        ping: { 'community.example': JSON.stringify({ status: 'up', delay: '32', loss: '0', time: '2026-09-18T04:39:00Z' }),
          'mirror.example': JSON.stringify({ status: 'down', delay: '-', loss: '100', time: '2026-09-18T04:39:00Z' }) },
      } }
      if (state.allowView && url.pathname === '/api/v2/nav/sites/4201/view' && !url.search) return { data: { site_id: 4201, view_count: 4321 } }
      if (state.allowHome && url.pathname === '/api/v2/nav/home' && url.search === `?lang=${state.lang}`) return { data: {
        schema_version: 4, generated_at: '2026-09-18T04:40:00Z', cache_state: {}, groups: [],
        spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] }, ping: {},
        saying: { content: '连接社区，分享创作。', author: 'GoFurry', language: state.lang }, hero: { desktop: null, mobile: null },
      } }
      return { status: 500 }
    }, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: 'https://site-groups-mirror.example' })
    try { await use(worker) }
    finally { worker.current?.releasePageTwo(); worker.current = null; await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ siteGroupsApp }, use) => { await use(siteGroupsApp.app.base) },
  siteGroups: async ({ page, context, siteGroupsApp }, use, testInfo) => {
    const { app } = siteGroupsApp
    let releasePageTwo = () => {}
    const state: State = { sites: makeSites(), variant: 'ready', lang: 'zh', upstream: [], allowView: false, allowHome: false,
      pageTwo: new Promise<void>(resolve => { releasePageTwo = resolve }), releasePageTwo: () => releasePageTwo() }
    siteGroupsApp.current = state
    const upstreamStart = app.requests.length
    const errors = captureBrowserErrors(page)
    const external: string[] = [], failed: string[] = [], assets: string[] = [], popups: string[] = []
    const browserAPI: Array<{ method: string, path: string, query: string }> = []
    const popupErrors: string[][] = []
    context.on('page', popup => { if (popup !== page) popupErrors.push(captureBrowserErrors(popup)) })
    context.on('requestfailed', request => failed.push(`${request.url()}: ${request.failure()?.errorText}`))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin === app.base && url.pathname.startsWith('/api/')) browserAPI.push({ method: request.method(), path: url.pathname, query: url.search })
    })
    await context.route('**/*', async route => {
      const request = route.request(), url = new URL(request.url())
      if (url.origin === app.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      if (request.method() === 'GET' && state.sites.some(site => url.href === `${primary}/${site.icon}`)) {
        assets.push(url.href)
        return route.fulfill({ contentType: 'image/svg+xml', body: icon })
      }
      if (state.allowView && url.href === 'https://community.example/' && request.method() === 'GET' && request.isNavigationRequest()) {
        popups.push(url.href)
        return route.fulfill({ contentType: 'text/html', body: '<title>Seeded Site Groups navigation</title>' })
      }
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const pageCalls = () => state.upstream.filter(url => url.pathname === groupPath).map(url => Number(url.searchParams.get('page')))
    const assertQuiet = () => {
      expect(app.requests.slice(upstreamStart)).toEqual(state.upstream)
      const pages = pageCalls()
      expect(pages).toEqual(pages.length === 1 ? [1] : [1, 2])
      for (const url of state.upstream.filter(url => url.pathname === groupPath)) {
        expect(Object.fromEntries(url.searchParams)).toEqual({ lang: state.lang, page: String(Number(url.searchParams.get('page'))), page_size: '24' })
      }
      expect(state.upstream.filter(url => url.pathname === pingPath).map(url => url.search)).toEqual([''])
      expect(state.upstream.filter(url => ![groupPath, pingPath].includes(url.pathname)).map(url => url.pathname + url.search)).toEqual([
        ...(state.allowView ? ['/api/v2/nav/sites/4201/view'] : []), ...(state.allowHome ? [`/api/v2/nav/home?lang=${state.lang}`] : []),
      ])
      expect(browserAPI.filter(call => call.path === groupPath).map(call => ({ method: call.method, page: new URLSearchParams(call.query).get('page') })))
        .toEqual(pages.length === 1 ? [] : [{ method: 'GET', page: '2' }])
      expect(browserAPI.filter(call => call.path !== groupPath)).toEqual([
        { method: 'GET', path: pingPath, query: '' },
        ...(state.allowView ? [{ method: 'POST', path: '/api/v2/nav/sites/4201/view', query: '' }] : []),
        ...(state.allowHome ? [{ method: 'GET', path: '/api/v2/nav/home', query: `?lang=${state.lang}` }] : []),
      ])
      expect(popups).toEqual(state.allowView ? ['https://community.example/'] : [])
      expect(external).toEqual([])
      expect(failed).toEqual([])
      expect([...errors, ...popupErrors.flat()]).toEqual([])
    }
    const settle = async (...targets: Locator[]) => {
      for (const target of targets) {
        await expect.poll(() => target.locator('img').evaluateAll(images => images.filter(image => {
          const box = image.getBoundingClientRect()
          return box.width > 0 && box.height > 0 && box.top < innerHeight && box.bottom > 0
        }).every(image => image.complete && image.naturalWidth > 0))).toBe(true)
        await target.evaluate(async element => {
          await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
          await Promise.all(element.getAnimations({ subtree: true }).filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
            .map(animation => animation.finished.catch(() => {})))
          await document.fonts.ready
          await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
        })
      }
    }
    try {
      await use({ async open({ width = 1440, height = 900, theme = 'light', variant = 'ready', lang = 'zh', mode = 'sfw' } = {}) {
        state.variant = variant; state.lang = lang
        await page.setViewportSize({ width, height })
        await context.addInitScript(({ origin, theme, mode, steamKey, sample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          localStorage.setItem('mode', mode)
          localStorage.setItem('nav-header-show-quick-access', '0')
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
            primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, mode, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
        const response = await page.goto(`${lang === 'en' ? '/en' : ''}/site-groups/42`, { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const ssr = await response!.text()
        // Restrict SSR evidence to rendered markup, not the serialized Nuxt payload.
        const rendered = ssr.slice(ssr.indexOf('<body'), ssr.indexOf('</body>')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g, '')
        expect(rendered).toContain(group.name)
        expect(rendered).toContain(group.info)
        if (variant === 'ready') {
          for (const site of state.sites.slice(0, 24)) expect(rendered).toContain(site.name)
          expect(rendered).toMatch(/site-group-total[^>]*>\s*30\s/)
          expect(rendered).not.toContain('Wolf After Dark')
        }
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect.poll(() => page.locator('html').evaluate(element => element.classList.contains('dark'))).toBe(theme === 'dark')
        await expect.poll(() => state.upstream.filter(url => url.pathname === pingPath).length).toBe(1)
        const content = page.locator('main.site-group-content')
        const root = content.locator('..'), panel = content.locator(':scope > section'), cards = panel.locator('.nav-site-card')
        const back = content.getByRole('button', { name: lang === 'en' ? 'Back Home' : '返回首页', exact: true })
        const loadMore = panel.getByRole('button')
        const empty = panel.getByText(variant === 'missing' ? '这个分组的缓存暂时不可用。' : '这个分组下暂时还没有可展示的站点。', { exact: true })
        await expect(content.getByRole('heading', { level: 1 })).toHaveText(group.name)
        await expect(content.locator('.site-group-summary')).toHaveText(group.info)
        await expect(content.locator('.site-group-total')).toHaveText(variant === 'ready' ? /^30\s/ : /^0\s/)
        await expect(cards).toHaveCount(variant === 'ready' ? 24 : 0)
        if (variant === 'ready') await expect(cards.first().locator('img')).toHaveAttribute('src', `${primary}/${state.sites[0]!.icon}`)
        await settle(content)
        assertQuiet()
        return {
          page, root, content, panel, cards, back, loadMore, empty, sites: state.sites, ssr, theme, assertQuiet, settle, pageCalls,
          async loadAll(beforeRelease) {
            expect(pageCalls(), 'Button contract starts before the observer requests page two').toEqual([1])
            // Real keyboard activation without scrolling into the observer's rootMargin.
            await loadMore.evaluate(element => (element as HTMLElement).focus({ preventScroll: true }))
            await expect(loadMore).toBeFocused()
            const next = page.waitForResponse(response => new URL(response.url()).pathname === groupPath && new URL(response.url()).searchParams.get('page') === '2')
            await page.keyboard.press('Enter')
            await expect.poll(pageCalls).toEqual([1, 2])
            await expect(loadMore).toBeDisabled()
            await beforeRelease?.()
            state.releasePageTwo()
            const result = await next
            const body = (await result.json()).data
            expect(body.items).toHaveLength(6)
            expect(body.has_more).toBe(false)
            await expect(cards).toHaveCount(30)
            await expect(cards.locator('h3')).toHaveText(state.sites.map(site => site.name))
            // Each unique fixture ID has a unique name and managed icon key;
            // assert the rendered identities without reaching into Vue internals.
            expect(await cards.locator('img').evaluateAll(images => images.map(image => image.getAttribute('src'))))
              .toEqual(state.sites.map(site => `${primary}/${site.icon}`))
            await expect(loadMore).toHaveCount(0)
            assertQuiet()
          },
          async openSite() {
            state.allowView = true
            const view = page.waitForResponse(response => new URL(response.url()).pathname === '/api/v2/nav/sites/4201/view')
            const popupEvent = page.waitForEvent('popup')
            await cards.first().click()
            const popup = await popupEvent
            await popup.waitForLoadState('load')
            await expect(popup).toHaveURL('https://community.example/')
            expect((await (await view).json()).data.view_count).toBe(4321)
            expect(await page.evaluate(() => JSON.parse(localStorage.getItem('recentSites')!)[0])).toEqual({ id: '4201', name: 'Wolf Community', url: 'https://community.example' })
            await popup.close()
            assertQuiet()
          },
          async goHome() {
            state.allowHome = true
            await back.click()
            await expect(page).toHaveURL(`${app.base}${lang === 'en' ? '/en' : '/'}`)
            await expect(page.locator('.nav-home-page')).toBeVisible()
            await expect.poll(() => state.upstream.filter(url => url.pathname === '/api/v2/nav/home').length).toBe(1)
            assertQuiet()
          },
          async visualClip() {
            await page.mouse.move(1, 1)
            await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
            await settle(content)
            expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
            const box = await content.boundingBox()
            expect(box).not.toBeNull()
            const y = Math.max(0, Math.floor(box!.y))
            assertQuiet()
            return { x: 0, y, width, height: Math.min(700, height - y) }
          },
        }
      } })
      assertQuiet()
    } finally {
      // Wide grids can naturally prefetch page two via IntersectionObserver.
      // Hold only that API response while first-page pixels are captured; never
      // replace the observer, mutate Vue state, or leave a gate held at teardown.
      state.releasePageTwo()
      if (pageCalls().includes(2) && !page.isClosed()) await expect(page.locator('main.site-group-content > section button')).toHaveCount(0)
      if (testInfo.status === testInfo.expectedStatus) assertQuiet()
      await testInfo.attach('site-groups-network.json', { contentType: 'application/json', body: JSON.stringify({
        upstream: state.upstream.map(String), browserAPI, assets, popups, external, failed, errors, popupErrors,
      }, null, 2) })
      siteGroupsApp.current = null
    }
  },
})

// Measured in real Chromium before migration. Shell overrides Games' nominal
// page backgrounds: the actual consumer root is transparent in both themes.
export async function assertSiteGroupAppearance(scene: Scene) {
  const { root, content, panel, cards, back, theme } = scene
  const dark = theme === 'dark'
  const title = dark ? 'rgba(241, 245, 249, 0.88)' : 'rgba(71, 42, 20, 0.92)'
  const accent = dark ? 'rgba(190, 208, 222, 0.72)' : 'rgba(124, 45, 18, 0.88)'
  const muted = dark ? 'rgba(203, 213, 225, 0.62)' : 'rgba(78, 60, 46, 0.76)'
  await css(root, { 'background-color': 'rgba(0, 0, 0, 0)', color: dark ? 'rgba(226, 232, 240, 0.88)' : 'rgba(57, 53, 48, 0.84)' })
  await css(content.locator('h1'), { color: title, 'font-size': '30px', 'font-weight': '600' })
  await css(content.locator('.site-group-total'), { color: accent, 'font-size': '14px', 'font-weight': '600' })
  await css(content.locator('.site-group-summary'), { color: muted, 'font-size': '14px', 'line-height': '24px' })
  await css(panel, { 'border-color': 'rgba(0, 0, 0, 0)', 'background-color': 'rgba(0, 0, 0, 0)', 'box-shadow': 'none' })
  await css(back, { color: dark ? 'rgba(226, 232, 240, 0.88)' : accent, 'background-color': 'rgba(0, 0, 0, 0)' })
  if (await cards.count()) {
    await css(cards.first(), { 'background-color': dark ? 'rgb(18, 30, 48)' : 'rgb(255, 244, 226)',
      'box-shadow': dark ? 'rgba(0, 0, 0, 0) 0px 0px 0px 0px inset' : 'rgba(126, 92, 58, 0.18) 0px 0px 0px 1px inset' })
    await css(cards.first().locator('h3'), { color: title, 'font-size': '16px', 'line-height': '24px', 'font-weight': '500' })
    await css(cards.first().locator('p'), { color: muted, 'font-size': '12px', 'line-height': '16.2px' })
  }
}

export async function css(target: Locator, properties: Record<string, string>) {
  for (const [property, value] of Object.entries(properties)) await expect(target).toHaveCSS(property, value)
}

export { expect }
