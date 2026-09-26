import { useMutation } from '@tanstack/react-query'
import { Link, useSearchParams } from 'react-router-dom'
import { useState } from 'react'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction } from '../../components/ui/dialog'
import { useAuth } from '../auth/auth-context'
import { collaborationError, sourceURL, transition, useIdea, useRefreshCollaboration } from './api'
import { statusLabels, type Idea } from './types'

export function ideaWorkspacePath(kind: 'game' | 'site', id: number | 'new', ideaID: string | number | null) {
  return `${kind === 'game' ? '/game/games' : '/nav/sites'}/${id}${ideaID ? `?idea=${encodeURIComponent(ideaID)}` : ''}`
}
export function reliableAppID(idea?: Idea) { return idea?.kind === 'game' && /^steam:[1-9]\d*$/.test(idea.source_key ?? '') ? Number(idea!.source_key!.slice(6)) : undefined }

// Called only after the formal API has returned success. Failure is a separate
// recoverable outcome, never propagated into the content creation mutation.
export async function linkCreatedIdea(idea: Idea | undefined, ideaID: string | null, kind: 'game' | 'site', resourceID: number) {
  if (!ideaID) return null
  try {
    if (!idea) throw new Error('想法上下文尚未加载')
    await transition(idea, 'link', { kind, resource_id: resourceID })
    return null
  } catch (error) { return `正式内容已创建，但想法关联未完成。请在此页面重新加载后使用“关联到当前内容”。${collaborationError(error)}` }
}

export function IdeaContextBanner({ kind, resourceID }: { kind: 'game' | 'site'; resourceID?: number }) {
  const [params] = useSearchParams()
  const auth = useAuth()
  const query = useIdea(params.get('idea'), auth.can('collaboration.read'))
  const refresh = useRefreshCollaboration()
  const [confirmRelease, setConfirmRelease] = useState(false)
  const mutation = useMutation({ mutationFn: (action: string) => transition(query.data!, action, action === 'link' ? { kind, resource_id: resourceID } : {}), onSuccess: async () => { setConfirmRelease(false); await refresh() } })
  if (!params.get('idea') || !auth.can('collaboration.read')) return null
  if (query.error) return <Alert tone="warning">想法上下文加载失败。<Button variant="secondary" onClick={() => void query.refetch()}>重新加载想法</Button></Alert>
  if (!query.data) return <p>正在加载想法上下文…</p>
  const idea = query.data
  const active = idea.status === 'idea' || idea.status === 'researching'
  const canLink = active && (idea.kind === kind || idea.kind === 'other')
  const url = sourceURL(idea.source, idea.source_key)
  return <section className="grid gap-2 rounded-md border bg-surface-muted p-4" aria-label="想法上下文">
    <div className="flex flex-wrap items-center gap-3"><Link className="font-medium text-primary" to="/collaboration">来自内容想法池</Link><span>{idea.title || idea.source}</span><span>{statusLabels[idea.status]}</span></div>
    {url && <a href={url} target="_blank" rel="noreferrer" className="break-all text-primary">打开来源：{idea.source}</a>}
    <p className="whitespace-pre-wrap text-sm">{idea.note}</p><p className="text-xs text-muted-foreground">加入者：{idea.creator_name} · 整理人：{idea.researcher_name || '未指定'}</p>
    {resourceID && active && <Alert tone="info">正式内容已存在；想法尚未关联，可以关联到当前内容。</Alert>}
    {mutation.error && <Alert tone="warning">{collaborationError(mutation.error)} <Button variant="secondary" onClick={() => { mutation.reset(); void query.refetch() }}>重新加载</Button></Alert>}
    {auth.can('collaboration.write') && <div className="flex gap-2">
      {resourceID && canLink && <Button disabled={mutation.isPending} onClick={() => mutation.mutate('link')}>关联到当前内容</Button>}
      {idea.status === 'researching' && <Button variant="secondary" disabled={mutation.isPending} onClick={() => idea.researching_by_account_id === auth.state?.identity?.account_id ? mutation.mutate('release') : setConfirmRelease(true)}>释放整理</Button>}
    </div>}
    <ConfirmAction open={confirmRelease} onOpenChange={setConfirmRelease} title="释放其他成员的整理状态" description={`当前由 ${idea.researcher_name} 整理，释放会记录审计。`} confirmLabel="确认释放" variant="primary" onConfirm={() => mutation.mutate('release')} busy={mutation.isPending} />
  </section>
}
