import assert from 'node:assert/strict'
import { readFile, mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { launchPerfBrowser, reportsDir } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { mockGameHome } from './fixtures/insights-overview.mjs'

const origins = { primary: 'https://primary.example', mirror: 'https://mirror.example' }
const iconKey = 'nav/sites/1/icon/' + 'a'.repeat(32) + '.svg'
const nextKey = iconKey.replace(/a{32}/, 'b'.repeat(32))
const heroKey = 'nav/hero/desktop/' + 'a'.repeat(32) + '.avif'
const patternKey = 'nav/patterns/' + 'a'.repeat(32) + '.svg'
const svg = '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="45"><rect width="120" height="45" fill="#78644b"/></svg>'
const probeBody = await readFile(new URL('./fixtures/cdn-probe.bin', import.meta.url))
const steamSource = 'https://shared.steamstatic.com/store_item_assets/steam/apps/570/header.jpg'
const app = await startInsightsFixtureApp(url => {
  if (url.pathname === '/api/v2/nav/home') return { data: {
    schema_version: 4, generated_at: new Date().toISOString(), cache_state: {},
    groups: [{ id: '1', name: 'Resource fixtures', info: '', priority: 1, sites: [{ id: '1', name: 'Fixture site', domain: 'example.com', info: 'Routing', icon: iconKey, nsfw: '0', welfare: '0', view_count: 0 }] }],
    spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
    saying: null, ping: {}, hero: { desktop: { id: '1', object_key: heroKey }, mobile: null },
  } }
  if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: { schema_version: 1, patterns: [{ id: '1', name: 'Fixture pattern', name_en: 'Fixture pattern', object_key: patternKey, light_color: '#123456', dark_color: '#abcdef', light_opacity: .12, dark_opacity: .08, default_size_px: 120 }] } }
  if (url.pathname.endsWith('/game/home')) {
    const home = mockGameHome()
    for (const game of home.panel.latest_games) game.header_url = steamSource + '?id=' + game.id
    return { data: home }
  }
  return { data: [] }
}, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: origins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: origins.mirror })
const browser = await launchPerfBrowser()
const output = join(reportsDir, 'resource-routing')
await mkdir(output, { recursive: true })
const errors = []
async function settleTab(page, index) {
  await page.waitForFunction(index => {
    const pages = document.querySelector('.preferences-pages')
    return pages && Math.abs(pages.scrollLeft - pages.clientWidth * index) < 2 && document.querySelectorAll('.preferences-tabs [role=tab]')[index]?.getAttribute('aria-selected') === 'true'
  }, index)
}
async function openRouting(page) {
  await page.locator('.gf-nav__mode-button').click()
  await page.getByRole('tab').last().click()
  await settleTab(page, 2)
}
async function save(page) {
  await page.locator('.gf-modal__header-actions .gf-button--primary').click()
  await page.locator('[data-resource-routing]').waitFor({ state: 'detached' })
}
async function cancel(page) {
  await page.locator('.gf-modal__header-actions .gf-button--ghost').click()
  await page.locator('[data-resource-routing]').waitFor({ state: 'detached' })
}
const cookieValue = async (context, name) => (await context.cookies()).find(cookie => cookie.name === name)?.value
async function changeResource(locator, name, value) {
  await locator.evaluate((element, { name, value }) => {
    const app = document.querySelector('#__nuxt').__vue_app__
    const find = vnode => {
      if (!vnode || typeof vnode !== 'object') return null
      const instance = vnode.component
      if (instance?.subTree.el === element && name in instance.props) return instance
      if (instance) { const found = find(instance.subTree); if (found) return found }
      if (vnode.suspense) { const found = find(vnode.suspense.activeBranch); if (found) return found }
      for (const child of Array.isArray(vnode.children) ? vnode.children : []) { const found = find(child); if (found) return found }
      return null
    }
    const instance = find(app._container._vnode)
    if (!instance) throw new Error('Resource component was not found')
    instance.props[name] = value
  }, { name, value })
}
async function setup(path = '/', initialCookies = [], legacy = null) {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1000 }, reducedMotion: 'reduce' })
  await context.addCookies(initialCookies.map(([name, value]) => ({ name, value, url: app.base })))
  const state = { primary: 160, mirror: 15, china: 180, global: 15, failed: false, blocked: '', release: null, probeRequests: [] }
  let release
  const gate = new Promise(resolve => { release = resolve })
  state.release = release
  await context.addInitScript(({ base, legacy }) => {
    if (location.origin !== base) return
    localStorage.setItem('gf_background_preference', JSON.stringify({ version: 1, source: 'server', pattern_id: '1', overrides: {} }))
    if (legacy) localStorage.setItem('gofurry:steam-shared-cdn-preference:v1', JSON.stringify(legacy))
  }, { base: app.base, legacy })
  await context.route('**/*', async route => {
    const url = new URL(route.request().url())
    if (url.origin === app.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
    const managed = Object.values(origins).includes(url.origin)
    const steam = url.hostname.startsWith('shared.')
    if (!managed && !steam) return route.abort()
    const probe = url.pathname.endsWith('cdn.bin') || url.searchParams.has('gf_probe')
    if (probe) {
      state.probeRequests.push(url.href)
      await gate
      const group = managed ? (url.origin === origins.primary ? 'primary' : 'mirror') : url.hostname.includes('eccdnx') ? 'china' : 'global'
      await delay(state[group])
      if (state.failed) return route.abort()
      try {
        await route.fulfill({ status: 200, contentType: managed ? 'application/octet-stream' : 'image/svg+xml', headers: { 'Access-Control-Allow-Origin': '*' }, body: managed ? probeBody : svg })
      } catch (error) {
        // A timed-out Image may have cancelled a gated fixture request already.
        if (!String(error).includes('Route is already handled')) throw error
      }
      return
    }
    if (state.blocked && url.href.includes(state.blocked)) return route.abort()
    return route.fulfill({ status: 200, contentType: 'image/svg+xml', headers: { 'Access-Control-Allow-Origin': '*' }, body: svg })
  })
  const page = await context.newPage()
  page.on('pageerror', error => errors.push(error.message))
  page.on('console', message => { if (/hydration.*mismatch/i.test(message.text())) errors.push(message.text()) })
  await page.goto(app.base + path, { waitUntil: 'domcontentloaded' })
  await page.waitForFunction(() => Boolean(document.querySelector('#__nuxt')?.__vue_app__))
  return { context, page, state }
}
try {
  const ssr = await (await fetch(app.base + '/', { headers: { Cookie: 'gf_asset_cdn=primary; gf_asset_cdn_mode=mirror' } })).text()
  assert(ssr.includes(origins.mirror + '/' + heroKey), 'SSR must honor pinned mode ahead of recommendation')
  const steamSSR = await (await fetch(app.base + '/games', { headers: { Cookie: 'gf_steam_asset_mode=global; gf_steam_asset_group=china' } })).text()
  assert(steamSSR.includes('https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/570/header.jpg'), 'Steam SSR must honor the pinned cookie')

  const { page, context, state } = await setup()
  const icon = page.locator('.nav-site-card__logo img').first()
  await icon.scrollIntoViewIfNeeded()
  await page.waitForFunction(() => {
    const icon = document.querySelector('.nav-site-card__logo img')
    return icon?.complete && icon.naturalWidth > 0 && document.querySelector('[data-pattern-status=server]')
  })
  const readAssets = () => page.evaluate(() => ({
    icon: document.querySelector('.nav-site-card__logo img').src,
    hero: document.querySelector('.nav-header__background--managed img').currentSrc,
    pattern: getComputedStyle(document.querySelector('.gf-public-background__pattern')).maskImage,
  }))
  const before = await readAssets()
  assert(Object.values(before).every(url => url.includes('primary.example')))
  await icon.evaluate(el => {
    window.assetSrcMutations = []
    new MutationObserver(records => window.assetSrcMutations.push(...records.map(() => el.src))).observe(el, { attributes: true, attributeFilter: ['src'] })
  })
  state.release()
  await page.waitForFunction(() => JSON.parse(localStorage.getItem('gf_asset_cdn_diagnostics') || 'null')?.selected === 'mirror')
  assert.deepEqual(await readAssets(), before, 'automatic probe changed a loaded icon/Hero/pattern')
  assert.equal(await cookieValue(context, 'gf_asset_cdn'), 'mirror')

  await openRouting(page)
  const tabs = page.getByRole('tab')
  assert.equal(await tabs.count(), 3)
  for (const [key, index] of [['Home', 0], ['ArrowLeft', 2], ['ArrowRight', 0], ['End', 2], ['ArrowLeft', 1], ['ArrowRight', 2]]) {
    await page.keyboard.press(key); await settleTab(page, index)
    assert.equal(await tabs.nth(index).evaluate(el => el === document.activeElement), true)
  }
  const sections = page.locator('[data-resource-routing] .resource-route')
  const assets = sections.nth(0), steam = sections.nth(1)
  await assets.getByRole('radio', { name: 'EdgeOne', exact: true }).check()
  await steam.getByRole('radio', { name: '全球线路', exact: true }).check()
  await cancel(page)
  assert.equal(await cookieValue(context, 'gf_asset_cdn_mode'), undefined)
  assert.equal(await cookieValue(context, 'gf_steam_asset_mode'), undefined)
  await openRouting(page)
  assert(await assets.getByRole('radio', { name: '自动优选' }).isChecked())
  state.primary = 10; state.mirror = 170
  await assets.getByRole('button', { name: '重新测速', exact: true }).click()
  await page.waitForFunction(() => JSON.parse(localStorage.getItem('gf_asset_cdn_diagnostics') || 'null')?.selected === 'primary')
  const diagnostics = await page.evaluate(() => localStorage.getItem('gf_asset_cdn_diagnostics'))
  assert(await assets.getByRole('button', { name: /秒后可重测/ }).isDisabled())
  await cancel(page)
  assert.equal(await page.evaluate(() => localStorage.getItem('gf_asset_cdn_diagnostics')), diagnostics)
  assert.deepEqual(await readAssets(), before, 'manual probe changed loaded assets')

  await openRouting(page)
  await assets.getByRole('radio', { name: 'Cloudflare' }).check()
  await steam.getByRole('radio', { name: '中国线路' }).check()
  await save(page)
  assert.equal(await cookieValue(context, 'gf_asset_cdn_mode'), 'mirror')
  assert.equal(await cookieValue(context, 'gf_steam_asset_mode'), 'china')
  assert.deepEqual(await readAssets(), before, 'saving a route changed loaded assets')
  assert.deepEqual(await page.evaluate(() => window.assetSrcMutations), [], 'loaded icon src was mutated')

  // Change the actual Vue prop: this is a new resource within the same component.
  await changeResource(icon, 'objectKey', nextKey)
  await page.waitForFunction(key => document.querySelector('.nav-site-card__logo img').src === 'https://mirror.example/' + key, nextKey)
  state.blocked = origins.mirror + '/' + iconKey
  await changeResource(icon, 'objectKey', iconKey)
  await page.waitForFunction(key => document.querySelector('.nav-site-card__logo img').src === 'https://primary.example/' + key, iconKey)
  assert.equal(await cookieValue(context, 'gf_asset_cdn_mode'), 'mirror', 'real fallback cleared pin')
  await openRouting(page)
  await assets.getByRole('radio', { name: 'Cloudflare' }).focus()
  assert.notEqual(await assets.getByRole('radio', { name: 'Cloudflare' }).evaluate(el => getComputedStyle(el.closest('label')).outlineStyle), 'none')
  await page.screenshot({ path: join(output, 'routing-light-desktop.png') })
  await page.setViewportSize({ width: 390, height: 844 })
  await settleTab(page, 2)
  assert.equal(await page.locator('.preferences-page').nth(2).evaluate(el => el.scrollWidth > el.clientWidth + 1), false)
  await page.screenshot({ path: join(output, 'routing-light-mobile.png') })
  await page.evaluate(() => document.documentElement.classList.add('dark'))
  await page.waitForTimeout(300)
  await steam.scrollIntoViewIfNeeded()
  await page.screenshot({ path: join(output, 'routing-dark-mobile.png') })
  await context.unrouteAll({ behavior: 'wait' }); await context.close()
  console.log('[routing] SSR, three-tab keyboard, Save/Cancel, diagnostics/cooldown, managed snapshots/fallback, light/dark/mobile PASS')

  const game = await setup('/games')
  const picture = game.page.locator('.game-group-page-live img').first()
  await picture.waitFor()
  await game.page.waitForFunction(() => document.querySelector('.game-group-page-live img')?.naturalWidth > 0)
  const first = await picture.getAttribute('src')
  assert(first.startsWith('https://shared.st.dl.eccdnx.com/'))
  game.state.release()
  await game.page.waitForFunction(() => JSON.parse(localStorage.getItem('gf_steam_asset_diagnostics_v1') || 'null')?.selected === 'global')
  assert.equal(await picture.getAttribute('src'), first, 'automatic probe replaced an already loaded Steam image')
  await openRouting(game.page)
  const gameSteam = game.page.locator('[data-resource-routing] .resource-route').nth(1)
  game.state.china = 5; game.state.global = 160
  await gameSteam.getByRole('button', { name: '重新测速', exact: true }).click()
  await game.page.waitForFunction(() => JSON.parse(localStorage.getItem('gf_steam_asset_diagnostics_v1') || 'null')?.selected === 'china')
  assert.equal(await picture.getAttribute('src'), first)
  await gameSteam.getByRole('radio', { name: '全球线路' }).check()
  await save(game.page)
  assert.equal(await picture.getAttribute('src'), first, 'Save replaced an already loaded Steam image')
  const newSource = steamSource + '?changed=1'
  await changeResource(picture, 'src', newSource)
  await game.page.waitForFunction(() => document.querySelector('.game-group-page-live img').src.includes('shared.akamai.steamstatic.com') && document.querySelector('.game-group-page-live img').src.includes('changed=1'))
  game.state.blocked = 'shared.akamai.steamstatic.com'
  await changeResource(picture, 'src', steamSource + '?changed=2')
  await game.page.waitForFunction(() => document.querySelector('.game-group-page-live img').src.includes('shared.cloudflare.steamstatic.com'))
  assert.equal(await cookieValue(game.context, 'gf_steam_asset_mode'), 'global')
  await game.context.unrouteAll({ behavior: 'wait' }); await game.context.close()
  console.log('[routing] Steam automatic/manual/Save snapshots, changed src and actual fallback PASS')

  const failed = await setup('/en/games', [['gf_asset_cdn_mode', 'mirror'], ['gf_steam_asset_mode', 'china']])
  failed.state.failed = true; failed.state.release()
  await failed.page.waitForFunction(() => localStorage.getItem('gf_steam_asset_diagnostics_v1'))
  await openRouting(failed.page)
  const failedSections = failed.page.locator('[data-resource-routing] .resource-route')
  assert(await failedSections.nth(0).getByRole('radio', { name: 'Cloudflare' }).isChecked())
  assert(await failedSections.nth(1).getByRole('radio', { name: 'China route' }).isChecked())
  assert.equal(await failed.page.locator('.resource-route__warning').count(), 2)
  await save(failed.page)
  assert.equal(await cookieValue(failed.context, 'gf_steam_asset_mode'), 'china')
  await failed.context.unrouteAll({ behavior: 'wait' }); await failed.context.close()

  const migrated = await setup('/games', [], { group: 'global', testedAt: Date.now() })
  await migrated.page.waitForFunction(() => document.cookie.includes('gf_steam_asset_group=global'))
  assert.equal(await cookieValue(migrated.context, 'gf_steam_asset_mode'), undefined, 'legacy automatic result was pinned')
  assert.equal(await migrated.page.evaluate(() => localStorage.getItem('gofurry:steam-shared-cdn-preference:v1')), null)
  assert((await migrated.page.locator('.game-group-page-live img').first().getAttribute('src')).startsWith('https://shared.st.dl.eccdnx.com/'), 'legacy migration replaced the hydrated snapshot')
  await openRouting(migrated.page)
  assert(await migrated.page.locator('[data-resource-routing] .resource-route').nth(1).getByRole('radio', { name: '自动优选' }).isChecked())
  migrated.state.release()
  await migrated.page.waitForFunction(() => localStorage.getItem('gf_steam_asset_diagnostics_v1') && localStorage.getItem('gf_asset_cdn_diagnostics'))
  await migrated.context.unrouteAll({ behavior: 'wait' }); await migrated.context.close()
  assert.deepEqual(errors, [])
  console.log('[routing] failed probes keep pins saveable, legacy recommendation-only migration, English UI and no hydration/browser errors PASS')
  console.log('Routing screenshots:', output)
} finally {
  for (const context of browser.contexts()) await context.unrouteAll({ behavior: 'ignoreErrors' })
  await browser.close()
  await app.close()
}
