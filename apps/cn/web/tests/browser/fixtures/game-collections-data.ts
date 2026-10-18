import type { GameCollectionDetail, GameCollectionHome, GameCollectionIndex, GameCollectionSummary, GameCollectionTimelineItem } from '../../../app/types/game'

export const collectionMetadata = { schema_version: 1 as const, generated_at: '2026-10-06T00:00:00Z', as_of_date: '2026-10-06' }
export function collectionSummary(index: number, media: string, lang = 'zh'): GameCollectionSummary {
  return { code: `collection-${index}`, name: lang === 'en' ? `Forest stories ${index}` : `森林故事 ${index}`,
    info: lang === 'en' ? 'Follow our companions through changing worlds and the games that shaped their journey.' : '跟随伙伴穿越不断变化的世界，沿着作品的发行脉络，重访那些值得记住的故事。',
    visible_game_count: index % 4, published_at: '2026-10-01T00:00:00Z',
    preview_games: Array.from({ length: index % 4 }, (_, i) => ({ game_id: `${index * 10 + i}`, name: `Preview ${index}-${i}`, header_url: `${media}/collection-${index}-${i}.svg` })),
  }
}
export function collectionHome(media: string, lang = 'zh'): GameCollectionHome {
  // Out-of-order slots protect numeric placement, without inserting empty tiles.
  return { ...collectionMetadata, slots: [3, 1, 5, 2, 4].map(slot => {
    const collection = collectionSummary(slot, media, lang)
    // Home only contains positive-visible collections in the selected mode.
    if (!collection.visible_game_count) {
      collection.visible_game_count = 1
      collection.preview_games = [{ game_id: String(slot * 10), name: `Preview ${slot}-0`, header_url: `${media}/collection-${slot}-0.svg` }]
    }
    return { slot, collection }
  }) }
}
export function collectionIndex(media: string, lang: string, mode: string, page: number): GameCollectionIndex {
  const items = (page === 1 ? [3, 2, 1, 4, 5, 6] : [6, 7, 8]).map(i => collectionSummary(i, media, lang))
  if (mode === 'nsfw') items[0] = { ...items[0]!, name: 'Adult collection', preview_games: [{ game_id: '999', name: 'Adult preview', header_url: `${media}/adult-preview.svg` }] }
  return { ...collectionMetadata, page, page_size: 24, total: 8, has_more: page === 1, items }
}
export function collectionDetail(media: string, lang = 'zh', mode = 'sfw', code = 'collection-3'): GameCollectionDetail {
  const collection = { ...collectionSummary(3, media, lang), code }
  const item = (id: number, phase: GameCollectionTimelineItem['phase'], precision?: 'day' | 'month' | 'quarter' | 'year', date = '2020-01-01', inferred = false): GameCollectionTimelineItem => ({
    game_id: String(id), name: `${lang === 'en' ? 'Journey' : '旅途'} ${id}`, summary: lang === 'en' ? 'A new chapter in a shared world.' : '在共同的世界里，开启新的旅程。', header_url: `${media}/game-${id}.svg`, phase,
    chronology: precision ? { source: phase === 'released' ? 'first_available' : 'release', precision, window_start: date,
      window_end: date, inferred } : null,
  })
  let items = [item(21, 'released', 'year', '2004-01-01', true), item(19, 'released', 'month', '2010-05-01'),
    item(7, 'released', 'day', '2020-01-02'), item(8, 'released', 'day', '2020-01-02'),
    item(30, 'released_unknown'), item(40, 'upcoming_overdue', 'month', '2026-08-01'),
    item(41, 'upcoming', 'quarter', '2027-04-01'), item(42, 'upcoming', 'year', '2028-01-01'),
    item(43, 'upcoming', 'day', '2029-05-18'), item(50, 'upcoming_tba'), item(60, 'unknown')]
  items = items.map((value, index) => ({ ...value,
    primary_tag: index % 4 < 2 ? { code: 'story', name: lang === 'en' ? 'Story' : '剧情' } : null,
    secondary_tag: index % 4 === 0 || index % 4 === 2 ? { code: 'visual-novel', name: lang === 'en' ? 'Visual novel' : '视觉小说' } : null,
    rating: index % 4 === 0 ? { average: 4.64, count: 28 } : null,
    online: index % 4 === 0 ? { count: 1234, collected_at: '2026-10-06T08:00:00Z' } : index % 4 === 2 ? { count: 0, collected_at: '2026-10-06T08:00:00Z' } : null,
    community_count: index % 4 === 0 ? 3 : index % 4 === 2 ? 1 : 0,
  }))
  if (code === 'flow-long') {
    items = Array.from({ length: 7 }, (_, i) => ({ ...items[i % 4]!, game_id: String(200 + i), phase: 'released' }))
    items[1] = { ...items[1]!, name: (lang === 'en' ? 'A very long journey through a changing world ' : '穿越不断变化的世界与伙伴一起踏上漫长旅程').repeat(4), summary: items[1]!.summary.repeat(8) }
  }
  if (code === 'unknown-only') items = [item(60, 'unknown')]
  if (code === 'adult-only') items = []
  if (mode === 'nsfw') items.push({ ...item(999, 'unknown'), name: 'Adult game', header_url: `${media}/adult-game.svg` })
  return { ...collectionMetadata, collection: { ...collection, visible_game_count: items.length }, items }
}
