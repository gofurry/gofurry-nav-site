export const CUSTOM_PANEL_CODE_STORAGE_KEY = 'customPanelCode'
export const CUSTOM_PANEL_CODE_EVENT = 'custom-panel-code-change'
export const CUSTOM_PANEL_HEIGHT_STORAGE_KEY = 'customPanelHeight'
export const CUSTOM_PANEL_HEIGHT_EVENT = 'custom-panel-height-change'
export const DEFAULT_CUSTOM_PANEL_HEIGHT = 320

export function loadCustomPanelCode(): string {
  return localStorage.getItem(CUSTOM_PANEL_CODE_STORAGE_KEY)?.trim() ?? ''
}

export function saveCustomPanelCode(code: string) {
  const trimmed = code.trim()

  if (trimmed) {
    localStorage.setItem(CUSTOM_PANEL_CODE_STORAGE_KEY, trimmed)
  } else {
    localStorage.removeItem(CUSTOM_PANEL_CODE_STORAGE_KEY)
  }

  window.dispatchEvent(new Event(CUSTOM_PANEL_CODE_EVENT))
}

export function normalizeCustomPanelHeight(value: string | number): number {
  const parsed = typeof value === 'number' ? value : Number.parseInt(value.trim(), 10)

  if (!Number.isFinite(parsed)) {
    return DEFAULT_CUSTOM_PANEL_HEIGHT
  }

  return Math.min(1200, Math.max(200, parsed))
}

export function loadCustomPanelHeight(): number {
  const saved = localStorage.getItem(CUSTOM_PANEL_HEIGHT_STORAGE_KEY)

  if (!saved) {
    return DEFAULT_CUSTOM_PANEL_HEIGHT
  }

  return normalizeCustomPanelHeight(saved)
}

export function saveCustomPanelHeight(value: string | number) {
  const normalized = normalizeCustomPanelHeight(value)
  localStorage.setItem(CUSTOM_PANEL_HEIGHT_STORAGE_KEY, String(normalized))
  window.dispatchEvent(new Event(CUSTOM_PANEL_HEIGHT_EVENT))
}
