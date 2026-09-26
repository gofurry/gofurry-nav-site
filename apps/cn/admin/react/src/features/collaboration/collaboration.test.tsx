import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider, useParams } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { ToastProvider } from '../../app/toast'
import { ApiError, getJSON, listJSON, sendJSON } from '../../lib/api'
import { GameWorkspacePage } from '../games/game-pages'
import { SiteWorkspacePage } from '../sites/site-pages'
import { WorkbenchPage } from '../workbench/workbench-page'
import { base } from './api'
import { CollaborationPage } from './collaboration-page'
import { BatchIdeaDialog, LinkIdeaDialog } from './idea-dialogs'
import { IdeaTableRow } from './idea-table'
import { IdeaContextBanner } from './idea-context'
import { SharedBoard, BoardNoteCard } from './board'
import type { BoardNote, Idea, IdeaInput, PreviewRow } from './types'

vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))
const auth = vi.hoisted(() => ({ capabilities: ['content.read', 'content.write', 'collaboration.read', 'collaboration.write'] }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (value: string) => auth.capabilities.includes(value), state: { identity: { account_id: 1 } } }) }))
const idea: Idea = { id: 7, kind: 'game', title: '线索', source: 'Steam:82', source_key: 'steam:82', note: '先确认详情', priority: 'normal', status: 'idea', version: 3, created_by_account_id: 1, researching_by_account_id: null, creator_name: '成员', researcher_name: '', linked_kind: null, linked_resource_id: null, created_at: '2026-09-26T00:00:00Z' }
const note: BoardNote = { id: 5, body: '共享文字', x: 0, y: 0, width: 320, height: 260, z_index: 0, version: 2, updated_by_account_id: 1 }
function setup(element: ReactNode, path = '/', routes: { path: string; element: ReactNode }[] = []) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  const router = createMemoryRouter([{ path: path.split('?')[0].replace(/\/new$/, '/:id'), element }, ...routes], { initialEntries: [path] })
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return { client, router }
}
function HandoffFixture({ kind }: { kind: 'game' | 'site' }) { const { id } = useParams(); return id === 'new' ? kind === 'game' ? <GameWorkspacePage /> : <SiteWorkspacePage /> : <IdeaContextBanner kind={kind} resourceID={42} /> }

beforeEach(() => { auth.capabilities = ['content.read', 'content.write', 'collaboration.read', 'collaboration.write'] })
afterEach(() => { cleanup(); vi.resetAllMocks() })

