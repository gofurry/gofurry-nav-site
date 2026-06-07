import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
export const rootDir = path.resolve(__dirname, '../..')
export const performanceDir = path.join(rootDir, 'docs/performance')
export const reportsDir = path.join(performanceDir, 'reports')
export const baselinePath = path.join(performanceDir, 'baseline.json')
export const budgetPath = path.join(performanceDir, 'budget.json')

const defaultViewport = { width: 1440, height: 900 }
const defaultSitePath = process.env.PERF_SITE_PATH || '/sites/1'

export const scenarios = [
  {
    id: 'home-initial',
    label: '首页首屏',
    path: '/',
    critical: true,
    blockedHeavyDependencies: ['md-editor-v3', 'echarts', 'hls.js']
  },
  {
    id: 'home-revealed',
    label: '首页内容区',
    path: '/',
    critical: true,
    action: 'reveal-home',
    waitAfterActionMs: 1800
  },
  {
    id: 'updates',
    label: '更新页',
    path: '/updates',
    critical: true
  },
  {
    id: 'about',
    label: '关于页',
    path: '/about',
    critical: true
  },
  {
    id: 'archive-empty',
    label: '知识库空页',
    path: '/archive',
    critical: true,
    blockedHeavyDependencies: ['md-editor-v3']
  },
  {
    id: 'games',
    label: '游戏首页',
    path: '/games',
    critical: true,
    blockedHeavyDependencies: ['hls.js']
  },
  {
    id: 'site-detail',
    label: '站点详情页',
    path: defaultSitePath,
    critical: false
  }
]

export const heavyDependencyPatterns = {
  'md-editor-v3': {
    url: [/md-editor-v3/i, /md-editor/i, /preview\.[\w-]+\.css/i],
    body: []
  },
  echarts: {
    url: [/echarts/i, /zrender/i],
    body: [/zrender/i, /apache echarts/i, /seriesType/i]
  },
  'hls.js': {
    url: [/hls(\.min)?\.js/i, /hls-js/i],
    body: [/LEVEL_LOADED/, /FRAG_LOADED/, /BUFFER_APPENDING/, /hlsDefaultConfig/i]
  }
}

export function parseArgs(argv = process.argv.slice(2)) {
  const args = {}

  for (let index = 0; index < argv.length; index += 1) {
    const item = argv[index]
    if (!item.startsWith('--')) {
      continue
    }

    const [rawKey, inlineValue] = item.slice(2).split('=')
    if (inlineValue !== undefined) {
      args[rawKey] = inlineValue
      continue
    }

    const next = argv[index + 1]
    if (next && !next.startsWith('--')) {
      args[rawKey] = next
      index += 1
    } else {
      args[rawKey] = true
    }
  }

  return args
}

export function normalizeBaseUrl(baseUrl = 'http://localhost:3000') {
  return String(baseUrl).replace(/\/$/, '')
}

export function toAbsoluteUrl(baseUrl, pathname) {
  if (/^https?:\/\//i.test(pathname)) {
    return pathname
  }

  return `${normalizeBaseUrl(baseUrl)}${pathname.startsWith('/') ? pathname : `/${pathname}`}`
}

export function bytesToKb(bytes) {
  return Math.round((bytes / 1024) * 10) / 10
}

export function bytesToMb(bytes) {
  return Math.round((bytes / 1024 / 1024) * 10) / 10
}

export function roundMs(value) {
  return Math.round(Number(value || 0) * 10) / 10
}

export async function launchPerfBrowser() {
  const baseLaunchOptions = {
    headless: true,
    args: [
      '--disable-background-timer-throttling',
      '--disable-renderer-backgrounding',
      '--disable-backgrounding-occluded-windows'
    ]
  }
  const channels = [
    process.env.PERF_BROWSER_CHANNEL,
    undefined,
    'chrome',
    'msedge'
  ].filter((channel, index, list) => list.indexOf(channel) === index)
  const errors = []

  for (const channel of channels) {
    const launchOptions = { ...baseLaunchOptions }
    if (channel) {
      launchOptions.channel = channel
    }

    try {
      return await chromium.launch(launchOptions)
    } catch (error) {
      errors.push(`${channel || 'playwright chromium'}: ${error.message.split('\n')[0]}`)
    }
  }

  throw new Error(`无法启动浏览器。请设置 PERF_BROWSER_CHANNEL=chrome，或运行 npx playwright install chromium。\n${errors.join('\n')}`)
}

