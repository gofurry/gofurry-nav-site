import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { ref } from 'vue'
import ts from 'typescript'
import { normalizeHeroID, normalizeHeroPreference, heroQuery, heroPreferenceKey } from '../app/utils/heroPreferences.ts'

for (const value of [undefined, null, 123, '', '0', '-1', '01', '1.1', '+1', ' 1', '1e3', '9223372036854775808']) assert.equal(normalizeHeroID(value), null)
for (const value of ['1', '9007199254740993', '9223372036854775807']) assert.equal(normalizeHeroID(value), value)
assert.equal(normalizeHeroPreference('invalid', null, null).mode, 'random')
assert.deepEqual(heroQuery(normalizeHeroPreference('local', '10', '20')), { hero_mode: 'local' })
assert.deepEqual(heroQuery(normalizeHeroPreference('random', '10', '20')), {})
assert.deepEqual(heroQuery(normalizeHeroPreference('fixed', '9007199254740993', '20')), { hero_desktop_id: '9007199254740993', hero_mobile_id: '20' })

// Execute the actual cookie/state composable; keep bigint decoding and unrelated
// route preferences outside Nuxt's default JSON number coercion.
const source = readFileSync(new URL('../app/composables/useHeroPreferences.ts', import.meta.url), 'utf8')
  .replace("'~/utils/heroPreferences'", JSON.stringify(new URL('../app/utils/heroPreferences.ts', import.meta.url).href))
const compiled = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText
const { useHeroPreferences } = await import('data:text/javascript;base64,' + Buffer.from(compiled).toString('base64'))
const states = new Map(), cookies = new Map(), options = new Map()
globalThis.useState = (name, init) => { if (!states.has(name)) states.set(name, ref(init())); return states.get(name) }
globalThis.useCookie = (name, config) => { options.set(name, config); if (!cookies.has(name)) cookies.set(name, ref(undefined)); return cookies.get(name) }
cookies.set('gf_hero_mode', ref('fixed'))
cookies.set('gf_hero_desktop_id', ref('9007199254740993'))
cookies.set('gf_hero_mobile_id', ref('20'))
cookies.set('gf_asset_cdn_mode', ref('mirror'))
const settings = useHeroPreferences()
assert.equal(settings.legacyPending.value, false)
assert.equal(options.get('gf_hero_desktop_id').decode('9007199254740993'), '9007199254740993')
assert.equal(heroPreferenceKey(settings.preference.value), 'fixed:9007199254740993:20')
settings.save({ ...settings.preference.value, mode: 'random' })
assert.equal(cookies.get('gf_hero_desktop_id').value, '9007199254740993', 'leaving Fixed must retain its IDs')
assert.equal(settings.revision.value, 1)
settings.save(settings.preference.value)
assert.equal(settings.revision.value, 1, 'unrelated modal Save must not reselect Hero')
assert.equal(cookies.get('gf_asset_cdn_mode').value, 'mirror')
assert.deepEqual([...options.keys()], ['gf_hero_mode', 'gf_hero_desktop_id', 'gf_hero_mobile_id'])
states.clear(); cookies.clear()
const fresh = useHeroPreferences()
assert.equal(fresh.legacyPending.value, true)
assert.equal(fresh.preference.value.mode, 'random')
fresh.save(fresh.preference.value)
assert.equal(fresh.legacyPending.value, false)
console.log('Hero preference normalization, string ID cookies, SSR query and isolated Save contracts PASS')
