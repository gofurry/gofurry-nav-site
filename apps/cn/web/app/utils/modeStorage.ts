export type DisplayMode = 'sfw' | 'nsfw'

export const MODE_STORAGE_KEY = 'mode'
export const MODE_CHANGE_EVENT = 'mode-change'

export function normalizeDisplayMode(value: unknown): DisplayMode {
  const rawValue = Array.isArray(value) ? value[0] : value
  return typeof rawValue === 'string' && rawValue.trim().toLowerCase() === 'nsfw'
    ? 'nsfw'
    : 'sfw'
}

function canUseModeStorage() {
  return typeof window !== 'undefined' && typeof localStorage !== 'undefined'
}

export function readMode() {
  if (!canUseModeStorage()) {
    return ''
  }

  return localStorage.getItem(MODE_STORAGE_KEY)?.trim() ?? ''
}

export function readDisplayMode(): DisplayMode {
  return normalizeDisplayMode(readMode())
}

export function writeMode(value: string) {
  if (!canUseModeStorage()) {
    return
  }

  const trimmed = value.trim()

  if (trimmed) {
    localStorage.setItem(MODE_STORAGE_KEY, trimmed)
  } else {
    localStorage.removeItem(MODE_STORAGE_KEY)
  }

  dispatchModeChange(trimmed)
}

export function dispatchModeChange(mode = readMode()) {
  if (!canUseModeStorage()) {
    return
  }

  window.dispatchEvent(
    new CustomEvent(MODE_CHANGE_EVENT, {
      detail: {
        mode,
        displayMode: normalizeDisplayMode(mode),
      },
    })
  )
}

export function subscribeModeChange(
  callback: (payload: { mode: string; displayMode: DisplayMode }) => void
) {
  if (!canUseModeStorage()) {
    return () => {}
  }

  const handleStorage = (event: StorageEvent) => {
    if (event.key !== MODE_STORAGE_KEY) {
      return
    }

    const mode = event.newValue?.trim() ?? ''
    callback({
      mode,
      displayMode: normalizeDisplayMode(mode),
    })
  }

  const handleCustomEvent = (event: Event) => {
    const customEvent = event as CustomEvent<{ mode?: string; displayMode?: DisplayMode }>
    const mode = customEvent.detail?.mode?.trim() ?? readMode()
    callback({
      mode,
      displayMode: customEvent.detail?.displayMode ?? normalizeDisplayMode(mode),
    })
  }

  window.addEventListener('storage', handleStorage)
  window.addEventListener(MODE_CHANGE_EVENT, handleCustomEvent)

  return () => {
    window.removeEventListener('storage', handleStorage)
    window.removeEventListener(MODE_CHANGE_EVENT, handleCustomEvent)
  }
}
