import { useQuery } from '@tanstack/react-query'
import { Database, HardDrive, LockKeyhole } from 'lucide-react'
import { useMemo, useState } from 'react'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { JsonBlock, Kpi, KpiGrid, OperationTabs } from '../../components/admin/operations'
import { Detail, DetailGrid, PageHeader, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { getJSON } from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { bytes, percent, statusTone } from '../operations/presentation'
import type { DataOpsOverview, DatabaseStatus, RelationSize } from '../operations/types'

export function DataOperationsPage() {
  const query = useQuery({ queryKey: ['dataops', 'overview'], queryFn: () => getJSON<DataOpsOverview>('/api/v1/dataops/overview'), refetchInterval: 60_000 })
  const [selectedKey, setSelectedKey] = useState('gfa')
  if (query.isLoading) return <LoadingState label="正在检查三个数据库…" />
  if (query.error || !query.data) return <ErrorState message={query.error?.message ?? '数据运维信息不可用'} onRetry={() => void query.refetch()} />
  const selected = query.data.databases.find((database) => database.key === selectedKey) ?? query.data.databases[0]
  return <div className="grid gap-6"><PageHeader title="数据运维" description="Issue #77 · 只读 PostgreSQL、Schema 与 Goose 迁移健康中心。" eyebrow="dataops.read" actions={<StatusBadge tone="info"><LockKeyhole className="size-3" />只读</StatusBadge>} /><Alert tone="info">本页不提供 SQL 执行、迁移 / 回滚、DROP / TRUNCATE、VACUUM 或连接终止能力。</Alert><KpiGrid>{query.data.databases.map((database) => <Kpi key={database.key} label={database.key.toUpperCase()} value={<StatusBadge tone={statusTone(database.health)}>{database.health}</StatusBadge>} detail={`${database.database_name || '未连接'} · ${database.migration.status}`} tone={database.health === 'healthy' ? 'success' : database.health === 'unavailable' ? 'danger' : 'warning'} />)}</KpiGrid><OperationTabs tabs={query.data.databases.map((database) => ({ key: database.key, label: database.key.toUpperCase() }))} active={selected?.key ?? selectedKey} onChange={setSelectedKey} />{selected && <DatabasePanel database={selected} />}<p className="text-right text-xs text-muted-foreground">检查时间 {formatDate(query.data.generated_at)} · 每 60 秒刷新</p></div>
}

function DatabasePanel({ database }: { database: DatabaseStatus }) {
  const columns = useMemo<AdminColumn<RelationSize>[]>(() => [{ key: 'name', header: 'Relation', render: (row) => <TechnicalLabel>{row.name}</TechnicalLabel> }, { key: 'table_bytes', header: 'Table', render: (row) => bytes(row.table_bytes) }, { key: 'index_bytes', header: 'Indexes', render: (row) => bytes(row.index_bytes) }, { key: 'total_bytes', header: 'Total', render: (row) => <span className="font-medium">{bytes(row.total_bytes)}</span> }], [])
  if (database.health === 'unavailable') return <ErrorState title={`${database.key} 不可用`} message={database.error || '数据库元数据读取失败'} />
  return <div className="grid gap-5"><Section title="数据库概览" actions={<Database className="size-5 text-muted-foreground" />}><DetailGrid><Detail label="Health"><StatusBadge tone={statusTone(database.health)}>{database.health}</StatusBadge></Detail><Detail label="PostgreSQL Version">{database.postgresql_version}</Detail><Detail label="Database Name">{database.database_name}</Detail><Detail label="Database Size">{bytes(database.database_size_bytes)}</Detail><Detail label="Connection Usage">{database.connection_usage == null ? '—' : percent(database.connection_usage)}</Detail><Detail label="Database Time">{formatDate(database.database_time)}</Detail></DetailGrid></Section><Section title="迁移状态" description="Goose 是唯一 Schema Owner。"><div className="grid gap-4 md:grid-cols-4"><Kpi label="Current Applied" value={database.migration.current_applied || '—'} /><Kpi label="Expected Repository" value={database.migration.expected} /><Kpi label="Pending" value={database.migration.pending_count} tone={database.migration.pending_count ? 'warning' : 'success'} /><Kpi label="Status" value={<StatusBadge tone={statusTone(database.migration.status)}>{database.migration.status}</StatusBadge>} /></div></Section><Section title="连接" description="安全的服务器元数据，不暴露 DSN 或凭据。"><DetailGrid><Detail label="Total Connections">{database.total_connections}</Detail><Detail label="Active Connections">{database.active_connections}</Detail><Detail label="Max Connections">{database.max_connections}</Detail></DetailGrid></Section><Section title="最大 Relations" description="限定 Top 10，不是通用 PostgreSQL 浏览器。" actions={<HardDrive className="size-5 text-muted-foreground" />}><DataTable data={database.relations} columns={columns} total={database.relations.length} page={1} pageSize={20} search="" onSearchChange={() => undefined} onPageChange={() => undefined} onPageSizeChange={() => undefined} searchable={false} /></Section><Section title="技术元数据"><JsonBlock value={{ key: database.key, health: database.health, postgresql_version: database.postgresql_version, database_name: database.database_name, database_time: database.database_time, connections: { total: database.total_connections, active: database.active_connections, max: database.max_connections }, migration: database.migration }} /></Section></div>
}
