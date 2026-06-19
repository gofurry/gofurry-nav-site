<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getJSON } from '../api'

type CollectRun = {
  id: string
  task_type: string
  status: string
  total_count: number
  success_count: number
  failed_count: number
  skipped_count: number
  partial_count: number
  duration_millis: number
  error_kind: string
  error_message: string
  started_at: string
  ended_at?: string
}

type TaskResult = {
  id: number
  run_id: string
  task_type: string
  status: string
  game_id: number
  appid: number
  upstream_status_code: number
  traffic_bucket: string
  retry_count: number
  duration_millis: number
  error_kind: string
  error_message: string
  started_at: string
  ended_at?: string
}

type StatusSummary = {
  task_type: string
  status: string
  count: number
}

type CollectStatus = {
  latest_run?: CollectRun
  latest_task_runs: CollectRun[]
  summary: StatusSummary[]
  generated_at: string
}

type GameStatus = {
  game_id: number
  appid: number
  name: string
  details_updated_at?: string
  localized: Array<{ lang: string; name: string; updated_at: string }>
  prices: Array<{ region: string; available: boolean; currency: string; final_amount: number; updated_at: string }>
  media_count: number
  news_count: number
  latest_news_at?: string
  latest_player_count?: { count: number; status: string; collected_at: string }
  latest_task_results: TaskResult[]
}

const loading = ref(false)
const error = ref('')
const status = ref<CollectStatus | null>(null)
const runs = ref<CollectRun[]>([])
const taskResults = ref<TaskResult[]>([])
const gameStatus = ref<GameStatus | null>(null)

const runStatusFilter = ref('')
const taskTypeFilter = ref('')
const resultStatusFilter = ref('')
const gameID = ref('')

