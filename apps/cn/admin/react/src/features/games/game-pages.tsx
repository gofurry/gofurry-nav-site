import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Braces, Check, LoaderCircle, Plus, RefreshCw, Trash2 } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Controller, useForm, type Resolver } from 'react-hook-form'
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
import { ConfirmAction } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { errorMessage, getJSON, listJSON, sendJSON } from '../../lib/api'
import type { Game, GameWorkspace, KeyValue, OptionItem } from '../../lib/types'
import { formatDate } from '../../lib/utils'
import { useAuth } from '../auth/auth-context'

const kvSchema = z.object({ key: z.string(), value: z.string() })
const gameContentSchema = z.object({
  name: z.string().trim().min(1, '请输入中文名称'), name_en: z.string(), info: z.string(), info_en: z.string(), appid: z.coerce.number().int().positive('Steam AppID 必须是正整数'),
  developers: z.array(z.string()), publishers: z.array(z.string()), header: z.string(), groups: z.array(kvSchema), links: z.array(kvSchema), resources: z.array(kvSchema),
})
type GameContentValues = z.infer<typeof gameContentSchema>

const gameClassificationSchema = z.object({ primary_tag: z.coerce.number().int().nonnegative(), secondary_tag: z.coerce.number().int().nonnegative(), tag_ids: z.array(z.string()), weight: z.coerce.number().int() })
type GameClassificationValues = z.infer<typeof gameClassificationSchema>

type SteamPrefill = Pick<Game, 'appid' | 'name' | 'name_en' | 'info' | 'info_en' | 'groups' | 'developers' | 'publishers' | 'header' | 'links'>

const tabs: WorkspaceTab[] = [
  { key: 'overview', label: '概览' }, { key: 'content', label: '内容' }, { key: 'classification', label: '分类与展示' },
  { key: 'steam', label: 'Steam 与采集' }, { key: 'status', label: '数据状态' }, { key: 'history', label: '历史' },
]

function emptyGame(): Game {
  return { id: 0, name: '', name_en: '', info: '', info_en: '', create_time: '', update_time: '', resources: [{ key: '', value: '' }], groups: [{ key: '', value: '' }], developers: [''], publishers: [''], appid: 0, header: '', links: [{ key: '', value: '' }], weight: 1000, primary_tag: 0, secondary_tag: 0 }
}

export function GameListPage() {
  const navigate = useNavigate()
  const auth = useAuth()
  const [params, setParams] = useSearchParams()
  const page = Math.max(1, Number(params.get('page') || 1))
  const pageSize = [20, 50, 100].includes(Number(params.get('page_size'))) ? Number(params.get('page_size')) : 50
  const search = params.get('search') ?? ''
  const query = useQuery({ queryKey: ['games', page, pageSize, search], queryFn: () => listJSON<Game>('/api/v1/game/games', page, pageSize, search) })
  const tagOptions = useQuery({ queryKey: ['options', 'tags'], queryFn: () => listJSON<OptionItem>('/api/v1/options/tags', 1, 200), staleTime: 60_000 })
  const tagNames = useMemo(() => new Map((tagOptions.data?.list ?? []).map((item) => [Number(item.id), item.label])), [tagOptions.data?.list])
  const columns = useMemo<AdminColumn<Game>[]>(() => [
    { key: 'id', header: 'ID', hidden: true }, { key: 'name', header: '名称', render: (game) => <div><p className="font-medium text-primary">{game.name}</p>{game.name_en && <p className="text-xs text-muted-foreground">{game.name_en}</p>}</div> },
    { key: 'appid', header: 'Steam AppID', render: (game) => <span className="font-mono text-xs">{game.appid}</span> },
    { key: 'primary_tag', header: '主要标签', render: (game) => tagNames.get(game.primary_tag) || <span className="text-muted-foreground">未配置</span> },
    { key: 'data_state', header: '数据状态', sortable: false, render: (game) => <StatusBadge tone={game.appid > 0 ? 'success' : 'warning'}>{game.appid > 0 ? 'Steam 已配置' : '缺少 AppID'}</StatusBadge> },
    { key: 'update_time', header: '最近更新', render: (game) => formatDate(game.update_time) },
  ], [tagNames])
  const set = (key: string, value: string) => { const next = new URLSearchParams(params); if (value) next.set(key, value); else next.delete(key); setParams(next, { replace: true }) }
  return <div className="grid gap-6"><PageHeader title="游戏" description="以游戏为中心维护内容、标签、Steam 与采集配置。" eyebrow="game.games" actions={auth.can('content.write') && <Button onClick={() => navigate('/game/games/new')}><Plus className="size-4" />新增游戏</Button>} /><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={pageSize} search={search} onSearchChange={(value) => { const next = new URLSearchParams(params); next.set('page', '1'); if (value) next.set('search', value); else next.delete('search'); setParams(next, { replace: true }) }} onPageChange={(value) => set('page', String(value))} onPageSizeChange={(value) => { const next = new URLSearchParams(params); next.set('page', '1'); next.set('page_size', String(value)); setParams(next, { replace: true }) }} onRowClick={(row) => navigate(`/game/games/${row.id}`)} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /></div>
}

