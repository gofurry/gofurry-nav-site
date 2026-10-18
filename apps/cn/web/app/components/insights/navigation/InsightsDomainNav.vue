<template>
  <nav class="insights-domain-nav" :data-domain="domain" :aria-label="$t(navLabel)">
    <NuxtLink
      v-for="item in items"
      :key="item.path"
      :to="localePath(item.path)"
      :aria-current="isInsightsDomainActive(route.path, item.path) ? 'page' : undefined"
      class="insights-domain-nav__link"
      @focus="revealInsightsNavigationLink"
    >
      {{ $t(item.label) }}
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { insightsDomainItems, isInsightsDomainActive, revealInsightsNavigationLink, type InsightsNavigationDomain } from './navigation'

const props = defineProps<{
  domain: InsightsNavigationDomain
}>()
const route = useRoute()
const localePath = useLocalePath()
const items = computed(() => insightsDomainItems[props.domain])
const navLabel = computed(() => props.domain === 'site'
  ? 'insights.siteIntelligence.navLabel'
  : 'insights.gameIntelligence.navLabel')
</script>
