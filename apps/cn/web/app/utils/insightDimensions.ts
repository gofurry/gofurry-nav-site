import type { InsightDimension } from '@/types/insights'

export function normalizeInsightSlice(dimension: InsightDimension, value: string): string | null {
  const normalized = value.trim()
  if (!normalized) return null
  if (dimension === 'country') {
    if (normalized === 'unknown') return normalized
    return /^[a-z]{2}$/i.test(normalized) ? normalized.toUpperCase() : null
  }
  if (dimension === 'nsfw') return ['nsfw', 'sfw', 'unknown'].includes(normalized) ? normalized : null
  if (dimension === 'public_interest') return ['public_interest', 'standard', 'unknown'].includes(normalized) ? normalized : null
  return normalized === 'unknown' || /^[1-9]\d*$/.test(normalized) ? normalized : null
}

export function formatInsightRatio(value: number | null) {
  return value === null ? '—' : `${(value * 100).toFixed(1)}%`
}
