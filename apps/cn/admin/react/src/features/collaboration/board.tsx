import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRef, useState, type PointerEvent } from 'react'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction } from '../../components/ui/dialog'
import { Textarea } from '../../components/ui/input'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { getJSON, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { base, collaborationError } from './api'
import type { BoardNote } from './types'

function payload(note: BoardNote) { return { body: note.body, x: note.x, y: note.y, width: note.width, height: note.height, z_index: note.z_index, version: note.version } }
export function boardGeometry(note: BoardNote, mode: 'drag' | 'resize', dx: number, dy: number): BoardNote {
  const clamp = (value: number, min: number, max: number) => Math.round(Math.max(min, Math.min(max, value)))
  return mode === 'drag' ? { ...note, x: clamp(note.x + dx, 0, 20000), y: clamp(note.y + dy, 0, 20000) } : { ...note, width: clamp(note.width + dx, 160, 1600), height: clamp(note.height + dy, 120, 1600) }
}

export function BoardNoteCard({ note, topZ, writable }: { note: BoardNote; topZ: number; writable: boolean }) {
  const client = useQueryClient()
  const [draft, setDraft] = useState<BoardNote | null>(null)
  const [deleting, setDeleting] = useState(false)
  const gesture = useRef<{ start: BoardNote; current: BoardNote; mode: 'drag' | 'resize'; x: number; y: number; pointer: number; previous: BoardNote | null } | null>(null)
  const mutation = useMutation({ mutationFn: ({ value, remove = false }: { value: BoardNote; remove?: boolean }) => sendJSON<BoardNote>(`${base}/board/notes/${note.id}`, remove ? 'DELETE' : 'PUT', remove ? { version: value.version } : payload(value)), onSuccess: async (saved, variables) => {
    client.setQueryData<BoardNote[]>(['collaboration', 'board'], (notes) => notes?.flatMap((item) => item.id === note.id ? variables.remove ? [] : [saved] : [item]))
    setDraft(null); setDeleting(false)
    await client.invalidateQueries({ queryKey: ['collaboration', 'board'] })
  } })
  const value = draft ?? note
  const start = (event: PointerEvent<HTMLButtonElement>, mode: 'drag' | 'resize') => {
    if (!writable || mutation.isPending || event.button !== 0) return
    event.preventDefault(); event.currentTarget.setPointerCapture(event.pointerId)
    const initial = { ...value, z_index: Math.min(topZ + 1, 1000000) }
    gesture.current = { start: initial, current: initial, mode, x: event.clientX, y: event.clientY, pointer: event.pointerId, previous: draft }; setDraft(initial)
  }
  const move = (event: PointerEvent<HTMLButtonElement>) => {
    const state = gesture.current; if (!state || state.pointer !== event.pointerId) return
    state.current = boardGeometry(state.start, state.mode, event.clientX - state.x, event.clientY - state.y); setDraft(state.current)
  }
  const finish = (event: PointerEvent<HTMLButtonElement>, canceled = false) => {
    const state = gesture.current; if (!state || state.pointer !== event.pointerId) return
    gesture.current = null
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId)
    if (canceled) { setDraft(state.previous); return }
    mutation.mutate({ value: state.current })
  }
  const handlers = { onPointerMove: move, onPointerUp: (event: PointerEvent<HTMLButtonElement>) => finish(event), onPointerCancel: (event: PointerEvent<HTMLButtonElement>) => finish(event, true) }
  return <article aria-label={`便笺 ${note.id}`} className="absolute flex flex-col gap-2 rounded-md border bg-surface p-3 shadow-sm" style={{ left: value.x, top: value.y, width: value.width, height: value.height, zIndex: value.z_index }}>
    <button type="button" aria-label={`移动便笺 ${note.id}`} disabled={!writable || mutation.isPending} className="touch-none cursor-move rounded bg-surface-muted p-1 text-left text-xs focus:ring-2 focus:ring-ring" onPointerDown={(event) => start(event, 'drag')} {...handlers} onKeyDown={(event) => {
      const movement: Record<string, [number, number]> = { ArrowLeft: [-10, 0], ArrowRight: [10, 0], ArrowUp: [0, -10], ArrowDown: [0, 10] }
      if (movement[event.key]) { event.preventDefault(); mutation.mutate({ value: boardGeometry(value, 'drag', ...movement[event.key]) }) }
    }}>拖动 · #{note.id}（方向键移动）</button>
    <Textarea aria-label={`便笺正文 ${note.id}`} className="min-h-0 flex-1 resize-none" value={value.body} maxLength={10000} readOnly={!writable} disabled={mutation.isPending} onChange={(e) => setDraft({ ...value, body: e.target.value })} />
    {mutation.error && <div role="alert" className="max-h-24 overflow-auto text-xs text-danger">{collaborationError(mutation.error)}<button type="button" className="underline" onClick={() => { setDraft(null); mutation.reset(); void client.invalidateQueries({ queryKey: ['collaboration', 'board'] }) }}>重新加载</button></div>}
    {writable && <div className="flex flex-wrap gap-1 text-xs"><Button variant="secondary" disabled={!draft || mutation.isPending} onClick={() => mutation.mutate({ value })}>保存正文</Button><Button variant="secondary" disabled={mutation.isPending} onClick={() => mutation.mutate({ value: { ...value, z_index: Math.min(topZ + 1, 1000000) } })}>置顶</Button><Button variant="secondary" disabled={mutation.isPending} onClick={() => setDeleting(true)}>删除</Button></div>}
    {writable && <button type="button" aria-label={`调整便笺尺寸 ${note.id}`} disabled={mutation.isPending} className="absolute bottom-0 right-0 size-5 touch-none cursor-se-resize text-muted-foreground" onPointerDown={(event) => start(event, 'resize')} {...handlers} onKeyDown={(event) => {
      if (['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key)) { event.preventDefault(); mutation.mutate({ value: boardGeometry(value, 'resize', event.key === 'ArrowRight' ? 10 : event.key === 'ArrowLeft' ? -10 : 0, event.key === 'ArrowDown' ? 10 : event.key === 'ArrowUp' ? -10 : 0) }) }
    }}>◢</button>}
    <ConfirmAction open={deleting} onOpenChange={setDeleting} title="删除共享便笺" description="其他成员也会看到删除结果。" busy={mutation.isPending} onConfirm={() => mutation.mutate({ value, remove: true })} />
  </article>
}

export function SharedBoard() {
  const auth = useAuth(), client = useQueryClient()
  const query = useQuery({ queryKey: ['collaboration', 'board'], queryFn: () => getJSON<BoardNote[]>(`${base}/board/notes`), refetchInterval: 12000 })
  const [body, setBody] = useState('')
  const create = useMutation({ mutationFn: () => sendJSON<BoardNote>(`${base}/board/notes`, 'POST', { body, x: 40, y: 40, width: 320, height: 260, z_index: Math.min(Math.max(0, ...(query.data ?? []).map((note) => note.z_index)) + 1, 1000000) }), onSuccess: async () => { setBody(''); await client.invalidateQueries({ queryKey: ['collaboration', 'board'] }) } })
  if (query.isLoading) return <LoadingState />
  if (query.error && !query.data) return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />
  const notes = query.data ?? []
  return <div className="grid gap-4"><p className="text-sm text-muted-foreground">全局共享文本便笺，每 12 秒刷新。拖动或调整尺寸后松开保存；正文使用保存按钮。</p>
    {query.error && <Alert tone="warning">刷新失败，保留当前便笺与未保存的编辑。<Button variant="secondary" onClick={() => void query.refetch()}>重试刷新</Button></Alert>}
    {auth.can('collaboration.write') && <div className="flex items-start gap-3"><Textarea aria-label="新便笺正文" value={body} maxLength={10000} onChange={(e) => setBody(e.target.value)} /><Button disabled={!body.trim() || create.isPending} onClick={() => create.mutate()}>新建便笺</Button></div>}
    {create.error && <Alert tone="danger">{collaborationError(create.error)}</Alert>}
    <div className="relative h-[65vh] isolate overflow-auto rounded-md border bg-surface-muted"><div className="relative" style={{ width: Math.max(4000, ...notes.map((note) => note.x + note.width + 200)), height: Math.max(3000, ...notes.map((note) => note.y + note.height + 200)) }}>{notes.map((note) => <BoardNoteCard key={note.id} note={note} topZ={Math.max(0, ...notes.map((item) => item.z_index))} writable={auth.can('collaboration.write')} />)}</div></div>
  </div>
}
