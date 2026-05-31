<template>
  <section class="">
    <!-- 核心指标 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      <div
          class="rounded-xl p-5"
          :class="responseColor"
      >
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-sm font-medium text-gray-500"> {{ performanceText.visitTiming }} </h3>
          <i class="fa fa-clock-o" :class="iconColor(responseColor)"></i>
        </div>
        <p class="text-2xl text-gray-700 font-bold">{{ metrics.responseTime }}</p>
        <p class="text-xs text-gray-500 mt-1"> {{ performanceText.visitTimingHint }} </p>
        <dl class="mt-3 space-y-1 text-xs text-gray-500">
          <div v-for="item in timingDetails" :key="item.label" class="flex justify-between gap-2">
            <dt>{{ item.label }}</dt>
            <dd class="font-medium text-gray-700">{{ item.value }}</dd>
          </div>
        </dl>
      </div>

      <div
          class="rounded-xl p-5"
          :class="statusColor"
      >
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-sm font-medium text-gray-500">{{ performanceText.responseStatus }}</h3>
          <i class="fa fa-check-circle" :class="iconColor(statusColor)"></i>
        </div>
        <p class="text-2xl text-gray-700 font-bold">HTTP {{ metrics.statusCode }}</p>
        <p class="text-xs text-gray-500 mt-1">{{ performanceText.responseStatusHint }}</p>
        <dl class="mt-3 space-y-1 text-xs text-gray-500">
          <div v-for="item in responseDetails" :key="item.label" class="flex justify-between gap-3">
            <dt>{{ item.label }}</dt>
            <dd class="truncate font-medium text-gray-700">{{ item.value }}</dd>
          </div>
        </dl>
      </div>

      <div
          class="rounded-xl p-5"
          :class="tlsColor"
      >
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-sm font-medium text-gray-500">{{ performanceText.secureTransport }}</h3>
          <i class="fa fa-shield" :class="iconColor(tlsColor)"></i>
        </div>
        <p class="text-2xl text-gray-700 font-bold">{{ metrics.tlsVersion }}</p>
        <p class="text-xs text-gray-500 mt-1">{{ tlsSubtitle }}</p>
        <dl class="mt-3 space-y-1 text-xs text-gray-500">
          <div v-for="item in tlsDetails" :key="item.label" class="flex justify-between gap-3">
            <dt>{{ item.label }}</dt>
            <dd class="truncate font-medium text-gray-700">{{ item.value }}</dd>
          </div>
        </dl>
      </div>

      <div
          class="rounded-xl p-5"
          :class="certColor"
      >
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-sm font-medium text-gray-500">{{ performanceText.certStatus }}</h3>
          <i class="fa fa-calendar" :class="iconColor(certColor)"></i>
        </div>
        <p class="text-2xl text-gray-700 font-bold">{{ metrics.certDaysLeft }}</p>
        <p class="text-xs text-gray-500 mt-1">{{ performanceText.certStatusHint }}</p>
        <dl class="mt-3 space-y-1 text-xs text-gray-500">
          <div v-for="item in certDetails" :key="item.label" class="flex justify-between gap-3">
            <dt>{{ item.label }}</dt>
            <dd class="truncate font-medium text-gray-700">{{ item.value }}</dd>
          </div>
        </dl>
      </div>
    </div>

    <slot name="after-metrics" />

    <div class="rounded-2xl bg-orange-100/45 p-5">
      <div class="mb-5 grid grid-cols-1 gap-4 xl:grid-cols-[1fr_auto_1fr] xl:items-end">
        <div class="text-center xl:text-left">
          <div class="mb-1 text-xs font-medium uppercase tracking-wide text-orange-500">
            {{ label('PING 延迟观测', 'Ping observation') }}
          </div>
          <h3 class="text-lg font-semibold text-gray-900">{{ t('site.performance.latencyTrend') }}</h3>
        </div>

        <div class="flex justify-center">
          <div class="inline-flex rounded-xl bg-orange-50 p-1">
            <button
              v-for="option in sampleOptions"
              :key="option.value"
              @click="changeSample(option.value)"
              :class="[
                'rounded-lg px-4 py-2 text-sm transition-colors',
                sampleType === option.value
                  ? 'bg-orange-200 text-gray-900'
                  : 'text-gray-600 hover:bg-orange-100'
              ]"
            >
              {{ option.label }}
            </button>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-2 sm:grid-cols-4 xl:justify-self-end">
          <div
            v-for="item in trendStats"
            :key="item.label"
            class="rounded-xl bg-orange-50 px-3 py-2"
          >
            <div class="whitespace-nowrap text-[11px] text-gray-500">{{ item.label }}</div>
            <div class="mt-1 whitespace-nowrap text-sm font-semibold text-gray-800">{{ item.value }}</div>
          </div>
        </div>
      </div>

      <div class="rounded-xl bg-orange-50/70 px-3 pb-4 pt-6">
        <div class="h-[360px] w-full" ref="latencyChartRef"></div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { CollectorEnvelope, PingRecord, PingStats, HttpRecord, TargetLatestResponse, TargetObservationsResponse } from '@/types/nav'
