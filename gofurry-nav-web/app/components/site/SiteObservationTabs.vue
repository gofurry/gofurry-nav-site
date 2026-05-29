<template>
  <section class="rounded-2xl bg-orange-100/45 p-5">
    <div class="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div>
          <div v-if="text.eyebrow" class="mb-1 text-xs font-medium uppercase tracking-wide text-orange-500">
          {{ text.eyebrow }}
        </div>
        <h3 class="text-lg font-semibold text-gray-900">{{ text.title }}</h3>
      </div>

      <div class="flex flex-wrap gap-1 rounded-xl bg-orange-50 p-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="rounded-lg px-4 py-2 text-sm transition-colors"
          :class="activeTab === tab.key ? 'bg-orange-200 text-gray-900' : 'text-gray-600 hover:bg-orange-100'"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <div class="rounded-xl bg-orange-50/70 p-4">
      <div v-if="activeTab === 'ping'" class="space-y-4">
        <MetricGrid :items="pingMetrics" />
      </div>

      <div v-else-if="activeTab === 'http'" class="space-y-5">
        <MetricGrid :items="httpMetrics" />
        <InfoList :items="httpHeaderItems" :empty-text="text.none" />
      </div>

      <div v-else-if="activeTab === 'tls'" class="space-y-4">
        <MetricGrid :items="tlsMetrics" />
      </div>

      <div v-else class="space-y-5">
        <MetricGrid :items="dnsMetrics" />
        <div v-if="dnsRecordGroups.length" class="space-y-4">
          <div v-for="group in dnsRecordGroups" :key="group.type">
            <h4 class="mb-2 text-sm font-semibold text-gray-500">{{ group.type }}</h4>
            <div class="overflow-hidden rounded-lg bg-orange-100">
              <div
                v-for="record in group.records"
                :key="record.key"
                class="grid grid-cols-[4.5rem_minmax(0,1fr)_4rem] gap-3 border-b border-orange-50 px-3 py-2 text-sm last:border-b-0"
              >
                <span class="font-semibold text-gray-500">{{ record.type }}</span>
                <span class="break-all font-mono text-gray-800">{{ record.value }}</span>
                <span class="text-right text-gray-500">{{ record.ttl }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="rounded-lg bg-orange-100 p-4 text-sm text-gray-500">{{ text.none }}</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, ref, type PropType } from 'vue'
import { i18n } from '@/main'
import type { DnsRecord, HttpRecord, PingRecord, TargetLatestResponse } from '@/types/nav'

type TabKey = 'ping' | 'http' | 'tls' | 'dns'
type MetricItem = { label: string; value: string; accent?: boolean }
type InfoItem = { label: string; value: string }

const props = defineProps<{
  pingRecord: PingRecord | null
  httpRecord: HttpRecord | null
  dnsRecord: DnsRecord | null
  targetLatestCore: TargetLatestResponse | null
}>()

const activeTab = ref<TabKey>('http')
const text = computed(() => ({
  eyebrow: '',
  title: label('观测详情', 'Observation details'),
  none: label('暂无数据', 'No data'),
}))
const tabs = computed<{ key: TabKey; label: string }[]>(() => [
  { key: 'ping', label: 'PING' },
  { key: 'http', label: 'HTTP' },
  { key: 'tls', label: 'TLS' },
  { key: 'dns', label: 'DNS' },
])

const pingPayload = computed(() => payload('ping'))
const httpPayload = computed(() => payload('http'))
const dnsPayload = computed(() => payload('dns'))

const pingMetrics = computed<MetricItem[]>(() => [
  { label: 'ICMP', value: stringValue(pingPayload.value.icmp_status), accent: true },
  { label: label('平均 RTT', 'Avg RTT'), value: msValue(pingPayload.value.avg_rtt_ms) },
  { label: label('最小 RTT', 'Min RTT'), value: msValue(pingPayload.value.min_rtt_ms) },
  { label: label('最大 RTT', 'Max RTT'), value: msValue(pingPayload.value.max_rtt_ms) },
  { label: label('抖动', 'Jitter'), value: msValue(pingPayload.value.jitter_ms) },
  { label: label('丢包率', 'Loss'), value: percentValue(pingPayload.value.loss_rate) },
  { label: label('数据包', 'Packets'), value: packetValue() },
  { label: label('解析 IP', 'Resolved IP'), value: stringValue(pingPayload.value.resolved_ip) },
])

