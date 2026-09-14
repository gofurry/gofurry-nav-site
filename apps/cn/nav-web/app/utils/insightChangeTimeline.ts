import type { InsightExplorerChange } from '@/types/insights'

// Only group adjacent dates: the API's cursor order and repeated entity events
// remain authoritative, including when another page extends the last date.
export function groupInsightChangeDates(items: readonly InsightExplorerChange[]) {
  const groups: { date: string; items: InsightExplorerChange[] }[] = []
  for (const item of items) {
    const previous = groups.at(-1)
    if (previous?.date === item.date) previous.items.push(item)
    else groups.push({ date: item.date, items: [item] })
  }
  return groups
}

export function formatInsightChangeDate(value: string, locale: string) {
  const date = new Date(`${value}T00:00:00Z`)
  if (!Number.isFinite(date.getTime())) return value
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-US' : 'zh-CN', {
    year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC',
  }).format(date)
}
