<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getJSON } from '../api'

type RunState = {
  state: string
  collector_id?: string
  job_id?: string
  protocol: string
  status?: string
  started_at?: string
  finished_at?: string
  duration_ms: number
  target_count: number
  success_count: number
  failure_count: number
  skipped_count: number
  error_count: number
  skip_reason?: string
}

type StatusSummary = {
  protocol: string
  status: string
  count: number
}

type CollectStatus = {
  latest_runs: RunState[]
  summary: StatusSummary[]
  generated_at: string
}

type Observation = {
  id: number
  site_id: number
  target: string
  protocol: string
  status: string
  observed_at: string
  duration_ms: number
  error_code?: string
  error_message?: string
  collector_id?: string
  job_id?: string
}

type SiteStatus = {
  site_id: number
  summary: {
    state: string
    status: string
    target_count: number
    status_counts: Record<string, number>
    reason_codes?: string[]
    reason_messages?: string[]
  }
  targets: Array<{ target: string; status: string; observed_at: string; reason_codes?: string[] }>
  generated_at: string
}

type TargetStatus = {
  site_id: number
  target: string
  summary: { state: string; status: string; reason_codes?: string[]; reason_messages?: string[] }
  latest_core: { state: string; protocols: Record<string, { status: string; observed_at: string; duration_ms: number; error_code?: string }> }
  latest_light: { state: string; protocols: Record<string, { status: string; observed_at: string; duration_ms: number; error_code?: string }> }
  generated_at: string
}

const loading = ref(false)
const error = ref('')
const status = ref<CollectStatus | null>(null)
const observations = ref<Observation[]>([])
const siteStatus = ref<SiteStatus | null>(null)
const targetStatus = ref<TargetStatus | null>(null)

const protocolFilter = ref('')
const statusFilter = ref('')
const siteIDFilter = ref('')
const targetFilter = ref('')
const siteID = ref('')
const target = ref('')

