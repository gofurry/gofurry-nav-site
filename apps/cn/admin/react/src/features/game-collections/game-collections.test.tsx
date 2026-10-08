import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { ApiError, getJSON, listJSON, sendJSON } from '../../lib/api'
import { collectionEndpoint, eligibleCollectionEndpoint, homeKey, loadCollectionWorkspace, workspaceKey } from './api'
import { CollectionListPage } from './collection-list-page'
import { CollectionWorkspacePage, CreateCollectionPage } from './collection-workspace-page'
import { CollectionHomeCurationPage } from './home-curation-page'
import type { CollectionComposition, CollectionHome, CollectionMember, CollectionRuleTag, CollectionWorkspace, GameCollection } from './types'

const auth = vi.hoisted(() => ({ capabilities: ['content.read', 'content.write', 'audit.read'] }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => auth.capabilities.includes(capability) }) }))
vi.mock('../../lib/api', async original => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))

function collection(id = 1): GameCollection {
  return { id, code: `collection-${id}`, name: `分区 ${id}`, name_en: `Collection ${id}`, info: '中文简介', info_en: 'Description', status: 'draft', version: 5, member_count: 2, sfw_member_count: 1, home_slot: null, published_at: null, archived_at: null, created_at: '2026-10-06T00:00:00Z', updated_at: '2026-10-06T00:00:00Z' }
}
let current: GameCollection
let members: CollectionMember[]
let ruleTags: CollectionRuleTag[]
let automatic: CollectionMember[]
let excluded: CollectionMember[]
let home: CollectionHome
const newGame: CollectionMember = { game_id: 3, name: '新增游戏', name_en: 'New Game', appid: 103, adult: true }
const newTag: CollectionRuleTag = { tag_id: 10, code: 'adventure', name: '冒险', name_en: 'Adventure', active: true }
function composition(version = current.version): CollectionComposition {
  const auto = ruleTags.some(tag => tag.active) ? automatic : []
  const ids = [...new Set([...auto, ...members].map(game => game.game_id))]
  const effective = ids.filter(id => !excluded.some(game => game.game_id === id)).map(id => {
    const manual = members.find(game => game.game_id === id)
    const match = auto.find(game => game.game_id === id)
    return { ...(manual ?? match)!, source: manual ? match ? 'both' as const : 'manual' as const : 'automatic' as const }
  })
  return { collection_id: current.id, version, home_slot: current.home_slot, rule_tags: ruleTags, manual_members: members, excluded_members: excluded, effective_members: effective, counts: { auto_matched: auto.length, manual_pinned: members.length, excluded: excluded.length, effective: effective.length, sfw_visible: effective.filter(game => !game.adult).length } }
}
beforeEach(() => {
  auth.capabilities = ['content.read', 'content.write', 'audit.read']
  current = collection()
  ruleTags = []; automatic = []; excluded = []
  members = [{ game_id: 1, name: '安全游戏', name_en: 'Safe Game', appid: 101, adult: false }, { game_id: 2, name: '成人游戏', name_en: 'Adult Game', appid: 102, adult: true }]
  home = { revision: 'original', slots: Array.from({ length: 5 }, (_, i) => ({ slot: i + 1, collection: i === 0 ? { ...collection(), status: 'published', home_slot: 1 } : null })) }
  vi.mocked(getJSON).mockImplementation(async path => {
    if (path.endsWith('/home-curation')) return home
    if (path.endsWith('/composition')) return composition()
    if (path.startsWith('/api/v1/game/games/')) return { game: { id: 3, name: '新增游戏', name_en: 'New Game', appid: 103 }, tags: [{ tag_id: 140, code: 'adult' }] }
    if (path.endsWith('/2')) return { ...collection(2), status: 'published' }
    return current
  })
  vi.mocked(listJSON).mockImplementation(async path => path.startsWith(collectionEndpoint)
    ? { list: [{ ...collection(), status: 'published' }, { ...collection(2), status: 'published' }], total: 51 }
    : path.endsWith('/tags') ? { list: [{ id: '10', label: '冒险' }], total: 1 } : { list: [{ id: '3', label: '新增游戏' }], total: 1 })
  vi.mocked(sendJSON).mockImplementation(async (path, _method, payload) => {
    if (path.endsWith('/home-curation')) {
      const data = payload as { slots: { slot: number; collection_id: number | null }[] }
      home = { revision: 'saved', slots: data.slots.map(s => ({ slot: s.slot, collection: s.collection_id ? collection(s.collection_id) : null })) }; return home
    }
    if (path === collectionEndpoint) { current = { ...current, ...payload as GameCollection, id: 42, version: 1 }; members = []; return current }
    if (path.endsWith('/composition')) {
      const input = payload as { tag_ids: number[]; manual_game_ids: number[]; excluded_game_ids: number[] }
      const all = [...members, ...automatic, ...excluded, newGame]
      ruleTags = input.tag_ids.map(id => [...ruleTags, newTag].find(tag => tag.tag_id === id)!)
      members = input.manual_game_ids.map(id => all.find(game => game.game_id === id)!)
      excluded = input.excluded_game_ids.map(id => all.find(game => game.game_id === id)!)
      current = { ...current, version: current.version + 1 }
      return composition()
    }
    if (path.endsWith('/publish')) current = { ...current, status: 'published' }
    else if (path.endsWith('/unpublish') || path.endsWith('/restore')) current = { ...current, status: 'draft' }
    else if (path.endsWith('/archive')) current = { ...current, status: 'archived' }
    else current = { ...current, ...payload as Partial<GameCollection> }
    current = { ...current, version: current.version + 1 }; return current
  })
})
afterEach(() => { cleanup(); vi.restoreAllMocks(); vi.resetAllMocks() })

