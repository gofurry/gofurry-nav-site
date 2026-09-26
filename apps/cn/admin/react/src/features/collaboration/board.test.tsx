import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import type { ReactNode } from 'react'
import type { ReactFlowProps } from '@xyflow/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { ApiError, getJSON, listJSON, sendJSON } from '../../lib/api'
import { SharedBoard } from './board'
import { BoardNodeDialog, BoardEdgeDialog } from './board-dialogs'
import { canvasNode, layoutInput, mergeBoardNodes, nodeInput, type BoardDocument, type BoardNode, type BoardEdge, type CanvasNode, type CanvasEdge } from './board-types'
import { base } from './api'

const runtime = vi.hoisted(() => ({ writable: true, flow: { fitView: vi.fn(), screenToFlowPosition: (point: { x: number; y: number }) => point } }))
vi.mock('../auth/auth-context', () => ({ useAuth: () => ({ can: (value: string) => runtime.writable || !value.endsWith('.write') }) }))
vi.mock('../../app/theme', () => ({ useTheme: () => ({ resolvedTheme: 'dark' }) }))
vi.mock('../../lib/api', async (original) => ({ ...await original<typeof import('../../lib/api')>(), getJSON: vi.fn(), listJSON: vi.fn(), sendJSON: vi.fn() }))
// React Flow's event boundary is exercised here; browser verification covers its
// real pointer/resize/connection implementation, which depends on layout metrics.
vi.mock('@xyflow/react', async (original) => ({ ...await original<typeof import('@xyflow/react')>(), useReactFlow: () => runtime.flow, ReactFlowProvider: ({ children }: { children: ReactNode }) => children, Background: () => null, Controls: () => null, MiniMap: () => null,
  ReactFlow: ({ nodes = [], onNodesChange, onNodeDoubleClick, onConnect, children }: ReactFlowProps<CanvasNode, CanvasEdge>) => <div>{nodes.map((node) => <div key={node.id}><button aria-label={`编辑节点 ${node.id}`} onDoubleClick={(event) => onNodeDoubleClick?.(event, node)}>{node.data.record.body}</button><button aria-label={`移动节点 ${node.id}`} onPointerDown={() => onNodesChange?.([{ id: node.id, type: 'position', position: { x: 10, y: 10 }, dragging: true }])} onPointerMove={() => onNodesChange?.([{ id: node.id, type: 'position', position: { x: 60, y: 40 }, dragging: true }])} onPointerUp={() => onNodesChange?.([{ id: node.id, type: 'position', position: { x: 60, y: 40 }, dragging: false }])}>移动</button><span data-testid={`position-${node.id}`}>{node.position.x},{node.position.y}</span></div>)}<button onClick={() => onConnect?.({ source: '1', target: '2', sourceHandle: 'right', targetHandle: 'left' })}>连接两个节点</button>{children}</div>,
}))
const node: BoardNode = { id: 1, kind: 'note', title: '便签', body: '共享文字', color: 'sand', rotation: 0, reference_kind: null, reference_id: null, x: 0, y: 0, width: 300, height: 220, z_index: 0, version: 3 }
const edge: BoardEdge = { id: 5, source_id: 1, target_id: 2, source_handle: 'right', target_handle: 'left', routing: 'curve', label: '', color: 'slate', arrow: true, version: 2 }
const document: BoardDocument = { nodes: [node, { ...node, id: 2 }], edges: [], references: [] }
function setup(element: ReactNode) { const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } }); render(<QueryClientProvider client={client}><MemoryRouter>{element}</MemoryRouter></QueryClientProvider>); return client }
beforeEach(() => { runtime.writable = true; vi.mocked(getJSON).mockResolvedValue(document) })
afterEach(() => { cleanup(); vi.resetAllMocks() })

