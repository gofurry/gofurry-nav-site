import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ToastProvider } from '../../app/toast'
import { sendForm } from '../../lib/api'
import { AssetEditor } from './asset-pages'
import { PatternPreview } from './pattern-preview'

const auth = vi.hoisted(() => ({ write: true }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: () => auth.write }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), sendForm: vi.fn() }))
beforeEach(() => { auth.write = true; vi.clearAllMocks(); URL.createObjectURL = vi.fn(() => 'blob:pattern-preview'); URL.revokeObjectURL = vi.fn() })
afterEach(cleanup)

function editor(kind: 'hero' | 'pattern' = 'pattern') {
  const saved = vi.fn()
  const router = createMemoryRouter([{ path: '/', element: <AssetEditor kind={kind} variant="mobile" asset={null} close={vi.fn()} saved={saved} /> }])
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
