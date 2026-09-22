import { test as base, expect, type Locator, type Page, type Request, type Route } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type Host = 'home' | 'search' | 'detail'
type Language = 'zh' | 'en'
type Theme = 'light' | 'dark'
type Outcome = 'success' | 'rejected' | 'unavailable'
type OpenOptions = { host?: Host, locale?: Language, theme?: Theme, width?: number }
type Evidence = { path: string, query: Record<string, string>, body: unknown }
type BrowserCall = Evidence & { method: string }
export type ReviewBody = { id: string, name: string, content: string, score: number }
type Submission = BrowserCall & { body: ReviewBody }
type Gate = { release(): void, waitReceived(): Promise<Submission>, promise: Promise<void>, received?: Submission, outcome: Outcome }
type State = { host: Host, locale: Language, reads: Evidence[], unexpected: string[], opened: boolean }
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Worker = { app: App, current: State | null }
export const rejectionText = '本地拒绝：请稍后重试'
export const reviewDraft = { name: 'Reviewer', content: '这是一条本地契约测试点评。', score: '4.3' }
const submitPath = '/api/v2/game/reviews/anonymous'
const unavailableDiagnostic = 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)'
const copy = {
  zh: { title: '发表评论', required: '称呼和评价内容不能为空', score: '评分必须在 0.0 ~ 5.0 之间',
    pending: '提交中...', submit: '提交点评', success: '点评提交成功, 感谢您的参与', failure: '网络异常, 提交失败' },
  en: { title: 'Rating', required: 'Name and content are required.', score: 'Score must be between 0.0 and 5.0.',
    pending: 'Commiting...', submit: 'Commit Rating', success: 'Review submitted. Thank you!', failure: 'Submission failed. Please try again.' },
}
const evidence = (path: string, query: Record<string, string> = {}, body: unknown = undefined): Evidence => ({ path, query, body })
const searchBody = () => ({ pageNum: 1, pageSize: 20, content: '', availability: 'available', score: false,
  remark_order: false, time_order: true, tag_list: [], lang: 'zh' })
function expectedReads(host: Host, lang: Language): Evidence[] {
  if (host === 'home') return [evidence('/api/v2/game/home', { lang, region: 'CN' })]
  if (host === 'search') return [evidence('/api/v2/game/tag-categories', { lang: 'zh' }), evidence('/api/v2/game/search/page', {}, searchBody())]
  return [evidence('/api/v2/game/info', { id: '82', lang: 'zh', region: 'CN', news_limit: '20' }),
    evidence('/api/v2/game/reviews', { id: '82', page: '1', limit: '5' }),
    evidence('/api/v2/game/recommend/similar', { id: '82', lang: 'zh', region: 'CN', limit: '8' }),
    evidence('/api/v2/game/games/82/insights'), evidence('/api/v2/game/games/82/view')]
}
const sorted = <T extends Evidence>(calls: T[]) => [...calls].sort((a, b) => a.path.localeCompare(b.path))
function record(request: Request): BrowserCall {
  const url = new URL(request.url())
  return { ...evidence(url.pathname, Object.fromEntries(url.searchParams), request.postData() ? request.postDataJSON() : undefined), method: request.method() }
}
function detailData() {
  return { id: '82', appid: 82, name: 'Review detail fixture', summary: 'A fixed review consumer', about_the_game: '<p>Fixed introduction</p>',
    site: { view_count: 1, resources: [], groups: [], links: [] }, platforms: { windows: true }, prices: [], news: [], tags: [],
    developers: [], publishers: [], media: { screenshots: [], movies: [], assets: [] }, requirements: {}, support_info: {}, extra: {} }
}
function responseFor(path: string, media: string) {
  if (path === '/api/v2/game/home') return mockGameHome(media)
  if (path === '/api/v2/game/search/page') return { total: 1, list: [{ id: '83', appid: 83, name: 'Review search fixture', info: 'A fixed search result',
    cover: `${media}/game-83.svg`, update_time: '2026-09-01', release_date: '2026-08-28', remark_count: 0, avg_score: 0,
    primary_tag: '', secondary_tag: '', release: null, first_available: null }] }
  if (path === '/api/v2/game/info') return detailData()
  if (path === '/api/v2/game/reviews') return { total: 0, remarks: [] }
  if (path === '/api/v2/game/games/82/view') return { view_count: 2 }
  if (path === '/api/v2/game/games/82/insights') return {
    game: { id: 82, name: detailData().name }, state: { free: null, windows: null, mac: null, linux: null, release: null, as_of: null },
    players: { current: null, peak_30d: null, average_30d: null, as_of: null, observed_days_30d: 0, sample_coverage_30d: null },
    price: null, regional_prices: { as_of: null, regions: [] }, recent_changes: [],
  }
  return [] // Only the exact categories/similar entries below can reach this branch.
}
const artwork = '<svg xmlns="http://www.w3.org/2000/svg" width="460" height="215"><rect width="460" height="215" fill="#394b44"/><circle cx="350" cy="62" r="28" fill="#c8b398"/></svg>'

