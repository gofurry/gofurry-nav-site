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
          :edge-provider-hints="overviewEdgeProviderHints"
          :keywords="overviewKeywords"
          :site-id="siteId"
        />
      </div>

      <div class="mx-10 mb-8">
        <SitePerformance
          v-if="sitePageData.siteHttpRecord"
          :domain="sitePageData.domain"
          :ping-record="sitePageData.sitePingRecord"
          :http-record="sitePageData.siteHttpRecord"
          :site-id="siteId"
          :target-latest-core="sitePageData.targetLatestCore"
        >
          <template v-if="sitePageData.siteHealthSummary || sitePageData.targetHealthSummary" #after-metrics>
            <div class="mb-8">
              <SiteHealthSummaryPanel
                :current-target="sitePageData.domain"
                :security-headers="securityHeaderItems"
                :site-summary="sitePageData.siteHealthSummary"
                :target-summary="sitePageData.targetHealthSummary"
              />
            </div>
          </template>
        </SitePerformance>
      </div>

      <div class="mx-10 mb-8">
        <SiteObservationTabs
          :dns-record="sitePageData.siteDnsRecord"
          :http-record="sitePageData.siteHttpRecord"
          :ping-record="sitePageData.sitePingRecord"
          :target-latest-core="sitePageData.targetLatestCore"
        />
      </div>

      <div class="mx-10 mb-8">
        <SiteMetadataProbePanel
          :http-record="sitePageData.siteHttpRecord"
          :light-probe-state="sitePageData.lightProbeState"
          :site-id="siteId"
          :target="sitePageData.domain"
          :target-latest-core="sitePageData.targetLatestCore"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import SiteHealthSummaryPanel from '@/components/site/SiteHealthSummaryPanel.vue'
import SiteMetadataProbePanel from '@/components/site/SiteMetadataProbePanel.vue'
import SiteObservationTabs from '@/components/site/SiteObservationTabs.vue'
import SiteOverview from '@/components/site/SiteOverview.vue'
import SitePerformance from '@/components/site/SitePerformance.vue'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'

const { t } = useI18n()
const { data, pending, error, siteId } = await useSiteDetailPage()
const navV2Api = useApi('navV2')
const sitePageData = computed(() => data.value!)
const pageRoot = ref<HTMLElement | null>(null)
const loadFailedText = computed(() => (t('common.loading') === 'Loading...' ? 'Failed to load site data.' : '站点数据加载失败。'))
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
  return sitePageData.value.targetHealthSummary?.edge_provider_hints ?? []
})
const securityHeaderItems = computed(() => {
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

onMounted(() => {
  watch(siteId, (value) => {
    void touchSiteView(value)
  }, { immediate: true })
})

async function touchSiteView(value: string) {
  if (!value) {
    return
  }

  try {
    const response = await navV2Api<{ site_id: number; view_count: number }>(`/nav/sites/${value}/view`, { method: 'POST' })
    if (data.value?.siteInfo && Number.isFinite(response.view_count)) {
      data.value.siteInfo.view_count = response.view_count
    }
  } catch {
    // 浏览量统计是旁路副作用，失败不影响详情页展示。
  }
}

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
