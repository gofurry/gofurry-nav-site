<template>
  <section class="metadata-probe-shell">
    <div class="site-info-tabs-panel">
      <div class="info-tabs-header">
        <h3 class="info-tabs-title">{{ activeInfoTabTitle }}</h3>
        <div class="info-tabs-nav">
          <button
            v-for="tab in infoTabs"
            :key="tab.key"
            type="button"
            class="info-tab-button"
            :class="{ 'is-active': activeInfoTab === tab.key }"
            @click="activeInfoTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <div v-if="activeInfoTab === 'metadata'" class="metadata-panel">
        <div v-if="pageInfoItems.length" class="metadata-list">
          <div
            v-for="item in pageInfoItems"
            :key="item.label"
            class="metadata-row"
          >
            <div class="metadata-label">{{ item.label }}</div>
            <div class="metadata-value">{{ Array.isArray(item.value) ? item.value.join(', ') : item.value }}</div>
          </div>
        </div>
        <div v-else class="metadata-empty">{{ label('暂无数据', 'No data') }}</div>
      </div>

      <div v-else-if="activeInfoTab === 'changes'" class="info-tab-body changes-panel">
        <div v-if="changesLoading" class="panel-empty">{{ label('变化事件加载中', 'Loading changes') }}</div>
        <div
          v-for="event in changeEvents"
          :key="event.key"
          class="change-event-row"
        >
          <div class="change-event-head">
            <span class="protocol-badge">{{ event.protocol }}</span>
            <span class="change-field">{{ event.field }}</span>
            <span class="change-time">{{ event.detectedAt }}</span>
          </div>
          <div class="change-value-grid">
            <div class="change-value-block">
              <p class="change-value-label">{{ label('旧值', 'Old value') }}</p>
              <p class="change-value-text">{{ event.oldValue }}</p>
            </div>
            <div class="change-value-block">
              <p class="change-value-label">{{ label('新值', 'New value') }}</p>
              <p class="change-value-text">{{ event.newValue }}</p>
            </div>
          </div>
        </div>
        <div v-if="!changesLoading && !changeEvents.length" class="panel-empty">{{ label('暂无变化事件', 'No change events') }}</div>
      </div>

      <div v-else class="info-tab-body history-panel">
        <div v-if="observationsLoading" class="panel-empty">{{ label('观测历史加载中', 'Loading history') }}</div>
        <section
          v-for="history in observationHistories"
          :key="history.protocol"
          class="history-section"
        >
          <div class="history-head">
            <h4 class="history-title">{{ history.title }}</h4>
            <div v-if="history.totalPages > 1" class="history-pager">
              <button
                type="button"
                class="history-page-button"
                :disabled="history.page <= 1"
                @click="setObservationPage(history.protocol, history.page - 1)"
              >
                {{ label('上一页', 'Prev') }}
              </button>
              <span class="history-page-count">{{ history.page }}/{{ history.totalPages }}</span>
              <button
                type="button"
                class="history-page-button"
                :disabled="history.page >= history.totalPages"
                @click="setObservationPage(history.protocol, history.page + 1)"
              >
                {{ label('下一页', 'Next') }}
              </button>
            </div>
          </div>
          <div v-if="history.items.length" class="history-list">
            <div
              v-for="item in history.visibleItems"
              :key="`${history.protocol}:${item.observed_at}:${item.duration_ms}`"
              class="history-row"
            >
              <span
                :aria-label="statusText(item.status)"
                :class="['status-dot', statusDotClass(item.status)]"
                :title="statusText(item.status)"
              />
              <span class="history-summary">{{ observationSummary(history.protocol, item) }}</span>
              <span class="history-time">{{ formatTime(item.observed_at) }}</span>
            </div>
          </div>
          <div v-else class="panel-empty">{{ label('暂无历史', 'No history') }}</div>
        </section>
      </div>
    </div>

    <div class="light-probe-panel">
      <div class="light-probe-header">
        <h3 class="info-tabs-title">{{ label('低频轻探测', 'Light probes') }}</h3>
      </div>
      <div v-if="lightProbeEntries.length" class="light-probe-grid">
        <button
          v-for="probe in lightProbeEntries"
          :key="probe.protocol"
          type="button"
          class="light-probe-card"
          @click="selectedProbe = probe"
        >
          <div class="light-probe-card-head">
            <span class="light-probe-card-title">{{ protocolName(probe.protocol) }}</span>
            <span
              :aria-label="statusText(probe.status)"
              :class="['status-dot', statusDotClass(probe.status)]"
              :title="statusText(probe.status)"
            />
          </div>
          <div v-if="probe.items.length" class="light-probe-facts">
            <div
              v-for="item in probe.items"
              :key="item.label"
              class="light-probe-fact"
            >
              <span class="light-probe-label">{{ item.label }}</span>
              <span class="light-probe-value">{{ Array.isArray(item.value) ? item.value.join(', ') : item.value }}</span>
            </div>
          </div>
          <div v-else class="light-probe-empty">{{ label('暂无数据', 'No data') }}</div>
          <div class="light-probe-detail-hint">
            {{ label('点击查看详情', 'Click for details') }}
          </div>
        </button>
      </div>
      <div v-else class="panel-empty">{{ label('暂无数据', 'No data') }}</div>
    </div>

    <Teleport to="body">
      <Transition name="probe-modal">
        <div
          v-if="selectedProbe"
          class="probe-modal-backdrop"
          @click.self="selectedProbe = null"
        >
          <article
            class="probe-modal-dialog"
            role="dialog"
            aria-modal="true"
          >
            <header class="probe-modal-header">
              <div>
                <p class="probe-modal-eyebrow">{{ label('低频轻探测详情', 'Light probe detail') }}</p>
                <h3 class="probe-modal-title">{{ protocolName(selectedProbe.protocol) }}</h3>
              </div>
              <div class="probe-modal-actions">
                <button
                  type="button"
                  class="probe-modal-close"
                  @click="selectedProbe = null"
                >
                  {{ label('关闭', 'Close') }}
                </button>
              </div>
            </header>

            <div class="probe-modal-body">
              <div class="probe-modal-summary-grid">
                <div class="probe-modal-summary-item">
                  <p class="probe-modal-summary-label">{{ label('观测时间', 'Observed') }}</p>
                  <p class="probe-modal-summary-value">{{ formatTime(selectedProbe.observedAt) }}</p>
                </div>
                <div class="probe-modal-summary-item">
                  <p class="probe-modal-summary-label">{{ label('耗时', 'Duration') }}</p>
                  <p class="probe-modal-summary-value">{{ formatDuration(selectedProbe.durationMs) }}</p>
                </div>
                <div class="probe-modal-summary-item">
                  <p class="probe-modal-summary-label">{{ label('结果', 'Result') }}</p>
                  <p class="probe-modal-summary-value">{{ selectedProbe.errorCode || statusText(selectedProbe.status) }}</p>
                </div>
              </div>

              <div v-if="selectedProbe.errorMessage" class="probe-modal-error">
                {{ selectedProbe.errorMessage }}
              </div>

              <div class="probe-modal-sections">
                <section
                  v-for="section in selectedProbeDetailSections"
                  :key="section.title"
                  class="probe-modal-section"
                >
                  <h4 class="probe-modal-section-title">{{ section.title }}</h4>
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
import { computed, defineComponent, h, ref, watch, type PropType } from 'vue'
import { i18n } from '@/main'
import type { CollectorEnvelope, HttpRecord, TargetChangesResponse, TargetLatestResponse, TargetObservationsResponse } from '@/types/nav'

