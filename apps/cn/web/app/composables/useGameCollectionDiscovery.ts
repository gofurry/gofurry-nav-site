import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import { readDisplayMode, subscribeModeChange } from '~/utils/modeStorage'
import type { GameCollectionCriteria, GameCollectionIndex, GameCollectionMode } from '~/types/game'

export const COLLECTION_SEARCH_DEBOUNCE_MS = 350

// One route-local owner for criteria, mode and pagination. Only accepted page-1
// results replace ready cards; pagination belongs to that exact generation.
export function useGameCollectionDiscovery(options: {
  initial: GameCollectionIndex | null
  initialError?: unknown
  load(mode: GameCollectionMode, page: number, criteria: GameCollectionCriteria): Promise<GameCollectionIndex>
}) {
  const criteria = shallowRef<GameCollectionCriteria>({ q: '', phase: 'all', sort: 'published_desc' })
  const inputValue = ref('')
  const snapshot = shallowRef(options.initial)
  const mode = ref<GameCollectionMode>('sfw')
  const pending = ref(false)
  const error = shallowRef<unknown>(options.initialError ?? null)
  const loadingMore = ref(false)
  const loadError = ref(false)
  let generation = 0
  let snapshotGeneration = 0
  let disposed = false
  let composing = false
  let interruptedDebounce = false
  let timer: ReturnType<typeof setTimeout> | undefined
  let unsubscribe: (() => void) | undefined
  const cancelTimer = () => { clearTimeout(timer); timer = undefined }
  function invalidate() {
    generation++
    loadingMore.value = false
    loadError.value = false
  }
  async function refresh() {
    cancelTimer()
    invalidate()
    const request = generation
    pending.value = true
    error.value = null
    try {
      const result = await options.load(mode.value, 1, { ...criteria.value })
      if (disposed || request !== generation) return
      snapshot.value = result
      snapshotGeneration = request
    } catch (cause) {
      if (!disposed && request === generation) error.value = cause
    } finally {
      if (!disposed && request === generation) pending.value = false
    }
  }
  function commitText(value: string) {
    inputValue.value = value
    const q = value.trim()
    if (q === criteria.value.q) return
    cancelTimer()
    criteria.value = { ...criteria.value, q }
    invalidate() // Invalidate old load-more responses before the debounce elapses.
    pending.value = true
    error.value = null
    timer = setTimeout(() => { void refresh() }, COLLECTION_SEARCH_DEBOUNCE_MS)
  }
  function onInput(event: Event) {
    const value = (event.target as HTMLInputElement).value
    inputValue.value = value
    if (!composing) commitText(value)
  }
  function onCompositionStart() { composing = true; interruptedDebounce = timer !== undefined; cancelTimer() }
  function onCompositionEnd(event: CompositionEvent) {
    composing = false
    const value = (event.target as HTMLInputElement).value
    // A prior English debounce may have been interrupted by composition.
    if (value.trim() === criteria.value.q && interruptedDebounce) {
      timer = setTimeout(() => { void refresh() }, COLLECTION_SEARCH_DEBOUNCE_MS)
    } else commitText(value)
  }
  function applyFilter(filter: Pick<GameCollectionCriteria, 'phase' | 'sort'>) {
    if (filter.phase === criteria.value.phase && filter.sort === criteria.value.sort) return
    criteria.value = { ...criteria.value, ...filter }
    void refresh()
  }
  async function loadMore() {
    if (!snapshot.value?.has_more || pending.value || loadingMore.value || error.value || snapshotGeneration !== generation) return
    const request = generation
    const current = snapshot.value
    loadingMore.value = true
    loadError.value = false
    try {
      const result = await options.load(mode.value, current.page + 1, { ...criteria.value })
      if (disposed || request !== generation) return
      const seen = new Set(current.items.map(item => item.code))
      const additions = result.items.filter(item => { if (seen.has(item.code)) return false; seen.add(item.code); return true })
      snapshot.value = { ...result, items: [...current.items, ...additions] }
    } catch {
      if (!disposed && request === generation) loadError.value = true
    } finally {
      if (!disposed && request === generation) loadingMore.value = false
    }
  }
  onMounted(() => {
    const change = (next: GameCollectionMode) => {
      if (next === mode.value) return
      mode.value = next
      void refresh()
    }
    try {
      unsubscribe = subscribeModeChange(({ displayMode }) => change(displayMode))
      change(readDisplayMode())
    } catch { /* Keep the SSR SFW snapshot if storage is unavailable. */ }
  })
  onBeforeUnmount(() => { disposed = true; invalidate(); cancelTimer(); unsubscribe?.() })
  return { criteria, inputValue, snapshot, mode, pending, error, loadingMore, loadError,
    refresh, applyFilter, loadMore, onInput, onCompositionStart, onCompositionEnd }
}
