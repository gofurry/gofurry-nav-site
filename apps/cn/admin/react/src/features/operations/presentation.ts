import type { CollectionRun } from './types'

export type SemanticTone = 'success' | 'warning' | 'danger' | 'info' | 'neutral'

const stateLabels: Record<string, string> = {
  positive: '支持', negative: '不支持', stale: '数据过期', not_probed: '未探测',
  probe_failed: '探测失败', unknown: '未知', not_applicable: '不适用',
}

const eventLabels: Record<string, string> = {
  ipv6_enabled: '启用了 IPv6', ipv6_disabled: '停用了 IPv6', tls13_enabled: '启用了 TLS 1.3', tls13_disabled: '停用了 TLS 1.3',
  security_txt_enabled: '提供了 security.txt', security_txt_disabled: '不再提供 security.txt', primary_target_changed: '变更了主采集目标',
  tls_certificate_changed: '更新了 TLS 证书', game_became_free: '由付费变为免费', game_became_paid: '由免费变为付费',
  windows_support_enabled: '新增 Windows 支持', windows_support_disabled: '取消 Windows 支持', linux_support_enabled: '新增 Linux 支持',
  linux_support_disabled: '取消 Linux 支持', game_released: '已发售', game_price_changed: '调整了价格',
}

export function metricStateLabel(state: string) { return stateLabels[state] ?? state }
export function eventLabel(code: string) { return eventLabels[code] ?? code.replaceAll('_', ' ') }
export function eventSentence(name: string, code: string) { return `${name || '未知对象'} ${eventLabel(code)}` }
export function percent(value: number | null | undefined) { return value == null ? '—' : `${(value * 100).toFixed(1)}%` }
export function duration(milliseconds: number) { return milliseconds >= 1000 ? `${(milliseconds / 1000).toFixed(1)} s` : `${milliseconds} ms` }
export function bytes(value: number) { if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(2)} GiB`; if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MiB`; if (value >= 1024) return `${(value / 1024).toFixed(1)} KiB`; return `${value} B` }
export function statusTone(status: string): SemanticTone {
  if (['success', 'healthy', 'current', 'completed', 'active', 'positive', 'enabled'].includes(status)) return 'success'
  if (['failed', 'failure', 'unavailable', 'danger', 'disabled', 'probe_failed'].includes(status)) return 'danger'
  if (['warning', 'degraded', 'behind', 'ahead', 'missed', 'stale', 'unknown', 'partial'].includes(status)) return 'warning'
  if (['running', 'queued', 'info'].includes(status)) return 'info'
  return 'neutral'
}
export function runCoverage(run: Pick<CollectionRun, 'expected_count' | 'success_count'>) { return run.expected_count === 0 ? null : run.success_count / run.expected_count }
export function safeJSON(value: unknown) { try { return JSON.stringify(value, null, 2) } catch { return '{}' } }