function parseContentLength(headers) {
  const value = headers['content-length']
  if (!value) {
    return 0
  }

  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
}

function isScriptResource(resource) {
  return resource.resourceType === 'script'
    || /\.m?js(\?|$)/i.test(resource.url)
    || resource.initiatorType === 'script'
}

function isImageResource(resource) {
  return resource.resourceType === 'image'
    || resource.initiatorType === 'img'
    || /\.(png|jpe?g|gif|webp|avif|svg)(\?|$)/i.test(resource.url)
}

function detectHeavyDependencies(scriptBodies, resources) {
  const hits = {}
  const resourceText = resources.map((resource) => resource.url).join('\n')
  const scriptText = scriptBodies.map((script) => `${script.url}\n${script.text}`).join('\n')

  for (const [name, patterns] of Object.entries(heavyDependencyPatterns)) {
    const matched = patterns.url.some((pattern) => pattern.test(resourceText))
      || (patterns.body || []).some((pattern) => pattern.test(scriptText))
    hits[name] = matched
  }

  return hits
}

function mergePerformanceEntries(resources, entries) {
  const byUrl = new Map(resources.map((resource) => [resource.url, resource]))

  for (const entry of entries) {
    const current = byUrl.get(entry.name) || {
      url: entry.name,
      resourceType: '',
      status: 0,
      transferSize: 0,
      decodedBodySize: 0,
      encodedBodySize: 0
    }

    current.initiatorType = entry.initiatorType || current.initiatorType || ''
    current.transferSize = Math.max(current.transferSize || 0, entry.transferSize || 0)
    current.decodedBodySize = Math.max(current.decodedBodySize || 0, entry.decodedBodySize || 0)
    current.encodedBodySize = Math.max(current.encodedBodySize || 0, entry.encodedBodySize || 0)
    byUrl.set(entry.name, current)
  }

  return [...byUrl.values()]
}

function summarizeMetrics(resources, scriptBodies, pageMetrics, cdpMetrics, longTasks, warnings) {
  const scripts = resources.filter(isScriptResource)
  const images = resources.filter(isImageResource)
  const jsTransferBytes = scripts.reduce((sum, resource) => {
    const scriptBody = scriptBodies.find((script) => script.url === resource.url)
    const fallbackSize = scriptBody?.bytes || 0
    return sum + Math.max(resource.transferSize || 0, resource.encodedBodySize || 0, fallbackSize)
  }, 0)
  const imageTransferBytes = images.reduce((sum, resource) => {
    return sum + Math.max(resource.transferSize || 0, resource.encodedBodySize || 0)
  }, 0)
  const longTaskTotalMs = longTasks.reduce((sum, task) => sum + Number(task.duration || 0), 0)

  return {
    domNodes: pageMetrics.domNodes,
    imagesInDom: pageMetrics.imagesInDom,
    jsRequests: scripts.length,
    jsTransferBytes,
    jsTransferKb: bytesToKb(jsTransferBytes),
    imageRequests: images.length,
    imageTransferBytes,
    imageTransferKb: bytesToKb(imageTransferBytes),
    longTaskCount: longTasks.length,
    longTaskTotalMs: roundMs(longTaskTotalMs),
    jsHeapUsedBytes: cdpMetrics.JSHeapUsedSize || 0,
    jsHeapUsedMb: bytesToMb(cdpMetrics.JSHeapUsedSize || 0),
    cdp: cdpMetrics,
    heavyDependencies: detectHeavyDependencies(scriptBodies, resources),
    warnings
  }
}

