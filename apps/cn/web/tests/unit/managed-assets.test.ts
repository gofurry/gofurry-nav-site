import { createHash } from 'node:crypto'
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { assetURL, assetCandidate, chooseAssetCDN, probeAssetCDNs, ASSET_PROBE_SHA256 } from '../../app/utils/managedAssets'
import { parseBackgroundPreference, patternAppearance } from '../../app/utils/backgroundPreferences'

const origins = { primary: 'https://primary.example.com', mirror: 'https://mirror.example.com' }
const key = `nav/hero/desktop/${'a'.repeat(32)}.avif`

describe('managed assets', () => {
  it('accepts managed keys and rejects external, traversing and legacy paths', () => {
    expect(assetURL(origins, 'primary', key)).toBe(`${origins.primary}/${key}`)
    for (const invalid of ['https://other.example.com/x', '../icon.svg', 'old-icon.ico', 'nav/bg/standard-bg-1.avif']) {
      expect(assetURL(origins, 'primary', invalid)).toBe('')
    }
  })

  it.each([
    [100, 90, 'primary'], [100, 80, 'mirror'], [400, 350, 'mirror'],
    [null, 400, 'mirror'], [100, null, 'primary'], [null, null, 'primary'],
  ] as const)('chooses CDN for primary=%s mirror=%s', (primary, mirror, expected) => {
    expect(chooseAssetCDN(primary, mirror)).toBe(expected)
  })

  it('falls back in order without retrying duplicate URLs', () => {
    expect(assetCandidate(origins, 'primary', key, new Set(['primary']), '/defaultLogo.svg').provider).toBe('mirror')
    expect(assetCandidate(origins, 'mirror', key, new Set(['primary', 'mirror']), '/defaultLogo.svg').url).toBe('/defaultLogo.svg')
    expect(assetCandidate(origins, 'primary', null, new Set(), '').url).toBe('')
    expect(assetCandidate(origins, 'primary', 'old-icon.ico', new Set(), '/defaultLogo.svg').url).toBe('/defaultLogo.svg')
    expect(assetCandidate({ primary: origins.primary, mirror: origins.primary }, 'primary', key, new Set(['primary']), '/defaultLogo.svg').url).toBe('/defaultLogo.svg')
  })

  it('validates the probe fixture and consumes every full CORS response body', async () => {
    const probe = readFileSync(new URL('../fixtures/cdn-probe.bin', import.meta.url))
    expect(probe.length).toBe(8192)
    expect(createHash('sha256').update(probe).digest('hex')).toBe(ASSET_PROBE_SHA256)
    let consumed = 0
    const requests: string[] = []
    await probeAssetCDNs(origins, async (url, options) => {
      expect(options?.credentials).toBe('omit')
      expect(options?.mode).toBe('cors')
      expect(String(url)).toMatch(/\/system\/probes\/cdn\.bin\?probe=/)
      requests.push(String(url))
      const response = new Response(new Uint8Array(probe))
      const body = response.arrayBuffer.bind(response)
      response.arrayBuffer = () => { consumed++; return body() }
      return response
    })
    expect(requests).toHaveLength(4)
    expect(consumed).toBe(4)
    expect(requests[0]?.startsWith(origins.primary) && requests[1]?.startsWith(origins.mirror)).toBe(true)
    const failure = await probeAssetCDNs(origins, async () => new Response(new Uint8Array(8192)))
    expect(failure.primaryMs).toBeNull()
    expect(failure.mirrorMs).toBeNull()
  })
})

describe('background preferences', () => {
  it('keeps string IDs and only explicit, valid overrides', () => {
    expect(parseBackgroundPreference(JSON.stringify({ version: 1, source: 'server', pattern_id: '9007199254740993', overrides: { opacity: 0, size_px: 200 }, blob: 'must not persist', object_key: key, light_color: '#ffffff' })))
      .toEqual({ version: 1, source: 'server', pattern_id: '9007199254740993', overrides: { opacity: 0, size_px: 200 } })
    expect(parseBackgroundPreference('{"version":2,"source":"local"}')).toEqual({ version: 1, source: 'default', overrides: {} })
    expect(parseBackgroundPreference('{"version":1,"source":"local","overrides":{"opacity":2,"size_px":0,"color":"url(https://x)"}}').overrides).toEqual({})
  })

  it('preserves zero opacity and follows catalog updates without overrides', () => {
    const defaults = { light_color: '#123456', dark_color: '#abcdef', light_opacity: 0.1, dark_opacity: 0.2, default_size_px: 160 }
    expect(patternAppearance(defaults, 'dark', { opacity: 0 })).toEqual({ color: '#abcdef', opacity: 0, size: 160 })
    expect(patternAppearance({ ...defaults, default_size_px: 240 }, 'light', {})).toEqual({ color: '#123456', opacity: 0.1, size: 240 })
  })
})