it('previews batches and defaults to skipping known rows on create', async () => {
  vi.mocked(sendJSON).mockImplementation(async (path, _method, body) => path.endsWith('batch-preview') ? (body as { items: IdeaInput[] }).items.map((row, index) => ({ ...row, index, valid: true, warnings: index ? ['resource_exists'] : [], errors: [], source_key: `steam:${index}`, display_hint: '', idea_match: null, resource_match: null, batch_duplicate_of: null } satisfies PreviewRow)) : { inserted_count: 1, skipped_count: 1 })
  setup(<BatchIdeaDialog close={vi.fn()} />)
  fireEvent.change(screen.getByLabelText('粘贴内容'), { target: { value: '123\n456' } })
  fireEvent.click(screen.getByRole('button', { name: '预览' }))
  await screen.findByText('正式 Game / Site 已存在')
  fireEvent.click(screen.getByRole('button', { name: '只加入新内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/ideas/batch`, 'POST', expect.objectContaining({ skip_known: true, items: expect.any(Array) })))
  expect(await screen.findByText('已加入 1 条，跳过 1 条。')).toBeInTheDocument()
})

it('provides a visible-delimiter example without replacing user input', async () => {
  setup(<BatchIdeaDialog close={vi.fn()} />)
  fireEvent.click(screen.getByRole('button', { name: '填入示例' }))
  expect(screen.getByLabelText('粘贴内容')).toHaveValue('求生之路 2 | steam:550 | 核对介绍\n传送门 2 | steam:620 | 补充资料')
  expect(screen.getByRole('button', { name: '填入示例' })).toBeDisabled()
  fireEvent.change(screen.getByLabelText('粘贴内容'), { target: { value: '我的内容' } })
  expect(screen.getByRole('button', { name: '填入示例' })).toBeDisabled()
  expect(sendJSON).not.toHaveBeenCalled()
})

it('blocks malformed space-separated rows before server preview and exposes parsed columns', async () => {
  vi.mocked(sendJSON).mockImplementation(async (_path, _method, body) => (body as { items: IdeaInput[] }).items.map((row, index) => ({ ...row, index, valid: true, warnings: [], errors: [], source_key: row.source, display_hint: row.source, idea_match: null, resource_match: null, batch_duplicate_of: null })))
  setup(<BatchIdeaDialog close={vi.fn()} />)
  fireEvent.change(screen.getByLabelText('粘贴内容'), { target: { value: '550\n620\ntest steam:666 测试' } })
  fireEvent.click(screen.getByRole('button', { name: '预览' }))
  expect(await screen.findByText(/第 3 行/)).toHaveTextContent('空格不会分列')
  expect(sendJSON).not.toHaveBeenCalled()
  expect(screen.getByRole('button', { name: '只加入新内容' })).toBeDisabled()
  fireEvent.change(screen.getByLabelText('粘贴内容'), { target: { value: 'test | steam:666 | 测试' } })
  fireEvent.click(screen.getByRole('button', { name: '预览' }))
  await screen.findByRole('columnheader', { name: '来源' })
  expect(screen.getByRole('cell', { name: 'test' })).toBeInTheDocument()
  expect(screen.getByRole('cell', { name: 'steam:666' })).toBeInTheDocument()
  expect(screen.getByRole('cell', { name: '测试' })).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledWith(`${base}/ideas/batch-preview`, 'POST', { items: [expect.objectContaining({ title: 'test', source: 'steam:666', note: '测试' })] })
})

it('requires confirmation to delete an idea and retains it on a version conflict', async () => {
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('stale', 409))
  setup(<table><tbody><IdeaTableRow idea={{ ...idea, status: 'landed', linked_kind: 'game', linked_resource_id: 42 }} /></tbody></table>)
  await userEvent.click(screen.getByRole('button', { name: '想法 7 的更多操作' }))
  await userEvent.click(await screen.findByRole('menuitem', { name: '删除想法' }))
  const dialog = await screen.findByRole('dialog', { name: '删除想法' })
  expect(within(dialog).getByText(/已关联的正式游戏或网站会保留/)).toBeInTheDocument()
  expect(sendJSON).not.toHaveBeenCalled()
  await userEvent.click(within(dialog).getByRole('button', { name: '确认删除' }))
  await screen.findByText('此内容刚刚被其他成员修改，请重新加载后重试。')
  expect(sendJSON).toHaveBeenCalledExactlyOnceWith(`${base}/ideas/7`, 'DELETE', { version: 3 })
  expect(within(screen.getByRole('table', { hidden: true })).getByText('线索')).toBeInTheDocument()
  expect(dialog).toBeInTheDocument()
  await userEvent.click(within(dialog).getByRole('button', { name: '重新加载' }))
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
})

it('hides edit and deletion from read-only members', async () => {
  auth.capabilities = ['collaboration.read']
  setup(<table><tbody><IdeaTableRow idea={idea} /></tbody></table>)
  expect(screen.queryByRole('button', { name: '编辑' })).not.toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: '想法 7 的更多操作' }))
  expect(screen.queryByRole('menuitem', { name: '删除想法' })).not.toBeInTheDocument()
})

it('keeps filters in the URL and uses active inventory by default', async () => {
  vi.mocked(getJSON).mockImplementation(async (url) => url.includes('/ideas?') ? { list: [], total: 0 } : { reserve_count: 100, researching_count: 2, landed_30d: 3 })
  const { router } = setup(<CollaborationPage />, '/collaboration')
  await screen.findByText('100 条储备 · 2 条正在整理 · 近 30 天已落地 3 条')
  expect(vi.mocked(getJSON).mock.calls.some(([url]) => url.includes('status=active'))).toBe(true)
  fireEvent.change(screen.getByLabelText('搜索想法'), { target: { value: 'Steam' } })
  await waitFor(() => expect(router.state.location.search).toContain('keyword=Steam'))
})

it('retains modal edits on 409 and explicitly reloads without retrying a stale write', async () => {
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('stale', 409))
  vi.mocked(getJSON).mockResolvedValue({ ...idea, version: 4, note: '其他成员的修改' })
  setup(<table><tbody><IdeaTableRow idea={idea} /></tbody></table>)
  fireEvent.click(screen.getByRole('button', { name: '编辑' }))
  fireEvent.change(screen.getByLabelText('备注'), { target: { value: '我的修改' } })
  fireEvent.click(screen.getByRole('button', { name: '保存' }))
  await screen.findByText('此内容刚刚被其他成员修改，请重新加载后重试。')
  expect(screen.getByLabelText('备注')).toHaveValue('我的修改')
  expect(sendJSON).toHaveBeenCalledWith(`${base}/ideas/7`, 'PUT', expect.objectContaining({ version: 3, note: '我的修改' }))
  fireEvent.click(screen.getByRole('button', { name: '重新加载' }))
  await waitFor(() => expect(screen.getByLabelText('备注')).toHaveValue('其他成员的修改'))
  expect(within(screen.getByRole('table', { hidden: true })).queryByRole('textbox', { hidden: true })).not.toBeInTheDocument()
  expect(screen.getByRole('dialog', { name: '编辑想法' })).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledTimes(1)
})

