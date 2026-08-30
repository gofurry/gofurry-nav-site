import { useQuery } from '@tanstack/react-query'
import { Braces, FileClock } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FilterBar, FilterField, JsonBlock, OperationTabs, RemotePicker } from '../../components/admin/operations'
import { Detail, DetailGrid, PageHeader, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { TechnicalDetails } from '../../components/admin/technical-details'
import { NativeSelect } from '../../components/ui/select'
import { getJSON } from '../../lib/api'
import type { OptionItem } from '../../lib/types'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'
import { changesTabs } from './capability-ux'
import { eventLabel, eventSentence, statusTone } from './presentation'
import type { ChangeCheckpoint, ChangeEvent, ChangeOverview, ChangePage, ChangeRegistry } from './types'


export function ChangesPage() {
  const auth = useAuth(); const [params, setParams] = useSearchParams(); const requested = params.get('tab') ?? 'recent'; const active = requested === 'technical' && !auth.can('changes.technical') ? 'recent' : changesTabs.some((tab) => tab.key === requested) ? requested : 'recent'
  const setTab = (tab: string) => { const next = new URLSearchParams(); if (tab !== 'recent') next.set('tab', tab); setParams(next, { replace: true }) }
  return <div className="grid gap-6"><PageHeader title="变化事件" description="用人可读的语义事件回答“最近发生了什么”。" eyebrow="changes.read-model" /><OperationTabs tabs={changesTabs} active={active} onChange={setTab} can={auth.can} />{active === 'recent' && <RecentChangesPanel />}{active === 'entities' && <EntityChangesPanel />}{active === 'technical' && auth.can('changes.technical') && <ChangeTechnicalPanel />}</div>
}

function ChangeTable({ domain, entityID = 0 }: { domain: 'game' | 'nav'; entityID?: number }) {
  const [page, setPage] = useState(1); const [eventCode, setEventCode] = useState(''); const [from, setFrom] = useState(''); const [through, setThrough] = useState(''); const [selected, setSelected] = useState<ChangeEvent | null>(null)
  const search = new URLSearchParams({ domain, page: String(page), page_size: '20' }); if (entityID) search.set('entity_id', String(entityID)); if (eventCode) search.set('event_code', eventCode); if (from) search.set('from', from); if (through) search.set('to', through)
  const query = useQuery({ queryKey: ['changes', 'events', domain, entityID, eventCode, from, through, page], queryFn: () => getJSON<ChangePage>(`/api/v1/changes/events?${search}`) })
  const columns = useMemo<AdminColumn<ChangeEvent>[]>(() => [
    { key: 'event_at', header: '时间', render: (row) => formatDate(row.event_at ?? row.projection_date) },
    { key: 'historical_name', header: '变化', render: (row) => <div><p className="font-medium text-primary">{eventSentence(row.historical_name, row.event_code)}</p><TechnicalLabel>{row.domain} · {row.detector_key} v{row.detector_version}</TechnicalLabel></div> },
    { key: 'event_code', header: '类型', render: (row) => <StatusBadge tone="info">{eventLabel(row.event_code)}</StatusBadge> },
    { key: 'scope_kind', header: '范围', render: (row) => `${row.scope_kind} · ${row.scope_key}` },
  ], [])
  return <div className="grid gap-4"><FilterBar><FilterField label="事件类型"><input className="h-9 rounded-md border bg-surface px-3 text-sm" value={eventCode} onChange={(event) => { setEventCode(event.target.value); setPage(1) }} placeholder="event_code" /></FilterField><FilterField label="开始日期"><input className="h-9 rounded-md border bg-surface px-3 text-sm" type="date" value={from} onChange={(event) => { setFrom(event.target.value); setPage(1) }} /></FilterField><FilterField label="结束日期"><input className="h-9 rounded-md border bg-surface px-3 text-sm" type="date" value={through} onChange={(event) => { setThrough(event.target.value); setPage(1) }} /></FilterField></FilterBar><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={20} search="" onSearchChange={() => undefined} onPageChange={setPage} onPageSizeChange={() => undefined} onRowClick={setSelected} searchable={false} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /><EventDrawer event={selected} onOpenChange={(open) => { if (!open) setSelected(null) }} /></div>
}

function RecentChangesPanel() {
  const [domain, setDomain] = useState<'game' | 'nav'>('nav')
  const overview = useQuery({ queryKey: ['changes', 'overview'], queryFn: () => getJSON<ChangeOverview[]>('/api/v1/changes/overview') })
  if (overview.isLoading) return <LoadingState />
  if (overview.error) return <ErrorState message={overview.error.message} />
  const domainOverview = overview.data?.filter((row) => row.domain === domain) ?? []
  return <div className="grid gap-5"><Section title="变化摘要" actions={<NativeSelect value={domain} onChange={(event) => setDomain(event.target.value as 'game' | 'nav')}><option value="nav">网站</option><option value="game">游戏</option></NativeSelect>}><div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">{domainOverview.map((row) => <div key={`${row.detector_key}-${row.detector_version}`} className="rounded-md border p-4"><div className="flex items-center justify-between"><StatusBadge tone={statusTone(row.status)}>{row.status}</StatusBadge><span className="text-2xl font-semibold">{row.latest_event_count}</span></div><p className="mt-3 text-sm font-medium">{row.description}</p><p className="mt-1 text-xs text-muted-foreground">{row.latest_projection_date ?? '尚无投影'} · 累计 {row.total_event_count}</p></div>)}</div></Section><ChangeTable domain={domain} /></div>
}

function EntityChangesPanel() {
  const [domain, setDomain] = useState<'game' | 'nav'>('nav'); const [entity, setEntity] = useState<OptionItem | null>(null)
  return <div className="grid gap-5"><Section title="选择实体" description="按网站或游戏查看可重现的语义变化历史。"><div className="grid gap-3 md:grid-cols-[12rem_1fr]"><FilterField label="Domain"><NativeSelect value={domain} onChange={(event) => { setDomain(event.target.value as 'game' | 'nav'); setEntity(null) }}><option value="nav">网站</option><option value="game">游戏</option></NativeSelect></FilterField><RemotePicker endpoint={domain === 'game' ? '/api/v1/options/games' : '/api/v1/options/sites'} value={entity} onChange={setEntity} placeholder={domain === 'game' ? '搜索游戏…' : '搜索网站…'} /></div></Section>{entity ? <ChangeTable key={`${domain}-${entity.id}`} domain={domain} entityID={Number(entity.id)} /> : <Section><div className="grid min-h-40 place-items-center text-sm text-muted-foreground"><FileClock className="mb-2 size-8" />请先选择实体。</div></Section>}</div>
}

function EventDrawer({ event, onOpenChange }: { event: ChangeEvent | null; onOpenChange: (open: boolean) => void }) {
  return <TechnicalDetails open={Boolean(event)} onOpenChange={onOpenChange} title={event ? eventSentence(event.historical_name, event.event_code) : '变化事件'} identifier={event?.event_key}>{event && <div className="grid gap-5"><Section title="业务摘要"><DetailGrid><Detail label="对象">{event.historical_name}</Detail><Detail label="变化">{eventLabel(event.event_code)}</Detail><Detail label="时间">{formatDate(event.event_at ?? event.projection_date)}</Detail></DetailGrid><div className="mt-5 grid gap-4 md:grid-cols-2"><div><p className="mb-2 text-xs text-muted-foreground">旧值</p><JsonBlock value={event.old_value} /></div><div><p className="mb-2 text-xs text-muted-foreground">新值</p><JsonBlock value={event.new_value} /></div></div></Section><Section title="技术详情"><DetailGrid><Detail label="detector">{event.detector_key} v{event.detector_version}</Detail><Detail label="event_code">{event.event_code}</Detail><Detail label="scope">{event.scope_kind} · {event.scope_key}</Detail><Detail label="projection_date">{event.projection_date}</Detail><Detail label="event_at">{formatDate(event.event_at)}</Detail><Detail label="time_basis">{event.time_basis}</Detail></DetailGrid><div className="mt-4"><JsonBlock value={event.source_versions} /></div></Section></div>}</TechnicalDetails>
}

function ChangeTechnicalPanel() {
  const registry = useQuery({ queryKey: ['changes', 'registry'], queryFn: () => getJSON<ChangeRegistry[]>('/api/v1/changes/registry') }); const checkpoints = useQuery({ queryKey: ['changes', 'checkpoints'], queryFn: () => getJSON<ChangeCheckpoint[]>('/api/v1/changes/checkpoints') }); const [selected, setSelected] = useState<ChangeRegistry | null>(null)
  const columns = useMemo<AdminColumn<ChangeRegistry>[]>(() => [{ key: 'description', header: 'Detector', render: (row) => <div><p className="font-medium text-primary">{row.description}</p><TechnicalLabel>{row.domain} · {row.detector_key} v{row.detector_version}</TechnicalLabel></div> }, { key: 'status', header: '状态', render: (row) => <StatusBadge tone={statusTone(row.status)}>{row.status}</StatusBadge> }, { key: 'source_kind', header: 'Source Kind' }, { key: 'source_contracts', header: 'Source Contracts', render: (row) => row.source_contracts.join(', ') }, { key: 'watermark_policy', header: 'Watermark' }], [])
  return <div className="grid gap-5"><Section title="Detector Registry" description="只读契约；本页不提供 rebuild/backfill 控制。" actions={<Braces className="size-5 text-muted-foreground" />}><DataTable data={registry.data ?? []} columns={columns} total={registry.data?.length ?? 0} page={1} pageSize={100} search="" onSearchChange={() => undefined} onPageChange={() => undefined} onPageSizeChange={() => undefined} onRowClick={setSelected} searchable={false} loading={registry.isLoading} error={registry.error?.message} /></Section><Section title="Checkpoints"><div className="grid gap-2">{checkpoints.data?.map((row) => <div key={`${row.domain}-${row.detector_key}-${row.detector_version}`} className="grid gap-2 rounded-md border p-3 md:grid-cols-4"><div><p className="font-medium">{row.detector_key}</p><TechnicalLabel>{row.domain} v{row.detector_version}</TechnicalLabel></div><div><p className="text-xs text-muted-foreground">Processed Through</p><p className="text-sm">{row.processed_through ?? '—'}</p></div><div><p className="text-xs text-muted-foreground">Upstream</p><p className="text-sm">{row.upstream_processed_through ?? '—'}</p></div><StatusBadge tone={row.lag_days ? 'warning' : 'success'}>{row.lag_days ?? 0}d lag</StatusBadge></div>)}</div></Section><TechnicalDetails open={Boolean(selected)} onOpenChange={(open) => { if (!open) setSelected(null) }} title={selected?.description ?? 'Detector'} identifier={selected ? `${selected.detector_key} v${selected.detector_version}` : undefined}>{selected && <div className="grid gap-4"><DetailGrid><Detail label="Detection Policy">{selected.detection_policy}</Detail><Detail label="Watermark Policy">{selected.watermark_policy}</Detail><Detail label="Processing Grain">{selected.processing_grain}</Detail></DetailGrid><JsonBlock value={selected} /></div>}</TechnicalDetails></div>
}
