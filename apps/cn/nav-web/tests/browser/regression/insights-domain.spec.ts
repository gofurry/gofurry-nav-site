import type { Page } from '@playwright/test'
import { mkdir, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { assertRuntimeSurface, revealImages } from '../fixtures/insights-runtime'
import { test, expect, domainMetricKeys, domainDimensionKeys, domainPulseCapsuleSizes, openRuntime, keyboardFocus } from '../fixtures/insights-domain'

const routes = [
  ['/insights/sites?metric=ipv6&range=30d&dimension=country', 'site'],
  ['/en/insights/sites?metric=tls13&range=90d&dimension=group', 'site'],
  ['/insights/games?metric=free&range=30d&dimension=primary_tag', 'game'],
  ['/en/insights/games?metric=linux&range=90d&dimension=tag', 'game'],
] as const
const afterMeasurements: Awaited<ReturnType<typeof measurePulse>>[] = []
async function measurePulse(page: Page) {
  await revealImages(page, '.insights-game-pulse .insight-entity-media')
  return page.locator('.insights-game-pulse').evaluate(root => {
    const css = getComputedStyle(root)
    return {
      width: innerWidth,
      overflow: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth,
      grid: { columns: css.gridTemplateColumns, rows: css.gridTemplateRows, gap: css.gap },
      items: [...root.querySelectorAll<HTMLElement>('[data-pulse]')].map(el => {
        const media = el.querySelector<HTMLElement>('.insight-entity-media'), img = media?.querySelector('img')
        const rect = img?.getBoundingClientRect(), scale = img && rect ? Math.min(rect.width / img.naturalWidth, rect.height / img.naturalHeight) : 0
        const body = el.querySelector<HTMLElement>('.insights-game-pulse__body'), style = media && getComputedStyle(media)
        return {
          role: el.dataset.pulse, box: el.getBoundingClientRect().toJSON(), body: body?.getBoundingClientRect().toJSON(),
          media: media?.getBoundingClientRect().toJSON(), image: rect?.toJSON(), natural: img ? [img.naturalWidth, img.naturalHeight] : null,
          source: media?.dataset.pulseSource, state: media?.dataset.mediaState,
          aspectRatio: style?.aspectRatio, maxHeight: style?.maxHeight, objectFit: img && getComputedStyle(img).objectFit,
          drawn: img && rect ? { width: img.naturalWidth * scale, height: img.naturalHeight * scale, sideBlank: (rect.width - img.naturalWidth * scale) / 2 } : null,
          textFits: [...el.querySelectorAll<HTMLElement>('.insights-game-pulse__body, strong, .insights-game-pulse__detail')].every(text => text.scrollWidth <= text.clientWidth + 1),
        }
      }),
    }
  })
}

async function pulseGeometry(page: Page, width: number) {
  const data = await measurePulse(page), [lead, discount, latest] = data.items
  expect(data.overflow).toBeLessThanOrEqual(1)
  expect(data.items.map(item => item.role)).toEqual(['players', 'discount', 'latest'])
  expect(latest!.box.top - discount!.box.bottom).toBeGreaterThanOrEqual(20)
  expect(latest!.box.top - discount!.box.bottom).toBeLessThanOrEqual(24.5)
  const tracks = data.grid.columns.split(' ').map(Number.parseFloat)
  if (width > 760) {
    expect(tracks).toHaveLength(2)
    expect(tracks[0]! / tracks[1]!).toBeCloseTo(7 / 5, 2)
    expect(Math.abs(lead!.box.top - discount!.box.top)).toBeLessThan(1)
    expect(discount!.box.left).toBeGreaterThanOrEqual(lead!.box.right)
  } else {
    expect(tracks).toHaveLength(1)
    expect(discount!.box.top).toBeGreaterThanOrEqual(lead!.box.bottom + 19)
  }
  for (const item of data.items) {
    expect(item.textFits).toBe(true)
    if (!item.media) { expect(item.box.height).toBeLessThan(110); continue }
    expect(item.media.height).toBeGreaterThan(20)
    expect(item.media.width).toBeLessThanOrEqual(item.box.width + 1)
    if (item.role !== 'players') {
      expect(item.media.width).toBeLessThanOrEqual(width > 1100 ? 144 : 88)
      expect(item.body.left).toBeGreaterThanOrEqual(item.media.right + 13)
      expect(Math.abs(item.media.top - item.body.top)).toBeLessThan(1)
    }
    if (item.image) {
      expect(item.objectFit).toBe('contain')
      // Compare actual contain drawing to both boxes, not just the CSS keyword.
      expect(Math.abs(item.image.width - item.drawn!.width)).toBeLessThan(1)
      expect(Math.abs(item.image.height - item.drawn!.height)).toBeLessThan(1)
      expect(Math.abs(item.media.width - item.drawn!.width)).toBeLessThan(1)
      expect(Math.abs(item.media.height - item.drawn!.height)).toBeLessThan(1)
      if (item.source === 'header') {
        expect(item.natural).toEqual([460, 215])
        expect(item.aspectRatio).toBe('460 / 215')
        expect(item.maxHeight).toBe('none')
        expect(item.media.height).toBeCloseTo(item.media.width * 215 / 460, 0)
      } else {
        expect(item.source).toBe('capsule')
        expect(item.media.height).toBeLessThanOrEqual(item.role === 'players' ? 300 : width > 1100 ? 92 : 72)
      }
    } else {
      expect(item.state).toBe('fallback')
      expect(item.media.height).toBeCloseTo(item.media.width * 215 / 460, 0)
      expect(item.media.height).toBeLessThan(400)
    }
  }
  return data
}

async function pulseEvidence(page: Page, name: string) {
  const dir = process.env.GOFURRY_DOMAIN_REVIEW_DIR
  if (!dir) return
  await mkdir(dir, { recursive: true })
  // Focus is asserted before export; review the resting composition separately.
  await page.locator('.insight-domain-pulse h2').click()
  await page.locator('.insight-domain-pulse').screenshot({ path: join(dir, `pulse-${name}.png`) })
}
async function layout(page: Page, width: number, domain: string) {
  const data = await page.evaluate(() => ({
    overflow: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth,
    chart: document.querySelector('.insights-trend')!.getBoundingClientRect().toJSON(),
    context: document.querySelector('.insight-trend-context')!.getBoundingClientRect().toJSON(),
    tracks: getComputedStyle(document.querySelector('.insight-trend-workspace')!).gridTemplateColumns.split(' ').map(Number.parseFloat),
  }))
  expect(data.overflow).toBeLessThanOrEqual(1)
  if (domain === 'site' ? width > 760 : width > 1100) {
    expect(data.tracks).toHaveLength(2)
    // Grid tracks measure the available columns, including the context's own
    // internal padding. Do not compare chart width to only the sidebar text box.
    const ratio = domain === 'site' ? 2 : 3
    expect(data.tracks[0]! / data.tracks[1]!).toBeGreaterThan(ratio - .15)
    expect(data.tracks[0]! / data.tracks[1]!).toBeLessThan(ratio + .15)
    expect(Math.abs(data.chart.width - data.tracks[0]!)).toBeLessThan(1)
    expect(Math.abs(data.context.width - data.tracks[1]!)).toBeLessThan(1)
    expect(data.context.left).toBeGreaterThanOrEqual(data.chart.right)
  } else {
    expect(data.tracks).toHaveLength(1)
    expect(data.context.top).toBeGreaterThanOrEqual(data.chart.bottom)
  }
}

async function foundationAppearance(page: Page, width: number, domain: string) {
  const result = await page.evaluate(() => {
    const css = (selector: string) => getComputedStyle(document.querySelector(selector)!)
    const rail = document.querySelector('[data-metric-rail]')!, railBox = rail.getBoundingClientRect()
    const items = [...rail.querySelectorAll('button')]
    return {
      title: { weight: css('h1').fontWeight, size: parseFloat(css('h1').fontSize) },
      intro: { weight: css('.insights-domain-header__intro').fontWeight, size: parseFloat(css('.insights-domain-header__intro').fontSize) },
      count: { weight: css('[data-domain-count]').fontWeight, size: parseFloat(css('[data-domain-count]').fontSize) },
      snapshot: css('.insights-domain-header__snapshot').fontSize,
      headings: [...document.querySelectorAll('.insights-trend h2, [data-dimension-explorer] h2')].map(el => getComputedStyle(el).fontWeight),
      values: [...rail.querySelectorAll('strong')].map(el => getComputedStyle(el).fontWeight),
      labels: [...document.querySelectorAll('.insight-dimension-bars__label strong')].map(el => getComputedStyle(el).fontWeight),
      scrollable: rail.scrollWidth > rail.clientWidth,
      scrollbar: css('[data-metric-rail]').scrollbarWidth,
      partialItem: items.some(el => { const box = el.getBoundingClientRect(); return box.left < railBox.right && box.right > railBox.right + 1 }),
      controls: [...items, ...document.querySelectorAll('[data-dimension]')].map(el => ({ height: el.getBoundingClientRect().height, background: getComputedStyle(el).backgroundColor, shadow: getComputedStyle(el).boxShadow })),
      selectedLine: getComputedStyle(rail.querySelector('[aria-pressed="true"]')!, '::after').height,
      textFits: [...document.querySelectorAll('[data-domain-header] p, [data-metric-key] > *, .insight-trend-context dd, .insight-dimension-bars__label')].every(el => el.scrollWidth <= el.clientWidth + 1),
      surfaces: ['[data-domain-header]', '[data-metric-rail]', '.insight-trend-workspace', '.insights-trend', '[data-dimension-explorer]'].map(selector => {
        const style = css(selector)
        return { background: style.backgroundColor, shadow: style.boxShadow, left: style.borderLeftWidth, right: style.borderRightWidth }
      }),
    }
  })
  expect(result.title.weight).toBe('700'); expect(result.title.size).toBeGreaterThanOrEqual(30); expect(result.title.size).toBeLessThanOrEqual(46)
  expect(result.intro.weight).toBe('400'); expect(result.intro.size).toBeGreaterThanOrEqual(14); expect(result.intro.size).toBeLessThanOrEqual(15)
  expect(result.count.weight).toBe('600'); expect(result.count.size).toBeGreaterThanOrEqual(25); expect(result.count.size).toBeLessThanOrEqual(28)
  expect(result.snapshot).toBe('12px')
  for (const weight of [...result.headings, ...result.values, ...result.labels]) expect(weight).toBe('600')
  expect(result.scrollbar).toBe('thin'); expect(result.selectedLine).toBe('2px'); expect(result.textFits).toBe(true)
  await expect(page.locator('[data-dimension][aria-pressed="true"]')).toHaveCSS('font-weight', '600')
  await expect(page.locator('[data-dimension][aria-pressed="true"]')).toHaveCSS('border-bottom-width', '2px')
  if (width === 390) { expect(result.scrollable).toBe(true); expect(result.partialItem).toBe(true) }
  if (width === 1440) expect(result.scrollable).toBe(false)
  for (const control of result.controls) {
    expect(control.height).toBeGreaterThanOrEqual(44)
    expect(control.background).toBe('rgba(0, 0, 0, 0)'); expect(control.shadow).toBe('none')
  }
  for (const surface of result.surfaces) expect(surface).toEqual({ background: 'rgba(0, 0, 0, 0)', shadow: 'none', left: '0px', right: '0px' })
  await expect(page.locator('[data-domain-header] time')).toHaveAttribute('datetime', '2026-09-09T04:00:00Z')
  await expect(page.locator('.insight-trend-context')).toContainText('2026-09-08')
  await expect(page.locator('.insight-trend-context')).not.toContainText('2026-09-09')
  await expect(page.locator('.insight-trend-context__delta')).toContainText('+4.2%')
  if (domain === 'site') {
    await expect(page.locator('.insight-trend-context')).toContainText('2026-06-01')
    await expect(page.locator('[data-metric-key="certificate_verified"] strong')).toHaveText('0.0%')
    await expect(page.locator('[data-metric-key="security_txt"] strong')).toHaveText('—')
  }
}

for (const [route, domain] of routes) for (const theme of ['light', 'dark'] as const) for (const width of [1440, 1024, 768, 390]) {
  test('Domain SSR/hydration, full raw data, focus ' + route + ' ' + theme + ' ' + width, async ({ page, context, request, runtime }) => {
    await page.setViewportSize({ width, height: 1000 })
    await context.addInitScript(value => localStorage.setItem('theme', value), theme)
    const ssr = await request.get(route), ssrHTML = await ssr.text()
    expect(ssr.status()).toBe(200)
    expect(ssrHTML).toMatch(/<h1>[^<]+<\/h1>/)
    for (const key of domainMetricKeys[domain]) expect(ssrHTML).toContain('data-metric-key="' + key + '"')
    expect(runtime.calls).toHaveLength(domain === 'site' ? 3 : 4)
    const offset = runtime.calls.length
    const html = await openRuntime(page, route)
    expect(html).toContain('data-domain-count>' + (domain === 'site' ? 238 : 213))
    expect(html).toContain('insights-domain-nav')
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('.insight-domain-activity .insight-activity-item')).toHaveCount(2)
    await expect(page.locator('.insight-domain-activity .insight-entity-media')).toHaveCount(2)
    await expect(page.locator('.overview-event, [data-activity-icon]')).toHaveCount(0)
    const prefix = route.startsWith('/en/') ? '/en' : ''
    expect(await page.locator('.insight-domain-continue a').evaluateAll(elements => elements.map(el => el.getAttribute('href'))))
      .toEqual((domain === 'game' ? ['/insights/games/players', '/insights/games/prices', '/insights/games/languages', '/insights/games/compare'] : ['/insights/sites/certificates', '/insights/sites/compare']).map(path => prefix + path))
    await expect(page.locator('[data-metric-key]')).toHaveCount(domainMetricKeys[domain].length)
    await expect(page.locator('[data-dimension]')).toHaveCount(domainDimensionKeys[domain].length)
    await expect(page.locator('[data-slice]')).toHaveCount(8)
    await expect(page.locator('[data-dimension-raw]')).not.toHaveAttribute('open')
    expect(await page.locator('[data-slice] progress').first().evaluate(el => ({ value: el.value, max: el.max })))
      .toEqual(domain === 'site' ? { value: .824, max: 1 } : { value: 46, max: 46 })
    await page.locator('[data-dimension-raw] summary').click()
    await expect(page.locator('[data-dimension-raw] tbody tr')).toHaveCount(10)
    if (width === 390) expect(await page.locator('.insight-dimension-raw__scroll').evaluate(el => el.scrollWidth > el.clientWidth)).toBe(true)
    await layout(page, width, domain)
    await foundationAppearance(page, width, domain)
    await assertRuntimeSurface(page, '.insights-domain-page', theme)
    const last = page.locator('[data-metric-key]').last()
    await keyboardFocus(last)
    expect(await last.evaluate(el => {
      const box = el.getBoundingClientRect(), rail = el.closest('[data-metric-rail]')!.getBoundingClientRect()
      const css = getComputedStyle(el)
      return box.left >= rail.left - 1 && box.right <= rail.right + 1
        && el.matches(':focus-visible') && css.outlineStyle === 'solid' && parseFloat(css.outlineWidth) >= 2
    })).toBe(true)
    await expect(page.locator('[data-metric-key][aria-pressed="true"]')).toHaveAttribute('data-metric-key', new URL(route, 'http://fixture.test').searchParams.get('metric')!)
    if (domain === 'game') {
      const measured = await pulseGeometry(page, width)
      await expect(page.locator('[data-pulse="players"]')).toContainText('1,240')
      await expect(page.locator('[data-pulse="players"]')).toContainText('2,800')
      await expect(page.locator('[data-pulse="discount"] b')).toHaveText(prefix ? '$5.99' : 'US$5.99')
      await expect(page.locator('[data-pulse="discount"]')).toContainText('50%')
      await expect(page.locator('[data-pulse="discount"]')).toContainText('USD')
      await expect(page.locator('[data-pulse="latest"]')).toContainText('2026-08-28')
      expect(await page.locator('a[data-pulse]').evaluateAll(links => links.map(link => link.getAttribute('href'))))
        .toEqual(['91', '92', '93'].map(id => `${prefix}/games/${id}`))
      for (const link of await page.locator('a[data-pulse]').all()) {
        await keyboardFocus(link)
        await expect(link).toHaveCSS('outline-width', '2px')
        await expect(link).toHaveCSS('box-shadow', 'none')
        await expect(link).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
        await expect(link.locator('img')).toHaveAttribute('alt', await link.locator('strong').innerText())
      }
      await pulseEvidence(page, `${prefix ? 'en' : 'zh'}-${theme}-${width}`)
      if (!prefix && theme === 'light' && process.env.GOFURRY_DOMAIN_REVIEW_DIR) {
        afterMeasurements.push(measured)
        await writeFile(join(process.env.GOFURRY_DOMAIN_REVIEW_DIR, '..', 'measure-after.json'), JSON.stringify(afterMeasurements, null, 2))
      }
    }
    // A few review images from this existing fixture, never accepted goldens.
    const reviewDir = process.env.GOFURRY_DOMAIN_REVIEW_DIR
    const representative = (domain === 'game' && !prefix && theme === 'light' && [1440, 390].includes(width))
      || (domain === 'game' && !prefix && theme === 'light' && width === 768)
      || (domain === 'game' && prefix && theme === 'dark' && width === 1024)
      || (domain === 'site' && !prefix && theme === 'light' && width === 1440)
      || (domain === 'site' && prefix && theme === 'dark' && width === 768)
    const alternate = (domain === 'game' && prefix && theme === 'dark' && width === 1440)
      || (domain === 'site' && !prefix && theme === 'dark' && width === 390)
    if (reviewDir && (representative || alternate)) {
      await mkdir(reviewDir, { recursive: true })
      await page.locator('[data-dimension-raw] summary').click()
      await page.locator('[data-domain-header] h1').click()
      await page.locator('[data-metric-rail]').evaluate(el => { el.scrollLeft = 0 })
      await revealImages(page)
      const name = `${domain}-${prefix ? 'en' : 'zh'}-${theme}-${width}`
      if (representative) await page.screenshot({ path: join(reviewDir, `full-${name}.png`), fullPage: true })
      await page.locator('[data-metric-rail]').screenshot({ path: join(reviewDir, `rail-${name}.png`) })
      await page.locator('.insight-trend-workspace').screenshot({ path: join(reviewDir, `trend-${name}.png`) })
      await page.locator('[data-dimension-explorer]').screenshot({ path: join(reviewDir, `dimensions-${name}.png`) })
    }
    expect(runtime.calls.slice(offset)).toHaveLength(domain === 'site' ? 3 : 4)
    expect(runtime.count('/game/home')).toBe(domain === 'game' ? 2 : 0)
    runtime.assertQuiet()
  })
}
for (const roles of [['players'], ['discount', 'latest'], ['players', 'latest'], ['players', 'discount'], []]) for (const width of [1440, 390]) {
  test(`Pulse missing roles ${roles.join('+') || 'none'} ${width}`, async ({ page, runtime }) => {
    runtime.state.pulseRoles = roles
    await page.setViewportSize({ width, height: 1000 })
    const html = await openRuntime(page, '/insights/games')
    expect(html).toContain('data-pulse="players"')
    expect(runtime.calls).toHaveLength(4)
    expect(await page.locator('a[data-pulse]').evaluateAll(links => links.map(link => link.getAttribute('data-pulse')))).toEqual(roles)
    const ids = { players: '91', discount: '92', latest: '93' }
    for (const [role, id] of Object.entries(ids)) {
      const item = page.locator(`[data-pulse="${role}"]`)
      if (roles.includes(role)) await expect(item).toHaveAttribute('href', `/games/${id}`)
      else {
        await expect(item).toContainText('生态观测数据暂不可用')
        await expect(item.locator('a, img')).toHaveCount(0)
      }
    }
    await pulseGeometry(page, width)
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-slice]')).toHaveCount(8)
    await pulseEvidence(page, `roles-${roles.join('-') || 'none'}-${width}`)
    runtime.assertQuiet()
  })
}

