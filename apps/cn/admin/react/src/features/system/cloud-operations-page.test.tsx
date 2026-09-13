import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { getJSON, sendJSON } from '../../lib/api'
import { CloudOperationsContent } from './cloud-operations-page'

const auth = vi.hoisted(() => ({ capabilities: ['cloudops.read', 'cloudops.manage'] }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => auth.capabilities.includes(capability) }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), sendJSON: vi.fn() }))
const store = { provider: 'cos', bucket: 'dev', region: 'test', public_base_url: 'https://assets.example.com', configured: true, reachable: true, probe_exists: true }
const overview = { primary: store, mirror: { ...store, provider: 'r2' }, edgeone: { configured: true, main_host: 'example.com', asset_host: 'assets.example.com' }, cloudflare: { configured: true, asset_host: 'mirror.example.com' } }
beforeEach(() => { vi.clearAllMocks(); auth.capabilities = ['cloudops.read', 'cloudops.manage']; vi.mocked(getJSON).mockResolvedValue({ total: 0, list: [] }); vi.mocked(sendJSON).mockResolvedValue({ result: { job_id: 'job', status: 'submitted' } }); vi.spyOn(window, 'confirm').mockReturnValue(true) })
afterEach(() => { cleanup(); vi.restoreAllMocks() })
function setup() { render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><CloudOperationsContent overview={overview} refresh={vi.fn()} /></QueryClientProvider>) }

it('clears the main host using ordinary host purge and hides full-zone purge from a developer', async () => {
  setup()
  expect(screen.queryByRole('button', { name: '清除整个 EdgeOne Zone 缓存' })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '清除主站缓存 · example.com' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/system/cloud/edgeone/purge', 'POST', { type: 'host', targets: ['example.com'] }))
})

it('uses a separate endpoint and explicit scope confirmation for an owner zone purge', async () => {
  auth.capabilities.push('cloudops.purge_all'); setup()
  fireEvent.click(screen.getByRole('button', { name: '清除整个 EdgeOne Zone 缓存' }))
  expect(window.confirm).toHaveBeenCalledWith(expect.stringContaining('包括所有加速域名'))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/system/cloud/edgeone/purge-all', 'POST', undefined))
})

it('shows storage facts and object inspection without mutation controls for read-only access', () => {
  auth.capabilities = ['cloudops.read']; setup()
  expect(screen.getByRole('button', { name: '检查' })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: '提交清缓存' })).not.toBeInTheDocument()
  expect(screen.queryByText(/同步率/)).not.toBeInTheDocument()
  expect(screen.queryByText(/purge everything/i)).not.toBeInTheDocument()
})
