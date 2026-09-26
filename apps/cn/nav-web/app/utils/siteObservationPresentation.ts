import type { CollectorEnvelope } from '~/types/nav'
import type { SiteDetailPageData } from '~/composables/useSiteDetailPage'

type Source = Pick<SiteDetailPageData, 'domain' | 'targetHealthSummary' | 'targetLatestCore' | 'lightProbeState'>
type Translate = (key: string) => string
export interface ObservationFact { key: string; label: string; value: string }
export const observationRecord = (value: unknown): Record<string, unknown> => value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : {}
const text = (value: unknown) => typeof value === 'string' ? value.trim() : ''
const strings = (value: unknown) => Array.isArray(value) ? value.map(text).filter(Boolean) : []
const records = (value: unknown) => Array.isArray(value) ? value.map(observationRecord).filter(row => Object.keys(row).length) : []
export const observationNumber = (value: unknown) => typeof value === 'number' && Number.isFinite(value) && value >= 0 ? value : null
const ms = (value: unknown) => observationNumber(value) === null ? '—' : `${value} ms`
// Collector loss_rate is already a percentage (0–100), including sub-one values.
const percent = (value: unknown) => observationNumber(value) === null ? '—' : `${Math.round((value as number) * 100) / 100}%`
export const observationTime = (value: unknown) => text(value) && text(value) !== '0001-01-01T00:00:00Z' && Number.isFinite(Date.parse(text(value)))
  ? new Date(text(value)).toISOString().slice(0, 19).replace('T', ' ') + ' UTC' : '—'
export const observationStatus = (value: unknown) => ['success', 'failure', 'skipped', 'healthy', 'warning', 'degraded', 'down', 'stale'].includes(text(value)) ? text(value) : 'unknown'

export function normalizeObservationHeaders(value: unknown) {
  const normalized = new Map<string, string[]>()
  for (const [key, raw] of Object.entries(observationRecord(value))) {
    const values = Array.isArray(raw) ? strings(raw) : text(raw) ? [text(raw)] : []
    const name = key.trim().toLowerCase()
    if (name && values.length) normalized.set(name, [...(normalized.get(name) ?? []), ...values])
  }
  return [...normalized].sort(([a], [b]) => a.localeCompare(b)).map(([key, values]) => ({ key, label: key, value: values.join('\n') }))
}

