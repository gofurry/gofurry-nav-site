import { isTransitionPageOpen, isTransitionSiteAlias, transitionNotice } from '#shared/utils/transitionMode'

export default defineNuxtRouteMiddleware(to => {
  if (useTransitionMode().value && !isTransitionPageOpen(to.path) && !isTransitionSiteAlias(to.path)) {
    throw createError({ statusCode: 503, statusMessage: 'Service temporarily unavailable', message: transitionNotice, fatal: true })
  }
})
