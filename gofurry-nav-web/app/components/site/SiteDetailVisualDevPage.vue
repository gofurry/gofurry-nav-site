<template>
  <div ref="pageRoot" class="visual-dev-page min-h-full overflow-x-hidden text-slate-900">
    <div class="visual-sky" aria-hidden="true">
      <svg class="visual-grid" viewBox="0 0 1200 720" preserveAspectRatio="none">
        <defs>
          <linearGradient id="devLine" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#ff8a3d" stop-opacity="0.18" />
            <stop offset="52%" stop-color="#6a7cff" stop-opacity="0.16" />
            <stop offset="100%" stop-color="#15b8a6" stop-opacity="0.12" />
          </linearGradient>
          <radialGradient id="devGlow" cx="45%" cy="28%" r="58%">
            <stop offset="0%" stop-color="#ffd49a" stop-opacity="0.55" />
            <stop offset="52%" stop-color="#f7f1e8" stop-opacity="0.28" />
            <stop offset="100%" stop-color="#f7f1e8" stop-opacity="0" />
          </radialGradient>
        </defs>
        <rect width="1200" height="720" fill="url(#devGlow)" />
        <path class="flow-line flow-line-a" d="M-40 146 C190 62 322 270 520 176 C744 70 816 244 1240 88" />
        <path class="flow-line flow-line-b" d="M-70 530 C156 406 298 596 498 470 C742 316 918 500 1270 354" />
        <path class="flow-line flow-line-c" d="M20 340 C210 236 398 382 592 300 C780 220 986 244 1180 182" />
        <g class="node-field">
          <circle cx="144" cy="128" r="3.5" />
          <circle cx="330" cy="258" r="2.5" />
          <circle cx="536" cy="178" r="3" />
          <circle cx="784" cy="250" r="2.7" />
          <circle cx="1018" cy="180" r="3.2" />
          <circle cx="242" cy="494" r="2.8" />
          <circle cx="566" cy="462" r="3.2" />
          <circle cx="902" cy="420" r="2.8" />
        </g>
      </svg>
    </div>

    <div v-if="pending" class="relative flex min-h-[68vh] items-center justify-center text-slate-500">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="error" class="relative flex min-h-[68vh] items-center justify-center text-red-500">
      {{ loadFailedText }}
    </div>

    <main v-else class="relative mx-auto w-full max-w-[1560px] px-5 py-8 sm:px-8 lg:px-10">
      <section class="dev-hero">
        <div class="hero-orbit" aria-hidden="true">
          <svg viewBox="0 0 220 220">
            <circle class="orbit-ring orbit-ring-a" cx="110" cy="110" r="80" />
            <circle class="orbit-ring orbit-ring-b" cx="110" cy="110" r="55" />
            <path class="orbit-pulse" d="M42 132 C78 52 136 44 178 102 C142 74 94 94 42 132Z" />
          </svg>
        </div>

        <div class="relative z-10 flex flex-col gap-7 lg:flex-row lg:items-end lg:justify-between">
          <div class="flex min-w-0 flex-col gap-5 md:flex-row md:items-center">
            <div class="logo-shell">
              <img
                v-if="sitePageData.siteInfo?.icon"
                :src="logoUrl(sitePageData.siteInfo.icon)"
                :alt="siteName"
                class="h-20 w-20 rounded-2xl object-cover"
              >
              <div v-else class="flex h-20 w-20 items-center justify-center rounded-2xl bg-orange-100 text-2xl font-bold">
                GF
              </div>
            </div>

            <div class="min-w-0">
              <div class="mb-3 flex flex-wrap items-center gap-2">
                <span v-if="primaryEdgeLabel" class="dev-pill dev-pill-strong">{{ primaryEdgeLabel }}</span>
                <span v-if="sitePageData.siteInfo?.nsfw === '1'" class="dev-pill">NSFW</span>
                <span v-else class="dev-pill">SFW</span>
                <span v-if="sitePageData.siteInfo?.welfare === '1'" class="dev-pill">{{ label('公益网站', 'Nonprofit') }}</span>
              </div>
              <h1 class="break-words text-3xl font-black tracking-normal text-slate-950 md:text-4xl">
                {{ siteName }}
              </h1>
              <div class="mt-2 font-mono text-sm text-slate-600">{{ sitePageData.domain }}</div>
              <p class="mt-4 max-w-4xl text-sm leading-7 text-slate-700 md:text-base">
                {{ sitePageData.siteInfo?.info || '-' }}
              </p>
              <div v-if="overviewKeywords.length" class="mt-4 flex flex-wrap gap-2">
                <span
                  v-for="keyword in overviewKeywords"
                  :key="keyword"
                  class="rounded-full bg-white/62 px-3 py-1 text-xs font-semibold text-orange-700 ring-1 ring-orange-200/70"
                >
                  {{ keyword }}
                </span>
              </div>
            </div>
          </div>

          <div class="flex shrink-0 flex-col items-start gap-3 lg:items-end">
            <a
              v-if="visitUrl"
              :href="visitUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="visit-button duration-500"
            >
              <span class="fa fa-link" aria-hidden="true"></span>
              {{ label('访问网站', 'Visit site') }}
            </a>
            <div class="text-xs font-semibold text-orange-600">
              {{ label('浏览量', 'Views') }}: {{ sitePageData.siteInfo?.view_count ?? 0 }}
            </div>
          </div>
        </div>
      </section>

      <section class="signal-grid">
        <article
          v-for="card in signalCards"
          :key="card.title"
          class="signal-card duration-500"
          :class="card.tone"
        >
          <p class="text-xs font-semibold text-slate-500">{{ card.eyebrow }}</p>
          <div class="mt-2 flex items-end justify-between gap-3">
            <h2 class="text-3xl font-black text-slate-900">{{ card.value }}</h2>
            <span class="text-xs font-semibold text-slate-500">{{ card.badge }}</span>
          </div>
          <dl class="mt-5 space-y-2 text-xs">
            <div v-for="item in card.items" :key="item.label" class="flex justify-between gap-4">
              <dt class="text-slate-500">{{ item.label }}</dt>
              <dd class="min-w-0 truncate font-mono font-semibold text-slate-800">{{ item.value }}</dd>
            </div>
          </dl>
        </article>
      </section>

      <section class="dev-health-grid">
        <div class="dev-panel">
          <SiteHealthSummaryPanel
            :current-target="sitePageData.domain"
            :security-headers="securityHeaderItems"
            :site-summary="sitePageData.siteHealthSummary"
            :target-summary="sitePageData.targetHealthSummary"
          />
        </div>
      </section>

      <section class="dev-section dev-performance">
        <SitePerformance
          v-if="sitePageData.siteHttpRecord"
          :domain="sitePageData.domain"
          :http-record="sitePageData.siteHttpRecord"
          :ping-record="sitePageData.sitePingRecord"
          :site-id="siteId"
          :target-latest-core="sitePageData.targetLatestCore"
        />
      </section>

      <section class="dev-section">
        <SiteObservationTabs
          :dns-record="sitePageData.siteDnsRecord"
          :http-record="sitePageData.siteHttpRecord"
          :ping-record="sitePageData.sitePingRecord"
          :target-latest-core="sitePageData.targetLatestCore"
        />
      </section>

      <section class="dev-section">
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
import SiteHealthSummaryPanel from '@/components/site/SiteHealthSummaryPanel.vue'
import SiteMetadataProbePanel from '@/components/site/SiteMetadataProbePanel.vue'
import SiteObservationTabs from '@/components/site/SiteObservationTabs.vue'
import SitePerformance from '@/components/site/SitePerformance.vue'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'