function LinesEditor({ value, onChange, placeholder }: { value: string[]; onChange: (value: string[]) => void; placeholder?: string }) {
  return <Textarea value={value.join('\n')} onChange={(event) => onChange(event.target.value.split('\n').map((item) => item.trim()).filter(Boolean))} placeholder={placeholder} />
}

function KeyValueEditor({ value, onChange }: { value: KeyValue[]; onChange: (value: KeyValue[]) => void }) {
  const items = value.length ? value : [{ key: '', value: '' }]
  return <div className="grid gap-2">{items.map((item, index) => <div key={index} className="grid grid-cols-[10rem_1fr_auto] gap-2"><Input value={item.key} onChange={(event) => onChange(items.map((entry, current) => current === index ? { ...entry, key: event.target.value } : entry))} placeholder="键" /><Input value={item.value} onChange={(event) => onChange(items.map((entry, current) => current === index ? { ...entry, value: event.target.value } : entry))} placeholder="值 / URL" /><Button type="button" variant="ghost" size="icon" aria-label="删除此项" onClick={() => onChange(items.filter((_, current) => current !== index))}><Trash2 className="size-4" /></Button></div>)}<Button type="button" variant="secondary" size="sm" className="w-fit" onClick={() => onChange([...items, { key: '', value: '' }])}><Plus className="size-3.5" />新增一项</Button></div>
}

function gameContentValues(game: Game): GameContentValues {
  return { name: game.name, name_en: game.name_en, info: game.info, info_en: game.info_en, appid: game.appid, developers: game.developers ?? [], publishers: game.publishers ?? [], header: game.header, groups: game.groups ?? [], links: game.links ?? [], resources: game.resources ?? [] }
}

