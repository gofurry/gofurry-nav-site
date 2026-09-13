import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useToast } from '../../app/toast'
import { FormField, Section } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { errorMessage, sendForm, sendJSON } from '../../lib/api'
import type { Site } from '../../lib/types'
import type { Publication } from './asset-pages'

export function SiteIconEditor({ site, onDirtyChange }: { site: Site; onDirtyChange: (dirty: boolean) => void }) {
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState('')
  const [error, setError] = useState('')
  const client = useQueryClient()
  const { toast } = useToast()
  useEffect(() => { onDirtyChange(file !== null); return () => onDirtyChange(false) }, [file, onDirtyChange])
  useEffect(() => { if (!file) { setPreview(''); return }; const url = URL.createObjectURL(file); setPreview(url); return () => URL.revokeObjectURL(url) }, [file])
  const mutation = useMutation({ mutationFn: async (clear: boolean) => {
    const endpoint = `/api/v1/nav/sites/${site.id}/icon`
    if (clear) return sendJSON<Publication>(endpoint, 'DELETE')
    if (!file) throw new Error('请选择图标文件')
    const body = new FormData(); body.set('file', file)
    return sendForm<Publication>(endpoint, body)
  }, onSuccess: async (result) => {
    setFile(null); setError('')
    await client.invalidateQueries({ queryKey: ['site'] })
    toast(result.warnings?.length ? `图标已发布到 COS；${result.warnings.join('；')}` : '图标已更新，公开页面将在缓存刷新后显示')
  }, onError: (err) => setError(errorMessage(err)) })
  return <Section title="网站图标" description="上传网站原生图标，保留原始格式。清除后使用默认图标。">
    <div className="grid gap-4">
      {error && <Alert tone="danger">{error}</Alert>}
      {(preview || site.icon_primary_url) && <img className="size-20 rounded object-contain" src={preview || site.icon_primary_url} alt="网站图标预览" onError={(event) => { if (!preview && site.icon_mirror_url && event.currentTarget.src !== site.icon_mirror_url) event.currentTarget.src = site.icon_mirror_url }} />}
      <FormField label="选择图标" help="最多 2 MiB；不转码，不裁剪。"><Input type="file" disabled={mutation.isPending} onChange={(event) => { const next = event.target.files?.[0]; if (next && next.size > 2 * 1024 * 1024) { setFile(null); setError('图标不能超过 2 MiB'); return }; setError(''); setFile(next ?? null) }} /></FormField>
      <div className="flex justify-end gap-2"><Button variant="secondary" disabled={!site.icon || mutation.isPending} onClick={() => { if (window.confirm('清除网站图标并使用默认图标？')) mutation.mutate(true) }}>清除图标</Button><Button disabled={!file || mutation.isPending} onClick={() => mutation.mutate(false)}>上传图标</Button></div>
    </div>
  </Section>
}
