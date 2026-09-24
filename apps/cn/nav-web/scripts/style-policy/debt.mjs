import { randomUUID } from 'node:crypto'
import { readFile, rename, rm, stat, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

export const RULES = Object.freeze([
  'tailwind-appearance',
  'tailwind-arbitrary-appearance',
  'raw-visual-value',
  'important',
  'deep-selector',
  'legacy-dark-entry',
])

const ruleSet = new Set(RULES)
const exceptionFields = ['path', 'rule', 'issue', 'reason', 'remove_when']

function record(value, label) {
  if (value === null || typeof value !== 'object' || Array.isArray(value)
    || ![Object.prototype, null].includes(Object.getPrototypeOf(value))) {
    throw new Error(`${label} must be an object`)
  }
}

function exactKeys(value, expected, label) {
  record(value, label)
  for (const key of Reflect.ownKeys(value)) {
    if (!expected.includes(key)) throw new Error(`${label} has unknown key ${String(key)}`)
  }
  for (const key of expected) {
    if (!Object.hasOwn(value, key)) throw new Error(`${label} is missing ${key}`)
  }
}

function sourcePath(value, label) {
  if (typeof value !== 'string' || !value.startsWith('app/')
    || !/\.(?:vue|ts|js|css|less)$/.test(value)
    || /[\\:*?{}]/.test(value)
    || [...value].some(character => character.charCodeAt(0) < 32 || character.charCodeAt(0) === 127)
    || value.split('/').some(part => !part || part === '.' || part === '..' || part.trim() !== part)) {
    throw new Error(`${label} must be an exact Nav Web-relative POSIX source-file path under app/`)
  }
  // Brackets are literal characters: Nuxt routes such as [id].vue are not globs.
  return value
}

function knownRule(value, label) {
  if (!ruleSet.has(value)) throw new Error(`${label} has unknown rule ${String(value)}`)
  return value
}

function emptyCounts() {
  return Object.fromEntries(RULES.map(rule => [rule, {}]))
}

function counts(value, label, { requireAll = false, allowZero = true } = {}) {
  record(value, label)
  for (const rule of Reflect.ownKeys(value)) knownRule(rule, label)
  const normalized = emptyCounts()
  for (const rule of RULES) {
    if (!Object.hasOwn(value, rule)) {
      if (requireAll) throw new Error(`${label} is missing rule ${rule}`)
      continue
    }
    record(value[rule], `${label}.${rule}`)
    for (const file of Reflect.ownKeys(value[rule]).sort()) {
      sourcePath(file, `${label}.${rule} path`)
      const count = value[rule][file]
      if (!Number.isSafeInteger(count) || count < (allowZero ? 0 : 1)) {
        throw new Error(`${label}.${rule}[${file}] must be a ${allowZero ? 'non-negative' : 'positive'} safe integer`)
      }
      if (count > 0) normalized[rule][file] = count
    }
  }
  return normalized
}

/** Validate parsed JSON without normalizing invalid paths or granting new budgets. */
export function validateManifest(input) {
  exactKeys(input, ['schema_version', 'baseline', 'exceptions'], 'manifest')
  if (input.schema_version !== 1) throw new Error(`Unsupported manifest schema_version: ${String(input.schema_version)}`)
  const baseline = counts(input.baseline, 'manifest.baseline', { requireAll: true, allowZero: false })
  if (!Array.isArray(input.exceptions)) throw new Error('manifest.exceptions must be an array')
  const seen = new Set()
  const exceptions = input.exceptions.map((entry, index) => {
    const label = `manifest.exceptions[${index}]`
    exactKeys(entry, exceptionFields, label)
    sourcePath(entry.path, `${label}.path`)
    knownRule(entry.rule, label)
    for (const field of ['issue', 'reason', 'remove_when']) {
      if (typeof entry[field] !== 'string' || !entry[field].trim()) {
        throw new Error(`${label}.${field} must be a non-empty string`)
      }
    }
    const key = `${entry.rule}\0${entry.path}`
    if (seen.has(key)) throw new Error(`${label} duplicates exception for ${entry.rule} ${entry.path}`)
    seen.add(key)
    return Object.fromEntries(exceptionFields.map(field => [field, entry[field]]))
  })
  return { schema_version: 1, baseline, exceptions }
}

/** Each finding is one authored occurrence; exceptions match only its rule and path. */
export function aggregateFindings(findings, manifest) {
  const validated = validateManifest(manifest)
  if (!Array.isArray(findings)) throw new Error('findings must be an array')
  const exemptions = new Set(validated.exceptions.map(entry => `${entry.rule}\0${entry.path}`))
  const actual = emptyCounts()
  for (const [index, finding] of findings.entries()) {
    record(finding, `findings[${index}]`)
    const rule = knownRule(finding.rule, `findings[${index}]`)
    const file = sourcePath(finding.file, `findings[${index}].file`)
    if (!exemptions.has(`${rule}\0${file}`)) {
      actual[rule][file] = (actual[rule][file] ?? 0) + 1
    }
  }
  return counts(actual, 'actual')
}

/** Missing rule/file pairs have budget zero; decreases are stale until recorded. */
export function compareDebt(actual, baseline) {
  const observed = counts(actual, 'actual')
  const budget = counts(baseline, 'baseline')
  const changes = []
  for (const rule of RULES) {
    const files = new Set([...Object.keys(observed[rule]), ...Object.keys(budget[rule])])
    for (const file of [...files].sort()) {
      const expected = budget[rule][file] ?? 0
      const current = observed[rule][file] ?? 0
      if (current !== expected) {
        changes.push({ rule, file, baseline: expected, actual: current, kind: current > expected ? 'regression' : 'stale' })
      }
    }
  }
  return changes
}

/** Atomically replace the manifest only after every rule/file passes the decrease gate. */
export async function updateBaseline(manifestPath, manifest, actual) {
  const validated = validateManifest(manifest)
  const baseline = counts(actual, 'actual')
  const changes = compareDebt(baseline, validated.baseline)
  const regressions = changes.filter(change => change.kind === 'regression')
  if (regressions.length) {
    throw new Error(`Refusing baseline increase:\n${regressions.map(change => `${change.rule} ${change.file}: ${change.baseline} -> ${change.actual}`).join('\n')}`)
  }
  if (!changes.length) return validated

  const target = path.resolve(manifestPath instanceof URL ? fileURLToPath(manifestPath) : manifestPath)
  const current = validateManifest(JSON.parse(await readFile(target, 'utf8')))
  if (JSON.stringify(current) !== JSON.stringify(validated)) {
    throw new Error('Manifest changed during the scan; refusing to overwrite it')
  }

  const updated = { ...validated, baseline }
  const temporary = path.join(path.dirname(target), `.${path.basename(target)}.${randomUUID()}.tmp`)
  const { mode } = await stat(target)
  try {
    await writeFile(temporary, `${JSON.stringify(updated, null, 2)}\n`, { flag: 'wx', mode })
    await rename(temporary, target)
  }
  finally {
    await rm(temporary, { force: true })
  }
  return updated
}
