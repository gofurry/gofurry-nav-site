<template>
  <section data-site-insight-trend :data-site-insight-trend-state="surfaceState" :data-site-trend-metric="metric" :data-site-trend-range="range" class="mt-6" aria-labelledby="site-ecosystem-trend-title">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-baseline gap-x-2 gap-y-1"><h4 id="site-ecosystem-trend-title" class="site-overview-title">{{ t('siteIntelligence.trendTitle') }}</h4><span class="site-detail-note">{{ t('siteIntelligence.trendScope') }}</span></div>
      <div class="site-detail-segmented flex" role="group" :aria-label="t('siteIntelligence.range')">
        <button v-for="option in siteInsightRanges" :key="option" type="button" :data-site-insight-range="option" :aria-pressed="option === range" class="site-intelligence-range" @click="emit('range', option)">{{ t('insights.ranges.' + option) }}</button>
      </div>
    </div>
    <div class="site-intelligence-chart-shell relative mt-3 min-w-0" :aria-busy="surfaceState === 'loading'">
      <div ref="element" data-site-insight-chart :data-site-chart-ready="ready && presentation.state === 'ready'" role="img" :aria-label="t('siteIntelligence.trendTitle')" :class="{ invisible: surfaceState !== 'ready' }" class="site-intelligence-chart w-full min-w-0" />
      <div v-if="surfaceState !== 'ready'" data-site-trend-message role="status" class="absolute inset-0 flex flex-col items-center justify-center gap-3 text-center">
        <p class="site-detail-note">{{ t('siteIntelligence.trendStates.' + surfaceState) }}</p>
        <button v-if="surfaceState === 'unavailable'" data-site-insight-trend-retry type="button" class="site-detail-text-link" @click="retry">{{ t('siteIntelligence.retry') }}</button>
      </div>
    </div>
    <p v-if="surfaceState === 'ready' && !values.some(value => value !== null)" data-site-trend-no-values class="site-detail-note mt-2">{{ t('siteIntelligence.noObservedValues') }}</p>
  </section>
</template>
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { useThemeStore } from '~/stores/theme'
import { siteInsightRanges, type SiteInsightMetric, type SiteInsightRange } from '~/utils/siteDetailRouteState'
import { formatChartPercent } from '~/utils/detailChartValues'
import { siteInsightAdoptionValues } from '~/utils/siteInsightsPresentation'
import type { SiteInsightTrendPresentation } from '~/composables/useSiteInsightTrend'
const props = defineProps<{ presentation: SiteInsightTrendPresentation; metric: SiteInsightMetric; range: SiteInsightRange }>()
const emit = defineEmits<{ range: [value: SiteInsightRange]; retry: [] }>()
const { t, locale } = useI18n(), theme = useThemeStore()
const element = ref<HTMLElement | null>(null), ready = ref(false), failed = ref(false)
const chart = shallowRef<import('echarts').ECharts | null>(null)
const values = computed(() => siteInsightAdoptionValues(props.presentation.points))
const surfaceState = computed(() => props.presentation.state !== 'ready' ? props.presentation.state : failed.value ? 'unavailable' : ready.value ? 'ready' : 'loading')
let active = false, revision = 0
let resize: ResizeObserver | null = null
async function renderChart() {
  const version = ++revision
  ready.value = false; failed.value = false
  await nextTick()
  if (!active || !element.value || props.presentation.state !== 'ready') return
  try {
    const echarts = await import('echarts')
    if (!active || !element.value || version !== revision) return
    const style = getComputedStyle(element.value), color = (key: string) => style.getPropertyValue(key).trim()
    if (!chart.value) chart.value = echarts.init(element.value, undefined, { renderer: 'canvas' })
    chart.value.setOption({ animation: false,
      grid: { top: 20, right: 16, bottom: 40, left: 48 },
      tooltip: { trigger: 'axis', confine: true, renderMode: 'richText', backgroundColor: color('--gf-surface'),
        borderWidth: 0, textStyle: { color: color('--gf-text-main') }, valueFormatter: formatChartPercent },
      xAxis: { type: 'category', boundaryGap: false, data: props.presentation.points.map(point => point.date), axisTick: { show: false },
        axisLine: { show: false }, axisLabel: { color: color('--gf-text-muted'), hideOverlap: true } },
      yAxis: { type: 'value', min: 0, max: 100, axisLabel: { color: color('--gf-text-muted'), formatter: '{value}%' }, splitLine: { lineStyle: { color: color('--site-detail-border'), opacity: 0.3 } } },
      series: [{ name: t('siteIntelligence.adoption'), type: 'line', data: values.value, connectNulls: false, symbolSize: 6,
        showSymbol: props.presentation.points.length <= 31, lineStyle: { width: 2, color: style.color }, itemStyle: { color: style.color }, areaStyle: { color: style.color, opacity: 0.05 } }],
    }, true)
    ready.value = true
  } catch { if (active && version === revision) failed.value = true }
}
function retry() { failed.value = false; emit('retry') }
onMounted(() => { active = true; resize = new ResizeObserver(() => chart.value?.resize()); if (element.value) resize.observe(element.value); void renderChart() })
watch(() => [props.metric, props.range, props.presentation, theme.theme, locale.value], () => { void renderChart() }, { flush: 'post' })
onBeforeUnmount(() => { active = false; revision++; resize?.disconnect(); chart.value?.dispose(); chart.value = null })
</script>
