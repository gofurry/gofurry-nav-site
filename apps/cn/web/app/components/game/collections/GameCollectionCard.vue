<template>
  <NuxtLink :to="localePath(`/games/collections/${collection.code}`)" class="game-collection-card game-collection-index-card flex min-w-0 flex-col overflow-hidden">
    <div class="game-collection-card__collage grid overflow-hidden" :style="{ gridTemplateColumns: `repeat(${Math.max(1, collection.preview_games.length)}, minmax(0, 1fr))` }">
      <SteamAssetImage v-for="game in collection.preview_games" :key="game.game_id" :src="game.header_url" :alt="game.name" loading="lazy" class="game-collection-image" />
      <span v-if="!collection.preview_games.length" class="game-collection-note flex items-center justify-center p-4 text-center">{{ t('game.collections.emptyWorks') }}</span>
    </div>
    <div class="flex flex-1 flex-col gap-3 px-2 py-3">
      <h2 class="game-collection-title line-clamp-2">{{ collection.name }}</h2>
      <p class="game-collection-copy line-clamp-3">{{ collection.info }}</p>
      <p class="game-collection-note mt-auto">{{ t('game.collections.count', { count: collection.visible_game_count }) }}</p>
    </div>
  </NuxtLink>
</template>
<script setup lang="ts">
import type { GameCollectionSummary } from '~/types/game'
import SteamAssetImage from '~/components/common/SteamAssetImage.vue'
defineProps<{ collection: GameCollectionSummary }>()
const { t } = useI18n()
const localePath = useLocalePath()
</script>
