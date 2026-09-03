import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { LoaderCircle, Plus } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Controller, useForm, type Resolver } from 'react-hook-form'
import { useParams, useSearchParams } from 'react-router-dom'
import { useToast } from '../../app/toast'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { RemoteSelect as SharedRemoteSelect } from '../../components/admin/operations'
import { FormField, FormSection, PageHeader, PageLayout } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { DateTimePicker } from '../../components/ui/date-picker'
import { ConfirmAction } from '../../components/ui/dialog'
import { Input, Textarea } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { Sheet } from '../../components/ui/sheet'
import { useAuth } from '../auth/auth-context'
import { errorMessage, getJSON, listJSON, sendJSON } from '../../lib/api'
import { formatDate, getAtPath } from '../../lib/utils'
import { findResource } from './definitions'
import type { ResourceDefinition, ResourceField, ResourceRecord } from './resource-types'

export function RemoteSelect({ endpoint, value, onChange, disabled }: { endpoint: string; value: unknown; onChange: (value: string) => void; disabled?: boolean }) {
  const selected = value === null || value === undefined ? '' : String(value)
  return <SharedRemoteSelect endpoint={endpoint} value={selected} onChange={(option) => onChange(option?.id ?? '')} disabled={disabled} placeholder="搜索远程选项…" />
}

function FieldControl({ field, value, onChange, disabled }: { field: ResourceField; value: unknown; onChange: (value: unknown) => void; disabled?: boolean }) {
  if (field.type === 'textarea') return <Textarea value={String(value ?? '')} onChange={(event) => onChange(event.target.value)} disabled={disabled} placeholder={field.placeholder} />
  if (field.type === 'select') return <Select value={String(value ?? '')} onValueChange={onChange} options={field.options ?? []} disabled={disabled} />
  if (field.type === 'remote-select' && field.optionEndpoint) return <RemoteSelect endpoint={field.optionEndpoint} value={value} onChange={onChange} disabled={disabled} />
  if (field.type === 'boolean') return <label className="flex h-9 items-center gap-2"><input type="checkbox" checked={Boolean(value)} onChange={(event) => onChange(event.target.checked)} disabled={disabled} className="size-4 accent-primary" /><span className="text-sm text-muted-foreground">{value ? '已启用' : '未启用'}</span></label>
  if (field.type === 'string-array') return <Textarea value={Array.isArray(value) ? value.join('\n') : ''} onChange={(event) => onChange(event.target.value.split('\n').map((item) => item.trim()).filter(Boolean))} disabled={disabled} placeholder={field.placeholder} />
  if (field.type === 'datetime') return <DateTimePicker value={String(value ?? '')} onValueChange={onChange} disabled={disabled} ariaLabel={field.label} />
  return <Input type={field.type === 'number' ? 'number' : 'text'} value={String(value ?? '')} onChange={(event) => onChange(event.target.value)} disabled={disabled} placeholder={field.placeholder} />
}

function groupFields(fields: ResourceField[]) {
  const result: Array<{ title: string; fields: ResourceField[] }> = []
  let current = { title: '基本信息', fields: [] as ResourceField[] }
  fields.forEach((field) => {
    if (field.section) { current = { title: field.section, fields: [] }; result.push(current) }
    if (result.length === 0) result.push(current)
    current.fields.push(field)
  })
  return result
}

export function resourceRecordID(record: ResourceRecord) {
  if (record.id === undefined || record.id === null || String(record.id).trim() === '') throw new Error('resource record id is required')
  return String(record.id)
}

function resourceEndpoint(endpoint: string, id: string) {
  return `${endpoint}/${encodeURIComponent(id)}`
}

function ResourceEditor({ definition, id, open, onOpenChange }: { definition: ResourceDefinition; id: string | null; open: boolean; onOpenChange: (open: boolean) => void }) {
  const client = useQueryClient()
  const { toast } = useToast()
  const [operationError, setOperationError] = useState('')
  const detail = useQuery({ queryKey: ['resource-detail', definition.key, id], queryFn: () => getJSON<ResourceRecord>(resourceEndpoint(definition.detailEndpoint, id!)), enabled: open && id !== null })
  const form = useForm<ResourceRecord>({ resolver: zodResolver(definition.schema as never) as Resolver<ResourceRecord>, defaultValues: definition.defaults })
  useEffect(() => {
    if (!open) return
    if (id === null) form.reset(structuredClone(definition.defaults))
    else if (detail.data) form.reset(detail.data)
  }, [definition.defaults, detail.data, form, id, open])
  useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => { if (form.formState.isDirty) event.preventDefault() }
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [form.formState.isDirty])
  const mutation = useMutation({ mutationFn: (payload: ResourceRecord) => id === null ? sendJSON(definition.detailEndpoint, 'POST', payload) : sendJSON(resourceEndpoint(definition.detailEndpoint, id), 'PUT', payload), onSuccess: async () => {
    await client.invalidateQueries({ queryKey: ['resource-list', definition.key] })
    if (id !== null) await client.invalidateQueries({ queryKey: ['resource-detail', definition.key, id] })
    form.reset(form.getValues())
    toast(`${definition.title}${id === null ? '已创建' : '已保存'}`)
    onOpenChange(false)
  }, onError: (error) => setOperationError(errorMessage(error)) })
  const requestClose = (next: boolean) => {
    if (!next && form.formState.isDirty && !window.confirm('存在未保存的修改，确定离开吗？')) return
    onOpenChange(next)
  }
  return <Sheet open={open} onOpenChange={requestClose} title={id === null ? `新增${definition.title}` : `编辑${definition.title}`} description={id === null ? '填写业务字段后提交' : `记录 #${id}`}>
    {detail.isLoading && id !== null ? <div className="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle className="size-4 animate-spin" />正在加载详情…</div> : <form className="grid gap-7" onSubmit={form.handleSubmit((values) => { setOperationError(''); mutation.mutate(values) })}>
      {operationError && <Alert tone="danger">{operationError}</Alert>}
      {groupFields(definition.fields).map((group) => <FormSection key={group.title} title={group.title}>{group.fields.map((field) => <Controller key={field.key} name={field.key} control={form.control} render={({ field: controlField, fieldState }) => <FormField label={field.label} help={field.help} error={fieldState.error?.message}><FieldControl field={field} value={controlField.value} onChange={controlField.onChange} disabled={mutation.isPending || (id !== null && field.key === 'id')} /></FormField>} />)}</FormSection>)}
      <div className="sticky bottom-0 -mx-5 -mb-5 flex justify-end gap-2 border-t bg-surface px-5 py-4"><Button type="button" variant="secondary" onClick={() => requestClose(false)}>取消</Button><Button type="submit" disabled={mutation.isPending || !form.formState.isDirty}>{mutation.isPending && <LoaderCircle className="size-4 animate-spin" />}{id === null ? '创建' : '保存修改'}</Button></div>
    </form>}
  </Sheet>
}

