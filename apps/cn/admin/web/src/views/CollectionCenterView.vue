<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import type { EChartsOption } from 'echarts'
import { getJSON, sendJSON } from '../api'
import EChart from '../components/EChart.vue'

type Overview = { running_count: number; queued_count: number; failed_24h: number; missed_24h: number }
type Instance = { domain: string; instance_id: string; collector_id: string; version: string; health: string; heartbeat_age_seconds: number; started_at: string }
type Schedule = { domain: string; id: number; job_key: string; name: string; enabled: boolean; schedule_kind: string; cron_expression?: string; interval_seconds?: number; anchor_at?: string; timezone: string; misfire_policy: string; misfire_grace_seconds: number; next_scheduled_for?: string; last_materialized_for?: string; last_status: string; last_success_coverage: number }
type Progress = { expected: number; attempted: number; success: number; partial: number; failed: number; skipped: number }
type Job = { domain: string; id: number; job_key: string; trigger: string; scope_type: string; scope_id?: number; target?: string; tasks: string[]; priority: number; scheduled_for?: string; status: string; requested_by: string; claimed_by?: string; created_at: string; run_id?: string; progress?: Progress }
type Run = { domain: string; id: string; job_id: number; job_key: string; trigger: string; scope_type: string; scope_id?: number; target?: string; status: string; expected_count: number; attempted_count: number; success_count: number; partial_count: number; failure_count: number; skipped_count: number; schedule_delay_ms: number; duration_ms: number; collector_instance_id: string; started_at: string }
type Result = { domain: string; id: number; run_id: string; task: string; entity_id: number; appid?: number; target?: string; status: string; observation_id?: number; duration_ms: number; error_kind: string; error_message: string; started_at: string }
type ChartPoint = { domain: string; job_id: number; job_key: string; job_status: string; run_status: string; expected: number; attempted: number; success: number; partial: number; failed: number; skipped: number; coverage: number; schedule_delay_ms: number; duration_ms: number; created_at: string }

const tabs = ['Overview', 'Schedules', 'Running / Queue', 'History', 'Manual'] as const
const activeTab = ref<(typeof tabs)[number]>('Overview')
const loading = ref(false)
const error = ref('')
const notice = ref('')
const overview = ref<Overview>({ running_count: 0, queued_count: 0, failed_24h: 0, missed_24h: 0 })
const instances = ref<Instance[]>([])
const schedules = ref<Schedule[]>([])
const jobs = ref<Job[]>([])
const runs = ref<Run[]>([])
const chartPoints = ref<ChartPoint[]>([])
const domainFilter = ref('')
const chartWindow = ref('24h')
const chartJobKey = ref('')
const historyStatus = ref('')
const historyTrigger = ref('')
const historySince = ref('')
const historyUntil = ref('')
const editing = ref<Schedule | null>(null)
const resultRun = ref<Run | null>(null)
const resultRows = ref<Result[]>([])
const resultSearch = reactive({ game_id: '', appid: '', site_id: '', target: '', protocol: '' })

const manual = reactive({ domain: 'game', scope_type: 'all', scope_id: '', target: '', tasks: ['details', 'news'] as string[] })
const gameTaskOptions = ['details', 'news', 'players']
const navTaskOptions = ['ping', 'http', 'dns', 'rdap', 'robots', 'security_txt', 'llms_txt', 'page_assets', 'port_check', 'waf_canary']
const manualTaskOptions = computed(() => manual.domain === 'game' ? gameTaskOptions : navTaskOptions)

const visibleJobs = computed(() => jobs.value.filter((item) => !domainFilter.value || item.domain === domainFilter.value))
const activeJobs = computed(() => visibleJobs.value.filter((item) => item.status === 'running' || item.status === 'queued'))
const visibleRuns = computed(() => runs.value.filter((item) => !domainFilter.value || item.domain === domainFilter.value))