import {i18n} from "@/main";

const t = (key: string) => i18n.global.t(key)

interface Props {
  pingRecord: PingRecord | null
  httpRecord: HttpRecord
  targetLatestCore?: TargetLatestResponse | null
  siteId?: string | number
  domain?: string
}

const props = defineProps<Props>()
const navV2Api = useApi('navV2')
const latencyChartRef = ref<HTMLElement | null>(null)
const chart = ref<echarts.ECharts | null>(null)
const sampleType = ref<'twenty' | 'sixty' | 'hundred'>('twenty')
const v2PingRecord = ref<PingRecord | null>(null)
const httpPayload = computed(() => asRecord(props.targetLatestCore?.protocols?.http?.payload))
const yesText = computed(() => i18n.global.locale.value === 'en' ? 'Yes' : '是')
const noText = computed(() => i18n.global.locale.value === 'en' ? 'No' : '否')

// 当前 ping 数据
const effectivePingRecord = computed(() => props.pingRecord || v2PingRecord.value)
const currentPing = computed<PingStats | null>(() => effectivePingRecord.value?.[sampleType.value] || null)
const sampleOptions = computed(() => [
  { value: 'twenty' as const, label: label('20次', '20') },
  { value: 'sixty' as const, label: label('60次', '60') },
  { value: 'hundred' as const, label: label('100次', '100') },
])
const trendStats = computed(() => {
  const latestPoint = currentPing.value?.DelayModel?.[0]
  return [
    { label: t('site.performance.averageLatency'), value: currentPing.value?.avgDelay || '-' },
    { label: t('site.performance.packetLossRate'), value: currentPing.value?.avgLoss || '-' },
    { label: label('最新延迟', 'Latest'), value: latestPoint?.delay ? `${latestPoint.delay}ms` : '-' },
    { label: label('样本数量', 'Samples'), value: String(currentPing.value?.DelayModel?.length ?? 0) },
  ]
})

// 核心指标
const metrics = computed(() => ({
  responseTime: formatMs(firstNumber(httpPayload.value.response_time_ms, parseNumber(props.httpRecord.responseTime))),
  statusCode: firstNumber(httpPayload.value.status_code, props.httpRecord.statusCode),
  tlsVersion: firstString(httpPayload.value.tls_version, props.httpRecord.tlsVersion, 'Unknown'),
  certDaysLeft: formatDays(firstNumber(httpPayload.value.cert_days_left, parseNumber(props.httpRecord.certDaysLeft)))
}))

const performanceText = computed(() => ({
  visitTiming: label('访问耗时', 'Visit Timing'),
  visitTimingHint: label('用户访问', 'Visitor access'),
  responseStatus: label('响应状态', 'Response Status'),
  responseStatusHint: label('HTTP 与内容协商', 'HTTP and content'),
  secureTransport: label('安全传输', 'Secure Transport'),
  certStatus: label('证书状态', 'Certificate Status'),
  certStatusHint: label('剩余有效期', 'Remaining validity'),
}))

const timingDetails = computed(() => [
  { label: 'DNS', value: formatMs(firstNumber(httpPayload.value.dns_lookup_ms)) },
  { label: 'TCP', value: formatMs(firstNumber(httpPayload.value.tcp_connect_ms)) },
  { label: 'TLS', value: formatMs(firstNumber(httpPayload.value.tls_handshake_ms)) },
  { label: 'TTFB', value: formatMs(firstNumber(httpPayload.value.ttfb_ms)) },
])

