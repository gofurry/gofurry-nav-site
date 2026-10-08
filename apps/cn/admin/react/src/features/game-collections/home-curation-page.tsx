import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../../app/toast'
import { RemoteSelect } from '../../components/admin/operations'
import { PageHeader, PageLayout, Section } from '../../components/admin/page'
import { LoadingState, ErrorState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { ApiError, errorMessage, getJSON, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { collectionEndpoint, eligibleCollectionEndpoint, homeKey, loadEligibleCollections } from './api'
import type { CollectionHome, GameCollection } from './types'

export function CollectionHomeCurationPage() {
  const query = useQuery({ queryKey: homeKey, queryFn: () => getJSON<CollectionHome>(`${collectionEndpoint}/home-curation`) })
  if (query.isLoading) return <LoadingState />
  if (!query.data) return <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} />
  return <CollectionHomeEditor data={query.data} reload={async () => { const result = await query.refetch(); if (result.error) throw result.error; return result.data! }} />
}
export function CollectionHomeEditor({ data, reload }: { data: CollectionHome; reload: () => Promise<CollectionHome> }) {
  const canWrite = useAuth().can('content.write')
  const client = useQueryClient()
  const { toast } = useToast()
  const [draft, setDraft] = useState<CollectionHome | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [conflict, setConflict] = useState(false)
  const current = draft ?? data
  useUnsavedChanges(draft !== null)
  const mutation = useMutation({
    retry: false,
    mutationFn: () => sendJSON<CollectionHome>(`${collectionEndpoint}/home-curation`, 'PUT', { revision: current.revision, slots: current.slots.map(s => ({ slot: s.slot, collection_id: s.collection?.id ?? null })) }),
    onMutate: () => client.cancelQueries({ queryKey: homeKey }),
    onSuccess: saved => { client.setQueryData(homeKey, saved); setDraft(null); setError(''); setConflict(false); toast('已保存'); void client.invalidateQueries({ queryKey: ['game-collections'] }) },
    onError: err => { setError(errorMessage(err)); setConflict(err instanceof ApiError && err.status === 409) },
  })
  const busy = loading || mutation.isPending
  const select = (slot: number, collection: GameCollection | null) => {
    if (collection && current.slots.some(s => s.slot !== slot && s.collection?.id === collection.id)) { setError('同一分区不可占据多个入口'); return }
    setDraft({ ...current, slots: current.slots.map(s => s.slot === slot ? { slot, collection } : s) })
  }
  const reloadExplicitly = async () => {
    if (draft && !window.confirm('放弃未保存的首页入口编排并重新加载？')) return
    setLoading(true)
    try { const fresh = await reload(); client.setQueryData(homeKey, fresh); setDraft(null); setError(''); setConflict(false) }
    catch (err) { setError(errorMessage(err)) }
    finally { setLoading(false) }
  }
  return <PageLayout>
    <Link to="/game/collections">返回游戏分区</Link>
    <PageHeader title="首页入口编排" actions={<><Button variant="secondary" disabled={busy} onClick={() => void reloadExplicitly()}>{draft ? '放弃修改并重新加载' : '重新加载'}</Button>{canWrite && <Button disabled={!draft || busy || conflict} onClick={() => mutation.mutate()}>保存编排</Button>}</>} />
    <p className="text-sm text-muted-foreground">仅可选择已发布且至少有一个 SFW 可见游戏的分区。</p>
    {error && <Alert tone="danger">{error}{conflict && ' 编排草稿已保留，请重新加载。'}</Alert>}
    {current.slots.map(slot => <Section key={slot.slot} title={`#${slot.slot}`} actions={canWrite && <Button variant="ghost" disabled={busy || !slot.collection} onClick={() => select(slot.slot, null)}>清空第 {slot.slot} 位</Button>}>
      {slot.collection ? <Link className="text-primary" to={`/game/collections/${slot.collection.id}`}>{slot.collection.name}</Link> : <p className="text-sm text-muted-foreground">空位</p>}
      {slot.collection && (slot.collection.status !== 'published' || slot.collection.sfw_member_count === 0) && <p className="mt-2 text-sm text-warning">已配置 · 当前不符合展示条件。位置保留，可更换或清空。</p>}
      {canWrite && <div className="mt-3"><RemoteSelect endpoint={eligibleCollectionEndpoint} loadOptions={loadEligibleCollections} value={slot.collection ? { id: String(slot.collection.id), label: slot.collection.name } : null} disabled={busy} debounceMs={300} placeholder={`选择或更换第 ${slot.slot} 位分区…`} excludeIDs={current.slots.filter(s => s.slot !== slot.slot && s.collection).map(s => String(s.collection!.id))} onChange={async option => {
        if (!option) { select(slot.slot, null); return }
        setLoading(true); setError('')
        try { select(slot.slot, await getJSON<GameCollection>(`${collectionEndpoint}/${option.id}`)) }
        catch (err) { setError(errorMessage(err)) }
        finally { setLoading(false) }
      }} /></div>}
    </Section>)}
    <Section title="#6 全部分区"><p className="text-sm text-muted-foreground">固定入口 · 无需配置</p></Section>
  </PageLayout>
}
