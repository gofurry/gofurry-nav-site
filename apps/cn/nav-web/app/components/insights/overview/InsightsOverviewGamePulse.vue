<template>
  <section class="overview-games" aria-labelledby="overview-games-title" data-overview-games>
    <div class="overview-section-heading">
      <div><p class="overview-kicker">{{ $t('insights.editorial.gamesKicker') }}</p><h2 id="overview-games-title">{{ $t('insights.editorial.gamesTitle') }}</h2></div>
      <NuxtLink :to="localePath('/insights/games')" class="overview-text-link">{{ $t('insights.editorial.explore') }} <span aria-hidden="true">↗</span></NuxtLink>
    </div>
    <InsightsGamePulse :panel="panel" />
  </section>
</template>

<script setup lang="ts">
import type { GameV2PanelRecord } from '@/types/game'
import InsightsGamePulse from '../domain/InsightsGamePulse.vue'
defineProps<{ panel: GameV2PanelRecord | null }>()
const localePath = useLocalePath()
</script>