for (const width of [1440, 1024, 768, 390]) {
  test(`Pulse capsule natural ratios ${width}`, async ({ page, runtime }) => {
    runtime.state.pulseMedia = 'capsule'
    await page.setViewportSize({ width, height: 1000 })
    await page.route('**/media/game-pulse-*-capsule.svg', route => {
      const id = new URL(route.request().url()).pathname.match(/game-pulse-(\d+)-capsule/)![1]!
      const [w, h] = domainPulseCapsuleSizes[Number(id) as keyof typeof domainPulseCapsuleSizes]
      return route.fulfill({ contentType: 'image/svg+xml', body: `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}"><rect width="100%" height="100%" fill="#394b44"/><rect x="1" y="1" width="${w - 2}" height="${h - 2}" fill="none" stroke="#c8b398"/><text x="12" y="40" fill="white">CAPSULE FIXTURE</text></svg>` })
    })
    await openRuntime(page, '/insights/games')
    const data = await pulseGeometry(page, width)
    expect(data.items.map(item => item.natural)).toEqual(Object.values(domainPulseCapsuleSizes))
    expect(runtime.calls).toHaveLength(4)
    await pulseEvidence(page, `capsule-${width}`)
    runtime.assertQuiet()
  })
}

for (const mode of ['header-with-capsule', 'capsule', 'missing']) for (const width of [1440, 390]) {
  test(`Pulse media fallback ${mode} ${width}`, async ({ page, runtime }) => {
    runtime.state.pulseMedia = mode
    runtime.failImages = mode !== 'missing'
    const images: string[] = []
    page.on('request', req => { if (req.resourceType() === 'image') images.push(new URL(req.url()).pathname) })
    await page.setViewportSize({ width, height: 1000 })
    await openRuntime(page, '/en/insights/games')
    await revealImages(page, '.insights-game-pulse .insight-entity-media')
    await expect(page.locator('.insights-game-pulse img')).toHaveCount(0)
    await expect(page.locator('.insights-game-pulse [data-media-state="fallback"]')).toHaveCount(3)
    await pulseGeometry(page, width)
    const pulseRequests = () => images.filter(path => path.includes('game-pulse-'))
    const requestsAtFallback = pulseRequests().length
    for (const [role, id, name] of [['players', '91', 'Active game fixture'], ['discount', '92', 'Discount game fixture'], ['latest', '93', 'Recent release fixture']]) {
      const link = page.locator(`a[data-pulse="${role}"]`)
      await expect(link).toHaveAttribute('href', `/en/games/${id}`)
      await expect(link.locator('[role="img"]')).toHaveAttribute('aria-label', name!)
      await expect(link.locator('strong')).toHaveText(name!)
      await keyboardFocus(link)
    }
    await expect(page.locator('[data-pulse="players"]')).toContainText('1,240')
    await expect(page.locator('[data-pulse="discount"] b')).toHaveText('$5.99')
    await expect(page.locator('[data-pulse="latest"]')).toContainText('2026-08-28')
    expect(pulseRequests()).toHaveLength(requestsAtFallback)
    if (mode === 'header-with-capsule') {
      // Resource identity is the contract: SSR/hydration can attempt the same
      // failed URL more than once. No capsule chain or post-fallback requests.
      expect([...new Set(pulseRequests())].sort()).toEqual(['91', '92', '93'].map(id => `/media/game-pulse-${id}-header.svg`))
      expect(images.some(path => path.includes('-capsule.svg'))).toBe(false)
    }
    expect(runtime.calls).toHaveLength(4)
    await pulseEvidence(page, `fallback-${mode}-${width}`)
    runtime.assertQuiet()
  })
}

