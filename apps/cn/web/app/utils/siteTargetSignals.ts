import { siteProtocolTone, type SiteDetailTone } from './siteDetailPresentation'

type Translate = (key: string, values?: Record<string, string | number>) => string
type EdgeHint = { provider: string; type: string; confidence: string }
const cdnProviders = new Set(['cloudflare', 'tencent_cloud', 'aliyun', 'aws_cloudfront', 'fastly'])

/** Display protocol status, not Target/Site health or latency classification. */
export function presentSiteProtocolStatus(value: unknown, stale: boolean | null | undefined, t: Translate) {
  const status = stale ? 'stale' : typeof value === 'string' && ['success', 'failure', 'skipped', 'healthy', 'warning', 'degraded', 'down', 'stale'].includes(value) ? value : 'unknown'
  const tone: SiteDetailTone = ['unknown', 'skipped'].includes(status) ? 'muted' : siteProtocolTone(status)
  return { status, statusLabel: t('siteDetail.states.' + status), tone, success: status === 'success' }
}

/** Only typed, reliable CDN hints qualify; provider identity cannot imply a CDN. */
export function presentSiteCdn(hints: readonly EdgeHint[], t: Translate) {
  const reliable = hints.filter(hint => hint.type === 'cdn' && ['high', 'medium'].includes(hint.confidence) && hint.provider.trim())
  const selected = reliable.find(hint => hint.confidence === 'high') ?? reliable[0]
  if (!selected) return null
  const provider = selected.provider.trim()
  return { provider, label: cdnProviders.has(provider) ? t('siteDetail.cdn.' + provider) : t('siteDetail.cdn.other', { provider }) }
}
