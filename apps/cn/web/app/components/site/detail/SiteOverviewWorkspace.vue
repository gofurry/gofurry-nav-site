<template>
  <div data-site-overview class="site-overview min-w-0 space-y-6">
    <div data-site-overview-composite class="site-detail-surface">
      <SiteOverviewHealth :health="presentation.health" :attention="presentation.attention" />
      <div class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
      <SiteCapabilitySnapshot :groups="presentation.capabilityGroups" :state="presentation.capabilityState" :insights-to="insightsTo" :retrying="retrying" @retry="emit('retry')" />
    </div>
    <SiteRecentChanges :items="presentation.recentChanges" :state="presentation.changesState" />
  </div>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import SiteOverviewHealth from './SiteOverviewHealth.vue'
import SiteCapabilitySnapshot from './SiteCapabilitySnapshot.vue'
import SiteRecentChanges from './SiteRecentChanges.vue'
defineProps<{ presentation: SiteOverviewPresentation; insightsTo: RouteLocationRaw; retrying: boolean }>()
const emit = defineEmits<{ retry: [] }>()
</script>
