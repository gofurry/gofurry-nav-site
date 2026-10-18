import assert from 'node:assert/strict'
import { mkdtemp, readFile, readdir, rm, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'
import test from 'node:test'
import { aggregateFindings, compareDebt, RULES, updateBaseline, validateManifest } from './debt.mjs'

const fileA = 'app/components/A.vue'
const fileB = 'app/components/B.vue'
const rule = 'tailwind-appearance'

function manifest(baseline = {}, exceptions = []) {
  return { schema_version: 1, baseline: Object.fromEntries(RULES.map(key => [key, baseline[key] ?? {}])), exceptions }
}

function exception(overrides = {}) {
  return {
    path: fileA,
    rule,
    issue: '#109',
    reason: 'Existing nested composition needs this temporary visual override.',
    remove_when: 'The Site Detail redesign replaces the nested composition.',
    ...overrides,
  }
}

async function temporaryManifest(t, input, contents = `${JSON.stringify(input, null, 2)}\n`) {
  const directory = await mkdtemp(path.join(tmpdir(), 'gofurry-style-debt-'))
  t.after(async () => {
    assert.equal(path.dirname(path.resolve(directory)), path.resolve(tmpdir()))
    assert.ok(path.basename(directory).startsWith('gofurry-style-debt-'))
    await rm(directory, { recursive: true, force: true })
  })
  const target = path.join(directory, 'frontend-style-debt.json')
  await writeFile(target, contents)
  return { directory, target, contents }
}

test('valid manifest is copied, ordered and keeps exact exception metadata', () => {
  const input = manifest({ [rule]: { [fileB]: 2, [fileA]: 1 } }, [exception()])
  const original = structuredClone(input)
  const result = validateManifest(input)
  assert.deepEqual(Object.keys(result.baseline), RULES)
  assert.deepEqual(Object.keys(result.baseline[rule]), [fileA, fileB])
  assert.deepEqual(result.exceptions, input.exceptions)
  result.exceptions[0].reason = 'changed only the copy'
  result.baseline[rule][fileA] = 99
  assert.deepEqual(input, original)
})

test('manifest schema rejects unknown/missing fields and rule families', () => {
  const invalid = [null, [], 'not parsed JSON', {}, { ...manifest(), schema_version: 2 }, { ...manifest(), schema_version: '1' }]
  invalid.push({ ...manifest(), ignored: true })
  invalid.push({ ...manifest(), baseline: {} })
  invalid.push({ ...manifest(), baseline: [] })
  invalid.push({ ...manifest(), baseline: { ...manifest().baseline, mystery: {} } })
  invalid.push({ ...manifest(), exceptions: {} })
  invalid.push({ schema_version: 1, baseline: manifest().baseline })
  for (const input of invalid) assert.throws(() => validateManifest(input))
})

test('manifest counts must be positive safe integers', () => {
  for (const value of [0, -1, 0.5, '1', true, null, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1]) {
    assert.throws(() => validateManifest(manifest({ [rule]: { [fileA]: value } })), /positive safe integer/)
  }
  assert.throws(() => validateManifest(manifest({ [rule]: [] })), /must be an object/)
  assert.equal(validateManifest(manifest({ [rule]: { [fileA]: Number.MAX_SAFE_INTEGER } })).baseline[rule][fileA], Number.MAX_SAFE_INTEGER)
})

test('paths reject traversal, directories, absolute forms and wildcard bypasses', () => {
  const invalid = ['', 'A.vue', '/app/A.vue', 'C:/app/A.vue', 'app\\A.vue', '../app/A.vue',
    'app/../A.vue', 'app/./A.vue', 'app//A.vue', 'app/components', 'app/A.vue/',
    'app/*.vue', 'app/**/A.vue', 'app/A?.vue', 'app/{A,B}.vue', 'app/ A.vue', 'app/A.vue ', 'app/A\0.vue']
  for (const file of invalid) {
    assert.throws(() => validateManifest(manifest({ [rule]: { [file]: 1 } })), /exact Nav Web-relative/)
    assert.throws(() => validateManifest(manifest({}, [exception({ path: file })])), /exact Nav Web-relative/)
  }
  const route = 'app/pages/sites/[id].vue'
  assert.equal(validateManifest(manifest({ [rule]: { [route]: 1 } })).baseline[rule][route], 1)
})

test('paths reject embedded control characters without rejecting Unicode filenames', () => {
  for (const code of [0, 9, 10, 13, 31, 127]) {
    const file = `app/components/A${String.fromCharCode(code)}B.vue`
    assert.throws(() => validateManifest(manifest({ [rule]: { [file]: 1 } })), /exact Nav Web-relative/)
    assert.throws(() => validateManifest(manifest({}, [exception({ path: file })])), /exact Nav Web-relative/)
  }
  const file = 'app/components/导航.vue'
  assert.equal(validateManifest(manifest({ [rule]: { [file]: 1 } })).baseline[rule][file], 1)
})

test('exceptions require complete, specific, non-empty metadata', () => {
  for (const field of ['path', 'rule', 'issue', 'reason', 'remove_when']) {
    const entry = exception()
    delete entry[field]
    assert.throws(() => validateManifest(manifest({}, [entry])), /missing/)
  }
  for (const field of ['issue', 'reason', 'remove_when']) {
    for (const value of ['', '  ', null, 109]) {
      assert.throws(() => validateManifest(manifest({}, [exception({ [field]: value })])), /non-empty string/)
    }
  }
  assert.throws(() => validateManifest(manifest({}, [exception({ rule: '*' })])), /unknown rule/)
  assert.throws(() => validateManifest(manifest({}, [exception({ ignore: true })])), /unknown key/)
  assert.throws(() => validateManifest(manifest({}, [exception(), exception()])), /duplicates exception/)
  assert.throws(() => validateManifest(manifest({}, [null])), /must be an object/)
})

test('aggregation counts occurrences and exempts only the exact rule/path pair', () => {
  const route = 'app/pages/sites/[id].vue'
  const findings = [
    { rule, file: fileA }, { rule: 'raw-visual-value', file: fileA },
    { rule, file: fileB }, { rule, file: fileB },
    { rule, file: route }, { rule, file: 'app/pages/sites/id.vue' },
  ]
  const input = manifest({}, [exception(), exception({ path: route })])
  const actual = aggregateFindings(findings, input)
  assert.deepEqual(Object.keys(actual), RULES)
  assert.deepEqual(actual[rule], { [fileB]: 2, 'app/pages/sites/id.vue': 1 })
  assert.deepEqual(actual['raw-visual-value'], { [fileA]: 1 })
  assert.deepEqual(actual.important, {})
})

test('invalid detector findings fail closed instead of disappearing', () => {
  assert.throws(() => aggregateFindings(null, manifest()), /must be an array/)
  assert.throws(() => aggregateFindings([null], manifest()), /must be an object/)
  assert.throws(() => aggregateFindings([{ rule: 'unknown', file: fileA }], manifest()), /unknown rule/)
  assert.throws(() => aggregateFindings([{ rule, file: '../A.vue' }], manifest()), /exact Nav Web-relative/)
  assert.deepEqual(aggregateFindings([], manifest()), manifest().baseline)
})

test('equal counts pass; missing new-file budgets default to zero', () => {
  assert.deepEqual(compareDebt({ [rule]: { [fileA]: 3 } }, { [rule]: { [fileA]: 3 } }), [])
  assert.deepEqual(compareDebt({}, {}), [])
  assert.deepEqual(compareDebt({ [rule]: { [fileA]: 1 } }, {}), [
    { rule, file: fileA, baseline: 0, actual: 1, kind: 'regression' },
  ])
})

test('regressions and stale budgets cannot offset across files or rules', () => {
  const actual = { [rule]: { [fileA]: 2, [fileB]: 2 }, important: { [fileA]: 1 } }
  const baseline = { [rule]: { [fileA]: 3, [fileB]: 1 }, important: { [fileB]: 1 } }
  assert.deepEqual(compareDebt(actual, baseline), [
    { rule, file: fileA, baseline: 3, actual: 2, kind: 'stale' },
    { rule, file: fileB, baseline: 1, actual: 2, kind: 'regression' },
    { rule: 'important', file: fileA, baseline: 0, actual: 1, kind: 'regression' },
    { rule: 'important', file: fileB, baseline: 1, actual: 0, kind: 'stale' },
  ])
})

test('comparison rejects invalid maps instead of treating them as zero', () => {
  for (const invalid of [null, [], { mystery: {} }, { [rule]: null }, { [rule]: { [fileA]: -1 } }, { [rule]: { [fileA]: 1.2 } }]) {
    assert.throws(() => compareDebt(invalid, {}))
    assert.throws(() => compareDebt({}, invalid))
  }
})

test('update only decreases, omits zero entries, preserves exceptions and writes deterministic JSON', async t => {
  const input = manifest({ [rule]: { [fileB]: 4, [fileA]: 3 }, important: { [fileA]: 1 } }, [exception({ rule: 'deep-selector' })])
  const { directory, target } = await temporaryManifest(t, input)
  const actual = { [rule]: { [fileB]: 2, [fileA]: 1 }, important: { [fileA]: 0 } }
  const updated = await updateBaseline(target, input, actual)
  const expected = manifest({ [rule]: { [fileA]: 1, [fileB]: 2 } }, input.exceptions)
  assert.deepEqual(updated, expected)
  assert.equal(await readFile(target, 'utf8'), `${JSON.stringify(expected, null, 2)}\n`)
  assert.deepEqual(await readdir(directory), ['frontend-style-debt.json'])
  assert.equal(input.baseline[rule][fileA], 3)
})

test('any increase prevents every write even when another file decreases', async t => {
  const input = manifest({ [rule]: { [fileA]: 3, [fileB]: 1 } })
  const { directory, target, contents } = await temporaryManifest(t, input, `  ${JSON.stringify(input)}  \r\n`)
  await assert.rejects(updateBaseline(target, input, { [rule]: { [fileA]: 1, [fileB]: 2 } }), /Refusing baseline increase/)
  assert.equal(await readFile(target, 'utf8'), contents)
  assert.deepEqual(await readdir(directory), ['frontend-style-debt.json'])
})

test('new-file debt is refused without modifying manifest bytes', async t => {
  const input = manifest({ [rule]: { [fileA]: 1 } })
  const { target, contents } = await temporaryManifest(t, input)
  await assert.rejects(updateBaseline(target, input, { [rule]: { [fileB]: 1 } }), /0 -> 1/)
  assert.equal(await readFile(target, 'utf8'), contents)
})

test('invalid actual or manifest fails before any write', async t => {
  const input = manifest({ [rule]: { [fileA]: 1 } })
  const { target, contents } = await temporaryManifest(t, input)
  await assert.rejects(updateBaseline(target, input, { [rule]: { [fileA]: -1 } }), /non-negative safe integer/)
  await assert.rejects(updateBaseline(target, { ...input, schema_version: 9 }, {}), /schema_version/)
  assert.equal(await readFile(target, 'utf8'), contents)
})

test('equal baseline preserves existing bytes without a formatting rewrite', async t => {
  const input = manifest({ [rule]: { [fileA]: 1 } })
  const { target, contents } = await temporaryManifest(t, input, JSON.stringify(input))
  assert.deepEqual(await updateBaseline(target, input, input.baseline), input)
  assert.equal(await readFile(target, 'utf8'), contents)
})

test('update refuses a manifest changed since the scan began', async t => {
  const input = manifest({ [rule]: { [fileA]: 2 } })
  const { target } = await temporaryManifest(t, input)
  const changed = `${JSON.stringify(manifest({ [rule]: { [fileA]: 1 } }))}\n`
  await writeFile(target, changed)
  await assert.rejects(updateBaseline(target, input, {}), /changed during the scan/)
  assert.equal(await readFile(target, 'utf8'), changed)
})

test('invalid JSON on disk fails closed and remains untouched', async t => {
  const input = manifest({ [rule]: { [fileA]: 1 } })
  const { target, contents } = await temporaryManifest(t, input, '{ broken JSON')
  await assert.rejects(updateBaseline(target, input, {}), SyntaxError)
  assert.equal(await readFile(target, 'utf8'), contents)
})
