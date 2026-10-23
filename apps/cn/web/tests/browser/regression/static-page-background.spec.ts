import { mkdir } from 'node:fs/promises'
import type { Page } from '@playwright/test'
import { test, expect } from '../fixtures/nav-shell'
import { test as gamesTest } from '../fixtures/games-home'
import { test as sitesTest } from '../fixtures/site-groups'
import { test as insightsTest } from '../fixtures/insights-domain'

for (const width of [1440, 1024, 768, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`static page background ignores retired preferences: ${width}px ${theme}`, async ({ page, context, navShell }) => {
    const source = theme === 'light' ? 'server' : 'local'
    await context.addInitScript(source => {
      localStorage.setItem('gf_background_preference', JSON.stringify({ version: 1, source, pattern_id: '1', overrides: { opacity: 1, size_px: 1, color: '#ff0000' } }))
      const opened: string[] = []
      Object.assign(window, { backgroundDatabaseReads: opened })
      const open = indexedDB.open.bind(indexedDB)
      indexedDB.open = (name, version) => { opened.push(name); return open(name, version) }
    }, source)
    await navShell.open({ width, theme })
    const html = await (await page.request.get('/terms')).text()
    expect(html).toContain('data-public-background')
    expect(html).toContain('data-pattern-status="default"')
    expect(html).toContain('data-page-dots')
    const canvas = page.locator('[data-public-background]')
    const fade = page.locator('[data-page-dots]')
    const left = fade.locator('.gf-public-background__dots--left')
    const right = fade.locator('.gf-public-background__dots--right')
    const alpha = width < 768 ? (theme === 'light' ? '0.12' : '0.1')
      : width < 1280 ? (theme === 'light' ? '0.16' : '0.15') : theme === 'light' ? '0.24' : '0.22'
    await expect(canvas).toHaveCount(1)
    await expect(fade).toHaveCount(1)
    await expect(canvas).toHaveCSS('position', 'fixed')
    await expect(canvas).toHaveCSS('background-color', theme === 'light' ? 'rgb(242, 227, 208)' : 'rgb(29, 31, 34)')
    await expect(canvas).toHaveCSS('opacity', '1')
    await expect(canvas).toHaveCSS('pointer-events', 'none')
    await expect(fade).toHaveCSS('position', 'absolute')
    await expect(fade).toHaveCSS('height', width < 768 ? '185px' : '520px')
    await expect(fade).toHaveCSS('pointer-events', 'none')
    await expect(fade).toHaveCSS('mask-image', /linear-gradient/)
    await expect(left).toHaveCSS('background-image', /radial-gradient/)
    expect(await left.evaluate(el => getComputedStyle(el).backgroundImage)).toContain(alpha)
    await expect(left).toHaveCSS('background-repeat', 'repeat')
    await expect(left).toHaveCSS('background-size', width < 768 ? '28px 28px' : '26px 26px')
    await expect(left).toHaveCSS('filter', 'none')
    await expect(fade).toHaveCSS('animation-name', 'none')
    await expect(right).toHaveCSS('display', width < 768 ? 'none' : 'block')
    if (width >= 768) {
      const boxes = await Promise.all([left.boundingBox(), right.boundingBox()])
      expect(boxes[0]!.width).toBeCloseTo(Math.min(width * .22, 340), 1)
      expect(boxes[1]!.width).toBeCloseTo(boxes[0]!.width, 1)
      expect(boxes[1]!.x - boxes[0]!.width).toBeGreaterThanOrEqual(width * .56 - 1)
      await expect(left).toHaveCSS('mask-image', /to right/)
      await expect(right).toHaveCSS('mask-image', /to left/)
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1)
    await page.evaluate(() => {
      window.dispatchEvent(new Event('gf-background-change'))
      window.dispatchEvent(new Event('focus'))
      window.dispatchEvent(new StorageEvent('storage', { key: 'gf_background_preference' }))
    })
    expect(await left.evaluate(el => getComputedStyle(el).backgroundImage)).toContain(alpha)
    expect(await page.evaluate(() => (window as Window & { backgroundDatabaseReads?: string[] }).backgroundDatabaseReads)).toEqual([])
    if (process.env.GOFURRY_149_REVIEW === '1') {
      await mkdir('../../../.cache/149-c1-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/149-c1-review/background-${width}-${theme}.png` })
    }
    await navShell.scrollTo(520)
    expect((await fade.boundingBox())!.y + (await fade.boundingBox())!.height).toBeLessThanOrEqual(0)
    expect((await canvas.boundingBox())!.y).toBe(0)
    await navShell.scrollTo(0)
    await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
    await expect(canvas).toHaveCSS('background-color', theme === 'dark' ? 'rgb(242, 227, 208)' : 'rgb(29, 31, 34)')
    if (width < 768) {
      await navShell.scrollTo(240)
      await navShell.openBottomMode()
    } else await page.locator('.gf-nav').getByRole('button', { name: '模式', exact: true }).click()
    await expect(page.locator('.gf-preferences-modal')).toBeVisible()
    await expect(page.locator('.gf-preferences-modal > .preferences-tabs').getByRole('tab')).toHaveCount(2)
    await page.locator('.gf-preferences-modal').getByRole('button', { name: '取消', exact: true }).click()
    await expect(page.locator('.gf-preferences-modal')).toHaveCount(0)
    navShell.assertQuiet()
  })
}

for (const config of [{ width: 1440, theme: 'dark' }, { width: 390, theme: 'light' }] as const) {
  test(`Home shares the static dot canvas: ${config.width}px ${config.theme}`, async ({ page, navShell }) => {
    await navShell.open({ ...config, route: '/' })
    await expect(page.locator('[data-public-background]')).toHaveCount(1)
    await expect(page.locator('[data-page-dots]')).toHaveCount(1)
    await expect(page.locator('[data-public-background]')).toHaveCSS('background-color', config.theme === 'dark' ? 'rgb(29, 31, 34)' : 'rgb(242, 227, 208)')
    await expect(page.locator('[data-page-dots]')).toHaveCSS('position', 'absolute')
    if (process.env.GOFURRY_149_REVIEW === '1') {
      await mkdir('../../../.cache/149-c1-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/149-c1-review/home-${config.width}-${config.theme}.png` })
    }
    await page.evaluate(() => window.scrollTo({ top: 600, behavior: 'instant' }))
    expect(await page.evaluate(() => scrollY)).toBeGreaterThan(0)
    expect((await page.locator('[data-page-dots]').boundingBox())!.y).toBe(-await page.evaluate(() => scrollY))
    expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1)
    navShell.assertQuiet()
  })
}

// Reuse each page's real API owner; only the shared canvas is asserted here.
async function reviewPageCanvas(page: Page, owner: string) {
  await page.waitForFunction(() => {
    const nuxt = (document.querySelector('#__nuxt') as Element & { __vue_app__?: { config: { globalProperties: { $nuxt?: { isHydrating: boolean } } } } })?.__vue_app__
    return nuxt?.config.globalProperties.$nuxt?.isHydrating === false
  })
  for (const width of [1440, 1024, 768, 390]) {
    await page.setViewportSize({ width, height: width < 768 ? 844 : 900 })
    for (const theme of ['dark', 'light'] as const) {
      await page.evaluate(() => window.scrollTo({ top: 0, behavior: 'instant' }))
      if (await page.locator('html').evaluate(el => el.classList.contains('dark')) !== (theme === 'dark')) {
        await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
      }
      const canvas = page.locator('[data-public-background]'), dots = page.locator('[data-page-dots]')
      await expect(canvas).toHaveCount(1)
      await expect(dots).toHaveCount(1)
      await expect(canvas).toHaveCSS('background-color', theme === 'dark' ? 'rgb(29, 31, 34)' : 'rgb(242, 227, 208)')
      await expect(dots.locator('.gf-public-background__dots--left')).toHaveCSS('background-image', /radial-gradient/)
      expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1)
      if (process.env.GOFURRY_149_REVIEW === '1' && ((width === 1440 && theme === 'dark') || (width === 390 && theme === 'light'))) {
        await mkdir('../../../.cache/149-c1-review', { recursive: true })
        await page.evaluate(() => document.fonts.ready)
        await page.screenshot({ path: `../../../.cache/149-c1-review/${owner}-${width}-${theme}.png` })
      }
      await page.evaluate(() => window.scrollTo({ top: 600, behavior: 'instant' }))
      const scroll = await page.evaluate(() => scrollY)
      expect((await dots.boundingBox())!.y + scroll).toBe(0)
      expect((await canvas.boundingBox())!.y).toBe(0)
    }
  }
}

gamesTest('Games shares the edge-dot and graphite canvas', async ({ page, gamesHome }) => {
  const scene = await gamesHome.open({ theme: 'dark' })
  expect(scene.rendered).toContain('data-page-dots')
  await reviewPageCanvas(page, 'games')
  scene.assertQuiet()
})

sitesTest('Sites shares the edge-dot and graphite canvas', async ({ page, siteGroups }) => {
  const scene = await siteGroups.open({ theme: 'dark', variant: 'empty' })
  expect(scene.ssr).toContain('data-page-dots')
  await reviewPageCanvas(page, 'sites')
  scene.assertQuiet()
})

insightsTest('Insights shares the edge-dot and graphite canvas', async ({ page, runtime }) => {
  const response = await page.goto('/insights/sites', { waitUntil: 'load' })
  expect(response?.status()).toBe(200)
  expect(await response!.text()).toContain('data-page-dots')
  await reviewPageCanvas(page, 'insights')
  runtime.assertQuiet()
})
