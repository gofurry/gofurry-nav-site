#!/usr/bin/env node
import { mkdir, writeFile } from 'node:fs/promises'
import path from 'node:path'
import {
  launchPerfBrowser,
  normalizeBaseUrl,
  parseArgs,
  reportsDir,
  toAbsoluteUrl
} from './shared.mjs'

const args = parseArgs()
const baseUrl = normalizeBaseUrl(args['base-url'] || process.env.PERF_BASE_URL || 'http://localhost:3000')
const outputName = `visual-${new Date().toISOString().replace(/[:.]/g, '-')}`
const outputDir = path.join(reportsDir, outputName)

const desktop = { width: 1440, height: 900 }
const mobile = { width: 390, height: 844 }
const legacyDarkSelectors = [
  '.games-page--dark',
  '.search-results--dark',
  '.is-dark-theme',
  '.spotlight-panels--dark'
]

const visualScenarios = [
  ...makePageScenarios({
    id: 'nav-home',
    label: '首页导航',
    path: '/',
    locale: 'zh-CN',
    rootSelector: '.nav-home-page',
    requiredSelectors: ['.nav-home-page', '.nav-header', '.search-box-shell'],
    optionalDataSelectors: ['.nav-content', '.nav-site-card', '.spotlight-panel'],
    action: 'reveal-home-nav'
  }),
  ...makePageScenarios({
    id: 'nav-home-en',
    label: '首页导航英文',
    path: '/en',
    locale: 'en-US',
    rootSelector: '.nav-home-page',
    requiredSelectors: ['.nav-home-page', '.nav-header', '.search-box-shell'],
    optionalDataSelectors: ['.nav-content', '.nav-site-card', '.spotlight-panel'],
    action: 'reveal-home-nav'
  }),
  ...makePageScenarios({
    id: 'games',
    label: '游戏首页',
    path: '/games',
    locale: 'zh-CN',
    rootSelector: '.games-page',
    requiredSelectors: ['.games-page', '.game-info-shell', '.game-sidebar-shell'],
    optionalDataSelectors: ['.game-card']
  }),
  ...makePageScenarios({
    id: 'games-search',
    label: '游戏搜索页',
    path: '/games/search',
    locale: 'zh-CN',
    rootSelector: '.games-search-page',
    requiredSelectors: ['.games-search-page', '.search-result-grid', '.gf-pagination'],
    optionalDataSelectors: ['.search-page-card']
  }),
  ...makePageScenarios({
    id: 'games-search-en',
    label: '游戏搜索页英文',
    path: '/en/games/search',
    locale: 'en-US',
    rootSelector: '.games-search-page',
    requiredSelectors: ['.games-search-page', '.search-result-grid', '.gf-pagination'],
    optionalDataSelectors: ['.search-page-card']
  })
]

function makePageScenarios(config) {
  return [
    {
      ...config,
      id: `${config.id}-light-desktop`,
      theme: 'light',
      viewport: desktop
    },
    {
      ...config,
      id: `${config.id}-dark-desktop`,
      theme: 'dark',
      viewport: desktop
    },
    {
      ...config,
      id: `${config.id}-light-mobile`,
      theme: 'light',
      viewport: mobile
    },
    {
      ...config,
      id: `${config.id}-dark-mobile`,
      theme: 'dark',
      viewport: mobile
    }
  ]
}

