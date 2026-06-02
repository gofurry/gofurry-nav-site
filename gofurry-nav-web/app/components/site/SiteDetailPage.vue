<template>
  <div ref="pageRoot" class="site-detail-page min-h-full overflow-x-hidden text-slate-900">
    <div class="visual-sky" aria-hidden="true">
      <svg class="visual-grid" viewBox="0 0 1200 720" preserveAspectRatio="none">
        <defs>
          <pattern id="quietGrid" width="96" height="96" patternUnits="userSpaceOnUse">
            <path d="M96 0H0V96" class="grid-line" />
          </pattern>
          <pattern id="quietDots" width="192" height="192" patternUnits="userSpaceOnUse">
            <circle cx="96" cy="96" r="2" class="grid-dot" />
          </pattern>
          <linearGradient id="quietDiagonal" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#fb8c2f" stop-opacity="0.10" />
            <stop offset="58%" stop-color="#4f6fed" stop-opacity="0.07" />
            <stop offset="100%" stop-color="#15b8a6" stop-opacity="0.05" />
          </linearGradient>
        </defs>
        <rect width="1200" height="720" fill="url(#quietGrid)" />
        <rect width="1200" height="720" fill="url(#quietDots)" />
        <path class="quiet-diagonal" d="M-80 610 C250 530 420 660 650 560 C840 478 970 512 1280 398" />
        <path class="quiet-diagonal is-cool" d="M-120 210 C220 140 360 272 590 202 C810 136 940 190 1270 82" />
      </svg>
    </div>

    <div v-if="pending" class="relative flex min-h-[68vh] items-center justify-center text-slate-500">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="error" class="relative flex min-h-[68vh] items-center justify-center text-red-500">
      {{ loadFailedText }}
    </div>

    <main v-else class="relative mx-auto w-full max-w-[1560px] px-5 py-8 sm:px-8 lg:px-10">
      <SiteDetailHero
        :badges="heroBadges"
        :domain="sitePageData.domain"
        :icon="sitePageData.siteInfo?.icon || undefined"
        :info="sitePageData.siteInfo?.info || undefined"
        :keywords="overviewKeywords"
        :logo-prefix="siteLogoPrefix"
        :site-id="siteId"
        :site-name="siteName"
        :switchable-domains="switchableDomains"
        :view-count="sitePageData.siteInfo?.view_count ?? 0"
        :visit-url="visitUrl"
      />

      <SiteSignalCards :cards="signalCards" />

      <SiteObservationOverviewPanel
        :protocol-availability-text="protocolAvailabilityText"
        :protocol-entries="protocolTrackEntries"
        :risk-messages="targetRiskMessages"
        :security-header-items="securityHeaderItems"
        :security-header-ratio="securityHeaderRatio"
        :strip-items="observationStripItems"
      />

      <SitePerformancePanel
        v-if="sitePageData.siteHttpRecord"
        :domain="sitePageData.domain"
        :http-record="sitePageData.siteHttpRecord"
        :ping-record="sitePageData.sitePingRecord"
        :site-id="siteId"
        :target-latest-core="sitePageData.targetLatestCore"
      />

      <section class="detail-observation-tabs">
        <SiteObservationTabs
          :dns-record="sitePageData.siteDnsRecord"
          :http-record="sitePageData.siteHttpRecord"
          :ping-record="sitePageData.sitePingRecord"
          :target-latest-core="sitePageData.targetLatestCore"
        />
      </section>

      <section class="detail-section">
        <SiteMetadataProbePanel
          :http-record="sitePageData.siteHttpRecord"
          :light-probe-state="sitePageData.lightProbeState"
          :site-id="siteId"
          :target="sitePageData.domain"
          :target-latest-core="sitePageData.targetLatestCore"
        />
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import SiteDetailHero from '@/components/site/SiteDetailHero.vue'
import SiteMetadataProbePanel from '@/components/site/SiteMetadataProbePanel.vue'
import SiteObservationOverviewPanel from '@/components/site/SiteObservationOverviewPanel.vue'
import SiteObservationTabs from '@/components/site/SiteObservationTabs.vue'
import SitePerformancePanel from '@/components/site/SitePerformancePanel.vue'
import SiteSignalCards from '@/components/site/SiteSignalCards.vue'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'

