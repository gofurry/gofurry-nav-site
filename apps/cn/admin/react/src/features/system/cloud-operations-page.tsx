import { ArrowClockwise, ArrowSquareOut, Copy, MagnifyingGlass, ArrowRight, Warning, CloudArrowUp, CloudCheck, ShieldWarning } from '@phosphor-icons/react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Detail, FormField, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { EmptyState, ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { StatusBadge, type StatusTone } from '../../components/admin/status'
import { Select } from '../../components/ui/select'
import { ConfirmAction } from '../../components/ui/dialog'
import { formatDate } from '../../lib/utils'
import { bytes } from '../operations/presentation'
import { Input, Textarea } from '../../components/ui/input'
import { errorMessage, getJSON, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'

type Store = { provider: string; bucket: string; region: string; public_base_url: string; configured: boolean; reachable: boolean; probe_exists: boolean; error?: string }
type CDN = { configured: boolean; main_host?: string; asset_host: string }
type Overview = { primary: Store; mirror: Store; edgeone: CDN; cloudflare: CDN }
type ObjectInfo = { size: number; content_type: string; sha256: string; asset_kind: string; cache_control: string }
type ObjectState = { state: string; error?: string; info?: ObjectInfo }
type Inspection = { object_key: string; primary_url: string; mirror_url: string; primary: ObjectState; mirror: ObjectState; comparison: string }
type Operation<T> = { result: T; warnings?: string[] }
type Purge = { job_id: string; request_id?: string; status: string; failures?: { reason: string; targets: string[] }[] }
type Task = { job_id: string; target: string; type: string; status: string; created_at: string; updated_at: string; failure?: string }
const base = '/api/v1/system/cloud'
const statusLabels: Record<string, string> = { ready: '存在', missing: '不存在', error: '查询失败', not_configured: '未配置', matching: '两端元数据一致', different: '两端元数据不一致', unknown: '暂时无法比较', submitted: '已提交', processing: '处理中', success: '成功', partial: '部分目标失败', failed: '失败', timeout: '超时', canceled: '已取消' }

function CloudStatus({ status }: { status: string }) {
  const tones: Record<string, StatusTone> = { ready: 'success', matching: 'success', success: 'success', missing: 'warning', different: 'warning', partial: 'warning', error: 'danger', failed: 'danger', timeout: 'danger', processing: 'info', submitted: 'info' }
  return <StatusBadge tone={tones[status] ?? 'neutral'}>{statusLabels[status] ?? status}</StatusBadge>
}

export function CloudOperationsPage() {
  const query = useQuery({ queryKey: ['cloud-overview'], queryFn: () => getJSON<Overview>(`${base}/overview`) })
  if (query.isLoading) return <LoadingState />
  if (!query.data) return <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} />
  return <CloudOperationsContent overview={query.data} refresh={() => void query.refetch()} />
}

export function CloudOperationsContent({ overview, refresh }: { overview: Overview; refresh: () => void }) {
  const [jobID, setJobID] = useState('')
  return <PageLayout>
    <PageHeader title="云资源" actions={<Button variant="secondary" onClick={refresh}><ArrowClockwise className="size-4" />刷新状态</Button>} />
    <p className="-mt-2 text-sm text-muted-foreground">查看资源存储状态、检查双端对象，并管理 CDN 缓存。</p>
    <div data-cloud-overview className="grid gap-4 lg:grid-cols-2">{(['primary', 'mirror'] as const).map((key) => {
      const store = overview[key]
      const Icon = key === 'primary' ? CloudArrowUp : CloudCheck
      return <Section key={key} title={key === 'primary' ? 'COS Primary' : 'R2 Mirror'} actions={<StatusBadge tone={!store.configured ? 'neutral' : store.reachable ? 'success' : 'danger'}>{!store.configured ? '未配置' : store.reachable ? '已连接' : '连接失败'}</StatusBadge>}>
        <div className="mb-4 flex items-center gap-3"><span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><Icon className="size-5" /></span><div><p className="text-sm font-medium">{key === 'primary' ? '主存储' : '镜像存储'}</p><p className="text-xs text-muted-foreground">{key === 'primary' ? '资源发布以 COS 成功为准' : '同步失败不阻止主站资源发布'}</p></div></div>
        <dl className="grid min-w-0 grid-cols-2 gap-4"><Detail label="存储桶"><span className="break-all font-mono text-xs">{store.bucket || '—'}</span></Detail><Detail label="区域">{store.region || '—'}</Detail><div className="col-span-2"><Detail label="公共 CDN 地址"><span className="break-all font-mono text-xs">{store.public_base_url || '—'}</span></Detail></div></dl>
        {store.reachable && <div className="mt-4 flex items-center gap-2 border-t pt-3 text-xs text-muted-foreground"><StatusBadge tone={store.probe_exists ? 'success' : 'warning'}>{store.probe_exists ? '探测文件就绪' : 'Probe 文件缺失'}</StatusBadge><span>CDN 连接探测</span></div>}
        {store.error && <div className="mt-3"><Alert tone="danger">{store.error}</Alert></div>}
      </Section>
    })}</div>
    <ObjectInspector />
    <div data-cloud-purge className="grid items-start gap-4 lg:grid-cols-2"><CDNPurge provider="edgeone" config={overview.edgeone} submitted={setJobID} /><CDNPurge provider="cloudflare" config={overview.cloudflare} submitted={() => {}} /></div>
    <PurgeTasks jobID={jobID} setJobID={setJobID} configured={overview.edgeone.configured} />
  </PageLayout>
}