const { locale, t } = useI18n()
const { data, pending, error, siteId } = await useSiteDetailPage()
const navV2Api = useApi('navV2')
const config = useRuntimeConfig()
const pageRoot = ref<HTMLElement | null>(null)
const sitePageData = computed(() => data.value!)
const siteName = computed(() => sitePageData.value.siteInfo?.name?.trim() || 'GoFurry')
const loadFailedText = computed(() => (t('common.loading') === 'Loading...' ? 'Failed to load site data.' : '站点数据加载失败。'))
const httpPayload = computed(() => asRecord(sitePageData.value.targetLatestCore?.protocols?.http?.payload))
const primaryEdgeLabel = computed(() => {
  const hint = sitePageData.value.targetHealthSummary?.edge_provider_hints?.[0]
  if (!hint) {
    return ''
  }
  return `${providerLabel(hint.provider)} · ${hint.hint_type?.toUpperCase() || 'EDGE'}`
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

function logoUrl(icon: string) {
  const prefix = String(config.public?.siteLogoPrefixUrl || '')
  if (/^https?:\/\//i.test(icon)) {
    return icon
  }
  return `${prefix}${icon}`
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
.visual-dev-page {
  --surface: rgba(255, 250, 242, 0.76);
  --surface-strong: rgba(255, 244, 226, 0.86);
  --ink-muted: #667085;
  --accent: #fb8c2f;
  position: relative;
  background:
    linear-gradient(180deg, rgba(255, 251, 245, 0.96) 0%, rgba(248, 240, 229, 0.96) 52%, rgba(255, 246, 234, 0.98) 100%);
}

.visual-sky {
  pointer-events: none;
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
}

.visual-grid {
  height: 100%;
  width: 100%;
  opacity: 0.9;
}

.flow-line {
  fill: none;
  stroke: url(#devLine);
  stroke-width: 2;
  stroke-dasharray: 12 18;
  animation: dash-flow 16s linear infinite;
}

.flow-line-b {
  animation-duration: 22s;
  animation-direction: reverse;
  opacity: 0.76;
}

.flow-line-c {
  animation-duration: 18s;
  opacity: 0.58;
}

.node-field circle {
  fill: rgba(251, 140, 47, 0.42);
  animation: node-breathe 4.8s ease-in-out infinite;
}

.node-field circle:nth-child(2n) {
  fill: rgba(79, 111, 237, 0.35);
  animation-delay: 1.4s;
}

.dev-hero {
  position: relative;
  overflow: hidden;
  border-radius: 32px;
  background:
    radial-gradient(circle at 88% 18%, rgba(106, 124, 255, 0.14), transparent 28%),
    radial-gradient(circle at 18% 24%, rgba(251, 140, 47, 0.20), transparent 26%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.74), rgba(255, 242, 219, 0.62));
  padding: clamp(1.5rem, 4vw, 3.5rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.18);
}

.hero-orbit {
  position: absolute;
  right: max(2rem, 6vw);
  top: 50%;
  width: min(28vw, 320px);
  transform: translateY(-50%);
  opacity: 0.42;
}

.hero-orbit svg {
  width: 100%;
}

.orbit-ring {
  fill: none;
  stroke: rgba(251, 140, 47, 0.42);
  stroke-width: 1.5;
  stroke-dasharray: 8 12;
  transform-origin: 110px 110px;
  animation: orbit-spin 28s linear infinite;
}

.orbit-ring-b {
  stroke: rgba(79, 111, 237, 0.36);
  animation-duration: 18s;
  animation-direction: reverse;
}

.orbit-pulse {
  fill: rgba(20, 184, 166, 0.16);
  stroke: rgba(20, 184, 166, 0.26);
  stroke-width: 1;
  animation: pulse-fade 5s ease-in-out infinite;
}

.logo-shell {
  border-radius: 24px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.72), rgba(255, 219, 170, 0.62));
  padding: 0.5rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.16);
}

