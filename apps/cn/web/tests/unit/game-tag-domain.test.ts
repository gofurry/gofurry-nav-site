import { expect, it } from 'vitest'
import { buildGameTagGroups, hasAdultTag } from '../../app/utils/gameTagDomain'

const tag = { id: '812345', code: 'adult', name: 'Adult', category_id: '99', category_code: 'future-category', category_name: 'Future', game_count: 1 }
const categories = [{ id: '99', code: 'future-category', name: 'Future', tags: [tag] }, { id: '3', code: 'empty', name: 'Empty', tags: [] }]

it('uses explicit categories, including empty/new categories, without synthetic parent tags', () => {
  const groups = buildGameTagGroups(categories, [812345])
  expect(groups.map(group => group.id)).toEqual(['99', '3'])
  expect(groups[0]!.children).toHaveLength(1)
  expect(groups[0]!.children[0]!.selected).toBe(true)
  expect(groups[1]!.children).toHaveLength(0)
  expect(groups[0]!.children.some(child => child.id === '99')).toBe(false)
  groups[0]!.expanded = true
  expect(buildGameTagGroups(categories, [], groups)[0]!.expanded).toBe(true)
  expect(buildGameTagGroups(categories, [], groups)[0]!.children[0]!.selected).toBe(false)
})

it('identifies Adult by code rather than the legacy numeric ID', () => {
  const legacyId = { id: '1014', code: 'unrelated' }
  expect(hasAdultTag([tag])).toBe(true)
  expect(hasAdultTag([legacyId])).toBe(false)
  expect(hasAdultTag(null)).toBe(false)
})
