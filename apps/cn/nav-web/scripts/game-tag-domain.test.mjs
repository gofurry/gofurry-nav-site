import assert from 'node:assert/strict'
import { buildGameTagGroups, hasAdultTag } from '../app/utils/gameTagDomain.ts'

const tag = { id: '812345', code: 'adult', name: 'Adult', category_id: '99', category_code: 'future-category', category_name: 'Future', game_count: 1 }
const categories = [{ id: '99', code: 'future-category', name: 'Future', tags: [tag] }, { id: '3', code: 'empty', name: 'Empty', tags: [] }]
const groups = buildGameTagGroups(categories, [812345])
assert.deepEqual(groups.map(group => group.id), ['99', '3'])
assert.equal(groups[0].children.length, 1)
assert.equal(groups[0].children[0].selected, true)
assert.equal(groups[1].children.length, 0)
assert.equal(groups[0].children.some(child => child.id === '99'), false)
groups[0].expanded = true
assert.equal(buildGameTagGroups(categories, [], groups)[0].expanded, true)
assert.equal(buildGameTagGroups(categories, [], groups)[0].children[0].selected, false)
assert.equal(hasAdultTag([tag]), true)
assert.equal(hasAdultTag([{ id: '1014', code: 'unrelated' }]), false)
assert.equal(hasAdultTag(null), false)
console.log('Game Tag domain: explicit categories, selection, and code-based Adult checks PASS')
