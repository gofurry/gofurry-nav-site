// Shared by browser interception and the local SSR fixture upstream. No CDN is required.
export function mockOverview(domain, mediaBase = 'https://media.example') {
  return {
    generated_at: domain === 'site' ? '2026-09-01T10:00:00Z' : '2026-09-01T12:00:00Z',
    entity_count: domain === 'site' ? 238 : 213,
    changes_7d: domain === 'site' ? 20 : 27,
    metrics: domain === 'site' ? [
      { key: 'tls13', value: 0.824, delta_30d: 0.042 },
      { key: 'ipv6', value: 0.63, delta_30d: -0.011 },
      { key: 'security_txt', value: null, delta_30d: null },
    ] : [],
    recent_changes: domain === 'site' ? [
      { type: 'site.ipv6.enabled', date: '2026-09-01', occurred_at: null, entity: { id: 41, name: 'Site fixture', visual: { kind: 'site_icon', asset: 'nav/sites/41/icon/' + 'a'.repeat(32) + '.svg' } }, detail: null },
      { type: 'site.tls13.disabled', date: '2026-08-31', occurred_at: null, entity: { id: 42, name: 'Site failure fixture' }, detail: null },
    ] : [
      { type: 'game.windows.added', date: '2026-09-01', occurred_at: '2026-09-01T12:00:00Z', entity: { id: 82, name: 'Game fixture — a long title across the ecosystem', visual: { kind: 'game_header', asset: `${mediaBase}/game.svg` } }, detail: null },
      { type: 'game.linux.added', date: '2026-08-31', occurred_at: '2026-08-31T12:00:00Z', entity: { id: 83, name: 'Game failure fixture' }, detail: null },
    ],
  }
}

export function mockGamePanel(mediaBase = 'https://media.example') {
  const game = (id, name) => ({
    id: String(id), appid: String(id), name, header_url: `${mediaBase}/game-${id}.svg`, capsule_url: '',
    release_date: '2026-08-28', online_count: { count: 1240, peak_count: 2800, status: 'success', collected_at: '2026-09-01T09:00:00Z' },
    price: { region: 'CN', currency: 'CNY', final_amount: 3800, available: true, is_free: false, discount_percent: 20 },
    prices: [{ region: 'US', currency: 'USD', final_amount: 599, initial_amount: 1198, available: true, is_free: false, discount_percent: 50 }],
  })
  const active = game(91, 'Active game fixture')
  const discount = game(92, 'Discount game fixture')
  const latest = game(93, 'Recent release fixture')
  return {
    top_online: [active], highest_discount: [active, discount], latest_games: [active, discount, latest],
    updated_games: [], popular_games: [], free_games: [], top_price: [], low_price: [], latest_news: [],
  }
}

export function mockGameHome(mediaBase = 'https://media.example') {
  return { panel: mockGamePanel(mediaBase), latest_news: { news_zh: [], news_en: [] }, latest_reviews: [] }
}
