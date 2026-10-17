import { test, expect } from '../fixtures/game-detail-contract'
import { mkdir, writeFile } from 'node:fs/promises'
import { join } from 'node:path'
import { keyboardFocus } from '../fixtures/insights-runtime'

for (const locale of ['zh', 'en'] as const) for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`F3 Game Insights style ownership ${locale} ${width} ${theme}`, async ({ detail }) => {
    await detail.open({ locale, width, theme }); await detail.tab('insights')
    const page = detail.page, root = page.locator('[data-game-insights]')
    await expect(root).toHaveAttribute('data-player-loaded-ranges', '30d')
    await expect(root).toHaveAttribute('data-price-loaded-ranges', 'CN:30d')
    await detail.settle(root)
    const group = page.locator('.game-insights-control-group .insights-ranges').first()
    const active = group.locator('[aria-pressed="true"]'), idle = group.locator('[aria-pressed="false"]').first()
    const read = (el: Element) => { const s = getComputedStyle(el); return { color: s.color, background: s.backgroundColor, padding: s.padding, weight: s.fontWeight, radius: s.borderRadius } }
    const styles = { active: await active.evaluate(read), idle: await idle.evaluate(read), hover: {} }
    await idle.hover(); styles.hover = await idle.evaluate(read)
    expect(styles.hover).toEqual(styles.active)
    expect(styles.active.background).not.toBe('rgba(0, 0, 0, 0)')
    expect(styles.idle.background).toBe('rgba(0, 0, 0, 0)')
    expect(styles.idle.color).not.toBe(styles.active.color)
    await keyboardFocus(idle)
    expect(await root.locator('.entity-insights-eyebrow').count()).toBe(0)
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1)
    const fonts = await root.locator('h2, h3, dt, dd, .game-insights-timeline__category, .game-insights-price-list__discount, .entity-insights-back-link').evaluateAll(els => els.map(el => ({ tag: el.tagName, class: el.className, weight: getComputedStyle(el).fontWeight })))
    for (const font of fonts) expect(['400', '500', '600', '700']).toContain(font.weight)
    await expect(root.locator('h2')).toHaveCSS('font-weight', '700')
    await expect(root.locator('.entity-insights-back-link')).toHaveCSS('font-weight', '600')
    await expect(root.locator('.game-insights-timeline__category').first()).toHaveCSS('font-weight', '500')
    const dir = process.env.GOFURRY_F3_REVIEW_DIR
    if (dir) {
      await mkdir(dir, { recursive: true })
      const key = `game-detail-${locale}-${theme}-${width}`
      await writeFile(join(dir, key + '.json'), JSON.stringify({ styles, fonts }, null, 2))
      await page.mouse.move(0, 0); await root.locator('h2').click()
      await page.screenshot({ path: join(dir, key + '.png') })
    }
    expect(detail.count('/games/82/insights/players', { range: '30d' })).toBe(1)
    expect(detail.count('/games/82/insights/prices', { range: '30d', region: 'CN' })).toBe(1)
    detail.assertQuiet()
  })
}

test('populated Facts preserve zero, priced zero, unknown, missing and platform states', async ({ detail }) => {
  await detail.open(); await detail.tab('insights')
  await expect(detail.page.locator('[data-current-players]')).toHaveText('0 人')
  await expect(detail.page.locator('[data-price-kind="priced"]')).toContainText('¥0.00')
  await expect(detail.page.locator('[data-price-kind="unknown"]')).toHaveCount(1)
  await expect(detail.page.locator('[data-price-kind="missing"]')).toHaveCount(1)
  await expect(detail.page.locator('[data-observed-low]')).toContainText('GoFurry 观测低价')
  await expect(detail.page.locator('.game-insights-platform-list')).toContainText('macOS')
  await expect(detail.page.locator('.game-insights-platform-list__state--unknown')).toHaveCount(1)
  await expect(detail.page.locator('[data-player-loaded-ranges]')).toHaveAttribute('data-player-loaded-ranges', '30d')
  detail.assertQuiet()
})

