import { finiteChartValue, formatChartNumber, formatChartPercent, formatGameAverage } from '../../../app/utils/detailChartValues'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) test(`Task C entry, Current Target signals and shared material ${width}-${theme}`, async ({ page, context, runtime }) => {
  Object.assign(runtime.state, { cdnScenario: 'reliable', protocolScenario: 'slow', observationRich: true, historyCount: 100, viewCount: 20457 })
  await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
  const prefix = width === 390 ? '' : '/en', success = prefix ? 'Success' : '成功', stale = prefix ? 'Stale' : '已过期'
  const path = prefix + '/site/41' + (width === 390 ? '?domain=target.example' : '')
  await open(page, path)
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await expect(page).toHaveURL(path)
  await expect(page.locator('[data-site-cdn]')).toHaveText('Cloudflare CDN')
  await expect(page.locator('[data-site-hero]')).not.toContainText(/confidence|medium|high|hint_type/)
  const trigger = page.locator('[data-site-target-trigger]')
  await expect(trigger).not.toHaveClass(/gf-button/)
  await expect(trigger).toHaveCSS('border-width', '0px')
  await expect(trigger).toHaveCSS('transition-duration', '0.5s, 0.5s'); await expect(trigger).toHaveCSS('transform', 'none')
  await expect(page.locator('[data-site-views]')).toHaveText('20,457')
  await assertAcrylic(page); await assertRuntimeSurface(page, '[data-site-detail]', theme)
  await review(page, `task-c-entry-${width}-${theme}`)
  await page.locator('[data-site-observation-tab="overview"]').click()
  for (const owner of ['[data-site-target-context]', '[data-site-observation-overview]']) {
    await expect(page.locator(owner + ' .site-detail-status-dot')).toHaveCount(3)
    await expect(page.locator(owner + ' [data-site-protocol-status].sr-only')).toHaveText([success, success])
    await expect(page.locator(owner + ' [data-site-protocol-status]:not(.sr-only)')).toHaveText('· ' + stale)
    await expect(page.locator(owner + ' .site-detail-status-dot[data-tone="good"]')).toHaveCount(2)
    await expect(page.locator(owner + ' .site-detail-status-dot[data-tone="warning"]')).toHaveCount(1)
  }
  await expect(page.locator('[data-site-protocol="ping"] dd')).toHaveAttribute('data-tone', 'bad')
  await expect(page.locator('[data-site-observation-protocol="ping"] dd').first()).toHaveAttribute('data-tone', 'bad')
  await review(page, `task-c-protocols-${width}-${theme}`)
  await trigger.hover()
  await expect.poll(() => trigger.evaluate(node => node.getAnimations().filter(animation => animation.playState === 'running').length)).toBe(0)
  await review(page, `task-c-hover-${width}-${theme}`)
  await trigger.click(); await expect(trigger).toHaveCSS('transition-duration', '0s')
  await page.keyboard.press('End'); await page.keyboard.press('Enter')
  await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
  await expect(page.locator('[data-site-cdn]')).toHaveText('Fastly CDN')
  await expect(trigger).toBeFocused()
  expect(runtime.count('/sites/41/detail')).toBe(2); expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
  expect(runtime.count('/observations')).toBe(1); expect(runtime.count('/trend')).toBe(0); expect(runtime.calls).toHaveLength(6)
  await assertRuntimeSurface(page, '[data-site-detail]', theme); runtime.assertQuiet()
})

for (const scenario of ['none', 'unreliable'] as const) test('Task C omits unreliable CDN badge: ' + scenario, async ({ page, runtime }) => {
  runtime.state.cdnScenario = scenario
  await open(page, '/en/site/41?tab=overview')
  await expect(page.locator('[data-site-cdn]')).toHaveCount(0)
  expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
})

for (const scenario of ['failure', 'missing'] as const) test('Task C protocol exceptions remain visible: ' + scenario, async ({ page, runtime }) => {
  runtime.state.protocolScenario = scenario
  await open(page, '/en/site/41?tab=observation&view=overview')
  for (const owner of ['[data-site-target-context]', '[data-site-observation-overview]']) {
    await expect(page.locator(owner + ' [data-site-protocol-status].sr-only')).toHaveCount(0)
    await expect(page.locator(owner + ' [data-site-protocol-status]')).toHaveText(Array(3).fill('· ' + (scenario === 'failure' ? 'Failed' : 'Unknown')))
  }
  expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
})

// Optional review artifacts use the Functional Browser owner, never Visual goldens.
async function review(page: Page, name: string) {
  const directory = process.env.GOFURRY_SITE_DETAIL_REVIEW_DIR
  if (!directory) return
  await mkdir(directory, { recursive: true })
  await settleRuntime(page)
  await page.evaluate(() => window.scrollTo(0, 0))
  await page.screenshot({ path: join(directory, name + '.png'), fullPage: true })
}
async function open(page: Page, path = '/en/site/41?tab=overview') {
  const view = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  await openRuntime(page, path)
  await (await view).finished()
}

