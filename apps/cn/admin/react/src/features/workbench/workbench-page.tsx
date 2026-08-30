import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { useQuery } from '@tanstack/react-query'
import { Activity, Gamepad2, Megaphone, Plus, Sparkles, UserRound } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { PageHeader, Section } from '../../components/admin/page'
import { StatusBadge, TechnicalLabel } from '../../components/admin/status'
import { Button } from '../../components/ui/button'
import { listJSON } from '../../lib/api'
import { useAuth } from '../auth/auth-context'

echarts.use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

async function contentTotals() {
  const [sites, games, notices, sayings] = await Promise.all([
    listJSON('/api/v1/nav/sites', 1, 1), listJSON('/api/v1/game/games', 1, 1),
    listJSON('/api/v1/nav/update-notices', 1, 1), listJSON('/api/v1/nav/sayings', 1, 1),
  ])
  return { sites: sites.total, games: games.total, notices: notices.total, sayings: sayings.total }
}

function ContentChart({ data }: { data?: { sites: number; games: number; notices: number; sayings: number } }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current || !data) return
    const chart = echarts.init(ref.current)
    const primary = getComputedStyle(document.documentElement).getPropertyValue('--primary').trim()
    chart.setOption({
      animationDuration: 250,
      grid: { left: 12, right: 12, top: 12, bottom: 24, containLabel: true },
      xAxis: { type: 'category', data: ['网站', '游戏', '公告', '金句'], axisTick: { show: false } },
      yAxis: { type: 'value', minInterval: 1, splitLine: { lineStyle: { color: 'rgba(127,127,127,.16)' } } },
      tooltip: { trigger: 'axis' },
      series: [{ type: 'bar', data: [data.sites, data.games, data.notices, data.sayings], itemStyle: { color: primary, borderRadius: [4, 4, 0, 0] }, barMaxWidth: 46 }],
    })
    const resize = () => chart.resize()
    window.addEventListener('resize', resize)
    return () => { window.removeEventListener('resize', resize); chart.dispose() }
  }, [data])
  return <div ref={ref} className="h-64 w-full" role="img" aria-label="内容数量柱状图" />
}

export function WorkbenchPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const query = useQuery({ queryKey: ['workbench-content-totals'], queryFn: contentTotals })
  const legacyOrigin = import.meta.env.VITE_LEGACY_ADMIN_ORIGIN || 'http://127.0.0.1:10099'
  const identity = auth.state?.identity
  return <div className="grid gap-6"><PageHeader title="工作台" description="内容运营入口与当前账号上下文。" eyebrow="workbench" /><div className="grid gap-5 xl:grid-cols-[1fr_2fr]"><Section title="当前用户"><div className="flex items-start gap-4"><div className="grid size-11 place-items-center rounded-full bg-primary/10 text-primary"><UserRound className="size-5" /></div><div><p className="font-semibold">{identity?.display_name}</p><p className="mt-1 text-sm text-muted-foreground">@{identity?.username}</p><div className="mt-3 flex items-center gap-2"><StatusBadge tone="success">{identity?.role}</StatusBadge><TechnicalLabel>{identity?.capabilities.length ?? 0} capabilities</TechnicalLabel></div></div></div></Section><Section title="快捷操作" description="仅显示当前 principal 具备 capability 的动作。"><div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">{auth.can('content.write') && <><Button className="h-auto justify-start p-4" onClick={() => navigate('/nav/sites/new')}><Plus className="size-4" />新增网站</Button><Button variant="secondary" className="h-auto justify-start p-4" onClick={() => navigate('/game/games/new')}><Gamepad2 className="size-4" />新增游戏</Button><Button variant="secondary" className="h-auto justify-start p-4" onClick={() => navigate('/nav/update-notices')}><Megaphone className="size-4" />发布公告</Button></>}{auth.can('collection.execute') && <a href={`${legacyOrigin}/collection`} className="inline-flex items-center gap-2 rounded-md border bg-surface px-4 py-3 text-sm font-medium hover:bg-surface-muted"><Activity className="size-4" />手动采集</a>}</div></Section></div><div className="grid gap-5 xl:grid-cols-[2fr_1fr]"><Section title="内容概览" description="来自现有内容列表 API 的低成本汇总。">{query.error ? <p className="py-10 text-center text-sm text-danger">{query.error.message}</p> : <ContentChart data={query.data} />}</Section><Section title="运营边界" description="P0.5.2-B 保持内容迁移聚焦。"><div className="grid gap-3"><div className="flex items-center gap-3 rounded-md bg-surface-muted p-3"><Sparkles className="size-4 text-primary" /><div><p className="text-sm font-medium">内容工作区</p><p className="text-xs text-muted-foreground">React 原生</p></div></div><div className="flex items-center gap-3 rounded-md bg-surface-muted p-3"><Activity className="size-4 text-info" /><div><p className="text-sm font-medium">数据与系统中心</p><p className="text-xs text-muted-foreground">P0.5.2-C 接续</p></div></div><Link to="/nav/sites" className="text-sm font-medium text-primary hover:underline">进入网站工作区 →</Link></div></Section></div></div>
}
