import { describe, expect, it } from 'vitest'
import { queryString, toRFC3339 } from './collection-query'

describe('Collection parity filters', () => {
  it('preserves chart, history, and result filter parameters', () => {
    expect(queryString({ window: '30d', domain: 'nav', job_key: 'nav.daily', target: 'example.org' }))
      .toBe('window=30d&domain=nav&job_key=nav.daily&target=example.org')
  })

  it('normalizes local date-time input to the backend RFC3339 contract', () => {
    const value = '2026-08-31T12:34'
    expect(toRFC3339(value)).toBe(new Date(value).toISOString())
    expect(toRFC3339('')).toBe('')
  })
})