function ObjectInspector() {
  const canManage = useAuth().can('cloudops.manage')
  const [key, setKey] = useState('')
  const [message, setMessage] = useState('')
  const lookup = useMutation({ mutationFn: (value: string) => getJSON<Inspection>(`${base}/object?key=${encodeURIComponent(value)}`) })
  const repair = useMutation({ mutationFn: (value: string) => sendJSON<Operation<Inspection>>(`${base}/object/repair-mirror`, 'POST', { key: value }), onSuccess: (response) => { setMessage(response.warnings?.join('；') || 'Mirror 已从 COS 修复'); lookup.mutate(response.result.object_key) } })
  const copy = async (url: string) => { try { await navigator.clipboard.writeText(url); setMessage('地址已复制') } catch { setMessage('无法访问剪贴板，请从资源链接复制地址') } }
  const item = lookup.data
  return <Section title="对象检查" description="输入完整 Object Key，比较 COS 与 R2 中的对象状态和元数据。">
    <div className="grid gap-4">
      <form className="flex flex-wrap items-end gap-3" onSubmit={(event) => { event.preventDefault(); setMessage(''); repair.reset(); lookup.mutate(key.trim()) }}><div className="min-w-0 basis-64 flex-1"><FormField label="Object Key"><Input className="font-mono text-xs" required value={key} onChange={(event) => setKey(event.target.value)} placeholder="nav/hero/desktop/….avif" /></FormField></div><Button disabled={lookup.isPending || repair.isPending}><MagnifyingGlass className="size-4" />{lookup.isPending ? '检查中…' : '检查'}</Button></form>
      {(lookup.error || repair.error) && <Alert tone="danger">{errorMessage(lookup.error || repair.error)}</Alert>}
      {message && <p role="status" className="text-sm text-primary">{message}</p>}
      {item ? <>
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-md bg-surface-muted px-3 py-2"><code className="min-w-0 break-all text-xs">{item.object_key}</code><CloudStatus status={item.comparison} /></div>
        <div className="grid gap-4 lg:grid-cols-2">{(['primary', 'mirror'] as const).map((key) => {
          const object = item[key]; const url = item[`${key}_url`]
          return <div key={key} className="flex min-w-0 flex-col gap-3 rounded-md border p-4"><div className="flex flex-wrap items-center justify-between gap-2"><h3 className="text-sm font-semibold">{key === 'primary' ? 'COS Primary' : 'R2 Mirror'}</h3><CloudStatus status={object.state} /></div>
            {object.error && <Alert tone="danger">{object.error}</Alert>}
            {object.info && <dl className="grid grid-cols-2 gap-3"><Detail label="大小"><span title={`${object.info.size.toLocaleString()} bytes`}>{bytes(object.info.size)}</span></Detail><Detail label="类型"><span className="break-all">{object.info.content_type}</span></Detail><Detail label="资源种类">{object.info.asset_kind || '—'}</Detail><Detail label="缓存策略"><span className="break-all text-xs">{object.info.cache_control || '—'}</span></Detail><div className="col-span-2"><Detail label="SHA256"><span className="break-all font-mono text-xs">{object.info.sha256 || '无元数据'}</span></Detail></div></dl>}
            <div className="mt-auto flex flex-wrap items-center gap-3 border-t pt-3"><a href={url} className="inline-flex items-center gap-1 text-sm text-primary hover:underline" target="_blank" rel="noreferrer">打开资源<ArrowSquareOut className="size-3.5" /></a><Button size="sm" variant="ghost" onClick={() => void copy(url)}><Copy className="size-3.5" />复制 {key === 'primary' ? 'Primary' : 'Mirror'} URL</Button></div>
          </div>
        })}</div>
        {canManage && <div className="flex flex-wrap items-center justify-between gap-3 border-t pt-3"><p className="text-xs text-muted-foreground">使用 COS 中的原始对象恢复 R2 镜像。</p><Button variant="secondary" disabled={repair.isPending || item.primary.state !== 'ready' || item.object_key === 'system/probes/cdn.bin'} onClick={() => repair.mutate(item.object_key)}><ArrowRight className="size-4" />从 COS 修复 R2</Button></div>}
      </> : <p className="text-xs text-muted-foreground">仅检查指定对象，不列出或删除存储桶中的文件。</p>}
    </div>
  </Section>
}

