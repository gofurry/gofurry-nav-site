import type { InsightDimensionBreakdown, InsightDomain } from '@/types/insights'

export function insightDimensionBars(breakdown: InsightDimensionBreakdown | null, domain: InsightDomain) {
  const maximum = domain === 'site' ? 1 : Math.max(1, ...(breakdown?.items ?? []).map(item => item.population))
  // Keep the authoritative backend order. Population bars are counts, not shares.
  return (breakdown?.items ?? []).slice(0, 8).map(item => {
    const value = domain === 'site' ? item.metric_value : item.population
    return { item, value, maximum, signal: value === null ? null : Math.max(0, Math.min(maximum, value)) }
  })
}
