import { describe, expect, it } from 'vitest'
import { steamPrefillValues } from './steam-prefill'

describe('Steam partial prefill', () => {
  it.each([
    { name: '中文', name_en: 'English' },
    { name_en: 'English', info_en: 'Description' },
    { name: '中文', info: '简介' },
    { header: 'https://example.com/header.jpg' },
  ])('applies all available fields: %j', (data) => {
    expect(steamPrefillValues(data)).toEqual(data)
  })

  it('preserves existing input when upstream fields are empty or absent', () => {
    const current = { name: '手工名称', developers: ['原开发者'], header: 'old.jpg', info: '原简介' }
    const result = { ...current, ...steamPrefillValues({ name: '', header: null, developers: [], info: '  ', name_en: 'English' } as never) }
    expect(result).toEqual({ ...current, name_en: 'English' })
  })
})
