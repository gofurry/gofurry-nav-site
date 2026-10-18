<template>
  <nav :aria-label="t('game.collections.heading')" class="game-collection-shortcuts grid grid-cols-3 gap-2">
    <NuxtLink v-for="entry in entries" :key="entry.collection.code" :to="localePath(`/games/collections/${entry.collection.code}`)" class="sidebar-action-button game-collection-shortcut min-w-0 text-center">
      <span class="line-clamp-2">{{ entry.collection.name }}</span>
    </NuxtLink>
    <NuxtLink :to="localePath('/games/collections')" class="sidebar-action-button game-collection-shortcut min-w-0 text-center">{{ t('game.collections.all') }}</NuxtLink>
  </nav>
</template>
<script setup lang="ts">
import type { GameCollectionHome } from '~/types/game'
const props = defineProps<{ collections: GameCollectionHome }>()
const { t } = useI18n()
const localePath = useLocalePath()
const entries = computed(() => [...props.collections.slots].sort((a, b) => a.slot - b.slot))
</script>
