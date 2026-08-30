import { Construction, ExternalLink } from 'lucide-react'
import { useLocation } from 'react-router-dom'
import { PageHeader } from '../components/admin/page'
import { StatusBadge } from '../components/admin/status'

const labels: Record<string, { title: string; stage: string; legacy?: string; message: string }> = {
  '/collection': { title: '采集中心', stage: 'P0.5.2-C', legacy: '/collection', message: '本阶段仅在 Site/Game Workspace 复用已有立即采集契约，不重构 Collection Center。' },
  '/metrics': { title: '数据指标', stage: 'P0.5.2-C', legacy: '/metrics', message: 'Metric Center 的信息架构与异常解释将在下一阶段迁移。' },
  '/changes': { title: '变化事件', stage: 'P0.5.2-C', legacy: '/changes', message: '当前 Site/Game History 已读取受支持的实体变化子集，完整中心随后迁移。' },
  '/system/data-operations': { title: '数据运维', stage: '#77 / P0.5.2-C', message: 'Data Operations 不在 P0.5.2-B 范围内。' },
  '/system/audit': { title: '操作审计', stage: 'P0.5.2-C', message: '完整 Audit UI 与实体级聚合将在后续阶段实现。' },
  '/system/accounts': { title: '账号与权限', stage: '后续阶段', message: 'P0.5.2-A 的账号与授权后端契约保持不变，本阶段不实现完整账号管理 UI。' },
}

export function PlaceholderPage() {
  const location = useLocation()
  const item = labels[location.pathname] ?? { title: '待迁移页面', stage: '后续阶段', message: '该页面不在本阶段实现范围。' }
  const legacyOrigin = import.meta.env.VITE_LEGACY_ADMIN_ORIGIN || 'http://127.0.0.1:10099'
  return <div className="grid gap-6"><PageHeader title={item.title} description="React Admin 分阶段迁移边界" /><div className="grid min-h-80 place-items-center rounded-lg border bg-surface p-8 text-center"><div><div className="mx-auto grid size-12 place-items-center rounded-full bg-surface-muted"><Construction className="size-5 text-muted-foreground" /></div><div className="mt-4"><StatusBadge tone="info">{item.stage}</StatusBadge></div><h2 className="mt-3 text-lg font-semibold">保留兼容入口</h2><p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-muted-foreground">{item.message}</p>{item.legacy && <a href={`${legacyOrigin}${item.legacy}`} className="mt-5 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline">打开现有 Vue 页面<ExternalLink className="size-4" /></a>}</div></div></div>
}

export function NotFoundPage() { return <div className="grid min-h-[60vh] place-items-center text-center"><div><p className="font-mono text-sm text-muted-foreground">404</p><h1 className="mt-2 text-2xl font-semibold">页面不存在</h1><a href="/" className="mt-4 inline-block text-sm text-primary hover:underline">返回工作台</a></div></div> }
