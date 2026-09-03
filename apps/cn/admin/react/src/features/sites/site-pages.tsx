import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Braces, Check, LoaderCircle, Pencil, Plus, RefreshCw, Star, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useForm, type Resolver } from 'react-hook-form'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { Detail, DetailGrid, FormField, FormSection, PageHeader, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { TechnicalDetails } from '../../components/admin/technical-details'
import { DataCenterPanel, HistoryPanel, WorkspaceTabs, type WorkspaceTab } from '../../components/admin/workspace'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { ConfirmAction, Dialog } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { errorMessage, getJSON, listJSON, sendJSON } from '../../lib/api'
import type { CollectorTarget, OptionItem, Site, SiteSummary, SiteWorkspace } from '../../lib/types'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'

const siteSchema = z.object({
  name: z.string().trim().min(1, '请输入中文名称'), name_en: z.string(), info: z.string(), info_en: z.string(), icon: z.string(),
})
type SiteValues = z.infer<typeof siteSchema>

const classificationSchema = z.object({ country: z.string(), nsfw: z.enum(['0', '1']), welfare: z.enum(['0', '1']), group_ids: z.array(z.string()), featured: z.boolean(), featured_weight: z.coerce.number().int() })
type ClassificationValues = z.infer<typeof classificationSchema>

const targetSchema = z.object({ name: z.string().trim().min(1, '请输入域名'), prefix: z.string(), proxy: z.enum(['0', '1']), tls: z.enum(['0', '1']) })
type TargetValues = z.infer<typeof targetSchema>

const tabs: WorkspaceTab[] = [
  { key: 'overview', label: '概览' }, { key: 'content', label: '内容' }, { key: 'classification', label: '分类与展示' },
  { key: 'collection', label: '采集' }, { key: 'status', label: '数据状态' }, { key: 'history', label: '历史' },
]

export function SiteListPage() {
  const navigate = useNavigate()
  const auth = useAuth()
  const [params, setParams] = useSearchParams()
  const page = Math.max(1, Number(params.get('page') || 1))
  const pageSize = [20, 50, 100].includes(Number(params.get('page_size'))) ? Number(params.get('page_size')) : 50
  const search = params.get('search') ?? ''
  const query = useQuery({ queryKey: ['site-summaries', page, pageSize, search], queryFn: () => listJSON<SiteSummary>('/api/v1/nav/site-summaries', page, pageSize, search) })
  const columns = useMemo<AdminColumn<SiteSummary>[]>(() => [
    { key: 'id', header: 'ID', hidden: true },
    { key: 'name', header: '名称', render: (site) => <div><p className="font-medium text-primary">{site.name}</p>{site.name_en && <p className="text-xs text-muted-foreground">{site.name_en}</p>}</div> },
    { key: 'primary_target', header: '域名', render: (site) => site.primary_target ? <span className="font-mono text-xs">{site.primary_target}</span> : <span className="text-muted-foreground">未配置</span> },
    { key: 'group_names', header: '分类 / 分组', render: (site) => <div className="flex flex-wrap gap-1">{site.group_names.length ? site.group_names.slice(0, 3).map((name) => <span key={name} className="rounded bg-surface-muted px-2 py-0.5 text-xs">{name}</span>) : <span className="text-muted-foreground">未分组</span>}</div> },
    { key: 'featured', header: '数据状态', render: (site) => <div className="flex gap-1.5"><StatusBadge tone={site.primary_target ? 'success' : 'warning'}>{site.primary_target ? '采集目标已配置' : '缺少主目标'}</StatusBadge>{site.featured && <StatusBadge tone="info">精选</StatusBadge>}</div> },
    { key: 'update_time', header: '最近更新', render: (site) => formatDate(site.update_time) },
  ], [])
  const set = (key: string, value: string) => { const next = new URLSearchParams(params); if (value) next.set(key, value); else next.delete(key); setParams(next, { replace: true }) }
  return <div className="grid gap-6"><PageHeader title="网站" description="以网站为中心维护内容、分类、展示与采集配置。" eyebrow="nav.sites" actions={auth.can('content.write') && <Button onClick={() => navigate('/nav/sites/new')}><Plus className="size-4" />新增网站</Button>} /><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={pageSize} search={search} onSearchChange={(value) => { const next = new URLSearchParams(params); next.set('page', '1'); if (value) next.set('search', value); else next.delete('search'); setParams(next, { replace: true }) }} onPageChange={(value) => set('page', String(value))} onPageSizeChange={(value) => { const next = new URLSearchParams(params); next.set('page', '1'); next.set('page_size', String(value)); setParams(next, { replace: true }) }} onRowClick={(row) => navigate(`/nav/sites/${row.id}`)} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /></div>
}

function SiteContentForm({ site, creating = false }: { site?: Site; creating?: boolean }) {
  const navigate = useNavigate()
  const client = useQueryClient()
  const { toast } = useToast()
  const [operationError, setOperationError] = useState('')
  const form = useForm<SiteValues>({ resolver: zodResolver(siteSchema), defaultValues: { name: site?.name ?? '', name_en: site?.name_en ?? '', info: site?.info ?? '', info_en: site?.info_en ?? '', icon: site?.icon ?? '' } })
  useEffect(() => form.reset({ name: site?.name ?? '', name_en: site?.name_en ?? '', info: site?.info ?? '', info_en: site?.info_en ?? '', icon: site?.icon ?? '' }), [form, site])
  useUnsavedChanges(form.formState.isDirty)
  const mutation = useMutation({ mutationFn: async (values: SiteValues) => {
    const payload = { ...values, icon: values.icon || null, country: site?.country ?? null, nsfw: site?.nsfw ?? '0', welfare: site?.welfare ?? '0' }
    return creating ? sendJSON<Site>('/api/v1/nav/sites', 'POST', payload) : sendJSON<Site>(`/api/v1/nav/sites/${site!.id}`, 'PUT', payload)
  }, onSuccess: async (saved) => { form.reset({ name: saved.name, name_en: saved.name_en, info: saved.info, info_en: saved.info_en, icon: saved.icon ?? '' }); await client.invalidateQueries({ queryKey: ['site'] }); await client.invalidateQueries({ queryKey: ['site-summaries'] }); toast(creating ? '网站已创建' : '网站内容已保存'); if (creating) navigate(`/nav/sites/${saved.id}`, { replace: true }) }, onError: (error) => setOperationError(errorMessage(error)) })
  return <Section title={creating ? '新增网站' : '内容'} description="按业务语言维护网站名称、简介与图标。"><form className="grid gap-6" onSubmit={form.handleSubmit((values) => { setOperationError(''); mutation.mutate(values) })}>{operationError && <Alert tone="danger">{operationError}</Alert>}<FormSection title="基本内容"><div className="grid gap-4 md:grid-cols-2"><FormField label="中文名称" required error={form.formState.errors.name?.message}><Input {...form.register('name')} /></FormField><FormField label="英文名称" error={form.formState.errors.name_en?.message}><Input {...form.register('name_en')} /></FormField></div><FormField label="中文简介" error={form.formState.errors.info?.message}><Textarea {...form.register('info')} /></FormField><FormField label="英文简介" error={form.formState.errors.info_en?.message}><Textarea {...form.register('info_en')} /></FormField></FormSection><FormSection title="媒体"><FormField label="Icon" help="填写现有图标 URL 或资源标识。"><Input {...form.register('icon')} /></FormField></FormSection><div className="flex justify-end"><Button disabled={mutation.isPending || !form.formState.isDirty}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}{creating ? '创建网站' : '保存内容'}</Button></div></form></Section>
}

