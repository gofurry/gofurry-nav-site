import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { ErrorState } from '../../components/admin/states'
import { useAuth } from './auth-context'
import { AuthLoading } from './auth-pages'

export function AuthGuard() {
  const auth = useAuth()
  const location = useLocation()
  if (auth.loading) return <AuthLoading />
  if (auth.error) return <ErrorState title="无法连接管理服务" message={auth.error.message} onRetry={() => void auth.reload()} />
  if (!auth.state?.initialized) return <Navigate to="/setup" replace />
  if (!auth.state.authenticated) return <Navigate to="/login" state={{ from: `${location.pathname}${location.search}` }} replace />
  return <Outlet />
}

export function CapabilityGuard({ capability }: { capability: string }) {
  const auth = useAuth()
  if (!auth.can(capability)) return <ErrorState title="无权访问" message="当前账号不具备访问此工作区所需的能力。" />
  return <Outlet />
}