function setup(path = '/game/collections/1') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 }, mutations: { retry: false } } })
  const router = createMemoryRouter([
    { path: '/game/collections', element: <CollectionListPage /> },
    { path: '/game/collections/new', element: <CreateCollectionPage /> },
    { path: '/game/collections/home-curation', element: <CollectionHomeCurationPage /> },
    { path: '/game/collections/:id', element: <CollectionWorkspacePage /> },
    { path: '/system/audit', element: <p>操作审计页</p> },
  ], { initialEntries: [path] })
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return { client, router }
}
async function ready() { await screen.findByRole('textbox', { name: '中文名称' }) }
const changeName = (name: string) => fireEvent.change(screen.getByRole('textbox', { name: '中文名称' }), { target: { value: name } })

it('keeps URL search composition-safe and resets pagination only at final commit', async () => {
  const { router } = setup('/game/collections?page_num=2&keyword=old&status=published')
  const input = await screen.findByRole('textbox', { name: '搜索游戏分区' })
  const status = screen.getByRole('combobox', { name: '分区状态' })
  expect(status).toHaveClass('w-40', 'shrink-0')
  expect(status).not.toHaveClass('w-full')
  expect(input).toHaveClass('flex-1', 'min-w-64')
  const toolbar = input.parentElement!.parentElement!
  expect(toolbar).toHaveClass('flex', 'flex-wrap')
  expect(toolbar).toContainElement(status)
  expect(toolbar).toContainElement(screen.getByRole('button', { name: '列' }))
  const header = screen.getByRole('heading', { name: '游戏分区' }).closest('header')!
  expect(header).toContainElement(screen.getByRole('link', { name: '首页入口编排' }))
  expect(header).toContainElement(screen.getByRole('link', { name: '新建游戏分区' }))
  fireEvent.compositionStart(input)
  fireEvent.change(input, { target: { value: 'zhong' } })
  fireEvent.change(input, { target: { value: '中' } })
  expect(router.state.location.search).toContain('page_num=2')
  expect(router.state.location.search).toContain('keyword=old')
  fireEvent.compositionEnd(input, { target: { value: '中文' } })
  fireEvent.change(input, { target: { value: '中文' } })
  await waitFor(() => expect(new URLSearchParams(router.state.location.search).get('keyword')).toBe('中文'))
  expect(router.state.location.search).toContain('page_num=1')
  expect(vi.mocked(listJSON).mock.calls.filter(([, , , q]) => q === '中文')).toHaveLength(1)
  expect(vi.mocked(listJSON).mock.calls.some(([, , , q]) => q === 'zhong' || q === '中')).toBe(false)
  const user = userEvent.setup(); await user.click(screen.getByRole('combobox', { name: '分区状态' })); await user.click(await screen.findByRole('option', { name: '草稿' }))
  expect(router.state.location.search).toContain('status=draft')
})

it('creates a Draft and navigates to its returned real ID', async () => {
  const { router } = setup('/game/collections/new')
  fireEvent.change(screen.getByRole('textbox', { name: /^Code/ }), { target: { value: 'new-collection' } })
  changeName('新的分区')
  fireEvent.change(screen.getByRole('textbox', { name: '英文名称' }), { target: { value: 'New Collection' } })
  fireEvent.click(screen.getByRole('button', { name: '创建草稿' }))
  await waitFor(() => expect(router.state.location.pathname).toBe('/game/collections/42'))
  expect(await screen.findByText('已保存', { exact: true })).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledWith(collectionEndpoint, 'POST', { code: 'new-collection', name: '新的分区', name_en: 'New Collection', info: '', info_en: '' })
})

