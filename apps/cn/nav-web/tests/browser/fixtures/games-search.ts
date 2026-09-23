import { test as base, expect, type Locator, type Page, type Request } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type Kind = 'advanced' | 'simple' | 'tags' | 'home' | 'detail' | 'reviews' | 'insights' | 'daily' | 'similar' | 'view'
type Body = Record<string, unknown>
type Call = { kind: Kind, path: string, query: Record<string, string>, body?: Body }
type BrowserCall = Call & { method: string, request: Request, reply?: Reply }
type Reply = 'success' | 'empty' | 'rejected' | 503
type Script = { kind: Kind, value: string, replies: Reply[], upstreamReplies: Reply[], serverIndex: number, browserIndex: number }
type Gate = {
  kind: Kind, value: string, promise: Promise<void>, received?: Call, request?: Request, completed: boolean,
  release(): void, waitReceived(): Promise<Call>, waitCompleted(): Promise<void>,
}
type State = { calls: Call[], gates: Gate[], scripts: Script[], unexpected: string[], legacyTags: boolean }
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Worker = { app: App, current: State | null }
const paths: Record<Kind, string> = { advanced: '/api/v2/game/search/page', simple: '/api/v2/game/search/simple',
  tags: '/api/v2/game/tag-categories', home: '/api/v2/game/home', detail: '/api/v2/game/info',
  reviews: '/api/v2/game/reviews', insights: '/api/v2/game/games/7100/insights', daily: '/api/v2/game/games/7100/insights/daily',
  similar: '/api/v2/game/recommend/similar', view: '/api/v2/game/games/7100/view' }
const kindFor = (path: string) => (Object.keys(paths) as Kind[]).find(kind => paths[kind] === path)
const keyFor = (call: Call) => String(call.kind === 'tags' ? call.query.lang : call.body?.[call.kind === 'simple' ? 'txt' : 'content'] ?? '')
const game = (index: number, media: string, name: string) => ({ id: String(7100 + index), appid: 7100 + index,
  name: `${name} ${index + 1}`, info: 'A deterministic search lifecycle result.', cover: `${media}/search-${7100 + index}.svg`,
  update_time: '2026-09-01', release_date: '2026-08-28', remark_count: index, avg_score: 4.2,
  primary_tag: '', secondary_tag: '', release: null, first_available: null })
function responseFor(call: Call, media: string, legacyTags: boolean) {
  if (call.kind === 'home') return mockGameHome(media)
  if (call.kind === 'similar') return []
  if (call.kind === 'view') return { view_count: 1 }
  if (call.kind === 'detail') return { id: 7100, appid: 7100, name: 'Result 1', description: 'Search destination',
    content: '<p>Search destination</p>', platforms: { windows: true }, prices: [], tags: [], assets: {}, reviews: [], support_info: {}, external: {} }
  if (call.kind === 'reviews') return { data: [], count: 0 }
  if (call.kind === 'insights') return { appid: 7100, as_of: '2026-09-01', metrics: [], prices: [], linux: null, mac: null, release: null, as_of_date: '2026-09-01' }
  if (call.kind === 'daily') return { appid: 7100, as_of: '2026-09-01', window_days: 30, observed_days_30d: 0, sample_coverage_30d: 0, windows: {} }
  if (call.kind === 'tags' && legacyTags) return [{ id: '99', code: 'future-category', name: 'Catalog Category', tags: [
    { id: '812345', code: 'sample-tag', name: 'Leaf Tag', category: 'future-category', category_name: 'Catalog Category', game_count: 4 },
  ] }]
  if (call.kind === 'tags') return [18, 2].map((count, group) => ({
    id: String(99 + group), code: `category-${group}`, name: `${call.query.lang} Category ${group + 1}`,
    tags: Array.from({ length: count }, (_, index) => ({ id: String(812345 + group * 100 + index),
      code: `leaf-${group}-${index}`, name: `${call.query.lang} Leaf ${group + 1}-${index + 1}`,
      category_id: String(99 + group), category_code: `category-${group}`, category_name: `${call.query.lang} Category ${group + 1}`, game_count: 4 })),
  }))
  if (call.kind === 'simple') return Array.from({ length: 4 }, (_, index) => game(index, media, String(call.body?.txt)))
  const size = Number(call.body?.pageSize), offset = (Number(call.body?.pageNum) - 1) * size
  return { total: 29, list: Array.from({ length: Math.max(0, Math.min(size, 29 - offset)) },
    (_, index) => game(offset + index, media, String(call.body?.content || 'Result'))) }
}
const artwork = '<svg xmlns="http://www.w3.org/2000/svg" width="460" height="215"><rect width="460" height="215" fill="#394b44"/></svg>'