function MultiOptions({ options, selected, onChange, disabled }: { options: OptionItem[]; selected: string[]; onChange: (next: string[]) => void; disabled?: boolean }) {
  return <div className="grid max-h-56 gap-1 overflow-auto rounded-md border p-2 md:grid-cols-2">{options.map((option) => { const active = selected.includes(String(option.id)); return <button key={option.id} type="button" disabled={disabled} onClick={() => onChange(active ? selected.filter((id) => id !== String(option.id)) : [...selected, String(option.id)])} className="flex items-center gap-2 rounded px-2 py-2 text-left text-sm hover:bg-surface-muted"><span className={active ? 'grid size-4 place-items-center rounded border border-primary bg-primary text-primary-foreground' : 'size-4 rounded border'}>{active && <Check className="size-3" />}</span>{option.label}</button> })}</div>
}

function SiteClassificationForm({ workspace }: { workspace: SiteWorkspace }) {
  const client = useQueryClient()
  const { toast } = useToast()
  const [operationError, setOperationError] = useState('')
  const options = useQuery({ queryKey: ['options', 'site-groups'], queryFn: () => listJSON<OptionItem>('/api/v1/options/site-groups', 1, 200) })
  const form = useForm<ClassificationValues>({ resolver: zodResolver(classificationSchema) as unknown as Resolver<ClassificationValues>, defaultValues: { country: workspace.site.country ?? '', nsfw: workspace.site.nsfw as '0' | '1', welfare: workspace.site.welfare as '0' | '1', group_ids: workspace.groups.map((group) => String(group.group_id)), featured: Boolean(workspace.featured), featured_weight: workspace.featured?.weight ?? 0 } })
  useUnsavedChanges(form.formState.isDirty)
  const mutation = useMutation({ mutationFn: async (values: ClassificationValues) => {
    const site = workspace.site
    await sendJSON(`/api/v1/nav/sites/${site.id}`, 'PUT', { name: site.name, name_en: site.name_en, info: site.info, info_en: site.info_en, icon: site.icon, country: values.country || null, nsfw: values.nsfw, welfare: values.welfare })
    await sendJSON('/api/v1/nav/site-group-maps/bulk-replace', 'PUT', { owner_id: site.id, ids: values.group_ids.map(Number) })
    if (values.featured && workspace.featured) await sendJSON(`/api/v1/nav/featured-sites/${workspace.featured.id}`, 'PUT', { site_id: site.id, weight: values.featured_weight })
    else if (values.featured) await sendJSON('/api/v1/nav/featured-sites', 'POST', { site_id: site.id, weight: values.featured_weight })
    else if (workspace.featured) await sendJSON(`/api/v1/nav/featured-sites/${workspace.featured.id}`, 'DELETE')
  }, onSuccess: async () => { await client.invalidateQueries({ queryKey: ['site', workspace.site.id] }); await client.invalidateQueries({ queryKey: ['site-summaries'] }); form.reset(form.getValues()); toast('分类与展示已保存') }, onError: (error) => setOperationError(errorMessage(error)) })
  const selected = form.watch('group_ids')
  return <Section title="分类与展示" description="以业务关系维护分组和精选状态，不直接操作映射表。"><form className="grid gap-6" onSubmit={form.handleSubmit((values) => { setOperationError(''); mutation.mutate(values) })}>{operationError && <Alert tone="danger">{operationError}</Alert>}<FormSection title="分类属性"><div className="grid gap-4 md:grid-cols-3"><FormField label="国家 / 地区"><Input {...form.register('country')} /></FormField><FormField label="NSFW"><Select value={form.watch('nsfw')} onValueChange={(value) => form.setValue('nsfw', value as ClassificationValues['nsfw'], { shouldDirty: true })} options={[{ value: '0', label: '否' }, { value: '1', label: '是' }]} /></FormField><FormField label="公益属性"><Select value={form.watch('welfare')} onValueChange={(value) => form.setValue('welfare', value as ClassificationValues['welfare'], { shouldDirty: true })} options={[{ value: '0', label: '否' }, { value: '1', label: '是' }]} /></FormField></div></FormSection><FormSection title="网站分组" description="保存后会一次替换当前网站的分组集合。"><MultiOptions options={options.data?.list ?? []} selected={selected} onChange={(value) => form.setValue('group_ids', value, { shouldDirty: true })} disabled={mutation.isPending} /></FormSection><FormSection title="精选展示"><label className="flex items-center gap-2 text-sm"><input type="checkbox" className="size-4 accent-primary" {...form.register('featured')} />加入精选网站</label>{form.watch('featured') && <FormField label="展示权重" help="权重越高越靠前。"><Input type="number" {...form.register('featured_weight')} /></FormField>}</FormSection><div className="flex justify-end"><Button disabled={mutation.isPending || !form.formState.isDirty}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}保存分类与展示</Button></div></form></Section>
}

