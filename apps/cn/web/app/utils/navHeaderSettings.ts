export const NAV_HEADER_QUICK_ACCESS_STORAGE_KEY = 'nav-header-show-quick-access'
export const NAV_HEADER_SETTINGS_EVENT = 'nav-header-settings-change'

function hasBrowserStorage() {
  return typeof window !== 'undefined' && typeof localStorage !== 'undefined'
}

export function readShowQuickAccess() {
  if (!hasBrowserStorage()) {
    return true
  }

  return localStorage.getItem(NAV_HEADER_QUICK_ACCESS_STORAGE_KEY) !== '0'
}

export function writeShowQuickAccess(value: boolean) {
  if (!hasBrowserStorage()) {
    return
  }

  localStorage.setItem(NAV_HEADER_QUICK_ACCESS_STORAGE_KEY, value ? '1' : '0')
  window.dispatchEvent(new CustomEvent(NAV_HEADER_SETTINGS_EVENT, {
    detail: { showQuickAccess: value },
  }))
}

export function subscribeNavHeaderSettingsChange(callback: (payload: { showQuickAccess: boolean }) => void) {
  if (!hasBrowserStorage()) {
    return () => {}
  }

  const handleCustomEvent = (event: Event) => {
    const customEvent = event as CustomEvent<{ showQuickAccess?: boolean }>
    callback({
      showQuickAccess: customEvent.detail?.showQuickAccess ?? readShowQuickAccess(),
    })
  }

  const handleStorageEvent = (event: StorageEvent) => {
    if (event.key !== NAV_HEADER_QUICK_ACCESS_STORAGE_KEY) {
      return
    }

    callback({ showQuickAccess: event.newValue !== '0' })
  }

  window.addEventListener(NAV_HEADER_SETTINGS_EVENT, handleCustomEvent)
  window.addEventListener('storage', handleStorageEvent)

  return () => {
    window.removeEventListener(NAV_HEADER_SETTINGS_EVENT, handleCustomEvent)
    window.removeEventListener('storage', handleStorageEvent)
  }
}
