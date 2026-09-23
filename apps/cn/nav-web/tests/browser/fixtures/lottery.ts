import { test as base, expect, type Locator, type Page, type Request, type Route } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type Theme = 'light' | 'dark'
type Locale = 'zh' | 'en'
type Reply = 'success' | 'empty' | 'rejected' | 'unavailable'
type Submission = { id: number, name: string, email: string, key: string }
type Gate = { promise: Promise<void>, received: boolean, completed: boolean, request?: Request,
  release(): void, waitReceived(): Promise<void>, waitCompleted(): Promise<void> }
type PostGate = Gate & { reply: Exclude<Reply, 'empty'>, body?: Submission }
type State = { replies: Reply[], reads: string[], unexpected: string[], gates: Gate[], elapsed: boolean }
type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Worker = { app: App, current: State | null }
const readPath = '/api/v2/game/prizes', postPath = `${readPath}/participation`
const networkDiagnostic = 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)'
export const lotteryNow = '2026-09-18T12:40:00+08:00'
export const lotteryRejection = '本地拒绝：抽奖密钥错误'
export const lotteryDraft: Submission = { id: 7301, name: 'Audit Member', email: 'audit@example.test', key: 'fixture-key' }

function payload(elapsed = false) {
  const prize = { title: 'Fixture Game Bundle', platform: 'Steam', count: 2 }
  return { active: [
    { lottery: { id: '7301', title: '社区游戏奖池', desc: '这是用于契约验收的固定奖池，报名后通过邮件完成激活。',
      start_time: '2026-09-17 12:40:00', end_time: '2026-09-19 12:40:00', prize }, count: 12,
    member: Array.from({ length: 12 }, (_, index) => ({ name: `Member ${index + 1}`, email: `m${index + 1}***@example.test` })) },
    { lottery: { id: '7302', title: '周末特别奖池', desc: '第二个固定奖池，目前还没有参与者。',
      start_time: elapsed ? '2026-09-15 12:40:00' : '2026-09-20 12:40:00',
      end_time: elapsed ? '2026-09-17 12:40:00' : '2026-09-22 12:40:00', prize: { ...prize, count: 1 } }, count: 0, member: [] },
  ], history: { prize_count: 9, prize: [
    { name: '秋日游戏活动', desc: '固定的历史活动与中奖记录。', end_time: '2026-09-10 18:00:00', prize, count: 24,
      winner: [{ name: 'Winner One', email: 'one***@example.test' }, { name: 'Winner Two', email: 'two***@example.test' }] },
    { name: '社区体验活动', desc: '历史活动的无中奖记录状态。', end_time: '2026-09-01 18:00:00', prize: { ...prize, count: 1 }, count: 3, winner: [] },
  ] } }
}
function makeGate(): Gate {
  let release!: () => void
  const promise = new Promise<void>(resolve => { release = resolve })
  const gate: Gate = { promise, release, received: false, completed: false,
    async waitReceived() { await expect.poll(() => gate.received && Boolean(gate.request)).toBe(true) },
    async waitCompleted() { await expect.poll(() => gate.completed).toBe(true) },
  }
  return gate
}
type OpenOptions = { theme?: Theme, locale?: Locale, width?: number, ready?: boolean, activation?: 'success' | 'fail', message?: string, elapsed?: boolean }
export type LotteryScene = {
  page: Page, root: Locator, modal: Locator, dialog: Locator, inputs: Locator, submit: Locator,
  theme: Theme, locale: Locale, paused: boolean, rendered: string, reads: string[], submissions: Submission[],
  planReads(replies: Reply[]): void, holdRead(): Gate, queueSubmit(reply?: PostGate['reply']): PostGate,
  open(options?: OpenOptions): Promise<void>, ready(): Promise<void>, openModal(index?: number): Promise<void>,
  fill(values?: Partial<Submission>): Promise<void>, expectAbort(gate: Gate): void, waitAborted(gate: Gate): Promise<void>,
  assertQuiet(counts?: { gets: number, upstream: number, posts?: number }): void,
}

