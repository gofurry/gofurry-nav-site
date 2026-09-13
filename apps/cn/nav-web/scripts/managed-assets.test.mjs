import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { assetURL, assetCandidate, chooseAssetCDN, probeAssetCDNs, ASSET_PROBE_SHA256 } from '../app/utils/managedAssets.ts'

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
