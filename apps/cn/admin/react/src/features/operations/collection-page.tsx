import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Activity, Ban, LoaderCircle, Play, RefreshCw, Settings2 } from 'lucide-react'
import { useMemo, useState, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useToast } from '../../app/toast'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FilterField, JsonBlock, Kpi, KpiGrid, OperationTabs, OperationsChart, RemoteSelect, type OperationTab } from '../../components/admin/operations'
import { Detail, DetailGrid, FormField, ListPanel, ListRow, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction, Dialog } from '../../components/ui/dialog'
import { DateTimePicker } from '../../components/ui/date-picker'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { Sheet } from '../../components/ui/sheet'
import { errorMessage, getJSON, sendJSON } from '../../lib/api'
import type { OptionItem, PageResult } from '../../lib/types'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'
import { collectionActionVisibility } from './capability-ux'
import { queryString, toRFC3339 } from './collection-query'
import { ACTIVE_COLLECTION_REFRESH_MS, collectionJobProgress, duration, percent, runCoverage, statusTone } from './presentation'
import type { CollectionChartPoint, CollectionJob, CollectionOverview, CollectionResult, CollectionRun, CollectionSchedule, CollectorInstance } from './types'

const tabs: OperationTab[] = [
  { key: 'overview', label: '概览' }, { key: 'schedules', label: '调度' }, { key: 'running', label: '运行中' },
  { key: 'history', label: '运行历史' }, { key: 'manual', label: '手动采集', capability: 'collection.execute' },
]

export function CollectionPage() {
  const auth = useAuth()
  const [params, setParams] = useSearchParams()
  const active = tabs.some((tab) => tab.key === params.get('tab')) ? params.get('tab')! : 'overview'
  const setTab = (tab: string) => { const next = new URLSearchParams(); if (tab !== 'overview') next.set('tab', tab); setParams(next, { replace: true }) }
  return <PageLayout><PageHeader title="采集中心" /><OperationTabs tabs={tabs} active={active} onChange={setTab} can={auth.can} />{active === 'overview' && <CollectionOverviewPanel />}{active === 'schedules' && <SchedulesPanel />}{active === 'running' && <RunningPanel />}{active === 'history' && <RunHistoryPanel />}{active === 'manual' && <ManualCollectionPanel />}</PageLayout>
}

