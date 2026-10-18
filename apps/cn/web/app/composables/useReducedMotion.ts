import { onBeforeUnmount, onMounted, readonly, ref } from 'vue'

// A consumer-owned subscription, not SSR state or a global animation controller.
export function useReducedMotion() {
  const reduced = ref(false)
  let media: MediaQueryList | undefined
  const sync = () => { reduced.value = media?.matches ?? false }

  onMounted(() => {
    media = window.matchMedia('(prefers-reduced-motion: reduce)')
    sync()
    media.addEventListener('change', sync)
  })
  onBeforeUnmount(() => media?.removeEventListener('change', sync))

  return readonly(reduced)
}
