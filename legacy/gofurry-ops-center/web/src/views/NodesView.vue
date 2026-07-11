<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import * as api from '../api'
import { formatTime, statusClass } from '../format'
import type { OpsNode } from '../types'

const rows = ref<OpsNode[]>([])
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    rows.value = await api.nodes()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="panel">
    <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
      <h2 class="font-semibold">节点清单</h2>
      <button class="icon-button" title="刷新" @click="load">
        <RefreshCw class="size-4" />
      </button>
    </div>
    <p v-if="error" class="p-4 text-sm text-[var(--ops-red)]">{{ error }}</p>
    <div v-if="loading && rows.length === 0" class="p-4"><div class="empty">加载中</div></div>
    <div v-else-if="rows.length === 0" class="p-4"><div class="empty">暂无节点</div></div>
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>节点</th>
            <th>区域</th>
            <th>角色</th>
            <th>状态</th>
            <th>Agent</th>
            <th>最后心跳</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="node in rows" :key="node.node_id">
            <td>
              <RouterLink class="font-semibold text-[var(--ops-teal)]" :to="`/nodes/${node.node_id}`">
                {{ node.display_name || node.node_id }}
              </RouterLink>
              <p class="text-xs text-[var(--ops-muted)]">{{ node.node_id }}</p>
            </td>
            <td>{{ node.region }}</td>
            <td>{{ node.role || '-' }}</td>
            <td><span :class="statusClass(node.status)">{{ node.status }}</span></td>
            <td>{{ node.agent_version || '-' }}</td>
            <td>{{ formatTime(node.last_seen_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
