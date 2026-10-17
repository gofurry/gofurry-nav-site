import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import NavSpotlightPanels from '../../app/components/nav/NavSpotlightPanels.vue'
import GameSearchResult from '../../app/components/game/search/GameSearchResult.vue'
import { useReducedMotion } from '../../app/composables/useReducedMotion'
import type { SearchPageResponseItem } from '../../app/types/game'
import type { Site } from '../../app/types/nav'

let media: MediaQueryList
beforeEach(() => {
  media = Object.assign(new EventTarget(), { matches: false }) as MediaQueryList
  const original = window.matchMedia.bind(window)
  vi.spyOn(window, 'matchMedia').mockImplementation(query => query === '(prefers-reduced-motion: reduce)' ? media : original(query))
})
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks() })
function reduce(value: boolean) {
  Object.assign(media, { matches: value })
  media.dispatchEvent(new Event('change'))
}
const sites: Site[] = Array.from({ length: 14 }, (_, i) => ({
  id: String(i + 1), name: `Site ${i + 1}`, domain: '', info: '', country: null,
  nsfw: '0', welfare: '0', icon: null, view_count: 0, create_time: '', update_time: '',
}))
const spotlight = { page_size: 6, featured: sites, popular: [], latest: [], random: [] }
const games = (start: number): SearchPageResponseItem[] => [start, start + 1].map(id => ({
  id: String(id), name: `Game ${id}`, info: '', cover: '', update_time: '', release_date: '',
  remark_count: 0, avg_score: 0, appid: id, primary_tag: '', secondary_tag: '', release: null, first_available: null,
}))
const searchProps = { gameList: games(1), currentPage: 1, pageSize: 2, pageDirection: 1 as const,
  totalPages: 3, total: 6, loading: false, status: 'success' as const }

it('keeps SSR deterministic and releases each mounted preference subscription', async () => {
  const harness = defineComponent({ setup() {
    const reduced = useReducedMotion()
    return () => h('output', String(reduced.value))
  } })
  expect(await renderToString(h(harness))).toBe('<output>false</output>')
  expect(window.matchMedia).not.toHaveBeenCalled()
  reduce(true)
  const add = vi.spyOn(media, 'addEventListener'), remove = vi.spyOn(media, 'removeEventListener')
  const view = await mountSuspended(harness)
  try {
    expect(view.text()).toBe('true')
    reduce(false); await nextTick(); expect(view.text()).toBe('false')
    reduce(true); await nextTick(); expect(view.text()).toBe('true')
  } finally { view.unmount() }
  expect(remove.mock.calls).toEqual(add.mock.calls)
})

it('settles Spotlight immediately on initial reduce and interruption without advancing the clock', async () => {
  reduce(true)
  const view = await mountSuspended(NavSpotlightPanels, { props: { spotlight, displayMode: 'sfw' } })
  vi.useFakeTimers()
  const next = () => view.findAll('.spotlight-panel__pager button')[1]!.trigger('click')
  const titles = () => view.find('.spotlight-panel__track').findAll('strong').map(node => node.text())
  try {
    await next()
    expect(titles()).toEqual(sites.slice(6, 12).map(site => site.name))
    expect(view.find('.spotlight-panel__track').findAll('.spotlight-panel__slide')).toHaveLength(1)
    reduce(false); await nextTick(); await next() // page 3, still moving
    await next() // latest selection wraps to page 1
    reduce(true); await nextTick()
    expect(titles()).toEqual(sites.slice(0, 6).map(site => site.name))
    await vi.advanceTimersByTimeAsync(1000)
    expect(titles()).toEqual(sites.slice(0, 6).map(site => site.name))
    reduce(false); await nextTick(); await next()
    expect(view.find('.spotlight-panel__track').classes()).toContain('spotlight-panel__track--next')
    await vi.advanceTimersByTimeAsync(559)
    expect(view.find('.spotlight-panel__track').findAll('.spotlight-panel__slide')).toHaveLength(2)
    await vi.advanceTimersByTimeAsync(1)
    expect(titles()).toEqual(sites.slice(6, 12).map(site => site.name))
  } finally { view.unmount() }
})