function CollectionOverviewPanel() {
  const [instanceView, setInstanceView] = useState<'current' | 'history'>('current')
  const [chartWindow, setChartWindow] = useState<'24h' | '7d' | '30d'>('7d')
  const [chartDomain, setChartDomain] = useState('')
  const [chartJobKey, setChartJobKey] = useState('')
  const overview = useQuery({ queryKey: ['collection', 'overview'], queryFn: () => getJSON<CollectionOverview>('/api/v1/collection/overview') })
  const instances = useQuery({ queryKey: ['collection', 'instances', instanceView], queryFn: () => getJSON<PageResult<CollectorInstance>>(`/api/v1/collection/instances?view=${instanceView}&page=1&page_size=50`) })
  const failed = useQuery({ queryKey: ['collection', 'failed-runs'], queryFn: () => getJSON<PageResult<CollectionRun>>('/api/v1/collection/runs?status=failed&page=1&page_size=5') })
  const charts = useQuery({ queryKey: ['collection', 'charts', chartWindow, chartDomain, chartJobKey], queryFn: () => getJSON<CollectionChartPoint[]>(`/api/v1/collection/charts/outcomes?${queryString({ window: chartWindow, domain: chartDomain, job_key: chartJobKey })}`) })
  if (overview.isLoading) return <LoadingState />
  if (overview.error || !overview.data) return <ErrorState message={overview.error?.message ?? '采集概览不可用'} onRetry={() => void overview.refetch()} />
  const unhealthy = (instances.data?.list ?? []).filter((instance) => instance.health !== 'healthy')
  const points = charts.data ?? []
  const instanceActions = <div className="flex gap-1"><Button size="sm" variant={instanceView === 'current' ? 'primary' : 'secondary'} onClick={() => setInstanceView('current')}>Current</Button><Button size="sm" variant={instanceView === 'history' ? 'primary' : 'secondary'} onClick={() => setInstanceView('history')}>Historical</Button></div>
  return <div className="grid gap-4">
    <KpiGrid><Kpi label="Collectors" value={instances.data?.total ?? '—'} detail={instanceView === 'current' ? '当前逻辑实例' : '历史生命周期'} tone={unhealthy.length ? 'warning' : 'success'} /><Kpi label="Running" value={overview.data.running_count} tone={overview.data.running_count ? 'info' : 'neutral'} /><Kpi label="Queued" value={overview.data.queued_count} tone={overview.data.queued_count ? 'info' : 'neutral'} /><Kpi label="Failed 24h" value={overview.data.failed_24h} tone={overview.data.failed_24h ? 'danger' : 'success'} /><Kpi label="Missed 24h" value={overview.data.missed_24h} tone={overview.data.missed_24h ? 'warning' : 'success'} /></KpiGrid>
    <Section title="采集趋势"><div className="grid gap-4"><div className="flex flex-wrap items-end gap-3"><FilterField label="图表窗口"><Select value={chartWindow} onValueChange={(value) => setChartWindow(value as '24h' | '7d' | '30d')} options={[{ value: '24h', label: '24 小时' }, { value: '7d', label: '7 天' }, { value: '30d', label: '30 天' }]} /></FilterField><FilterField label="Domain"><Select value={chartDomain} onValueChange={setChartDomain} options={[{ value: '', label: '全部' }, { value: 'game', label: 'Game' }, { value: 'nav', label: 'Nav' }]} /></FilterField><FilterField label="Job Key"><Input value={chartJobKey} onChange={(event) => setChartJobKey(event.target.value)} placeholder="可选 job_key" /></FilterField></div><div className="grid divide-y border-t xl:grid-cols-3 xl:divide-x xl:divide-y-0"><ChartPane title="Outcome" detail="成功结果数" ><OperationsChart className="h-52" points={points.map((point) => ({ label: formatDate(point.created_at), value: point.success }))} unit="" label="采集成功数趋势" loading={charts.isLoading} error={charts.error?.message} /></ChartPane><ChartPane title="Coverage" detail="成功 / 应采集"><OperationsChart className="h-52" points={points.map((point) => ({ label: formatDate(point.created_at), value: point.coverage == null ? null : Number((point.coverage * 100).toFixed(1)) }))} unit="%" label="采集覆盖率趋势" loading={charts.isLoading} error={charts.error?.message} /></ChartPane><ChartPane title="Timing" detail="运行时长"><OperationsChart className="h-52" points={points.map((point) => ({ label: formatDate(point.created_at), value: Number((point.duration_ms / 1000).toFixed(1)) }))} unit="s" label="采集时长趋势" loading={charts.isLoading} error={charts.error?.message} /></ChartPane></div></div></Section>
    <div className="grid gap-4 xl:grid-cols-2"><ListPanel title="需要关注">{overview.data.failed_24h === 0 && overview.data.missed_24h === 0 && unhealthy.length === 0 ? <ListRow><p className="text-sm text-muted-foreground">当前没有需要关注的采集状态。</p></ListRow> : <>{overview.data.failed_24h > 0 && <Attention tone="danger" title="最近失败" detail={`${overview.data.failed_24h} 个任务在 24 小时内失败`} />}{overview.data.missed_24h > 0 && <Attention tone="warning" title="错过调度" detail={`${overview.data.missed_24h} 个计划时隙未执行`} />}{unhealthy.map((instance) => <Attention key={instance.instance_id} tone="warning" title={`${instance.domain} Collector 异常`} detail={`${instance.hostname} · 心跳 ${instance.heartbeat_age_seconds}s`} />)}</>}</ListPanel><ListPanel title="最近失败">{failed.isLoading ? <ListRow><p className="text-sm text-muted-foreground">加载中…</p></ListRow> : (failed.data?.list ?? []).length === 0 ? <ListRow><p className="text-sm text-muted-foreground">暂无失败记录。</p></ListRow> : failed.data!.list.map((run) => <ListRow key={`${run.domain}-${run.id}`} className="flex items-center justify-between gap-3"><div className="min-w-0"><p className="text-sm font-medium">{run.job_key}</p><p className="mt-0.5 truncate text-xs text-muted-foreground">{formatDate(run.started_at)} · {run.error_message || run.error_kind || '未记录原因'}</p></div><StatusBadge tone="danger">{run.status}</StatusBadge></ListRow>)}</ListPanel></div>
    <InstanceTable page={instances.data} loading={instances.isLoading} error={instances.error?.message} actions={instanceActions} />
  </div>
}