it('keeps Code immutable and saves content with the displayed version', async () => {
  setup(); await ready()
  expect(screen.getByRole('textbox', { name: /Code/ })).toHaveAttribute('readonly')
  expect(screen.getByRole('link', { name: '操作审计' })).toHaveAttribute('href', '/system/audit?resource=gfg_game_collection')
  changeName('新名称'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', { version: 5, name: '新名称', name_en: 'Collection 1', info: '中文简介', info_en: 'Description' }))
  expect(await screen.findByText('已保存', { exact: true })).toBeInTheDocument()
  expect((vi.mocked(sendJSON).mock.calls[0][2] as Record<string, unknown>).code).toBeUndefined()
})

it('uses complete membership and advances shared baseVersion while retaining dirty content', async () => {
  setup(); await ready(); changeName('保留的内容草稿')
  fireEvent.click(screen.getByRole('button', { name: '移除固定 成人游戏' }))
  expect(screen.getByRole('button', { name: '发布' })).toBeDisabled()
  fireEvent.click(screen.getByRole('button', { name: '保存收录设置' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [], manual_game_ids: [1], excluded_game_ids: [] }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存内容' })).toBeEnabled())
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('保留的内容草稿')
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 6, name: '保留的内容草稿' })))
})

it('content save preserves membership draft and rebases its later complete-set save', async () => {
  setup(); await ready(); fireEvent.click(screen.getByRole('button', { name: '移除固定 成人游戏' })); changeName('保存内容先')
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存收录设置' })).toBeEnabled())
  expect(screen.queryByText('成人游戏')).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '保存收录设置' }))
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 6, tag_ids: [], manual_game_ids: [1], excluded_game_ids: [] }))
})

it('adds games with code-based Adult inspection without drag or order controls', async () => {
  setup(); await ready()
  fireEvent.focus(screen.getByPlaceholderText('搜索并固定游戏…'))
  fireEvent.click(await screen.findByRole('option', { name: '新增游戏' }))
  await screen.findByRole('button', { name: '移除固定 新增游戏' })
  expect(getJSON).toHaveBeenCalledWith('/api/v1/game/games/3/workspace')
  expect(screen.getAllByText('Adult')).toHaveLength(2)
  expect(screen.queryByRole('button', { name: /上移|下移|排序/ })).not.toBeInTheDocument()
  expect(document.querySelector('[draggable="true"]')).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: '保存收录设置' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [], manual_game_ids: [1, 2, 3], excluded_game_ids: [] }))
})

it('background version changes do not overwrite dirty drafts or silently rebase; 409 retains both', async () => {
  const { client } = setup(); await ready(); changeName('本地草稿'); fireEvent.click(screen.getByRole('button', { name: '移除固定 成人游戏' }))
  await act(async () => { client.setQueryData<CollectionWorkspace>(workspaceKey(1), { collection: { ...current, name: '他人修改', version: 9 }, composition: composition(9) }) })
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('此游戏分区已被其他操作修改，请重新加载后重试。', 409))
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  expect(await screen.findByText(/草稿已保留/)).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledTimes(1)
  expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 5 }))
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('本地草稿')
  expect(screen.queryByText('成人游戏')).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: '保存收录设置' })).toBeDisabled()
  vi.spyOn(window, 'confirm').mockReturnValue(true)
  current = { ...current, name: '服务器最新', version: 9 }
  fireEvent.click(screen.getByRole('button', { name: '放弃修改并重新加载' }))
  await waitFor(() => expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('服务器最新'))
  expect(screen.getByText('成人游戏')).toBeInTheDocument()
  expect(screen.queryByText(/草稿已保留/)).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: '重新加载' })).toBeEnabled()
  vi.mocked(sendJSON).mockResolvedValue({ ...current, version: 10 })
  changeName('重新编辑'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 9 })))
})

it('clean background refresh advances data and version normally', async () => {
  const { client } = setup(); await ready()
  await act(async () => { client.setQueryData(workspaceKey(1), { collection: { ...current, name: '刷新名称', version: 7 }, composition: composition(7) }) })
  await waitFor(() => expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('刷新名称'))
  changeName('更新'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 7 })))
})

