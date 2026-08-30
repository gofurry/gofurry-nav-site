import { describe, expect, it } from 'vitest'
import { eventSentence, metricStateLabel, percent, runCoverage } from './presentation'

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
})
