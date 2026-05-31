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
        <div class="relative z-10 flex min-w-0 flex-col gap-5 md:flex-row md:items-center">
          <div class="flex shrink-0 justify-center md:self-center">
            <div class="logo-shell">
              <img
                v-if="sitePageData.siteInfo?.icon"
                :src="logoUrl(sitePageData.siteInfo.icon)"
                :alt="siteName"
                class="h-20 w-20 rounded-lg object-contain"
              >
              <div v-else class="flex h-20 w-20 items-center justify-center rounded-lg text-2xl font-bold text-slate-700">
                GF
              </div>
            </div>
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex min-w-0 flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0 flex-1">
                <div class="hero-title-row">
                  <h1 class="mr-2 break-words text-xl font-black tracking-normal text-slate-950 md:text-2xl">
                    {{ siteName }}
                  </h1>
                  <div class="flex flex-wrap items-center gap-2.5">
                    <span
                      v-for="badge in heroBadges"
                      :key="badge.label"
                      class="dev-pill"
                      :class="badge.class"
                    >
                      {{ badge.label }}
                    </span>
                  </div>
                </div>

                <div class="relative mt-2 flex w-full flex-wrap items-center gap-x-5 gap-y-2 lg:w-fit">
                  <div
                    class="group/domain relative w-fit"
                    @pointerenter="openDomainCard"
                    @pointerleave="scheduleCloseDomainCard"
                    @focusin="openDomainCard"
                    @focusout="scheduleCloseDomainCard"
                  >
                    <button
                      type="button"
                      class="flex items-center font-mono text-sm text-slate-500 transition-colors duration-500 hover:text-orange-500"
                      @click="copyToClipboard(sitePageData.domain)"
                    >
                      <span>{{ sitePageData.domain }}</span>
                      <span class="ml-2 text-xs text-slate-400 opacity-0 transition-opacity duration-500 group-hover/domain:opacity-100">
                        {{ t('common.copy') }}
                      </span>
                    </button>

                    <transition name="domain-card">
                      <div
                        v-show="showDomainCard"
                        class="absolute left-0 top-full z-30 w-[min(22rem,calc(100vw-3rem))] pt-3"
                        @pointerenter="openDomainCard"
                        @pointerleave="scheduleCloseDomainCard"
                      >
                        <div class="absolute left-0 top-0 h-3 w-full" />
                        <div class="domain-popover">
                          <div class="mb-2 px-1 text-xs font-semibold text-orange-500">
                            {{ label('采集域名', 'Collected domains') }}
                          </div>
                          <div class="domain-list-scroll flex max-h-72 flex-col gap-1 overflow-y-auto pr-1">
                            <NuxtLink
                              v-for="domain in switchableDomains"
                              :key="domain"
                              :to="domainLink(domain)"
                              class="rounded-lg px-3 py-2 font-mono text-xs text-slate-700 transition-colors duration-500 hover:bg-orange-100/80 hover:text-orange-700"
                              :class="{ 'bg-orange-100/80 text-orange-700': domain === sitePageData.domain }"
                            >
                              {{ domain }}
                            </NuxtLink>
                          </div>
                        </div>
                      </div>
                    </transition>

                    <transition name="fade">
                      <div
                        v-if="copied"
                        class="absolute -top-7 left-0 rounded bg-slate-900 px-2 py-0.5 text-xs text-white"
                      >
                        {{ t('common.copied') }}
                      </div>
                    </transition>
                  </div>
                  <div class="text-xs text-orange-500 lg:hidden">
                    {{ label('浏览量', 'Views') }}: {{ sitePageData.siteInfo?.view_count ?? 0 }}
                  </div>
                </div>
              </div>

              <div class="hero-actions">
                <a
                  v-if="visitUrl"
                  :href="visitUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="visit-button duration-500"
                >
                  <img src="@/assets/svgs/go.svg" alt="" class="h-5 w-5 opacity-90">
                  {{ label('访问网站', 'Visit site') }}
                </a>
                <div class="hidden text-xs text-orange-500 lg:block">
                  {{ label('浏览量', 'Views') }}: {{ sitePageData.siteInfo?.view_count ?? 0 }}
                </div>
              </div>
            </div>

            <p class="max-w-6xl text-sm leading-7 text-slate-700 md:text-base">
              {{ sitePageData.siteInfo?.info || '-' }}
            </p>

            <div v-if="overviewKeywords.length" class="flex flex-wrap gap-2">
              <span
                v-for="keyword in overviewKeywords"
                :key="keyword"
                class="keyword-chip"
              >
                {{ keyword }}
              </span>
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

      <section class="dev-observation-overview">
        <div class="overview-strip">
          <div
            v-for="item in observationStripItems"
            :key="item.label"
            class="overview-stat"
            :class="item.tone"
          >
            <div class="text-[11px] text-slate-500">{{ item.label }}</div>
            <div class="truncate text-sm font-bold text-slate-900">{{ item.value }}</div>
          </div>
        </div>

        <div class="overview-detail-grid">
          <div class="protocol-rail-panel">
            <div class="mb-4 flex items-center justify-between gap-3">
              <div>
                <h3 class="text-lg font-black text-slate-950">{{ label('当前采集', 'Current Checks') }}</h3>
              </div>
              <div class="rounded-full bg-orange-100 px-3 py-1 text-xs text-orange-700">
                {{ protocolAvailabilityText }}
              </div>
            </div>

            <div class="protocol-rail">
              <article
                v-for="entry in protocolTrackEntries"
                :key="entry.protocol"
                class="protocol-node"
                :class="entry.tone"
              >
                <div class="min-w-0">
                  <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
                    <div class="flex items-center gap-2">
                      <span class="protocol-dot" />
                      <strong class="text-sm text-slate-950">{{ entry.label }}</strong>
                    </div>
                    <span class="font-mono text-[11px] text-slate-500">{{ entry.observedAt }}</span>
                  </div>
                  <div class="mt-2 grid gap-1 text-xs text-slate-600 sm:grid-cols-2">
                    <span>{{ label('耗时', 'Time') }}: <b>{{ entry.duration }}</b></span>
                    <span>{{ label('过期阈值', 'Stale') }}: <b>{{ entry.staleAfter }}</b></span>
                  </div>
                </div>
              </article>
            </div>

            <div class="signal-note mt-4 px-4 py-3 text-xs text-slate-600">
              <div class="mb-1 font-semibold text-slate-700">{{ label('观测信号', 'Signals') }}</div>
              <div v-if="targetRiskMessages.length" class="space-y-1">
                <div v-for="message in targetRiskMessages" :key="message">{{ message }}</div>
              </div>
              <div v-else>{{ label('暂无需要关注的信号', 'No notable signals') }}</div>
            </div>
          </div>

          <div class="security-matrix-panel">
            <div class="mb-4 flex items-center justify-between gap-3">
              <div>
                <h3 class="text-lg font-black text-slate-950">{{ label('安全响应头', 'Security Headers') }}</h3>
              </div>
              <div class="rounded-full bg-orange-100 px-3 py-1 text-xs text-orange-700">
                {{ securityHeaderRatio }}
              </div>
            </div>

            <div class="security-matrix">
              <div
                v-for="item in securityHeaderItems"
                :key="item.label"
                class="security-header-cell"
                :class="{ 'is-ok': item.ok }"
              >
                <span class="status-dot" />
                <span class="min-w-0 truncate">{{ item.label }}</span>
                <span class="ml-auto text-xs">{{ item.ok ? label('是', 'Yes') : label('否', 'No') }}</span>
              </div>
            </div>
          </div>
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
const heroBadges = computed(() => {
  const badges: { label: string; class: string }[] = []
  if (primaryEdgeLabel.value) {
    badges.push({ label: primaryEdgeLabel.value, class: 'dev-pill-edge' })
  }

  badges.push({
    label: sitePageData.value.siteInfo?.nsfw === '1' ? 'NSFW' : 'SFW',
    class: sitePageData.value.siteInfo?.nsfw === '1' ? 'dev-pill-risk' : 'dev-pill-sfw',
  })

  if (sitePageData.value.siteInfo?.welfare === '1') {
    badges.push({ label: label('公益网站', 'Nonprofit'), class: 'dev-pill-welfare' })
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
const copied = ref(false)
const showDomainCard = ref(false)
let domainCardCloseTimer: ReturnType<typeof setTimeout> | null = null

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

function copyToClipboard(text: string) {
  if (!text) {
    return
  }

  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    void navigator.clipboard.writeText(text)
  }
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 1800)
}

function openDomainCard() {
  if (switchableDomains.value.length <= 1) {
    return
  }

  if (domainCardCloseTimer) {
    clearTimeout(domainCardCloseTimer)
    domainCardCloseTimer = null
  }

  showDomainCard.value = true
}

function scheduleCloseDomainCard() {
  if (domainCardCloseTimer) {
    clearTimeout(domainCardCloseTimer)
  }

  domainCardCloseTimer = setTimeout(() => {
    showDomainCard.value = false
  }, 320)
}

function domainLink(domain: string) {
  return `/site/${encodeURIComponent(String(siteId.value))}/${encodeURIComponent(domain)}/dev`
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
  overflow: visible;
  border-radius: 0.5rem;
  background:
    radial-gradient(circle at 18% 24%, rgba(251, 140, 47, 0.20), transparent 26%),
    linear-gradient(135deg, rgba(255, 255, 255, 0.74), rgba(255, 242, 219, 0.62));
  padding: clamp(1rem, 2vw, 1.6rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.18);
}

.logo-shell {
  border-radius: 0.5rem;
  background: #ffedd5;
  padding: 0.45rem;
}

.hero-title-row {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.55rem;
}

@media (min-width: 640px) {
  .hero-title-row {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.65rem;
  }
}

.dev-pill {
  border-radius: 999px;
  padding: 0.25rem 0.7rem;
  font-size: 0.75rem;
  font-weight: 400;
  white-space: nowrap;
}

.dev-pill-edge {
  background: #fdba74;
  color: #9a3412;
}

.dev-pill-sfw {
  background: #fed7aa;
  color: #c2410c;
}

.dev-pill-risk {
  background: rgba(254, 226, 226, 0.80);
  color: #b91c1c;
}

.dev-pill-welfare {
  background: #fde68a;
  color: #b45309;
}

.keyword-chip {
  border-radius: 999px;
  background: #fed7aa;
  padding: 0.25rem 0.75rem;
  color: #9a3412;
  font-size: 0.75rem;
  font-weight: 400;
}

.hero-actions {
  display: flex;
  flex-shrink: 0;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.6rem;
}

@media (min-width: 1024px) {
  .hero-actions {
    align-items: flex-end;
    margin-left: auto;
    padding-left: 2rem;
  }
}

.domain-popover {
  border-radius: 1rem;
  background: rgba(255, 247, 237, 0.96);
  padding: 0.75rem;
  color: #1f2937;
  backdrop-filter: blur(14px);
}

.domain-card-enter-active,
.domain-card-leave-active {
  transition: opacity 180ms ease, transform 180ms ease;
}

.domain-card-enter-from,
.domain-card-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 160ms ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.domain-list-scroll {
  scrollbar-width: thin;
}

.visit-button {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.5rem;
  background: rgba(254, 215, 170, 0.62);
  padding: 0.55rem 1rem;
  color: #111827;
  font-size: 0.875rem;
  font-weight: 700;
}

.visit-button:hover {
  background: rgba(254, 215, 170, 0.88);
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

.dev-observation-overview,
.dev-section {
  margin-top: 1.5rem;
}

.dev-section {
  border-radius: 28px;
  background: rgba(255, 247, 235, 0.72);
  padding: clamp(1rem, 2vw, 1.5rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.12);
}

.dev-observation-overview {
  border-radius: 8px;
  background:
    radial-gradient(circle at 10% 0%, rgba(251, 140, 47, 0.10), transparent 36%),
    rgba(255, 247, 235, 0.70);
  padding: clamp(1rem, 2vw, 1.5rem);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.12);
}

.overview-strip {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

@media (min-width: 900px) {
  .overview-strip {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.overview-stat {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-radius: 0.55rem;
  background: rgba(255, 232, 196, 0.72);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.10);
  padding: 0.72rem 0.82rem;
}

.overview-stat.is-ok {
  background: rgba(220, 252, 231, 0.64);
  box-shadow: inset 0 0 0 1px rgba(16, 185, 129, 0.10);
}

.overview-stat.is-warn {
  background: rgba(253, 224, 71, 0.34);
  box-shadow: inset 0 0 0 1px rgba(245, 158, 11, 0.12);
}

.overview-detail-grid {
  margin-top: 0.9rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.9rem;
}

@media (min-width: 1180px) {
  .overview-detail-grid {
    grid-template-columns: minmax(0, 1.55fr) minmax(22rem, 0.9fr);
  }
}

.protocol-rail-panel,
.security-matrix-panel {
  border-radius: 0.75rem;
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.protocol-rail {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.65rem;
}

@media (min-width: 760px) {
  .protocol-rail {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.protocol-node {
  position: relative;
  border-radius: 0.65rem;
  background: rgba(255, 230, 191, 0.70);
  padding: 0.85rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.08);
}

.protocol-dot {
  height: 0.75rem;
  width: 0.75rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: #f59e0b;
}

.protocol-node.is-ok .protocol-dot,
.security-header-cell.is-ok .status-dot {
  background: #10b981;
}

.protocol-node.is-bad .protocol-dot {
  background: #f43f5e;
}

.security-matrix {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.6rem;
}

@media (min-width: 640px) {
  .security-matrix {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.security-header-cell {
  display: flex;
  min-height: 2.6rem;
  align-items: center;
  gap: 0.55rem;
  border-radius: 0.55rem;
  background: rgba(255, 230, 191, 0.68);
  padding: 0.65rem 0.8rem;
  font-size: 0.78rem;
  color: #475569;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.07);
}

.signal-note {
  border-radius: 0.55rem;
  background: rgba(255, 230, 191, 0.66);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.07);
}

.status-dot {
  height: 0.5rem;
  width: 0.5rem;
  flex-shrink: 0;
  border-radius: 999px;
  background: #cbd5e1;
}

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
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  background:
    linear-gradient(118deg, rgba(255, 237, 213, 0.55), rgba(255, 250, 242, 0.78) 48%, rgba(239, 246, 255, 0.32)),
    rgba(255, 247, 235, 0.70);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.12);
}

.dev-performance {
  background: transparent;
  padding: 0;
  box-shadow: none;
}

.dev-performance :deep(section > .rounded-2xl::before) {
  content: "";
  pointer-events: none;
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg, transparent, rgba(79, 111, 237, 0.08), transparent),
    repeating-linear-gradient(90deg, rgba(79, 111, 237, 0.055) 0 1px, transparent 1px 76px);
  mask-image: linear-gradient(to bottom, transparent 0, #000 18%, #000 82%, transparent 100%);
  opacity: 0.5;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child) {
  position: relative;
  z-index: 1;
  margin-bottom: 1rem;
  align-items: center;
  gap: 0.8rem;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:first-child) {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:first-child > div) {
  display: none;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:first-child h3) {
  color: #0f172a;
  font-size: clamp(1.08rem, 1.4vw, 1.28rem);
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:nth-child(2)) {
  align-self: center;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:nth-child(2) > div) {
  gap: 0.2rem;
  border-radius: 8px;
  background: rgba(255, 232, 196, 0.34);
  box-shadow: none;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child button) {
  border-radius: 7px;
  padding: 0.62rem 1rem;
  font-size: 0.88rem;
  transition-duration: 500ms;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child button.bg-orange-200) {
  background: #fdba74;
  color: #111827;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
  gap: 0.55rem;
}

.dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:last-child > div) {
  min-width: 4.8rem;
  border-radius: 7px;
  background: rgba(255, 232, 196, 0.38);
  box-shadow: none;
}

.dev-performance :deep(section > .rounded-2xl > .rounded-xl) {
  position: relative;
  z-index: 1;
  overflow: hidden;
  border-radius: 0;
  background:
    linear-gradient(rgba(255, 250, 242, 0.30), rgba(255, 250, 242, 0.30)),
    repeating-linear-gradient(0deg, transparent 0 39px, rgba(148, 163, 184, 0.13) 40px),
    repeating-linear-gradient(90deg, transparent 0 71px, rgba(79, 111, 237, 0.08) 72px);
  box-shadow: none;
}

.dev-performance :deep(section > .rounded-2xl.bg-orange-100\/45) {
  padding: clamp(1rem, 1.6vw, 1.25rem);
}

.dev-performance :deep(section > .rounded-2xl > .rounded-xl.bg-orange-50\/70) {
  margin-top: 0.2rem;
  padding: 0.5rem 0 0;
  background-color: transparent;
}

.dev-performance :deep(section > .rounded-2xl > .rounded-xl > div) {
  height: 320px;
}

@media (max-width: 1023px) {
  .dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
    justify-self: center;
  }

  .dev-performance :deep(section > .rounded-2xl > .rounded-xl > div) {
    height: 300px;
  }
}

@media (max-width: 640px) {
  .dev-performance :deep(section > .rounded-2xl) {
    padding: 1rem;
  }

  .dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:first-child) {
    text-align: center;
  }

  .dev-performance :deep(section > .rounded-2xl > .grid:first-child > div:last-child) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    justify-self: stretch;
  }

  .dev-performance :deep(section > .rounded-2xl > .rounded-xl) {
    padding-inline: 0.35rem;
  }

  .dev-performance :deep(section > .rounded-2xl > .rounded-xl > div) {
    height: 260px;
  }
}

@keyframes dash-flow {
  to {
    stroke-dashoffset: -180;
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
