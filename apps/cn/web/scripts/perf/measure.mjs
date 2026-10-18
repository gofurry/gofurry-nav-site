#!/usr/bin/env node
import { baselinePath, compactMeasurement, parseArgs, runMeasurement, writeJson, writeMeasurementReports } from './shared.mjs'

const args = parseArgs()
const baseUrl = args['base-url'] || process.env.PERF_BASE_URL || 'http://localhost:3000'
const measurement = await runMeasurement({ baseUrl })
const reports = await writeMeasurementReports(measurement, args.baseline ? 'baseline' : 'measure')

console.log(`[perf] JSON 报告：${reports.jsonPath}`)
console.log(`[perf] Markdown 报告：${reports.markdownPath}`)

if (args.baseline) {
  await writeJson(baselinePath, compactMeasurement(measurement))
  console.log(`[perf] 已刷新基线：${baselinePath}`)
}
