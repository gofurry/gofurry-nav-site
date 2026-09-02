import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { listJSON } from '../../lib/api'
import { RemoteSelect } from './resource-page'

vi.mock('../../lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../lib/api')>()
  return { ...actual, listJSON: vi.fn() }
})

describe('Resource Engine remote options', () => {
  beforeEach(() => vi.mocked(listJSON).mockResolvedValue({ list: [], total: 0 }))

  it('queries the existing option endpoint with operator search text', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/games" value="" onChange={() => undefined} /></QueryClientProvider>)
    await waitFor(() => expect(listJSON).toHaveBeenCalledWith('/api/v1/options/games', 1, 50, ''))
    await userEvent.type(screen.getByPlaceholderText('搜索远程选项…'), 'steam')
    await waitFor(() => expect(listJSON).toHaveBeenLastCalledWith('/api/v1/options/games', 1, 50, 'steam'))
  })
})
