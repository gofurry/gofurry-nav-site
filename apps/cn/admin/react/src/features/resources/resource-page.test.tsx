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
    const combobox = screen.getByPlaceholderText('搜索远程选项…')
    expect(combobox).toHaveAttribute('autocomplete', 'off')
    await userEvent.type(combobox, 'steam')
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

  it('exposes group-centric curation from the group list', async () => {
    vi.mocked(listJSON).mockResolvedValue({ list: [{ id: 42, name: '社区' }], total: 1 })
    renderResource('nav', 'site-groups')
    expect(await screen.findByRole('link', { name: '首页编排' })).toHaveAttribute('href', '/nav/site-groups/42/curation')
  })

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

describe('Tag domain resources', () => {
  beforeEach(() => { vi.clearAllMocks(); authTestState.canWrite = true })
  it.each([['tags', '标签'], ['tag-categories', '标签类别']] as const)('offers archive/restore instead of deletion for %s', async (resource, title) => {
    const row = { id: 12, code: 'stable-code', name: '测试条目', name_en: 'Test', archived_at: null }
    vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [row] })
    vi.mocked(sendJSON).mockResolvedValue(row)
    const view = renderResource('game', resource)
    await userEvent.click(await screen.findByRole('button', { name: '归档' }))
    expect(screen.getByRole('heading', { name: `归档${title}` })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: '确认归档' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/${resource}/12`, 'DELETE'))
    view.unmount()
    vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [{ ...row, archived_at: '2026-09-16' }] })
    renderResource('game', resource)
    await userEvent.click(await screen.findByRole('button', { name: '恢复' }))
    await userEvent.click(screen.getByRole('button', { name: '确认恢复' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/${resource}/12/restore`, 'POST'))
  })
  it('keeps existing tag code immutable and resolves the category name', async () => {
    const row = { id: 12, code: 'stable-code', name: '现有标签', name_en: 'Tag', info: '', info_en: '', category_id: 2 }
    vi.mocked(listJSON).mockImplementation(async endpoint => endpoint.includes('/options/') ? { total: 1, list: [{ id: '2', label: '物种' }] } : { total: 1, list: [row] })
    vi.mocked(getJSON).mockResolvedValue(row)
    vi.mocked(sendJSON).mockResolvedValue(row)
    renderResource('game', 'tags')
    await userEvent.click(await screen.findByText('现有标签'))
    expect(await screen.findByLabelText(/^Code/)).toBeDisabled()
    await waitFor(() => expect(screen.getByRole('combobox')).toHaveValue('物种'))
    const input = screen.getByLabelText('中文名称')
    await userEvent.clear(input); await userEvent.type(input, '新名称')
    await userEvent.click(screen.getByRole('button', { name: '保存修改' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/tags/12', 'PUT', expect.objectContaining({ code: 'stable-code', name: '新名称', category_id: 2 })))
  })
})

it.each([['tags', '标签'], ['tag-categories', '标签类别']] as const)('creates %s with an explicit code and no manual ID', async (resource, title) => {
  vi.clearAllMocks(); authTestState.canWrite = true
  vi.mocked(listJSON).mockImplementation(async endpoint => endpoint.includes('/options/') ? { total: 1, list: [{ id: '2', label: '物种' }] } : { total: 0, list: [] })
  vi.mocked(sendJSON).mockResolvedValue({ id: 777 })
  renderResource('game', resource)
  await userEvent.click(screen.getByRole('button', { name: `新增${title}` }))
  const code = await screen.findByLabelText(/^Code/)
  expect(code).toBeEnabled()
  expect(screen.queryByLabelText('标签 ID')).not.toBeInTheDocument()
  await userEvent.type(code, 'new-identity')
  await userEvent.type(screen.getByLabelText('中文名称'), '新条目')
  await userEvent.type(screen.getByLabelText('英文名称'), 'New Item')
  if (resource === 'tags') {
    await userEvent.click(screen.getByRole('combobox'))
    await userEvent.click(await screen.findByRole('option', { name: '物种' }))
  }
  await userEvent.click(screen.getByRole('button', { name: '创建' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/game/${resource}`, 'POST', expect.objectContaining({ code: 'new-identity', name: '新条目' })))
  expect(vi.mocked(sendJSON).mock.calls[0]?.[2]).not.toHaveProperty('id')
})
