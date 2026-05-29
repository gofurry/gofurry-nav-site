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

    <!-- 延迟时序图 -->
    <div class="rounded-xl p-5">
      <div class="flex flex-col md:flex-row md:items-center md:justify-between mb-4">
        <div class="flex items-center mb-2 md:mb-0">
          <h3 class="font-semibold mr-4">{{ t('site.performance.latencyTrend') }}</h3>
          <div class="flex gap-2">
            <button
                v-for="type in ['twenty','sixty','hundred']"
                :key="type"
                @click="changeSample(type as 'twenty'|'sixty'|'hundred')"
                :class="[
                'px-3 py-1 rounded-lg text-sm',
                sampleType === type
                  ? 'bg-orange-200 text-gray-800'
                  : 'bg-orange-50 text-gray-700 hover:bg-orange-100'
              ]"
            >
              {{ type === 'twenty' ? '20次' : type === 'sixty' ? '60次' : '100次' }}
            </button>
          </div>
        </div>

        <!-- 平均延迟 / 丢包率 -->
        <div class="text-xs text-gray-500 md:ml-4">
          {{ t('site.performance.averageLatency') }}: {{ currentPing?.avgDelay || 'Unknown' }}，
          {{ t('site.performance.packetLossRate') }}: {{ currentPing?.avgLoss || 'Unknown' }}
        </div>
      </div>

      <div class="h-90 w-full" ref="latencyChartRef"></div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { PingRecord, PingStats, HttpRecord, TargetLatestResponse } from '@/types/nav'
import {i18n} from "@/main";

const t = (key: string) => i18n.global.t(key)

interface Props {
  pingRecord: PingRecord
  httpRecord: HttpRecord
  targetLatestCore?: TargetLatestResponse | null
}

const props = defineProps<Props>()
const latencyChartRef = ref<HTMLElement | null>(null)
const chart = ref<echarts.ECharts | null>(null)
const sampleType = ref<'twenty' | 'sixty' | 'hundred'>('twenty')
const httpPayload = computed(() => asRecord(props.targetLatestCore?.protocols?.http?.payload))
const yesText = computed(() => i18n.global.locale.value === 'en' ? 'Yes' : '是')
const noText = computed(() => i18n.global.locale.value === 'en' ? 'No' : '否')

// 当前 ping 数据
const currentPing = computed<PingStats | null>(() => props.pingRecord?.[sampleType.value] || null)

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
    tooltip: {
      trigger: 'item',
      axisPointer: { type: 'none' },
      confine: true,
      backgroundColor: 'rgba(0,0,0,0.75)',
      borderRadius: 8,
      padding: [8, 12],
      textStyle: { color: '#fff', fontSize: 12, lineHeight: 18 },
      formatter: (params: any) => {
        const point = params.data
        if (!point) return '无数据'
        return `
        <div style="line-height:1.5">
          <div><strong>`+t('site.performance.time')+`:</strong> ${point.time}</div>
          <div><strong>`+t('site.performance.status')+`:</strong> ${point.status}</div>
          <div><strong>`+t('site.performance.packetLossRate')+`:</strong> ${point.loss}%</div>
          <div><strong>`+t('site.performance.latency')+`:</strong> ${point.value} ms</div>
        </div>
      `
      },
      extraCssText: 'max-width: 220px; white-space: normal;'
    },
    grid: { left: '3%', right: '3%', top: '8%', bottom: '10%', containLabel: true },
    xAxis: { type: 'category', data: times, axisLabel: { color: '#888', fontSize: 10 } },
    yAxis: { type: 'value', name: t('site.performance.latency')+' (ms)', axisLabel: { color: '#888' }, splitLine: { lineStyle: { type: 'dashed', color: '#ccc' } } },
    series: [{
      name: t('site.performance.latency'),
      type: 'line',
      data: seriesData,
      smooth: true,
      symbol: 'circle',
      symbolSize: 5,
      lineStyle: { width: 2 },
      areaStyle: { opacity: 0.15 },
      emphasis: { itemStyle: { symbolSize: 8, color: '#f97316' } }
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
})

onBeforeUnmount(() => {
  if (resizeObserver && latencyChartRef.value) resizeObserver.unobserve(latencyChartRef.value)
  resizeObserver = null
  if (chart.value)  chart.value.dispose()
})

// 监听 ping 数据变化
watch(() => props.pingRecord, updateChart, { deep: true })
watch(sampleType, updateChart)
</script>