test('region and range caches survive tab switches without redundant history calls', async ({ detail }) => {
  await detail.open(); expect(detail.count('/games/82/insights/players')).toBe(0); await detail.tab('insights')
  const root = detail.page.locator('[data-game-insights]')
  await expect(detail.page.locator('[data-price-region-summary]')).toHaveCount(3)
  await expect(root).toHaveAttribute('data-price-loaded-ranges', 'CN:30d')
  await detail.page.getByRole('button', { name: '美国', exact: true }).click()
  await expect(root).toHaveAttribute('data-price-loaded-ranges', /US:30d/)
  await detail.page.getByRole('button', { name: '香港', exact: true }).click()
  await expect(root).toHaveAttribute('data-price-loaded-ranges', /HK:30d/)
  await detail.page.locator('[data-game-insights-range="90d"]').click()
  await expect(root).toHaveAttribute('data-player-loaded-ranges', '30d,90d')
  await detail.page.locator('[data-game-insights-range="30d"]').click()
  await detail.page.getByRole('button', { name: '中国', exact: true }).click()
  await detail.tab('intro'); await detail.tab('insights')
  expect(detail.count('/games/82/insights/players', { range: '30d' })).toBe(1)
  expect(detail.count('/games/82/insights/prices', { range: '30d', region: 'CN' })).toBe(1)
  expect(detail.count('/games/82/insights/prices', { range: '30d', region: 'US' })).toBe(1)
  expect(detail.count('/games/82/insights/prices', { range: '30d', region: 'HK' })).toBe(1)
  await expect(root).toHaveAttribute('data-price-region', 'CN'); detail.assertQuiet()
})

test('independent history failure, retry and out-of-order responses preserve the selected state', async ({ detail }) => {
  const recover = detail.fail('/games/82/insights/players', { range: '30d' })
  await detail.open(); await detail.tab('insights')
  const player = detail.page.locator('[data-player-history]'), price = detail.page.locator('[data-price-history]')
  await expect(player.getByRole('button')).toBeVisible(); await expect(price.locator('canvas')).toBeVisible()
  recover(); detail.points(1); await player.getByRole('button').click()
  await expect(player.locator('.game-insights-chart-state')).toBeVisible()
  const held = detail.hold('/games/82/insights/players', { range: '90d' }); detail.points(0)
  await detail.page.locator('[data-game-insights-range="90d"]').click(); await held.wait()
  await detail.page.locator('[data-game-insights-range="180d"]').click()
  await expect(detail.page.locator('[data-player-loaded-ranges]')).toHaveAttribute('data-player-loaded-ranges', /180d/)
  held.release(); await expect.poll(() => held.completed).toBe(true)
  await expect(detail.page.locator('[data-game-insights-range="180d"]')).toHaveAttribute('aria-pressed', 'true')
  await expect(player.locator('.game-insights-chart-state')).toBeVisible()
  const restorePrice = detail.fail('/games/82/insights/prices', { region: 'US', range: '180d' })
  await detail.page.getByRole('button', { name: '美国', exact: true }).click()
  await expect(price.getByRole('button')).toBeVisible()
  const playerReads = detail.count('/games/82/insights/players')
  restorePrice(); detail.points(8); await price.getByRole('button').click()
  await expect(price.locator('canvas')).toBeVisible()
  expect(detail.count('/games/82/insights/players')).toBe(playerReads)
  detail.assertQuiet()
})

