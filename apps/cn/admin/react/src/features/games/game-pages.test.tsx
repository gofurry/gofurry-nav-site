import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { listJSON, sendJSON } from '../../lib/api'
import type { Game } from '../../lib/types'
import { GameClassificationForm, GameContentForm } from './game-pages'

vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), listJSON: vi.fn(), sendJSON: vi.fn() }))
afterEach(() => { cleanup(); vi.resetAllMocks() })
const game: Game = { id: 1, name: '游戏', name_en: 'Game', info: '', info_en: '', create_time: '', update_time: '', resources: [], groups: [], developers: [], publishers: [], appid: 82, header: '', links: [], weight: 1, primary_tag: 1, secondary_tag: 0 }

function setup(element: React.ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  const router = createMemoryRouter([{ path: '/', element }])
  render(<QueryClientProvider client={client}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
}

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
  vi.mocked(sendJSON).mockResolvedValue(game)
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
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/game/tag-maps/bulk-replace', 'PUT', { owner_id: 1, ids: [1, 2] }))
})