type InfoItem = { label: string; value: string | string[] }
type DetailSection = { title: string; items: InfoItem[] }
type InfoTabKey = 'metadata' | 'changes' | 'history'
type ObservationProtocol = 'ping' | 'http' | 'dns'
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
  siteId: string | number
  target: string
}>()

const navV2Api = useApi('navV2')
const httpPayload = computed(() => asRecord(props.targetLatestCore?.protocols?.http?.payload))
const httpStatus = computed(() => props.targetLatestCore?.protocols?.http?.status || '')
const lightProtocols = computed(() => props.lightProbeState?.protocols ?? {})
const activeInfoTab = ref<InfoTabKey>('metadata')
const targetChanges = ref<TargetChangesResponse | null>(null)
const targetObservations = ref<Record<ObservationProtocol, TargetObservationsResponse | null>>({
  ping: null,
  http: null,
  dns: null,
})
const changesLoading = ref(false)
const observationsLoading = ref(false)
const observationPages = ref<Record<ObservationProtocol, number>>({
  ping: 1,
  http: 1,
  dns: 1,
})
const selectedProbe = ref<LightProbeEntry | null>(null)
const infoTabs = computed<{ key: InfoTabKey; label: string }[]>(() => [
  { key: 'metadata', label: label('页面元信息', 'Metadata') },
  { key: 'changes', label: label('变化事件', 'Changes') },
  { key: 'history', label: label('观测历史', 'History') },
])
const activeInfoTabTitle = computed(() => infoTabs.value.find(tab => tab.key === activeInfoTab.value)?.label ?? label('页面元信息', 'Metadata'))
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
const changeEvents = computed(() => arrayRecords(targetChanges.value?.events).slice(0, 12).map((raw, index) => ({
  key: stringValue(raw.event_id || raw.id || `${raw.protocol || 'event'}:${index}`),
  protocol: protocolLabel(stringValue(raw.protocol)),
  field: stringValue(raw.field || raw.category || raw.change_type),
  oldValue: inlineValue(raw.old_value),
  newValue: inlineValue(raw.new_value),
  detectedAt: formatTime(stringValue(raw.detected_at || raw.observed_at || raw.create_time)),
})))
const observationPageSize = 4
const observationHistories = computed(() => (['ping', 'http', 'dns'] as ObservationProtocol[]).map((protocol) => {
  const items = targetObservations.value[protocol]?.items ?? []
  const totalPages = Math.max(1, Math.ceil(items.length / observationPageSize))
  const page = Math.min(Math.max(observationPages.value[protocol] || 1, 1), totalPages)
  const start = (page - 1) * observationPageSize
  return {
    protocol,
    title: protocolLabel(protocol),
    items,
    visibleItems: items.slice(start, start + observationPageSize),
    page,
    totalPages,
  }
}))