test('Timeline and real chart themes retain date precision, ordering and responsive canvas', async ({ detail }) => {
  await detail.open(); await detail.tab('insights')
  const timeline = detail.page.locator('[data-entity-timeline]')
  await expect(timeline).toHaveAttribute('data-timeline-mode', 'compact')
  const before = await timeline.locator('time').allTextContents()
  expect(before[0]).toBe('2026-09-18'); expect(before[1]).toContain(':')
  await timeline.locator('button').last().click(); await expect(timeline).toHaveAttribute('data-timeline-mode', 'list')
  await expect(timeline.locator('button').last()).toHaveAttribute('aria-pressed', 'true')
  await expect(timeline.locator('button').last()).toHaveClass(/insights-ranges__button--active/)
  expect(await timeline.locator('time').allTextContents()).toEqual(before)
  const canvas = detail.page.locator('[data-player-history] canvas')
  await expect(canvas).toBeVisible(); await detail.settle(detail.page.locator('[data-player-history]'))
  const tooltipAppearance = async (kind: 'player' | 'price', dark: boolean) => {
    const root = detail.page.locator(`[data-${kind}-history]`), target = root.locator('canvas')
    await target.hover({ position: { x: (await target.boundingBox())!.width / 2, y: 150 } })
    const tooltip = root.locator('.game-insights-chart > div').last()
    await expect(tooltip).toContainText('2026-09')
    await expect(tooltip).toHaveCSS('background-color', dark ? 'rgba(15, 23, 42, 0.96)' : 'rgba(255, 250, 242, 0.98)')
    await expect(tooltip).toHaveCSS('color', dark ? 'rgb(226, 232, 240)' : 'rgb(41, 37, 36)')
    await expect(tooltip).toHaveCSS('border-color', kind === 'player'
      ? dark ? 'rgba(125, 211, 252, 0.28)' : 'rgba(217, 119, 6, 0.22)'
      : dark ? 'rgba(167, 243, 208, 0.28)' : 'rgba(21, 128, 61, 0.22)')
    await detail.page.mouse.move(1, 1); await expect(tooltip).toHaveCSS('visibility', 'hidden')
  }
  await tooltipAppearance('player', false); await tooltipAppearance('price', false)
  const light = await canvas.evaluate(el => (el as HTMLCanvasElement).toDataURL())
  const priceCanvas = detail.page.locator('[data-price-history] canvas')
  const lightPrice = await priceCanvas.evaluate(el => (el as HTMLCanvasElement).toDataURL())
  await detail.page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await expect(detail.page.locator('html')).toHaveClass(/dark/)
  await expect.poll(() => canvas.evaluate(el => (el as HTMLCanvasElement).toDataURL())).not.toBe(light)
  await expect.poll(() => priceCanvas.evaluate(el => (el as HTMLCanvasElement).toDataURL())).not.toBe(lightPrice)
  const width = (await canvas.boundingBox())!.width
  await detail.page.setViewportSize({ width: 700, height: 900 })
  await expect.poll(async () => (await canvas.boundingBox())!.width).not.toBe(width)
  await detail.settle(detail.page.locator('[data-player-history]'))
  await tooltipAppearance('player', true); await tooltipAppearance('price', true)
  detail.assertQuiet()
})


for (const [value, expected] of [[35386.394, '35,386.4'], [42.5, '42.5'], [42, '42']] as const) test('Task D 30-day average precision ' + value, async ({ detail }) => {
  detail.average(value)
  await detail.open(); await detail.tab('insights')
  await expect(detail.page.locator('[data-average-players]')).toHaveText(expected + ' 人')
  await expect(detail.page.locator('[data-current-players]')).toHaveText('0 人')
  await expect(detail.page.locator('.game-insights-overview__metrics dd').first()).toHaveText('120 人')
  detail.assertQuiet()
})

test('Task D real player and price chart tooltips retain missing evidence as a dash', async ({ detail }) => {
  detail.missingChartEvidence()
  await detail.open(); await detail.tab('insights')
  for (const [kind, left] of [['player', 58], ['price', 62]] as const) {
    const root = detail.page.locator(`[data-${kind}-history]`), canvas = root.locator('canvas')
    await expect(canvas).toBeVisible(); await detail.settle(root)
    const width = (await canvas.boundingBox())!.width
    await canvas.hover({ position: { x: left + (width - left - 18) * 2 / 7, y: 150 } })
    const tooltip = root.locator('.game-insights-chart > div').last()
    await expect(tooltip).toContainText('2026-09-03')
    await expect(tooltip).toContainText('—')
    await expect(tooltip).not.toContainText(/undefined|null|NaN|Infinity/)
    expect((await tooltip.textContent())!.match(/—/g)).toHaveLength(kind === 'player' ? 2 : 1)
    await detail.page.mouse.move(0, 0)
  }
  detail.assertQuiet()
})