async function assertAcrylic(page: Page) {
  const panels = '[data-site-identity-note], [data-site-target-context], .site-detail-surface, [data-site-overview-composite], .site-observation-composite, .site-observation-evidence'
  // Theme hydration can start the same color transition as hover. Observe its end,
  // without fixed sleep or a hard-coded color snapshot.
  await page.mouse.move(0, 0)
  await expect.poll(() => page.locator(panels).evaluateAll(nodes => new Set(nodes.map(node => getComputedStyle(node).backgroundColor)).size)).toBe(1)
  const material = await page.locator('[data-site-detail]').evaluate(root => {
    const css = (selector: string) => getComputedStyle(root.querySelector(selector)!)
    const alpha = (value: string) => Number(value.match(/\/\s*([\d.]+)\)$/)?.[1] ?? value.match(/rgba\([^)]*,\s*([\d.]+)\)$/)?.[1] ?? 1)
    const a = css('[data-site-identity-note]'), b = css('[data-site-target-context]')
    const borders = Array.from(root.querySelectorAll('[data-site-identity-note], [data-site-hero], [data-site-health-strip], [data-site-health], [data-site-target-context], .site-detail-surface, [data-site-overview-composite], .site-observation-composite, .site-observation-evidence, [data-site-workspace-subnav-header] [role="tablist"], .site-observation-facts > div, .site-observation-evidence-list > div, .site-security-row, .site-intelligence-row, .site-intelligence-chart-shell')).map(node => {
      const s = getComputedStyle(node)
      return { hook: node.tagName + ' ' + node.className, widths: [s.borderTopWidth, s.borderRightWidth, s.borderBottomWidth, s.borderLeftWidth], shadow: s.boxShadow }
    })
    const panels = Array.from(root.querySelectorAll('[data-site-identity-note], [data-site-target-context], .site-detail-surface, [data-site-overview-composite], .site-observation-composite, .site-observation-evidence')).map(node => ({ fill: getComputedStyle(node).backgroundColor, image: getComputedStyle(node).backgroundImage }))
    return { a: alpha(a.backgroundColor), b: alpha(b.backgroundColor), borders, panels }
  })
  expect(material.a).toBeLessThan(0.7); expect(material.b).toBe(material.a); expect(material.b).toBeGreaterThan(0)
  expect(new Set(material.panels.map(panel => panel.fill)).size).toBe(1)
  expect(material.panels.every(panel => panel.image === 'none')).toBe(true)
  for (const item of material.borders) {
    expect(item.widths, item.hook).toEqual(['0px', '0px', '0px', '0px'])
    expect(item.shadow, item.hook).toBe('none')
  }
}

async function assertTitleNav(page: Page, domain: 'observation' | 'security') {
  const header = page.locator(`[data-site-${domain}] > [data-site-workspace-subnav-header]`)
  await expect(header.locator('h2')).toHaveCount(1)
  await expect(header.locator('[role="tablist"]')).toHaveCount(1)
  await expect(page.locator('[data-site-workspace] > h2')).toHaveCount(0)
  await expect(header.locator('h2')).toBeVisible()
  await expect(header.locator('[role="tablist"]')).toBeVisible()
  await expect(header.locator('[role="tablist"]')).toHaveCSS('overflow-x', 'auto')
  await expect(header.locator('[role="tablist"]')).toHaveCSS('scrollbar-width', 'none')
}

