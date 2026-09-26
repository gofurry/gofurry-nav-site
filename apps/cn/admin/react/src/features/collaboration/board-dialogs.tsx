import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { RemoteSelect } from '../../components/admin/operations'
import { FormField } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Dialog } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { getJSON, sendJSON } from '../../lib/api'
import type { OptionItem, PageResult } from '../../lib/types'
import { useAuth } from '../auth/auth-context'
import { base, collaborationError } from './api'
import { boardColors, boardKinds, nodeInput, type BoardEdge, type BoardNode, type BoardNodeInput, type BoardColor, type ReferenceKind } from './board-types'
import { statusLabels, type Idea } from './types'

function IdeaPicker({ onChange }: { onChange: (item: OptionItem) => void }) {
  const [search, setSearch] = useState(''), [keyword, setKeyword] = useState('')
  useEffect(() => { const timer = window.setTimeout(() => setKeyword(search), 300); return () => window.clearTimeout(timer) }, [search])
  const query = useQuery({ queryKey: ['collaboration', 'board-idea-picker', keyword], queryFn: () => getJSON<PageResult<Idea>>(`${base}/ideas?status=all&page_size=10&keyword=${encodeURIComponent(keyword)}`) })
  return <div className="grid gap-2"><Input aria-label="搜索想法卡片" placeholder="搜索内容想法…" value={search} onChange={(event) => setSearch(event.target.value)} /><div className="admin-scroll max-h-40 overflow-auto rounded-md border p-1" role="listbox" aria-label="内容想法">{query.isLoading ? <p className="p-2 text-sm">加载中…</p> : query.error ? <p className="p-2 text-sm text-danger">{query.error.message}</p> : !query.data?.list.length ? <p className="p-2 text-sm text-muted-foreground">无匹配想法</p> : query.data.list.map((idea) => <button className="block w-full rounded px-2 py-2 text-left hover:bg-surface-muted" type="button" role="option" aria-selected={false} key={idea.id} onClick={() => onChange({ id: String(idea.id), label: idea.title || idea.source || `想法 #${idea.id}` })}><span className="block truncate text-sm">{idea.title || idea.source}</span><span className="text-xs text-muted-foreground">{idea.kind} · {statusLabels[idea.status]}</span></button>)}</div></div>
}