for (const locale of ['zh', 'en']) for (const width of [1440, 1024, 768, 390]) {
  test(`Pulse long names and observed zero ${locale} ${width}`, async ({ page, runtime }) => {
    runtime.state.pulseLongNames = locale
    runtime.state.pulseZero = true
    await page.setViewportSize({ width, height: 1000 })
    await openRuntime(page, `${locale === 'en' ? '/en' : ''}/insights/games`)
    await pulseGeometry(page, width)
    await expect(page.locator('.insights-game-pulse__count')).toHaveText(locale === 'en' ? '0 players' : '0 位玩家')
    await expect(page.locator('[data-pulse="discount"] b')).toHaveText(locale === 'en' ? '$5.99' : 'US$5.99')
    await expect(page.locator('[data-pulse="discount"]')).toContainText('50%')
    await expect(page.locator('[data-pulse="discount"]')).toContainText('USD')
    expect(await page.locator('[data-pulse="players"] strong').evaluate(el => {
      const name = el.getBoundingClientRect(), count = document.querySelector('.insights-game-pulse__count')!.getBoundingClientRect()
      return name.right <= count.left || name.bottom <= count.top || count.bottom <= name.top
    })).toBe(true)
    expect(runtime.calls).toHaveLength(4)
    await pulseEvidence(page, `long-name-${locale}-${width}`)
    runtime.assertQuiet()
  })
}

