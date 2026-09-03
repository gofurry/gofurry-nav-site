import { useQuery } from '@tanstack/react-query'
import { Braces, CircleGauge } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FilterBar, FilterField, JsonBlock, Kpi, KpiGrid, OperationTabs, OperationsChart } from '../../components/admin/operations'
import { Detail, DetailGrid, PageHeader, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { TechnicalDetails } from '../../components/admin/technical-details'
import { DatePicker } from '../../components/ui/date-picker'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { getJSON } from '../../lib/api'
import type { PageResult } from '../../lib/types'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'
import { metricsTabs } from './capability-ux'
import { metricStateLabel, percent, statusTone } from './presentation'
import type { MetricCheckpoint, MetricDaily, MetricEntity, MetricOverview, MetricRegistry } from './types'

const metricLabels: Record<string, string> = { free_game_share: '免费游戏占比', windows_support: 'Windows 支持率', linux_support: 'Linux 支持率', ipv6_adoption: 'IPv6 支持率', tls13_adoption: 'TLS 1.3 支持率', security_txt_adoption: 'security.txt 覆盖率' }
const dimensionOptions = { game: ['primary_tag_id', 'tag_id'], nav: ['site_country', 'group_id', 'nsfw', 'welfare'] }
const metricLabel = (key: string) => metricLabels[key] ?? key

export function MetricsPage() {
  const auth = useAuth(); const [params, setParams] = useSearchParams()
  const requested = params.get('tab') ?? 'overview'; const active = requested === 'technical' && !auth.can('metrics.technical') ? 'overview' : metricsTabs.some((tab) => tab.key === requested) ? requested : 'overview'
  const setTab = (tab: string) => { const next = new URLSearchParams(); if (tab !== 'overview') next.set('tab', tab); setParams(next, { replace: true }) }
  return <div className="grid gap-6"><PageHeader title="数据指标" description="用可解释的历史指标回答“数据现在说明什么”。" eyebrow="metrics.read-model" /><OperationTabs tabs={metricsTabs} active={active} onChange={setTab} can={auth.can} />{active === 'overview' && <MetricOverviewPanel />}{active === 'results' && <MetricResultsPanel />}{active === 'entities' && <MetricEntitiesPanel />}{active === 'technical' && auth.can('metrics.technical') && <MetricTechnicalPanel />}</div>
}

function useMetricOverview() { return useQuery({ queryKey: ['metrics', 'overview'], queryFn: () => getJSON<MetricOverview[]>('/api/v1/metrics/overview') }) }

function MetricOverviewPanel() {
  const query = useMetricOverview()
  if (query.isLoading) return <LoadingState />
  if (query.error || !query.data) return <ErrorState message={query.error?.message ?? '指标概览不可用'} onRetry={() => void query.refetch()} />
  return <div className="grid gap-5">{query.data.map((metric) => <Section key={`${metric.domain}-${metric.metric_key}-${metric.metric_version}`} title={metricLabel(metric.metric_key)} description={metric.description} actions={<StatusBadge tone={metric.lag_days ? 'warning' : 'success'}>{metric.lag_days ? `滞后 ${metric.lag_days} 天` : '已同步'}</StatusBadge>}><div className="grid gap-5 xl:grid-cols-[1fr_1.4fr]"><div><p className="text-4xl font-semibold tracking-tight">{percent(metric.adoption_rate)}</p><p className="mt-2 text-sm text-muted-foreground">覆盖率 {percent(metric.coverage_rate)}</p><TechnicalLabel>{metric.metric_key} · v{metric.metric_version} · {metric.domain}</TechnicalLabel></div><KpiGrid><Kpi label="支持" value={metric.positive_count} tone="success" /><Kpi label="不支持" value={metric.negative_count} /><Kpi label="数据过期" value={metric.stale_count} tone={metric.stale_count ? 'warning' : 'neutral'} /><Kpi label="探测失败" value={metric.probe_failed_count} tone={metric.probe_failed_count ? 'danger' : 'neutral'} /><Kpi label="未知" value={metric.unknown_count} tone={metric.unknown_count ? 'warning' : 'neutral'} /></KpiGrid></div></Section>)}</div>
}

function MetricPicker({ metrics, value, onChange }: { metrics: MetricOverview[]; value: string; onChange: (value: string) => void }) {
  return <Select value={value} onValueChange={onChange} options={metrics.map((metric) => ({ value: `${metric.domain}:${metric.metric_key}:${metric.metric_version}`, label: `${metricLabel(metric.metric_key)} · ${metric.domain} v${metric.metric_version}` }))} />
}

function parseMetric(value: string) { const [domain = '', metric = '', version = '0'] = value.split(':'); return { domain, metric, version: Number(version) } }

function MetricResultsPanel() {
  const overview = useMetricOverview(); const [selection, setSelection] = useState(''); const [dimensionKey, setDimensionKey] = useState('global'); const [dimensionValue, setDimensionValue] = useState('all'); const [from, setFrom] = useState(''); const [through, setThrough] = useState(''); const [page, setPage] = useState(1)
  const metrics = overview.data ?? []; const selected = selection || (metrics[0] ? `${metrics[0].domain}:${metrics[0].metric_key}:${metrics[0].metric_version}` : ''); const parsed = parseMetric(selected)
  const search = new URLSearchParams({ domain: parsed.domain, metric: parsed.metric, version: String(parsed.version), dimension_key: dimensionKey || 'global', dimension_value: dimensionValue || 'all', page: String(page), page_size: '20' }); if (from) search.set('from', from); if (through) search.set('to', through)
  const query = useQuery({ queryKey: ['metrics', 'daily', selected, dimensionKey, dimensionValue, from, through, page], queryFn: () => getJSON<PageResult<MetricDaily>>(`/api/v1/metrics/daily?${search}`), enabled: Boolean(parsed.metric) })
  const columns = useMemo<AdminColumn<MetricDaily>[]>(() => [{ key: 'fact_date', header: '日期' }, { key: 'population_count', header: '总体' }, { key: 'eligible_count', header: '可评估' }, { key: 'adoption_rate', header: '支持率', render: (row) => percent(row.adoption_rate) }, { key: 'coverage_rate', header: '覆盖率', render: (row) => percent(row.coverage_rate) }, { key: 'unknown_count', header: '未知' }, { key: 'computed_at', header: '计算时间', render: (row) => formatDate(row.computed_at) }], [])
  if (overview.isLoading) return <LoadingState />
  if (overview.error) return <ErrorState message={overview.error.message} />
  const dimensions = ['global', ...(dimensionOptions[parsed.domain as keyof typeof dimensionOptions] ?? [])].map((value) => ({ value, label: value === 'global' ? '全局' : value }))
  return <div className="grid gap-5"><FilterBar><FilterField label="指标"><MetricPicker metrics={metrics} value={selected} onChange={(value) => { setSelection(value); setDimensionKey('global'); setDimensionValue('all'); setPage(1) }} /></FilterField><FilterField label="维度"><Select value={dimensionKey} onValueChange={(value) => { setDimensionKey(value); setDimensionValue(value === 'global' ? 'all' : ''); setPage(1) }} options={dimensions} /></FilterField>{dimensionKey !== 'global' && <FilterField label="维度值"><Input value={dimensionValue} onChange={(event) => { setDimensionValue(event.target.value); setPage(1) }} placeholder="dimension_value" /></FilterField>}<FilterField label="开始日期"><DatePicker value={from} onValueChange={setFrom} ariaLabel="指标开始日期" /></FilterField><FilterField label="结束日期"><DatePicker value={through} onValueChange={setThrough} ariaLabel="指标结束日期" /></FilterField></FilterBar><Section title="趋势" description="指标率（%）；空分母保持不可用。"><OperationsChart points={(query.data?.list ?? []).slice().reverse().map((row) => ({ label: row.fact_date, value: row.adoption_rate == null ? null : Number((row.adoption_rate * 100).toFixed(1)) }))} unit="%" label="指标趋势" loading={query.isLoading} error={query.error?.message} /></Section><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={20} search="" onSearchChange={() => undefined} onPageChange={setPage} onPageSizeChange={() => undefined} searchable={false} loading={query.isLoading} error={query.error?.message} /></div>
}

function MetricEntitiesPanel() {
  const overview = useMetricOverview(); const [selection, setSelection] = useState(''); const [state, setState] = useState(''); const [page, setPage] = useState(1); const [selected, setSelected] = useState<MetricEntity | null>(null)
  const metrics = overview.data ?? []; const chosen = selection || (metrics[0] ? `${metrics[0].domain}:${metrics[0].metric_key}:${metrics[0].metric_version}` : ''); const parsed = parseMetric(chosen); const metadata = metrics.find((item) => item.domain === parsed.domain && item.metric_key === parsed.metric && item.metric_version === parsed.version); const factDate = metadata?.latest_fact_date ?? ''
  const search = new URLSearchParams({ domain: parsed.domain, metric: parsed.metric, version: String(parsed.version), fact_date: factDate, page: String(page), page_size: '20' }); if (state) search.set('state', state)
  const query = useQuery({ queryKey: ['metrics', 'entities', chosen, factDate, state, page], queryFn: () => getJSON<PageResult<MetricEntity>>(`/api/v1/metrics/entities?${search}`), enabled: Boolean(parsed.metric && factDate) })
  const columns = useMemo<AdminColumn<MetricEntity>[]>(() => [{ key: 'historical_name', header: '实体', render: (row) => <div><p className="font-medium text-primary">{row.historical_name}</p><TechnicalLabel>{row.domain} #{row.entity_id}</TechnicalLabel></div> }, { key: 'state', header: '状态', render: (row) => <StatusBadge tone={statusTone(row.state)}>{metricStateLabel(row.state)}</StatusBadge> }, { key: 'reason_code', header: '原因', render: (row) => row.reason_code.replaceAll('_', ' ') }, { key: 'source_observed_at', header: 'Evidence Time', render: (row) => formatDate(row.source_observed_at) }], [])
  if (overview.isLoading) return <LoadingState />
  if (!factDate && !overview.isLoading) return <AlertMetricEmpty />
  const states = [{ value: '', label: '全部' }, ...['positive', 'negative', 'stale', 'not_probed', 'probe_failed', 'unknown', 'not_applicable'].map((value) => ({ value, label: metricStateLabel(value) }))]
  return <div className="grid gap-5"><FilterBar><FilterField label="指标"><MetricPicker metrics={metrics} value={chosen} onChange={(value) => { setSelection(value); setPage(1) }} /></FilterField><FilterField label="状态"><Select value={state} onValueChange={(value) => { setState(value); setPage(1) }} options={states} /></FilterField><div className="pb-2 text-xs text-muted-foreground">Fact Date: {factDate}</div></FilterBar><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={20} search="" onSearchChange={() => undefined} onPageChange={setPage} onPageSizeChange={() => undefined} onRowClick={setSelected} searchable={false} loading={query.isLoading} error={query.error?.message} /><TechnicalDetails open={Boolean(selected)} onOpenChange={(open) => { if (!open) setSelected(null) }} title={selected?.historical_name ?? '实体'} identifier={selected ? `${selected.domain} #${selected.entity_id}` : undefined}>{selected && <div className="grid gap-5"><DetailGrid><Detail label="State">{metricStateLabel(selected.state)}</Detail><Detail label="reason_code">{selected.reason_code}</Detail><Detail label="observed_at">{formatDate(selected.source_observed_at)}</Detail><Detail label="evaluated_at">{formatDate(selected.evaluated_at)}</Detail></DetailGrid><JsonBlock value={selected.dimension_values} /><JsonBlock value={selected.source_projection_versions} /></div>}</TechnicalDetails></div>
}

function AlertMetricEmpty() { return <Section><div className="grid min-h-40 place-items-center text-sm text-muted-foreground"><CircleGauge className="mb-2 size-8" />当前活跃指标尚无已物化日期。</div></Section> }

function MetricTechnicalPanel() {
  const registry = useQuery({ queryKey: ['metrics', 'registry'], queryFn: () => getJSON<MetricRegistry[]>('/api/v1/metrics/registry') }); const checkpoints = useQuery({ queryKey: ['metrics', 'checkpoints'], queryFn: () => getJSON<MetricCheckpoint[]>('/api/v1/metrics/checkpoints') }); const [selected, setSelected] = useState<MetricRegistry | null>(null)
  const columns = useMemo<AdminColumn<MetricRegistry>[]>(() => [{ key: 'description', header: '指标', render: (row) => <div><p className="font-medium text-primary">{metricLabel(row.metric_key)}</p><TechnicalLabel>{row.domain} · {row.metric_key} v{row.metric_version}</TechnicalLabel></div> }, { key: 'status', header: '状态', render: (row) => <StatusBadge tone={statusTone(row.status)}>{row.status}</StatusBadge> }, { key: 'source_facts', header: 'Source Facts', render: (row) => row.source_facts.join(', ') }, { key: 'freshness_seconds', header: 'Freshness', render: (row) => row.freshness_seconds ? `${row.freshness_seconds}s` : '—' }, { key: 'allowed_dimensions', header: 'Dimensions', render: (row) => row.allowed_dimensions.join(', ') || 'global' }], [])
  return <div className="grid gap-5"><Section title="技术契约" description="只读 Registry 与版本契约；不提供 rebuild/backfill 操作。" actions={<Braces className="size-5 text-muted-foreground" />}><DataTable data={registry.data ?? []} columns={columns} total={registry.data?.length ?? 0} page={1} pageSize={100} search="" onSearchChange={() => undefined} onPageChange={() => undefined} onPageSizeChange={() => undefined} onRowClick={setSelected} searchable={false} loading={registry.isLoading} error={registry.error?.message} /></Section><Section title="Checkpoints"><div className="grid gap-2">{checkpoints.data?.map((checkpoint) => <div key={`${checkpoint.domain}-${checkpoint.metric_key}-${checkpoint.metric_version}`} className="grid gap-2 rounded-md border p-3 md:grid-cols-4"><div><p className="font-medium">{metricLabel(checkpoint.metric_key)}</p><TechnicalLabel>{checkpoint.domain} v{checkpoint.metric_version}</TechnicalLabel></div><div><p className="text-xs text-muted-foreground">Processed Through</p><p className="text-sm">{checkpoint.processed_through ?? '—'}</p></div><div><p className="text-xs text-muted-foreground">Upstream</p><p className="text-sm">{checkpoint.upstream_processed_through ?? '—'}</p></div><StatusBadge tone={checkpoint.lag_days ? 'warning' : 'success'}>{checkpoint.lag_days ?? 0}d lag</StatusBadge></div>)}</div></Section><TechnicalDetails open={Boolean(selected)} onOpenChange={(open) => { if (!open) setSelected(null) }} title={selected ? metricLabel(selected.metric_key) : '指标'} identifier={selected ? `${selected.metric_key} v${selected.metric_version}` : undefined}>{selected && <div className="grid gap-4"><DetailGrid><Detail label="Eligibility Policy">{selected.eligibility_policy}</Detail><Detail label="State Policy">{selected.state_policy}</Detail><Detail label="Coverage Policy">{selected.coverage_policy}</Detail></DetailGrid><JsonBlock value={selected} /></div>}</TechnicalDetails></div>
}
