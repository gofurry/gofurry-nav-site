<template>
  <section class="site-observation-tabs">
    <div class="observation-header">
      <div>
        <div v-if="text.eyebrow" class="observation-eyebrow">
          {{ text.eyebrow }}
        </div>
        <h3 class="observation-title">{{ text.title }}</h3>
      </div>

      <div class="observation-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="observation-tab"
          :class="{ 'is-active': activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <div class="observation-content">
      <div v-if="activeTab === 'ping'" class="space-y-4">
        <SiteObservationMetricGrid :items="pingMetrics" />
      </div>

      <div v-else-if="activeTab === 'http'" class="space-y-5">
        <SiteObservationMetricGrid :items="httpMetrics" />
        <SiteObservationInfoList :items="httpHeaderItems" :empty-text="text.none" />
      </div>

      <div v-else-if="activeTab === 'tls'" class="space-y-4">
        <SiteObservationMetricGrid :items="tlsMetrics" />
      </div>

      <div v-else class="space-y-5">
        <SiteObservationMetricGrid :items="dnsMetrics" />
        <div v-if="dnsRecordGroups.length" class="space-y-4">
          <div v-for="group in dnsRecordGroups" :key="group.type" class="dns-record-card">
            <h4 class="record-heading">{{ group.type }}</h4>
            <div class="record-list">
              <div
                v-for="record in group.records"
                :key="record.key"
                class="record-row"
              >
                <span class="font-semibold text-gray-500 dark:text-slate-400">{{ record.type }}</span>
                <span class="break-all font-mono text-gray-800 dark:text-slate-200">{{ record.value }}</span>
                <span class="text-right text-gray-500 dark:text-slate-400">{{ record.ttl }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">{{ text.none }}</div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { i18n } from '@/main'
import SiteObservationInfoList from '@/components/site/SiteObservationInfoList.vue'
import SiteObservationMetricGrid from '@/components/site/SiteObservationMetricGrid.vue'
import type { DnsRecord, HttpRecord, PingRecord, TargetLatestResponse } from '@/types/nav'
import type { ObservationInfoItem, ObservationMetricItem, ObservationTone } from './detailTypes'

type TabKey = 'ping' | 'http' | 'tls' | 'dns'
type MetricItem = ObservationMetricItem
type InfoItem = ObservationInfoItem

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
  { label: 'ICMP', value: stringValue(pingPayload.value.icmp_status), tone: pingPayload.value.icmp_status === 'reachable' ? 'good' : 'warn' },
  { label: label('平均 RTT', 'Avg RTT'), value: msValue(pingPayload.value.avg_rtt_ms), tone: latencyTone(pingPayload.value.avg_rtt_ms) },
  { label: label('最小 RTT', 'Min RTT'), value: msValue(pingPayload.value.min_rtt_ms), tone: latencyTone(pingPayload.value.min_rtt_ms) },
  { label: label('最大 RTT', 'Max RTT'), value: msValue(pingPayload.value.max_rtt_ms), tone: latencyTone(pingPayload.value.max_rtt_ms) },
  { label: label('抖动', 'Jitter'), value: msValue(pingPayload.value.jitter_ms), tone: latencyTone(pingPayload.value.jitter_ms, 30, 80) },
  { label: label('丢包率', 'Loss'), value: percentValue(pingPayload.value.loss_rate), tone: lossTone(pingPayload.value.loss_rate) },
  { label: label('数据包', 'Packets'), value: packetValue(), tone: 'normal' },
  { label: label('解析 IP', 'Resolved IP'), value: stringValue(pingPayload.value.resolved_ip), tone: 'normal' },
])

const httpMetrics = computed<MetricItem[]>(() => [
  { label: label('状态码', 'Status'), value: numberValue(firstValue(httpPayload.value.status_code, props.httpRecord?.statusCode)), tone: statusTone(firstValue(httpPayload.value.status_code, props.httpRecord?.statusCode)) },
  { label: label('响应耗时', 'Response'), value: msValue(firstValue(httpPayload.value.response_time_ms, parseNumber(props.httpRecord?.responseTime))), tone: latencyTone(firstValue(httpPayload.value.response_time_ms, parseNumber(props.httpRecord?.responseTime))) },
  { label: label('DNS 查询', 'DNS Lookup'), value: msValue(httpPayload.value.dns_lookup_ms), tone: latencyTone(httpPayload.value.dns_lookup_ms, 200, 1000) },
  { label: label('TCP 连接', 'TCP Connect'), value: msValue(httpPayload.value.tcp_connect_ms), tone: latencyTone(httpPayload.value.tcp_connect_ms, 300, 1200) },
  { label: label('TLS 握手', 'TLS Handshake'), value: msValue(httpPayload.value.tls_handshake_ms), tone: latencyTone(httpPayload.value.tls_handshake_ms, 400, 1500) },
  { label: 'TTFB', value: msValue(httpPayload.value.ttfb_ms), tone: latencyTone(httpPayload.value.ttfb_ms) },
  { label: label('传输耗时', 'Transfer'), value: msValue(httpPayload.value.transfer_ms), tone: latencyTone(httpPayload.value.transfer_ms) },
  { label: label('响应体', 'Body'), value: bytesValue(firstValue(httpPayload.value.body_read_bytes, props.httpRecord?.contentLength)), tone: 'normal' },
  { label: label('HTTP 协议', 'Protocol'), value: stringValue(httpPayload.value.http_protocol), tone: 'normal' },
  { label: label('远端 IP', 'Remote IP'), value: stringValue(httpPayload.value.remote_ip), tone: 'normal' },
  { label: 'Content-Type', value: firstString(httpPayload.value.content_type, headerValue('content-type')), tone: 'good' },
  { label: label('最终 URL', 'Final URL'), value: firstString(httpPayload.value.final_url, props.httpRecord?.url), tone: 'normal' },
])

