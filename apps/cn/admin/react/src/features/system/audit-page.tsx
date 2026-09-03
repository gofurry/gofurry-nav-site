import { useQuery } from '@tanstack/react-query'
import { History, ShieldCheck } from 'lucide-react'
import { useMemo, useState } from 'react'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FilterField, JsonBlock } from '../../components/admin/operations'
import { Detail, DetailGrid, PageHeader, Section } from '../../components/admin/page'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { TechnicalDetails } from '../../components/admin/technical-details'
import { DatePicker } from '../../components/ui/date-picker'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { getJSON } from '../../lib/api'
import type { PageResult } from '../../lib/types'
import { formatDate } from '../../lib/utils'
import type { AuditLog } from '../operations/types'

function auditJSON(value: string) { try { return JSON.parse(value || '{}') as unknown } catch { return {} } }

export function AuditPage() {
  const [page, setPage] = useState(1); const [operator, setOperator] = useState(''); const [role, setRole] = useState(''); const [action, setAction] = useState(''); const [resource, setResource] = useState(''); const [from, setFrom] = useState(''); const [through, setThrough] = useState(''); const [selected, setSelected] = useState<AuditLog | null>(null)
  const search = new URLSearchParams({ page: String(page), page_size: '20' }); if (operator) search.set('operator', operator); if (role) search.set('role', role); if (action) search.set('action', action); if (resource) search.set('resource', resource); if (from) search.set('from', from); if (through) search.set('to', through)
  const query = useQuery({ queryKey: ['audit', page, operator, role, action, resource, from, through], queryFn: () => getJSON<PageResult<AuditLog>>(`/api/v1/audit/logs?${search}`) })
  const columns = useMemo<AdminColumn<AuditLog>[]>(() => [
    { key: 'created_at', header: '时间', render: (row) => formatDate(row.created_at) },
    { key: 'operator_name', header: '操作者', render: (row) => <div><p className="font-medium text-primary">{row.operator_name || row.operator}</p><TechnicalLabel>{row.operator}</TechnicalLabel></div> },
    { key: 'operator_role', header: '角色', render: (row) => <StatusBadge>{row.operator_role || '未知'}</StatusBadge> },
    { key: 'action', header: '操作', render: (row) => row.action.replaceAll('.', ' / ') },
    { key: 'resource', header: '对象', render: (row) => <div><p>{row.resource}</p><TechnicalLabel>{row.target_id || '—'}</TechnicalLabel></div> },
  ], [])
  const resetPage = <T,>(setter: (value: T) => void) => (value: T) => { setter(value); setPage(1) }
  const toolbar = <><FilterField label="历史角色"><Select value={role} onValueChange={resetPage(setRole)} options={[{ value: '', label: '全部角色' }, { value: 'owner', label: 'Owner' }, { value: 'developer', label: 'Developer' }, { value: 'operator', label: 'Operator' }, { value: 'system', label: 'System' }]} /></FilterField><FilterField label="Action"><Input value={action} onChange={(event) => resetPage(setAction)(event.target.value)} placeholder="account.role" /></FilterField><FilterField label="Resource"><Input value={resource} onChange={(event) => resetPage(setResource)(event.target.value)} placeholder="gfa_admin_account" /></FilterField><FilterField label="开始"><DatePicker value={from} onValueChange={resetPage(setFrom)} ariaLabel="审计开始日期" /></FilterField><FilterField label="结束"><DatePicker value={through} onValueChange={resetPage(setThrough)} ariaLabel="审计结束日期" /></FilterField></>
  return <div className="grid gap-6"><PageHeader title="操作审计" description="使用操作发生时的账号名与角色快照解读历史。" eyebrow="audit.read" actions={<ShieldCheck className="size-5 text-muted-foreground" />} /><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={20} search={operator} onSearchChange={resetPage(setOperator)} onPageChange={setPage} onPageSizeChange={() => undefined} searchPlaceholder="搜索历史操作者…" toolbar={toolbar} onRowClick={setSelected} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /><TechnicalDetails open={Boolean(selected)} onOpenChange={(open) => { if (!open) setSelected(null) }} title={selected?.action ?? '审计详情'} identifier={selected ? `audit #${selected.id}` : undefined}>{selected && <div className="grid gap-5"><Section title="业务摘要" actions={<History className="size-5 text-muted-foreground" />}><DetailGrid><Detail label="操作者">{selected.operator_name || selected.operator}</Detail><Detail label="历史角色">{selected.operator_role || '未知'}</Detail><Detail label="操作">{selected.action}</Detail><Detail label="对象">{selected.resource} · {selected.target_id || '—'}</Detail><Detail label="时间">{formatDate(selected.created_at)}</Detail></DetailGrid></Section><Section title="Before / After Diff"><div className="grid gap-4 xl:grid-cols-2"><div><p className="mb-2 text-xs font-medium text-muted-foreground">Before</p><JsonBlock value={auditJSON(selected.before_data)} /></div><div><p className="mb-2 text-xs font-medium text-muted-foreground">After</p><JsonBlock value={auditJSON(selected.after_data)} /></div></div></Section><Section title="Technical Metadata"><JsonBlock value={{ id: selected.id, operator: selected.operator, operator_account_id: selected.operator_account_id, operator_name: selected.operator_name, operator_role: selected.operator_role, session_version: selected.session_version, request_id: selected.request_id, ip_address: selected.ip_address, user_agent: selected.user_agent }} /></Section></div>}</TechnicalDetails></div>
}
