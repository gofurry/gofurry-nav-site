import assert from 'node:assert/strict'
import { execFileSync, spawnSync } from 'node:child_process'
import { mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import { extractSource, discoverSources, isPolicySource, readSourceFacts } from './source.mjs'
import { detectCssFacts } from './css.mjs'
import { detectTailwindFacts } from './tailwind.mjs'
import { checkPolicy, formatReport, detectStyleFacts } from './policy.mjs'
import { RULES } from './debt.mjs'

const root = fileURLToPath(new URL('../../', import.meta.url))
const cli = path.join(root, 'scripts/style-policy.mjs')
const empty = () => ({ schema_version: 1, baseline: Object.fromEntries(RULES.map(rule => [rule, {}])), exceptions: [] })
const facts = source => extractSource(source, 'app/components/Fixture.vue').facts
const classes = source => facts(source).filter(fact => fact.kind === 'class').map(fact => fact.value)
const colors = source => detectCssFacts(facts(source)).filter(finding => finding.rule === 'raw-visual-value').map(finding => finding.value)

test('static classes and transition props are extracted, comments/messages are not', () => {
  assert.deepEqual(classes(`<template><!-- class="bg-red-500" --><div class="flex rounded-lg" enter-active-class="opacity-70" :title="'bg-white'" /></template>`), ['flex rounded-lg', 'opacity-70'])
})

test('bound arrays, object keys, ternaries and logical class values', () => {
  const source = `<template><div :class="['text-sm', ok ? 'bg-white' : 'bg-black', ok && 'rounded-lg', { 'font-bold': ok, italic: ok }]" /></template>`
  assert.deepEqual(classes(source), ['text-sm', 'bg-white', 'bg-black', 'rounded-lg', 'font-bold', 'italic'])
})

test('local const/ref/computed/function returns follow actual class sinks once per authored literal', () => {
  const source = `<script setup lang="ts">
  const unrelated = 'bg-purple-500'
  const size = 'text-sm'
  function tone() { if (active) return 'bg-white'; return 'bg-black' }
  const current = computed(() => tone())
  </script><template><div :class="[size, current]" /><div :class="current" /></template>`
  assert.deepEqual(classes(source), ['text-sm', 'bg-white', 'bg-black'])
})

test('Footer-like v-for property dataflow ignores unused object fields', () => {
  const source = `<script setup>const entries = [{ hoverClass: 'hover:opacity-70', message: 'bg-red-500' }, { hoverClass: 'hover:scale-105' }]</script>
  <template><i v-for="(entry, index) in entries" :class="entry.hoverClass" /></template>`
  assert.deepEqual(classes(source), ['hover:opacity-70', 'hover:scale-105'])
})

test('function-local indexed palette and independent lexical scopes', () => {
  const source = `<script setup lang="ts">
  const colors = ['bg-red-500']
  function tagClass(index: number) { const colors = ['bg-white text-black', 'bg-black text-white']; return colors[index % colors.length] }
  </script><template><span :class="tagClass(1)" /></template>`
  assert.deepEqual(classes(source), ['bg-white text-black', 'bg-black text-white'])
})

test('static template literals resolve; unsupported runtime strings are unresolved', () => {
  const source = '<script setup>const tone = "white"; const computedClass = `bg-${tone}`</script><template><i :class="computedClass" /><i :class="`tone-${api.value}`" /></template>'
  assert.deepEqual(classes(source), ['bg-white'])
  assert.ok(extractSource(source, 'app/example.vue').unresolved.length > 0)
})

test('recursive aliases terminate without guessing class strings', () => {
  assert.deepEqual(classes('<script setup>const a = b; const b = a</script><template><i :class="a" /></template>'), [])
  const result = extractSource('<script setup>const a = [a]</script><template><i :style="a" /></template>', 'app/recursive.vue')
  assert.deepEqual(result.facts, [])
  assert.ok(result.unresolved.length > 0)
})

test('object lookup resolves const keys and ref object properties', () => {
  const source = `<script setup>
  const tone = 'success'; const options = {success: 'bg-green-500', error: 'bg-red-500'}
  const settings = ref({button: 'text-sm'}); const current = computed(() => settings.value.button)
  </script><template><i :class="[options[tone], current]" /></template>`
  assert.deepEqual(classes(source), ['bg-green-500', 'text-sm'])
})

test('pure interpolation aliases share an authored occurrence and recursive templates terminate', () => {
  assert.deepEqual(classes('<script setup>const c="text-sm"</script><template><i :class="c" /><i :class="`${c}`" /></template>'), ['text-sm'])
  assert.deepEqual(classes('<script setup>const c=`bg-${c}`</script><template><i :class="c" /></template>'), [])
})

test('direct DOM class sinks in scripts use local dataflow', () => {
  const source = `const c='bg-white'; document.body.classList.add(c, 'games-page--dark'); document.body.className='text-sm'; document.body.setAttribute('class', 'font-bold'); const message='bg-black'`
  const extracted = extractSource(source, 'app/dom.ts').facts.filter(fact => fact.kind === 'class')
  assert.deepEqual(extracted.map(fact => fact.value), ['bg-white', 'games-page--dark', 'text-sm', 'font-bold'])
})

test('class palette arbitrary colors are not triple-counted as raw values', async () => {
  const source = `<script setup>const colors=['bg-[#fff]']</script><template><i :class="colors" /></template>`
  assert.deepEqual(classes(source), ['bg-[#fff]'])
  const findings = await detectStyleFacts(facts(source))
  assert.deepEqual(findings.map(finding => finding.rule), ['tailwind-appearance', 'tailwind-arbitrary-appearance'])
})

test('invalid class use cannot hide a genuine raw color sharing the literal', async () => {
  const source = `<script setup>const ink='#fff'</script><template><i :class="ink" :style="{color:ink}" /></template>`
  const findings = await detectStyleFacts(facts(source))
  assert.deepEqual(findings.map(finding => finding.rule), ['raw-visual-value'])
})

test('script/chart visual objects and semantic palettes exclude validation messages', () => {
  const source = `<script setup>
  const message = 'Invalid color: #9c846a'
  const palette = ['#abc', '#ffffff']
  const chart = { textStyle: { color: '#123' }, lineStyle: { color: 'rgb(1 2 3)' }, title: '#aaa' }
  const defaults = ref({ light_color: '#000000', dark_color: '#ffffff' })
  </script><template><div /></template>`
  assert.deepEqual(colors(source), ['#abc', '#ffffff', '#123', 'rgb(1 2 3)', '#000000', '#ffffff'])
})

test('inline/bound styles and SVG paint count only visual data', () => {
  const source = `<template><svg style="color: #abc" :style="{backgroundColor: '#fff'}" fill="#123" :stroke="'#456'" title="#aaa" /></template>`
  assert.deepEqual(colors(source), ['#abc', '#fff', '#123', '#456'])
})

test('script literal referenced by several style sinks is one authored occurrence', () => {
  assert.deepEqual(colors(`<script setup>const ink = '#123'; const option = { color: ink }</script><template><i :style="{color: ink}" /></template>`), ['#123'])
})

test('emitted noscript style goes through the CSS parser', () => {
  const source = `<script setup>useHead({noscript: [{innerHTML: '<style>.error{opacity:1!important;color:#fff}</style>'}]})</script><template><div /></template>`
  const findings = detectCssFacts(facts(source))
  assert.equal(findings.filter(finding => finding.rule === 'important').length, 1)
  assert.deepEqual(findings.filter(finding => finding.rule === 'raw-visual-value').map(finding => finding.value), ['#fff'])
})

test('embedded CSS is parsed from actual HTML sinks, not documentation strings', () => {
  const source = `<script setup>
  const help = 'Example: <style>.demo{color:#fff!important}</style>'
  const html = '<style>.visible{color:#abc!important}</style>'
  document.body.innerHTML = html
  </script><template><div v-html="html" /></template>`
  const findings = detectCssFacts(facts(source))
  assert.equal(findings.filter(finding => finding.rule === 'important').length, 1)
  assert.deepEqual(findings.filter(finding => finding.rule === 'raw-visual-value').map(finding => finding.value), ['#abc'])
})

test('literal inline CSS is parsed for annotations as well as colors', () => {
  const findings = detectCssFacts(facts('<template><i style="color:#abc!important" /></template>'))
  assert.equal(findings.filter(finding => finding.rule === 'important').length, 1)
  assert.deepEqual(findings.filter(finding => finding.rule === 'raw-visual-value').map(finding => finding.value), ['#abc'])
})

test('bound CSS strings and direct style DOM sinks retain important annotations', () => {
  const source = `<script setup>
  const style = 'color:#abc!important'; element.setAttribute('style', style)
  element.style.cssText = 'opacity:1!important'
  </script><template><i :style="[style, { backgroundColor: '#fff' }]" /><i :style="'opacity:0!important'" /></template>`
  const findings = detectCssFacts(facts(source))
  assert.equal(findings.filter(finding => finding.rule === 'important').length, 3)
  assert.deepEqual(findings.filter(finding => finding.rule === 'raw-visual-value').map(finding => finding.value), ['#abc', '#fff'])
})

test('source line numbers point to script and template occurrences', async () => {
  const source = `<script setup>\nconst c = 'bg-white'\n</script>\n<template>\n<div :class="c" class="text-sm" />\n</template>`
  const findings = await detectTailwindFacts(facts(source))
  assert.equal(findings.find(finding => finding.value === 'bg-white').line, 2)
  assert.equal(findings.find(finding => finding.value === 'text-sm').line, 5)
})

test('invalid SFC, script, bound expression and style fail closed', () => {
  for (const source of [
    '<template><div></template>',
    '<script setup>const x = </script><template><div /></template>',
    '<template><div :class="[" /></template>',
    '<template><div /></template><style>.a{color:</style>',
  ]) assert.throws(() => facts(source))
})

test('unsupported external blocks and style/template languages cannot silently bypass', () => {
  for (const source of [
    '<template><div /></template><style src="./other.css" />',
    '<template lang="pug">div</template>',
    '<template><div /></template><style lang="scss">.a{color:red}</style>',
  ]) assert.throws(() => facts(source))
})

test('discovery preserves P0 exclusions without excluding domain or ambient source', () => {
  for (const file of ['app/components/experimental/ambient/example.vue', 'app/components/site/Detail.vue', 'app/pages/[id].vue', 'app/new.ts']) assert.ok(isPolicySource(file))
  for (const file of ['app/assets/js/china.js', 'app/types/a.d.ts', 'app/a.test.ts', 'app/a.spec.js', 'app/__tests__/a.vue', 'app/fixtures/a.css', 'app/generated/a.ts', 'server/a.ts', 'app/icon.svg']) assert.equal(isPolicySource(file), false)
})

async function fixture(t, source = '<template><div class="flex" /></template>') {
  const dir = await mkdtemp(path.join(os.tmpdir(), 'gofurry-style-policy-'))
  t.after(() => rm(dir, { recursive: true, force: true }))
  execFileSync('git', ['init', '--quiet', dir])
  await mkdir(path.join(dir, 'app/components'), { recursive: true })
  await writeFile(path.join(dir, 'app/components/New.vue'), source)
  await writeFile(path.join(dir, 'frontend-style-debt.json'), JSON.stringify(empty()))
  return dir
}

test('new untracked source is included and receives zero budget', async t => {
  const dir = await fixture(t, '<template><div class="rounded-lg" /></template>')
  assert.deepEqual(discoverSources(dir), ['app/components/New.vue'])
  const result = await checkPolicy(dir)
  assert.equal(result.ok, false)
  assert.deepEqual(result.differences[0], { rule: 'tailwind-appearance', file: 'app/components/New.vue', baseline: 0, actual: 1, kind: 'regression' })
  assert.match(formatReport(result), /line 1: rounded-lg/)
})

test('unreadable source and invalid manifest fail closed', async t => {
  const dir = await fixture(t)
  await assert.rejects(readSourceFacts(dir, ['app/missing.vue']), /ENOENT/)
  await writeFile(path.join(dir, 'frontend-style-debt.json'), '{')
  await assert.rejects(checkPolicy(dir), SyntaxError)
})

test('CLI fails on tooling errors, regressions and unknown options without rewriting manifest', async t => {
  const dir = await fixture(t, '<template><div class="bg-white" /></template>')
  const manifestPath = path.join(dir, 'frontend-style-debt.json')
  const original = await readFile(manifestPath, 'utf8')
  for (const args of [[], ['--update-baseline'], ['--accept-all']]) {
    const result = spawnSync(process.execPath, [cli, ...args], { cwd: dir, encoding: 'utf8' })
    assert.equal(result.status, 1, result.stderr)
    assert.equal(await readFile(manifestPath, 'utf8'), original)
  }
  await writeFile(manifestPath, '{')
  const result = spawnSync(process.execPath, [cli], { cwd: dir, encoding: 'utf8' })
  assert.equal(result.status, 1)
  assert.match(result.stderr, /tooling failure/)
})

test('current application tree exactly matches every checked-in rule/file budget', async () => {
  const result = await checkPolicy(root)
  assert.equal(result.ok, true, formatReport(result))
  assert.deepEqual(result.differences, [])
})

test('Modal retirement updater removes only its stale entry and grants no new file debt', async t => {
  const dir = await fixture(t)
  const oldFile = 'app/assets/styles/components/modal.less'
  const current = JSON.parse(await readFile(path.join(root, 'frontend-style-debt.json'), 'utf8'))
  const before = structuredClone(current)
  before.baseline['raw-visual-value'][oldFile] = 29
  const target = path.join(dir, 'frontend-style-debt.json')
  await writeFile(target, JSON.stringify(before))
  const { updateBaseline } = await import('./debt.mjs')
  await updateBaseline(target, before, current.baseline)
  assert.deepEqual(JSON.parse(await readFile(target, 'utf8')), current)
})