export type SearchScene = {
  page: Page, input: Locator, filter: Locator, keyword: Locator, pageSize: Locator, rendered: string,
  calls(kind: Kind): Call[], hold(kind: Exclude<Kind, 'home'>, value: string): Gate,
  respond(kind: Exclude<Kind, 'home'>, value: string, replies: Reply[]): void,
  open(options?: { home?: boolean, query?: Record<string, string>, locale?: 'zh' | 'en', theme?: 'light' | 'dark', ready?: boolean, legacyTags?: boolean, fixedNow?: string }): Promise<void>,
  openFilter(): Promise<void>, cancel(): Promise<void>, apply(): Promise<void>,
  leaf(name: string): Locator, sort(name: string): Locator,
  waitResults(name: string, count?: number): Promise<void>,
  expectAbort(gate: Gate): void, waitAborted(gate: Gate): Promise<void>,
  leaveForHome(): Promise<void>, forwardToHome(): Promise<void>, assertQuiet(): void,
  allowDetail(): void, allowSteam(): string,
}

export const test = base.extend<{ search: SearchScene }, { searchApp: Worker }>({
  timezoneId: 'UTC',
  // Only the server and active-scenario reference survive a test. All gates/data are test-owned.
  // eslint-disable-next-line no-empty-pattern -- Playwright requires destructured fixture arguments.
  searchApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp(async (url: URL, media: string, body?: Body) => {
      const state = worker.current
      if (!state) throw new Error('Search API request without an active test')
      const kind = kindFor(url.pathname)
      if (!kind) { state.unexpected.push(`upstream ${url.pathname}`); return { status: 500 } }
      const call: Call = { kind, path: url.pathname, query: Object.fromEntries(url.searchParams), body }
      state.calls.push(call)
      const script = state.scripts.find(item => item.kind === kind && item.value === keyFor(call))
      const reply = script?.upstreamReplies[script.serverIndex++] ?? 'success'
      const gate = state.gates.find(item => item.kind === kind && item.value === keyFor(call) && !item.received)
      if (gate) { gate.received = call; await gate.promise }
      if (gate) gate.completed = true
      if (reply === 503) return { status: 503 }
      // The upstream helper uses an explicit status to emit a non-success envelope.
      // HTTP 200 exercises useApi's business rejection without transport retry.
      if (reply === 'rejected') return { status: 200 }
      const data = reply === 'empty'
        ? kind === 'advanced' ? { total: 0, list: [] } : []
        : responseFor(call, media, state.legacyTags)
      return { data }
    }, { TZ: 'UTC' })
    try { await use(worker) } finally { await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ searchApp }, use) => { await use(searchApp.app.base) },
  search: async ({ page, context, searchApp }, use, testInfo) => {
    expect(searchApp.current).toBeNull()
    const state: State = { calls: [], gates: [], scripts: [], unexpected: [], legacyTags: false }
    searchApp.current = state
    const { app } = searchApp
    const expectedNetworkURLs = new Set<string>()
    const errors = captureBrowserErrors(page, expectedNetworkURLs)
    const browserCalls: BrowserCall[] = [], external: string[] = []
    const failed: { request: Request, error: string | undefined }[] = []
    const expectedAborts = new Set<Request>()
    const expectedHTTP = new Set<Request>(), receivedHTTP = new Set<Request>()
    const networkDiagnostic = 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)'
    const rawConsole: { type: string, text: string, url: string }[] = []
    const assets = new Set([...Array.from({ length: 29 }, (_, index) => `${app.upstreamUrl}/media/search-${7100 + index}.svg`),
      ...[91, 92, 93].map(id => `${app.upstreamUrl}/media/game-${id}.svg`)])
    let opened = false, expectedHomeCalls = 0, expectedBrowserHomeCalls = 0
    let detailAllowed = false, popupAllowed = false, popupCount = 0
    const steamURL = 'https://store.steampowered.com/app/7100'
    page.on('console', message => rawConsole.push({ type: message.type(), text: message.text(), url: message.location().url }))
    const popupErrors: string[][] = []
    context.on('page', popup => {
      if (!popupAllowed || ++popupCount !== 1) state.unexpected.push('unexpected popup')
      popupErrors.push(captureBrowserErrors(popup))
    })
    context.on('requestfailed', request => failed.push({ request, error: request.failure()?.errorText }))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin !== app.base || !url.pathname.startsWith('/api/')) return
      const kind = kindFor(url.pathname)
      if (!kind) { state.unexpected.push(`browser ${request.method()} ${url.pathname}`); return }
      if (['detail', 'reviews', 'insights', 'daily', 'similar', 'view'].includes(kind) && !detailAllowed) state.unexpected.push(`unexpected detail ${url.pathname}`)
      if (['detail', 'reviews', 'similar'].includes(kind)) expect(url.searchParams.get('id')).toBe('7100')
      const call: BrowserCall = { kind, path: url.pathname, query: Object.fromEntries(url.searchParams),
        body: request.postData() ? request.postDataJSON() : undefined, method: request.method(), request }
      browserCalls.push(call)
      const script = state.scripts.find(item => item.kind === kind && item.value === keyFor(call))
      call.reply = script?.replies[script.browserIndex++] ?? 'success'
      if (call.reply === 503) {
        expectedHTTP.add(request)
        expectedNetworkURLs.add(request.url())
      }
      expect(call.method).toBe(['advanced', 'simple', 'view'].includes(kind) ? 'POST' : 'GET')
      const gate = state.gates.find(item => item.kind === kind && item.value === keyFor(call) && !item.request)
      if (gate) gate.request = request
    })
    context.on('response', response => {
      if (response.status() < 400 || new URL(response.url()).origin !== app.base) return
      if (response.status() === 503 && expectedHTTP.has(response.request())) receivedHTTP.add(response.request())
      else state.unexpected.push(`HTTP ${response.status()} ${response.url()}`)
    })
    await context.route('**/*', route => {
      const request = route.request(), url = new URL(request.url())
      if (url.origin === app.base) {
        if (url.pathname.startsWith('/api/') && !kindFor(url.pathname)) return route.abort('blockedbyclient')
        return route.continue()
      }
      if (request.method() === 'GET' && assets.has(url.href)) return route.fulfill({ contentType: 'image/svg+xml', body: artwork })
      if (popupAllowed && request.isNavigationRequest() && request.method() === 'GET' && url.href === steamURL) {
        return route.fulfill({ contentType: 'text/html', body: '<!doctype html><title>Isolated Steam destination</title>' })
      }
      if (['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      return route.abort('blockedbyclient')
    })
    const seedStorage = (theme: 'light' | 'dark', fixedNow?: string) => context.addInitScript(({ origin, steamKey, sample, theme, fixedNow }) => {
      if (location.origin !== origin) return
      localStorage.setItem('theme', theme)
      // Cache timestamps use the scenario clock, independent of init-script ordering.
      const checkedAt = fixedNow ? Date.parse(fixedNow) : Date.now()
      localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
        primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
      localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
        china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
    }, { origin: app.base, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0], theme, fixedNow })
    const filter = page.locator('.game-search-filter-panel')
    const assertQuiet = () => {
      expect(errors).toEqual([])
      const diagnostics = rawConsole.filter(item => item.type === 'error')
      expect(diagnostics.filter(item => item.text !== networkDiagnostic || !expectedNetworkURLs.has(item.url))).toEqual([])
      // Console events lack Request identity. Bound the exact diagnostic quota to verified
      // request instances/statuses, so another failure at the same URL cannot be ignored.
      expect(receivedHTTP).toEqual(new Set([...expectedHTTP].filter(request => !expectedAborts.has(request))))
      for (const url of expectedNetworkURLs) {
        expect(diagnostics.filter(item => item.url === url)).toHaveLength([...receivedHTTP].filter(request => request.url() === url).length)
      }
      expect(external).toEqual([]); expect(state.unexpected).toEqual([])
      expect(popupCount).toBe(popupAllowed ? 1 : 0); expect(popupErrors.flat()).toEqual([])
      expect(failed).toHaveLength(expectedAborts.size)
      for (const request of expectedAborts) expect(failed.filter(item => item.request === request)).toHaveLength(1)
      for (const failure of failed) {
        expect(expectedAborts.has(failure.request)).toBe(true)
        expect(failure.error).toBe('net::ERR_ABORTED')
      }
      // The real Nitro proxy retries a failing GET once, independently of the browser's
      // ofetch retry. Preserve and account for both transport layers exactly.
      const wire = (call: Call) => JSON.stringify({ path: call.path, query: call.query, body: call.body })
      const mounted = state.calls.filter(call => call.kind !== 'home')
      expect(browserCalls.filter(call => call.kind !== 'home').flatMap(call =>
        call.kind === 'tags' && call.reply === 503 ? [wire(call), wire(call)] : [wire(call)],
      ).sort()).toEqual(mounted.map(wire).sort())
      expect(state.calls.filter(call => call.kind === 'home')).toHaveLength(expectedHomeCalls)
      expect(browserCalls.filter(call => call.kind === 'home')).toHaveLength(expectedBrowserHomeCalls)
      for (const call of state.calls.filter(item => item.kind === 'home')) expect(call.query).toEqual({ lang: 'zh', region: 'CN' })
      for (const gate of state.gates) { expect(gate.received).toBeDefined(); expect(gate.completed).toBe(true) }
    }
    const scene: SearchScene = {
      page, filter, rendered: '', input: page.locator('.game-sidebar-search-input'),
      allowDetail() { detailAllowed = true },
      allowSteam() { popupAllowed = true; return steamURL },
      keyword: filter.locator('.game-search-filter-input').first(), pageSize: filter.locator('.game-search-filter-input').nth(1),
      calls: kind => state.calls.filter(call => call.kind === kind),
      respond(kind, value, replies) {
        expect(state.scripts.some(script => script.kind === kind && script.value === value)).toBe(false)
        const upstreamReplies = replies.flatMap(reply => kind === 'tags' && reply === 503 ? [reply, reply] : [reply])
        state.scripts.push({ kind, value, replies, upstreamReplies, serverIndex: 0, browserIndex: 0 })
      },
      hold(kind, value) {
        let release!: () => void
        const promise = new Promise<void>(resolve => { release = resolve })
        const gate: Gate = { kind, value, promise, release, completed: false,
          async waitReceived() {
            await expect.poll(() => Boolean(gate.received && gate.request)).toBe(true)
            return gate.received!
          },
          async waitCompleted() { await expect.poll(() => gate.completed).toBe(true) },
        }
        state.gates.push(gate)
        return gate
      },
      async open({ home = false, query = {}, locale = 'zh', theme = 'light', ready = true, legacyTags = false, fixedNow } = {}) {
        expect(opened).toBe(false); opened = true
        state.legacyTags = legacyTags
        if (fixedNow) await page.clock.setFixedTime(new Date(fixedNow))
        await seedStorage(theme, fixedNow)
        if (home) expectedHomeCalls++
        const path = (locale === 'en' ? '/en' : '') + (home ? '/games' : `/games/search?${new URLSearchParams({ pageSize: '4', ...query })}`)
        const response = await page.goto(path, { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const html = await response!.text()
        scene.rendered = html
        if (home) expect(html).toContain('Active game fixture')
        else expect(html).not.toContain('A deterministic search lifecycle result.')
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        if (theme === 'dark') await expect(page.locator('html')).toHaveClass(/\bdark\b/)
        else await expect(page.locator('html')).not.toHaveClass(/\bdark\b/)
        await expect(scene.input).toBeVisible()
        if (home) expect(scene.calls('home')).toHaveLength(1)
        else if (ready) {
          await expect.poll(() => scene.calls('advanced').length).toBe(1)
          await expect.poll(() => scene.calls('tags').length).toBe(1)
          await scene.waitResults(query.content || 'Result', Number(query.pageSize || 4))
        }
      },
      async openFilter() { await page.locator('.search-filter-button').click(); await expect(filter).toBeVisible() },
      async cancel() { await filter.locator('.game-search-filter-action--ghost').click(); await expect(filter).toHaveCount(0) },
      async apply() { await filter.locator('.game-search-filter-action--primary').click(); await expect(filter).toHaveCount(0) },
      leaf: name => filter.locator('.game-search-filter-chip').filter({ hasText: new RegExp(`^${name} 4$`) }),
      sort: name => filter.getByRole('button', { name, exact: true }),
      async waitResults(name, count = 4) {
        await expect(page.locator('.search-results')).not.toHaveClass(/search-results--(?:pending|sliding)/)
        const titles = page.locator('.search-result-page-slide:not([aria-hidden="true"]) .search-page-title')
        await expect(titles).toHaveCount(count)
        await expect(titles.first()).toContainText(name)
      },
      expectAbort(gate) { expect(gate.request).toBeDefined(); expectedAborts.add(gate.request!) },
      async waitAborted(gate) { await expect.poll(() => failed.some(item => item.request === gate.request)).toBe(true) },
      async leaveForHome() {
        expectedHomeCalls++; expectedBrowserHomeCalls++
        await page.locator('.gf-nav__link').filter({ hasText: /^热门兽游$/ }).click()
        await expect(page).toHaveURL(`${app.base}/games`)
        await expect(page.locator('.games-search-page')).toHaveCount(0)
        await expect(page.locator('.game-group-page-live').first()).toBeVisible()
      },
      async forwardToHome() {
        expectedHomeCalls++; expectedBrowserHomeCalls++
        await page.goForward()
        await expect(page).toHaveURL(`${app.base}/games`)
        await expect(page.locator('.game-group-page-live').first()).toBeVisible()
      },
      assertQuiet,
    }
    try { await use(scene) }
    finally {
      for (const gate of state.gates) gate.release()
      // Our upstream gates are finite and fully released; context teardown owns route cleanup.
      for (const gate of state.gates.filter(item => item.received)) await gate.waitCompleted()
      try {
        if (testInfo.status === testInfo.expectedStatus && opened) assertQuiet()
      } finally {
        await testInfo.attach('search-lifecycle-evidence.json', { contentType: 'application/json', body: JSON.stringify({
          upstream: state.calls, browserCalls: browserCalls.map(({ request: _request, ...call }) => call),
          failed: failed.map(item => ({ url: item.request.url(), body: item.request.postData(), error: item.error,
            expected: expectedAborts.has(item.request) })), rawConsole, errors, external, unexpected: state.unexpected,
          injectedHTTP: [...expectedHTTP].map(request => ({ url: request.url(), method: request.method(), body: request.postData(), received: receivedHTTP.has(request) })),
        }, null, 2) })
        searchApp.current = null
      }
    }
  },
})