async function assertCapabilityStates(page: Page, owner: string) {
  const section = page.locator(owner)
  await expect(section.locator('[data-site-capability]')).toHaveCount(7)
  for (const [key, state] of [['ipv6', 'supported'], ['http2', 'unsupported'], ['tls13', 'stale'], ['certificate_verified', 'not_probed'], ['hsts', 'unavailable'], ['csp', 'unknown'], ['security_txt', 'not_applicable']]) {
    await expect(section.locator(`[data-site-capability="${key}"]`)).toHaveAttribute('data-site-capability-state', state!)
  }
}

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) {
  const suffix = `${width}-${theme}`
  test(`final shell, Overview and Observation presentation ${suffix}`, async ({ page, context, runtime }) => {
    runtime.state.observationRich = true; runtime.state.fullCapabilities = true; runtime.state.manyChanges = true; runtime.state.historyCount = 100
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await open(page)
    await expect(page.locator('[data-site-identity-note] [data-site-hero]')).toHaveCount(1)
    await expect(page.locator('[data-site-identity-note] [data-site-health-strip]')).toHaveCount(1)
    await expect(page.locator('[data-site-identity-note] [aria-hidden="true"] > span')).toHaveCount(3)
    await assertAcrylic(page)
    await assertCapabilityStates(page, '[data-site-capability-snapshot]')
    await expect(page.locator('[data-site-health]')).toHaveCount(6)
    await expect(page.locator('[data-site-health] .site-detail-note')).toHaveCount(0)
    await expect(page.locator('[data-site-health="certificate"] .site-detail-health__value')).toHaveText('Not observed')
    await expect(page.locator('[data-site-target-trigger]')).toContainText('target.example')
    await expect(page.locator('[data-site-target-context]')).not.toContainText(/shared_canonical|same_host|redirect_to_external|confidence|cloudflare/)
    await expect(page.locator('[data-site-hero] .site-detail-hero__description')).toBeVisible()
    await expect(page.locator('[data-site-hero] .site-detail-hero__meta')).toBeVisible()
    await expect(page.locator('[data-site-hero] .site-detail-hero__icon')).toBeVisible()
    await expect(page.locator('[data-site-capability-snapshot] [data-site-overview-insights]')).toBeVisible()
    await expect(page.locator('[data-site-change][data-site-change-category]')).toHaveCount(4)
    await taskDChanges(page, '[data-site-recent-changes]', 4, width >= 768)
    await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `overview-${suffix}`)
    await page.locator('[data-site-primary-tab="observation"]').click()
    await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
    await page.locator('[data-site-observation-tab="overview"]').click()
    await expect(page.locator('[data-site-observation]')).not.toContainText('Current Target ·')
    await assertTitleNav(page, 'observation')
    await assertAcrylic(page)
    await expect(page.locator('[data-site-observation-nav]')).toHaveCSS('border-width', '0px')
    await expect(page.locator('[data-site-observation-protocol]')).toHaveCount(3)
    await expect(page.locator('[data-site-observation-risks]')).toContainText('DNS observations are missing or stale')
    await expect(page.locator('[data-site-observation-risks]')).not.toContainText('后端旧中文')
    await review(page, `observation-overview-${suffix}`)
    for (const view of ['performance', 'http', 'dns', 'web']) {
      await page.locator(`[data-site-observation-tab="${view}"]`).click()
      if (view === 'performance') {
        await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
        await expect(page.locator('[data-site-performance] [data-site-performance-sample="20"]')).toHaveAttribute('aria-pressed', 'true')
      }
      if (view === 'http') {
        const redirects = page.locator('[data-site-redirect-node]')
        await expect(redirects).toHaveCount(2)
        await expect(redirects.first()).not.toContainText('1.')
        await expect(redirects.locator('[data-direction="right"][aria-hidden="true"]')).toHaveCount(1)
        await expect(redirects.last().locator('[data-direction]')).toHaveCount(0)
      }
      if (view === 'dns') {
        await expect(page.locator('[data-site-dns-risks] h3')).toHaveText('DNS Observation Signals')
        for (const code of ['ptr_empty', 'low_ttl', 'collector_reported_signal']) await expect(page.locator(`[data-site-dns-signal="${code}"] [data-tone]`)).toHaveAttribute('data-tone', 'neutral')
        for (const code of ['private_ip', 'nxdomain_with_answer']) await expect(page.locator(`[data-site-dns-signal="${code}"] p[data-tone]`)).toHaveAttribute('data-tone', 'warning')
      }
      if (view === 'web') {
        for (const probe of await page.locator('[data-site-web-probe]').all()) {
          expect(await probe.locator(':scope > dl > div').count()).toBeLessThanOrEqual(5)
          await expect(probe.locator('details')).not.toHaveAttribute('open')
        }
      }
      await assertAcrylic(page)
      await assertTitleNav(page, 'observation')
      await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `observation-${view}-${suffix}`)
    }
    expect(runtime.count('/sites/41/detail')).toBe(1); expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.count('/observations')).toBe(1); expect(runtime.count('/trend')).toBe(0); expect(runtime.calls).toHaveLength(5)
    runtime.assertQuiet()
  })

  test(`final Security evidence presentation ${suffix}`, async ({ page, context, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, days: 12, txt: 'issues', long: true })
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await open(page, '/en/site/41?tab=security')
    await assertTitleNav(page, 'security'); await assertAcrylic(page)
    await expect(page.locator('[data-site-security-summary-plane] [data-site-security-summary]')).toHaveCount(6)
    await expect(page.locator('[data-site-security-summary="transport"]')).toContainText('TLS 1.3')
    await expect(page.locator('[data-site-security-scope]')).toHaveCount(0)
    await expect(page.locator('[data-site-security]')).not.toContainText('Current Target ·')
    await expect(page.locator('[data-site-security-summary="headers"]')).toContainText('6 / 6 observed')
    await expect(page.locator('[data-site-security-overview]')).not.toContainText(/X-Frame-Options|Content-Security-Policy|Security Score/)
    await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `security-overview-${suffix}`)
    for (const view of ['tls', 'web', 'exposure']) {
      await page.locator(`[data-site-security-tab="${view}"]`).click()
      await expect(page.locator('[data-site-security-scope]')).toHaveCount(0)
      if (view === 'tls') {
        await expect(page.locator('[data-site-security-transport-composite] [data-site-tls-state]')).toHaveText('TLS 1.3')
        await expect(page.locator('[data-site-security-transport-composite] [data-site-certificate-verification]')).toHaveCount(1)
        await expect(page.locator('[data-site-security-transport-composite] [data-site-certificate-expiry]')).toHaveCount(1)
        await expect(page.locator('[data-site-certificate-expiry]')).toHaveAttribute('data-tone', 'warning')
        await expect(page.locator('[data-site-health="certificate"] .site-detail-health__value')).toHaveAttribute('data-tone', 'warning')
        await page.locator('[data-site-certificate-crypto] summary').click()
        await expect(page.locator('[data-site-certificate-crypto] [data-site-evidence="cert_fingerprint_sha256"]')).toBeVisible()
      }
      if (view === 'web') {
        await expect(page.locator('[data-site-security-header]')).toHaveCount(6)
        await expect(page.locator('[data-site-security-header][data-state="present"]')).toHaveCount(6)
        await expect(page.locator('[data-site-security-txt-details]')).not.toHaveAttribute('open')
        await expect(page.locator('[data-site-security-txt-validation]')).toContainText('Contact information is missing or invalid')
        await expect(page.locator('[data-site-security-txt-validation]')).not.toContainText('contact_missing_or_invalid')
        await expect(page.locator('[data-site-security-headers] > div').first().locator('[data-site-security-raw-headers]')).toBeVisible()
      }
      if (view === 'exposure') {
        await expect(page.locator('[data-site-port-check] h3')).toHaveText('Port Observation')
        await expect(page.locator('[data-site-waf-canary] h3')).toHaveText('Request Behavior Check')
        await expect(page.locator('[data-site-port-result]')).toHaveCount(5)
        await expect(page.locator('[data-site-port-result][data-state="open"] dd[data-tone]')).toHaveAttribute('data-tone', 'neutral')
        await expect(page.locator('[data-site-waf-canary]')).toContainText('1 / 1 expected behaviors matched')
        await expect(page.locator('[data-site-waf-canary]')).not.toContainText(/WAF Enabled|WAF Detected|Protected by WAF/)
      }
      await assertAcrylic(page); await assertTitleNav(page, 'security')
      await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `security-${view}-${suffix}`)
    }
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })

  test(`final Site intelligence analysis presentation ${suffix}`, async ({ page, context, runtime }) => {
    runtime.state.fullCapabilities = true; runtime.state.intelligence.rich = true
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    await open(page)
    expect(runtime.count('/trend')).toBe(0)
    await page.locator('[data-site-primary-tab="insights"]').click()
    await expect(page.locator('[data-site-insight-trend]')).toHaveAttribute('data-site-insight-trend-state', 'ready')
    await expect(page.locator('[data-site-insights-workspace] > .site-detail-surface')).toHaveCount(1)
    await expect(page.locator('[data-site-insights-header] h2')).toHaveCount(1)
    await expect(page.locator('[data-site-workspace] > h2')).toHaveCount(0)
    await assertAcrylic(page)
    await assertCapabilityStates(page, '[data-site-capability-matrix]')
    await expect(page.locator('[data-site-capability="ipv6"]')).toHaveAttribute('aria-pressed', 'true')
    await expect(page.locator('[data-site-capability-selected="true"]')).toHaveCount(1)
    await expect(page.locator('[data-site-insight-group] > h4')).toHaveText(['Network', 'Transport', 'Web policy'])
    await expect(page.locator('[data-site-capability-matrix] [data-site-capability]')).toHaveCount(7)
    await expect(page.locator('[data-site-capability="tls13"]')).not.toContainText('Transport')
    await expect(page.locator('[data-site-capability-detail] [data-site-insight-trend]')).toHaveCount(1)
    await expect(page.locator('.site-detail-segmented [data-site-insight-range="30d"]')).toHaveAttribute('aria-pressed', 'true')
    const neutral = await page.locator('[data-site-capability-adoption], [data-site-capability-coverage]').evaluateAll(nodes => nodes.map(node => getComputedStyle(node).color))
    expect(new Set(neutral).size).toBe(1)
    const chartColor = await page.locator('[data-site-insight-chart]').evaluate(node => {
      const style = getComputedStyle(node), marker = document.createElement('span')
      node.appendChild(marker); marker.style.color = 'var(--site-detail-info)'
      const info = getComputedStyle(marker).color; marker.style.color = 'var(--site-detail-positive)'
      const positive = getComputedStyle(marker).color; marker.remove()
      return { actual: style.color, info, positive }
    })
    expect(chartColor.actual).toBe(chartColor.info); expect(chartColor.actual).not.toBe(chartColor.positive)
    await expect(page.locator('[data-site-insight-change]')).toHaveCount(6)
    await taskDChanges(page, '[data-site-insight-changes]', 6, width >= 768)
    await expect(page.locator('[data-site-insights-workspace]')).not.toContainText(/score|ranking|percentile/i)
    await assertRuntimeSurface(page, '[data-site-detail]', theme); await review(page, `insights-${suffix}`)
    expect(runtime.count('/trend')).toBe(1); expect(runtime.calls).toHaveLength(5); runtime.assertQuiet()
  })
}

