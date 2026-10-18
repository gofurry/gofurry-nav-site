import type { Page } from '@playwright/test'
import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface } from '../fixtures/site-detail'
const tab = (page: Page, view: string) => page.locator(`[data-site-security-tab="${view}"]`)
async function openSecurity(page: Page, view = 'overview', domain = '') {
  const counted = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  const html = await openRuntime(page, '/en/site/41?tab=security' + (view === 'overview' ? '' : '&view=' + view) + (domain ? '&domain=' + domain : ''))
  await (await counted).finished()
  return html
}
async function activeView(page: Page, view: string) {
  await expect(tab(page, view)).toHaveAttribute('aria-selected', 'true')
  await expect(tab(page, view)).toHaveAttribute('tabindex', '0')
  await expect(page.locator('[data-site-security-tab][tabindex="0"]')).toHaveCount(1)
  await expect(page.locator('[data-site-security-view]')).toHaveAttribute('data-site-security-view', view)
}
for (const view of ['overview', 'tls', 'web', 'exposure']) test('Security SSR owns existing Detail evidence only: ' + view, async ({ request, runtime }) => {
  runtime.state.security.enabled = true
  const response = await request.get('/en/site/41?tab=security&view=' + view)
  expect(response.status()).toBe(200)
  const html = await response.text()
  expect(html).toContain(`data-site-security-view="${view}"`)
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/nav/sites/41/detail', '/api/v2/nav/sites/41/insights', '/api/v2/nav/sites/41/recommendations'])
  runtime.assertQuiet()
})
test('Security router, back/forward, reload and keyboard retain Target with no view fetch', async ({ page, runtime }) => {
  runtime.state.security.enabled = true
  await openSecurity(page, 'web', 'alt.example')
  await tab(page, 'tls').click(); await activeView(page, 'tls')
  await tab(page, 'exposure').click(); await activeView(page, 'exposure')
  await page.goBack(); await activeView(page, 'tls'); await page.goForward(); await activeView(page, 'exposure')
  await tab(page, 'tls').focus()
  for (const [key, view] of [['ArrowRight', 'web'], ['End', 'exposure'], ['Home', 'overview'], ['ArrowLeft', 'exposure']]) {
    await page.keyboard.press(key!); await activeView(page, view!); await expect(tab(page, view!)).toBeFocused()
    if (view === 'overview') expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ domain: 'alt.example', tab: 'security' })
  }
  expect(runtime.calls).toHaveLength(4)
  const counted = page.waitForResponse(response => new URL(response.url()).pathname.endsWith('/sites/41/view'))
  expect((await page.reload({ waitUntil: 'domcontentloaded' }))?.status()).toBe(200)
  await (await counted).finished(); await activeView(page, 'exposure')
  expect(runtime.calls).toHaveLength(8)
  await tab(page, 'web').click(); await page.locator('[data-site-security-raw-headers]').click()
  await expect(page.locator('[data-site-observation-view]')).toHaveAttribute('data-site-observation-view', 'http')
  expect(Object.fromEntries(new URL(page.url()).searchParams)).toEqual({ domain: 'alt.example', tab: 'observation', view: 'http' })
  expect(runtime.calls).toHaveLength(8)
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', 'https://go-furry.com/en/site/41')
  runtime.assertQuiet()
})
test('raw headers keep an implicit primary Target without creating a Detail fetch', async ({ page, runtime }) => {
  runtime.state.security.enabled = true
  await openSecurity(page, 'web')
  await page.locator('[data-site-security-raw-headers]').click()
  await expect(page.locator('[data-site-observation-view]')).toHaveAttribute('data-site-observation-view', 'http')
  await settleRuntime(page); expect(runtime.calls).toHaveLength(4)
  runtime.assertQuiet()
})
for (const view of ['overview', 'tls', 'web', 'exposure']) test('Security Target switch adds only Detail and retains ' + view, async ({ page, runtime }) => {
  runtime.state.security.enabled = true
  await openSecurity(page, view)
  const held = runtime.hold(url => url.pathname.endsWith('/detail') && url.searchParams.get('target') === 'alt.example')
  try {
    await page.locator('[data-site-target-trigger]').click(); await page.locator('[data-site-target-option="alt.example"]').click(); await held.wait()
    await expect(page.locator('[data-site-security]')).toHaveAttribute('data-site-security-target', 'target.example')
    held.release(); await held.done()
    await expect(page.locator('[data-site-security]')).toHaveAttribute('data-site-security-target', 'alt.example')
    await activeView(page, view)
    if (view === 'tls') await expect(page.locator('[data-site-security-tls]')).toContainText('TLS 1.2')
    if (view === 'web') await expect(page.locator('[data-site-security-txt]')).toContainText('security@alt.example')
    await settleRuntime(page)
    expect(runtime.calls).toHaveLength(5); expect(runtime.count('/sites/41/detail')).toBe(2)
    expect(runtime.count('/sites/41/insights')).toBe(1); expect(runtime.count('/sites/41/view')).toBe(1)
    expect(runtime.count('/observations')).toBe(0)
    runtime.assertQuiet()
  } finally { held.release() }
})
for (const [tls, state] of [['verified', 'verified'], ['failed', 'failed'], ['missing_verification', 'not_observed'], ['not_collected', 'not_observed'], ['missing', 'not_observed'], ['not_tls', 'not_applicable']] as const) {
  test('TLS verification preserves ' + tls, async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, tls })
    await openSecurity(page, 'tls')
    await expect(page.locator('[data-site-certificate-verification]')).toHaveAttribute('data-state', state)
    if (tls === 'failed') await expect(page.locator('[data-site-certificate]')).toContainText('unknown_authority')
    if (tls === 'not_collected' || tls === 'not_tls') {
      await expect(page.locator('[data-site-certificate-expiry]')).toHaveAttribute('data-expiry', 'not_observed')
      const strip = page.locator('[data-site-health="certificate"]')
      await expect(strip.locator('.site-detail-health__value')).toHaveAttribute('data-tone', 'neutral')
      await expect(strip).toContainText('Not observed')
      await expect(strip).not.toContainText('0 days')
    }
    await expect(page.locator('[data-site-security-tab]')).toHaveCount(4)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