it.each([['draft', '发布', 'publish'], ['draft', '归档', 'archive'], ['published', '取消发布', 'unpublish'], ['archived', '恢复为草稿', 'restore']] as const)('confirms %s lifecycle %s before a versioned write', async (status, label, action) => {
  current.status = status; setup(); await ready()
  fireEvent.click(screen.getByRole('button', { name: label }))
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(within(await screen.findByRole('dialog')).getByRole('button', { name: '确认操作' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/${action}`, 'POST', { version: 5 }))
})

it('read-only users inspect complete content and members without write or audit actions', async () => {
  auth.capabilities = ['content.read']; setup(); await ready()
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('分区 1')
  expect(screen.getByRole('textbox', { name: '中文名称' })).toBeDisabled()
  expect(screen.getByText('成人游戏')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /保存|发布|归档|恢复|移除/ })).not.toBeInTheDocument()
  expect(screen.queryByRole('link', { name: '操作审计' })).not.toBeInTheDocument()
})

it('archived workspace is read-only with Restore as the only mutation', async () => {
  current.status = 'archived'; setup(); await ready()
  expect(screen.getByRole('textbox', { name: '中文简介' })).toBeDisabled()
  expect(screen.getByText('安全游戏')).toBeInTheDocument()
  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /保存内容|保存收录设置|发布|归档|移除/ })).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: '恢复为草稿' })).toBeEnabled()
})

it('protects dirty content/member navigation and unload and disables lifecycle', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  const { router } = setup(); await ready(); fireEvent.click(screen.getByRole('button', { name: '移除固定 成人游戏' }))
  expect(screen.getByRole('button', { name: '发布' })).toBeDisabled()
  expect(screen.getByRole('button', { name: '归档' })).toBeDisabled()
  const event = new Event('beforeunload', { cancelable: true }); window.dispatchEvent(event); expect(event.defaultPrevented).toBe(true)
  fireEvent.click(screen.getByRole('link', { name: '返回' })); await waitFor(() => expect(confirm).toHaveBeenCalled())
  expect(router.state.location.pathname).toBe('/game/collections/1')
})

it('home has five editable slots, fixed sixth, eligible picker, and full placement payload', async () => {
  setup('/game/collections/home-curation'); await screen.findByRole('heading', { name: '#6 全部分区' })
  expect(screen.getAllByRole('combobox')).toHaveLength(5)
  expect(screen.queryByText(/缓存刷新|公开页面将在/)).not.toBeInTheDocument()
  expect(screen.getByText('固定入口 · 无需配置')).toBeInTheDocument()
  expect(listJSON).toHaveBeenCalledWith(eligibleCollectionEndpoint, 1, 50, '')
  fireEvent.focus(screen.getAllByRole('combobox')[1])
  expect(screen.queryByRole('option', { name: /分区 1/ })).not.toBeInTheDocument()
  fireEvent.click(await screen.findByRole('option', { name: /分区 2/ }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存编排' })).toBeEnabled())
  fireEvent.click(screen.getByRole('button', { name: '清空第 1 位' }))
  fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/home-curation`, 'PUT', { revision: 'original', slots: Array.from({ length: 5 }, (_, i) => ({ slot: i + 1, collection_id: i === 1 ? 2 : null })) }))
})

it('home dirty draft keeps original revision through background changes and 409', async () => {
  const { client } = setup('/game/collections/home-curation'); await screen.findByText('#6 全部分区')
  fireEvent.click(screen.getByRole('button', { name: '清空第 1 位' }))
  await act(async () => { client.setQueryData(homeKey, { ...home, revision: 'external' }) })
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('编排已修改', 409))
  fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
  expect(await screen.findByText(/编排草稿已保留/)).toBeInTheDocument()
  expect(sendJSON).toHaveBeenCalledTimes(1)
  expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/home-curation`, 'PUT', expect.objectContaining({ revision: 'original' }))
  expect(screen.getByRole('button', { name: '清空第 1 位' })).toBeDisabled()
  const event = new Event('beforeunload', { cancelable: true }); window.dispatchEvent(event); expect(event.defaultPrevented).toBe(true)
})

it('home is inspectable but not editable for content.read', async () => {
  auth.capabilities = ['content.read']; setup('/game/collections/home-curation'); await screen.findByText('#6 全部分区')
  expect(screen.getByRole('link', { name: '分区 1' })).toBeInTheDocument()
  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: '保存编排' })).not.toBeInTheDocument()
})

it('does not combine mismatched content and membership versions', async () => {
  vi.mocked(getJSON).mockImplementation(async path => path.endsWith('/composition') ? composition(9) : current)
  await expect(loadCollectionWorkspace(1)).rejects.toThrow('正在更新')
  expect(getJSON).toHaveBeenCalledTimes(4)
})


it('reloads clean workspace from header without confirmation', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  setup(); await ready()
  const reload = screen.getByRole('button', { name: '重新加载' })
  expect(reload.closest('header')).not.toBeNull()
  expect(screen.queryByText('公开刷新说明')).not.toBeInTheDocument()
  const calls = vi.mocked(getJSON).mock.calls.length
  current = { ...current, name: '服务器变更', version: 8 }
  fireEvent.click(reload)
  await waitFor(() => expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('服务器变更'))
  expect(getJSON).toHaveBeenCalledTimes(calls + 2)
  expect(confirm).not.toHaveBeenCalled()
  changeName('基线验证'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 8 })))
})

it('cancel discard retains both drafts; confirmed reload replaces them together', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  setup(); await ready(); changeName('内容草稿')
  fireEvent.click(screen.getByRole('button', { name: '移除固定 成人游戏' }))
  const calls = vi.mocked(getJSON).mock.calls.length
  const reload = screen.getByRole('button', { name: '放弃修改并重新加载' })
  expect(reload.closest('header')).not.toBeNull()
  fireEvent.click(reload)
  expect(confirm).toHaveBeenCalledExactlyOnceWith('放弃未保存的内容和成员修改，并重新加载服务器上的最新版本？')
  expect(getJSON).toHaveBeenCalledTimes(calls)
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('内容草稿')
  expect(screen.queryByText('成人游戏')).not.toBeInTheDocument()
  current = { ...current, name: '最新版本', version: 12 }
  confirm.mockReturnValue(true); fireEvent.click(reload)
  await waitFor(() => expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('最新版本'))
  expect(screen.getByText('成人游戏')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: '保存收录设置' })).toBeDisabled()
  expect(screen.getByRole('button', { name: '发布' })).toBeEnabled()
})

function largeMembership() {
  members = Array.from({ length: 45 }, (_, index) => ({ game_id: index + 100, name: `作品 ${index + 1}`, name_en: `Game ${index + 1}`, appid: 50000 + index, adult: false }))
  current.member_count = 45
}
const memberRows = () => within(screen.getByRole('list', { name: '已收录游戏' })).getAllByRole('listitem')
const nextMemberPage = () => fireEvent.click(screen.getByRole('button', { name: '成员下一页' }))

it('pages local membership 20/20/5 and searches names/AppID only after composition commits without requests', async () => {
  largeMembership(); setup(); await ready()
  expect(memberRows()).toHaveLength(20); nextMemberPage()
  expect(memberRows()).toHaveLength(20); nextMemberPage()
  expect(memberRows()).toHaveLength(5)
  expect(screen.getByText('显示 41–45，共 45 个')).toBeInTheDocument()
  const reads = vi.mocked(getJSON).mock.calls.length
  const searches = vi.mocked(listJSON).mock.calls.length
  const input = screen.getByRole('textbox', { name: '搜索已收录游戏' })
  fireEvent.compositionStart(input)
  fireEvent.change(input, { target: { value: 'zuopin' } })
  fireEvent.change(input, { target: { value: '作品 2' } })
  expect(memberRows()).toHaveLength(5)
  fireEvent.compositionEnd(input, { target: { value: '作品 23' } })
  expect(memberRows()).toHaveLength(1)
  expect(screen.getByText('1 / 1')).toBeInTheDocument()
  expect(screen.getByRole('link', { name: '作品 23' })).toBeInTheDocument()
  fireEvent.change(input, { target: { value: 'gAmE 41' } })
  expect(screen.getByRole('link', { name: '作品 41' })).toBeInTheDocument()
  fireEvent.change(input, { target: { value: '50044' } })
  expect(screen.getByRole('link', { name: '作品 45' })).toBeInTheDocument()
  fireEvent.change(input, { target: { value: 'no match' } })
  expect(screen.getByText('未找到匹配的已收录游戏')).toBeInTheDocument()
  expect(getJSON).toHaveBeenCalledTimes(reads)
  expect(listJSON).toHaveBeenCalledTimes(searches)
  expect(sendJSON).not.toHaveBeenCalled()
})

it('preserves full draft across page removals and add, then saves the canonical complete set', async () => {
  largeMembership(); const originalIDs = members.map(member => member.game_id)
  setup(); await ready()
  fireEvent.click(screen.getByRole('button', { name: '移除固定 作品 1' }))
  nextMemberPage(); nextMemberPage()
  fireEvent.click(screen.getByRole('button', { name: '移除固定 作品 45' }))
  fireEvent.focus(screen.getByPlaceholderText('搜索并固定游戏…'))
  fireEvent.click(await screen.findByRole('option', { name: '新增游戏' }))
  await screen.findByRole('button', { name: '移除固定 新增游戏' })
  expect(screen.getByText('显示 41–44，共 44 个')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '保存收录设置' }))
  const expected = [3, ...originalIDs.filter(id => id !== 100 && id !== 144)]
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [], manual_game_ids: expected, excluded_game_ids: [] }))
  expect(await screen.findByText('已保存', { exact: true })).toBeInTheDocument()
})

it('clamps a removed last page to the last valid page', async () => {
  largeMembership(); setup(); await ready(); nextMemberPage(); nextMemberPage()
  for (let index = 41; index <= 45; index++) fireEvent.click(screen.getByRole('button', { name: `移除固定 作品 ${index}` }))
  expect(memberRows()).toHaveLength(20)
  expect(screen.getByText('显示 21–40，共 40 个')).toBeInTheDocument()
  expect(screen.getByText('2 / 2')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: '成员下一页' })).toBeDisabled()
})

it('uses only a short success toast when a member save clears the Home slot', async () => {
  current = { ...current, home_slot: 1 }
  setup(); await ready()
  vi.mocked(sendJSON).mockResolvedValue({ ...composition(6), home_slot: null })
  fireEvent.click(screen.getByRole('button', { name: '移除固定 安全游戏' }))
  fireEvent.click(screen.getByRole('button', { name: '保存收录设置' }))
  expect(await screen.findByText('已保存', { exact: true })).toBeInTheDocument()
  expect(screen.queryByText(/缓存刷新|公开页面将在|分钟|小时|撤出首页/)).not.toBeInTheDocument()
})

it('Home curation success toast only confirms saving', async () => {
  setup('/game/collections/home-curation'); await screen.findByText('#6 全部分区')
  fireEvent.click(screen.getByRole('button', { name: '清空第 1 位' }))
  fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
  expect(await screen.findByText('已保存', { exact: true })).toBeInTheDocument()
  expect(screen.queryByText(/缓存刷新|公开页面将在|分钟|小时/)).not.toBeInTheDocument()
})


it('loads Composition for legacy manual collections without calling the compatibility members API', async () => {
  setup(); await ready()
  expect(screen.getByText('未绑定标签，仅使用人工固定成员。')).toBeInTheDocument()
  expect(within(screen.getByRole('list', { name: '已收录游戏' })).getAllByText('人工')).toHaveLength(2)
  expect(getJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`)
  expect(vi.mocked(getJSON).mock.calls.some(([url]) => url.endsWith('/members'))).toBe(false)
})