const tlsMetrics = computed<MetricItem[]>(() => [
  { label: label('证书已采集', 'Collected'), value: boolText(httpPayload.value.cert_collected), tone: httpPayload.value.cert_collected === true ? 'good' : 'warn' },
  { label: label('证书已校验', 'Verified'), value: boolText(httpPayload.value.cert_verified), tone: httpPayload.value.cert_verified === true ? 'good' : 'warn' },
  { label: label('握手状态', 'Handshake'), value: stringValue(httpPayload.value.tls_handshake), tone: httpPayload.value.tls_handshake === 'collected' ? 'good' : 'normal' },
  { label: label('校验错误', 'Verify Error'), value: stringValue(httpPayload.value.verify_error), tone: httpPayload.value.verify_error ? 'warn' : 'good' },
  { label: label('TLS 版本', 'TLS Version'), value: firstString(httpPayload.value.tls_version, props.httpRecord?.tlsVersion), tone: 'good' },
  { label: label('密码套件', 'Cipher'), value: firstString(httpPayload.value.cipher_suite, props.httpRecord?.cipherSuite), tone: 'normal' },
  { label: label('生效时间', 'Not Before'), value: dateValue(httpPayload.value.cert_not_before), tone: 'normal' },
  { label: label('到期时间', 'Not After'), value: dateValue(firstValue(httpPayload.value.cert_not_after, httpPayload.value.cert_expiry, props.httpRecord?.certExpiry)), tone: 'normal' },
  { label: label('剩余天数', 'Days Left'), value: dayValue(firstValue(httpPayload.value.cert_days_left, parseNumber(props.httpRecord?.certDaysLeft))), tone: certDaysTone(firstValue(httpPayload.value.cert_days_left, parseNumber(props.httpRecord?.certDaysLeft))) },
  { label: label('签发者', 'Issuer'), value: firstString(httpPayload.value.cert_issuer_cn, httpPayload.value.cert_issuer, props.httpRecord?.certIssuer), tone: 'normal' },
  { label: 'SAN', value: numberValue(firstValue(httpPayload.value.cert_san_count, props.httpRecord?.certDNSNames?.length)), tone: 'normal' },
  { label: label('证书链长度', 'Chain Length'), value: numberValue(httpPayload.value.cert_chain_length), tone: 'normal' },
])