watch(activeInfoTab, (tab) => {
  if (tab === 'changes') {
    void loadChanges()
  }
  if (tab === 'history') {
    void loadObservations()
  }
})

watch(() => [props.siteId, props.target], () => {
  targetChanges.value = null
  targetObservations.value = { ping: null, http: null, dns: null }
  observationPages.value = { ping: 1, http: 1, dns: 1 }
  if (activeInfoTab.value === 'changes') {
    void loadChanges()
  }
  if (activeInfoTab.value === 'history') {
    void loadObservations()
  }
})

const InfoList = defineComponent({
  props: {
    items: { type: Array as PropType<InfoItem[]>, default: () => [] },
    emptyText: { type: String, default: '-' },
    compact: { type: Boolean, default: false },
  },
  setup(componentProps) {
    return () => componentProps.items.length
      ? h('div', { class: componentProps.compact ? 'modal-info-list is-compact' : 'modal-info-list' },
        componentProps.items.map(item => h('div', { class: 'modal-info-row' }, [
          h('span', { class: 'modal-info-label' }, item.label),
          h('span', { class: 'modal-info-value' }, Array.isArray(item.value) ? item.value.join(', ') : item.value),
        ])))
      : h('div', { class: 'modal-empty' }, componentProps.emptyText)
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

function protocolLabel(protocol: string) {
  const normalized = protocol.toLowerCase()
  const map: Record<string, string> = {
    ping: 'PING',
    http: 'HTTP',
    dns: 'DNS',
  }
  return map[normalized] || protocolName(protocol).toUpperCase()
}

function setObservationPage(protocol: ObservationProtocol, page: number) {
  const history = observationHistories.value.find(item => item.protocol === protocol)
  const totalPages = history?.totalPages ?? 1
  observationPages.value[protocol] = Math.min(Math.max(page, 1), totalPages)
}

async function loadChanges() {
  if (targetChanges.value || changesLoading.value || !props.siteId || !props.target) {
    return
  }

  changesLoading.value = true
  try {
    targetChanges.value = await navV2Api<TargetChangesResponse>(targetPath('/changes'))
  } catch {
    targetChanges.value = null
  } finally {
    changesLoading.value = false
  }
}

async function loadObservations() {
  if (observationsLoading.value || !props.siteId || !props.target) {
    return
  }
  const hasAllProtocols = (['ping', 'http', 'dns'] as ObservationProtocol[]).every(protocol => targetObservations.value[protocol])
  if (hasAllProtocols) {
    return
  }

  observationsLoading.value = true
  try {
    const [ping, http, dns] = await Promise.all((['ping', 'http', 'dns'] as ObservationProtocol[]).map(protocol =>
      navV2Api<TargetObservationsResponse>(targetPath('/observations'), {
        query: {
          protocol,
          limit: 8,
          payload_mode: 'preview',
        },
      }).catch(() => null)
    ))
    targetObservations.value = {
      ping: ping ?? null,
      http: http ?? null,
      dns: dns ?? null,
    }
  } finally {
    observationsLoading.value = false
  }
}

function targetPath(suffix: string) {
  return `/nav/sites/${props.siteId}/targets/${encodeURIComponent(props.target)}${suffix}`
}

function observationSummary(protocol: string, envelope: CollectorEnvelope) {
  const payload = asRecord(envelope.payload)
  const normalized = protocol.toLowerCase()

  if (normalized === 'ping') {
    return compactParts([
      valueWithLabel(label('平均延迟', 'Avg RTT'), millisValue(payload.avg_rtt_ms || payload.delay_ms)),
      valueWithLabel(label('丢包', 'Loss'), percentLike(payload.loss_rate || payload.legacy_loss)),
      valueWithLabel(label('耗时', 'Duration'), formatDuration(envelope.duration_ms)),
    ])
  }

  if (normalized === 'http') {
    return compactParts([
      valueWithLabel(label('状态码', 'Status'), numberValue(payload.status_code)),
      valueWithLabel(label('响应', 'Response'), millisValue(payload.response_time_ms || envelope.duration_ms)),
      valueWithLabel(label('协议', 'Protocol'), stringValue(payload.http_protocol)),
      valueWithLabel(label('类型', 'Type'), stringValue(payload.content_type)),
    ])
  }

  if (normalized === 'dns') {
    const flags = stringArray(payload.risk_flags)
    return compactParts([
      flags.length ? valueWithLabel(label('风险', 'Risk'), flags.join(', ')) : '',
      valueWithLabel('A', recordCount(payload.A)),
      valueWithLabel('AAAA', recordCount(payload.AAAA)),
      valueWithLabel('NS', recordCount(payload.NS)),
      valueWithLabel(label('耗时', 'Duration'), formatDuration(envelope.duration_ms)),
    ])
  }

  return compactParts([
    valueWithLabel(label('耗时', 'Duration'), formatDuration(envelope.duration_ms)),
    stringValue(envelope.error_code),
  ])
}

function recordCount(value: unknown) {
  return Array.isArray(value) ? String(value.length) : '-'
}

function millisValue(value: unknown) {
  const number = numberFromAny(value)
  return number > 0 ? `${Math.round(number)}ms` : '-'
}

function percentLike(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value <= 1 ? `${Math.round(value * 100)}%` : `${value}%`
  }
  return stringValue(value)
}

function valueWithLabel(name: string, value: string) {
  return value && value !== '-' ? `${name}: ${value}` : ''
}

function compactParts(parts: string[]) {
  return parts.filter(Boolean).join(' / ') || '-'
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

function statusDotClass(status: string) {
  if (status === 'success') return 'is-success'
  if (status === 'failure') return 'is-failure'
  if (status === 'skipped') return 'is-skipped'
  return 'is-unknown'
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

function inlineValue(value: unknown) {
  const displayed = displayValue(value)
  return Array.isArray(displayed) ? displayed.join(', ') || '-' : displayed
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
.metadata-probe-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: clamp(1.5rem, 3vw, 2.3rem);
}

.site-info-tabs-panel {
  min-width: 0;
}

.light-probe-panel {
  min-width: 0;
}

.light-probe-header {
  display: flex;
  align-items: center;
  min-height: 2.45rem;
}

.info-tabs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.25rem;
  min-height: 2.45rem;
}

.info-tabs-title {
  color: #0f172a;
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.35;
}

.info-tabs-nav {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.35rem;
  border-radius: 8px;
  background: transparent;
}

.info-tab-button {
  border-radius: 8px;
  padding: 0.55rem 0.85rem;
  color: #475569;
  font-size: 0.9rem;
  font-weight: 600;
  transition: background-color 500ms ease, color 500ms ease;
}

.info-tab-button:hover,
.info-tab-button.is-active {
  background: #fdba74;
  color: #111827;
}

.metadata-panel {
  margin-top: 1.15rem;
}

.metadata-list {
  display: grid;
  gap: 0.74rem;
}

.metadata-row {
  display: grid;
  grid-template-columns: minmax(8rem, 13rem) minmax(0, 1fr);
  gap: clamp(1rem, 4vw, 2.4rem);
  align-items: start;
  min-width: 0;
}

.metadata-label {
  min-width: 0;
  color: #64748b;
  font-size: 0.92rem;
  font-weight: 800;
  line-height: 1.55;
}

.metadata-value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.9rem;
  line-height: 1.55;
}

.metadata-empty {
  color: #64748b;
  font-size: 0.9rem;
}

.info-tab-body {
  margin-top: 1.25rem;
}

.panel-empty {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 0.9rem 0;
  color: #64748b;
  font-size: 0.9rem;
}

.change-event-row {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 0.95rem 0;
}

.change-event-row:first-of-type {
  border-top: 0;
  padding-top: 0;
}

.change-event-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.protocol-badge {
  border-radius: 999px;
  background: rgba(255, 237, 213, 0.76);
  padding: 0.22rem 0.6rem;
  color: #9a4a12;
  font-size: 0.74rem;
  font-weight: 800;
}

.change-field {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-size: 0.92rem;
  font-weight: 800;
}

.change-time {
  margin-left: auto;
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.78rem;
}

.change-value-grid {
  margin-top: 0.8rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.9rem 1.2rem;
}

.change-value-block {
  min-width: 0;
  border-left: 2px solid rgba(251, 140, 47, 0.18);
  padding-left: 0.75rem;
}

.change-value-label {
  color: #64748b;
  font-size: 0.76rem;
  font-weight: 800;
}

.change-value-text {
  margin-top: 0.25rem;
  min-width: 0;
  overflow-wrap: anywhere;
  color: #1f2937;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.55;
}

.history-section {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 1rem 0;
}

.history-section:first-of-type {
  border-top: 0;
  padding-top: 0;
}

.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.history-title {
  color: #111827;
  font-size: 0.94rem;
  font-weight: 800;
}

.history-pager {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.78rem;
}

.history-page-button {
  border-radius: 8px;
  background: transparent;
  padding: 0.25rem 0.55rem;
  color: #475569;
  transition: background-color 500ms ease, color 500ms ease;
}

.history-page-button:hover:not(:disabled) {
  background: rgba(255, 237, 213, 0.72);
  color: #9a4a12;
}

.history-page-button:disabled {
  cursor: not-allowed;
  color: #cbd5e1;
}

.history-page-count {
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

.history-list {
  margin-top: 0.7rem;
}

.history-row {
  display: grid;
  grid-template-columns: 1rem minmax(0, 1fr) 8.5rem;
  gap: 0.75rem;
  align-items: center;
  border-bottom: 1px solid rgba(251, 140, 47, 0.10);
  border-left: 2px solid transparent;
  padding: 0.6rem 0.75rem 0.6rem 0.65rem;
  transition: background-color 500ms ease, border-color 500ms ease;
}

.history-row:hover {
  background: rgba(255, 237, 213, 0.52);
  border-left-color: rgba(251, 140, 47, 0.45);
}

.history-row:last-child {
  border-bottom: 0;
}

.history-summary {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #1f2937;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.55;
}

.history-time {
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.78rem;
  line-height: 1.55;
  text-align: right;
}

.light-probe-grid {
  margin-top: 1.15rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.light-probe-card {
  display: block;
  width: 100%;
  min-width: 0;
  break-inside: avoid;
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.42);
  padding: 0.74rem 0.8rem;
  text-align: left;
  transition: background-color 500ms ease, box-shadow 500ms ease;
}

.light-probe-card:hover,
.light-probe-card:focus-visible {
  background: rgba(255, 247, 235, 0.72);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.22), 0 0 0 4px rgba(251, 140, 47, 0.055);
  outline: none;
}

.light-probe-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
}

.light-probe-card-title {
  min-width: 0;
  color: #111827;
  font-size: 0.92rem;
  font-weight: 800;
}

.light-probe-facts {
  margin-top: 0.62rem;
  display: grid;
  gap: 0.3rem;
}

.light-probe-fact {
  display: grid;
  grid-template-columns: minmax(4.4rem, 0.42fr) minmax(0, 1fr);
  gap: 0.5rem;
  align-items: start;
  min-width: 0;
  font-size: 0.84rem;
  line-height: 1.45;
}

.light-probe-label {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #64748b;
  font-weight: 800;
}

.light-probe-value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

.light-probe-empty {
  margin-top: 0.7rem;
  color: #64748b;
  font-size: 0.84rem;
}

.light-probe-detail-hint {
  margin-top: 0.65rem;
  color: #ea580c;
  font-size: 0.74rem;
  font-weight: 800;
  opacity: 0;
  transition: opacity 500ms ease;
}

.light-probe-card:hover .light-probe-detail-hint,
.light-probe-card:focus-visible .light-probe-detail-hint {
  opacity: 1;
}

.status-dot {
  display: inline-block;
  width: 0.58rem;
  height: 0.58rem;
  flex: 0 0 auto;
  border-radius: 999px;
  vertical-align: middle;
}

.status-dot.is-success {
  background: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.14);
}