function hybrid() { ruleTags = [newTag]; automatic = [members[0], newGame] }
const rowFor = (name: string) => screen.getByRole('link', { name }).closest('li')!
const saveComposition = () => fireEvent.click(screen.getByRole('button', { name: '保存收录设置' }))

it('shows deduped sources and keeps automatic matches after removing a pin; overrides stay disjoint', async () => {
  hybrid(); setup(); await ready()
  expect(within(rowFor('安全游戏')).getByText('自动 + 人工')).toBeInTheDocument()
  expect(within(rowFor('成人游戏')).getByText('人工')).toBeInTheDocument()
  expect(within(rowFor('新增游戏')).getByText('自动')).toBeInTheDocument()
  expect(memberRows()).toHaveLength(3)
  expect(screen.getByText('有效收录').nextElementSibling).toHaveTextContent('3')
  expect(screen.getByText('SFW 可见').nextElementSibling).toHaveTextContent('1')
  fireEvent.click(screen.getByRole('button', { name: '移除固定 安全游戏' }))
  expect(within(rowFor('安全游戏')).getByText('自动')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '排除 成人游戏' }))
  expect(within(screen.getByRole('list', { name: '已排除游戏' })).getByText('成人游戏')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '固定保留 新增游戏' }))
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [10], manual_game_ids: [3], excluded_game_ids: [2] }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存收录设置' })).toBeDisabled())
  expect(within(rowFor('新增游戏')).getByText('自动 + 人工')).toBeInTheDocument()
})