test('Domain Site artwork failure retains shared identity and entity links', async ({ page, runtime }) => {
  runtime.failImages = true
  await openRuntime(page, '/insights/sites')
  await revealImages(page, '.insight-domain-activity .insight-entity-media')
  await expect(page.locator('.insight-domain-activity img')).toHaveCount(0)
  await expect(page.locator('.insight-domain-activity [role="img"][aria-label]')).toHaveCount(2)
  await expect(page.locator('.insight-domain-activity [data-change-link]').first()).toHaveAttribute('href', '/site/41')
  await expect(page.locator('.insight-domain-activity [data-change-link]').first()).toContainText('Site fixture')
  await keyboardFocus(page.locator('.insight-domain-activity [data-change-link]').first())
  runtime.assertQuiet()
})
for (const domain of ['site', 'game'] as const) {
  test('Domain missing overview and breakdown stay independent ' + domain, async ({ page, runtime }) => {
    const path = '/insights/' + (domain === 'site' ? 'sites' : 'games')
    runtime.state.overviewFailure = true
    await openRuntime(page, path)
    await expect(page.locator('[data-domain-count]')).toHaveText('—')
    await expect(page.locator('[data-domain-header] time')).toHaveCount(0)
    await expect(page.locator('.insight-trend-context__value')).toHaveText('—')
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-slice]')).toHaveCount(8)
    runtime.state.overviewFailure = false; runtime.state.breakdownFailure = true
    await openRuntime(page, path)
    await expect(page.locator('[data-domain-count]')).toHaveText(domain === 'site' ? '238' : '213')
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-dimension-explorer] .insights-empty-state')).toHaveText('生态观测数据暂不可用')
    await expect(page.locator('[data-slice], [data-dimension-raw]')).toHaveCount(0)
    runtime.assertQuiet()
  })
  test('Domain selected slice SSR and deferred all-history ' + domain, async ({ page, request, runtime }) => {
    const path = '/insights/' + (domain === 'site' ? 'sites' : 'games')
    const metric = domain === 'site' ? 'tls13' : 'linux', dimension = domain === 'site' ? 'group' : 'tag'
    const response = await request.get(path + '?metric=' + metric + '&dimension=' + dimension + '&slice=31&range=30d')
    expect(response.status()).toBe(200)
    expect(runtime.calls).toHaveLength(domain === 'site' ? 4 : 5)
    const offset = runtime.calls.length
    const html = await openRuntime(page, path + '?metric=' + metric + '&dimension=' + dimension + '&slice=31&range=all')
    expect(html).toContain('aria-busy="true"')
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    const calls = runtime.calls.slice(offset)
    expect(calls.filter(call => call.url.pathname.endsWith('/trend'))).toHaveLength(2)
    // A request-only SSR pass cannot execute the client-deferred histories.
    const before = runtime.calls.length
    await request.get(path + '?metric=' + metric + '&dimension=' + dimension + '&slice=31&range=all')
    expect(runtime.calls.slice(before).some(call => call.url.pathname.endsWith('/trend'))).toBe(false)
    runtime.assertQuiet()
  })
  test('Domain query normalization, slice toggling and locale preserve state ' + domain, async ({ page, runtime }) => {
    const path = '/insights/' + (domain === 'site' ? 'sites' : 'games')
    const metric = domain === 'site' ? 'tls13' : 'linux', dimension = domain === 'site' ? 'group' : 'tag'
    await openRuntime(page, path + '?metric=invalid&range=bad&dimension=bad&slice=bad&keep=yes')
    await expect(page).toHaveURL(url => url.searchParams.get('metric') === domainMetricKeys[domain][0]
      && url.searchParams.get('dimension') === domainDimensionKeys[domain][0] && !url.searchParams.has('slice'))
    await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-metric', domainMetricKeys[domain][0]!)
    await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-dimension', domainDimensionKeys[domain][0]!)
    if (domain === 'game') {
      await page.locator('[data-metric-key="mac"]').click()
      await expect(page).toHaveURL(url => url.searchParams.get('metric') === 'mac')
      await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-metric', 'mac')
      await expect(page.locator('.insights-trend canvas')).toBeVisible()
    } else {
      await page.locator('[data-metric-key="security_txt"]').click()
      await expect(page.locator('.insights-domain-page')).toHaveAttribute('data-selected-metric', 'security_txt')
    }
    await page.locator('[data-metric-key="' + metric + '"]').click()
    await page.locator('[data-range="90d"]').click()
    await page.locator('[data-dimension="' + dimension + '"]').click()
    await expect(page.locator('[data-overlapping]')).toBeVisible()
    await page.locator('[data-slice="31"]').click()
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    await expect(page.locator('[data-slice="31"] strong')).toHaveCSS('text-decoration-line', 'underline')
    await expect(page.locator('[data-slice="31"] strong')).toHaveCSS('text-decoration-thickness', '2px')
    await page.locator('[data-slice="31"]').click()
    await expect(page.locator('.insights-slice-trend')).toHaveCount(0)
    await page.locator('[data-slice="31"]').click()
    const main = page.waitForResponse(res => res.url().includes('/metrics/' + metric + '/trend?range=all'))
    const slice = page.waitForResponse(res => res.url().includes('/breakdown/' + dimension + '/31/trend?range=all'))
    await page.locator('[data-range="all"]').click()
    await Promise.all([main, slice])
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    await page.locator('button:has(img[alt="EN"])').first().click()
    await expect(page).toHaveURL(url => url.pathname === '/en' + path)
    await expect(page.locator('.insights-slice-trend canvas')).toBeVisible()
    const query = new URL(page.url()).searchParams
    for (const [key, value] of Object.entries({ metric, range: 'all', dimension, slice: '31', keep: 'yes' })) expect(query.get(key)).toBe(value)
    runtime.assertQuiet()
  })
}
for (const [mode, text] of [['empty', '暂无可用历史数据'], ['one', '正在积累历史数据'], ['fail', '生态观测数据暂不可用']]) {
  test('Domain independent history state ' + mode, async ({ page, runtime }) => {
    runtime.state.trendMode = mode!
    await openRuntime(page, routes[0][0])
    await expect(page.locator('.insights-trend').getByText(text!, { exact: true })).toBeVisible()
    runtime.assertQuiet()
  })
}
test('Domain panel failure and gated history leave independent content and Dark chart usable', async ({ page, runtime }) => {
  runtime.state.panelFailure = true
  await openRuntime(page, routes[2][0])
  await expect(page.locator('.insights-game-pulse__unavailable')).toBeVisible()
  await expect(page.locator('.insights-trend canvas')).toBeVisible()
  await expect(page.locator('[data-slice]')).toHaveCount(8)
  expect(await page.locator('[data-domain-activity] [data-change-link]').count()).toBeGreaterThan(0)
  runtime.state.panelFailure = false
  await openRuntime(page, routes[0][0])
  const gate = runtime.hold(url => url.pathname.endsWith('/tls13/trend'))
  await page.locator('[data-metric-key="tls13"]').click(); await gate.wait()
  await expect(page.locator('.insights-trend [aria-busy="true"]')).toBeVisible()
  gate.release(); await gate.done()
  await expect(page.locator('.insights-trend canvas')).toBeVisible()
  await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  runtime.assertQuiet()
})

for (const [route, domain] of [routes[0], routes[2]]) for (const width of [1440, 390]) {
  test(`Domain Dark ready shell ${domain} ${width}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width, height: width === 390 ? 844 : 900 })
    await context.addInitScript(() => localStorage.setItem('theme', 'dark'))
    await openRuntime(page, route)
    await expect(page.locator('.insights-trend canvas')).toBeVisible()
    await expect(page.locator('[data-metric-rail]')).toBeVisible()
    await expect(page.locator('[data-dimension-explorer]')).toBeVisible()
    await expect(page.locator('.insights-data-info')).toBeVisible()
    await assertRuntimeSurface(page, '.insights-domain-page', 'dark')
    await layout(page, width, domain)
    expect(runtime.calls).toHaveLength(domain === 'site' ? 3 : 4)
    runtime.assertQuiet()
  })
}
