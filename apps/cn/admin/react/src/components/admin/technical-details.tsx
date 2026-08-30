import { Braces } from 'lucide-react'
import type { ReactNode } from 'react'
import { Sheet } from '../ui/sheet'

export function TechnicalDetails({ open, onOpenChange, title, identifier, children }: { open: boolean; onOpenChange: (open: boolean) => void; title: string; identifier?: string; children: ReactNode }) {
  return <Sheet open={open} onOpenChange={onOpenChange} title={`技术详情 · ${title}`} description={identifier}>
    <div className="mb-5 flex items-center gap-2 rounded-md bg-surface-muted p-3 text-xs text-muted-foreground"><Braces className="size-4" />技术字段用于排障和数据核对，不影响日常内容操作。</div>
    {children}
  </Sheet>
}