function GameContentForm({ game, creating = false }: { game: Game; creating?: boolean }) {
  const navigate = useNavigate()
  const client = useQueryClient()
  const { toast } = useToast()
  const [operationError, setOperationError] = useState('')
  const form = useForm<GameContentValues>({ resolver: zodResolver(gameContentSchema) as unknown as Resolver<GameContentValues>, defaultValues: gameContentValues(game) })
  useEffect(() => form.reset(gameContentValues(game)), [form, game])
  useUnsavedChanges(form.formState.isDirty)
  const mutation = useMutation({ mutationFn: (values: GameContentValues) => {
    const payload = { ...values, developers: values.developers.filter(Boolean), publishers: values.publishers.filter(Boolean), groups: values.groups.filter((item) => item.key || item.value), links: values.links.filter((item) => item.key || item.value), resources: values.resources.filter((item) => item.key || item.value), weight: game.weight, primary_tag: game.primary_tag, secondary_tag: game.secondary_tag }
    return creating ? sendJSON<Game>('/api/v1/game/games', 'POST', payload) : sendJSON<Game>(`/api/v1/game/games/${game.id}`, 'PUT', payload)
  }, onSuccess: async (saved) => { form.reset(gameContentValues(saved)); await client.invalidateQueries({ queryKey: ['game'] }); await client.invalidateQueries({ queryKey: ['games'] }); toast(creating ? '游戏已创建' : '游戏内容已保存'); if (creating) navigate(`/game/games/${saved.id}`, { replace: true }) }, onError: (error) => setOperationError(errorMessage(error)) })
  const prefill = useMutation({ mutationFn: () => getJSON<SteamPrefill>(`/api/v1/game/games/steam-prefill?appid=${form.getValues('appid')}`), onSuccess: (data) => { (['name', 'name_en', 'info', 'info_en', 'developers', 'publishers', 'header', 'groups', 'links'] as const).forEach((key) => form.setValue(key, data[key], { shouldDirty: true })); toast('已加载 Steam 预填内容', 'info') }, onError: (error) => toast(errorMessage(error), 'danger') })
  return <Section title={creating ? '新增游戏' : '内容'} description="按业务意义组织基本内容、创作者、媒体与外部资源。"><form className="grid gap-7" onSubmit={form.handleSubmit((values) => { setOperationError(''); mutation.mutate(values) })}>{operationError && <Alert tone="danger">{operationError}</Alert>}<FormSection title="基本内容"><div className="grid gap-4 md:grid-cols-2"><FormField label="中文名称" required error={form.formState.errors.name?.message}><Input {...form.register('name')} /></FormField><FormField label="英文名称"><Input {...form.register('name_en')} /></FormField></div><FormField label="中文简介"><Textarea {...form.register('info')} /></FormField><FormField label="英文简介"><Textarea {...form.register('info_en')} /></FormField><div className="grid gap-4 md:grid-cols-[1fr_auto]"><FormField label="Steam AppID" required error={form.formState.errors.appid?.message}><Input type="number" {...form.register('appid')} /></FormField><Button type="button" variant="secondary" className="self-end" disabled={prefill.isPending || Number(form.watch('appid')) <= 0} onClick={() => prefill.mutate()}>{prefill.isPending && <LoaderCircle className="size-4 animate-spin" />}从 Steam 预填</Button></div></FormSection><FormSection title="创作者"><Controller name="developers" control={form.control} render={({ field }) => <FormField label="开发者" help="每行一个开发者。"><LinesEditor value={field.value} onChange={field.onChange} /></FormField>} /><Controller name="publishers" control={form.control} render={({ field }) => <FormField label="发行商" help="每行一个发行商。"><LinesEditor value={field.value} onChange={field.onChange} /></FormField>} /></FormSection><FormSection title="媒体"><FormField label="封面 / Header URL"><Input {...form.register('header')} /></FormField></FormSection><FormSection title="外部资源"><Controller name="groups" control={form.control} render={({ field }) => <FormField label="社群"><KeyValueEditor value={field.value} onChange={field.onChange} /></FormField>} /><Controller name="links" control={form.control} render={({ field }) => <FormField label="第三方链接"><KeyValueEditor value={field.value} onChange={field.onChange} /></FormField>} /><Controller name="resources" control={form.control} render={({ field }) => <FormField label="资源"><KeyValueEditor value={field.value} onChange={field.onChange} /></FormField>} /></FormSection><div className="flex justify-end"><Button disabled={mutation.isPending || !form.formState.isDirty}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}{creating ? '创建游戏' : '保存内容'}</Button></div></form></Section>
}

function TagMultiSelect({ options, selected, onChange }: { options: OptionItem[]; selected: string[]; onChange: (value: string[]) => void }) {
  return <div className="grid max-h-64 gap-1 overflow-auto rounded-md border p-2 md:grid-cols-2 xl:grid-cols-3">{options.map((option) => { const active = selected.includes(String(option.id)); return <button key={option.id} type="button" onClick={() => onChange(active ? selected.filter((id) => id !== String(option.id)) : [...selected, String(option.id)])} className="flex items-center gap-2 rounded px-2 py-2 text-sm hover:bg-surface-muted"><span className={active ? 'grid size-4 place-items-center rounded border border-primary bg-primary text-primary-foreground' : 'size-4 rounded border'}>{active && <Check className="size-3" />}</span>{option.label}</button> })}</div>
}

