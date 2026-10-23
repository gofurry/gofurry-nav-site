import { nextTick, onBeforeUnmount, watch } from 'vue'
import { useRoute } from '#app'
import { useThemeStore } from '@/stores/theme'
import { createThemeReveal } from '@/utils/themeReveal'

export function useThemeReveal() {
  const theme = useThemeStore()
  const route = useRoute()
  const reveal = createThemeReveal()
  watch(() => route.fullPath, reveal.cancel, { flush: 'sync' })
  onBeforeUnmount(reveal.dispose)
  return (event: MouseEvent) => {
    const target = theme.theme === 'light' ? 'dark' : 'light'
    reveal.toggle(event.currentTarget as HTMLElement | null, async () => {
      theme.setTheme(target)
      await nextTick()
    })
  }
}
