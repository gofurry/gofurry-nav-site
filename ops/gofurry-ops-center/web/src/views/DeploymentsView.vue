<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import * as api from '../api'
import { formatTime, statusClass } from '../format'
import type { DeployEvent } from '../types'

const rows = ref<DeployEvent[]>([])
const error = ref('')

async function load() {
  error.value = ''
  try {
    rows.value = await api.deployments(80)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  }
}

onMounted(load)
</script>

<template>
  <div class="panel">
    <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
      <h2 class="font-semibold">部署事件</h2>
      <button class="icon-button" title="刷新" @click="load">
        <RefreshCw class="size-4" />
      </button>
    </div>
    <p v-if="error" class="p-4 text-sm text-[var(--ops-red)]">{{ error }}</p>
    <div v-if="rows.length === 0" class="p-4"><div class="empty">暂无部署事件</div></div>
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>服务</th>
            <th>区域</th>
            <th>节点</th>
            <th>状态</th>
            <th>版本</th>
            <th>时间</th>
            <th>消息</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="event in rows" :key="event.id">
            <td class="font-medium">{{ event.service_name }}</td>
            <td>{{ event.region }}</td>
            <td>{{ event.node_id || '-' }}</td>
            <td><span :class="statusClass(event.status)">{{ event.status }}</span></td>
            <td>{{ event.version || '-' }}</td>
            <td>{{ formatTime(event.created_at) }}</td>
            <td class="max-w-md text-sm text-[var(--ops-muted)]">{{ event.message || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
