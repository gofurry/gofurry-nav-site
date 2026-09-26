import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { getJSON, listJSON, sendJSON } from '../../lib/api'
import type { Game } from '../../lib/types'
import { GameClassificationForm, GameContentForm } from './game-pages'

vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: () => true }) }))

afterEach(() => { cleanup(); vi.resetAllMocks() })
const game: Game = { id: 1, name: '游戏', name_en: 'Game', info: '', info_en: '', create_time: '', update_time: '', resources: [], groups: [], developers: [], publishers: [], appid: 82, header: '', links: [], weight: 1, primary_tag: 1, secondary_tag: 0 }

function setup(element: React.ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  const router = createMemoryRouter([{ path: '/', element }])
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
}

it('shows partial Steam warnings, preserves manual fields, and clears warnings after a complete retry', async () => {
  const warning = '中文详情未完整获取，请检查名称和简介，重试或手动填写。'
  vi.mocked(getJSON)
    .mockResolvedValueOnce({ name: '', info: '', header: 'https://example.test/header.jpg', warnings: [warning] })
    .mockResolvedValueOnce({ name: 'Steam 中文名称', info: 'Steam 中文简介', name_en: 'Steam name', info_en: 'Steam description', header: 'https://example.test/header.jpg' })
  vi.mocked(sendJSON).mockResolvedValue(game)
  setup(<GameContentForm game={{ ...game, info: '手工简介' }} />)
  fireEvent.click(screen.getByRole('button', { name: '从 Steam 预填' }))
  expect(await screen.findByText(warning)).toBeInTheDocument()
  expect(screen.getByText('已加载部分 Steam 内容，请检查提示')).toBeInTheDocument()
  expect(screen.queryByText('已加载 Steam 预填内容')).not.toBeInTheDocument()
  expect(screen.getByRole('textbox', { name: /^中文名称/ })).toHaveValue('游戏')
  expect(screen.getByRole('textbox', { name: '中文简介' })).toHaveValue('手工简介')
  expect(screen.getByRole('textbox', { name: '封面 / Header URL' })).toHaveValue('https://example.test/header.jpg')
  expect(getJSON).toHaveBeenCalledWith('/api/v1/game/games/steam-prefill?appid=82')

  fireEvent.click(screen.getByRole('button', { name: '从 Steam 预填' }))
  await waitFor(() => expect(screen.getByRole('textbox', { name: /^中文名称/ })).toHaveValue('Steam 中文名称'))
  expect(screen.queryByText(warning)).not.toBeInTheDocument()
  expect(screen.getByText('已加载 Steam 预填内容')).toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalled())
  expect(vi.mocked(sendJSON).mock.calls[0][2]).not.toHaveProperty('warnings')
})

it('preserves the content form and reports a failed Steam lookup', async () => {
  vi.mocked(getJSON).mockRejectedValue(new Error('Steam 暂时不可用'))
  setup(<GameContentForm game={game} />)
  fireEvent.click(screen.getByRole('button', { name: '从 Steam 预填' }))
  expect(await screen.findByText('Steam 暂时不可用')).toBeInTheDocument()
  expect(screen.getByRole('textbox', { name: /^中文名称/ })).toHaveValue('游戏')
  expect(screen.queryByText('已加载 Steam 预填内容')).not.toBeInTheDocument()
  expect(sendJSON).not.toHaveBeenCalled()
})

it.each(['中文简介', '英文简介'])('blocks overlong %s on the actual content form, then saves 400 characters', async (label) => {
  vi.mocked(sendJSON).mockResolvedValue(game)
  setup(<GameContentForm game={game} />)
  fireEvent.change(screen.getByRole('textbox', { name: label }), { target: { value: '🐺'.repeat(401) } })
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(screen.getByText('简介最多 400 个字符')).toBeInTheDocument())
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.change(screen.getByRole('textbox', { name: label }), { target: { value: '🐺'.repeat(400) } })
  fireEvent.click(screen.getByRole('button', { name: '保存内容' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/games/1', 'PUT', expect.objectContaining({ [label === '中文简介' ? 'info' : 'info_en']: '🐺'.repeat(400) })))
})

it('wires both remote selectors and saves all selected tags after local filtering', async () => {
  vi.mocked(listJSON).mockResolvedValue({ total: 2, list: [{ id: '1', label: '冒险' }, { id: '2', label: '解谜' }] })
  vi.mocked(sendJSON).mockResolvedValue({ game, tags: [{ tag_id: 1, tag_name: '冒险' }, { tag_id: 2, tag_name: '解谜' }] })
  setup(<GameClassificationForm workspace={{ game, tags: [{ tag_id: 1, tag_name: '冒险' }] } as Parameters<typeof GameClassificationForm>[0]['workspace']} />)
  await screen.findByRole('checkbox', { name: '冒险' })
  expect(screen.getByPlaceholderText('搜索主要标签…')).toHaveValue('冒险')
  expect(screen.getByPlaceholderText('搜索次要标签…')).toHaveValue('')
  expect(listJSON).toHaveBeenCalledWith('/api/v1/options/tags', 1, 10, '')
  const requests = vi.mocked(listJSON).mock.calls.length
  fireEvent.change(screen.getByLabelText('搜索全部标签'), { target: { value: '解谜' } })
  fireEvent.click(screen.getByRole('checkbox', { name: '解谜' }))
  expect(listJSON).toHaveBeenCalledTimes(requests)
  fireEvent.click(screen.getByRole('button', { name: '保存分类与展示' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/games/1/classification', 'PUT', { weight: 1, primary_tag_id: 1, secondary_tag_id: null, tag_ids: [1, 2] }))
})
