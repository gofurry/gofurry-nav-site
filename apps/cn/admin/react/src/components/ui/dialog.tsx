import { Dialog as BaseDialog } from '@base-ui/react/dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'
import { Button } from './button'

export function Dialog({ open, onOpenChange, title, description, children, footer }: {
  open: boolean; onOpenChange: (open: boolean) => void; title: string; description?: string
  children: ReactNode; footer?: ReactNode
}) {
  return <BaseDialog.Root open={open} onOpenChange={onOpenChange}>
    <BaseDialog.Portal>
      <BaseDialog.Backdrop className="overlay-open fixed inset-0 z-50 bg-black/45" />
      <BaseDialog.Popup className="dialog-open fixed left-1/2 top-1/2 z-50 w-[min(34rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-lg border bg-surface shadow-2xl outline-none">
        <div className="flex items-start justify-between gap-4 border-b p-5">
          <div><BaseDialog.Title className="text-base font-semibold">{title}</BaseDialog.Title>{description && <BaseDialog.Description className="mt-1 text-sm text-muted-foreground">{description}</BaseDialog.Description>}</div>
          <BaseDialog.Close aria-label="关闭" className="rounded p-1 hover:bg-surface-muted"><X className="size-4" /></BaseDialog.Close>
        </div>
        <div className="p-5">{children}</div>
        {footer && <div className="flex justify-end gap-2 border-t p-4">{footer}</div>}
      </BaseDialog.Popup>
    </BaseDialog.Portal>
  </BaseDialog.Root>
}

export function ConfirmAction({ open, onOpenChange, title, description, busy, onConfirm }: {
  open: boolean; onOpenChange: (open: boolean) => void; title: string; description: string
  busy?: boolean; onConfirm: () => void
}) {
  return <Dialog open={open} onOpenChange={onOpenChange} title={title} description={description} footer={<>
    <Button variant="secondary" onClick={() => onOpenChange(false)}>取消</Button>
    <Button variant="danger" disabled={busy} onClick={onConfirm}>{busy ? '处理中…' : '确认删除'}</Button>
  </>}><p className="text-sm text-muted-foreground">此操作会立即提交到服务器，请确认目标无误。</p></Dialog>
}
