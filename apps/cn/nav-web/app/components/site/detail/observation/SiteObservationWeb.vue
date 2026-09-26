<template>
  <div data-site-web class="space-y-6">
    <section data-site-web-metadata>
      <h3 class="site-overview-title mb-2">{{ t('siteObservation.sections.metadata') }}</h3>
      <SiteObservationFacts :items="presentation.metadata" />
    </section>
    <section v-for="probe in presentation.probes" :key="probe.protocol" :data-site-web-probe="probe.protocol" class="site-observation-probe">
      <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
        <h3 class="site-overview-title">{{ t('siteObservation.probes.' + probe.protocol) }}</h3>
        <p class="site-detail-note">{{ probe.statusLabel }} · {{ probe.duration }} · {{ probe.observed }}</p>
      </div>
      <div v-for="section in probe.sections" :key="section.key" class="mt-3">
        <details v-if="section.disclosure" class="site-observation-disclosure">
          <summary>{{ section.title }}</summary><SiteObservationFacts :items="section.items" class="mt-2" />
        </details>
        <template v-else>
          <h4 v-if="probe.protocol === 'page_assets'" class="site-detail-label mb-2">{{ section.title }}</h4>
          <SiteObservationFacts :items="section.items" />
        </template>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation['web'] }>()
const { t } = useI18n()
</script>