.status-dot.is-failure {
  background: #ef4444;
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.12);
}

.status-dot.is-skipped {
  background: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.12);
}

.status-dot.is-unknown {
  background: #94a3b8;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.12);
}

.probe-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 42, 0.46);
  padding: 1.5rem 1rem;
  backdrop-filter: blur(6px);
}

.probe-modal-dialog {
  display: flex;
  width: min(100%, 56rem);
  max-height: 88vh;
  flex-direction: column;
  overflow: hidden;
  border-radius: 8px;
  background:
    radial-gradient(circle at 8% 0%, rgba(251, 140, 47, 0.08), transparent 30%),
    linear-gradient(120deg, rgba(255, 247, 235, 0.88), rgba(255, 250, 242, 0.94)),
    rgba(255, 247, 235, 0.90);
  color: #111827;
  box-shadow:
    inset 0 0 0 1px rgba(251, 140, 47, 0.16),
    0 24px 70px rgba(15, 23, 42, 0.22);
}

.probe-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(251, 140, 47, 0.14);
  padding: 1rem 1.25rem 0.9rem;
}

.probe-modal-eyebrow {
  color: #ea580c;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.35;
}

.probe-modal-title {
  margin-top: 0.18rem;
  color: #111827;
  font-size: 1.22rem;
  font-weight: 850;
  line-height: 1.25;
}

