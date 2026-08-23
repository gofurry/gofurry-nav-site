import { normalizeDisplayMode, writeMode } from '@/utils/modeStorage'

export default defineNuxtRouteMiddleware((to) => {
  if (to.query.mode == null) {
    return
  }

  writeMode(normalizeDisplayMode(to.query.mode))

  const query = { ...to.query }
  delete query.mode

  return navigateTo(
    {
      path: to.path,
      query,
      hash: to.hash,
    },
    {
      replace: true,
    }
  )
})
