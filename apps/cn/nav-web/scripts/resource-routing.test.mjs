import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createRenderer, h, nextTick, ref } from 'vue'
import { createI18n } from 'vue-i18n'
import ts from 'typescript'
import { resolveAssetPreferred, probeAssetCDNs, readAssetDiagnostics } from '../app/utils/managedAssets.ts'
import { resolveSteamPreferred, steamSharedAssetCandidates, STEAM_SHARED_CDN_GROUP_PREFIXES } from '../app/utils/steamAssets.ts'
import { chooseSteamGroup, legacySteamRecommendation, probeSteamGroups, readSteamDiagnostics, STEAM_PROBE_PATHS } from '../app/utils/steamAssetRouting.ts'
import { createRouteProbe, isFresh, ROUTE_TTL_MS } from '../app/utils/resourceRouting.ts'

for (const preferred of ['primary', 'mirror']) {
  assert.equal(resolveAssetPreferred('auto', preferred), preferred)
  assert.equal(resolveAssetPreferred('primary', preferred), 'primary')
  assert.equal(resolveAssetPreferred('mirror', preferred), 'mirror')
}
assert.equal(resolveAssetPreferred('auto', null), 'primary')
const failedProbe = await probeAssetCDNs({ primary: 'https://primary.example', mirror: 'https://mirror.example' }, async () => new Response('', { status: 503 }), () => 0, 'mirror')
assert.equal(failedProbe.selected, 'mirror', 'both failures retain the previous recommendation')
assert.equal(failedProbe.primaryState, 'failed')
assert.equal(failedProbe.mirrorState, 'failed')
assert.equal(readAssetDiagnostics({ ...failedProbe, primaryMs: 'bad' }), null)
assert.equal(readAssetDiagnostics({ ...failedProbe, primaryState: undefined }).primaryState, 'failed', 'old diagnostics should remain readable')

const source = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/570/hash/header.jpg?v=2#cover'
for (const group of ['china', 'global']) {
  assert.equal(resolveSteamPreferred('auto', group, group === 'china' ? 'en' : 'zh'), group, 'locale cannot override history')
  for (const pin of ['china', 'global']) assert.equal(resolveSteamPreferred(pin, group, 'zh'), pin)
  const candidates = steamSharedAssetCandidates(source, group)
  const other = group === 'china' ? 'global' : 'china'
  assert.deepEqual(candidates.map(url => new URL(url).origin), [...STEAM_SHARED_CDN_GROUP_PREFIXES[group], ...STEAM_SHARED_CDN_GROUP_PREFIXES[other]])
  assert(candidates.every(url => url.endsWith('/store_item_assets/steam/apps/570/hash/header.jpg?v=2#cover')))
}
assert.equal(resolveSteamPreferred('auto', null, 'en'), 'global')
assert.equal(resolveSteamPreferred('auto', null, 'zh'), 'china')
assert.deepEqual(steamSharedAssetCandidates('https://other.example/a.jpg'), ['https://other.example/a.jpg'])
assert.deepEqual(steamSharedAssetCandidates(null), [])
assert.equal(chooseSteamGroup(null, null, 'global'), 'global')
assert.equal(chooseSteamGroup(80, 40, 'china'), 'global')
assert.equal(chooseSteamGroup(40, null, 'global'), 'china')

const stamp = 1_000_000_000
assert.equal(legacySteamRecommendation({ group: 'global', testedAt: stamp }, stamp), 'global')
assert.equal(resolveSteamPreferred('auto', legacySteamRecommendation({ group: 'global', testedAt: stamp }, stamp), 'zh'), 'global')
for (const raw of [null, {}, { group: 'pin', testedAt: stamp }, { group: 'china', testedAt: stamp + 1 }, { group: 'china', testedAt: stamp - 6 * 3600_000 }]) assert.equal(legacySteamRecommendation(raw, stamp), null)

