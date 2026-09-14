import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowDown, ArrowUp } from '@phosphor-icons/react'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useToast } from '../../app/toast'
import { PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { errorMessage, getJSON, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'

export type CurationSite = { site_id: string; name: string; deleted: boolean }
export type GroupCuration = { name: string; revision: string; sites: CurationSite[] }

export function GroupCurationPage() {
  const { id } = useParams()
  const endpoint = `/api/v1/nav/site-groups/${id}/curation`
  const query = useQuery({ queryKey: ['group-curation', id], queryFn: () => getJSON<GroupCuration>(endpoint) })
  if (query.isLoading) return <LoadingState />
  if (!query.data) return <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} />
  return <GroupCurationEditor key={id} data={query.data} endpoint={endpoint} reload={() => void query.refetch()} />
}

export function GroupCurationEditor({ data, endpoint, reload }: { data: GroupCuration; endpoint: string; reload: () => void }) {
  const client = useQueryClient()
  const { toast } = useToast()
  const canWrite = useAuth().can('content.write')
  // Keep the revision with the draft: background refetches must not silently
  // replace an operator's work or make a stale draft look current.
  const [draft, setDraft] = useState<GroupCuration | null>(null)
  const [error, setError] = useState('')
  const current = draft ?? data
  const sites = current.sites.filter((site) => !site.deleted)
  const inactive = current.sites.filter((site) => site.deleted)
  useUnsavedChanges(draft !== null)
  const mutation = useMutation({
    mutationFn: () => sendJSON<GroupCuration>(endpoint, 'PUT', { revision: current.revision, site_ids: [...sites, ...inactive].map((site) => site.site_id) }),
    onSuccess: async (saved) => {
      client.setQueryData(['group-curation', endpoint.split('/').at(-2)], saved)
      setDraft(null)
      setError('')
      await client.invalidateQueries({ queryKey: ['site'] })
      toast('首页编排已保存，公开首页将在缓存刷新后更新')
      reload()
    },
    onError: (err) => setError(errorMessage(err)),
  })
  const move = (index: number, to: number) => {
    const next = [...sites]
    const [site] = next.splice(index, 1)
    next.splice(to, 0, site)
    setDraft({ ...current, sites: [...next, ...inactive] })
  }
  const rows = (items: CurationSite[], offset: number) => <ol className="divide-y divide-border" start={offset + 1}>{items.map((site, localIndex) => {
    const index = offset + localIndex
    return <li key={site.site_id} className="flex flex-wrap items-center gap-3 py-3">
      <span className="w-6 text-sm tabular-nums text-muted-foreground">{index + 1}</span>
      <Link className="min-w-0 flex-1 break-words text-sm hover:text-primary" to={`/nav/sites/${site.site_id}`}>{site.name}</Link>
      {canWrite && <div className="flex gap-1">
        {index >= 8 && <Button variant="secondary" disabled={mutation.isPending} onClick={() => move(index, 0)}>移至首页首位</Button>}
        <Button variant="ghost" aria-label={`上移 ${site.name}`} disabled={index === 0 || mutation.isPending} onClick={() => move(index, index - 1)}><ArrowUp className="size-4" /></Button>
        <Button variant="ghost" aria-label={`下移 ${site.name}`} disabled={index === sites.length - 1 || mutation.isPending} onClick={() => move(index, index + 1)}><ArrowDown className="size-4" /></Button>
      </div>}
    </li>
  })}</ol>
  return <PageLayout>
    <Link className="text-sm text-muted-foreground hover:text-primary" to="/nav/site-groups">返回网站分组</Link>
    <PageHeader title={`${current.name} · 首页编排`} actions={<div className="flex gap-2"><Button variant="secondary" disabled={mutation.isPending} onClick={() => {
      if (draft && !window.confirm('放弃未保存的编排并重新加载？')) return
      setDraft(null); setError(''); reload()
    }}>重新加载</Button>{canWrite && <Button disabled={!draft || mutation.isPending} onClick={() => mutation.mutate()}>保存编排</Button>}</div>} />
    {error && <Alert tone="danger">{error}</Alert>}
    <p className="text-sm text-muted-foreground">上下移动站点调整顺序。前 8 个展示在首页，其余站点仍保留分组关系。保存后等待公开首页缓存刷新。</p>
    <Section title={`首页展示（${Math.min(sites.length, 8)}/8）`}>{sites.length ? rows(sites.slice(0, 8), 0) : <p className="text-sm text-muted-foreground">暂无可展示站点，请先在站点详情中加入此分组。</p>}</Section>
    <Section title={`其余站点（${Math.max(0, sites.length - 8)}）`}>{sites.length > 8 ? rows(sites.slice(8), 8) : <p className="text-sm text-muted-foreground">该分组全部可用站点均已进入首页预览。</p>}</Section>
    {inactive.length > 0 && <Section title={`已停用站点（${inactive.length}）`}><p className="text-sm text-muted-foreground">保留分组关系，不参与首页展示。</p><ul>{inactive.map((site) => <li className="py-2 text-sm" key={site.site_id}>{site.name}</li>)}</ul></Section>}
  </PageLayout>
}
