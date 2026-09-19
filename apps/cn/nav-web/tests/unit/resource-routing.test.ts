import { afterEach, describe, expect, it, vi } from 'vitest'
import { resolveAssetPreferred, probeAssetCDNs, readAssetDiagnostics } from '../../app/utils/managedAssets'
import { resolveSteamPreferred, steamSharedAssetCandidates, STEAM_SHARED_CDN_GROUP_PREFIXES } from '../../app/utils/steamAssets'
import { chooseSteamGroup, legacySteamRecommendation, probeSteamGroups, readSteamDiagnostics, STEAM_PROBE_PATHS } from '../../app/utils/steamAssetRouting'
import { averageProbe, createRouteProbe, isFresh, ROUTE_TTL_MS, type ProbeRecord } from '../../app/utils/resourceRouting'

const stamp = 1_000_000_000
const source = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/570/hash/header.jpg?v=2#cover'
afterEach(() => vi.unstubAllGlobals())

describe('route policy and history', () => {
  it('keeps pins separate from recommendations and retains history on failed probes', async () => {
    for (const preferred of ['primary', 'mirror'] as const) {
      expect(resolveAssetPreferred('auto', preferred)).toBe(preferred)
      expect(resolveAssetPreferred('primary', preferred)).toBe('primary')
      expect(resolveAssetPreferred('mirror', preferred)).toBe('mirror')
    }
    expect(resolveAssetPreferred('auto', null)).toBe('primary')
    const failed = await probeAssetCDNs({ primary: 'https://primary.example', mirror: 'https://mirror.example' }, async () => new Response('', { status: 503 }), () => 0, 'mirror')
    expect(failed.selected).toBe('mirror')
    expect(failed.primaryState).toBe('failed')
    expect(failed.mirrorState).toBe('failed')
    expect(readAssetDiagnostics({ ...failed, primaryMs: 'bad' })).toBeNull()
    expect(readAssetDiagnostics({ ...failed, primaryState: undefined })?.primaryState).toBe('failed')
  })

  it('keeps Steam logical groups and path/query/hash through all fallback hosts', () => {
    for (const group of ['china', 'global'] as const) {
      expect(resolveSteamPreferred('auto', group, group === 'china' ? 'en' : 'zh')).toBe(group)
      for (const pin of ['china', 'global'] as const) expect(resolveSteamPreferred(pin, group, 'zh')).toBe(pin)
      const candidates = steamSharedAssetCandidates(source, group)
      const other = group === 'china' ? 'global' : 'china'
      expect(candidates.map(url => new URL(url).origin)).toEqual([...STEAM_SHARED_CDN_GROUP_PREFIXES[group], ...STEAM_SHARED_CDN_GROUP_PREFIXES[other]])
      expect(candidates.every(url => url.endsWith('/store_item_assets/steam/apps/570/hash/header.jpg?v=2#cover'))).toBe(true)
    }
    expect(resolveSteamPreferred('auto', null, 'en')).toBe('global')
    expect(resolveSteamPreferred('auto', null, 'zh')).toBe('china')
    expect(steamSharedAssetCandidates('https://other.example/a.jpg')).toEqual(['https://other.example/a.jpg'])
    expect(steamSharedAssetCandidates(null)).toEqual([])
    expect(chooseSteamGroup(null, null, 'global')).toBe('global')
    expect(chooseSteamGroup(80, 40, 'china')).toBe('global')
    expect(chooseSteamGroup(40, null, 'global')).toBe('china')
  })

  it('migrates fresh legacy history only as recommendation', () => {
    const legacy = { group: 'global', testedAt: stamp }
    expect(legacySteamRecommendation(legacy, stamp)).toBe('global')
    expect(resolveSteamPreferred('auto', legacySteamRecommendation(legacy, stamp), 'zh')).toBe('global')
    for (const raw of [null, {}, { group: 'pin', testedAt: stamp }, { group: 'china', testedAt: stamp + 1 }, { group: 'china', testedAt: stamp - 6 * 3600_000 }]) {
      expect(legacySteamRecommendation(raw, stamp)).toBeNull()
    }
  })
})

