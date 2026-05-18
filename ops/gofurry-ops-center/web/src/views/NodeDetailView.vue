<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as api from '../api'
import { formatTime, statusClass } from '../format'
import type { OpsNode } from '../types'

const props = defineProps<{ id: string }>()
const node = ref<OpsNode | null>(null)
const loading = ref(false)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    node.value = await api.node(props.id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.id, load)
</script>

<template>
  <div class="space-y-5">
    <div v-if="error" class="panel p-4 text-sm text-[var(--ops-red)]">{{ error }}</div>
    <div v-if="loading && !node" class="empty">加载中</div>
    <section v-if="node" class="panel p-5">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p class="text-sm text-[var(--ops-muted)]">{{ node.region }} · {{ node.role || 'node' }}</p>
          <h2 class="mt-1 text-2xl font-semibold">{{ node.display_name || node.node_id }}</h2>
          <p class="mt-1 text-sm text-[var(--ops-muted)]">{{ node.node_id }}</p>
        </div>
        <span :class="statusClass(node.status)">{{ node.status }}</span>
      </div>
      <dl class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="metric">
          <dt class="text-sm text-[var(--ops-muted)]">Agent</dt>
          <dd class="mt-2 font-semibold">{{ node.agent_version || '-' }}</dd>
        </div>
        <div class="metric">
          <dt class="text-sm text-[var(--ops-muted)]">最后心跳</dt>
          <dd class="mt-2 font-semibold">{{ formatTime(node.last_seen_at) }}</dd>
        </div>
        <div class="metric">
          <dt class="text-sm text-[var(--ops-muted)]">更新时间</dt>
          <dd class="mt-2 font-semibold">{{ formatTime(node.updated_at) }}</dd>
        </div>
        <div class="metric">
          <dt class="text-sm text-[var(--ops-muted)]">区域</dt>
          <dd class="mt-2 font-semibold">{{ node.region }}</dd>
        </div>
      </dl>
    </section>
  </div>
</template>
