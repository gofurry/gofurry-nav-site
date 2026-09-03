import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { getJSON } from '../../lib/api'
import type { DataOpsOverview, DatabaseStatus } from '../operations/types'
import { DataOperationsPage } from './data-operations-page'

vi.mock('../../lib/api', async (importOriginal) => ({ ...await importOriginal<typeof import('../../lib/api')>(), getJSON: vi.fn() }))
afterEach(cleanup)

function database(key: string): DatabaseStatus {
  return { key, health: 'healthy', postgresql_version: '17', database_name: key, database_size_bytes: 1024, total_connections: 2, active_connections: 1, max_connections: 100, connection_usage: 0.02, migration: { current_applied: 3, expected: 3, pending_count: 0, status: 'current' }, relations: [] }
}

describe('Data Operations database cards', () => {
  it('selects on first click and opens technical details on the selected card', async () => {
    vi.mocked(getJSON).mockResolvedValue({ generated_at: '2026-09-03T00:00:00Z', databases: [database('gfa'), database('gfn'), database('gfg')] } satisfies DataOpsOverview)
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(<QueryClientProvider client={client}><DataOperationsPage /></QueryClientProvider>)

    const gfn = await screen.findByRole('button', { name: /GFN/ })
    expect(gfn).toHaveAttribute('aria-pressed', 'false')
    await userEvent.click(gfn)
    expect(gfn).toHaveAttribute('aria-pressed', 'true')
    expect(screen.queryByText('技术详情 · GFN')).not.toBeInTheDocument()
    await userEvent.click(gfn)
    await waitFor(() => expect(screen.getByText('技术详情 · GFN')).toBeInTheDocument())
  })
})