function ChartPane({ title, detail, children }: { title: string; detail: string; children: ReactNode }) { return <div className="min-w-0 px-3 pt-4 first:pl-0 last:pr-0"><div className="flex items-baseline justify-between gap-2 px-2"><h3 className="text-sm font-semibold">{title}</h3><span className="text-xs text-muted-foreground">{detail}</span></div>{children}</div> }

function Attention({ tone, title, detail }: { tone: 'warning' | 'danger'; title: string; detail: string }) { return <ListRow className="flex items-center justify-between gap-3"><div><p className="text-sm font-medium">{title}</p><p className="mt-0.5 text-xs text-muted-foreground">{detail}</p></div><StatusBadge tone={tone}>{tone === 'danger' ? '需处理' : '需关注'}</StatusBadge></ListRow> }

function InstanceTable({ page, loading, error, actions }: { page?: PageResult<CollectorInstance>; loading?: boolean; error?: string; actions: ReactNode }) {
  const columns = useMemo<AdminColumn<CollectorInstance>[]>(() => [
    { key: 'collector_id', header: 'Collector', render: (row) => <div><p className="font-medium">{row.domain} · {row.collector_id}</p><p className="font-mono text-xs text-muted-foreground">{row.hostname}</p></div> },
    { key: 'health', header: '状态', render: (row) => <StatusBadge tone={statusTone(row.health)}>{row.health}</StatusBadge> },
    { key: 'version', header: '版本', render: (row) => <div><p>{row.version || '—'}</p><TechnicalLabel>{row.commit_sha || '—'}</TechnicalLabel></div> },
    { key: 'last_heartbeat_at', header: '最近心跳', render: (row) => <div><p>{formatDate(row.last_heartbeat_at)}</p><p className="text-xs text-muted-foreground">{row.heartbeat_age_seconds}s ago</p></div> },
  ], [])
  return <DataTable title="Collector 状态" headerActions={actions} data={page?.list ?? []} columns={columns} total={page?.total ?? 0} page={1} pageSize={50} search="" onSearchChange={() => undefined} onPageChange={() => undefined} onPageSizeChange={() => undefined} searchable={false} loading={loading} error={error} />
}

function schedulePayload(schedule: CollectionSchedule, enabled = schedule.enabled) {
  return { enabled, schedule_kind: schedule.schedule_kind, cron_expression: schedule.cron_expression ?? null, interval_seconds: schedule.interval_seconds ?? null, anchor_at: schedule.anchor_at ?? null, timezone: schedule.timezone, misfire_policy: schedule.misfire_policy, misfire_grace_seconds: schedule.misfire_grace_seconds }
}

