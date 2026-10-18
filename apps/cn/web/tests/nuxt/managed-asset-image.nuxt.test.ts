import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import ManagedAssetImage from '../../app/components/common/ManagedAssetImage.vue'

// happy-dom does not load images. Model pending resources until explicit events;
// real browser coverage owns cached/SSR completion and actual network failures.
beforeEach(() => vi.spyOn(HTMLImageElement.prototype, 'complete', 'get').mockReturnValue(false))
afterEach(() => vi.restoreAllMocks())

it('keeps existing single-image attrs/load listeners and Primary-to-Mirror fallback', async () => {
  const load = vi.fn(), key = `nav/hero/desktop/${'a'.repeat(32)}.avif`
  const wrapper = await mountSuspended(ManagedAssetImage, { props: { objectKey: key, fallback: '' }, attrs: { class: 'consumer-image', onLoad: load } })
  try {
    expect(wrapper.find('picture').exists()).toBe(false)
    expect(wrapper.get('img').classes()).toContain('consumer-image')
    expect(wrapper.get('img').attributes('src')).toBe(`https://primary.example/${key}`)
    await wrapper.get('img').trigger('load')
    expect(load).toHaveBeenCalledTimes(1)
    await wrapper.get('img').trigger('error')
    expect(wrapper.get('img').attributes('src')).toBe(`https://mirror.example/${key}`)
    await wrapper.get('img').trigger('error')
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.emitted('exhausted')).toHaveLength(1)
  } finally { wrapper.unmount() }
})

it('routes only the selected picture source on failure and preserves the other viewport resource', async () => {
  let resized!: () => void
  const query = { matches: true, addEventListener: vi.fn((_event, callback) => { resized = callback }), removeEventListener: vi.fn() }
  vi.spyOn(window, 'matchMedia').mockReturnValue(query as unknown as MediaQueryList)
  const desktop = `game/showcase/1/desktop/${'a'.repeat(32)}.avif`, mobile = `game/showcase/1/mobile/${'b'.repeat(32)}.avif`
  const wrapper = await mountSuspended(ManagedAssetImage, { props: { objectKey: desktop, mobileObjectKey: mobile, fallback: '' }, attrs: { class: 'consumer-image' } })
  try {
    expect(wrapper.findAll('img')).toHaveLength(1)
    expect(wrapper.get('source').attributes('media')).toBe('(max-width: 1023px)')
    expect(wrapper.get('img').classes()).toContain('consumer-image')
    await wrapper.get('img').trigger('error')
    expect(wrapper.get('source').attributes('srcset')).toBe(`https://mirror.example/${mobile}`)
    expect(wrapper.get('img').attributes('src')).toBe(`https://primary.example/${desktop}`)
    await wrapper.get('img').trigger('error')
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.emitted('exhausted')).toHaveLength(1)
    query.matches = false; resized(); await wrapper.vm.$nextTick()
    expect(wrapper.get('img').attributes('src')).toBe(`https://primary.example/${desktop}`)
  } finally { wrapper.unmount() }
  expect(query.removeEventListener).toHaveBeenCalledWith('change', resized)
})
