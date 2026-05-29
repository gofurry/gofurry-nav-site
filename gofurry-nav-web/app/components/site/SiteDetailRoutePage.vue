<template>
  <div ref="pageRoot" class="flex min-h-full w-full flex-col overflow-x-hidden bg-orange-50 text-gray-800">
    <div v-if="pending" class="flex flex-1 items-center justify-center text-gray-500">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="error" class="flex flex-1 items-center justify-center text-red-500">
      {{ loadFailedText }}
    </div>

    <div v-else class="contents">
      <div class="mx-10 my-8">
        <SiteOverview
          v-if="sitePageData.siteInfo"
          :site="{
            name: sitePageData.siteInfo.name || '',
            icon: sitePageData.siteInfo.icon || undefined,
            domain: sitePageData.domain || '',
            welfare: sitePageData.siteInfo.welfare === '1',
            nsfw: sitePageData.siteInfo.nsfw === '1',
            description: sitePageData.siteInfo.info || '',
            viewCount: sitePageData.siteInfo.view_count ?? 0,
          }"
          :domain-options="domainOptions"
          :domain-route-suffix="domainRouteSuffix"
          :enable-domain-switcher="enableDomainSwitcher"
          :edge-provider-hints="overviewEdgeProviderHints"
          :keywords="overviewKeywords"
          :site-id="siteId"
        />
      </div>

      <div v-if="sitePageData.siteHealthSummary || sitePageData.targetHealthSummary" class="mx-10 mb-8">
        <SiteHealthSummaryPanel
          :current-target="sitePageData.domain"
          :mode="healthSummaryMode"
          :security-headers="securityHeaderItems"
          :site-summary="sitePageData.siteHealthSummary"
          :target-summary="sitePageData.targetHealthSummary"
        />
      </div>

      <div class="mx-10 mb-8">
        <SitePerformance
          v-if="sitePageData.sitePingRecord && sitePageData.siteHttpRecord"
          :ping-record="sitePageData.sitePingRecord"
          :http-record="sitePageData.siteHttpRecord"
          :target-latest-core="sitePageData.targetLatestCore"
          :mode="healthSummaryMode"
        />
      </div>

      <div class="mx-10 mb-8">
        <SiteHttpPanel
          v-if="sitePageData.siteHttpRecord"
          :record="sitePageData.siteHttpRecord"
        />
      </div>

      <div class="mx-10 mb-8">
        <SiteDnsPanel
          v-if="sitePageData.siteDnsRecord"
          :record="sitePageData.siteDnsRecord"
        />
      </div>

      <div class="mb-8 mr-4 flex flex-wrap items-center justify-center gap-3 text-orange-800">
        <button
          class="flex items-center justify-center gap-2 rounded-lg bg-orange-300 px-4 py-2 text-sm transition-colors hover:bg-orange-200"
          @click="generateReport"
        >
          {{ t('common.save') }}
        </button>

        <button
          class="flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm transition-colors hover:bg-orange-100"
          @click="() => refresh()"
        >
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import SiteDnsPanel from '@/components/site/SiteDnsPanel.vue'
import SiteHealthSummaryPanel from '@/components/site/SiteHealthSummaryPanel.vue'
import SiteHttpPanel from '@/components/site/SiteHttpPanel.vue'
import SiteOverview from '@/components/site/SiteOverview.vue'
import SitePerformance from '@/components/site/SitePerformance.vue'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'

const props = withDefaults(defineProps<{
  enableDomainSwitcher?: boolean
  domainRouteSuffix?: string
}>(), {
  enableDomainSwitcher: false,
  domainRouteSuffix: '',
})

