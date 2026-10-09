<template>
  <NuxtLink :to="localePath(`/games/${item.game_id}`)" class="overview-game-visual block min-w-0" :aria-label="$t('insights.overviewVisuals.viewGame', { name })" data-overview-game-visual :data-game-id="item.game_id">
    <SteamAssetImage :src="item.visual.asset" :alt="name" width="460" height="215" :loading="priority ? 'eager' : 'lazy'" :fetchpriority="priority ? 'high' : 'auto'" class="overview-game-visual__image block aspect-[460/215] w-full" @error="$emit('failed')" />
    <span class="overview-game-visual__caption mt-3 block min-w-0">
      <span class="overview-game-visual__label block">{{ $t('insights.overviewVisuals.collectedGame') }}</span>
      <strong class="mt-1 block">{{ name }}</strong>
    </span>
  </NuxtLink>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import type { InsightFeaturedVisual } from '@/types/insights'
const props = defineProps<{ item: InsightFeaturedVisual, priority?: boolean }>()
defineEmits<{ failed: [] }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const name = computed(() => locale.value === 'en' ? props.item.name_en || props.item.name : props.item.name)
</script>