export { expect }

export async function assertSearchAppearance(scene: SearchScene, theme: 'light' | 'dark') {
  const { page } = scene
  // Shell's later rule owns the canvas; the Search root remains transparent.
  await expect(page.locator('.games-search-page')).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  const card = page.locator('.search-result-page-slide:not([aria-hidden="true"]) .search-page-card').first()
  await expect(card).toHaveCSS('border-radius', '14.72px')
  await expect(card).toHaveCSS('background-color', theme === 'dark' ? 'rgba(226, 232, 240, 0.067)' : 'rgba(255, 250, 242, 0.42)')
  await expect(card.locator('.search-page-title')).toHaveCSS('font-weight', '750')
  await expect(card.locator('.search-page-desc')).toHaveCSS('font-size', '12.48px')
  await expect(card.locator('.search-page-desc')).toHaveCSS('line-height', '16.848px')
  await expect(page.locator('.search-filter-button')).toHaveCSS('font-size', '14.4px')
  await expect(page.locator('.game-search-page-button').first()).toHaveCSS('font-size', '13.44px')
  await expect(page.locator('.game-search-page-button').first()).toHaveCSS('line-height', '20.16px')
}

export async function assertFilterAppearance(scene: SearchScene, theme: 'light' | 'dark') {
  const { filter } = scene
  await expect(filter).toHaveCSS('background-color', theme === 'dark' ? 'rgba(15, 23, 42, 0.94)' : 'rgba(255, 250, 242, 0.94)')
  await expect(filter).toHaveCSS('border-radius', '16px')
  await expect(filter).toHaveCSS('backdrop-filter', 'blur(18px)')
  await expect(filter.locator('.game-search-filter-action').first()).toHaveCSS('font-size', '13.76px')
  await expect(filter.locator('.game-search-filter-action').first()).toHaveCSS('line-height', '20.64px')
  await expect(filter.locator('.game-search-filter-chip').first()).toHaveCSS('font-size', '12.16px')
  await expect(filter.locator('.game-search-filter-chip').first()).toHaveCSS('line-height', '18.24px')
  await expect(filter.locator('.dp__input').first()).toHaveCSS('font-size', '13.76px')
  await expect(filter.locator('.dp__input').first()).toHaveCSS('border-radius', '12.48px')
}

