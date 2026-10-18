import { onBeforeUnmount, onMounted, toValue, watch, type MaybeRefOrGetter, type Ref } from 'vue'
import type { GameShowcaseClickSource, GameShowcaseItem } from '~/types/game'
import { submitGameShowcaseEvent } from '~/services/game'

const SESSION_KEY = 'gofurry-showcase-session'
const UUID_V4 = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/

export function useGameShowcaseTracking(
  root: Ref<HTMLElement | null>,
  snapshotId: MaybeRefOrGetter<string>,
  active: MaybeRefOrGetter<GameShowcaseItem | undefined>,
) {
  let observer: IntersectionObserver | undefined
  let timer: ReturnType<typeof setTimeout> | undefined
  let visible = false
  let mounted = false
  let memorySession: string | undefined
  const impressions = new Set<string>()

  function session() {
    if (typeof globalThis.crypto?.randomUUID !== 'function') return undefined
    if (memorySession) return memorySession
    try {
      const saved = sessionStorage.getItem(SESSION_KEY)
      if (saved && UUID_V4.test(saved)) return (memorySession = saved)
    } catch { /* Storage may be blocked; retain only a page-instance session. */ }
    try {
      const id = crypto.randomUUID()
      if (!UUID_V4.test(id)) return undefined
      memorySession = id
      try { sessionStorage.setItem(SESSION_KEY, id) } catch { /* Memory fallback. */ }
      return id
    } catch { return undefined }
  }

  function cancel() {
    if (timer !== undefined) clearTimeout(timer)
    timer = undefined
  }

  function eligible() {
    return mounted && visible && document.visibilityState === 'visible'
  }

  function evaluate() {
    cancel()
    const item = toValue(active)
    const identity = `${toValue(snapshotId)}:${item?.key}`
    if (!eligible() || !item?.tracking_token || impressions.has(identity)) return
    timer = setTimeout(() => {
      timer = undefined
      if (!eligible() || toValue(active)?.key !== item.key || `${toValue(snapshotId)}:${item.key}` !== identity) return
      const sessionID = session()
      if (!sessionID) return
      impressions.add(identity)
      void submitGameShowcaseEvent({ tracking_token: item.tracking_token, session_id: sessionID, event: 'impression' })
    }, 1000)
  }

  watch([() => toValue(snapshotId), () => toValue(active)?.key], evaluate, { flush: 'sync' })
  onMounted(() => {
    mounted = true
    document.addEventListener('visibilitychange', evaluate)
    if (typeof IntersectionObserver !== 'undefined' && root.value) {
      observer = new IntersectionObserver(entries => {
        // A batch may contain exit and re-entry for the same root. Process every
        // threshold crossing so even a brief interruption restarts the interval.
        for (const entry of entries) {
          if (entry.target !== root.value) continue
          const next = entry.isIntersecting && entry.intersectionRatio >= 0.5
          if (next !== visible) { visible = next; evaluate() }
        }
      }, { threshold: [0, 0.5] })
      observer.observe(root.value)
    }
  })
  onBeforeUnmount(() => {
    mounted = false
    cancel()
    observer?.disconnect()
    document.removeEventListener('visibilitychange', evaluate)
  })

  function click(source: GameShowcaseClickSource, item = toValue(active)) {
    if (!mounted || !item?.tracking_token) return
    const sessionID = session()
    if (!sessionID) return
    void submitGameShowcaseEvent({ tracking_token: item.tracking_token, session_id: sessionID, event: 'click', source })
  }

  return { click }
}
