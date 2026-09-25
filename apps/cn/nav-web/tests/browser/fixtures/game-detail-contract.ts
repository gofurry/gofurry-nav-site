import { test as base, expect, type Locator, type Page, type Request } from '@playwright/test'
import { fileURLToPath } from 'node:url'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type Theme = 'light' | 'dark'
type Call = { path: string, query: Record<string, string> }
type Gate = { received: boolean, completed: boolean, promise: Promise<void>, release(): void, wait(): Promise<void> }
type Rule = { path: string, query: Record<string, string>, fail: boolean, gate?: Gate }
type State = { reads: Call[], unexpected: string[], rules: Rule[], empty: boolean, adult: boolean, historyPoints: number }
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Worker = { app: App, current: State | null }
export const detailNow = '2026-09-18T12:40:00+08:00'
const prefix = '/api/v2/game'
const image = '<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360"><rect width="640" height="360" fill="#394b44"/><circle cx="490" cy="85" r="45" fill="#c8b398"/><path d="M0 360V240L180 80l190 280M290 360l160-180 190 150v30" fill="#778978"/><text x="36" y="304" fill="#fff" font-family="sans-serif" font-size="24">GAME · DETAIL FIXTURE</text></svg>'
const assetNames = ['cover', 'similar', 'poster', ...Array.from({ length: 12 }, (_, i) => `shot-${i}`)]
const title = (id: string, lang: string) => lang === 'en' ? `Forest Journey ${id}` : `林间旅途 ${id}`
const matches = (url: URL, rule: Pick<Rule, 'path' | 'query'>) => url.pathname === prefix + rule.path
  && Object.entries(rule.query).every(([key, value]) => url.searchParams.get(key) === value)
