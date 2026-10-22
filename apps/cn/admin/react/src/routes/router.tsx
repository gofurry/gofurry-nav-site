import { lazy } from 'react'
import { createBrowserRouter } from 'react-router-dom'
import { AppShell } from '../components/admin/app-shell'
import { AuthGuard, CapabilityGuard } from '../features/auth/auth-guard'
import { BootstrapPage, LoginPage } from '../features/auth/auth-pages'
import { DATAOPS_READ_CAPABILITY } from '../lib/capabilities'
import { NotFoundPage } from '../pages/not-found-page'

const CollaborationPage = lazy(() => import('../features/collaboration/collaboration-page').then((module) => ({ default: module.CollaborationPage })))
const WorkbenchPage = lazy(() => import('../features/workbench/workbench-page').then((module) => ({ default: module.WorkbenchPage })))
const SiteListPage = lazy(() => import('../features/sites/site-pages').then((module) => ({ default: module.SiteListPage })))
const SiteWorkspacePage = lazy(() => import('../features/sites/site-pages').then((module) => ({ default: module.SiteWorkspacePage })))
const GroupCurationPage = lazy(() => import('../features/sites/group-curation-page').then((module) => ({ default: module.GroupCurationPage })))
const HeroAssetsPage = lazy(() => import('../features/assets/asset-pages').then((module) => ({ default: module.HeroAssetsPage })))
const GameListPage = lazy(() => import('../features/games/game-pages').then((module) => ({ default: module.GameListPage })))
const GameWorkspacePage = lazy(() => import('../features/games/game-pages').then((module) => ({ default: module.GameWorkspacePage })))
const CollectionListPage = lazy(() => import('../features/game-collections/collection-list-page').then(module => ({ default: module.CollectionListPage })))
const CreateCollectionPage = lazy(() => import('../features/game-collections/collection-workspace-page').then(module => ({ default: module.CreateCollectionPage })))
const CollectionWorkspacePage = lazy(() => import('../features/game-collections/collection-workspace-page').then(module => ({ default: module.CollectionWorkspacePage })))
const CollectionHomeCurationPage = lazy(() => import('../features/game-collections/home-curation-page').then(module => ({ default: module.CollectionHomeCurationPage })))
const ShowcasePage = lazy(() => import('../features/showcase/showcase-page').then((module) => ({ default: module.ShowcasePage })))
const ShowcaseCampaignPage = lazy(() => import('../features/showcase/showcase-campaign-page').then((module) => ({ default: module.ShowcaseCampaignPage })))
const ReleaseNoteListPage = lazy(() => import('../features/release-notes/release-note-pages').then((module) => ({ default: module.ReleaseNoteListPage })))
const ReleaseNoteEditorPage = lazy(() => import('../features/release-notes/release-note-pages').then((module) => ({ default: module.ReleaseNoteEditorPage })))
const ResourceEngineBoundary = lazy(() => import('../features/resources/resource-page').then((module) => ({ default: module.ResourceEngineBoundary })))
const CollectionPage = lazy(() => import('../features/operations/collection-page').then((module) => ({ default: module.CollectionPage })))
const MetricsPage = lazy(() => import('../features/operations/metrics-page').then((module) => ({ default: module.MetricsPage })))
const ChangesPage = lazy(() => import('../features/operations/changes-page').then((module) => ({ default: module.ChangesPage })))
const DataOperationsPage = lazy(() => import('../features/system/data-operations-page').then((module) => ({ default: module.DataOperationsPage })))
const CloudOperationsPage = lazy(() => import('../features/system/cloud-operations-page').then((module) => ({ default: module.CloudOperationsPage })))
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
          { path: 'nav/site-groups/:id/curation', element: <GroupCurationPage /> },
          { path: 'nav/hero-assets', element: <HeroAssetsPage /> },
          { path: 'nav/update-notices', element: <ReleaseNoteListPage /> },
          { path: 'nav/update-notices/new', element: <ReleaseNoteEditorPage /> },
          { path: 'nav/update-notices/:id', element: <ReleaseNoteEditorPage /> },
          { path: 'nav/:resource', element: <ResourceEngineBoundary section="nav" /> },
          { path: 'game/games', element: <GameListPage /> },
          { path: 'game/games/:id', element: <GameWorkspacePage /> },
          { path: 'game/collections', element: <CollectionListPage /> },
          { path: 'game/collections/new', element: <CreateCollectionPage /> },
          { path: 'game/collections/home-curation', element: <CollectionHomeCurationPage /> },
          { path: 'game/collections/:id', element: <CollectionWorkspacePage /> },
          { path: 'game/showcase', element: <ShowcasePage /> },
          { path: 'game/showcase/:id', element: <ShowcaseCampaignPage /> },
          { path: 'game/:resource', element: <ResourceEngineBoundary section="game" /> },
        ] },
        { element: <CapabilityGuard capability="collaboration.read" />, children: [{ path: 'collaboration', element: <CollaborationPage /> }] },
        { element: <CapabilityGuard capability="collection.read" />, children: [{ path: 'collection', element: <CollectionPage /> }] },
        { element: <CapabilityGuard capability="metrics.read" />, children: [{ path: 'metrics', element: <MetricsPage /> }] },
        { element: <CapabilityGuard capability="changes.read" />, children: [{ path: 'changes', element: <ChangesPage /> }] },
        { element: <CapabilityGuard capability={DATAOPS_READ_CAPABILITY} />, children: [{ path: 'system/data-operations', element: <DataOperationsPage /> }] },
        { element: <CapabilityGuard capability="cloudops.read" />, children: [{ path: 'system/cloud', element: <CloudOperationsPage /> }] },
        { element: <CapabilityGuard capability="audit.read" />, children: [{ path: 'system/audit', element: <AuditPage /> }] },
        { element: <CapabilityGuard capability="account.manage" />, children: [{ path: 'system/accounts', element: <AccountsPage /> }] },
        { path: '*', element: <NotFoundPage /> },
      ],
    }],
  },
])
