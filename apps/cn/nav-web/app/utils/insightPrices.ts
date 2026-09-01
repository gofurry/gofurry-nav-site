import type { GameInsightPrice, GameInsightPricePoint } from '@/types/insights'

export type PublicPriceDisplay =
  | { kind: 'free', amount: null }
  | { kind: 'priced', amount: number }
  | { kind: 'unavailable', amount: null }

export function publicPriceDisplay(price: GameInsightPrice | GameInsightPricePoint | null): PublicPriceDisplay {
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
  return new Intl.NumberFormat(locale === 'en' ? 'en-US' : 'zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount / 100)
}