const dnsMetrics = computed<MetricItem[]>(() => [
  { label: 'A', value: recordCount('A'), tone: recordCount('A') === '0' ? 'warn' : 'normal' },
  { label: 'AAAA', value: recordCount('AAAA'), tone: 'normal' },
  { label: label('CNAME 深度', 'CNAME Depth'), value: numberValue(dnsPayload.value.cname_chain_depth), tone: 'good' },
  { label: label('CNAME 终点', 'CNAME Terminal'), value: stringValue(dnsPayload.value.cname_terminal), tone: 'normal' },
  { label: label('MX 主机', 'MX Hosts'), value: mxHostsValue(), tone: mxHostsValue() === '-' ? 'warn' : 'normal' },
  { label: label('NS 主机', 'NS Hosts'), value: nsHostsValue(), tone: nsHostsValue() === '-' ? 'warn' : 'normal' },
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
    .map((key) => ({ label: key, value: headers[key]?.join(', ') ?? '', tone: headerTone(key) }))
    .filter((item) => item.value)
    .slice(0, 8)
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

function latencyTone(value: unknown, normalLimit = 1000, warnLimit = 3000): ObservationTone {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  if (!Number.isFinite(parsed)) {
    return 'normal'
  }
  if (parsed <= normalLimit) {
    return 'good'
  }
  if (parsed <= warnLimit) {
    return 'normal'
  }
  return 'warn'
}

function lossTone(value: unknown): ObservationTone {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  if (!Number.isFinite(parsed)) {
    return 'normal'
  }
  const normalized = parsed <= 1 ? parsed * 100 : parsed
  if (normalized <= 0) {
    return 'good'
  }
  if (normalized <= 5) {
    return 'normal'
  }
  return 'warn'
}

function statusTone(value: unknown): ObservationTone {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  if (!Number.isFinite(parsed)) {
    return 'normal'
  }
  if (parsed >= 200 && parsed < 400) {
    return 'good'
  }
  if (parsed >= 400 && parsed < 500) {
    return 'normal'
  }
  return 'warn'
}

function certDaysTone(value: unknown): ObservationTone {
  const parsed = typeof value === 'number' ? value : parseNumber(value)
  if (!Number.isFinite(parsed)) {
    return 'normal'
  }
  if (parsed > 30) {
    return 'good'
  }
  if (parsed > 7) {
    return 'normal'
  }
  return 'warn'
}

function headerTone(key: string): ObservationTone {
  if (key === 'x-powered-by') {
    return 'warn'
  }
  if (['content-type', 'etag', 'last-modified'].includes(key)) {
    return 'good'
  }
  return 'normal'
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.site-observation-tabs {
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

:global(html.dark .site-observation-tabs){
  background:
    radial-gradient(circle at 8% 0%, rgba(251, 146, 60, 0.12), transparent 30%),
    linear-gradient(120deg, rgba(15, 23, 42, 0.84), rgba(30, 41, 59, 0.72)),
    rgba(15, 23, 42, 0.82);
  box-shadow:
    inset 0 0 0 1px rgba(251, 146, 60, 0.14),
    0 18px 52px rgba(0, 0, 0, 0.18);
}

.observation-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.observation-eyebrow {
  margin-bottom: 0.25rem;
  color: #ea580c;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.observation-title {
  color: #0f172a;
  font-size: 1.05rem;
  font-weight: 800;
}

:global(html.dark .observation-title){
  color: #f8fafc;
}

.observation-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.78);
  padding: 0.25rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.06);
}

:global(html.dark .observation-tabs){
  background: rgba(15, 23, 42, 0.66);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.12);
}

.observation-tab {
  border-radius: 7px;
  padding: 0.55rem 0.95rem;
  color: #475569;
  font-size: 0.9rem;
  transition-duration: 500ms;
}

:global(html.dark .observation-tab){
  color: #cbd5e1;
}

.observation-tab:hover,
.observation-tab.is-active {
  background: #fdba74;
  color: #111827;
}

:global(html.dark .observation-tab:hover),
:global(html.dark .observation-tab.is-active){
  background: rgba(251, 146, 60, 0.26);
  color: #fff7ed;
}

.observation-content {
  margin-top: 1.1rem;
}

.empty-state {
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.54);
  padding: 1rem;
  color: #64748b;
  font-size: 0.875rem;
}

:global(html.dark .empty-state){
  background: rgba(15, 23, 42, 0.54);
  color: #94a3b8;
}

.dns-record-card {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  border-radius: 0;
  background: transparent;
  padding: 0.9rem 0 0;
  box-shadow: none;
  transition: none;
}

:global(html.dark .dns-record-card){
  border-top-color: rgba(251, 146, 60, 0.16);
}

.dns-record-card:first-child {
  border-top: 0;
  padding-top: 0;
}

.dns-record-card:hover {
  background: transparent;
  box-shadow: none;
}

.record-heading {
  margin-bottom: 0.45rem;
  color: #475569;
  font-size: 0.86rem;
  font-weight: 800;
}

:global(html.dark .record-heading){
  color: #cbd5e1;
}

.record-list {
  overflow: visible;
  border-radius: 0;
  background: transparent;
}

.record-row {
  display: grid;
  grid-template-columns: 4.5rem minmax(0, 1fr) 4rem;
  gap: 0.75rem;
  border-bottom: 1px solid rgba(251, 140, 47, 0.10);
  border-left: 2px solid transparent;
  padding: 0.58rem 0.85rem 0.58rem 0.75rem;
  font-size: 0.86rem;
  transition: background-color 500ms ease, border-color 500ms ease;
}

:global(html.dark .record-row){
  border-bottom-color: rgba(148, 163, 184, 0.12);
}

.record-row:hover {
  background: rgba(255, 237, 213, 0.68);
  border-left-color: rgba(251, 140, 47, 0.58);
}

:global(html.dark .record-row:hover){
  background: rgba(251, 146, 60, 0.12);
  border-left-color: rgba(251, 146, 60, 0.54);
}

.record-row:hover span:first-child {
  color: #9a4a12;
}

:global(html.dark .record-row:hover span:first-child){
  color: #fdba74;
}

@media (max-width: 640px) {
  .observation-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .observation-tabs {
    width: 100%;
    flex-wrap: nowrap;
    overflow-x: auto;
  }
}
</style>
