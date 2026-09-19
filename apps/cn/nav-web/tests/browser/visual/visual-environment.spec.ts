import { test, expect } from '@playwright/test'

test('real Chromium matches the visual environment contract', async ({ browser, page }) => {
  expect(browser.browserType().name()).toBe('chromium')
  if (process.env.CI || process.env.GOFURRY_VISUAL_ENV === 'pinned') {
    expect(process.platform).toBe('linux')
    expect(Number(process.versions.node.split('.')[0])).toBe(24)
  }

  await page.setContent('<!doctype html><html lang="zh-CN"><title>Visual environment</title><body>环境检查</body></html>')
  expect(page.viewportSize()).toEqual({ width: 1440, height: 900 })
  expect(await page.evaluate(() => ({
    width: window.innerWidth,
    height: window.innerHeight,
    dpr: window.devicePixelRatio,
    language: navigator.language,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    reducedMotion: matchMedia('(prefers-reduced-motion: reduce)').matches,
    light: matchMedia('(prefers-color-scheme: light)').matches,
  }))).toEqual({
    width: 1440,
    height: 900,
    dpr: 1,
    language: 'zh-CN',
    timezone: 'UTC',
    reducedMotion: true,
    light: true,
  })
})
