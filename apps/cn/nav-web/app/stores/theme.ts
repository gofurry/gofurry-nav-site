export const useThemeStore = defineStore('theme', () => {
  const theme = ref<'light' | 'dark'>('light')

  function setTheme(value: 'light' | 'dark') {
    theme.value = value

    if (import.meta.client) {
      document.documentElement.classList.toggle('dark', value === 'dark')
      localStorage.setItem('theme', value)
    }
  }

  function initTheme() {
    if (!import.meta.client) return

    const saved = localStorage.getItem('theme')
    setTheme(saved === 'dark' ? 'dark' : 'light')
  }

  return {
    theme,
    setTheme,
    initTheme
  }
})
