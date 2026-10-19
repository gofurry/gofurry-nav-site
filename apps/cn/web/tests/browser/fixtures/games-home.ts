import { test as base, expect, type Locator, type Page } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import type { GameHomeApiResponse, GameV2ListItem, GameV2NewsItem, GameV2PriceView, GameShowcaseSnapshot, GameShowcaseEvent } from '../../../app/types/game'
import { collectionHome } from './game-collections-data'
import { makeShowcase, showcaseOrigins, type ShowcaseScenario } from './games-home-showcase-data'
import { steamSharedAssetCandidates } from '../../../app/utils/steamAssets'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

const assetOrigin = 'https://games-home-assets.example'
const homePath = '/api/v2/game/home'
const showcasePath = `${homePath}/showcase`
const collectionsPath = '/api/v2/game/collections/home'
type CollectionsScenario = 'populated' | 'empty' | 'sparse' | 'failure' | 'transient' | 'slow'
const eventPath = `${showcasePath}/events`
const collectedAt = '2026-09-18T04:40:00Z'
export const groupNames = ['最近发售', '最近收录', '免费专区', '热门排行'] as const
const englishGroupNames = ['Latest Release', 'Recently Added', 'Free to Play', 'Hot Ranking'] as const
export const gamesHomeClosureTime = '2026-09-18T04:40:00.000Z'

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

type Dataset = 'core' | 'news-single' | 'news-populated' | 'reviews-populated' | 'layout-stress'
type Language = 'zh' | 'en'
type OpenOptions = { width?: number, height?: number, theme?: Theme, locale?: Language, dataset?: Dataset, showcase?: ShowcaseScenario, collections?: CollectionsScenario, mode?: string }

function makeNews(lang: Language): GameV2NewsItem[] {
  const titles = lang === 'zh'
    ? ['森林旅途：秋日更新 & 新章节', '星港漫游：开发日志第二期', '山谷来信：试玩版现已开放']
    : ['Forest Journey: Autumn Update & New Chapter', 'Starport: Developer Diary Two', 'Valley Letters: Demo Available']
  const dates = ['2026-09-17T12:30:00Z', '2026-09-16T08:15:00Z', '2026-09-15T16:05:00Z']
  return titles.map((headline, index) => ({
    id: `news-${lang}-${index + 1}`, game_id: String(6101 + index), appid: String(6101 + index), lang,
    game_name: ['森林旅途', '星港漫游', '山谷来信'][index]!, header_url: `${assetOrigin}/news/${index + 1}.svg`,
    event_gid: `event-${lang}-${index + 1}`, headline,
    summary: lang === 'zh'
      ? '开发团队带来了新的故事章节、场景与角色，也分享了旅途中的创作过程。欢迎和伙伴一起探索，在熟悉的世界中发现更多惊喜。'
      : 'Discover new chapters, locations and characters. The development team shares stories from the journey and invites everyone to explore the updated world together.',
    plain_text: '', html: index === 0 ? '<p>探索<strong>森林</strong>&nbsp;&amp;&nbsp;星空。</p>' : '',
    url: `https://games-home-news.example/${lang}/${index + 1}`, tags: [],
    published_at: dates[index]!, updated_at: dates[index]!, comment_count: 0, vote_up_count: 0, vote_down_count: 0,
  }))
}

