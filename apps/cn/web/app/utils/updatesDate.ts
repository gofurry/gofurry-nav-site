// Release dates and calendar groups use the China site's zone on SSR and client.
const timeZone = 'Asia/Shanghai'

export function parseUpdatesDate(value: string): Date | null {
  if (!value) return null
  const normalized = value.trim().replace(' ', 'T')
  const date = new Date(/(?:Z|[+-]\d{2}:?\d{2})$/i.test(normalized) ? normalized : `${normalized}+08:00`)
  return Number.isNaN(date.getTime()) ? null : date
}

export function updatesMonth(value: string): string {
  const date = parseUpdatesDate(value)
  if (!date) return '—'
  const parts = new Intl.DateTimeFormat('en', { timeZone, year: 'numeric', month: '2-digit' }).formatToParts(date)
  return `${parts.find(part => part.type === 'year')!.value} / ${parts.find(part => part.type === 'month')!.value}`
}

export function formatUpdatesFullDate(value: string, localeCode: string, unavailable = '—'): string {
  const date = parseUpdatesDate(value)
  if (!date) return unavailable
  return new Intl.DateTimeFormat(localeCode, {
    timeZone, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23',
  }).format(date) + ' UTC+8'
}

export function formatUpdatesDate(value: string, localeCode: string): string {
  const date = parseUpdatesDate(value)
  return date ? new Intl.DateTimeFormat(localeCode, { timeZone, year: 'numeric', month: '2-digit', day: '2-digit' }).format(date) : '—'
}
