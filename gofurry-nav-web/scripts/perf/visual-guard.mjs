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
  '.archive-page--dark',
  '.nav-home-page--dark',
  '.gf-static-page--dark',
  '.lottery-page--dark'
]
const legacyDarkClassNames = legacyDarkSelectors.map((selector) => selector.replace(/^\./, ''))
const migratedTailwindDebtPaths = new Set([
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
  ...makePageScenarios({
    id: 'archive-empty',
    label: '知识库空态',
    path: '/archive',
    locale: 'zh-CN',
    rootSelector: '.archive-page',
    requiredSelectors: ['.archive-page', '.archive-sidebar', '.archive-workspace', '.empty-conversation', '.ask-bar'],
    optionalDataSelectors: [],
    expectActiveNav: false
  }),
  ...makePageScenarios({
    id: 'archive-guide',
    label: '知识库说明弹窗',
    path: '/archive',
    locale: 'zh-CN',
    rootSelector: '.archive-page',
    requiredSelectors: ['.archive-page', '.archive-sidebar', '.archive-workspace', '.guide-panel', '.ask-bar'],
    optionalDataSelectors: [],
    action: 'open-archive-guide',
    expectActiveNav: false
  }),
  ...makePageScenarios({
    id: 'archive-session',
    label: '知识库会话与引用',
    path: '/archive',
    locale: 'zh-CN',
    rootSelector: '.archive-page',
    requiredSelectors: [
      '.archive-page',
      '.history-row',
      '.conversation-list',
      '.conversation',
      '.answer-markdown',
      '.citations-block',
      '.citation-item',
      '.citation-markdown',
      '.error-message',
      '.ask-bar'
    ],
    optionalDataSelectors: [],
    action: 'open-archive-seeded-session',
    seedArchiveSessions: true,
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
    '- `/`、`/en`、`/games`、`/games/prize`、`/games/search`、`/en/games/search`、`/about`、`/updates`、`/archive` 和法律页移动端的关键容器存在。',
    '- 不再出现 `games-page--dark`、`search-results--dark`、`is-dark-theme`、`spotlight-panels--dark`、`about-page--dark`、`legal-page--dark`、`updates-page--dark`、`archive-page--dark`、`lottery-page--dark` 等旧暗色入口。',
    '- 源码静态扫描不允许新增旧暗色 class、`:global(.dark ...)`、未登记 `:deep(...)` 或已迁移抽奖页的复杂颜色类。',
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
        pushSourceFailure(failures, sourcePath, lineNumberForIndex(content, match.index), `已迁移抽奖页仍包含复杂颜色类 ${match[0]}`)
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

  if (scenario.action === 'open-archive-guide') {
    await page.click('.doc-button')
    await page.waitForSelector('.guide-panel', { timeout: 5000 }).catch(() => {})
    await page.waitForTimeout(600)
    return
  }

  if (scenario.action === 'open-archive-seeded-session') {
    await page.waitForSelector('.history-item', { timeout: 5000 }).catch(() => {})
    await page.click('.history-item').catch(() => {})
    await page.waitForSelector('.conversation', { timeout: 5000 }).catch(() => {})
    await page.click('.citation-item summary').catch(() => {})
    await page.waitForSelector('.citation-markdown', { timeout: 8000 }).catch(() => {})
    await page.waitForTimeout(900)
  }
}

function makeArchiveSeedSessions() {
  const now = Date.UTC(2026, 5, 14, 4, 0, 0)
  return [
    {
      id: 'visual-archive-session',
      title: 'Archive style guard conversation',
      createdAt: now,
      updatedAt: now + 2000,
      messages: [
        {
          id: 'visual-archive-answer',
          question: '如何检查 archive 样式迁移边界？',
          answer: [
            '## 样式隔离检查',
            '',
            '- 根节点使用 `.archive-page` 承载页面 token。',
            '- Markdown 预览只在 archive 页面内改写。',
            '- session sidebar、citation 和输入区保持独立视觉。',
            '',
            '| 区域 | 状态 |',
            '| --- | --- |',
            '| citation | 已隔离 |',
            '| markdown | 已收口 |',
            '',
            '> 引用预览需要保持可读，并且不能污染其他页面。',
            '',
            '`html.dark` 是唯一暗色入口。'
          ].join('\n'),
          citations: [
            {
              source_type: 'docs',
              title: 'Style system roadmap',
              url: 'https://example.com/archive-style',
              chunk_index: 3,
              score: 0.94,
              snippet: [
                '### 引用片段',
                '',
                'Archive citation markdown **只在页面内** 覆盖。',
                '',
                '```css',
                '.archive-page .citation-markdown {}',
                '```'
              ].join('\n')
            }
          ],
          createdAt: now,
          updatedAt: now + 1000,
          status: 'done'
        },
        {
          id: 'visual-archive-error',
          question: '接口失败时用户还能看到什么？',
          answer: '',
          citations: [],
          createdAt: now + 1000,
          updatedAt: now + 2000,
          status: 'error',
          error: '本地 RAG 接口不可用，已保留问题和错误提示。'
        }
      ]
    }
  ]
}

async function runScenario(browser, scenario) {
  const context = await browser.newContext({
    viewport: scenario.viewport,
    deviceScaleFactor: 1,
    reducedMotion: 'reduce',
    locale: scenario.locale
  })
  await context.addInitScript(({ theme, archiveSessions }) => {
    try {
      localStorage.setItem('theme', theme)
      if (archiveSessions) {
        localStorage.setItem('gofurry.archive.chat.sessions.v1', JSON.stringify(archiveSessions))
        localStorage.removeItem('gofurry.archive.chat.records.v1')
      } else {
        localStorage.removeItem('gofurry.archive.chat.sessions.v1')
      }
    } catch {
      // Ignore storage access errors on transient browser origins.
    }
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, {
    theme: scenario.theme,
    archiveSessions: scenario.seedArchiveSessions ? makeArchiveSeedSessions() : null
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