function gate(): Gate {
  let release!: () => void
  const result: Gate = { received: false, completed: false, promise: new Promise<void>(resolve => { release = resolve }),
    release: () => release(), async wait() { await expect.poll(() => result.received).toBe(true) } }
  return result
}
function payload(url: URL, media: string, state: State) {
  const id = url.searchParams.get('id') || url.pathname.match(/games\/(\d+)/)?.[1] || '82'
  const lang = url.searchParams.get('lang') || 'zh', path = url.pathname.slice(prefix.length)
  if (path === '/info') return {
    id, appid: Number(id), name: title(id, lang), summary: '一段穿越森林与小镇的旅途，与伙伴寻找失落的故事。',
    short_description: '探索、相遇与选择，构成属于你的冒险。', type: 'game', is_free: false,
    about_the_game: state.empty ? '' : '<h2>探索森林中的故事</h2><p>与伙伴一起穿越小镇，在每次选择中发现新的风景。</p><p><a href="https://detail.example.test/intro">阅读开发日志</a></p><blockquote>每一段旅途都有值得记住的相遇。</blockquote>',
    site: { view_count: 120, create_time: '2026-08-01', update_time: '2026-09-18', resources: [{ key: '试玩下载', value: 'https://detail.example.test/demo' }, { key: '开发日志', value: 'https://detail.example.test/intro' }],
      groups: [{ key: 'discord', value: 'https://detail.example.test/community' }], links: [{ key: 'official', value: 'https://detail.example.test/official' }] },
    release: { coming_soon: false, date: '2026-08-18', availability: 'available', precision: 'day', exact_date: '2026-08-18', year: 2026, month: 8, quarter: null, window_start: '2026-08-18', window_end: '2026-08-18', raw_text: '', observed_at: '2026-08-18' },
    first_available: { precision: 'day', exact_date: '2026-08-18', year: 2026, month: 8, quarter: null, window_start: '2026-08-18', window_end: '2026-08-18', source: 'observed_transition', inferred: false },
    platforms: { windows: true, mac: true, linux: false }, languages: [
      { code: 'zh-Hans', steam_name: 'Simplified Chinese', steam_api_code: 'schinese', steam_web_code: 'schinese', tier: 'platform', interface_supported: true, subtitles_supported: true, full_audio_supported: false },
      { code: 'en', steam_name: 'English', steam_api_code: 'english', steam_web_code: 'english', tier: 'platform', interface_supported: true, subtitles_supported: true, full_audio_supported: true }],
    developers: ['Forest Studio'], publishers: ['Forest Publisher'], online_count: { count: 42, collected_at: '2026-09-18T04:00:00Z' },
    tags: Array.from({ length: 10 }, (_, i) => ({ id: String(700 + i), code: state.adult && i === 0 ? 'adult' : `detail-${i}`, category_code: 'genre', role: 'normal', name: `标签 ${i + 1}`, desc: `标签 ${i + 1} 的固定描述` })),
    prices: [{ region: 'CN', available: true, currency: 'CNY', final_amount: 3200, final_formatted: '¥32.00' }],
    media: { library_cover_url: `${media}/detail-cover.svg`, screenshots: state.empty ? [] : Array.from({ length: 12 }, (_, i) => ({ id: i + 1, thumbnail_url: `${media}/detail-shot-${i}.svg`, url: `${media}/detail-shot-${i}.svg` })),
      movies: state.empty ? [] : [{ id: 1, name: '旅途预告', thumbnail_url: `${media}/detail-poster.svg`, extra: { webm_max_url: `${media}/detail-trailer.webm` } }], assets: [] },
    news: state.empty ? [] : Array.from({ length: 7 }, (_, i) => ({ headline: `林间消息 ${i + 1}`, published_at: `2026-09-${String(18 - i).padStart(2, '0')}T04:00:00Z`, html: '<p>这一版本新增了森林路线与伙伴故事，感谢每一位参与旅途的玩家。</p>', url: 'https://detail.example.test/news' })),
    requirements: { pc: { minimum: '<p>Windows 10 · 8 GB RAM</p>', recommended: '<p>Windows 11 · 16 GB RAM</p>' } },
    support_info: { url: 'https://detail.example.test/support', email: 'support@example.test' }, website: 'https://detail.example.test/official',
    extra: { ratings: [{ board: 'esrb', rating: 'T', required_age: '13' }], content_descriptors: ['Fantasy themes'] },
  }
  if (path === '/reviews') {
    const page = Number(url.searchParams.get('page'))
    return { total: state.empty ? 0 : 7, avg_score: state.empty ? 0 : 4.2, page_num: page,
      remarks: state.empty ? [] : Array.from({ length: page === 1 ? 5 : 2 }, (_, i) => ({ name: `玩家 ${i + (page - 1) * 5 + 1}`, content: `一段值得记住的旅程，期待下一次更新。${id}`, score: i === 0 ? 4.5 : 4,
        ip: '192.0.2.*', region: '中国', create_time: `2026-09-${String(18 - i).padStart(2, '0')} 12:00:00` })) }
  }
  if (path === '/recommend/similar') return state.empty ? [] : Array.from({ length: 6 }, (_, i) => ({ id: String(83 + i), appid: String(83 + i), name: title(String(83 + i), lang), summary: '探索新的故事，结识同行的伙伴。', display_score: .9 - i / 20,
    library_cover_url: i === 5 ? '' : `${media}/detail-similar.svg`, reasons: [{ label: '共同标签', value: '剧情丰富' }] }))
  if (/^\/games\/\d+\/view$/.test(path)) return { view_count: 121 }
  if (/^\/games\/\d+\/insights$/.test(path)) return {
    game: { id: Number(id), name: title(id, lang) }, state: { free: false, windows: true, mac: true, linux: null, release: 'available', as_of: '2026-09-18' },
    players: { current: 0, peak_30d: 120, average_30d: 42.5, as_of: '2026-09-18T04:00:00Z', fact_through: '2026-09-17', eligible_from_30d: '2026-08-19', observed_days_30d: 28, successful_samples_30d: 112, sample_coverage_30d: .93 },
    price: null, regional_prices: { as_of: '2026-09-18', regions: [
      { region: 'CN', available: true, state: 'priced', currency: 'CNY', initial_amount: 4200, final_amount: 0, discount_percent: 100, observed_low: { amount: 0, currency: 'CNY', first_seen: '2026-09-01', observed_since: '2026-09-01', initial_amount: 4200, discount_percent: 100 } },
      { region: 'US', available: true, state: 'unknown', currency: null, initial_amount: null, final_amount: null, discount_percent: null, observed_low: null },
      { region: 'HK', available: false, state: null, currency: null, initial_amount: null, final_amount: null, discount_percent: null, observed_low: null },
    ] }, recent_changes: Array.from({ length: 7 }, (_, i) => ({ type: i % 2 ? 'game.discount.started' : 'game.release.available', date: `2026-09-${18 - i}`, occurred_at: i % 2 ? `2026-09-${18 - i}T04:00:00Z` : null })),
  }
  if (/\/insights\/(players|prices)$/.test(path)) {
    const player = path.endsWith('/players'), region = url.searchParams.get('region') || 'CN'
    return { region, requested_range: url.searchParams.get('range'), available_from: '2026-09-01', available_through: '2026-09-08',
      points: Array.from({ length: state.historyPoints }, (_, i) => ({ date: `2026-09-${String(i + 1).padStart(2, '0')}`, ...(player
        ? { min: i, max: [24, 42, 32, 0, 65, 56, 88, 71][i % 8], avg: i === 3 ? null : 12 + i * 4 }
        : { state: i === 3 ? 'unknown' : i === 4 ? 'free' : 'priced', currency: i === 3 ? null : region === 'CN' ? 'CNY' : region === 'US' ? 'USD' : 'HKD', initial_amount: i === 3 ? null : 4200, final_amount: i === 3 ? null : i === 4 ? 0 : 3200 - i * 100, discount_percent: i === 3 ? null : 20 }) })) }
  }
  state.unexpected.push(`unhandled upstream ${url.href}`); return null
}
function assertQuery(url: URL) {
  const path = url.pathname.slice(prefix.length), query = Object.fromEntries(url.searchParams)
  if (path === '/info') {
    expect(query).toEqual({ id: expect.stringMatching(/^8[2-8]$/), lang: expect.stringMatching(/^(zh|en)$/), region: 'CN', news_limit: '20' })
  } else if (path === '/reviews') {
    expect(query).toEqual({ id: expect.stringMatching(/^8[2-8]$/), page: expect.stringMatching(/^[12]$/), limit: '5' })
  } else if (path === '/recommend/similar') {
    expect(query).toEqual({ id: expect.stringMatching(/^8[2-8]$/), lang: expect.stringMatching(/^(zh|en)$/), region: 'CN', limit: '8' })
  } else if (/\/insights\/players$/.test(path)) {
    expect(query).toEqual({ range: expect.stringMatching(/^(30d|90d|180d|1y|3y|5y)$/) })
  } else if (/\/insights\/prices$/.test(path)) {
    expect(query).toEqual({ range: expect.stringMatching(/^(30d|90d|180d|1y|3y|5y)$/), region: expect.stringMatching(/^(CN|US|HK)$/) })
  } else if (/^\/games\/8[2-8]\/(view|insights)$/.test(path)) expect(query).toEqual({})
  else throw new Error(`Unplanned Detail API ${url.href}`)
}
export type DetailScene = {
  page: Page, root: Locator, theme: Theme, html: string, reads: Call[], browserCalls: { method: string, url: string }[],
  open(options?: { theme?: Theme, width?: number, locale?: 'zh' | 'en', adult?: boolean, empty?: boolean }): Promise<void>,
  tab(key: string): Promise<void>, fail(path: string, query?: Record<string, string>): () => void,
  hold(path: string, query?: Record<string, string>): Gate, points(count: number): void,
  count(path: string, query?: Record<string, string>): number, assertQuiet(): void, assertInitialReads(): void,
  popup(action: () => Promise<unknown>, url: string): Promise<void>, movieGate(): Gate, failMovie(): void,
  settle(target?: Locator): Promise<void>, clip(target: Locator, margin?: number): Promise<{ x: number, y: number, width: number, height: number }>,
}
export const test = base.extend<{ detail: DetailScene }, { detailApp: Worker }>({
  timezoneId: 'Asia/Shanghai',
  // eslint-disable-next-line no-empty-pattern -- Playwright requires destructured fixture arguments.
  detailApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp(async (url: URL, media: string) => {
      const state = worker.current
      if (!state) throw new Error('Detail request without a scenario')
      try { assertQuery(url) } catch (error) { state.unexpected.push(String(error)); return { status: 500 } }
      state.reads.push({ path: url.pathname.slice(prefix.length), query: Object.fromEntries(url.searchParams) })
      const active = state.rules.find(rule => matches(url, rule) && rule.gate && !rule.gate.received)
      if (active?.gate) { active.gate.received = true; await active.gate.promise; active.gate.completed = true }
      if (state.rules.some(rule => matches(url, rule) && rule.fail)) return { status: 503 }
      return { data: payload(url, media, state) }
    })
    try { await use(worker) } finally { await worker.app.close() }
  }, { scope: 'worker' }],
  detail: async ({ page, context, detailApp }, use, testInfo) => {
    expect(detailApp.current).toBeNull()
    const state: State = { reads: [], unexpected: [], rules: [], empty: false, adult: false, historyPoints: 8 }
    detailApp.current = state
    const { app } = detailApp, expectedURLs = new Set<string>(), errors = captureBrowserErrors(page, expectedURLs)
    const rawErrors: { text: string, url: string }[] = [], external: string[] = [], failed: Request[] = [], httpFailures: Request[] = []
    const gates: Gate[] = [], browserCalls: DetailScene['browserCalls'] = [], popups: string[] = [], allowedPopups = new Set<string>()
    let movie: Gate | undefined, movieFailed = false, opened = false
    page.on('console', message => { if (message.type() === 'error') rawErrors.push({ text: message.text(), url: message.location().url }) })
    context.on('requestfailed', request => failed.push(request))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin === app.base && url.pathname.startsWith('/api/')) browserCalls.push({ method: request.method(), url: url.pathname + url.search })
    })
    context.on('response', response => { if (response.status() >= 400) httpFailures.push(response.request()) })
    await context.route('**/*', async route => {
      const request = route.request(), url = new URL(request.url())
      if (allowedPopups.has(url.href) && request.isNavigationRequest()) { popups.push(url.href); return route.fulfill({ contentType: 'text/html', body: '<p>Detail popup fixture</p>' }) }
      if (url.origin === app.upstreamUrl && assetNames.some(name => url.pathname === `/media/detail-${name}.svg`)) return route.fulfill({ contentType: 'image/svg+xml', body: image })
      if (url.origin === app.upstreamUrl && url.pathname === '/media/detail-trailer.webm') {
        if (movieFailed) { expectedURLs.add(url.href); return route.fulfill({ status: 503, body: 'Injected media failure' }) }
        if (movie) { movie.received = true; await movie.promise }
        await route.fulfill({ contentType: 'video/webm', path: fileURLToPath(new URL('../../fixtures/game-detail-trailer.webm', import.meta.url)) })
        if (movie) movie.completed = true
        return
      }
      if (url.origin !== app.base) { external.push(url.href); return route.abort('blockedbyclient') }
      if (url.pathname.startsWith('/api/')) {
        const path = url.pathname.slice(prefix.length), valid = url.pathname.startsWith(prefix) && (/^\/(info|reviews|recommend\/similar)$/.test(path) || /^\/games\/\d+\/(view|insights(?:\/(players|prices))?)$/.test(path))
        if (!valid || request.method() !== (path.endsWith('/view') ? 'POST' : 'GET')) { state.unexpected.push(`browser ${request.method()} ${url.href}`); return route.abort('blockedbyclient') }
        if (state.rules.some(rule => matches(url, rule) && rule.fail)) expectedURLs.add(url.href)
      }
      return route.continue()
    })
    const scene: DetailScene = { page, root: page.locator('.game-detail-page'), theme: 'light', html: '', reads: state.reads, browserCalls,
      async open({ theme = 'light', width = 1440, locale = 'zh', adult = false, empty = false } = {}) {
        expect(opened).toBe(false); opened = true; state.adult = adult; state.empty = empty; scene.theme = theme
        await page.setViewportSize({ width, height: width < 768 ? 844 : 900 })
        await page.clock.setFixedTime(new Date(detailNow))
        await context.addInitScript(({ origin, theme, checkedAt, steamKey, sample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample, china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, checkedAt: Date.parse(detailNow), steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
        const response = await page.goto(`${app.base}${locale === 'en' ? '/en' : ''}/games/82`, { waitUntil: 'load' })
        expect(response?.status()).toBe(200); scene.html = await response!.text()
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect(page.locator('h1')).toHaveText(title('82', locale))
        await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
        await expect.poll(() => scene.count('/games/82/view')).toBe(1)
      },
      async tab(key) { await page.locator(`[data-game-tab="${key}"]`).click(); await expect(page.locator('.game-detail-tab--active')).toHaveAttribute('data-game-tab', key) },
      fail(path, query = {}) { const rule = { path, query, fail: true }; state.rules.push(rule); return () => { rule.fail = false } },
      hold(path, query = {}) { const held = gate(); gates.push(held); state.rules.push({ path, query, fail: false, gate: held }); return held },
      points(count) { state.historyPoints = count },
      count(path, query = {}) { return state.reads.filter(call => call.path === path && Object.entries(query).every(([k, v]) => call.query[k] === v)).length },
      assertQuiet() {
        expect(errors).toEqual([]); expect(external).toEqual([])
        expect(failed.map(request => ({ url: request.url(), error: request.failure()?.errorText }))).toEqual([])
        expect(state.unexpected).toEqual([])
        for (const request of httpFailures) expect(expectedURLs.has(request.url())).toBe(true)
        for (const entry of rawErrors) {
          expect(expectedURLs.has(entry.url)).toBe(true)
          expect(entry.text).toBe('Failed to load resource: the server responded with a status of 503 (Service Unavailable)')
        }
        expect(rawErrors).toHaveLength(httpFailures.length)
        for (const url of expectedURLs) expect(rawErrors.filter(entry => entry.url === url)).toHaveLength(httpFailures.filter(request => request.url() === url).length)
        for (const held of gates) { expect(held.received).toBe(true); expect(held.completed).toBe(true) }
      },
      assertInitialReads() {
        expect(state.reads.map(call => call.path).sort()).toEqual(['/info', '/reviews', '/recommend/similar', '/games/82/insights', '/games/82/view'].sort())
        expect(browserCalls).toEqual([{ method: 'POST', url: `${prefix}/games/82/view` }])
        expect(state.reads.find(call => call.path === '/info')!.query).toEqual({ id: '82', lang: 'zh', region: 'CN', news_limit: '20' })
      },
      async popup(action, url) {
        allowedPopups.add(url); const pending = context.waitForEvent('page'); await action(); const popup = await pending
        await popup.waitForURL(url); await popup.waitForLoadState('load'); expect(popups).toContain(url); await popup.close()
      },
      movieGate() { movie = gate(); gates.push(movie); return movie },
      failMovie() { movie = undefined; movieFailed = true },
      async settle(target = scene.root) {
        await target.evaluate(async root => {
          await Promise.all([...root.querySelectorAll('img')].filter(img => img.getBoundingClientRect().bottom > 0 && img.getBoundingClientRect().top < innerHeight).map(img => img.decode()))
          await document.fonts.ready
          await Promise.all(root.getAnimations({ subtree: true }).filter(a => a.effect?.getComputedTiming().iterations !== Infinity).map(a => a.finished.catch(() => {})))
          await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
        })
      },
      async clip(target, margin = 12) {
        await target.evaluate(el => {
          el.scrollIntoView({ block: 'start', behavior: 'instant' })
          const nav = document.querySelector('.gf-nav')?.getBoundingClientRect().height ?? 0
          window.scrollBy({ top: -nav - 16, behavior: 'instant' })
        }); await scene.settle(target)
        const box = await target.boundingBox(); expect(box).not.toBeNull(); const viewport = page.viewportSize()!
        const x = Math.max(0, Math.floor(box!.x - margin)), y = Math.max(0, Math.floor(box!.y - margin))
        return { x, y, width: Math.min(viewport.width - x, Math.ceil(box!.width + 2 * margin)), height: Math.min(viewport.height - y, Math.ceil(box!.height + 2 * margin)) }
      },
    }
    try { await use(scene) } finally {
      for (const held of gates) held.release()
      for (const held of gates.filter(item => item.received)) await expect.poll(() => held.completed).toBe(true)
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('detail-evidence.json', { body: JSON.stringify({ reads: state.reads, browserCalls, unexpected: state.unexpected, external, errors, rawErrors }, null, 2), contentType: 'application/json' })
        await testInfo.attach('detail-nitro.log', { body: app.logs(), contentType: 'text/plain' })
      }
      detailApp.current = null
    }
  },
})
export { expect }

