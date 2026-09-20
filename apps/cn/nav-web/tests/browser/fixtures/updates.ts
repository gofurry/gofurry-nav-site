import { test as base, expect, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import type { NavUpdateNotice, NavUpdatesResponse } from '../../../app/types/nav'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

const notice = (id: number, publishedAt: string, title: string, body: string): NavUpdateNotice => ({
  id, title, body, published_at: publishedAt, create_time: publishedAt, update_time: publishedAt,
})

// Fixed dates/order exercise the real six-item batch, multiline copy and older year.
const fixtureItems = (): NavUpdateNotice[] => [
  notice(109, '2026-09-18T12:34:00Z', '导航首页与资源路由体验完成一轮稳定性调整',
    '重新整理资源加载路径与导航交互，在不改变现有使用方式的前提下，提高弱网环境中的反馈一致性。'),
  notice(108, '2026-08-30T08:05:00Z', '站点观测详情补充更清晰的状态说明',
    '补充可用性、协议状态和最近观测时间的展示，让异常信息更容易定位。'),
  notice(107, '2026-07-14T16:20:00Z', '游戏情报页更新筛选与区域价格说明',
    '筛选条件现在会保持在地址栏中。\n区域价格信息继续按公开数据展示，不进行货币换算。'),
  notice(106, '2026-06-01T03:45:00Z', '导航收录信息完成一批校正',
    '修正部分站点名称、简介与跳转地址，并整理重复标签。'),
  notice(105, '2026-05-12T11:10:00Z', '资源图片加载策略调整',
    '优化主线路与镜像线路之间的选择逻辑，保持已有资源地址兼容。'),
  notice(104, '2026-03-28T09:00:00Z', '移动端导航细节修复',
    '修复窄屏下部分入口拥挤和文字换行问题。'),
  notice(103, '2026-01-15T13:30:00Z', '更新公告页面上线',
    '新增按年份浏览的更新记录，方便查看 GoFurry 持续迭代内容。'),
  notice(102, '2025-12-20T06:00:00Z', '年末数据维护完成',
    '完成一轮站点元数据整理，并修正部分历史观测记录的展示。'),
  notice(101, '2025-10-01T15:15:00Z', '站点观测任务调整',
    '调整部分定时观测任务的调度方式，保持公开页面展示逻辑不变。'),
]

type UpdatesApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type UpdatesOptions = { width?: number; height?: number; theme?: 'light' | 'dark' }
type UpdatesScenario = {
  items: NavUpdateNotice[]
  errors: string[]
  root: Locator
  timeline: Locator
  entries: Locator
  latest: Locator
  loadMore: Locator
  yearControl(year: string): Locator
  yearEntries(year: string): Locator
  open(options?: UpdatesOptions): Promise<{ ssrHTML: string }>
  assertInitial(): Promise<void>
  assertQuiet(): void
  updatesCalls(): URL[]
}

export const test = base.extend<{ updates: UpdatesScenario }, { updatesApp: UpdatesApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  updatesApp: [async ({}, use) => {
    const app = await startInsightsFixtureApp(url => {
      if (url.pathname !== '/api/v2/nav/updates' || url.searchParams.get('lang') !== 'zh') return { status: 500 }
      const data: NavUpdatesResponse = {
        schema_version: 1, state: 'ready', generated_at: '2026-09-18T12:40:00Z', items: fixtureItems(),
      }
      return { data }
    })
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ updatesApp }, use) => { await use(updatesApp.base) },
  updates: async ({ page, context, updatesApp }, use, testInfo) => {
    const items = fixtureItems()
    const errors = captureBrowserErrors(page)
    const external: string[] = [], failed: string[] = [], clientUpdates: string[] = []
    const upstreamStart = updatesApp.requests.length
    const upstreamCalls = () => updatesApp.requests.slice(upstreamStart)
    const updatesCalls = () => upstreamCalls().filter(url => url.pathname === '/api/v2/nav/updates')
    page.on('requestfailed', request => failed.push(`${request.url()}: ${request.failure()?.errorText}`))
    page.on('request', request => {
      if (new URL(request.url()).pathname === '/api/v2/nav/updates') clientUpdates.push(request.url())
    })
    await context.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === updatesApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      await route.abort('blockedbyclient')
    })

    const root = page.locator('.updates-page')
    const timeline = root.locator('.updates-timeline-section')
    const entries = timeline.getByRole('article')
    const latest = timeline.getByText('最新', { exact: true })
    const loadMore = timeline.getByRole('button', { name: '加载更多', exact: true })
    const yearControl = (year: string) => timeline.getByRole('button').filter({ hasText: year })
    const yearEntries = (year: string) => timeline.getByRole('listitem')
      .filter({ has: page.getByRole('button').filter({ hasText: year }) }).getByRole('article')
    const assertQuiet = () => {
      expect(updatesCalls(), 'Exactly one real SSR Updates request').toHaveLength(1)
      expect(updatesCalls()[0]!.searchParams.toString()).toBe('lang=zh')
      expect(upstreamCalls(), 'No unrelated business API requests').toHaveLength(1)
      expect(clientUpdates, 'Hydration must reuse the SSR payload').toEqual([])
      expect(external, 'Updates must not depend on external resources').toEqual([])
      expect(failed, 'No failed resource requests are expected').toEqual([])
      expect(errors, 'SSR/hydration/rendering must have no browser errors').toEqual([])
    }
    const assertInitial = async () => {
      await expect(root).toHaveCount(1)
      await expect(root).toBeVisible()
      await expect(timeline).toHaveAttribute('aria-busy', 'false')
      const summary = root.getByRole('region', { name: '更新公告概览', exact: true })
      await expect(summary).toContainText('公告数量')
      await expect(summary.getByText('9', { exact: true })).toBeVisible()
      await expect(yearControl('2026')).toBeVisible()
      await expect(yearControl('2025')).toBeVisible()
      // Production has no aria-expanded; actual year content proves expansion.
      await expect(yearEntries('2026')).toHaveCount(6)
      await expect(yearEntries('2025')).toHaveCount(0)
      await expect(entries).toHaveCount(6)
      await expect(latest).toHaveCount(1)
      await expect(latest).toBeVisible()
      await expect(loadMore).toBeVisible()
      assertQuiet()
    }
    try {
      await use({
        items, errors, root, timeline, entries, latest, loadMore, yearControl, yearEntries,
        assertInitial, assertQuiet, updatesCalls,
        async open({ width = 1440, height = 900, theme = 'light' } = {}) {
          await page.setViewportSize({ width, height })
          await context.addInitScript(({ origin, theme, steamDiagnosticsKey, steamSample }) => {
            if (location.origin !== origin) return
            localStorage.setItem('theme', theme)
            // Only routing history is fresh. Notice timestamps above stay fixed;
            // real plugins use their TTL, with no explicit route pin or mock.
            const checkedAt = Date.now()
            localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({
              selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20,
              primaryState: 'success', mirrorState: 'success',
            }))
            localStorage.setItem(steamDiagnosticsKey, JSON.stringify({
              version: 1, selected: 'china', checkedAt, sample: steamSample,
              china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
            }))
          }, { origin: updatesApp.base, theme, steamDiagnosticsKey: STEAM_DIAGNOSTICS_KEY, steamSample: STEAM_PROBE_PATHS[0] })
          const response = await page.goto('/updates', { waitUntil: 'load' })
          expect(response?.status()).toBe(200)
          const ssrHTML = await response!.text()
          expect(ssrHTML).toContain(items[0]!.title)
          await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
          await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
          await assertInitial()
          return { ssrHTML }
        },
      })
      assertQuiet()
    } finally {
      await testInfo.attach('updates-requests.json', {
        body: JSON.stringify({ upstream: upstreamCalls().map(url => url.href), clientUpdates, external, failed, errors }),
        contentType: 'application/json',
      })
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('updates-app.log', { body: updatesApp.logs(), contentType: 'text/plain' })
      }
      // No gates/tasks: Playwright owns test-scoped context and route disposal.
    }
  },
})

export { expect }
