import { computed, ref, watch } from 'vue'
import { i18n } from '@/main'
import type { CollectorEnvelope, HttpRecord, TargetChangesResponse, TargetLatestResponse, TargetObservationsResponse } from '@/types/nav'
import type { DetailInfoItem, DetailSection, LightProbeEntry, ObservationProtocol } from '@/components/site/detailTypes'

type InfoItem = DetailInfoItem
type InfoTabKey = 'metadata' | 'changes' | 'history'

export type SiteMetadataProbePanelProps = {
  httpRecord: HttpRecord | null
  targetLatestCore: TargetLatestResponse | null
  lightProbeState: TargetLatestResponse | null
  siteId: string | number
  target: string
}

export function useSiteMetadataProbePanel(props: SiteMetadataProbePanelProps) {
  const navV2Api = useApi('navV2')
  const httpPayload = computed(() => asRecord(props.targetLatestCore?.protocols?.http?.payload))
  const lightProtocols = computed(() => props.lightProbeState?.protocols ?? {})

  function headerValue(key: string) {
    const normalizedKey = key.toLowerCase()
    for (const [name, values] of Object.entries(props.httpRecord?.headers ?? {})) {
      if (name.toLowerCase() === normalizedKey && values?.length) {
        return values[0]
      }
    }
    return ''
  }

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

  const lightProbeOrder = ['rdap', 'robots', 'security_txt', 'llms_txt', 'page_assets', 'port_check', 'waf_canary']
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

  return {
    activeInfoTab,
    activeInfoTabTitle,
    changeEvents,
    changesLoading,
    infoTabs,
    label,
    lightProbeDetailSections,
    lightProbeEntries,
    observationHistories,
    observationsLoading,
    observationSummary,
    pageInfoItems,
    setObservationPage,
  }
}

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
    case 'llms_txt':
      return compactItems([
        { label: 'Exists', value: boolText(payload.exists) },
        { label: 'Title', value: stringValue(payload.title) },
        { label: 'Headings', value: numberValue(payload.heading_count) },
        { label: 'Links', value: numberValue(payload.link_count) },
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
    case 'llms_txt':
      return compactSections([
        {
          title: 'llms.txt',
          items: compactItems([
            { label: label('存在', 'Exists'), value: boolText(payload.exists) },
            { label: label('路径', 'Path'), value: stringValue(payload.path) },
            { label: label('状态码', 'Status'), value: numberValue(payload.status_code) },
            { label: 'Content-Type', value: stringValue(payload.content_type) },
            { label: label('标题', 'Title'), value: stringValue(payload.title) },
            { label: label('章节数', 'Headings'), value: numberValue(payload.heading_count) },
            { label: label('链接数', 'Links'), value: numberValue(payload.link_count) },
            { label: 'Optional', value: boolText(payload.optional_section_present) },
            { label: label('读取字节', 'Bytes read'), value: bytesText(payload.body_read_bytes) },
            { label: label('响应截断', 'Truncated'), value: boolText(payload.body_truncated) },
            { label: 'SHA256', value: stringValue(payload.sha256) },
          ]),
        },
        {
          title: label('章节', 'Headings'),
          items: stringArray(payload.headings).map((value, index) => ({ label: `#${index + 1}`, value })),
        },
        {
          title: label('链接样例', 'Link samples'),
          items: stringArray(payload.links).map((value, index) => ({ label: `#${index + 1}`, value })),
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
    llms_txt: 'llms.txt',
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

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