const { t } = useI18n()
const { data, pending, error, refresh, siteId } = await useSiteDetailPage({
  includeV2Latest: props.enableDomainSwitcher,
})
const sitePageData = computed(() => data.value!)
const pageRoot = ref<HTMLElement | null>(null)
const loadFailedText = computed(() => (t('common.loading') === 'Loading...' ? 'Failed to load site data.' : '站点数据加载失败。'))
const enableDomainSwitcher = computed(() => props.enableDomainSwitcher)
const domainRouteSuffix = computed(() => props.domainRouteSuffix)
const healthSummaryMode = computed(() => (props.enableDomainSwitcher ? 'v2' : 'default'))
const domainOptions = computed(() => {
  const targets = sitePageData.value.siteHealthSummary?.targets ?? []
  const seen = new Set<string>()
  const options: string[] = []

  for (const target of targets) {
    const value = target.target?.trim()
    if (!value || seen.has(value)) {
      continue
    }

    seen.add(value)
    options.push(value)
  }

  const currentDomain = sitePageData.value.domain?.trim()
  if (currentDomain && !seen.has(currentDomain)) {
    options.unshift(currentDomain)
  }

  return options
})
const overviewKeywords = computed(() => {
  if (!props.enableDomainSwitcher) {
    return []
  }

  const rawKeywords = sitePageData.value.siteHttpRecord?.meta?.keywords
  if (!rawKeywords) {
    return []
  }

  return rawKeywords
    .split(/[,，;；]/)
    .map(keyword => keyword.trim())
    .filter(Boolean)
    .filter((keyword, index, list) => list.indexOf(keyword) === index)
    .slice(0, 12)
})
const overviewEdgeProviderHints = computed(() => {
  return props.enableDomainSwitcher ? sitePageData.value.targetHealthSummary?.edge_provider_hints ?? [] : []
})
const securityHeaderItems = computed(() => {
  if (!props.enableDomainSwitcher) {
    return []
  }

  const compactSummary = securityHeadersCompact()
  if (Object.keys(compactSummary).length) {
    return [
      { label: 'HSTS', ok: Boolean(compactSummary.strict_transport_security) },
      { label: 'CSP', ok: Boolean(compactSummary.content_security_policy) },
      { label: 'X-Frame-Options', ok: Boolean(compactSummary.x_frame_options) },
      { label: 'X-Content-Type-Options', ok: Boolean(compactSummary.x_content_type_options) },
      { label: 'Referrer-Policy', ok: Boolean(compactSummary.referrer_policy) },
      { label: 'Permissions-Policy', ok: Boolean(compactSummary.permissions_policy) },
    ]
  }

  const detailedSummary = securityHeaderSummary()
  if (Object.keys(detailedSummary).length) {
    return [
      { label: 'HSTS', ok: Boolean(asRecord(detailedSummary.hsts).present) },
      { label: 'CSP', ok: Boolean(asRecord(detailedSummary.content_security_policy).present) },
      { label: 'X-Frame-Options', ok: Boolean(asRecord(detailedSummary.x_frame_options).present) },
      { label: 'X-Content-Type-Options', ok: Boolean(asRecord(detailedSummary.x_content_type_options).present) },
      { label: 'Referrer-Policy', ok: Boolean(asRecord(detailedSummary.referrer_policy).present) },
      { label: 'Permissions-Policy', ok: Boolean(asRecord(detailedSummary.permissions_policy).present) },
    ]
  }

  const headers = normalizeHeaders(sitePageData.value.siteHttpRecord?.headers)
  return [
    { label: 'HSTS', ok: hasHeader(headers, 'strict-transport-security') },
    { label: 'CSP', ok: hasHeader(headers, 'content-security-policy') },
    { label: 'X-Frame-Options', ok: hasHeader(headers, 'x-frame-options') },
    { label: 'X-Content-Type-Options', ok: hasHeader(headers, 'x-content-type-options') },
    { label: 'Referrer-Policy', ok: hasHeader(headers, 'referrer-policy') },
    { label: 'Permissions-Policy', ok: hasHeader(headers, 'permissions-policy') },
  ]
})

const seoTitle = computed(() => {
  const name = sitePageData.value.siteInfo?.name?.trim()
  return name ? `${name} - GoFurry` : `${t('site.title')} - GoFurry`
})
const seoDescription = computed(() => {
  const description = sitePageData.value.siteInfo?.info?.trim() ?? ''
  return description.slice(0, 160)
})

useSeoMeta({
  title: () => seoTitle.value,
  description: () => seoDescription.value,
  ogTitle: () => seoTitle.value,
  ogDescription: () => seoDescription.value,
})

const generateReport = () => {}

function normalizeHeaders(headers?: Record<string, string[]>) {
  const normalized: Record<string, string[]> = {}

  for (const [key, value] of Object.entries(headers ?? {})) {
    normalized[key.toLowerCase()] = Array.isArray(value) ? value : [String(value)]
  }

  return normalized
}

function hasHeader(headers: Record<string, string[]>, key: string) {
  return (headers[key]?.length ?? 0) > 0
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}

function securityHeaderSummary() {
  const httpPayload = asRecord(sitePageData.value.targetLatestCore?.protocols?.http?.payload)
  return asRecord(httpPayload.security_header_summary)
}

function securityHeadersCompact() {
  const httpPayload = asRecord(sitePageData.value.targetLatestCore?.protocols?.http?.payload)
  return asRecord(httpPayload.security_headers)
}
</script>
