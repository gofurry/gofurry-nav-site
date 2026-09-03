import { Menu } from '@base-ui/react/menu'
import { CaretLeft, CaretRight, Chats, ClockCounterClockwise, Database, GameController, Gauge, GearSix, ListBullets, MagnifyingGlass, Megaphone, Moon, Pulse, Quotes, ShieldCheck, SignOut, Sparkle, SquaresFour, Sun, Tag, UserCircle, Wrench } from '@phosphor-icons/react'
import { Suspense, useEffect, useState, type ComponentType } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useTheme, type ThemeMode } from '../../app/theme'
import { useAuth } from '../../features/auth/auth-context'
import { DATAOPS_READ_CAPABILITY } from '../../lib/capabilities'
import { isGlobalSearchShortcut } from '../../lib/keyboard'
import { cn } from '../../lib/utils'
import { Button } from '../ui/button'
import { GlobalSearch } from './global-search'
import { LoadingState } from './states'

type NavEntry = { label: string; href: string; icon: ComponentType<{ className?: string }>; capability?: string }
type NavGroup = { label: string; entries: NavEntry[] }

export const navigationGroups: NavGroup[] = [
  { label: '', entries: [{ label: '工作台', href: '/', icon: SquaresFour, capability: 'content.read' }] },
  { label: '导航内容', entries: [
    { label: '网站', href: '/nav/sites', icon: Sparkle, capability: 'content.read' },
    { label: '网站分组', href: '/nav/site-groups', icon: ListBullets, capability: 'content.read' },
    { label: '更新公告', href: '/nav/update-notices', icon: Megaphone, capability: 'content.read' },
    { label: '金句', href: '/nav/sayings', icon: Quotes, capability: 'content.read' },
  ] },
  { label: '游戏内容', entries: [
    { label: '游戏', href: '/game/games', icon: GameController, capability: 'content.read' },
    { label: '标签', href: '/game/tags', icon: Tag, capability: 'content.read' },
    { label: '评论', href: '/game/comments', icon: Chats, capability: 'content.read' },
    { label: '抽奖', href: '/game/prizes', icon: Sparkle, capability: 'content.read' },
  ] },
  { label: '数据运营', entries: [
    { label: '采集', href: '/collection', icon: Pulse, capability: 'collection.read' },
    { label: '数据指标', href: '/metrics', icon: Gauge, capability: 'metrics.read' },
    { label: '变化事件', href: '/changes', icon: ClockCounterClockwise, capability: 'changes.read' },
  ] },
  { label: '系统', entries: [
    { label: '数据运维', href: '/system/data-operations', icon: Database, capability: DATAOPS_READ_CAPABILITY },
    { label: '操作审计', href: '/system/audit', icon: ShieldCheck, capability: 'audit.read' },
    { label: '账号与权限', href: '/system/accounts', icon: UserCircle, capability: 'account.manage' },
  ] },
]

export function capabilityAwareNavigation(can: (capability: string) => boolean) {
  return navigationGroups.map((group) => ({ ...group, entries: group.entries.filter((entry) => !entry.capability || can(entry.capability)) })).filter((group) => group.entries.length > 0)
}

const breadcrumbLabels: Record<string, string> = { nav: '导航内容', game: '游戏内容', sites: '网站', games: '游戏', 'site-groups': '网站分组', 'update-notices': '更新公告', sayings: '金句', tags: '标签', comments: '评论', prizes: '抽奖', collection: '采集', metrics: '数据指标', changes: '变化事件', system: '系统', 'data-operations': '数据运维', audit: '操作审计', accounts: '账号与权限' }

function Breadcrumbs() {
  const location = useLocation()
  const parts = location.pathname.split('/').filter(Boolean)
  if (parts.length === 0) return <span>工作台</span>
  return <div className="flex min-w-0 items-center gap-1.5 overflow-hidden text-sm text-muted-foreground">{parts.map((part, index) => <span key={`${part}-${index}`} className="flex min-w-0 items-center gap-1.5">{index > 0 && <span>/</span>}<span className={cn('truncate', index === parts.length - 1 && 'text-foreground')}>{/^\d+$/.test(part) ? `#${part}` : (breadcrumbLabels[part] ?? part)}</span></span>)}</div>
}

export function AdminBrand({ collapsed }: { collapsed: boolean }) {
  return <div className={cn('flex h-14 items-center border-b', collapsed ? 'justify-center px-2' : 'px-4')}>
    <p className={cn('truncate font-semibold leading-none tracking-tight', collapsed ? 'text-sm' : 'text-base')}>{collapsed ? 'GF' : 'GoFurry'}</p>
  </div>
}

