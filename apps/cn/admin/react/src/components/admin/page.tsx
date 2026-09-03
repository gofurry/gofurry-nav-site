import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

export function PageHeader({ title, actions }: { title: string; description?: string; actions?: ReactNode; eyebrow?: string }) {
  return <header data-admin-page-header className="-mb-2 flex min-h-9 items-center justify-between gap-6">
    <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
    {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
  </header>
}

export function Section({ title, description, actions, children, className }: { title?: string; description?: string; actions?: ReactNode; children: ReactNode; className?: string }) {
  return <section className={cn('rounded-lg border bg-surface', className)}>
    {(title || description || actions) && <div className="flex items-start justify-between gap-4 border-b px-5 py-4"><div>{title && <h2 className="font-semibold">{title}</h2>}{description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}</div>{actions}</div>}
    <div className="p-5">{children}</div>
  </section>
}

export function FormField({ label, help, error, required, children }: { label: string; help?: string; error?: string; required?: boolean; children: ReactNode }) {
  return <label className="grid gap-1.5 text-sm"><span className="font-medium">{label}{required && <span className="ml-1 text-danger">*</span>}</span>{children}{help && !error && <span className="text-xs text-muted-foreground">{help}</span>}{error && <span className="text-xs text-danger">{error}</span>}</label>
}

export function FormSection({ title, description, children }: { title: string; description?: string; children: ReactNode }) {
  return <fieldset className="grid gap-4 border-0 p-0"><legend className="mb-1 w-full border-b pb-2 text-sm font-semibold">{title}</legend>{description && <p className="-mt-2 text-xs text-muted-foreground">{description}</p>}{children}</fieldset>
}

export function DetailGrid({ children }: { children: ReactNode }) { return <dl className="grid gap-x-8 gap-y-4 md:grid-cols-2 xl:grid-cols-3">{children}</dl> }
export function Detail({ label, children, technical }: { label: string; children: ReactNode; technical?: string }) {
  return <div><dt className="text-xs text-muted-foreground">{label}</dt><dd className="mt-1 text-sm font-medium">{children}</dd>{technical && <dd className="mt-0.5 font-mono text-[11px] text-muted-foreground">{technical}</dd>}</div>
}
