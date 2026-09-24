import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/browser/visual',
  fullyParallel: false,
  forbidOnly: Boolean(process.env.CI),
  retries: 0,
  workers: 1,
  timeout: 60_000,
  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-visual-report', open: 'never' }],
  ],
  outputDir: 'visual-test-results',
  // Comparisons must not silently create a missing baseline, including in CI.
  updateSnapshots: 'none',
  snapshotPathTemplate: '{testDir}/__snapshots__/{testFilePath}/{arg}{ext}',
  use: {
    headless: true,
    viewport: { width: 1440, height: 900 },
    locale: 'zh-CN',
    timezoneId: 'UTC',
    deviceScaleFactor: 1,
    colorScheme: 'light',
    contextOptions: { reducedMotion: 'reduce' },
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'off',
  },
  expect: {
    toHaveScreenshot: { animations: 'disabled', caret: 'hide', scale: 'css' },
  },
  projects: [{ name: 'visual-chromium', use: { browserName: 'chromium' } }],
})
