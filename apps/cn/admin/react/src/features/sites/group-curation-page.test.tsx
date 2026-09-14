import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { sendJSON } from '../../lib/api'
import { GroupCurationEditor, type GroupCuration } from './group-curation-page'

const auth = vi.hoisted(() => ({ canWrite: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: () => auth.canWrite }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), sendJSON: vi.fn() }))

const data: GroupCuration = { name: '社区', revision: 'original', sites: Array.from({ length: 10 }, (_, i) => ({ site_id: String(i + 1), name: `站点${i + 1}`, deleted: false })) }

function setup(value = data) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  const reload = vi.fn()
  const router = createMemoryRouter([{ path: '/', element: <GroupCurationEditor data={value} endpoint="/api/v1/nav/site-groups/1/curation" reload={reload} /> }])
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return { reload }
}

describe('homepage group curation', () => {
  afterEach(cleanup)
  beforeEach(() => { vi.clearAllMocks(); auth.canWrite = true; vi.mocked(sendJSON).mockResolvedValue(data) })

  it('moves across the Top 8 boundary and saves every membership without exposing weights', async () => {
    setup()
    const lists = screen.getAllByRole('list')
    expect(within(lists[0]).getAllByRole('listitem')).toHaveLength(8)
    expect(within(lists[1]).getAllByRole('listitem')).toHaveLength(2)
    expect(screen.getByRole('button', { name: '保存编排' })).toBeDisabled()
    fireEvent.click(screen.getByRole('button', { name: '上移 站点9' }))
    expect(within(screen.getAllByRole('list')[0]).getByText('站点9')).toBeInTheDocument()
    expect(within(screen.getAllByRole('list')[1]).getByText('站点8')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/nav/site-groups/1/curation', 'PUT', { revision: 'original', site_ids: ['1','2','3','4','5','6','7','9','8','10'] }))
  })

  it('can promote a remaining site directly and retains disabled membership', () => {
    setup({ ...data, sites: [...data.sites, { site_id: '11', name: '停用站点', deleted: true }] })
    fireEvent.click(screen.getAllByRole('button', { name: '移至首页首位' })[0])
    expect(within(screen.getAllByRole('list')[0]).getAllByRole('listitem')[0]).toHaveTextContent('站点9')
    expect(screen.getByText('停用站点')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '上移 停用站点' })).not.toBeInTheDocument()
  })

  it('keeps a rejected draft visible for conflict recovery', async () => {
    vi.mocked(sendJSON).mockRejectedValue(new Error('分组内容或排序已变化，请重新加载后编排'))
    setup()
    fireEvent.click(screen.getByRole('button', { name: '上移 站点9' }))
    fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
    expect(await screen.findByText('分组内容或排序已变化，请重新加载后编排')).toBeInTheDocument()
    expect(within(screen.getAllByRole('list')[0]).getByText('站点9')).toBeInTheDocument()
  })

  it('shows read-only users the preview without mutation controls', () => {
    auth.canWrite = false
    setup()
    expect(screen.getByText('首页展示（8/8）')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '保存编排' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '上移 站点2' })).not.toBeInTheDocument()
  })
})
