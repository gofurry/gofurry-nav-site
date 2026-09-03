import { Select as BaseSelect } from '@base-ui/react/select'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from '../../lib/utils'

export type SelectOption = { label: string; value: string }

export function Select({ value, onValueChange, options, placeholder = '请选择', disabled, name, className, ariaLabel }: {
  value?: string; onValueChange: (value: string) => void; options: SelectOption[]
  placeholder?: string; disabled?: boolean; name?: string; className?: string; ariaLabel?: string
}) {
  return <BaseSelect.Root value={value ?? ''} onValueChange={(next) => next !== null && onValueChange(next)} disabled={disabled} name={name}>
    <BaseSelect.Trigger aria-label={ariaLabel} className={cn('flex h-9 w-full items-center justify-between gap-2 rounded-md border bg-surface px-3 text-left text-sm outline-none focus:ring-2 focus:ring-ring disabled:opacity-60', className)}>
      <BaseSelect.Value placeholder={placeholder}>{(selected: string) => options.find((item) => item.value === selected)?.label ?? placeholder}</BaseSelect.Value>
      <BaseSelect.Icon><ChevronDown className="size-4 text-muted-foreground" /></BaseSelect.Icon>
    </BaseSelect.Trigger>
    <BaseSelect.Portal>
      <BaseSelect.Positioner className="z-[80] outline-none" sideOffset={4}>
        <BaseSelect.Popup className="max-h-72 min-w-[var(--anchor-width)] overflow-auto rounded-md border bg-surface p-1 shadow-xl">
          {options.map((option) => <BaseSelect.Item key={option.value} value={option.value} className="grid cursor-default grid-cols-[1rem_1fr] items-center gap-2 rounded px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-surface-muted">
            <BaseSelect.ItemIndicator><Check className="size-4 text-primary" /></BaseSelect.ItemIndicator>
            <BaseSelect.ItemText>{option.label}</BaseSelect.ItemText>
          </BaseSelect.Item>)}
        </BaseSelect.Popup>
      </BaseSelect.Positioner>
    </BaseSelect.Portal>
  </BaseSelect.Root>
}