function CDNPurge({ provider, config, submitted }: { provider: 'edgeone' | 'cloudflare'; config: CDN; submitted: (id: string) => void }) {
  const auth = useAuth()
  const canManage = auth.can('cloudops.manage')
  const [type, setType] = useState<'url' | 'prefix' | 'host'>('url')
  const [targets, setTargets] = useState('')
  const [confirmation, setConfirmation] = useState<{ type: string; targets: string[] } | 'all' | null>(null)
  const mutation = useMutation({ mutationFn: (request: { type: string; targets: string[] } | 'all') => sendJSON<Operation<Purge>>(`${base}/${provider}/${request === 'all' ? 'purge-all' : 'purge'}`, 'POST', request === 'all' ? undefined : request), onSuccess: (response) => { if (response.result.job_id) submitted(response.result.job_id) } })
  const purge = (request: { type: string; targets: string[] } | 'all') => setConfirmation(request)
  const name = provider === 'edgeone' ? 'EdgeOne' : 'Cloudflare'
  return <><Section title={`${name} 清缓存`} description={config.asset_host || '尚未配置资产域名'} actions={<StatusBadge tone={config.configured ? 'info' : 'neutral'}>{config.configured ? '已配置' : '未配置'}</StatusBadge>}>
    <div className="grid gap-4">
      {mutation.error && <Alert tone="danger">{errorMessage(mutation.error)}</Alert>}
      {mutation.data && <div className="grid gap-2 rounded-md border bg-surface-muted p-3 text-sm" role="status"><div className="flex flex-wrap items-center gap-2"><CloudStatus status={mutation.data.result.status} /><code className="break-all text-xs">{mutation.data.result.job_id}</code></div>{mutation.data.result.request_id && <p className="break-all text-xs text-muted-foreground">请求 ID：{mutation.data.result.request_id}</p>}{mutation.data.result.failures?.map((failure, i) => <Alert key={i} tone="danger">{failure.reason}：{failure.targets.join('、')}</Alert>)}{mutation.data.warnings?.map((warning) => <Alert key={warning} tone="warning">{warning}</Alert>)}</div>}
      {canManage ? <fieldset disabled={!config.configured || mutation.isPending} className="grid gap-4">
        <FormField label="清除方式"><Select ariaLabel={`${name} 清除方式`} value={type} disabled={!config.configured || mutation.isPending} onValueChange={(value) => setType(value as typeof type)} options={[{ value: 'url', label: 'URL · 指定资源' }, { value: 'prefix', label: 'Prefix · 路径前缀' }, { value: 'host', label: 'Host · 指定域名' }]} /></FormField>
        <FormField label="清缓存目标" help={type === 'host' ? '每行一个已配置域名，不带协议或路径；最多 20 个。' : '每行一个包含 https:// 的完整地址；最多 20 个。'}><Textarea className="min-h-28 font-mono text-xs" value={targets} onChange={(event) => setTargets(event.target.value)} placeholder={type === 'host' ? config.asset_host : `https://${config.asset_host}/nav/`} /></FormField>
        <div className="flex justify-end"><Button disabled={!targets.trim()} onClick={() => purge({ type, targets: targets.split(/\r?\n/).map((target) => target.trim()).filter(Boolean) })}><ArrowClockwise className="size-4" />提交清缓存</Button></div>
        {provider === 'edgeone' && config.main_host && <div className="grid gap-2 border-t pt-3"><p className="text-xs text-muted-foreground">快捷操作 · 仅清除主站域名</p><Button variant="secondary" className="h-auto min-h-9 whitespace-normal break-all" onClick={() => purge({ type: 'host', targets: [config.main_host!] })}>清除主站缓存 · {config.main_host}</Button></div>}
      </fieldset> : <p className="flex items-center gap-2 text-sm text-muted-foreground"><ShieldWarning className="size-4" />当前账号仅可查看配置与任务。</p>}
      {provider === 'edgeone' && auth.can('cloudops.purge_all') && <aside aria-label="高风险操作" className="grid gap-3 rounded-md border border-danger/30 bg-danger/5 p-4"><div className="flex items-center gap-2 text-sm font-semibold text-danger"><Warning className="size-4" />高风险操作</div><p className="text-xs leading-relaxed text-muted-foreground">覆盖整个 EdgeOne Zone 的所有加速域名。此操作的范围大于主站或指定资源清缓存。</p><Button variant="danger" className="h-auto min-h-9 whitespace-normal" disabled={!config.configured || mutation.isPending} onClick={() => purge('all')}>清除整个 EdgeOne Zone 缓存</Button></aside>}
    </div>
  </Section><ConfirmAction open={confirmation !== null} onOpenChange={(open) => { if (!open) setConfirmation(null) }} title={confirmation === 'all' ? '清除整个 EdgeOne Zone 缓存' : `${name} 清缓存确认`} description={confirmation === 'all' ? '范围：整个 EdgeOne Zone（包括所有加速域名）。确认清除该范围的全部缓存？' : `请核对以下 ${confirmation?.targets.length ?? 0} 个目标，确认后提交缓存清除任务。`} confirmLabel="确认清缓存" variant={confirmation === 'all' ? 'danger' : 'primary'} onConfirm={() => { if (confirmation) { mutation.mutate(confirmation); setConfirmation(null) } }} >{confirmation && confirmation !== 'all' && <ul className="mt-3 divide-y rounded-md border bg-surface-muted px-3">{confirmation.targets.map((target, index) => <li key={index} className="break-all py-2 font-mono text-xs">{target}</li>)}</ul>}</ConfirmAction></>
}