function SchedulesPanel() {
  const auth = useAuth()
  const access = collectionActionVisibility(auth.can)
  const { toast } = useToast()
  const client = useQueryClient()
  const [editing, setEditing] = useState<CollectionSchedule | null>(null)
  const [toggling, setToggling] = useState<CollectionSchedule | null>(null)
  const query = useQuery({ queryKey: ['collection', 'schedules'], queryFn: () => getJSON<CollectionSchedule[]>('/api/v1/collection/schedules') })
  const mutation = useMutation({ mutationFn: ({ action, schedule }: { action: 'run' | 'toggle'; schedule: CollectionSchedule }) => action === 'run' ? sendJSON(`/api/v1/collection/schedules/${schedule.domain}/${schedule.id}/run`, 'POST') : sendJSON(`/api/v1/collection/schedules/${schedule.domain}/${schedule.id}`, 'PUT', schedulePayload(schedule, !schedule.enabled)), onSuccess: async (_, variables) => { await client.invalidateQueries({ queryKey: ['collection'] }); toast(variables.action === 'run' ? '已创建手动运行，调度相位保持不变' : '调度状态已更新'); setToggling(null) }, onError: (error) => toast(errorMessage(error), 'danger') })
  if (query.isLoading) return <LoadingState />
  if (query.error) return <ErrorState message={query.error.message} onRetry={() => void query.refetch()} />
  return <><ListPanel title="采集调度">{query.data?.map((schedule) => <ListRow key={`${schedule.domain}-${schedule.id}`} className="grid gap-3 xl:grid-cols-[1.4fr_1fr_1fr_auto] xl:items-center"><div><div className="flex items-center gap-2"><p className="font-medium">{schedule.name}</p><StatusBadge tone={schedule.enabled ? 'success' : 'neutral'}>{schedule.enabled ? '已启用' : '已停用'}</StatusBadge></div><p className="mt-0.5 font-mono text-xs text-muted-foreground">{schedule.domain} · {schedule.job_key} · v{schedule.version}</p></div><div className="text-sm"><p>{schedule.schedule_kind === 'cron' ? schedule.cron_expression : `${schedule.interval_seconds ?? 0}s`}</p><p className="text-xs text-muted-foreground">下次 {formatDate(schedule.next_scheduled_for)}</p></div><div className="text-sm"><div className="flex items-center gap-2"><StatusBadge tone={statusTone(schedule.last_status)}>{schedule.last_status || '无记录'}</StatusBadge><span>{percent(schedule.last_success_coverage)}</span></div><p className="mt-0.5 text-xs text-muted-foreground">{schedule.last_success_count}/{schedule.last_expected_count} coverage</p></div><div className="flex flex-wrap justify-end gap-2">{access.runNow && <Button size="sm" disabled={mutation.isPending} onClick={() => mutation.mutate({ action: 'run', schedule })}><Play className="size-3.5" />Run Now</Button>}{access.control && <><Button size="sm" variant="secondary" onClick={() => setEditing(schedule)}><Settings2 className="size-3.5" />编辑</Button><Button size="sm" variant="secondary" disabled={mutation.isPending} onClick={() => setToggling(schedule)}>{schedule.enabled ? '停用' : '启用'}</Button></>}</div></ListRow>)}</ListPanel>{editing && <ScheduleDialog key={`${editing.domain}-${editing.id}`} schedule={editing} onOpenChange={(open) => { if (!open) setEditing(null) }} />}<ConfirmAction open={Boolean(toggling)} onOpenChange={(open) => { if (!open) setToggling(null) }} title={toggling?.enabled ? '停用采集调度' : '启用采集调度'} description={toggling?.enabled ? `${toggling?.name ?? ''} 停用后不会再创建新的计划任务，现有任务不受影响。` : `${toggling?.name ?? ''} 启用后将按当前调度契约创建计划任务。`} busy={mutation.isPending} confirmLabel={toggling?.enabled ? '确认停用' : '确认启用'} variant={toggling?.enabled ? 'danger' : 'primary'} onConfirm={() => toggling && mutation.mutate({ action: 'toggle', schedule: toggling })} /></>
}

