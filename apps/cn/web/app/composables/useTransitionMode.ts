import { isTransitionMode } from '#shared/utils/transitionMode'

export function useTransitionMode() {
  const config = useRuntimeConfig()
  return computed(() => isTransitionMode(config.public.transitionMode))
}
