export const BACKGROUND_PREFERENCE_KEY = 'gf_background_preference'
export const BACKGROUND_CHANGE_EVENT = 'gf-background-change'
export const DEFAULT_PATTERN_URL = '/web/background/gofurry-pattern.svg'
export type BackgroundOverrides = { color?: string; opacity?: number; size_px?: number }
export type BackgroundPreference = { version: 1; source: 'default' | 'server' | 'local'; pattern_id?: string; overrides: BackgroundOverrides }
export type LocalBackground = { blob: Blob; kind: 'svg' | 'raster'; name: string }
export type PatternDefaults = { light_color: string; dark_color: string; light_opacity: number; dark_opacity: number; default_size_px: number }

export function defaultBackgroundPreference(): BackgroundPreference { return { version: 1, source: 'default', overrides: {} } }
export function parseBackgroundPreference(raw: string | null): BackgroundPreference {
  try {
    const input = JSON.parse(raw || '')
    if (input?.version !== 1 || !['default', 'server', 'local'].includes(input.source)) return defaultBackgroundPreference()
    const result: BackgroundPreference = { version: 1, source: input.source, overrides: {} }
    if (result.source === 'server') {
      if (typeof input.pattern_id !== 'string' || !/^[1-9]\d*$/.test(input.pattern_id)) return defaultBackgroundPreference()
      result.pattern_id = input.pattern_id
    }
    const overrides = input.overrides ?? {}
    if (typeof overrides.color === 'string' && /^#[a-f\d]{6}$/i.test(overrides.color)) result.overrides.color = overrides.color
    if (typeof overrides.opacity === 'number' && Number.isFinite(overrides.opacity) && overrides.opacity >= 0 && overrides.opacity <= 1) result.overrides.opacity = overrides.opacity
    if (Number.isInteger(overrides.size_px) && overrides.size_px > 0 && overrides.size_px <= 10000) result.overrides.size_px = overrides.size_px
    return result
  } catch { return defaultBackgroundPreference() }
}

export function patternAppearance(defaults: PatternDefaults, theme: 'light' | 'dark', overrides: BackgroundOverrides) {
  return { color: overrides.color ?? defaults[`${theme}_color`], opacity: overrides.opacity ?? defaults[`${theme}_opacity`], size: overrides.size_px ?? defaults.default_size_px }
}
export function readBackgroundPreference() {
  if (typeof window === 'undefined') return defaultBackgroundPreference()
  try { return parseBackgroundPreference(localStorage.getItem(BACKGROUND_PREFERENCE_KEY)) } catch { return defaultBackgroundPreference() }
}

const databaseName = 'gofurry-page-background'
const storeName = 'local-background'
async function localStore<T>(mode: IDBTransactionMode, run: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  const database = await new Promise<IDBDatabase>((resolve, reject) => {
    const open = indexedDB.open(databaseName, 1)
    open.onupgradeneeded = () => { open.result.createObjectStore(storeName) }
    open.onsuccess = () => resolve(open.result)
    open.onerror = () => reject(new Error('无法打开本地背景存储'))
    open.onblocked = () => reject(new Error('本地背景存储被其他页面占用'))
  })
  return new Promise<T>((resolve, reject) => {
    const transaction = database.transaction(storeName, mode)
    const request = run(transaction.objectStore(storeName))
    transaction.oncomplete = () => { database.close(); resolve(request.result) }
    transaction.onerror = transaction.onabort = () => { database.close(); reject(new Error('无法保存或读取本地背景')) }
  })
}
export async function readLocalBackground(): Promise<LocalBackground | null> {
  const item = await localStore<LocalBackground | undefined>('readonly', (store) => store.get('selected'))
  return item?.blob instanceof Blob && (item.kind === 'svg' || item.kind === 'raster') ? item : null
}
export async function saveBackgroundPreference(preference: BackgroundPreference, file?: LocalBackground | null, removeLocal = false) {
  if (file) await localStore('readwrite', (store) => store.put(file, 'selected'))
  if (removeLocal) await localStore('readwrite', (store) => store.delete('selected'))
  // Normalize to the exact contract; never persist Blob data, object URLs or catalog defaults here.
  localStorage.setItem(BACKGROUND_PREFERENCE_KEY, JSON.stringify(parseBackgroundPreference(JSON.stringify(preference))))
  window.dispatchEvent(new Event(BACKGROUND_CHANGE_EVENT))
}

export function validateLocalSVG(text: string) {
  if (text.length > 512 * 1024 || /<!|<\?(?!xml\s)/i.test(text)) throw new Error('SVG 必须是自包含图案，不能带声明、实体或处理指令')
  const doc = new DOMParser().parseFromString(text, 'image/svg+xml')
  const allowed = new Set(['svg', 'g', 'path', 'rect', 'circle', 'ellipse', 'line', 'polyline', 'polygon', 'defs', 'use', 'symbol', 'pattern', 'mask', 'clipPath', 'title', 'desc'])
  if (doc.querySelector('parsererror') || doc.documentElement.localName !== 'svg') throw new Error('无效的 SVG 文件')
  for (const node of doc.querySelectorAll('*')) {
    if (!allowed.has(node.localName) || (node.namespaceURI && node.namespaceURI !== 'http://www.w3.org/2000/svg')) throw new Error('SVG 含有不支持的元素')
    for (const attr of node.attributes) {
      const name = attr.localName.toLowerCase(), value = attr.value.trim()
      if (attr.name === 'xmlns' || attr.name === 'xmlns:xlink') continue
      if (name.startsWith('on') || name === 'style' || name === 'base' || value.includes('\\') || value.includes(':') || /@import/i.test(value) || (name === 'href' && !/^#[\w-]+$/.test(value)) || /url\s*\(\s*(?!#[\w-]+\s*\))/i.test(value)) throw new Error('SVG 不允许脚本、样式或外部资源')
    }
  }
}

export async function prepareLocalBackground(file: File): Promise<LocalBackground> {
  if (!file.size || file.size > 10 * 1024 * 1024) throw new Error('请选择不超过 10 MiB 的图片')
  if (file.type === 'image/svg+xml' || file.name.toLowerCase().endsWith('.svg')) {
    const text = await file.text(); validateLocalSVG(text)
    return { blob: new Blob([text], { type: 'image/svg+xml' }), kind: 'svg', name: file.name }
  }
  if (!/^image\/(png|jpeg|webp|avif|gif|bmp)$/.test(file.type)) throw new Error('支持 SVG、PNG、JPEG、WebP、AVIF、GIF 或 BMP')
  const bitmap = await createImageBitmap(file)
  bitmap.close()
  return { blob: file, kind: 'raster', name: file.name }
}