/** Current Target evidence only; no Site aggregate, I/O, risk inference or TLS interpretation. */
export function presentSiteObservation(source: Source, t: Translate) {
  const target = source.domain
  const summary = source.targetHealthSummary?.target === target ? source.targetHealthSummary : null
  const latest = source.targetLatestCore?.target === target ? source.targetLatestCore : null
  const light = source.lightProbeState?.target === target ? source.lightProbeState : null
  const matching = (envelope?: CollectorEnvelope) => envelope?.target && envelope.target !== target ? undefined : envelope
  const core = (protocol: string) => matching(latest?.protocols?.[protocol])
  const httpEnvelope = core('http'), dnsEnvelope = core('dns'), pingEnvelope = core('ping')
  const http = observationRecord(httpEnvelope?.payload), dns = observationRecord(dnsEnvelope?.payload), ping = observationRecord(pingEnvelope?.payload)
  const label = (key: string) => t('siteObservation.fields.' + key)
  const display = (value: unknown): string => typeof value === 'boolean' ? t('siteObservation.' + (value ? 'yes' : 'no'))
    : typeof value === 'number' && Number.isFinite(value) ? String(value) : Array.isArray(value) ? strings(value).join('\n') || '—' : text(value) || '—'
  const fact = (key: string, value: unknown): ObservationFact => ({ key, label: label(key), value: display(value) })
  const compact = (items: ObservationFact[]) => items.filter(item => item.value !== '—')
  const protocols = ['ping', 'http', 'dns'].map(protocol => {
    const state = summary?.protocols?.[protocol], envelope = core(protocol)
    const status = observationStatus(state?.status || envelope?.status)
    const stale = state?.stale ?? (latest?.state === 'stale' ? true : null)
    return { protocol, status, statusLabel: t('siteDetail.states.' + status), duration: ms(state?.duration_ms ?? envelope?.duration_ms),
      observed: observationTime(state?.observed_at || envelope?.observed_at),
      freshness: t('siteObservation.' + (stale === true ? 'stale' : stale === false ? 'fresh' : 'freshnessUnknown')) }
  })
  const status = observationStatus(summary?.state === 'missing' ? null : summary?.status)
  const messages = [...strings(summary?.reason_messages), ...strings(latest?.reason_messages)]
  const risks = [...new Set(messages.length ? messages : [...strings(summary?.reason_codes), ...strings(latest?.reason_codes)])]
  const headers = normalizeObservationHeaders(http.headers)
  const header = (key: string) => headers.find(item => item.key === key)?.value
  const response = observationNumber(http.response_time_ms) ?? observationNumber(httpEnvelope?.duration_ms)
  const endpoint = compact([fact('finalUrl', http.final_url), fact('protocol', http.http_protocol), fact('remoteIp', http.remote_ip),
    fact('resolvedIp', ping.resolved_ip), fact('contentType', http.content_type || header('content-type'))])
  // Independent measured stages share one scale; they are not summed into a fabricated total.
  const timingValues = [http.dns_lookup_ms, http.tcp_connect_ms, http.tls_handshake_ms, http.ttfb_ms, http.transfer_ms, response].map(observationNumber)
  const timingMax = Math.max(1, ...timingValues.filter((value): value is number => value !== null))
  const timings = ['dns', 'tcp', 'tls', 'ttfb', 'transfer', 'total'].map((key, index) => ({ key,
    label: t('siteObservation.timings.' + key), value: timingValues[index] ?? null,
    text: ms(timingValues[index]), fraction: timingValues[index] === null ? 0 : timingValues[index]! / timingMax }))
  const kpis = [fact('response', ms(response)), fact('rtt', ms(ping.avg_rtt_ms)), fact('jitter', ms(ping.jitter_ms)), fact('loss', percent(ping.loss_rate))]
  const httpFacts = [fact('status', http.status_code), fact('protocol', http.http_protocol), fact('response', ms(response)),
    fact('finalUrl', http.final_url), fact('remoteIp', http.remote_ip), fact('contentType', http.content_type || header('content-type')),
    fact('bytes', observationNumber(http.body_read_bytes) === null ? undefined : `${http.body_read_bytes} B`),
    fact('server', http.server || header('server')), fact('observed', observationTime(httpEnvelope?.observed_at))]
  const commonHeaders = headers.filter(item => ['content-type', 'content-length', 'content-encoding', 'server', 'cache-control', 'location', 'date', 'etag', 'last-modified'].includes(item.key))

  const groupTypes = ['A', 'AAAA', 'CNAME', 'MX', 'NS', 'TXT', 'CAA', 'SOA']
  const groups = new Map<string, { value: string; ttl: string; details: ObservationFact[] }[]>()
  const seen = new Set<string>(), chains: string[][] = []
  const walk = (rows: Record<string, unknown>[], fallback: string, chain: string[] = [], depth = 0) => {
    if (depth > 20) return
    for (const row of rows) {
      const type = text(row.type).toUpperCase() || fallback, value = text(row.value)
      const identity = `${type}:${value}:${row.ttl}`
      if (value && groupTypes.includes(type) && !seen.has(identity)) {
        seen.add(identity)
        const details = compact(['asn', 'isp', 'country', 'reverse_ptr', 'dnssec', 'provider_type', 'hijacked', 'risk_flags'].map(key => fact(key, row[key])))
        groups.set(type, [...(groups.get(type) ?? []), { value, ttl: observationNumber(row.ttl) === null ? '—' : `${row.ttl} s`, details }])
      }
      const next = value ? [...chain, value] : chain
      const children = records(row.children)
      if (type === 'CNAME' || chain.length) {
        if (!children.length && next.length) chains.push([target, ...next])
        walk(children, type, next, depth + 1)
      } else walk(children, type, [], depth + 1)
    }
  }
  for (const type of groupTypes) walk(records(dns[type]), type)
  const dnsGroups = groupTypes.flatMap(type => groups.has(type) ? [{ type, records: groups.get(type)! }] : [])
  const dnsFacts = [
    ...['A', 'AAAA', 'CNAME', 'MX', 'NS'].map(type => ({ key: type, label: type, value: Array.isArray(dns[type]) ? String(groups.get(type)?.length ?? 0) : '—' })),
    fact('duration', ms(dnsEnvelope?.duration_ms)), fact('cnameDepth', dns.cname_chain_depth), fact('cnameTerminal', dns.cname_terminal),
  ]
  const meta = observationRecord(http.meta)
  const metadata = [fact('title', http.title), fact('description', meta.description), fact('charset', http.html_charset || meta.charset), fact('keywords', meta.keywords)]
  const webProbes = ['robots', 'llms_txt', 'page_assets', 'rdap'].map(protocol => {
    const envelope = matching(light?.protocols?.[protocol]), data = observationRecord(envelope?.payload)
    const sections: { key: string; title: string; items: ObservationFact[]; disclosure: boolean }[] = []
    const section = (key: string, items: ObservationFact[], disclosure = false) => sections.push({ key, title: t('siteObservation.sections.' + key), items, disclosure })
    if (protocol === 'robots') {
      section('robots', [fact('exists', data.exists), fact('status', data.status_code), fact('sitemaps', data.sitemap_count), fact('userAgentStar', data.user_agent_star_present), fact('disallowAll', data.global_disallow_all)])
      if (strings(data.sitemaps).length) section('sitemaps', [fact('sitemaps', data.sitemaps)], true)
    } else if (protocol === 'llms_txt') {
      section('llms', [fact('exists', data.exists), fact('path', data.path), fact('status', data.status_code), fact('title', data.title),
        fact('headings', data.heading_count), fact('links', data.link_count), fact('optional', data.optional_section_present), fact('bytes', data.body_read_bytes)])
      if (strings(data.headings).length || strings(data.links).length) section('contents', compact([fact('headings', data.headings), fact('links', data.links)]), true)
    } else if (protocol === 'page_assets') {
      const icon = observationRecord(data.icon), manifest = observationRecord(data.manifest)
      section('favicon', [fact('exists', icon.exists), fact('url', icon.source_url), fact('contentType', icon.content_type), fact('status', icon.status_code)])
      section('manifest', [fact('exists', manifest.exists), fact('url', manifest.source_url), fact('name', manifest.name || manifest.short_name),
        fact('display', manifest.display), fact('themeColor', manifest.theme_color), fact('startUrl', manifest.start_url), fact('scope', manifest.scope), fact('icons', manifest.icons_count)])
    } else {
      section('rdap', [fact('registrableDomain', data.registrable_domain), fact('registrar', data.registrar), fact('expires', data.expires_at),
        fact('statuses', data.statuses), fact('nameservers', data.nameservers), fact('dnssecDelegation', data.dnssec_delegation_signed)])
    }
    return { protocol, status: observationStatus(envelope?.status), statusLabel: t('siteDetail.states.' + observationStatus(envelope?.status)),
      observed: observationTime(envelope?.observed_at), duration: ms(envelope?.duration_ms), sections }
  })
  return { target, status, statusLabel: t('siteDetail.states.' + status), protocols, risks, endpoint, timings, kpis,
    http: { facts: httpFacts, headers, commonHeaders, redirects: strings(http.redirect_chain) },
    dns: { facts: dnsFacts, groups: dnsGroups, chains: [...new Map(chains.map(chain => [JSON.stringify(chain), chain])).values()], risks: strings(dns.risk_flags) },
    web: { metadata, probes: webProbes } }
}
export type SiteObservationPresentation = ReturnType<typeof presentSiteObservation>

export function presentPingHistory(items: CollectorEnvelope[], target: string) {
  return items.filter(item => (!item.target || item.target === target) && (!item.protocol || item.protocol === 'ping'))
    .map((item, index) => {
      const data = observationRecord(item.payload)
      return { key: `${item.observed_at}:${index}`, timestamp: Date.parse(item.observed_at) || 0,
        time: observationTime(item.observed_at), status: observationStatus(item.status),
        rtt: observationNumber(data.avg_rtt_ms), loss: observationNumber(data.loss_rate) }
    }).sort((a, b) => b.timestamp - a.timestamp)
}
export type PingHistoryRow = ReturnType<typeof presentPingHistory>[number]
