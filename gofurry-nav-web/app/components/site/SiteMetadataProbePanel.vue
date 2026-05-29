<template>
  <section class="grid grid-cols-1 gap-5 xl:grid-cols-[1fr_1.05fr]">
    <div class="rounded-2xl bg-orange-100/45 p-5">
      <div class="mb-4 flex items-center justify-between gap-3">
        <h3 class="text-lg font-semibold text-gray-900">{{ label('页面元信息', 'Page metadata') }}</h3>
        <span :class="['rounded-full px-2.5 py-1 text-xs font-semibold', statusClass(httpStatus)]">
          {{ statusText(httpStatus) }}
        </span>
      </div>
      <InfoList :items="pageInfoItems" :empty-text="label('暂无数据', 'No data')" />
    </div>

    <div class="rounded-2xl bg-orange-100/45 p-5">
      <h3 class="mb-4 text-lg font-semibold text-gray-900">{{ label('低频轻探测', 'Light probes') }}</h3>
      <div v-if="lightProbeEntries.length" class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <button
          v-for="probe in lightProbeEntries"
          :key="probe.protocol"
          type="button"
          class="group cursor-pointer rounded-xl bg-orange-50/80 p-4 text-left transition-[background-color,box-shadow,color] duration-500 hover:bg-orange-50 hover:ring-2 hover:ring-orange-300/55 focus:outline-none focus:ring-2 focus:ring-orange-300/70"
          @click="selectedProbe = probe"
        >
          <div class="mb-3 flex items-center justify-between gap-2">
            <span class="text-sm font-semibold text-gray-900">{{ protocolName(probe.protocol) }}</span>
            <span :class="['rounded-full px-2.5 py-1 text-xs font-semibold', statusClass(probe.status)]">
              {{ statusText(probe.status) }}
            </span>
          </div>
          <InfoList compact :items="probe.items" :empty-text="label('暂无数据', 'No data')" />
          <div class="mt-3 text-xs font-semibold text-orange-600 opacity-0 transition-opacity duration-500 group-hover:opacity-100 group-focus-visible:opacity-100">
            {{ label('点击查看详情', 'Click for details') }}
          </div>
        </button>
      </div>
      <div v-else class="rounded-xl bg-orange-50/80 p-4 text-sm text-gray-500">{{ label('暂无数据', 'No data') }}</div>
    </div>

    <Teleport to="body">
      <Transition name="probe-modal">
        <div
          v-if="selectedProbe"
          class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/45 px-4 py-6 backdrop-blur-sm"
          @click.self="selectedProbe = null"
        >
          <article
            class="flex max-h-[88vh] w-full max-w-4xl flex-col overflow-hidden rounded-2xl bg-orange-50 text-gray-900"
            role="dialog"
            aria-modal="true"
          >
            <header class="flex items-start justify-between gap-4 bg-orange-100/70 px-5 py-4">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wide text-orange-600">{{ label('低频轻探测详情', 'Light probe detail') }}</p>
                <h3 class="mt-1 text-xl font-semibold">{{ protocolName(selectedProbe.protocol) }}</h3>
              </div>
              <div class="flex items-center gap-3">
                <span :class="['rounded-full px-2.5 py-1 text-xs font-semibold', statusClass(selectedProbe.status)]">
                  {{ statusText(selectedProbe.status) }}
                </span>
                <button
                  type="button"
                  class="rounded-full bg-orange-50 px-3 py-1.5 text-sm text-gray-700 transition-colors hover:bg-orange-200"
                  @click="selectedProbe = null"
                >
                  {{ label('关闭', 'Close') }}
                </button>
              </div>
            </header>

            <div class="overflow-y-auto px-5 py-5">
              <div class="mb-4 grid grid-cols-1 gap-3 md:grid-cols-3">
                <div class="rounded-xl bg-orange-100/45 p-3">
                  <p class="text-xs font-semibold text-gray-500">{{ label('观测时间', 'Observed') }}</p>
                  <p class="mt-1 break-words font-mono text-sm">{{ formatTime(selectedProbe.observedAt) }}</p>
                </div>
                <div class="rounded-xl bg-orange-100/45 p-3">
                  <p class="text-xs font-semibold text-gray-500">{{ label('耗时', 'Duration') }}</p>
                  <p class="mt-1 font-mono text-sm">{{ formatDuration(selectedProbe.durationMs) }}</p>
                </div>
                <div class="rounded-xl bg-orange-100/45 p-3">
                  <p class="text-xs font-semibold text-gray-500">{{ label('结果', 'Result') }}</p>
                  <p class="mt-1 font-mono text-sm">{{ selectedProbe.errorCode || statusText(selectedProbe.status) }}</p>
                </div>
              </div>

              <div v-if="selectedProbe.errorMessage" class="mb-4 rounded-xl bg-red-50 p-3 text-sm text-red-800">
                {{ selectedProbe.errorMessage }}
              </div>

              <div class="space-y-4">
                <section
                  v-for="section in selectedProbeDetailSections"
                  :key="section.title"
                  class="rounded-xl bg-orange-100/35 p-4"
                >
                  <h4 class="mb-3 text-sm font-semibold text-gray-900">{{ section.title }}</h4>
                  <InfoList :items="section.items" :empty-text="label('暂无数据', 'No data')" />
                </section>
              </div>
            </div>
          </article>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, type PropType } from 'vue'
