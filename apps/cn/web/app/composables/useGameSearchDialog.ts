import { nextTick, onBeforeUnmount, watch, type Ref } from 'vue'

// Search's two body-mounted dialogs share behavior, not a new global modal system.
export function useGameSearchDialog(panel: Ref<HTMLElement | null>, options: {
  dismiss(): void
  initialFocus(): HTMLElement | null
  fallbackFocus?(): HTMLElement | null
  consumeEscape?(): boolean
}) {
  let cleanup: (() => void) | undefined
  let disposed = false
  watch(panel, async element => {
    cleanup?.(); cleanup = undefined
    if (!element || disposed) return
    const trigger = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const root = document.querySelector<HTMLElement>('#__nuxt')
    const wasInert = root?.inert ?? false
    const html = document.documentElement
    const overflow = html.style.overflow
    const gutter = html.style.scrollbarGutter
    const scroll = { left: window.scrollX, top: window.scrollY }
    // Retain the scrollbar allocation while locking the document scroller.
    html.style.scrollbarGutter = 'stable'
    html.style.overflow = 'hidden'
    const focusable = () => [...element.querySelectorAll<HTMLElement>(
      'button, input, select, textarea, a[href], [tabindex]',
    )].filter(node => node.tabIndex >= 0 && !node.matches(':disabled') && node.getClientRects().length
      && getComputedStyle(node).visibility !== 'hidden' && !node.closest('[inert]'))
    const focusInside = () => (options.initialFocus() ?? focusable()[0] ?? element).focus({ preventScroll: true })
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault(); event.stopImmediatePropagation()
        if (!options.consumeEscape?.()) options.dismiss()
      } else if (event.key === 'Tab') {
        const nodes = focusable(), first = nodes[0], last = nodes.at(-1)
        if (!first || !last) { event.preventDefault(); element.focus(); return }
        if (event.shiftKey && (document.activeElement === first || !element.contains(document.activeElement))) {
          event.preventDefault(); last.focus()
        } else if (!event.shiftKey && (document.activeElement === last || !element.contains(document.activeElement))) {
          event.preventDefault(); first.focus()
        }
      }
    }
    const onFocus = (event: FocusEvent) => {
      if (!element.contains(event.target as Node)) focusInside()
    }
    let active = true
    cleanup = () => {
      if (!active) return
      active = false
      document.removeEventListener('keydown', onKey, true)
      document.removeEventListener('focusin', onFocus)
      if (root) root.inert = wasInert
      html.style.overflow = overflow; html.style.scrollbarGutter = gutter
      window.scrollTo({ ...scroll, behavior: 'instant' })
      void nextTick(() => {
        if (panel.value && !disposed) return
        const target = trigger?.isConnected && !trigger.closest('[inert]') ? trigger : options.fallbackFocus?.()
        if (target?.isConnected) target.focus({ preventScroll: true })
      })
    }
    await nextTick()
    if (!active || disposed || panel.value !== element) return
    focusInside()
    if (root) root.inert = true
    document.addEventListener('keydown', onKey, true)
    document.addEventListener('focusin', onFocus)
  }, { flush: 'post' })
  onBeforeUnmount(() => { disposed = true; cleanup?.(); cleanup = undefined })
}