export type ReviewScene = {
  page: Page, dialog: Locator, backdrop: Locator, name: Locator, content: Locator, score: Locator, submit: Locator,
  close: Locator, feedback: Locator, text: typeof copy.zh | typeof copy.en, submissions: Submission[], rendered: string,
  openDialog(index?: number): Promise<void>, closeDialog(backdrop?: boolean): Promise<void>,
  fill(values?: Partial<typeof reviewDraft>): Promise<void>, queue(outcome?: Outcome): Gate,
  settle(): Promise<void>, assertQuiet(): void, assertBounds(): Promise<void>,
  visualClip(focused?: boolean): Promise<{ x: number, y: number, width: number, height: number }>,
}

export const test = base.extend<{ review: { open(options?: OpenOptions): Promise<ReviewScene> } }, { reviewApp: Worker }>({
  // Mutable responses belong to a test scenario; the worker only owns the server and its active-scenario reference.
  // eslint-disable-next-line no-empty-pattern -- Playwright fixture arguments must be destructured.
  reviewApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp((url: URL, media: string, body: unknown) => {
      const state = worker.current
      if (!state) throw new Error('Review request without a scenario')
      const call = evidence(url.pathname, Object.fromEntries(url.searchParams), body)
      state.reads.push(call)
      if (!expectedReads(state.host, state.locale).some(item => item.path === call.path)) {
        state.unexpected.push(`upstream ${url.pathname}`)
        return { status: 500 }
      }
      return { data: responseFor(call.path, media) }
    })
    try { await use(worker) }
    finally { await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ reviewApp }, use) => { await use(reviewApp.app.base) },
  review: async ({ page, context, reviewApp }, use, testInfo) => {
    expect(reviewApp.current, 'The previous scenario must be released').toBeNull()
    const state: State = { host: 'home', locale: 'zh', reads: [], unexpected: [], opened: false }
    reviewApp.current = state
    const { app } = reviewApp, upstreamStart = app.requests.length
    const expectedFailures = new Set<string>(), errors = captureBrowserErrors(page, expectedFailures)
    const rawConsole: { type: string, text: string, url: string }[] = []
    const browserCalls: BrowserCall[] = [], external: string[] = [], failed: string[] = [], handlerErrors: string[] = []
    const submissions: Submission[] = [], gates: Gate[] = [], tasks = new Set<Promise<void>>()
    const assets = new Set([91, 92, 93, 83].map(id => `${app.upstreamUrl}/media/game-${id}.svg`))
    page.on('console', message => rawConsole.push({ type: message.type(), text: message.text(), url: message.location().url }))
    context.on('page', () => state.unexpected.push('unexpected popup'))
    context.on('requestfailed', request => failed.push(`${request.method()} ${request.url()}: ${request.failure()?.errorText}`))
    context.on('request', request => {
      if (request.url().startsWith(`${app.base}/api/`)) browserCalls.push(record(request))
    })
    const queue = (outcome: Outcome = 'success'): Gate => {
      let release!: () => void
      const promise = new Promise<void>(resolve => { release = resolve })
      const gate: Gate = { outcome, promise, release, async waitReceived() {
        await expect.poll(() => gate.received, 'The actual submit request must reach the gate').toBeDefined()
        return gate.received!
      } }
      gates.push(gate)
      if (outcome === 'unavailable') expectedFailures.add(`${app.base}${submitPath}`)
      return gate
    }
    const handleSubmission = async (route: Route) => {
      const call = record(route.request()), gate = gates.find(item => !item.received)
      if (call.method !== 'POST' || Object.keys(call.query).length || !gate) {
        state.unexpected.push(`unplanned submit ${JSON.stringify(call)}`)
        await route.abort('blockedbyclient')
        return
      }
      const body = call.body as ReviewBody
      expect(Object.keys(body).sort()).toEqual(['content', 'id', 'name', 'score'])
      expect(typeof body.id).toBe('string'); expect(typeof body.name).toBe('string'); expect(typeof body.content).toBe('string')
      expect(typeof body.score).toBe('number'); expect(Number.isFinite(body.score)).toBe(true)
      const submission = { ...call, body }
      submissions.push(submission)
      gate.received = submission
      await gate.promise
      await route.fulfill({ status: gate.outcome === 'unavailable' ? 503 : 200, contentType: 'application/json',
        body: JSON.stringify(gate.outcome === 'success' ? { code: 1, data: 'ok' } : { code: 0, data: rejectionText }) })
    }
    await context.route('**/*', async route => {
      const request = route.request(), url = new URL(request.url())
      if (url.origin === app.base && url.pathname === submitPath) {
        const task = handleSubmission(route).catch(error => { handlerErrors.push(String(error)) })
        tasks.add(task)
        try { await task } finally { tasks.delete(task) }
        return
      }
      if (url.origin === app.base) {
        if (url.pathname.startsWith('/api/')) {
          const allowed = state.host === 'search' ? ['/api/v2/game/tag-categories', '/api/v2/game/search/page']
            : state.host === 'detail' ? ['/api/v2/game/games/82/view'] : []
          if (!allowed.includes(url.pathname)) {
            state.unexpected.push(`browser ${request.method()} ${url.pathname}`)
            return route.abort('blockedbyclient')
          }
        }
        return route.continue()
      }
      if (assets.has(url.href) && request.method() === 'GET') return route.fulfill({ contentType: 'image/svg+xml', body: artwork })
      if (['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const assertQuiet = () => {
      expect(state.unexpected).toEqual([]); expect(external).toEqual([]); expect(failed).toEqual([])
      expect(errors).toEqual([]); expect(handlerErrors).toEqual([])
      const expected = expectedReads(state.host, state.locale)
      expect(sorted(state.reads)).toEqual(sorted(expected))
      expect(app.requests.slice(upstreamStart).map((url: URL) => url.pathname).sort()).toEqual(expected.map(item => item.path).sort())
      const expectedBrowserReads = expected.filter(item => state.host === 'search' || item.path.endsWith('/view'))
        .map(item => ({ ...item, method: item.path.endsWith('/view') || item.path.endsWith('/search/page') ? 'POST' : 'GET' }))
      expect(sorted(browserCalls.filter(item => item.path !== submitPath))).toEqual(sorted(expectedBrowserReads))
      expect(browserCalls.filter(item => item.path === submitPath)).toEqual(submissions)
      expect(submissions).toHaveLength(gates.length)
      expect(rawConsole.filter(item => item.type === 'error')).toEqual(gates.filter(item => item.outcome === 'unavailable')
        .map(() => ({ type: 'error', text: unavailableDiagnostic, url: `${app.base}${submitPath}` })))
    }
    try {
      await use({ async open({ host = 'home', locale = 'zh', theme = 'light', width = 1440 } = {}) {
        expect(state.opened, 'One host navigation per test').toBe(false)
        if (host !== 'home') expect(locale).toBe('zh')
        Object.assign(state, { host, locale, opened: true })
        await page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
        await context.addInitScript(({ origin, theme, steamKey, sample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
            primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
        const path = host === 'home' ? `${locale === 'en' ? '/en' : ''}/games` : host === 'search' ? '/games/search' : '/games/82'
        const response = await page.goto(path, { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const html = await response!.text()
        const rendered = html.slice(html.indexOf('<body'), html.indexOf('</body>')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g, '')
        expect(rendered).not.toContain('class="review-dialog"')
        if (host === 'home') expect(rendered).toContain('Active game fixture')
        if (host === 'search') expect(rendered).not.toContain('Review search fixture')
        if (host === 'detail') expect(rendered).toContain(detailData().name)
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
        await expect.poll(() => state.reads.length).toBe(expectedReads(host, locale).length)
        const dialog = page.locator('.review-dialog'), backdrop = dialog.locator('..')
        const name = dialog.locator('input').first(), content = dialog.locator('textarea'), score = dialog.locator('.review-score-field')
        const submit = dialog.locator('.review-submit'), close = dialog.locator('.review-dialog__close')
        const feedback = submit.locator('..').locator(':scope > p')
        const settle = () => settleReview(backdrop)
        const assertBounds = async () => {
          const geometry = await backdrop.evaluate(el => {
            const b = el.getBoundingClientRect()
            return { x: b.x, y: b.y, width: b.width, height: b.height, w: innerWidth, h: innerHeight,
              overflow: document.documentElement.scrollWidth > innerWidth, filter: getComputedStyle(el).backdropFilter }
          })
          expect(geometry).toMatchObject({ x: 0, y: 0, width, height: width === 390 ? 844 : 900, overflow: false, filter: 'blur(2px)' })
          for (const target of [dialog, name, content, score, submit, close, ...await feedback.all()]) {
            const box = await target.boundingBox()
            expect(box).not.toBeNull()
            expect(box!.x).toBeGreaterThanOrEqual(-1); expect(box!.y).toBeGreaterThanOrEqual(-1)
            expect(box!.x + box!.width).toBeLessThanOrEqual(geometry.w + 1)
            expect(box!.y + box!.height).toBeLessThanOrEqual(geometry.h + 1)
          }
        }
        const scene: ReviewScene = { page, dialog, backdrop, name, content, score, submit, close, feedback, text: copy[locale], submissions, rendered,
          queue, settle, assertQuiet, assertBounds,
          async openDialog(index = 0) {
            await expect(dialog).toHaveCount(0)
            const trigger = host === 'home' ? page.locator('.game-group-page-live .review-float-button').nth(index)
              : host === 'search' ? page.locator('.search-result-page-slide:not([aria-hidden="true"]) .search-review-button').first()
                : page.locator('.game-detail-action--secondary').first()
            await trigger.focus(); await trigger.click()
            await expect(dialog).toBeVisible()
            expect(await dialog.evaluate(el => ({ inNuxt: Boolean(el.closest('#__nuxt')), inGames: Boolean(el.closest('.games-page')),
              body: el.parentElement?.parentElement === document.body }))).toEqual({ inNuxt: false, inGames: false, body: true })
            await expect(dialog.getByRole('heading')).toHaveText(copy[locale].title)
            await expect(dialog.locator(':scope > div').first().locator('p')).toHaveText(host === 'home'
              ? ['Active game fixture', 'Discount game fixture', 'Recent release fixture'][index]!
              : host === 'search' ? 'Review search fixture' : detailData().name)
            await settle()
          },
          async closeDialog(onBackdrop = false) {
            if (onBackdrop) await backdrop.click({ position: { x: 2, y: 2 } })
            else await close.click()
            await expect(dialog).toHaveCount(0)
          },
          async fill(values = {}) {
            const draft = { ...reviewDraft, ...values }
            await name.fill(draft.name); await content.fill(draft.content); await score.fill(draft.score)
          },
          async visualClip(focused = false) {
            await page.addStyleTag({ content: '#__nuxt { visibility: hidden; } body { background: var(--gf-page-background); }' })
            if (focused) await name.focus()
            else await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
            await page.mouse.move(1, 1)
            await settle(); await assertBounds()
            if (focused) await expect(name).toBeFocused()
            assertQuiet()
            const box = (await dialog.boundingBox())!, viewport = page.viewportSize()!
            const x = Math.max(0, box.x - 48), y = Math.max(0, box.y - 48)
            return { x, y, width: Math.min(viewport.width, box.x + box.width + 48) - x,
              height: Math.min(viewport.height, box.y + box.height + 48) - y }
          },
        }
        await expect(dialog).toHaveCount(0)
        assertQuiet()
        return scene
      } })
    } finally {
      // Only our finite submit handlers are awaited, after all gates are released.
      // Context teardown owns other route registrations; no unrouteAll(wait).
      for (const gate of gates) gate.release()
      await Promise.all([...tasks])
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('review-evidence.json', { body: JSON.stringify({ reads: state.reads, browserCalls, submissions,
          unexpected: state.unexpected, external, failed, errors, rawConsole, handlerErrors }, null, 2), contentType: 'application/json' })
        await testInfo.attach('review-nitro.log', { body: app.logs(), contentType: 'text/plain' })
      }
      try { if (state.opened && testInfo.status === testInfo.expectedStatus) assertQuiet() }
      finally { reviewApp.current = null }
    }
  },
})

export { expect }
export async function settleReview(target: Locator) {
  await target.evaluate(async el => {
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
    await Promise.all(el.getAnimations({ subtree: true }).filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
      .map(animation => animation.finished.catch(() => {})))
    await document.fonts.ready
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
  })
}
async function css(target: Locator, values: Record<string, string>) {
  for (const [property, value] of Object.entries(values)) await expect(target).toHaveCSS(property, value)
}
export async function assertReviewAppearance(scene: ReviewScene) {
  const { page, dialog, backdrop, name, content, score, submit, close } = scene
  await page.mouse.move(1, 1)
  await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
  await scene.settle()
  await css(dialog, { width: page.viewportSize()!.width === 390 ? '358px' : '480px', padding: '19.2px', 'border-width': '1px',
    'border-color': 'rgba(126, 92, 58, 0.18)', 'border-radius': '16.8px', 'background-color': 'rgba(255, 250, 242, 0.92)',
    'box-shadow': 'rgba(45, 28, 12, 0.2) 0px 22px 70px 0px', 'backdrop-filter': 'blur(12px)',
    'animation-duration': '0.18s', 'animation-timing-function': 'cubic-bezier(0.22, 1, 0.36, 1)' })
  await css(backdrop, { 'padding-left': '16px', 'padding-right': '16px', 'backdrop-filter': 'blur(2px)',
    'background-color': 'oklab(0.147 0.00261104 0.00303026 / 0.3)' })
  await css(dialog.getByRole('heading'), { 'font-size': '18px', 'line-height': '28.0001px', 'font-weight': '700', color: 'oklch(0.266 0.079 36.259)' })
  await css(dialog.locator(':scope > div').first().locator('p'), { 'font-size': '14px', 'line-height': '20px', 'font-weight': '400', color: 'oklch(0.553 0.013 58.071)' })
  await css(dialog.locator('label > span'), { 'font-size': '14px', 'line-height': '20px', 'font-weight': '500', color: 'oklch(0.444 0.011 73.639)' })
  for (const field of [name, content, score]) await css(field, { 'font-size': '16px', 'line-height': '24px', 'font-weight': '400',
    color: 'rgba(41, 37, 36, 0.92)', 'border-radius': '12px', 'border-color': 'rgba(126, 92, 58, 0.18)',
    'background-color': 'rgba(255, 255, 255, 0.46)', 'transition-property': 'border-color, background-color', 'transition-duration': '0.16s, 0.16s' })
  await css(submit, { 'font-size': '16px', 'line-height': '24px', 'font-weight': '700', color: 'rgb(255, 255, 255)',
    'border-radius': '12px', 'background-color': 'rgba(180, 83, 9, 0.86)', opacity: '1' })
  await css(close, { 'font-size': '16px', 'line-height': '24px', 'font-weight': '400', color: 'rgba(120, 83, 53, 0.72)' })
  await scene.assertBounds()
}
export async function assertReviewFocus(scene: ReviewScene) {
  await scene.name.focus(); await scene.settle()
  await expect(scene.name).toBeFocused()
  await css(scene.name, { 'border-color': 'rgba(249, 115, 22, 0.42)', 'background-color': 'rgba(255, 255, 255, 0.68)' })
}
export async function assertReviewFeedback(scene: ReviewScene, message: string, success = false) {
  await expect(scene.feedback).toHaveText(message)
  await css(scene.feedback, { 'font-size': '14px', 'line-height': '20px', color: success ? 'oklch(0.627 0.194 149.214)' : 'oklch(0.637 0.237 25.331)' })
}
