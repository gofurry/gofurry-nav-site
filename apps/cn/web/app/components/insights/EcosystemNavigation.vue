<template>
  <div class="ecosystem-navigation" :data-context="context || 'none'">
    <InsightsPrimaryNav />
    <InsightsDomainNav v-if="context" :domain="context" />
  </div>
</template>

<script setup lang="ts">
import InsightsPrimaryNav from './navigation/InsightsPrimaryNav.vue'
import InsightsDomainNav from './navigation/InsightsDomainNav.vue'

defineProps<{
  context?: 'site' | 'game'
}>()
</script>
