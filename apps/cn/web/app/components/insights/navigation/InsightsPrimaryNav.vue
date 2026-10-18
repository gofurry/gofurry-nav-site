<template>
  <nav class="insights-primary-nav" :aria-label="$t('insights.navLabel')">
    <NuxtLink
      v-for="item in insightsPrimaryItems"
      :key="item.path"
      :to="localePath(item.path)"
      :aria-current="isInsightsPrimaryActive(route.path, item.path) ? 'page' : undefined"
      class="insights-primary-nav__link"
      @focus="revealInsightsNavigationLink"
    >
      {{ $t(item.label) }}
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
import { insightsPrimaryItems, isInsightsPrimaryActive, revealInsightsNavigationLink } from './navigation'

const route = useRoute()
const localePath = useLocalePath()
</script>
