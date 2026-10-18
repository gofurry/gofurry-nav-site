#!/usr/bin/env node
import { readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import {
  budgetPath,
  evaluateBudget,
  parseArgs,
  renderMarkdownReport,
  reportsDir,
  runMeasurement,
  writeMeasurementReports
} from './shared.mjs'

const args = parseArgs()
const baseUrl = args['base-url'] || process.env.PERF_BASE_URL || 'http://localhost:3000'
const budget = JSON.parse(await readFile(budgetPath, 'utf8'))
const measurement = await runMeasurement({ baseUrl })
const guardResult = evaluateBudget(measurement, budget)
const reports = await writeMeasurementReports(measurement, 'guard')
const markdownWithGuard = renderMarkdownReport(measurement, guardResult)

await writeFile(reports.markdownPath, markdownWithGuard, 'utf8')
await writeFile(path.join(reportsDir, 'latest-guard.md'), markdownWithGuard, 'utf8')

console.log(`[perf] JSON 报告：${reports.jsonPath}`)
console.log(`[perf] Markdown 报告：${reports.markdownPath}`)

if (guardResult.warnings.length) {
  console.warn(`[perf] 警告 ${guardResult.warnings.length} 项：`)
  for (const warning of guardResult.warnings) {
    console.warn(`  - ${warning.scenarioLabel}: ${warning.message}`)
  }
}

if (guardResult.failures.length) {
  console.error(`[perf] 回归守卫失败 ${guardResult.failures.length} 项：`)
  for (const failure of guardResult.failures) {
    console.error(`  - ${failure.scenarioLabel}: ${failure.message}`)
  }
  process.exitCode = 1
} else {
  console.log('[perf] 回归守卫通过')
}
