import type { GameShowcaseItem, GameShowcaseSnapshot } from '../../../app/types/game'

export const showcaseOrigins = { primary: 'https://showcase-primary.example', mirror: 'https://showcase-mirror.example' }
export const showcaseSteam = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/6101/showcase/header.jpg?v=2'
export type ShowcaseScenario = 'empty' | 'single-editorial' | 'automatic-upcoming' | 'managed-sponsored-tabletop'
  | 'managed-sponsored-merchandise' | 'four-items' | 'long-content' | 'media-failure' | 'steam-failure'
  | 'varied-content' | 'primary-failure' | 'showcase-failure' | 'showcase-transient' | 'slow-showcase' | 'tracking-failure' | 'mobile-missing' | 'two-managed'

export function makeShowcase(scenario: ShowcaseScenario, lang: 'zh' | 'en'): GameShowcaseSnapshot {
  const snapshot: GameShowcaseSnapshot = { schema_version: 1, snapshot_id: 'a'.repeat(32), generated_at: '2026-10-04T04:00:00Z', valid_until: '2026-10-04T04:05:00Z', items: [] }
  if (scenario === 'empty' || scenario === 'showcase-failure') return snapshot
  const en = lang === 'en'
  const editorial: GameShowcaseItem = {
    key: 'campaign:9007199254740993', campaign_id: '9007199254740993', game_id: '6101', source: 'managed', reason: 'editorial', content_type: 'game', sponsored: false,
    title: en ? 'Echoes of the Wild' : '荒野回声', summary: en ? 'A quiet journey through the forest, with old friends and new stories.' : '与伙伴踏上森林旅途，在星光与古老遗迹间发现新的故事。',
    tags: en ? ['Adventure', 'Furry', 'Story'] : ['冒险', '兽人', '剧情'], editorial_note: en ? 'A thoughtful journey worth sharing.' : '一段值得与伙伴分享的温柔旅程。',
    artwork: { kind: 'managed', desktop_object_key: `game/showcase/9007199254740993/desktop/${'a'.repeat(32)}.avif`, mobile_object_key: `game/showcase/9007199254740993/mobile/${'b'.repeat(32)}.avif`, focal_x: .25, focal_y: .75 },
    primary_action: { type: 'game', game_id: '6101' }, secondary_action: { type: 'steam', target: 'https://store.steampowered.com/app/6101' },
    position: 1, tracking_token: 'opaque-signed-editorial',
  }
  const upcoming: GameShowcaseItem = { ...editorial, key: 'automatic:6102', campaign_id: undefined, game_id: '6102', source: 'automatic', reason: 'upcoming',
    title: en ? 'Tomorrow in the Valley' : '山谷的明天', editorial_note: undefined, artwork: { kind: 'steam', url: showcaseSteam }, primary_action: { type: 'game', game_id: '6102' }, tracking_token: 'opaque-signed-upcoming',
    release: { availability: 'upcoming', precision: 'quarter', exact_date: null, year: 2027, month: null, quarter: 2, window_start: '2027-04-01', window_end: '2027-06-30' } }
  const sponsored: GameShowcaseItem = { ...editorial, key: 'campaign:9007199254740994', campaign_id: '9007199254740994', game_id: undefined, reason: 'sponsored', content_type: 'tabletop', sponsored: true,
    title: en ? 'Woodland Companions' : '林间伙伴', summary: en ? 'Gather around the table and write your own woodland adventure.' : '与朋友围坐桌前，共同书写属于你们的森林冒险。', editorial_note: undefined,
    artwork: { ...editorial.artwork, desktop_object_key: `game/showcase/9007199254740994/desktop/${'c'.repeat(32)}.avif`, mobile_object_key: `game/showcase/9007199254740994/mobile/${'d'.repeat(32)}.avif` },
    primary_action: { type: 'project', target: 'https://showcase-destination.example/project' }, secondary_action: { type: 'kickstarter', target: 'https://showcase-destination.example/kickstarter' }, tracking_token: 'opaque-signed-sponsored' }
  const merchandise: GameShowcaseItem = { ...sponsored, content_type: 'merchandise', title: en ? 'A Little Forest Friend' : '口袋里的森林伙伴', primary_action: { type: 'product', target: 'https://showcase-destination.example/product' } }
  const trending: GameShowcaseItem = { ...upcoming, key: 'automatic:6103', game_id: '6103', primary_action: { type: 'game', game_id: '6103' }, reason: 'trending', release: undefined, title: en ? 'Starport Letters' : '星港来信', tracking_token: 'opaque-signed-trending' }
  if (scenario === 'automatic-upcoming' || scenario === 'steam-failure') snapshot.items = [upcoming]
  else if (scenario === 'managed-sponsored-tabletop') snapshot.items = [sponsored]
  else if (scenario === 'managed-sponsored-merchandise') snapshot.items = [merchandise]
  else if (scenario === 'two-managed') snapshot.items = [editorial, sponsored]
  else if (scenario === 'four-items' || scenario === 'media-failure' || scenario === 'tracking-failure') snapshot.items = [editorial, upcoming, sponsored, trending]
  else snapshot.items = [editorial]
  if (scenario === 'varied-content') {
    snapshot.items = [
      { ...trending, title: en ? 'Quiet' : '静谧', summary: en ? 'A journey.' : '一段旅程。', tags: [], secondary_action: undefined },
      { ...editorial, title: (en ? 'Across the forest and beyond the stars ' : '穿越森林与星海的漫长冒险').repeat(4), summary: editorial.summary.repeat(12), release: upcoming.release, editorial_note: editorial.editorial_note!.repeat(6) },
      sponsored,
    ]
  }
  if (scenario === 'long-content') {
    editorial.title = (en ? 'Across the forest and beyond the stars ' : '穿越森林与星海的漫长冒险').repeat(8)
    editorial.summary = (en ? 'There are many more stories to discover. ' : '在无数个温柔的日夜里寻找彼此与更多值得珍藏的回忆。').repeat(12)
    editorial.editorial_note = editorial.summary
    editorial.tags = ['A'.repeat(90), 'B'.repeat(90), 'C'.repeat(90)]
  }
  if (scenario === 'mobile-missing') delete editorial.artwork.mobile_object_key
  snapshot.items.forEach((item, index) => { item.position = index + 1 })
  return snapshot
}
