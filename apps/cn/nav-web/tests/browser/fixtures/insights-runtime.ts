import { test as base, expect, type Page, type Request, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'
import { fileURLToPath } from 'node:url'

type App = Awaited<ReturnType<typeof startInsightsFixtureApp>>
export interface Reply { status?: number; data?: unknown }
export interface APICall { url: URL; body?: unknown; status: number; completed: boolean }
interface Held {
  match(url: URL, body: unknown): boolean
  promise: Promise<void>
  received: boolean
  completed: boolean
  request?: Request
  aborted: boolean
  expectAbort(): void
  release(): void
  wait(): Promise<void>
  done(): Promise<void>
}
export interface Runtime<S> {
  state: S
  app: App
  calls: APICall[]
  unexpected: string[]
  assets: Set<string>
  failImages: boolean
  expectedURLs: Set<string>
  holds: Held[]
  hold(match: Held['match']): Held
  count(path: string): number
  assertQuiet(): void
}
interface Worker<S> { app: App; current: Runtime<S> | null }
const staticReadiness = new WeakMap<Page, () => Promise<void>>()

// Only transport and diagnostics are shared; each domain owns its response plan.
export function runtimeTest<S>(
  initial: () => S,
  allowed: (url: URL) => boolean,
  respond: (url: URL, media: string, body: unknown, state: S) => Reply | Promise<Reply>,
  runtimeOverrides: Record<string, string> = {},
) {
  return base.extend<{ runtime: Runtime<S> }, { runtimeApp: Worker<S> }>({
    // eslint-disable-next-line no-empty-pattern -- Playwright requires destructured fixture arguments.
    runtimeApp: [async ({}, use) => {
      const worker = { current: null } as Worker<S>
      worker.app = await startInsightsFixtureApp(async (url: URL, media: string, body: unknown) => {
        const scene = worker.current
        if (!scene) throw new Error('Request without an active runtime scenario')
        if (!allowed(url)) { scene.unexpected.push(url.href); return { status: 503 } }
        const call: APICall = { url, body, status: 200, completed: false }
        scene.calls.push(call)
        // Resolve the response against this request's scenario before holding it.
        const result = await respond(url, media, body, scene.state)
        call.status = result.status ?? 200
        if (result.status) scene.expectedURLs.add(new URL(url.pathname + url.search, worker.app.base).href)
        collectAssets(result.data, scene.assets, worker.app.upstreamUrl)
        const gate = scene.holds.find(item => !item.received && item.match(url, body))
        if (gate) { gate.received = true; await gate.promise; gate.completed = true }
        call.completed = true
        return result
      }, runtimeOverrides)
      try { await use(worker) } finally { await worker.app.close() }
    }, { scope: 'worker' }],
    runtime: async ({ runtimeApp }, use, testInfo) => {
      const scene: Runtime<S> = {
        app: runtimeApp.app, state: initial(), calls: [], unexpected: [],
        assets: new Set(), failImages: false, expectedURLs: new Set(), holds: [],
        hold(match) {
          let release!: () => void
          const promise = new Promise<void>(resolve => { release = resolve })
          const gate: Held = { match, promise, release, received: false, completed: false, aborted: false,
            expectAbort() { gate.aborted = true },
            async wait() { await expect.poll(() => gate.received).toBe(true) },
            async done() {
              await expect.poll(() => gate.completed).toBe(true)
              if (gate.request && !gate.aborted) {
                const response = await gate.request.response()
                expect(response).not.toBeNull()
                expect(await response!.finished()).toBeNull()
              }
            },
          }
          scene.holds.push(gate); return gate
        },
        count(path) { return scene.calls.filter(call => call.url.pathname.endsWith(path)).length },
        assertQuiet() { expect(scene.unexpected).toEqual([]) },
      }
      runtimeApp.current = scene
      try { await use(scene) } finally {
        for (const gate of scene.holds) gate.release()
        await expect.poll(() => scene.holds.every(gate => !gate.received || gate.completed)).toBe(true)
        if (testInfo.status !== testInfo.expectedStatus) {
          await testInfo.attach('runtime-upstream.json', { body: JSON.stringify({ calls: scene.calls, unexpected: scene.unexpected }), contentType: 'application/json' })
          await testInfo.attach('nitro.log', { body: scene.app.logs(), contentType: 'text/plain' })
        }
        runtimeApp.current = null
      }
    },
    baseURL: async ({ runtime }, use) => { await use(runtime.app.base) },
    page: async ({ page, context, runtime }, use, testInfo) => {
      const external: string[] = [], unexpectedHTTP: string[] = []
      const recoveryProbes: string[] = []
      const injected = new Set<Request>(), received = new Set<Request>()
      const staticRequests = new Set<Request>()
      const failed: Request[] = []
      const raw: Array<{ text: string; url: string }> = []
      const errors = captureBrowserErrors(page, runtime.expectedURLs)
      page.on('console', message => {
        if (message.type() === 'error') raw.push({ text: message.text(), url: message.location().url })
      })
      context.on('requestfinished', request => staticRequests.delete(request))
      context.on('requestfailed', request => { staticRequests.delete(request); failed.push(request) })
      context.on('request', request => {
        const url = new URL(request.url())
        if (url.origin === runtime.app.base && url.pathname.startsWith('/_nuxt/')
          && ['script', 'stylesheet', 'font'].includes(request.resourceType())) staticRequests.add(request)
        if (url.origin === runtime.app.base && url.pathname.startsWith('/api/')) {
          if (!allowed(url)) runtime.unexpected.push('browser ' + request.method() + ' ' + url.href)
          const method = request.postData() || url.pathname.endsWith('/view') ? 'POST' : 'GET'
          if (request.method() !== method) runtime.unexpected.push('method ' + request.method() + ' ' + url.href)
          const gate = runtime.holds.find(item => !item.request && item.match(url, request.postData() ? request.postDataJSON() : undefined))
          if (gate) gate.request = request
        }
      })
      context.on('response', response => {
        if (response.status() < 400) return
        if (response.status() === 503 && runtime.expectedURLs.has(response.url())) received.add(response.request())
        else unexpectedHTTP.push(response.status() + ' ' + response.url())
      })
      await context.addInitScript(({ origin, steamKey, sample }) => {
        if (location.origin !== origin) return
        if (!localStorage.getItem('theme')) localStorage.setItem('theme', 'light')
        const checkedAt = Date.now()
        localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({ selected: 'primary', checkedAt,
          primaryMs: 10, mirrorMs: 20, primaryState: 'success', mirrorState: 'success' }))
        localStorage.setItem(steamKey, JSON.stringify({ version: 1, selected: 'china', checkedAt, sample,
          china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' } }))
      }, { origin: runtime.app.base, steamKey: STEAM_DIAGNOSTICS_KEY, sample: STEAM_PROBE_PATHS[0] })
      await context.route('**/*', route => {
        const request = route.request(), url = new URL(request.url())
        const image = runtime.assets.has(url.href) || (url.origin === runtime.app.base && url.pathname === '/defaultLogo.svg')
        if (runtime.failImages && image) {
          injected.add(request); runtime.expectedURLs.add(url.href)
          return route.abort('failed')
        }
        // An actual renderer failure deliberately invalidates the diagnostic TTL.
        // Only this injected failure may exercise the exact local managed probe.
        if (injected.size > 0 && url.origin === runtime.app.upstreamUrl
          && url.pathname === '/system/probes/cdn.bin' && request.method() === 'GET'
          && [...url.searchParams.keys()].join() === 'probe'
          && /^[a-f0-9-]{36}$/.test(url.searchParams.get('probe') ?? '')) {
          recoveryProbes.push(url.href)
          return route.fulfill({ contentType: 'application/octet-stream', headers: { 'access-control-allow-origin': '*' },
            path: fileURLToPath(new URL('../../fixtures/cdn-probe.bin', import.meta.url)) })
        }
        if (image && url.pathname.startsWith('/nav/patterns/')) {
          return route.fulfill({ contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><path d="M1 1h3v3H1z"/></svg>' })
        }
        if (url.origin === runtime.app.base || image || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
        external.push(url.href)
        return route.abort('blockedbyclient')
      })
      const originalQuiet = runtime.assertQuiet
      staticReadiness.set(page, async () => {
        await page.evaluate(async () => { await document.fonts.ready; await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))) })
        // Wait only for observed production chunks/fonts, not a global network-idle window.
        await expect.poll(() => staticRequests.size).toBe(0)
      })
      runtime.assertQuiet = () => {
        originalQuiet()
        expect(errors).toEqual([]); expect(external).toEqual([]); expect(unexpectedHTTP).toEqual([])
        const cancelled = runtime.holds.filter(gate => gate.aborted).map(gate => {
          expect(gate.request).toBeDefined(); expect(gate.request!.failure()?.errorText).toBe('net::ERR_ABORTED')
          return gate.request!
        })
        expect(new Set(failed)).toEqual(new Set([...injected, ...cancelled]))
        expect(recoveryProbes.length).toBeLessThanOrEqual(4)
        for (const request of injected) expect(request.failure()?.errorText).toBe('net::ERR_FAILED')
        const expected = [...injected, ...received]
        for (const diagnostic of raw) {
          const text = diagnostic.text
          expect(text === 'Failed to load resource: net::ERR_FAILED'
            || text === 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)').toBe(true)
          expect(expected.some(request => request.url() === diagnostic.url)).toBe(true)
        }
        for (const url of new Set(expected.map(request => request.url()))) {
          expect(raw.filter(item => item.url === url)).toHaveLength(expected.filter(request => request.url() === url).length)
        }
      }
      try { await use(page) } finally {
        if (testInfo.status !== testInfo.expectedStatus) {
          await testInfo.attach('runtime-browser.json', { body: JSON.stringify({ errors, external, unexpectedHTTP, raw,
            failed: failed.map(request => ({ url: request.url(), error: request.failure() })) }), contentType: 'application/json' })
        }
      }
    },
  })
}

function collectAssets(value: unknown, assets: Set<string>, upstream: string) {
  if (typeof value === 'string') {
    if (value.startsWith(upstream + '/media/')) assets.add(value)
    else if (/^nav\/sites\/\d+\/icon\/[a-f0-9]{32}\.[a-z]+$/.test(value)) assets.add(upstream + '/' + value)
    else if (/^nav\/patterns\/[a-f0-9]{32}\.svg$/.test(value)) assets.add(upstream + '/' + value)
  } else if (Array.isArray(value)) for (const item of value) collectAssets(item, assets, upstream)
  else if (value && typeof value === 'object') for (const item of Object.values(value)) collectAssets(item, assets, upstream)
}

export async function openRuntime(page: Page, path: string) {
  if (page.url() !== 'about:blank') await settleRuntime(page)
  const response = await page.goto(path, { waitUntil: 'load' })
  expect(response?.status()).toBe(200)
  const html = await response!.text()
  expect(html).toMatch(/<h1\b[^>]*>[\s\S]*?<\/h1>/)
  await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
  await expect(page.locator('main h1').first()).toBeVisible()
  await settleRuntime(page)
  return html
}
export async function settleRuntime(page: Page) { await staticReadiness.get(page)?.() }

// Page-shell contract only; domain specs still assert their own ready content.
export async function assertRuntimeSurface(page: Page, selector: string, theme: 'light' | 'dark') {
  await settleRuntime(page)
  await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
  const root = page.locator(selector)
  await expect(root).toHaveCount(1)
  await expect(root).toBeVisible()
  const box = await root.boundingBox()
  expect(box!.width).toBeGreaterThan(0)
  expect(box!.height).toBeGreaterThan(0)
  expect((await root.innerText()).trim().length).toBeGreaterThan(20)
  expect(await page.evaluate(() => Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - document.documentElement.clientWidth)).toBeLessThanOrEqual(1)
}

export async function revealImages(page: Page, selector = '.insight-entity-media') {
  for (const media of await page.locator(selector).all()) {
    await media.scrollIntoViewIfNeeded()
    await expect.poll(() => media.locator('img').evaluateAll(elements =>
      elements.every(image => (image as HTMLImageElement).complete && (image as HTMLImageElement).naturalWidth > 0))).toBe(true)
  }
}
export async function keyboardFocus(locator: Locator) {
  await locator.page().keyboard.press('Tab')
  await locator.focus()
  await expect(locator).toBeFocused()
  await expect.poll(() => locator.evaluate(el => getComputedStyle(el).outlineStyle)).not.toBe('none')
}
export { expect }