const responseDetails = computed(() => [
  { label: label('协议', 'Protocol'), value: firstString(httpPayload.value.http_protocol, '-') },
  { label: label('重定向', 'Redirects'), value: String(firstNumber(httpPayload.value.redirect_count, props.httpRecord.redirects?.length ?? 0)) },
  { label: label('类型', 'Type'), value: firstString(httpPayload.value.content_type, headerValue('content-type'), '-') },
  { label: label('压缩', 'Encoding'), value: firstString(httpPayload.value.content_encoding, boolMaybeText(httpPayload.value.compressed), '-') },
])

const tlsSubtitle = computed(() => {
  const verified = httpPayload.value.cert_verified
  if (verified === true) {
    return label('证书校验', 'Certificate verification')
  }
  if (verified === false && firstString(httpPayload.value.verify_error)) {
    return label('证书校验需关注', 'Certificate needs review')
  }
  return label('安全加密协议', 'Secure protocol')
})

const tlsDetails = computed(() => [
  { label: label('校验', 'Verified'), value: boolText(httpPayload.value.cert_verified) },
  { label: label('握手', 'Handshake'), value: firstString(httpPayload.value.tls_handshake, '-') },
  { label: label('套件', 'Cipher'), value: firstString(httpPayload.value.cipher_suite, '-') },
  { label: label('签名', 'Signature'), value: firstString(httpPayload.value.cert_signature_algorithm, httpPayload.value.cert_sig_alg, '-') },
])

const certDetails = computed(() => [
  { label: label('签发者', 'Issuer'), value: firstString(httpPayload.value.cert_issuer_cn, httpPayload.value.cert_issuer, props.httpRecord.certIssuer, '-') },
  { label: label('到期', 'Expires'), value: shortDate(firstString(httpPayload.value.cert_not_after, httpPayload.value.cert_expiry, props.httpRecord.certExpiry)) },
  { label: label('SAN', 'SAN'), value: numberOrDash(firstNumber(httpPayload.value.cert_san_count, props.httpRecord.certDNSNames?.length)) },
  { label: label('链长', 'Chain'), value: numberOrDash(firstNumber(httpPayload.value.cert_chain_length)) },
])

// 核心颜色函数
const getColor = (metric: string, value: any) => {
  switch (metric) {
    case 'responseTime': {
      const time = parseInt(value)
      if (time < 300) return 'border-l-4 border-green-300 bg-green-50'
      if (time < 800) return 'border-l-4 border-yellow-300 bg-yellow-50'
      return 'border-l-4 border-red-300 bg-red-50'
    }
    case 'statusCode': {
      if (value >= 200 && value < 300) return 'border-l-4 border-green-300 bg-green-50'
      if (value >= 300 && value < 400) return 'border-l-4 border-yellow-300 bg-yellow-50'
      return 'border-l-4 border-red-300 bg-red-50'
    }
    case 'tlsVersion': {
      if (value.includes('1.3')) return 'border-l-4 border-green-300 bg-green-50'
      if (value.includes('1.2')) return 'border-l-4 border-yellow-300 bg-yellow-50'
      return 'border-l-4 border-red-300 bg-red-50'
    }
    case 'certDaysLeft': {
      const days = parseInt(String(value).replace(/[^0-9]/g, ''), 10) || 0
      if (days > 90) return 'border-l-4 border-green-300 bg-green-50'
      if (days > 30) return 'border-l-4 border-yellow-300 bg-yellow-50'
      return 'border-l-4 border-red-300 bg-red-50'
    }
  }
  return ''
}

const responseColor = computed(() => getColor('responseTime', metrics.value.responseTime))
const statusColor = computed(() => getColor('statusCode', metrics.value.statusCode))
const tlsColor = computed(() => getColor('tlsVersion', metrics.value.tlsVersion))
const certColor = computed(() => getColor('certDaysLeft', metrics.value.certDaysLeft))
const iconColor = (bgClass: string) =>
    bgClass.includes('green') ? 'text-green-500' : bgClass.includes('yellow') ? 'text-yellow-500' : 'text-red-500'

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
    return yesText.value
  }
  if (value === false) {
    return noText.value
  }
  return '-'
}

function boolMaybeText(value: unknown) {
  if (typeof value === 'boolean') {
    return boolText(value)
  }
  return ''
}

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}

