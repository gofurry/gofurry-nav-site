import { useQuery } from '@tanstack/react-query'
import { Database, LockKeyhole } from 'lucide-react'
import { useMemo, useState, type ReactNode } from 'react'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { JsonBlock } from '../../components/admin/operations'
import { Detail, DetailGrid, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { TechnicalDetails } from '../../components/admin/technical-details'
import { getJSON } from '../../lib/api'
import { cn, formatDate } from '../../lib/utils'
import { bytes, percent, statusTone } from '../operations/presentation'
import type { DataOpsOverview, DatabaseStatus, RelationSize } from '../operations/types'

export function DataOperationsPage() {
  const query = useQuery({ queryKey: ['dataops', 'overview'], queryFn: () => getJSON<DataOpsOverview>('/api/v1/dataops/overview'), refetchInterval: 60_000 })
  const [selectedKey, setSelectedKey] = useState('gfa')
  const [technical, setTechnical] = useState<DatabaseStatus | null>(null)
  if (query.isLoading) return <LoadingState label="正在检查三个数据库…" />
  if (query.error || !query.data) return <ErrorState message={query.error?.message ?? '数据运维信息不可用'} onRetry={() => void query.refetch()} />
  const selected = query.data.databases.find((database) => database.key === selectedKey) ?? query.data.databases[0]
  const selectDatabase = (database: DatabaseStatus) => {
    if (database.key === selected?.key) setTechnical(database)
    else { setSelectedKey(database.key); setTechnical(null) }
  }
  return <PageLayout>
    <PageHeader title="数据运维" actions={<StatusBadge tone="info"><LockKeyhole className="size-3" />只读</StatusBadge>} />
    <div className="grid gap-3 md:grid-cols-3">{query.data.databases.map((database) => {
      const active = database.key === selected?.key
      return <button key={database.key} type="button" aria-pressed={active} onClick={() => selectDatabase(database)} className={cn('rounded-lg border bg-surface p-4 text-left transition-colors hover:bg-surface-muted', active && 'border-primary bg-primary/5 ring-1 ring-primary/30')}><div className="flex items-center justify-between gap-3"><span className="flex items-center gap-2 font-semibold"><Database className="size-4 text-muted-foreground" />{database.key.toUpperCase()}</span><StatusBadge tone={statusTone(database.health)}>{database.health}</StatusBadge></div><p className="mt-3 truncate text-sm">{database.database_name || '未连接'}</p><p className="mt-1 text-xs text-muted-foreground">迁移 {database.migration.status} · Pending {database.migration.pending_count}</p></button>
    })}</div>
    {selected && <DatabasePanel database={selected} />}
    <p className="text-right text-xs text-muted-foreground">检查时间 {formatDate(query.data.generated_at)} · 每 60 秒刷新</p>
    <TechnicalDetails open={Boolean(technical)} onOpenChange={(open) => { if (!open) setTechnical(null) }} title={technical?.key.toUpperCase() ?? '数据库'} identifier={technical?.database_name} showHint={false}>{technical && <JsonBlock value={{ key: technical.key, health: technical.health, postgresql_version: technical.postgresql_version, database_name: technical.database_name, database_time: technical.database_time, database_size_bytes: technical.database_size_bytes, connections: { total: technical.total_connections, active: technical.active_connections, max: technical.max_connections, usage: technical.connection_usage }, migration: technical.migration, relations: technical.relations }} />}</TechnicalDetails>
  </PageLayout>
}

function DatabasePanel({ database }: { database: DatabaseStatus }) {
  const columns = useMemo<AdminColumn<RelationSize>[]>(() => [{ key: 'name', header: 'Relation', render: (row) => <TechnicalLabel>{row.name}</TechnicalLabel> }, { key: 'table_bytes', header: 'Table', render: (row) => bytes(row.table_bytes) }, { key: 'index_bytes', header: 'Indexes', render: (row) => bytes(row.index_bytes) }, { key: 'total_bytes', header: 'Total', render: (row) => <span className="font-medium">{bytes(row.total_bytes)}</span> }], [])
  if (database.health === 'unavailable') return <ErrorState title={`${database.key} 不可用`} message={database.error || '数据库元数据读取失败'} />
  return <div className="grid gap-5">
    <div className="grid gap-5 xl:grid-cols-2">
      <Section title="状态与连接"><DetailGrid><Detail label="Health"><StatusBadge tone={statusTone(database.health)}>{database.health}</StatusBadge></Detail><Detail label="Database">{database.database_name}</Detail><Detail label="PostgreSQL">{database.postgresql_version}</Detail><Detail label="Database Size">{bytes(database.database_size_bytes)}</Detail><Detail label="Connections">{database.active_connections} active / {database.total_connections} total</Detail><Detail label="Connection Usage">{database.connection_usage == null ? '—' : percent(database.connection_usage)}</Detail><Detail label="Max Connections">{database.max_connections}</Detail><Detail label="Database Time">{formatDate(database.database_time)}</Detail></DetailGrid></Section>
      <Section title="迁移状态"><div className="grid gap-4 sm:grid-cols-2"><OperationValue label="Current Applied" value={database.migration.current_applied || '—'} /><OperationValue label="Expected Repository" value={database.migration.expected} /><OperationValue label="Pending" value={database.migration.pending_count} tone={database.migration.pending_count ? 'warning' : 'success'} /><OperationValue label="Status" value={<StatusBadge tone={statusTone(database.migration.status)}>{database.migration.status}</StatusBadge>} /></div></Section>
    </div>
    <div><div className="mb-3 flex items-center justify-between"><h2 className="font-semibold">最大 Relations</h2><span className="text-xs text-muted-foreground">{database.key.toUpperCase()}</span></div><DataTable data={database.relations} columns={columns} total={database.relations.length} page={1} pageSize={20} search="" onSearchChange={() => undefined} onPageChange={() => undefined} onPageSizeChange={() => undefined} searchable={false} /></div>
  </div>
}

function OperationValue({ label, value, tone }: { label: string; value: ReactNode; tone?: 'success' | 'warning' }) {
  return <div className={cn('rounded-md bg-surface-muted p-4', tone === 'warning' && 'text-warning', tone === 'success' && 'text-success')}><p className="text-xs text-muted-foreground">{label}</p><div className="mt-2 text-xl font-semibold">{value}</div></div>
}
