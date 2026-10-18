import { onMounted, onBeforeUnmount, ref, shallowRef } from 'vue'
import { readDisplayMode, subscribeModeChange } from '~/utils/modeStorage'
import type { GameCollectionMode } from '~/types/game'

// Each route instance owns one SSR SFW snapshot. No URL or global store state.
// The generation also lets pagination reject responses from an older mode.
export function useGameCollectionModeRefresh<T>(options: {
  initial: T | null
  initialError?: unknown
  load(mode: GameCollectionMode): Promise<T>
  onError?(error: unknown): void
}) {
  const snapshot = shallowRef<T | null>(options.initial)
  const mode = ref<GameCollectionMode>('sfw')
  const snapshotMode = ref<GameCollectionMode>('sfw')
  const revision = ref(0)
  const pending = ref(false)
  const error = shallowRef<unknown>(options.initialError ?? null)
  let disposed = false
  let unsubscribe: (() => void) | undefined

  async function refresh() {
    const generation = ++revision.value
    const requestedMode = mode.value
    pending.value = true
    error.value = null
    try {
      const result = await options.load(requestedMode)
      if (disposed || generation !== revision.value) return
      snapshot.value = result
      snapshotMode.value = requestedMode
    } catch (cause) {
      if (disposed || generation !== revision.value) return
      error.value = cause
      options.onError?.(cause)
    } finally {
      if (!disposed && generation === revision.value) pending.value = false
    }
  }

  onMounted(() => {
    const change = (next: GameCollectionMode) => {
      if (next === mode.value) return
      mode.value = next
      void refresh()
    }
    // Blocked local storage keeps the valid SFW snapshot usable.
    try {
      unsubscribe = subscribeModeChange(({ displayMode }) => change(displayMode))
      change(readDisplayMode())
    } catch { /* SFW is the privacy-safe default. */ }
  })
  onBeforeUnmount(() => {
    disposed = true
    revision.value++
    unsubscribe?.()
  })
  return { snapshot, mode, snapshotMode, revision, pending, error, refresh }
}
