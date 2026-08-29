<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getJSON } from '../api'

type MetricCounts = {
  population_count: number
  eligible_count: number
  not_applicable_count: number
  positive_count: number
  negative_count: number
  stale_count: number
  not_probed_count: number
  probe_failed_count: number
  unknown_count: number
  adoption_rate: number | null
  coverage_rate: number | null
}

type Overview = MetricCounts & {
  domain: 'game' | 'nav'
  metric_key: string
  metric_version: number
  description: string
  processed_through: string | null
  upstream_processed_through: string | null
  lag_days: number | null
  latest_fact_date: string | null
}

type Registry = {
  domain: 'game' | 'nav'
  metric_key: string
  metric_version: number
  metric_kind: string
  entity_level: string
  source_facts: string[]
  eligibility_policy: string
  state_policy: string
  freshness_seconds: number | null
  allowed_dimensions: string[]
  status: string
  description: string
}

type Checkpoint = {
  domain: 'game' | 'nav'
  metric_key: string
  metric_version: number
  status: string
  source_start_date: string | null
  processed_through: string | null
  upstream_processed_through: string | null
  lag_days: number | null
  updated_at: string
}

type Daily = MetricCounts & {
  domain: 'game' | 'nav'
  metric_key: string
  metric_version: number
  fact_date: string
  dimension_key: string
  dimension_value: string
  computed_at: string
}

type Entity = {
  domain: 'game' | 'nav'
  entity_id: number
  historical_name: string
  state: string
  reason_code: string
  source_observed_at: string | null
  dimension_values: Record<string, unknown>
  source_projection_versions: Record<string, number>
}

type EntityPage = { total: number; list: Entity[] }
type Tab = 'overview' | 'registry' | 'checkpoints' | 'daily' | 'entities'

const tabs: Array<{ key: Tab; label: string }> = [
  { key: 'overview', label: 'Overview' },
  { key: 'registry', label: 'Registry' },
  { key: 'checkpoints', label: 'Checkpoints' },
  { key: 'daily', label: 'Daily Results' },
  { key: 'entities', label: 'Entity Explorer' },
]
const states = ['', 'positive', 'negative', 'stale', 'not_probed', 'probe_failed', 'unknown', 'not_applicable']

const activeTab = ref<Tab>('overview')
const loading = ref(false)
const error = ref('')
const overview = ref<Overview[]>([])
const registry = ref<Registry[]>([])
const checkpoints = ref<Checkpoint[]>([])
const daily = ref<Daily[]>([])
const entityPage = ref<EntityPage>({ total: 0, list: [] })

const dailyFilter = ref({ domain: 'nav', metric: '', version: 1, from: '', to: '', dimensionKey: 'global', dimensionValue: 'all' })
const entityFilter = ref({ domain: 'nav', metric: '', version: 1, factDate: '', state: '', reasonCode: '', page: 1, pageSize: 50 })

const activeRegistry = computed(() => registry.value.filter((item) => item.status === 'active'))
const dailyMetrics = computed(() => activeRegistry.value.filter((item) => item.domain === dailyFilter.value.domain))
const entityMetrics = computed(() => activeRegistry.value.filter((item) => item.domain === entityFilter.value.domain))

async function loadFoundation() {
  loading.value = true
  error.value = ''
  try {
    ;[overview.value, registry.value, checkpoints.value] = await Promise.all([
      getJSON<Overview[]>('/api/v1/metrics/overview'),
      getJSON<Registry[]>('/api/v1/metrics/registry'),
      getJSON<Checkpoint[]>('/api/v1/metrics/checkpoints'),
    ])
    seedFilters()
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause)
  } finally {
    loading.value = false
  }
}

function seedFilters() {
  const navMetric = activeRegistry.value.find((item) => item.domain === 'nav') ?? activeRegistry.value[0]
  if (!dailyFilter.value.metric && navMetric) {
    dailyFilter.value.domain = navMetric.domain
    dailyFilter.value.metric = navMetric.metric_key
    dailyFilter.value.version = navMetric.metric_version
  }
  if (!entityFilter.value.metric && navMetric) {
    entityFilter.value.domain = navMetric.domain
    entityFilter.value.metric = navMetric.metric_key
    entityFilter.value.version = navMetric.metric_version
    entityFilter.value.factDate = overview.value.find((item) => item.domain === navMetric.domain && item.metric_key === navMetric.metric_key)?.latest_fact_date ?? ''
  }
}

