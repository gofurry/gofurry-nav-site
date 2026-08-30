import { useQuery } from '@tanstack/react-query'
import { Clock3, ExternalLink } from 'lucide-react'
import type { ReactNode } from 'react'
import { getJSON } from '../../lib/api'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../../features/auth/auth-context'
import { Alert } from '../ui/alert'
import { EmptyState, LoadingState } from './states'
import { StatusBadge, TechnicalLabel } from './status'

export type WorkspaceTab = { key: string; label: string }
export function WorkspaceTabs({ tabs, active, onChange }: { tabs: WorkspaceTab[]; active: string; onChange: (key: string) => void }) {
  return <div className="flex gap-1 border-b" role="tablist">{tabs.map((tab) => <button key={tab.key} type="button" role="tab" aria-selected={active === tab.key} onClick={() => onChange(tab.key)} className={active === tab.key ? 'border-b-2 border-primary px-3 py-2.5 text-sm font-medium text-primary' : 'border-b-2 border-transparent px-3 py-2.5 text-sm text-muted-foreground hover:text-foreground'}>{tab.label}</button>)}</div>
}

type ChangeEvent = { event_key: string; detector_key: string; detector_version: number; entity_id: number; historical_name: string; event_at?: string; projection_date: string; event_code: string; scope_kind: string; scope_key: string; materialized_at: string }
type ChangePage = { list: ChangeEvent[]; total: number }

export function HistoryPanel({ domain, entityId }: { domain: 'nav' | 'game'; entityId: number }) {
  const auth = useAuth()
  const enabled = auth.can('changes.read')
  const query = useQuery({ queryKey: ['entity-history', domain, entityId], queryFn: () => getJSON<ChangePage>(`/api/v1/changes/events?domain=${domain}&entity_id=${entityId}&page=1&page_size=20`), enabled })
  if (!enabled) return <Alert tone="warning">当前账号没有 <TechnicalLabel>changes.read</TechnicalLabel>，无法读取变化事件。操作审计聚合将在 P0.5.2-C 接入。</Alert>
  if (query.isLoading) return <LoadingState label="正在加载变化记录…" />
  if (query.error) return <Alert tone="danger">{query.error.message}</Alert>
  if (!query.data?.list.length) return <EmptyState title="暂无可用历史" message="当前实体没有已物化的变化事件；操作审计时间线将在 P0.5.2-C 聚合。" />
  return <div className="rounded-md border bg-surface">{query.data.list.map((event) => <article key={event.event_key} className="flex gap-3 border-b p-4 last:border-0"><div className="mt-0.5 grid size-8 shrink-0 place-items-center rounded-full bg-surface-muted"><Clock3 className="size-4 text-muted-foreground" /></div><div className="min-w-0 flex-1"><div className="flex items-center justify-between gap-4"><p className="font-medium">{event.event_code}</p><span className="text-xs text-muted-foreground">{formatDate(event.event_at || event.materialized_at)}</span></div><p className="mt-1 text-sm text-muted-foreground">{event.historical_name || `实体 #${event.entity_id}`} · {event.scope_kind}/{event.scope_key}</p><TechnicalLabel>{event.detector_key} · v{event.detector_version}</TechnicalLabel></div></article>)}</div>
}

export function DeferredPanel({ title, children, legacyPath }: { title: string; children: ReactNode; legacyPath: string }) {
  const legacyOrigin = import.meta.env.VITE_LEGACY_ADMIN_ORIGIN || 'http://127.0.0.1:10099'
  return <div className="rounded-lg border bg-surface p-6"><StatusBadge tone="info">分阶段迁移</StatusBadge><h2 className="mt-3 text-lg font-semibold">{title}</h2><div className="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">{children}</div><a href={`${legacyOrigin}${legacyPath}`} className="mt-4 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline">打开现有中心<ExternalLink className="size-3.5" /></a></div>
}
