import type { SiteHealthSummary, TargetHealthSummary, TargetLatestResponse } from '~/types/nav'

type Tone = 'neutral' | 'good' | 'warning' | 'bad'
type Translate = (key: string) => string
interface TargetSource {
  domain: string
  siteHealthSummary: SiteHealthSummary | null
  targetHealthSummary: TargetHealthSummary | null
  targetLatestCore: TargetLatestResponse | null
}

const record = (value: unknown): Record<string, unknown> => value && typeof value === 'object' && !Array.isArray(value)
  ? value as Record<string, unknown> : {}
const text = (value: unknown) => typeof value === 'string' ? value.trim() : ''
const number = (value: unknown) => typeof value === 'number' && Number.isFinite(value) ? value : null
const duration = (value: number | null) => value === null ? '—' : `${Math.round(value)} ms`
const tone = (status: string): Tone => status === 'healthy' || status === 'success' ? 'good'
  : status === 'down' || status === 'failure' ? 'bad'
    : ['warning', 'degraded', 'stale'].includes(status) ? 'warning' : 'neutral'
const date = (...values: unknown[]) => {
  for (const value of values) {
    const raw = text(value)
    // Missing Go summaries serialize time.Time's zero value, not null.
    if (raw && raw !== '0001-01-01T00:00:00Z' && Number.isFinite(Date.parse(raw))) {
      return new Date(raw).toISOString().slice(0, 19).replace('T', ' ') + ' UTC'
    }
  }
  return '—'
}

/** One Current Target projection for the strip and both responsive contexts.
 * Site summary supplies the selector catalog/relations, never Target health.
 * Keep missing evidence nullable; legacy HttpRecord fills missing fields with zero.
 */
export function presentSiteTarget(source: TargetSource, t: Translate) {
  const target = source.domain
  const summary = source.targetHealthSummary?.target === target ? source.targetHealthSummary : null
  const latest = source.targetLatestCore?.target === target ? source.targetLatestCore : null
  const protocols = latest?.protocols ?? {}
  const http = protocols.http?.target && protocols.http.target !== target ? undefined : protocols.http
  const payload = record(http?.payload)
  const status = summary?.state === 'stale' ? 'stale' : summary?.state === 'ready' ? summary.status : 'unknown'
  const stateLabel = (value: string) => t(`siteDetail.states.${['healthy', 'warning', 'degraded', 'unknown', 'down', 'stale', 'success', 'failure', 'skipped'].includes(value) ? value : 'unknown'}`)
  const statusLabel = stateLabel(status)
  const latency = number(payload.response_time_ms) ?? number(http?.duration_ms)
  const httpStatus = number(payload.status_code)
  const httpProtocol = text(payload.http_protocol)
  const tlsVersion = text(payload.tls_version)
  const certificateDays = number(payload.cert_days_left)
  const certificateVerified = typeof payload.cert_verified === 'boolean' ? payload.cert_verified : null
  const certificateLabel = certificateVerified === null ? t('siteDetail.notObserved')
    : t(certificateVerified ? 'siteDetail.verified' : 'siteDetail.notVerified')
  const observedAt = date(summary?.observed_at, http?.observed_at)
  const protocolStates = ['ping', 'http', 'dns'].map(protocol => {
    const current = summary?.protocols?.[protocol]
    const envelope = protocols[protocol]
    const matching = envelope?.target && envelope.target !== target ? undefined : envelope
    const state = current?.stale ? 'stale' : current?.status || matching?.status || 'unknown'
    return { protocol, label: protocol === 'ping' ? 'Ping' : protocol.toUpperCase(), status: state,
      statusLabel: stateLabel(state), tone: tone(state), duration: duration(number(current?.duration_ms) ?? number(matching?.duration_ms)),
      observedAt: date(current?.observed_at, matching?.observed_at) }
  })
  const relation = (host: string) => {
    const item = (source.siteHealthSummary?.targets ?? []).find(item => item.target === host)
    const hints = host === target ? summary?.target_relation_hints ?? item?.target_relation_hints : item?.target_relation_hints
    const canonical = host === target ? summary?.canonical_target_hint ?? item?.canonical_target_hint : item?.canonical_target_hint
    const relations = (hints ?? []).map(hint => [hint.relation, hint.related_host || hint.value].filter(Boolean).join(' · '))
    if (canonical) relations.push([canonical.relation, canonical.preferred_host || canonical.canonical_host || canonical.final_host].filter(Boolean).join(' · '))
    for (const hint of source.siteHealthSummary?.target_relation_hints ?? []) {
      if ((hint.targets ?? []).includes(host)) relations.push(`${hint.relation} · ${hint.host}`)
    }
    return [...new Set(relations)].join('; ')
  }
  const targets = [...new Set([target, ...(source.siteHealthSummary?.targets ?? []).map(item => item.target)].filter(Boolean))]
  const edgeProviderHints = (summary?.edge_provider_hints ?? []).map(hint => ({
    provider: hint.provider, type: hint.hint_type, confidence: hint.confidence,
    evidence: (hint.evidence ?? []).map(item => `${item.source}: ${item.field} = ${item.value}`).join('; '),
  }))
  const finalUrl = text(payload.final_url)
  const visitUrl = /^https?:\/\//i.test(finalUrl) ? finalUrl : target ? `https://${target}` : ''
  const health = [
    { key: 'status', label: t('siteDetail.status'), value: statusLabel, detail: '', tone: tone(status) },
    { key: 'latency', label: t('siteDetail.latency'), value: duration(latency), detail: '', tone: 'neutral' as Tone },
    { key: 'http', label: 'HTTP', value: httpStatus === null ? '—' : `HTTP ${httpStatus}`, detail: httpProtocol, tone: httpStatus === null ? 'neutral' as Tone : httpStatus >= 400 ? 'warning' as Tone : 'neutral' as Tone },
    { key: 'tls', label: 'TLS', value: tlsVersion || '—', detail: '', tone: 'neutral' as Tone },
    { key: 'certificate', label: t('siteDetail.certificate'), value: certificateDays === null ? '—' : `${certificateDays} ${t('siteDetail.days')}`, detail: certificateLabel,
      tone: certificateVerified === false || (certificateDays !== null && certificateDays <= 30) ? 'warning' as Tone : 'neutral' as Tone },
    { key: 'observed', label: t('siteDetail.observed'), value: observedAt, detail: '', tone: 'neutral' as Tone },
  ]
  return { target, status, statusLabel, tone: tone(status), latency, httpStatus, httpProtocol, tlsVersion,
    certificateDays, certificateVerified, certificateLabel, observedAt, protocolStates, edgeProviderHints,
    targetList: targets.map(host => ({ target: host, relation: relation(host) })), targetRelation: relation(target), health, visitUrl }
}

export type SiteTargetPresentation = ReturnType<typeof presentSiteTarget>
