import { expect, it } from 'vitest'
import { normalizeHeroID, normalizeHeroPreference, heroQuery, heroPreferenceKey } from '../../app/utils/heroPreferences'

it.each([undefined, null, 123, '', '0', '-1', '01', '1.1', '+1', ' 1', '1e3', '9223372036854775808'])('rejects invalid Hero ID %s', value => {
  expect(normalizeHeroID(value)).toBeNull()
})

it.each(['1', '9007199254740993', '9223372036854775807'])('preserves valid bigint ID %s as a string', value => {
  expect(normalizeHeroID(value)).toBe(value)
})

it('normalizes modes and builds the independent SSR query and preference identity', () => {
  expect(normalizeHeroPreference('invalid', null, null).mode).toBe('random')
  expect(heroQuery(normalizeHeroPreference('local', '10', '20'))).toEqual({ hero_mode: 'local' })
  expect(heroQuery(normalizeHeroPreference('random', '10', '20'))).toEqual({})
  const fixed = normalizeHeroPreference('fixed', '9007199254740993', '20')
  expect(heroQuery(fixed)).toEqual({ hero_desktop_id: '9007199254740993', hero_mobile_id: '20' })
  expect(heroPreferenceKey(fixed)).toBe('fixed:9007199254740993:20')
})
