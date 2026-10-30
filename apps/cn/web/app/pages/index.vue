<template>
  <TransitionHome v-if="transitionMode" />
  <NavHomePage v-else />
</template>

<script setup lang="ts">
import NavHomePage from '@/components/nav/NavHomePage.vue'
import TransitionHome from '~/components/transition/TransitionHome.vue'
const transitionMode = useTransitionMode()
</script>