function makeDataset(dataset: Dataset, lang: Language): GameHomeApiResponse {
  const data = makeHome()
  if (dataset === 'news-populated' || dataset === 'news-single') {
    const count = dataset === 'news-single' ? 1 : 3
    data.latest_news = { news_zh: makeNews('zh').slice(0, count), news_en: makeNews('en').slice(0, count) }
    data.panel.latest_news = lang === 'zh' ? data.latest_news.news_zh : data.latest_news.news_en
  }
  if (dataset === 'reviews-populated') {
    const names = [data.panel.latest_games[0]!.name, '星港漫游', '山谷来信']
    const times = ['2026-09-18 04:35:00', '2026-09-18 02:40:00', '2026-09-15 04:40:00']
    const bodies = ['这段旅途让我记住了每一位伙伴，也喜欢那些安静的场景与细节。长评论继续描述森林、星光与夏天的回忆，最后一段文字应当自然截断，不应挤压旁边的封面或下方的地区与时间。',
      '喜欢新的故事和角色。', '音乐和画面都很温柔，期待下一段旅程，也期待再次遇见熟悉的伙伴。']
    data.latest_reviews = names.map((game_name, index) => ({ game_name, time: times[index]!,
      content: bodies[index]!, region: `测试区域 ${String.fromCharCode(65 + index)}`,
      ip: ['192.0.2.10', '198.51.100.20', '203.0.113.30'][index]!, score: [4.2, 3.5, 5][index]!,
      game_cover: `${assetOrigin}/reviews/${index + 1}.svg` }))
  }
  if (dataset === 'layout-stress') {
    for (const [groupIndex, key] of (['latest_games', 'updated_games', 'free_games', 'popular_games'] as const).entries()) {
      data.panel[key] = Array.from({ length: 17 }, (_, index) => ({ ...game(7100 + groupIndex * 100 + index,
        `布局检查 ${index + 1} — 旅途中的森林、星光与永不结束的夏日回忆`),
      name_en: `Layout check ${index + 1} — a long journey through the forest and summer memories`,
      comment_count: index + 1 }))
    }
  }
  return data
}

function cover(id: string) {
  const colors = ['#394b44', '#4b5269', '#785949']
  return `<svg xmlns="http://www.w3.org/2000/svg" width="460" height="215" viewBox="0 0 460 215"><rect width="460" height="215" fill="${colors[Number(id) % 3]}"/><circle cx="350" cy="62" r="28" fill="#d9c6a3"/><path d="M0 215V155L110 58l120 157M190 215l125-123 145 110v13" fill="#819484"/><path d="M0 195L95 132l78 83H0" fill="#263b37"/></svg>`
}

type Theme = 'light' | 'dark'
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type State = { data: GameHomeApiResponse, upstream: URL[], lang: Language, showcase: GameShowcaseSnapshot, scenario: ShowcaseScenario,
  events: GameShowcaseEvent[], parallel: Promise<void>, releaseReads(): void, destination?: string,
  showcaseHold: Promise<void>, releaseShowcase(): void, showcasePending: boolean, collections: CollectionsScenario, collectionsHold: Promise<void>, releaseCollections(): void }
type Worker = { app: App, current: State | null }
export type GamesHomeScene = {
  page: Page, root: Locator, groups: Locator, stats: Locator, sidebar: Locator, dock: Locator, theme: Theme,
  data: GameHomeApiResponse, group(index: number): Locator, cards(index: number): Locator,
  settle(...targets: Locator[]): Promise<void>, assertQuiet(): void, visualReady(): Promise<void>,
  news: Locator, reviews: Locator, rendered: string, locale: Language,
  expectNewsPopup(url: string): void,
  showcase: Locator, snapshot: GameShowcaseSnapshot, events: GameShowcaseEvent[], assets: string[],
  readonly showcasePending: boolean, releaseShowcase(): void, releaseCollections(): void,
  expectShowcasePopup(url: string): void, expectGameDestination(id: string): void,
  holdArtwork(url: string): { requested: Promise<void>; release(): void },
  closureClip(target: Locator): Promise<{ x: number, y: number, width: number, height: number }>,
}