it('restores an excluded automatic member only after Backend confirms it, without materializing a manual pin', async () => {
  hybrid(); excluded = [newGame]; setup(); await ready()
  fireEvent.click(screen.getByRole('button', { name: '取消排除 新增游戏' }))
  expect(screen.queryByRole('link', { name: '新增游戏' })).not.toBeInTheDocument()
  expect(screen.getByText(/数量为已保存结果/)).toBeInTheDocument()
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [10], manual_game_ids: [1, 2], excluded_game_ids: [] }))
  expect(await screen.findByRole('link', { name: '新增游戏' })).toBeInTheDocument()
  expect(within(rowFor('新增游戏')).getByText('自动')).toBeInTheDocument()
})

it('pinning an excluded game atomically removes its exclusion', async () => {
  hybrid(); excluded = [newGame]; setup(); await ready()
  fireEvent.click(screen.getByRole('button', { name: '固定保留 新增游戏' }))
  expect(screen.queryByRole('list', { name: '已排除游戏' })).not.toBeInTheDocument()
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [10], manual_game_ids: [1, 2, 3], excluded_game_ids: [] }))
})

it('retains paused Tag rules while saving overrides and allows explicitly detaching them', async () => {
  ruleTags = [{ ...newTag, tag_id: 9, name: '旧标签', active: false }]
  setup(); await ready()
  expect(screen.getByText('失效')).toBeInTheDocument()
  const picker = screen.getByPlaceholderText('搜索并绑定标签…')
  fireEvent.focus(picker)
  expect(screen.queryByRole('option', { name: '旧标签' })).not.toBeInTheDocument()
  fireEvent.click(await screen.findByRole('option', { name: '冒险' }))
  fireEvent.click(screen.getByRole('button', { name: '移除固定 成人游戏' }))
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [9, 10], manual_game_ids: [1], excluded_game_ids: [] }))
  await waitFor(() => expect(screen.getByRole('button', { name: '解绑 旧标签' })).toBeEnabled())
  expect(screen.getByText('失效')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '解绑 旧标签' }))
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 6, tag_ids: [10], manual_game_ids: [1], excluded_game_ids: [] }))
})