function GameClassificationForm({ workspace }: { workspace: GameWorkspace }) {
  const client = useQueryClient()
  const { toast } = useToast()
  const [operationError, setOperationError] = useState('')
  const options = useQuery({ queryKey: ['options', 'tags'], queryFn: () => listJSON<OptionItem>('/api/v1/options/tags', 1, 200) })
  const form = useForm<GameClassificationValues>({ resolver: zodResolver(gameClassificationSchema) as unknown as Resolver<GameClassificationValues>, defaultValues: { primary_tag: workspace.game.primary_tag, secondary_tag: workspace.game.secondary_tag, tag_ids: workspace.tags.map((tag) => String(tag.tag_id)), weight: workspace.game.weight } })
  useUnsavedChanges(form.formState.isDirty)
  const mutation = useMutation({ mutationFn: async (values: GameClassificationValues) => {
    const game = workspace.game
    await sendJSON(`/api/v1/game/games/${game.id}`, 'PUT', { name: game.name, name_en: game.name_en, info: game.info, info_en: game.info_en, resources: game.resources, groups: game.groups, developers: game.developers, publishers: game.publishers, appid: game.appid, header: game.header, links: game.links, weight: values.weight, primary_tag: values.primary_tag, secondary_tag: values.secondary_tag })
    await sendJSON('/api/v1/game/tag-maps/bulk-replace', 'PUT', { owner_id: game.id, ids: values.tag_ids.map(Number) })
  }, onSuccess: async () => { await client.invalidateQueries({ queryKey: ['game', workspace.game.id] }); await client.invalidateQueries({ queryKey: ['games'] }); form.reset(form.getValues()); toast('游戏分类与展示已保存') }, onError: (error) => setOperationError(errorMessage(error)) })
  const tagOptions = options.data?.list ?? []
  const selectOptions = [{ value: '0', label: '无' }, ...tagOptions.map((option) => ({ value: String(option.id), label: option.label }))]
  return <Section title="分类与展示" description="主标签、副标签与完整标签集合在同一工作流维护。"><form className="grid gap-6" onSubmit={form.handleSubmit((values) => { setOperationError(''); mutation.mutate(values) })}>{operationError && <Alert tone="danger">{operationError}</Alert>}<FormSection title="主要分类"><div className="grid gap-4 md:grid-cols-3"><FormField label="主要标签"><Select value={String(form.watch('primary_tag'))} onValueChange={(value) => form.setValue('primary_tag', Number(value), { shouldDirty: true })} options={selectOptions} /></FormField><FormField label="次要标签"><Select value={String(form.watch('secondary_tag'))} onValueChange={(value) => form.setValue('secondary_tag', Number(value), { shouldDirty: true })} options={selectOptions} /></FormField><FormField label="展示权重"><Input type="number" {...form.register('weight')} /></FormField></div></FormSection><FormSection title="全部标签" description="保存后一次替换 Game 的 Tag Map 集合。"><TagMultiSelect options={tagOptions} selected={form.watch('tag_ids')} onChange={(value) => form.setValue('tag_ids', value, { shouldDirty: true })} /></FormSection><div className="flex justify-end"><Button disabled={mutation.isPending || !form.formState.isDirty}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}保存分类与展示</Button></div></form></Section>
}

