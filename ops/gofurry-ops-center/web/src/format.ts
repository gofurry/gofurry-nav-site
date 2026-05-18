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
