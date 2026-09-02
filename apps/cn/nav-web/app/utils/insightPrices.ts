import type { GameInsightPrice, GameInsightPricePoint, GameInsightRegionalPrice } from '@/types/insights'

export type PublicPriceDisplay =
  | { kind: 'free', amount: null }
  | { kind: 'priced', amount: number }
  | { kind: 'unavailable', amount: null }

export function publicPriceDisplay(price: GameInsightPrice | GameInsightPricePoint | GameInsightRegionalPrice | null): PublicPriceDisplay {
  if (!price || price.state === 'unknown' || price.state === 'unpriced') {
    return { kind: 'unavailable', amount: null }
  }
  if (price.state === 'free') {
    return { kind: 'free', amount: null }
  }
  if (price.state === 'priced' && price.final_amount !== null && Number.isFinite(price.final_amount)) {
    return { kind: 'priced', amount: price.final_amount }
  }
  return { kind: 'unavailable', amount: null }
}

export function formatCnyMinorAmount(amount: number, locale: string) {
  return formatMinorAmount(amount, 'CNY', locale)
}

export function formatMinorAmount(amount: number, currency: string, locale: string) {
  return new Intl.NumberFormat(locale === 'en' ? 'en-US' : 'zh-CN', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount / 100)
}

export function priceSegmentKey(point: GameInsightPricePoint) {
  if (point.state === 'priced' && point.currency) return `priced:${point.currency}`
  if (point.state === 'free') return 'free'
  return null
}
