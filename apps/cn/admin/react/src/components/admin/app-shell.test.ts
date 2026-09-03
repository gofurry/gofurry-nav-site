import { createElement } from 'react'
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { DATAOPS_READ_CAPABILITY } from '../../lib/capabilities'
import { isGlobalSearchShortcut } from '../../lib/keyboard'
import { AdminBrand, capabilityAwareHeaderActions, capabilityAwareNavigation, logoutAndRedirect } from './app-shell'

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

  it('keeps Ctrl/Cmd+K and slash global-search shortcuts without hijacking form input', () => {
    expect(isGlobalSearchShortcut({ key: 'k', ctrlKey: true, metaKey: false }, true)).toBe(true)
    expect(isGlobalSearchShortcut({ key: 'k', ctrlKey: false, metaKey: true }, false)).toBe(true)
    expect(isGlobalSearchShortcut({ key: '/', ctrlKey: false, metaKey: false }, false)).toBe(true)
    expect(isGlobalSearchShortcut({ key: '/', ctrlKey: false, metaKey: false }, true)).toBe(false)
  })

  it('shows a text-only brand that follows the sidebar state', () => {
    const { rerender } = render(createElement(AdminBrand, { collapsed: false }))
    expect(screen.getByText('GoFurry')).toBeInTheDocument()
    expect(screen.queryByText('GF')).not.toBeInTheDocument()
    expect(screen.queryByText('GoFurry Admin')).not.toBeInTheDocument()
    expect(screen.queryByText('V3 CONTENT')).not.toBeInTheDocument()

    rerender(createElement(AdminBrand, { collapsed: true }))
    expect(screen.getByText('GF')).toBeInTheDocument()
    expect(screen.queryByText('GoFurry')).not.toBeInTheDocument()
  })

  it('shows the Collection header shortcut only with collection.read', () => {
    expect(capabilityAwareHeaderActions((capability) => capability === 'collection.read').map((action) => action.href)).toEqual(['/collection'])
    expect(capabilityAwareHeaderActions(() => false)).toEqual([])
  })

  it('replaces the protected route only after logout state has been cleared', async () => {
    const events: string[] = []
    const logout = vi.fn(async () => { events.push('logout') })
    const navigate = vi.fn(() => { events.push('navigate') })

    await logoutAndRedirect(logout, navigate)

    expect(events).toEqual(['logout', 'navigate'])
    expect(navigate).toHaveBeenCalledWith('/login', { replace: true })
  })
})
