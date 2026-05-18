<script setup lang="ts">
import { Activity, AlertTriangle, Gauge, HardDrive, MemoryStick, Server } from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import * as api from '../api'
import MetricChart from '../components/MetricChart.vue'
import RangeControl from '../components/RangeControl.vue'
import {
  alertClass,
  formatDurationSeconds,
  formatNumber,
  formatPercent,
  formatTime,
  statusClass,
} from '../format'
import type { MetricSeries, MetricsRange, Overview, OverviewMetrics, TopResource } from '../types'
import { useAutoRefresh } from '../useAutoRefresh'

const range = ref<MetricsRange>('1h')
const overview = ref<Overview | null>(null)
const metrics = ref<OverviewMetrics | null>(null)
const loading = ref(false)
const error = ref('')

const services = computed(() => overview.value?.services ?? [])
const alerts = computed(() => overview.value?.alerts ?? [])
const latest = computed(() => metrics.value?.latest_system)
const freshness = computed(() => formatDurationSeconds(metrics.value?.sample_freshness_seconds))
const serviceTotal = computed(() => services.value.length || sumCounts(metrics.value?.service_status_counts ?? []))
const serviceOk = computed(() => countByName(metrics.value?.service_status_counts ?? [], ['ok', 'up', 'success', 'healthy']))
const serviceBad = computed(() => serviceTotal.value - serviceOk.value)
const alertTotal = computed(() => sumCounts(metrics.value?.alert_level_counts ?? []))
const cpuSeries = computed<MetricSeries[]>(() => [{ name: 'CPU', unit: '%', points: metrics.value?.cpu_trend ?? [] }])
const memorySeries = computed<MetricSeries[]>(() => [{ name: 'Memory', unit: '%', points: metrics.value?.memory_trend ?? [] }])
const loadLatencySeries = computed<MetricSeries[]>(() => [
  { name: 'Load', points: metrics.value?.load_trend ?? [] },
  { name: 'Latency', unit: 'ms', points: metrics.value?.latency_trend ?? [] },
])
const diskSeries = computed<MetricSeries[]>(() => [{ name: 'Disk High', unit: '%', points: metrics.value?.disk_trend ?? [] }])

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [nextMetrics, nextOverview] = await Promise.all([api.overviewMetrics(range.value), api.overview()])
    metrics.value = nextMetrics
    overview.value = nextOverview
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function countByName(items: Array<{ name: string; count: number }>, names: string[]) {
  const accepted = new Set(names)
  return items.reduce((total, item) => total + (accepted.has(item.name.toLowerCase()) ? item.count : 0), 0)
}

function sumCounts(items: Array<{ count: number }>) {
  return items.reduce((total, item) => total + item.count, 0)
}

function resourceLabel(item: TopResource) {
  return item.name ? `${item.node_id} · ${item.name}` : item.node_id
}

function resourceValue(item: TopResource) {
  return item.unit === '%' ? formatPercent(item.value) : formatNumber(item.value)
}

onMounted(load)
watch(range, load)
useAutoRefresh(load)
</script>