async function refresh(realtime = false) {
  if (!realtime) loading.value = true
  error.value = ''
  try {
    const [overviewData, instanceData, scheduleData, jobData, runData] = await Promise.all([
      getJSON<Overview>('/api/v1/collection/overview'),
      getJSON<Instance[]>('/api/v1/collection/instances'),
      getJSON<Schedule[]>('/api/v1/collection/schedules'),
      getJSON<Job[]>('/api/v1/collection/jobs?limit=200'),
      getJSON<Run[]>(historyRunsURL()),
    ])
    overview.value = overviewData
    instances.value = instanceData
    schedules.value = scheduleData
    jobs.value = jobData
    runs.value = runData
    await loadCharts()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '加载失败'
  } finally {
    loading.value = false
  }
}

function historyRunsURL() {
  const params = new URLSearchParams({ limit: '200' })
  if (domainFilter.value) params.set('domain', domainFilter.value)
  if (chartJobKey.value.trim()) params.set('job_key', chartJobKey.value.trim())
  if (historyStatus.value) params.set('status', historyStatus.value)
  if (historyTrigger.value.trim()) params.set('trigger', historyTrigger.value.trim())
  if (historySince.value) params.set('since', new Date(historySince.value).toISOString())
  if (historyUntil.value) params.set('until', new Date(historyUntil.value).toISOString())
  return `/api/v1/collection/runs?${params}`
}

async function loadHistory() {
  loading.value = true
  error.value = ''
  try {
    runs.value = await getJSON<Run[]>(historyRunsURL())
    await loadCharts()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '历史加载失败'
  } finally {
    loading.value = false
  }
}

async function loadCharts() {
  const params = new URLSearchParams({ window: chartWindow.value })
  if (domainFilter.value) params.set('domain', domainFilter.value)
  if (chartJobKey.value.trim()) params.set('job_key', chartJobKey.value.trim())
  chartPoints.value = await getJSON<ChartPoint[]>(`/api/v1/collection/charts/outcomes?${params}`)
}

async function runNow(schedule: Schedule) {
  await mutate(() => sendJSON(`/api/v1/collection/schedules/${schedule.domain}/${schedule.id}/run`, 'POST'))
}

async function toggleSchedule(schedule: Schedule) {
  await saveSchedule({ ...schedule, enabled: !schedule.enabled })
}

async function saveSchedule(schedule = editing.value) {
  if (!schedule) return
  await mutate(() => sendJSON(`/api/v1/collection/schedules/${schedule.domain}/${schedule.id}`, 'PUT', {
    enabled: schedule.enabled,
    schedule_kind: schedule.schedule_kind,
    cron_expression: schedule.schedule_kind === 'cron' ? schedule.cron_expression : null,
    interval_seconds: schedule.schedule_kind === 'interval' ? Number(schedule.interval_seconds) : null,
    anchor_at: schedule.schedule_kind === 'interval' ? schedule.anchor_at : null,
    timezone: schedule.timezone,
    misfire_policy: schedule.misfire_policy,
    misfire_grace_seconds: Number(schedule.misfire_grace_seconds),
  }))
  editing.value = null
}

async function cancelJob(job: Job) {
  await mutate(() => sendJSON(`/api/v1/collection/jobs/${job.domain}/${job.id}/cancel`, 'POST'))
}

function isPointInTime(run: Run) {
  return run.job_key === 'game.players' || ['nav.ping', 'nav.http', 'nav.dns', 'nav.port_check'].includes(run.job_key)
}

async function retryRun(run: Run) {
  if (isPointInTime(run)) {
    const task = run.job_key === 'game.players' ? 'players' : run.job_key.replace('nav.', '')
    await mutate(() => sendJSON('/api/v1/collection/jobs', 'POST', {
      domain: run.domain,
      scope_type: run.scope_type,
      scope_id: run.scope_type === 'all' ? null : run.scope_id,
      target: run.scope_type === 'target' ? run.target : null,
      tasks: [task],
    }))
    return
  }
  await mutate(() => sendJSON(`/api/v1/collection/jobs/${run.domain}/${run.job_id}/retry`, 'POST'))
}