export async function assertDetailAppearance(scene: DetailScene) {
  const { page, theme } = scene
  await expect(scene.root).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  for (const [selector, size, height, weight] of [
    ['.game-detail-title', '24px', '31.9999px', '700'], ['.game-detail-summary', '14px', '22.75px', '400'],
    ['.game-detail-tag', '12px', '16px', '400'], ['.game-detail-tab', '14px', '20px', '400'],
    ['.game-detail-similar-title', '14px', '20px', '500'], ['.game-detail-similar-summary', '12px', '16px', '400'],
    ['.game-detail-similar-reason', '11px', '15.7143px', '400'],
  ]) {
    const node = page.locator(selector!).first()
    await expect(node).toHaveCSS('font-size', size!); await expect(node).toHaveCSS('line-height', height!); await expect(node).toHaveCSS('font-weight', weight!)
  }
  await expect(page.locator('.game-detail-cover')).toHaveCSS('border-radius', '13.12px')
  await expect(page.locator('.game-detail-cover')).toHaveCSS('box-shadow', theme === 'dark' ? 'none' : 'rgba(91, 62, 28, 0.03) 0px 4px 12px 0px')
  await expect(page.locator('.game-detail-tag').first()).toHaveCSS('border-radius', '8.8px')
  await expect(page.locator('.game-detail-tag-tip').first()).toHaveCSS('background-color', theme === 'dark' ? 'rgba(2, 6, 23, 0.96)' : 'rgba(31, 41, 55, 0.94)')
  await expect(page.locator('.game-detail-action--primary')).toHaveCSS('font-weight', '750')
  await expect(page.locator('.link-tag').first()).toHaveCSS('border-radius', '9.92px')
  await expect(page.locator('.game-detail-title')).toHaveCSS('color', theme === 'dark' ? 'rgba(241, 245, 249, 0.88)' : 'rgba(71, 42, 20, 0.92)')
  const card = page.locator('.game-detail-content-card, .game-detail-comment, .game-detail-news-item').first()
  if (await card.count()) {
    await expect(card).toHaveCSS('background-color', theme === 'dark' ? 'rgba(226, 232, 240, 0.067)' : 'rgba(255, 250, 242, 0.42)')
    await expect(card).toHaveCSS('border-radius', '11.52px')
    await expect(card).toHaveCSS('box-shadow', theme === 'dark' ? 'none' : 'rgba(91, 62, 28, 0.03) 0px 4px 12px 0px')
  }
  for (const [selector, size, height, weight] of [
    ['.game-detail-load-more', '14px', '20px', '700'], ['.game-detail-news-title', '18px', '28.0001px', '700'],
    ['.game-detail-info-label', '14px', '20px', '760'], ['.game-detail-comment-body', '14px', '20px', '400'],
    ['.game-detail-intro-card', '14px', '24.08px', '400'],
  ]) {
    const node = page.locator(selector!).first()
    if (!await node.count()) continue
    await expect(node).toHaveCSS('font-size', size!); await expect(node).toHaveCSS('line-height', height!); await expect(node).toHaveCSS('font-weight', weight!)
  }
}
