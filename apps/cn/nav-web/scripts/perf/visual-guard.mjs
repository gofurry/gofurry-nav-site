#!/usr/bin/env node
import { mkdir, readFile, readdir, writeFile } from 'node:fs/promises'
import path from 'node:path'
import {
  launchPerfBrowser,
  normalizeBaseUrl,
  parseArgs,
  reportsDir,
  rootDir,
  toAbsoluteUrl
} from './shared.mjs'

const args = parseArgs()
const baseUrl = normalizeBaseUrl(args['base-url'] || process.env.PERF_BASE_URL || 'http://localhost:3000')
const outputName = `visual-${new Date().toISOString().replace(/[:.]/g, '-')}`
const outputDir = path.join(reportsDir, outputName)

const desktop = { width: 1440, height: 900 }
const mobile = { width: 390, height: 844 }
const sourceScanRoot = path.join(rootDir, 'app')
const sourceScanExtensions = new Set(['.vue', '.ts', '.less'])
const legacyDarkSelectors = [
  '.games-page--dark',
  '.search-results--dark',
  '.is-dark-theme',
  '.spotlight-panels--dark',
  '.about-page--dark',
  '.legal-page--dark',
  '.updates-page--dark',
  '.nav-home-page--dark',
  '.gf-static-page--dark',
  '.lottery-page--dark'
]
const legacyDarkClassNames = legacyDarkSelectors.map((selector) => selector.replace(/^\./, ''))
const migratedTailwindDebtPaths = new Set([
  'app/pages/games/[id].vue',
  'app/components/common/LinkTag.vue',
  'app/components/common/SiteIconList.vue',
  'app/components/game/detail/GameDetailHeader.vue',
  'app/components/game/detail/GameDetailMain.vue',
  'app/components/game/detail/GameSidebarLinks.vue',
  'app/components/game/detail/GameSidebarSimilar.vue',
  'app/components/game/detail/tabs/GameTabComment.vue',
  'app/components/game/detail/tabs/GameTabDetail.vue',
  'app/components/game/detail/tabs/GameTabGallery.vue',
  'app/components/game/detail/tabs/GameTabIntro.vue',
  'app/components/game/detail/tabs/GameTabNews.vue',
  'app/pages/games/prize/index.vue',
  'app/pages/games/prize/activation.vue',
  'app/components/game/lottery/LotteryJoinModal.vue'
])
const retainedDeepSelectorAllowlist = new Map([
  ['app/components/site/SiteDetailPage.vue', 'Site detail renders nested card surfaces from child panels until the detail page receives its own Less migration.'],
  ['app/components/site/SitePerformancePanel.vue', 'Performance panel wraps the shared SitePerformance child component and keeps targeted child overrides local to that wrapper.']
])
const migratedTailwindDebtPattern = /\b(?:dark:|(?:bg|border)-\[[^\]]+\]|(?:bg|text|border)-(?:black|white|slate|stone|gray|zinc|neutral|red|orange|amber|yellow|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)(?:-[^\s"'`<>]+)?|(?:bg|border)-transparent|shadow-\[[^\]]+\]|ring-(?:black|white|slate|stone|gray|red|orange|amber|yellow|green|emerald|sky|blue|violet|rose|[0-9]))/

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
    id: 'game-detail',
    label: '游戏详情页',
    path: '/games/1',
    locale: 'zh-CN',
    rootSelector: '.game-detail-page',
    requiredSelectors: ['.games-page', '.game-detail-page', '.game-detail-hero', '.game-detail-tabs'],
    optionalDataSelectors: ['.game-detail-sidebar-card', '.game-detail-tag', '.game-detail-similar-item']
  }),
  ...makePageScenarios({
    id: 'lottery',
    label: '游戏抽奖页',
    path: '/games/prize',
    locale: 'zh-CN',
    rootSelector: '.lottery-page',
    requiredSelectors: ['.lottery-page', '.lottery-hero', '.lottery-section'],
    optionalDataSelectors: ['.lottery-pool', '.lottery-history__row'],
    expectActiveNav: false
  }),
  ...makeMobilePageScenarios({
    id: 'lottery-activation',
    label: '游戏抽奖激活移动端',
    path: '/games/prize/activation?status=success',
    locale: 'zh-CN',
    rootSelector: '.lottery-activation-page',
    requiredSelectors: ['.lottery-activation-page', '.activation-card', '.activation-status', '.activation-card__link'],
    optionalDataSelectors: [],
    expectActiveNav: false
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
  }),
  ...makePageScenarios({
    id: 'about',
    label: '关于页',
    path: '/about',
    locale: 'zh-CN',
    rootSelector: '.about-page',
    requiredSelectors: ['.gf-static-page', '.about-page', '.about-panel', '.about-link'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makePageScenarios({
    id: 'about-en',
    label: '关于页英文',
    path: '/en/about',
    locale: 'en-US',
    rootSelector: '.about-page',
    requiredSelectors: ['.gf-static-page', '.about-page', '.about-panel', '.about-link'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makePageScenarios({
    id: 'steam-zone',
    label: '兽游专区',
    path: '/steam',
    locale: 'zh-CN',
    rootSelector: '.steam-zone-page',
    requiredSelectors: ['.games-page', '.steam-zone-page', '.steam-zone-panel', '.steam-zone-card'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makePageScenarios({
    id: 'updates',
    label: '更新公告',
    path: '/updates',
    locale: 'zh-CN',
    rootSelector: '.updates-page',
    requiredSelectors: ['.updates-page', '.updates-summary-shell', '.updates-timeline-section'],
    optionalDataSelectors: ['.updates-entry'],
    expectActiveNav: false
  }),
  ...makePageScenarios({
    id: 'updates-en',
    label: '更新公告英文',
    path: '/en/updates',
    locale: 'en-US',
    rootSelector: '.updates-page',
    requiredSelectors: ['.updates-page', '.updates-summary-shell', '.updates-timeline-section'],
    optionalDataSelectors: ['.updates-entry'],
    expectActiveNav: false
  }),
  ...makeMobilePageScenarios({
    id: 'terms',
    label: '服务条款移动端',
    path: '/terms',
    locale: 'zh-CN',
    rootSelector: '.legal-page',
    requiredSelectors: ['.gf-static-page', '.legal-page', '.legal-panel', '.gf-static-section-list'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makeMobilePageScenarios({
    id: 'terms-en',
    label: '服务条款英文移动端',
    path: '/en/terms',
    locale: 'en-US',
    rootSelector: '.legal-page',
    requiredSelectors: ['.gf-static-page', '.legal-page', '.legal-panel', '.gf-static-section-list'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makeMobilePageScenarios({
    id: 'privacy',
    label: '隐私政策移动端',
    path: '/privacy',
    locale: 'zh-CN',
    rootSelector: '.legal-page',
    requiredSelectors: ['.gf-static-page', '.legal-page', '.legal-panel', '.gf-static-section-list'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makeMobilePageScenarios({
    id: 'privacy-en',
    label: '隐私政策英文移动端',
    path: '/en/privacy',
    locale: 'en-US',
    rootSelector: '.legal-page',
    requiredSelectors: ['.gf-static-page', '.legal-page', '.legal-panel', '.gf-static-section-list'],
    optionalDataSelectors: [],
    expectActiveNav: false
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

function makeMobilePageScenarios(config) {
  return [
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
    '- `/`、`/en`、`/games`、`/games/1`、`/games/prize`、`/games/search`、`/en/games/search`、`/steam`、`/about`、`/updates` 和法律页移动端的关键容器存在。',
    '- 不再出现 `games-page--dark`、`search-results--dark`、`is-dark-theme`、`spotlight-panels--dark`、`about-page--dark`、`legal-page--dark`、`updates-page--dark`、`lottery-page--dark` 等旧暗色入口。',
    '- 源码静态扫描不允许新增旧暗色 class、`:global(.dark ...)`、未登记 `:deep(...)` 或已迁移游戏详情/抽奖页的复杂颜色类。',
    '- 桌面与移动端不出现明显横向溢出。',
    '- 数据卡片为空只记为 warning，方便在本地后端无数据时继续保留截图。'
  )

  return `${lines.join('\n')}\n`
}

async function listSourceFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true })
  const files = []

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name)
    if (entry.isDirectory()) {
      files.push(...await listSourceFiles(fullPath))
      continue
    }

    if (entry.isFile() && sourceScanExtensions.has(path.extname(entry.name))) {
      files.push(fullPath)
    }
  }

  return files
}

function toSourcePath(filePath) {
  return path.relative(rootDir, filePath).replace(/\\/g, '/')
}

function lineNumberForIndex(content, index) {
  return content.slice(0, index).split(/\r?\n/).length
}

function pushSourceFailure(failures, filePath, line, message) {
  failures.push({
    file: filePath,
    line,
    message
  })
}

async function collectSourceDebtFailures() {
  const failures = []
  const files = await listSourceFiles(sourceScanRoot)

  for (const file of files) {
    const sourcePath = toSourcePath(file)
    const content = await readFile(file, 'utf8')

    for (const className of legacyDarkClassNames) {
      const index = content.indexOf(className)
      if (index >= 0) {
        pushSourceFailure(failures, sourcePath, lineNumberForIndex(content, index), `检测到废弃暗色入口 ${className}`)
      }
    }

    for (const match of content.matchAll(/:global\(\.dark\b/g)) {
      pushSourceFailure(failures, sourcePath, lineNumberForIndex(content, match.index || 0), '请使用 :global(html.dark ...) 或迁移到 Less token')
    }

    for (const match of content.matchAll(/:deep\(/g)) {
      if (!retainedDeepSelectorAllowlist.has(sourcePath)) {
        pushSourceFailure(failures, sourcePath, lineNumberForIndex(content, match.index || 0), '未登记的 :deep(...) 覆盖')
      }
    }

    if (migratedTailwindDebtPaths.has(sourcePath)) {
      const match = content.match(migratedTailwindDebtPattern)
      if (match?.index !== undefined) {
        pushSourceFailure(failures, sourcePath, lineNumberForIndex(content, match.index), `已迁移游戏详情/抽奖页仍包含复杂颜色类 ${match[0]}`)
      }
    }
  }

  return failures
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

  if (scenario.expectActiveNav !== false && snapshot.activeNavCount < 1) {
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
  if (scenario.action === 'reveal-home-nav') {
    await page.mouse.wheel(0, 1100)
    await page.waitForSelector('.nav-content', { timeout: 5000 }).catch(() => {})
    await page.waitForTimeout(1800)
    return
  }

}

async function runScenario(browser, scenario) {
  const context = await browser.newContext({
    viewport: scenario.viewport,
    deviceScaleFactor: 1,
    reducedMotion: 'reduce',
    locale: scenario.locale
  })
  await context.addInitScript(({ theme }) => {
    try {
      localStorage.setItem('theme', theme)
    } catch {
      // Ignore storage access errors on transient browser origins.
    }
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, {
    theme: scenario.theme
  })

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

const sourceDebtFailures = await collectSourceDebtFailures()
if (sourceDebtFailures.length) {
  console.error(`[visual] 源码样式债守卫失败 ${sourceDebtFailures.length} 项：`)
  for (const failure of sourceDebtFailures) {
    console.error(`  - ${failure.file}:${failure.line} ${failure.message}`)
  }
  process.exit(1)
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
