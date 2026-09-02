export function isGlobalSearchShortcut(event: Pick<KeyboardEvent, 'key' | 'metaKey' | 'ctrlKey'>, typing: boolean) {
  return ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') || (event.key === '/' && !typing)
}