.probe-modal-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.85rem;
}

.probe-modal-close {
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.78);
  padding: 0.42rem 0.72rem;
  color: #475569;
  font-size: 0.86rem;
  font-weight: 700;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.10);
  transition: background-color 500ms ease, color 500ms ease, box-shadow 500ms ease;
}

.probe-modal-close:hover,
.probe-modal-close:focus-visible {
  background: #fdba74;
  color: #111827;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.18), 0 0 0 4px rgba(251, 140, 47, 0.06);
  outline: none;
}

.probe-modal-body {
  overflow-y: auto;
  padding: 1rem 1.25rem 1.25rem;
}

.probe-modal-summary-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
}

.probe-modal-summary-item {
  min-width: 0;
  padding: 0.72rem 0;
}

.probe-modal-summary-label {
  color: #64748b;
  font-size: 0.76rem;
  font-weight: 800;
  line-height: 1.45;
}

.probe-modal-summary-value {
  margin-top: 0.18rem;
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.45;
}

.probe-modal-error {
  margin-top: 1rem;
  border-left: 2px solid rgba(239, 68, 68, 0.34);
  background: rgba(254, 242, 242, 0.62);
  padding: 0.72rem 0.85rem;
  color: #991b1b;
  font-size: 0.88rem;
  line-height: 1.55;
}

