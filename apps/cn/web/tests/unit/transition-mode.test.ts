import { describe, expect, it } from 'vitest'
import { isTransitionMode, isTransitionPageOpen, isTransitionSiteAlias } from '../../shared/utils/transitionMode'

describe('transition runtime policy', () => {
  it('requires an explicit true, including runtime environment strings', () => {
    for (const value of [undefined, null, false, 'false', '', '0', '1']) expect(isTransitionMode(value)).toBe(false)
    expect(isTransitionMode(true)).toBe(true)
    expect(isTransitionMode('true')).toBe(true)
  })
  it('opens only localized Home and positive-ID detail pages', () => {
    for (const path of ['/', '/en', '/en/', '/site/41', '/en/site/41/', '/games/82', '/en/games/82']) expect(isTransitionPageOpen(path), path).toBe(true)
    for (const path of ['/games', '/games/search', '/games/prize', '/games/82/news', '/site/0', '/site/abc', '/sites', '/updates/1', '/en/about', '/insights']) expect(isTransitionPageOpen(path), path).toBe(false)
  })
  it('retains only the two existing Site detail aliases', () => {
    for (const path of ['/sites/41', '/en/sites/41', '/site/41/target.example', '/en/site/41/target.example']) expect(isTransitionSiteAlias(path), path).toBe(true)
    for (const path of ['/sites', '/sites/abc', '/site/41/target.example/history', '/games/search']) expect(isTransitionSiteAlias(path), path).toBe(false)
  })
})