function TargetEditor({ siteId, target, open, onOpenChange }: { siteId: number; target: CollectorTarget | null; open: boolean; onOpenChange: (open: boolean) => void }) {
  const client = useQueryClient()
  const { toast } = useToast()
  const [operationError, setOperationError] = useState('')
  const form = useForm<TargetValues>({ resolver: zodResolver(targetSchema), defaultValues: { name: '', prefix: '', proxy: '0', tls: '1' } })
  useEffect(() => form.reset({ name: target?.name ?? '', prefix: target?.prefix ?? '', proxy: (target?.proxy ?? '0') as '0' | '1', tls: (target?.tls ?? '1') as '0' | '1' }), [form, target, open])
  const mutation = useMutation({ mutationFn: (values: TargetValues) => sendJSON(target ? `/api/v1/nav/collector-domains/${target.id}` : '/api/v1/nav/collector-domains', target ? 'PUT' : 'POST', { site_id: siteId, name: values.name, prefix: values.prefix || null, proxy: values.proxy, tls: values.tls }), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['site', siteId] }); toast(target ? '采集目标已保存' : '采集目标已创建'); onOpenChange(false) }, onError: (error) => setOperationError(errorMessage(error)) })
  return <Dialog open={open} onOpenChange={onOpenChange} title={target ? '编辑采集目标' : '新增采集目标'} description="域名、TLS 与代理配置" footer={<><Button variant="secondary" onClick={() => onOpenChange(false)}>取消</Button><Button disabled={mutation.isPending} onClick={form.handleSubmit((values) => mutation.mutate(values))}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}保存</Button></>}>
    <form className="grid gap-4" onSubmit={(event) => event.preventDefault()}>{operationError && <Alert tone="danger">{operationError}</Alert>}<FormField label="域名" required error={form.formState.errors.name?.message}><Input {...form.register('name')} placeholder="example.com" /></FormField><FormField label="前缀" help="例如 www.；留空表示无前缀。"><Input {...form.register('prefix')} /></FormField><div className="grid grid-cols-2 gap-4"><FormField label="TLS"><Select value={form.watch('tls')} onValueChange={(value) => form.setValue('tls', value as TargetValues['tls'], { shouldDirty: true })} options={[{ value: '1', label: '启用' }, { value: '0', label: '禁用' }]} /></FormField><FormField label="代理"><Select value={form.watch('proxy')} onValueChange={(value) => form.setValue('proxy', value as TargetValues['proxy'], { shouldDirty: true })} options={[{ value: '0', label: '直连' }, { value: '1', label: '使用代理' }]} /></FormField></div></form>
  </Dialog>
}

