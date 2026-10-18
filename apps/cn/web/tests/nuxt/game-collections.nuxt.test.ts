import { afterEach, expect, it, vi } from 'vitest'
import { useNuxtApp } from '#app'
import { defineComponent, h, nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { getGameCollectionHome, getGameCollections, getGameCollectionDetail } from '../../app/services/game'
import { useGameCollectionModeRefresh } from '../../app/composables/useGameCollectionModeRefresh'
import TimelineItem from '../../app/components/game/collections/GameCollectionTimelineItem.vue'
import GameSidebarLinks from '../../app/components/game/detail/GameSidebarLinks.vue'
import type { GameBaseInfoResponse } from '../../app/types/game'
import Timeline from '../../app/components/game/collections/GameCollectionTimeline.vue'
import { collectionDetail } from '../browser/fixtures/game-collections-data'
import { connector, compactPlacement } from '../../app/utils/serpentineSequence'

afterEach(() => { localStorage.removeItem('mode'); vi.unstubAllGlobals(); vi.restoreAllMocks() })

it('owns independent timeouts, strict locale and SFW Home query', async () => {
  const fetcher = vi.fn().mockResolvedValue({ code: 1, data: {} })
  vi.stubGlobal('$fetch', fetcher)
  await useNuxtApp().runWithContext(() => getGameCollectionHome('bad'))
  expect(fetcher).toHaveBeenLastCalledWith('/game/collections/home', expect.objectContaining({ query: { lang: 'zh', mode: 'sfw' }, retry: 0, timeout: 1000 }))
  await useNuxtApp().runWithContext(() => getGameCollections('en', 'nsfw', 2))
  expect(fetcher).toHaveBeenLastCalledWith('/game/collections', expect.objectContaining({ query: { lang: 'en', mode: 'nsfw', page: 2, page_size: 24, q: '', phase: 'all', sort: 'published_desc' }, retry: 0, timeout: 8000 }))
  await useNuxtApp().runWithContext(() => getGameCollectionDetail('a/b', 'zh', 'sfw'))
  expect(fetcher).toHaveBeenLastCalledWith('/game/collections/a%2Fb', expect.objectContaining({ query: { lang: 'zh', mode: 'sfw' }, retry: 0, timeout: 8000 }))
})

function gate<T>() { let resolve!: (value: T) => void; const promise = new Promise<T>(r => { resolve = r }); return { promise, resolve } }
const change = (mode: string) => window.dispatchEvent(new CustomEvent('mode-change', { detail: { mode } }))
async function mountMode(load: (mode: 'sfw' | 'nsfw') => Promise<string>, onError = vi.fn()) {
  let state!: ReturnType<typeof useGameCollectionModeRefresh<string>>
  const wrapper = await mountSuspended(defineComponent({ setup() {
    state = useGameCollectionModeRefresh({ initial: 'SFW', load, onError })
    return () => h('p', state.snapshot.value ?? '')
  } }))
  return { wrapper, state }
}
it('SFW hydration makes no read; NSFW makes exactly one', async () => {
  const load = vi.fn().mockResolvedValue('NSFW')
  const sfw = await mountMode(load)
  expect(load).not.toHaveBeenCalled(); sfw.wrapper.unmount()
  localStorage.setItem('mode', 'nsfw')
  const nsfw = await mountMode(load)
  expect(load).toHaveBeenCalledExactlyOnceWith('nsfw')
  await nextTick(); expect(nsfw.state.snapshot.value).toBe('NSFW'); nsfw.wrapper.unmount()
})
it('keeps ready content, ignores stale mode responses and cleans up listeners/inflight completions', async () => {
  const older = gate<string>(), newer = gate<string>(), after = gate<string>()
  const load = vi.fn().mockReturnValueOnce(older.promise).mockReturnValueOnce(newer.promise).mockReturnValueOnce(after.promise)
  const { wrapper, state } = await mountMode(load)
  change('nsfw'); expect(state.snapshot.value).toBe('SFW')
  change('sfw'); newer.resolve('NEW SFW'); await newer.promise; await nextTick()
  older.resolve('STALE NSFW'); await older.promise; await nextTick()
  expect(state.snapshot.value).toBe('NEW SFW')
  change('nsfw'); wrapper.unmount(); after.resolve('DISPOSED'); await after.promise; await nextTick()
  change('sfw'); expect(load).toHaveBeenCalledTimes(3); expect(state.snapshot.value).toBe('NEW SFW')
})
it('retains the old snapshot on failure and retries the current mode', async () => {
  const failure = new Error('503'), load = vi.fn().mockRejectedValueOnce(failure).mockResolvedValueOnce('NSFW')
  const onError = vi.fn(), view = await mountMode(load, onError)
  change('nsfw'); await nextTick(); await nextTick()
  expect(view.state.snapshot.value).toBe('SFW'); expect(onError).toHaveBeenCalledWith(failure)
  await view.state.refresh(); expect(view.state.snapshot.value).toBe('NSFW')
  expect(load.mock.calls).toEqual([['nsfw'], ['nsfw']]); view.wrapper.unmount()
})
it.each([0, 1, 2, 3, 4, 6, 7])('uses shared geometry with %i ordered nodes, preserving Backend order', async count => {
  const detail = collectionDetail('local')
  detail.items = Array.from({ length: count }, (_, i) => ({ ...detail.items[0]!, game_id: String(100 - i) }))
  const wrapper = await mountSuspended(Timeline, { props: { detail }, global: { stubs: { GameCollectionTimelineItem: true } } })
  const nodes = wrapper.findAll('li')
  expect(nodes.map(n => n.attributes('data-game-id'))).toEqual(detail.items.map(i => i.game_id))
  nodes.forEach((node, i) => {
    expect(node.attributes('data-connector')).toBe(connector(i, count))
    expect((node.element as HTMLElement).style.getPropertyValue('--timeline-column')).toBe(String(compactPlacement(i)['--timeline-column']))
  })
  wrapper.unmount()
})

it('renders detail metadata accessibly and tolerates pre-upgrade cached items', async () => {
  const item = collectionDetail('local').items[0]!
  const wrapper = await mountSuspended(TimelineItem, { props: { item } })
  expect(wrapper.findAll('.game-collection-timeline-tag').map(tag => tag.text())).toEqual(['剧情', '视觉小说'])
  expect(wrapper.find('.game-collection-timeline-metrics').text()).toContain('1,234')
  expect(wrapper.findAll('svg').every(icon => icon.attributes('aria-hidden') === 'true')).toBe(true)
  expect(wrapper.find('[aria-label="评分 4.6，共 28 条评价"]').exists()).toBe(true)
  const observed = wrapper.find('[aria-label^="最近观测在线人数"]')
  expect(observed.attributes('aria-label')).toContain('最近观测在线人数 1,234，采集于')
  for (const part of ['2026/10/06', '08:00', 'UTC']) expect(observed.attributes('aria-label')).toContain(part)
  expect(observed.attributes('title')).toBe(observed.attributes('aria-label'))
  await wrapper.setProps({ item: { ...item, primary_tag:null, secondary_tag:null, rating:null, online:{ count:0,collected_at:'2026-10-06T00:00:00Z' }, community_count:0 } })
  expect(wrapper.findAll('.game-collection-timeline-tag')).toHaveLength(0)
  expect(wrapper.find('.game-collection-timeline-metrics').text()).toBe('0')
  expect(wrapper.find('[aria-label^="最近观测在线人数 0，采集于"]').exists()).toBe(true)
  await wrapper.setProps({ item: { ...item, online: { count: 1234, collected_at: 'invalid' } } })
  expect(wrapper.find('[aria-label="最近观测在线人数 1,234"]').exists()).toBe(true)
  const legacy = { ...item }; delete legacy.primary_tag; delete legacy.secondary_tag; delete legacy.rating; delete legacy.online; delete legacy.community_count
  await wrapper.setProps({ item:legacy })
  expect(wrapper.find('.game-collection-timeline-metrics').text()).toBe('')
  expect(wrapper.find('.game-collection-title').text()).toBe(item.name)
  wrapper.unmount()
})

it('Game Detail community entries use the same trimmed key/value rule as Timeline counts', async () => {
  const groups = [{ key: 'discord', value: 'https://valid.example' }, { key: ' telegram ', value: ' https://second.example ' }, {}, { key: 'discord', value: '' }, { key: '', value: 'https://x' }, { key: '   ', value: 'https://x' }, { key: 'discord', value: '   ' }]
  const wrapper = await mountSuspended(GameSidebarLinks, {
    props: { game: { groups, links: [], resources: [] } as unknown as GameBaseInfoResponse },
    global: { stubs: { GameCommentDialog: true } },
  })
  expect(wrapper.findAll('.site-icon-link').map(link => link.attributes('href'))).toEqual(['https://valid.example', ' https://second.example '])
  await wrapper.setProps({ game: { groups: null, links: [], resources: [] } as unknown as GameBaseInfoResponse })
  expect(wrapper.findAll('.site-icon-link')).toHaveLength(0)
  wrapper.unmount()
})
