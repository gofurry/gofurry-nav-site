import { describe, expect, it } from 'vitest'
import { DATAOPS_READ_CAPABILITY } from '../../lib/capabilities'
import { capabilityAwareNavigation } from './app-shell'

describe('capability-aware navigation', () => {
  it('shows content routes without reproducing role checks', () => {
    const capabilities = new Set(['content.read', 'content.write', 'collection.read'])
    const entries = capabilityAwareNavigation((capability) => capabilities.has(capability)).flatMap((group) => group.entries)
    expect(entries.map((entry) => entry.href)).toContain('/nav/sites')
    expect(entries.map((entry) => entry.href)).toContain('/collection')
    expect(entries.map((entry) => entry.href)).not.toContain('/system/accounts')
  })

  it('reveals owner-controlled areas only through native capabilities', () => {
    const capabilities = new Set(['content.read', 'account.manage', 'audit.read'])
    const entries = capabilityAwareNavigation((capability) => capabilities.has(capability)).flatMap((group) => group.entries)
    expect(entries.map((entry) => entry.href)).toContain('/system/accounts')
    expect(entries.map((entry) => entry.href)).toContain('/system/audit')
  })

  it('uses only the canonical dataops.read capability', () => {
    const canonical = capabilityAwareNavigation((capability) => capability === DATAOPS_READ_CAPABILITY).flatMap((group) => group.entries)
    const legacyAlias = capabilityAwareNavigation((capability) => capability === 'data_ops.read').flatMap((group) => group.entries)

    expect(DATAOPS_READ_CAPABILITY).toBe('dataops.read')
    expect(canonical.map((entry) => entry.href)).toContain('/system/data-operations')
    expect(legacyAlias.map((entry) => entry.href)).not.toContain('/system/data-operations')
  })

  it('exposes only Operator operational areas', () => {
    const capabilities = new Set(['content.read', 'collection.read', 'collection.execute', 'metrics.read', 'changes.read'])
    const paths = capabilityAwareNavigation((capability) => capabilities.has(capability)).flatMap((group) => group.entries.map((entry) => entry.href))
    expect(paths).toEqual(expect.arrayContaining(['/collection', '/metrics', '/changes']))
    expect(paths).not.toEqual(expect.arrayContaining(['/system/data-operations', '/system/audit', '/system/accounts']))
  })

  it('adds engineering areas for Developer but keeps accounts Owner-only', () => {
    const capabilities = new Set(['content.read', 'collection.read', 'metrics.read', 'changes.read', 'dataops.read', 'audit.read'])
    const paths = capabilityAwareNavigation((capability) => capabilities.has(capability)).flatMap((group) => group.entries.map((entry) => entry.href))
    expect(paths).toEqual(expect.arrayContaining(['/system/data-operations', '/system/audit']))
    expect(paths).not.toContain('/system/accounts')
  })

  it('shows account governance only with account.manage', () => {
    const paths = capabilityAwareNavigation((capability) => capability === 'account.manage').flatMap((group) => group.entries.map((entry) => entry.href))
    expect(paths).toContain('/system/accounts')
  })
})
