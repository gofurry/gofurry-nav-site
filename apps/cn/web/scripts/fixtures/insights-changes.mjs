export const changeCategories = {
  site: { capability: 'site.ipv6.enabled', target: 'site.primary_target.changed', certificate: 'site.tls_certificate.changed' },
  game: { pricing_model: 'game.free.enabled', platform: 'game.linux.added', release: 'game.release.changed', price: 'game.price.decreased', discount: 'game.discount.started' },
}

export async function changesFixtureResponse(url, mediaBase, state) {
  if (!/^\/api\/v2\/(nav|game)\/insights\/changes$/.test(url.pathname)) return { status: 503 }
  const domain = url.pathname.includes('/nav/') ? 'site' : 'game'
  const cursor = url.searchParams.get('cursor')
  const category = url.searchParams.get('category') || Object.keys(changeCategories[domain])[0]
  if (state.delayRange === url.searchParams.get('range')) await new Promise(resolve => setTimeout(resolve, 450))
  if (state.failure || (cursor && state.moreFailure)) return { status: 503 }
  if (state.empty) return { data: { items: [], next_cursor: null } }
  const dates = cursor ? ['2026-09-08', '2026-09-07'] : ['2026-09-09', '2026-09-09', '2026-09-08', '2026-09-08']
  const items = dates.map((date, index) => ({
    domain, category, type: changeCategories[domain][category], date,
    occurred_at: index === 0 && !cursor ? date + 'T13:32:00Z' : null,
    entity: {
      id: cursor ? 34 + index : [31, 31, 32, 33][index],
      name: index === 2 ? '' : index === 3 ? `${domain} · 公开生态记录 LongEntityNameWithoutSpacesForResponsiveVerification` : `${domain === 'site' ? 'Furry Archive' : 'Moonlit Journey'} ${cursor ? index + 5 : index + 1} · ${url.searchParams.get('range')}`,
      ...(index === 0 || index === 3 ? { visual: { kind: domain === 'site' ? 'site_icon' : 'game_header', asset: domain === 'site' ? 'nav/sites/' + (cursor ? 34 + index : [31, 31, 32, 33][index]) + '/icon/' + 'a'.repeat(32) + '.svg' : mediaBase + '/change-header.svg' } } : {}),
    },
    detail: null,
  }))
  return { data: { items, next_cursor: cursor ? null : 'fixture-next-page' } }
}