it('does not guess matches for a newly selected rule before saving', async () => {
  automatic = [newGame]; setup(); await ready()
  fireEvent.focus(screen.getByPlaceholderText('搜索并绑定标签…'))
  fireEvent.click(await screen.findByRole('option', { name: '冒险' }))
  expect(screen.queryByRole('link', { name: '新增游戏' })).not.toBeInTheDocument()
  expect(screen.getByText(/收录设置尚未保存/)).toBeInTheDocument()
  saveComposition()
  expect(await screen.findByRole('link', { name: '新增游戏' })).toBeInTheDocument()
  expect(within(rowFor('新增游戏')).getByText('自动')).toBeInTheDocument()
})

it('keeps remote Tag and Game searches independent and IME-safe through their debounce', async () => {
  setup(); await ready()
  await waitFor(() => expect(listJSON).toHaveBeenCalledWith('/api/v1/options/tags', 1, 20, ''))
  for (const [placeholder, endpoint, word] of [['搜索并绑定标签…', '/api/v1/options/tags', '冒险'], ['搜索并固定游戏…', '/api/v1/options/games', '游戏']] as const) {
    const input = screen.getByPlaceholderText(placeholder)
    fireEvent.focus(input)
    vi.mocked(listJSON).mockClear()
    vi.useFakeTimers()
    try {
      fireEvent.compositionStart(input)
      fireEvent.change(input, { target: { value: 'pinyin' } })
      fireEvent.keyDown(input, { key: 'Enter', isComposing: true })
      await act(async () => { await vi.advanceTimersByTimeAsync(350) })
      expect(listJSON).not.toHaveBeenCalled()
      fireEvent.compositionEnd(input, { target: { value: word } })
      fireEvent.change(input, { target: { value: word } })
      await act(async () => { await vi.advanceTimersByTimeAsync(300) })
      expect(listJSON).toHaveBeenCalledExactlyOnceWith(endpoint, 1, 20, word)
    } finally { vi.useRealTimers() }
  }
  expect(sendJSON).not.toHaveBeenCalled()
  expect(screen.getByRole('button', { name: '保存收录设置' })).toBeDisabled()
})