let urls = []
let active = 0, maxActive = 0
const success = await probeSteamGroups('china', async url => {
  urls.push(url); maxActive = Math.max(maxActive, ++active)
  await Promise.resolve(); active--
  return { state: 'success', ms: url.includes('eccdnx') ? 80 : 20 }
}, () => stamp)
assert.equal(success.selected, 'global')
assert.equal(success.global.ms, 20)
assert.equal(maxActive, 2, 'groups run in parallel, rounds stay sequential')
assert.equal(urls.length, 4)
assert.equal(new Set(urls).size, 4, 'the two rounds must not reuse the same cached URL')
assert(urls.every(url => url.includes(STEAM_PROBE_PATHS[0])))
const automaticURLs = [...urls]
urls = []
await probeSteamGroups('global', async url => { urls.push(url); return { state: 'success', ms: 20 } }, () => stamp, true)
assert(urls.every(url => !automaticURLs.includes(url)), 'manual test within the same minute must not reuse automatic image cache')
urls = []
const fallbackSample = await probeSteamGroups('global', async url => {
  urls.push(url)
  return url.includes('/570/') ? { state: 'failed', ms: null } : { state: 'success', ms: 20 }
})
assert.equal(urls.length, 8)
assert.equal(fallbackSample.sample, STEAM_PROBE_PATHS[1])
urls = []
const exhausted = await probeSteamGroups('global', async url => { urls.push(url); return { state: 'timeout', ms: null } })
assert.equal(urls.length, 16, 'sample fallback must be bounded')
assert.equal(exhausted.selected, 'global')
assert.equal(exhausted.china.state, 'timeout')
assert.deepEqual(readSteamDiagnostics(exhausted), exhausted)
assert.equal(readSteamDiagnostics({ ...exhausted, china: {} }), null)

// Exercise actual Vue composables, not a duplicate of their snapshot algorithm.
async function loadComposable(name) {
  let source = readFileSync(new URL('../app/composables/' + name + '.ts', import.meta.url), 'utf8')
  source = source.replace(/from ['"](~\/[^'"]+|vue(?:-i18n)?)['"]/g, (_match, specifier) => {
    const url = specifier.startsWith('~/') ? new URL('../app/' + specifier.slice(2) + '.ts', import.meta.url).href : import.meta.resolve(specifier)
    return 'from ' + JSON.stringify(url)
  })
  const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText
  return import('data:text/javascript;base64,' + Buffer.from(compiled).toString('base64'))
}
const { useManagedAsset } = await loadComposable('useManagedAsset')
const { useSteamAsset } = await loadComposable('useSteamAsset')
const renderer = createRenderer({
  createElement: () => ({ props: {} }), createText: text => ({ text }), createComment: text => ({ text }),
  insert() {}, remove() {}, setText() {}, setElementText() {}, parentNode: () => null, nextSibling: () => null,
  patchProp(el, key, _old, value) { el.props[key] = value },
})
const assetMode = ref('auto'), assetRecommendation = ref('primary')
const steamMode = ref('auto'), steamRecommendation = ref('china')
let failures = 0
globalThis.useNuxtApp = () => ({
  $assetCDN: { origins: { primary: 'https://primary.example', mirror: 'https://mirror.example' }, resolvePreferred: () => resolveAssetPreferred(assetMode.value, assetRecommendation.value), reportFailure: () => { failures++ } },
  $steamAssetRoute: { resolvePreferred: locale => resolveSteamPreferred(steamMode.value, steamRecommendation.value, locale), reportFailure: () => { failures++ } },
})
const key = ref('nav/hero/desktop/' + 'a'.repeat(32) + '.avif'), steamSource = ref(source)
let managed, steam
const app = renderer.createApp({ setup() { managed = useManagedAsset(key, '/default.svg'); steam = useSteamAsset(steamSource); return () => h('img', { src: managed.src.value, steam: steam.src.value }) } })
app.use(createI18n({ legacy: false, locale: 'zh', messages: { zh: {}, en: {} } }))
app.mount({})
const originalManaged = managed.src.value, originalSteam = steam.src.value
assetRecommendation.value = 'mirror'; steamRecommendation.value = 'global'
await nextTick()
assert.equal(managed.src.value, originalManaged, 'automatic/manual recommendation must not reload the current key')
assert.equal(steam.src.value, originalSteam, 'automatic/manual recommendation must not reload the current src')
assetMode.value = 'mirror'; steamMode.value = 'global'
await nextTick()
assert.equal(managed.src.value, originalManaged, 'Save must not reload managed images')
assert.equal(steam.src.value, originalSteam, 'Save must not reload Steam images')
key.value = 'nav/hero/desktop/' + 'b'.repeat(32) + '.avif'
steamSource.value = source.replace('header.jpg', 'capsule.jpg')
await nextTick()
assert(managed.src.value.startsWith('https://mirror.example/'))
assert(steam.src.value.startsWith(STEAM_SHARED_CDN_GROUP_PREFIXES.global[0]))
managed.onError()
assert(managed.src.value.startsWith('https://primary.example/'))
managed.onError()
assert.equal(managed.src.value, '/default.svg')
assert.equal(assetMode.value, 'mirror', 'fallback must not clear a pin')
for (const prefix of [...STEAM_SHARED_CDN_GROUP_PREFIXES.global, ...STEAM_SHARED_CDN_GROUP_PREFIXES.china].slice(1)) {
  assert.equal(steam.onError(), true)
  assert(steam.src.value.startsWith(prefix))
}
assert.equal(steam.onError(), false)
assert.equal(steamMode.value, 'global')
assert(failures > 0)
app.unmount()
delete globalThis.useNuxtApp