async function installLongTaskObserver(page) {
  await page.addInitScript(() => {
    window.__gofurryPerfLongTasks = []

    try {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          window.__gofurryPerfLongTasks.push({
            name: entry.name,
            startTime: entry.startTime,
            duration: entry.duration
          })
        }
      })

      observer.observe({ type: 'longtask', buffered: true })
    } catch {
      window.__gofurryPerfLongTasksUnsupported = true
    }
  })
}

async function performScenarioAction(page, scenario) {
  if (scenario.action !== 'reveal-home') {
    return
  }

  await page.mouse.wheel(0, 900)
  await page.waitForTimeout(scenario.waitAfterActionMs || 1500)
}

async function collectScenario(browser, baseUrl, scenario, options = {}) {
  const warnings = []
  const resources = []
  const scriptBodyPromises = []
  const context = await browser.newContext({
    viewport: options.viewport || defaultViewport,
    deviceScaleFactor: 1,
    reducedMotion: 'reduce',
    locale: 'zh-CN'
  })
  const page = await context.newPage()
  const client = await context.newCDPSession(page)

  await installLongTaskObserver(page)
  await client.send('Performance.enable')

  page.on('response', (response) => {
    const request = response.request()
    const url = response.url()
    const resourceType = request.resourceType()
    const headers = response.headers()
    const resource = {
      url,
      status: response.status(),
      resourceType,
      transferSize: parseContentLength(headers),
      contentType: headers['content-type'] || ''
    }

    resources.push(resource)

    if (resourceType === 'script' || /\.m?js(\?|$)/i.test(url)) {
      scriptBodyPromises.push(
        response.text()
          .then((text) => ({ url, text, bytes: Buffer.byteLength(text) }))
          .catch(() => ({ url, text: '', bytes: 0 }))
      )
    }
  })

  const startedAt = Date.now()
  const url = toAbsoluteUrl(baseUrl, scenario.path)
  let status = 0

  try {
    const response = await page.goto(url, {
      waitUntil: 'domcontentloaded',
      timeout: options.timeoutMs || 30000
    })
    status = response?.status() || 0

    if (status >= 400) {
      warnings.push(`页面返回 HTTP ${status}`)
    }

    await page.waitForLoadState('networkidle', { timeout: options.networkIdleTimeoutMs || 10000 })
      .catch(() => warnings.push('networkidle 等待超时，已按当前页面状态采集'))

    await page.waitForTimeout(options.settleMs || 1200)
    await performScenarioAction(page, scenario)
  } catch (error) {
    warnings.push(`页面打开失败：${error.message}`)
  }

  const scriptBodies = (await Promise.allSettled(scriptBodyPromises))
    .filter((result) => result.status === 'fulfilled')
    .map((result) => result.value)

  const performanceEntries = await page.evaluate(() => {
    return performance.getEntriesByType('resource').map((entry) => ({
      name: entry.name,
      initiatorType: entry.initiatorType,
      transferSize: entry.transferSize || 0,
      encodedBodySize: entry.encodedBodySize || 0,
      decodedBodySize: entry.decodedBodySize || 0
    }))
  }).catch(() => [])
  const mergedResources = mergePerformanceEntries(resources, performanceEntries)

  const pageMetrics = await page.evaluate(() => ({
    domNodes: document.querySelectorAll('*').length,
    imagesInDom: document.querySelectorAll('img').length
  })).catch(() => ({ domNodes: 0, imagesInDom: 0 }))

  const longTasks = await page.evaluate(() => window.__gofurryPerfLongTasks || []).catch(() => [])
  const cdpMetricList = await client.send('Performance.getMetrics').catch(() => ({ metrics: [] }))
  const cdpMetrics = Object.fromEntries((cdpMetricList.metrics || []).map((metric) => [metric.name, metric.value]))
  const metrics = summarizeMetrics(mergedResources, scriptBodies, pageMetrics, cdpMetrics, longTasks, warnings)

  await context.close()

  return {
    id: scenario.id,
    label: scenario.label,
    path: scenario.path,
    url,
    critical: scenario.critical,
    status,
    durationMs: Date.now() - startedAt,
    metrics,
    resources: mergedResources.map((resource) => ({
      url: resource.url,
      resourceType: resource.resourceType || resource.initiatorType || '',
      status: resource.status || 0,
      transferSize: resource.transferSize || 0,
      encodedBodySize: resource.encodedBodySize || 0
    }))
  }
}

