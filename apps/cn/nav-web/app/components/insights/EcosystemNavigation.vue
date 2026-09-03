<template>
  <div class="ecosystem-navigation" :data-context="context || 'none'">
    <InsightsNav />
    <SiteIntelligenceNav v-if="context === 'site'" />
    <GameIntelligenceNav v-else-if="context === 'game'" />
  </div>
</template>

<script setup lang="ts">
import GameIntelligenceNav from '@/components/insights/GameIntelligenceNav.vue'
import InsightsNav from '@/components/insights/InsightsNav.vue'
import SiteIntelligenceNav from '@/components/insights/SiteIntelligenceNav.vue'

defineProps<{
  context?: 'site' | 'game'
}>()
</script>
