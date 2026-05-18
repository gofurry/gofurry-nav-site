<script setup lang="ts">
import { Activity, Boxes, Database, Gauge, HardDrive, Network, ShieldCheck } from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import * as api from '../api'
import MetricChart from '../components/MetricChart.vue'
import RangeControl from '../components/RangeControl.vue'
import {
  booleanText,
  formatBytes,
  formatDurationSeconds,
  formatNumber,
  formatPercent,
  formatRate,
  formatTime,
  statusClass,
} from '../format'
import type { MetricSeries, MetricsRange, NodeMetrics } from '../types'
import { useAutoRefresh } from '../useAutoRefresh'

const props = defineProps<{ id: string }>()
const range = ref<MetricsRange>('1h')
const metrics = ref<NodeMetrics | null>(null)
const loading = ref(false)
const error = ref('')

const node = computed(() => metrics.value?.node)
const latest = computed(() => metrics.value?.latest)
const system = computed(() => latest.value?.system)
const freshness = computed(() => formatDurationSeconds(metrics.value?.sample_freshness_seconds))
const cpuSeries = computed<MetricSeries[]>(() => [{ name: 'CPU', unit: '%', points: metrics.value?.trends.cpu ?? [] }])
const memorySeries = computed<MetricSeries[]>(() => [{ name: 'Memory', unit: '%', points: metrics.value?.trends.memory ?? [] }])
const loadSeries = computed<MetricSeries[]>(() => [{ name: 'Load', points: metrics.value?.trends.load ?? [] }])
const networkSeries = computed<MetricSeries[]>(() => [
  { name: 'RX', unit: 'B/s', points: metrics.value?.trends.network_rx ?? [] },
  { name: 'TX', unit: 'B/s', points: metrics.value?.trends.network_tx ?? [] },
])
const serviceLatencySeries = computed<MetricSeries[]>(() => metrics.value?.trends.service_latency ?? [])