const successRate = computed(() => {
  const runs = status.value?.latest_runs || []
  const total = runs.reduce((sum, item) => sum + (item.target_count || 0), 0)
  const success = runs.reduce((sum, item) => sum + (item.success_count || 0), 0)
  if (!total) return '-'
  return `${Math.round((success / total) * 100)}%`
})

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({ limit: '50' })
    if (protocolFilter.value) params.set('protocol', protocolFilter.value)
    if (statusFilter.value) params.set('status', statusFilter.value)
    if (siteIDFilter.value) params.set('site_id', siteIDFilter.value)
    if (targetFilter.value) params.set('target', targetFilter.value)

    const [statusData, observationData] = await Promise.all([
      getJSON<CollectStatus>('/api/v1/nav/collect/status'),
      getJSON<Observation[]>(`/api/v1/nav/collect/observations?${params.toString()}`),
    ])
    status.value = statusData
    observations.value = observationData
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function loadSiteStatus() {
  const id = siteID.value.trim()
  if (!id) return
  loading.value = true
  error.value = ''
  try {
    siteStatus.value = await getJSON<SiteStatus>(`/api/v1/nav/collect/sites/${encodeURIComponent(id)}/status`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function loadTargetStatus() {
  const id = siteID.value.trim()
  const value = target.value.trim()
  if (!id || !value) return
  loading.value = true
  error.value = ''
  try {
    targetStatus.value = await getJSON<TargetStatus>(`/api/v1/nav/collect/sites/${encodeURIComponent(id)}/targets/${encodeURIComponent(value)}/status`)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function formatTime(value?: string) {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

function formatDuration(value?: number) {
  if (!value) return '-'
  if (value < 1000) return `${value} ms`
  return `${(value / 1000).toFixed(1)} s`
}

function statusClass(value?: string) {
  if (value === 'success' || value === 'complete' || value === 'healthy') return 'text-emerald-300'
  if (value === 'warning' || value === 'skipped') return 'text-[var(--warning)]'
  if (value === 'failure' || value === 'failed' || value === 'down' || value === 'degraded') return 'text-[var(--danger)]'
  return 'text-[var(--text-muted)]'
}

function protocolEntries(group?: TargetStatus['latest_core']) {
  return Object.entries(group?.protocols || {})
}

onMounted(refresh)
</script>

<template>
  <main class="space-y-5">
    <header class="flex flex-col gap-3 border-b border-[var(--line)] pb-4 md:flex-row md:items-end md:justify-between">
      <div>
        <div class="text-xs uppercase tracking-[0.25em] text-[var(--accent)]">Nav V2</div>
        <h1 class="mt-2 text-2xl font-semibold">导航采集观测</h1>
        <p class="mt-1 text-sm text-[var(--text-muted)]">查看 nav 采集器协议运行状态、站点健康摘要和最近观测。</p>
      </div>
      <button class="border border-[var(--accent)] px-4 py-2 text-sm text-[var(--accent)] disabled:opacity-50" :disabled="loading" @click="refresh">
        {{ loading ? '刷新中' : '刷新' }}
      </button>
    </header>

    <div v-if="error" class="border border-[var(--danger)] bg-[var(--danger)]/10 px-4 py-3 text-sm text-[var(--danger)]">
      {{ error }}
    </div>

    <section class="grid gap-4 md:grid-cols-4">
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">协议数</div>
        <div class="mt-2 text-lg font-semibold">{{ status?.latest_runs?.length || 0 }}</div>
      </div>
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">最近目标</div>
        <div class="mt-2 text-lg font-semibold">{{ status?.latest_runs?.reduce((sum, item) => sum + (item.target_count || 0), 0) || 0 }}</div>
      </div>
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">成功率</div>
        <div class="mt-2 text-lg font-semibold">{{ successRate }}</div>
      </div>
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">生成时间</div>
        <div class="mt-2 text-sm font-semibold">{{ formatTime(status?.generated_at) }}</div>
      </div>
    </section>

    <section class="grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <h2 class="text-lg font-semibold">协议最新运行</h2>
        <div class="mt-3 overflow-x-auto">
          <table class="w-full min-w-[760px] text-left text-sm">
            <thead class="text-xs uppercase text-[var(--text-muted)]">
              <tr>
                <th class="py-2">协议</th>
                <th class="py-2">状态</th>
                <th class="py-2">目标</th>
                <th class="py-2">成功</th>
                <th class="py-2">失败</th>
                <th class="py-2">耗时</th>
                <th class="py-2">Collector</th>
                <th class="py-2">开始</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in status?.latest_runs || []" :key="item.protocol" class="border-t border-[var(--line)]">
                <td class="py-2 font-mono text-xs">{{ item.protocol }}</td>
                <td class="py-2" :class="statusClass(item.status || item.state)">{{ item.status || item.state }}</td>
                <td class="py-2">{{ item.target_count }}</td>
                <td class="py-2">{{ item.success_count }}</td>
                <td class="py-2">{{ item.failure_count }}</td>
                <td class="py-2">{{ formatDuration(item.duration_ms) }}</td>
                <td class="max-w-[160px] truncate py-2">{{ item.collector_id || '-' }}</td>
                <td class="py-2">{{ formatTime(item.started_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <h2 class="text-lg font-semibold">近 7 天观测摘要</h2>
        <div class="mt-3 grid gap-2 sm:grid-cols-2">
          <div v-for="item in status?.summary || []" :key="`${item.protocol}-${item.status}`" class="border border-[var(--line)] bg-[var(--panel-strong)] px-3 py-2">
            <div class="font-mono text-xs">{{ item.protocol }}</div>
            <div class="mt-1 flex items-center justify-between text-sm">
              <span :class="statusClass(item.status)">{{ item.status }}</span>
              <span class="font-semibold">{{ item.count }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="border border-[var(--line)] bg-[var(--panel)] p-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <h2 class="text-lg font-semibold">最近观测</h2>
        <div class="flex flex-wrap gap-2">
          <input v-model="siteIDFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="site_id" />
          <input v-model="targetFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="target" />
          <input v-model="protocolFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="protocol" />
          <select v-model="statusFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm">
            <option value="">全部状态</option>
            <option value="success">success</option>
            <option value="failure">failure</option>
          </select>
          <button class="border border-[var(--line-strong)] px-3 py-2 text-sm" @click="refresh">筛选</button>
        </div>
      </div>
      <div class="mt-3 overflow-x-auto">
        <table class="w-full min-w-[960px] text-left text-sm">
          <thead class="text-xs uppercase text-[var(--text-muted)]">
            <tr>
              <th class="py-2">站点</th>
              <th class="py-2">Target</th>
              <th class="py-2">协议</th>
              <th class="py-2">状态</th>
              <th class="py-2">耗时</th>
              <th class="py-2">错误</th>
              <th class="py-2">Job</th>
              <th class="py-2">时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in observations" :key="item.id" class="border-t border-[var(--line)]">
              <td class="py-2">{{ item.site_id }}</td>
              <td class="max-w-[260px] truncate py-2">{{ item.target }}</td>
              <td class="py-2 font-mono text-xs">{{ item.protocol }}</td>
              <td class="py-2" :class="statusClass(item.status)">{{ item.status }}</td>
              <td class="py-2">{{ formatDuration(item.duration_ms) }}</td>
              <td class="max-w-[260px] truncate py-2">{{ item.error_code || item.error_message || '-' }}</td>
              <td class="max-w-[180px] truncate py-2 font-mono text-xs">{{ item.job_id || '-' }}</td>
              <td class="py-2">{{ formatTime(item.observed_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="border border-[var(--line)] bg-[var(--panel)] p-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <h2 class="text-lg font-semibold">站点/目标排查</h2>
        <div class="flex flex-wrap gap-2">
          <input v-model="siteID" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="site id" />
          <input v-model="target" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="target" />
          <button class="border border-[var(--line-strong)] px-3 py-2 text-sm" @click="loadSiteStatus">查站点</button>
          <button class="border border-[var(--line-strong)] px-3 py-2 text-sm" @click="loadTargetStatus">查目标</button>
        </div>
      </div>

      <div class="mt-4 grid gap-4 lg:grid-cols-2">
        <div v-if="siteStatus" class="border border-[var(--line)] bg-[var(--panel-strong)] p-3 text-sm">
          <div class="flex items-center justify-between">
            <div class="text-lg font-semibold">Site {{ siteStatus.site_id }}</div>
            <div :class="statusClass(siteStatus.summary.status)">{{ siteStatus.summary.status }}</div>
          </div>
          <div class="mt-2 text-[var(--text-muted)]">targets {{ siteStatus.summary.target_count }} / state {{ siteStatus.summary.state }}</div>
          <div class="mt-3 flex flex-wrap gap-2">
            <span v-for="targetItem in siteStatus.targets" :key="targetItem.target" class="border border-[var(--line)] px-2 py-1" :class="statusClass(targetItem.status)">
              {{ targetItem.target }} · {{ targetItem.status }}
            </span>
          </div>
        </div>

        <div v-if="targetStatus" class="border border-[var(--line)] bg-[var(--panel-strong)] p-3 text-sm">
          <div class="flex items-center justify-between">
            <div class="max-w-[70%] truncate text-lg font-semibold">{{ targetStatus.target }}</div>
            <div :class="statusClass(targetStatus.summary.status)">{{ targetStatus.summary.status }}</div>
          </div>
          <div class="mt-3 grid gap-2 sm:grid-cols-2">
            <div v-for="[protocol, item] in protocolEntries(targetStatus.latest_core)" :key="`core-${protocol}`" class="border border-[var(--line)] px-2 py-1">
              <div class="font-mono text-xs">{{ protocol }}</div>
              <div :class="statusClass(item.status)">{{ item.status }} · {{ formatDuration(item.duration_ms) }}</div>
            </div>
            <div v-for="[protocol, item] in protocolEntries(targetStatus.latest_light)" :key="`light-${protocol}`" class="border border-[var(--line)] px-2 py-1">
              <div class="font-mono text-xs">{{ protocol }}</div>
              <div :class="statusClass(item.status)">{{ item.status }} · {{ formatDuration(item.duration_ms) }}</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>
