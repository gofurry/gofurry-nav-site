import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CollectorEnvelope, DnsRecord, HttpRecord, PingRecord, SiteHealthSummary, SiteInfo, SiteV2DetailResponse, SiteV2Info, TargetHealthSummary, TargetLatestResponse } from '~/types/nav'

export interface SiteDetailPageData {
  siteInfo: SiteInfo | null
  domain: string
  sitePingRecord: PingRecord | null
  siteHttpRecord: HttpRecord | null
  siteDnsRecord: DnsRecord | null
  siteHealthSummary: SiteHealthSummary | null
  targetHealthSummary: TargetHealthSummary | null
  targetLatestCore: TargetLatestResponse | null
  lightProbeState: TargetLatestResponse | null
}

function extractRouteParam(value: unknown): string {
  const rawValue = Array.isArray(value) ? value[0] : value
  if (typeof rawValue !== 'string') {
    return ''
  }

  try {
    return decodeURIComponent(rawValue).trim()
  } catch {
    return rawValue.trim()
  }
}

export async function useSiteDetailPage() {
  const route = useRoute()
  const { locale } = useI18n()
  const navV2Api = useApi('navV2')

  const siteId = computed(() => String(route.params.id ?? ''))
  const pathDomain = computed(() => extractRouteParam(route.params.domain))
  const queryDomain = computed(() => {
    const value = route.query.domain
    return typeof value === 'string' ? value : ''
  })
  const selectedDomain = computed(() => pathDomain.value || queryDomain.value)
  const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))

  const asyncData = await useAsyncData<SiteDetailPageData>(
    () => `site-detail:${route.path}:${siteId.value}:${selectedDomain.value}:${lang.value}:v2`,
    async () => {
      if (!siteId.value) {
        throw new Error('invalid site id')
      }

      const detail = await navV2Api<SiteV2DetailResponse>(`/nav/sites/${siteId.value}/detail`, {
        query: {
          lang: lang.value,
          target: selectedDomain.value || undefined,
          payload_mode: 'preview',
        },
      })
      const resolvedDomain = detail.selected_target || selectedDomain.value

      return {
        siteInfo: toSiteInfo(detail.site),
        domain: resolvedDomain,
        sitePingRecord: null,
        siteHttpRecord: toHttpRecord(detail.latest_core?.protocols?.http, resolvedDomain),
        siteDnsRecord: null,
        siteHealthSummary: detail.site_summary,
        targetHealthSummary: detail.target_summary,
        targetLatestCore: detail.latest_core,
        lightProbeState: detail.light_probe_state,
      }
    },
    {
      watch: [siteId, selectedDomain, lang],
      default: () => ({
        siteInfo: null,
        domain: '',
        sitePingRecord: null,
        siteHttpRecord: null,
        siteDnsRecord: null,
        siteHealthSummary: null,
        targetHealthSummary: null,
        targetLatestCore: null,
        lightProbeState: null,
      }),
    }
  )

  return {
    ...asyncData,
    siteId,
    queryDomain,
    pathDomain,
    lang,
  }
}

function toSiteInfo(site: SiteV2Info | null | undefined): SiteInfo | null {
  if (!site) {
    return null
  }
  return {
    name: site.name,
    info: site.info,
    icon: site.icon,
    country: site.country,
    nsfw: site.nsfw,
    welfare: site.welfare,
    view_count: site.view_count,
  }
}

function toHttpRecord(envelope: CollectorEnvelope | undefined, domain: string): HttpRecord | null {
  const payload = asRecord(envelope?.payload)
  if (!Object.keys(payload).length) {
    return null
  }
  const meta = asRecord(payload.meta)
  const headers = normalizeHeaderArrays(payload.headers)
  return {
    domain,
    url: stringValue(payload.url || payload.final_url),
    statusCode: numberValue(payload.status_code),
    responseTime: `${numberValue(payload.response_time_ms || envelope?.duration_ms)}ms`,
    contentLength: numberValue(payload.body_read_bytes),
    title: stringValue(payload.title),
    server: stringValue(payload.server || headerValue(headers, 'server')),
    redirects: stringArray(payload.redirects || payload.redirect_chain),
    headers,
    meta: {
      charset: stringValue(meta.charset),
      description: stringValue(meta.description),
      keywords: stringValue(meta.keywords),
    },
    tlsVersion: stringValue(payload.tls_version),
    cipherSuite: stringValue(payload.cipher_suite),
    certExpiry: stringValue(payload.cert_not_after || payload.cert_expiry),
    certDaysLeft: String(numberValue(payload.cert_days_left)),
    certIssuer: stringValue(payload.cert_issuer_cn || payload.cert_issuer),
    certIssuerOrg: stringArray(payload.cert_issuer_org),
    certDNSNames: stringArray(payload.cert_dns_names),
    certPubKeyAlg: stringValue(payload.cert_public_key_algorithm || payload.cert_pub_key_alg),
    certSigAlg: stringValue(payload.cert_signature_algorithm || payload.cert_sig_alg),
    certEmail: stringValue(payload.cert_email) || null,
    certIsCA: Boolean(payload.cert_is_ca),
  }
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}

function stringValue(value: unknown) {
  return typeof value === 'string' ? value.trim() : ''
}

function numberValue(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return Math.round(value)
  }
  if (typeof value === 'string') {
    const matched = value.match(/-?\d+(\.\d+)?/)
    return matched ? Math.round(Number(matched[0])) : 0
  }
  return 0
}

function stringArray(value: unknown) {
  return Array.isArray(value) ? value.map(String).filter(Boolean) : []
}

function normalizeHeaderArrays(value: unknown): Record<string, string[]> {
  const headers = asRecord(value)
  const normalized: Record<string, string[]> = {}
  for (const [key, rawValue] of Object.entries(headers)) {
    normalized[key] = Array.isArray(rawValue) ? rawValue.map(String) : [String(rawValue)]
  }
  return normalized
}

function headerValue(headers: Record<string, string[]>, key: string) {
  const normalizedKey = key.toLowerCase()
  for (const [name, values] of Object.entries(headers)) {
    if (name.toLowerCase() === normalizedKey && values.length) {
      return values[0]
    }
  }
  return ''
}
