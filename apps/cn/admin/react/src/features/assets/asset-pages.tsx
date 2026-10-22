import { CheckCircle, MinusCircle, GearSix, Eye } from '@phosphor-icons/react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useToast } from '../../app/toast'
import { FormField, PageHeader, PageLayout, Section } from '../../components/admin/page'
import { ErrorState, LoadingState } from '../../components/admin/states'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { ConfirmAction } from '../../components/ui/dialog'
import { FilePicker } from '../../components/admin/file-picker'
import { useUnsavedChanges } from '../../hooks/use-unsaved-changes'
import { errorMessage, getJSON, listJSON, sendForm, sendJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'

type Metadata = { name: string; enabled: boolean }
export type ManagedAsset = Metadata & { id: string; variant?: 'desktop' | 'mobile'; object_key: string; primary_url: string; mirror_url: string }
export type Publication = { object_key: string; primary: string; mirror: string; warnings?: string[]; item?: ManagedAsset }
const defaults: Metadata = { name: '', enabled: true }
const endpoint = '/api/v1/nav/hero-assets'

export function HeroAssetsPage() {
  const canWrite = useAuth().can('content.write')
  const { toast } = useToast()
  const [variant, setVariant] = useState<'desktop' | 'mobile'>('desktop')
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<ManagedAsset | null | undefined>()
  const query = useQuery({ queryKey: ['managed-assets', 'hero', variant, page], queryFn: () => listJSON<ManagedAsset>(`${endpoint}?variant=${variant}`, page, 20) })
  const toggle = useMutation({
    mutationFn: async (asset: ManagedAsset) => {
      const current = await getJSON<ManagedAsset>(`${endpoint}/${asset.id}`)
      const metadata = { name: current.name, enabled: !asset.enabled }
      return sendJSON(`${endpoint}/${asset.id}`, 'PUT', metadata)
    },
    onSuccess: () => { void query.refetch() },
    onError: (err) => toast(errorMessage(err), 'danger'),
  })
  return <PageLayout>
    <PageHeader title="首页 Hero" actions={canWrite && selected === undefined && <Button onClick={() => setSelected(null)}>新增 Hero</Button>} />
    <div className="flex gap-2">{(['desktop', 'mobile'] as const).map((value) => <Button key={value} variant={variant === value ? 'primary' : 'secondary'} disabled={selected !== undefined} onClick={() => { setVariant(value); setPage(1) }}>{value === 'desktop' ? '桌面资源池' : '移动资源池'}</Button>)}</div>
    <p className="text-sm text-muted-foreground">首页从当前设备的已启用资源中随机选择。桌面与移动资源互相独立。仅接受 AVIF。</p>
    {selected !== undefined && <AssetEditor key={`${variant}-${selected?.id ?? 'new'}`} variant={variant} asset={selected} close={() => setSelected(undefined)} saved={() => { setSelected(undefined); void query.refetch() }} />}
    {query.isLoading ? <LoadingState /> : query.error ? <ErrorState message={errorMessage(query.error)} onRetry={() => void query.refetch()} /> : <>
      <div data-asset-grid={variant} className={`grid gap-4 ${variant === 'mobile' ? 'grid-cols-1 md:grid-cols-2 xl:grid-cols-4' : 'grid-cols-1 lg:grid-cols-2'}`}>{query.data?.list.map((asset) => <AssetPreviewCard key={asset.id} asset={asset} variant={variant} canWrite={canWrite} disabled={selected !== undefined || toggle.isPending} manage={() => setSelected(asset)} toggle={() => toggle.mutate(asset)} />)}</div>
      {query.data?.total === 0 && <Section>暂无资源，可新增第一项。</Section>}
      <div className="flex items-center justify-end gap-3"><span className="text-sm">共 {query.data?.total ?? 0} 项 · 第 {page} 页</span><Button variant="secondary" disabled={page === 1 || selected !== undefined} onClick={() => setPage(page - 1)}>上一页</Button><Button variant="secondary" disabled={page * 20 >= (query.data?.total ?? 0) || selected !== undefined} onClick={() => setPage(page + 1)}>下一页</Button></div>
    </>}
  </PageLayout>
}

export function AssetPreviewCard({ asset, variant, canWrite, disabled, manage, toggle }: { asset: ManagedAsset; variant: 'desktop' | 'mobile'; canWrite: boolean; disabled: boolean; manage: () => void; toggle: () => void }) {
  return <article data-asset-card className={`relative isolate overflow-hidden rounded-lg border bg-surface ${variant === 'mobile' ? 'aspect-square' : 'aspect-video'}`}>
    <img className="absolute inset-0 h-full w-full object-cover" src={asset.primary_url} alt={asset.name} onError={(event) => { if (asset.mirror_url && event.currentTarget.src !== asset.mirror_url) event.currentTarget.src = asset.mirror_url }} />
    <div className="absolute inset-x-0 top-0 flex items-start justify-between gap-2 bg-gradient-to-b from-black/75 to-transparent p-3 pb-8 text-white">
      <div className="min-w-0"><h2 className="truncate text-sm font-semibold" title={asset.name}>{asset.name}</h2></div>
      <div className="flex shrink-0 gap-1">
        <Button type="button" variant="ghost" size="icon" className="size-8 bg-black/40 text-white hover:bg-black/65" aria-label={`${asset.enabled ? '停用' : '启用'} ${asset.name}`} aria-pressed={asset.enabled} title={asset.enabled ? '已启用' : '已停用'} disabled={!canWrite || disabled} onClick={toggle}>{asset.enabled ? <CheckCircle className="size-5" weight="fill" /> : <MinusCircle className="size-5" />}</Button>
        <Button type="button" variant="ghost" size="icon" className="size-8 bg-black/40 text-white hover:bg-black/65" aria-label={`${canWrite ? '管理' : '查看'} ${asset.name}`} title={canWrite ? '管理' : '查看'} disabled={disabled} onClick={manage}>{canWrite ? <GearSix className="size-5" /> : <Eye className="size-5" />}</Button>
      </div>
    </div>
  </article>
}

export function AssetEditor({ variant, asset, close, saved }: { variant: 'desktop' | 'mobile'; asset: ManagedAsset | null; close: () => void; saved: () => void }) {
  const canWrite = useAuth().can('content.write')
  const { toast } = useToast()
  const [draft, setDraft] = useState<Metadata>(() => asset ? { name: asset.name, enabled: asset.enabled } : { ...defaults })
  const [initial] = useState(() => JSON.stringify(draft))
  const [file, setFile] = useState<File | null>(null)
  const [localURL, setLocalURL] = useState('')
  const [error, setError] = useState('')
  const [confirmation, setConfirmation] = useState<'delete' | 'file' | null>(null)
  const dirty = JSON.stringify(draft) !== initial || file !== null
  useUnsavedChanges(dirty)
  useEffect(() => { if (!file) { setLocalURL(''); return }; const url = URL.createObjectURL(file); setLocalURL(url); return () => URL.revokeObjectURL(url) }, [file])
  const metadata = draft
  const mutation = useMutation({
    mutationFn: async (action: 'save' | 'file' | 'delete') => {
      if (action === 'delete') return sendJSON<Publication>(`${endpoint}/${asset!.id}`, 'DELETE')
      if (asset && action === 'save') return sendJSON<Publication>(`${endpoint}/${asset.id}`, 'PUT', metadata)
      if (!file) throw new Error('请先选择文件')
      const body = new FormData(); body.set('file', file)
      if (!asset) { for (const [key, value] of Object.entries(metadata)) body.set(key, String(value)); body.set('variant', variant) }
      return sendForm<Publication>(`${endpoint}${asset ? `/${asset.id}/file` : ''}`, body)
    },
    onSuccess: (result) => { setConfirmation(null); toast(result.warnings?.length ? `已发布到 COS；${result.warnings.join('；')}。可在云资源中修复 Mirror。` : '资源已保存'); saved() },
    onError: (err) => { setConfirmation(null); setError(errorMessage(err)) },
  })
  const update = <K extends keyof Metadata>(key: K, value: Metadata[K]) => setDraft((current) => ({ ...current, [key]: value }))
  const selectFile = (candidate: File | null) => {
    setFile(null); setError(''); if (!candidate) return
    try {
      if (candidate.size > 5 * 1024 * 1024) throw new Error('文件不能超过 5 MiB')
      if (!candidate.name.toLowerCase().endsWith('.avif')) throw new Error('Hero 只接受 AVIF 文件')
      setFile(candidate)
    } catch (err) { setError(errorMessage(err)) }
  }
  const previewURL = localURL || asset?.primary_url
  return <><Section title={asset ? `管理 ${asset.name}` : '新增资源'}>
    <form className="grid gap-3" onSubmit={(event) => { event.preventDefault(); setError(''); mutation.mutate('save') }}>
      {error && <Alert tone="danger">{error}</Alert>}
      <div className="grid items-start gap-4 lg:grid-cols-2">
        <fieldset disabled={!canWrite || mutation.isPending} className="grid min-w-0 gap-3">
          <div className="grid grid-cols-[minmax(0,1fr)_auto] items-end gap-3">
            <FormField label="名称" required><Input required maxLength={120} value={draft.name} onChange={(event) => update('name', event.target.value)} /></FormField>
            <label className="flex h-9 items-center gap-2 text-sm"><input type="checkbox" checked={draft.enabled} onChange={(event) => update('enabled', event.target.checked)} />启用</label>
          </div>
          <FilePicker label={asset ? '选择替换文件' : '选择文件'} file={file} help={asset ? '选择后预览，确认后替换。' : '选择后预览，创建时上传。'} accept=".avif,image/avif" disabled={!canWrite || mutation.isPending} onSelect={selectFile} />
        </fieldset>
        <div className={`w-full min-w-0 ${variant === 'mobile' ? 'mx-auto max-w-64' : ''}`}>
          {previewURL ? (<img className={`w-full rounded-md border object-cover ${variant === 'mobile' ? 'aspect-square' : 'aspect-video'}`} src={previewURL} alt="Hero 预览" onError={(event) => { if (!localURL && asset?.mirror_url && event.currentTarget.src !== asset.mirror_url) event.currentTarget.src = asset.mirror_url }} />) : <div className={`grid place-items-center rounded-md border border-dashed text-sm text-muted-foreground ${variant === 'mobile' ? 'aspect-square' : 'aspect-video'}`}>选择文件后预览</div>}
        </div>
      </div>
      {asset && <details className="text-xs text-muted-foreground"><summary>资源地址</summary><p className="my-2 break-all">{asset.object_key}</p><a className="block break-all underline" href={asset.primary_url} target="_blank" rel="noreferrer">COS / 中国 CDN</a><a className="block break-all underline" href={asset.mirror_url} target="_blank" rel="noreferrer">R2 / Mirror CDN</a></details>}
      <div className="flex flex-wrap justify-end gap-2 border-t pt-3">
        {asset && canWrite && <Button type="button" variant="danger" disabled={mutation.isPending} onClick={() => setConfirmation('delete')}>删除</Button>}
        <Button type="button" variant="secondary" disabled={mutation.isPending} onClick={() => { if (!dirty || window.confirm('放弃未保存的改动？')) close() }}>关闭</Button>
        {canWrite && asset && <Button type="button" variant="secondary" disabled={!file || mutation.isPending} onClick={() => setConfirmation('file')}>替换文件</Button>}
        {canWrite && <Button disabled={mutation.isPending || (!asset && !file)}>{asset ? '保存信息' : '创建资源'}</Button>}
      </div>
    </form>
  </Section><ConfirmAction open={confirmation !== null} onOpenChange={(open) => { if (!open && !mutation.isPending) setConfirmation(null) }} title={confirmation === 'delete' ? '删除资源' : '替换资源文件'} description={confirmation === 'delete' ? `确定删除“${asset?.name}”？删除后不再展示此资源。` : `确定用“${file?.name}”替换“${asset?.name}”的文件？名称和启用状态不受影响。`} confirmLabel={confirmation === 'delete' ? '确认删除' : '确认替换'} variant={confirmation === 'delete' ? 'danger' : 'primary'} busy={mutation.isPending} onConfirm={() => { if (confirmation) mutation.mutate(confirmation) }} /></>
}