for (const [days, expiry] of [[31, 'normal'], [30, 'attention'], [7, 'warning'], [0, 'expired']] as const) {
  test('Certificate expiry ' + expiry + ' never changes verification', async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, days })
    await openSecurity(page, 'tls')
    await expect(page.locator('[data-site-certificate-expiry]')).toHaveAttribute('data-expiry', expiry)
    await expect(page.locator('[data-site-certificate-verification]')).toHaveAttribute('data-state', 'verified')
    await tab(page, 'overview').click()
    await expect(page.locator('[data-site-security-attention-item="expiry"]')).toHaveCount(expiry === 'normal' ? 0 : 1)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
for (const [legacyDays, days, state] of [[undefined, 45, 'normal'], [7, 7, 'warning']] as const) {
  test('Certificate observation-time fallback shares all displays: ' + (legacyDays === undefined ? 'V2 missing days' : 'explicit days win'), async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, days: legacyDays })
    const html = await openSecurity(page, 'tls')
    expect(html).toContain(`${days} days left`)
    await expect(page.locator('[data-site-health="certificate"]')).toContainText(`${days} days left`)
    await expect(page.locator('[data-site-certificate-expiry]')).toHaveText(`${days} days left`)
    await expect(page.locator('[data-site-certificate-expiry]')).toHaveAttribute('data-expiry', state)
    await expect(page.locator('[data-site-certificate-verification]')).toHaveAttribute('data-state', 'verified')
    await page.locator('[data-site-transport-details] summary').click()
    await expect(page.locator('[data-site-evidence="cert_days_left"] dd')).toHaveText(String(days))
    await tab(page, 'overview').click()
    await expect(page.locator('[data-site-security-summary="certificate"]')).toContainText(`${days} days left`)
    expect(runtime.calls).toHaveLength(4)
    runtime.assertQuiet()
  })
}
for (const [headers, found, missing] of [['all', 6, 0], ['some', 2, 4], ['missing', 0, 6], ['not_observed', 0, 0]] as const) {
  test('Security header evidence: ' + headers, async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, headers })
    await openSecurity(page, 'web')
    await expect(page.locator('[data-site-security-header]')).toHaveCount(6)
    await expect(page.locator('[data-site-security-header][data-state="present"]')).toHaveCount(found)
    await expect(page.locator('[data-site-security-header][data-state="missing"]')).toHaveCount(missing)
    if (headers === 'not_observed') await expect(page.locator('[data-site-security-header][data-state="not_observed"]')).toHaveCount(6)
    await expect(page.locator('[data-site-security-headers]')).not.toContainText(/Dangerous|Unsafe|Critical/)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
for (const [txt, state] of [['found', 'found'], ['issues', 'found_with_issues'], ['not_found', 'not_found'], ['unavailable', 'unavailable'], ['not_observed', 'not_observed']] as const) {
  test('security.txt distinguishes ' + txt, async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, txt })
    await openSecurity(page, 'web')
    await expect(page.locator('[data-site-security-txt]')).toHaveAttribute('data-state', state)
    if (txt === 'issues') {
      await expect(page.locator('[data-site-security-txt-validation]')).toContainText('Contact information is missing or invalid')
      await expect(page.locator('[data-site-security-txt-validation]')).not.toContainText('contact_missing_or_invalid')
    }
    if (txt === 'unavailable') await expect(page.locator('[data-site-security-txt]')).toContainText('Fixture probe unavailable')
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
for (const [ports, state] of [['empty', 'empty'], ['skipped', 'skipped'], ['unavailable', 'unavailable'], ['not_observed', 'not_observed']] as const) {
  test('Port probe distinguishes ' + ports, async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, ports })
    await openSecurity(page, 'exposure')
    await expect(page.locator('[data-site-port-check]')).toHaveAttribute('data-state', state)
    if (ports === 'empty') await expect(page.locator('[data-site-port-result]')).toHaveCount(0)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
