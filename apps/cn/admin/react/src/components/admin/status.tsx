import { Circle, CircleAlert, CircleCheck, CircleDashed, Info } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

export type StatusTone = 'success' | 'warning' | 'danger' | 'info' | 'neutral'
const icons = { success: CircleCheck, warning: CircleAlert, danger: CircleAlert, info: Info, neutral: CircleDashed }
const toneClasses: Record<StatusTone, string> = {
  success: 'border-success/30 bg-success/8 text-success', warning: 'border-warning/30 bg-warning/8 text-warning',
  danger: 'border-danger/30 bg-danger/8 text-danger', info: 'border-info/30 bg-info/8 text-info', neutral: 'text-muted-foreground',
}

export function StatusBadge({ tone = 'neutral', children }: { tone?: StatusTone; children: ReactNode }) {
  const Icon = icons[tone]
  return <span className={cn('inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs font-medium', toneClasses[tone])}><Icon className="size-3" />{children}</span>
}

export function TechnicalLabel({ children }: { children: ReactNode }) { return <span className="font-mono text-[11px] text-muted-foreground">{children}</span> }
export function StatusDot({ tone = 'neutral' }: { tone?: StatusTone }) { return <Circle className={cn('size-2 fill-current', toneClasses[tone])} /> }