// Only target-scoped work: no global Home chart animations or network-idle wait.
export async function searchClip(scene: SearchScene, selector: string, margin = 12) {
  const target = scene.page.locator(selector)
  await expect(target.first()).toBeVisible()
  await expect.poll(() => target.evaluateAll(elements => elements.every(element =>
    [...element.querySelectorAll('img')].filter(img => img.getClientRects().length && getComputedStyle(img).visibility !== 'hidden')
      .every(img => img.complete && img.naturalWidth > 0),
  ))).toBe(true)
  await target.evaluateAll(async elements => {
    await document.fonts.ready
    await Promise.all(elements.flatMap(element => element.getAnimations({ subtree: true }))
      .filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
      .map(animation => animation.finished.catch(() => {})))
    await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)))
  })
  const box = await target.evaluateAll((elements, margin) => {
    const rects = elements.map(element => element.getBoundingClientRect())
    const x = Math.max(0, Math.floor(Math.min(...rects.map(rect => rect.left)) - margin))
    const y = Math.max(0, Math.floor(Math.min(...rects.map(rect => rect.top)) - margin))
    const right = Math.min(innerWidth, Math.ceil(Math.max(...rects.map(rect => rect.right)) + margin))
    const bottom = Math.min(innerHeight, Math.ceil(Math.max(...rects.map(rect => rect.bottom)) + margin))
    return { x, y, width: right - x, height: bottom - y }
  }, margin)
  expect(box.width).toBeGreaterThan(0); expect(box.height).toBeGreaterThan(0)
  scene.assertQuiet()
  return box
}
