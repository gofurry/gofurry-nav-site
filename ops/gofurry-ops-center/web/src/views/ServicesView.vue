<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import * as api from '../api'
import { formatTime, statusClass } from '../format'
import type { ServiceStatus } from '../types'

const rows = ref<ServiceStatus[]>([])
const error = ref('')

async function load() {
  error.value = ''
  try {
    rows.value = await api.services()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  }
}

onMounted(load)
</script>

<template>
  <div class="panel">
    <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
      <h2 class="font-semibold">服务状态</h2>
      <button class="icon-button" title="刷新" @click="load">
        <RefreshCw class="size-4" />
      </button>
    </div>
    <p v-if="error" class="p-4 text-sm text-[var(--ops-red)]">{{ error }}</p>
    <div v-if="rows.length === 0" class="p-4"><div class="empty">暂无服务样本</div></div>
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>服务</th>
            <th>类型</th>
            <th>节点</th>
            <th>状态</th>
            <th>失败次数</th>
            <th>最后成功</th>
            <th>消息</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="service in rows" :key="service.key">
            <td class="font-medium">{{ service.name }}</td>
            <td>{{ service.service_type }}</td>
            <td>{{ service.node_id }}</td>
            <td><span :class="statusClass(service.status)">{{ service.status }}</span></td>
            <td>{{ service.failure_count }}</td>
            <td>{{ formatTime(service.last_ok_at) }}</td>
            <td class="max-w-md text-sm text-[var(--ops-muted)]">{{ service.message || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