.probe-modal-sections {
  margin-top: 1rem;
  display: grid;
  gap: 0.95rem;
}

.probe-modal-section {
  min-width: 0;
  border-top: 1px solid rgba(251, 140, 47, 0.14);
  padding-top: 0.95rem;
}

.probe-modal-section:first-child {
  border-top: 0;
  padding-top: 0;
}

.probe-modal-section-title {
  margin-bottom: 0.55rem;
  color: #111827;
  font-size: 0.94rem;
  font-weight: 850;
  line-height: 1.35;
}

:deep(.modal-info-list) {
  display: grid;
}

:deep(.modal-info-row) {
  display: grid;
  grid-template-columns: minmax(7.5rem, 12rem) minmax(0, 1fr);
  gap: 1rem;
  align-items: start;
  min-width: 0;
  border-bottom: 1px solid rgba(251, 140, 47, 0.10);
  border-left: 2px solid transparent;
  padding: 0.48rem 0.65rem 0.48rem 0.55rem;
  transition: background-color 500ms ease, border-color 500ms ease;
}

:deep(.modal-info-row:hover) {
  background: rgba(255, 237, 213, 0.48);
  border-left-color: rgba(251, 140, 47, 0.42);
}

:deep(.modal-info-row:last-child) {
  border-bottom: 0;
}