for (const locale of ['en', 'zh']) test('localized multi-target Attention ' + locale, async ({ page, runtime }) => {
  runtime.state.summaryScenario = 'mixed'; runtime.state.extraTargets = ['third.example']
  await open(page, locale === 'en' ? '/en/site/41?tab=overview' : '/site/41?tab=overview')
  await expect(page.locator('[data-site-overview-attention] li')).toHaveText([locale === 'en' ? '2 targets · HTTP is currently unreachable' : '2 个目标 · HTTP 当前无法访问'])
  await expect(page.locator('[data-site-overview-attention]')).not.toContainText('后端旧中文')
  await review(page, `overview-attention-${locale}`)
  expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
})

for (const locale of ['en', 'zh']) test('first-round finite title nav and localized request evidence ' + locale, async ({ page, runtime }) => {
  runtime.state.security.enabled = true
  await page.setViewportSize({ width: 390, height: 844 })
  await open(page, (locale === 'en' ? '/en' : '') + '/site/41?tab=security')
  const nav = page.locator('[data-site-security-nav]')
  await assertTitleNav(page, 'security')
  // Short Chinese labels fit at 390px; narrower space exercises actual overflow in both locales.
  await page.setViewportSize({ width: 320, height: 844 })
  await assertTitleNav(page, 'security')
  expect(await nav.evaluate(node => node.scrollWidth > node.clientWidth)).toBe(true)
  await page.locator('[data-site-security-tab="overview"]').focus()
  await page.keyboard.press('End')
  const exposure = page.locator('[data-site-security-tab="exposure"]')
  await expect(exposure).toHaveAttribute('aria-selected', 'true'); await expect(exposure).toBeFocused()
  await expect(exposure).toHaveCSS('transition-duration', '0s')
  expect(await nav.evaluate(node => node.scrollLeft)).toBeGreaterThan(0)
  await expect(page.locator('[data-site-waf-canary] h3')).toHaveText(locale === 'en' ? 'Request Behavior Check' : '请求行为校验')
  await expect(page.locator('[data-site-port-check] h3')).toHaveText(locale === 'en' ? 'Port Observation' : '端口观测')
  await page.keyboard.press('Home'); await expect(page.locator('[data-site-security-tab="overview"]')).toBeFocused()
  await page.locator('[data-site-primary-tab="observation"]').click()
  await assertTitleNav(page, 'observation')
  await expect(page.locator('[data-site-performance-history-state]')).toHaveAttribute('data-site-performance-history-state', 'ready')
  await page.locator('[data-site-observation-tab="overview"]').focus(); await page.keyboard.press('End')
  await expect(page.locator('[data-site-observation-tab="web"]')).toHaveAttribute('aria-selected', 'true')
  await expect(page.locator('[data-site-observation-tab="web"]')).toBeFocused()
  await assertAcrylic(page); await assertRuntimeSurface(page, '[data-site-detail]', 'light')
  expect(runtime.count('/observations')).toBe(1); expect(runtime.calls).toHaveLength(5); runtime.assertQuiet()
})