async function loadResults(run = resultRun.value) {
  if (!run) return
  resultRun.value = run
  resultRows.value = []
  const params = new URLSearchParams({ limit: '500' })
  if (run.domain === 'game') {
    if (resultSearch.game_id) params.set('game_id', resultSearch.game_id)
    if (resultSearch.appid) params.set('appid', resultSearch.appid)
  } else {
    if (resultSearch.site_id) params.set('site_id', resultSearch.site_id)
    if (resultSearch.target.trim()) params.set('target', resultSearch.target.trim())
    if (resultSearch.protocol.trim()) params.set('protocol', resultSearch.protocol.trim())
  }
  loading.value = true
  error.value = ''
  try {
    resultRows.value = await getJSON<Result[]>(`/api/v1/collection/runs/${run.domain}/${run.id}/results?${params}`)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '结果加载失败'
  } finally {
    loading.value = false
  }
}

async function createManual() {
  const scopeID = manual.scope_type === 'all' ? null : Number(manual.scope_id)
  await mutate(async () => {
    const created = await sendJSON<Job[]>('/api/v1/collection/jobs', 'POST', {
      domain: manual.domain,
      scope_type: manual.scope_type,
      scope_id: scopeID,
      target: manual.scope_type === 'target' ? manual.target.trim() : null,
      tasks: manual.tasks,
    })
    notice.value = `已创建 Job：${created.map((item) => `${item.domain}#${item.id}`).join(', ')}`
  })
}

async function mutate(action: () => Promise<unknown>) {
  loading.value = true
  error.value = ''
  notice.value = ''
  try {
    await action()
    await refresh(true)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : '操作失败'
  } finally {
    loading.value = false
  }
}

function setManualDomain(domain: string) {
  manual.domain = domain
  manual.scope_type = 'all'
  manual.scope_id = ''
  manual.target = ''
  manual.tasks = domain === 'game' ? ['details', 'news'] : ['ping']
}

function onManualDomainChange(event: Event) {
  setManualDomain((event.target as HTMLSelectElement).value)
}

function toggleManualTask(task: string) {
  manual.tasks = manual.tasks.includes(task) ? manual.tasks.filter((item) => item !== task) : [...manual.tasks, task]
}

function statusClass(status: string) {
  if (status === 'success' || status === 'online') return 'text-emerald-300'
  if (status === 'partial' || status === 'queued' || status === 'degraded' || status === 'missed') return 'text-[var(--warning)]'
  if (status === 'failed' || status === 'offline' || status === 'canceled') return 'text-[var(--danger)]'
  return 'text-[var(--text-muted)]'
}

function formatTime(value?: string) { return value ? new Date(value).toLocaleString() : '-' }
function formatDuration(value: number) { return value < 1000 ? `${value} ms` : `${(value / 1000).toFixed(1)} s` }
function coverage(success: number, expected: number) { return expected ? `${Math.round(success / expected * 100)}%` : '0%' }
function progressFor(job: Job) { return job.progress || { expected: 0, attempted: 0, success: 0, partial: 0, failed: 0, skipped: 0 } }

function chartBase(): EChartsOption {
  return {
    backgroundColor: 'transparent',
    textStyle: { color: '#aeb9c5' },
    tooltip: { trigger: 'axis' },
    grid: { left: 48, right: 18, top: 24, bottom: 48 },
    xAxis: {
      type: 'category',
      data: chartPoints.value.map((item) => new Date(item.created_at).toLocaleString()),
      axisLabel: { rotate: 25 },
    },
    yAxis: { type: 'value' },
  }
}

