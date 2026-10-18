import { describe, it, expect } from 'vitest'
import { groupUpdatesByMonth, updateCommit, updateDetailSeo } from '../../app/utils/updatePresentation'
import { formatUpdatesDate, formatUpdatesFullDate, updatesMonth } from '../../app/utils/updatesDate'
import { authoritativePageStatus } from '../../app/utils/authoritativePageError'
import { parseUpdateInventory } from '../../server/utils/sitemapInventory'

const item = { id: 1, title: 'Release', summary: '', version: null, commit_sha: null, published_at: '2026-08-31T23:30:00Z' }

describe('public Release Notes projection', () => {
  it('groups by the China-site calendar and preserves API order', () => {
    expect(updatesMonth(item.published_at)).toBe('2026 / 09')
    expect(updatesMonth('2026-09-01 07:30:00')).toBe('2026 / 09')
    expect(formatUpdatesFullDate(item.published_at, 'en-US')).toContain('07:30')
    expect(formatUpdatesDate(item.published_at, 'zh-CN')).toBe('2026/09/01')
    expect(formatUpdatesDate('invalid', 'en-US')).toBe('—')
    const groups = groupUpdatesByMonth([item, { ...item, id: 2 }, { ...item, id: 3, published_at: '2025-12-20T06:00:00Z' }])
    expect(groups.map(group => [group.month, group.items.map(release => release.id)])).toEqual([['2026 / 09', [1, 2]], ['2025 / 12', [3]]])
    expect(updatesMonth('invalid')).toBe('—')
  })
  it('uses short SHA text and full SHA destination without accepting malformed links', () => {
    expect(updateCommit(null)).toBeNull()
    expect(updateCommit('bad/url')).toBeNull()
    expect(updateCommit('abcdef0123456789')).toEqual({ label: 'abcdef0', href: 'https://github.com/gofurry/gofurry-nav-site/commit/abcdef0123456789' })
  })
  it('builds localized SEO from metadata only, with optional version and summary', () => {
    expect(updateDetailSeo(item, 'en')).toEqual({ title: 'Release | GoFurry', description: 'Read GoFurry feature updates, experience improvements and maintenance notes.' })
    expect(updateDetailSeo({ ...item, version: 'September', summary: 'A concise summary' }, 'zh')).toEqual({ title: 'September — Release | GoFurry', description: 'A concise summary' })
    expect(updateDetailSeo(item, 'zh').description).toContain('GoFurry')
  })
  it('preserves authoritative 404 and maps unavailable detail to 503', () => {
    expect(authoritativePageStatus({ statusCode: 404 }, 'update')).toBe(404)
    expect(authoritativePageStatus({ response: { status: 404 } }, 'update')).toBe(404)
    expect(authoritativePageStatus(new Error('offline'), 'update')).toBe(503)
  })
})

describe('Release Notes sitemap inventory fails closed', () => {
  const envelope = (state: string, items: unknown) => ({ code: 1, data: { schema_version: 1, state, items } })
  it('accepts public ready and success-empty inventory', () => {
    expect(parseUpdateInventory(envelope('ready', [item]))).toEqual([{ id: '1' }])
    expect(parseUpdateInventory(envelope('empty', []))).toEqual([])
  })
  it.each([
    null, { code: 0, data: {} }, { code: 1, data: {} }, envelope('error', []), envelope('ready', {}),
    envelope('ready', []), envelope('empty', [item]), envelope('ready', [{ title: 'missing' }]),
    ...[null, 0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1, '1', '../evil'].map(id => envelope('ready', [{ id }])),
  ])('rejects malformed inventory %j', value => { expect(() => parseUpdateInventory(value)).toThrow() })
})
