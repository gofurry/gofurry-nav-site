import { describe, expect, it } from 'vitest'
import { emptyGameShowcase, showcaseContextKey, showcaseDestination, showcaseFocalPoint, showcaseReleaseDate } from '../../app/utils/gameShowcasePresentation'
import { assetURL, assetCandidate } from '../../app/utils/managedAssets'
import type { GameShowcaseItem, GameShowcaseRelease } from '../../app/types/game'

describe('Showcase presentation keeps the frozen public contract', () => {
  it('keeps independent empty arrays and bigint string route IDs', () => {
    expect(emptyGameShowcase().items).toEqual([])
    expect(emptyGameShowcase().items).not.toBe(emptyGameShowcase().items)
    expect(showcaseDestination({ type: 'game', game_id: '9007199254740993' })).toEqual({ path: '/games/9007199254740993', external: false })
    expect(showcaseDestination({ type: 'game', game_id: 'a/b' })?.path).toBe('/games/a%2Fb')
    expect(showcaseDestination({ type: 'game', target: 'https://example.com' })).toBeNull()
  })
  it('only renders valid external actions and never invents a target', () => {
    for (const target of ['', 'javascript:alert(1)', '//example.com', 'http://example.com', 'https://user:pass@example.com']) {
      expect(showcaseDestination({ type: 'website', target })).toBeNull()
    }
    expect(showcaseDestination({ type: 'steam', target: 'https://store.steampowered.com/app/570' }))
      .toEqual({ path: 'https://store.steampowered.com/app/570', external: true })
    expect(showcaseDestination(undefined)).toBeNull()
  })
  it('routes every reason and sponsored content type to localized context', () => {
    for (const [reason, key] of Object.entries({ editorial: 'editorial', upcoming: 'upcoming', new_release: 'newRelease', trending: 'trending' })) {
      expect(showcaseContextKey({ reason, sponsored: false } as GameShowcaseItem)).toBe(`gameShowcase.context.${key}`)
    }
    for (const content_type of ['game', 'crowdfunding', 'tabletop', 'merchandise', 'other']) {
      expect(showcaseContextKey({ reason: 'editorial', sponsored: true, content_type } as GameShowcaseItem))
        .toBe(`gameShowcase.context.sponsored.${content_type}`)
    }
  })
  it('formats canonical calendar precision without guessed dates', () => {
    const release: GameShowcaseRelease = { availability: 'upcoming', precision: 'day', exact_date: '2027-02-14', year: 2027, month: 2, quarter: null, window_start: null, window_end: null }
    expect(showcaseReleaseDate(release, 'en')).toBe('Feb 14, 2027')
    expect(showcaseReleaseDate(release, 'zh')).toBe('2027年2月14日')
    expect(showcaseReleaseDate({ ...release, exact_date: '2027-02-30' }, 'en')).toBe('')
    expect(showcaseReleaseDate({ ...release, precision: 'quarter', quarter: 3 }, 'en')).toBe('Q3 2027')
    expect(showcaseReleaseDate({ ...release, precision: 'year' }, 'zh')).toBe('2027 年')
    expect(showcaseReleaseDate({ ...release, precision: 'month' }, 'en')).toBe('Feb 2027')
    expect(showcaseReleaseDate({ ...release, precision: 'tba' }, 'en')).toBe('')
    expect(showcaseReleaseDate({ ...release, availability: 'unknown' }, 'zh')).toBe('')
  })
  it('uses only Showcase AVIF keys through the existing provider fallback chain', () => {
    const origins = { primary: 'https://primary.example', mirror: 'https://mirror.example' }
    for (const variant of ['desktop', 'mobile']) {
      const key = `game/showcase/9007199254740993/${variant}/${'a'.repeat(32)}.avif`
      expect(assetURL(origins, 'primary', key)).toBe(`${origins.primary}/${key}`)
      expect(assetCandidate(origins, 'primary', key, new Set(['primary']), '').url).toBe(`${origins.mirror}/${key}`)
      expect(assetCandidate(origins, 'primary', key, new Set(['primary', 'mirror']), '').url).toBe('')
      for (const invalid of [key.replace('/9007199254740993/', '/0/'), key.replace('.avif', '.jpg'), `${key}?x=1`, key.replace(`/${variant}/`, '/../')]) {
        expect(assetURL(origins, 'primary', invalid)).toBe('')
      }
    }
    expect([undefined, NaN, -1, 0, .25, 1, 2].map(showcaseFocalPoint)).toEqual([50, 50, 0, 0, 25, 100, 100])
  })
})
