import { useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { FormField, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
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

export function CloudOperationsPage() {
  const query = useQuery({ queryKey: ['cloud-overview'], queryFn: () => getJSON<Overview>(`${base}/overview`) })
  if (query.isLoading) return <LoadingState />
  if (!query.data) return <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} />
  return <CloudOperationsContent overview={query.data} refresh={() => void query.refetch()} />
}

export function CloudOperationsContent({ overview, refresh }: { overview: Overview; refresh: () => void }) {
  const [jobID, setJobID] = useState('')
  return <PageLayout>
    <PageHeader title="云资源" actions={<Button variant="secondary" onClick={refresh}>刷新状态</Button>} />
    <div className="grid gap-4 lg:grid-cols-2">{(['primary', 'mirror'] as const).map((key) => { const store = overview[key]; return <Section key={key} title={key === 'primary' ? 'COS · Primary' : 'R2 · Mirror'}>
      <p className="text-sm">{!store.configured ? '未配置' : store.reachable ? '已连接' : '连接失败'}{store.reachable && !store.probe_exists && ' · Probe 文件缺失'}</p>
      <dl className="mt-3 grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm"><dt>Bucket</dt><dd className="break-all">{store.bucket || '—'}</dd><dt>Region</dt><dd>{store.region || '—'}</dd><dt>CDN</dt><dd className="break-all">{store.public_base_url || '—'}</dd></dl>
      {store.error && <Alert tone="danger">{store.error}</Alert>}
    </Section> })}</div>
    <ObjectInspector />
    <div className="grid items-start gap-4 xl:grid-cols-2"><CDNPurge provider="edgeone" config={overview.edgeone} submitted={setJobID} /><CDNPurge provider="cloudflare" config={overview.cloudflare} submitted={() => {}} /></div>
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
  return <Section title="对象检查" description="按完整 Object Key 实时检查两个存储桶；不会列出或删除桶内文件。">
    <form className="flex flex-wrap items-end gap-3" onSubmit={(event) => { event.preventDefault(); setMessage(''); repair.reset(); lookup.mutate(key.trim()) }}><div className="min-w-60 flex-1"><FormField label="Object Key"><Input required value={key} onChange={(event) => setKey(event.target.value)} placeholder="nav/hero/desktop/… .avif" /></FormField></div><Button disabled={lookup.isPending || repair.isPending}>检查</Button></form>
    {(lookup.error || repair.error) && <Alert tone="danger">{errorMessage(lookup.error || repair.error)}</Alert>}
    {message && <p role="status" className="my-3 text-sm">{message}</p>}
    {item && <div className="mt-4 grid gap-4">
      <p className="break-all text-sm">{item.object_key} · {statusLabels[item.comparison]}</p>
      <div className="grid gap-4 lg:grid-cols-2">{(['primary', 'mirror'] as const).map((key) => { const object = item[key]; const url = item[`${key}_url`]; return <div key={key} className="rounded-md border p-3 text-sm"><p className="font-medium">{key === 'primary' ? 'COS' : 'R2'} · {statusLabels[object.state] ?? object.state}</p>{object.error && <p className="mt-2 text-danger">{object.error}</p>}{object.info && <dl className="mt-3 grid gap-2"><div><dt>大小</dt><dd>{object.info.size.toLocaleString()} bytes</dd></div><div><dt>类型</dt><dd>{object.info.content_type}</dd></div><div><dt>SHA256</dt><dd className="break-all font-mono text-xs">{object.info.sha256 || '无元数据'}</dd></div><div><dt>资源种类</dt><dd>{object.info.asset_kind || '—'}</dd></div><div><dt>缓存策略</dt><dd className="break-all">{object.info.cache_control || '—'}</dd></div></dl>}<div className="mt-3 flex flex-wrap gap-3"><a href={url} className="self-center underline" target="_blank" rel="noreferrer">打开资源</a><Button variant="secondary" onClick={() => void copy(url)}>复制 {key === 'primary' ? 'Primary' : 'Mirror'} URL</Button></div></div> })}</div>
      {canManage && <div><Button disabled={repair.isPending || item.primary.state !== 'ready' || item.object_key === 'system/probes/cdn.bin'} onClick={() => repair.mutate(item.object_key)}>从 COS 修复 R2</Button></div>}
    </div>}
  </Section>
}

function CDNPurge({ provider, config, submitted }: { provider: 'edgeone' | 'cloudflare'; config: CDN; submitted: (id: string) => void }) {
  const auth = useAuth()
  const canManage = auth.can('cloudops.manage')
  const [type, setType] = useState<'url' | 'prefix' | 'host'>('url')
  const [targets, setTargets] = useState('')
  const mutation = useMutation({ mutationFn: (request: { type: string; targets: string[] } | 'all') => sendJSON<Operation<Purge>>(`${base}/${provider}/${request === 'all' ? 'purge-all' : 'purge'}`, 'POST', request === 'all' ? undefined : request), onSuccess: (response) => { if (response.result.job_id) submitted(response.result.job_id) } })
  const purge = (request: { type: string; targets: string[] } | 'all') => { const scope = request === 'all' ? '整个 EdgeOne Zone（包括所有加速域名）' : request.targets.join('\n'); if (window.confirm(`确认清除以下缓存？\n${scope}`)) mutation.mutate(request) }
  return <Section title={provider === 'edgeone' ? 'EdgeOne' : 'Cloudflare'} description={config.configured ? `已配置 · ${config.asset_host}` : '未配置'}>
    {mutation.error && <Alert tone="danger">{errorMessage(mutation.error)}</Alert>}
    {mutation.data && <div className="mb-3 grid gap-2 text-sm" role="status"><p>{statusLabels[mutation.data.result.status] ?? mutation.data.result.status} · {mutation.data.result.job_id}</p>{mutation.data.result.failures?.map((failure, i) => <Alert key={i} tone="danger">{failure.reason}：{failure.targets.join('、')}</Alert>)}{mutation.data.warnings?.map((warning) => <Alert key={warning} tone="warning">{warning}</Alert>)}</div>}
    {canManage && <fieldset disabled={!config.configured || mutation.isPending} className="grid gap-3">
      <FormField label="清除方式"><select className="h-9 rounded-md border bg-surface px-3" value={type} onChange={(event) => setType(event.target.value as typeof type)}><option value="url">URL · 指定资源</option><option value="prefix">Prefix · 路径前缀</option><option value="host">Host · 指定域名</option></select></FormField>
      <FormField label="目标（每行一个，最多 20 个）" help={type === 'host' ? '填写已配置的域名，不带协议或路径。' : '填写包含 https:// 的完整地址。'}><Textarea value={targets} onChange={(event) => setTargets(event.target.value)} placeholder={type === 'host' ? config.asset_host : `https://${config.asset_host}/nav/`} /></FormField>
      <Button disabled={!targets.trim()} onClick={() => purge({ type, targets: targets.split(/\r?\n/).map((target) => target.trim()).filter(Boolean) })}>提交清缓存</Button>
      {provider === 'edgeone' && config.main_host && <Button variant="secondary" onClick={() => purge({ type: 'host', targets: [config.main_host!] })}>清除主站缓存 · {config.main_host}</Button>}
    </fieldset>}
    {provider === 'edgeone' && auth.can('cloudops.purge_all') && <div className="mt-5 border-t pt-4"><p className="mb-2 text-xs text-muted-foreground">此操作覆盖整个 EdgeOne Zone 的所有域名。</p><Button variant="danger" disabled={!config.configured || mutation.isPending} onClick={() => purge('all')}>清除整个 EdgeOne Zone 缓存</Button></div>}
  </Section>
}

function PurgeTasks({ jobID, setJobID, configured }: { jobID: string; setJobID: (id: string) => void; configured: boolean }) {
  const [page, setPage] = useState(1)
  const query = useQuery({ queryKey: ['cloud-purge-tasks', jobID, page], queryFn: () => getJSON<{ total: number; list: Task[] }>(`${base}/edgeone/purge-tasks?job_id=${encodeURIComponent(jobID)}&page_num=${page}`), enabled: configured })
  return <Section title="EdgeOne 清缓存任务" actions={<Button variant="secondary" disabled={!configured || query.isFetching} onClick={() => void query.refetch()}>刷新任务</Button>}>
    <FormField label="任务 ID" help="留空显示过去 24 小时的任务。"><Input value={jobID} onChange={(event) => { setJobID(event.target.value); setPage(1) }} /></FormField>
    {query.error && <Alert tone="danger">{errorMessage(query.error)}</Alert>}
    <div className="mt-4 divide-y">{query.data?.list.map((task, index) => <div key={`${task.job_id}-${index}`} className="grid gap-1 py-3 text-sm"><p className="break-all">{task.target}</p><p>{statusLabels[task.status] ?? task.status} · {task.type} · {task.created_at}</p><p className="font-mono text-xs text-muted-foreground">{task.job_id}</p>{task.failure && <p className="text-danger">{task.failure}</p>}</div>)}</div>
    {query.data?.total === 0 && <p className="py-3 text-sm text-muted-foreground">暂无任务；新提交的任务可能需要稍后刷新。</p>}
    <div className="flex justify-end gap-2"><Button variant="secondary" disabled={page === 1} onClick={() => setPage(page - 1)}>上一页</Button><Button variant="secondary" disabled={page * 20 >= (query.data?.total ?? 0)} onClick={() => setPage(page + 1)}>下一页</Button></div>
  </Section>
}
