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
        <div
          v-for="probe in lightProbeEntries"
          :key="probe.protocol"
          class="rounded-xl bg-orange-50/80 p-4"
        >
          <div class="mb-3 flex items-center justify-between gap-2">
            <span class="text-sm font-semibold text-gray-900">{{ protocolName(probe.protocol) }}</span>
            <span :class="['rounded-full px-2.5 py-1 text-xs font-semibold', statusClass(probe.status)]">
              {{ statusText(probe.status) }}
            </span>
          </div>
          <InfoList compact :items="probe.items" :empty-text="label('暂无数据', 'No data')" />
        </div>
      </div>
      <div v-else class="rounded-xl bg-orange-50/80 p-4 text-sm text-gray-500">{{ label('暂无数据', 'No data') }}</div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, type PropType } from 'vue'
import { i18n } from '@/main'
import type { HttpRecord, TargetLatestResponse } from '@/types/nav'

type InfoItem = { label: string; value: string | string[] }

const props = defineProps<{
  httpRecord: HttpRecord | null
  targetLatestCore: TargetLatestResponse | null
  lightProbeState: TargetLatestResponse | null
}>()

const httpPayload = computed(() => asRecord(props.targetLatestCore?.protocols?.http?.payload))
const httpStatus = computed(() => props.targetLatestCore?.protocols?.http?.status || '')
const lightProtocols = computed(() => props.lightProbeState?.protocols ?? {})
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
      items: lightProbeItems(protocol, envelope.payload),
    }
  }))

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

function boolText(value: unknown) {
  if (value === true) return label('是', 'Yes')
  if (value === false) return label('否', 'No')
  return '-'
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
