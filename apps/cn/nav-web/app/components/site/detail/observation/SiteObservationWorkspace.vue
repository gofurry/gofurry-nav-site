<template>
  <div data-site-observation :data-site-observation-target="presentation.target" class="site-observation min-w-0">
    <SiteObservationNav :active="view" @select="emit('select', $event)" />
    <div class="mt-3 flex flex-wrap justify-between gap-2"><p class="site-detail-note break-all">{{ t('siteObservation.currentTarget') }} · {{ presentation.target }}</p></div>
    <section id="site-observation-panel" :data-site-observation-view="view" role="tabpanel" :aria-labelledby="'observation-tab-' + view" tabindex="0" class="site-observation-panel mt-5 min-w-0">
      <SiteObservationOverview v-if="view === 'overview'" :presentation="presentation" />
      <SiteObservationPerformance v-else-if="view === 'performance'" :presentation="presentation" :history="history" @sample="emit('sample', $event)" @retry="emit('retry')" />
      <SiteObservationHttp v-else-if="view === 'http'" :presentation="presentation.http" />
      <SiteObservationDns v-else-if="view === 'dns'" :presentation="presentation.dns" />
      <SiteObservationWeb v-else :presentation="presentation.web" />
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationView } from '~/utils/siteDetailRouteState'
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import type { SiteHistoryPresentation, SiteHistorySample } from '~/composables/useSiteObservationHistory'
import SiteObservationNav from './SiteObservationNav.vue'
import SiteObservationOverview from './SiteObservationOverview.vue'
import SiteObservationPerformance from './SiteObservationPerformance.vue'
import SiteObservationHttp from './SiteObservationHttp.vue'
import SiteObservationDns from './SiteObservationDns.vue'
import SiteObservationWeb from './SiteObservationWeb.vue'
defineProps<{ view: SiteObservationView; presentation: SiteObservationPresentation; history: SiteHistoryPresentation }>()
const emit = defineEmits<{ select: [view: SiteObservationView]; sample: [sample: SiteHistorySample]; retry: [] }>()
const { t } = useI18n()
</script>