import { i18n } from '@/main'
import type { HttpRecord, TargetLatestResponse } from '@/types/nav'

type InfoItem = { label: string; value: string | string[] }
type DetailSection = { title: string; items: InfoItem[] }
type LightProbeEntry = {
  protocol: string
  status: string
  payload: unknown
  observedAt: string
  durationMs: number
  errorCode: string
  errorMessage: string
  items: InfoItem[]
}

const props = defineProps<{
  httpRecord: HttpRecord | null
  targetLatestCore: TargetLatestResponse | null
  lightProbeState: TargetLatestResponse | null
}>()

const httpPayload = computed(() => asRecord(props.targetLatestCore?.protocols?.http?.payload))
const httpStatus = computed(() => props.targetLatestCore?.protocols?.http?.status || '')
const lightProtocols = computed(() => props.lightProbeState?.protocols ?? {})
const selectedProbe = ref<LightProbeEntry | null>(null)
const pageInfoItems = computed<InfoItem[]>(() => {
  const meta = asRecord(httpPayload.value.meta)
  const openGraph = asRecord(httpPayload.value.open_graph || httpPayload.value.openGraph)
  const twitterCard = asRecord(httpPayload.value.twitter_card || httpPayload.value.twitterCard)
  const cachePolicy = asRecord(httpPayload.value.cache_policy)
  return compactItems([
    { label: 'Title', value: firstString(httpPayload.value.title, props.httpRecord?.title) },
    { label: 'Description', value: firstString(meta.description, props.httpRecord?.meta?.description) },
    { label: 'Keywords', value: firstString(meta.keywords, props.httpRecord?.meta?.keywords) },
    { label: 'Application', value: stringValue(meta.application_name) },
    { label: 'Theme Color', value: stringValue(meta.theme_color) },
    { label: 'Robots', value: stringValue(meta.robots || httpPayload.value.robots_meta_policy) },
    { label: 'Viewport', value: stringValue(meta.viewport) },
    { label: 'Canonical', value: stringValue(httpPayload.value.canonical_url || httpPayload.value.canonicalUrl) },
    { label: 'HTML Lang', value: stringValue(httpPayload.value.html_lang || httpPayload.value.htmlLang) },
    { label: 'OpenGraph', value: firstString(openGraph.title, openGraph.type) },
    { label: 'Twitter Card', value: stringValue(twitterCard.card) },
    { label: 'Cache-Control', value: firstString(cachePolicy.cache_control, headerValue('cache-control')) },
  ])
})