const outcomeOption = computed<EChartsOption>(() => ({
  ...chartBase(),
  legend: { data: ['success', 'partial', 'failed', 'skipped', 'missed'], textStyle: { color: '#aeb9c5' } },
  series: [
    ...(['success', 'partial', 'failed', 'skipped'] as const).map((key) => ({
      name: key,
      type: 'bar' as const,
      stack: 'outcome',
      data: chartPoints.value.map((item) => item[key]),
    })),
    { name: 'missed', type: 'bar', stack: 'outcome', data: chartPoints.value.map((item) => item.job_status === 'missed' ? 1 : 0) },
  ],
}))
const coverageOption = computed<EChartsOption>(() => ({ ...chartBase(), series: [{ name: 'success coverage', type: 'line', smooth: true, data: chartPoints.value.map((item) => Number((item.coverage * 100).toFixed(2))) }], yAxis: { type: 'value', min: 0, max: 100, axisLabel: { formatter: '{value}%' } } }))
const timingOption = computed<EChartsOption>(() => ({ ...chartBase(), legend: { data: ['schedule delay', 'duration'], textStyle: { color: '#aeb9c5' } }, series: [{ name: 'schedule delay', type: 'line', data: chartPoints.value.map((item) => item.schedule_delay_ms) }, { name: 'duration', type: 'line', data: chartPoints.value.map((item) => item.duration_ms) }] }))

let timer = 0
onMounted(async () => { await refresh(); timer = window.setInterval(() => refresh(true), 3000) })
onBeforeUnmount(() => window.clearInterval(timer))
</script>