const { locale, t } = useI18n()
const { data, pending, error, siteId } = await useSiteDetailPage()
const navV2Api = useApi('navV2')
const config = useRuntimeConfig()
const pageRoot = ref<HTMLElement | null>(null)
const sitePageData = computed(() => data.value!)
const siteName = computed(() => sitePageData.value.siteInfo?.name?.trim() || 'GoFurry')
const siteLogoPrefix = computed(() => String(config.public?.siteLogoPrefixUrl || ''))
const loadFailedText = computed(() => (t('common.loading') === 'Loading...' ? 'Failed to load site data.' : '站点数据加载失败。'))
const httpPayload = computed(() => asRecord(sitePageData.value.targetLatestCore?.protocols?.http?.payload))
const primaryEdgeLabel = computed(() => {
  const hint = sitePageData.value.targetHealthSummary?.edge_provider_hints?.[0]
  if (!hint) {
    return ''
  }
  return `${providerLabel(hint.provider)} · ${hint.hint_type?.toUpperCase() || 'EDGE'}`
})
const heroBadges = computed(() => {
  const badges: { label: string; class: string }[] = []
  if (primaryEdgeLabel.value) {
    badges.push({ label: primaryEdgeLabel.value, class: 'detail-pill-edge' })
  }

  badges.push({
    label: sitePageData.value.siteInfo?.nsfw === '1' ? 'NSFW' : 'SFW',
    class: sitePageData.value.siteInfo?.nsfw === '1' ? 'detail-pill-risk' : 'detail-pill-sfw',
  })

  if (sitePageData.value.siteInfo?.welfare === '1') {
    badges.push({ label: label('公益网站', 'Nonprofit'), class: 'detail-pill-welfare' })
  }

  return badges
})
const visitUrl = computed(() => {
  const url = firstString(httpPayload.value.final_url, sitePageData.value.siteHttpRecord?.url)
  if (url) {
    return url
  }
  const domain = sitePageData.value.domain
  return domain ? `https://${domain}` : ''
})
const overviewKeywords = computed(() => {
  const rawKeywords = firstString(sitePageData.value.siteHttpRecord?.meta?.keywords, asRecord(httpPayload.value.meta).keywords)
  if (!rawKeywords) {
    return []
  }

  return rawKeywords
    .split(/[,，;；]/)
    .map(keyword => keyword.trim())
    .filter(Boolean)
    .filter((keyword, index, list) => list.indexOf(keyword) === index)
    .slice(0, 10)
})
const switchableDomains = computed(() => {
  const seen = new Set<string>()
  const domains: string[] = []
  for (const target of sitePageData.value.siteHealthSummary?.targets ?? []) {
    const value = target.target?.trim()
    if (!value || seen.has(value)) {
      continue
    }

    seen.add(value)
    domains.push(value)
  }

  const current = sitePageData.value.domain?.trim()
  if (current && !seen.has(current)) {
    domains.unshift(current)
  }

  return domains
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
const protocolTrackEntries = computed(() => {
  const protocols = asRecord(sitePageData.value.targetHealthSummary?.protocols)
  return ['ping', 'http', 'dns']
    .map((protocol) => {
      const summary = asRecord(protocols[protocol])
      if (!Object.keys(summary).length) {
        return null
      }

      const stale = Boolean(summary.stale)
      const status = firstString(summary.status, 'unknown')
      return {
        protocol,
        label: protocol.toUpperCase(),
        status: stale ? label('已过期', 'Stale') : protocolStatusText(status),
        duration: formatMs(firstNumber(summary.duration_ms)),
        observedAt: formatDateTime(firstString(summary.observed_at)),
        staleAfter: `${firstNumber(summary.stale_after_seconds) ?? '-'}s`,
        tone: status === 'success' && !stale ? 'is-ok' : status === 'failure' ? 'is-bad' : 'is-warn',
      }
    })
    .filter(Boolean) as Array<{
      protocol: string
      label: string
      status: string
      duration: string
      observedAt: string
      staleAfter: string
      tone: string
    }>
})
const protocolAvailableCount = computed(() => protocolTrackEntries.value.filter(entry => entry.tone === 'is-ok').length)
const protocolAvailabilityText = computed(() => `${protocolAvailableCount.value}/${protocolTrackEntries.value.length || 3}`)
const targetRiskMessages = computed(() => {
  const summary = sitePageData.value.targetHealthSummary
  const messages = summary?.reason_messages?.filter(Boolean) ?? []
  if (messages.length) {
    return messages
  }

  return summary?.reason_codes?.filter(Boolean) ?? []
})
const securityHeaderOkCount = computed(() => securityHeaderItems.value.filter(item => item.ok).length)
const securityHeaderRatio = computed(() => `${securityHeaderOkCount.value}/${securityHeaderItems.value.length || 6}`)
const observationStripItems = computed(() => [
  {
    label: label('当前目标', 'Target'),
    value: sitePageData.value.domain || '-',
    tone: '',
  },
  {
    label: label('最近观测', 'Observed'),
    value: formatDateTime(firstString(sitePageData.value.targetHealthSummary?.observed_at)),
    tone: '',
  },
  {
    label: label('协议可用', 'Protocols'),
    value: protocolAvailabilityText.value,
    tone: protocolAvailableCount.value === protocolTrackEntries.value.length ? 'is-ok' : 'is-warn',
  },
  {
    label: label('风险信号', 'Signals'),
    value: targetRiskMessages.value.length ? `${targetRiskMessages.value.length}${label('项', '')}` : label('无', 'None'),
    tone: targetRiskMessages.value.length ? 'is-warn' : 'is-ok',
  },
])
const signalCards = computed(() => [
  {
    eyebrow: label('用户访问', 'Visitor access'),
    title: label('访问耗时', 'Latency'),
    value: formatMs(firstNumber(httpPayload.value.response_time_ms, sitePageData.value.targetLatestCore?.protocols?.http?.duration_ms)),
    badge: label('端到端', 'E2E'),
    tone: latencyTone(firstNumber(httpPayload.value.response_time_ms, 0)),
    items: [
      { label: 'DNS', value: formatMs(firstNumber(httpPayload.value.dns_lookup_ms)) },
      { label: 'TCP', value: formatMs(firstNumber(httpPayload.value.tcp_connect_ms)) },
      { label: 'TLS', value: formatMs(firstNumber(httpPayload.value.tls_handshake_ms)) },
      { label: 'TTFB', value: formatMs(firstNumber(httpPayload.value.ttfb_ms)) },
    ],
  },
  {
    eyebrow: label('内容响应', 'Content response'),
    title: 'HTTP',
    value: `HTTP ${firstNumber(httpPayload.value.status_code, sitePageData.value.siteHttpRecord?.statusCode) ?? '-'}`,
    badge: firstString(httpPayload.value.http_protocol, '-'),
    tone: 'tone-green',
    items: [
      { label: label('重定向', 'Redirects'), value: String(firstNumber(httpPayload.value.redirect_count, sitePageData.value.siteHttpRecord?.redirects?.length) ?? '-') },
      { label: label('类型', 'Type'), value: firstString(httpPayload.value.content_type, headerValue('content-type'), '-') },
      { label: label('压缩', 'Encoding'), value: firstString(httpPayload.value.content_encoding, boolMaybeText(httpPayload.value.compressed), '-') },
      { label: label('读取', 'Body'), value: bytesText(firstNumber(httpPayload.value.body_read_bytes, sitePageData.value.siteHttpRecord?.contentLength)) },
    ],
  },
  {
    eyebrow: label('安全传输', 'Secure transport'),
    title: 'TLS',
    value: firstString(httpPayload.value.tls_version, sitePageData.value.siteHttpRecord?.tlsVersion, '-'),
    badge: httpPayload.value.cert_verified === false ? label('需关注', 'Review') : label('已校验', 'Verified'),
    tone: httpPayload.value.cert_verified === false ? 'tone-amber' : 'tone-green',
    items: [
      { label: label('证书校验', 'Verified'), value: boolText(httpPayload.value.cert_verified) },
      { label: label('握手', 'Handshake'), value: firstString(httpPayload.value.tls_handshake, '-') },
      { label: label('套件', 'Cipher'), value: firstString(httpPayload.value.cipher_suite, '-') },
      { label: label('签名', 'Signature'), value: firstString(httpPayload.value.cert_signature_algorithm, httpPayload.value.cert_sig_alg, '-') },
    ],
  },
  {
    eyebrow: label('证书状态', 'Certificate'),
    title: label('有效期', 'Validity'),
    value: formatDays(firstNumber(httpPayload.value.cert_days_left, sitePageData.value.siteHttpRecord?.certDaysLeft)),
    badge: label('剩余', 'Left'),
    tone: certTone(firstNumber(httpPayload.value.cert_days_left, sitePageData.value.siteHttpRecord?.certDaysLeft)),
    items: [
      { label: label('签发者', 'Issuer'), value: firstString(httpPayload.value.cert_issuer_cn, httpPayload.value.cert_issuer, sitePageData.value.siteHttpRecord?.certIssuer, '-') },
      { label: label('到期', 'Expires'), value: shortDate(firstString(httpPayload.value.cert_not_after, httpPayload.value.cert_expiry, sitePageData.value.siteHttpRecord?.certExpiry)) },
      { label: 'SAN', value: numberOrDash(firstNumber(httpPayload.value.cert_san_count, sitePageData.value.siteHttpRecord?.certDNSNames?.length)) },
      { label: label('链长', 'Chain'), value: numberOrDash(firstNumber(httpPayload.value.cert_chain_length)) },
    ],
  },
])
const seoTitle = computed(() => `${siteName.value} - GoFurry`)
const seoDescription = computed(() => (sitePageData.value.siteInfo?.info?.trim() ?? '').slice(0, 160))

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

function securityHeaderSummary() {
  return asRecord(httpPayload.value.security_header_summary)
}

function securityHeadersCompact() {
  return asRecord(httpPayload.value.security_headers)
}

function headerValue(key: string) {
  const normalizedKey = key.toLowerCase()
  for (const [name, values] of Object.entries(sitePageData.value.siteHttpRecord?.headers ?? {})) {
    if (name.toLowerCase() === normalizedKey && values?.length) {
      return values[0]
    }
  }
  return ''
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}

function firstNumber(...values: unknown[]) {
  for (const value of values) {
    const parsed = typeof value === 'number' ? value : parseNumber(value)
    if (Number.isFinite(parsed)) {
      return parsed
    }
  }
  return null
}

function parseNumber(value: unknown) {
  if (typeof value === 'number') {
    return value
  }
  if (typeof value !== 'string') {
    return Number.NaN
  }
  const matched = value.match(/-?\d+(\.\d+)?/)
  return matched ? Number(matched[0]) : Number.NaN
}

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function formatMs(value: number | null) {
  return value === null ? '-' : `${Math.round(value)}ms`
}

function formatDays(value: number | null) {
  if (value === null) {
    return '-'
  }
  return `${Math.round(value)}${label('天', 'd')}`
}

function numberOrDash(value: number | null) {
  return value === null ? '-' : String(Math.round(value))
}

function shortDate(value: string) {
  if (!value) {
    return '-'
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '').slice(0, 10)
}

function formatDateTime(value: string) {
  if (!value) {
    return '-'
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '').slice(0, 19)
}

function protocolStatusText(status: string) {
  const map: Record<string, string> = {
    success: label('成功', 'Success'),
    failure: label('失败', 'Failed'),
    skipped: label('跳过', 'Skipped'),
    unknown: label('未知', 'Unknown'),
  }
  return map[status] ?? status
}

function boolText(value: unknown) {
  if (value === true) {
    return label('是', 'Yes')
  }
  if (value === false) {
    return label('否', 'No')
  }
  return '-'
}

function boolMaybeText(value: unknown) {
  if (typeof value === 'boolean') {
    return boolText(value)
  }
  return ''
}

function bytesText(value: number | null) {
  if (!value || value <= 0) {
    return '-'
  }
  if (value >= 1024 * 1024) {
    return `${(value / 1024 / 1024).toFixed(1)} MB`
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(1)} KB`
  }
  return `${Math.round(value)} B`
}

function latencyTone(value: number | null) {
  if (value !== null && value >= 800) {
    return 'tone-rose'
  }
  if (value !== null && value >= 300) {
    return 'tone-amber'
  }
  return 'tone-green'
}

function certTone(value: number | null) {
  if (value !== null && value <= 30) {
    return 'tone-rose'
  }
  if (value !== null && value <= 90) {
    return 'tone-amber'
  }
  return 'tone-green'
}

function providerLabel(provider: string) {
  const map: Record<string, string> = {
    cloudflare: 'Cloudflare',
    tencent_cloud: label('腾讯云', 'Tencent Cloud'),
    aliyun: label('阿里云', 'Aliyun'),
    aws_cloudfront: 'CloudFront',
    fastly: 'Fastly',
    vercel: 'Vercel',
    netlify: 'Netlify',
    github_pages: 'GitHub Pages',
  }
  return map[provider] || provider
}

function label(zh: string, en: string) {
  return locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.site-detail-page {
  --surface: rgba(255, 250, 242, 0.76);
  --surface-strong: rgba(255, 244, 226, 0.86);
  --ink-muted: #667085;
  --accent: #fb8c2f;
  isolation: isolate;
  position: relative;
  background:
    linear-gradient(180deg, rgba(255, 251, 245, 0.96) 0%, rgba(248, 240, 229, 0.96) 52%, rgba(255, 246, 234, 0.98) 100%);
}

.site-detail-page > main {
  z-index: 1;
}

.visual-sky {
  pointer-events: none;
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  opacity: 1;
  background:
    radial-gradient(circle at 18% 12%, rgba(251, 140, 47, 0.09), transparent 30%),
    radial-gradient(circle at 82% 18%, rgba(79, 111, 237, 0.07), transparent 28%),
    repeating-linear-gradient(0deg, transparent 0 95px, rgba(251, 140, 47, 0.055) 96px),
    repeating-linear-gradient(90deg, transparent 0 95px, rgba(79, 111, 237, 0.045) 96px);
}

.visual-grid {
  height: 100%;
  width: 100%;
  opacity: 1;
  mask-image: linear-gradient(90deg, transparent 0, #000 10%, #000 90%, transparent 100%);
}

.grid-line {
  fill: none;
  stroke: rgba(251, 140, 47, 0.11);
  stroke-width: 1;
}

.grid-dot {
  fill: rgba(251, 140, 47, 0.22);
}

.quiet-diagonal {
  fill: none;
  stroke: url(#quietDiagonal);
  stroke-width: 2;
  stroke-dasharray: 14 30;
  opacity: 0.72;
}

.quiet-diagonal.is-cool {
  opacity: 0.56;
}

.detail-section,
.detail-observation-tabs {
  margin-top: 1.5rem;
}

.detail-section {
  border-radius: 8px;
  background:
    radial-gradient(circle at 8% 0%, rgba(251, 140, 47, 0.08), transparent 30%),
    linear-gradient(120deg, rgba(255, 247, 235, 0.80), rgba(255, 250, 242, 0.88)),
    rgba(255, 247, 235, 0.80);
  padding: clamp(1rem, 2vw, 1.5rem);
  box-shadow:
    inset 0 0 0 1px rgba(251, 140, 47, 0.16),
    0 16px 42px rgba(124, 45, 18, 0.04);
}

.detail-section :deep(.rounded-2xl.bg-orange-100\/45),
.detail-section :deep(.rounded-xl.bg-orange-50\/80),
.detail-section :deep(.rounded-xl.bg-orange-50\/70),
.detail-section :deep(.rounded-lg.bg-orange-100),
.detail-section :deep(.rounded-lg.bg-orange-100\/35),
.detail-section :deep(.rounded-md.bg-orange-50),
.detail-section :deep(.rounded-md.bg-orange-100\/45),
.detail-section :deep(.rounded-xl.bg-orange-100\/45) {
  background-color: rgba(255, 250, 242, 0.70);
}

</style>
