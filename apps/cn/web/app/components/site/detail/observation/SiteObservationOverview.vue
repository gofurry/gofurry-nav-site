<template>
  <div data-site-observation-overview data-site-observation-overview-composite class="site-observation-composite">
    <section data-site-observation-current>
      <h3 class="site-observation-heading">{{ t('siteObservation.currentObservation') }}</h3>
      <dl class="site-observation-protocol-strip mt-4 grid min-w-0 gap-4 md:grid-cols-3">
        <div v-for="item in presentation.protocols" :key="item.protocol" :data-site-observation-protocol="item.protocol" class="min-w-0">
          <dt class="site-observation-caption site-detail-protocol-heading flex flex-wrap items-center gap-2">
            <span class="site-detail-status-dot shrink-0" :data-tone="item.tone" aria-hidden="true" />{{ item.protocol.toUpperCase() }}
            <span data-site-protocol-status :class="{ 'sr-only': item.success }"><span v-if="!item.success" aria-hidden="true">· </span>{{ item.statusLabel }}</span>
          </dt>
          <dd class="site-observation-measure mt-2" :data-tone="item.durationTone"><span class="sr-only">{{ t('siteObservation.fields.duration') }}</span>{{ item.duration }}</dd>
          <dd data-site-protocol-freshness class="site-observation-caption mt-2"><span class="sr-only">{{ t('siteObservation.freshness') }}</span>{{ item.freshness }}</dd>
          <dd class="site-observation-caption mt-1 break-words"><span class="sr-only">{{ t('siteObservation.fields.observed') }}</span>{{ item.observed }}</dd>
        </div>
      </dl>
    </section>
    <div v-if="presentation.endpoint.length" class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
    <section v-if="presentation.endpoint.length" data-site-observation-endpoint aria-labelledby="site-endpoint-title">
      <h3 id="site-endpoint-title" class="site-observation-heading mb-3">{{ t('siteObservation.endpoint') }}</h3>
      <SiteObservationEvidence :items="presentation.endpoint" />
    </section>
    <div v-if="presentation.risks.length" class="site-detail-inset-separator mx-3 my-5" aria-hidden="true" />
    <section v-if="presentation.risks.length" data-site-observation-risks class="site-observation-attention-rail">
      <h3 class="site-observation-heading flex items-center gap-2"><PhWarning class="site-detail-icon" data-tone="warning" aria-hidden="true" />{{ t('siteObservation.risks') }}</h3>
      <ul class="mt-2 space-y-2"><li v-for="message in presentation.risks" :key="message" class="break-words">{{ message }}</li></ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import { PhWarning } from '@phosphor-icons/vue'
import SiteObservationEvidence from './SiteObservationEvidence.vue'
defineProps<{ presentation: SiteObservationPresentation }>()
const { t } = useI18n()
</script>
