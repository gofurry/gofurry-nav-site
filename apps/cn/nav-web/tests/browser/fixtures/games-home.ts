import { test as base, expect, type Locator, type Page } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import type { GameHomeApiResponse, GameV2ListItem, GameV2PriceView } from '../../../app/types/game'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

const assetOrigin = 'https://games-home-assets.example'
const homePath = '/api/v2/game/home'
const collectedAt = '2026-09-18T04:40:00Z'
export const groupNames = ['最近发售', '最近收录', '免费专区', '热门排行'] as const

function price(free: boolean, discount: boolean, region = 'CN'): GameV2PriceView {
  const amount = region === 'CN' ? 4800 : 1200
  return { region, currency: region === 'CN' ? 'CNY' : 'USD', available: true, is_free: free,
    initial_amount: free ? 0 : amount, final_amount: free ? 0 : discount ? amount / 2 : amount,
    discount_percent: discount ? 50 : 0, initial_formatted: '', final_formatted: '',
    collected_at: collectedAt, updated_at: collectedAt }
}

function game(id: number, name: string, options: { free?: boolean, upcoming?: boolean, discount?: boolean } = {}): GameV2ListItem {
  const { free = false, upcoming = false, discount = false } = options
  return { id: String(id), appid: String(id), name, name_zh: name, name_en: `Game ${id}`,
    summary: '与旅伴穿越森林、探索遗迹，在漫长的旅途中发现新的故事。长简介用于保护窄卡片的两行内容区域，文字不应推开评分或相邻卡片。',
    header_url: `${assetOrigin}/covers/${id}.svg`, capsule_url: '', release_date: upcoming ? '' : '2026-09-01',
    release: { availability: upcoming ? 'upcoming' : 'available', precision: 'day', exact_date: upcoming ? '2027-02-14' : '2026-09-01',
      year: upcoming ? 2027 : 2026, month: upcoming ? 2 : 9, quarter: null, window_start: null, window_end: null,
      raw_text: '', observed_at: collectedAt },
    first_available: upcoming ? null : { precision: 'day', exact_date: '2026-09-01', year: 2026, month: 9, quarter: null,
      window_start: '2026-09-01', window_end: '2026-09-01', source: 'steam_backfill', inferred: false },
    developers: ['Forest Studio'], publishers: ['Forest Studio'], platforms: { windows: true, mac: false, linux: true },
    price: price(free, discount), prices: [price(free, discount, 'US')],
    online_count: { count: 1234 + id, peak_count: 4000 + id, status: 'success', collected_at: collectedAt },
    tags: [], avg_score: 4.2, comment_count: 12, updated_at: collectedAt }
}

function makeHome(): GameHomeApiResponse {
  const latest = Array.from({ length: 9 }, (_, i) => game(6101 + i, i === 0
    ? '森林旅途：与星光同行的漫长冒险与永不结束的夏日回忆' : `最近发售 ${i + 1}`, { discount: i === 1 }))
  const recent = Array.from({ length: 9 }, (_, i) => game(6201 + i, `最近收录 ${i + 1}`, { upcoming: i === 0 }))
  const free = Array.from({ length: 9 }, (_, i) => game(6301 + i, `免费游戏 ${i + 1}`, { free: true }))
  const hot = Array.from({ length: 9 }, (_, i) => game(6401 + i, `热门游戏 ${i + 1}`))
  return { panel: { latest_games: latest, updated_games: recent, free_games: free, popular_games: hot,
    top_online: [...hot, ...latest.slice(0, 7)], top_price: [latest[0]!], highest_discount: [latest[1]!],
    low_price: [latest[0]!, latest[1]!], latest_news: [] },
  latest_news: { news_zh: [], news_en: [] }, latest_reviews: [] }
}

function cover(id: string) {
  const colors = ['#394b44', '#4b5269', '#785949']
  return `<svg xmlns="http://www.w3.org/2000/svg" width="460" height="215" viewBox="0 0 460 215"><rect width="460" height="215" fill="${colors[Number(id) % 3]}"/><circle cx="350" cy="62" r="28" fill="#d9c6a3"/><path d="M0 215V155L110 58l120 157M190 215l125-123 145 110v13" fill="#819484"/><path d="M0 195L95 132l78 83H0" fill="#263b37"/></svg>`
}

type Theme = 'light' | 'dark'
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type State = { data: GameHomeApiResponse, upstream: URL[] }
type Worker = { app: App, current: State | null }
export type GamesHomeScene = {
  page: Page, root: Locator, groups: Locator, stats: Locator, sidebar: Locator, dock: Locator, theme: Theme,
  data: GameHomeApiResponse, group(index: number): Locator, cards(index: number): Locator,
  settle(...targets: Locator[]): Promise<void>, assertQuiet(): void, visualReady(): Promise<void>,
}