<template>
  <div class="space-y-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="fine-label">Monitoring</p>
        <h2 class="mt-1 text-xl font-semibold">实时监控面板</h2>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-sm text-[var(--ops-muted)]">刷新 {{ freshness }}</span>
        <RangeControl v-model="range" />
      </div>
    </div>

    <div v-if="error" class="panel border-[var(--ops-red)] p-4 text-sm text-[var(--ops-red)]">{{ error }}</div>
    <div v-if="loading && !metrics" class="empty">加载中</div>

    <template v-if="metrics">
      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">节点</p>
            <Server class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ overview?.nodes_total ?? 0 }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">Down {{ overview?.nodes_down ?? 0 }} · Samples {{ latest?.nodes_reported ?? 0 }}</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">CPU</p>
            <Gauge class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ formatPercent(latest?.cpu_avg) }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">Load {{ formatNumber(latest?.load1_avg) }}</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">内存</p>
            <MemoryStick class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ formatPercent(latest?.memory_avg) }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">平均使用率</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">磁盘高水位</p>
            <HardDrive class="size-5 text-[var(--ops-amber)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ formatPercent(metrics.highest_disk?.value) }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">{{ metrics.highest_disk ? resourceLabel(metrics.highest_disk) : '-' }}</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">服务 / 告警</p>
            <Activity class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ serviceOk }}/{{ serviceTotal }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">{{ serviceBad }} 异常 · {{ alertTotal }} 告警</p>
        </div>
      </section>

      <section class="grid gap-5 xl:grid-cols-2">
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">CPU 趋势</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="cpuSeries" />
        </div>
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">内存趋势</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="memorySeries" />
        </div>
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">Load / 延迟</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="loadLatencySeries" />
        </div>
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">磁盘高水位趋势</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="diskSeries" />
        </div>
      </section>

      <section class="grid gap-5 xl:grid-cols-[1fr_1fr_0.9fr]">
        <div class="panel p-4">
          <h3 class="section-title">资源排行</h3>
          <div class="mt-4 grid gap-4 md:grid-cols-3 xl:grid-cols-1">
            <div>
              <p class="fine-label">CPU</p>
              <div class="mt-3 space-y-3">
                <div v-for="item in metrics.top_cpu" :key="`cpu-${resourceLabel(item)}`" class="space-y-2">
                  <div class="flex justify-between gap-3 text-sm">
                    <span class="truncate">{{ resourceLabel(item) }}</span>
                    <span>{{ resourceValue(item) }}</span>
                  </div>
                  <div class="meter"><span :style="{ width: `${Math.min(item.value, 100)}%` }"></span></div>
                </div>
                <div v-if="metrics.top_cpu.length === 0" class="empty py-5">暂无 CPU 样本</div>
              </div>
            </div>
            <div>
              <p class="fine-label">Memory</p>
              <div class="mt-3 space-y-3">
                <div v-for="item in metrics.top_memory" :key="`mem-${resourceLabel(item)}`" class="space-y-2">
                  <div class="flex justify-between gap-3 text-sm">
                    <span class="truncate">{{ resourceLabel(item) }}</span>
                    <span>{{ resourceValue(item) }}</span>
                  </div>
                  <div class="meter"><span :style="{ width: `${Math.min(item.value, 100)}%` }"></span></div>
                </div>
                <div v-if="metrics.top_memory.length === 0" class="empty py-5">暂无内存样本</div>
              </div>
            </div>
            <div>
              <p class="fine-label">Disk</p>
              <div class="mt-3 space-y-3">
                <div v-for="item in metrics.top_disk" :key="`disk-${resourceLabel(item)}`" class="space-y-2">
                  <div class="flex justify-between gap-3 text-sm">
                    <span class="truncate">{{ resourceLabel(item) }}</span>
                    <span>{{ resourceValue(item) }}</span>
                  </div>
                  <div class="meter warn"><span :style="{ width: `${Math.min(item.value, 100)}%` }"></span></div>
                </div>
                <div v-if="metrics.top_disk.length === 0" class="empty py-5">暂无磁盘样本</div>
              </div>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
            <h3 class="section-title">服务健康</h3>
            <span :class="statusClass(overview?.status)">{{ overview?.status || 'unknown' }}</span>
          </div>
          <div v-if="services.length === 0" class="p-4"><div class="empty">暂无服务样本</div></div>
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>服务</th>
                  <th>节点</th>
                  <th>状态</th>
                  <th>延迟</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="service in services.slice(0, 8)" :key="service.key">
                  <td>
                    <p class="font-medium">{{ service.name }}</p>
                    <p class="text-xs text-[var(--ops-muted)]">{{ service.service_type }}</p>
                  </td>
                  <td>{{ service.node_id }}</td>
                  <td><span :class="statusClass(service.status)">{{ service.status }}</span></td>
                  <td>{{ service.latency_ms || 0 }} ms</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="panel">
          <div class="flex items-center gap-2 border-b border-[var(--ops-border)] px-4 py-3">
            <AlertTriangle class="size-4 text-[var(--ops-amber)]" />
            <h3 class="section-title">活跃告警</h3>
          </div>
          <div v-if="alerts.length === 0" class="p-4"><div class="empty">暂无活跃告警</div></div>
          <div v-else class="divide-y divide-[var(--ops-border)]">
            <div v-for="alert in alerts.slice(0, 8)" :key="alert.key" class="space-y-2 p-4">
              <div class="flex items-center justify-between gap-3">
                <p class="font-medium">{{ alert.title }}</p>
                <span :class="alertClass(alert.level)">{{ alert.level }}</span>
              </div>
              <p class="text-sm text-[var(--ops-muted)]">{{ alert.message || alert.type }}</p>
              <p class="text-xs text-[var(--ops-muted)]">{{ alert.node_id || alert.region }} · {{ formatTime(alert.last_seen_at) }}</p>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
