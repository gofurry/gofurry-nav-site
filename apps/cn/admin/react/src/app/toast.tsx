import { CheckCircle2, CircleAlert, Info, X } from 'lucide-react'
import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { cn } from '../lib/utils'

type ToastTone = 'success' | 'danger' | 'info'
type ToastItem = { id: number; message: string; tone: ToastTone }
const ToastContext = createContext<{ toast: (message: string, tone?: ToastTone) => void } | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([])
  const toast = useCallback((message: string, tone: ToastTone = 'success') => {
    const id = Date.now() + Math.random()
    setItems((current) => [...current, { id, message, tone }])
    window.setTimeout(() => setItems((current) => current.filter((item) => item.id !== id)), 4200)
  }, [])
  const value = useMemo(() => ({ toast }), [toast])
  return <ToastContext.Provider value={value}>
    {children}
    <div className="fixed right-5 top-5 z-[100] grid w-80 gap-2" aria-live="polite">
      {items.map((item) => {
        const Icon = item.tone === 'success' ? CheckCircle2 : item.tone === 'danger' ? CircleAlert : Info
        return <div key={item.id} className="flex items-start gap-3 rounded-md border bg-surface p-3 shadow-lg">
          <Icon className={cn('mt-0.5 size-4 shrink-0', item.tone === 'success' && 'text-success', item.tone === 'danger' && 'text-danger', item.tone === 'info' && 'text-info')} />
          <p className="min-w-0 flex-1 text-sm">{item.message}</p>
          <button aria-label="关闭通知" onClick={() => setItems((current) => current.filter((entry) => entry.id !== item.id))}><X className="size-4 text-muted-foreground" /></button>
        </div>
      })}
    </div>
  </ToastContext.Provider>
}

export function useToast() {
  const context = useContext(ToastContext)
  if (!context) throw new Error('useToast must be used within ToastProvider')
  return context
}