function renderMarkdownReport(report) {
  const lines = [
    '# GoFurry Nav Web 视觉截图守卫',
    '',
    `- 生成时间：${report.generatedAt}`,
    `- 基准地址：${report.baseUrl}`,
    `- 输出目录：${report.outputDir}`,
    '',
    '## 截图清单',
    '',
    '| 场景 | 路径 | 主题 | 视口 | 截图 | 状态 | 说明 |',
    '| --- | --- | --- | --- | --- | --- | --- |'
  ]

  for (const scenario of report.scenarios) {
    const status = scenario.failures.length ? '失败' : '通过'
    const notes = [...scenario.failures, ...scenario.warnings]
      .map((item) => item.message)
      .join('<br>') || '-'
    lines.push([
      scenario.label,
      scenario.path,
      scenario.theme,
      `${scenario.viewport.width}x${scenario.viewport.height}`,
      scenario.screenshot,
      status,
      notes
    ].join(' | ').replace(/^/, '| ').replace(/$/, ' |'))
  }

  lines.push(
    '',
    '## 检查项',
    '',
    '- `html.dark` 与截图主题一致。',
    '- `/`、`/en`、`/games`、`/games/search`、`/en/games/search` 的关键容器存在。',
    '- 不再出现 `games-page--dark`、`search-results--dark`、`is-dark-theme`、`spotlight-panels--dark` 等旧暗色入口。',
    '- 桌面与移动端不出现明显横向溢出。',
    '- 数据卡片为空只记为 warning，方便在本地后端无数据时继续保留截图。'
  )

  return `${lines.join('\n')}\n`
}

async function inspectPage(page, scenario) {
  return await page.evaluate(({ rootSelector, requiredSelectors, optionalDataSelectors, theme, legacySelectors }) => {
    const documentElement = document.documentElement
    const body = document.body
    const root = document.querySelector(rootSelector)
    const rootBox = root?.getBoundingClientRect()
    const horizontalOverflow = Math.max(documentElement.scrollWidth, body.scrollWidth) - documentElement.clientWidth
    const optionalCounts = Object.fromEntries(
      optionalDataSelectors.map((selector) => [selector, document.querySelectorAll(selector).length])
    )

    return {
      expectedTheme: theme,
      actualTheme: documentElement.classList.contains('dark') ? 'dark' : 'light',
      title: document.title,
      rootBox: rootBox
        ? {
            width: Math.round(rootBox.width),
            height: Math.round(rootBox.height)
          }
        : null,
      missingRequired: requiredSelectors.filter((selector) => !document.querySelector(selector)),
      optionalCounts,
      legacyDarkCount: legacySelectors.reduce(
        (count, selector) => count + document.querySelectorAll(selector).length,
        0
      ),
      activeNavCount: document.querySelectorAll('.gf-nav__link--active, .gf-nav__mobile-link--active').length,
      horizontalOverflow,
      bodyTextLength: body.innerText.trim().length
    }
  }, {
    rootSelector: scenario.rootSelector,
    requiredSelectors: scenario.requiredSelectors,
    optionalDataSelectors: scenario.optionalDataSelectors,
    theme: scenario.theme,
    legacySelectors: legacyDarkSelectors
  })
}

function evaluateSnapshot(snapshot, scenario) {
  const failures = []
  const warnings = []
  const pushFailure = (message) => failures.push({ message })
  const pushWarning = (message) => warnings.push({ message })

  if (snapshot.actualTheme !== scenario.theme) {
    pushFailure(`主题不一致：期望 ${scenario.theme}，实际 ${snapshot.actualTheme}`)
  }

  if (!snapshot.rootBox || snapshot.rootBox.width < 1 || snapshot.rootBox.height < 1) {
    pushFailure(`根容器 ${scenario.rootSelector} 不存在或尺寸异常`)
  }

  if (snapshot.missingRequired.length) {
    pushFailure(`缺少关键选择器：${snapshot.missingRequired.join(', ')}`)
  }

  if (snapshot.legacyDarkCount > 0) {
    pushFailure(`检测到 ${snapshot.legacyDarkCount} 个旧暗色入口节点`)
  }

  if (snapshot.horizontalOverflow > 2) {
    pushFailure(`页面横向溢出 ${Math.round(snapshot.horizontalOverflow)}px`)
  }

  if (snapshot.activeNavCount < 1) {
    pushWarning('未检测到导航 active 状态')
  }

  if (snapshot.bodyTextLength < 20) {
    pushWarning('页面文本内容过少，可能仍在加载或接口无数据')
  }

  for (const [selector, count] of Object.entries(snapshot.optionalCounts)) {
    if (count < 1) {
      pushWarning(`${selector} 数据卡片为空`)
    }
  }

  return { failures, warnings }
}

async function performScenarioAction(page, scenario) {
  if (scenario.action !== 'reveal-home-nav') {
    return
  }

  await page.mouse.wheel(0, 1100)
  await page.waitForSelector('.nav-content', { timeout: 5000 }).catch(() => {})
  await page.waitForTimeout(1800)
}

