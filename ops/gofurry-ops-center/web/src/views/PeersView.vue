<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { onMounted, ref } from 'vue'
import * as api from '../api'
import { formatTime, statusClass } from '../format'
import type { PeerSummary } from '../types'

const rows = ref<PeerSummary[]>([])
const error = ref('')

async function load() {
  error.value = ''
  try {
    rows.value = await api.peerStatus()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  }
}

onMounted(load)
</script>

<template>
  <div class="panel">
    <div class="flex items-center justify-between border-b border-[var(--ops-border)] px-4 py-3">
      <h2 class="font-semibold">Peer 摘要</h2>
      <button class="icon-button" title="刷新" @click="load">
        <RefreshCw class="size-4" />
      </button>
    </div>
    <p v-if="error" class="p-4 text-sm text-[var(--ops-red)]">{{ error }}</p>
    <div v-if="rows.length === 0" class="p-4"><div class="empty">暂无 Peer 数据</div></div>
    <div v-else class="grid gap-3 p-4 lg:grid-cols-2">
      <article v-for="peer in rows" :key="`${peer.region}-${peer.center_id}`" class="metric">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-sm text-[var(--ops-muted)]">{{ peer.region }}</p>
            <h2 class="mt-1 font-semibold">{{ peer.center_id }}</h2>
          </div>
          <span :class="statusClass(peer.status)">{{ peer.status }}</span>
        </div>
        <dl class="mt-4 grid grid-cols-2 gap-3 text-sm">
          <div>
            <dt class="text-[var(--ops-muted)]">节点</dt>
            <dd class="font-semibold">{{ peer.nodes_total }} / Down {{ peer.nodes_down }}</dd>
          </div>
          <div>
            <dt class="text-[var(--ops-muted)]">告警</dt>
            <dd class="font-semibold">C{{ peer.critical_alerts }} / W{{ peer.warning_alerts }}</dd>
          </div>
          <div>
            <dt class="text-[var(--ops-muted)]">同步</dt>
            <dd class="font-semibold">{{ peer.last_sync_status || '-' }}</dd>
          </div>
          <div>
            <dt class="text-[var(--ops-muted)]">更新时间</dt>
            <dd class="font-semibold">{{ formatTime(peer.updated_at) }}</dd>
          </div>
        </dl>
      </article>
    </div>
  </div>
</template>
