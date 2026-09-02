import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { useQuery } from '@tanstack/react-query'
import { Check, Clipboard, Search } from 'lucide-react'
import { useEffect, useRef, useState, type ReactNode } from 'react'
import type { OptionItem } from '../../lib/types'
import { listJSON } from '../../lib/api'
import { cn } from '../../lib/utils'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { EmptyState, ErrorState, LoadingState } from './states'

echarts.use([LineChart, GridComponent, TooltipComponent, CanvasRenderer])

export type OperationTab = { key: string; label: string; capability?: string }

export function OperationTabs({ tabs, active, onChange, can = () => true }: { tabs: OperationTab[]; active: string; onChange: (key: string) => void; can?: (capability: string) => boolean }) {
  const visible = tabs.filter((tab) => !tab.capability || can(tab.capability))
  return <div className="flex flex-wrap gap-1 rounded-lg border bg-surface p-1">{visible.map((tab) => <button key={tab.key} type="button" onClick={() => onChange(tab.key)} className={cn('rounded-md px-3 py-2 text-sm font-medium text-muted-foreground', active === tab.key && 'bg-primary text-primary-foreground')}>{tab.label}</button>)}</div>
}

export function KpiGrid({ children }: { children: ReactNode }) { return <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">{children}</div> }
export function Kpi({ label, value, detail, tone = 'neutral' }: { label: string; value: ReactNode; detail?: string; tone?: 'neutral' | 'success' | 'warning' | 'danger' | 'info' }) {
  return <div className={cn('rounded-lg border bg-surface p-4', tone === 'warning' && 'border-warning/40', tone === 'danger' && 'border-danger/40', tone === 'success' && 'border-success/40', tone === 'info' && 'border-info/40')}><p className="text-xs text-muted-foreground">{label}</p><p className="mt-2 text-2xl font-semibold tracking-tight">{value}</p>{detail && <p className="mt-1 text-xs text-muted-foreground">{detail}</p>}</div>
}

export function FilterBar({ children }: { children: ReactNode }) { return <div className="flex flex-wrap items-end gap-3 rounded-lg border bg-surface p-4">{children}</div> }
export function FilterField({ label, children }: { label: string; children: ReactNode }) { return <label className="grid min-w-36 gap-1 text-xs text-muted-foreground"><span>{label}</span>{children}</label> }

export function JsonBlock({ value, label = '复制 JSON' }: { value: unknown; label?: string }) {
  const [copied, setCopied] = useState(false)
  const text = typeof value === 'string' ? value : JSON.stringify(value, null, 2)
  const copy = async () => { await navigator.clipboard?.writeText(text); setCopied(true); window.setTimeout(() => setCopied(false), 1200) }
  return <div className="relative"><Button variant="ghost" size="sm" className="absolute right-2 top-2" onClick={() => void copy()}><Clipboard className="size-3.5" />{copied ? '已复制' : label}</Button><pre className="max-h-80 overflow-auto rounded-md bg-surface-muted p-4 pr-24 font-mono text-xs leading-6">{text || '{}'}</pre></div>
}

export function OperationsChart({ points, unit, label, loading, error }: { points: Array<{ label: string; value: number | null }>; unit: string; label: string; loading?: boolean; error?: string }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!ref.current || points.length === 0) return
    const chart = echarts.init(ref.current)
    const primary = getComputedStyle(document.documentElement).getPropertyValue('--primary').trim()
    chart.setOption({
      animationDuration: 200,
      grid: { left: 12, right: 18, top: 18, bottom: 25, containLabel: true },
      xAxis: { type: 'category', data: points.map((point) => point.label), axisTick: { show: false } },
      yAxis: { type: 'value', axisLabel: { formatter: `{value}${unit}` }, splitLine: { lineStyle: { color: 'rgba(127,127,127,.16)' } } },
      tooltip: { trigger: 'axis', valueFormatter: (value: unknown) => value == null ? '不可用' : `${value}${unit}` },
      series: [{ type: 'line', data: points.map((point) => point.value), connectNulls: false, showSymbol: points.length < 16, symbolSize: 6, lineStyle: { color: primary, width: 2 }, itemStyle: { color: primary } }],
    })
    const resize = () => chart.resize()
    window.addEventListener('resize', resize)
    return () => { window.removeEventListener('resize', resize); chart.dispose() }
  }, [points, unit])
  if (loading) return <LoadingState />
  if (error) return <ErrorState message={error} />
  if (points.length === 0) return <EmptyState />
  return <div ref={ref} className="h-64 w-full" role="img" aria-label={label} />
}

export function RemotePicker({ endpoint, value, onChange, placeholder = '搜索并选择…', disabled }: { endpoint: string; value: OptionItem | null; onChange: (value: OptionItem | null) => void; placeholder?: string; disabled?: boolean }) {
  const [search, setSearch] = useState('')
  const query = useQuery({ queryKey: ['remote-picker', endpoint, search], queryFn: () => listJSON<OptionItem>(endpoint, 1, 50, search), enabled: !disabled })
  return <div className="grid gap-2"><div className="relative"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><Input className="pl-9" value={search} disabled={disabled} onChange={(event) => setSearch(event.target.value)} placeholder={placeholder} /></div>{value && <button type="button" onClick={() => onChange(null)} className="flex items-center justify-between rounded-md border border-primary/30 bg-primary/5 px-3 py-2 text-left text-sm"><span><span className="font-medium">{value.label}</span>{value.extra && <span className="ml-2 text-xs text-muted-foreground">{value.extra}</span>}</span><span className="text-xs text-primary">清除</span></button>}<div className="max-h-48 overflow-auto rounded-md border bg-surface p-1">{query.isLoading ? <p className="p-3 text-sm text-muted-foreground">加载中…</p> : query.error ? <p className="p-3 text-sm text-danger">{query.error.message}</p> : (query.data?.list ?? []).length === 0 ? <p className="p-3 text-sm text-muted-foreground">无匹配项</p> : query.data!.list.map((option) => <button key={option.id} type="button" onClick={() => onChange(option)} className="flex w-full items-center gap-2 rounded px-2 py-2 text-left text-sm hover:bg-surface-muted"><span className="grid size-4 place-items-center">{String(value?.id) === String(option.id) && <Check className="size-4 text-primary" />}</span><span className="min-w-0"><span className="block truncate font-medium">{option.label}</span>{option.extra && <span className="block truncate text-xs text-muted-foreground">{option.extra}</span>}</span></button>)}</div></div>
}