function headerValue(key: string) {
  const normalizedKey = key.toLowerCase()
  for (const [name, values] of Object.entries(props.httpRecord.headers ?? {})) {
    if (name.toLowerCase() === normalizedKey && values?.length) {
      return values[0]
    }
  }
  return ''
}

// 初始化图表
function initChart() {
  if (!latencyChartRef.value) return
  if (chart.value) chart.value.dispose()
  chart.value = echarts.init(latencyChartRef.value, undefined, { renderer: 'canvas' })
}

// 更新图表
function updateChart() {
  if (!currentPing.value || !chart.value) return

  const tooltipData = [...currentPing.value.DelayModel].reverse()
  const seriesData = tooltipData.map(d => ({
    value: Number(d.delay) || 0,
    time: d.time || 'Unknown',
    status: d.status || 'Unknown',
    loss: d.loss || '0'
  }))
  const times = tooltipData.map(d => (d.time?.split(' ')[1]) || 'Unknown')

  chart.value.setOption({
    color: ['#4f6fed'],
    tooltip: {
      trigger: 'item',
      axisPointer: {
        type: 'line',
        lineStyle: { color: '#f59e0b', width: 1, type: 'dashed' }
      },
      confine: true,
      backgroundColor: 'rgba(255, 251, 245, 0.96)',
      borderColor: 'rgba(251, 146, 60, 0.28)',
      borderWidth: 1,
      borderRadius: 10,
      padding: [10, 12],
      textStyle: { color: '#374151', fontSize: 12, lineHeight: 18 },
      formatter: (params: any) => {
        const point = params.data
        if (!point) return label('无数据', 'No data')
        return `
        <div style="line-height:1.5">
          <div><strong>`+t('site.performance.time')+`:</strong> ${point.time}</div>
          <div><strong>`+t('site.performance.status')+`:</strong> ${point.status}</div>
          <div><strong>`+t('site.performance.packetLossRate')+`:</strong> ${point.loss}%</div>
          <div><strong>`+t('site.performance.latency')+`:</strong> ${point.value} ms</div>
        </div>
      `
      },
      extraCssText: 'max-width: 240px; white-space: normal; backdrop-filter: blur(10px);'
    },
    grid: { left: 48, right: 24, top: 42, bottom: 42, containLabel: true },
    xAxis: {
      type: 'category',
      data: times,
      boundaryGap: false,
      axisTick: { show: false },
      axisLine: { lineStyle: { color: '#e5d4bd' } },
      axisLabel: { color: '#8b8178', fontSize: 10, hideOverlap: true }
    },
    yAxis: {
      type: 'value',
      name: t('site.performance.latency')+' (ms)',
      nameTextStyle: { color: '#8b8178', fontSize: 11, padding: [0, 0, 14, 0] },
      axisLabel: { color: '#8b8178' },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { type: 'dashed', color: 'rgba(148, 163, 184, 0.38)' } }
    },
    series: [{
      name: t('site.performance.latency'),
      type: 'line',
      data: seriesData,
      smooth: true,
      symbol: 'circle',
      symbolSize: 5,
      showSymbol: true,
      lineStyle: { width: 2.5, color: '#4f6fed' },
      itemStyle: { color: '#4f6fed', borderColor: '#fff7ed', borderWidth: 1.5 },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(79, 111, 237, 0.26)' },
          { offset: 1, color: 'rgba(249, 115, 22, 0.04)' }
        ])
      },
      emphasis: {
        focus: 'series',
        itemStyle: { symbolSize: 8, color: '#f97316', borderColor: '#fff7ed', borderWidth: 2 }
      }
    }]
  })

  // 鼠标事件监听
  chart.value.getZr().off('mousemove') // 避免重复绑定
  chart.value.getZr().on('mousemove', (event) => {
    const pointInPixel: [number, number] = [
      event.offsetX ?? 0,
      event.offsetY ?? 0
    ];
    let nearestIndex = -1;
    let minDist = Infinity;

    seriesData.forEach((d, i) => {
      const pixel = chart.value!.convertToPixel({ seriesIndex: 0 }, [i, d.value]);
      if (!pixel) return;

      // 分解像素坐标时增加类型保护
      if (Array.isArray(pixel) && pixel.length === 2) {
        const px = Number(pixel[0]) || 0;
        const py = Number(pixel[1]) || 0;
        const dist = Math.hypot(px - pointInPixel[0], py - pointInPixel[1]);

        if (dist < minDist && dist <= 10) {
          minDist = dist;
          nearestIndex = i;
        }
      }
    });

    if (nearestIndex !== -1) {
      chart.value!.dispatchAction({ type: 'highlight', seriesIndex: 0, dataIndex: nearestIndex });
      chart.value!.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: nearestIndex });
    } else {
      chart.value!.dispatchAction({ type: 'downplay', seriesIndex: 0 });
      chart.value!.dispatchAction({ type: 'hideTip' });
    }
  });
}


