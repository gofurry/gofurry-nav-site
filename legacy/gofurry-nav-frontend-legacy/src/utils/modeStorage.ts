export type DisplayMode = 'sfw' | 'nsfw'

export const MODE_STORAGE_KEY = 'mode'
export const MODE_CHANGE_EVENT = 'mode-change'

export function readMode() {
  return localStorage.getItem(MODE_STORAGE_KEY)?.trim() ?? ''
}

export function readDisplayMode(): DisplayMode {
  return readMode() === 'nsfw' ? 'nsfw' : 'sfw'
}

export function writeMode(value: string) {
  const trimmed = value.trim()

  if (trimmed) {
    localStorage.setItem(MODE_STORAGE_KEY, trimmed)
  } else {
    localStorage.removeItem(MODE_STORAGE_KEY)
  }

  dispatchModeChange(trimmed)
}

export function dispatchModeChange(mode = readMode()) {
  window.dispatchEvent(
    new CustomEvent(MODE_CHANGE_EVENT, {
      detail: {
        mode,
        displayMode: mode === 'nsfw' ? 'nsfw' : 'sfw',
      },
    })
  )
}

export function subscribeModeChange(
  callback: (payload: { mode: string; displayMode: DisplayMode }) => void
) {
  const handleStorage = (event: StorageEvent) => {
    if (event.key !== MODE_STORAGE_KEY) {
      return
    }

    const mode = event.newValue?.trim() ?? ''
    callback({
      mode,
      displayMode: mode === 'nsfw' ? 'nsfw' : 'sfw',
    })
  }

  const handleCustomEvent = (event: Event) => {
    const customEvent = event as CustomEvent<{ mode?: string; displayMode?: DisplayMode }>
    const mode = customEvent.detail?.mode?.trim() ?? readMode()
    callback({
      mode,
      displayMode: customEvent.detail?.displayMode ?? (mode === 'nsfw' ? 'nsfw' : 'sfw'),
    })
  }

  window.addEventListener('storage', handleStorage)
  window.addEventListener(MODE_CHANGE_EVENT, handleCustomEvent)

  return () => {
    window.removeEventListener('storage', handleStorage)
    window.removeEventListener(MODE_CHANGE_EVENT, handleCustomEvent)
  }
}