function syncDailyMetric() {
  const item = dailyMetrics.value[0]
  dailyFilter.value.metric = item?.metric_key ?? ''
  dailyFilter.value.version = item?.metric_version ?? 1
}

function syncDailyVersion() {
  const item = dailyMetrics.value.find((candidate) => candidate.metric_key === dailyFilter.value.metric)
  dailyFilter.value.version = item?.metric_version ?? 1
}

function syncEntityMetric() {
  const item = entityMetrics.value[0]
  entityFilter.value.metric = item?.metric_key ?? ''
  entityFilter.value.version = item?.metric_version ?? 1
  entityFilter.value.factDate = overview.value.find((row) => row.domain === item?.domain && row.metric_key === item?.metric_key)?.latest_fact_date ?? ''
  entityFilter.value.page = 1
}

function syncEntityVersion() {
  const item = entityMetrics.value.find((candidate) => candidate.metric_key === entityFilter.value.metric)
  entityFilter.value.version = item?.metric_version ?? 1
  entityFilter.value.factDate = overview.value.find((row) => row.domain === item?.domain && row.metric_key === item?.metric_key)?.latest_fact_date ?? ''
  entityFilter.value.page = 1
}

async function loadDaily() {
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      domain: dailyFilter.value.domain,
      metric: dailyFilter.value.metric,
      version: String(dailyFilter.value.version),
      dimension_key: dailyFilter.value.dimensionKey || 'global',
      dimension_value: dailyFilter.value.dimensionValue || 'all',
    })
    if (dailyFilter.value.from) params.set('from', dailyFilter.value.from)
    if (dailyFilter.value.to) params.set('to', dailyFilter.value.to)
    daily.value = await getJSON<Daily[]>(`/api/v1/metrics/daily?${params}`)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause)
  } finally {
    loading.value = false
  }
}

async function loadEntities(resetPage = false) {
  if (resetPage) entityFilter.value.page = 1
  if (!entityFilter.value.metric || !entityFilter.value.factDate) return
  loading.value = true
  error.value = ''
  try {
    const params = new URLSearchParams({
      domain: entityFilter.value.domain,
      metric: entityFilter.value.metric,
      version: String(entityFilter.value.version),
      fact_date: entityFilter.value.factDate,
      page: String(entityFilter.value.page),
      page_size: String(entityFilter.value.pageSize),
    })
    if (entityFilter.value.state) params.set('state', entityFilter.value.state)
    if (entityFilter.value.reasonCode) params.set('reason_code', entityFilter.value.reasonCode)
    entityPage.value = await getJSON<EntityPage>(`/api/v1/metrics/entities?${params}`)
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause)
  } finally {
    loading.value = false
  }
}

function selectTab(tab: Tab) {
  activeTab.value = tab
  if (tab === 'daily' && daily.value.length === 0) void loadDaily()
  if (tab === 'entities' && entityPage.value.list.length === 0) void loadEntities()
}

function formatRate(value: number | null) {
  return value === null ? '—' : `${(value * 100).toFixed(1)}%`
}

function stateClass(state: string) {
  if (state === 'positive') return 'text-emerald-300'
  if (state === 'negative' || state === 'probe_failed') return 'text-[var(--danger)]'
  if (state === 'stale' || state === 'not_probed') return 'text-[var(--warning)]'
  return 'text-[var(--text-muted)]'
}

function previousPage() {
  if (entityFilter.value.page > 1) {
    entityFilter.value.page--
    void loadEntities()
  }
}

function nextPage() {
  if (entityFilter.value.page * entityFilter.value.pageSize < entityPage.value.total) {
    entityFilter.value.page++
    void loadEntities()
  }
}

onMounted(loadFoundation)
</script>

