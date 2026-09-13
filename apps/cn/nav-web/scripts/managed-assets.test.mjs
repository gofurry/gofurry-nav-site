import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { assetURL, assetCandidate, chooseAssetCDN, probeAssetCDNs, ASSET_PROBE_SHA256 } from '../app/utils/managedAssets.ts'
import { parseBackgroundPreference, patternAppearance } from '../app/utils/backgroundPreferences.ts'

const origins = { primary: 'https://primary.example.com', mirror: 'https://mirror.example.com' }
const key = `nav/hero/desktop/${'a'.repeat(32)}.avif`
assert.equal(assetURL(origins, 'primary', key), `${origins.primary}/${key}`)
for (const invalid of ['https://other.example.com/x', '../icon.svg', 'old-icon.ico', 'nav/bg/standard-bg-1.avif']) assert.equal(assetURL(origins, 'primary', invalid), '')
assert.equal(chooseAssetCDN(100, 90), 'primary')
assert.equal(chooseAssetCDN(100, 80), 'mirror')
assert.equal(chooseAssetCDN(400, 350), 'mirror')
assert.equal(chooseAssetCDN(null, 400), 'mirror')
assert.equal(chooseAssetCDN(100, null), 'primary')
assert.equal(chooseAssetCDN(null, null), 'primary')
assert.equal(assetCandidate(origins, 'primary', key, new Set(['primary']), '/defaultLogo.svg').provider, 'mirror')
assert.equal(assetCandidate(origins, 'mirror', key, new Set(['primary', 'mirror']), '/defaultLogo.svg').url, '/defaultLogo.svg')
assert.equal(assetCandidate(origins, 'primary', null, new Set(), '').url, '')
assert.equal(assetCandidate(origins, 'primary', 'old-icon.ico', new Set(), '/defaultLogo.svg').url, '/defaultLogo.svg')

const probe = readFileSync(new URL('./fixtures/cdn-probe.bin', import.meta.url))
assert.equal(probe.length, 8192)
assert.equal(createHash('sha256').update(probe).digest('hex'), ASSET_PROBE_SHA256)
let consumed = 0
const requests = []
await probeAssetCDNs(origins, async (url, options) => {
  assert.equal(options.credentials, 'omit')
  assert.equal(options.mode, 'cors')
  assert.match(url, /\/system\/probes\/cdn\.bin\?probe=/)
  requests.push(url)
  return { ok: true, arrayBuffer: async () => { consumed++; return probe.buffer.slice(probe.byteOffset, probe.byteOffset + probe.byteLength) } }
})
assert.equal(requests.length, 4)
assert.equal(consumed, 4, 'every probe must consume the complete body')
assert.ok(requests[0].startsWith(origins.primary) && requests[1].startsWith(origins.mirror))
const failure = await probeAssetCDNs(origins, async () => new Response(new Uint8Array(8192)))
assert.equal(failure.primaryMs, null, 'wrong probe bytes must not count as a healthy CDN')
assert.equal(failure.mirrorMs, null)
console.log('Managed asset key, fallback, selection, and full-body probe contracts passed')

const preference = parseBackgroundPreference(JSON.stringify({ version: 1, source: 'server', pattern_id: '9007199254740993', overrides: { opacity: 0, size_px: 200 }, blob: 'must not persist', object_key: key, light_color: '#ffffff' }))
assert.deepEqual(preference, { version: 1, source: 'server', pattern_id: '9007199254740993', overrides: { opacity: 0, size_px: 200 } })
assert.deepEqual(parseBackgroundPreference('{"version":2,"source":"local"}'), { version: 1, source: 'default', overrides: {} })
assert.deepEqual(parseBackgroundPreference('{"version":1,"source":"local","overrides":{"opacity":2,"size_px":0,"color":"url(https://x)"}}').overrides, {})
const defaults = { light_color: '#123456', dark_color: '#abcdef', light_opacity: 0.1, dark_opacity: 0.2, default_size_px: 160 }
assert.deepEqual(patternAppearance(defaults, 'dark', { opacity: 0 }), { color: '#abcdef', opacity: 0, size: 160 })
assert.deepEqual(patternAppearance({ ...defaults, default_size_px: 240 }, 'light', {}), { color: '#123456', opacity: 0.1, size: 240 }, 'catalog updates must flow through when no override exists')
console.log('Background preference normalization and explicit override contracts passed')
