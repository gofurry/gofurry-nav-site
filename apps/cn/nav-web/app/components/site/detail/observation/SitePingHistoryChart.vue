<template>
  <div class="relative min-w-0" :aria-busy="!ready">
    <p v-if="!ready" role="status" class="site-detail-note absolute inset-x-0 top-4 text-center">{{ t('siteObservation.chartLoading') }}</p>
    <div ref="element" data-site-performance-chart :data-site-chart-ready="ready" role="img" :aria-label="t('siteObservation.history')" class="site-observation-chart w-full min-w-0" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { useThemeStore } from '~/stores/theme'
import type { PingHistoryRow } from '~/utils/siteObservationPresentation'
const props = defineProps<{ rows: PingHistoryRow[] }>()
const { t, locale } = useI18n()
const theme = useThemeStore()
const element = ref<HTMLElement | null>(null), ready = ref(false)
const chart = shallowRef<import('echarts').ECharts | null>(null)
const themeKey = computed(() => theme.theme)
let active = false, revision = 0
let resize: ResizeObserver | null = null
async function renderChart() {
  const version = ++revision
  await nextTick()
  if (!active || !element.value) return
  const echarts = await import('echarts')
  if (!active || !element.value || version !== revision) return
  const styles = getComputedStyle(element.value)
  const color = (name: string) => styles.getPropertyValue(name).trim()
  if (!chart.value) chart.value = echarts.init(element.value, undefined, { renderer: 'canvas' })
  const rows = [...props.rows].reverse()
  chart.value.setOption({ animation: false,
    grid: { top: 24, right: 16, bottom: 40, left: 52 },
    tooltip: { trigger: 'axis', confine: true, renderMode: 'richText', backgroundColor: color('--gf-surface'),
      borderColor: color('--gf-border'), textStyle: { color: color('--gf-text-main') } },
    xAxis: { type: 'category', data: rows.map(row => row.time), axisTick: { show: false },
      axisLine: { lineStyle: { color: color('--gf-border') } },
      axisLabel: { color: color('--gf-text-muted'), hideOverlap: true, formatter: (value: string) => value.slice(5, 16) } },
    yAxis: { type: 'value', min: 0, name: 'ms', nameTextStyle: { color: color('--gf-text-muted') },
      axisLabel: { color: color('--gf-text-muted') }, splitLine: { lineStyle: { color: color('--gf-border') } } },
    series: [{ name: t('siteObservation.fields.rtt'), type: 'line', data: rows.map(row => row.rtt), connectNulls: false,
      symbolSize: 6, showSymbol: rows.length <= 20, lineStyle: { color: color('--site-detail-positive'), width: 2 },
      itemStyle: { color: color('--site-detail-positive') } }],
  }, true)
  ready.value = true
}
onMounted(() => {
  active = true
  resize = new ResizeObserver(() => chart.value?.resize())
  if (element.value) resize.observe(element.value)
  void renderChart()
})
watch(() => [props.rows, themeKey.value, locale.value], () => { void renderChart() }, { flush: 'post' })
onBeforeUnmount(() => { active = false; revision++; resize?.disconnect(); chart.value?.dispose(); chart.value = null })
</script>
