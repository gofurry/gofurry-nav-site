import { afterEach, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useNuxtApp } from '#app'
import NavHeader from '../../app/components/nav/NavHeader.vue'
import { CUSTOM_SITES_STORAGE_KEY, loadCustomSites, saveCustomSites } from '../../app/utils/customSites'

const seed = [
  { id: 'art', name: 'Artwork', url: 'https://art.example/' },
  { id: 'wiki', name: 'Wiki', url: 'https://wiki.example/' },
]
const mount = () => mountSuspended(NavHeader, { attachTo: document.body, global: { stubs: { HeroBackground: true, SearchBox: true } } })
afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); localStorage.clear() })

it('restores the persisted keyboard order on a fresh component and returns focus on Escape', async () => {
  saveCustomSites(seed)
  const view = await mount()
  try {
    const opener = view.find<HTMLButtonElement>('.quick-site-manage')
    opener.element.focus(); await opener.trigger('click'); await nextTick()
    const grip = view.find<HTMLButtonElement>('.quick-modal-grip')
    grip.element.focus(); await grip.trigger('keydown', { key: 'ArrowDown' }); await nextTick()
    expect(loadCustomSites()).toEqual([...seed].reverse())
    expect(document.activeElement).toBe(view.findAll('.quick-modal-grip')[1]!.element)
    await grip.trigger('keydown', { key: 'Escape' }); await nextTick()
    expect(document.activeElement).toBe(opener.element)
  } finally { view.unmount() }
  const restored = await mount()
  try {
    await restored.find('.quick-site-manage').trigger('click'); await nextTick()
    expect(restored.findAll('.quick-modal-item-name').map(node => node.text())).toEqual(['Wiki', 'Artwork'])
  } finally { restored.unmount() }
})

it('uses pointer intent beyond the paint boundary to move a taller row past the final shorter row', async () => {
  saveCustomSites(seed)
  const view = await mount()
  try {
    await view.find('.quick-site-manage').trigger('click'); await nextTick()
    const grip = view.find<HTMLButtonElement>('.quick-modal-grip')
    let captured = false
    Object.assign(grip.element, {
      setPointerCapture: () => { captured = true },
      hasPointerCapture: () => captured,
      releasePointerCapture: () => { captured = false },
    })
    const rows = view.findAll<HTMLElement>('.quick-modal-item')
    vi.spyOn(rows[0]!.element, 'getBoundingClientRect').mockReturnValue({ top: 0, height: 80 } as DOMRect)
    vi.spyOn(rows[1]!.element, 'getBoundingClientRect').mockReturnValue({ top: 88, height: 50 } as DOMRect)
    vi.spyOn(view.find('.quick-modal-list').element, 'getBoundingClientRect').mockReturnValue({ top: 0, bottom: 300 } as DOMRect)
    let frame: FrameRequestCallback = () => {}
    vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => { frame = callback; return 721 })
    vi.stubGlobal('cancelAnimationFrame', vi.fn())
    const write = vi.spyOn(localStorage, 'setItem')
    await grip.trigger('pointerdown', { button: 0, isPrimary: true, pointerId: 1, clientY: 40 })
    await grip.trigger('pointermove', { pointerId: 1, clientY: 120 })
    frame(0); await nextTick()
    expect(rows[0]!.element.style.transform).toBe('translateY(58px)')
    expect(rows[1]!.element.style.transform).toBe('translateY(-88px)')
    expect(loadCustomSites()).toEqual(seed)
    expect(write.mock.calls.filter(([key]) => key === CUSTOM_SITES_STORAGE_KEY)).toHaveLength(0)
    await grip.trigger('pointerup', { pointerId: 1, clientY: 120 }); await nextTick()
    expect(loadCustomSites()).toEqual([...seed].reverse())
    expect(write.mock.calls.filter(([key]) => key === CUSTOM_SITES_STORAGE_KEY)).toHaveLength(1)
    expect(captured).toBe(false)
  } finally { view.unmount() }
})

for (const cancel of ['route', 'unmount'] as const) {
  it(`cancels captured pointer work without saving on ${cancel}`, async () => {
    saveCustomSites(seed)
    const view = await mount()
    let mounted = true
    try {
      await view.find('.quick-site-manage').trigger('click'); await nextTick()
      const grip = view.find<HTMLButtonElement>('.quick-modal-grip')
      let captured = false
      Object.assign(grip.element, {
        setPointerCapture: vi.fn(() => { captured = true }),
        hasPointerCapture: vi.fn(() => captured),
        releasePointerCapture: vi.fn(() => { captured = false }),
      })
      view.findAll('.quick-modal-item').forEach((node, index) => vi.spyOn(node.element, 'getBoundingClientRect').mockReturnValue({ top: index * 68, height: 60 } as DOMRect))
      const raf = vi.fn(() => 721), stop = vi.fn()
      vi.stubGlobal('requestAnimationFrame', raf); vi.stubGlobal('cancelAnimationFrame', stop)
      const write = vi.spyOn(localStorage, 'setItem')
      await grip.trigger('pointerdown', { button: 0, isPrimary: true, pointerId: 1, clientY: 30 })
      expect(captured).toBe(true)
      expect(raf).toHaveBeenCalled()
      if (cancel === 'route') { await useNuxtApp().$router.push('/terms'); await nextTick() }
      else { view.unmount(); mounted = false }
      expect(captured).toBe(false)
      expect(stop).toHaveBeenCalledWith(721)
      expect(write.mock.calls.filter(([key]) => key === CUSTOM_SITES_STORAGE_KEY)).toHaveLength(0)
      expect(loadCustomSites()).toEqual(seed)
      if (mounted) expect(view.find('.quick-modal-backdrop').exists()).toBe(false)
    } finally { if (mounted) view.unmount() }
  })
}