export const test = base.extend<{ gamesHome: { open(options?: { width?: number, theme?: Theme }): Promise<GamesHomeScene> } }, { gamesHomeApp: Worker }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires destructured fixture arguments.
  gamesHomeApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp((url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Games Home API requires an active scenario')
      state.upstream.push(url)
      if (url.pathname === homePath && url.searchParams.size === 2
        && url.searchParams.get('lang') === 'zh' && url.searchParams.get('region') === 'CN') return { data: state.data }
      return { status: 500 }
    })
    try { await use(worker) }
    finally { worker.current = null; await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ gamesHomeApp }, use) => { await use(gamesHomeApp.app.base) },
  gamesHome: async ({ page, context, gamesHomeApp }, use, testInfo) => {
    const { app } = gamesHomeApp
    const state: State = { data: makeHome(), upstream: [] }
    gamesHomeApp.current = state
    const upstreamStart = app.requests.length
    const errors = captureBrowserErrors(page)
    const external: string[] = [], failed: string[] = [], assets: string[] = [], browserAPI: string[] = []
    const { panel } = state.data
    const allGames = [...panel.latest_games, ...panel.updated_games, ...panel.free_games, ...panel.popular_games!]
    const covers = new Map(allGames.map(item => [item.header_url, item.id]))
    context.on('requestfailed', request => failed.push(`${request.url()}: ${request.failure()?.errorText}`))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin === app.base && url.pathname.startsWith('/api/')) browserAPI.push(`${request.method()} ${url.pathname}${url.search}`)
    })
    await context.route('**/*', async route => {
      const request = route.request(), url = new URL(request.url())
      if (url.origin === app.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      const id = covers.get(url.href)
      if (id && request.method() === 'GET') {
        assets.push(url.href)
        return route.fulfill({ contentType: 'image/svg+xml', body: cover(id) })
      }
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const assertQuiet = () => {
      expect(app.requests.slice(upstreamStart)).toEqual(state.upstream)
      expect(state.upstream.map(url => url.pathname)).toEqual([homePath])
      expect(Object.fromEntries(state.upstream[0]!.searchParams)).toEqual({ lang: 'zh', region: 'CN' })
      expect(browserAPI, 'SSR payload must satisfy hydration, pagination and statistics without browser refetch').toEqual([])
      expect(external).toEqual([])
      expect(failed).toEqual([])
      expect(errors).toEqual([])
    }
    const settle = async (...targets: Locator[]) => {
      for (const target of targets) {
        await expect.poll(() => target.locator('img').evaluateAll(images => images.filter(image => {
          const box = image.getBoundingClientRect()
          return box.width > 0 && box.height > 0 && box.top < innerHeight && box.bottom > 0 && box.left < innerWidth && box.right > 0
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
      await use({ async open({ width = 1440, theme = 'light' } = {}) {
        await page.setViewportSize({ width, height: 900 })
        await context.addInitScript(({ origin, theme, steamKey, sample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
            primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
        const response = await page.goto('/games', { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const ssr = await response!.text()
        const rendered = ssr.slice(ssr.indexOf('<body'), ssr.indexOf('</body>')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g, '')
        for (const name of groupNames) expect(rendered).toContain(name)
        const ssrCards = [...rendered.matchAll(/<p class="game-card__title[^>]*>([^<]*)<\/p>/g)].map(match => match[1]!.trim())
        expect(ssrCards).toEqual([state.data.panel.latest_games, state.data.panel.updated_games, state.data.panel.free_games, state.data.panel.popular_games!]
          .flatMap(list => list.slice(0, 8).map(item => item.name)))
        expect(rendered).toContain(state.data.panel.latest_games[0]!.summary)
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect.poll(() => page.locator('html').evaluate(element => element.classList.contains('dark'))).toBe(theme === 'dark')
        const root = page.locator('.games-page'), groups = root.locator('.game-info-group'), stats = root.locator('.game-stats-section')
        const sidebar = root.locator('.game-sidebar-shell'), dock = root.locator('.game-tool-dock')
        const group = (index: number) => groups.nth(index)
        const cards = (index: number) => group(index).locator('.game-group-page-live .game-card')
        await expect(groups).toHaveCount(4)
        await expect(groups.locator('h3')).toHaveText([...groupNames])
        for (let index = 0; index < 4; index++) await expect(cards(index)).toHaveCount(8)
        await expect(stats.locator('.stats-type-tab--active')).toHaveText('在线人数')
        await settle(root)
        assertQuiet()
        return { page, root, groups, stats, sidebar, dock, theme, data: state.data, group, cards, settle, assertQuiet,
          async visualReady() {
            await page.evaluate(() => { (document.activeElement as HTMLElement | null)?.blur(); window.scrollTo({ top: 0, behavior: 'instant' }) })
            await page.mouse.move(1, 1)
            await settle(root)
            expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
            assertQuiet()
          },
        }
      } })
      assertQuiet()
    } finally {
      await testInfo.attach('games-home-network.json', { contentType: 'application/json', body: JSON.stringify({
        upstream: state.upstream.map(String), browserAPI, assets, external, failed, errors,
      }, null, 2) })
      gamesHomeApp.current = null
    }
  },
})

export { expect }

// Chromium measurements from the unchanged P6 entry build. These intentionally
// freeze effective appearance (including the transparent sidebar), not token intent.
export async function assertGamesHomeAppearance(scene: GamesHomeScene) {
  const { theme, sidebar, dock } = scene
  const dark = theme === 'dark', card = scene.cards(0).first()
  await css(card, { 'background-color': dark ? 'rgba(226, 232, 240, 0.067)' : 'rgba(255, 250, 242, 0.4)',
    'border-color': 'rgba(0, 0, 0, 0)', 'border-width': '1px', 'border-radius': '12px',
    'box-shadow': dark ? 'none' : 'rgba(91, 62, 28, 0.03) 0px 4px 12px 0px',
    'transition-property': 'background-color, border-color' })
  await css(card.locator('.game-card__title'), { 'font-size': '14px', 'line-height': '20px', 'font-weight': '600',
    color: dark ? 'rgba(241, 245, 249, 0.88)' : 'rgba(71, 42, 20, 0.92)' })
  await css(card.locator('.game-card__desc'), { 'font-size': '12px', 'line-height': '16px', height: '32px',
    color: dark ? 'rgba(203, 213, 225, 0.62)' : 'rgba(120, 83, 53, 0.58)' })
  await css(sidebar, { 'background-color': 'rgba(0, 0, 0, 0)', 'background-image': 'none',
    'border-width': '0px', 'border-radius': '0px', 'box-shadow': 'none', 'backdrop-filter': 'none' })
  for (const target of [dock.locator('.game-tool-button'), dock.locator('.game-tool-feedback')]) {
    await css(target, { 'background-color': dark ? 'rgba(15, 23, 42, 0.76)' : 'rgba(255, 255, 255, 0.7)',
      'border-color': dark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(255, 255, 255, 0.55)', 'border-radius': '10.4px',
      'box-shadow': dark ? 'rgba(2, 6, 23, 0.28) 0px 12px 32px 0px' : 'rgba(76, 42, 18, 0.14) 0px 12px 32px 0px',
      'backdrop-filter': 'blur(18px)', color: dark ? 'rgb(219, 228, 240)' : 'rgb(51, 65, 85)' })
  }
  await assertStatsAppearance(scene)
}

export async function assertStatsAppearance(scene: GamesHomeScene) {
  const { stats, theme } = scene, dark = theme === 'dark'
  await css(stats.locator('.stats-type-tab--active'), { 'font-size': '14px', 'line-height': '20px', 'font-weight': '600',
    color: dark ? 'rgba(226, 232, 240, 0.9)' : 'rgb(124, 45, 18)', 'background-color': 'rgba(0, 0, 0, 0)' })
  await css(stats.locator('.stats-type-tab--idle'), { color: dark ? 'rgba(180, 213, 226, 0.62)' : 'rgba(154, 52, 18, 0.74)' })
  await css(stats.locator('.stats-page-tab--active'), { 'background-color': dark ? 'rgba(30, 41, 59, 0.52)' : 'rgba(255, 255, 255, 0.58)',
    'box-shadow': dark ? 'rgba(255, 255, 255, 0.08) 0px 1px 0px 0px inset' : 'rgba(255, 255, 255, 0.62) 0px 1px 0px 0px inset' })
}

export async function assertCardHover(scene: GamesHomeScene) {
  const card = scene.cards(0).first()
  await card.hover()
  await css(card, { 'background-color': scene.theme === 'dark' ? 'rgba(148, 163, 184, 0.18)' : 'rgba(255, 235, 205, 0.82)',
    'border-color': 'rgba(0, 0, 0, 0)', 'box-shadow': scene.theme === 'dark' ? 'none' : 'rgba(91, 62, 28, 0.03) 0px 4px 12px 0px' })
  await scene.page.mouse.move(1, 1)
  await scene.settle(card)
}

async function css(target: Locator, properties: Record<string, string>) {
  for (const [property, value] of Object.entries(properties)) await expect(target).toHaveCSS(property, value)
}