test('first-round unavailable Insights retains its borderless explorer and shared inline retry', async ({ page, runtime }) => {
  runtime.state.siteInsightsFailure = true
  await open(page, '/en/site/41?tab=insights')
  await expect(page.locator('[data-site-insights-state]')).toHaveAttribute('data-site-insights-state', 'unavailable')
  await expect(page.locator('[data-site-capability-matrix] [data-site-capability]')).toHaveCount(7)
  await expect(page.locator('[data-site-insights-header] [data-site-insights-retry]')).toBeVisible()
  await expect(page.locator('[data-site-insight-trend]')).toHaveAttribute('data-site-insight-trend-state', 'ready')
  await assertAcrylic(page)
  const before = { detail: runtime.count('/sites/41/detail'), insights: runtime.count('/sites/41/insights'), view: runtime.count('/sites/41/view'), trend: runtime.count('/trend') }
  runtime.state.siteInsightsFailure = false
  await page.locator('[data-site-insights-retry]').click()
  await expect(page.locator('[data-site-insights-state]')).toHaveAttribute('data-site-insights-state', 'ready')
  expect(runtime.count('/sites/41/insights')).toBe(before.insights + 1)
  expect(runtime.count('/sites/41/detail')).toBe(before.detail); expect(runtime.count('/sites/41/view')).toBe(before.view); expect(runtime.count('/trend')).toBe(before.trend)
  runtime.assertQuiet()
})