it.each(['reduce', 'unmount'] as const)('invalidates Spotlight nextTick before it can queue a RAF on %s', async action => {
  const view = await mountSuspended(NavSpotlightPanels, { props: { spotlight, displayMode: 'sfw' } })
  vi.useFakeTimers()
  try {
    await view.findAll('.spotlight-panel__pager button')[1]!.trigger('click')
    const raf = vi.spyOn(globalThis, 'requestAnimationFrame')
    vi.advanceTimersByTime(560) // callback queued nextTick, but no RAF exists yet
    if (action === 'reduce') reduce(true)
    else view.unmount()
    await nextTick()
    expect(raf).not.toHaveBeenCalled()
    if (action === 'reduce') expect(view.find('.spotlight-panel__track').findAll('.spotlight-panel__slide')).toHaveLength(1)
  } finally { if (action !== 'unmount') view.unmount() }
})

it('Search keeps real pending data static across restoration and resumes motion on the next page', async () => {
  const view = await mountSuspended(GameSearchResult, { props: searchProps })
  vi.useFakeTimers()
  try {
    await view.setProps({ currentPage: 2, loading: true, status: 'pending' })
    expect(view.find('.search-results').classes()).toContain('search-results--sliding')
    reduce(true); await nextTick()
    expect(view.findAll('.search-result-page-slide')).toHaveLength(1)
    expect(view.findAll('.search-page-title')).toHaveLength(0)
    expect(view.findAll('.search-page-card--skeleton')).toHaveLength(2)
    reduce(false); await nextTick()
    await vi.advanceTimersByTimeAsync(1000)
    expect(view.findAll('.search-page-title')).toHaveLength(0)
    expect(view.attributes('aria-busy')).toBe('true')
    await view.setProps({ currentPage: 3 }) // newer pending page, still no accepted result
    await view.setProps({ gameList: games(5), loading: false, status: 'success' })
    expect(view.findAll('.search-page-title').map(node => node.text())).toEqual(['Game 5', 'Game 6'])
    expect(view.findAll('.search-result-page-slide')).toHaveLength(1)
    await view.setProps({ currentPage: 2, loading: true, status: 'pending', pageDirection: -1 })
    expect(view.find('.search-result-page-track').classes()).toContain('search-result-page-track--prev')
    await view.setProps({ gameList: games(3), loading: false, status: 'success' })
    reduce(true); await nextTick() // data ready, cancel the visual completion timer
    expect(view.findAll('.search-page-title').map(node => node.text())).toEqual(['Game 3', 'Game 4'])
    expect(view.findAll('.search-result-page-slide')).toHaveLength(1)
    await vi.advanceTimersByTimeAsync(1000)
    expect(view.findAll('.search-page-title').map(node => node.text())).toEqual(['Game 3', 'Game 4'])
  } finally { view.unmount() }
})

it('Reduced Search does not relabel old cards while the route is still starting the next request', async () => {
  reduce(true)
  const view = await mountSuspended(GameSearchResult, { props: searchProps })
  vi.useFakeTimers()
  try {
    await view.setProps({ currentPage: 2 })
    expect(view.findAll('.search-page-title')).toHaveLength(0)
    expect(view.findAll('.search-page-card--skeleton')).toHaveLength(2)
    await view.setProps({ loading: true, status: 'pending' })
    reduce(false); await nextTick()
    await view.setProps({ gameList: games(3), loading: false, status: 'success' })
    expect(view.findAll('.search-page-title').map(node => node.text())).toEqual(['Game 3', 'Game 4'])
    expect(view.findAll('.search-result-page-slide')).toHaveLength(1)
    reduce(true); await nextTick()
    // A batched ready page + list is already authoritative; no phantom pending.
    await view.setProps({ currentPage: 3, gameList: games(5) })
    expect(view.findAll('.search-page-title').map(node => node.text())).toEqual(['Game 5', 'Game 6'])
    expect(view.findAll('.search-page-card--skeleton')).toHaveLength(0)
  } finally { view.unmount() }
})

it.each(['reduce', 'unmount'] as const)('invalidates Search nextTick before it can queue a RAF on %s', async action => {
  const view = await mountSuspended(GameSearchResult, { props: searchProps })
  vi.useFakeTimers()
  try {
    await view.setProps({ gameList: games(3) })
    const raf = vi.spyOn(globalThis, 'requestAnimationFrame')
    vi.advanceTimersByTime(460)
    if (action === 'reduce') reduce(true)
    else view.unmount()
    await nextTick()
    expect(raf).not.toHaveBeenCalled()
    if (action === 'reduce') expect(view.findAll('.search-page-title').map(node => node.text())).toEqual(['Game 3', 'Game 4'])
  } finally { if (action !== 'unmount') view.unmount() }
})