it('persists a drag only on pointer release and retains the local position after 409', async () => {
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('stale', 409))
  const client = setup(<SharedBoard />)
  const move = await screen.findByRole('button', { name: '移动节点 1' })
  fireEvent.pointerDown(move); fireEvent.pointerMove(move)
  expect(sendJSON).not.toHaveBeenCalled()
  fireEvent.pointerUp(move)
  await screen.findByText('此内容刚刚被其他成员修改，请重新加载后重试。')
  expect(sendJSON).toHaveBeenCalledExactlyOnceWith(`${base}/board/nodes/layout`, 'PUT', { nodes: [expect.objectContaining({ id: 1, version: 3, x: 60, y: 40 })] })
  vi.mocked(getJSON).mockResolvedValue({ ...document, nodes: [{ ...node, x: 900, version: 4 }, document.nodes[1]] })
  await client.invalidateQueries({ queryKey: ['collaboration', 'board'] })
  expect(screen.getByTestId('position-1')).toHaveTextContent('60,40')
  fireEvent.click(screen.getByRole('button', { name: '放弃本地移动并重新加载' }))
  await waitFor(() => expect(screen.getByTestId('position-1')).toHaveTextContent('900,0'))
})
it('preserves a dirty node deleted remotely while updating other nodes', () => {
  const local = [canvasNode({ ...node, x: 40 }), canvasNode({ ...node, id: 2 })]
  const merged = mergeBoardNodes({ ...document, nodes: [{ ...node, id: 2, x: 400, version: 4 }] }, local, new Set(['1']))
  expect(merged.find((item) => item.id === '1')?.position.x).toBe(40)
  expect(merged.find((item) => item.id === '2')?.position.x).toBe(400)
  expect(layoutInput(local[0], { width: 450, height: 280 })).toMatchObject({ id: 1, version: 3, x: 40, width: 450, height: 280 })
})
it('creates a version-independent edge between existing nodes and exposes the save error', async () => {
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('端点已删除', 409))
  setup(<SharedBoard />)
  fireEvent.click(await screen.findByRole('button', { name: '连接两个节点' }))
  await screen.findByText('此内容刚刚被其他成员修改，请重新加载后重试。')
  expect(sendJSON).toHaveBeenCalledWith(`${base}/board/edges`, 'POST', expect.objectContaining({ source_id: 1, target_id: 2, source_handle: 'right', target_handle: 'left' }))
})
it('keeps node editing drafts on version conflict without resubmitting', async () => {
  vi.mocked(sendJSON).mockRejectedValue(new ApiError('stale', 409))
  setup(<BoardNodeDialog node={node} initial={nodeInput(node)} close={vi.fn()} />)
  fireEvent.change(screen.getByLabelText('正文'), { target: { value: '未保存草稿' } })
  fireEvent.click(screen.getByRole('button', { name: '保存' }))
  await screen.findByText('此内容刚刚被其他成员修改，请重新加载后重试。')
  expect(screen.getByLabelText('正文')).toHaveValue('未保存草稿')
  expect(sendJSON).toHaveBeenCalledExactlyOnceWith(`${base}/board/nodes/1`, 'PUT', expect.objectContaining({ body: '未保存草稿', version: 3 }))
})
it('creates reference cards using the existing options API and only writes GFA', async () => {
  const card = { ...node, kind: 'card' as const, reference_kind: 'game' as const, reference_id: 0 }
  vi.mocked(listJSON).mockResolvedValue({ total: 1, list: [{ id: '42', label: '正式游戏' }] })
  vi.mocked(sendJSON).mockResolvedValue({ ...card, reference_id: 42 })
  setup(<BoardNodeDialog initial={nodeInput(card)} close={vi.fn()} />)
  fireEvent.click(await screen.findByRole('option', { name: '正式游戏' }))
  fireEvent.click(screen.getByRole('button', { name: '保存' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/board/nodes`, 'POST', expect.objectContaining({ reference_kind: 'game', reference_id: 42 })))
  expect(listJSON).toHaveBeenCalledWith('/api/v1/options/games', 1, 10, '')
  expect(vi.mocked(sendJSON).mock.calls).toHaveLength(1)
})
it('saves an edited connection label with its displayed version', async () => {
  vi.mocked(sendJSON).mockResolvedValue(edge)
  setup(<BoardEdgeDialog edge={edge} close={vi.fn()} />)
  fireEvent.change(screen.getByLabelText('连线标签'), { target: { value: '参考资料' } })
  fireEvent.click(screen.getByRole('button', { name: '保存' }))
  await waitFor(() => expect(sendJSON).toHaveBeenCalledWith(`${base}/board/edges/5`, 'PUT', expect.objectContaining({ label: '参考资料', version: 2 })))
})
it('keeps the canvas readable but removes editing tools without write capability', async () => {
  runtime.writable = false
  setup(<SharedBoard />)
  await screen.findAllByText('共享文字')
  expect(screen.queryByRole('button', { name: '新建便签' })).not.toBeInTheDocument()
  fireEvent.click(screen.getByRole('button', { name: '连接两个节点' }))
  expect(sendJSON).not.toHaveBeenCalled()
})
