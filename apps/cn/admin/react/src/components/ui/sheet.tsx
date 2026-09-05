import { Dialog as BaseDialog } from '@base-ui/react/dialog'
import { X } from '@phosphor-icons/react'
import type { ReactNode } from 'react'

export function Sheet({ open, onOpenChange, title, description, children, footer }: {
  open: boolean; onOpenChange: (open: boolean) => void; title: string; description?: string; children: ReactNode; footer?: ReactNode
}) {
  return <BaseDialog.Root open={open} onOpenChange={onOpenChange}>
    <BaseDialog.Portal>
      <BaseDialog.Backdrop className="overlay-open fixed inset-0 z-50 bg-black/40" />
      <BaseDialog.Popup className="sheet-open fixed inset-y-0 right-0 z-50 flex h-dvh w-[min(42rem,92vw)] flex-col overflow-hidden border-l bg-surface shadow-2xl outline-none">
        <div className="flex shrink-0 items-start justify-between gap-4 border-b p-5">
          <div><BaseDialog.Title className="text-lg font-semibold">{title}</BaseDialog.Title>{description && <BaseDialog.Description className="mt-1 text-sm text-muted-foreground">{description}</BaseDialog.Description>}</div>
          <BaseDialog.Close aria-label="关闭" className="rounded p-1 hover:bg-surface-muted"><X className="size-5" /></BaseDialog.Close>
        </div>
        <div className="admin-scroll min-h-0 flex-1 overflow-y-auto p-5">{children}</div>
        {footer && <div className="shrink-0 border-t bg-surface px-5 py-4">{footer}</div>}
      </BaseDialog.Popup>
    </BaseDialog.Portal>
  </BaseDialog.Root>
}
