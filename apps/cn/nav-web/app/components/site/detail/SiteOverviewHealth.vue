<template>
  <section data-site-overview-health :data-site-summary-state="health.summaryState" :data-site-status="health.status" class="site-overview-health" aria-labelledby="site-overview-health-title">
    <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
      <h3 id="site-overview-health-title" class="site-overview-title">{{ t('siteOverview.health') }}</h3>
      <span class="site-detail-note">{{ t('siteOverview.scope') }}</span>
    </div>
    <div class="mt-3 flex flex-wrap items-baseline gap-x-4 gap-y-1">
      <p class="site-overview-health__status" :data-tone="health.tone">{{ health.statusLabel }}</p>
      <p v-if="health.freshnessLabel" class="site-overview-state" data-tone="warning">{{ health.freshnessLabel }}</p>
    </div>
    <div class="mt-2 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-2">
      <p v-if="health.summaryText" data-site-status-distribution class="site-overview-summary">{{ health.summaryText }}</p>
      <p class="site-detail-note">{{ t('siteOverview.updated') }} <time :datetime="health.generatedAt || undefined">{{ health.generatedLabel }}</time></p>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
defineProps<{ health: SiteOverviewPresentation['health'] }>()
const { t } = useI18n()
</script>
