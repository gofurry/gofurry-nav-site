import { expect, it } from 'vitest'
import { parsePaste } from './parser'
import { sourceURL } from './api'

it('accepts complete sources without canonicalizing or guessing kinds', () => {
  expect(parsePaste('123\r\nSteam:456\nhttps://store.steampowered.com/app/82/x\nwww.example.test', 'game').map(({ title, source }) => [title, source])).toEqual([['', '123'], ['', 'Steam:456'], ['', 'https://store.steampowered.com/app/82/x'], ['', 'www.example.test']])
  expect(parsePaste('123 | |', 'other')[0].title).toBe('123')
})
it('splits visible pipes while retaining spaces inside titles and notes', () => {
  expect(parsePaste('求生之路 2 | steam:550 | 核对 中文 介绍\r\n传送门 2 | 620\n标题 | | 备注\n| 123 |\n', 'game')).toMatchObject([
    { title: '求生之路 2', source: 'steam:550', note: '核对 中文 介绍' },
    { title: '传送门 2', source: '620', note: '' },
    { title: '标题', source: '', note: '备注' },
    { title: '', source: '123', note: '' },
  ])
})
it('does not silently turn space-separated columns into an idea title', () => {
  expect(() => parsePaste('550\n\n test steam:666 测试', 'game')).toThrow('第 3 行')
  expect(() => parsePaste('test steam:666 测试', 'game')).toThrow('空格不会分列')
  expect(() => parsePaste('标题\tsteam:666\t测试', 'game')).toThrow('Tab')
  expect(() => parsePaste('标题｜steam:666｜测试', 'game')).toThrow('名称 | 来源 | 备注')
})
it('requires at most three columns and a title or source', () => {
  expect(() => parsePaste('标题 | 123 | 备注 | 意外多列', 'game')).toThrow('最多三列')
  expect(() => parsePaste('| | 仅备注', 'other')).toThrow('名称和来源')
})
it('supports escaped literal pipes and preserves ordinary backslashes and quotes', () => {
  expect(parsePaste(String.raw`A \| B | https://example.test | "引号" C:\notes`, 'site')[0]).toMatchObject({ title: 'A | B', note: String.raw`"引号" C:\notes` })
  expect(parsePaste(String.raw`A | | 反斜线 \\`, 'other')[0].note).toBe('反斜线 \\')
})
it('bounds batches and summarizes errors with physical line numbers', () => {
  expect(parsePaste(Array(500).fill('line | |').join('\n'), 'other')).toHaveLength(500)
  expect(() => parsePaste(Array(501).fill('line | |').join('\n'), 'other')).toThrow('500')
  expect(() => parsePaste(' \n', 'other')).toThrow('粘贴')
  expect(() => parsePaste(Array(6).fill('bad line').join('\n'), 'game')).toThrow('另有 1 行格式错误')
})
it('never renders executable or credential-bearing source links', () => {
  expect(sourceURL('javascript:alert(1)')).toBeUndefined()
  expect(sourceURL('https://name:secret@example.test')).toBeUndefined()
  expect(sourceURL('Steam:82', 'steam:82')).toBe('https://store.steampowered.com/app/82/')
})