export async function runMeasurement(options = {}) {
  const baseUrl = normalizeBaseUrl(options.baseUrl || 'http://localhost:3000')
  const browser = await launchPerfBrowser()
  const results = []

  try {
    for (const scenario of scenarios) {
      console.log(`[perf] 测量 ${scenario.label} ${scenario.path}`)
      results.push(await collectScenario(browser, baseUrl, scenario, options))
    }
  } finally {
    await browser.close()
  }

  return {
    generatedAt: new Date().toISOString(),
    baseUrl,
    viewport: options.viewport || defaultViewport,
    scenarios: results
  }
}

export async function loadJson(filePath, fallback = null) {
  try {
    return JSON.parse(await readFile(filePath, 'utf8'))
  } catch {
    return fallback
  }
}

export async function writeJson(filePath, data) {
  await mkdir(path.dirname(filePath), { recursive: true })
  await writeFile(filePath, `${JSON.stringify(data, null, 2)}\n`, 'utf8')
}

export function makeReportName(prefix = 'measure') {
  const stamp = new Date().toISOString().replace(/[:.]/g, '-')
  return `${prefix}-${stamp}`
}

export async function writeMeasurementReports(measurement, prefix = 'measure') {
  const name = makeReportName(prefix)
  const jsonPath = path.join(reportsDir, `${name}.json`)
  const markdownPath = path.join(reportsDir, `${name}.md`)

  await mkdir(reportsDir, { recursive: true })
  await writeJson(jsonPath, measurement)
  await writeFile(markdownPath, renderMarkdownReport(measurement), 'utf8')

  return { jsonPath, markdownPath }
}

export function compactMeasurement(measurement) {
  return {
    generatedAt: measurement.generatedAt,
    baseUrl: measurement.baseUrl,
    viewport: measurement.viewport,
    scenarios: measurement.scenarios.map((scenario) => ({
      id: scenario.id,
      label: scenario.label,
      path: scenario.path,
      critical: scenario.critical,
      status: scenario.status,
      durationMs: scenario.durationMs,
      metrics: {
        domNodes: scenario.metrics.domNodes,
        imagesInDom: scenario.metrics.imagesInDom,
        jsRequests: scenario.metrics.jsRequests,
        jsTransferKb: scenario.metrics.jsTransferKb,
        imageRequests: scenario.metrics.imageRequests,
        imageTransferKb: scenario.metrics.imageTransferKb,
        longTaskCount: scenario.metrics.longTaskCount,
        longTaskTotalMs: scenario.metrics.longTaskTotalMs,
        jsHeapUsedMb: scenario.metrics.jsHeapUsedMb,
        heavyDependencies: scenario.metrics.heavyDependencies,
        warnings: scenario.metrics.warnings
      }
    }))
  }
}

