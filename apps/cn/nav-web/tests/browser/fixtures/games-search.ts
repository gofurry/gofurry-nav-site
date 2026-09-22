import { test as base, expect, type Locator, type Page, type Request } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type Kind = 'advanced' | 'simple' | 'tags' | 'home'
type Body = Record<string, unknown>
type Call = { kind: Kind, path: string, query: Record<string, string>, body?: Body }
type BrowserCall = Call & { method: string, request: Request }
type Gate = {
  kind: Kind, value: string, promise: Promise<void>, received?: Call, request?: Request, completed: boolean,
  release(): void, waitReceived(): Promise<Call>, waitCompleted(): Promise<void>,
}
type State = { calls: Call[], gates: Gate[], unexpected: string[] }
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Worker = { app: App, current: State | null }
const paths: Record<Kind, string> = { advanced: '/api/v2/game/search/page', simple: '/api/v2/game/search/simple',
  tags: '/api/v2/game/tag-categories', home: '/api/v2/game/home' }
const kindFor = (path: string) => (Object.keys(paths) as Kind[]).find(kind => paths[kind] === path)
const keyFor = (call: Call) => String(call.kind === 'tags' ? call.query.lang : call.body?.[call.kind === 'simple' ? 'txt' : 'content'] ?? '')
const game = (index: number, media: string, name: string) => ({ id: String(7100 + index), appid: 7100 + index,
  name: `${name} ${index + 1}`, info: 'A deterministic search lifecycle result.', cover: `${media}/search-${7100 + index}.svg`,
  update_time: '2026-09-01', release_date: '2026-08-28', remark_count: index, avg_score: 4.2,
  primary_tag: '', secondary_tag: '', release: null, first_available: null })