async function load() {
  loading.value = true
  error.value = ''
  try {
    metrics.value = await api.nodeMetrics(props.id, range.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function meterWidth(value?: number) {
  return `${Math.min(Math.max(value ?? 0, 0), 100)}%`
}

onMounted(load)
watch(() => props.id, load)
watch(range, load)
useAutoRefresh(load)
</script>

<template>
  <div class="space-y-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="fine-label">Node Monitor</p>
        <h2 class="mt-1 text-xl font-semibold">{{ node?.display_name || node?.node_id || props.id }}</h2>
        <p class="mt-1 text-sm text-[var(--ops-muted)]">{{ node?.region || '-' }} · {{ node?.role || 'node' }} · 刷新 {{ freshness }}</p>
      </div>
      <RangeControl v-model="range" />
    </div>

    <div v-if="error" class="panel p-4 text-sm text-[var(--ops-red)]">{{ error }}</div>
    <div v-if="loading && !metrics" class="empty">加载中</div>

    <template v-if="metrics">
      <section class="panel p-5">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p class="text-sm text-[var(--ops-muted)]">{{ node?.node_id }}</p>
            <h3 class="mt-1 text-2xl font-semibold">{{ node?.display_name || node?.node_id }}</h3>
            <p class="mt-1 text-sm text-[var(--ops-muted)]">Agent {{ node?.agent_version || '-' }} · 最后心跳 {{ formatTime(node?.last_seen_at) }}</p>
          </div>
          <span :class="statusClass(node?.status)">{{ node?.status || 'unknown' }}</span>
        </div>
        <dl class="mt-6 grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <div class="metric">
            <dt class="flex items-center justify-between text-sm text-[var(--ops-muted)]">
              CPU <Gauge class="size-4" />
            </dt>
            <dd class="mt-2 text-2xl font-semibold">{{ formatPercent(system?.cpu_usage) }}</dd>
            <div class="meter mt-3"><span :style="{ width: meterWidth(system?.cpu_usage) }"></span></div>
          </div>
          <div class="metric">
            <dt class="flex items-center justify-between text-sm text-[var(--ops-muted)]">
              内存 <Activity class="size-4" />
            </dt>
            <dd class="mt-2 text-2xl font-semibold">{{ formatPercent(system?.memory_usage) }}</dd>
            <p class="mt-1 text-xs text-[var(--ops-muted)]">{{ formatBytes(system?.memory_used) }} / {{ formatBytes(system?.memory_total) }}</p>
            <div class="meter mt-3"><span :style="{ width: meterWidth(system?.memory_usage) }"></span></div>
          </div>
          <div class="metric">
            <dt class="text-sm text-[var(--ops-muted)]">Load</dt>
            <dd class="mt-2 text-2xl font-semibold">{{ formatNumber(system?.load1) }}</dd>
            <p class="mt-1 text-xs text-[var(--ops-muted)]">{{ formatNumber(system?.load5) }} / {{ formatNumber(system?.load15) }}</p>
          </div>
          <div class="metric">
            <dt class="text-sm text-[var(--ops-muted)]">Uptime</dt>
            <dd class="mt-2 text-2xl font-semibold">{{ formatDurationSeconds(system?.uptime_seconds) }}</dd>
            <p class="mt-1 text-xs text-[var(--ops-muted)]">采样 {{ formatTime(system?.reported_at) }}</p>
          </div>
          <div class="metric">
            <dt class="text-sm text-[var(--ops-muted)]">样本</dt>
            <dd class="mt-2 text-2xl font-semibold">{{ freshness }}</dd>
            <p class="mt-1 text-xs text-[var(--ops-muted)]">{{ formatTime(metrics.last_sample_at) }}</p>
          </div>
        </dl>
      </section>

      <section class="grid gap-5 xl:grid-cols-2">
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">CPU</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="cpuSeries" />
        </div>
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">内存</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="memorySeries" />
        </div>
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">Load</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="loadSeries" />
        </div>
        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">网络吞吐</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="networkSeries" />
        </div>
      </section>

      <section class="grid gap-5 xl:grid-cols-[0.9fr_1.1fr]">
        <div class="panel p-4">
          <div class="mb-3 flex items-center gap-2">
            <HardDrive class="size-4 text-[var(--ops-amber)]" />
            <h3 class="section-title">磁盘分区</h3>
          </div>
          <div v-if="latest?.disks.length === 0" class="empty">暂无磁盘样本</div>
          <div v-else class="space-y-4">
            <div v-for="disk in latest?.disks" :key="disk.mount" class="space-y-2">
              <div class="flex justify-between gap-3 text-sm">
                <span class="font-medium">{{ disk.mount }}</span>
                <span>{{ formatPercent(disk.usage) }}</span>
              </div>
              <div class="meter" :class="{ warn: disk.usage >= 85 }"><span :style="{ width: meterWidth(disk.usage) }"></span></div>
              <p class="text-xs text-[var(--ops-muted)]">{{ formatBytes(disk.used) }} / {{ formatBytes(disk.total) }} · inode {{ formatPercent(disk.inode_usage) }}</p>
            </div>
          </div>
        </div>

        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">磁盘趋势</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="metrics.trends.disk_usage" />
        </div>
      </section>

      <section class="grid gap-5 xl:grid-cols-2">
        <div class="panel">
          <div class="flex items-center gap-2 border-b border-[var(--ops-border)] px-4 py-3">
            <Network class="size-4 text-[var(--ops-teal)]" />
            <h3 class="section-title">网络接口</h3>
          </div>
          <div v-if="latest?.networks.length === 0" class="p-4"><div class="empty">暂无网络样本</div></div>
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>接口</th>
                  <th>RX</th>
                  <th>TX</th>
                  <th>包</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="network in latest?.networks" :key="network.name">
                  <td class="font-medium">{{ network.name }}</td>
                  <td>{{ formatRate(network.rx_bytes_per_sec) }}</td>
                  <td>{{ formatRate(network.tx_bytes_per_sec) }}</td>
                  <td class="text-sm text-[var(--ops-muted)]">{{ network.packets_recv }} / {{ network.packets_sent }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="panel">
          <div class="flex items-center gap-2 border-b border-[var(--ops-border)] px-4 py-3">
            <Boxes class="size-4 text-[var(--ops-teal)]" />
            <h3 class="section-title">Docker 容器</h3>
          </div>
          <div v-if="latest?.docker.length === 0" class="p-4"><div class="empty">暂无容器样本</div></div>
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>容器</th>
                  <th>状态</th>
                  <th>健康</th>
                  <th>重启</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in latest?.docker" :key="item.name">
                  <td>
                    <p class="font-medium">{{ item.name }}</p>
                    <p class="text-xs text-[var(--ops-muted)]">{{ item.error_message || item.status }}</p>
                  </td>
                  <td><span :class="statusClass(item.running ? 'ok' : 'down')">{{ item.running ? 'running' : 'down' }}</span></td>
                  <td>{{ item.health_status || '-' }}</td>
                  <td>{{ item.restart_count }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section class="grid gap-5 xl:grid-cols-[1.1fr_0.9fr]">
        <div class="panel">
          <div class="flex items-center gap-2 border-b border-[var(--ops-border)] px-4 py-3">
            <Database class="size-4 text-[var(--ops-teal)]" />
            <h3 class="section-title">服务检查</h3>
          </div>
          <div v-if="latest?.http_checks.length === 0 && latest?.service_checks.length === 0" class="p-4"><div class="empty">暂无服务检查样本</div></div>
          <div v-else class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>服务</th>
                  <th>状态</th>
                  <th>延迟</th>
                  <th>指标</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in latest?.http_checks" :key="`http-${item.name}`">
                  <td>
                    <p class="font-medium">{{ item.name }}</p>
                    <p class="text-xs text-[var(--ops-muted)]">http · {{ item.status_code || '-' }}</p>
                  </td>
                  <td><span :class="statusClass(item.status)">{{ item.status }}</span></td>
                  <td>{{ item.latency_ms }} ms</td>
                  <td class="text-sm text-[var(--ops-muted)]">{{ item.error_message || item.url || '-' }}</td>
                </tr>
                <tr v-for="item in latest?.service_checks" :key="`${item.service_type}-${item.name}`">
                  <td>
                    <p class="font-medium">{{ item.name }}</p>
                    <p class="text-xs text-[var(--ops-muted)]">{{ item.service_type }}</p>
                  </td>
                  <td><span :class="statusClass(item.status)">{{ item.status }}</span></td>
                  <td>{{ item.latency_ms }} ms</td>
                  <td class="text-sm text-[var(--ops-muted)]">
                    conn {{ item.connections || 0 }} · mem {{ formatBytes(item.memory_used) }} · keys {{ item.key_count || 0 }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="panel p-4">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="section-title">服务延迟趋势</h3>
            <span class="text-xs text-[var(--ops-muted)]">{{ range }}</span>
          </div>
          <MetricChart :series="serviceLatencySeries" />
        </div>
      </section>

      <section class="panel">
        <div class="flex items-center gap-2 border-b border-[var(--ops-border)] px-4 py-3">
          <ShieldCheck class="size-4 text-[var(--ops-teal)]" />
          <h3 class="section-title">证书</h3>
        </div>
        <div v-if="latest?.certs.length === 0" class="p-4"><div class="empty">暂无证书样本</div></div>
        <div v-else class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>主机</th>
                <th>状态</th>
                <th>剩余</th>
                <th>匹配</th>
                <th>过期时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="cert in latest?.certs" :key="cert.name">
                <td class="font-medium">{{ cert.name }}</td>
                <td>{{ cert.host }}</td>
                <td><span :class="statusClass(cert.status)">{{ cert.status }}</span></td>
                <td>{{ cert.days_remaining }} d</td>
                <td>{{ booleanText(cert.matched_name) }}</td>
                <td>{{ formatTime(cert.expires_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>
  </div>
</template>
