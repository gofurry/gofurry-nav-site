/** Display missing/non-finite evidence as a gap, never an invented zero. */
export function finiteChartValue(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}
export function formatChartPercent(value: unknown) {
  const number = finiteChartValue(value)
  return number === null ? '—' : `${number}%`
}
export function formatChartNumber(value: unknown, locale: string, maximumFractionDigits = 3) {
  const number = finiteChartValue(value)
  return number === null ? '—' : new Intl.NumberFormat(locale, { minimumFractionDigits: 0, maximumFractionDigits }).format(number)
}
export function formatGameAverage(value: unknown, locale: string) {
  return formatChartNumber(value, locale, 1)
}