function ScheduleDialog({ schedule, onOpenChange }: { schedule: CollectionSchedule; onOpenChange: (open: boolean) => void }) {
  const client = useQueryClient()
  const { toast } = useToast()
  const [draft, setDraft] = useState<CollectionSchedule>(schedule)
  const mutation = useMutation({ mutationFn: () => sendJSON(`/api/v1/collection/schedules/${draft!.domain}/${draft!.id}`, 'PUT', schedulePayload(draft!)), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['collection', 'schedules'] }); toast('调度已更新'); onOpenChange(false) }, onError: (error) => toast(errorMessage(error), 'danger') })
  return <Dialog open onOpenChange={onOpenChange} title="编辑调度" description={schedule.name} footer={<><Button variant="secondary" onClick={() => onOpenChange(false)}>取消</Button><Button disabled={mutation.isPending} onClick={() => mutation.mutate()}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}保存</Button></>}><div className="grid gap-4"><FormField label="类型"><Select value={draft.schedule_kind} onValueChange={(value) => setDraft({ ...draft, schedule_kind: value })} options={[{ value: 'interval', label: 'Interval' }, { value: 'cron', label: 'Cron' }]} /></FormField>{draft.schedule_kind === 'cron' ? <FormField label="Cron"><Input value={draft.cron_expression ?? ''} onChange={(event) => setDraft({ ...draft, cron_expression: event.target.value })} /></FormField> : <FormField label="间隔（秒）"><Input type="number" value={draft.interval_seconds ?? 60} onChange={(event) => setDraft({ ...draft, interval_seconds: Number(event.target.value) })} /></FormField>}<FormField label="时区"><Input value={draft.timezone} onChange={(event) => setDraft({ ...draft, timezone: event.target.value })} /></FormField><FormField label="Misfire Policy"><Select value={draft.misfire_policy} onValueChange={(value) => setDraft({ ...draft, misfire_policy: value })} options={[{ value: 'skip', label: 'skip' }, { value: 'catch_up_once', label: 'catch_up_once' }]} /></FormField></div></Dialog>
}