const httpMetrics = computed<MetricItem[]>(() => [
  { label: label('状态码', 'Status'), value: numberValue(firstValue(httpPayload.value.status_code, props.httpRecord?.statusCode)), accent: true },
  { label: label('响应耗时', 'Response'), value: msValue(firstValue(httpPayload.value.response_time_ms, parseNumber(props.httpRecord?.responseTime))) },
  { label: label('DNS 查询', 'DNS Lookup'), value: msValue(httpPayload.value.dns_lookup_ms) },
  { label: label('TCP 连接', 'TCP Connect'), value: msValue(httpPayload.value.tcp_connect_ms) },
  { label: label('TLS 握手', 'TLS Handshake'), value: msValue(httpPayload.value.tls_handshake_ms) },
  { label: 'TTFB', value: msValue(httpPayload.value.ttfb_ms) },
  { label: label('传输耗时', 'Transfer'), value: msValue(httpPayload.value.transfer_ms) },
  { label: label('响应体', 'Body'), value: bytesValue(firstValue(httpPayload.value.body_read_bytes, props.httpRecord?.contentLength)) },
  { label: label('HTTP 协议', 'Protocol'), value: stringValue(httpPayload.value.http_protocol) },
  { label: label('远端 IP', 'Remote IP'), value: stringValue(httpPayload.value.remote_ip) },
  { label: 'Content-Type', value: firstString(httpPayload.value.content_type, headerValue('content-type')) },
  { label: label('最终 URL', 'Final URL'), value: firstString(httpPayload.value.final_url, props.httpRecord?.url) },
])

const tlsMetrics = computed<MetricItem[]>(() => [
  { label: label('证书已采集', 'Collected'), value: boolText(httpPayload.value.cert_collected) },
  { label: label('证书已校验', 'Verified'), value: boolText(httpPayload.value.cert_verified), accent: true },
  { label: label('握手状态', 'Handshake'), value: stringValue(httpPayload.value.tls_handshake) },
  { label: label('校验错误', 'Verify Error'), value: stringValue(httpPayload.value.verify_error) },
  { label: label('TLS 版本', 'TLS Version'), value: firstString(httpPayload.value.tls_version, props.httpRecord?.tlsVersion) },
  { label: label('密码套件', 'Cipher'), value: firstString(httpPayload.value.cipher_suite, props.httpRecord?.cipherSuite) },
  { label: label('生效时间', 'Not Before'), value: dateValue(httpPayload.value.cert_not_before) },
  { label: label('到期时间', 'Not After'), value: dateValue(firstValue(httpPayload.value.cert_not_after, httpPayload.value.cert_expiry, props.httpRecord?.certExpiry)) },
  { label: label('剩余天数', 'Days Left'), value: dayValue(firstValue(httpPayload.value.cert_days_left, parseNumber(props.httpRecord?.certDaysLeft))) },
  { label: label('签发者', 'Issuer'), value: firstString(httpPayload.value.cert_issuer_cn, httpPayload.value.cert_issuer, props.httpRecord?.certIssuer) },
  { label: 'SAN', value: numberValue(firstValue(httpPayload.value.cert_san_count, props.httpRecord?.certDNSNames?.length)) },
  { label: label('证书链长度', 'Chain Length'), value: numberValue(httpPayload.value.cert_chain_length) },
])

const dnsMetrics = computed<MetricItem[]>(() => [
  { label: 'A', value: recordCount('A') },
  { label: 'AAAA', value: recordCount('AAAA') },
  { label: label('CNAME 深度', 'CNAME Depth'), value: numberValue(dnsPayload.value.cname_chain_depth) },
  { label: label('CNAME 终点', 'CNAME Terminal'), value: stringValue(dnsPayload.value.cname_terminal) },
  { label: label('MX 主机', 'MX Hosts'), value: mxHostsValue() },
  { label: label('NS 主机', 'NS Hosts'), value: nsHostsValue() },
])
const dnsRecordGroups = computed(() => {
  const groups: { type: string; records: { key: string; type: string; value: string; ttl: string }[] }[] = []
  for (const type of ['A', 'AAAA', 'CNAME', 'MX', 'NS', 'TXT', 'CAA', 'SOA']) {
    const rows = dnsRows(type).slice(0, 8)
    if (!rows.length) {
      continue
    }
    groups.push({
      type,
      records: rows.map((row, index) => ({
        key: `${type}:${index}:${stringValue(row.value)}`,
        type: firstString(row.type, type),
        value: firstString(row.value, '-'),
        ttl: row.ttl === undefined ? '-' : `${row.ttl}s`,
      })),
    })
  }
  return groups
})

const httpHeaderItems = computed<InfoItem[]>(() => {
  const headers = normalizeHeaders(props.httpRecord?.headers)
  const preferred = ['server', 'content-type', 'cache-control', 'vary', 'etag', 'last-modified', 'alt-svc', 'x-powered-by']
  return preferred
    .map((key) => ({ label: key, value: headers[key]?.join(', ') ?? '' }))
    .filter((item) => item.value)
    .slice(0, 8)
})