it('preserves Tag and exclusion drafts with content across background refresh, shared save version, and 409', async () => {
  hybrid(); const { client } = setup(); await ready()
  fireEvent.click(screen.getByRole('button', { name: '解绑 冒险' }))
  fireEvent.click(screen.getByRole('button', { name: '排除 安全游戏' }))
  changeName('本地内容')
  await act(async () => { client.setQueryData(workspaceKey(1), { collection: { ...current, version: 9 }, composition: composition(9) }) })
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1`, 'PUT', expect.objectContaining({ version: 5 })))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存收录设置' })).toBeEnabled())
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('版本冲突', 409))
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenLastCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 6, tag_ids: [], manual_game_ids: [2], excluded_game_ids: [1] }))
  expect(await screen.findByText(/草稿已保留/)).toBeInTheDocument()
  expect(screen.getByRole('textbox', { name: '中文名称' })).toHaveValue('本地内容')
  expect(screen.queryByRole('button', { name: '解绑 冒险' })).not.toBeInTheDocument()
  expect(screen.getByRole('button', { name: '取消排除 安全游戏' })).toBeInTheDocument()
  expect(screen.getByRole('button', { name: '保存收录设置' })).toBeDisabled()
})

it('shows Backend publish validation without losing composition and diagnoses passive member loss', async () => {
  members = []; ruleTags = [{ ...newTag, active: false }]; setup(); await ready()
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('已发布游戏分区须至少有效收录两个游戏', 400))
  fireEvent.click(screen.getByRole('button', { name: '发布' }))
  fireEvent.click(within(await screen.findByRole('dialog')).getByRole('button', { name: '确认操作' }))
  expect(await screen.findByText(/须至少有效收录两个/)).toBeInTheDocument()
  expect(screen.getByText('失效')).toBeInTheDocument()
  expect(screen.getByRole('button', { name: '发布' })).toBeEnabled()
})

it('published collections with passive member loss still allow content editing', async () => {
  members = []; current.status = 'published'; setup(); await ready()
  expect(screen.getByText(/当前有效成员不足两部/)).toBeInTheDocument()
  changeName('仍可编辑'); fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  expect(await screen.findByText('已保存', { exact: true })).toBeInTheDocument()
})

it.each(['reader', 'archived'])('keeps hybrid inspection available to %s without composition write controls', async mode => {
  hybrid(); excluded = [newGame]; ruleTags.push({ ...newTag, tag_id: 9, name: '已归档规则', active: false })
  if (mode === 'reader') auth.capabilities = ['content.read']
  else current.status = 'archived'
  setup(); await ready()
  expect(screen.getByText('失效')).toBeInTheDocument()
  expect(screen.getByText('自动 + 人工')).toBeInTheDocument()
  expect(screen.getByRole('list', { name: '已排除游戏' })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /^(固定保留 |移除固定 |排除 |取消排除 |解绑 |保存收录)/ })).not.toBeInTheDocument()
  expect(screen.queryByRole('combobox')).not.toBeInTheDocument()
})

it('retains an existing ineligible Home slot while saving another placement', async () => {
  home.slots[0].collection!.sfw_member_count = 0
  setup('/game/collections/home-curation'); await screen.findByText('#6 全部分区')
  expect(screen.getByText(/已配置 · 当前不符合展示条件/)).toBeInTheDocument()
  expect(screen.getByRole('link', { name: '分区 1' })).toBeInTheDocument()
  fireEvent.focus(screen.getAllByRole('combobox')[1])
  fireEvent.click(await screen.findByRole('option', { name: /分区 2/ }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存编排' })).toBeEnabled())
  fireEvent.click(screen.getByRole('button', { name: '保存编排' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/home-curation`, 'PUT', { revision: 'original', slots: [
    { slot: 1, collection_id: 1 }, { slot: 2, collection_id: 2 }, { slot: 3, collection_id: null }, { slot: 4, collection_id: null }, { slot: 5, collection_id: null },
  ] }))
})

it('warns for 100+ automatic members without truncating or pinning them', async () => {
  ruleTags = [newTag]; members = []
  automatic = Array.from({ length: 105 }, (_, i) => ({ game_id: 100 + i, name: `自动作品 ${i + 1}`, name_en: `Auto ${i + 1}`, appid: i, adult: false }))
  setup(); await ready()
  expect(screen.getByText(/不会截断成员/)).toBeInTheDocument()
  expect(memberRows()).toHaveLength(20)
  for (let i = 0; i < 5; i++) nextMemberPage()
  expect(memberRows()).toHaveLength(5)
  fireEvent.click(screen.getByRole('button', { name: '固定保留 自动作品 105' }))
  saveComposition()
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${collectionEndpoint}/1/composition`, 'PUT', { version: 5, tag_ids: [10], manual_game_ids: [204], excluded_game_ids: [] }))
  await waitFor(() => expect(screen.getByRole('button', { name: '保存收录设置' })).toBeDisabled())
  expect(screen.getByText('显示 101–105，共 105 个')).toBeInTheDocument()
})