export function BoardNodeDialog({ node, initial, close }: { node?: BoardNode; initial: BoardNodeInput; close: () => void }) {
  const client = useQueryClient(), auth = useAuth()
  const [input, setInput] = useState(() => node ? nodeInput(node) : initial)
  const [selection, setSelection] = useState<OptionItem | null>(node?.reference_id ? { id: String(node.reference_id), label: `当前引用 #${node.reference_id}` } : null)
  const save = useMutation({ mutationFn: () => sendJSON<BoardNode>(node ? `${base}/board/nodes/${node.id}` : `${base}/board/nodes`, node ? 'PUT' : 'POST', input), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['collaboration', 'board'] }); close() } })
  const set = (values: Partial<BoardNodeInput>) => setInput((current) => ({ ...current, ...values }))
  const references = [{ value: '', label: '自由卡片' }, { value: 'idea', label: '内容想法' }, ...(auth.can('content.read') ? [{ value: 'game', label: '正式游戏' }, { value: 'site', label: '正式网站' }] : [])]
  const pick = (item: OptionItem | null) => { setSelection(item); set({ reference_id: item ? Number(item.id) : 0 }) }
  const valid = (['rectangle', 'ellipse', 'arrow'].includes(input.kind) || Boolean(input.title.trim() || input.body.trim() || input.reference_id)) && (!input.reference_kind || input.reference_id > 0)
  return <Dialog open title={`${node ? '编辑' : '新建'}${boardKinds[input.kind]}`} onOpenChange={(open) => !open && !save.isPending && close()} footer={<><Button variant="secondary" disabled={save.isPending} onClick={close}>取消</Button><Button disabled={!valid || save.isPending} onClick={() => save.mutate()}>保存</Button></>}>
    <fieldset disabled={save.isPending} className="grid min-w-0 gap-4">
      {input.kind === 'card' && <><FormField label="卡片内容"><Select value={input.reference_kind} options={references} onValueChange={(kind) => { set({ reference_kind: kind as ReferenceKind | '', reference_id: 0 }); setSelection(null) }} /></FormField>{input.reference_kind === 'idea' && <IdeaPicker onChange={pick} />}{(input.reference_kind === 'game' || input.reference_kind === 'site') && <RemoteSelect key={input.reference_kind} resultsLayout="inline" endpoint={`/api/v1/options/${input.reference_kind === 'game' ? 'games' : 'sites'}`} value={selection} onChange={pick} pageSize={10} debounceMs={300} />}{selection && <p className="truncate text-xs text-muted-foreground" title={selection.label}>引用：{selection.label}</p>}</>}
      {!input.reference_kind && <FormField label="标题"><Input maxLength={200} value={input.title} onChange={(event) => set({ title: event.target.value })} /></FormField>}
      {['note', 'card', 'text'].includes(input.kind) && <FormField label="正文"><Textarea rows={5} maxLength={10000} value={input.body} onChange={(event) => set({ body: event.target.value })} /></FormField>}
      <FormField label="颜色"><Select options={boardColors} value={input.color} onValueChange={(color) => set({ color: color as BoardColor })} /></FormField>
      {input.kind === 'arrow' && <FormField label="方向"><Select options={[{ value: '0', label: '向右' }, { value: '90', label: '向下' }, { value: '180', label: '向左' }, { value: '270', label: '向上' }]} value={String(input.rotation)} onValueChange={(rotation) => set({ rotation: Number(rotation) })} /></FormField>}
    </fieldset>
    {save.error && <Alert tone="warning"><div>{collaborationError(save.error)}<p className="mt-1 text-xs">当前草稿仍保留；重新加载将放弃本次编辑。</p><Button variant="secondary" onClick={() => { void client.invalidateQueries({ queryKey: ['collaboration', 'board'] }); close() }}>重新加载</Button></div></Alert>}
  </Dialog>
}

export function BoardEdgeDialog({ edge, close }: { edge: BoardEdge; close: () => void }) {
  const client = useQueryClient()
  const [label, setLabel] = useState(edge.label), [routing, setRouting] = useState(edge.routing), [color, setColor] = useState(edge.color), [arrow, setArrow] = useState(edge.arrow)
  const save = useMutation({ mutationFn: () => sendJSON(`${base}/board/edges/${edge.id}`, 'PUT', { source_id: edge.source_id, target_id: edge.target_id, source_handle: edge.source_handle, target_handle: edge.target_handle, version: edge.version, label, routing, color, arrow }), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['collaboration', 'board'] }); close() } })
  return <Dialog open title="编辑连线" onOpenChange={(open) => !open && !save.isPending && close()} footer={<><Button variant="secondary" disabled={save.isPending} onClick={close}>取消</Button><Button disabled={save.isPending} onClick={() => save.mutate()}>保存</Button></>}><fieldset disabled={save.isPending} className="grid gap-4"><FormField label="连线标签"><Input maxLength={200} value={label} onChange={(event) => setLabel(event.target.value)} /></FormField><FormField label="连线样式"><Select options={[{ value: 'curve', label: '曲线' }, { value: 'step', label: '折线' }]} value={routing} onValueChange={(value) => setRouting(value as 'curve' | 'step')} /></FormField><FormField label="颜色"><Select options={boardColors} value={color} onValueChange={(value) => setColor(value as BoardColor)} /></FormField><label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={arrow} onChange={(event) => setArrow(event.target.checked)} />显示箭头</label></fieldset>{save.error && <Alert tone="warning">{collaborationError(save.error)}<p className="mt-1 text-xs">当前草稿仍保留；重新加载将放弃本次编辑。</p><Button variant="secondary" onClick={() => { void client.invalidateQueries({ queryKey: ['collaboration', 'board'] }); close() }}>重新加载</Button></Alert>}</Dialog>
}