describe('Steam probes', () => {
  it('runs two sequential rounds with parallel groups and unique automatic/manual URLs', async () => {
    const urls: string[] = []
    let active = 0, maxActive = 0
    const result = await probeSteamGroups('china', async url => {
      urls.push(url); maxActive = Math.max(maxActive, ++active)
      await Promise.resolve(); active--
      return { state: 'success', ms: url.includes('eccdnx') ? 80 : 20 }
    }, () => stamp)
    expect(result.selected).toBe('global')
    expect(result.global.ms).toBe(20)
    expect(maxActive).toBe(2)
    expect(urls).toHaveLength(4)
    expect(new Set(urls).size).toBe(4)
    expect(urls.every(url => url.includes(STEAM_PROBE_PATHS[0]!))).toBe(true)
    const manualURLs: string[] = []
    await probeSteamGroups('global', async url => { manualURLs.push(url); return { state: 'success', ms: 20 } }, () => stamp, true)
    expect(manualURLs.every(url => !urls.includes(url))).toBe(true)
  })

  it('uses bounded sample fallback and retains timeout diagnostics/history', async () => {
    const urls: string[] = []
    const fallback = await probeSteamGroups('global', async url => {
      urls.push(url)
      return url.includes('/570/') ? { state: 'failed', ms: null } : { state: 'success', ms: 20 }
    })
    expect(urls).toHaveLength(8)
    expect(fallback.sample).toBe(STEAM_PROBE_PATHS[1])
    urls.length = 0
    const exhausted = await probeSteamGroups('global', async url => { urls.push(url); return { state: 'timeout', ms: null } })
    expect(urls).toHaveLength(16)
    expect(exhausted.selected).toBe('global')
    expect(exhausted.china.state).toBe('timeout')
    expect(readSteamDiagnostics(exhausted)).toEqual(exhausted)
    expect(readSteamDiagnostics({ ...exhausted, china: {} })).toBeNull()
  })

  it('averages complete successful rounds and preserves failure/timeout semantics', () => {
    expect(averageProbe([{ state: 'success', ms: 10 }, { state: 'success', ms: 30 }])).toEqual({ state: 'success', ms: 20 })
    expect(averageProbe([{ state: 'success', ms: 10 }, { state: 'failed', ms: null }])).toEqual({ state: 'failed', ms: null })
    expect(averageProbe([{ state: 'failed', ms: null }, { state: 'timeout', ms: null }])).toEqual({ state: 'timeout', ms: null })
  })
})

it('schedules only after mount/idle, deduplicates, preserves TTL/cooldown, and allows explicit offline probes', async () => {
  let clock = stamp, runs = 0
  let saved: ProbeRecord | null = null
  let finish!: () => void
  const idle: (() => void)[] = []
  const browserNavigator = { onLine: true, connection: { saveData: false } }
  vi.stubGlobal('navigator', browserNavigator)
  vi.stubGlobal('window', { requestIdleCallback: (callback: () => void) => idle.push(callback) })
  const runner = createRouteProbe(() => {
    runs++
    return new Promise<ProbeRecord>(resolve => { finish = () => resolve({ checkedAt: clock }) })
  }, value => { saved = value }, () => clock)
  await runner.probe(true)
  expect(runs).toBe(0)
  runner.start(null)
  runner.scheduleProbe()
  expect(idle).toHaveLength(1)
  const pending = runner.probe(true)
  expect(runner.probe(true)).toBe(pending)
  expect(runs).toBe(1)
  finish(); await pending
  idle.shift()!()
  expect(runs).toBe(1)
  expect(saved).toMatchObject({ lastManualAt: clock })
  const restoredRun = vi.fn(async () => ({ checkedAt: clock }))
  const restored = createRouteProbe(restoredRun, () => {}, () => clock)
  restored.start(saved)
  await restored.probe(true)
  expect(restoredRun).not.toHaveBeenCalled()
  clock += 59_999
  await runner.probe(true)
  expect(runs).toBe(1)
  clock++
  const second = runner.probe(true); finish(); await second
  expect(runs).toBe(2)
  expect(isFresh(clock, clock + ROUTE_TTL_MS - 1)).toBe(true)
  expect(isFresh(clock, clock + ROUTE_TTL_MS)).toBe(false)
  browserNavigator.onLine = false
  runner.markStale()
  expect(idle).toHaveLength(0)
  expect(runner.diagnostics.value?.checkedAt).toBe(clock)
  clock += 60_000
  const offlineManual = runner.probe(true); finish(); await offlineManual
  expect(runs).toBe(3)
  browserNavigator.onLine = true; browserNavigator.connection.saveData = true
  runner.markStale()
  expect(idle).toHaveLength(0)
  browserNavigator.connection.saveData = false
  runner.scheduleProbe()
  expect(idle).toHaveLength(1)
})
