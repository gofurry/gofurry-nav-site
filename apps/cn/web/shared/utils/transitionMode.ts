export const transitionNotice = '网站目前处于建设中，部分内容与服务暂未开放。'

export function isTransitionMode(value: unknown): boolean {
  return value === true || value === 'true'
}

export function transitionPagePath(path: string): string {
  return path.replace(/^\/en(?=\/|$)/, '').replace(/\/+$/, '') || '/'
}

export function isTransitionPageOpen(path: string): boolean {
  const page = transitionPagePath(path)
  return page === '/' || /^\/(site|games)\/[1-9]\d*$/.test(page)
}

// Keep the existing Site aliases, including the old target-address redirect.
export function isTransitionSiteAlias(path: string): boolean {
  return /^\/sites\/[1-9]\d*$/.test(transitionPagePath(path))
    || /^\/site\/[1-9]\d*\/[^/]+$/.test(transitionPagePath(path))
}