<template>
  <main class="space-y-5">
    <header class="page-header flex-col md:flex-row">
      <div><h1 class="text-2xl font-semibold">Collection Center</h1><p class="mt-1 text-sm text-[var(--text-muted)]">统一管理 Game / Nav 的调度、队列、运行历史与采集下发。</p></div>
      <div class="flex gap-2"><select v-model="domainFilter" class="ui-control px-3 py-2 text-sm" @change="loadCharts"><option value="">全部域</option><option value="game">Game</option><option value="nav">Nav</option></select><button class="ui-button ui-button--primary px-4 py-2 text-sm" @click="refresh()">{{ loading ? '刷新中' : '刷新' }}</button></div>
    </header>
    <div v-if="error" class="status-alert px-4 py-3 text-sm text-[var(--danger)]">{{ error }}</div>
    <div v-if="notice" class="status-alert px-4 py-3 text-sm text-emerald-300">{{ notice }}</div>
    <div class="flex flex-wrap gap-2"><button v-for="tab in tabs" :key="tab" class="ui-button px-3 py-2 text-sm" :class="{ 'ui-button--primary': activeTab === tab }" @click="activeTab = tab">{{ tab }}</button></div>

    <template v-if="activeTab === 'Overview'">
      <section class="status-rail grid grid-cols-2 md:grid-cols-4"><div v-for="[label, value] in [['Running', overview.running_count], ['Queued', overview.queued_count], ['Failed 24h', overview.failed_24h], ['Missed 24h', overview.missed_24h]]" :key="String(label)" class="status-rail__item"><div class="text-xs text-[var(--text-muted)]">{{ label }}</div><div class="mt-2 text-xl font-semibold">{{ value }}</div></div></section>
      <section class="workspace-section"><h2 class="text-lg font-semibold">Collector Instances</h2><div class="mt-3 overflow-x-auto"><table class="w-full min-w-[800px] text-left text-sm"><thead class="text-xs uppercase text-[var(--text-muted)]"><tr><th>Domain</th><th>Collector</th><th>Instance</th><th>Version</th><th>Health</th><th>Heartbeat</th><th>Started</th></tr></thead><tbody><tr v-for="item in instances" :key="`${item.domain}-${item.instance_id}`" class="border-t border-[var(--line)]"><td class="py-2">{{ item.domain }}</td><td>{{ item.collector_id }}</td><td class="font-mono text-xs">{{ item.instance_id }}</td><td>{{ item.version }}</td><td :class="statusClass(item.health)">{{ item.health }}</td><td>{{ item.heartbeat_age_seconds }}s</td><td>{{ formatTime(item.started_at) }}</td></tr></tbody></table></div></section>
      <section class="grid gap-4 xl:grid-cols-3"><div class="workspace-section"><h2 class="font-semibold">Outcome</h2><EChart :option="outcomeOption" /></div><div class="workspace-section"><h2 class="font-semibold">Coverage</h2><EChart :option="coverageOption" /></div><div class="workspace-section"><h2 class="font-semibold">Timing</h2><EChart :option="timingOption" /></div></section>
    </template>

    <section v-else-if="activeTab === 'Schedules'" class="workspace-section"><h2 class="text-lg font-semibold">Schedules</h2><div class="mt-3 overflow-x-auto"><table class="w-full min-w-[1050px] text-left text-sm"><thead class="text-xs uppercase text-[var(--text-muted)]"><tr><th>Domain</th><th>Job</th><th>Enabled</th><th>Schedule</th><th>Policy</th><th>Next</th><th>Last</th><th>Coverage</th><th>Actions</th></tr></thead><tbody><tr v-for="item in schedules" :key="`${item.domain}-${item.id}`" class="border-t border-[var(--line)]"><td class="py-2">{{ item.domain }}</td><td class="font-mono text-xs">{{ item.job_key }}</td><td :class="item.enabled ? 'text-emerald-300' : 'text-[var(--text-muted)]'">{{ item.enabled }}</td><td>{{ item.schedule_kind === 'cron' ? item.cron_expression : `${item.interval_seconds}s @ ${formatTime(item.anchor_at)}` }}</td><td>{{ item.misfire_policy }}</td><td>{{ formatTime(item.next_scheduled_for) }}</td><td :class="statusClass(item.last_status)">{{ item.last_status || '-' }}</td><td>{{ Math.round(item.last_success_coverage * 100) }}%</td><td class="space-x-2"><button class="ui-button px-2 py-1" @click="editing = { ...item }">Edit</button><button class="ui-button px-2 py-1" @click="toggleSchedule(item)">{{ item.enabled ? 'Disable' : 'Enable' }}</button><button class="ui-button px-2 py-1" @click="runNow(item)">Run Now</button></td></tr></tbody></table></div></section>

    <section v-else-if="activeTab === 'Running / Queue'" class="workspace-section"><h2 class="text-lg font-semibold">Running / Queue</h2><div class="mt-3 overflow-x-auto"><table class="w-full min-w-[1050px] text-left text-sm"><thead class="text-xs uppercase text-[var(--text-muted)]"><tr><th>Priority</th><th>Domain</th><th>Job</th><th>Trigger</th><th>Scope</th><th>Status</th><th>Progress</th><th>Scheduled / Created</th><th>Action</th></tr></thead><tbody><tr v-for="item in activeJobs" :key="`${item.domain}-${item.id}`" class="border-t border-[var(--line)]"><td class="py-2">{{ item.priority }}</td><td>{{ item.domain }}</td><td class="font-mono text-xs">{{ item.job_key }} #{{ item.id }}</td><td>{{ item.trigger }}</td><td>{{ item.scope_type }} {{ item.scope_id || '' }} {{ item.target || '' }}</td><td :class="statusClass(item.status)">{{ item.status }}</td><td class="min-w-[220px]"><div>{{ progressFor(item).attempted }}/{{ progressFor(item).expected }} · ✓{{ progressFor(item).success }} ◐{{ progressFor(item).partial }} ✕{{ progressFor(item).failed }}</div><div class="mt-1 h-1.5 bg-[var(--control)]"><div class="h-full bg-emerald-400" :style="{ width: `${progressFor(item).expected ? Math.min(100, progressFor(item).attempted / progressFor(item).expected * 100) : 0}%` }" /></div></td><td>{{ formatTime(item.scheduled_for || item.created_at) }}</td><td><button class="ui-button px-2 py-1" @click="cancelJob(item)">Cancel</button></td></tr></tbody></table></div></section>

    <section v-else-if="activeTab === 'History'" class="workspace-section"><div class="flex flex-wrap items-center justify-between gap-2"><h2 class="text-lg font-semibold">Run History</h2><div class="flex flex-wrap gap-2"><select v-model="chartWindow" class="ui-control px-2 py-1" @change="loadCharts"><option value="24h">24h</option><option value="7d">7d</option><option value="30d">30d</option></select><input v-model="chartJobKey" class="ui-control px-2 py-1" placeholder="job_key" /><select v-model="historyStatus" class="ui-control px-2 py-1"><option value="">all status</option><option v-for="value in ['success','partial','failed','canceled','running']" :key="value" :value="value">{{ value }}</option></select><input v-model="historyTrigger" class="ui-control px-2 py-1" placeholder="trigger" /><input v-model="historySince" type="datetime-local" class="ui-control px-2 py-1" /><input v-model="historyUntil" type="datetime-local" class="ui-control px-2 py-1" /><button class="ui-button ui-button--primary px-3 py-1" @click="loadHistory">Apply</button></div></div><div class="mt-3 overflow-x-auto"><table class="w-full min-w-[1150px] text-left text-sm"><thead class="text-xs uppercase text-[var(--text-muted)]"><tr><th>Time</th><th>Domain / Job</th><th>Trigger</th><th>Status</th><th>Coverage</th><th>Counters</th><th>Duration</th><th>Delay</th><th>Collector</th><th>Action</th></tr></thead><tbody><tr v-for="item in visibleRuns" :key="item.id" class="border-t border-[var(--line)]"><td class="py-2">{{ formatTime(item.started_at) }}</td><td>{{ item.domain }} · <span class="font-mono text-xs">{{ item.job_key }}</span></td><td>{{ item.trigger }}</td><td :class="statusClass(item.status)">{{ item.status }}</td><td>{{ coverage(item.success_count, item.expected_count) }}</td><td>E{{ item.expected_count }} / A{{ item.attempted_count }} / S{{ item.success_count }} / P{{ item.partial_count }} / F{{ item.failure_count }}</td><td>{{ formatDuration(item.duration_ms) }}</td><td>{{ formatDuration(item.schedule_delay_ms) }}</td><td class="font-mono text-xs">{{ item.collector_instance_id }}</td><td class="space-x-2"><button class="ui-button px-2 py-1" @click="loadResults(item)">Results</button><button v-if="!['running'].includes(item.status)" class="ui-button px-2 py-1" @click="retryRun(item)">{{ isPointInTime(item) ? 'Run Again' : 'Retry' }}</button></td></tr></tbody></table></div></section>

    <section v-else class="workspace-section"><h2 class="text-lg font-semibold">Manual Collection</h2><div class="mt-4 grid gap-4 md:grid-cols-2"><div class="detail-block space-y-3"><label class="block text-sm">Domain<select :value="manual.domain" class="ui-control mt-1 w-full px-3 py-2" @change="onManualDomainChange"><option value="game">Game</option><option value="nav">Nav</option></select></label><label class="block text-sm">Scope<select v-model="manual.scope_type" class="ui-control mt-1 w-full px-3 py-2"><option value="all">all</option><option v-if="manual.domain === 'game'" value="game">game</option><option v-if="manual.domain === 'nav'" value="site">site</option><option v-if="manual.domain === 'nav'" value="target">target</option></select></label><input v-if="manual.scope_type !== 'all'" v-model="manual.scope_id" class="ui-control w-full px-3 py-2" placeholder="scope id" /><input v-if="manual.scope_type === 'target'" v-model="manual.target" class="ui-control w-full px-3 py-2" placeholder="target host" /></div><div class="detail-block"><div class="text-sm">Tasks / Protocols</div><div class="mt-2 flex flex-wrap gap-2"><button v-for="task in manualTaskOptions" :key="task" class="ui-button px-2 py-1 text-xs" :class="{ 'ui-button--primary': manual.tasks.includes(task) }" @click="toggleManualTask(task)">{{ task }}</button></div><button class="ui-button ui-button--primary mt-5 px-4 py-2" :disabled="loading || !manual.tasks.length" @click="createManual">Create Job{{ manual.domain === 'nav' && manual.tasks.length > 1 ? 's' : '' }}</button></div></div></section>

    <div v-if="editing" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"><div class="workspace-section w-full max-w-xl space-y-3"><h2 class="text-lg font-semibold">Edit {{ editing.job_key }}</h2><label class="block text-sm">Kind<select v-model="editing.schedule_kind" class="ui-control mt-1 w-full px-3 py-2"><option value="cron">cron</option><option value="interval">interval</option></select></label><input v-if="editing.schedule_kind === 'cron'" v-model="editing.cron_expression" class="ui-control w-full px-3 py-2" placeholder="0 3 * * *" /><template v-else><input v-model.number="editing.interval_seconds" type="number" class="ui-control w-full px-3 py-2" placeholder="interval seconds" /><input v-model="editing.anchor_at" class="ui-control w-full px-3 py-2" placeholder="anchor RFC3339" /></template><input v-model="editing.timezone" class="ui-control w-full px-3 py-2" placeholder="timezone" /><select v-model="editing.misfire_policy" class="ui-control w-full px-3 py-2"><option value="skip">skip</option><option value="catch_up_once">catch_up_once</option></select><input v-model.number="editing.misfire_grace_seconds" type="number" class="ui-control w-full px-3 py-2" placeholder="misfire grace seconds" /><div class="flex justify-end gap-2"><button class="ui-button px-3 py-2" @click="editing = null">Cancel</button><button class="ui-button ui-button--primary px-3 py-2" @click="saveSchedule()">Save</button></div></div></div>
    <div v-if="resultRun" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"><div class="workspace-section max-h-[90vh] w-full max-w-6xl overflow-auto"><div class="flex items-center justify-between gap-3"><div><h2 class="text-lg font-semibold">Task Results</h2><p class="font-mono text-xs text-[var(--text-muted)]">{{ resultRun.domain }} · {{ resultRun.id }}</p></div><button class="ui-button px-3 py-2" @click="resultRun = null">Close</button></div><div class="mt-3 flex flex-wrap gap-2"><template v-if="resultRun.domain === 'game'"><input v-model="resultSearch.game_id" class="ui-control px-2 py-1" placeholder="game_id" /><input v-model="resultSearch.appid" class="ui-control px-2 py-1" placeholder="appid" /></template><template v-else><input v-model="resultSearch.site_id" class="ui-control px-2 py-1" placeholder="site_id" /><input v-model="resultSearch.target" class="ui-control px-2 py-1" placeholder="target" /><input v-model="resultSearch.protocol" class="ui-control px-2 py-1" placeholder="protocol" /></template><button class="ui-button ui-button--primary px-3 py-1" @click="loadResults()">Search</button></div><div class="mt-3 overflow-x-auto"><table class="w-full min-w-[950px] text-left text-sm"><thead class="text-xs uppercase text-[var(--text-muted)]"><tr><th>Started</th><th>Task</th><th>Entity</th><th>Target / AppID</th><th>Status</th><th>Observation</th><th>Duration</th><th>Error</th></tr></thead><tbody><tr v-for="item in resultRows" :key="item.id" class="border-t border-[var(--line)]"><td class="py-2">{{ formatTime(item.started_at) }}</td><td>{{ item.task }}</td><td>{{ item.entity_id }}</td><td>{{ item.target || item.appid || '-' }}</td><td :class="statusClass(item.status)">{{ item.status }}</td><td>{{ item.observation_id || '-' }}</td><td>{{ formatDuration(item.duration_ms) }}</td><td>{{ item.error_kind }} {{ item.error_message }}</td></tr></tbody></table></div></div></div>
  </main>
</template>
