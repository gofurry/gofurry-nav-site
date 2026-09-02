<template>
  <nav class="site-intelligence-nav" :aria-label="$t('insights.siteIntelligence.navLabel')">
    <NuxtLink v-for="item in items" :key="item.path" :to="localePath(item.path)" :class="{ active: isActive(item.path) }">
      {{ $t(item.label) }}
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
const route = useRoute()
const localePath = useLocalePath()
const items = [
  { path: '/insights/sites', label: 'insights.siteIntelligence.ecosystem' },
  { path: '/insights/sites/certificates', label: 'insights.siteIntelligence.certificates' },
  { path: '/insights/sites/compare', label: 'insights.siteIntelligence.compare' },
]
function isActive(path: string) { return (route.path.replace(/^\/en(?=\/|$)/, '') || '/') === path }
</script>
