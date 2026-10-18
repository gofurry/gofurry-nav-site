import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { getJSON, listJSON, sendJSON } from '../../lib/api'
import type { SiteWorkspace } from '../../lib/types'
import { SiteListPage, SiteWorkspacePage } from './site-pages'

const auth = vi.hoisted(() => ({ write: true, history: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (cap: string) => cap === 'changes.read' ? auth.history : auth.write }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))

const workspace: SiteWorkspace = {
  site: { id: 7, name: '归档示例', name_en: 'Archive Example', info: '原有中文简介', info_en: 'Original description', create_time: '2026-01-01', update_time: '2026-10-18', country: 'CN', nsfw: '0', welfare: '1', icon: 'icon.svg', icon_primary_url: '/icon.svg', deleted: true, deleted_at: '2026-10-18T01:00:00Z', deletion_reason_code: 'unreachable', deletion_reason: '人工检查记录' },
  former_primary_target: 'old.example.test', targets: [{ id: 9, site_id: 7, name: 'example.test', prefix: 'old.', proxy: '0', tls: '1', primary: false }], groups: [], featured: null,
}
function setup(path: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  const router = createMemoryRouter([{ path: '/nav/sites', element: <SiteListPage /> }, { path: '/nav/sites/:id', element: <SiteWorkspacePage /> }], { initialEntries: [path] })
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return router
}

describe('Site archival', () => {
  afterEach(cleanup)
  beforeEach(() => {
    vi.clearAllMocks(); auth.write = true; auth.history = true
    vi.mocked(getJSON).mockImplementation(async (path) => path.includes('/changes/events') ? { list: [], total: 0 } : workspace)
    vi.mocked(listJSON).mockResolvedValue({ list: [], total: 0 })
    vi.mocked(sendJSON).mockResolvedValue({})
  })
  it('keeps archived pages readable without mounting mutation controls for a writer', async () => {
    setup('/nav/sites/7')
    expect(await screen.findByText('已归档 · 只读')).toBeInTheDocument()
    expect(screen.getByText('无法访问')).toBeInTheDocument()
    expect(screen.getByText('人工检查记录')).toBeInTheDocument()
    for (const tab of ['内容', '分类与展示', '采集', '数据状态', '历史']) {
      fireEvent.click(screen.getByRole('tab', { name: tab }))
      expect(screen.queryByRole('button', { name: /归档|保存|采集|设为|上传|清除|删除|新增目标/ })).not.toBeInTheDocument()
      if (tab === '内容') {
        expect(screen.getByText('原有中文简介')).toBeInTheDocument()
        expect(screen.getByRole('img', { name: '归档示例 图标' })).toHaveAttribute('src', '/icon.svg')
      }
      if (tab === '采集') expect(screen.getByText('历史 / 停用配置')).toBeInTheDocument()
    }
    await screen.findByText('暂无可用历史')
    expect(sendJSON).not.toHaveBeenCalled()
  })
  it('does not invent legacy reasons or bypass the existing history capability', async () => {
    auth.history = false
    vi.mocked(getJSON).mockResolvedValue({ ...workspace, site: { ...workspace.site, deletion_reason_code: null, deletion_reason: null }, former_primary_target: '' })
    setup('/nav/sites/7?tab=history')
    expect(await screen.findByText('未记录（历史归档）')).toBeInTheDocument()
    expect(screen.getByText('无历史主目标')).toBeInTheDocument()
    expect(getJSON).not.toHaveBeenCalledWith(expect.stringContaining('/changes/events'))
  })
  it('requires an explicit reason and sends a trimmed DELETE body; failures keep the draft', async () => {
    vi.mocked(getJSON).mockResolvedValue({ ...workspace, site: { ...workspace.site, deleted: false } })
    vi.mocked(sendJSON).mockRejectedValueOnce(new Error('audit unavailable'))
    const user = userEvent.setup()
    setup('/nav/sites/7')
    await user.click(await screen.findByRole('button', { name: '归档' }))
    await user.click(screen.getByRole('button', { name: '确认归档' }))
    expect(await screen.findByText('请选择归档原因')).toBeInTheDocument()
    expect(sendJSON).not.toHaveBeenCalled()
    await user.click(screen.getByRole('combobox', { name: '归档原因' }))
    await user.click(await screen.findByRole('option', { name: '无法访问' }))
    fireEvent.change(screen.getByRole('textbox', { name: '归档说明' }), { target: { value: '  核对记录  ' } })
    await user.click(screen.getByRole('button', { name: '确认归档' }))
    expect(await screen.findByText('audit unavailable')).toBeInTheDocument()
    expect(sendJSON).toHaveBeenCalledWith('/api/v1/nav/sites/7', 'DELETE', { deletion_reason_code: 'unreachable', deletion_reason: '核对记录' })
    expect(screen.getByRole('textbox', { name: '归档说明' })).toHaveValue('  核对记录  ')
    vi.mocked(getJSON).mockResolvedValue(workspace)
    await user.click(screen.getByRole('button', { name: '确认归档' }))
    expect(await screen.findByText('已归档 · 只读')).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
  it('persists status/search in URL, resets pagination and keys requests by status', async () => {
    const user = userEvent.setup()
    const router = setup('/nav/sites?search=old.example.test&page=3')
    await waitFor(() => expect(listJSON).toHaveBeenCalledWith('/api/v1/nav/site-summaries?status=active', 3, 50, 'old.example.test'))
    await user.click(screen.getByRole('combobox', { name: '网站状态' }))
    await user.click(await screen.findByRole('option', { name: '已归档' }))
    await waitFor(() => expect(listJSON).toHaveBeenCalledWith('/api/v1/nav/site-summaries?status=archived', 1, 50, 'old.example.test'))
    expect(router.state.location.search).toContain('status=archived')
    expect(router.state.location.search).toContain('page=1')
  })
  it('rejects an overlong note and cancel discards the draft without sending a request', async () => {
    vi.mocked(getJSON).mockResolvedValue({ ...workspace, site: { ...workspace.site, deleted: false } })
    const user = userEvent.setup()
    setup('/nav/sites/7')
    await user.click(await screen.findByRole('button', { name: '归档' }))
    await user.click(screen.getByRole('combobox', { name: '归档原因' }))
    await user.click(await screen.findByRole('option', { name: '其他原因' }))
    fireEvent.change(screen.getByRole('textbox', { name: '归档说明' }), { target: { value: '🦊'.repeat(1001) } })
    await user.click(screen.getByRole('button', { name: '确认归档' }))
    expect(await screen.findByText('说明最多 1000 字符')).toBeInTheDocument()
    expect(sendJSON).not.toHaveBeenCalled()
    await user.click(screen.getByRole('button', { name: '取消' }))
    await user.click(screen.getByRole('button', { name: '归档' }))
    expect(screen.getByRole('textbox', { name: '归档说明' })).toHaveValue('')
    expect(screen.getByRole('combobox', { name: '归档原因' })).toHaveTextContent('请选择原因')
  })
  it('does not expose archive to a content reader', async () => {
    auth.write = false
    vi.mocked(getJSON).mockResolvedValue({ ...workspace, site: { ...workspace.site, deleted: false } })
    setup('/nav/sites/7?tab=content')
    expect(await screen.findByText('原有中文简介')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /归档|保存内容/ })).not.toBeInTheDocument()
  })
})
