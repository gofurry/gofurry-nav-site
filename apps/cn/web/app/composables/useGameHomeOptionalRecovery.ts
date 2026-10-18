import { computed, onBeforeUnmount, onMounted, shallowRef, watch, type Ref } from 'vue'

// undefined: this identity's SSR/navigation read is pending; null: unavailable.
// A valid empty payload is still success and never starts recovery.
export function useGameHomeOptionalRecovery<T>(
  identity: Ref<string>, snapshot: Ref<T | null | undefined>,
  load: (identity: string, signal: AbortSignal) => Promise<T>,
) {
  const nuxtApp = useNuxtApp()
  const recovered = shallowRef<T | null>(null)
  let mounted = false, attempted = false, generation = 0
  let controller: AbortController | undefined
  const invalidate = () => {
    generation++
    controller?.abort()
    controller = undefined
    recovered.value = null
  }
  const recover = () => {
    if (!mounted || attempted || snapshot.value !== null) return
    attempted = true
    const current = ++generation
    controller = new AbortController()
    void nuxtApp.runWithContext(() => load(identity.value, controller!.signal)).then(value => {
      if (mounted && current === generation) recovered.value = value
    }).catch(() => { /* Optional recovery is silent and never retries. */ })
  }
  watch(identity, () => { invalidate(); attempted = false }, { flush: 'sync' })
  watch([identity, snapshot], recover, { flush: 'post' })
  onMounted(() => { mounted = true; recover() })
  onBeforeUnmount(() => { mounted = false; invalidate() })
  return computed(() => snapshot.value ?? recovered.value)
}