it('offers navigation after another researcher conflicts without taking ownership', async () => {
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('该内容目前由成员乙整理', 409))
  const { router } = setup(<table><tbody><IdeaTableRow idea={idea} /></tbody></table>, '/', [{ path: '/game/games/new', element: <p>录入页面</p> }])
  fireEvent.click(screen.getByRole('button', { name: '录入游戏' }))
  await screen.findByText('该内容目前由成员乙整理')
  fireEvent.click(screen.getByRole('button', { name: '仍然打开录入页面' }))
  await waitFor(() => expect(router.state.location.search).toBe('?idea=7'))
  expect(sendJSON).toHaveBeenCalledTimes(1)
})

it('reuses existing game options for manual linking and supports Other landing', async () => {
  vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [{ id: '42', label: '正式游戏' }] })
  vi.mocked(sendJSON).mockResolvedValue({ ...idea, status: 'landed' })
  setup(<LinkIdeaDialog idea={idea} close={vi.fn()} />)
  expect(await screen.findByRole('listbox')).not.toHaveClass('absolute')
  fireEvent.click(await screen.findByRole('option', { name: '正式游戏' }))
  fireEvent.click(screen.getByRole('button', { name: '确认关联' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/ideas/7/link`, 'POST', { version: 3, kind: 'game', resource_id: 42 }))
  expect(listJSON).toHaveBeenCalledWith('/api/v1/options/games', 1, 10, '')
  cleanup()
  setup(<table><tbody><IdeaTableRow idea={{ ...idea, kind: 'other' }} /></tbody></table>)
  await userEvent.click(screen.getByRole('button', { name: '想法 7 的更多操作' }))
  await userEvent.click(await screen.findByRole('menuitem', { name: '标记已落地' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/ideas/7/land`, 'POST', { version: 3 }))
})

