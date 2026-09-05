import type { GameDetailInsightRange } from '@/types/insights'

export const gameDetailInsightRanges: GameDetailInsightRange[] = ['30d', '90d', '180d', '1y', '3y', '5y']

export function formatGameInsightAxisDate(value: string, range: GameDetailInsightRange) {
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(value)
  if (!match) return value

  const [, year, month, day] = match
  if (range === '30d' || range === '90d' || range === '180d') {
    return `${month}-${day}`
  }
  return `${year}-${month}`
}
