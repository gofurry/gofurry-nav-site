import { useQuery } from '@tanstack/react-query'
import { Activity, ArrowRight, Gamepad2, Megaphone, Plus, ShieldCheck, Sparkles } from 'lucide-react'
import { Link } from 'react-router-dom'
import { PageHeader, PageLayout, Section } from '../../components/admin/page'
import { EmptyState, ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { getJSON } from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'
import { eventSentence, statusTone } from '../operations/presentation'
import type { WorkbenchSummary } from '../operations/types'

export function WorkbenchPage() {
  const auth = useAuth()
  const query = useQuery({ queryKey: ['workbench', 'summary', auth.state?.identity?.session_version], queryFn: () => getJSON<WorkbenchSummary>('/api/v1/workbench/summary'), refetchInterval: 60_000 })
  if (query.isLoading) return <LoadingState label="正在汇总需要关注的事项…" />
  if (query.error || !query.data) return <ErrorState message={query.error?.message ?? '工作台不可用'} onRetry={() => void query.refetch()} />
  const summary = query.data
  const quickActions = [
    ...(auth.can('content.write') ? [
      { href: '/nav/sites/new', label: '新增网站', icon: Plus },
      { href: '/game/games/new', label: '新增游戏', icon: Gamepad2 },
      { href: '/nav/update-notices', label: '发布公告', icon: Megaphone },
    ] : []),
    ...(auth.can('collection.execute') ? [{ href: '/collection?tab=manual', label: '手动采集', icon: Activity }] : []),
  ]
  return <PageLayout>
    <PageHeader title="工作台" />
    <div className="grid gap-5 xl:grid-cols-2">
      <Section title="快捷操作">{quickActions.length === 0 ? <EmptyState title="暂无可用操作" /> : <div className="-my-2 divide-y">{quickActions.map((item) => <Link key={item.href} to={item.href} className="flex items-center gap-3 py-3 text-sm hover:text-primary"><item.icon className="size-4 text-muted-foreground" /><span className="flex-1 font-medium">{item.label}</span><ArrowRight className="size-4 text-muted-foreground" /></Link>)}</div>}</Section>
      <Section title="系统状态摘要">{summary.system_status.length === 0 ? <EmptyState title="暂无状态信息" /> : <div className="-my-2 divide-y">{summary.system_status.map((item) => <Link key={item.key} to={item.href} className="flex items-center gap-3 py-3 hover:bg-surface-muted/50"><div className="min-w-0 flex-1"><p className="text-sm font-medium">{item.label}</p><p className="mt-1 truncate text-xs text-muted-foreground">{item.summary}</p></div><StatusBadge tone={statusTone(item.status)}>{item.status}</StatusBadge></Link>)}</div>}</Section>
    </div>
    <Section title="需要关注">{summary.attention.length === 0 ? <p className="text-sm text-success">当前没有对你可见的待处理异常。</p> : <div className="-my-2 divide-y">{summary.attention.map((item) => <Link key={item.key} to={item.href} className="flex items-center justify-between gap-4 py-3 hover:text-primary"><div className="min-w-0"><div className="flex items-center gap-2"><StatusBadge tone={statusTone(item.tone)}>{item.tone}</StatusBadge><p className="truncate text-sm font-medium">{item.title}</p></div><p className="mt-1 truncate text-xs text-muted-foreground">{item.summary}</p></div><ArrowRight className="size-4 shrink-0 text-muted-foreground" /></Link>)}</div>}</Section>
    <div className="grid gap-5 xl:grid-cols-2">
      <Section title="最近变化" actions={<Link className="text-sm font-medium text-primary" to="/changes">全部变化 →</Link>}>{summary.recent_changes.length === 0 ? <EmptyState title="暂无变化" /> : <div className="-my-2 divide-y">{summary.recent_changes.map((change) => <Link key={change.event_key} to={`/changes?domain=${change.domain}`} className="flex items-center gap-3 py-3 hover:text-primary"><Sparkles className="size-4 text-info" /><div className="min-w-0 flex-1"><p className="truncate text-sm font-medium">{eventSentence(change.historical_name, change.event_code)}</p><p className="mt-1 text-xs text-muted-foreground">{formatDate(change.event_at ?? change.projection_date)}</p></div><TechnicalLabel>{change.domain}</TechnicalLabel></Link>)}</div>}</Section>
      <Section title="最近操作" actions={auth.can('audit.read') && <Link className="text-sm font-medium text-primary" to="/system/audit">审计日志 →</Link>}>{summary.recent_operations.length === 0 ? <EmptyState title="无可见操作" /> : <div className="-my-2 divide-y">{summary.recent_operations.map((operation) => <div key={operation.id} className="flex items-center gap-3 py-3"><ShieldCheck className="size-4 text-muted-foreground" /><div className="min-w-0 flex-1"><p className="truncate text-sm font-medium">{operation.operator_name} · {operation.action}</p><p className="mt-1 text-xs text-muted-foreground">{operation.resource} · {formatDate(operation.created_at)}</p></div><StatusBadge>{operation.operator_role}</StatusBadge></div>)}</div>}</Section>
    </div>
  </PageLayout>
}
