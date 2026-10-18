<template>
  <section data-site-capability-detail :data-site-selected-metric="capability.key" class="site-detail-surface" aria-labelledby="site-capability-detail-title">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h3 id="site-capability-detail-title" class="site-overview-title">{{ capability.label }}</h3>
      <p data-site-selected-fact :data-site-capability-state="capability.state" :data-tone="capability.tone" class="site-overview-state inline-flex items-center gap-2 break-words"><span class="site-overview-dot" aria-hidden="true" />{{ capability.stateLabel }}</p>
    </div>
    <p class="site-detail-note mt-1">{{ t('siteIntelligence.factDate') }} · <time :datetime="capability.dateTime">{{ capability.date }}</time></p>
    <dl class="site-intelligence-facts mt-4 grid min-w-0 grid-cols-2 gap-4">
      <div class="min-w-0"><dt class="site-detail-label"><SiteInsightHelp :label="t('siteIntelligence.sameDayAdoption')" :help="t('siteIntelligence.adoptionHelp')" /></dt><dd data-site-selected-adoption class="site-intelligence-analysis-value mt-1">{{ capability.adoption }}</dd></div>
      <div class="min-w-0"><dt class="site-detail-label"><SiteInsightHelp :label="t('siteIntelligence.coverage')" :help="t('siteIntelligence.coverageHelp')" /></dt><dd data-site-selected-coverage class="site-intelligence-analysis-value mt-1">{{ capability.coverage }}</dd></div>
    </dl>
    <slot />
  </section>
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import SiteInsightHelp from './SiteInsightHelp.vue'
defineProps<{ capability: SiteInsightsPresentation['selected'] }>()
const { t } = useI18n()
</script>