it.each([{ kind: 'game' as const, fails: true }, { kind: 'site' as const, fails: true }, { kind: 'game' as const, fails: false }, { kind: 'site' as const, fails: false }])('prefills safe $kind context and preserves formal success (link fails: $fails)', async ({ kind, fails }) => {
  const current = { ...idea, kind, title: '来源名称' }
  vi.mocked(getJSON).mockResolvedValue(current)
  const formal = { id: 42, name: '完整名称', name_en: '', info: '', info_en: '', appid: 82, developers: [], publishers: [], groups: [], links: [], resources: [], header: '', weight: 0 }
  vi.mocked(sendJSON).mockImplementation(async (path) => { if (path.includes('/collaboration/')) { if (fails) throw new ApiError('GFA unavailable', 500); const landed = { ...current, status: 'landed' }; vi.mocked(getJSON).mockResolvedValue(landed); return landed } return formal })
  const path = kind === 'game' ? '/game/games' : '/nav/sites'
  const { router } = setup(<HandoffFixture kind={kind} />, `${path}/new?idea=7`)
  if (kind === 'game') {
    await waitFor(() => expect(screen.getByRole('spinbutton', { name: /Steam AppID/ })).toHaveValue(82))
    expect(screen.getByRole('textbox', { name: /中文名称/ })).toHaveValue('')
  } else await waitFor(() => expect(screen.getByRole('textbox', { name: /中文名称/ })).toHaveValue('来源名称'))
  expect(vi.mocked(getJSON).mock.calls.every(([url]) => !url.includes('steam-prefill'))).toBe(true)
  fireEvent.change(screen.getByRole('textbox', { name: /中文名称/ }), { target: { value: '完整名称' } })
  fireEvent.click(screen.getByRole('button', { name: kind === 'game' ? '创建游戏' : '创建网站' }))
  await waitFor(() => expect(router.state.location.pathname).toBe(`${path}/42`))
  expect(router.state.location.search).toBe('?idea=7')
  if (fails) expect(await screen.findByText(/正式内容已创建，但想法关联未完成/)).toBeInTheDocument()
  else expect(await screen.findByText('已落地')).toBeInTheDocument()
  expect(vi.mocked(sendJSON).mock.calls.some(([url, method]) => method === 'DELETE' || url.includes('collector-domains'))).toBe(false)
  vi.mocked(sendJSON).mockResolvedValue({ ...current, status: 'landed' })
  if (fails) fireEvent.click(screen.getByRole('button', { name: '关联到当前内容' }))
  else expect(screen.queryByRole('button', { name: '关联到当前内容' })).not.toBeInTheDocument()
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${base}/ideas/7/link`, 'POST', { version: 3, kind, resource_id: 42 }))
  expect(vi.mocked(sendJSON).mock.calls.filter(([url]) => url === `/api/v1${path}`).length).toBe(1)
})

it('shows neutral Workbench inventory only with collaboration.read', async () => {
  vi.mocked(getJSON).mockResolvedValue({ collaboration: { reserve_count: 100, researching_count: 1, landed_30d: 2 }, attention: [], system_status: [], recent_changes: [], recent_operations: [] })
  setup(<WorkbenchPage />)
  await screen.findByText('100 条储备 · 1 条正在整理 · 近 30 天已落地 2 条')
  expect(screen.getByText('当前没有对你可见的待处理异常。')).toBeInTheDocument()
  cleanup(); auth.capabilities = ['content.read']; setup(<WorkbenchPage />)
  await screen.findByText('当前没有对你可见的待处理异常。')
  expect(screen.queryByText('内容储备')).not.toBeInTheDocument()
})

it('creates and edits a text note with versioned saves', async () => {
  vi.mocked(getJSON).mockResolvedValue([note]); vi.mocked(sendJSON).mockResolvedValue({ ...note, version: 3 })
  setup(<SharedBoard />)
  fireEvent.change(await screen.findByLabelText('新便笺正文'), { target: { value: '新便笺' } })
  fireEvent.click(screen.getByRole('button', { name: '新建便笺' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/board/notes`, 'POST', expect.objectContaining({ body: '新便笺' })))
  fireEvent.change(screen.getByLabelText('便笺正文 5'), { target: { value: '修改正文' } })
  fireEvent.click(screen.getByRole('button', { name: '保存正文' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/board/notes/5`, 'PUT', expect.objectContaining({ body: '修改正文', version: 2 })))
})

it.each(['移动便笺 5', '调整便笺尺寸 5'])('persists %s only on pointerup, then handles conflict and versioned deletion', async (label) => {
  class TestPointerEvent extends MouseEvent { pointerId = 1 }
  const previous = window.PointerEvent
  window.PointerEvent = TestPointerEvent as unknown as typeof PointerEvent
  if (!HTMLElement.prototype.setPointerCapture) Object.defineProperty(HTMLElement.prototype, 'setPointerCapture', { configurable: true, value: () => {} })
  if (!HTMLElement.prototype.hasPointerCapture) Object.defineProperty(HTMLElement.prototype, 'hasPointerCapture', { configurable: true, value: () => false })
  const capture = vi.spyOn(HTMLElement.prototype, 'setPointerCapture').mockImplementation(() => {})
  const has = vi.spyOn(HTMLElement.prototype, 'hasPointerCapture').mockReturnValue(false)
  try {
    vi.mocked(sendJSON).mockRejectedValue(new ApiError('stale', 409))
    setup(<BoardNoteCard note={note} topZ={0} writable />)
    const handle = screen.getByLabelText(label)
    fireEvent.pointerDown(handle, { button: 0, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(handle, { clientX: 30, clientY: 40 })
    expect(sendJSON).not.toHaveBeenCalled()
    fireEvent.pointerUp(handle, { clientX: 30, clientY: 40 })
    await waitFor(() => expect(sendJSON).toHaveBeenCalledTimes(1))
    expect(sendJSON).toHaveBeenCalledWith(`${base}/board/notes/5`, 'PUT', expect.objectContaining(label.startsWith('移动') ? { version: 2, x: 30, y: 40 } : { version: 2, width: 350, height: 300 }))
    await screen.findByText('此内容刚刚被其他成员修改，请重新加载后重试。')
    fireEvent.click(screen.getByRole('button', { name: '重新加载' }))
    vi.mocked(sendJSON).mockResolvedValue(note)
    fireEvent.click(screen.getByRole('button', { name: '删除' }))
    fireEvent.click(screen.getByRole('button', { name: '确认删除' }))
    await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/board/notes/5`, 'DELETE', { version: 2 }))
  } finally { window.PointerEvent = previous; capture.mockRestore(); has.mockRestore() }
})

it('keeps unsaved Board text mounted when a background refresh fails', async () => {
 vi.mocked(getJSON).mockResolvedValue([note])
 const { client } = setup(<SharedBoard />)
 const input = await screen.findByLabelText('便笺正文 5')
 fireEvent.change(input, { target: { value: '尚未保存的正文' } })
 vi.mocked(getJSON).mockRejectedValue(new Error('temporary network failure'))
 await client.invalidateQueries({ queryKey: ['collaboration', 'board'] })
 await screen.findByText('刷新失败，保留当前便笺与未保存的编辑。')
 expect(screen.getByLabelText('便笺正文 5')).toHaveValue('尚未保存的正文')
 expect(sendJSON).not.toHaveBeenCalled()
})
