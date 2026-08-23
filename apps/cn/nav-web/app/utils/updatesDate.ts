export function parseUpdatesDate(value: string) {
  if (!value) {
    return null
  }

  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const parsed = new Date(normalized)
  if (Number.isNaN(parsed.getTime())) {
    return null
  }

  return parsed
}

export function formatUpdatesMonthDay(value: string, localeCode: string) {
  const date = parseUpdatesDate(value)
  if (!date) {
    return '--.--'
  }

  return new Intl.DateTimeFormat(localeCode, {
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

export function formatUpdatesClock(value: string, localeCode: string) {
  const date = parseUpdatesDate(value)
  if (!date) {
    return '--:--'
  }

  return new Intl.DateTimeFormat(localeCode, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

export function formatUpdatesYear(value: string) {
  const date = parseUpdatesDate(value)
  if (!date) {
    return '----'
  }

  return String(date.getFullYear())
}

export function formatUpdatesFullDate(value: string, localeCode: string, unavailable: string) {
  const date = parseUpdatesDate(value)
  if (!date) {
    return value || unavailable
  }

  return new Intl.DateTimeFormat(localeCode, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}
