#!/usr/bin/env node
import { mkdir } from 'node:fs/promises'
import path from 'node:path'
import { launchPerfBrowser, parseArgs, reportsDir, toAbsoluteUrl } from './shared.mjs'

const args = parseArgs()
const baseUrl = args['base-url'] || process.env.PERF_BASE_URL || 'http://localhost:3000'
const targetPath = args.path || '/'
const waitMs = Number(args.wait || 5000)
const traceName = `trace-${targetPath.replace(/[^a-z0-9]+/gi, '-').replace(/^-|-$/g, '') || 'home'}-${new Date().toISOString().replace(/[:.]/g, '-')}.zip`
const tracePath = path.join(reportsDir, traceName)

await mkdir(reportsDir, { recursive: true })

const browser = await launchPerfBrowser()
const context = await browser.newContext({
  viewport: { width: 1440, height: 900 },
  deviceScaleFactor: 1,
  reducedMotion: 'reduce',
  locale: 'zh-CN'
})
const page = await context.newPage()

try {
  await context.tracing.start({
    screenshots: true,
    snapshots: true,
    sources: false
  })

  await page.goto(toAbsoluteUrl(baseUrl, targetPath), {
    waitUntil: 'domcontentloaded',
    timeout: 30000
  })
  await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => undefined)

  if (args.reveal) {
    await page.evaluate(() => {
      window.scrollTo({ top: Math.max(window.innerHeight + 80, 900), behavior: 'instant' })
    })
  }

  await page.waitForTimeout(waitMs)
  await context.tracing.stop({ path: tracePath })
  console.log(`[perf] Trace 文件：${tracePath}`)
} finally {
  await context.close()
  await browser.close()
}