async function runScenario(browser, scenario) {
  const context = await browser.newContext({
    viewport: scenario.viewport,
    deviceScaleFactor: 1,
    reducedMotion: 'reduce',
    locale: scenario.locale
  })
  await context.addInitScript((theme) => {
    try {
      localStorage.setItem('theme', theme)
    } catch {
      // Ignore storage access errors on transient browser origins.
    }
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, scenario.theme)

  const page = await context.newPage()
  const warnings = []
  const url = toAbsoluteUrl(baseUrl, scenario.path)
  const screenshotName = `${scenario.id}.png`
  const screenshotPath = path.join(outputDir, screenshotName)
  let status = 0
  let snapshot = null

  try {
    const response = await page.goto(url, {
      waitUntil: 'domcontentloaded',
      timeout: 30000
    })
    status = response?.status() || 0

    if (status >= 400) {
      warnings.push({ message: `页面返回 HTTP ${status}` })
    }

    await page.waitForLoadState('networkidle', { timeout: 10000 })
      .catch(() => warnings.push({ message: 'networkidle 等待超时，已按当前页面状态截图' }))
    await page.waitForTimeout(1200)
    await performScenarioAction(page, scenario)
    snapshot = await inspectPage(page, scenario)
    await page.screenshot({ path: screenshotPath, fullPage: true })
  } catch (error) {
    warnings.push({ message: `页面截图失败：${error.message}` })
  } finally {
    await context.close()
  }

  const evaluated = snapshot
    ? evaluateSnapshot(snapshot, scenario)
    : { failures: [{ message: '未能采集页面快照' }], warnings: [] }

  return {
    id: scenario.id,
    label: scenario.label,
    path: scenario.path,
    url,
    theme: scenario.theme,
    locale: scenario.locale,
    viewport: scenario.viewport,
    status,
    screenshot: screenshotName,
    snapshot,
    failures: evaluated.failures,
    warnings: [...warnings, ...evaluated.warnings]
  }
}

await mkdir(outputDir, { recursive: true })

const browser = await launchPerfBrowser()
const results = []

try {
  for (const scenario of visualScenarios) {
    console.log(`[visual] 截图 ${scenario.label} ${scenario.path} ${scenario.theme} ${scenario.viewport.width}x${scenario.viewport.height}`)
    results.push(await runScenario(browser, scenario))
  }
} finally {
  await browser.close()
}

const report = {
  generatedAt: new Date().toISOString(),
  baseUrl,
  outputDir,
  scenarios: results
}
const markdown = renderMarkdownReport(report)
const manifestPath = path.join(outputDir, 'manifest.json')
const markdownPath = path.join(outputDir, 'manifest.md')

await writeFile(manifestPath, `${JSON.stringify(report, null, 2)}\n`, 'utf8')
await writeFile(markdownPath, markdown, 'utf8')
await writeFile(path.join(reportsDir, 'latest-visual.md'), markdown, 'utf8')

const failures = results.flatMap((scenario) => (
  scenario.failures.map((failure) => ({ scenario, failure }))
))
const warnings = results.flatMap((scenario) => (
  scenario.warnings.map((warning) => ({ scenario, warning }))
))

console.log(`[visual] 截图目录：${outputDir}`)
console.log(`[visual] Manifest：${manifestPath}`)

if (warnings.length) {
  console.warn(`[visual] 警告 ${warnings.length} 项：`)
  for (const { scenario, warning } of warnings) {
    console.warn(`  - ${scenario.label} ${scenario.theme} ${scenario.viewport.width}x${scenario.viewport.height}: ${warning.message}`)
  }
}

if (failures.length) {
  console.error(`[visual] 视觉守卫失败 ${failures.length} 项：`)
  for (const { scenario, failure } of failures) {
    console.error(`  - ${scenario.label} ${scenario.theme} ${scenario.viewport.width}x${scenario.viewport.height}: ${failure.message}`)
  }
  process.exitCode = 1
} else {
  console.log('[visual] 视觉守卫通过')
}