.dev-pill {
  border-radius: 999px;
  background: rgba(255, 245, 224, 0.78);
  padding: 0.35rem 0.75rem;
  color: #9a4a08;
  font-size: 0.75rem;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.18);
}

.dev-pill-strong {
  background: rgba(255, 224, 183, 0.88);
  color: #7c2d12;
}

.visit-button {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 18px;
  background: #111827;
  padding: 0.85rem 1.1rem;
  color: #fffaf2;
  font-size: 0.875rem;
  font-weight: 800;
}

.visit-button:hover {
  background: #fb8c2f;
  color: #111827;
}

.signal-grid {
  margin-top: 1.5rem;
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 1rem;
}

@media (min-width: 760px) {
  .signal-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .signal-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.signal-card {
  min-height: 13rem;
  border-radius: 24px;
  padding: 1.35rem;
  background: var(--surface);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.7);
}

.signal-card:hover {
  background: rgba(255, 255, 255, 0.88);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.28), 0 0 0 6px rgba(251, 140, 47, 0.07);
}

.tone-green {
  border-left: 4px solid rgba(52, 211, 153, 0.86);
}

.tone-amber {
  border-left: 4px solid rgba(251, 191, 36, 0.9);
}

.tone-rose {
  border-left: 4px solid rgba(251, 113, 133, 0.9);
}

.dev-health-grid,
.dev-section {
  margin-top: 1.5rem;
}

.dev-panel,
.dev-section {
  border-radius: 28px;
  background: rgba(255, 247, 235, 0.72);
  padding: clamp(1rem, 2vw, 1.5rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.12);
}

.dev-health-grid :deep(.rounded-xl.bg-orange-50),
.dev-section :deep(.rounded-2xl.bg-orange-100\/45),
.dev-section :deep(.rounded-xl.bg-orange-50\/80),
.dev-section :deep(.rounded-xl.bg-orange-50\/70),
.dev-section :deep(.rounded-lg.bg-orange-100),
.dev-section :deep(.rounded-lg.bg-orange-100\/35),
.dev-section :deep(.rounded-md.bg-orange-50),
.dev-section :deep(.rounded-md.bg-orange-100\/45),
.dev-section :deep(.rounded-xl.bg-orange-100\/45) {
  background-color: rgba(255, 250, 242, 0.70);
}

.dev-performance :deep(section > .grid:first-child) {
  display: none;
}

.dev-performance :deep(section > .rounded-2xl) {
  background:
    radial-gradient(circle at 78% 0%, rgba(79, 111, 237, 0.09), transparent 34%),
    rgba(255, 247, 235, 0.56);
}

@keyframes dash-flow {
  to {
    stroke-dashoffset: -180;
  }
}

@keyframes orbit-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse-fade {
  0%,
  100% {
    opacity: 0.52;
  }
  50% {
    opacity: 0.88;
  }
}

@keyframes node-breathe {
  0%,
  100% {
    opacity: 0.35;
  }
  50% {
    opacity: 0.9;
  }
}
</style>
