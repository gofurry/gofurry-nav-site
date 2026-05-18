<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import * as api from '../api'
import { alertClass, formatTime } from '../format'
import type { AlertState } from '../types'

const rows = ref<AlertState[]>([])
const activeOnly = ref(true)
const error = ref('')

async function load() {
  error.value = ''
  try {
    rows.value = await api.alerts(activeOnly.value)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  }
}

onMounted(load)
</script>

<template>
  <div class="panel">
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--ops-border)] px-4 py-3">
      <h2 class="font-semibold">告警状态</h2>
      <div class="flex items-center gap-2">
        <label class="flex items-center gap-2 text-sm text-[var(--ops-muted)]">
          <input v-model="activeOnly" type="checkbox" @change="load" />
          仅活跃
        </label>
        <button class="icon-button" title="刷新" @click="load">
          <RefreshCw class="size-4" />
        </button>
      </div>
    </div>
    <p v-if="error" class="p-4 text-sm text-[var(--ops-red)]">{{ error }}</p>
    <div v-if="rows.length === 0" class="p-4"><div class="empty">暂无告警</div></div>
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>告警</th>
            <th>级别</th>
            <th>节点</th>
            <th>状态</th>
            <th>首次</th>
            <th>最近</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="alert in rows" :key="alert.key">
            <td>
              <p class="font-medium">{{ alert.title }}</p>
              <p class="text-sm text-[var(--ops-muted)]">{{ alert.message || alert.type }}</p>
            </td>
            <td><span :class="alertClass(alert.level)">{{ alert.level }}</span></td>
            <td>{{ alert.node_id || alert.region }}</td>
            <td>{{ alert.status }}</td>
            <td>{{ formatTime(alert.first_seen_at) }}</td>
            <td>{{ formatTime(alert.last_seen_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
