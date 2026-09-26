import { computed, onBeforeUnmount, onMounted, ref, shallowReactive, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useApi } from './useApi'
import type { TargetObservationsResponse } from '~/types/nav'
import { presentPingHistory, type PingHistoryRow } from '~/utils/siteObservationPresentation'

export type SiteHistoryState = 'loading' | 'ready' | 'empty' | 'unavailable'
export type SiteHistorySample = 20 | 60 | 100
interface Entry { state: SiteHistoryState; rows: PingHistoryRow[] }
export interface SiteHistoryPresentation extends Entry { sample: SiteHistorySample; total: number }

/** One page-session cache. Captured request keys can only update their own entry. */
export function useSiteObservationHistory(input: {
  siteId: MaybeRefOrGetter<string>; target: MaybeRefOrGetter<string>; active: MaybeRefOrGetter<boolean>
}) {
  const api = useApi('navV2')
  const hydrated = ref(false), sample = ref<SiteHistorySample>(20)
  const entries = shallowReactive(new Map<string, Entry>())
  const key = computed(() => JSON.stringify([toValue(input.siteId), toValue(input.target), 'ping']))
  const current = computed<Entry>(() => entries.get(key.value) ?? { state: 'loading', rows: [] })
  const rows = computed(() => current.value.rows.slice(0, sample.value))
  let alive = true
  async function load(force = false) {
    if (!hydrated.value || !toValue(input.active)) return
    const siteId = toValue(input.siteId), target = toValue(input.target), requestKey = key.value
    if (!siteId || !target || (entries.has(requestKey) && !force)) return
    if (force && entries.get(requestKey)?.state === 'loading') return
    entries.set(requestKey, { state: 'loading', rows: [] })
    try {
      const response = await api<TargetObservationsResponse>(`/nav/sites/${siteId}/targets/${encodeURIComponent(target)}/observations`, {
        query: { protocol: 'ping', limit: 100, payload_mode: 'preview' }, retry: 0,
      })
      if ((response.target && response.target !== target) || (response.protocol && response.protocol !== 'ping')) throw new Error('Observation identity mismatch')
      const result = presentPingHistory(response.items ?? [], target)
      if (alive) entries.set(requestKey, { state: result.length ? 'ready' : 'empty', rows: result })
    } catch {
      if (alive) entries.set(requestKey, { state: 'unavailable', rows: [] })
    }
  }
  onMounted(() => { hydrated.value = true })
  onBeforeUnmount(() => { alive = false })
  watch(key, () => { sample.value = 20 })
  watch([hydrated, key, () => toValue(input.active)], () => { void load() }, { immediate: true, flush: 'post' })
  return { state: computed(() => current.value.state), rows, sample, total: computed(() => current.value.rows.length),
    selectSample: (value: SiteHistorySample) => { sample.value = value }, retry: () => load(true) }
}
export type SiteObservationHistory = ReturnType<typeof useSiteObservationHistory>
