<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import * as api from '../api'
import { booleanText, formatTime, statusClass } from '../format'
import type { SyncRun } from '../types'

const rows = ref<SyncRun[]>([])
const error = ref('')

async function load() {
  error.value = ''
  try {
    rows.value = await api.syncRuns(80)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  }
}

onMounted(load)
</script>

<template>
  <div class="panel">
    <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
      <h2 class="font-semibold">同步事件</h2>
      <button class="icon-button" title="刷新" @click="load">
        <RefreshCw class="size-4" />
      </button>
    </div>
    <p v-if="error" class="p-4 text-sm text-[var(--ops-red)]">{{ error }}</p>
    <div v-if="rows.length === 0" class="p-4"><div class="empty">暂无同步事件</div></div>
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>同步</th>
            <th>区域</th>
            <th>状态</th>
            <th>数量</th>
            <th>Checksum</th>
            <th>完成时间</th>
            <th>错误</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="run in rows" :key="run.id">
            <td>
              <p class="font-medium">{{ run.sync_name }}</p>
              <p class="text-xs text-[var(--ops-muted)]">{{ run.version || '-' }}</p>
            </td>
            <td>{{ run.region }}</td>
            <td><span :class="statusClass(run.status)">{{ run.status }}</span></td>
            <td>{{ run.items_total }}</td>
            <td>{{ booleanText(run.checksum_ok) }}</td>
            <td>{{ formatTime(run.finished_at || run.created_at) }}</td>
            <td class="max-w-md text-sm text-[var(--ops-muted)]">{{ run.error_message || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
