import { computed, onMounted, ref, watch, type ComputedRef } from 'vue'
import type { LocationQueryRaw } from 'vue-router'
import type {
  GameInsightDimension,
  GameInsightMetricKey,
  InsightDimension,
  InsightDimensionBreakdown,
  InsightDimensionTrend,
  InsightRange,
  NavInsightMetricKey,
  SiteInsightDimension,
} from '@/types/insights'
import { normalizeInsightSlice } from '@/utils/insightDimensions'

type DimensionMetricKey = NavInsightMetricKey | GameInsightMetricKey
type DomainDimension = SiteInsightDimension | GameInsightDimension

interface DimensionOptions<Key extends DimensionMetricKey, Dimension extends DomainDimension> {
  domain: 'site' | 'game'
  metric: ComputedRef<Key>
  range: ComputedRef<InsightRange>
  defaultDimension: Dimension
  dimensions: readonly Dimension[]
  getBreakdown: (metric: Key, dimension: Dimension) => Promise<InsightDimensionBreakdown>
  getSliceTrend: (metric: Key, dimension: Dimension, value: string, range: InsightRange) => Promise<InsightDimensionTrend>
}

interface DimensionSnapshot {
  breakdown: InsightDimensionBreakdown | null
  breakdownUnavailable: boolean
  sliceTrend: InsightDimensionTrend | null
  sliceTrendUnavailable: boolean
  sliceTrendDeferred: boolean
}

export async function useInsightsDimensions<Key extends DimensionMetricKey, Dimension extends DomainDimension>(
  options: DimensionOptions<Key, Dimension>,
) {
  const route = useRoute()
  const router = useRouter()
  const initialDimension = normalizeDimension(queryValue(route.query.dimension), options.dimensions, options.defaultDimension)
  const initialSlice = normalizeInsightSlice(initialDimension, queryValue(route.query.slice))
  const initialMetric = options.metric.value
  const initialRange = options.range.value
  const snapshotKey = `insights:dimensions:${options.domain}:${initialMetric}:${initialRange}:${initialDimension}:${initialSlice ?? 'none'}`

  const snapshot = useAsyncData<DimensionSnapshot>(snapshotKey, async () => {
    const sliceTrendDeferred = Boolean(initialSlice) && initialRange === 'all'
    const [breakdownResult, trendResult] = await Promise.allSettled([
      options.getBreakdown(initialMetric, initialDimension),
      initialSlice && !sliceTrendDeferred
        ? options.getSliceTrend(initialMetric, initialDimension, initialSlice, initialRange)
        : Promise.resolve(null),
    ])
    return {
      breakdown: breakdownResult.status === 'fulfilled' ? breakdownResult.value : null,
      breakdownUnavailable: breakdownResult.status === 'rejected',
      sliceTrend: trendResult.status === 'fulfilled' ? trendResult.value : null,
      sliceTrendUnavailable: Boolean(initialSlice) && !sliceTrendDeferred && trendResult.status === 'rejected',
      sliceTrendDeferred,
    }
  }, {
    default: () => ({
      breakdown: null,
      breakdownUnavailable: true,
      sliceTrend: null,
      sliceTrendUnavailable: false,
      sliceTrendDeferred: Boolean(initialSlice) && initialRange === 'all',
    }),
  })

  const breakdown = ref<InsightDimensionBreakdown | null>(null)
  const breakdownUnavailable = ref(false)
  const breakdownLoading = ref(false)
  const sliceTrend = ref<InsightDimensionTrend | null>(null)
  const sliceTrendUnavailable = ref(false)
  const sliceTrendLoading = ref(Boolean(initialSlice) && initialRange === 'all')
  const interactionReady = ref(false)
  let breakdownSequence = 0
  let trendSequence = 0

  const selectedDimension = computed<Dimension>(() =>
    normalizeDimension(queryValue(route.query.dimension), options.dimensions, options.defaultDimension),
  )
  const selectedSlice = computed<string | null>(() =>
    normalizeInsightSlice(selectedDimension.value, queryValue(route.query.slice)),
  )

  async function loadBreakdown() {
    const sequence = ++breakdownSequence
    breakdownLoading.value = true
    breakdownUnavailable.value = false
    breakdown.value = null
    try {
      const response = await options.getBreakdown(options.metric.value, selectedDimension.value)
      if (sequence === breakdownSequence) breakdown.value = response
    } catch {
      if (sequence === breakdownSequence) breakdownUnavailable.value = true
    } finally {
      if (sequence === breakdownSequence) breakdownLoading.value = false
    }
  }

  async function loadSliceTrend() {
    const slice = selectedSlice.value
    const sequence = ++trendSequence
    sliceTrend.value = null
    sliceTrendUnavailable.value = false
    if (!slice) {
      sliceTrendLoading.value = false
      return
    }
    sliceTrendLoading.value = true
    try {
      const response = await options.getSliceTrend(options.metric.value, selectedDimension.value, slice, options.range.value)
      if (sequence === trendSequence) sliceTrend.value = response
    } catch {
      if (sequence === trendSequence) sliceTrendUnavailable.value = true
    } finally {
      if (sequence === trendSequence) sliceTrendLoading.value = false
    }
  }

  function selectDimension(dimension: InsightDimension) {
    if (!options.dimensions.includes(dimension as Dimension) || dimension === selectedDimension.value) return
    const query: LocationQueryRaw = { ...route.query, dimension }
    delete query.slice
    void router.push({ path: route.path, query, hash: route.hash })
  }

  function selectSlice(value: string) {
    const normalized = normalizeInsightSlice(selectedDimension.value, value)
    if (!normalized) return
    const query: LocationQueryRaw = { ...route.query }
    if (normalized === selectedSlice.value) delete query.slice
    else query.slice = normalized
    void router.push({ path: route.path, query, hash: route.hash })
  }

  watch([options.metric, selectedDimension], () => {
    if (interactionReady.value) void loadBreakdown()
  })
  watch([options.metric, options.range, selectedDimension, selectedSlice], () => {
    if (interactionReady.value) void loadSliceTrend()
  })

  onMounted(async () => {
    const rawDimension = queryValue(route.query.dimension)
    const rawSlice = queryValue(route.query.slice)
    if (rawDimension !== selectedDimension.value || rawSlice !== (selectedSlice.value ?? '')) {
      const query: LocationQueryRaw = {
        ...route.query,
        metric: options.metric.value,
        range: options.range.value,
        dimension: selectedDimension.value,
      }
      if (selectedSlice.value) query.slice = selectedSlice.value
      else delete query.slice
      await router.replace({ path: route.path, query, hash: route.hash })
    }
    interactionReady.value = true
    if (initialSlice && initialRange === 'all') await loadSliceTrend()
  })

  const { data } = await snapshot
  breakdown.value = data.value.breakdown
  breakdownUnavailable.value = data.value.breakdownUnavailable
  sliceTrend.value = data.value.sliceTrend
  sliceTrendUnavailable.value = data.value.sliceTrendUnavailable
  sliceTrendLoading.value = data.value.sliceTrendDeferred

  return {
    dimensions: options.dimensions,
    selectedDimension,
    selectedSlice,
    breakdown,
    breakdownUnavailable,
    breakdownLoading,
    sliceTrend,
    sliceTrendUnavailable,
    sliceTrendLoading,
    selectDimension,
    selectSlice,
  }
}

function queryValue(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function normalizeDimension<Dimension extends DomainDimension>(
  value: string,
  dimensions: readonly Dimension[],
  fallback: Dimension,
): Dimension {
  return dimensions.includes(value as Dimension) ? value as Dimension : fallback
}
