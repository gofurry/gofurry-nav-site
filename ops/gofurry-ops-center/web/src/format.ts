export function formatTime(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(date)
}

export function statusClass(status?: string) {
  switch ((status || '').toLowerCase()) {
    case 'ok':
    case 'up':
    case 'success':
    case 'healthy':
      return 'badge badge-ok'
    case 'warning':
    case 'degraded':
    case 'pending':
      return 'badge badge-warn'
    case 'down':
    case 'critical':
    case 'failed':
    case 'error':
      return 'badge badge-bad'
    default:
      return 'badge badge-muted'
  }
}

export function alertClass(level?: string) {
  switch ((level || '').toLowerCase()) {
    case 'critical':
      return 'badge badge-bad'
    case 'warning':
      return 'badge badge-warn'
    default:
      return 'badge badge-muted'
  }
}

export function booleanText(value?: boolean) {
  if (value === undefined) return '-'
  return value ? '通过' : '失败'
}

export function formatPercent(value?: number, digits = 1) {
  if (value === undefined || value === null || Number.isNaN(value)) return '-'
  return `${value.toFixed(digits)}%`
}

export function formatNumber(value?: number, digits = 1) {
  if (value === undefined || value === null || Number.isNaN(value)) return '-'
  return value.toFixed(digits)
}

export function formatBytes(value?: number) {
  if (value === undefined || value === null || Number.isNaN(value)) return '-'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let next = Math.max(0, value)
  let index = 0
  while (next >= 1024 && index < units.length - 1) {
    next /= 1024
    index += 1
  }
  return `${next.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

export function formatRate(value?: number) {
  if (value === undefined || value === null || Number.isNaN(value)) return '-'
  return `${formatBytes(value)}/s`
}

export function formatDurationSeconds(value?: number) {
  if (value === undefined || value === null || Number.isNaN(value)) return '-'
  if (value < 60) return `${Math.max(0, Math.round(value))}s`
  const minutes = Math.floor(value / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  if (hours < 48) return `${hours}h`
  return `${Math.floor(hours / 24)}d`
}