// 切换抽样类型
function changeSample(type: 'twenty' | 'sixty' | 'hundred') {
  sampleType.value = type
  updateChart()
}

// ResizeObserver 自动适应尺寸
let resizeObserver: ResizeObserver | null = null

onMounted(async () => {
  await nextTick()
  initChart()
  // 执行两次才能正确初始化 TODO: 需要改进
  changeSample(sampleType.value)
  changeSample(sampleType.value)

  if (latencyChartRef.value) {
    resizeObserver = new ResizeObserver(() => chart.value?.resize())
    resizeObserver.observe(latencyChartRef.value)
  }
  void loadV2PingHistory()
})

onBeforeUnmount(() => {
  if (resizeObserver && latencyChartRef.value) resizeObserver.unobserve(latencyChartRef.value)
  resizeObserver = null
  if (chart.value)  chart.value.dispose()
})

// 监听 ping 数据变化
watch(() => props.pingRecord, updateChart, { deep: true })
watch(() => [props.siteId, props.domain], () => {
  v2PingRecord.value = null
  void loadV2PingHistory()
})
watch(sampleType, updateChart)

async function loadV2PingHistory() {
  if (props.pingRecord || !props.siteId || !props.domain) {
    return
  }

  try {
    const response = await navV2Api<TargetObservationsResponse>(`/nav/sites/${props.siteId}/targets/${encodeURIComponent(props.domain)}/observations`, {
      query: {
        protocol: 'ping',
        limit: 100,
        payload_mode: 'preview',
      },
    })
    v2PingRecord.value = buildPingRecordFromObservations(response.items ?? [])
    await nextTick()
    updateChart()
  } catch {
    v2PingRecord.value = emptyPingRecord()
  }
}

function buildPingRecordFromObservations(items: CollectorEnvelope[]): PingRecord {
  const points = items.map(toDelayPoint).filter(point => point.delay !== '')
  return {
    twenty: buildPingStats(points.slice(0, 20)),
    sixty: buildPingStats(points.slice(0, 60)),
    hundred: buildPingStats(points.slice(0, 100)),
  }
}

function toDelayPoint(item: CollectorEnvelope) {
  const payload = asRecord(item.payload)
  const delay = firstNumber(payload.avg_rtt_ms, payload.delay_ms, item.duration_ms)
  const loss = firstNumber(payload.loss_rate, payload.legacy_loss)
  return {
    status: item.status,
    time: formatObservedTime(item.observed_at),
    loss: formatLossNumber(loss),
    delay: delay === null ? '' : String(Math.round(delay)),
  }
}

function buildPingStats(points: { status: string; time: string; loss: string; delay: string }[]): PingStats {
  const delayValues = points.map(point => parseNumber(point.delay)).filter(Number.isFinite)
  const lossValues = points.map(point => parseNumber(point.loss)).filter(Number.isFinite)
  return {
    DelayModel: points,
    avgDelay: delayValues.length ? `${Math.round(delayValues.reduce((sum, value) => sum + value, 0) / delayValues.length)}ms` : '-',
    avgLoss: lossValues.length ? `${Math.round(lossValues.reduce((sum, value) => sum + value, 0) / lossValues.length)}%` : '-',
  }
}

function emptyPingRecord(): PingRecord {
  return {
    twenty: buildPingStats([]),
    sixty: buildPingStats([]),
    hundred: buildPingStats([]),
  }
}

function formatObservedTime(value: string) {
  if (!value) {
    return ''
  }
  return value.replace('T', ' ').replace(/\.\d+.*$/, '')
}

function formatLossNumber(value: number | null) {
  if (value === null) {
    return '0'
  }
  return value <= 1 ? String(Math.round(value * 100)) : String(Math.round(value))
}
</script>