const lightProbeOrder = ['rdap', 'robots', 'security_txt', 'page_assets', 'port_check', 'waf_canary']
const lightProbeEntries = computed(() => lightProbeOrder
  .flatMap((protocol) => {
    const envelope = lightProtocols.value[protocol]
    if (!envelope) {
      return []
    }

    return {
      protocol,
      status: envelope.status,
      payload: envelope.payload,
      observedAt: envelope.observed_at,
      durationMs: envelope.duration_ms,
      errorCode: envelope.error_code || '',
      errorMessage: envelope.error_message || '',
      items: lightProbeItems(protocol, envelope.payload),
    }
  }))
const selectedProbeDetailSections = computed(() => selectedProbe.value ? lightProbeDetailSections(selectedProbe.value) : [])

const InfoList = defineComponent({
  props: {
    items: { type: Array as PropType<InfoItem[]>, default: () => [] },
    emptyText: { type: String, default: '-' },
    compact: { type: Boolean, default: false },
  },
  setup(componentProps) {
    return () => componentProps.items.length
      ? h('div', { class: componentProps.compact ? 'space-y-1.5' : 'space-y-2.5' },
        componentProps.items.map(item => h('div', { class: 'grid grid-cols-[8rem_minmax(0,1fr)] gap-3 text-sm' }, [
          h('span', { class: 'font-semibold text-gray-500' }, item.label),
          h('span', { class: 'break-words font-mono text-gray-800' }, Array.isArray(item.value) ? item.value.join(', ') : item.value),
        ])))
      : h('div', { class: 'text-sm text-gray-500' }, componentProps.emptyText)
  },
})

function lightProbeItems(protocol: string, rawPayload: unknown): InfoItem[] {
  const payload = asRecord(rawPayload)
  switch (protocol) {
    case 'rdap':
      return compactItems([
        { label: 'Domain', value: stringValue(payload.registrable_domain) },
        { label: 'Registrar', value: stringValue(payload.registrar) },
        { label: 'Expires', value: stringValue(payload.expires_at) },
        { label: 'Statuses', value: stringArray(payload.statuses).slice(0, 4) },
        { label: 'Nameservers', value: stringArray(payload.nameservers).slice(0, 6) },
      ])
    case 'robots':
      return compactItems([
        { label: 'Exists', value: boolText(payload.exists) },
        { label: 'Status', value: numberValue(payload.status_code) },
        { label: 'Sitemaps', value: numberValue(payload.sitemap_count) },
        { label: 'Disallow All', value: boolText(payload.global_disallow_all) },
      ])
    case 'security_txt':
      return compactItems([
        { label: 'Exists', value: boolText(payload.exists) },
        { label: 'Path', value: stringValue(payload.path_used) },
        { label: 'Contact', value: stringArray(payload.contact).slice(0, 2) },
        { label: 'Expires', value: stringValue(payload.expires) },
      ])
    case 'page_assets': {
      const icon = asRecord(payload.icon)
      const manifest = asRecord(payload.manifest)
      return compactItems([
        { label: 'Icon', value: boolText(icon.exists) },
        { label: 'Icon Type', value: stringValue(icon.content_type) },
        { label: 'Manifest', value: boolText(manifest.exists) },
        { label: 'Manifest Name', value: stringValue(manifest.name || manifest.short_name) },
      ])
    }
    case 'port_check':
      return compactItems([
        { label: 'Ports', value: `${numberValue(payload.ports_checked)}/${numberValue(payload.ports_configured)}` },
        { label: 'Open', value: numberValue(payload.open_count) },
        { label: 'Closed', value: numberValue(payload.closed_count) },
        { label: 'Timeout', value: numberValue(payload.timeout_count) },
      ])
    case 'waf_canary':
      return compactItems([
        { label: 'Cases', value: `${numberValue(payload.cases_executed)}/${numberValue(payload.cases_total)}` },
        { label: 'Blocked', value: numberValue(payload.blocked_count) },
        { label: 'Matched', value: numberValue(payload.expected_blocked_matched_count) },
        { label: 'Unexpected Pass', value: numberValue(payload.unexpected_pass_count) },
      ])
    default:
      return []
  }
}

