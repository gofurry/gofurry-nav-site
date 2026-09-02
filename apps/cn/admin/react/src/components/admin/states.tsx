import { CircleAlert, Inbox, LoaderCircle, RotateCw } from 'lucide-react'
import { Button } from '../ui/button'

export function LoadingState({ label = '正在加载…' }: { label?: string }) {
  return <div className="grid min-h-48 place-items-center rounded-md border bg-surface"><div className="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle className="size-4 animate-spin" />{label}</div></div>
}

export function EmptyState({ title = '暂无数据', message = '当前条件下没有可显示的记录。' }: { title?: string; message?: string }) {
  return <div className="grid min-h-48 place-items-center rounded-md border bg-surface text-center"><div><Inbox className="mx-auto mb-3 size-8 text-muted-foreground" /><p className="font-medium">{title}</p><p className="mt-1 text-sm text-muted-foreground">{message}</p></div></div>
}

export function ErrorState({ title = '加载失败', message, onRetry }: { title?: string; message: string; onRetry?: () => void }) {
  return <div className="grid min-h-48 place-items-center rounded-md border border-danger/30 bg-surface p-6 text-center"><div><CircleAlert className="mx-auto mb-3 size-8 text-danger" /><p className="font-medium">{title}</p><p className="mt-1 max-w-xl text-sm text-muted-foreground">{message}</p>{onRetry && <Button className="mt-4" variant="secondary" onClick={onRetry}><RotateCw className="size-4" />重试</Button>}</div></div>
}
