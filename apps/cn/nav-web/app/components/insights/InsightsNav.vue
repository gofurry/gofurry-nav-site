<template>
  <nav class="insights-nav" :aria-label="$t('sidebar.insights')">
    <NuxtLink
      v-for="item in items"
      :key="item.path"
      :to="localePath(item.path)"
      class="insights-nav__link"
      :class="{ 'insights-nav__link--active': isActive(item.path) }"
    >
      {{ $t(item.label) }}
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
const route = useRoute()
const localePath = useLocalePath()

const items = [
  { label: 'insights.nav.overview', path: '/insights' },
  { label: 'insights.nav.sites', path: '/insights/sites' },
  { label: 'insights.nav.games', path: '/insights/games' },
]

function isActive(path: string) {
  const normalized = route.path.replace(/^\/en(?=\/|$)/, '') || '/'
  return normalized === path
}
</script>
