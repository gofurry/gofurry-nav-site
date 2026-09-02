export function parseInsightCompareIDs(value: unknown): number[] | null {
  if (value === undefined || value === null || value === '') return []
  if (typeof value !== 'string') return null
  const result: number[] = []
  const seen = new Set<number>()
  for (const part of value.split(',')) {
    const normalized = part.trim()
    if (!/^[1-9]\d*$/.test(normalized)) return null
    const id = Number(normalized)
    if (!Number.isSafeInteger(id)) return null
    if (!seen.has(id)) {
      seen.add(id)
      result.push(id)
    }
  }
  return result.length <= 4 ? result : null
}

export function insightCompareReady(ids: number[]) {
  return ids.length >= 2 && ids.length <= 4
}
