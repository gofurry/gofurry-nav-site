import { expect, it } from 'vitest'
import { collectionDate, collectionTime, collectionTimelineSections } from '../../app/utils/gameCollectionPresentation'
import { collectionDetail } from '../browser/fixtures/game-collections-data'
import zh from '../../i18n/locales/zh.json'
import en from '../../i18n/locales/en.json'

it('formats date-only facts without local timezone shifts or invented precision', () => {
  const previous = process.env.TZ
  process.env.TZ = 'America/Los_Angeles'
  try {
    expect(collectionDate('2020-01-02', 'en')).toBe('January 2, 2020')
    expect(collectionDate('2020-01-02', 'zh')).toBe('2020年1月2日')
    expect(collectionDate('2027-05-01', 'en', 'month')).toBe('May 2027')
    expect(collectionDate('2027-04-01', 'en', 'quarter')).toBe('Q2 2027')
    expect(collectionDate('2027-04-01', 'zh', 'quarter')).toBe('2027 Q2')
    expect(collectionDate('2004-01-01', 'zh', 'year')).toBe('2004')
  } finally { if (previous === undefined) delete process.env.TZ; else process.env.TZ = previous }
})
it.each([['zh', zh], ['en', en]] as const)('uses inferred/overdue/unknown wording from %s', (lang, messages) => {
  const t = (key: string, values?: Record<string, string | number>) => {
    const copy = messages.game.collections.time[key.split('.').at(-1)! as keyof typeof messages.game.collections.time]
    return copy.replace('{time}', String(values?.time ?? ''))
  }
  const items = collectionDetail('local').items
  expect(collectionTime(items[0]!, lang, t)).toBe(lang === 'zh' ? '约 2004' : 'Approx. 2004')
  expect(collectionTime(items[5]!, lang, t)).toContain(lang === 'zh' ? '当前仍未发售' : 'still upcoming')
  for (const i of [4, 9, 10]) expect(collectionTime(items[i]!, lang, t)).toBe(t('game.collections.time.' + items[i]!.phase))
})
it('partitions without sorting IDs or crossing unknown/NOW boundaries', () => {
  const source = collectionDetail('local').items
  const original = structuredClone(source)
  const sections = collectionTimelineSections(source)
  expect(sections.map(s => [s.key, s.snake, s.items.map(i => i.game_id)])).toEqual([
    ['past', true, ['21', '19', '7', '8']], ['releasedUnknown', false, ['30']],
    ['future', true, ['40', '41', '42', '43']], ['tba', false, ['50']], ['unknown', false, ['60']],
  ])
  expect(source).toEqual(original)
})
