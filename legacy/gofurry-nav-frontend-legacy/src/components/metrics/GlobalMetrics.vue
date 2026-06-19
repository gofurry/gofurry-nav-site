<template>
  <div class="space-y-6">
    <NodeMetrics :node-cards="nodeCards" />

    <section class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <ServiceSummary :title='"Nav "+t("metrics.service")' :items="navSummary" />
      <ServiceSummary :title='"Game "+t("metrics.service")' :items="gameSummary" />
    </section>

    <MetricsTrend />

    <section class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <PathMetricsTable
          :title='"Nav API "+t("metrics.avgResponse")+" (1h)"'
          :data="sortedNavPath"
      />
      <PathMetricsTable
          :title='"Game API "+t("metrics.avgResponse")+" (1h)"'
          :data="sortedGamePath"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getPromMetrics } from '@/utils/api/stat.ts'
import type { PromMetricsModel } from '@/types/stat.ts'
import NodeMetrics from "@/components/metrics/NodeMetrics.vue";
import ServiceSummary from "@/components/metrics/ServiceSummary.vue";
import PathMetricsTable from "@/components/metrics/PathMetricsTable.vue";
import MetricsTrend from "@/components/metrics/MetricsTrend.vue";
import { i18n } from '@/main.ts'

const { t } = i18n.global
const metrics = ref<PromMetricsModel | null>(null)

onMounted(async () => {
  metrics.value = await getPromMetrics()
})

// ================= Node Cards =================
const nodeCards = computed(() => {
  const n = metrics.value?.node
  if (!n) return []

  return [
    { label: t("metrics.cpuUsage"), value: Number(n.cpu_usage).toFixed(2), suffix: '%' },
    { label: t("metrics.memoryUsage"), value: formatBytes(n.mem_usage), suffix: '' },
    { label: t("metrics.diskUsage"), value: Number(n.disk_usage).toFixed(2), suffix: '%' },
    { label: t("metrics.netIn")+'(1d)', value: formatBytes(n.net_rx_1d), suffix: '' },
    { label: t("metrics.netOut")+'(1d)', value: formatBytes(n.net_tx_1d), suffix: '' },
    { label: t("metrics.uptime"), value: formatUptime(n.uptime), suffix: '' }
  ]
})

 // ================= Summary =================
const navSummary = computed(() => {
  const s = metrics.value?.nav
  if (!s) return []

  return [
    { label: t("metrics.avgResponse")+'(1h)', value: formatDuration(s.avg_response_1h) },
    { label: t("metrics.p99Response")+'(1h)', value: formatDuration(s.p99_response_1h) },
    { label: t("metrics.p95Response")+'(1h)', value: formatDuration(s.p95_response_1h) },
    { label: t("metrics.errorRate")+'(1h)', value: formatFailRate(s.fail_rate_1h) },
    { label: t("metrics.requestCount")+'(1d)', value: formatCount(s.http_requests_1d), suffix: ' '+t("common.times") },
    { label: t("metrics.requestCount")+'(7d)', value: formatCount(s.http_requests_7d), suffix: ' '+t("common.times") }
  ]
})

const gameSummary = computed(() => {
  const s = metrics.value?.game
  if (!s) return []

  return [
    { label: t("metrics.avgResponse")+'(1h)', value: formatDuration(s.avg_response_1h) },
    { label: t("metrics.p99Response")+'(1h)', value: formatDuration(s.p99_response_1h) },
    { label: t("metrics.p95Response")+'(1h)', value: formatDuration(s.p95_response_1h) },
    { label: t("metrics.errorRate")+'(1h)', value: formatFailRate(s.fail_rate_1h) },
    { label: t("metrics.requestCount")+'(1d)', value: formatCount(s.http_requests_1d), suffix: ' '+t("common.times") },
    { label: t("metrics.requestCount")+'(7d)', value: formatCount(s.http_requests_7d), suffix: ' '+t("common.times") }
  ]
})

 // ================= Path Sort =================
const sortedNavPath = computed(() => sortByValue(metrics.value?.nav_path.avg_response_1h))
const sortedGamePath = computed(() => sortByValue(metrics.value?.game_path.avg_response_1h))

function sortByValue(data?: Record<string, string>) {
  if (!data) return {}
  return Object.fromEntries(
      Object.entries(data).sort((a, b) => Number(b[1]) - Number(a[1]))
  )
}

 // ================= Utils =================
function formatBytes(val?: string) {
  if (!val) return '-'
  const num = Number(val)
  if (num < 1024) return `${num.toFixed(0)} B`
  if (num < 1024 ** 2) return `${(num / 1024).toFixed(1)} KB`
  if (num < 1024 ** 3) return `${(num / 1024 ** 2).toFixed(1)} MB`
  return `${(num / 1024 ** 3).toFixed(1)} GB`
}

function formatUptime(val?: string) {
  if (!val) return '-'
  const sec = Number(val)
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  return `${d}`+t("common.shortDay")+` ${h}`+t("common.shortHour")
}

function formatDuration(sec?: string) {
  if (!sec) return '-'
  const s = Number(sec)
  if (s < 1) return `${Math.round(s * 1000)} ms`
  return `${s.toFixed(2)} s`
}

function formatFailRate(rate?: string) {
  if (!rate) return '-'
  const r = Number(rate)
  if (r >= 0.01) {
    return (r * 100).toFixed(2) + '%'
  }
  return (r * 10000).toFixed(2) + '‱'
}

function formatCount(val?: string) {
  if (!val) return '-'
  const n = Math.floor(Number(val))
  if (n >= 10000) {
    return `${(n / 10000).toFixed(1)}`+t("common.tenThousand")
  }
  return `${n}`
}

</script>
