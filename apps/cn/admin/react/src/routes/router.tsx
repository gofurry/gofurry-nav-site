import { lazy } from 'react'
import { createBrowserRouter } from 'react-router-dom'
import { AppShell } from '../components/admin/app-shell'
import { AuthGuard, CapabilityGuard } from '../features/auth/auth-guard'
import { BootstrapPage, LoginPage } from '../features/auth/auth-pages'
import { NotFoundPage, PlaceholderPage } from '../pages/placeholder-page'

const WorkbenchPage = lazy(() => import('../features/workbench/workbench-page').then((module) => ({ default: module.WorkbenchPage })))
const SiteListPage = lazy(() => import('../features/sites/site-pages').then((module) => ({ default: module.SiteListPage })))
const SiteWorkspacePage = lazy(() => import('../features/sites/site-pages').then((module) => ({ default: module.SiteWorkspacePage })))
const GameListPage = lazy(() => import('../features/games/game-pages').then((module) => ({ default: module.GameListPage })))
const GameWorkspacePage = lazy(() => import('../features/games/game-pages').then((module) => ({ default: module.GameWorkspacePage })))
const ResourceEngineBoundary = lazy(() => import('../features/resources/resource-page').then((module) => ({ default: module.ResourceEngineBoundary })))

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
        { element: <CapabilityGuard capability="collection.read" />, children: [{ path: 'collection', element: <PlaceholderPage /> }] },
        { element: <CapabilityGuard capability="metrics.read" />, children: [{ path: 'metrics', element: <PlaceholderPage /> }] },
        { element: <CapabilityGuard capability="changes.read" />, children: [{ path: 'changes', element: <PlaceholderPage /> }] },
        { element: <CapabilityGuard capability="data_ops.read" />, children: [{ path: 'system/data-operations', element: <PlaceholderPage /> }] },
        { element: <CapabilityGuard capability="audit.read" />, children: [{ path: 'system/audit', element: <PlaceholderPage /> }] },
        { element: <CapabilityGuard capability="account.manage" />, children: [{ path: 'system/accounts', element: <PlaceholderPage /> }] },
        { path: '*', element: <NotFoundPage /> },
      ],
    }],
  },
])