function RunningPanel() {
  const auth = useAuth()
  const access = collectionActionVisibility(auth.can)
  const client = useQueryClient()
  const { toast } = useToast()
  const [cancelling, setCancelling] = useState<CollectionJob | null>(null)
  const query = useQuery({ queryKey: ['collection', 'active-jobs'], queryFn: async () => { const [running, queued] = await Promise.all([getJSON<CollectionJob[]>('/api/v1/collection/jobs?status=running&limit=100'), getJSON<CollectionJob[]>('/api/v1/collection/jobs?status=queued&limit=100')]); return [...running, ...queued] }, refetchInterval: ACTIVE_COLLECTION_REFRESH_MS })
  const cancel = useMutation({ mutationFn: (job: CollectionJob) => sendJSON(`/api/v1/collection/jobs/${job.domain}/${job.id}/cancel`, 'POST'), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['collection'] }); toast('已请求取消任务'); setCancelling(null) }, onError: (error) => toast(errorMessage(error), 'danger') })
  const columns = useMemo<AdminColumn<CollectionJob>[]>(() => [
    { key: 'job_key', header: '任务', render: (row) => <div><p className="font-medium">{row.job_key}</p><TechnicalLabel>{row.domain} · job #{row.id}</TechnicalLabel></div> },
    { key: 'scope_type', header: '范围', render: (row) => `${row.scope_type}${row.scope_id ? ` #${row.scope_id}` : ''}${row.target ? ` · ${row.target}` : ''}` },
    { key: 'status', header: '状态', render: (row) => <StatusBadge tone={statusTone(row.status)}>{row.status}</StatusBadge> },
    { key: 'progress', header: '进度', render: (row) => <CollectionProgressCell job={row} /> },
    { key: 'created_at', header: '创建 / 开始', render: (row) => formatDate(row.created_at) },
    { key: '_control', header: '', sortable: false, render: (row) => access.control ? <Button size="sm" variant="secondary" disabled={cancel.isPending} onClick={(event) => { event.stopPropagation(); setCancelling(row) }}><Ban className="size-3.5" />取消</Button> : null },
  ], [access.control, cancel])
  return <><DataTable data={query.data ?? []} columns={columns} total={query.data?.length ?? 0} page={1} pageSize={100} search="" onSearchChange={() => undefined} onPageChange={() => undefined} onPageSizeChange={() => undefined} searchable={false} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /><ConfirmAction open={Boolean(cancelling)} onOpenChange={(open) => { if (!open) setCancelling(null) }} title="取消采集任务" description={`任务 ${cancelling?.job_key ?? ''} #${cancelling?.id ?? ''} 将收到取消请求；已经完成的结果不会回滚。`} busy={cancel.isPending} confirmLabel="确认取消任务" onConfirm={() => cancelling && cancel.mutate(cancelling)} /></>
}

function CollectionProgressCell({ job }: { job: CollectionJob }) {
  const progress = collectionJobProgress(job.status, job.progress)
  if (progress.state === 'queued') return <span className="text-sm text-muted-foreground">排队中</span>
  if (progress.state === 'waiting') return <span className="text-sm text-muted-foreground">等待进度</span>
  return <div className="w-40" aria-label={`进度 ${progress.percentage}% · ${progress.attempted} / ${progress.expected}`}><div className="mb-1 flex items-center justify-between text-xs"><span>{progress.percentage}%</span><span className="text-muted-foreground">{progress.attempted} / {progress.expected}</span></div><div className="h-2 overflow-hidden rounded-full bg-surface-muted" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={progress.percentage}><div className="h-full rounded-full bg-primary transition-[width]" style={{ width: `${progress.percentage}%` }} /></div></div>
}

function RunHistoryPanel() {
  const [params, setParams] = useSearchParams()
  const page = Math.max(1, Number(params.get('page') || 1)); const pageSize = [20, 50, 100].includes(Number(params.get('page_size'))) ? Number(params.get('page_size')) : 20
  const domain = params.get('domain') ?? ''; const status = params.get('status') ?? ''; const trigger = params.get('trigger') ?? ''; const jobKey = params.get('job_key') ?? ''; const since = params.get('since') ?? ''; const until = params.get('until') ?? ''
  const [selected, setSelected] = useState<CollectionRun | null>(null)
  const query = useQuery({ queryKey: ['collection', 'runs', page, pageSize, domain, status, trigger, jobKey, since, until], queryFn: () => getJSON<PageResult<CollectionRun>>(`/api/v1/collection/runs?${queryString({ page, page_size: pageSize, domain, status, trigger, job_key: jobKey, since: toRFC3339(since), until: toRFC3339(until) })}`) })
  const set = (key: string, value: string) => { const next = new URLSearchParams(params); next.set('tab', 'history'); next.set('page', '1'); if (value) next.set(key, value); else next.delete(key); setParams(next, { replace: true }) }
  const columns = useMemo<AdminColumn<CollectionRun>[]>(() => [
    { key: 'started_at', header: 'Started', render: (row) => formatDate(row.started_at) },
    { key: 'job_key', header: 'Domain / Job', render: (row) => <div><p className="font-medium text-primary">{row.job_key}</p><TechnicalLabel>{row.domain} · {row.id}</TechnicalLabel></div> },
    { key: 'trigger', header: 'Trigger' }, { key: 'status', header: 'Status', render: (row) => <StatusBadge tone={statusTone(row.status)}>{row.status}</StatusBadge> },
    { key: 'coverage', header: 'Coverage', render: (row) => percent(runCoverage(row)) }, { key: 'duration_ms', header: 'Duration', render: (row) => duration(row.duration_ms) },
    { key: 'collector_instance_id', header: 'Collector', render: (row) => <TechnicalLabel>{row.collector_instance_id}</TechnicalLabel> },
  ], [])
  const toolbar = <><Select className="w-32" value={domain} onValueChange={(value) => set('domain', value)} options={[{ value: '', label: '全部域' }, { value: 'game', label: 'Game' }, { value: 'nav', label: 'Nav' }]} ariaLabel="运行域" /><Select className="w-32" value={status} onValueChange={(value) => set('status', value)} options={['', 'success', 'failed', 'partial', 'canceled', 'running'].map((value) => ({ value, label: value || '全部状态' }))} ariaLabel="运行状态" /><Select className="w-32" value={trigger} onValueChange={(value) => set('trigger', value)} options={['', 'scheduled', 'manual', 'retry'].map((value) => ({ value, label: value || '全部触发' }))} ariaLabel="触发方式" /><Input className="w-44" value={jobKey} onChange={(event) => set('job_key', event.target.value)} placeholder="job_key" /><div className="w-48"><DateTimePicker value={since} onValueChange={(value) => set('since', value)} ariaLabel="开始时间" /></div><div className="w-48"><DateTimePicker value={until} onValueChange={(value) => set('until', value)} ariaLabel="结束时间" /></div></>
  return <><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={pageSize} search="" onSearchChange={() => undefined} onPageChange={(value) => set('page', String(value))} onPageSizeChange={(value) => set('page_size', String(value))} onRowClick={setSelected} searchable={false} toolbar={toolbar} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /><RunDrawer key={selected ? `${selected.domain}-${selected.id}` : 'closed'} run={selected} onOpenChange={(open) => { if (!open) setSelected(null) }} /></>
}

function RunDrawer({ run, onOpenChange }: { run: CollectionRun | null; onOpenChange: (open: boolean) => void }) {
  const auth = useAuth(); const client = useQueryClient(); const { toast } = useToast(); const [page, setPage] = useState(1); const [pageSize, setPageSize] = useState(20); const [filters, setFilters] = useState({ game_id: '', appid: '', site_id: '', target: '', protocol: '' })
  const resultQuery = queryString({ page, page_size: pageSize, game_id: filters.game_id, appid: filters.appid, site_id: filters.site_id, target: filters.target, protocol: filters.protocol })
  const results = useQuery({ queryKey: ['collection', 'run-results', run?.domain, run?.id, page, pageSize, filters], queryFn: () => getJSON<PageResult<CollectionResult>>(`/api/v1/collection/runs/${run!.domain}/${run!.id}/results?${resultQuery}`), enabled: Boolean(run) })
  const retry = useMutation({ mutationFn: () => sendJSON(`/api/v1/collection/jobs/${run!.domain}/${run!.job_id}/retry`, 'POST'), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['collection'] }); toast('已创建重试任务') }, onError: (error) => toast(errorMessage(error), 'danger') })
  const columns = useMemo<AdminColumn<CollectionResult>[]>(() => [{ key: 'task', header: '任务' }, { key: 'entity_id', header: '实体', render: (row) => row.target || row.appid || row.entity_id }, { key: 'status', header: '状态', render: (row) => <StatusBadge tone={statusTone(row.status)}>{row.status}</StatusBadge> }, { key: 'duration_ms', header: '时长', render: (row) => duration(row.duration_ms) }, { key: 'error_message', header: '错误', render: (row) => row.error_message || '—' }], [])
  const setFilter = (key: keyof typeof filters, value: string) => { setFilters((current) => ({ ...current, [key]: value })); setPage(1) }
  const toolbar = run?.domain === 'game' ? <><Input className="w-32" value={filters.game_id} onChange={(event) => setFilter('game_id', event.target.value)} placeholder="game_id" /><Input className="w-32" value={filters.appid} onChange={(event) => setFilter('appid', event.target.value)} placeholder="appid" /></> : <><Input className="w-32" value={filters.site_id} onChange={(event) => setFilter('site_id', event.target.value)} placeholder="site_id" /><Input className="w-44" value={filters.target} onChange={(event) => setFilter('target', event.target.value)} placeholder="target" /><Input className="w-36" value={filters.protocol} onChange={(event) => setFilter('protocol', event.target.value)} placeholder="protocol" /></>
  return <Sheet open={Boolean(run)} onOpenChange={onOpenChange} title={`运行详情 · ${run?.job_key ?? ''}`} description={run?.id}><div className="grid gap-5">{run && <><Section title="摘要" actions={auth.can('collection.execute') && ['failed', 'partial', 'canceled', 'cancelled'].includes(run.status) ? <Button size="sm" onClick={() => retry.mutate()} disabled={retry.isPending}><RefreshCw className="size-3.5" />重试</Button> : undefined}><DetailGrid><Detail label="状态"><StatusBadge tone={statusTone(run.status)}>{run.status}</StatusBadge></Detail><Detail label="Coverage">{percent(runCoverage(run))}</Detail><Detail label="时长">{duration(run.duration_ms)}</Detail><Detail label="触发">{run.trigger}</Detail><Detail label="开始">{formatDate(run.started_at)}</Detail><Detail label="Collector">{run.collector_instance_id}</Detail></DetailGrid></Section>{run.error_message && <Alert tone="danger"><strong>{run.error_kind || '运行错误'}：</strong>{run.error_message}</Alert>}<Section title="任务结果" description="服务端真实筛选与分页。"><DataTable data={results.data?.list ?? []} columns={columns} total={results.data?.total ?? 0} page={page} pageSize={pageSize} search="" onSearchChange={() => undefined} onPageChange={setPage} onPageSizeChange={(value) => { setPageSize(value); setPage(1) }} searchable={false} toolbar={toolbar} loading={results.isLoading} error={results.error?.message} /></Section><Section title="技术详情"><JsonBlock value={run} /></Section></>}</div></Sheet>
}

