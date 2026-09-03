import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { useToast } from '../../app/toast'
import { FormField } from '../../components/admin/page'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Dialog } from '../../components/ui/dialog'
import { Input } from '../../components/ui/input'
import { errorMessage, sendJSON } from '../../lib/api'
import type { AuthState } from '../../lib/types'
import { useAuth } from './auth-context'

export const selfUsernameSchema = z.object({
  username: z.string().trim().regex(/^[a-z0-9][a-z0-9._-]{2,63}$/, '用户名需为 3–64 位小写字母、数字或 ._-'),
  current_password: z.string().min(1, '请输入当前密码'),
})

export const selfPasswordSchema = z.object({
  current_password: z.string().min(1, '请输入当前密码'),
  new_password: z.string().min(1, '请输入新密码'),
  confirm_password: z.string().min(1, '请再次输入新密码'),
}).refine((value) => value.new_password === value.confirm_password, { path: ['confirm_password'], message: '两次输入的新密码不一致' })

type UsernameValues = z.infer<typeof selfUsernameSchema>
type PasswordValues = z.infer<typeof selfPasswordSchema>

export async function changeOwnUsername(values: UsernameValues) {
  return sendJSON<AuthState>('/api/v1/auth/self/username', 'PUT', values)
}

export async function changeOwnPassword(values: PasswordValues) {
  return sendJSON('/api/v1/auth/self/password', 'POST', { current_password: values.current_password, new_password: values.new_password })
}

export function SelfUsernameDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const auth = useAuth()
  const { toast } = useToast()
  const form = useForm<UsernameValues>({ resolver: zodResolver(selfUsernameSchema), defaultValues: { username: auth.state?.identity?.username ?? '', current_password: '' } })
  const close = () => { form.reset({ username: auth.state?.identity?.username ?? '', current_password: '' }); onOpenChange(false) }
  const mutation = useMutation({ mutationFn: changeOwnUsername, onSuccess: (state) => { auth.setState(state); toast('用户名已更新'); form.reset({ username: state.identity?.username ?? '', current_password: '' }); onOpenChange(false) } })
  return <Dialog open={open} onOpenChange={(next) => { if (!next) close() }} title="修改用户名" footer={<><Button variant="secondary" onClick={close}>取消</Button><Button disabled={mutation.isPending} onClick={form.handleSubmit((values) => mutation.mutate(values))}>{mutation.isPending ? '保存中…' : '保存'}</Button></>}>
    <form className="grid gap-4" onSubmit={(event) => event.preventDefault()}>{mutation.error && <Alert tone="danger">{errorMessage(mutation.error)}</Alert>}<FormField label="新用户名" required error={form.formState.errors.username?.message}><Input autoComplete="username" {...form.register('username')} /></FormField><FormField label="当前密码" required error={form.formState.errors.current_password?.message}><Input type="password" autoComplete="current-password" {...form.register('current_password')} /></FormField></form>
  </Dialog>
}

export function SelfPasswordDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const auth = useAuth()
  const navigate = useNavigate()
  const form = useForm<PasswordValues>({ resolver: zodResolver(selfPasswordSchema), defaultValues: { current_password: '', new_password: '', confirm_password: '' } })
  const close = () => { form.reset(); onOpenChange(false) }
  const mutation = useMutation({ mutationFn: changeOwnPassword, onSuccess: async () => { await auth.clearSession(); navigate('/login', { replace: true }) } })
  return <Dialog open={open} onOpenChange={(next) => { if (!next) close() }} title="修改密码" footer={<><Button variant="secondary" onClick={close}>取消</Button><Button disabled={mutation.isPending} onClick={form.handleSubmit((values) => mutation.mutate(values))}>{mutation.isPending ? '修改中…' : '修改密码'}</Button></>}>
    <form className="grid gap-4" onSubmit={(event) => event.preventDefault()}>{mutation.error && <Alert tone="danger">{errorMessage(mutation.error)}</Alert>}<FormField label="当前密码" required error={form.formState.errors.current_password?.message}><Input type="password" autoComplete="current-password" {...form.register('current_password')} /></FormField><FormField label="新密码" required error={form.formState.errors.new_password?.message}><Input type="password" autoComplete="new-password" {...form.register('new_password')} /></FormField><FormField label="确认新密码" required error={form.formState.errors.confirm_password?.message}><Input type="password" autoComplete="new-password" {...form.register('confirm_password')} /></FormField></form>
  </Dialog>
}