function GameSteamCollection({ game }: { game: Game }) {
  const auth = useAuth()
  const { toast } = useToast()
  const collection = useMutation({ mutationFn: () => sendJSON('/api/v1/collection/jobs', 'POST', { domain: 'game', scope_type: 'game', scope_id: game.id, target: null, tasks: ['details', 'news', 'players'] }), onSuccess: () => toast('已创建游戏采集任务'), onError: (error) => toast(errorMessage(error), 'danger') })
  return <div className="grid gap-5"><Section title="Steam 标识"><DetailGrid><Detail label="Steam AppID" technical="gfg_game.appid">{game.appid}</Detail><Detail label="Canonical Release"><StatusBadge>当前 Admin API 未暴露</StatusBadge></Detail><Detail label="First Available"><StatusBadge>当前 Admin API 未暴露</StatusBadge></Detail><Detail label="Canonical Languages"><StatusBadge>当前 Admin API 未暴露</StatusBadge></Detail></DetailGrid></Section><Section title="采集能力" description="沿用现有 Game Collection 的 details / news / players 任务语义。"><div className="mb-4 flex flex-wrap gap-2"><StatusBadge tone="info">Details</StatusBadge><StatusBadge tone="info">News</StatusBadge><StatusBadge tone="info">Players</StatusBadge></div>{auth.can('collection.execute') ? <Button disabled={collection.isPending} onClick={() => collection.mutate()}><RefreshCw className={collection.isPending ? 'size-4 animate-spin' : 'size-4'} />立即刷新</Button> : <Alert tone="warning">当前账号没有 <TechnicalLabel>collection.execute</TechnicalLabel>。</Alert>}</Section></div>
}

function GameOverview({ workspace }: { workspace: GameWorkspace }) {
  return <div className="grid gap-5"><Section title="游戏概览"><DetailGrid><Detail label="中文名称">{workspace.game.name}</Detail><Detail label="英文名称">{workspace.game.name_en || '—'}</Detail><Detail label="Steam AppID" technical="gfg_game.appid">{workspace.game.appid}</Detail><Detail label="主要标签" technical={`tag_id · ${workspace.game.primary_tag}`}>{workspace.tags.find((tag) => tag.tag_id === workspace.game.primary_tag)?.tag_name || '未配置'}</Detail><Detail label="标签集合">{workspace.tags.length ? workspace.tags.map((tag) => tag.tag_name).join('、') : '未配置'}</Detail><Detail label="最近更新">{formatDate(workspace.game.update_time)}</Detail></DetailGrid></Section><Section title="内容摘要"><div className="grid gap-3 md:grid-cols-4"><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">开发者</p><p className="mt-1 text-2xl font-semibold">{workspace.game.developers.length}</p></div><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">发行商</p><p className="mt-1 text-2xl font-semibold">{workspace.game.publishers.length}</p></div><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">标签</p><p className="mt-1 text-2xl font-semibold">{workspace.tags.length}</p></div><div className="rounded-md bg-surface-muted p-4"><p className="text-xs text-muted-foreground">外部资源</p><p className="mt-1 text-2xl font-semibold">{workspace.game.links.length + workspace.game.resources.length}</p></div></div></Section></div>
}

function GameDataStatus({ game }: { game: Game }) {
  return <div className="grid gap-5"><Section title="当前可确认状态" description="不推断尚未由 Admin API 暴露的事实值。"><div className="grid gap-3 md:grid-cols-3"><div className="rounded-md border p-4"><p className="text-sm font-medium">Steam 标识</p><div className="mt-2"><StatusBadge tone={game.appid > 0 ? 'success' : 'warning'}>{game.appid > 0 ? '可采集' : '缺失'}</StatusBadge></div><TechnicalLabel>game.appid</TechnicalLabel></div><div className="rounded-md border p-4"><p className="text-sm font-medium">内容完整性</p><div className="mt-2"><StatusBadge tone={game.name && game.info ? 'success' : 'warning'}>{game.name && game.info ? '基本内容完整' : '需要补充'}</StatusBadge></div><TechnicalLabel>content readiness</TechnicalLabel></div><div className="rounded-md border p-4"><p className="text-sm font-medium">分类状态</p><div className="mt-2"><StatusBadge tone={game.primary_tag > 0 ? 'success' : 'neutral'}>{game.primary_tag > 0 ? '已配置主标签' : '未配置'}</StatusBadge></div><TechnicalLabel>primary_tag</TechnicalLabel></div></div></Section><DataCenterPanel title="Player / Price / Release Facts" to="/metrics?tab=entities">Windows、Linux、Free state、Player Facts、Price Facts、Release Facts 与 Latest Fact 在原生 React 数据中心查看。</DataCenterPanel></div>
}

