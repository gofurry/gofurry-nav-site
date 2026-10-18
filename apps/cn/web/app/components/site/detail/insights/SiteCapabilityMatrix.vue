<template>
  <section data-site-capability-matrix class="min-w-0" aria-labelledby="site-matrix-title">
    <h3 id="site-matrix-title" class="site-overview-title">{{ t('siteIntelligence.matrix') }}</h3>
    <div class="site-intelligence-matrix-heading mt-3 flex flex-wrap gap-3 lg:grid lg:grid-cols-[minmax(0,1.3fr)_repeat(3,minmax(0,1fr))_minmax(0,1.1fr)]">
      <span class="hidden lg:block">{{ t('siteIntelligence.capability') }}</span><span class="hidden lg:block">{{ t('siteIntelligence.siteState') }}</span>
      <SiteInsightHelp :label="t('siteIntelligence.adoptionShort')" :help="t('siteIntelligence.adoptionHelp')" />
      <SiteInsightHelp :label="t('siteIntelligence.coverage')" :help="t('siteIntelligence.coverageHelp')" />
      <span class="hidden lg:block">{{ t('siteIntelligence.factDate') }}</span>
    </div>
    <section v-for="group in groups" :key="group.key" :data-site-insight-group="group.key" class="mt-4 space-y-1.5">
      <h4 class="site-detail-label mb-1">{{ group.label }}</h4>
    <button v-for="row in group.rows" :key="row.key" type="button" :data-site-capability="row.key" :data-site-capability-state="row.state"
      :data-site-capability-selected="selected === row.key" :aria-pressed="selected === row.key" :aria-label="row.label + ': ' + row.stateLabel"
      class="site-intelligence-row grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto] items-baseline gap-x-3 gap-y-1 text-left lg:grid-cols-[minmax(0,1.3fr)_repeat(3,minmax(0,1fr))_minmax(0,1.1fr)]" @click="emit('select', row.key)">
      <span class="site-intelligence-capability min-w-0 break-words">{{ row.label }}</span>
      <span class="site-overview-state inline-flex min-w-0 items-center gap-1 break-words" :data-tone="row.tone"><span class="site-overview-dot shrink-0" aria-hidden="true" />{{ row.stateLabel }}</span>
      <span class="site-intelligence-context col-span-2 flex flex-wrap gap-x-3 gap-y-1 lg:contents">
        <span class="inline-flex min-w-0 items-baseline gap-1"><span class="site-detail-note lg:hidden">{{ t('siteIntelligence.adoptionShort') }}</span><span data-site-capability-adoption>{{ row.adoption }}</span></span>
        <span class="inline-flex min-w-0 items-baseline gap-1"><span class="site-detail-note lg:hidden">{{ t('siteIntelligence.coverage') }}</span><span data-site-capability-coverage>{{ row.coverage }}</span></span>
        <time data-site-capability-date :datetime="row.dateTime" :aria-label="row.date" class="min-w-0 break-words"><template v-if="row.dateTime"><span class="lg:hidden">{{ row.shortDate }}</span><span class="hidden lg:inline">{{ row.date }}</span></template><template v-else>—</template></time>
      </span>
    </button>
    </section>
  </section>
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import type { SiteInsightMetric } from '~/utils/siteDetailRouteState'
import SiteInsightHelp from './SiteInsightHelp.vue'
defineProps<{ groups: SiteInsightsPresentation['groups']; selected: SiteInsightMetric }>()
const emit = defineEmits<{ select: [metric: SiteInsightMetric] }>()
const { t } = useI18n()
</script>
