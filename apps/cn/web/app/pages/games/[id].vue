<template>
  <TransitionDetail v-if="transitionMode" kind="game" />
  <GameDetailPage v-else />
</template>

<script setup lang="ts">
import GameDetailPage from '~/components/game/GameDetailPage.vue'
import TransitionDetail from '~/components/transition/TransitionDetail.vue'
import { parsePositiveEntityRouteId } from '~/utils/routeIdentity'

definePageMeta({ validate: route => parsePositiveEntityRouteId(route.params.id) !== null })
const transitionMode = useTransitionMode()
</script>
