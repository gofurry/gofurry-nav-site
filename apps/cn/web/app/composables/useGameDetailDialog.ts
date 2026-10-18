import { nextTick, onBeforeUnmount, watch, type Ref } from 'vue'

// Detail's Lightbox and adult confirmation are body-mounted. Keep their local
// focus/scroll lifetime independent of Search and Review dialog owners.
export function useGameDetailDialog(panel: Ref<HTMLElement | null>, dismiss: () => void) {
  let cleanup: (() => void) | undefined
  let disposed = false
  watch(panel, async element => {
    cleanup?.(); cleanup = undefined
    if (!element || disposed) return
    const trigger = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const root = document.querySelector<HTMLElement>('#__nuxt')
    const wasInert = root?.inert ?? false
    const html = document.documentElement, overflow = html.style.overflow, gutter = html.style.scrollbarGutter
    const scroll = { left: window.scrollX, top: window.scrollY }
    html.style.scrollbarGutter = 'stable'; html.style.overflow = 'hidden'
    const focusable = () => [...element.querySelectorAll<HTMLElement>('button, a[href], [tabindex]')]
      .filter(node => node.tabIndex >= 0 && !node.matches(':disabled') && node.getClientRects().length && !node.closest('[inert]'))
    const focusInside = () => (focusable()[0] ?? element).focus({ preventScroll: true })
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') { event.preventDefault(); event.stopImmediatePropagation(); dismiss() }
      if (event.key === 'Tab') {
        const nodes = focusable(), first = nodes[0], last = nodes.at(-1)
        if (!first || !last) { event.preventDefault(); element.focus(); return }
        if (event.shiftKey && (document.activeElement === first || !element.contains(document.activeElement))) { event.preventDefault(); last.focus() }
        else if (!event.shiftKey && (document.activeElement === last || !element.contains(document.activeElement))) { event.preventDefault(); first.focus() }
      }
    }
    const onFocus = (event: FocusEvent) => { if (!element.contains(event.target as Node)) focusInside() }
    let active = true
    cleanup = () => {
      active = false
      document.removeEventListener('keydown', onKey, true); document.removeEventListener('focusin', onFocus)
      if (root) root.inert = wasInert
      html.style.overflow = overflow; html.style.scrollbarGutter = gutter
      window.scrollTo({ ...scroll, behavior: 'instant' })
      void nextTick(() => {
        if (!disposed && !panel.value && trigger?.isConnected && !trigger.closest('[inert]')) trigger.focus({ preventScroll: true })
      })
    }
    await nextTick()
    if (!active || disposed || panel.value !== element) return
    focusInside()
    if (root) root.inert = true
    document.addEventListener('keydown', onKey, true); document.addEventListener('focusin', onFocus)
  }, { flush: 'post' })
  onBeforeUnmount(() => { disposed = true; cleanup?.(); cleanup = undefined })
}
