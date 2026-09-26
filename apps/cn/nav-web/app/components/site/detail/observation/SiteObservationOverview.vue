<template>
  <div data-site-observation-overview class="space-y-6">
    <div class="flex flex-wrap items-baseline justify-between gap-2">
      <h3 class="site-overview-title">{{ t('siteObservation.currentTarget') }}</h3>
      <p data-site-observation-status>{{ presentation.statusLabel }}</p>
    </div>
    <dl class="site-observation-protocols">
      <div v-for="item in presentation.protocols" :key="item.protocol" :data-site-observation-protocol="item.protocol" class="grid grid-cols-2 gap-x-4 gap-y-2 sm:grid-cols-[5rem_1fr_1fr_2fr_1fr]">
        <dt class="site-overview-title">{{ item.protocol.toUpperCase() }}</dt>
        <dd>{{ item.statusLabel }}</dd>
        <dd><span class="site-detail-note block">{{ t('siteObservation.fields.duration') }}</span>{{ item.duration }}</dd>
        <dd><span class="site-detail-note block">{{ t('siteObservation.fields.observed') }}</span>{{ item.observed }}</dd>
        <dd data-site-protocol-freshness><span class="site-detail-note block">{{ t('siteObservation.freshness') }}</span>{{ item.freshness }}</dd>
      </div>
    </dl>
    <section v-if="presentation.endpoint.length" aria-labelledby="site-endpoint-title">
      <h3 id="site-endpoint-title" class="site-overview-title mb-2">{{ t('siteObservation.endpoint') }}</h3>
      <SiteObservationFacts :items="presentation.endpoint" />
    </section>
    <section v-if="presentation.risks.length" data-site-observation-risks class="site-overview-attention">
      <h3 class="site-overview-title">{{ t('siteObservation.risks') }}</h3>
      <ul class="mt-2 space-y-2"><li v-for="message in presentation.risks" :key="message" class="break-words">{{ message }}</li></ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import SiteObservationFacts from './SiteObservationFacts.vue'
defineProps<{ presentation: SiteObservationPresentation }>()
const { t } = useI18n()
</script>
