import { describe, expect, it } from 'vitest'
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
})