export function AppShell() {
  const auth = useAuth()
  const navigate = useNavigate()
  const { mode, setMode } = useTheme()
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem('gofurry-admin-sidebar') === 'collapsed')
  const [searchOpen, setSearchOpen] = useState(false)
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null
      const typing = target?.matches('input, textarea, select, [contenteditable="true"]')
      if (isGlobalSearchShortcut(event, Boolean(typing))) { event.preventDefault(); setSearchOpen(true) }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])
  const toggleSidebar = () => setCollapsed((current) => { const next = !current; localStorage.setItem('gofurry-admin-sidebar', next ? 'collapsed' : 'expanded'); return next })
  const logout = async () => { await auth.logout(); navigate('/login', { replace: true }) }
  const themeIcon = mode === 'dark' ? Moon : mode === 'light' ? Sun : GearSix
  const ThemeIcon = themeIcon
  const visibleGroups = capabilityAwareNavigation(auth.can)

  return <div data-admin-shell className="h-dvh overflow-hidden bg-background">
    <aside data-admin-sidebar className={cn('fixed inset-y-0 left-0 z-30 flex h-dvh flex-col border-r bg-surface transition-[width] duration-200', collapsed ? 'w-16' : 'w-60')}>
      <AdminBrand collapsed={collapsed} />
      <nav className="admin-scroll min-h-0 flex-1 overflow-y-auto p-2">{visibleGroups.map((group) => {
        return <div key={group.label || 'workbench'} className="mb-4">{group.label && !collapsed && <p className="mb-1 px-2 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">{group.label}</p>}{group.entries.map((entry) => <NavLink key={entry.href} to={entry.href} end={entry.href === '/'} title={collapsed ? entry.label : undefined} className={({ isActive }) => cn('mb-0.5 flex h-9 items-center gap-3 rounded-md px-2 text-sm text-muted-foreground hover:bg-surface-muted hover:text-foreground', collapsed && 'justify-center px-0', isActive && 'bg-primary/10 font-medium text-primary')}><entry.icon className="size-4 shrink-0" />{!collapsed && <span>{entry.label}</span>}</NavLink>)}</div>
      })}</nav>
      <button type="button" onClick={toggleSidebar} className="flex h-11 shrink-0 items-center justify-center border-t text-muted-foreground hover:bg-surface-muted" aria-label={collapsed ? '展开侧边栏' : '收起侧边栏'}>{collapsed ? <CaretRight className="size-4" /> : <><CaretLeft className="mr-2 size-4" /><span className="text-xs">收起导航</span></>}</button>
    </aside>
    <div data-admin-workspace className={cn('flex h-dvh min-w-0 flex-col overflow-hidden transition-[padding] duration-200', collapsed ? 'pl-16' : 'pl-60')}>
      <header className="z-20 flex h-14 min-w-0 shrink-0 items-center justify-between gap-3 border-b bg-background/92 px-3 backdrop-blur sm:px-6"><Breadcrumbs /><div className="flex shrink-0 items-center gap-1 sm:gap-2">
        <Button variant="secondary" className="w-9 justify-center overflow-hidden px-0 text-muted-foreground xl:w-56 xl:justify-start xl:px-3" onClick={() => setSearchOpen(true)} aria-label="全局搜索"><MagnifyingGlass className="size-4 shrink-0" /><span className="hidden flex-1 text-left xl:inline">全局搜索</span><kbd className="hidden rounded border bg-surface-muted px-1.5 font-mono text-[10px] xl:inline">Ctrl K</kbd></Button>
        <Button variant="ghost" size="icon" title={auth.can(DATAOPS_READ_CAPABILITY) ? '数据库与迁移状态' : '当前账号无 dataops.read'} disabled={!auth.can(DATAOPS_READ_CAPABILITY)} onClick={() => navigate('/system/data-operations')}><Wrench className="size-4" /></Button>
        <Menu.Root><Menu.Trigger render={<Button variant="ghost" size="icon" aria-label="主题设置" />}><ThemeIcon className="size-4" /></Menu.Trigger><Menu.Portal><Menu.Positioner className="z-40" sideOffset={4} align="end"><Menu.Popup className="min-w-36 rounded-md border bg-surface p-1 shadow-lg outline-none"><Menu.RadioGroup value={mode} onValueChange={(value) => setMode(value as ThemeMode)}>{([['system', '跟随系统'], ['light', '浅色'], ['dark', '深色']] as Array<[ThemeMode, string]>).map(([value, label]) => <Menu.RadioItem key={value} value={value} closeOnClick className="cursor-default rounded px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-surface-muted">{label}</Menu.RadioItem>)}</Menu.RadioGroup></Menu.Popup></Menu.Positioner></Menu.Portal></Menu.Root>
        <Menu.Root><Menu.Trigger render={<Button variant="secondary" size="icon" aria-label="账号菜单" />}><UserCircle className="size-4" /></Menu.Trigger><Menu.Portal><Menu.Positioner className="z-40" sideOffset={4} align="end"><Menu.Popup className="min-w-52 rounded-md border bg-surface p-1 shadow-lg outline-none"><div className="border-b px-2 py-2"><p className="text-sm font-medium">{auth.state?.identity?.display_name}</p><p className="mt-0.5 font-mono text-xs text-muted-foreground">{auth.state?.identity?.username} · {auth.state?.identity?.role}</p></div><Menu.Item onClick={() => void logout()} className="mt-1 flex cursor-default items-center gap-2 rounded px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-surface-muted"><SignOut className="size-4" />退出登录</Menu.Item></Menu.Popup></Menu.Positioner></Menu.Portal></Menu.Root>
      </div></header>
      <main data-admin-workspace-scroll className="admin-scroll min-h-0 flex-1 overflow-x-hidden overflow-y-auto"><div className="mx-auto w-full max-w-[1680px] p-3 sm:p-6"><Suspense fallback={<LoadingState label="正在加载工作区…" />}><Outlet /></Suspense></div></main>
    </div>
    <GlobalSearch open={searchOpen} onOpenChange={setSearchOpen} />
  </div>
}
