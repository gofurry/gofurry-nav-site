import { CircleAlert, Info } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

export function Alert({ children, tone = 'info' }: { children: ReactNode; tone?: 'info' | 'danger' | 'warning' }) {
  const Icon = tone === 'info' ? Info : CircleAlert
  const toneClass = tone === 'info' ? 'border-info/35 bg-info/8 text-info' : tone === 'danger' ? 'border-danger/35 bg-danger/8 text-danger' : 'border-warning/40 bg-warning/8 text-warning'
  return <div role="alert" className={cn('flex gap-3 rounded-md border p-3 text-sm', toneClass)}>
    <Icon className="mt-0.5 size-4 shrink-0" /><div className="text-foreground">{children}</div>
  </div>
}
