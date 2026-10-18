import { ref, shallowRef } from 'vue'

export type ProbeState = 'never' | 'probing' | 'success' | 'timeout' | 'failed' | 'stale'
export type ProbeMeasurement = { ms: number | null; state: 'success' | 'timeout' | 'failed' }
export function isProbeMeasurement(value: unknown): value is ProbeMeasurement {
  if (!value || typeof value !== 'object') return false
  const measurement = value as ProbeMeasurement
  return measurement.state === 'success' ? typeof measurement.ms === 'number' && Number.isFinite(measurement.ms) && measurement.ms >= 0
    : (measurement.state === 'failed' || measurement.state === 'timeout') && measurement.ms === null
}
export const ROUTE_TTL_MS = 12 * 60 * 60 * 1000
export const ROUTE_MODE_MAX_AGE = 365 * 24 * 60 * 60
export const MANUAL_PROBE_COOLDOWN_MS = 60_000

export function isFresh(checkedAt: number, now = Date.now()) {
  return Number.isFinite(checkedAt) && checkedAt > 0 && checkedAt <= now && now - checkedAt < ROUTE_TTL_MS
}

export function averageProbe(rounds: ProbeMeasurement[]): ProbeMeasurement {
  if (rounds.some(result => result.state === 'timeout')) return { ms: null, state: 'timeout' }
  if (rounds.some(result => result.ms === null)) return { ms: null, state: 'failed' }
  return { ms: rounds.reduce((sum, result) => sum + result.ms!, 0) / rounds.length, state: 'success' }
}

export function readRouteStorage(key: string): unknown {
  try { return JSON.parse(localStorage.getItem(key) || 'null') } catch { return null }
}
export function writeRouteStorage(key: string, value: unknown) {
  try { localStorage.setItem(key, JSON.stringify(value)) } catch { /* Storage may be disabled. */ }
}

export type ProbeRecord = { checkedAt: number; lastManualAt?: number; stale?: boolean }
// Only the browser scheduling lifecycle is shared; each domain owns its probe,
// recommendation policy, cookies and fallback candidates.
export function createRouteProbe<T extends ProbeRecord>(run: (manual: boolean) => Promise<T>, persist: (result: T) => void, now = Date.now) {
  const diagnostics = shallowRef<T | null>(null)
  const probing = ref(false)
  const lastManualAt = ref(0)
  let pending: Promise<void> | null = null
  let scheduled = false
  let mounted = false
  const probe = (manual = false): Promise<void> => {
    if (!mounted) return Promise.resolve()
    if (pending) return pending
    if (manual && now() - lastManualAt.value < MANUAL_PROBE_COOLDOWN_MS) return Promise.resolve()
    if (manual) lastManualAt.value = now()
    probing.value = true
    pending = run(manual).then(result => {
      diagnostics.value = { ...result, lastManualAt: lastManualAt.value }
      persist(diagnostics.value)
    }).finally(() => { probing.value = false; pending = null })
    return pending
  }
  const scheduleProbe = () => {
    if (!mounted || scheduled || pending || (diagnostics.value && !diagnostics.value.stale && isFresh(diagnostics.value.checkedAt, now()))) return
    const connection = (navigator as Navigator & { connection?: { saveData?: boolean } }).connection
    if (navigator.onLine === false || connection?.saveData) return
    scheduled = true
    const runIdle = () => {
      scheduled = false
      if (navigator.onLine !== false && !connection?.saveData && (!diagnostics.value || diagnostics.value.stale || !isFresh(diagnostics.value.checkedAt, now()))) void probe()
    }
    if ('requestIdleCallback' in window) window.requestIdleCallback(runIdle, { timeout: 2000 })
    else setTimeout(runIdle, 250)
  }
  const start = (stored: T | null) => {
    mounted = true
    diagnostics.value = stored
    const previous = stored?.lastManualAt || 0
    lastManualAt.value = typeof previous === 'number' && Number.isFinite(previous) && previous >= 0 && previous <= now() ? previous : 0
    scheduleProbe()
  }
  const markStale = () => {
    if (diagnostics.value) { diagnostics.value = { ...diagnostics.value, stale: true }; persist(diagnostics.value) }
    scheduleProbe()
  }
  return { diagnostics, probing, lastManualAt, probe, scheduleProbe, start, markStale }
}
