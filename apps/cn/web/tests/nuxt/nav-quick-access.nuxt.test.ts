import { mountSuspended } from '@nuxt/test-utils/runtime'
import { expect, it } from 'vitest'
import NavQuickAccess from '../../app/components/nav/NavQuickAccess.vue'

it('keeps invalid URLs static, falls back on icon errors and emits original recent/custom sites', async () => {
  const recentSites = [
    { id: 'recent', name: 'Recent', url: 'https://recent.example/path?q=value#section' },
    { id: 'bad-recent', name: 'Broken Recent', url: 'not a URL' },
  ]
  const customSites = [
    { id: 'custom', name: 'Custom', url: 'https://custom.example/path' },
    { id: 'bad-custom', name: 'Invalid Custom', url: 'javascript:alert(1)' },
  ]
  const view = await mountSuspended(NavQuickAccess, { props: { recentSites, customSites } })
  try {
    expect(view.findAll('img')).toHaveLength(2)
    expect(view.find('[title="Broken Recent"] img').exists()).toBe(false)
    expect(view.find('[title="Broken Recent"] .quick-site-fallback').text()).toBe('B')
    expect(view.find('[title="Invalid Custom"] img').exists()).toBe(false)
    expect(view.find('[title="Invalid Custom"] .quick-site-fallback').text()).toBe('I')
    for (const [name, hostname] of [['Recent', 'recent.example'], ['Custom', 'custom.example']]) {
      const tile = view.find(`[title="${name}"]`)
      expect(tile.find('img').attributes('src')).toBe(`https://a.favicon.im/${hostname}?larger=true&throw-error-on-404=true`)
      await tile.find('img').trigger('error')
      expect(tile.find('img').exists()).toBe(false)
      expect(tile.find('.quick-site-fallback').text()).toBe(name!.slice(0, 1))
      await tile.trigger('click')
    }
    expect(view.emitted('visit-recent')).toEqual([[recentSites[0]]])
    expect(view.emitted('visit-custom')).toEqual([[customSites[0]]])
    expect(view.findAll('img')).toHaveLength(0)
  } finally { view.unmount() }
})
