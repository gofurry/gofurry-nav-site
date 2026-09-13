import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { sendForm, sendJSON } from '../../lib/api'
import { AssetEditor, type ManagedAsset } from './asset-pages'
import { SiteIconEditor } from './site-icon-editor'
import type { Site } from '../../lib/types'
import { PatternPreview } from './pattern-preview'

const auth = vi.hoisted(() => ({ write: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: () => auth.write }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), sendForm: vi.fn(), sendJSON: vi.fn() }))
beforeEach(() => { auth.write = true; vi.clearAllMocks(); URL.createObjectURL = vi.fn(() => 'blob:pattern-preview'); URL.revokeObjectURL = vi.fn() })
afterEach(cleanup)

function editor(kind: 'hero' | 'pattern' = 'pattern', asset: ManagedAsset | null = null) {
  const saved = vi.fn()
  const router = createMemoryRouter([{ path: '/', element: <AssetEditor kind={kind} variant="mobile" asset={asset} close={vi.fn()} saved={saved} /> }])
  render(<QueryClientProvider client={new QueryClient()}><ToastProvider><RouterProvider router={router} /></ToastProvider></QueryClientProvider>)
  return saved
}

it('uses an actual repeating mask for both themes with the edited settings', () => {
  const { container, rerender } = render(<PatternPreview url="blob:pattern-preview" appearance={{ light_color: '#123456', dark_color: '#abcdef', light_opacity: 0.1, dark_opacity: 0.4, default_size_px: 96 }} />)
  const light = container.querySelector('[data-pattern-mask="light"]') as HTMLElement
  expect(light.style.maskRepeat).toBe('repeat')
  expect(light.style.maskImage).toContain('blob:pattern-preview')
  expect(light.style.opacity).toBe('0.1')
  rerender(<PatternPreview url="blob:pattern-preview" appearance={{ light_color: '#111111', dark_color: '#ffffff', light_opacity: 0.2, dark_opacity: 0.4, default_size_px: 128 }} />)
  expect(light.style.maskSize).toBe('128px')
  expect(light.style.opacity).toBe('0.2')
})

it('previews a selected pattern without uploading, then submits metadata and retains mirror warnings', async () => {
  const saved = editor()
  fireEvent.change(screen.getByLabelText(/英文名称/), { target: { value: 'Pattern' } })
  fireEvent.change(screen.getByLabelText(/^名称/), { target: { value: '图案' } })
  const file = new File(['<?xml version="1.0"?><!DOCTYPE svg><!-- exported drawing --><svg xmlns="http://www.w3.org/2000/svg"><style>.shape{fill:black}</style><path class="shape" d="M0 0h10v10z"/></svg>'], 'pattern.svg', { type: 'image/svg+xml' })
  fireEvent.change(screen.getByLabelText(/选择文件/), { target: { files: [file] } })
  await waitFor(() => expect(screen.getByLabelText('背景亮暗预览')).toBeInTheDocument())
  expect(sendForm).not.toHaveBeenCalled()
  expect(URL.createObjectURL).toHaveBeenCalledWith(file)
  vi.mocked(sendForm).mockResolvedValue({ primary: 'ready', mirror: 'failed', warnings: ['R2 mirror sync failed'] })
  fireEvent.click(screen.getByRole('button', { name: '创建资源' }))
  await waitFor(() => expect(sendForm).toHaveBeenCalledOnce())
  const [path, body] = vi.mocked(sendForm).mock.calls[0]
  expect(path).toBe('/api/v1/nav/background-patterns')
  expect(body.get('file')).toBe(file)
  expect(body.get('light_color')).toBe('#9c846a')
  expect(saved).toHaveBeenCalledOnce()
  expect(await screen.findByText(/已发布到 COS；R2 mirror sync failed/)).toBeInTheDocument()
})

it('does not expose publish controls without content.write', () => {
  auth.write = false; editor('hero')
  expect(screen.queryByRole('button', { name: '创建资源' })).not.toBeInTheDocument()
  expect(screen.getByLabelText(/选择文件/)).toBeDisabled()
})

const existingAsset: ManagedAsset = { id: '42', name: '测试资源', name_en: 'Test asset', enabled: true, sort_order: 0, light_color: '#123456', dark_color: '#abcdef', light_opacity: .1, dark_opacity: .2, default_size_px: 96, object_key: 'nav/patterns/test.svg', primary_url: 'https://assets.example.com/test.svg', mirror_url: 'https://mirror.example.com/test.svg' }

it.each(['hero', 'pattern'] as const)('confirms %s replacement and deletion before sending any mutation', async (kind) => {
  editor(kind, existingAsset)
  const file = new File(['original bytes'], kind === 'hero' ? 'next.avif' : 'next.svg')
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
  expect(vi.mocked(sendForm).mock.calls[0][0]).toBe(`/api/v1/nav/${kind === 'hero' ? 'hero-assets' : 'background-patterns'}/42/file`)
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
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/nav/${kind === 'hero' ? 'hero-assets' : 'background-patterns'}/42`, 'DELETE'))
})

it('keeps text-based numeric editing and blocks invalid pattern metadata', async () => {
  editor('pattern', existingAsset)
  expect(document.querySelector('input[type="number"]')).toBeNull()
  const opacity = screen.getByLabelText('浅色透明度')
  fireEvent.change(opacity, { target: { value: '0.' } })
  expect(opacity).toHaveValue('0.')
  fireEvent.change(opacity, { target: { value: '0.065' } })
  fireEvent.change(screen.getByLabelText('平铺尺寸（px）'), { target: { value: '0' } })
  fireEvent.click(screen.getByRole('button', { name: '保存信息' }))
  expect(await screen.findByText(/透明度须为/)).toBeInTheDocument()
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.change(screen.getByLabelText('平铺尺寸（px）'), { target: { value: '160' } })
  vi.mocked(sendJSON).mockResolvedValue({})
  fireEvent.click(screen.getByRole('button', { name: '保存信息' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/nav/background-patterns/42', 'PUT', expect.objectContaining({ light_opacity: .065, default_size_px: 160 })))
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
