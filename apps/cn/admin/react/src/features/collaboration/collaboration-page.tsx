import { useQuery } from '@tanstack/react-query'
import { lazy, Suspense, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { PageLayout, PageHeader } from '../../components/admin/page'
import { LoadingState, ErrorState } from '../../components/admin/states'
import { WorkspaceTabs } from '../../components/admin/workspace'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { getJSON } from '../../lib/api'
import type { PageResult } from '../../lib/types'
import { useAuth } from '../auth/auth-context'
import { base } from './api'
import { IdeaEditorDialog, BatchIdeaDialog } from './idea-dialogs'
import { IdeaTableRow } from './idea-table'
import './collaboration.css'
import { kindOptions, priorityOptions, statusLabels, type Idea, type Inventory } from './types'

const SharedBoard = lazy(() => import('./board').then((module) => ({ default: module.SharedBoard })))

export function CollaborationPage() {
  const [params, setParams] = useSearchParams()
  const tab = params.get('tab') === 'board' ? 'board' : 'ideas'
  return <PageLayout className="min-w-0"><PageHeader title="协作中心" description="储备内容线索，协作整理与落地。" /><WorkspaceTabs tabs={[{ key: 'ideas', label: '内容想法池' }, { key: 'board', label: '共享画板' }]} active={tab} onChange={(next) => { const value = new URLSearchParams(params); value.set('tab', next); setParams(value) }} />{tab === 'board' ? <Suspense fallback={<LoadingState />}><SharedBoard /></Suspense> : <Ideas />}</PageLayout>
}
function Ideas() {
  const auth = useAuth()
  const [params, setParams] = useSearchParams()
  const [dialog, setDialog] = useState<'single' | 'batch' | null>(null)
  const filters = new URLSearchParams({ page_num: params.get('page_num') || '1', page_size: '50', keyword: params.get('keyword') || '', kind: params.get('kind') || '', status: params.get('status') || 'active', priority: params.get('priority') || '', researcher: params.get('researcher') || 'all' })
  const page = Number(filters.get('page_num')) || 1
  const query = useQuery({ queryKey: ['collaboration', 'ideas', filters.toString()], queryFn: () => getJSON<PageResult<Idea>>(`${base}/ideas?${filters}`) })
  const summary = useQuery({ queryKey: ['collaboration', 'summary'], queryFn: () => getJSON<Inventory>(`${base}/summary`) })
  const set = (key: string, value: string) => { const next = new URLSearchParams(params); next.set(key, value); if (key !== 'page_num') next.set('page_num', '1'); setParams(next, { replace: true }) }
  return <div className="idea-inventory grid gap-4">
    <div className="flex flex-wrap items-center justify-between gap-3"><p className="text-sm">{summary.data ? `${summary.data.reserve_count} 条储备 · ${summary.data.researching_count} 条正在整理 · 近 30 天已落地 ${summary.data.landed_30d} 条` : '内容库存'}</p>{auth.can('collaboration.write') && <div className="flex gap-2"><Button variant="secondary" onClick={() => setDialog('single')}>快速加入</Button><Button onClick={() => setDialog('batch')}>批量加入</Button></div>}</div>
    <div className="idea-filters"><Input aria-label="搜索想法" placeholder="搜索名称、来源、备注…" value={filters.get('keyword')!} onChange={(e) => set('keyword', e.target.value)} /><Select ariaLabel="类型筛选" options={[{ value: '', label: '全部类型' }, ...kindOptions]} value={filters.get('kind')!} onValueChange={(value) => set('kind', value)} /><Select ariaLabel="状态筛选" options={[{ value: 'active', label: '储备 + 整理中' }, { value: 'all', label: '全部状态' }, ...Object.entries(statusLabels).map(([value, label]) => ({ value, label }))]} value={filters.get('status')!} onValueChange={(value) => set('status', value)} /><Select ariaLabel="优先级筛选" options={[{ value: '', label: '全部优先级' }, ...priorityOptions]} value={filters.get('priority')!} onValueChange={(value) => set('priority', value)} /><Select ariaLabel="整理人筛选" options={[{ value: 'all', label: '全部整理人' }, { value: 'me', label: '我在整理' }, { value: 'unassigned', label: '未指定整理人' }]} value={filters.get('researcher')!} onValueChange={(value) => set('researcher', value)} /></div>
    {query.isLoading ? <LoadingState /> : query.error ? <ErrorState message={query.error.message} onRetry={() => void query.refetch()} /> : <><div className="min-w-0 overflow-hidden rounded-md border bg-surface"><table className="idea-table text-left text-sm"><thead className="bg-surface-muted text-xs text-muted-foreground"><tr>{[['kind', '类型'], ['title', '名称 / 线索'], ['source', '来源'], ['note', '备注'], ['priority', '优先级'], ['status', '状态'], ['researcher', '整理人'], ['creator', '加入者'], ['created', '加入时间'], ['actions', '操作']].map(([key, label]) => <th className={`idea-${key}`} key={key}>{label}</th>)}</tr></thead><tbody>{query.data?.list.map((idea) => <IdeaTableRow key={idea.id} idea={idea} />)}</tbody></table>{!query.data?.list.length && <p className="p-6 text-sm text-muted-foreground">暂无符合条件的想法。</p>}</div><div className="flex flex-wrap items-center justify-end gap-3 text-sm"><span>共 {query.data?.total ?? 0} 条 · 第 {page} 页</span><Button variant="secondary" disabled={page <= 1} onClick={() => set('page_num', String(page - 1))}>上一页</Button><Button variant="secondary" disabled={page * 50 >= (query.data?.total ?? 0)} onClick={() => set('page_num', String(page + 1))}>下一页</Button></div></>}
    {dialog === 'single' && <IdeaEditorDialog close={() => setDialog(null)} />}{dialog === 'batch' && <BatchIdeaDialog close={() => setDialog(null)} />}
  </div>
}
