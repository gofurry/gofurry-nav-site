import { isTransitionMode, isTransitionPageOpen, isTransitionSiteAlias, transitionNotice } from '../../shared/utils/transitionMode'

export default defineEventHandler(event => {
  if (!isTransitionMode(useRuntimeConfig(event).public.transitionMode)) return
  const path = getRequestURL(event).pathname
  // API proxies and runtime/static resources remain independent of page access.
  if (/^\/(api\/|_nuxt\/|_i18n\/|__nuxt|__vite|@|web\/|assets\/|\.well-known\/)/.test(path)
    || ['/healthz', '/robots.txt', '/sitemap.xml', '/manifest.webmanifest', '/llms.txt', '/favicon.ico', '/logo-mini.png', '/logo-mini.svg', '/defaultLogo.svg', '/og-image.jpg'].includes(path)) return
  // Never cache HTML across an operational mode switch (including open pages).
  setHeader(event, 'cache-control', 'no-store')
  if (isTransitionPageOpen(path) || isTransitionSiteAlias(path)) return
  setResponseStatus(event, 503, 'Service Unavailable')
  setHeader(event, 'retry-after', '3600')
  setHeader(event, 'x-robots-tag', 'noindex, nofollow')
  setHeader(event, 'content-type', 'text/html; charset=utf-8')
  return `<!doctype html><html lang="zh-CN"><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="robots" content="noindex, nofollow"><title>GoFurry · 建设中</title><body><main><h1>GoFurry</h1><p>${transitionNotice}</p></main></body></html>`
})