function SiteCollection({ workspace }: { workspace: SiteWorkspace }) {
  const auth = useAuth()
  const client = useQueryClient()
  const { toast } = useToast()
  const [editor, setEditor] = useState<{ open: boolean; target: CollectorTarget | null }>({ open: false, target: null })
  const [deleting, setDeleting] = useState<CollectorTarget | null>(null)
  const action = useMutation({ mutationFn: ({ type, target }: { type: 'primary' | 'delete' | 'collect'; target?: CollectorTarget }) => type === 'primary' ? sendJSON(`/api/v1/nav/collector-domains/${target!.id}/primary`, 'POST', {}) : type === 'delete' ? sendJSON(`/api/v1/nav/collector-domains/${target!.id}`, 'DELETE') : sendJSON('/api/v1/collection/jobs', 'POST', { domain: 'nav', scope_type: 'site', scope_id: workspace.site.id, target: null, tasks: ['ping', 'http', 'dns', 'security_txt'] }), onSuccess: async (_, variables) => { await client.invalidateQueries({ queryKey: ['site', workspace.site.id] }); toast(variables.type === 'collect' ? '已创建网站采集任务' : variables.type === 'primary' ? '主采集目标已更新' : '采集目标已删除'); setDeleting(null) }, onError: (error) => toast(errorMessage(error), 'danger') })
  return <div className="grid gap-5"><Section title="采集配置" description="维护当前网站的采集目标、TLS 与代理。" actions={auth.can('content.write') && <Button variant="secondary" onClick={() => setEditor({ open: true, target: null })}><Plus className="size-4" />新增目标</Button>}>
    <div className="grid gap-2">{workspace.targets.length === 0 ? <Alert tone="warning">尚未配置采集目标。</Alert> : workspace.targets.map((target) => <div key={target.id} className="flex items-center justify-between rounded-md border p-3"><div><div className="flex items-center gap-2"><span className="font-mono text-sm">{`${target.prefix ?? ''}${target.name}`}</span>{target.primary && <StatusBadge tone="success">Primary</StatusBadge>}</div><div className="mt-1 flex gap-3 text-xs text-muted-foreground"><span>TLS {target.tls === '1' ? '开启' : '关闭'}</span><span>{target.proxy === '1' ? '代理' : '直连'}</span><TechnicalLabel>target #{target.id}</TechnicalLabel></div></div>{auth.can('content.write') && <div className="flex items-center gap-1">{!target.primary && <Button variant="ghost" size="sm" disabled={action.isPending} onClick={() => action.mutate({ type: 'primary', target })}><Star className="size-3.5" />设为主目标</Button>}<Button variant="ghost" size="icon" onClick={() => setEditor({ open: true, target })}><Pencil className="size-4" /></Button><Button variant="ghost" size="icon" onClick={() => setDeleting(target)}><Trash2 className="size-4 text-danger" /></Button></div>}</div>)}</div>
  </Section>{auth.can('collection.execute') && <Section title="立即采集" description="使用现有 Collection 契约创建 HTTP、DNS、Ping 与 security.txt 任务。"><Button disabled={action.isPending || workspace.targets.length === 0} onClick={() => action.mutate({ type: 'collect' })}><RefreshCw className={action.isPending ? 'size-4 animate-spin' : 'size-4'} />立即采集</Button></Section>}<TargetEditor siteId={workspace.site.id} target={editor.target} open={editor.open} onOpenChange={(open) => setEditor((current) => ({ ...current, open }))} /><ConfirmAction open={Boolean(deleting)} onOpenChange={(open) => { if (!open) setDeleting(null) }} title="删除采集目标" description={`确定删除 ${deleting ? `${deleting.prefix ?? ''}${deleting.name}` : ''} 吗？`} busy={action.isPending} onConfirm={() => action.mutate({ type: 'delete', target: deleting! })} /></div>
}

function SiteOverview({ workspace }: { workspace: SiteWorkspace }) {
  const primary = workspace.targets.find((target) => target.primary)
  return <div className="grid gap-5"><Section title="网站概览"><DetailGrid><Detail label="中文名称">{workspace.site.name}</Detail><Detail label="英文名称">{workspace.site.name_en || '—'}</Detail><Detail label="主域名" technical={primary ? `collector_domain #${primary.id}` : 'primary_target · unknown'}>{primary ? `${primary.prefix ?? ''}${primary.name}` : <StatusBadge tone="warning">未配置</StatusBadge>}</Detail><Detail label="网站分组">{workspace.groups.length ? workspace.groups.map((group) => group.group_name).join('、') : '未分组'}</Detail><Detail label="精选展示">{workspace.featured ? <StatusBadge tone="info">精选 · 权重 {workspace.featured.weight}</StatusBadge> : <StatusBadge>普通</StatusBadge>}</Detail><Detail label="最近更新">{formatDate(workspace.site.update_time)}</Detail></DetailGrid></Section><Section title="采集摘要"><div className="grid gap-3 md:grid-cols-4"><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">Targets</p><p className="mt-1 text-2xl font-semibold">{workspace.targets.length}</p></div><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">TLS</p><p className="mt-1 text-sm font-medium">{primary ? (primary.tls === '1' ? '已启用' : '未启用') : '未知'}</p></div><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">Proxy</p><p className="mt-1 text-sm font-medium">{primary ? (primary.proxy === '1' ? '代理' : '直连') : '未知'}</p></div><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">Groups</p><p className="mt-1 text-2xl font-semibold">{workspace.groups.length}</p></div></div></Section></div>
}

function SiteDataStatus({ workspace }: { workspace: SiteWorkspace }) {
  const primary = workspace.targets.find((target) => target.primary)
  return <div className="grid gap-5"><Section title="当前可确认状态" description="只展示现有 API 可以可靠确认的配置状态。"><div className="grid gap-3 md:grid-cols-3"><div className="rounded-md border p-4"><p className="text-sm font-medium">主采集目标</p><div className="mt-2"><StatusBadge tone={primary ? 'success' : 'warning'}>{primary ? '已配置' : '未知'}</StatusBadge></div><TechnicalLabel>active_site_primary_target_v1</TechnicalLabel></div><div className="rounded-md border p-4"><p className="text-sm font-medium">TLS 采集</p><div className="mt-2"><StatusBadge tone={primary?.tls === '1' ? 'success' : primary ? 'warning' : 'neutral'}>{primary ? (primary.tls === '1' ? '已开启' : '未开启') : 'N/A'}</StatusBadge></div><TechnicalLabel>collector_domain.tls</TechnicalLabel></div><div className="rounded-md border p-4"><p className="text-sm font-medium">采集覆盖</p><div className="mt-2"><StatusBadge tone={workspace.targets.length ? 'info' : 'warning'}>{workspace.targets.length} 个 Target</StatusBadge></div><TechnicalLabel>active_target_count</TechnicalLabel></div></div></Section><DataCenterPanel title="Facts / Metrics / Changes 状态" to="/metrics?tab=entities">IPv6、TLS 1.3、security.txt、Coverage 与 Latest Fact 由现有数据域维护；实体指标和异常解释在原生 React 数据中心查看。</DataCenterPanel></div>
}

export function SiteWorkspacePage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const auth = useAuth()
  const [params, setParams] = useSearchParams()
  const [technicalOpen, setTechnicalOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const client = useQueryClient()
  const { toast } = useToast()
  const numericId = Number(id)
  const query = useQuery({ queryKey: ['site', numericId], queryFn: () => getJSON<SiteWorkspace>(`/api/v1/nav/sites/${numericId}/workspace`), enabled: id !== 'new' && Number.isInteger(numericId) && numericId > 0 })
  const deleteMutation = useMutation({ mutationFn: () => sendJSON(`/api/v1/nav/sites/${numericId}`, 'DELETE'), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['site-summaries'] }); toast('网站已删除'); navigate('/nav/sites', { replace: true }) }, onError: (error) => toast(errorMessage(error), 'danger') })
  if (id === 'new') return <div className="grid gap-6"><PageHeader title="新增网站" description="创建基础内容后进入完整 Site Workspace。" actions={<Button variant="secondary" onClick={() => navigate('/nav/sites')}><ArrowLeft className="size-4" />返回列表</Button>} /><SiteContentForm creating /></div>
  if (!Number.isInteger(numericId) || numericId <= 0) return <ErrorState title="网站标识无效" message="URL 中的网站 ID 必须是正整数。" />
  if (query.isLoading) return <LoadingState label="正在加载 Site Workspace…" />
  if (query.error || !query.data) return <ErrorState message={query.error?.message ?? '网站不存在'} onRetry={() => void query.refetch()} />
  const workspace = query.data
  const active = tabs.some((tab) => tab.key === params.get('tab')) ? params.get('tab')! : 'overview'
  const setTab = (tab: string) => { const next = new URLSearchParams(params); if (tab === 'overview') next.delete('tab'); else next.set('tab', tab); setParams(next, { replace: true }) }
  return <div className="grid gap-5"><PageHeader title={workspace.site.name} description={workspace.site.name_en || 'Site Workspace'} eyebrow={`nav.site · #${workspace.site.id}`} actions={<><Button variant="secondary" onClick={() => navigate('/nav/sites')}><ArrowLeft className="size-4" />返回列表</Button><Button variant="secondary" onClick={() => setTechnicalOpen(true)}><Braces className="size-4" />技术详情</Button>{auth.can('content.write') && <Button variant="danger" onClick={() => setDeleteOpen(true)}><Trash2 className="size-4" />删除</Button>}</>} /><WorkspaceTabs tabs={tabs} active={active} onChange={setTab} />{active === 'overview' && <SiteOverview workspace={workspace} />}{active === 'content' && (auth.can('content.write') ? <SiteContentForm site={workspace.site} /> : <Alert tone="warning">当前账号只有只读权限。</Alert>)}{active === 'classification' && (auth.can('content.write') ? <SiteClassificationForm workspace={workspace} /> : <Alert tone="warning">当前账号只有只读权限。</Alert>)}{active === 'collection' && <SiteCollection workspace={workspace} />}{active === 'status' && <SiteDataStatus workspace={workspace} />}{active === 'history' && <HistoryPanel domain="nav" entityId={workspace.site.id} />}<TechnicalDetails open={technicalOpen} onOpenChange={setTechnicalOpen} title={workspace.site.name} identifier={`gfn_site.id · ${workspace.site.id}`}><pre className="overflow-auto rounded-md bg-surface-muted p-4 font-mono text-xs leading-6">{JSON.stringify(workspace, null, 2)}</pre></TechnicalDetails><ConfirmAction open={deleteOpen} onOpenChange={setDeleteOpen} title="删除网站" description={`确定删除 ${workspace.site.name} 吗？相关历史事实会按现有后端语义保留。`} busy={deleteMutation.isPending} onConfirm={() => deleteMutation.mutate()} /></div>
}