function PurgeTasks({ jobID, setJobID, configured }: { jobID: string; setJobID: (id: string) => void; configured: boolean }) {
  const [page, setPage] = useState(1)
  const query = useQuery({ queryKey: ['cloud-purge-tasks', jobID, page], queryFn: () => getJSON<{ total: number; list: Task[] }>(`${base}/edgeone/purge-tasks?job_id=${encodeURIComponent(jobID)}&page_num=${page}`), enabled: configured })
  const columns: AdminColumn<Task>[] = [
    { key: 'target', header: '目标 / 任务', sortable: false, render: (task) => <div className="max-w-lg py-3"><p className="break-all font-medium">{task.target}</p><p className="mt-1 break-all font-mono text-xs text-muted-foreground">{task.job_id}</p>{task.failure && <p className="mt-2 text-xs text-danger">{task.failure}</p>}</div> },
    { key: 'type', header: '方式', sortable: false, render: (task) => <span className="text-xs uppercase text-muted-foreground">{task.type}</span> },
    { key: 'status', header: '状态', sortable: false, render: (task) => <span className="whitespace-nowrap"><CloudStatus status={task.status} /></span> },
    { key: 'created_at', header: '提交 / 更新时间', sortable: false, render: (task) => <div className="whitespace-nowrap py-3 text-xs"><p>{formatDate(task.created_at)}</p><p className="mt-1 text-muted-foreground">{formatDate(task.updated_at)}</p></div> },
  ]
  return <Section title="EdgeOne 清缓存任务" description="按任务查看缓存清除进度与失败原因。" actions={<Button variant="secondary" disabled={!configured || query.isFetching} onClick={() => void query.refetch()}><ArrowClockwise className="size-4" />刷新任务</Button>}>
    <div className="grid min-w-0 gap-4">
      <div className="max-w-lg"><FormField label="任务 ID" help="留空显示过去 24 小时的任务。"><Input value={jobID} onChange={(event) => { setJobID(event.target.value); setPage(1) }} placeholder="按任务 ID 筛选" /></FormField></div>
      {!configured ? <EmptyState title="EdgeOne 尚未配置" message="配置后可查看清缓存任务。" /> : <DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={20} search="" onSearchChange={() => {}} onPageChange={setPage} searchable={false} loading={query.isLoading} error={query.error ? errorMessage(query.error) : undefined} onRetry={() => void query.refetch()} />}
      {configured && query.data?.total === 0 && <p className="text-xs text-muted-foreground">新提交的任务可能需要稍后刷新。</p>}
    </div>
  </Section>
}
