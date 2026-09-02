import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { AuthProvider, useAuth } from './auth-context'

describe('auth capability integration', () => {
  afterEach(() => vi.unstubAllGlobals())

  it('consumes the current-principal capability list from auth state', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ code: 1, message: '', data: { initialized: true, authenticated: true, identity: { account_id: 3, username: 'operator', display_name: 'Operator', role: 'operator', status: 'active', session_version: 1, capabilities: ['content.read', 'content.write'] } } }), { status: 200, headers: { 'Content-Type': 'application/json' } })))
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const wrapper = ({ children }: { children: ReactNode }) => <QueryClientProvider client={client}><AuthProvider>{children}</AuthProvider></QueryClientProvider>
    const { result } = renderHook(() => useAuth(), { wrapper })
    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(result.current.can('content.write')).toBe(true)
    expect(result.current.can('account.manage')).toBe(false)
  })
})