function responseFor(call: Call, media: string) {
  if (call.kind === 'home') return mockGameHome(media)
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
  page: Page, input: Locator, filter: Locator, keyword: Locator, pageSize: Locator,
  calls(kind: Kind): Call[], hold(kind: Exclude<Kind, 'home'>, value: string): Gate,
  open(options?: { home?: boolean, query?: Record<string, string> }): Promise<void>,
  openFilter(): Promise<void>, cancel(): Promise<void>, apply(): Promise<void>,
  leaf(name: string): Locator, sort(name: string): Locator,
  waitResults(name: string, count?: number): Promise<void>,
  expectAbort(gate: Gate): void, waitAborted(gate: Gate): Promise<void>,
  leaveForHome(): Promise<void>, assertQuiet(): void,
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
      const gate = state.gates.find(item => item.kind === kind && item.value === keyFor(call) && !item.received)
      if (gate) { gate.received = call; await gate.promise }
      const data = responseFor(call, media)
      if (gate) gate.completed = true
      return { data }
    }, { TZ: 'UTC' })
    try { await use(worker) } finally { await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ searchApp }, use) => { await use(searchApp.app.base) },
  search: async ({ page, context, searchApp }, use, testInfo) => {
    expect(searchApp.current).toBeNull()
    const state: State = { calls: [], gates: [], unexpected: [] }
    searchApp.current = state
    const { app } = searchApp
    const errors = captureBrowserErrors(page)
    const browserCalls: BrowserCall[] = [], external: string[] = []
    const failed: { request: Request, error: string | undefined }[] = []
    const expectedAborts = new Set<Request>()
    const rawConsole: { type: string, text: string, url: string }[] = []
    const assets = new Set([...Array.from({ length: 29 }, (_, index) => `${app.upstreamUrl}/media/search-${7100 + index}.svg`),
      ...[91, 92, 93].map(id => `${app.upstreamUrl}/media/game-${id}.svg`)])
    let opened = false, expectedHomeCalls = 0, expectedBrowserHomeCalls = 0
    page.on('console', message => rawConsole.push({ type: message.type(), text: message.text(), url: message.location().url }))
    context.on('page', () => state.unexpected.push('unexpected popup'))
    context.on('requestfailed', request => failed.push({ request, error: request.failure()?.errorText }))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin !== app.base || !url.pathname.startsWith('/api/')) return
      const kind = kindFor(url.pathname)
      if (!kind) { state.unexpected.push(`browser ${request.method()} ${url.pathname}`); return }
      const call: BrowserCall = { kind, path: url.pathname, query: Object.fromEntries(url.searchParams),
        body: request.postData() ? request.postDataJSON() : undefined, method: request.method(), request }
      browserCalls.push(call)
      expect(call.method).toBe(['advanced', 'simple'].includes(kind) ? 'POST' : 'GET')
      const gate = state.gates.find(item => item.kind === kind && item.value === keyFor(call) && !item.request)
      if (gate) gate.request = request
    })
    await context.route('**/*', route => {
      const request = route.request(), url = new URL(request.url())
      if (url.origin === app.base) {
        if (url.pathname.startsWith('/api/') && !kindFor(url.pathname)) return route.abort('blockedbyclient')
        return route.continue()
      }
      if (request.method() === 'GET' && assets.has(url.href)) return route.fulfill({ contentType: 'image/svg+xml', body: artwork })
      if (['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      return route.abort('blockedbyclient')
    })
    await context.addInitScript(({ origin, steamKey, sample }) => {
      if (location.origin !== origin) return
      localStorage.setItem('theme', 'light')
      const checkedAt = Date.now()
      localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
        primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
      localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
        china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
    }, { origin: app.base, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
    const filter = page.locator('.game-search-filter-panel')
    const assertQuiet = () => {
      expect(errors).toEqual([])
      expect(rawConsole.filter(item => item.type === 'error')).toEqual([])
      expect(external).toEqual([]); expect(state.unexpected).toEqual([])
      expect(failed).toHaveLength(expectedAborts.size)
      for (const request of expectedAborts) expect(failed.filter(item => item.request === request)).toHaveLength(1)
      for (const failure of failed) {
        expect(expectedAborts.has(failure.request)).toBe(true)
        expect(failure.error).toBe('net::ERR_ABORTED')
      }
      // Every browser call (including canceled ones) reached the real local API once.
      const wire = (call: Call) => JSON.stringify({ path: call.path, query: call.query, body: call.body })
      const mounted = state.calls.filter(call => call.kind !== 'home')
      expect(browserCalls.filter(call => call.kind !== 'home').map(wire).sort()).toEqual(mounted.map(wire).sort())
      expect(state.calls.filter(call => call.kind === 'home')).toHaveLength(expectedHomeCalls)
      expect(browserCalls.filter(call => call.kind === 'home')).toHaveLength(expectedBrowserHomeCalls)
      for (const call of state.calls.filter(item => item.kind === 'home')) expect(call.query).toEqual({ lang: 'zh', region: 'CN' })
      for (const gate of state.gates) { expect(gate.received).toBeDefined(); expect(gate.completed).toBe(true) }
    }
    const scene: SearchScene = {
      page, filter, input: page.locator('.game-sidebar-search-input'),
      keyword: filter.locator('.game-search-filter-input').first(), pageSize: filter.locator('.game-search-filter-input').nth(1),
      calls: kind => state.calls.filter(call => call.kind === kind),
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
      async open({ home = false, query = {} } = {}) {
        expect(opened).toBe(false); opened = true
        if (home) expectedHomeCalls++
        const path = home ? '/games' : `/games/search?${new URLSearchParams({ pageSize: '4', ...query })}`
        const response = await page.goto(path, { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const html = await response!.text()
        if (home) expect(html).toContain('Active game fixture')
        else expect(html).not.toContain('A deterministic search lifecycle result.')
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect(page.locator('html')).not.toHaveClass(/\bdark\b/)
        await expect(scene.input).toBeVisible()
        if (home) expect(scene.calls('home')).toHaveLength(1)
        else {
          await expect.poll(() => scene.calls('advanced').length).toBe(1)
          await expect.poll(() => scene.calls('tags').length).toBe(1)
          await scene.waitResults(query.content || 'Result', Number(query.pageSize || 4))
        }
      },
      async openFilter() { await page.locator('.search-filter-button').click(); await expect(filter).toBeVisible() },
      async cancel() { await filter.locator('.game-search-filter-action--ghost').click(); await expect(filter).toHaveCount(0) },
      async apply() { await filter.locator('.game-search-filter-action--primary').click(); await expect(filter).toHaveCount(0) },
      leaf: name => filter.locator('.game-search-filter-chip').filter({ hasText: new RegExp(`^${name} 4$`) }),
      sort: name => filter.locator('span.game-search-filter-chip').filter({ hasText: new RegExp(`^${name}$`) }),
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
        }, null, 2) })
        searchApp.current = null
      }
    }
  },
})

export { expect }
