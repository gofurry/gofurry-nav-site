import { afterEach, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useNuxtApp } from '#app'
import NavUtilities from '../../app/components/nav/NavUtilities.vue'
import { writeMode } from '../../app/utils/modeStorage'
import type { Site } from '../../app/types/nav'

const directory = vi.hoisted(() => vi.fn())
vi.mock('../../app/services/nav', () => ({ getNavSiteDirectory: directory }))
afterEach(() => { vi.restoreAllMocks(); directory.mockReset(); localStorage.removeItem('mode') })
const entries = Array.from({ length: 12 }, (_, i) => ({ id: String(i), name: `Site ${i}`, info: 'wolf community', nsfw: i === 0 ? '1' : '0', icon: '', domain: 'example.com' } as Site))
function desktop() {
  const media = Object.assign(new EventTarget(), { matches: false }) as MediaQueryList
  vi.spyOn(window, 'matchMedia').mockReturnValue(media)
  return media
}

it('loads only on open, caches the locale and preserves live mode filtering and the eight-item limit', async () => {
  desktop()
  writeMode('sfw')
  directory.mockResolvedValue(entries)
  const view = await mountSuspended(NavUtilities, { attachTo: document.body })
  try {
    expect(directory).not.toHaveBeenCalled()
    const trigger = view.find('.gf-nav__utility-row button')
    await trigger.trigger('click'); await flushPromises()
    await view.find('input').setValue('wolf')
    expect(view.findAll('.gf-nav__search-results strong').map(node => node.text())).toEqual(entries.slice(1, 9).map(site => site.name))
    writeMode('nsfw'); await nextTick()
    expect(view.findAll('.gf-nav__search-results strong').map(node => node.text())).toEqual(entries.slice(0, 8).map(site => site.name))
    await view.find('input').trigger('keydown', { key: 'Escape' })
    expect(view.find('section').exists()).toBe(false)
    expect(document.activeElement).toBe(trigger.element)
    await trigger.trigger('click'); await flushPromises()
    expect(directory).toHaveBeenCalledTimes(1)
    expect(view.find('input').element.value).toBe('')
    await view.setProps({ blocked: true })
    expect(view.find('section').exists()).toBe(false)
  } finally { view.unmount() }
})

it('invalidates pending responses across close, rapid reopen, locale change and unmount', async () => {
  const media = desktop()
  const remove = vi.spyOn(media, 'removeEventListener')
  const i18n = useNuxtApp().$i18n
  const previous = i18n.locale.value
  await i18n.loadLocaleMessages('en')
  i18n.locale.value = 'zh'
  const pending: Array<{ resolve: (sites: Site[]) => void, signal: AbortSignal }> = []
  directory.mockImplementation((_lang: string, signal: AbortSignal) => new Promise<Site[]>(resolve => pending.push({ resolve, signal })))
  const view = await mountSuspended(NavUtilities)
  try {
    const trigger = view.find('.gf-nav__utility-row button')
    await trigger.trigger('click'); await flushPromises()
    expect(view.find('section').attributes('aria-busy')).toBe('true')
    await trigger.trigger('click')
    expect(pending[0]!.signal.aborted).toBe(true)
    await trigger.trigger('click'); await flushPromises()
    pending[0]!.resolve(entries); await flushPromises()
    expect(view.find('section').attributes('aria-busy')).toBe('true')
    expect(view.findAll('.gf-nav__search-results button')).toHaveLength(0)
    i18n.locale.value = 'en'; await nextTick()
    expect(pending[1]!.signal.aborted).toBe(true)
    pending[1]!.resolve(entries); await flushPromises()
    expect(view.find('section').exists()).toBe(false)
    await trigger.trigger('click'); await flushPromises()
    expect(directory.mock.calls.map(call => call[0])).toEqual(['zh', 'zh', 'en'])
  } finally { view.unmount(); i18n.locale.value = previous }
  expect(pending[2]!.signal.aborted).toBe(true)
  pending[2]!.resolve(entries); await flushPromises()
  expect(remove).toHaveBeenCalledWith('change', expect.any(Function))
})
