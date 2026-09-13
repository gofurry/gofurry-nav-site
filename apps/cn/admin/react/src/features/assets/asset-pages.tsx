import { useMutation, useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useToast } from '../../app/toast'
import { FormField, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { errorMessage, listJSON, sendForm, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'
import { PatternPreview, type PatternAppearance } from './pattern-preview'

type Kind = 'hero' | 'pattern'
type Metadata = PatternAppearance & { name: string; name_en: string; enabled: boolean; sort_order: number }
export type ManagedAsset = Metadata & { id: string; variant?: 'desktop' | 'mobile'; object_key: string; primary_url: string; mirror_url: string }
export type Publication = { object_key: string; primary: string; mirror: string; warnings?: string[]; item?: ManagedAsset }
const defaults: Metadata = { name: '', name_en: '', enabled: true, sort_order: 0, light_color: '#9c846a', dark_color: '#ac9680', light_opacity: 0.1, dark_opacity: 0.1, default_size_px: 96 }
const endpointFor = (kind: Kind) => `/api/v1/nav/${kind === 'hero' ? 'hero-assets' : 'background-patterns'}`

export function HeroAssetsPage() { return <AssetPage kind="hero" /> }
export function BackgroundPatternsPage() { return <AssetPage kind="pattern" /> }

function AssetPage({ kind }: { kind: Kind }) {
  const canWrite = useAuth().can('content.write')
  const [variant, setVariant] = useState<'desktop' | 'mobile'>('desktop')
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<ManagedAsset | null | undefined>()
  const endpoint = endpointFor(kind)
  const query = useQuery({ queryKey: ['managed-assets', kind, variant, page], queryFn: () => listJSON<ManagedAsset>(`${endpoint}${kind === 'hero' ? `?variant=${variant}` : ''}`, page, 20) })
  return <PageLayout>
    <PageHeader title={kind === 'hero' ? '首页 Hero' : '背景图案'} actions={canWrite && selected === undefined && <Button onClick={() => setSelected(null)}>新增{kind === 'hero' ? ' Hero' : '图案'}</Button>} />
    {kind === 'hero' && <div className="flex gap-2">{(['desktop', 'mobile'] as const).map((value) => <Button key={value} variant={variant === value ? 'primary' : 'secondary'} disabled={selected !== undefined} onClick={() => { setVariant(value); setPage(1) }}>{value === 'desktop' ? '桌面资源池' : '移动资源池'}</Button>)}</div>}
    <p className="text-sm text-muted-foreground">{kind === 'hero' ? '首页从当前设备的已启用资源中随机选择。桌面与移动资源互相独立。仅接受 AVIF。' : 'SVG 图案可平铺展示，颜色、透明度和尺寸作为用户选择此图案时的默认外观。'}</p>
    {selected !== undefined && <AssetEditor key={`${kind}-${variant}-${selected?.id ?? 'new'}`} kind={kind} variant={variant} asset={selected} close={() => setSelected(undefined)} saved={() => { setSelected(undefined); void query.refetch() }} />}
    {query.isLoading ? <LoadingState /> : query.error ? <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} /> : <>
      <div className="grid gap-4 lg:grid-cols-2">{query.data?.list.map((asset) => <Section key={asset.id} title={asset.name} actions={<Button variant="secondary" disabled={selected !== undefined} onClick={() => setSelected(asset)}>{canWrite ? '管理' : '查看'}</Button>}>
        {kind === 'pattern' ? <PatternPreview url={asset.primary_url} appearance={asset} /> : <img className="h-40 w-full rounded object-cover" src={asset.primary_url} alt={asset.name} onError={(event) => { if (asset.mirror_url && event.currentTarget.src !== asset.mirror_url) event.currentTarget.src = asset.mirror_url }} />}
        <p className="mt-3 text-sm">{asset.enabled ? '已启用' : '已停用'}{kind === 'pattern' && ` · ${asset.name_en}`}</p>
      </Section>)}</div>
      {query.data?.total === 0 && <Section>暂无资源，可新增第一项。</Section>}
      <div className="flex items-center justify-end gap-3"><span className="text-sm">共 {query.data?.total ?? 0} 项 · 第 {page} 页</span><Button variant="secondary" disabled={page === 1 || selected !== undefined} onClick={() => setPage(page - 1)}>上一页</Button><Button variant="secondary" disabled={page * 20 >= (query.data?.total ?? 0) || selected !== undefined} onClick={() => setPage(page + 1)}>下一页</Button></div>
    </>}
  </PageLayout>
}

export function AssetEditor({ kind, variant, asset, close, saved }: { kind: Kind; variant: 'desktop' | 'mobile'; asset: ManagedAsset | null; close: () => void; saved: () => void }) {
  const canWrite = useAuth().can('content.write')
  const { toast } = useToast()
  const [draft, setDraft] = useState<Metadata>(() => asset ? { name: asset.name, name_en: asset.name_en ?? '', enabled: asset.enabled, sort_order: asset.sort_order ?? 0, light_color: asset.light_color, dark_color: asset.dark_color, light_opacity: asset.light_opacity, dark_opacity: asset.dark_opacity, default_size_px: asset.default_size_px } : { ...defaults })
  const [initial] = useState(() => JSON.stringify(draft))
  const [file, setFile] = useState<File | null>(null)
  const [localURL, setLocalURL] = useState('')
  const [error, setError] = useState('')
  const [validating, setValidating] = useState(false)
  const dirty = JSON.stringify(draft) !== initial || file !== null
  useUnsavedChanges(dirty)
  useEffect(() => { if (!file) { setLocalURL(''); return }; const url = URL.createObjectURL(file); setLocalURL(url); return () => URL.revokeObjectURL(url) }, [file])
  const metadata = kind === 'hero' ? { name: draft.name, enabled: draft.enabled } : draft
  const mutation = useMutation({
    mutationFn: async (action: 'save' | 'file' | 'delete') => {
      const endpoint = endpointFor(kind)
      if (action === 'delete') return sendJSON<Publication>(`${endpoint}/${asset!.id}`, 'DELETE')
      if (asset && action === 'save') return sendJSON<Publication>(`${endpoint}/${asset.id}`, 'PUT', metadata)
      if (!file) throw new Error('请先选择文件')
      const body = new FormData(); body.set('file', file)
      if (!asset) { for (const [key, value] of Object.entries(metadata)) body.set(key, String(value)); if (kind === 'hero') body.set('variant', variant) }
      return sendForm<Publication>(`${endpoint}${asset ? `/${asset.id}/file` : ''}`, body)
    },
    onSuccess: (result) => { toast(result.warnings?.length ? `已发布到 COS；${result.warnings.join('；')}。可在云资源中修复 Mirror。` : '资源已保存'); saved() },
    onError: (err) => setError(errorMessage(err)),
  })
  const update = <K extends keyof Metadata>(key: K, value: Metadata[K]) => setDraft((current) => ({ ...current, [key]: value }))
  const selectFile = async (candidate?: File) => {
    setFile(null); setError(''); if (!candidate) return
    setValidating(true)
    try {
      if (candidate.size > 5 * 1024 * 1024) throw new Error('文件不能超过 5 MiB')
      if (kind === 'hero' && !candidate.name.toLowerCase().endsWith('.avif')) throw new Error('Hero 只接受 AVIF 文件')
      setFile(candidate)
    } catch (err) { setError(errorMessage(err)) } finally { setValidating(false) }
  }
  const previewURL = localURL || asset?.primary_url
  return <Section title={asset ? `管理 ${asset.name}` : '新增资源'}>
    <form className="grid gap-4" onSubmit={(event) => { event.preventDefault(); setError(''); mutation.mutate('save') }}>
      {error && <Alert tone="danger">{error}</Alert>}
      <fieldset disabled={!canWrite || mutation.isPending} className="grid gap-4">
        <div className="grid gap-4 sm:grid-cols-2"><FormField label="名称" required><Input required maxLength={120} value={draft.name} onChange={(event) => update('name', event.target.value)} /></FormField>{kind === 'pattern' && <FormField label="英文名称" required><Input required maxLength={120} value={draft.name_en} onChange={(event) => update('name_en', event.target.value)} /></FormField>}</div>
        <label className="flex items-center gap-2 text-sm"><input type="checkbox" checked={draft.enabled} onChange={(event) => update('enabled', event.target.checked)} />启用</label>
        {kind === 'pattern' && <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">{(['light', 'dark'] as const).map((theme) => <div key={theme} className="grid gap-3"><FormField label={`${theme === 'light' ? '浅色' : '深色'}图案颜色`}><Input type="color" value={draft[`${theme}_color`]} onChange={(event) => update(`${theme}_color`, event.target.value)} /></FormField><FormField label={`${theme === 'light' ? '浅色' : '深色'}透明度`}><Input required type="number" min={0} max={1} step={0.01} value={draft[`${theme}_opacity`]} onChange={(event) => update(`${theme}_opacity`, event.target.valueAsNumber)} /></FormField></div>)}<div className="grid gap-3"><FormField label="平铺尺寸（px）"><Input required type="number" min={1} step={1} value={draft.default_size_px} onChange={(event) => update('default_size_px', event.target.valueAsNumber)} /></FormField><FormField label="目录顺序" help="数字越小越靠前"><Input required type="number" step={1} value={draft.sort_order} onChange={(event) => update('sort_order', event.target.valueAsNumber)} /></FormField></div></div>}
        <FormField label={asset ? '选择替换文件' : '选择文件'} help={asset ? '预览满意后点击“替换文件”。外观和名称通过“保存信息”单独保存。' : '选择后可预览，点击“创建资源”才会上传。'}><Input type="file" accept={kind === 'hero' ? '.avif,image/avif' : '.svg,image/svg+xml'} onChange={(event) => void selectFile(event.target.files?.[0])} /></FormField>
      </fieldset>
      {previewURL && (kind === 'pattern' ? <PatternPreview url={previewURL} appearance={draft} /> : <img className="max-h-72 rounded object-contain" src={previewURL} alt="Hero 预览" />)}
      {asset && <details className="text-xs text-muted-foreground"><summary>资源地址</summary><p className="my-2 break-all">{asset.object_key}</p><a className="block break-all underline" href={asset.primary_url} target="_blank" rel="noreferrer">COS / 中国 CDN</a><a className="block break-all underline" href={asset.mirror_url} target="_blank" rel="noreferrer">R2 / Mirror CDN</a></details>}
      <div className="flex flex-wrap justify-end gap-2">
        {asset && canWrite && <Button type="button" variant="danger" disabled={mutation.isPending} onClick={() => { if (window.confirm('删除后将不再供用户选择，云端文件会保留。确定删除？')) mutation.mutate('delete') }}>删除</Button>}
        <Button type="button" variant="secondary" disabled={mutation.isPending} onClick={() => { if (!dirty || window.confirm('放弃未保存的改动？')) close() }}>关闭</Button>
        {canWrite && asset && <Button type="button" variant="secondary" disabled={!file || validating || mutation.isPending} onClick={() => mutation.mutate('file')}>替换文件</Button>}
        {canWrite && <Button disabled={validating || mutation.isPending || (!asset && !file)}>{asset ? '保存信息' : '创建资源'}</Button>}
      </div>
    </form>
  </Section>
}
