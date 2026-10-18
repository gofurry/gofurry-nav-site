<template>
  <section data-site-overview-health data-site-overview-status-composite :data-site-summary-state="health.summaryState" :data-site-status="health.status" class="min-w-0" aria-labelledby="site-overview-health-title">
    <div data-site-overview-status-grid class="site-overview-status-grid grid min-w-0 grid-cols-2 gap-x-5 gap-y-4 md:grid-cols-3">
      <div class="site-overview-status-item col-span-2 min-w-0 md:col-span-1">
        <h3 id="site-overview-health-title" class="site-overview-heading">{{ t('siteOverview.health') }}</h3>
        <p class="site-overview-status-value mt-3 inline-flex items-center gap-2" :data-tone="health.tone"><span class="site-overview-dot shrink-0" aria-hidden="true" />{{ health.statusLabel }}</p>
      </div>
      <dl class="site-overview-status-item min-w-0">
        <dt class="site-overview-meta-label">{{ t('siteOverview.observedTargets') }}</dt>
        <dd data-site-overview-target-count class="site-overview-status-value mt-3 break-words">{{ health.targetCount === null ? t('siteOverview.countUnavailable') : t('siteOverview.targetCount', { count: health.targetCount }) }}</dd>
      </dl>
      <dl class="site-overview-status-item min-w-0">
        <dt class="site-overview-meta-label">{{ t('siteOverview.updated') }}</dt>
        <dd class="site-overview-updated mt-3 break-words"><time :datetime="health.generatedAt || undefined">{{ health.generatedLabel }}</time></dd>
      </dl>
    </div>
    <section v-if="attention.length" data-site-overview-attention class="site-overview-attention-row mt-4 flex flex-col gap-x-5 gap-y-2 sm:flex-row" aria-labelledby="site-attention-title">
      <h4 id="site-attention-title" class="site-overview-meta-label inline-flex shrink-0 items-start gap-2"><PhWarning class="site-detail-icon shrink-0" data-tone="warning" aria-hidden="true" />{{ t('siteOverview.attention') }}</h4>
      <ul class="min-w-0 flex-1 space-y-2">
        <li v-for="item in attention" :key="item.key" :data-site-attention-target="item.target" class="site-overview-attention-message break-words">{{ item.message }}</li>
      </ul>
    </section>
  </section>
</template>

<script setup lang="ts">
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import { PhWarning } from '@phosphor-icons/vue'
defineProps<{ health: SiteOverviewPresentation['health']; attention: SiteOverviewPresentation['attention'] }>()
const { t } = useI18n()
</script>