test('Task D chart formatters preserve missing values and average precision', () => {
  for (const value of [null, undefined, '4', NaN, Infinity, -Infinity, {}, []]) {
    expect(finiteChartValue(value)).toBeNull()
    expect(formatChartPercent(value)).toBe('—')
    expect(formatChartNumber(value, 'en')).toBe('—')
    expect(formatGameAverage(value, 'en')).toBe('—')
  }
  expect(formatChartPercent(0)).toBe('0%'); expect(formatChartNumber(0, 'en')).toBe('0')
  expect([35386.394, 42.5, 42].map(value => formatGameAverage(value, 'en'))).toEqual(['35,386.4', '42.5', '42'])
})

async function taskDTooltip(page: Page) {
  const trigger = page.locator('[data-site-help-trigger]').first(), tip = page.getByRole('tooltip')
  await expect(trigger).not.toHaveAttribute('title')
  await trigger.hover(); await expect(tip).toBeVisible()
  await expect(trigger).toHaveAttribute('aria-describedby', await tip.getAttribute('id') as string)
  const bounds = await tip.boundingBox(); expect(bounds!.x).toBeGreaterThanOrEqual(0)
  expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(page.viewportSize()!.width)
  await page.keyboard.press('Escape'); await expect(tip).toHaveCount(0)
  await page.mouse.move(0, 0); await trigger.focus(); await expect(tip).toBeVisible()
  await trigger.evaluate(node => (node as HTMLElement).blur()); await expect(tip).toHaveCount(0)
  await trigger.click(); await expect(tip).toBeVisible()
  await page.keyboard.press('Escape'); await expect(tip).toHaveCount(0)
}

async function taskDChanges(page: Page, selector: string, count: number, desktop = true) {
  const section = page.locator(selector), rows = section.locator('[data-site-change-stream] > li')
  await expect(rows).toHaveCount(count)
  await expect(rows.locator('time')).toHaveCount(count)
  for (const row of await rows.all()) {
    await expect(row).toHaveAttribute('data-site-change-category', /^(capability|target|certificate|unknown)$/)
    await expect(row.locator('time')).toHaveAttribute('datetime', /.+/)
    await expect(row.locator('time')).toHaveAttribute('data-precise', /^(true|false)$/)
  }
  const dates = await rows.locator('time').allTextContents()
  if (desktop) {
    await expect(section.locator('[data-site-change-mode="compact"]')).toHaveAttribute('aria-pressed', 'true')
    await section.locator('[data-site-change-mode="list"]').click()
    await expect(section.locator('[data-site-change-stream]')).toHaveAttribute('data-mode', 'list')
    await expect(section.locator('[data-site-change-mode="list"]')).toHaveAttribute('aria-pressed', 'true')
    await expect(rows.locator('[data-direction]')).toHaveCount(0)
    expect(await rows.locator('time').allTextContents()).toEqual(dates)
    await section.locator('[data-site-change-mode="compact"]').click()
    await expect(section.locator('[data-site-change-stream]')).toHaveAttribute('data-mode', 'compact')
    await expect(rows.locator('[data-direction][aria-hidden="true"]')).toHaveCount(count - 1)
    await expect(rows.last().locator('[data-direction]')).toHaveCount(0)
    expect(await rows.locator('time').allTextContents()).toEqual(dates)
  } else {
    await expect(section.locator('[data-site-change-modes]')).toBeHidden()
  }
}

async function taskDTrendGap(page: Page) {
  const chart = page.locator('[data-site-insight-chart]')
  // ECharts richText tooltips paint on canvas; observe actual rendered text, not an HTML substitute.
  await chart.evaluate(node => {
    const original = CanvasRenderingContext2D.prototype.fillText
    const capture = { text: [] as string[], restore: () => { CanvasRenderingContext2D.prototype.fillText = original } }
    ;(window as typeof window & { taskDCanvas?: typeof capture }).taskDCanvas = capture
    CanvasRenderingContext2D.prototype.fillText = function (text, x, y, maxWidth) {
      if (node.contains(this.canvas)) capture.text.push(text)
      if (maxWidth === undefined) original.call(this, text, x, y)
      else original.call(this, text, x, y, maxWidth)
    }
  })
  try {
    const width = (await chart.boundingBox())!.width
    await chart.hover({ position: { x: (width + 32) / 2, y: 100 } })
    const text = () => page.evaluate(() => (window as typeof window & { taskDCanvas: { text: string[] } }).taskDCanvas.text)
    await expect.poll(text).toContain('2026-09-02')
    await expect.poll(text).toContain('—')
    expect((await text()).join(' ')).not.toMatch(/undefined|null|NaN|Infinity/)
  } finally {
    await page.evaluate(() => {
      const target = window as typeof window & { taskDCanvas?: { restore: () => void } }
      target.taskDCanvas?.restore(); delete target.taskDCanvas
    })
    await page.mouse.move(0, 0)
  }
}

