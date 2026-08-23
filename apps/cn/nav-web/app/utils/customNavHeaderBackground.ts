type FileSystemPermissionMode = 'read' | 'readwrite'

interface FileSystemHandleLike {
  kind: 'file' | 'directory'
  name: string
  queryPermission?: (descriptor?: { mode?: FileSystemPermissionMode }) => Promise<PermissionState>
  requestPermission?: (descriptor?: { mode?: FileSystemPermissionMode }) => Promise<PermissionState>
}

export interface FileSystemFileHandleLike extends FileSystemHandleLike {
  kind: 'file'
  getFile: () => Promise<File>
}

export interface FileSystemDirectoryHandleLike extends FileSystemHandleLike {
  kind: 'directory'
  values: () => AsyncIterable<FileSystemHandleLike>
}

type PickerWindow = Window & {
  showDirectoryPicker?: () => Promise<FileSystemDirectoryHandleLike>
}

export interface CustomNavHeaderBackgroundSelection {
  handle: FileSystemDirectoryHandleLike
  folderName: string
}

export const CUSTOM_NAV_HEADER_BG_EVENT = 'custom-nav-header-bg-change'

const CUSTOM_NAV_HEADER_BG_META_KEY = 'customNavHeaderBgFolderName'
const DB_NAME = 'gofurry-custom-nav-header-bg'
const STORE_NAME = 'directoryHandles'
const STORE_KEY = 'nav-header-bg'
const CACHE_KEY = 'nav-header-bg-cache'

const imageExtensionPattern = /\.(avif|bmp|gif|jpeg|jpg|png|svg|webp)$/i

function dispatchCustomNavHeaderBackgroundChange() {
  window.dispatchEvent(new Event(CUSTOM_NAV_HEADER_BG_EVENT))
}

function openDatabase() {
  return new Promise<IDBDatabase>((resolve, reject) => {
    const request = window.indexedDB.open(DB_NAME, 1)

    request.onupgradeneeded = () => {
      const database = request.result
      if (!database.objectStoreNames.contains(STORE_NAME)) {
        database.createObjectStore(STORE_NAME)
      }
    }

    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

async function withStore<T>(
  mode: IDBTransactionMode,
  run: (store: IDBObjectStore) => IDBRequest<T> | void
) {
  const database = await openDatabase()

  return new Promise<T | undefined>((resolve, reject) => {
    const transaction = database.transaction(STORE_NAME, mode)
    const store = transaction.objectStore(STORE_NAME)
    const request = run(store)

    transaction.oncomplete = () => {
      database.close()
      if (request) {
        resolve(request.result)
      } else {
        resolve(undefined)
      }
    }

    transaction.onerror = () => {
      database.close()
      reject(transaction.error)
    }

    transaction.onabort = () => {
      database.close()
      reject(transaction.error)
    }
  })
}

async function ensureReadPermission(
  handle: FileSystemHandleLike,
  shouldRequest = false
) {
  const permission = await handle.queryPermission?.({ mode: 'read' })

  if (permission === 'granted') {
    return true
  }

  if (!shouldRequest || !handle.requestPermission) {
    return false
  }

  return (await handle.requestPermission({ mode: 'read' })) === 'granted'
}

async function saveDirectoryHandle(handle: FileSystemDirectoryHandleLike) {
  await withStore('readwrite', store => store.put(handle, STORE_KEY))
}

async function saveCachedBackground(file: File) {
  await withStore('readwrite', store => store.put(file, CACHE_KEY))
}

async function loadDirectoryHandle() {
  return (await withStore<FileSystemDirectoryHandleLike | undefined>(
    'readonly',
    store => store.get(STORE_KEY)
  )) ?? null
}

async function loadCachedBackground() {
  return (await withStore<Blob | undefined>(
    'readonly',
    store => store.get(CACHE_KEY)
  )) ?? null
}

async function deleteCachedBackground() {
  await withStore('readwrite', store => store.delete(CACHE_KEY))
}

async function deleteSavedBackground() {
  await withStore('readwrite', (store) => {
    store.delete(STORE_KEY)
    store.delete(CACHE_KEY)
  })
}

async function collectImageFiles(handle: FileSystemDirectoryHandleLike) {
  const files: File[] = []

  for await (const entry of handle.values()) {
    if (entry.kind === 'file') {
      const file = await (entry as FileSystemFileHandleLike).getFile()
      const isImage = file.type.startsWith('image/') || imageExtensionPattern.test(file.name)

      if (isImage) {
        files.push(file)
      }
      continue
    }

    files.push(...await collectImageFiles(entry as FileSystemDirectoryHandleLike))
  }

  return files
}

export function supportsCustomNavHeaderBackground() {
  if (typeof window === 'undefined') {
    return false
  }

  return typeof (window as PickerWindow).showDirectoryPicker === 'function' && 'indexedDB' in window
}

export function loadCustomNavHeaderBackgroundMeta() {
  const folderName = localStorage.getItem(CUSTOM_NAV_HEADER_BG_META_KEY) ?? ''

  return {
    folderName,
    enabled: Boolean(folderName),
  }
}

export async function pickCustomNavHeaderBackgroundDirectory() {
  if (!supportsCustomNavHeaderBackground()) {
    return null
  }

  const handle = await (window as PickerWindow).showDirectoryPicker?.()
  if (!handle) {
    return null
  }

  const hasPermission = await ensureReadPermission(handle, true)
  if (!hasPermission) {
    return null
  }

  return {
    handle,
    folderName: handle.name,
  } satisfies CustomNavHeaderBackgroundSelection
}

export async function saveCustomNavHeaderBackgroundDirectory(
  selection: CustomNavHeaderBackgroundSelection
) {
  await saveDirectoryHandle(selection.handle)
  await deleteCachedBackground()

  try {
    const files = await collectImageFiles(selection.handle)
    const randomFile = files[Math.floor(Math.random() * files.length)]
    if (randomFile) {
      await saveCachedBackground(randomFile)
    }
  } catch (error) {
    console.error('Cache custom nav header background err:', error)
  }

  localStorage.setItem(CUSTOM_NAV_HEADER_BG_META_KEY, selection.folderName)
  dispatchCustomNavHeaderBackgroundChange()
}

export async function clearCustomNavHeaderBackgroundDirectory() {
  await deleteSavedBackground()
  localStorage.removeItem(CUSTOM_NAV_HEADER_BG_META_KEY)
  dispatchCustomNavHeaderBackgroundChange()
}

export async function loadRandomCustomNavHeaderBackground() {
  if (typeof window === 'undefined' || !('indexedDB' in window)) {
    return null
  }

  const handle = await loadDirectoryHandle()
  if (handle && await ensureReadPermission(handle)) {
    try {
      const files = await collectImageFiles(handle)
      const randomFile = files[Math.floor(Math.random() * files.length)]
      if (randomFile) {
        await saveCachedBackground(randomFile)
        return URL.createObjectURL(randomFile)
      }
    } catch (error) {
      console.error('Read custom nav header background directory err:', error)
    }
  }

  const cachedBackground = await loadCachedBackground()
  return cachedBackground ? URL.createObjectURL(cachedBackground) : null
}
