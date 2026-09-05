import { Popover } from '@base-ui/react/popover'
import { Minus, Plus } from '@phosphor-icons/react'
import { CalendarDays, ChevronLeft, ChevronRight, Clock3 } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import { cn } from '../../lib/utils'
import { Button } from './button'
import { Select } from './select'

const weekdays = ['一', '二', '三', '四', '五', '六', '日']
const hourOptions = Array.from({ length: 24 }, (_, hour) => ({ value: String(hour).padStart(2, '0'), label: String(hour).padStart(2, '0') }))
const minuteOptions = Array.from({ length: 60 }, (_, minute) => ({ value: String(minute).padStart(2, '0'), label: String(minute).padStart(2, '0') }))

function datePart(value: string) {
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(value)
  return match ? match[0] : ''
}

function timePart(value: string) {
  const match = /T(\d{2}):(\d{2})/.exec(value.replace(' ', 'T'))
  return { hour: match?.[1] ?? '00', minute: match?.[2] ?? '00' }
}

function localDate(value = '') {
  const date = datePart(value)
  if (date) {
    const [year, month, day] = date.split('-').map(Number)
    return new Date(year, month - 1, day)
  }
  return new Date()
}

function dateValue(value: Date) {
  return [value.getFullYear(), String(value.getMonth() + 1).padStart(2, '0'), String(value.getDate()).padStart(2, '0')].join('-')
}

function dateTimeDraft(value: string, now: Date) {
  if (datePart(value)) {
    return { date: datePart(value), month: localDate(value), ...timePart(value) }
  }

  return {
    date: dateValue(now),
    month: now,
    hour: String(now.getHours()).padStart(2, '0'),
    minute: String(now.getMinutes()).padStart(2, '0'),
  }
}

function adjustTimePart(value: string, offset: number, maximum: number) {
  const adjusted = Math.min(maximum, Math.max(0, Number(value) + offset))
  return String(adjusted).padStart(2, '0')
}

export function calendarMonthDays(month: Date) {
  const first = new Date(month.getFullYear(), month.getMonth(), 1)
  const mondayOffset = (first.getDay() + 6) % 7
  const start = new Date(first)
  start.setDate(first.getDate() - mondayOffset)
  return Array.from({ length: 42 }, (_, index) => {
    const day = new Date(start)
    day.setDate(start.getDate() + index)
    return day
  })
}

function CalendarGrid({ month, selected, onMonthChange, onSelect }: { month: Date; selected: string; onMonthChange: (month: Date) => void; onSelect: (value: string) => void }) {
  const days = calendarMonthDays(month)
  const move = (offset: number) => onMonthChange(new Date(month.getFullYear(), month.getMonth() + offset, 1))
  return <div className="w-72 p-3">
    <div className="mb-3 flex items-center justify-between"><Button type="button" variant="ghost" size="icon" aria-label="上个月" onClick={() => move(-1)}><ChevronLeft className="size-4" /></Button><p className="text-sm font-medium">{month.getFullYear()} 年 {month.getMonth() + 1} 月</p><Button type="button" variant="ghost" size="icon" aria-label="下个月" onClick={() => move(1)}><ChevronRight className="size-4" /></Button></div>
    <div className="grid grid-cols-7 gap-1">{weekdays.map((day) => <span key={day} className="grid h-8 place-items-center text-xs text-muted-foreground">{day}</span>)}{days.map((day) => {
      const value = dateValue(day)
      return <button key={value} type="button" onClick={() => onSelect(value)} className={cn('grid h-8 place-items-center rounded text-sm hover:bg-surface-muted', day.getMonth() !== month.getMonth() && 'text-muted-foreground/55', selected === value && 'bg-primary font-medium text-primary-foreground hover:bg-primary')}>{day.getDate()}</button>
    })}</div>
  </div>
}

function PickerTrigger({ value, placeholder, icon, ariaLabel, disabled }: { value: string; placeholder: string; icon: ReactNode; ariaLabel: string; disabled?: boolean }) {
  return <Popover.Trigger disabled={disabled} aria-label={ariaLabel} className="flex h-9 w-full items-center gap-2 rounded-md border bg-surface px-3 text-left text-sm outline-none focus:ring-2 focus:ring-ring disabled:opacity-60"><span className="text-muted-foreground">{icon}</span><span className={cn('truncate', !value && 'text-muted-foreground')}>{value || placeholder}</span></Popover.Trigger>
}

