import { setTimeout as delay } from 'node:timers/promises'
export const compareSiteKeys = ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified']
export const compareGroups = { site: ['capabilities', 'certificate'], game: ['basic', 'platform', 'activity', 'price', 'language'] }
const order = [2, 1, 4, 3, 5, 6]
const names = { 1: 'Cedar Archive', 2: 'Fox Atlas — a long shared identity / 一个需要自然换行的实体名称', 3: 'Fox', 4: 'Fox Den', 5: 'Some Fox', 6: 'Canyon' }
const entity = (domain, id, media) => ({ id, name: names[id] || 'Entity ' + id, ...(id !== 1 ? { visual: { kind: domain === 'site' ? 'site_icon' : 'game_header', asset: domain === 'site' ? 'nav/sites/' + id + '/icon/' + 'a'.repeat(32) + '.svg' : media + '/game-' + id + '.svg' } } : {}) })

export async function compareFixtureResponse(url, media, body, state = {}) {
  const path = url.pathname
  if (path.endsWith('/nav/sites/directory')) {
    if (state.directoryFailure) return { status: 503 }
    return { data: order.map(id => ({ id: String(id), name: names[id], domain: id === 5 ? 'fox.example' : 'site-' + id + '.example', icon: id === 1 ? null : 'nav/sites/' + id + '/icon/' + 'a'.repeat(32) + '.svg', info: '', country: null, nsfw: 'false', welfare: 'false', view_count: 0, create_time: '', update_time: '' })) }
  }
  if (path.endsWith('/game/search/simple')) {
    state.searches?.push(body?.txt)
    if (body?.txt === 'slow' && !state.controlledTiming) await delay(650)
    if (state.searchFailure) return { status: 503 }
    return { data: order.map(id => ({ id: String(id), name: body?.txt === 'slow' ? 'Stale result ' + id : names[id], info: 'A short public search description / 公开简介', cover: id === 1 ? '' : media + '/game-' + id + '.svg' })) }
  }
  if (!path.endsWith('/insights/compare')) return { status: 503 }
  if (state.compareFailure) return { status: 503 }
  if (state.compareDelay) await delay(state.compareDelay)
  const ids = url.searchParams.get('ids').split(',').map(Number)
  if (state.reverse) ids.reverse()
  const status = state.insufficient ? 'insufficient_data' : 'ready'
  if (path.includes('/nav/')) return { data: { status, as_of: state.insufficient ? null : '2026-09-08', sites: ids.map(id => ({ site: entity('site', id, media),
    capabilities: compareSiteKeys.map((key, i) => ({ key, state: i === 1 ? 'unknown' : i === 2 ? 'unsupported' : 'supported' })),
    certificate: id === 1 ? null : { target: 'long-primary-infrastructure-target-that-must-wrap.example.test:443', not_after: '2026-09-10T00:00:00Z', days_to_expiry: 1, expiry_status: 'expires_within_7d', verified: false, verification_issue: 'hostname_mismatch', issuer: 'Example CA', observed_at: '2026-09-08T23:00:00Z' },
  })) } }
  const region = url.searchParams.get('region') || 'CN', currency = { CN: 'CNY', US: 'USD', HK: 'HKD' }[region]
  return { data: { status, region, state_as_of: state.insufficient ? null : '2026-09-08', player_snapshot_scheduled_for: '2026-09-09T04:00:00Z', player_fact_through: '2026-09-08', games: ids.map(id => ({ game: entity('game', id, media),
    state: { free: id === 1, windows: true, mac: false, linux: null, release: 'available' },
    players: { current_available: id !== 1, current: id === 1 ? null : id === 2 ? 0 : 1524, observed_at: '2026-09-09T04:01:00Z', peak_30d: id === 2 ? 0 : 1692, average_30d: 624.2, eligible_from_30d: '2026-08-10', observed_days_30d: 27, successful_samples_30d: 180, sample_coverage_30d: id === 1 ? null : .75 },
    price: { region, available: true, state: id === 1 ? 'free' : 'priced', currency: id === 1 ? null : currency, initial_amount: id === 1 ? null : 999, final_amount: id === 1 ? null : id === 2 ? 0 : 599, discount_percent: id === 1 ? null : id === 2 ? 100 : 40, observed_low: id === 1 ? null : { amount: id === 2 ? 0 : 399, currency, first_seen: '2026-09-01', observed_since: '2026-08-01', initial_amount: 999, discount_percent: 60 } },
    languages: { evidence: id === 1 ? 'stale' : 'fresh', supported: ['en','zh-CN'], explicit_full_audio: id === 1 ? [] : ['en'], unknown_names: id === 1 ? ['Unmapped fixture'] : [] },
  })) } }
}
