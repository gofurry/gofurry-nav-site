import type { GameShowcaseAction, GameShowcaseItem, GameShowcaseRelease, GameShowcaseSecondaryAction, GameShowcaseSnapshot } from '../types/game'

export function emptyGameShowcase(): GameShowcaseSnapshot {
  return { schema_version: 1, snapshot_id: '', generated_at: '', valid_until: '', items: [] }
}

export function showcaseContextKey(item: GameShowcaseItem) {
  if (item.sponsored || item.reason === 'sponsored') return `gameShowcase.context.sponsored.${item.content_type}`
  const reason = item.reason === 'new_release' ? 'newRelease' : item.reason
  return `gameShowcase.context.${reason}`
}

export function showcaseDestination(action: GameShowcaseAction | GameShowcaseSecondaryAction | undefined) {
  if (!action) return null
  if (action.type === 'game') {
    return action.game_id ? { path: `/games/${encodeURIComponent(action.game_id)}`, external: false } : null
  }
  // Stage A validates HTTPS. Also fail closed at the link boundary for malformed data.
  try {
    const url = new URL(action.target || '')
    return url.protocol === 'https:' && !url.username && !url.password
      ? { path: action.target!, external: true } : null
  } catch { return null }
}

export function showcaseFocalPoint(value: number | undefined) {
  return Number.isFinite(value) ? Math.min(1, Math.max(0, value!)) * 100 : 50
}

// Canonical calendar precision only; never infer from prose or the browser clock.
export function showcaseReleaseDate(release: GameShowcaseRelease | undefined, locale: string): string {
  if (!release || !['available', 'upcoming'].includes(release.availability)) return ''
  const en = locale === 'en'
  if (release.precision === 'day' && release.exact_date && /^\d{4}-\d{2}-\d{2}$/.test(release.exact_date)) {
    const date = new Date(`${release.exact_date}T00:00:00Z`)
    if (!Number.isFinite(date.getTime()) || date.toISOString().slice(0, 10) !== release.exact_date) return ''
    return new Intl.DateTimeFormat(en ? 'en-US' : 'zh-CN', { timeZone: 'UTC', year: 'numeric', month: 'short', day: 'numeric' }).format(date)
  }
  const year = release.year
  if (!Number.isInteger(year) || year! < 1) return ''
  if (release.precision === 'quarter' && Number.isInteger(release.quarter) && release.quarter! >= 1 && release.quarter! <= 4) {
    return en ? `Q${release.quarter} ${year}` : `${year} 年第 ${release.quarter} 季度`
  }
  if (release.precision === 'month' && Number.isInteger(release.month) && release.month! >= 1 && release.month! <= 12) {
    return new Intl.DateTimeFormat(en ? 'en-US' : 'zh-CN', { timeZone: 'UTC', year: 'numeric', month: 'short' })
      .format(new Date(`${String(year).padStart(4, '0')}-${String(release.month).padStart(2, '0')}-01T00:00:00Z`))
  }
  return release.precision === 'year' ? (en ? String(year) : `${year} 年`) : ''
}