function ManualCollectionPanel() {
  const { toast } = useToast(); const client = useQueryClient()
  const [domain, setDomain] = useState<'game' | 'nav'>('game'); const [scope, setScope] = useState('all')
  const [entity, setEntity] = useState<OptionItem | null>(null); const [target, setTarget] = useState<OptionItem | null>(null); const [tasks, setTasks] = useState<string[]>(['details'])
  const taskOptions = domain === 'game' ? ['details', 'news', 'players'] : ['ping', 'http', 'dns', 'rdap', 'robots', 'security_txt', 'llms_txt', 'page_assets', 'port_check', 'waf_canary']
  const mutation = useMutation({ mutationFn: () => sendJSON<CollectionJob[]>('/api/v1/collection/jobs', 'POST', { domain, scope_type: scope, scope_id: scope === 'all' ? null : Number(entity!.id), target: scope === 'target' ? target?.label : null, tasks }), onSuccess: async (jobs) => { await client.invalidateQueries({ queryKey: ['collection'] }); toast(`已创建 ${jobs.length} 个手动采集任务`) }, onError: (error) => toast(errorMessage(error), 'danger') })
  const changeDomain = (value: 'game' | 'nav') => { setDomain(value); setScope('all'); setEntity(null); setTarget(null); setTasks(value === 'game' ? ['details'] : ['ping']) }
  const toggle = (task: string) => setTasks((current) => current.includes(task) ? current.filter((item) => item !== task) : [...current, task])
  const valid = tasks.length > 0 && (scope === 'all' || entity) && (scope !== 'target' || target)
  const scopeOptions = domain === 'game' ? [{ value: 'all', label: '全部' }, { value: 'game', label: '指定游戏' }] : [{ value: 'all', label: '全部' }, { value: 'site', label: '指定网站' }, { value: 'target', label: '指定 Target' }]
  return <div className="grid gap-5 xl:grid-cols-[1fr_1.4fr]"><Section title="采集对象" description="先选择业务对象，再选择任务，无需记忆内部 ID。"><div className="grid gap-4"><FilterField label="Domain"><Select value={domain} onValueChange={(value) => changeDomain(value as 'game' | 'nav')} options={[{ value: 'game', label: 'Game' }, { value: 'nav', label: 'Nav' }]} /></FilterField><FilterField label="Scope"><Select value={scope} onValueChange={(value) => { setScope(value); setEntity(null); setTarget(null) }} options={scopeOptions} /></FilterField>{scope !== 'all' && <RemoteSelect endpoint={domain === 'game' ? '/api/v1/options/games' : '/api/v1/options/sites'} value={entity} onChange={(value) => { setEntity(value); setTarget(null) }} placeholder={domain === 'game' ? '搜索游戏…' : '搜索网站…'} />}{scope === 'target' && entity && <RemoteSelect endpoint={`/api/v1/options/site-targets?site_id=${entity.id}`} value={target} onChange={setTarget} placeholder="搜索 Target…" />}</div></Section><Section title="任务 / 协议" description="手动执行使用现有 Collection 业务语义。"><div className="flex flex-wrap gap-2">{taskOptions.map((task) => <button key={task} type="button" onClick={() => toggle(task)} className={tasks.includes(task) ? 'rounded-md border border-primary bg-primary px-3 py-2 text-sm text-primary-foreground' : 'rounded-md border bg-surface px-3 py-2 text-sm hover:bg-surface-muted'}>{task}</button>)}</div><div className="mt-6 flex items-center justify-between"><p className="text-xs text-muted-foreground">{domain} · {scope} · {tasks.length} tasks</p><Button disabled={!valid || mutation.isPending} onClick={() => mutation.mutate()}>{mutation.isPending ? <LoaderCircle className="size-4 animate-spin" /> : <Activity className="size-4" />}创建采集任务</Button></div></Section></div>
}
