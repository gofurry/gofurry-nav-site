export const useAppStore = defineStore('app', () => {
  const isNavigationRevealed = ref(true)

  function setNavigationRevealed(value: boolean) {
    isNavigationRevealed.value = value
  }

  return {
    isNavigationRevealed,
    setNavigationRevealed
  }
})
