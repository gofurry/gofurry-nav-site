import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { KeyRound, Plus, Shield, UserCog } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { DataTable, type AdminColumn } from '../../components/admin/data-table'
import { FormField, PageHeader, PageLayout } from '../../components/admin/page'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Dialog } from '../../components/ui/dialog'
import { Input } from '../../components/ui/input'
import { Select } from '../../components/ui/select'
import { errorMessage, listJSON, sendJSON } from '../../lib/api'
import { formatDate } from '../../lib/utils'
import type { Account } from '../operations/types'

const accountSchema = z.object({ username: z.string().regex(/^[a-z0-9][a-z0-9._-]{2,63}$/, '用户名需为 3–64 位小写字母、数字或 ._-'), display_name: z.string().trim().min(1, '请输入显示名称').max(128), role: z.enum(['owner', 'developer', 'operator']), password: z.string().min(1, '请输入初始密码') })
type AccountValues = z.infer<typeof accountSchema>
type AccountAction = 'display' | 'role' | 'status' | 'password' | 'revoke'

export function AccountsPage() {
  const [page, setPage] = useState(1); const [search, setSearch] = useState(''); const [createOpen, setCreateOpen] = useState(false); const [target, setTarget] = useState<Account | null>(null); const [action, setAction] = useState<AccountAction | null>(null)
  const query = useQuery({ queryKey: ['accounts', page, search], queryFn: () => listJSON<Account>('/api/v1/auth/accounts/', page, 20, search) })
  const columns = useMemo<AdminColumn<Account>[]>(() => [
    { key: 'display_name', header: 'User', render: (row) => <div><p className="font-medium text-primary">{row.display_name}</p><TechnicalLabel>account #{row.id}</TechnicalLabel></div> },
    { key: 'username', header: 'Username', render: (row) => <span className="font-mono text-xs">@{row.username}</span> },
    { key: 'role', header: 'Role', render: (row) => <StatusBadge tone={row.role === 'owner' ? 'info' : 'neutral'}>{row.role}</StatusBadge> },
    { key: 'status', header: 'Status', render: (row) => <StatusBadge tone={row.status === 'active' ? 'success' : 'danger'}>{row.status}</StatusBadge> },
    { key: 'last_login_at', header: 'Last Login', render: (row) => formatDate(row.last_login_at) },
  ], [])
  const start = (account: Account, next: AccountAction) => { setTarget(account); setAction(next) }
  return <PageLayout><PageHeader title="账号与权限" description="管理固定 Owner / Developer / Operator 角色，不创建自定义权限。" eyebrow="account.manage" actions={<Button onClick={() => setCreateOpen(true)}><Plus className="size-4" />创建账号</Button>} /><DataTable data={query.data?.list ?? []} columns={columns} total={query.data?.total ?? 0} page={page} pageSize={20} search={search} onSearchChange={(value) => { setSearch(value); setPage(1) }} onPageChange={setPage} onPageSizeChange={() => undefined} searchPlaceholder="搜索用户名或显示名称…" onRowClick={(row) => { setTarget(row); setAction('display') }} loading={query.isLoading} error={query.error?.message} onRetry={() => void query.refetch()} /><CreateAccountDialog open={createOpen} onOpenChange={setCreateOpen} />{target && action && <AccountActionDialog key={`${target.id}-${action}`} account={target} action={action} onActionChange={(next) => { setAction(next); if (!next) setTarget(null) }} start={start} />}</PageLayout>
}

function CreateAccountDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const client = useQueryClient(); const { toast } = useToast(); const [operationError, setOperationError] = useState('')
  const form = useForm<AccountValues>({ resolver: zodResolver(accountSchema), defaultValues: { username: '', display_name: '', role: 'operator', password: '' } })
  const mutation = useMutation({ mutationFn: (values: AccountValues) => sendJSON<Account>('/api/v1/auth/accounts/', 'POST', values), onSuccess: async () => { await client.invalidateQueries({ queryKey: ['accounts'] }); toast('账号已创建'); form.reset(); setOperationError(''); onOpenChange(false) }, onError: (error) => setOperationError(errorMessage(error)) })
  return <Dialog open={open} onOpenChange={onOpenChange} title="创建账号" footer={<><Button variant="secondary" onClick={() => onOpenChange(false)}>取消</Button><Button disabled={mutation.isPending} onClick={form.handleSubmit((values) => { setOperationError(''); mutation.mutate(values) })}>创建</Button></>}><form className="grid gap-4" onSubmit={(event) => event.preventDefault()}>{operationError && <Alert tone="danger">{operationError}</Alert>}<FormField label="Username" required error={form.formState.errors.username?.message}><Input autoComplete="off" {...form.register('username')} /></FormField><FormField label="显示名称" required error={form.formState.errors.display_name?.message}><Input {...form.register('display_name')} /></FormField><FormField label="角色"><Select value={form.watch('role')} onValueChange={(value) => form.setValue('role', value as AccountValues['role'], { shouldDirty: true })} options={[{ value: 'operator', label: 'Operator' }, { value: 'developer', label: 'Developer' }, { value: 'owner', label: 'Owner' }]} /></FormField><FormField label="初始密码" required error={form.formState.errors.password?.message}><Input type="password" autoComplete="new-password" {...form.register('password')} /></FormField></form></Dialog>
}

