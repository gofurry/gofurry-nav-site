<template>
  <NuxtPage v-if="hasChildRoute" />
  <TransitionDetail v-else-if="transitionMode" kind="site" />
  <SiteDetailPage v-else />
</template>

<script setup lang="ts">
import SiteDetailPage from '@/components/site/SiteDetailPage.vue'
import TransitionDetail from '~/components/transition/TransitionDetail.vue'
import { parsePositiveEntityRouteId } from '~/utils/routeIdentity'

definePageMeta({
  validate: route => parsePositiveEntityRouteId(route.params.id) !== null,
})

const route = useRoute()
const transitionMode = useTransitionMode()
if (!parsePositiveEntityRouteId(route.params.id)) {
  throw createError({
    statusCode: 404,
    statusMessage: 'Site not found',
  })
}

const hasChildRoute = computed(() => {
  const domain = Array.isArray(route.params.domain) ? route.params.domain[0] : route.params.domain
  return typeof domain === 'string' && domain.trim().length > 0
})
</script>