test('Task D Overview and Observation composites retain the default/history contract', async ({ page, runtime }) => {
  Object.assign(runtime.state, { observationRich: true, fullCapabilities: true, manyChanges: true, historyCount: 100, redirectCount: 7, cdnScenario: 'reliable' })
  await open(page)
  const composite = page.locator('[data-site-overview-composite]')
  await expect(composite.locator('[data-site-overview-health], [data-site-capability-snapshot]')).toHaveCount(2)
  await expect(composite.locator('[data-site-overview-attention]')).toHaveCount(0)
  await expect(composite.locator('[data-site-status-distribution]')).toHaveCount(0)
  expect(await composite.locator('[data-site-overview-status-grid] > *').evaluateAll(nodes => nodes.map(node => node.children.length))).toEqual([2, 2, 2])
  await expect(composite).not.toContainText('Site-wide')
  await expect(page.locator('[data-site-hero-heading] .site-detail-hero__domain')).toHaveText('target.example')
  await expect(page.locator('[data-site-hero-meta] [data-site-views]')).toHaveCount(1)
  await expect(page.locator('[data-site-hero-meta] [data-site-cdn]')).toHaveText('Cloudflare CDN')
  await taskDChanges(page, '[data-site-recent-changes]', 4)
  expect(runtime.calls).toHaveLength(4)
  await review(page, 'task-d-overview-1440-light')
  await page.locator('[data-site-primary-tab="observation"]').click()
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await expect(page.locator('[data-site-performance-waterfall] h4, [data-site-performance-waterfall] [data-site-help-trigger]')).toHaveCount(0)
  await expect(page.locator('[data-site-performance-history-state]')).toHaveAttribute('data-site-performance-points', '20')
  for (const sample of [60, 100, 20]) {
    await page.locator(`[data-site-performance-sample="${sample}"]`).click()
    await expect(page.locator('[data-site-performance-history-state]')).toHaveAttribute('data-site-performance-points', String(sample))
  }
  await page.locator('[data-site-history-table] summary').click()
  await expect(page.locator('[data-site-history-table] tbody tr')).toHaveCount(20)
  expect(runtime.count('/observations')).toBe(1)
  expect(Object.fromEntries(runtime.calls.find(call => call.url.pathname.endsWith('/observations'))!.url.searchParams)).toEqual({ protocol: 'ping', limit: '100', payload_mode: 'preview' })
  await review(page, 'task-d-performance-1440-light')
  await page.locator('[data-site-observation-tab="overview"]').click()
  const observation = page.locator('[data-site-observation-overview-composite]')
  await expect(observation.locator('[data-site-observation-current], [data-site-observation-endpoint], [data-site-observation-risks]')).toHaveCount(3)
  await review(page, 'task-d-observation-overview-1440-light')
  await page.locator('[data-site-observation-tab="http"]').click()
  await expect(page.locator('[data-site-http-composite] [data-site-http-response], [data-site-http-composite] [data-site-http-redirects], [data-site-http-composite] [data-site-http-headers]')).toHaveCount(3)
  const redirects = page.locator('[data-site-redirect-node]')
  await expect(redirects).toHaveCount(7)
  expect(await redirects.locator('[data-direction]').evaluateAll(nodes => nodes.map(node => node.getAttribute('data-direction')))).toEqual(['right', 'right', 'down', 'left', 'left', 'down'])
  await expect(redirects.last().locator('[data-direction]')).toHaveCount(0)
  await expect(page.locator('[data-site-http-redirects]')).not.toContainText(/301|302/)
  await page.locator('[data-site-http-all-headers] summary').click()
  await expect(page.locator('[data-site-http-all-headers] dl')).toBeVisible()
  await review(page, 'task-d-http-1440-light')
  await page.locator('[data-site-observation-tab="dns"]').click()
  await expect(page.locator('[data-site-dns-ledger] [data-site-dns-risks]')).toHaveCount(1)
  await expect(page.locator('[data-site-dns-signal="ptr_empty"]')).not.toContainText('ptr_empty')
  await expect(page.locator('[data-site-dns-signal="ptr_empty"] p')).toHaveAttribute('data-tone', 'neutral')
  await expect(page.locator('[data-site-dns-signal="private_ip"] p')).toHaveAttribute('data-tone', 'warning')
  await review(page, 'task-d-dns-1440-light')
  await assertRuntimeSurface(page, '[data-site-detail]', 'light')
  expect(runtime.calls).toHaveLength(5); runtime.assertQuiet()
})

