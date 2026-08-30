import { zodResolver } from '@hookform/resolvers/zod'
import { LoaderCircle, ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { Alert } from '../../components/ui/alert'
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { errorMessage } from '../../lib/api'
import { useAuth } from './auth-context'

const loginSchema = z.object({ username: z.string().trim().min(1, '请输入用户名'), password: z.string().min(1, '请输入密码') })
const bootstrapSchema = z.object({
  username: z.string().trim().regex(/^[a-z0-9._-]{3,64}$/, '请输入 3–64 位小写字母、数字、点、下划线或连字符'),
  display_name: z.string().trim().min(1, '请输入显示名称').max(128, '显示名称不能超过 128 字符'),
  password: z.string().min(1, '请输入密码'),
})

type LoginValues = z.infer<typeof loginSchema>
type BootstrapValues = z.infer<typeof bootstrapSchema>

function AuthFrame({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <main className="grid min-h-screen place-items-center bg-background p-6">
    <section className="w-full max-w-md rounded-lg border bg-surface p-7 shadow-sm">
      <div className="mb-6 flex items-center gap-3"><div className="grid size-10 place-items-center rounded-md bg-primary text-primary-foreground"><ShieldCheck className="size-5" /></div><div><h1 className="text-xl font-semibold">{title}</h1><p className="mt-0.5 text-sm text-muted-foreground">{description}</p></div></div>
      {children}
      <p className="mt-6 text-center font-mono text-xs text-muted-foreground">GoFurry Admin · React Foundation</p>
    </section>
  </main>
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="grid gap-1.5 text-sm"><span className="font-medium">{label}</span>{children}{error && <span className="text-xs text-danger">{error}</span>}</label>
}

export function LoginPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [operationError, setOperationError] = useState('')
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginValues>({ resolver: zodResolver(loginSchema), defaultValues: { username: '', password: '' } })
  if (!auth.loading && auth.state && !auth.state.initialized) return <Navigate to="/setup" replace />
  if (auth.state?.authenticated) return <Navigate to="/" replace />
  const submit = handleSubmit(async (values) => {
    setOperationError('')
    try {
      await auth.login(values)
      const destination = (location.state as { from?: string } | null)?.from ?? '/'
      navigate(destination, { replace: true })
    } catch (error) { setOperationError(errorMessage(error)) }
  })
  return <AuthFrame title="登录管理后台" description="使用 GoFurry Admin 账号继续">
    <form className="grid gap-4" onSubmit={submit}>
      {operationError && <Alert tone="danger">{operationError}</Alert>}
      <Field label="用户名" error={errors.username?.message}><Input autoFocus autoComplete="username" {...register('username')} /></Field>
      <Field label="密码" error={errors.password?.message}><Input type="password" autoComplete="current-password" {...register('password')} /></Field>
      <Button className="mt-2" disabled={isSubmitting}>{isSubmitting && <LoaderCircle className="size-4 animate-spin" />}登录</Button>
    </form>
  </AuthFrame>
}

export function BootstrapPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const [operationError, setOperationError] = useState('')
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<BootstrapValues>({ resolver: zodResolver(bootstrapSchema), defaultValues: { username: '', display_name: '', password: '' } })
  if (auth.state?.initialized) return <Navigate to={auth.state.authenticated ? '/' : '/login'} replace />
  const submit = handleSubmit(async (values) => {
    setOperationError('')
    try { await auth.bootstrap(values); navigate('/login', { replace: true }) } catch (error) { setOperationError(errorMessage(error)) }
  })
  return <AuthFrame title="初始化管理后台" description="创建首个 Owner 账号，仅可执行一次">
    <form className="grid gap-4" onSubmit={submit}>
      {operationError && <Alert tone="danger">{operationError}</Alert>}
      <Field label="用户名" error={errors.username?.message}><Input autoFocus autoComplete="username" {...register('username')} /></Field>
      <Field label="显示名称" error={errors.display_name?.message}><Input autoComplete="name" {...register('display_name')} /></Field>
      <Field label="密码" error={errors.password?.message}><Input type="password" autoComplete="new-password" {...register('password')} /></Field>
      <Button className="mt-2" disabled={isSubmitting}>{isSubmitting && <LoaderCircle className="size-4 animate-spin" />}创建 Owner</Button>
    </form>
  </AuthFrame>
}

export function AuthLoading() {
  return <main className="grid min-h-screen place-items-center"><div className="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle className="size-4 animate-spin" />正在验证登录状态…</div></main>
}
