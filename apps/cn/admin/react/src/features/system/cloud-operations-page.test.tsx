import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { getJSON, sendJSON } from '../../lib/api'
import { CloudOperationsContent } from './cloud-operations-page'

const auth = vi.hoisted(() => ({ capabilities: ['cloudops.read', 'cloudops.manage'] }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (capability: string) => auth.capabilities.includes(capability) }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), sendJSON: vi.fn() }))
const store = { provider: 'cos', bucket: 'dev', region: 'test', public_base_url: 'https://assets.example.com', configured: true, reachable: true, probe_exists: true }
const overview = { primary: store, mirror: { ...store, provider: 'r2' }, edgeone: { configured: true, main_host: 'example.com', asset_host: 'assets.example.com' }, cloudflare: { configured: true, asset_host: 'mirror.example.com' } }
beforeEach(() => { vi.clearAllMocks(); auth.capabilities = ['cloudops.read', 'cloudops.manage']; vi.mocked(getJSON).mockResolvedValue({ total: 0, list: [] }); vi.mocked(sendJSON).mockResolvedValue({ result: { job_id: 'job', status: 'submitted' } }) })
afterEach(() => { cleanup(); vi.restoreAllMocks() })
function setup() { render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><CloudOperationsContent overview={overview} refresh={vi.fn()} /></QueryClientProvider>) }

it('clears the main host using ordinary host purge and hides full-zone purge from a developer', async () => {
  setup()
  expect(screen.queryByRole('button', { name: '清除整个 EdgeOne Zone 缓存' })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '清除主站缓存 · example.com' }))
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(within(await screen.findByRole('dialog')).getByRole('button', { name: '确认清缓存' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/system/cloud/edgeone/purge', 'POST', { type: 'host', targets: ['example.com'] }))
})

it('uses a separate endpoint and explicit scope confirmation for an owner zone purge', async () => {
  const user = userEvent.setup()
  auth.capabilities.push('cloudops.purge_all'); setup()
  const toggle = screen.getByRole('button', { name: '高风险操作' })
  expect(toggle).toHaveAttribute('aria-expanded', 'false')
  expect(screen.queryByRole('button', { name: '清除整个 EdgeOne Zone 缓存' })).not.toBeInTheDocument()
  toggle.focus()
  await user.keyboard('{Enter}')
  expect(toggle).toHaveAttribute('aria-expanded', 'true')
  await user.keyboard(' ')
  expect(toggle).toHaveAttribute('aria-expanded', 'false')
  expect(screen.queryByRole('button', { name: '清除整个 EdgeOne Zone 缓存' })).not.toBeInTheDocument()
  expect(sendJSON).not.toHaveBeenCalled()
  await user.click(toggle)
  fireEvent.click(screen.getByRole('button', { name: '清除整个 EdgeOne Zone 缓存' }))
  const dialog = await screen.findByRole('dialog')
  expect(within(dialog).getByText(/包括所有加速域名/)).toBeInTheDocument()
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.click(within(dialog).getByRole('button', { name: '确认清缓存' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith('/api/v1/system/cloud/edgeone/purge-all', 'POST', undefined))
})

it('shows storage facts and object inspection without mutation controls for read-only access', () => {
  auth.capabilities = ['cloudops.read']; setup()
  expect(screen.getByRole('button', { name: '检查' })).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: '提交清缓存' })).not.toBeInTheDocument()
  expect(screen.queryByText(/同步率/)).not.toBeInTheDocument()
  expect(screen.queryByText(/purge everything/i)).not.toBeInTheDocument()
})

it.each(['EdgeOne', 'Cloudflare'])('uses the shared Select for %s and preserves confirmed request payloads', async (provider) => {
  const user = userEvent.setup(); setup()
  const section = screen.getByRole('heading', { name: `${provider} 清缓存` }).closest('section')!
  expect(section.querySelector('select')).toBeNull()
  await user.click(within(section).getByRole('combobox', { name: `${provider} 清除方式` }))
  await user.click(await screen.findByRole('option', { name: 'Prefix · 路径前缀' }))
  await user.type(within(section).getByLabelText(/清缓存目标/), 'https://assets.example.com/nav/\nhttps://assets.example.com/other/')
  await user.click(within(section).getByRole('button', { name: '提交清缓存' }))
  let dialog = await screen.findByRole('dialog')
  expect(sendJSON).not.toHaveBeenCalled()
  await user.click(within(dialog).getByRole('button', { name: '取消' }))
  expect(sendJSON).not.toHaveBeenCalled()
  await user.click(within(section).getByRole('button', { name: '提交清缓存' }))
  dialog = await screen.findByRole('dialog')
  await user.click(within(dialog).getByRole('button', { name: '确认清缓存' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`/api/v1/system/cloud/${provider.toLowerCase()}/purge`, 'POST', { type: 'prefix', targets: ['https://assets.example.com/nav/', 'https://assets.example.com/other/'] }))
})

it('separates the full-zone action and renders task status, failure and identifiers', async () => {
  auth.capabilities.push('cloudops.purge_all')
  vi.mocked(getJSON).mockResolvedValue({ total: 1, list: [{ job_id: 'task-42', target: 'https://assets.example.com/nav/icon.svg', type: 'url', status: 'failed', created_at: '2026-09-14T00:00:00Z', updated_at: '2026-09-14T00:01:00Z', failure: 'Provider rejected target' }] })
  setup()
  const danger = screen.getByRole('complementary', { name: '高风险操作' })
  fireEvent.click(within(danger).getByRole('button', { name: '高风险操作' }))
  expect(within(danger).getByRole('button', { name: '清除整个 EdgeOne Zone 缓存' })).toBeInTheDocument()
  expect(within(danger).queryByRole('button', { name: /清除主站/ })).not.toBeInTheDocument()
  expect(await screen.findByText('task-42')).toBeInTheDocument()
  expect(screen.getByText('Provider rejected target')).toBeInTheDocument()
  expect(screen.getByText('失败')).toBeInTheDocument()
  expect(screen.queryByRole('combobox', { name: '每页条数' })).not.toBeInTheDocument()
})
