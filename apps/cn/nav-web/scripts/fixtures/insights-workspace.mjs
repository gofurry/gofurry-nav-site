export const workspaceMetricKeys = ['latest_observed', 'peak_30d', 'average_30d']
export const workspaceRegionKeys = ['CN', 'US', 'HK']
const asOf = '2026-09-08'
const observed = '2026-09-09T04:00:00Z'
const game = (i, media) => ({ id: 101 + i, name: i === 1 ? 'Across the Quiet Mountains — A Very Long Game Name / 越过寂静山脉的漫长旅途' : 'Game ' + (i + 1), ...(i % 3 !== 2 ? { visual: { kind: 'game_header', asset: media + '/game-' + i + '.svg' } } : {}) })
const site = (i) => ({ id: 201 + i, name: i === 1 ? 'Long Site Identity / 很长的网站名称用于验证自然换行' : 'Site ' + (i + 1), ...(i % 2 === 0 ? { visual: { kind: 'site_icon', asset: 'nav/sites/' + (201 + i) + '/icon/' + 'a'.repeat(32) + '.svg' } } : {}) })
export function workspaceFixtureResponse(url, media, state = {}) {
  const path = url.pathname
  if (state.failure && path.endsWith(state.failure)) return { status: 503 }
  const empty = state.empty
  if (path.endsWith('/players/ranking')) {
    const metric = url.searchParams.get('metric') || 'latest_observed'
    const latest = metric === 'latest_observed'
    return { data: { metric, basis: latest ? 'scheduled_snapshot' : 'finalized_daily_facts', snapshot_scheduled_for: latest ? observed : null, latest_slot_scheduled_for: latest ? observed : null,
      observed_from: latest ? '2026-09-09T03:58:00Z' : null, observed_through: latest ? observed : null, window_from: latest ? null : '2026-08-10', window_through: latest ? null : asOf,
      population: empty ? 0 : 213, ranked: empty ? 0 : 189, entity_coverage: empty ? null : 189 / 213,
      items: empty ? [] : Array.from({ length: 20 }, (_, i) => ({ rank: i + 1, game: game(i, media), value: i === 19 ? 0 : 1524 / (i + 1), observed_at: latest ? observed : null, eligible_from: latest ? null : '2026-08-10', observed_days: latest ? null : 27, successful_samples: latest ? null : 180, sample_coverage: latest || i === 1 ? null : 0.75 })),
    } }
  }
  if (path.endsWith('/prices/overview')) return { data: { region: url.searchParams.get('region'), as_of: asOf, population: 213, priced: 140, free: 39, unpriced: 12, unknown: 10, unavailable: 12, known: 191, coverage: 191 / 213, discounted: 28, discounted_share: 0.2 } }
  if (path.endsWith('/prices/discounts')) {
    const region = url.searchParams.get('region') || 'CN', currency = { CN: 'CNY', US: 'USD', HK: 'HKD' }[region]
    return { data: { region, as_of: asOf, items: empty ? [] : Array.from({ length: 6 }, (_, i) => ({ game: game(i, media), currency, initial_amount: 999, final_amount: i === 0 ? 0 : 599, discount_percent: i === 0 ? 100 : 40, observed_low: i === 2 ? null : { amount: i === 0 ? 0 : 399, currency, first_seen: asOf, observed_since: '2026-08-01', initial_amount: 999, discount_percent: 60 } })) } }
  }
  if (path.endsWith('/languages/overview')) return { data: { as_of: asOf, freshness_seconds: 259200, population: 213, fresh: 189, stale: 12, unobserved: 12, coverage: 189 / 213, fully_normalized_games: 180, normalization_coverage: 180 / 189, unmapped_games: 9, unmapped_entries: 14,
    items: empty ? [] : ['en', 'zh-Hans', 'ja', 'de', 'fr', 'es', 'pt', 'ru', 'ko', 'it', 'pl', 'th', 'tr', 'nl'].map((code, i) => ({ code, steam_name: code, supported_games: 180 - i * 12, share: i === 11 ? null : (180 - i * 12) / 189, explicit_full_audio_games: i === 10 ? 0 : 23 - i, explicit_full_audio_share: i === 11 ? null : i === 10 ? 0 : (23 - i) / 189 })) } }
  if (path.endsWith('/certificates/overview')) {
    const item = i => ({ site: site(i), target: i === 1 ? 'very-long-infrastructure-target-that-must-wrap-without-changing-the-page-width.example.test:443' : 'site-' + i + '.example.test:443', not_after: '2026-09-' + String(8 + i).padStart(2, '0') + 'T00:00:00Z', days_to_expiry: i - 1, expiry_status: i === 0 ? 'expired' : 'expires_within_7d', verified: false, verification_issue: i === 0 ? 'hostname_mismatch' : 'unknown_authority', issuer: 'Example issuer', observed_at: observed })
    return { data: { as_of: asOf, reference_at: observed, freshness_seconds: 172800, population: 238, eligible: 238, verification: { known: 229, verified: 221, failed: 8, coverage: 229 / 238 }, expiry: { known: 229, coverage: 229 / 238, expired: 2, expires_within_7d: 6, expires_in_8_30d: 18, later: 203 }, quality: { not_applicable: 0, stale: 3, not_probed: 2, probe_failed: 3, unknown: 1 }, expiry_attention: empty ? [] : [0, 1, 2].map(item), verification_issues: empty ? [] : [0, 1].map(item) } }
  }
  return { status: 503 }
}