export const test = base.extend<{ gamesHome: { open(options?: OpenOptions): Promise<GamesHomeScene> } }, {
  gamesHomeApp: Worker, gamesHomeFixedNow: string | undefined,
}>({
  gamesHomeFixedNow: [undefined, { scope: 'worker', option: true }],
  gamesHomeApp: [async ({ gamesHomeFixedNow }, use) => {
    const worker = { current: null } as Worker
    const clockEnvironment = gamesHomeFixedNow === undefined ? {} : {
      TZ: 'UTC', GOFURRY_TEST_GAMES_HOME_NOW: gamesHomeFixedNow,
      NODE_OPTIONS: [process.env.NODE_OPTIONS, `--import="${new URL('./games-home-clock.mjs', import.meta.url).href}"`].filter(Boolean).join(' '),
    }
    worker.app = await startInsightsFixtureApp(async (url: URL, _media: string, body?: GameShowcaseEvent) => {
      const state = worker.current
      if (!state) throw new Error('Games Home API requires an active scenario')
      state.upstream.push(url)
      if ([homePath, showcasePath, collectionsPath].includes(url.pathname)) {
        // No response can complete until all three reads start: prove parallel SSR.
        if (state.upstream.filter(value => [homePath, showcasePath, collectionsPath].includes(value.pathname)).length === 3) state.releaseReads()
        await state.parallel
      }
      if (url.pathname === homePath && url.searchParams.size === 2
        && url.searchParams.get('lang') === state.lang && url.searchParams.get('region') === 'CN') return { data: state.data }
      if (url.pathname === showcasePath && url.searchParams.size === 2
        && url.searchParams.get('lang') === state.lang && url.searchParams.get('region') === 'CN') {
        if (state.scenario === 'slow-showcase') {
          state.showcasePending = true
          await state.showcaseHold
          state.showcasePending = false
        }
        return (state.scenario === 'showcase-failure' || (state.scenario === 'showcase-transient' && state.upstream.filter(v => v.pathname === showcasePath).length === 1)) ? { status: 503 } : { data: state.showcase }
      }
      if (url.pathname === collectionsPath && url.searchParams.size === 2 && url.searchParams.get('lang') === state.lang && url.searchParams.get('mode') === 'sfw') {
        if (state.collections === 'slow') await state.collectionsHold
        if (state.collections === 'failure' || (state.collections === 'transient' && state.upstream.filter(v => v.pathname === collectionsPath).length === 1)) return { status: 503 }
        const home = collectionHome(assetOrigin, state.lang)
        for (const slot of home.slots) slot.collection.preview_games = []
        if (state.collections === 'empty') home.slots = []
        if (state.collections === 'sparse') home.slots = home.slots.filter(s => s.slot === 1 || s.slot === 3)
        return { data: home }
      }
      if (url.pathname === eventPath && body) {
        state.events.push(body)
        return { status: state.scenario === 'tracking-failure' ? 500 : 204 }
      }
      if (state.destination) {
        const id = state.destination
        if (url.pathname === '/api/v2/game/info' && url.searchParams.get('id') === id) return { data: {
          id, appid: id, name: 'Showcase destination', summary: 'A local game destination', site: { view_count: 1, resources: [], groups: [], links: [] },
          platforms: {}, prices: [], news: [], tags: [], developers: [], publishers: [], media: { screenshots: [], movies: [], assets: [] }, requirements: {}, support_info: {}, extra: {},
        } }
        if (url.pathname === '/api/v2/game/reviews' && url.searchParams.get('id') === id) return { data: { total: 0, remarks: [] } }
        if (url.pathname === '/api/v2/game/recommend/similar' && url.searchParams.get('id') === id) return { data: [] }
        if (url.pathname === `/api/v2/game/games/${id}/view`) return { data: { view_count: 2 } }
        if (url.pathname === `/api/v2/game/games/${id}/insights`) return { data: {
          game: { id: Number(id), name: 'Showcase destination' }, state: { free: null, windows: null, mac: null, linux: null, release: null, as_of: null },
          players: { current: null, peak_30d: null, average_30d: null, as_of: null, observed_days_30d: 0, sample_coverage_30d: null },
          price: null, regional_prices: { as_of: null, regions: [] }, recent_changes: [],
        } }
      }
      return { status: 500 }
    }, { ...clockEnvironment, NUXT_PUBLIC_ASSET_PRIMARY_BASE: showcaseOrigins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: showcaseOrigins.mirror })
    try { await use(worker) }
    finally { worker.current = null; await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ gamesHomeApp }, use) => { await use(gamesHomeApp.app.base) },
  gamesHome: async ({ page, context, gamesHomeApp, gamesHomeFixedNow }, use, testInfo) => {
    const { app } = gamesHomeApp
    let releaseReads!: () => void
    const parallel = new Promise<void>(resolve => { releaseReads = resolve })
    let releaseShowcase!: () => void
    const showcaseHold = new Promise<void>(resolve => { releaseShowcase = resolve })
    let releaseCollections!: () => void
    const collectionsHold = new Promise<void>(resolve => { releaseCollections = resolve })
    const state: State = { data: makeHome(), upstream: [], lang: 'zh', showcase: makeShowcase('empty', 'zh'), scenario: 'empty', events: [], parallel, releaseReads,
      showcaseHold, releaseShowcase, showcasePending: false, collections: 'populated', collectionsHold, releaseCollections }
    gamesHomeApp.current = state
    const upstreamStart = app.requests.length
    const expectedMediaFailures = new Set<string>()
    const expectedNetworkFailures = new Set<string>()
    const errors = captureBrowserErrors(page, expectedNetworkFailures)
    const diagnostics: { text: string, url: string }[] = []
    page.on('console', message => { if (message.type() === 'error') diagnostics.push({ text: message.text(), url: message.location().url }) })
    const external: string[] = [], failed: string[] = [], assets: string[] = [], browserAPI: string[] = []
    const covers = new Map<string, string>()
    const artworkGates = new Map<string, { requested: Promise<void>; arrived(): void; pending: Promise<void>; release(): void }>()
    const expectedPopups: string[] = [], popupRequests: string[] = [], popups: Page[] = []
    // Keep the live arrays for popup errors, including errors arriving after page creation.
    const popupErrors: string[][] = []
    context.on('page', popup => { popups.push(popup); popupErrors.push(captureBrowserErrors(popup)) })
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
        const gate = artworkGates.get(url.href)
        if (gate) { gate.arrived(); await gate.pending }
        if (expectedMediaFailures.has(url.href)) return route.fulfill({ status: 503, body: 'Injected artwork failure' })
        return route.fulfill({ contentType: 'image/svg+xml', body: cover(id) })
      }
      if (expectedPopups.includes(url.href) && request.method() === 'GET' && request.isNavigationRequest()) {
        popupRequests.push(url.href)
        return route.fulfill({ contentType: 'text/html', body: '<!doctype html><html><head><link rel="icon" href="data:,"></head><body>Fixture news</body></html>' })
      }
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const assertQuiet = () => {
      expect(app.requests.slice(upstreamStart)).toEqual(state.upstream)
      const reads = state.upstream.filter(url => [homePath, showcasePath, collectionsPath].includes(url.pathname))
      const recoveryPaths = [
        ...(['slow-showcase', 'showcase-failure', 'showcase-transient'].includes(state.scenario) ? [showcasePath] : []),
        ...(['slow', 'failure', 'transient'].includes(state.collections) ? [collectionsPath] : []),
      ]
      expect(reads.slice(0, 3).map(url => url.pathname).sort()).toEqual([homePath, showcasePath, collectionsPath].sort())
      expect(reads.slice(3).map(url => url.pathname).sort()).toEqual(recoveryPaths.sort())
      for (const url of reads) expect(Object.fromEntries(url.searchParams)).toEqual(url.pathname === collectionsPath ? { lang: state.lang, mode: 'sfw' } : { lang: state.lang, region: 'CN' })
      const detail = state.upstream.filter(url => ![homePath, showcasePath, collectionsPath, eventPath].includes(url.pathname))
      const detailPaths = state.destination ? ['/api/v2/game/info', '/api/v2/game/reviews', '/api/v2/game/recommend/similar',
        `/api/v2/game/games/${state.destination}/insights`, `/api/v2/game/games/${state.destination}/view`] : []
      expect(detail.map(url => url.pathname).sort()).toEqual([...detailPaths].sort())
      const sideEffects = state.events.map(() => `POST ${eventPath}`)
      expect(browserAPI.filter(call => !detail.some(url => call.endsWith(url.pathname + url.search))).sort(), 'Exactly one recovery per unavailable slice; none for SSR success').toEqual([...sideEffects, ...reads.slice(3).map(url => `GET ${url.pathname}${url.search}`)].sort())
      for (const event of state.events) {
        expect(Object.keys(event).sort()).toEqual((event.event === 'click' ? ['tracking_token', 'session_id', 'event', 'source'] : ['tracking_token', 'session_id', 'event']).sort())
        expect(state.showcase.items.map(item => item.tracking_token)).toContain(event.tracking_token)
        expect(event.session_id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/)
        expect(['click', 'impression']).toContain(event.event)
        if (event.event === 'click') expect(['artwork', 'title', 'primary', 'secondary']).toContain(event.source)
      }
      expect(external).toEqual([])
      expect(failed).toEqual([])
      const eventDiagnostic = 'Failed to load resource: the server responded with a status of 500 (Internal Server Error)'
      const eventFailures = diagnostics.filter(value => value.text === eventDiagnostic && value.url === `${app.base}${eventPath}`)
      expect(eventFailures.length).toBe(state.scenario === 'tracking-failure' ? state.events.length : 0)
      expect(errors.filter(value => value !== `error: ${eventDiagnostic}`)).toEqual([])
      expect(errors.filter(value => value === `error: ${eventDiagnostic}`)).toHaveLength(eventFailures.length)
      const recoveryFailures = diagnostics.filter(value => value.text === 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)' && value.url.startsWith(`${app.base}/api/`))
      expect(recoveryFailures).toHaveLength(Number(state.collections === 'failure') + Number(state.scenario === 'showcase-failure'))
      for (const url of expectedMediaFailures) {
        expect(diagnostics.filter(value => value.url === url)).toHaveLength(assets.filter(value => value === url).length)
      }
      expect(popupErrors.flat()).toEqual([])
      expect(popupRequests).toEqual(expectedPopups)
      expect(popups).toHaveLength(expectedPopups.length)
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
      await use({ async open({ width = 1440, height = 900, theme = 'light', locale = 'zh', dataset = 'core', showcase = 'empty', collections = 'populated', mode = 'sfw' } = {}) {
        expect(state.upstream, 'Each scenario opens once').toHaveLength(0)
        state.lang = locale
        state.data = makeDataset(dataset, locale)
        state.collections = collections
        if (collections === 'failure') expectedNetworkFailures.add(`${app.base}${collectionsPath}?lang=${locale}&mode=sfw`)
        if (showcase === 'showcase-failure') expectedNetworkFailures.add(`${app.base}${showcasePath}?lang=${locale}&region=CN`)
        state.scenario = showcase
        state.showcase = makeShowcase(showcase, locale)
        for (const item of state.showcase.items) {
          const urls = item.artwork.kind === 'managed'
            ? [item.artwork.desktop_object_key, item.artwork.mobile_object_key].filter(Boolean).flatMap(key => Object.values(showcaseOrigins).map(origin => `${origin}/${key}`))
            : steamSharedAssetCandidates(item.artwork.url, locale === 'en' ? 'global' : 'china')
          for (const url of urls) {
            covers.set(url, item.game_id || '6101')
            if (showcase === 'media-failure' || showcase === 'steam-failure' || (showcase === 'primary-failure' && url.startsWith(showcaseOrigins.primary))) expectedMediaFailures.add(url)
          }
        }
        for (const url of expectedMediaFailures) expectedNetworkFailures.add(url)
        const { panel } = state.data
        for (const item of [...panel.latest_games, ...panel.updated_games, ...panel.free_games, ...panel.popular_games!,
          ...panel.top_online, ...panel.top_price, ...panel.highest_discount, ...panel.low_price]) covers.set(item.header_url, item.id)
        for (const item of [...state.data.latest_news.news_zh, ...state.data.latest_news.news_en]) covers.set(item.header_url, item.game_id)
        for (const [index, item] of state.data.latest_reviews.entries()) covers.set(item.game_cover, String(6101 + index))
        if (dataset === 'reviews-populated') expect(gamesHomeFixedNow, 'Reviews need the shared SSR/browser clock').toBeDefined()
        if (gamesHomeFixedNow !== undefined) await page.clock.setFixedTime(new Date(gamesHomeFixedNow))
        await page.setViewportSize({ width, height })
        await context.addInitScript(({ origin, theme, steamKey, sample, mediaFailure, mode }) => {
          if (location.origin !== origin) return
          // Real fallback, with Save-Data excluding unrelated failure-triggered probes.
          if (mediaFailure) Object.defineProperty(navigator, 'connection', { value: { saveData: true }, configurable: true })
          localStorage.setItem('theme', theme)
          localStorage.setItem('mode', mode)
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
            primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0], mediaFailure: expectedMediaFailures.size > 0, mode })
        // The real Nitro fetch must time out while its upstream gate stays held.
        // Fail before the old eight-second budget; do not fake browser time for SSR.
        const response = await page.goto(locale === 'en' ? '/en/games' : '/games', {
          waitUntil: 'load', ...(showcase === 'slow-showcase' || collections === 'slow' ? { timeout: 4000 } : {}),
        })
        expect(response?.status()).toBe(200)
        const ssr = await response!.text()
        const rendered = ssr.slice(ssr.indexOf('<body'), ssr.indexOf('</body>')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g, '')
        const names = locale === 'en' ? englishGroupNames : groupNames
        for (const name of names) expect(rendered).toContain(name)
        const ssrCards = [...rendered.matchAll(/<p class="game-card__title[^>]*>([^<]*)<\/p>/g)].map(match => match[1]!.trim())
        expect(ssrCards).toEqual([state.data.panel.latest_games, state.data.panel.updated_games, state.data.panel.free_games, state.data.panel.popular_games!]
          .flatMap(list => list.slice(0, 8).map(item => locale === 'en' ? item.name_en! : item.name)))
        expect(rendered).toContain(state.data.panel.latest_games[0]!.summary)
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect.poll(() => page.locator('html').evaluate(element => element.classList.contains('dark'))).toBe(theme === 'dark')
        const root = page.locator('.games-page'), groups = root.locator('.game-info-group'), stats = root.locator('.game-stats-section')
        const sidebar = root.locator('.game-sidebar-shell'), dock = page.locator('.gf-nav__utilities')
        const group = (index: number) => groups.nth(index)
        const cards = (index: number) => group(index).locator('.game-group-page-live .game-card')
        await expect(groups).toHaveCount(4)
        await expect(groups.locator('h3')).toHaveText([...names])
        for (let index = 0; index < 4; index++) await expect(cards(index)).toHaveCount(8)
        await expect(stats.locator('.stats-type-tab--active')).toHaveText(locale === 'en' ? 'Player Count' : '在线人数')
        const recoveryCount = Number(['slow-showcase', 'showcase-failure', 'showcase-transient'].includes(showcase)) + Number(['slow', 'failure', 'transient'].includes(collections))
        await expect.poll(() => browserAPI.filter(call => call.startsWith('GET ')).length).toBe(recoveryCount)
        await expect.poll(() => diagnostics.filter(value => value.text === 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)' && value.url.startsWith(`${app.base}/api/`)).length).toBe(Number(collections === 'failure') + Number(showcase === 'showcase-failure'))
        await settle(root)
        if (showcase === 'empty' || showcase === 'showcase-failure') assertQuiet()
        return { page, root, groups, stats, sidebar, dock, theme, data: state.data, group, cards, settle, assertQuiet,
          rendered, locale, news: root.locator('.game-news-panel'),
          showcase: root.locator('.game-home-showcase'), snapshot: state.showcase, events: state.events, assets,
          get showcasePending() { return state.showcasePending }, releaseShowcase, releaseCollections,
          expectShowcasePopup(url) {
            expect(state.showcase.items.flatMap(item => [item.primary_action.target, item.secondary_action?.target])).toContain(url)
            expectedPopups.push(url)
          },
          expectGameDestination(id) { state.destination = id },
          holdArtwork(url) {
            expect(covers.has(url), 'Gate only a declared fixture asset').toBe(true)
            let arrived!: () => void, release!: () => void
            const requested = new Promise<void>(resolve => { arrived = resolve })
            const pending = new Promise<void>(resolve => { release = resolve })
            const gate = { requested, pending, arrived, release }
            artworkGates.set(url, gate)
            return gate
          },
          reviews: sidebar.getByRole('heading', { name: locale === 'en' ? 'Latest Reviews' : '最新评论', exact: true }).locator('..'),
          expectNewsPopup(url) {
            expect(state.data.latest_news[locale === 'en' ? 'news_en' : 'news_zh'].map(item => item.url)).toContain(url)
            expectedPopups.push(url)
          },
          async closureClip(target) {
            await target.evaluate(el => el.scrollIntoView({ block: 'center', behavior: 'instant' }))
            await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
            await page.mouse.move(1, 1)
            await settle(target)
            const rect = await target.boundingBox()
            expect(rect).not.toBeNull()
            const viewport = page.viewportSize()!
            expect(rect!.y).toBeGreaterThanOrEqual(0)
            expect(rect!.y + rect!.height).toBeLessThanOrEqual(viewport.height)
            expect(rect!.x).toBeGreaterThanOrEqual(0)
            expect(rect!.x + rect!.width).toBeLessThanOrEqual(viewport.width)
            const x = Math.max(0, Math.floor(rect!.x - 12)), y = Math.max(0, Math.floor(rect!.y - 12))
            const right = Math.min(viewport.width, Math.ceil(rect!.x + rect!.width + 12))
            const bottom = Math.min(viewport.height, Math.ceil(rect!.y + rect!.height + 12))
            expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
            assertQuiet()
            return { x, y, width: right - x, height: bottom - y }
          },
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
      releaseReads()
      releaseShowcase()
      releaseCollections()
      for (const gate of artworkGates.values()) gate.release()
      await testInfo.attach('games-home-network.json', { contentType: 'application/json', body: JSON.stringify({
        upstream: state.upstream.map(String), browserAPI, events: state.events, assets, external, failed, errors, diagnostics, popupErrors, expectedPopups, popupRequests,
      }, null, 2) })
      for (const popup of popups) if (!popup.isClosed()) await popup.close()
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
  await expect(scene.page.locator('.game-tool-dock')).toHaveCount(0)
  await expect(dock.locator('.gf-nav__utility-row > a, .gf-nav__utility-row > button')).toHaveCount(3)
  await css(scene.page.locator('.gf-nav__actions > .gf-nav__shell:visible'), {
    'background-color': dark ? 'rgba(15, 22, 33, 0.97)' : 'rgba(24, 33, 47, 0.96)',
    'border-width': '0px', 'box-shadow': 'none', 'background-image': 'none', 'backdrop-filter': 'blur(10px)' })
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

// P6.1.3: measured in pinned Chromium, not inferred from utility names.
export async function assertNewsAppearance(scene: GamesHomeScene) {
  const { page, news, theme } = scene, dark = theme === 'dark'
  const strong = dark ? 'rgba(241, 245, 249, 0.88)' : 'rgba(71, 42, 20, 0.92)'
  const muted = dark ? 'rgba(203, 213, 225, 0.62)' : 'rgba(120, 83, 53, 0.58)'
  await news.evaluate(el => el.scrollIntoView({ block: 'center', behavior: 'instant' }))
  await page.mouse.move(1, 1); await scene.settle(news)
  await css(news.locator('.game-news-title'), { 'font-size': '18px', 'line-height': '28.0001px', 'font-weight': '700', color: strong })
  await css(news.locator('.game-news-desc'), { 'font-size': '14px', 'line-height': '20px', 'font-weight': '400', color: muted })
  await css(news.locator('.news-pager__count'), { 'font-size': '12px', 'line-height': '16px', 'font-weight': '600', color: muted })
  await css(news.locator('.news-viewport'), { 'overflow-x': 'hidden' })
  const reduced = await page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches)
  await css(news.locator('.news-track'), reduced
    ? { 'transition-property': 'none', 'transition-duration': '0s' }
    : { 'transition-property': 'transform', 'transition-duration': '0.34s',
        'transition-timing-function': 'cubic-bezier(0.22, 1, 0.36, 1)' })
  const card = news.locator('.news-card').first()
  const normal = dark ? 'rgba(226, 232, 240, 0.067)' : 'rgba(255, 250, 242, 0.4)'
  const shadow = dark ? 'none' : 'rgba(91, 62, 28, 0.03) 0px 4px 12px 0px'
  await css(card, { 'background-color': normal, 'border-color': 'rgba(0, 0, 0, 0)', 'border-width': '1px',
    'border-radius': '12.48px', 'box-shadow': shadow, 'backdrop-filter': 'blur(1px)',
    width: page.viewportSize()!.width < 640 ? '248px' : page.viewportSize()!.width < 1024 ? '312px' : '328px',
    'transition-property': 'background-color, border-color', 'transition-duration': '0.18s, 0.18s' })
  await css(card.locator('.news-card__title'), { 'font-size': '16px', 'line-height': '21.12px', 'font-weight': '700',
    color: strong, '-webkit-line-clamp': '2', 'min-height': '40.8px' })
  await css(card.locator('.news-card__summary'), { 'font-size': '14px', 'line-height': '20.3px', color: muted,
    '-webkit-line-clamp': '3', 'min-height': '58.4px' })
  await css(card.locator('.news-card__meta'), { 'font-size': '11.52px', 'line-height': '13.824px', color: muted })
  await css(news.locator('.news-nav-button:disabled'), { opacity: '0.36' })
  await css(news.locator('.news-progress-track'), { 'border-radius': '999px', 'background-color': dark
    ? 'color(srgb 0.886275 0.909804 0.941176 / 0.0953726)' : 'color(srgb 0.494118 0.360784 0.227451 / 0.102902)' })
  await css(news.locator('.news-progress-fill'), { 'border-radius': '999px', 'background-color': dark
    ? 'color(srgb 0.886275 0.909804 0.941176 / 0.396863)' : 'color(srgb 0.486275 0.176471 0.0705882 / 0.44)',
  'transition-property': 'width', 'transition-duration': '0.22s' })
  await card.hover(); await scene.settle(news)
  await css(card, { 'background-color': dark ? 'rgba(148, 163, 184, 0.18)' : 'rgba(255, 235, 205, 0.82)',
    'border-color': 'rgba(0, 0, 0, 0)', 'box-shadow': shadow })
  await page.mouse.move(1, 1); await scene.settle(news)
  await css(card, { 'background-color': normal })
}

export async function assertLatestReviewsAppearance(scene: GamesHomeScene) {
  const { page, reviews, theme } = scene, dark = theme === 'dark'
  const strong = dark ? 'rgba(241, 245, 249, 0.88)' : 'rgba(71, 42, 20, 0.92)'
  const muted = dark ? 'rgba(203, 213, 225, 0.62)' : 'rgba(120, 83, 53, 0.58)'
  await reviews.evaluate(el => el.scrollIntoView({ block: 'center', behavior: 'instant' }))
  await page.mouse.move(1, 1); await scene.settle(reviews)
  await css(reviews.getByRole('heading'), { 'font-size': '13.44px', 'line-height': '19.2px', 'font-weight': '700', color: strong })
  const item = reviews.locator('.latest-review-item').first()
  const normal = dark ? 'rgba(226, 232, 240, 0.067)' : 'rgba(255, 250, 242, 0.4)'
  const shadow = dark ? 'none' : 'rgba(91, 62, 28, 0.03) 0px 4px 12px 0px'
  await css(item, { 'background-color': normal, 'border-width': '1px', 'border-color': 'rgba(0, 0, 0, 0)',
    'border-radius': '13.12px', 'box-shadow': shadow, padding: '10.88px', gap: '12px' })
  await css(item.locator('.latest-review-item__identity'), { width: '88px', gap: '4px' })
  await css(item.locator('.latest-review-item__media'), { width: '88px', height: '52px', 'border-radius': '6px',
    'background-color': dark ? 'color(srgb 0.580392 0.639216 0.721569 / 0.137098)' : 'color(srgb 1 0.921569 0.803922 / 0.622902)' })
  await css(item.locator('img'), { 'object-fit': 'cover' })
  await css(item.locator('.latest-review-item__title'), { 'font-size': '12px', 'line-height': '16px',
    'font-weight': '600', color: strong, 'text-overflow': 'ellipsis' })
  await css(item.locator('.latest-review-item__body'), { 'font-size': '14px', 'line-height': '19.25px', color: muted, '-webkit-line-clamp': '2' })
  await css(item.locator('.latest-review-item__meta'), { 'font-size': '12px', 'line-height': '16px', color: muted })
  await css(scene.sidebar, { 'background-color': 'rgba(0, 0, 0, 0)', 'border-width': '0px', 'box-shadow': 'none', 'backdrop-filter': 'none' })
  await item.hover(); await scene.settle(reviews)
  await css(item, { 'background-color': dark ? 'rgba(148, 163, 184, 0.18)' : 'rgba(255, 235, 205, 0.82)',
    'border-color': 'rgba(0, 0, 0, 0)', 'box-shadow': shadow })
  await page.mouse.move(1, 1); await scene.settle(reviews)
  await css(item, { 'background-color': normal })
}
