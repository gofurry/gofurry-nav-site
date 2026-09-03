import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { getJSON, listJSON, sendJSON } from '../../lib/api'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '../../app/toast'
import { ResourcePage, RemoteSelect } from './resource-page'

vi.mock('../../lib/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../lib/api')>()
  return { ...actual, getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }
})

afterEach(cleanup)

const authTestState = vi.hoisted(() => ({ canWrite: false }))
vi.mock('../auth/auth-context', () => ({
  useAuth: () => ({ can: (capability: string) => capability === 'content.write' && authTestState.canWrite }),
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
  beforeEach(() => { vi.clearAllMocks(); authTestState.canWrite = false; vi.mocked(listJSON).mockResolvedValue({ list: [], total: 0 }) })

  it('queries the existing option endpoint with operator search text', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/games" value="" onChange={() => undefined} /></QueryClientProvider>)
    await waitFor(() => expect(listJSON).toHaveBeenCalledWith('/api/v1/options/games', 1, 50, ''))
    await userEvent.type(screen.getByPlaceholderText('搜索远程选项…'), 'steam')
    await waitFor(() => expect(listJSON).toHaveBeenLastCalledWith('/api/v1/options/games', 1, 50, 'steam'))
  })

  it('resolves an existing ID into one searchable combobox with layered metadata', async () => {
    vi.mocked(listJSON).mockResolvedValue({ list: [{ id: '307', label: 'Hubert', extra: 'Hubert · AppID 764930 · ID 307' }], total: 1 })
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(<QueryClientProvider client={client}><RemoteSelect endpoint="/api/v1/options/games" value="307" onChange={() => undefined} /></QueryClientProvider>)

    const combobox = await screen.findByRole('combobox')
    await waitFor(() => expect(combobox).toHaveValue('Hubert'))
    await userEvent.click(combobox)
    expect(screen.getByRole('option', { name: /Hubert.*AppID 764930.*ID 307/ })).toBeInTheDocument()
  })
})

describe('Resource Engine route definitions', () => {
  beforeEach(() => { vi.clearAllMocks(); authTestState.canWrite = false; vi.mocked(listJSON).mockResolvedValue({ list: [], total: 0 }) })

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

  it('opens and updates a comment with its full 64-bit string ID', async () => {
    authTestState.canWrite = true
    const id = '9007199254740993'
    const comment = { id, game_id: 307, region: 'CN', name: 'Tester', ip: '127.0.0.1', score: 5, content: '精确 ID 评论', create_time: '2026-09-03T00:00:00Z' }
    vi.mocked(listJSON).mockImplementation(async (endpoint) => endpoint === '/api/v1/game/comments' ? { list: [comment], total: 1 } : { list: [], total: 0 })
    vi.mocked(getJSON).mockResolvedValue(comment)
    vi.mocked(sendJSON).mockResolvedValue(comment)

    renderResource('game', 'comments')
    await userEvent.click(await screen.findByText('精确 ID 评论'))

    await waitFor(() => expect(getJSON).toHaveBeenCalledWith(`/api/v1/game/comments/${id}`))
    const content = await screen.findByLabelText('内容')
    await userEvent.clear(content)
    await userEvent.type(content, '已更新评论')
    await userEvent.click(screen.getByRole('button', { name: '保存修改' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/comments/${id}`, 'PUT', expect.objectContaining({ content: '已更新评论' })))
  })

  it('deletes a comment with its full 64-bit string ID', async () => {
    authTestState.canWrite = true
    const id = '9007199254740993'
    vi.mocked(listJSON).mockResolvedValue({ list: [{ id, game_id: 307, name: 'Tester', score: 5, content: '待删除评论', create_time: '2026-09-03T00:00:00Z' }], total: 1 })
    vi.mocked(sendJSON).mockResolvedValue(undefined)

    renderResource('game', 'comments')
    await userEvent.click(await screen.findByRole('button', { name: '更多操作' }))
    await userEvent.click(await screen.findByText('删除'))
    await userEvent.click(await screen.findByRole('button', { name: '确认删除' }))

    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/comments/${id}`, 'DELETE'))
  })
})
