<script setup lang="ts">
import { AlertTriangle, Activity, GitBranch, Server } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import * as api from '../api'
import { alertClass, formatTime, statusClass } from '../format'
import type { Overview } from '../types'

const data = ref<Overview | null>(null)
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    data.value = await api.overview()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">
    <div v-if="error" class="panel border-[var(--ops-red)] p-4 text-sm text-[var(--ops-red)]">{{ error }}</div>
    <div v-if="loading && !data" class="empty">加载中</div>

    <template v-if="data">
      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">节点</p>
            <Server class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ data.nodes_total }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">Down {{ data.nodes_down }}</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">服务</p>
            <Activity class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ data.services.length }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">最近心跳 {{ formatTime(data.last_heartbeat_at) }}</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">告警</p>
            <AlertTriangle class="size-5 text-[var(--ops-amber)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ data.critical_alerts + data.warning_alerts }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">Critical {{ data.critical_alerts }} / Warning {{ data.warning_alerts }}</p>
        </div>
        <div class="metric">
          <div class="flex items-center justify-between">
            <p class="text-sm font-medium text-[var(--ops-muted)]">同步</p>
            <GitBranch class="size-5 text-[var(--ops-teal)]" />
          </div>
          <p class="mt-3 text-3xl font-semibold">{{ data.last_sync?.status || '-' }}</p>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">{{ data.last_sync?.sync_name || data.region }}</p>
        </div>
      </section>

      <section class="grid gap-5 xl:grid-cols-[1.1fr_0.9fr]">
        <div class="panel">
          <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
            <h2 class="font-semibold">服务状态</h2>
            <span :class="statusClass(data.status)">{{ data.status }}</span>
          </div>
          <div class="table-wrap">
            <table class="data-table">
              <thead>
                <tr>
                  <th>服务</th>
                  <th>节点</th>
                  <th>状态</th>
                  <th>延迟</th>
                  <th>更新时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="service in data.services.slice(0, 8)" :key="service.key">
                  <td>
                    <p class="font-medium">{{ service.name }}</p>
                    <p class="text-xs text-[var(--ops-muted)]">{{ service.service_type }}</p>
                  </td>
                  <td>{{ service.node_id }}</td>
                  <td><span :class="statusClass(service.status)">{{ service.status }}</span></td>
                  <td>{{ service.latency_ms || 0 }} ms</td>
                  <td>{{ formatTime(service.updated_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="data.services.length === 0" class="p-4">
            <div class="empty">暂无服务样本</div>
          </div>
        </div>

        <div class="panel">
          <div class="border-b border-[var(--ops-border)] px-4 py-3">
            <h2 class="font-semibold">活跃告警</h2>
          </div>
          <div class="divide-y divide-[#edf0ed]">
            <div v-for="alert in data.alerts.slice(0, 8)" :key="alert.key" class="space-y-2 p-4">
              <div class="flex items-center justify-between gap-3">
                <p class="font-medium">{{ alert.title }}</p>
                <span :class="alertClass(alert.level)">{{ alert.level }}</span>
              </div>
              <p class="text-sm text-[var(--ops-muted)]">{{ alert.message || alert.type }}</p>
              <p class="text-xs text-[var(--ops-muted)]">{{ alert.node_id || alert.region }} · {{ formatTime(alert.last_seen_at) }}</p>
            </div>
          </div>
          <div v-if="data.alerts.length === 0" class="p-4">
            <div class="empty">暂无活跃告警</div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