const MetricGrid = defineComponent({
  props: {
    items: { type: Array as PropType<MetricItem[]>, default: () => [] },
  },
  setup(componentProps) {
    return () => h('div', { class: 'grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4' },
      componentProps.items.map(item => h('div', {
        class: [
          'rounded-lg p-3',
          item.accent ? 'bg-orange-100' : 'bg-orange-50',
        ],
      }, [
        h('div', { class: 'mb-1 text-[11px] font-semibold uppercase tracking-wide text-gray-500' }, item.label),
        h('div', { class: 'break-words text-base font-semibold text-gray-900' }, item.value || '-'),
      ]))
    )
  },
})

const InfoList = defineComponent({
  props: {
    items: { type: Array as PropType<InfoItem[]>, default: () => [] },
    emptyText: { type: String, default: '-' },
  },
  setup(componentProps) {
    return () => componentProps.items.length
      ? h('div', { class: 'grid grid-cols-1 gap-2 lg:grid-cols-2' },
        componentProps.items.map(item => h('div', { class: 'grid grid-cols-[9rem_minmax(0,1fr)] gap-3 rounded-lg bg-orange-100 px-3 py-2 text-sm' }, [
          h('span', { class: 'font-semibold text-gray-500' }, item.label),
          h('span', { class: 'break-words font-mono text-gray-800' }, item.value),
        ])))
      : h('div', { class: 'rounded-lg bg-orange-100 p-4 text-sm text-gray-500' }, componentProps.emptyText)
  },
})

function payload(protocol: string) {
  return asRecord(props.targetLatestCore?.protocols?.[protocol]?.payload)
}

function asRecord(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {}
}

function arrayValue(value: unknown): Record<string, any>[] {
  return Array.isArray(value) ? value.map(asRecord) : []
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value) ? value.map(String).filter(Boolean) : []
}

function firstValue(...values: unknown[]) {
  return values.find(value => value !== undefined && value !== null && value !== '')
}

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return '-'
}

function stringValue(value: unknown) {
  return firstString(value)
}

function numberValue(value: unknown) {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  return Number.isFinite(parsed) ? String(Math.round(parsed)) : '-'
}

function msValue(value: unknown) {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  return Number.isFinite(parsed) ? `${Math.round(parsed)}ms` : '-'
}

function dayValue(value: unknown) {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  return Number.isFinite(parsed) ? `${Math.round(parsed)}${label('天', 'd')}` : '-'
}

function percentValue(value: unknown) {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  if (!Number.isFinite(parsed)) {
    return '-'
  }
  return parsed <= 1 ? `${(parsed * 100).toFixed(2)}%` : `${parsed.toFixed(2)}%`
}

function bytesValue(value: unknown) {
  const bytes = typeof value === 'number' ? value : parseNumber(value)
  if (!Number.isFinite(bytes)) {
    return '-'
  }
  if (bytes < 1024) return `${Math.round(bytes)} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function boolText(value: unknown) {
  if (value === true) return label('是', 'Yes')
  if (value === false) return label('否', 'No')
  return '-'
}

function dateValue(value: unknown) {
  if (typeof value !== 'string' || !value) {
    return '-'
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
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

function packetValue() {
  const sent = numberValue(pingPayload.value.packets_sent)
  const recv = numberValue(pingPayload.value.packets_recv)
  return sent === '-' && recv === '-' ? '-' : `${recv}/${sent}`
}

function dnsRows(type: string) {
  const v2Rows = arrayValue(dnsPayload.value[type])
  if (v2Rows.length) {
    return v2Rows
  }
  const legacyKey: Record<string, keyof DnsRecord> = {
    A: 'a',
    AAAA: 'AAAA',
    CNAME: 'CNAME',
    MX: 'MX',
    NS: 'ns',
    TXT: 'txt',
    CAA: 'caa',
    SOA: 'SOA',
  }
  const key = legacyKey[type]
  return key ? arrayValue(props.dnsRecord?.[key]) : []
}

function recordCount(type: string) {
  return String(dnsRows(type).length)
}

function mxHostsValue() {
  const values = stringArray(dnsPayload.value.mx_hosts)
  if (values.length) {
    return values.join(', ')
  }
  return dnsRows('MX').map(row => firstString(row.value)).filter(value => value !== '-').join(', ') || '-'
}

function nsHostsValue() {
  const values = stringArray(dnsPayload.value.name_server_hosts)
  if (values.length) {
    return values.join(', ')
  }
  return dnsRows('NS').map(row => firstString(row.value)).filter(value => value !== '-').join(', ') || '-'
}

function normalizeHeaders(headers?: Record<string, string[]>) {
  const normalized: Record<string, string[]> = {}
  for (const [key, values] of Object.entries(headers ?? {})) {
    normalized[key.toLowerCase()] = Array.isArray(values) ? values.map(String) : [String(values)]
  }
  return normalized
}

function headerValue(key: string) {
  return normalizeHeaders(props.httpRecord?.headers)[key.toLowerCase()]?.[0] ?? ''
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>