test('Task D Security evidence and Insights help/changes preserve request ownership', async ({ page, runtime }) => {
  Object.assign(runtime.state.security, { enabled: true, days: 7, stale: true })
  runtime.state.intelligence.rich = true
  await open(page, '/en/site/41?tab=security')
  await expect(page.locator('[data-site-security-overview-composite] [data-site-security-attention]')).toHaveCount(1)
  await expect(page.locator('[data-site-security-scope]')).toHaveCount(0)
  await review(page, 'task-d-security-overview-1440-light')
  await page.locator('[data-site-security-tab="tls"]').click()
  const transport = page.locator('[data-site-security-transport-composite]')
  await expect(transport).toBeVisible()
  expect(await transport.locator(':scope > div > div').evaluateAll(nodes => nodes.map(node => node.children.length))).toEqual([2, 2, 2])
  await expect(transport.locator('[data-site-transport-details]')).not.toHaveAttribute('open')
  await taskDTooltip(page)
  await transport.locator('[data-site-transport-details] summary').click()
  await expect(transport.locator('[data-site-evidence="cipher_suite"]')).toBeVisible()
  await expect(page.locator('[data-site-certificate-identity] [data-site-certificate-crypto]')).toHaveCount(1)
  await expect(page.locator('[data-site-certificate-crypto]')).not.toHaveAttribute('open')
  await page.locator('[data-site-certificate-crypto] summary').click()
  await expect(page.locator('[data-site-evidence="cert_fingerprint_sha256"]')).toBeVisible()
  await review(page, 'task-d-tls-1440-light')
  await page.locator('[data-site-primary-tab="insights"]').click()
  await expect(page.locator('[data-site-insight-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await expect(page.locator('[data-site-insights-header]')).not.toContainText('Site-wide data')
  await taskDTrendGap(page)
  await taskDTooltip(page)
  await taskDChanges(page, '[data-site-insight-changes]', 6)
  await review(page, 'task-d-insights-1440-light')
  const dates = await page.locator('[data-site-insight-change] time').allTextContents()
  await page.locator('[data-site-target-trigger]').click(); await page.locator('[data-site-target-option="alt.example"]').click()
  await expect(page.locator('[data-site-detail]')).toHaveAttribute('data-site-target', 'alt.example')
  expect(await page.locator('[data-site-insight-change] time').allTextContents()).toEqual(dates)
  expect(runtime.count('/sites/41/detail')).toBe(2); expect(runtime.count('/sites/41/insights')).toBe(1)
  expect(runtime.count('/sites/41/view')).toBe(1); expect(runtime.count('/trend')).toBe(1); expect(runtime.count('/observations')).toBe(0)
  await assertRuntimeSurface(page, '[data-site-detail]', 'light'); runtime.assertQuiet()
})

for (const [width, theme] of [[390, 'light'], [1440, 'dark']] as const) test(`Task D focused responsive composition/help ${width}-${theme}`, async ({ page, context, runtime }) => {
  Object.assign(runtime.state, { fullCapabilities: true, manyChanges: true, summaryScenario: 'mixed', observationRich: true, redirectCount: 7 })
  await page.setViewportSize({ width, height: 900 }); await context.addInitScript(value => localStorage.setItem('theme', value), theme)
  await open(page)
  await expect(page.locator('[data-site-overview-composite] [data-site-overview-attention]')).toBeVisible()
  await taskDChanges(page, '[data-site-recent-changes]', 4, width >= 768)
  await review(page, `task-d-overview-${width}-${theme}`)
  // Direct secondary deep link avoids activating Performance while checking HTTP layout.
  await page.goto('/en/site/41?tab=observation&view=http', { waitUntil: 'load' })
  await expect(page.locator('[data-site-redirect-node]')).toHaveCount(7)
  await expect(page.locator('[data-site-redirect-node] [data-direction][aria-hidden="true"]')).toHaveCount(6)
  await expect(page.locator('[data-site-redirect-node]').last().locator('[data-direction]')).toHaveCount(0)
  await assertRuntimeSurface(page, '[data-site-detail]', theme)
  await review(page, `task-d-http-${width}-${theme}`)
  await page.locator('[data-site-primary-tab="insights"]').click()
  await expect(page.locator('[data-site-insight-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await taskDTooltip(page)
  await assertRuntimeSurface(page, '[data-site-detail]', theme)
  await review(page, `task-d-help-${width}-${theme}`)
  runtime.assertQuiet()
})