const latestRun = computed(() => status.value?.latest_run)

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const runParams = new URLSearchParams({ limit: '20' })
    if (runStatusFilter.value) runParams.set('status', runStatusFilter.value)
    if (taskTypeFilter.value) runParams.set('task_type', taskTypeFilter.value)

    const resultParams = new URLSearchParams({ limit: '50' })
    if (resultStatusFilter.value) resultParams.set('status', resultStatusFilter.value)
    if (taskTypeFilter.value) resultParams.set('task_type', taskTypeFilter.value)

    const [statusData, runData, resultData] = await Promise.all([
      getJSON<CollectStatus>('/api/v1/game/collect/status'),
      getJSON<CollectRun[]>(`/api/v1/game/collect/runs?${runParams.toString()}`),
      getJSON<TaskResult[]>(`/api/v1/game/collect/task-results?${resultParams.toString()}`),
    ])
    status.value = statusData
    runs.value = runData
    taskResults.value = resultData
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function loadGameStatus() {
  const id = gameID.value.trim()
  if (!id) return
  loading.value = true
  error.value = ''
  try {
    gameStatus.value = await getJSON<GameStatus>(`/api/v1/game/collect/games/${encodeURIComponent(id)}/status`)
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

function statusClass(value: string) {
  if (value === 'success') return 'text-emerald-300'
  if (value === 'partial') return 'text-[var(--warning)]'
  if (value === 'failed') return 'text-[var(--danger)]'
  return 'text-[var(--text-muted)]'
}

onMounted(refresh)
</script>

<template>
  <main class="space-y-5">
    <header class="flex flex-col gap-3 border-b border-[var(--line)] pb-4 md:flex-row md:items-end md:justify-between">
      <div>
        <div class="text-xs uppercase tracking-[0.25em] text-[var(--accent)]">Game V2</div>
        <h1 class="mt-2 text-2xl font-semibold">采集观测</h1>
        <p class="mt-1 text-sm text-[var(--text-muted)]">查看 v2 采集批次、任务结果和单游戏数据新鲜度。</p>
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
        <div class="text-xs text-[var(--text-muted)]">最新批次</div>
        <div class="mt-2 truncate text-lg font-semibold">{{ latestRun?.id || '-' }}</div>
        <div class="mt-1 text-sm" :class="statusClass(latestRun?.status || '')">{{ latestRun?.status || '-' }}</div>
      </div>
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">任务类型</div>
        <div class="mt-2 text-lg font-semibold">{{ latestRun?.task_type || '-' }}</div>
        <div class="mt-1 text-sm text-[var(--text-muted)]">{{ formatTime(latestRun?.started_at) }}</div>
      </div>
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">成功 / 失败</div>
        <div class="mt-2 text-lg font-semibold">{{ latestRun?.success_count ?? '-' }} / {{ latestRun?.failed_count ?? '-' }}</div>
        <div class="mt-1 text-sm text-[var(--text-muted)]">total {{ latestRun?.total_count ?? '-' }}</div>
      </div>
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <div class="text-xs text-[var(--text-muted)]">耗时</div>
        <div class="mt-2 text-lg font-semibold">{{ formatDuration(latestRun?.duration_millis) }}</div>
        <div class="mt-1 text-sm text-[var(--text-muted)]">{{ formatTime(status?.generated_at) }}</div>
      </div>
    </section>

    <section class="grid gap-4 lg:grid-cols-[1fr_1.2fr]">
      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <h2 class="text-lg font-semibold">任务最新批次</h2>
        <div class="mt-3 overflow-x-auto">
          <table class="w-full min-w-[640px] text-left text-sm">
            <thead class="text-xs uppercase text-[var(--text-muted)]">
              <tr>
                <th class="py-2">任务</th>
                <th class="py-2">状态</th>
                <th class="py-2">成功</th>
                <th class="py-2">失败</th>
                <th class="py-2">开始</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in status?.latest_task_runs || []" :key="item.id" class="border-t border-[var(--line)]">
                <td class="py-2">{{ item.task_type }}</td>
                <td class="py-2" :class="statusClass(item.status)">{{ item.status }}</td>
                <td class="py-2">{{ item.success_count }}</td>
                <td class="py-2">{{ item.failed_count }}</td>
                <td class="py-2">{{ formatTime(item.started_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="border border-[var(--line)] bg-[var(--panel)] p-4">
        <h2 class="text-lg font-semibold">近 7 天任务结果摘要</h2>
        <div class="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="item in status?.summary || []" :key="`${item.task_type}-${item.status}`" class="border border-[var(--line)] bg-[var(--panel-strong)] px-3 py-2">
            <div class="text-sm">{{ item.task_type }}</div>
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
        <h2 class="text-lg font-semibold">采集批次</h2>
        <div class="flex flex-wrap gap-2">
          <input v-model="taskTypeFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="task_type" />
          <select v-model="runStatusFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm">
            <option value="">全部状态</option>
            <option value="success">success</option>
            <option value="partial">partial</option>
            <option value="failed">failed</option>
          </select>
          <button class="border border-[var(--line-strong)] px-3 py-2 text-sm" @click="refresh">筛选</button>
        </div>
      </div>
      <div class="mt-3 overflow-x-auto">
        <table class="w-full min-w-[920px] text-left text-sm">
          <thead class="text-xs uppercase text-[var(--text-muted)]">
            <tr>
              <th class="py-2">Run ID</th>
              <th class="py-2">任务</th>
              <th class="py-2">状态</th>
              <th class="py-2">Total</th>
              <th class="py-2">Success</th>
              <th class="py-2">Failed</th>
              <th class="py-2">Duration</th>
              <th class="py-2">Started</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in runs" :key="item.id" class="border-t border-[var(--line)]">
              <td class="max-w-[260px] truncate py-2 font-mono text-xs">{{ item.id }}</td>
              <td class="py-2">{{ item.task_type }}</td>
              <td class="py-2" :class="statusClass(item.status)">{{ item.status }}</td>
              <td class="py-2">{{ item.total_count }}</td>
              <td class="py-2">{{ item.success_count }}</td>
              <td class="py-2">{{ item.failed_count }}</td>
              <td class="py-2">{{ formatDuration(item.duration_millis) }}</td>
              <td class="py-2">{{ formatTime(item.started_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="border border-[var(--line)] bg-[var(--panel)] p-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <h2 class="text-lg font-semibold">任务结果</h2>
        <div class="flex flex-wrap gap-2">
          <select v-model="resultStatusFilter" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm">
            <option value="">全部状态</option>
            <option value="success">success</option>
            <option value="partial">partial</option>
            <option value="failed">failed</option>
            <option value="skipped">skipped</option>
          </select>
          <button class="border border-[var(--line-strong)] px-3 py-2 text-sm" @click="refresh">筛选</button>
        </div>
      </div>
      <div class="mt-3 overflow-x-auto">
        <table class="w-full min-w-[960px] text-left text-sm">
          <thead class="text-xs uppercase text-[var(--text-muted)]">
            <tr>
              <th class="py-2">游戏</th>
              <th class="py-2">AppID</th>
              <th class="py-2">任务</th>
              <th class="py-2">状态</th>
              <th class="py-2">HTTP</th>
              <th class="py-2">Bucket</th>
              <th class="py-2">Retry</th>
              <th class="py-2">错误</th>
              <th class="py-2">Started</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in taskResults" :key="item.id" class="border-t border-[var(--line)]">
              <td class="py-2">{{ item.game_id }}</td>
              <td class="py-2">{{ item.appid }}</td>
              <td class="py-2">{{ item.task_type }}</td>
              <td class="py-2" :class="statusClass(item.status)">{{ item.status }}</td>
              <td class="py-2">{{ item.upstream_status_code || '-' }}</td>
              <td class="py-2">{{ item.traffic_bucket || '-' }}</td>
              <td class="py-2">{{ item.retry_count }}</td>
              <td class="max-w-[260px] truncate py-2">{{ item.error_kind || item.error_message || '-' }}</td>
              <td class="py-2">{{ formatTime(item.started_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="border border-[var(--line)] bg-[var(--panel)] p-4">
      <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <h2 class="text-lg font-semibold">单游戏状态</h2>
        <div class="flex gap-2">
          <input v-model="gameID" class="border border-[var(--line)] bg-[var(--bg)] px-3 py-2 text-sm" placeholder="game id" @keyup.enter="loadGameStatus" />
          <button class="border border-[var(--line-strong)] px-3 py-2 text-sm" @click="loadGameStatus">查询</button>
        </div>
      </div>
      <div v-if="gameStatus" class="mt-4 grid gap-4 md:grid-cols-2">
        <div class="border border-[var(--line)] bg-[var(--panel-strong)] p-3 text-sm">
          <div class="text-lg font-semibold">{{ gameStatus.name }}</div>
          <div class="mt-2 text-[var(--text-muted)]">game_id {{ gameStatus.game_id }} / appid {{ gameStatus.appid }}</div>
          <div class="mt-2 text-[var(--text-muted)]">详情更新 {{ formatTime(gameStatus.details_updated_at) }}</div>
          <div class="mt-2 text-[var(--text-muted)]">媒体 {{ gameStatus.media_count }} / 新闻 {{ gameStatus.news_count }} / 最新新闻 {{ formatTime(gameStatus.latest_news_at) }}</div>
        </div>
        <div class="border border-[var(--line)] bg-[var(--panel-strong)] p-3 text-sm">
          <div class="font-semibold">价格区域</div>
          <div class="mt-2 flex flex-wrap gap-2">
            <span v-for="price in gameStatus.prices" :key="price.region" class="border border-[var(--line)] px-2 py-1" :class="price.available ? 'text-emerald-300' : 'text-[var(--text-muted)]'">
              {{ price.region }} {{ price.currency || '-' }} {{ price.final_amount }}
            </span>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>
