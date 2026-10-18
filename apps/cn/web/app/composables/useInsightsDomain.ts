import { computed, onMounted, ref, watch } from 'vue'
import type {
  GameInsightMetricKey,
  InsightDomain,
  InsightFeedItem,
  InsightMetric,
  InsightMetricKey,
  InsightMetricTrend,
  InsightOverview,
  InsightRange,
  NavInsightMetricKey,
} from '@/types/insights'

type DomainMetricKey = NavInsightMetricKey | GameInsightMetricKey

interface DomainOptions<Key extends DomainMetricKey> {
  domain: InsightDomain
  defaultMetric: Key
  metrics: readonly Key[]
  getOverview: () => Promise<InsightOverview>
  getTrend: (metric: Key, range: InsightRange) => Promise<InsightMetricTrend>
}

interface DomainSnapshot {
  overview: InsightOverview | null
  overviewUnavailable: boolean
  trend: InsightMetricTrend | null
  trendUnavailable: boolean
  trendDeferred: boolean
}

export async function useInsightsDomain<Key extends DomainMetricKey>(options: DomainOptions<Key>) {
  const route = useRoute()
  const router = useRouter()
  const initialMetric = normalizeMetric(queryValue(route.query.metric), options.metrics, options.defaultMetric)
  const initialRange = normalizeRange(queryValue(route.query.range))
  const snapshotKey = `insights:${options.domain}:${initialMetric}:${initialRange}`

  const snapshot = useAsyncData<DomainSnapshot>(snapshotKey, async () => {
    const trendDeferred = initialRange === 'all'
    const [overviewResult, trendResult] = await Promise.allSettled([
      options.getOverview(),
      trendDeferred ? Promise.resolve(null) : options.getTrend(initialMetric, initialRange),
    ])
    return {
      overview: overviewResult.status === 'fulfilled' ? overviewResult.value : null,
      overviewUnavailable: overviewResult.status === 'rejected',
      trend: trendResult.status === 'fulfilled' ? trendResult.value : null,
      trendUnavailable: !trendDeferred && trendResult.status === 'rejected',
      trendDeferred,
    }
  }, {
    default: () => ({
      overview: null,
      overviewUnavailable: true,
      trend: null,
      trendUnavailable: false,
      trendDeferred: initialRange === 'all',
    }),
  })

  const overview = ref<InsightOverview | null>(null)
  const overviewUnavailable = ref(false)
  const trend = ref<InsightMetricTrend | null>(null)
  const trendUnavailable = ref(false)
  const trendLoading = ref(initialRange === 'all')
  const interactionReady = ref(false)
  let requestSequence = 0

  const selectedMetric = computed<Key>(() =>
    normalizeMetric(queryValue(route.query.metric), options.metrics, options.defaultMetric),
  )
  const selectedRange = computed<InsightRange>(() => normalizeRange(queryValue(route.query.range)))
  const metricsByKey = computed(() => new Map(
    (overview.value?.metrics ?? []).map(metric => [metric.key, metric]),
  ))
  const selectedMetricData = computed<InsightMetric | null>(() =>
    metricsByKey.value.get(selectedMetric.value) ?? null,
  )
  const recentChanges = computed<InsightFeedItem[]>(() =>
    (overview.value?.recent_changes ?? []).map(change => ({ ...change, domain: options.domain })),
  )

  async function loadTrend() {
    const sequence = ++requestSequence
    trendLoading.value = true
    trendUnavailable.value = false
    trend.value = null
    try {
      const response = await options.getTrend(selectedMetric.value, selectedRange.value)
      if (sequence === requestSequence) trend.value = response
    } catch {
      if (sequence === requestSequence) trendUnavailable.value = true
    } finally {
      if (sequence === requestSequence) trendLoading.value = false
    }
  }

  function selectMetric(metric: InsightMetricKey) {
    if (!options.metrics.includes(metric as Key) || metric === selectedMetric.value) return
    void router.push({
      path: route.path,
      query: { ...route.query, metric, range: selectedRange.value },
      hash: route.hash,
    })
  }

  function selectRange(range: InsightRange) {
    if (range === selectedRange.value) return
    void router.push({
      path: route.path,
      query: { ...route.query, metric: selectedMetric.value, range },
      hash: route.hash,
    })
  }

  watch([selectedMetric, selectedRange], () => {
    if (interactionReady.value) void loadTrend()
  })

  onMounted(async () => {
    const rawMetric = queryValue(route.query.metric)
    const rawRange = queryValue(route.query.range)
    if (rawMetric !== selectedMetric.value || rawRange !== selectedRange.value) {
      await router.replace({
        path: route.path,
        query: { ...route.query, metric: selectedMetric.value, range: selectedRange.value },
        hash: route.hash,
      })
    }
    interactionReady.value = true
    if (initialRange === 'all') await loadTrend()
  })

  const { data } = await snapshot
  overview.value = data.value.overview
  overviewUnavailable.value = data.value.overviewUnavailable
  trend.value = data.value.trend
  trendUnavailable.value = data.value.trendUnavailable
  trendLoading.value = data.value.trendDeferred

  return {
    metrics: options.metrics,
    overview,
    overviewUnavailable,
    trend,
    trendUnavailable,
    trendLoading,
    selectedMetric,
    selectedRange,
    selectedMetricData,
    metricsByKey,
    recentChanges,
    selectMetric,
    selectRange,
  }
}

function queryValue(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function normalizeMetric<Key extends DomainMetricKey>(value: string, metrics: readonly Key[], fallback: Key): Key {
  return metrics.includes(value as Key) ? value as Key : fallback
}

function normalizeRange(value: string): InsightRange {
  return value === '90d' || value === 'all' ? value : '30d'
}
