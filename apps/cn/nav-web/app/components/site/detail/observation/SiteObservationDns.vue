<template>
  <div data-site-dns class="space-y-6">
    <SiteObservationFacts :items="presentation.facts" />
    <section v-if="presentation.chains.length" data-site-dns-chain>
      <h3 class="site-overview-title">{{ t('siteObservation.resolution') }}</h3>
      <ol v-for="(chain, index) in presentation.chains" :key="index" class="site-observation-chain mt-3 flex flex-wrap gap-2">
        <li v-for="(value, step) in chain" :key="step" class="min-w-0 break-words"><span v-if="step" aria-hidden="true">→ </span>{{ value }}</li>
      </ol>
    </section>
    <section v-for="group in presentation.groups" :key="group.type" :data-site-dns-group="group.type">
      <h3 class="site-overview-title mb-2">{{ group.type }} <span class="site-detail-note">({{ group.records.length }})</span></h3>
      <div v-for="(record, index) in group.records" :key="index" class="site-observation-record">
        <div class="flex flex-wrap justify-between gap-2"><span class="min-w-0 break-all">{{ record.value }}</span><span class="site-detail-note">TTL {{ record.ttl }}</span></div>
        <details v-if="record.details.length" class="site-observation-disclosure mt-2">
          <summary>{{ t('siteObservation.recordDetails') }}</summary>
          <SiteObservationFacts :items="record.details" class="mt-2" />
        </details>
      </div>
    </section>
    <p v-if="!presentation.groups.length" class="site-detail-note">{{ t('siteObservation.noEvidence') }}</p>
    <section v-if="presentation.risks.length" data-site-dns-risks class="site-overview-attention">
      <h3 class="site-overview-title">{{ t('siteObservation.risks') }}</h3>
      <ul class="mt-2 space-y-1"><li v-for="risk in presentation.risks" :key="risk" class="break-words">{{ risk }}</li></ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation['dns'] }>()
const { t } = useI18n()
</script>