function AccountActionDialog({ account, action, onActionChange, start }: { account: Account | null; action: AccountAction | null; onActionChange: (action: AccountAction | null) => void; start: (account: Account, action: AccountAction) => void }) {
  const initialValue = action === 'display' ? account?.display_name ?? '' : action === 'role' ? account?.role ?? '' : ''
  const client = useQueryClient(); const { toast } = useToast(); const [value, setValue] = useState(initialValue); const [operationError, setOperationError] = useState('')
  const close = () => { setValue(''); setOperationError(''); onActionChange(null) }
  const endpoint = () => { if (!account || !action) return { path: '', method: 'POST' as const, body: {} }; if (action === 'display') return { path: `/${account.id}/display-name`, method: 'PUT' as const, body: { display_name: value } }; if (action === 'role') return { path: `/${account.id}/role`, method: 'PUT' as const, body: { role: value } }; if (action === 'status') return { path: `/${account.id}/status`, method: 'PUT' as const, body: { status: account.status === 'active' ? 'disabled' : 'active' } }; if (action === 'password') return { path: `/${account.id}/password`, method: 'POST' as const, body: { password: value } }; return { path: `/${account.id}/revoke-sessions`, method: 'POST' as const, body: {} } }
  const mutation = useMutation({ mutationFn: () => { const target = endpoint(); return sendJSON<Account>(`/api/v1/auth/accounts${target.path}`, target.method, target.body) }, onSuccess: async () => { await client.invalidateQueries({ queryKey: ['accounts'] }); toast(action === 'revoke' ? '已撤销现有会话' : '账号已更新'); close() }, onError: (error) => setOperationError(errorMessage(error)) })
  if (!account || !action) return null
  const title = { display: '编辑显示名称', role: '变更角色', status: account.status === 'active' ? '停用账号' : '启用账号', password: '重置密码', revoke: '撤销会话' }[action]
  const valid = !['display', 'password'].includes(action) || value.trim().length > 0
  return <Dialog open title={title} description={`${account.display_name} · @${account.username}`} onOpenChange={(open) => { if (!open) close() }} footer={<><Button variant="secondary" onClick={close}>取消</Button><Button variant={['status', 'password', 'revoke'].includes(action) ? 'danger' : 'primary'} disabled={!valid || mutation.isPending} onClick={() => mutation.mutate()}>确认</Button></>}><div className="grid gap-4">{operationError && <Alert tone="danger">{operationError}</Alert>}{action === 'display' && <FormField label="显示名称"><Input defaultValue={account.display_name} onChange={(event) => setValue(event.target.value)} /></FormField>}{action === 'role' && <><Alert tone="warning">变更角色会立即撤销该账号的已有会话。最后一个活跃 Owner 由后端事务保护。</Alert><FormField label="新角色"><Select value={value} onValueChange={setValue} options={[{ value: 'operator', label: 'Operator' }, { value: 'developer', label: 'Developer' }, { value: 'owner', label: 'Owner' }]} /></FormField></>}{action === 'status' && <Alert tone="warning">{account.status === 'active' ? '停用后该用户将立即无法登录，已有会话同时失效。' : '启用后该用户可再次登录。'}</Alert>}{action === 'password' && <><Alert tone="warning">重置密码会撤销该账号的全部现有会话；密码不会进入审计详情。</Alert><FormField label="新密码" required><Input type="password" autoComplete="new-password" value={value} onChange={(event) => setValue(event.target.value)} /></FormField></>}{action === 'revoke' && <Alert tone="warning">所有已签发登录会话将立即失效，用户需要重新登录。</Alert>}<div className="grid grid-cols-2 gap-2 border-t pt-4"><Button variant="secondary" onClick={() => start(account, 'display')}><UserCog className="size-4" />显示名称</Button><Button variant="secondary" onClick={() => start(account, 'role')}><Shield className="size-4" />角色</Button><Button variant="secondary" onClick={() => start(account, 'status')}>{account.status === 'active' ? '停用' : '启用'}</Button><Button variant="secondary" onClick={() => start(account, 'password')}><KeyRound className="size-4" />密码</Button><Button className="col-span-2" variant="secondary" onClick={() => start(account, 'revoke')}>撤销会话</Button></div></div></Dialog>
}