:deep(.modal-info-label) {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #64748b;
  font-size: 0.88rem;
  font-weight: 800;
  line-height: 1.5;
}

:deep(.modal-info-value) {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.5;
}

:deep(.modal-empty) {
  color: #64748b;
  font-size: 0.88rem;
}

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

@media (max-width: 640px) {
  .probe-modal-backdrop {
    align-items: stretch;
    padding: 0.75rem;
  }

  .probe-modal-dialog {
    max-height: calc(100vh - 1.5rem);
  }

  .probe-modal-header {
    align-items: flex-start;
    padding: 0.95rem 1rem 0.85rem;
  }

  .probe-modal-body {
    padding: 0.9rem 1rem 1rem;
  }

  :deep(.modal-info-row) {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.12rem;
  }

  .info-tabs-header {
    flex-direction: column;
  }

  .info-tabs-nav {
    width: 100%;
    justify-content: flex-start;
  }

  .info-tab-button {
    flex: 1 1 auto;
  }

  .metadata-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.18rem;
  }

  .change-time {
    margin-left: 0;
    width: 100%;
  }

  .change-value-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .history-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .history-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.35rem;
  }

  .history-time {
    text-align: left;
  }
}

@media (min-width: 768px) {
  .probe-modal-summary-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .probe-modal-summary-item {
    padding: 0.72rem 1rem;
  }

  .probe-modal-summary-item:first-child {
    padding-left: 0;
  }

  .change-value-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .light-probe-grid {
    display: block;
    column-count: 2;
    column-gap: 0.75rem;
  }

  .light-probe-card {
    margin-bottom: 0.75rem;
  }
}

@media (min-width: 1280px) {
  .metadata-probe-shell {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.05fr);
  }
}
</style>
