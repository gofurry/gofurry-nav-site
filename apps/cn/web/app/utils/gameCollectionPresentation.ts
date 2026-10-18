import type { GameCollectionChronology, GameCollectionTimelineItem } from '~/types/game'

type Translate = (key: string, values?: Record<string, string | number>) => string

// Date-only facts must never pass through the browser's local timezone.
export function collectionDate(date: string, locale: string, precision: GameCollectionChronology['precision'] = 'day') {
  const [year = 0, month = 1, day = 1] = date.split('-').map(Number)
  if (precision === 'year') return String(year)
  if (precision === 'quarter') return locale === 'en' ? `Q${Math.ceil(month / 3)} ${year}` : `${year} Q${Math.ceil(month / 3)}`
  if (locale !== 'en') return `${year}年${month}月${precision === 'day' ? `${day}日` : ''}`
  return new Intl.DateTimeFormat(locale === 'en' ? 'en-US' : 'zh-CN', {
    timeZone: 'UTC', year: 'numeric', month: locale === 'en' ? 'long' : 'numeric',
    ...(precision === 'day' ? { day: 'numeric' as const } : {}),
  }).format(new Date(Date.UTC(year, month - 1, day)))
}

export function collectionTime(item: GameCollectionTimelineItem, locale: string, t: Translate) {
  const key = 'game.collections.time.'
  if (item.phase === 'released_unknown' || item.phase === 'upcoming_tba' || item.phase === 'unknown') return t(key + item.phase)
  if (!item.chronology) return t(key + 'unknown')
  const fact = item.chronology
  let time = collectionDate(fact.window_start, locale, fact.precision)
  if (fact.inferred) time = t(key + 'inferred', { time })
  return item.phase === 'upcoming_overdue' ? t(key + 'overdue', { time }) : time
}

// Stable partition only: each phase's Backend chronology remains authoritative.
export function collectionTimelineSections(items: GameCollectionTimelineItem[]) {
  return [
    { key: 'past', snake: true, items: items.filter(item => item.phase === 'released') },
    { key: 'releasedUnknown', snake: false, items: items.filter(item => item.phase === 'released_unknown') },
    { key: 'future', snake: true, items: items.filter(item => item.phase === 'upcoming_overdue' || item.phase === 'upcoming') },
    { key: 'tba', snake: false, items: items.filter(item => item.phase === 'upcoming_tba') },
    { key: 'unknown', snake: false, items: items.filter(item => item.phase === 'unknown') },
  ]
}
