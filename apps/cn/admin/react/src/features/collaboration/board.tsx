import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Background, BackgroundVariant, ConnectionMode, Controls, MarkerType, MiniMap, ReactFlow, ReactFlowProvider, applyNodeChanges, useReactFlow, type Connection, type NodeChange, type ResizeParams } from '@xyflow/react'
import { ArrowRight, ArrowsInSimple, ArrowsOutSimple, Cards, Circle, Cursor, Hand, Note, Rectangle, TextT } from '@phosphor-icons/react'
import '@xyflow/react/dist/style.css'
import { useTheme } from '../../app/theme'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction } from '../../components/ui/dialog'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { getJSON, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { base, collaborationError } from './api'
import { BoardActions } from './board-context'
import { BoardEdgeDialog, BoardNodeDialog } from './board-dialogs'
import { CanvasEdgeView, CanvasNodeView } from './board-elements'
import { boardKinds, canvasNode, colorValues, layoutInput, mergeBoardNodes, newNode, nodeInput, type BoardDocument, type BoardEdge, type BoardKind, type BoardLayout, type BoardNode, type BoardNodeInput, type BoardReference, type CanvasEdge, type CanvasNode } from './board-types'
import './board.css'

const nodeTypes = { canvas: CanvasNodeView }, edgeTypes = { relationship: CanvasEdgeView }
const boardKey = ['collaboration', 'board']
const tools = [{ kind: 'note', icon: Note }, { kind: 'card', icon: Cards }, { kind: 'text', icon: TextT }, { kind: 'rectangle', icon: Rectangle }, { kind: 'ellipse', icon: Circle }, { kind: 'arrow', icon: ArrowRight }] as const
const empty: BoardDocument = { nodes: [], edges: [], references: [] }
export function SharedBoard() { return <ReactFlowProvider><BoardCanvas /></ReactFlowProvider> }

function BoardCanvas() {
  const auth = useAuth(), client = useQueryClient(), navigate = useNavigate(), flow = useReactFlow<CanvasNode, CanvasEdge>(), theme = useTheme()
  const query = useQuery({ queryKey: boardKey, queryFn: () => getJSON<BoardDocument>(`${base}/board`), refetchInterval: 12000, refetchOnWindowFocus: false })
  const [nodes, setNodes] = useState<CanvasNode[]>([]), nodesRef = useRef<CanvasNode[]>([])
  const dirty = useRef(new Set<string>()), saving = useRef(false), framed = useRef(false)
  const [selectedEdge, setSelectedEdge] = useState<string | null>(null)
  const [mode, setMode] = useState<'select' | 'pan'>('select'), [expanded, setExpanded] = useState(false)
  const [editor, setEditor] = useState<{ node?: BoardNode; initial: BoardNodeInput } | null>(null), [edgeEditor, setEdgeEditor] = useState<BoardEdge | null>(null)
  const [deleting, setDeleting] = useState<{ type: 'nodes' | 'edges'; id: number; version: number } | null>(null)
  const host = useRef<HTMLDivElement>(null)
  const replaceNodes = useCallback((next: CanvasNode[]) => { nodesRef.current = next; setNodes(next) }, [])
  useEffect(() => { if (query.data) replaceNodes(mergeBoardNodes(query.data, nodesRef.current, dirty.current)) }, [query.data, replaceNodes])
  useEffect(() => { if (query.data && !framed.current) { framed.current = true; void flow.fitView({ duration: 0, padding: .2, maxZoom: 1 }) } }, [query.data, flow])
  useEffect(() => { if (!expanded) return; const escape = (event: KeyboardEvent) => { if (event.key === 'Escape' && !document.querySelector('[role="dialog"]')) setExpanded(false) }; window.addEventListener('keydown', escape); return () => window.removeEventListener('keydown', escape) }, [expanded])

  const layout = useMutation({ mutationFn: (items: BoardLayout[]) => sendJSON<BoardNode[]>(`${base}/board/nodes/layout`, 'PUT', { nodes: items }), onSuccess: async (saved) => {
    const updated = new Map(saved.map((node) => [String(node.id), node]))
    saved.forEach((node) => dirty.current.delete(String(node.id)))
    replaceNodes(nodesRef.current.map((node) => updated.has(node.id) ? canvasNode(updated.get(node.id)!, query.data?.references, node) : node))
    client.setQueryData<BoardDocument>(boardKey, (current) => current && ({ ...current, nodes: current.nodes.map((node) => updated.get(String(node.id)) ?? node) }))
    await client.invalidateQueries({ queryKey: boardKey })
  }, onSettled: () => { saving.current = false } })
  const createEdge = useMutation({ mutationFn: (connection: Connection) => sendJSON<BoardEdge>(`${base}/board/edges`, 'POST', { source_id: Number(connection.source), target_id: Number(connection.target), source_handle: connection.sourceHandle, target_handle: connection.targetHandle, routing: 'curve', label: '', color: 'slate', arrow: true }), onSuccess: async () => { await client.invalidateQueries({ queryKey: boardKey }) } })
  const remove = useMutation({ mutationFn: () => sendJSON(`${base}/board/${deleting!.type}/${deleting!.id}`, 'DELETE', { version: deleting!.version }), onSuccess: async () => { setDeleting(null); setSelectedEdge(null); await client.invalidateQueries({ queryKey: boardKey }) } })
  const busy = layout.isPending || createEdge.isPending || remove.isPending
  const writable = auth.can('collaboration.write') && !busy && !layout.error
  const commitLayout = (items: BoardLayout[]) => {
    if (!items.length || saving.current || !writable) return
    saving.current = true
    items.forEach((item) => dirty.current.add(String(item.id)))
    layout.mutate(items)
  }
  const onNodesChange = (changes: NodeChange<CanvasNode>[]) => {
    const allowed = changes.filter((change) => change.type !== 'remove' && (!saving.current || change.type === 'select' || change.type === 'dimensions'))
    const next = applyNodeChanges(allowed, nodesRef.current)
    const finalIDs = new Set<string>()
    for (const change of allowed) if (change.type === 'position' && change.position) {
      dirty.current.add(change.id)
      // React Flow emits dragging=false on pointer release and keyboard moves.
      if (change.dragging === false) finalIDs.add(change.id)
    }
    replaceNodes(next)
    if (finalIDs.size) commitLayout(next.filter((node) => finalIDs.has(node.id)).map((node) => layoutInput(node)))
  }
  const resize = (id: string, params: ResizeParams) => {
    const node = nodesRef.current.find((item) => item.id === id); if (!node) return
    const geometry = { x: Math.round(params.x), y: Math.round(params.y), width: Math.round(params.width), height: Math.round(params.height) }
    replaceNodes(nodesRef.current.map((item) => item.id === id ? { ...item, position: { x: geometry.x, y: geometry.y }, width: geometry.width, height: geometry.height } : item))
    commitLayout([layoutInput(node, geometry)])
  }
  const raise = (id: string) => { const node = nodesRef.current.find((item) => item.id === id); if (node) commitLayout([layoutInput(node, { z_index: Math.min(Math.max(0, ...nodesRef.current.map((item) => item.zIndex ?? 0)) + 1, 1000000) })]) }
  const openReference = (ref: BoardReference) => {
    if (ref.kind === 'idea') navigate(`/collaboration?tab=ideas&status=all&keyword=${encodeURIComponent(ref.title)}`)
    else if (auth.can('content.read')) navigate(ref.kind === 'game' ? `/game/games/${ref.id}` : `/nav/sites/${ref.id}`)
  }
  const add = (kind: BoardKind) => {
    const rect = host.current?.getBoundingClientRect(); if (!rect) return
    const position = flow.screenToFlowPosition({ x: rect.left + rect.width / 2 - 120, y: rect.top + rect.height / 2 - 100 })
    const z = ['rectangle', 'ellipse', 'arrow'].includes(kind) ? 0 : Math.min(Math.max(0, ...nodes.map((node) => node.zIndex ?? 0)) + 1, 1000000)
    setEditor({ initial: newNode(kind, position, z) })
  }
  const reload = () => { dirty.current.clear(); layout.reset(); createEdge.reset(); replaceNodes(mergeBoardNodes(query.data ?? empty, [], dirty.current)); void query.refetch() }
  const edges = useMemo<CanvasEdge[]>(() => (query.data?.edges ?? []).map((record) => ({ id: String(record.id), type: 'relationship', source: String(record.source_id), target: String(record.target_id), sourceHandle: record.source_handle, targetHandle: record.target_handle, selected: String(record.id) === selectedEdge, data: { record }, markerEnd: record.arrow ? { type: MarkerType.ArrowClosed, color: colorValues[record.color] } : undefined })), [query.data?.edges, selectedEdge])
  if (query.isLoading) return <LoadingState />
  if (query.error && !query.data) return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />
  const error = layout.error || createEdge.error
  return <BoardActions.Provider value={{ writable, openReference, editNode: (node) => setEditor({ node, initial: nodeInput(node) }), editEdge: setEdgeEditor, deleteNode: (node) => { remove.reset(); setDeleting({ type: 'nodes', id: node.id, version: node.version }) }, deleteEdge: (edge) => { remove.reset(); setDeleting({ type: 'edges', id: edge.id, version: edge.version }) }, raise, beginResize: (id) => dirty.current.add(id), resize }}>
    <section className={`collaboration-canvas${expanded ? ' is-expanded' : ''}`} aria-label="共享画布">
      <div className="canvas-topbar"><div className="flex flex-wrap items-center gap-1" role="toolbar" aria-label="画布工具">
        <Button size="icon" variant={mode === 'select' ? 'primary' : 'ghost'} aria-label="选择工具" title="选择 / 框选" onClick={() => setMode('select')}><Cursor className="size-4" /></Button><Button size="icon" variant={mode === 'pan' ? 'primary' : 'ghost'} aria-label="平移工具" title="平移（按住空格）" onClick={() => setMode('pan')}><Hand className="size-4" /></Button>
        {auth.can('collaboration.write') && <><span className="mx-1 h-5 border-l" />{tools.map(({ kind, icon: Icon }) => <Button key={kind} size="icon" variant="ghost" disabled={!writable} aria-label={`新建${boardKinds[kind]}`} title={`新建${boardKinds[kind]}`} onClick={() => add(kind)}><Icon className="size-4" /></Button>)}</>}
      </div><div className="flex items-center gap-2"><span className="canvas-save-state" role="status">{busy ? '保存中…' : error ? '未保存' : nodes.some((node) => node.dragging || node.resizing) ? '调整中…' : query.error ? '刷新失败' : '已同步'}</span><Button size="icon" variant="ghost" aria-label={expanded ? '退出全屏画布' : '全屏画布'} title={expanded ? '退出全屏' : '全屏画布'} onClick={() => setExpanded(!expanded)}>{expanded ? <ArrowsInSimple className="size-4" /> : <ArrowsOutSimple className="size-4" />}</Button></div></div>
      {(error || query.error) && <Alert tone="warning"><div>{error ? collaborationError(error) : '刷新失败，保留当前画布和草稿。'}<Button variant="secondary" size="sm" className="ml-2" onClick={reload}>{layout.error ? '放弃本地移动并重新加载' : '重新加载'}</Button></div></Alert>}
      <div className="canvas-stage" ref={host}>
        <ReactFlow<CanvasNode, CanvasEdge> nodes={nodes} edges={edges} nodeTypes={nodeTypes} edgeTypes={edgeTypes} onNodesChange={onNodesChange} onEdgesChange={(changes) => { const selection = changes.find((change) => change.type === 'select' && change.selected); if (selection?.type === 'select') setSelectedEdge(selection.id) }} onNodeDoubleClick={(_event, node) => writable && setEditor({ node: node.data.record, initial: nodeInput(node.data.record) })} onEdgeDoubleClick={(_event, edge) => writable && setEdgeEditor(edge.data!.record)} onConnect={(connection) => writable && createEdge.mutate(connection)} isValidConnection={(connection) => connection.source !== connection.target && !edges.some((edge) => edge.source === connection.source && edge.target === connection.target && edge.sourceHandle === connection.sourceHandle && edge.targetHandle === connection.targetHandle)} onNodeClick={() => setSelectedEdge(null)} onPaneClick={() => setSelectedEdge(null)} connectionMode={ConnectionMode.Loose} nodesDraggable={writable && mode === 'select'} nodesConnectable={writable && mode === 'select'} edgesReconnectable={false} deleteKeyCode={null} selectionOnDrag={mode === 'select'} panOnDrag={mode === 'pan' ? true : [1, 2]} panActivationKeyCode="Space" selectionKeyCode="Shift" snapToGrid snapGrid={[10, 10]} nodeExtent={[[-100000, -100000], [100000, 100000]]} minZoom={.15} maxZoom={2} fitView fitViewOptions={{ maxZoom: 1, padding: .2 }} colorMode={theme.resolvedTheme} proOptions={{ hideAttribution: true }} ariaLabelConfig={{ 'controls.zoomIn.ariaLabel': '放大画布', 'controls.zoomOut.ariaLabel': '缩小画布', 'controls.fitView.ariaLabel': '适应全部内容', 'minimap.ariaLabel': '画布小地图' }}>
          <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="var(--border)" /><Controls showInteractive={false} /><MiniMap nodeColor={(node) => colorValues[(node as CanvasNode).data.record.color]} pannable zoomable />
          {!nodes.length && <div className="canvas-empty nodrag nopan"><p>从一张便签开始</p><span>{auth.can('collaboration.write') ? '添加便签或内容卡片，拖动连接点建立联系。' : '这里还没有共享内容。'}</span>{auth.can('collaboration.write') && <Button disabled={!writable} onClick={() => add('note')}>添加第一张便签</Button>}</div>}
        </ReactFlow>
      </div>
    </section>
    {editor && <BoardNodeDialog node={editor.node} initial={editor.initial} close={() => setEditor(null)} />}
    {edgeEditor && <BoardEdgeDialog edge={edgeEditor} close={() => setEdgeEditor(null)} />}
    <ConfirmAction open={Boolean(deleting)} onOpenChange={(open) => !open && !remove.isPending && setDeleting(null)} title={deleting?.type === 'nodes' ? '删除画布元素' : '删除连线'} description={deleting?.type === 'nodes' ? '该元素及相连的连线将从共享画布移除，引用的想法、游戏或网站会保留。' : '这条连线将从共享画布移除。'} busy={remove.isPending} onConfirm={() => remove.mutate()}>{remove.error && <Alert tone="warning">{collaborationError(remove.error)}<Button variant="secondary" onClick={() => { setDeleting(null); remove.reset(); void query.refetch() }}>重新加载</Button></Alert>}</ConfirmAction>
  </BoardActions.Provider>
}
