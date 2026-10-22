import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { sendForm, sendJSON } from '../../lib/api'
import { AssetEditor, type ManagedAsset } from './asset-pages'
import { SiteIconEditor } from './site-icon-editor'
import type { Site } from '../../lib/types'

const auth = vi.hoisted(() => ({ write: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: () => auth.write }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), sendForm: vi.fn(), sendJSON: vi.fn() }))
beforeEach(() => { auth.write = true; vi.clearAllMocks(); URL.createObjectURL = vi.fn(() => 'blob:hero-preview'); URL.revokeObjectURL = vi.fn() })
afterEach(cleanup)

function editor(asset: ManagedAsset | null = null) {
  const saved = vi.fn()
  const router = createMemoryRouter([{ path: '/', element: <AssetEditor variant="mobile" asset={asset} close={vi.fn()} saved={saved} /> }])
  render(<QueryClientProvider client={new QueryClient()}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return saved
}

it('previews Hero locally, publishes only Hero metadata and retains mirror warnings', async () => {
  const saved = editor()
  fireEvent.change(screen.getByLabelText(/^名称/), { target: { value: 'Mobile Hero' } })
  const file = new File(['original AVIF bytes'], 'hero.avif', { type: 'image/avif' })
  fireEvent.change(screen.getByLabelText(/选择文件/), { target: { files: [file] } })
  expect(await screen.findByAltText('Hero 预览')).toHaveAttribute('src', 'blob:hero-preview')
  expect(sendForm).not.toHaveBeenCalled()
  vi.mocked(sendForm).mockResolvedValue({ primary: 'ready', mirror: 'failed', warnings: ['R2 mirror sync failed'] })
  fireEvent.click(screen.getByRole('button', { name: '创建资源' }))
  await waitFor(() => expect(sendForm).toHaveBeenCalledOnce())
  const [path, body] = vi.mocked(sendForm).mock.calls[0]
  expect(path).toBe('/api/v1/nav/hero-assets')
  expect(Object.fromEntries(body)).toEqual({ name: 'Mobile Hero', enabled: 'true', variant: 'mobile', file })
  expect(saved).toHaveBeenCalledOnce()
  expect(await screen.findByText(/已发布到 COS；R2 mirror sync failed/)).toBeInTheDocument()
})

it('does not expose publish controls without content.write', () => {
  auth.write = false; editor()
  expect(screen.queryByRole('button', { name: '创建资源' })).not.toBeInTheDocument()
  expect(screen.getByLabelText(/选择文件/)).toBeDisabled()
})

const existingAsset: ManagedAsset = { id: '42', name: '测试资源', enabled: true, object_key: 'nav/hero/mobile/test.avif', primary_url: 'https://assets.example.com/test.avif', mirror_url: 'https://mirror.example.com/test.avif' }

it('confirms Hero replacement and deletion before sending any mutation', async () => {
  editor(existingAsset)
  const file = new File(['original bytes'], 'next.avif')
  fireEvent.change(screen.getByLabelText('选择替换文件'), { target: { files: [file] } })
  fireEvent.click(screen.getByRole('button', { name: '替换文件' }))
  let dialog = await screen.findByRole('dialog')
  expect(within(dialog).getByText(/next\./)).toBeInTheDocument()
  expect(sendForm).not.toHaveBeenCalled()
  fireEvent.click(within(dialog).getByRole('button', { name: '取消' }))
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  expect(sendForm).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: '替换文件' }))
  dialog = await screen.findByRole('dialog')
  vi.mocked(sendForm).mockResolvedValue({ primary: 'ready', mirror: 'ready' })
  fireEvent.click(within(dialog).getByRole('button', { name: '确认替换' }))
  await waitFor(() => expect(sendForm).toHaveBeenCalledOnce())
  expect(vi.mocked(sendForm).mock.calls[0][0]).toBe('/api/v1/nav/hero-assets/42/file')
  expect(vi.mocked(sendForm).mock.calls[0][1].get('file')).toBe(file)
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  fireEvent.click(screen.getByRole('button', { name: '删除' }))
  dialog = await screen.findByRole('dialog')
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(within(dialog).getByRole('button', { name: '取消' }))
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: '删除' }))
  dialog = await screen.findByRole('dialog')
  vi.mocked(sendJSON).mockResolvedValue({})
  fireEvent.click(within(dialog).getByRole('button', { name: '确认删除' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/nav/hero-assets/42', 'DELETE'))
})

it('uses the compact file picker and confirms clearing a site icon', async () => {
  const site = { id: 106, name: '测试网站', icon: 'nav/sites/106/icon/test.svg', icon_primary_url: 'https://assets.example.com/icon.svg' } as Site
  render(<QueryClientProvider client={new QueryClient()}><ToastProvider><SiteIconEditor site={site} onDirtyChange={vi.fn()} /></ToastProvider></QueryClientProvider>)
  expect(screen.getByText('最多 2 MiB')).toBeInTheDocument()
  expect(screen.getByText('上传网站图标，保留原始格式，清除后使用默认图标。')).toBeInTheDocument()
  expect(screen.getByLabelText('选择图标')).toHaveClass('hidden')
  expect(screen.getByRole('button', { name: '浏览文件' }).querySelector('svg')).not.toBeNull()
  fireEvent.click(screen.getByRole('button', { name: '清除图标' }))
  let dialog = await screen.findByRole('dialog')
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(within(dialog).getByRole('button', { name: '取消' }))
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  fireEvent.click(screen.getByRole('button', { name: '清除图标' }))
  dialog = await screen.findByRole('dialog')
  vi.mocked(sendJSON).mockRejectedValueOnce(new Error('清除失败，请重试'))
  fireEvent.click(within(dialog).getByRole('button', { name: '确认清除' }))
  expect(await screen.findByText('清除失败，请重试')).toBeVisible()
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  fireEvent.click(screen.getByRole('button', { name: '清除图标' }))
  dialog = await screen.findByRole('dialog')
  vi.mocked(sendJSON).mockResolvedValue({})
  fireEvent.click(within(dialog).getByRole('button', { name: '确认清除' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/nav/sites/106/icon', 'DELETE'))
})