// Shared probe scheduling: mount/idle only, deduplication, TTL, persisted cooldown,
// offline/save-data automatic skip, manual override.
let clock = stamp, runs = 0, saved = null, finish
const idle = []
const navigatorDescriptor = Object.getOwnPropertyDescriptor(globalThis, 'navigator')
Object.defineProperty(globalThis, 'navigator', { configurable: true, value: { onLine: true, connection: { saveData: false } } })
globalThis.window = { requestIdleCallback: callback => idle.push(callback) }
const runner = createRouteProbe(() => { runs++; return new Promise(resolve => { finish = () => resolve({ checkedAt: clock }) }) }, value => { saved = value }, () => clock)
await runner.probe(true)
assert.equal(runs, 0)
runner.start(null)
runner.scheduleProbe()
assert.equal(idle.length, 1)
const pending = runner.probe(true)
assert.strictEqual(runner.probe(true), pending)
assert.equal(runs, 1)
finish(); await pending
idle.shift()()
assert.equal(runs, 1, 'queued automatic probe must not run after a fresh manual test')
assert.equal(saved.lastManualAt, clock)
const restored = createRouteProbe(async () => { throw new Error('cooldown was lost across reload') }, () => {}, () => clock)
restored.start(saved)
await restored.probe(true)
clock += 59_999
await runner.probe(true)
assert.equal(runs, 1)
clock++
const second = runner.probe(true); finish(); await second
assert.equal(runs, 2)
assert(isFresh(clock, clock + ROUTE_TTL_MS - 1))
assert(!isFresh(clock, clock + ROUTE_TTL_MS))
navigator.onLine = false
runner.markStale()
assert.equal(idle.length, 0)
assert.equal(runner.diagnostics.value.checkedAt, clock, 'stale diagnostics must retain the real checked time')
clock += 60_000
const offlineManual = runner.probe(true); finish(); await offlineManual
assert.equal(runs, 3, 'manual/explicit probes may run offline and report failure normally')
navigator.onLine = true; navigator.connection.saveData = true
runner.markStale()
assert.equal(idle.length, 0)
navigator.connection.saveData = false
runner.scheduleProbe()
assert.equal(idle.length, 1)
delete globalThis.window
if (navigatorDescriptor) Object.defineProperty(globalThis, 'navigator', navigatorDescriptor)
else delete globalThis.navigator
console.log('Resource routing: resolver, probes, legacy migration, actual Vue snapshots, fallback, mount/idle, TTL and cooldown PASS')
