import { lazy } from 'react'
import { createBrowserRouter } from 'react-router-dom'
import { AppShell } from '../components/admin/app-shell'
import { AuthGuard, CapabilityGuard } from '../features/auth/auth-guard'
import { BootstrapPage, LoginPage } from '../features/auth/auth-pages'
import { DATAOPS_READ_CAPABILITY } from '../lib/capabilities'
import { NotFoundPage } from '../pages/placeholder-page'

const WorkbenchPage = lazy(() => import('../features/workbench/workbench-page').then((module) => ({ default: module.WorkbenchPage })))
const SiteListPage = lazy(() => import('../features/sites/site-pages').then((module) => ({ default: module.SiteListPage })))
const SiteWorkspacePage = lazy(() => import('../features/sites/site-pages').then((module) => ({ default: module.SiteWorkspacePage })))
const GameListPage = lazy(() => import('../features/games/game-pages').then((module) => ({ default: module.GameListPage })))
const GameWorkspacePage = lazy(() => import('../features/games/game-pages').then((module) => ({ default: module.GameWorkspacePage })))
const ResourceEngineBoundary = lazy(() => import('../features/resources/resource-page').then((module) => ({ default: module.ResourceEngineBoundary })))
const CollectionPage = lazy(() => import('../features/operations/collection-page').then((module) => ({ default: module.CollectionPage })))
const MetricsPage = lazy(() => import('../features/operations/metrics-page').then((module) => ({ default: module.MetricsPage })))
const ChangesPage = lazy(() => import('../features/operations/changes-page').then((module) => ({ default: module.ChangesPage })))
const DataOperationsPage = lazy(() => import('../features/system/data-operations-page').then((module) => ({ default: module.DataOperationsPage })))
const AuditPage = lazy(() => import('../features/system/audit-page').then((module) => ({ default: module.AuditPage })))
const AccountsPage = lazy(() => import('../features/system/accounts-page').then((module) => ({ default: module.AccountsPage })))

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/setup', element: <BootstrapPage /> },
  {
    element: <AuthGuard />,
    children: [{
      element: <AppShell />,
      children: [
        { element: <CapabilityGuard capability="content.read" />, children: [
          { index: true, element: <WorkbenchPage /> },
          { path: 'nav/sites', element: <SiteListPage /> },
          { path: 'nav/sites/:id', element: <SiteWorkspacePage /> },
          { path: 'nav/:resource', element: <ResourceEngineBoundary /> },
          { path: 'game/games', element: <GameListPage /> },
          { path: 'game/games/:id', element: <GameWorkspacePage /> },
          { path: 'game/:resource', element: <ResourceEngineBoundary /> },
        ] },
        { element: <CapabilityGuard capability="collection.read" />, children: [{ path: 'collection', element: <CollectionPage /> }] },
        { element: <CapabilityGuard capability="metrics.read" />, children: [{ path: 'metrics', element: <MetricsPage /> }] },
        { element: <CapabilityGuard capability="changes.read" />, children: [{ path: 'changes', element: <ChangesPage /> }] },
        { element: <CapabilityGuard capability={DATAOPS_READ_CAPABILITY} />, children: [{ path: 'system/data-operations', element: <DataOperationsPage /> }] },
        { element: <CapabilityGuard capability="audit.read" />, children: [{ path: 'system/audit', element: <AuditPage /> }] },
        { element: <CapabilityGuard capability="account.manage" />, children: [{ path: 'system/accounts', element: <AccountsPage /> }] },
        { path: '*', element: <NotFoundPage /> },
      ],
    }],
  },
])