<template>
  <main>
    <header class="page-header">
      <div>
        <div class="page-kicker">P0.3 · read-only verification</div>
        <h1 class="mt-1 text-2xl font-semibold">Metric Center</h1>
        <p class="mt-2 text-sm text-[var(--text-muted)]">Versioned metric contracts, checkpoints, daily counts and explainable entity states.</p>
      </div>
      <button class="ui-button px-4" :disabled="loading" @click="loadFoundation">{{ loading ? 'Loading…' : 'Refresh' }}</button>
    </header>

    <div class="mt-5 flex flex-wrap gap-2 border-b border-[var(--line)] pb-3">
      <button v-for="tab in tabs" :key="tab.key" class="ui-button px-4" :class="{ 'ui-button--primary': activeTab === tab.key }" @click="selectTab(tab.key)">{{ tab.label }}</button>
    </div>
    <p v-if="error" class="status-alert mt-4 px-4 py-3 text-sm text-[var(--danger)]">{{ error }}</p>

    <section v-if="activeTab === 'overview'" class="mt-5 grid gap-4 xl:grid-cols-3 md:grid-cols-2">
      <article v-for="item in overview" :key="`${item.domain}-${item.metric_key}`" class="data-surface border border-[var(--line)] p-4">
        <div class="flex items-start justify-between gap-3"><div><p class="page-kicker">{{ item.domain }} · v{{ item.metric_version }}</p><h2 class="mt-1 text-lg font-semibold">{{ item.metric_key }}</h2></div><span class="font-mono text-xs text-[var(--text-muted)]">{{ item.latest_fact_date ?? 'not computed' }}</span></div>
        <p class="mt-2 min-h-10 text-xs text-[var(--text-muted)]">{{ item.description }}</p>
        <div class="mt-4 grid grid-cols-2 gap-2"><div class="summary-cell"><span class="text-xs text-[var(--text-muted)]">Adoption</span><strong class="mt-1 block text-xl">{{ formatRate(item.adoption_rate) }}</strong></div><div class="summary-cell"><span class="text-xs text-[var(--text-muted)]">Coverage</span><strong class="mt-1 block text-xl">{{ formatRate(item.coverage_rate) }}</strong></div></div>
        <dl class="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-xs"><template v-for="key in ['population_count','positive_count','negative_count','stale_count','not_probed_count','probe_failed_count','unknown_count','not_applicable_count']" :key="key"><dt class="text-[var(--text-muted)]">{{ key.replaceAll('_', ' ') }}</dt><dd class="text-right font-mono">{{ item[key as keyof MetricCounts] }}</dd></template></dl>
        <p class="mt-4 border-t border-[var(--line)] pt-3 text-xs text-[var(--text-muted)]">Checkpoint {{ item.processed_through ?? '—' }} · upstream {{ item.upstream_processed_through ?? '—' }} · lag {{ item.lag_days ?? '—' }}</p>
      </article>
    </section>

    <section v-else-if="activeTab === 'registry'" class="resource-table-section"><div class="resource-table-scroll"><table class="resource-table min-w-[1200px]"><thead><tr><th>Domain</th><th>Metric</th><th>Contract</th><th>Sources</th><th>Policies</th><th>Freshness</th><th>Dimensions</th><th>Status</th></tr></thead><tbody><tr v-for="item in registry" :key="`${item.domain}-${item.metric_key}-${item.metric_version}`"><td>{{ item.domain }}</td><td><strong>{{ item.metric_key }}</strong><div class="font-mono text-xs text-[var(--text-muted)]">v{{ item.metric_version }} · {{ item.metric_kind }} · {{ item.entity_level }}</div></td><td class="max-w-72 text-xs">{{ item.description }}</td><td class="font-mono text-xs">{{ item.source_facts.join(', ') }}</td><td class="font-mono text-xs">{{ item.eligibility_policy }}<br>{{ item.state_policy }}</td><td>{{ item.freshness_seconds ?? '—' }}s</td><td>{{ item.allowed_dimensions.join(', ') }}</td><td :class="item.status === 'active' ? 'text-emerald-300' : 'text-[var(--text-muted)]'">{{ item.status }}</td></tr></tbody></table></div></section>

    <section v-else-if="activeTab === 'checkpoints'" class="resource-table-section"><div class="resource-table-scroll"><table class="resource-table"><thead><tr><th>Domain</th><th>Metric</th><th>Source Start</th><th>Processed</th><th>Upstream</th><th>Lag</th><th>Status</th><th>Updated</th></tr></thead><tbody><tr v-for="item in checkpoints" :key="`${item.domain}-${item.metric_key}-${item.metric_version}`"><td>{{ item.domain }}</td><td>{{ item.metric_key }} v{{ item.metric_version }}</td><td>{{ item.source_start_date ?? '—' }}</td><td>{{ item.processed_through ?? '—' }}</td><td>{{ item.upstream_processed_through ?? '—' }}</td><td>{{ item.lag_days ?? '—' }}</td><td>{{ item.status }}</td><td class="font-mono text-xs">{{ item.updated_at }}</td></tr></tbody></table></div></section>

    <section v-else-if="activeTab === 'daily'" class="mt-5">
      <div class="flex flex-wrap gap-2"><select v-model="dailyFilter.domain" class="ui-control px-3" @change="syncDailyMetric"><option value="game">game</option><option value="nav">nav</option></select><select v-model="dailyFilter.metric" class="ui-control min-w-52 px-3" @change="syncDailyVersion"><option v-for="item in dailyMetrics" :key="item.metric_key" :value="item.metric_key">{{ item.metric_key }} v{{ item.metric_version }}</option></select><input v-model="dailyFilter.from" type="date" class="ui-control px-3"><input v-model="dailyFilter.to" type="date" class="ui-control px-3"><input v-model="dailyFilter.dimensionKey" class="ui-control px-3" placeholder="dimension key"><input v-model="dailyFilter.dimensionValue" class="ui-control px-3" placeholder="dimension value"><button class="ui-button ui-button--primary px-4" @click="loadDaily">Query</button></div>
      <div class="resource-table-section"><div class="resource-table-scroll"><table class="resource-table min-w-[1250px]"><thead><tr><th>Date</th><th>Metric</th><th>Dimension</th><th>Population</th><th>Eligible</th><th>Positive</th><th>Negative</th><th>Stale</th><th>Not Probed</th><th>Failed</th><th>Unknown</th><th>N/A</th><th>Adoption</th><th>Coverage</th></tr></thead><tbody><tr v-for="item in daily" :key="`${item.domain}-${item.metric_key}-${item.fact_date}-${item.dimension_key}-${item.dimension_value}`"><td>{{ item.fact_date }}</td><td>{{ item.domain }}/{{ item.metric_key }} v{{ item.metric_version }}</td><td>{{ item.dimension_key }}/{{ item.dimension_value }}</td><td>{{ item.population_count }}</td><td>{{ item.eligible_count }}</td><td>{{ item.positive_count }}</td><td>{{ item.negative_count }}</td><td>{{ item.stale_count }}</td><td>{{ item.not_probed_count }}</td><td>{{ item.probe_failed_count }}</td><td>{{ item.unknown_count }}</td><td>{{ item.not_applicable_count }}</td><td>{{ formatRate(item.adoption_rate) }}</td><td>{{ formatRate(item.coverage_rate) }}</td></tr></tbody></table></div></div>
    </section>

    <section v-else class="mt-5">
      <div class="flex flex-wrap gap-2"><select v-model="entityFilter.domain" class="ui-control px-3" @change="syncEntityMetric"><option value="game">game</option><option value="nav">nav</option></select><select v-model="entityFilter.metric" class="ui-control min-w-52 px-3" @change="syncEntityVersion"><option v-for="item in entityMetrics" :key="item.metric_key" :value="item.metric_key">{{ item.metric_key }} v{{ item.metric_version }}</option></select><input v-model="entityFilter.factDate" type="date" class="ui-control px-3"><select v-model="entityFilter.state" class="ui-control px-3"><option v-for="state in states" :key="state" :value="state">{{ state || 'all states' }}</option></select><input v-model="entityFilter.reasonCode" class="ui-control px-3" placeholder="reason code"><button class="ui-button ui-button--primary px-4" @click="loadEntities(true)">Query</button></div>
      <div class="resource-table-section"><div class="resource-table-scroll"><table class="resource-table min-w-[1100px]"><thead><tr><th>Entity</th><th>State</th><th>Reason</th><th>Evidence</th><th>Dimensions</th><th>Projection Versions</th></tr></thead><tbody><tr v-for="item in entityPage.list" :key="`${item.domain}-${item.entity_id}`"><td><strong>{{ item.historical_name }}</strong><div class="font-mono text-xs text-[var(--text-muted)]">{{ item.domain }} #{{ item.entity_id }}</div></td><td><span class="inline-flex rounded-full border border-current/30 px-2 py-1 font-mono text-xs" :class="stateClass(item.state)">{{ item.state }}</span></td><td class="font-mono text-xs">{{ item.reason_code }}</td><td class="font-mono text-xs">{{ item.source_observed_at ?? '—' }}</td><td><pre class="max-w-80 whitespace-pre-wrap text-xs">{{ JSON.stringify(item.dimension_values) }}</pre></td><td><pre class="max-w-64 whitespace-pre-wrap text-xs">{{ JSON.stringify(item.source_projection_versions) }}</pre></td></tr></tbody></table></div><div class="resource-pagination"><span>{{ entityPage.total }} entities · page {{ entityFilter.page }}</span><button class="ui-button px-3" :disabled="entityFilter.page <= 1" @click="previousPage">Previous</button><button class="ui-button px-3" :disabled="entityFilter.page * entityFilter.pageSize >= entityPage.total" @click="nextPage">Next</button></div></div>
    </section>
  </main>
</template>