function lightProbeDetailSections(probe: LightProbeEntry): DetailSection[] {
  const payload = asRecord(probe.payload)
  switch (probe.protocol) {
    case 'rdap':
      return compactSections([
        {
          title: 'RDAP',
          items: compactItems([
            { label: '注册域', value: stringValue(payload.registrable_domain) },
            { label: 'RDAP 服务', value: stringValue(payload.rdap_server) },
            { label: '状态码', value: numberValue(payload.status_code) },
            { label: '注册商', value: stringValue(payload.registrar) },
            { label: '到期时间', value: stringValue(payload.expires_at) },
            { label: '域名状态', value: stringArray(payload.statuses) },
            { label: 'NameServer', value: stringArray(payload.nameservers) },
            { label: 'DNSSEC 签名', value: boolText(payload.dnssec_delegation_signed) },
            { label: '原文截断', value: boolText(payload.raw_truncated) },
          ]),
        },
        {
          title: 'RDAP 事件',
          items: mapRecordItems(asRecord(payload.events_summary)),
        },
      ])
    case 'robots':
      return compactSections([
        {
          title: 'robots.txt',
          items: compactItems([
            { label: '存在', value: boolText(payload.exists) },
            { label: '状态码', value: numberValue(payload.status_code) },
            { label: 'Content-Type', value: stringValue(payload.content_type) },
            { label: 'Sitemap 数量', value: numberValue(payload.sitemap_count) },
            { label: 'User-agent:*', value: boolText(payload.user_agent_star_present) },
            { label: '全站 Disallow', value: boolText(payload.global_disallow_all) },
            { label: '响应截断', value: boolText(payload.body_truncated) },
          ]),
        },
        {
          title: 'Sitemap',
          items: stringArray(payload.sitemaps).map((value, index) => ({ label: `#${index + 1}`, value })),
        },
      ])
    case 'security_txt':
      return compactSections([
        {
          title: 'security.txt',
          items: compactItems([
            { label: '存在', value: boolText(payload.exists) },
            { label: '路径', value: stringValue(payload.path_used) },
            { label: '状态码', value: numberValue(payload.status_code) },
            { label: 'Content-Type', value: stringValue(payload.content_type) },
            { label: 'Contact', value: stringArray(payload.contact) },
            { label: 'Expires', value: stringValue(payload.expires) },
            { label: 'Policy', value: stringValue(payload.policy) },
            { label: '语言', value: stringArray(payload.preferred_languages) },
            { label: 'Canonical', value: stringValue(payload.canonical) },
            { label: '响应截断', value: boolText(payload.body_truncated) },
          ]),
        },
      ])
    case 'page_assets': {
      const icon = asRecord(payload.icon)
      const manifest = asRecord(payload.manifest)
      return compactSections([
        {
          title: 'Favicon',
          items: compactItems([
            { label: '存在', value: boolText(icon.exists) },
            { label: 'URL', value: stringValue(icon.source_url) },
            { label: 'Rel', value: stringValue(icon.selected_rel) },
            { label: 'Sizes', value: stringValue(icon.selected_sizes) },
            { label: '状态码', value: numberValue(icon.status_code) },
            { label: 'Content-Type', value: stringValue(icon.content_type) },
            { label: '读取字节', value: bytesText(icon.body_read_bytes) },
            { label: '响应截断', value: boolText(icon.body_truncated) },
            { label: 'SHA256', value: stringValue(icon.sha256) },
            { label: '跳过原因', value: stringValue(icon.skipped_reason) },
          ]),
        },
        {
          title: 'Manifest',
          items: compactItems([
            { label: '存在', value: boolText(manifest.exists) },
            { label: 'URL', value: stringValue(manifest.source_url) },
            { label: '状态码', value: numberValue(manifest.status_code) },
            { label: 'Content-Type', value: stringValue(manifest.content_type) },
            { label: '读取字节', value: bytesText(manifest.body_read_bytes) },
            { label: '响应截断', value: boolText(manifest.body_truncated) },
            { label: '名称', value: stringValue(manifest.name) },
            { label: '短名称', value: stringValue(manifest.short_name) },
            { label: '主题色', value: stringValue(manifest.theme_color) },
            { label: '背景色', value: stringValue(manifest.background_color) },
            { label: '显示模式', value: stringValue(manifest.display) },
            { label: 'Start URL', value: stringValue(manifest.start_url) },
            { label: 'Scope', value: stringValue(manifest.scope) },
            { label: '图标数量', value: numberValue(manifest.icons_count) },
            { label: 'SHA256', value: stringValue(manifest.sha256) },
            { label: '解析错误', value: stringValue(manifest.manifest_decode_error) },
          ]),
        },
      ])
    }
    case 'port_check':
      return compactSections([
        {
          title: '端口概览',
          items: compactItems([
            { label: '配置端口', value: numberValue(payload.ports_configured) },
            { label: '检查端口', value: numberValue(payload.ports_checked) },
            { label: '开放', value: numberValue(payload.open_count) },
            { label: '关闭', value: numberValue(payload.closed_count) },
            { label: '超时', value: numberValue(payload.timeout_count) },
            { label: '疑似过滤', value: numberValue(payload.filtered_suspected_count) },
            { label: '跳过', value: numberValue(payload.skipped_count) },
            { label: '非法端口', value: numberValue(payload.invalid_port_count) },
            { label: '重复端口', value: numberValue(payload.duplicate_port_count) },
            { label: '端口截断', value: boolText(payload.truncated) },
          ]),
        },
        {
          title: '端口明细',
          items: arrayRecords(payload.results).map(portResultItem),
        },
      ])
    case 'waf_canary':
      return compactSections([
        {
          title: 'WAF Canary 概览',
          items: compactItems([
            { label: '路径', value: stringValue(payload.canary_path) },
            { label: '用例', value: `${numberValue(payload.cases_executed)}/${numberValue(payload.cases_total)}` },
            { label: '预期拦截', value: numberValue(payload.expected_blocked_count) },
            { label: '匹配预期', value: numberValue(payload.expected_blocked_matched_count) },
            { label: '实际拦截', value: numberValue(payload.blocked_count) },
            { label: '意外放行', value: numberValue(payload.unexpected_pass_count) },
            { label: '网络错误', value: numberValue(payload.network_error_count) },
            { label: '状态码异常', value: numberValue(payload.status_code_unexpected_count) },
            { label: '目标截断', value: boolText(payload.target_run_truncated) },
          ]),
        },
        {
          title: '用例明细',
          items: arrayRecords(payload.cases).map(wafCaseItem),
        },
      ])
    default:
      return [{ title: protocolName(probe.protocol), items: mapRecordItems(payload) }]
  }
}

