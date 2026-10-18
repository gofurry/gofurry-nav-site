import { computed, onBeforeUnmount, onMounted, ref, toValue, watch, type MaybeRefOrGetter, type Ref } from 'vue'

export const AUTO_ADVANCE_MS = 6000

interface AutoplayState {
  snapshotId: MaybeRefOrGetter<string>
  index: MaybeRefOrGetter<number>
  itemCount: MaybeRefOrGetter<number>
  artworkReady: MaybeRefOrGetter<boolean>
}

// Only owns when to advance. Artwork readiness comes from its existing owner;
// qualified impressions and their separate observer remain Tracking's concern.
export function useGameShowcaseAutoplay(root: Ref<HTMLElement | null>, state: AutoplayState, advance: () => void) {
  const mounted = ref(false)
  const userPaused = ref(false)
  const reducedMotion = ref(true)
  const showControl = computed(() => mounted.value && toValue(state.itemCount) > 1 && !reducedMotion.value)
  let timer: ReturnType<typeof setTimeout> | undefined
  let observer: IntersectionObserver | undefined
  let media: MediaQueryList | undefined
  let element: HTMLElement | null = null
  let visible = false
  let hovered = false
  let focused = false

  function cancel() {
    if (timer !== undefined) clearTimeout(timer)
    timer = undefined
  }
  function eligible() {
    const index = toValue(state.index), count = toValue(state.itemCount)
    return mounted.value && count > 1 && index >= 0 && index < count
      && toValue(state.artworkReady) && visible && document.visibilityState === 'visible'
      && !hovered && !focused && !userPaused.value && !reducedMotion.value
  }
  function restart() {
    cancel()
    if (!eligible()) return
    const snapshot = toValue(state.snapshotId), index = toValue(state.index)
    timer = setTimeout(() => {
      timer = undefined
      if (eligible() && snapshot === toValue(state.snapshotId) && index === toValue(state.index)) advance()
    }, AUTO_ADVANCE_MS)
  }
  function pointerEnter(event: PointerEvent) {
    if (event.pointerType !== 'mouse') return
    hovered = true; restart()
  }
  function pointerLeave(event: PointerEvent) {
    if (event.pointerType !== 'mouse') return
    hovered = false; restart()
  }
  function focusIn() { focused = true; restart() }
  function focusOut(event: FocusEvent) {
    focused = event.relatedTarget instanceof Node && !!element?.contains(event.relatedTarget)
    restart()
  }
  function motionChanged() {
    // Hiding a focused Pause/Play button may remove it without focusout.
    // Re-read focus before re-enabling motion instead of retaining that stale flag.
    focused = !!element?.contains(document.activeElement)
    reducedMotion.value = media?.matches ?? true
  }
  function togglePause() { userPaused.value = !userPaused.value }

  watch([() => toValue(state.snapshotId), () => toValue(state.index), () => toValue(state.itemCount),
    () => toValue(state.artworkReady), userPaused, reducedMotion], restart, { flush: 'sync' })

  onMounted(() => {
    element = root.value
    if (!element) return
    media = window.matchMedia('(prefers-reduced-motion: reduce)')
    motionChanged()
    media.addEventListener('change', motionChanged)
    hovered = element.matches(':hover')
    focused = element.contains(document.activeElement)
    element.addEventListener('pointerenter', pointerEnter)
    element.addEventListener('pointerleave', pointerLeave)
    element.addEventListener('focusin', focusIn)
    element.addEventListener('focusout', focusOut)
    document.addEventListener('visibilitychange', restart)
    mounted.value = true
    if (typeof IntersectionObserver !== 'undefined') {
      observer = new IntersectionObserver(entries => {
        for (const entry of entries) {
          if (entry.target !== element) continue
          const next = entry.isIntersecting && entry.intersectionRatio >= 0.5
          if (next !== visible) { visible = next; restart() }
        }
      }, { threshold: [0, 0.5] })
      observer.observe(element)
    }
  })
  onBeforeUnmount(() => {
    mounted.value = false
    cancel()
    observer?.disconnect()
    media?.removeEventListener('change', motionChanged)
    document.removeEventListener('visibilitychange', restart)
    element?.removeEventListener('pointerenter', pointerEnter)
    element?.removeEventListener('pointerleave', pointerLeave)
    element?.removeEventListener('focusin', focusIn)
    element?.removeEventListener('focusout', focusOut)
  })
  return { userPaused, showControl, togglePause, restart }
}
