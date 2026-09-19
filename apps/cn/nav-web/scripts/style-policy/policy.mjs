import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { readSourceFacts } from './source.mjs'
import { detectCssFacts } from './css.mjs'
import { detectTailwindFacts } from './tailwind.mjs'
import { validateManifest, aggregateFindings, compareDebt, updateBaseline, RULES } from './debt.mjs'

export async function detectStyleFacts(facts) {
  const tailwind = await detectTailwindFacts(facts)
  const arbitrary = tailwind.filter(finding => finding.rule === 'tailwind-arbitrary-appearance' && finding.offset !== undefined)
  // A literal can feed both class and visual sinks. Only real compiler-validated
  // arbitrary utilities exclude their contained colors from a third rule hit.
  const css = detectCssFacts(facts).filter(finding => finding.rule !== 'raw-visual-value' || finding.offset === undefined
    || !arbitrary.some(utility => utility.file === finding.file && utility.offset <= finding.offset && finding.offset < utility.offset + utility.value.length))
  return [...css, ...tailwind]
}

export async function scanPolicy(root, files) {
  const source = await readSourceFacts(root, files)
  const findings = await detectStyleFacts(source.facts)
  findings.sort((a, b) => a.file.localeCompare(b.file) || a.line - b.line || a.rule.localeCompare(b.rule))
  return { ...source, findings }
}

export async function checkPolicy(root, { update = false } = {}) {
  const manifestPath = path.join(root, 'frontend-style-debt.json')
  const manifest = validateManifest(JSON.parse(await readFile(manifestPath, 'utf8')))
  const scan = await scanPolicy(root)
  const actual = aggregateFindings(scan.findings, manifest)
  const differences = compareDebt(actual, manifest.baseline)
  if (update && !differences.some(item => item.kind === 'regression')) await updateBaseline(manifestPath, manifest, actual)
  return { ...scan, actual, differences, ok: differences.length === 0 || (update && differences.every(item => item.kind === 'stale')) }
}

export function formatReport(result) {
  const lines = [`Style policy ${result.ok ? 'passed' : 'failed'} (${result.files.length} source files).`]
  for (const rule of RULES) {
    const counts = Object.values(result.actual[rule])
    lines.push(`  ${rule}: ${counts.length} files / ${counts.reduce((a, b) => a + b, 0)} occurrences`)
  }
  for (const difference of result.differences) {
    const { rule, file, baseline, actual, kind } = difference
    lines.push(`\n${kind}: ${rule}\n  ${file}\n  baseline: ${baseline}; actual: ${actual}`)
    if (kind === 'stale') lines.push(result.ok ? '  Stored budget lowered.' : '  Lower the stored budget with npm run style:policy:update.')
    else for (const finding of result.findings.filter(item => item.rule === rule && item.file === file)) {
      lines.push(`  line ${finding.line}: ${finding.value} — ${finding.message}`)
    }
  }
  return lines.join('\n')
}