function compactSections(sections: DetailSection[]) {
  return sections.filter(section => section.items.length)
}

function compactItems(items: InfoItem[]) {
  return items.filter(item => Array.isArray(item.value) ? item.value.length : item.value && item.value !== '-')
}

function protocolName(protocol: string) {
  const map: Record<string, string> = {
    rdap: 'RDAP',
    robots: 'robots.txt',
    security_txt: 'security.txt',
    page_assets: 'Page assets',
    port_check: 'Port check',
    waf_canary: 'WAF canary',
  }
  return map[protocol] || protocol
}

function statusText(status: string) {
  if (status === 'success') return label('成功', 'Success')
  if (status === 'failure') return label('失败', 'Failure')
  if (status === 'skipped') return label('跳过', 'Skipped')
  return status || '-'
}

function statusClass(status: string) {
  if (status === 'success') return 'bg-green-100 text-green-800'
  if (status === 'failure') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-700'
}

function mapRecordItems(record: Record<string, any>): InfoItem[] {
  return compactItems(Object.entries(record).map(([key, value]) => ({
    label: key,
    value: displayValue(value),
  })))
}

function arrayRecords(value: unknown): Record<string, any>[] {
  return Array.isArray(value)
    ? value.filter(item => item && typeof item === 'object' && !Array.isArray(item)).map(item => item as Record<string, any>)
    : []
}