type ResourceSection = 'nav' | 'game'

export function ResourcePage({ section, resource }: { section: ResourceSection; resource?: string }) {
  const definition = findResource(section, resource)
  const auth = useAuth()
  const client = useQueryClient()
  const { toast } = useToast()
  const [params, setParams] = useSearchParams()
  const [editor, setEditor] = useState<{ open: boolean; id: string | null }>({ open: false, id: null })
  const [deleting, setDeleting] = useState<ResourceRecord | null>(null)
  const page = Math.max(1, Number(params.get('page') || 1))
  const pageSize = [20, 50, 100].includes(Number(params.get('page_size'))) ? Number(params.get('page_size')) : 50
  const search = params.get('search') ?? ''
  const query = useQuery({ queryKey: ['resource-list', definition?.key, page, pageSize, search], queryFn: () => listJSON<ResourceRecord>(definition!.listEndpoint, page, pageSize, search), enabled: Boolean(definition) })
  const deleteMutation = useMutation({ mutationFn: () => sendJSON(resourceEndpoint(definition!.detailEndpoint, resourceRecordID(deleting!)), 'DELETE'), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['resource-list', definition?.key] }); toast(`${definition?.title}已删除`); setDeleting(null) }, onError: (error) => toast(errorMessage(error), 'danger') })
  const columns = useMemo<AdminColumn<ResourceRecord>[]>(() => (definition?.columns ?? []).map((column) => ({ key: column.key, header: column.label, hidden: column.hidden, render: (row) => column.format ? column.format(getAtPath(row, column.key), row) : column.key.includes('time') ? formatDate(getAtPath(row, column.key)) : String(getAtPath(row, column.key) ?? '—') })), [definition])
  if (!definition) return <Alert tone="danger">未找到资源定义。</Alert>
  const changeParam = (key: string, value: string) => { const next = new URLSearchParams(params); if (value) next.set(key, value); else next.delete(key); setParams(next, { replace: true }) }
  const canWrite = auth.can('content.write')
  return <PageLayout>
    <PageHeader title={definition.title} description={definition.description} eyebrow={`${definition.section}.${definition.key}`} actions={canWrite && <Button onClick={() => setEditor({ open: true, id: null })}><Plus className="size-4" />新增{definition.title}</Button>} />
    <DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={pageSize} search={search} onSearchChange={(value) => { const next = new URLSearchParams(params); next.set('page', '1'); if (value) next.set('search', value); else next.delete('search'); setParams(next, { replace: true }) }} onPageChange={(value) => changeParam('page', String(value))} onPageSizeChange={(value) => { const next = new URLSearchParams(params); next.set('page', '1'); next.set('page_size', String(value)); setParams(next, { replace: true }) }} onEdit={canWrite ? (row) => setEditor({ open: true, id: resourceRecordID(row) }) : undefined} onDelete={canWrite ? setDeleting : undefined} onRowClick={canWrite ? (row) => setEditor({ open: true, id: resourceRecordID(row) }) : undefined} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} />
    <ResourceEditor definition={definition} id={editor.id} open={editor.open} onOpenChange={(open) => setEditor((current) => ({ ...current, open }))} />
    <ConfirmAction open={Boolean(deleting)} onOpenChange={(open) => { if (!open) setDeleting(null) }} title={`删除${definition.title}`} description={`确定删除记录 #${deleting?.id ?? ''} 吗？`} busy={deleteMutation.isPending} onConfirm={() => deleteMutation.mutate()} />
  </PageLayout>
}

export function ResourceEngineBoundary({ section }: { section: ResourceSection }) {
  const { resource } = useParams()
  return <ResourcePage section={section} resource={resource} />
}
