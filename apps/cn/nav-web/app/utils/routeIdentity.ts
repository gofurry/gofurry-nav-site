export function parsePositiveEntityRouteId(value: unknown): string | null {
  if (typeof value !== 'string' && typeof value !== 'number') {
    return null
  }

  const id = String(value).trim()
  return /^[1-9]\d*$/.test(id) ? id : null
}