function ExistingGameWorkspace({ numericId }: { numericId: number }) {
  const navigate = useNavigate()
  const auth = useAuth()
  const client = useQueryClient()
  const { toast } = useToast()
  const [params, setParams] = useSearchParams()
  const [technicalOpen, setTechnicalOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const query = useQuery({ queryKey: ['game', numericId], queryFn: () => getJSON<GameWorkspace>(`/api/v1/game/games/${numericId}/workspace`) })
  const deleteMutation = useMutation({ mutationFn: () => sendJSON(`/api/v1/game/games/${numericId}`, 'DELETE'), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['games'] }); toast('游戏已删除'); navigate('/game/games', { replace: true }) }, onError: (error) => toast(errorMessage(error), 'danger') })
  if (query.isLoading) return <LoadingState label="正在加载 Game Workspace…" />
  if (query.error || !query.data) return <ErrorState message={query.error?.message ?? '游戏不存在'} onRetry={() => void query.refetch()} />
  const workspace = query.data
  const active = tabs.some((tab) => tab.key === params.get('tab')) ? params.get('tab')! : 'overview'
  const setTab = (tab: string) => { const next = new URLSearchParams(params); if (tab === 'overview') next.delete('tab'); else next.set('tab', tab); setParams(next, { replace: true }) }
  return <div className="grid gap-5"><PageHeader title={workspace.game.name} description={workspace.game.name_en || 'Game Workspace'} eyebrow={`game.game · #${workspace.game.id}`} actions={<><Button variant="secondary" onClick={() => navigate('/game/games')}><ArrowLeft className="size-4" />返回列表</Button><Button variant="secondary" onClick={() => setTechnicalOpen(true)}><Braces className="size-4" />技术详情</Button>{auth.can('content.write') && <Button variant="danger" onClick={() => setDeleteOpen(true)}><Trash2 className="size-4" />删除</Button>}</>} /><WorkspaceTabs tabs={tabs} active={active} onChange={setTab} />{active === 'overview' && <GameOverview workspace={workspace} />}{active === 'content' && (auth.can('content.write') ? <GameContentForm game={workspace.game} /> : <Alert tone="warning">当前账号只有只读权限。</Alert>)}{active === 'classification' && (auth.can('content.write') ? <GameClassificationForm workspace={workspace} /> : <Alert tone="warning">当前账号只有只读权限。</Alert>)}{active === 'steam' && <GameSteamCollection game={workspace.game} />}{active === 'status' && <GameDataStatus game={workspace.game} />}{active === 'history' && <HistoryPanel domain="game" entityId={workspace.game.id} />}<TechnicalDetails open={technicalOpen} onOpenChange={setTechnicalOpen} title={workspace.game.name} identifier={`gfg_game.id · ${workspace.game.id}`}><pre className="overflow-auto rounded-md bg-surface-muted p-4 font-mono text-xs leading-6">{JSON.stringify(workspace, null, 2)}</pre></TechnicalDetails><ConfirmAction open={deleteOpen} onOpenChange={setDeleteOpen} title="删除游戏" description={`确定删除 ${workspace.game.name} 吗？历史数据按现有后端语义保留。`} busy={deleteMutation.isPending} onConfirm={() => deleteMutation.mutate()} /></div>
}

export function GameWorkspacePage() {
  const { id } = useParams()
  const navigate = useNavigate()
  if (id === 'new') return <div className="grid gap-6"><PageHeader title="新增游戏" description="可先填写 Steam AppID 获取预填内容，再进入完整 Game Workspace。" actions={<Button variant="secondary" onClick={() => navigate('/game/games')}><ArrowLeft className="size-4" />返回列表</Button>} /><GameContentForm game={emptyGame()} creating /></div>
  const numericId = Number(id)
  if (!Number.isInteger(numericId) || numericId <= 0) return <ErrorState title="游戏标识无效" message="URL 中的游戏 ID 必须是正整数。" />
  return <ExistingGameWorkspace numericId={numericId} />
}