export const test = base.extend<{ lottery: LotteryScene }, { lotteryApp: Worker }>({
  timezoneId: 'Asia/Shanghai',
  // eslint-disable-next-line no-empty-pattern -- Playwright fixture arguments must be destructured.
  lotteryApp: [async ({}, use) => {
    const worker = { current: null } as Worker
    worker.app = await startInsightsFixtureApp(async (url: URL) => {
      const state = worker.current
      if (!state) throw new Error('Lottery request outside its test scenario')
      state.reads.push(url.pathname)
      if (url.pathname !== readPath || url.search) { state.unexpected.push(`upstream ${url.href}`); return { status: 500 } }
      const reply = state.replies.shift() ?? 'success'
      const gate = state.gates.find(item => !item.received)
      if (gate) { gate.received = true; await gate.promise; gate.completed = true }
      if (reply === 'unavailable') return { status: 503 }
      if (reply === 'rejected') return { status: 200 }
      return { data: reply === 'empty' ? { active: [], history: { prize_count: 0, prize: [] } } : payload(state.elapsed) }
    }, { TZ: 'Asia/Shanghai' })
    try { await use(worker) } finally { await worker.app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ lotteryApp }, use) => { await use(lotteryApp.app.base) },
  lottery: async ({ page, context, lotteryApp }, use, testInfo) => {
    expect(lotteryApp.current).toBeNull()
    const state: State = { replies: [], reads: [], unexpected: [], gates: [], elapsed: false }
    lotteryApp.current = state
    const { app } = lotteryApp, upstreamStart = app.requests.length
    const expectedURLs = new Set<string>(), errors = captureBrowserErrors(page, expectedURLs)
    const rawConsole: { type: string, text: string, url: string }[] = []
    const gets: Request[] = [], posts: PostGate[] = [], submissions: Submission[] = []
    const failed: Request[] = [], aborted = new Set<Request>(), received503 = new Set<Request>()
    const external: string[] = [], handlers: string[] = [], tasks = new Set<Promise<void>>()
    let opened = false, read503Budget = 0
    page.on('console', m => rawConsole.push({ type: m.type(), text: m.text(), url: m.location().url }))
    context.on('page', () => state.unexpected.push('unexpected popup'))
    context.on('requestfailed', request => failed.push(request))
    context.on('request', request => {
      const url = new URL(request.url())
      if (url.origin !== app.base || !url.pathname.startsWith('/api/')) return
      if (url.pathname === readPath && request.method() === 'GET' && !url.search) {
        gets.push(request)
        const gate = state.gates.find(item => !item.request)
        if (gate) gate.request = request
      } else if (url.pathname !== postPath || request.method() !== 'POST' || url.search) state.unexpected.push(`browser ${request.method()} ${url.href}`)
    })
    context.on('response', response => {
      if (response.status() < 400) return
      const request = response.request()
      if (response.status() === 503 && (gets.includes(request) && read503Budget > 0
        || posts.some(gate => gate.request === request && gate.reply === 'unavailable'))) received503.add(request)
      else state.unexpected.push(`HTTP ${response.status()} ${response.url()}`)
    })
    const handlePost = async (route: Route) => {
      const gate = posts.find(item => !item.received), request = route.request()
      if (!gate || request.method() !== 'POST' || new URL(request.url()).search) {
        state.unexpected.push('unplanned Lottery submission'); await route.abort('blockedbyclient'); return
      }
      const body = request.postDataJSON() as Submission
      expect(Object.keys(body).sort()).toEqual(['email', 'id', 'key', 'name'])
      expect([7301, 7302]).toContain(body.id)
      for (const field of ['name', 'email', 'key'] as const) expect(typeof body[field]).toBe('string')
      submissions.push(body); Object.assign(gate, { body, request, received: true })
      await gate.promise
      await route.fulfill({ status: gate.reply === 'unavailable' ? 503 : 200, contentType: 'application/json',
        body: JSON.stringify(gate.reply === 'success' ? { code: 1, data: '' } : { code: 0, data: lotteryRejection }) })
      gate.completed = true
    }
    await context.route('**/*', async route => {
      const request = route.request(), url = new URL(request.url())
      if (url.origin !== app.base) {
        if (['data:', 'blob:'].includes(url.protocol)) return route.continue()
        external.push(url.href); return route.abort('blockedbyclient')
      }
      if (url.pathname === postPath) {
        const task = handlePost(route).catch(error => { handlers.push(String(error)) })
        tasks.add(task); try { await task } finally { tasks.delete(task) }
        return
      }
      if (url.pathname.startsWith('/api/') && (url.pathname !== readPath || request.method() !== 'GET' || url.search)) {
        state.unexpected.push(`blocked ${request.method()} ${url.href}`); return route.abort('blockedbyclient')
      }
      return route.continue()
    })
    const modal = page.locator('.lottery-modal'), dialog = page.locator('.lottery-modal__dialog')
    const scene: LotteryScene = {
      page, modal, dialog, root: page.locator('.lottery-page'), inputs: dialog.locator('input'), submit: dialog.locator('.lottery-modal__button--primary'),
      theme: 'light', locale: 'zh', paused: false, rendered: '', reads: state.reads, submissions,
      planReads(replies) {
        expect(state.reads).toHaveLength(0)
        state.replies = replies.flatMap(reply => reply === 'unavailable' ? Array<Reply>(4).fill(reply) : [reply])
        read503Budget = replies.filter(reply => reply === 'unavailable').length * 2
        if (read503Budget) expectedURLs.add(`${app.base}${readPath}`)
      },
      holdRead() { const gate = makeGate(); state.gates.push(gate); return gate },
      queueSubmit(reply = 'success') {
        const gate = Object.assign(makeGate(), { reply })
        posts.push(gate)
        if (reply === 'unavailable') expectedURLs.add(`${app.base}${postPath}`)
        return gate
      },
      async open({ theme = 'light', locale = 'zh', width = 1440, ready = true, activation, message, elapsed = false } = {}) {
        expect(opened).toBe(false); opened = true
        Object.assign(scene, { theme, locale, paused: Boolean(activation) }); state.elapsed = elapsed
        await page.setViewportSize({ width, height: width < 768 ? 844 : 900 })
        if (activation) { await page.clock.install({ time: new Date(lotteryNow) }); await page.clock.pauseAt(new Date(lotteryNow)) }
        else await page.clock.setFixedTime(new Date(lotteryNow))
        await context.addInitScript(({ origin, theme, checkedAt, steamKey, sample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
          localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample, china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
        }, { origin: app.base, theme, checkedAt: Date.parse(lotteryNow), steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
        const path = `${locale === 'en' ? '/en' : ''}/games/prize${activation ? `/activation?${new URLSearchParams({ status: activation, ...(message === undefined ? {} : { msg: message }) })}` : ''}`
        const response = await page.goto(path, { waitUntil: 'load' })
        expect(response?.status()).toBe(200); expect(response!.headers()['x-robots-tag']).toBe('noindex, follow')
        scene.rendered = await response!.text()
        expect(scene.rendered).not.toContain('class="lottery-page ')
        expect(scene.rendered).not.toContain('class="activation-card ')
        expect(scene.rendered).not.toContain('社区游戏奖池')
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
        await expect(page.locator('meta[name="robots"]')).toHaveAttribute('content', /noindex/)
        await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
        if (activation) { await expect(page.locator('.activation-card')).toBeVisible(); expect(state.reads).toHaveLength(0) }
        else if (ready) await scene.ready()
      },
      async ready() { await expect(page.locator('.lottery-page__loading')).toHaveCount(0); await expect(page.locator('.lottery-pool')).toHaveCount(2) },
      async openModal(index = 0) { await page.locator('.lottery-pool').nth(index).click(); await expect(dialog).toBeVisible() },
      async fill(values = {}) {
        const body = { ...lotteryDraft, ...values }
        for (const [index, field] of (['key', 'name', 'email'] as const).entries()) await scene.inputs.nth(index).fill(body[field])
      },
      expectAbort(gate) { expect(gate.request).toBeDefined(); aborted.add(gate.request!) },
      async waitAborted(gate) { await expect.poll(() => failed.includes(gate.request!)).toBe(true) },
      assertQuiet(counts = { gets: scene.paused ? 0 : 1, upstream: scene.paused ? 0 : 1, posts: 0 }) {
        expect(errors).toEqual([]); expect(external).toEqual([]); expect(handlers).toEqual([]); expect(state.unexpected).toEqual([])
        expect(gets).toHaveLength(counts.gets); expect(state.reads).toEqual(Array(counts.upstream).fill(readPath))
        expect(app.requests.slice(upstreamStart).map((url: URL) => url.pathname)).toEqual(state.reads)
        expect(submissions).toHaveLength(counts.posts ?? 0); expect(posts).toHaveLength(counts.posts ?? 0)
        expect(failed).toHaveLength(aborted.size)
        for (const request of failed) { expect(aborted.has(request)).toBe(true); expect(request.failure()?.errorText).toBe('net::ERR_ABORTED') }
        expect(received503.size).toBe(read503Budget + posts.filter(gate => gate.reply === 'unavailable').length)
        for (const url of expectedURLs) {
          expect(rawConsole.filter(item => item.type === 'error' && item.url === url)).toEqual([...received503].filter(request => request.url() === url).map(() => ({ type: 'error', text: networkDiagnostic, url })))
        }
        expect(rawConsole.filter(item => item.type === 'error' && !expectedURLs.has(item.url))).toEqual([])
        for (const gate of [...state.gates, ...posts]) { expect(gate.received).toBe(true); expect(gate.completed).toBe(true) }
      },
    }
    try { await use(scene) }
    finally {
      for (const gate of [...state.gates, ...posts]) gate.release()
      for (const gate of state.gates.filter(item => item.received)) await gate.waitCompleted()
      await Promise.allSettled(tasks)
      if (testInfo.status === testInfo.expectedStatus) expect(handlers).toEqual([])
      lotteryApp.current = null
    }
  },
})
export { expect }

export async function settleLottery(scene: LotteryScene, target: Locator) {
  await target.evaluate(async root => {
    if (document.activeElement instanceof HTMLElement) document.activeElement.blur()
    await Promise.all([...root.querySelectorAll('img')].map(img => img.complete ? Promise.resolve() : img.decode()))
    await document.fonts.ready
    await Promise.all(root.getAnimations({ subtree: true }).filter(a => a.effect?.getComputedTiming().iterations !== Infinity).map(a => a.finished.catch(() => {})))
  })
  await scene.page.mouse.move(1, 1)
  // Register both frames before advancing a paused clock; locator resolution is async.
  const readiness = await target.evaluateHandle(() => ({ frames: new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))) }))
  try {
    if (scene.paused) await scene.page.clock.runFor(64)
    await readiness.evaluate(async value => { await value.frames })
  } finally { await readiness.dispose() }
  expect(await scene.page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
}

export async function assertLotteryAppearance(scene: LotteryScene) {
  await expect(scene.root).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  const title = scene.page.locator('.lottery-page__title'), mobile = scene.page.viewportSize()!.width < 640
  await expect(title).toHaveCSS('font-size', mobile ? '36px' : '60px'); await expect(title).toHaveCSS('line-height', mobile ? '45px' : '75px')
  await expect(title).toHaveCSS('font-weight', '600')
  for (const [selector, size, height, weight] of [
    ['.lottery-summary__label', '11px', '16.5px', '400'], ['.lottery-summary__value', '30px', '36px', '600'],
    ['.lottery-pool__title', '18px', '28px', '600'], ['.lottery-pool__desc', '14px', '24px', '400'],
    ['.lottery-meta', '14px', '20px', '400'], ['.lottery-history__title', '16px', '24px', '600'], ['.lottery-winner', '11px', '16.5px', '400'],
  ]) {
    const node = scene.page.locator(selector!).first()
    await expect(node).toHaveCSS('font-size', size!); await expect(node).toHaveCSS('line-height', height!); await expect(node).toHaveCSS('font-weight', weight!)
  }
  const pool = scene.page.locator('.lottery-pool').first()
  await expect(pool).toHaveCSS('border-radius', '8px'); await expect(pool).toHaveCSS('backdrop-filter', 'blur(24px)')
  await expect(pool).not.toHaveCSS('box-shadow', 'none'); await expect(pool).toHaveCSS('transition-duration', '0.3s')
  await expect(pool).toHaveCSS('background-color', scene.theme === 'dark' ? 'rgba(226, 232, 240, 0.07)' : 'rgba(255, 255, 255, 0.055)')
  await expect(scene.page.locator('.lottery-summary')).toHaveCSS('background-color', 'rgba(255, 255, 255, 0.1)')
  await expect(scene.page.locator('.lottery-history')).toHaveCSS('background-color', scene.theme === 'dark' ? 'rgba(2, 6, 23, 0.36)' : 'rgba(0, 0, 0, 0.24)')
}

export async function assertLotteryModal(scene: LotteryScene) {
  await expect(scene.dialog).toHaveAttribute('role', 'dialog'); await expect(scene.dialog).toHaveAttribute('aria-modal', 'true')
  expect(await scene.modal.evaluate(el => el.parentElement === document.body)).toBe(true)
  expect(await scene.modal.evaluate(el => {
    const box = el.getBoundingClientRect()
    return box.x === 0 && box.y === 0 && box.width === innerWidth && box.height === innerHeight
      && [[2, 2], [innerWidth / 2, 20], [innerWidth - 2, innerHeight - 2]].every(([x, y]) => el.contains(document.elementFromPoint(x!, y!)))
  })).toBe(true)
  await expect(scene.dialog).toHaveCSS('border-radius', '12px'); await expect(scene.dialog).toHaveCSS('backdrop-filter', 'blur(24px)')
  await expect(scene.modal).toHaveCSS('backdrop-filter', 'blur(12px)')
  await expect(scene.dialog).toHaveCSS('background-color', scene.theme === 'dark' ? 'rgba(14, 20, 28, 0.94)' : 'rgba(21, 20, 18, 0.92)')
  await expect(scene.dialog).toHaveCSS('box-shadow', 'rgba(0, 0, 0, 0.45) 0px 24px 90px 0px')
  for (const input of [...await scene.inputs.all(), scene.submit, scene.dialog.locator('.lottery-modal__button--secondary')]) {
    await expect(input).toHaveCSS('font-size', '16px'); await expect(input).toHaveCSS('line-height', '24px'); await expect(input).toHaveCSS('font-weight', '400')
    await expect(input).toHaveCSS('border-radius', '8px'); await expect(input).toHaveCSS('transition-duration', '0.15s')
  }
}