export function renderMarkdownReport(measurement, guardResult = null) {
  const lines = [
    '# GoFurry Nav Web 前端性能测量报告',
    '',
    `- 生成时间：${measurement.generatedAt}`,
    `- 基准地址：${measurement.baseUrl}`,
    `- 视口：${measurement.viewport.width}x${measurement.viewport.height}`,
    '',
    '## 指标摘要',
    '',
    '| 场景 | DOM | JS 请求 | JS KB | 图片请求 | 图片 KB | Long Task | Long Task ms | JS Heap MB | 重型依赖 | 警告 |',
    '| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |'
  ]

  for (const scenario of measurement.scenarios) {
    const metrics = scenario.metrics
    const heavy = Object.entries(metrics.heavyDependencies)
      .filter(([, matched]) => matched)
      .map(([name]) => name)
      .join(', ') || '-'
    const warnings = metrics.warnings.length ? metrics.warnings.join('; ') : '-'

    lines.push([
      scenario.label,
      metrics.domNodes,
      metrics.jsRequests,
      metrics.jsTransferKb,
      metrics.imageRequests,
      metrics.imageTransferKb,
      metrics.longTaskCount,
      metrics.longTaskTotalMs,
      metrics.jsHeapUsedMb,
      heavy,
      warnings
    ].join(' | ').replace(/^/, '| ').replace(/$/, ' |'))
  }

  if (guardResult) {
    lines.push('', '## 回归守卫结果', '')

    if (!guardResult.failures.length) {
      lines.push('结果：通过。')
    } else {
      lines.push(`结果：失败，共 ${guardResult.failures.length} 项。`, '')
      for (const failure of guardResult.failures) {
        lines.push(`- ${failure.scenarioLabel}：${failure.message}`)
      }
    }

    if (guardResult.warnings.length) {
      lines.push('', '### 警告', '')
      for (const warning of guardResult.warnings) {
        lines.push(`- ${warning.scenarioLabel}：${warning.message}`)
      }
    }
  }

  lines.push(
    '',
    '## 说明',
    '',
    '- GPU 占用不做自动硬阈值；如需排查 paint/composite，请运行 `npm run perf:trace`。',
    '- JS transfer size 优先使用 PerformanceResourceTiming，缺失时使用响应头或脚本文本大小兜底。',
    '- 动态详情页依赖本地后端数据，失败会进入 warning，不会阻断核心页面测量。'
  )

  return `${lines.join('\n')}\n`
}

export function evaluateBudget(measurement, budget) {
  const failures = []
  const warnings = []
  const globalBudget = budget.global || {}
  const scenarioBudgets = budget.scenarios || {}

  for (const scenario of measurement.scenarios) {
    const scenarioBudget = {
      ...globalBudget,
      ...(scenarioBudgets[scenario.id] || {})
    }
    const metrics = scenario.metrics
    const pushIssue = (collection, message) => {
      collection.push({
        scenarioId: scenario.id,
        scenarioLabel: scenario.label,
        message
      })
    }
    if (metrics.warnings.length) {
      for (const warning of metrics.warnings) {
        const isBlockingWarning = /页面打开失败|HTTP\s+[45]\d\d/.test(warning)
        pushIssue(scenario.critical && isBlockingWarning ? failures : warnings, warning)
      }
    }

    const checks = [
      ['maxDomNodes', metrics.domNodes, 'DOM 节点数'],
      ['maxJsRequests', metrics.jsRequests, 'JS 请求数'],
      ['maxJsTransferKb', metrics.jsTransferKb, 'JS 传输 KB'],
      ['maxImageRequests', metrics.imageRequests, '图片请求数'],
      ['maxLongTaskCount', metrics.longTaskCount, 'Long Task 数量'],
      ['maxLongTaskTotalMs', metrics.longTaskTotalMs, 'Long Task 总耗时 ms'],
      ['maxJsHeapUsedMb', metrics.jsHeapUsedMb, 'JS Heap MB']
    ]

    for (const [budgetKey, actual, label] of checks) {
      const limit = scenarioBudget[budgetKey]
      if (typeof limit === 'number' && actual > limit) {
        pushIssue(scenario.critical ? failures : warnings, `${label} ${actual} 超过预算 ${limit}`)
      }
    }

    const blockedDependencies = [
      ...(scenario.blockedHeavyDependencies || []),
      ...(scenarioBudget.blockedHeavyDependencies || [])
    ]
    for (const dependency of new Set(blockedDependencies)) {
      if (metrics.heavyDependencies[dependency]) {
        pushIssue(scenario.critical ? failures : warnings, `不应加载重型依赖 ${dependency}`)
      }
    }
  }

  return { failures, warnings }
}
