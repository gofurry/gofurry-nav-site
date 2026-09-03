import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { listJSON } from '../../lib/api'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '../../app/toast'
import { ResourcePage, RemoteSelect } from './resource-page'

vi.mock('../../lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../lib/api')>()
  return { ...actual, listJSON: vi.fn() }
})

vi.mock('../auth/auth-context', () => ({
  useAuth: () => ({ can: () => false }),
}))

function renderResource(section: 'nav' | 'game', resource: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <MemoryRouter>
      <QueryClientProvider client={client}>
        <ToastProvider><ResourcePage section={section} resource={resource} /></ToastProvider>
      </QueryClientProvider>
    </MemoryRouter>,
  )
}

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

describe('Resource Engine route definitions', () => {
  beforeEach(() => vi.mocked(listJSON).mockResolvedValue({ list: [], total: 0 }))

  it.each([
    ['nav', 'site-groups', '网站分组'],
    ['nav', 'update-notices', '更新公告'],
    ['nav', 'sayings', '金句'],
    ['game', 'tags', '标签'],
    ['game', 'comments', '评论'],
    ['game', 'prizes', '抽奖'],
  ] as const)('renders %s/%s with the correct definition', (section, resource, title) => {
    renderResource(section, resource)
    expect(screen.getByRole('heading', { name: title })).toBeInTheDocument()
    expect(screen.queryByText('未找到资源定义。')).not.toBeInTheDocument()
  })
})
