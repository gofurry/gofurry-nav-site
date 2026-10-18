import { computed, onBeforeUnmount, onMounted, ref, shallowReactive, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { getNavInsightsTrend } from '~/services/nav'
import type { InsightRange, InsightTrendPoint, NavInsightMetricKey } from '~/types/insights'

export interface SiteInsightTrendPresentation {
  state: 'loading' | 'ready' | 'empty' | 'unavailable'
  points: InsightTrendPoint[]
}

/** Page-session ecosystem cache. A late response can only populate its own identity. */
export function useSiteInsightTrend(input: {
  active: MaybeRefOrGetter<boolean>; metric: MaybeRefOrGetter<NavInsightMetricKey>; range: MaybeRefOrGetter<InsightRange>
}) {
  const nuxt = useNuxtApp(), hydrated = ref(false)
  const entries = shallowReactive(new Map<string, SiteInsightTrendPresentation>())
  const key = computed(() => JSON.stringify([toValue(input.metric), toValue(input.range)]))
  const current = computed<SiteInsightTrendPresentation>(() => entries.get(key.value) ?? { state: 'loading', points: [] })
  let alive = true
  async function load(force = false) {
    if (!hydrated.value || !toValue(input.active)) return
    const identity = key.value, metric = toValue(input.metric), range = toValue(input.range)
    if (entries.has(identity) && (!force || entries.get(identity)?.state === 'loading')) return
    entries.set(identity, { state: 'loading', points: [] })
    try {
      const result = await nuxt.runWithContext(() => getNavInsightsTrend(metric, range))
      if (result.key !== metric || result.requested_range !== range) throw new Error('Ecosystem trend identity mismatch')
      if (alive) entries.set(identity, { state: result.points.length ? 'ready' : 'empty', points: result.points })
    } catch {
      if (alive) entries.set(identity, { state: 'unavailable', points: [] })
    }
  }
  onMounted(() => { hydrated.value = true })
  onBeforeUnmount(() => { alive = false })
  watch([hydrated, key, () => toValue(input.active)], () => { void load() }, { immediate: true, flush: 'post' })
  return { current, retry: () => load(true) }
}