function portResultItem(result: Record<string, any>): InfoItem {
  const port = numberValue(result.port)
  const service = stringValue(result.service_hint)
  const status = stringValue(result.status)
  const duration = formatDuration(numberFromAny(result.duration_ms))
  const error = stringValue(result.error_code || result.error_message)
  return {
    label: service && service !== '-' ? `${port} ${service}` : port,
    value: [status, duration, error].filter(value => value && value !== '-').join(' / ') || '-',
  }
}

function wafCaseItem(result: Record<string, any>): InfoItem {
  const method = stringValue(result.method)
  const status = numberValue(result.status_code)
  const blocked = boolText(result.blocked)
  const matched = boolText(result.matched_expected)
  const duration = formatDuration(numberFromAny(result.duration_ms))
  return {
    label: stringValue(result.case_id),
    value: [
      stringValue(result.category),
      method !== '-' ? method : '',
      status !== '-' ? `HTTP ${status}` : '',
      `${label('拦截', 'Blocked')}: ${blocked}`,
      `${label('命中预期', 'Expected')}: ${matched}`,
      duration,
    ].filter(Boolean).join(' / '),
  }
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.map(String).filter(Boolean) : []
}

function compactString(value: unknown) {
  return typeof value === 'string' ? value.trim() : ''
}

function firstString(...values: unknown[]) {
  for (const value of values) {
    const text = compactString(value)
    if (text) return text
  }
  return '-'
}

function stringValue(value: unknown) {
  return firstString(value)
}

function numberValue(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) return String(Math.round(value))
  if (typeof value === 'string' && value.trim()) return value
  return '-'
}

function numberFromAny(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim()) {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }
  return 0
}

function bytesText(value: unknown) {
  const bytes = numberFromAny(value)
  if (!bytes) return '-'
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${bytes} B`
}

function boolText(value: unknown) {
  if (value === true) return label('是', 'Yes')
  if (value === false) return label('否', 'No')
  return '-'
}

function displayValue(value: unknown): string | string[] {
  if (Array.isArray(value)) {
    return value.map(item => displayValue(item)).flat().map(String).filter(Boolean)
  }
  if (value && typeof value === 'object') {
    return JSON.stringify(value)
  }
  if (typeof value === 'boolean') {
    return boolText(value)
  }
  if (typeof value === 'number') {
    return numberValue(value)
  }
  return stringValue(value)
}

function formatTime(value: string) {
  if (!value) return '-'
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
}

function formatDuration(value: number) {
  return value > 0 ? `${Math.round(value)}ms` : '-'
}

function headerValue(key: string) {
  const normalizedKey = key.toLowerCase()
  for (const [name, values] of Object.entries(props.httpRecord?.headers ?? {})) {
    if (name.toLowerCase() === normalizedKey && values?.length) {
      return values[0]
    }
  }
  return ''
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.probe-modal-enter-active,
.probe-modal-leave-active {
  transition: opacity 160ms ease;
}

.probe-modal-enter-active article,
.probe-modal-leave-active article {
  transition: transform 160ms ease;
}

.probe-modal-enter-from,
.probe-modal-leave-to {
  opacity: 0;
}

.probe-modal-enter-from article,
.probe-modal-leave-to article {
  transform: translateY(8px) scale(0.98);
}
</style>
