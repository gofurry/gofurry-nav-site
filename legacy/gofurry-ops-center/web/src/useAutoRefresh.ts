import { onBeforeUnmount, onMounted } from 'vue'

export function useAutoRefresh(load: () => void | Promise<void>, intervalMs = 30_000) {
  let timer: ReturnType<typeof setInterval> | undefined

  function clear() {
    if (timer) {
      clearInterval(timer)
      timer = undefined
    }
  }

  function start() {
    clear()
    if (document.hidden) return
    timer = setInterval(() => {
      if (!document.hidden) void load()
    }, intervalMs)
  }

  function handleVisibilityChange() {
    if (document.hidden) {
      clear()
      return
    }
    void load()
    start()
  }

  onMounted(() => {
    start()
    document.addEventListener('visibilitychange', handleVisibilityChange)
  })

  onBeforeUnmount(() => {
    clear()
    document.removeEventListener('visibilitychange', handleVisibilityChange)
  })
}
