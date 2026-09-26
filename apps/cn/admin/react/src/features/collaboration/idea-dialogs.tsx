import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { RemoteSelect } from '../../components/admin/operations'
import { FormField } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Dialog } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { getJSON, sendJSON } from '../../lib/api'
import type { OptionItem } from '../../lib/types'
import { base, collaborationError, preview, transition, useRefreshCollaboration } from './api'
import { parsePaste } from './parser'
import { ideaInput, kindOptions, priorityOptions, warningLabels, type Idea, type IdeaInput, type Kind, type PreviewRow } from './types'

export function PreviewRows({ rows }: { rows: PreviewRow[] }) {
  return <div className="max-h-72 overflow-auto"><table className="w-full table-fixed text-left text-xs"><thead><tr><th className="w-8 py-2">条</th><th className="w-[22%] p-2">名称</th><th className="w-1/4 p-2">来源</th><th className="w-1/5 p-2">备注</th><th className="p-2">检查结果</th></tr></thead><tbody>{rows.map((row) => <tr key={row.index} className="border-t align-top"><td className="py-2">{row.index + 1}</td><td className="p-2"><div className="truncate" title={row.title}>{row.title || '—'}</div></td><td className="p-2"><div className="truncate" title={row.source}>{row.source || '—'}</div>{row.display_hint && row.display_hint !== row.source && <div className="truncate text-muted-foreground" title={row.display_hint}>{row.display_hint}</div>}</td><td className="p-2"><div className="truncate" title={row.note}>{row.note || '—'}</div></td><td className="break-words p-2">{!row.valid ? row.errors.join('；') : row.warnings.length ? row.warnings.map((warning) => warningLabels[warning]).join('；') : '新内容'}{row.idea_match && <div>想法 #{row.idea_match.id}</div>}{row.resource_match && <div className="truncate" title={row.resource_match.title}>正式内容 #{row.resource_match.id} {row.resource_match.title}</div>}</td></tr>)}</tbody></table></div>
}

const batchExamples: Record<Kind, string> = {
  game: '求生之路 2 | steam:550 | 核对介绍\n传送门 2 | steam:620 | 补充资料',
  site: '示例站点 | https://example.com | 核对站点介绍\n文档站点 | https://docs.example.com | 独立子域',
  other: '整理投稿规范 | | 补充说明\n收集活动线索 | | 待整理',
}

export function IdeaEditorDialog({ idea, close }: { idea?: Idea; close: () => void }) {
  const refresh = useRefreshCollaboration()
  const [input, setInput] = useState<IdeaInput>(() => idea ? ideaInput(idea) : { kind: 'game', title: '', source: '', note: '', priority: 'normal' })
  const [checked, setChecked] = useState<PreviewRow[] | null>(null)
  const check = useMutation({ mutationFn: () => preview([input]), onSuccess: setChecked })
  const save = useMutation({ mutationFn: () => sendJSON(idea ? `${base}/ideas/${idea.id}` : `${base}/ideas`, idea ? 'PUT' : 'POST', input), onSuccess: async () => { await refresh(); close() } })
  const reload = useMutation({ mutationFn: () => getJSON<Idea>(`${base}/ideas/${idea!.id}`), onSuccess: (fresh) => { setInput(ideaInput(fresh)); setChecked(null); save.reset(); check.reset() } })
  const busy = check.isPending || save.isPending || reload.isPending
  const set = (values: Partial<IdeaInput>) => { setInput({ ...input, ...values }); setChecked(null); check.reset() }
  return <Dialog open onOpenChange={(open) => !open && !busy && close()} title={idea ? '编辑想法' : '快速加入想法'} footer={<><Button variant="secondary" disabled={busy} onClick={close}>取消</Button>{!idea && <Button variant="secondary" disabled={busy} onClick={() => check.mutate()}>检查重复</Button>}<Button disabled={busy || (idea ? !input.title.trim() && !input.source.trim() : !checked?.[0].valid)} onClick={() => save.mutate()}>{idea ? '保存' : '加入储备'}</Button></>}>
    <div className="grid gap-4"><FormField label="类型"><Select disabled={busy || idea?.status === 'landed'} options={kindOptions} value={input.kind} onValueChange={(kind) => set({ kind: kind as Kind })} /></FormField><FormField label="名称 / 线索"><Input disabled={busy} maxLength={500} value={input.title} onChange={(e) => set({ title: e.target.value })} /></FormField><FormField label="来源"><Input disabled={busy} maxLength={2048} value={input.source} onChange={(e) => set({ source: e.target.value })} /></FormField><FormField label="备注"><Textarea disabled={busy} maxLength={10000} rows={5} value={input.note} onChange={(e) => set({ note: e.target.value })} /></FormField><FormField label="优先级"><Select disabled={busy} options={priorityOptions} value={input.priority} onValueChange={(value) => set({ priority: value as IdeaInput['priority'] })} /></FormField></div>
    {(check.error || save.error || reload.error) && <Alert tone="danger">{collaborationError(reload.error || save.error || check.error)}{idea && <Button variant="secondary" disabled={busy} onClick={() => reload.mutate()}>重新加载</Button>}</Alert>}{checked && <><p className="my-3 text-sm">重复仅是提示，仍可加入。</p><PreviewRows rows={checked} /></>}
  </Dialog>
}

export function BatchIdeaDialog({ close }: { close: () => void }) {
  const refresh = useRefreshCollaboration()
  const [kind, setKind] = useState<Kind>('game')
  const [text, setText] = useState('')
  const [items, setItems] = useState<IdeaInput[]>([])
  const [rows, setRows] = useState<PreviewRow[] | null>(null)
  const [result, setResult] = useState<{ inserted_count: number; skipped_count: number } | null>(null)
  const check = useMutation({ mutationFn: async () => { const candidates = parsePaste(text, kind); setItems(candidates); return preview(candidates) }, onSuccess: setRows })
  const create = useMutation({ mutationFn: (skipKnown: boolean) => sendJSON<{ inserted_count: number; skipped_count: number }>(`${base}/ideas/batch`, 'POST', { items, skip_known: skipKnown }), onSuccess: async (value) => { setResult(value); setRows(null); await refresh() } })
  const resetPreview = () => { setRows(null); setResult(null); check.reset(); create.reset() }
  const summary = rows && { total: rows.length, fresh: rows.filter((row) => row.valid && !row.warnings.length).length, invalid: rows.filter((row) => !row.valid).length }
  return <Dialog open onOpenChange={(open) => !open && !check.isPending && !create.isPending && close()} title="批量加入内容想法" description="每行一条，用竖线 | 分隔。每批最多 500 条。" footer={<>
    <Button variant="secondary" disabled={check.isPending || create.isPending} onClick={() => check.mutate()}>预览</Button>
    <Button variant="secondary" disabled={!rows || check.isPending || create.isPending} onClick={() => create.mutate(false)}>全部加入</Button>
    <Button disabled={!rows || check.isPending || create.isPending} onClick={() => create.mutate(true)}>只加入新内容</Button>
  </>}><div className="grid gap-4"><Select ariaLabel="批量类型" options={kindOptions} value={kind} disabled={check.isPending || create.isPending} onValueChange={(value) => { setKind(value as Kind); resetPreview() }} /><div className="rounded-md border bg-surface-muted p-3 text-sm"><p className="font-mono">名称 | 来源 | 备注</p><p className="mt-2 text-xs text-muted-foreground">空格属于内容，不会分列；备注可省略，来源留空时保留竖线。也可每行只填 Steam AppID 或网址。</p><Button className="mt-2" size="sm" variant="secondary" disabled={Boolean(text.trim()) || check.isPending || create.isPending} onClick={() => { setText(batchExamples[kind]); resetPreview() }}>填入示例</Button></div><Textarea aria-label="粘贴内容" placeholder={batchExamples[kind]} rows={7} value={text} disabled={check.isPending || create.isPending} onChange={(e) => { setText(e.target.value); resetPreview() }} />
    <p className="text-xs text-muted-foreground">先预览检查结果。“只加入新内容”跳过重复项；“全部加入”允许重复，均跳过无效行。</p>
    {(check.error || create.error) && <Alert tone="danger"><span className="whitespace-pre-line">{collaborationError(check.error || create.error)}</span></Alert>}
    {result && <Alert tone="info">已加入 {result.inserted_count} 条，跳过 {result.skipped_count} 条。</Alert>}
    {summary && rows && <><p className="text-sm">共 {summary.total} 条 · 新内容 {summary.fresh} · 无效 {summary.invalid}</p><p className="text-xs text-muted-foreground">{Object.entries(warningLabels).map(([code, label]) => `${label} ${rows.filter((row) => row.warnings.includes(code)).length}`).join(' · ')}</p><PreviewRows rows={rows} /></>}
  </div></Dialog>
}

export function LinkIdeaDialog({ idea, close }: { idea: Idea; close: () => void }) {
  const refresh = useRefreshCollaboration()
  const [kind, setKind] = useState<'game' | 'site'>(idea.kind === 'site' ? 'site' : 'game')
  const [selection, setSelection] = useState<OptionItem | null>(null)
  const mutation = useMutation({ mutationFn: () => transition(idea, 'link', { kind, resource_id: Number(selection!.id) }), onSuccess: async () => { await refresh(); close() } })
  return <Dialog open onOpenChange={(open) => !open && !mutation.isPending && close()} title="关联已有正式内容" description="搜索并选择已创建的游戏或网站。关联后，想法将标记为已落地。" footer={<><Button variant="secondary" disabled={mutation.isPending} onClick={close}>取消</Button><Button disabled={!selection || mutation.isPending} onClick={() => mutation.mutate()}>确认关联</Button></>}><div className="grid min-w-0 gap-4">
    {idea.kind === 'other' && <Select ariaLabel="关联类型" options={kindOptions.slice(0, 2)} value={kind} onValueChange={(value) => { setKind(value as 'game' | 'site'); setSelection(null) }} />}
    <RemoteSelect key={kind} resultsLayout="inline" endpoint={`/api/v1/options/${kind === 'game' ? 'games' : 'sites'}`} value={selection} onChange={setSelection} disabled={mutation.isPending} pageSize={10} debounceMs={300} />
    {selection && <p className="truncate text-sm text-muted-foreground" title={selection.label}>已选择：{selection.label}</p>}
    {mutation.error && <Alert tone="warning">{collaborationError(mutation.error)}<Button variant="secondary" onClick={() => { void refresh(); close() }}>重新加载</Button></Alert>}
  </div></Dialog>
}