export function DatePicker({ value, onValueChange, placeholder = '选择日期', disabled, ariaLabel = '选择日期' }: { value: string; onValueChange: (value: string) => void; placeholder?: string; disabled?: boolean; ariaLabel?: string }) {
  const [open, setOpen] = useState(false)
  const [month, setMonth] = useState(() => localDate(value))
  const selected = datePart(value)
  return <Popover.Root open={open} onOpenChange={(next) => { setOpen(next); if (next) setMonth(localDate(value)) }}>
    <PickerTrigger value={selected} placeholder={placeholder} icon={<CalendarDays className="size-4" />} ariaLabel={ariaLabel} disabled={disabled} />
    {disabled ? null : <Popover.Portal><Popover.Positioner className="z-[90]" sideOffset={4} align="start"><Popover.Popup className="rounded-lg border bg-surface shadow-xl outline-none"><CalendarGrid month={month} selected={selected} onMonthChange={setMonth} onSelect={(next) => { onValueChange(next); setOpen(false) }} />{selected && <div className="flex justify-end border-t p-2"><Button type="button" variant="ghost" size="sm" onClick={() => { onValueChange(''); setOpen(false) }}>清除</Button></div>}</Popover.Popup></Popover.Positioner></Popover.Portal>}
  </Popover.Root>
}

export function DateTimePicker({ value, onValueChange, placeholder = '选择日期和时间', disabled, ariaLabel = '选择日期和时间', now = () => new Date() }: { value: string; onValueChange: (value: string) => void; placeholder?: string; disabled?: boolean; ariaLabel?: string; now?: () => Date }) {
  const initialDraft = dateTimeDraft(value, now())
  const [open, setOpen] = useState(false)
  const [month, setMonth] = useState(initialDraft.month)
  const [draftDate, setDraftDate] = useState(initialDraft.date)
  const [hour, setHour] = useState(initialDraft.hour)
  const [minute, setMinute] = useState(initialDraft.minute)
  const display = value ? `${datePart(value)} ${timePart(value).hour}:${timePart(value).minute}` : ''
  const openPicker = (next: boolean) => {
    setOpen(next)
    if (next) {
      const draft = dateTimeDraft(value, now())
      setMonth(draft.month)
      setDraftDate(draft.date)
      setHour(draft.hour)
      setMinute(draft.minute)
    }
  }
  return <Popover.Root open={open} onOpenChange={openPicker}>
    <PickerTrigger value={display} placeholder={placeholder} icon={<Clock3 className="size-4" />} ariaLabel={ariaLabel} disabled={disabled} />
    {disabled ? null : <Popover.Portal><Popover.Positioner className="z-[90]" sideOffset={4} align="start"><Popover.Popup className="rounded-lg border bg-surface shadow-xl outline-none">
      <CalendarGrid month={month} selected={draftDate} onMonthChange={setMonth} onSelect={setDraftDate} />
      <div className="flex items-end gap-2 border-t p-3"><div className="grid min-w-0 flex-1 gap-1 text-xs text-muted-foreground"><span>小时</span><span className="flex items-center gap-1"><Button type="button" variant="ghost" size="icon" className="size-8 shrink-0" aria-label="小时减一" onClick={() => setHour((current) => adjustTimePart(current, -1, 23))}><Minus className="size-3.5" aria-hidden="true" /></Button><span className="min-w-0 flex-1"><Select value={hour} onValueChange={setHour} options={hourOptions} ariaLabel="小时" /></span><Button type="button" variant="ghost" size="icon" className="size-8 shrink-0" aria-label="小时加一" onClick={() => setHour((current) => adjustTimePart(current, 1, 23))}><Plus className="size-3.5" aria-hidden="true" /></Button></span></div><span className="pb-2">:</span><div className="grid min-w-0 flex-1 gap-1 text-xs text-muted-foreground"><span>分钟</span><span className="flex items-center gap-1"><Button type="button" variant="ghost" size="icon" className="size-8 shrink-0" aria-label="分钟减一" onClick={() => setMinute((current) => adjustTimePart(current, -1, 59))}><Minus className="size-3.5" aria-hidden="true" /></Button><span className="min-w-0 flex-1"><Select value={minute} onValueChange={setMinute} options={minuteOptions} ariaLabel="分钟" /></span><Button type="button" variant="ghost" size="icon" className="size-8 shrink-0" aria-label="分钟加一" onClick={() => setMinute((current) => adjustTimePart(current, 1, 59))}><Plus className="size-3.5" aria-hidden="true" /></Button></span></div></div>
      <div className="flex justify-between border-t p-2"><Button type="button" variant="ghost" size="sm" onClick={() => { onValueChange(''); setOpen(false) }}>清除</Button><Button type="button" size="sm" disabled={!draftDate} onClick={() => { onValueChange(`${draftDate}T${hour}:${minute}`); setOpen(false) }}>确定</Button></div>
    </Popover.Popup></Popover.Positioner></Popover.Portal>}
  </Popover.Root>
}
