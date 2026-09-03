import { describe, expect, it } from 'vitest'
import { ACTIVE_COLLECTION_REFRESH_MS, collectionJobProgress, eventSentence, metricStateLabel, percent, runCoverage } from './presentation'

describe('operational presentation semantics', () => {
  it('never presents unknown metric state as false', () => {
    expect(metricStateLabel('unknown')).toBe('未知')
    expect(metricStateLabel('unknown')).not.toBe('不支持')
  })

  it('keeps zero-denominator coverage unavailable', () => {
    expect(runCoverage({ expected_count: 0, success_count: 0 })).toBeNull()
    expect(percent(runCoverage({ expected_count: 0, success_count: 0 }))).toBe('—')
    expect(percent(runCoverage({ expected_count: 4, success_count: 0 }))).toBe('0.0%')
  })

  it('renders a human-readable event sentence', () => {
    expect(eventSentence('FurryFans', 'ipv6_enabled')).toBe('FurryFans 启用了 IPv6')
  })

  it('normalizes live collection progress without inventing queued progress', () => {
    expect(collectionJobProgress('running', { attempted: 3, expected: 8 })).toEqual({ state: 'progress', attempted: 3, expected: 8, percentage: 38 })
    expect(collectionJobProgress('running')).toEqual({ state: 'waiting', attempted: 0, expected: 0, percentage: null })
    expect(collectionJobProgress('queued', { attempted: 1, expected: 2 })).toEqual({ state: 'queued', attempted: 0, expected: 0, percentage: null })
    expect(ACTIVE_COLLECTION_REFRESH_MS).toBeGreaterThanOrEqual(2_000)
    expect(ACTIVE_COLLECTION_REFRESH_MS).toBeLessThanOrEqual(3_000)
  })
})
