import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DnsItem, DnsRecord, HttpRecord, PingRecord, Site, SiteHealthSummary, SiteInfo, TargetHealthSummary } from '~/types/nav'
import { safeJsonParse } from '~/utils/util'

export interface SiteDetailPageData {
  siteInfo: SiteInfo | null
  domain: string
  sitePingRecord: PingRecord | null
  siteHttpRecord: HttpRecord | null
  siteDnsRecord: DnsRecord | null
  siteHealthSummary: SiteHealthSummary | null
  targetHealthSummary: TargetHealthSummary | null
}

function extractPrimaryDomain(rawDomain: unknown): string {
  if (typeof rawDomain !== 'string' || !rawDomain.trim()) {
    return ''
  }

  try {
    const parsed = JSON.parse(rawDomain)

    if (Array.isArray(parsed)) {
      return typeof parsed[0] === 'string' ? parsed[0] : rawDomain
    }

    if (Array.isArray(parsed?.domain)) {
      return typeof parsed.domain[0] === 'string' ? parsed.domain[0] : rawDomain
    }
  } catch {
    return rawDomain
  }

  return rawDomain
}

function parseDnsRecord(record: DnsRecord | null): DnsRecord | null {
  if (!record) {
    return null
  }

  const parsedRecord = { ...record }

  for (const key in parsedRecord) {
    const value = parsedRecord[key as keyof DnsRecord]
    if (typeof value === 'string') {
      parsedRecord[key as keyof DnsRecord] = safeJsonParse<DnsItem[]>(value) as never
    }
  }

  return parsedRecord
}

export async function useSiteDetailPage() {
  const route = useRoute()
  const { locale } = useI18n()
  const config = useRuntimeConfig()
  const navApi = useApi('nav')
  const navV2Api = useApi('navV2')

  const siteId = computed(() => String(route.params.id ?? ''))
  const queryDomain = computed(() => {
    const value = route.query.domain
    return typeof value === 'string' ? value : ''
  })
  const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
  const healthSummaryEnabled = computed(() => {
    const enabled = config.public.navHealthSummaryEnabled
    return String(enabled) === 'true'
  })

  const asyncData = await useAsyncData<SiteDetailPageData>(
    () => `site-detail:${route.path}:${siteId.value}:${queryDomain.value}:${lang.value}`,
    async () => {
      const siteInfo = await navApi<SiteInfo>('/nav/site/getSiteDetail', {
        query: {
          id: siteId.value,
          lang: lang.value,
        },
      }).catch(() => null)

      let resolvedDomain = queryDomain.value
      if (!resolvedDomain && siteId.value) {
        const siteList = await navApi<Site[]>('/nav/page/site/list', {
          query: {
            lang: lang.value,
          },
        }).catch(() => [])
        const matchedSite = siteList.find((site) => String(site.id) === siteId.value)
        resolvedDomain = extractPrimaryDomain(matchedSite?.domain)
      }

      if (!resolvedDomain) {
        const siteHealthSummary = healthSummaryEnabled.value && siteId.value
          ? await navV2Api<SiteHealthSummary>(`/nav/sites/${siteId.value}/summary`).catch(() => null)
          : null
        return {
          siteInfo,
          domain: '',
          sitePingRecord: null,
          siteHttpRecord: null,
          siteDnsRecord: null,
          siteHealthSummary,
          targetHealthSummary: null,
        }
      }

      const [httpRecord, dnsRecord, pingRecord, siteHealthSummary, targetHealthSummary] = await Promise.all([
        navApi<HttpRecord>('/nav/site/getSiteHttpRecord', {
          query: {
            domain: resolvedDomain,
          },
        }).catch(() => null),
        navApi<DnsRecord>('/nav/site/getSiteDnsRecord', {
          query: {
            domain: resolvedDomain,
          },
        }).catch(() => null),
        navApi<PingRecord>('/nav/site/getSitePingRecord', {
          query: {
            domain: resolvedDomain,
          },
        }).catch(() => null),
        healthSummaryEnabled.value && siteId.value
          ? navV2Api<SiteHealthSummary>(`/nav/sites/${siteId.value}/summary`).catch(() => null)
          : Promise.resolve(null),
        healthSummaryEnabled.value && siteId.value
          ? navV2Api<TargetHealthSummary>(`/nav/sites/${siteId.value}/targets/${encodeURIComponent(resolvedDomain)}/summary`).catch(() => null)
          : Promise.resolve(null),
      ])

      return {
        siteInfo,
        domain: resolvedDomain,
        sitePingRecord: pingRecord,
        siteHttpRecord: safeJsonParse<HttpRecord>(httpRecord),
        siteDnsRecord: parseDnsRecord(dnsRecord),
        siteHealthSummary,
        targetHealthSummary,
      }
    },
    {
      watch: [siteId, queryDomain, lang],
      default: () => ({
        siteInfo: null,
        domain: '',
        sitePingRecord: null,
        siteHttpRecord: null,
        siteDnsRecord: null,
        siteHealthSummary: null,
        targetHealthSummary: null,
      }),
    }
  )

  return {
    ...asyncData,
    siteId,
    queryDomain,
    lang,
    healthSummaryEnabled,
  }
}
