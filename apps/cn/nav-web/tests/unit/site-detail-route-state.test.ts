import { describe, expect, it } from 'vitest'
import {
  buildSiteDetailQuery, parseSiteDetailRouteState, selectSiteDetailTab, selectSiteDetailTarget, selectSiteObservationView,
} from '../../app/utils/siteDetailRouteState'
import { siteEntityPath, siteTargetPath } from '../../app/utils/siteRoutes'

describe('Site Detail route state', () => {
  it('selects Observation views through the owner while retaining Target and clearing foreign state', () => {
    const state = parseSiteDetailRouteState({ domain: 'a.example', tab: 'insights', metric: 'csp', range: 'all' })
    expect(buildSiteDetailQuery(selectSiteObservationView(state, 'http'))).toEqual({ domain: 'a.example', tab: 'observation', view: 'http' })
    expect(buildSiteDetailQuery(selectSiteObservationView(state, 'overview'))).toEqual({ domain: 'a.example', tab: 'observation' })
  })
  it('defaults to the entity overview and omits default query values', () => {
    expect(parseSiteDetailRouteState({})).toEqual({ domain: '', tab: 'overview' })
    for (const query of [{}, { tab: 'overview' }, { tab: 'banana', view: 'dns', metric: 'csp', range: 'all' }]) {
      expect(buildSiteDetailQuery(parseSiteDetailRouteState(query))).toEqual({})
    }
    expect(buildSiteDetailQuery(parseSiteDetailRouteState({ tab: 'observation', view: 'overview' }))).toEqual({ tab: 'observation' })
    expect(buildSiteDetailQuery(parseSiteDetailRouteState({ tab: 'security', view: 'overview' }))).toEqual({ tab: 'security' })
    expect(buildSiteDetailQuery(parseSiteDetailRouteState({ tab: 'insights', metric: 'ipv6', range: '30d' }))).toEqual({ tab: 'insights' })
  })

  it.each([
    ['observation', 'overview'], ['observation', 'performance'], ['observation', 'http'],
    ['observation', 'dns'], ['observation', 'web'],
    ['security', 'overview'], ['security', 'tls'], ['security', 'web'], ['security', 'exposure'],
  ])('accepts %s/%s and removes foreign workspace query', (tab, view) => {
    const state = parseSiteDetailRouteState({ tab, view, domain: ' a.example ', metric: 'tls13', range: 'all', extra: 'x' })
    expect(state).toEqual({ domain: 'a.example', tab, view })
    const query = buildSiteDetailQuery(state)
    expect(query).not.toHaveProperty('metric')
    expect(query).not.toHaveProperty('range')
    expect(query).not.toHaveProperty('extra')
    expect(parseSiteDetailRouteState(query)).toEqual(state)
  })

  for (const metric of ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified']) {
    it.each(['30d', '90d', 'all'])(`accepts ${metric}/%s and removes view`, range => {
      const state = parseSiteDetailRouteState({ tab: 'insights', metric, range, view: 'dns' })
      expect(state).toEqual({ domain: '', tab: 'insights', metric, range })
      expect(parseSiteDetailRouteState(buildSiteDetailQuery(state))).toEqual(state)
    })
  }

  it.each([
    [{ tab: 'security', view: 'dns' }, { domain: '', tab: 'security', view: 'overview' }],
    [{ tab: 'observation', view: 'tls' }, { domain: '', tab: 'observation', view: 'overview' }],
    [{ tab: 'insights', metric: 'banana', range: '7d' }, { domain: '', tab: 'insights', metric: 'ipv6', range: '30d' }],
    [{ tab: null, domain: null }, { domain: '', tab: 'overview' }],
    [{ tab: ['security', 'insights'], view: ['tls', 'dns'], domain: ['first.example', 'second.example'] },
      { domain: 'first.example', tab: 'security', view: 'tls' }],
    [{ tab: [null, 'insights'], domain: [null, 'second.example'] }, { domain: '', tab: 'overview' }],
    [{ tab: 1, domain: {}, view: true }, { domain: '', tab: 'overview' }],
  ])('normalizes invalid and repeated UI state %j', (query, expected) => {
    expect(parseSiteDetailRouteState(query)).toEqual(expected)
  })

  it.each([
    { tab: 'overview' }, { tab: 'observation', view: 'dns' }, { tab: 'security', view: 'tls' },
    { tab: 'insights', metric: 'certificate_verified', range: '90d' },
  ])('preserves the active workspace when switching target: %j', query => {
    const original = parseSiteDetailRouteState({ ...query, domain: 'a.example' })
    const selected = selectSiteDetailTarget(original, 'b.example')
    expect(selected).toEqual({ ...original, domain: 'b.example' })
    expect(parseSiteDetailRouteState(buildSiteDetailQuery(selected))).toEqual(selected)
    expect(original.domain).toBe('a.example')
    const url = new URL(siteTargetPath(41, 'b.example', { ...query, domain: 'a.example' }), 'https://go-furry.com')
    expect(parseSiteDetailRouteState(Object.fromEntries(url.searchParams))).toEqual(selected)
  })

  it('clears prior workspace state on primary tab changes, including shared view names', () => {
    for (const view of ['dns', 'web']) {
      const state = parseSiteDetailRouteState({ domain: 'a.example', tab: 'observation', view })
      expect(buildSiteDetailQuery(selectSiteDetailTab(state, 'security'))).toEqual({ domain: 'a.example', tab: 'security' })
      expect(selectSiteDetailTab(state, 'observation')).toEqual(state)
    }
    const insights = parseSiteDetailRouteState({ domain: 'a.example', tab: 'insights', metric: 'tls13', range: 'all' })
    expect(buildSiteDetailQuery(selectSiteDetailTab(insights, 'observation'))).toEqual({ domain: 'a.example', tab: 'observation' })
    expect(buildSiteDetailQuery(selectSiteDetailTab(insights, 'overview'))).toEqual({ domain: 'a.example' })
  })

  it('does not silently erase or double-decode business targets; backend owns validation', () => {
    for (const domain of ['not-owned.example', '%61.example', '{"domain":[]}', 'a.example/?x=1&y=2']) {
      const path = siteTargetPath(41, domain)
      const parsed = parseSiteDetailRouteState(Object.fromEntries(new URL(path, 'https://go-furry.com').searchParams))
      expect(parsed.domain).toBe(domain)
    }
    expect(siteTargetPath(41, '  ')).toBe('/site/41')
    expect(siteEntityPath(41)).toBe('/site/41')
  })
})
