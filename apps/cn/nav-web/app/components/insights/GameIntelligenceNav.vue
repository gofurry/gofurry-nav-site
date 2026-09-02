<template>
  <nav class="game-intelligence-nav" :aria-label="$t('insights.gameIntelligence.navLabel')">
    <NuxtLink v-for="item in items" :key="item.path" :to="localePath(item.path)" :class="{ active: isActive(item.path) }">
      {{ $t(item.label) }}
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
const route = useRoute()
const localePath = useLocalePath()
const items = [
  { path: '/insights/games', label: 'insights.gameIntelligence.ecosystem' },
  { path: '/insights/games/players', label: 'insights.gameIntelligence.players' },
  { path: '/insights/games/prices', label: 'insights.gameIntelligence.prices' },
  { path: '/insights/games/languages', label: 'insights.gameIntelligence.languages' },
]
function isActive(path: string) { return (route.path.replace(/^\/en(?=\/|$)/, '') || '/') === path }
</script>