for (const [waf, state] of [['matched', 'matched'], ['unexpected_pass', 'mismatch'], ['network_error', 'mismatch'], ['unexpected_status', 'mismatch'], ['truncated', 'incomplete'], ['unavailable', 'unavailable'], ['not_observed', 'not_observed']] as const) {
  test('WAF canary evidence: ' + waf, async ({ page, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, waf })
    await openSecurity(page, 'exposure')
    await expect(page.locator('[data-site-waf-canary]')).toHaveAttribute('data-state', state)
    await expect(page.locator('[data-site-waf-cases]')).not.toHaveAttribute('open')
    if (waf === 'truncated') await expect(page.locator('[data-site-waf-truncated]')).toContainText('Incomplete evidence')
    if (waf === 'matched') await expect(page.locator('[data-site-waf-canary]')).toContainText('Canary behavior matched expected blocking')
    await expect(page.locator('[data-site-security]')).not.toContainText(/Security Score|WAF Enabled|WAF Detected|Protected by WAF|Highly Secure|Poor Security/)
    await tab(page, 'overview').click()
    await expect(page.locator('[data-site-security-attention]')).toHaveCount(state === 'mismatch' ? 1 : 0)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
for (const slice of ['insights', 'view']) test('Security survives optional ' + slice + ' failure', async ({ page, runtime }) => {
  runtime.state.security.enabled = true
  runtime.state.siteInsightsFailure = slice === 'insights'; runtime.state.viewFailure = slice === 'view'
  await openSecurity(page, 'tls')
  await expect(page.locator('[data-site-certificate-verification]')).toHaveAttribute('data-state', 'verified')
  expect(runtime.count('/sites/41/detail')).toBe(1); expect(runtime.count('/observations')).toBe(0)
  runtime.assertQuiet()
})
for (const failure of ['invalid-target', 'detail-failure']) test('Security keeps authoritative ' + failure, async ({ request, runtime }) => {
  runtime.state.security.enabled = true
  if (failure === 'detail-failure') runtime.state.failure = 'site'
  const response = await request.get('/en/site/41?tab=security&view=tls' + (failure === 'invalid-target' ? '&domain=foreign.example' : ''))
  expect(response.status()).toBe(failure === 'invalid-target' ? 404 : 503)
  expect(await response.text()).not.toContain('data-site-security-view')
  runtime.assertQuiet()
})
for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark'] as const) {
  test(`Security disclosures and long evidence ${width} ${theme}`, async ({ page, context, runtime }) => {
    Object.assign(runtime.state.security, { enabled: true, long: true })
    await page.setViewportSize({ width, height: 900 }); await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    await openSecurity(page, 'tls')
    for (const section of ['san', 'chain', 'crypto']) {
      const details = page.locator(`[data-site-certificate-${section}]`)
      await expect(details).not.toHaveAttribute('open'); await details.locator('summary').focus(); await page.keyboard.press('Enter')
      await expect(details).toHaveAttribute('open', '')
    }
    await expect(page.locator('[data-site-certificate-crypto] [data-site-evidence="ocsp_stapled"]')).toContainText('No')
    await expect(page.locator('[data-site-certificate-crypto] [data-site-evidence="sct_count"]')).toContainText('0')
    await expect(page.locator('[data-site-evidence="cert_fingerprint_sha256"] dd')).toHaveCSS('word-break', 'break-all')
    await assertRuntimeSurface(page, '[data-site-security]', theme)
    await tab(page, 'web').click()
    await expect(page.locator('[data-site-security-header="content_security_policy"]')).toContainText('report-uri')
    await expect(page.locator('[data-site-security-txt]')).toContainText('security@target.example')
    await assertRuntimeSurface(page, '[data-site-security]', theme)
    await tab(page, 'exposure').click()
    await expect(page.locator('[data-site-port-result]')).toHaveCount(5)
    for (const [state, tone] of [['open', 'neutral'], ['closed', 'muted'], ['timeout', 'warning'], ['filtered_suspected', 'muted'], ['skipped', 'muted']]) {
      await expect(page.locator(`[data-site-port-result][data-state="${state}"] dd[data-tone]`)).toHaveAttribute('data-tone', tone!)
    }
    await page.locator('[data-site-port-metadata] summary').click()
    await expect(page.locator('[data-site-port-metadata]')).toContainText('Invalid ports')
    await page.locator('[data-site-waf-cases] summary').click()
    await expect(page.locator('[data-site-waf-case]')).toHaveCount(2)
    await assertRuntimeSurface(page, '[data-site-security]', theme)
    expect(runtime.calls).toHaveLength(4); runtime.assertQuiet()
  })
}
