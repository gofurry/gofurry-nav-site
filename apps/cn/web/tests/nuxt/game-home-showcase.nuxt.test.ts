import { expect, it, vi } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import GameHomeShowcase from '../../app/components/game/main/content/GameHomeShowcase.vue'
import { makeShowcase } from '../browser/fixtures/games-home-showcase-data'

vi.mock('../../app/services/game', () => ({ submitGameShowcaseEvent: vi.fn().mockResolvedValue(undefined) }))

it('resets index and silent live region for a new snapshot and on a new page instance', async () => {
  const snapshot = makeShowcase('four-items', 'en')
  const wrapper = await mountSuspended(GameHomeShowcase, { props: { snapshot } })
  try {
    expect(wrapper.get('h2').text()).toBe(snapshot.items[0]!.title)
    expect(wrapper.get('[aria-live]').text()).toBe('')
    await wrapper.get('.game-home-showcase__controls button:last-child').trigger('click')
    expect(wrapper.get('h2').text()).toBe(snapshot.items[1]!.title)
    expect(wrapper.get('[aria-live]').text()).toContain(snapshot.items[1]!.title)
    await wrapper.setProps({ snapshot: { ...snapshot, snapshot_id: 'b'.repeat(32) } })
    expect(wrapper.get('h2').text()).toBe(snapshot.items[0]!.title)
    expect(wrapper.get('[aria-live]').text()).toBe('')
    await wrapper.get('.game-home-showcase__controls button:last-child').trigger('click')
  } finally { wrapper.unmount() }
  const reentered = await mountSuspended(GameHomeShowcase, { props: { snapshot } })
  try { expect(reentered.get('h2').text()).toBe(snapshot.items[0]!.title) }
  finally { reentered.unmount() }
})

it('hides an editorial note on Sponsored and emits plain text even for markup-like copy', async () => {
  const snapshot = makeShowcase('managed-sponsored-tabletop', 'en')
  snapshot.items[0]!.editorial_note = 'must never render'
  snapshot.items[0]!.title = '<script>not HTML</script>'
  const wrapper = await mountSuspended(GameHomeShowcase, { props: { snapshot } })
  try {
    expect(wrapper.get('.game-home-showcase__note').text()).toBe('')
    expect(wrapper.get('h2').text()).toBe('<script>not HTML</script>')
    expect(wrapper.find('script').exists()).toBe(false)
    expect(wrapper.get('.gf-button').attributes('rel')).toBe('noopener noreferrer')
  } finally { wrapper.unmount() }
})
